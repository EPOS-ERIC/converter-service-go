package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// brokerConfig holds the RabbitMQ connection configuration
type brokerConfig struct {
	host     string
	user     string
	password string
	vhost    string
}

// connectToBroker dials RabbitMQ and returns the established connection
func connectToBroker(cfg brokerConfig) (*amqp.Connection, error) {
	return amqp.Dial("amqp://" + cfg.user + ":" + cfg.password + "@" + cfg.host + "/" + cfg.vhost)
}

// initExternalAccessQueue declares the "externalAccess" exchange, the "map" queue, binds them,
// and sets QoS on the channel. Returns the declared queue
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
		"map",
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = channel.QueueBind(
		q.Name,           // queue
		"#.map",          // routing key
		"externalAccess", // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	err = channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	return q
}

// initProcessingServiceQueue declares the "processService" exchange, the "processing" queue, binds them,
// and sets QoS on the channel. Returns the declared queue
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
		"processing",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	err = channel.QueueBind(
		q.Name,
		"#.processing",
		"processService",
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	err = channel.Qos(
		1,
		0,
		false,
	)
	failOnError(err, "Failed to set QoS")

	return q
}

// initResourcesServiceQueue declares the "metadataService" exchange, the "resources" queue, binds them,
// and sets QoS on the channel. Returns the declared queue
func initResourcesServiceQueue(channel *amqp.Channel) amqp.Queue {
	err := channel.ExchangeDeclare(
		"metadataService",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := channel.QueueDeclare(
		"resources",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	err = channel.QueueBind(
		q.Name,
		"#.resources",
		"metadataService",
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	err = channel.Qos(
		1,
		0,
		false,
	)
	failOnError(err, "Failed to set QoS")

	return q
}
