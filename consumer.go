package main

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/epos-eu/converter-service/connection"
	"github.com/epos-eu/converter-service/handler"
	"github.com/epos-eu/converter-service/loggers"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handleExternalAccessMsgs(ch *amqp.Channel, msgs <-chan amqp.Delivery) {
	for m := range msgs {
		loggers.EA_LOGGER.Info("Received message")
		newRouting := strings.Split(m.RoutingKey, ".")
		routingReturn := buildRoutingKey(newRouting, "access_return")

		loggers.EA_LOGGER.Info("Handling message")
		response, err := handler.Handler(string(m.Body))
		if err != nil {
			loggers.EA_LOGGER.Error("Failed to convert the message", "error", err)
			err := publishError(ch, "externalAccess", routingReturn, m, err)
			if err != nil {
				loggers.EA_LOGGER.Error("Failed to publish the error message", "error", err)
			}

			loggers.EA_LOGGER.Info("Message handled with error")
			continue
		}

		loggers.EA_LOGGER.Debug("Sending converted message", "message", string(response))
		err = ch.Publish(
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
			loggers.EA_LOGGER.Error("Failed to publish the converted message", "error", err)
			continue
		}

		loggers.EA_LOGGER.Info("Message handled")
	}
}

// TODO
func handleProcessingServiceMsgs(ch *amqp.Channel, msgs <-chan amqp.Delivery) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for m := range msgs {
		loggers.PS_LOGGER.Info("Received message")
		newRouting := strings.Split(m.RoutingKey, ".")
		routingReturn := buildRoutingKey(newRouting, "processing_return")

		// TODO: handle message
		loggers.PS_LOGGER.Info("Handling RESOURCES message")

		loggers.PS_LOGGER.Info("Sending converted message")
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
			loggers.PS_LOGGER.Error("Failed to publish the converted message", "error", err)
			continue
		}

		loggers.PS_LOGGER.Info("Message handled")
	}
}

type Relation struct {
	PluginID     string `json:"pluginId"`
	InputFormat  string `json:"inputFormat"`
	OutputFormat string `json:"outputFormat"`
}

type Plugin struct {
	DistributionID string     `json:"distributionId"`
	Relations      []Relation `json:"relations"`
}

type PluginRelation []Plugin

type ResourcesMsg struct {
	Plugins string `json:"plugins"`
}

func handleResourcesServiceMsgs(ch *amqp.Channel, msgs <-chan amqp.Delivery) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for m := range msgs {
		loggers.RS_LOGGER.Info("Received message")
		newRouting := strings.Split(m.RoutingKey, ".")
		routingReturn := buildRoutingKey(newRouting, "map_return")

		var resourcesMsg ResourcesMsg
		err := json.Unmarshal(m.Body, &resourcesMsg)
		if err != nil || resourcesMsg.Plugins != "all" {
			loggers.RS_LOGGER.Error("Failed to process the message", "error", err)
			err := publishError(ch, "metadataService", routingReturn, m, err)
			if err != nil {
				loggers.RS_LOGGER.Error("Failed to publish the error message", "error", err)
			}

			loggers.RS_LOGGER.Info("Message handled with error")
			continue
		}

		// get all plugin relations
		relations, err := connection.GetPluginRelationForEnabledPlugins()
		if err != nil {
			loggers.RS_LOGGER.Error("Failed to get plugin relations", "error", err)
			err := publishError(ch, "metadataService", routingReturn, m, err)
			if err != nil {
				loggers.RS_LOGGER.Error("Failed to publish the error message", "error", err)
			}

			loggers.RS_LOGGER.Info("Message handled with error")
			continue
		}
		// group them by OperationID
		operations := make(map[string][]Relation)
		for _, relation := range relations {
			operations[relation.RelationID] = append(operations[relation.RelationID], Relation{
				PluginID:     relation.PluginID,
				InputFormat:  relation.InputFormat,
				OutputFormat: relation.OutputFormat,
			})
		}

		responseStr := make([]Plugin, 0)
		for k, v := range operations {
			responseStr = append(responseStr, Plugin{
				DistributionID: k,
				Relations:      v,
			})
		}

		response, err := json.Marshal(responseStr)
		if err != nil {
			loggers.RS_LOGGER.Error("Failed to marshal response", "error", err)
			err := publishError(ch, "metadataService", routingReturn, m, err)
			if err != nil {
				loggers.RS_LOGGER.Error("Failed to publish the error message", "error", err)
			}

			loggers.RS_LOGGER.Info("Message handled with error")
			continue
		}

		// loggers.RS_LOGGER.Debug("Sending converted message", "message", response)
		err = ch.PublishWithContext(
			ctx,
			"metadataService",
			routingReturn,
			false,
			false,
			amqp.Publishing{
				ContentType:   "application/json",
				CorrelationId: m.CorrelationId,
				Body:          response,
				Headers:       m.Headers,
			},
		)
		if err != nil {
			loggers.RS_LOGGER.Error("Failed to publish the converted message", "error", err)
			continue
		}

		loggers.RS_LOGGER.Info("Message handled")
	}
}

func buildRoutingKey(sections []string, suffix string) string {
	if len(sections) == 0 {
		return suffix
	}
	return strings.Join(sections[:len(sections)-1], ".") + "." + suffix
}

func publishError(ch *amqp.Channel, exchange, routingKey string, d amqp.Delivery, err error) error {
	return ch.Publish(
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
