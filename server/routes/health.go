package routes

import (
	"fmt"
	"net/http"

	"github.com/epos-eu/converter-service/db"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

type HealthHandler struct {
	RabbitConn *amqp.Connection
}

// Health check
func (h *HealthHandler) Health(c *gin.Context) {
	err := health(h.RabbitConn)
	if err != nil {
		c.String(http.StatusInternalServerError, "Unhealthy: ", err.Error())
		return
	} else {
		c.String(http.StatusOK, "Healthy")
		return
	}
}

func health(rabbitConn *amqp.Connection) error {
	_, err := rabbitConn.Channel()
	if err != nil {
		return fmt.Errorf("rabbit not connected")
	}

	db := db.Get()

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("can't get underlying sql.DB: %w", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		return fmt.Errorf("can't ping database: %w", err)
	}

	return nil
}
