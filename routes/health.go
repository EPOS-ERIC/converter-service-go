package routes

import (
	"fmt"
	"net/http"

	"github.com/epos-eu/converter-service/connection"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

type HealthHandler struct {
	RabbitConn *amqp.Connection
}

// Health check
//
//	@Summary		Check the health of the service
//	@Description	Check the health of the RabbitMQ connection and the database connection
//	@Tags			health
//	@Produce		json
//	@Success		200	{string}	string	"Healthy"
//	@Failure		500	{object}	HTTPError
//	@Router			/health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	err := health(h.RabbitConn)
	if err != nil {
		c.String(http.StatusInternalServerError, "Unhealthy: %w", err)
		return
	} else {
		c.String(http.StatusOK, "Healthy")
		return
	}
}

func health(rabbitConn *amqp.Connection) error {
	// Check the rabbit connection
	_, err := rabbitConn.Channel()
	if err != nil {
		// Unhealthy: rabbit not connected
		return fmt.Errorf("rabbit not connected")
	}

	// Check the connection to the db
	_, err = connection.Connect()
	if err != nil {
		return fmt.Errorf("can't connect to database")
	}

	return nil
}
