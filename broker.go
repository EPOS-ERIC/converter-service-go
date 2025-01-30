package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type brokerConfig struct {
	host     string
	user     string
	password string
	vhost    string
}

// connectToBroker dials RabbitMQ and returns the connection
func connectToBroker(cfg brokerConfig) (*amqp.Connection, error) {
	return amqp.Dial("amqp://" + cfg.user + ":" + cfg.password + "@" + cfg.host + "/" + cfg.vhost)
}

func initExternalAccessQueue(channel *amqp.Channel) amqp.Queue {
	err := channel.ExchangeDeclare(
		"externalAccess",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := channel.QueueDeclare(
		"map",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	err = channel.QueueBind(
		q.Name,
		"#.map",
		"externalAccess",
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

// TODO
func initProcessingServiceQueue(channel *amqp.Channel) amqp.Queue {
	err := channel.ExchangeDeclare(
		"processService",
		"topic",
		true,
		false,
		false,
		false,
		nil,
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
		"resurces",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	err = channel.QueueBind(
		q.Name,
		"#.map",
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
