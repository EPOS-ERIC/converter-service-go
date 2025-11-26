package routes

import (
	"fmt"
	"net/http"

	"github.com/epos-eu/converter-service/db"
	"github.com/epos-eu/converter-service/logging"
	"github.com/epos-eu/converter-service/rabbit"
	"github.com/gin-gonic/gin"
)

var healthLog = logging.Get("health")

type HealthHandler struct {
	Broker *rabbit.BrokerConfig
}

// Health check
func (h *HealthHandler) Health(c *gin.Context) {
	err := health(h.Broker)
	if err != nil {
		healthLog.Error("health check failed", "error", err)
		c.String(http.StatusServiceUnavailable, "Unhealthy: ", err.Error())
		return
	} else {
		c.String(http.StatusOK, "Healthy")
		return
	}
}

func health(broker *rabbit.BrokerConfig) error {
	ch, err := broker.Conn.Channel()
	if err != nil {
		return fmt.Errorf("can't open rabbit channel: %w", err)
	}
	err = ch.Close()
	if err != nil {
		return fmt.Errorf("can't close rabbit channel: %w", err)
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
