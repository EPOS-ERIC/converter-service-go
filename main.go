package main

import (
	"context"
	"github.com/epos-eu/converter-service/docs"
	"github.com/epos-eu/converter-service/handler"
	"github.com/epos-eu/converter-service/loggers"
	"github.com/epos-eu/converter-service/routes"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

type brokerConfig struct {
	host     string
	user     string
	password string
	vhost    string
}

func main() {
	loggers.EA_LOGGER.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	loggers.PS_LOGGER.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	// Default config overridden by env variables
	config := brokerConfig{
		host:     "rabbitmq",
		user:     "changeme",
		password: "changeme",
		vhost:    "changeme",
	}

	if val, res := os.LookupEnv("BROKER_HOST"); res == true {
		config.host = val
	}

	if val, res := os.LookupEnv("BROKER_USERNAME"); res == true {
		config.user = val
	}

	if val, res := os.LookupEnv("BROKER_PASSWORD"); res == true {
		config.password = val
	}

	if val, res := os.LookupEnv("BROKER_VHOST"); res == true {
		config.vhost = val
	}
	conn, err := amqp.Dial("amqp://" + config.user + ":" + config.password + "@" + config.host + "/" + config.vhost)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	externalAccessQueue := initExternalAccessQueue(ch)

	externalAccessMsgs, err := ch.Consume(
		externalAccessQueue.Name, // queue
		"",                       // consumer
		false,                    // auto-ack
		false,                    // exclusive
		false,                    // no-local
		false,                    // no-wait
		nil,                      // args
	)
	failOnError(err, "Failed to register a consumer")

	processingServiceQueue := initProcessingServiceQueue(ch)

	processingServiceMsgs, err := ch.Consume(
		processingServiceQueue.Name, // queue
		"",                          // consumer
		false,                       // auto-ack
		false,                       // exclusive
		false,                       // no-local
		false,                       // no-wait
		nil,                         // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for d := range externalAccessMsgs {
			loggers.EA_LOGGER.Println("Received message")
			newRouting := strings.Split(d.RoutingKey, ".")
			routingReturn := ""
			for section := range newRouting {
				if section < len(newRouting)-1 {
					routingReturn += newRouting[section] + "."
				}
			}
			routingReturn += "access_return"
			loggers.EA_LOGGER.Println("Handling message")
			response, err := handler.Handler(string(d.Body))
			if err != nil {
				loggers.EA_LOGGER.Printf("Failed to convert the message: %v\n", err)
				err = ch.PublishWithContext(ctx,
					"externalAccess", // exchange
					routingReturn,    // routing key
					false,            // mandatory
					false,            // immediate
					amqp.Publishing{
						ContentType:   "application/json",
						CorrelationId: d.CorrelationId,
						Body:          []byte(err.Error()),
						Headers:       d.Headers,
					})
				if err != nil {
					loggers.EA_LOGGER.Printf("Failed to publish the error message: %v\n", err)
					continue
				}

				err = d.Ack(true)
				if err != nil {
					loggers.EA_LOGGER.Printf("Error acknowledging: %v\n", err)
					continue
				}
				loggers.EA_LOGGER.Println("Message handled with error\n")
				continue
			}

			loggers.EA_LOGGER.Println("Sending converted message")
			err = ch.PublishWithContext(ctx,
				"externalAccess", // exchange
				routingReturn,    // routing key
				false,            // mandatory
				false,            // immediate
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          []byte(response),
					Headers:       d.Headers,
				})
			if err != nil {
				loggers.EA_LOGGER.Printf("Failed to publish the converted message: %v\n", err)
				continue
			}

			err = d.Ack(true)
			if err != nil {
				loggers.EA_LOGGER.Printf("Error acknowledging: %v\n", err)
				continue
			}
			loggers.EA_LOGGER.Println("Message handled\n")
		}
	}()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for d := range processingServiceMsgs {
			loggers.PS_LOGGER.Println("Received message")
			newRouting := strings.Split(d.RoutingKey, ".")
			routingReturn := ""
			for section := range newRouting {
				if section < len(newRouting)-1 {
					routingReturn += newRouting[section] + "."
				}
			}
			routingReturn += "processing_return"

			// TODO: handle message

			loggers.PS_LOGGER.Println("Sending converted message")
			err = ch.PublishWithContext(ctx,
				"processService", // exchange
				routingReturn,    // routing key
				false,            // mandatory
				false,            // immediate
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          []byte("TODO"),
					Headers:       d.Headers,
				})
			if err != nil {
				loggers.PS_LOGGER.Printf("Failed to publish the converted message: %v\n", err)
				return
			}

			err = d.Ack(true)
			if err != nil {
				loggers.PS_LOGGER.Printf("Error acknowledging: %v\n", err)
				return
			}
			loggers.PS_LOGGER.Println("Message handled\n")
		}
	}()

	// Start the service
	go func() {
		loggers.SERVICE_LOGGER.Printf("Server started")

		r := gin.Default()
		docs.SwaggerInfo.BasePath = "/api/converter-service/v1"

		// Routes
		v1 := r.Group("/api/converter-service/v1")
		{
			v1.GET("/plugins", routes.GetAllPlugins)
			v1.GET("/plugins/:id", routes.GetPlugin)

			v1.GET("/plugin-relations", routes.GetAllPluginRelations)
			v1.GET("/plugin-relations/:id", routes.GetPluginRelations)

			// Enable and disable plugins
			v1.POST("/plugins/:id/enable", routes.EnablePlugin)
			v1.POST("/plugins/:id/disable", routes.DisablePlugin)

			// Health check injecting the rabbit connection
			healthHandler := routes.HealthHandler{
				RabbitConn: conn,
			}
			v1.GET("/health", healthHandler.Health)
		}

		//	@title		Converter Service API
		//	@version	1.0
		//	@BasePath	/api/converter-service/v1

		// Swagger
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
		r.GET("/api/converter-service/v1/api-docs", func(c *gin.Context) {
			c.Redirect(http.StatusPermanentRedirect, "/swagger/doc.json")
		})
		r.GET("/api/converter-service/v1", func(c *gin.Context) {
			c.Redirect(http.StatusPermanentRedirect, "/swagger/index.html")
		})

		err := r.Run(":8080")
		if err != nil {
			loggers.SERVICE_LOGGER.Printf("ERROR: %v", err)
		}
	}()

	log.Printf(" [*] Converter Service ready to accept requests")
	<-forever
}

func initExternalAccessQueue(channel *amqp.Channel) amqp.Queue {
	err := channel.ExchangeDeclare(
		"externalAccess", // name
		"topic",          // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := channel.QueueDeclare(
		"map", // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = channel.QueueBind(
		q.Name,           // queue name
		"#.map",          // routing key
		"externalAccess", // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

	err = channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	return q
}

func initProcessingServiceQueue(channel *amqp.Channel) amqp.Queue {
	err := channel.ExchangeDeclare(
		"processService", // name
		"topic",          // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := channel.QueueDeclare(
		"processing", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = channel.QueueBind(
		q.Name,           // queue name
		"#.processing",   // routing key
		"processService", // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

	err = channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	return q
}
