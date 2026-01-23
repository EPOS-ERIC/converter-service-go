package server

import (
	_ "embed"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/epos-eu/converter-service/logging"
	"github.com/epos-eu/converter-service/rabbit"
	"github.com/epos-eu/converter-service/server/routes"
	"github.com/gin-gonic/gin"
)

//go:embed openapi.json
var openAPISpec []byte

var log = logging.Get("server")

// slogGinMiddleware creates a structured logging middleware using the logging package
func slogGinMiddleware() gin.HandlerFunc {
	// Get a logger specifically for HTTP requests
	httpLog := logging.Get("http")

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if strings.Contains(path, "actuator/health") {
			return
		}

		latency := time.Since(start)
		clientIP := c.ClientIP()
		if raw != "" {
			path = path + "?" + raw
		}

		status := c.Writer.Status()
		var level slog.Level
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		default:
			level = slog.LevelInfo
		}

		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.String("client_ip", clientIP),
			slog.Duration("latency", latency),
			slog.Int64("response_size", int64(c.Writer.Size())),
		}

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			attrs = append(attrs,
				slog.String("error", err.Error()),
				slog.Any("error_type", err.Type),
			)
		}

		httpLog.LogAttrs(c.Request.Context(), level, "HTTP request", attrs...)
	}
}

// customRecoveryMiddleware handles panics with structured logging
func customRecoveryMiddleware() gin.HandlerFunc {
	recoveryLog := logging.Get("recovery")

	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		recoveryLog.Error("panic recovered",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("client_ip", c.ClientIP()),
			slog.Any("panic", recovered),
		)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// StartServer initializes the Gin engine and starts listening on :8080.
// The RabbitMQ connection is passed for health checks.
func StartServer(broker *rabbit.BrokerConfig) {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	r := gin.New()

	r.Use(customRecoveryMiddleware())

	r.Use(slogGinMiddleware())

	// Routes
	v1 := r.Group("/api/converter-service/v1")
	{
		// Plugin CRUD endpoints
		v1.POST("/plugins", routes.CreatePlugin)
		v1.GET("/plugins", routes.GetAllPlugins)
		v1.GET("/plugins/:plugin_id", routes.GetPlugin)
		v1.PUT("/plugins/:plugin_id", routes.UpdatePlugin)
		v1.DELETE("/plugins/:plugin_id", routes.DeletePlugin)

		// Plugin Relations CRUD endpoints
		v1.POST("/plugin-relations", routes.CreatePluginRelation)
		v1.GET("/plugin-relations", routes.GetAllPluginRelations)
		v1.GET("/plugin-relations/:relation_id", routes.GetPluginRelation)
		v1.PUT("/plugin-relations/:relation_id", routes.UpdatePluginRelation)
		v1.DELETE("/plugin-relations/distribution/:relation_id", routes.DeleteRelationsByDistributionID)
		v1.DELETE("/plugin-relations/:relation_id", routes.DeletePluginRelation)

		// Distribution endpoints
		v1.GET("/distributions/:instance_id", routes.GetDistributionByInstanceID)

		// Enable and disable plugins
		v1.POST("/plugins/:plugin_id/enable", routes.EnablePlugin)
		v1.POST("/plugins/:plugin_id/disable", routes.DisablePlugin)

		// Health check
		healthHandler := routes.HealthHandler{
			Broker: broker,
		}
		v1.GET("/actuator/health", healthHandler.Health)

		v1.GET("/api-docs", func(c *gin.Context) {
			c.Data(http.StatusOK, "application/json", openAPISpec)
		})
	}

	log.Info("starting server", slog.String("port", "8080"))

	//	@title		Converter Service API
	//	@version	1.0
	//	@BasePath	/api/converter-service/v1

	err := r.Run(":8080")
	if err != nil {
		log.Error("failed to start server", slog.Any("error", err))
		panic(err)
	}
}
