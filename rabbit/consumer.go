package rabbit

import (
	"fmt"
	"os"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (b *BrokerConfig) handleMessages(
	queue *amqp.Queue,
	exchangeName string,
	routingKeySuffix string,
	handler func([]byte) ([]byte, error),
) {
	hostname, _ := os.Hostname()
	consumerTag := fmt.Sprintf("%s-%s-%d", queue.Name, hostname, time.Now().Unix())

	msgs, err := b.consumeChan.Consume(
		queue.Name,
		consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "consume "+queue.Name)

	// launch a new goroutine for each message received. We can assume we won't have more than maxMessages
	// goroutines at the same time because we set qos for the channel
	for d := range msgs {
		go func(delivery amqp.Delivery) {
			logger.Info("message received", "exchange", exchangeName, "queue", queue.Name)

			resp, err := handler(delivery.Body)
			if err != nil {
				logger.Error("handler failed", "error", err)
				err = delivery.Nack(false, false) // don't re‑queue for retry
				if err != nil {
					logger.Error("error nack-ing", "error", err)
				}
				return
			}

			logger.Debug("message handled successfully")

			rk := buildRoutingKey(delivery.RoutingKey, routingKeySuffix)
			err = b.publishChan.Publish(
				exchangeName,
				rk,
				false,
				false,
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: delivery.CorrelationId,
					Body:          resp,
					Headers:       delivery.Headers,
				},
			)
			if err != nil {
				logger.Error("publish failed", "error", err)
				err = delivery.Nack(false, false) // don't re‑queue for retry
				if err != nil {
					logger.Error("error nack-ing", "error", err)
				}
				return
			}

			logger.Debug("message sent successfully")

			if err = delivery.Ack(false); err != nil {
				logger.Error("ack failed", "error", err)
				return
			}

			logger.Debug("message acknowledged successfully")
		}(d)
	}
}

func buildRoutingKey(in, suffix string) string {
	parts := strings.Split(in, ".")
	if len(parts) == 0 {
		return suffix
	}
	return strings.Join(parts[:len(parts)-1], ".") + "." + suffix
}
