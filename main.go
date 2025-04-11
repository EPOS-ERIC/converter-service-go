package main

import (
	"log"
	"os"

	"github.com/epos-eu/converter-service/loggers"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func main() {
	loggers.InitSlog()

	// Read broker configuration from environment or fall back to defaults
	config := brokerConfig{
		host:     getEnv("BROKER_HOST", "rabbitmq"),
		user:     getEnv("BROKER_USERNAME", "changeme"),
		password: getEnv("BROKER_PASSWORD", "changeme"),
		vhost:    getEnv("BROKER_VHOST", "changeme"),
	}

	// Connect to RabbitMQ
	conn, err := connectToBroker(config)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// Open channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Initialize queues
	externalAccessQueue := initExternalAccessQueue(ch)
	// processingServiceQueue := initProcessingServiceQueue(ch) // TODO
	resourcesServiceQueue := initResourcesServiceQueue(ch)

	// Consumers
	externalAccessMsgs, err := ch.Consume(
		externalAccessQueue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register External Access consumer")

	// TODO
	// processingServiceMsgs, err := ch.Consume(
	// 	processingServiceQueue.Name,
	// 	"",
	// 	false,
	// 	false,
	// 	false,
	// 	false,
	// 	nil,
	// )
	// failOnError(err, "Failed to register Processing Service consumer")

	resourcesServiceMsgs, err := ch.Consume(
		resourcesServiceQueue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register Resources Service consumer")

	// Consume from each queue
	go handleExternalAccessMsgs(ch, externalAccessMsgs)
	// go handleProcessingServiceMsgs(ch, processingServiceMsgs)	// TODO
	go handleResourcesServiceMsgs(ch, resourcesServiceMsgs)

	go startServer(conn)

	select {}
}
