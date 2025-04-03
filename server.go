package main

import (
	_ "embed"
	"net/http"

	"github.com/epos-eu/converter-service/loggers"
	"github.com/epos-eu/converter-service/routes"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Embed the OpenAPI 3.0 JSON specification file
//
//go:embed openapi.json
var openAPISpec []byte

// startServer initializes the Gin engine and starts listening on :8080.
// The RabbitMQ connection is passed for health checks.
func startServer(conn *amqp.Connection) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Routes
	v1 := r.Group("/api/converter-service/v1")
	{
		// Plugin CRUD endpoints with consistent naming
		v1.POST("/plugins", routes.CreatePlugin)
		v1.GET("/plugins", routes.GetAllPlugins)
		v1.GET("/plugins/:plugin_id", routes.GetPlugin)
		v1.PUT("/plugins/:plugin_id", routes.UpdatePlugin)
		v1.DELETE("/plugins/:plugin_id", routes.DeletePlugin)

		// Plugin Relations CRUD endpoints
		v1.POST("/plugin-relations", routes.CreatePluginRelation)
		v1.GET("/plugin-relations", routes.GetAllPluginRelations)
		v1.GET("/plugin-relations/:plugin_id", routes.GetPluginRelation)
		v1.PUT("/plugin-relations/:plugin_id", routes.UpdatePluginRelation)
		v1.DELETE("/plugin-relations/:plugin_id", routes.DeletePluginRelation)

		// Enable and disable plugins
		v1.POST("/plugins/:plugin_id/enable", routes.EnablePlugin)
		v1.POST("/plugins/:plugin_id/disable", routes.DisablePlugin)

		// Health check injecting the RabbitMQ connection
		healthHandler := routes.HealthHandler{
			RabbitConn: conn,
		}
		v1.GET("/actuator/health", healthHandler.Health)

		v1.GET("/api-docs", func(c *gin.Context) {
			c.Data(http.StatusOK, "application/json", openAPISpec)
		})
	}

	//	@title		Converter Service API
	//	@version	1.0
	//	@BasePath	/api/converter-service/v1

	// Swagger endpoints

	err := r.Run(":8080")
	if err != nil {
		loggers.API_LOGGER.Error("Error initializing server", "error", err)
	}
}
