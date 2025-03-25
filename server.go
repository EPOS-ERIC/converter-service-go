package main

import (
	"net/http"

	"github.com/epos-eu/converter-service/docs"
	"github.com/epos-eu/converter-service/loggers"
	"github.com/epos-eu/converter-service/routes"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// startServer initializes the Gin engine and starts listening on :8080.
// The RabbitMQ connection is passed for health checks
func startServer(conn *amqp.Connection) {
	r := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/converter-service/v1"

	// Routes
	v1 := r.Group("/api/converter-service/v1")
	{
		// Plugin CRUDs
		v1.POST("/plugin", routes.CreatePlugin)
		v1.GET("/plugins", routes.GetAllPlugins)
		v1.GET("/plugins/:id", routes.GetPlugin)
		v1.PUT("/plugins/:id", routes.UpdatePlugin)
		v1.DELETE("/plugins/:id", routes.DeletePlugin)

		// Plugin Relations CRUDs
		v1.POST("/plugin-relation", routes.CreatePluginRelation)
		v1.GET("/plugin-relations", routes.GetAllPluginRelations)
		v1.GET("/plugin-relations/:id", routes.GetPluginRelations)
		v1.PUT("/plugin-relations/:id", routes.UpdatePluginRelation)
		v1.DELETE("/plugin-relations/:id", routes.DeletePluginRelation)

		// Enable and disable plugins
		v1.POST("/plugins/:id/enable", routes.EnablePlugin)
		v1.POST("/plugins/:id/disable", routes.DisablePlugin)

		// Health check injecting the rabbit connection
		healthHandler := routes.HealthHandler{
			RabbitConn: conn,
		}
		v1.GET("/health", healthHandler.Health)
	}

	// @title		Converter Service API
	// @version		1.0
	// @BasePath	/api/converter-service/v1

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	r.GET("/api/converter-service/v1/api-docs", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/swagger/doc.json")
	})
	r.GET("/api/converter-service/v1", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/swagger/index.html")
	})

	err := r.Run(":8080")
	if err != nil {
		loggers.API_LOGGER.Error("Error initializing server", "error", err)
	}
}
