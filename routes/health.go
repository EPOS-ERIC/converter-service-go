package routes

import (
	"fmt"
	"github.com/epos-eu/converter-service/connection"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"net/http"
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
func (this *HealthHandler) Health(c *gin.Context) {
	err := health(c, this.RabbitConn)
	if err != nil {
		c.String(http.StatusInternalServerError, "Unhealthy: "+err.Error())
		return
	} else {
		c.String(http.StatusOK, "Healthy")
		return
	}
}

func health(c *gin.Context, rabbitConn *amqp.Connection) error {
	// Check the rabbit connection
	_, err := rabbitConn.Channel()
	if err != nil {
		// Unhealthy: rabbit not connected
		return fmt.Errorf("rabbit not connected")
	}

	// Check the connection to the db
	db, err := connection.Connect()
	if err != nil {
		return fmt.Errorf("can't connect to database")
	}
	err = db.Ping(c)
	if err != nil {
		return fmt.Errorf("can't connect to database")
	}

	return nil
}

