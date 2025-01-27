package main

import (
	"context"
	"strings"
	"time"

	"github.com/epos-eu/converter-service/handler"
	"github.com/epos-eu/converter-service/loggers"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handleExternalAccessMsgs(ch *amqp.Channel, msgs <-chan amqp.Delivery) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for m := range msgs {
		loggers.EA_LOGGER.Println("Received message")
		newRouting := strings.Split(m.RoutingKey, ".")
		routingReturn := buildRoutingKey(newRouting, "access_return")

		loggers.EA_LOGGER.Println("Handling message")
		response, err := handler.Handler(string(m.Body))
		if err != nil {
			loggers.EA_LOGGER.Printf("Failed to convert the message: %v\n", err)
			err := publishError(ch, ctx, "externalAccess", routingReturn, m, err)
			if err != nil {
				loggers.EA_LOGGER.Printf("Failed to publish the error message: %v\n", err)
			}

			err = m.Ack(true)
			if err != nil {
				loggers.EA_LOGGER.Printf("Error acknowledging: %v\n", err)
				continue
			}
			loggers.EA_LOGGER.Println("Message handled with error\n")
			continue
		}

		loggers.EA_LOGGER.Println("Sending converted message")
		err = ch.PublishWithContext(
			ctx,
			"externalAccess",
			routingReturn,
			false,
			false,
			amqp.Publishing{
				ContentType:   "application/json",
				CorrelationId: m.CorrelationId,
				Body:          []byte(response),
				Headers:       m.Headers,
			},
		)
		if err != nil {
			loggers.EA_LOGGER.Printf("Failed to publish the converted message: %v\n", err)
			continue
		}

		err = m.Ack(true)
		if err != nil {
			loggers.EA_LOGGER.Printf("Error acknowledging: %v\n", err)
			continue
		}
		loggers.EA_LOGGER.Println("Message handled\n")
	}
}

// Example consumer for processingService
func handleProcessingServiceMsgs(ch *amqp.Channel, msgs <-chan amqp.Delivery) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for m := range msgs {
		loggers.PS_LOGGER.Println("Received message")
		newRouting := strings.Split(m.RoutingKey, ".")
		routingReturn := buildRoutingKey(newRouting, "processing_return")

		// TODO: handle message

		loggers.PS_LOGGER.Println("Sending converted message")
		err := ch.PublishWithContext(
			ctx,
			"processService",
			routingReturn,
			false,
			false,
			amqp.Publishing{
				ContentType:   "application/json",
				CorrelationId: m.CorrelationId,
				Body:          []byte("TODO"),
				Headers:       m.Headers,
			},
		)
		if err != nil {
			loggers.PS_LOGGER.Printf("Failed to publish the converted message: %v\n", err)
			return
		}

		err = m.Ack(true)
		if err != nil {
			loggers.PS_LOGGER.Printf("Error acknowledging: %v\n", err)
			return
		}
		loggers.PS_LOGGER.Println("Message handled\n")
	}
}

// Example consumer for resourcesService
func handleResourcesServiceMsgs(ch *amqp.Channel, msgs <-chan amqp.Delivery) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for m := range msgs {
		loggers.RS_LOGGER.Println("Received message")
		newRouting := strings.Split(m.RoutingKey, ".")
		routingReturn := buildRoutingKey(newRouting, "metadata_service")

		// TODO: handle message

		loggers.RS_LOGGER.Println("Sending converted message")
		err := ch.PublishWithContext(
			ctx,
			"metadataService",
			routingReturn,
			false,
			false,
			amqp.Publishing{
				ContentType:   "application/json",
				CorrelationId: m.CorrelationId,
				Body:          []byte("TODO"),
				Headers:       m.Headers,
			},
		)
		if err != nil {
			loggers.RS_LOGGER.Printf("Failed to publish the converted message: %v\n", err)
			return
		}

		err = m.Ack(true)
		if err != nil {
			loggers.RS_LOGGER.Printf("Error acknowledging: %v\n", err)
			return
		}
		loggers.RS_LOGGER.Println("Message handled\n")
	}
}

func buildRoutingKey(sections []string, suffix string) string {
	if len(sections) == 0 {
		return suffix
	}
	return strings.Join(sections[:len(sections)-1], ".") + "." + suffix
}

// Helper to publish error messages back
func publishError(ch *amqp.Channel, ctx context.Context, exchange, routingKey string, d amqp.Delivery, err error) error {
	return ch.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: d.CorrelationId,
			Body:          []byte(err.Error()),
			Headers:       d.Headers,
		},
	)
}
