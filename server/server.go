package server

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"time"

	"github.com/epos-eu/converter-service/rabbit"
	"github.com/epos-eu/converter-service/server/routes"
	"github.com/gin-gonic/gin"
)

// Embed the OpenAPI 3.0 JSON specification file
//
//go:embed openapi.json
var openAPISpec []byte

type LogEntry struct {
	Component string `json:"component"`
	Time      string `json:"time"`
	Level     string `json:"level"`
	Status    int    `json:"status"`
	Latency   string `json:"latency"`
	ClientIP  string `json:"client_ip"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	BodySize  int    `json:"body_size"`
	Error     string `json:"error,omitempty"`
}

// startServer initializes the Gin engine and starts listening on :8080.
// The RabbitMQ connection is passed for health checks.
func StartServer(broker *rabbit.BrokerConfig) {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	r := gin.New()
	r.Use(gin.Recovery())
	// use a custom json logger for gin
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		entry := LogEntry{
			Component: "API",
			Time:      param.TimeStamp.Format(time.RFC3339),
			Level:     "INFO",
			Status:    param.StatusCode,
			Latency:   param.Latency.String(),
			ClientIP:  param.ClientIP,
			Method:    param.Method,
			Path:      param.Path,
			BodySize:  param.BodySize,
			Error:     param.ErrorMessage,
		}

		jsonLog, _ := json.Marshal(entry)
		return string(jsonLog) + "\n"
	}))

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
		v1.GET("/plugin-relations/:relation_id", routes.GetPluginRelation)
		v1.PUT("/plugin-relations/:relation_id", routes.UpdatePluginRelation)
		v1.DELETE("/plugin-relations/:relation_id", routes.DeletePluginRelation)

		// Enable and disable plugins
		v1.POST("/plugins/:plugin_id/enable", routes.EnablePlugin)
		v1.POST("/plugins/:plugin_id/disable", routes.DisablePlugin)

		// Health check injecting the RabbitMQ connection
		healthHandler := routes.HealthHandler{
			RabbitConn: broker.Conn,
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
		panic(err)
	}
}
