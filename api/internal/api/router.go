package api

import (
	"github.com/dustinleblanc/go-bespin/internal/websocket"
	"github.com/gin-gonic/gin"
)

// SetupRouter sets up the API router
func SetupRouter(handlers *Handlers, wsServer *websocket.Server) *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Webhook-Signature, X-Webhook-Event")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// WebSocket handler
	r.GET("/socket.io/*any", func(c *gin.Context) {
		wsServer.ServeWs(c.Writer, c.Request)
	})

	// API routes
	api := r.Group("/api")
	{
		webhooks := api.Group("/webhooks")
		{
			webhooks.POST("/:source", handlers.ReceiveWebhook)
			webhooks.GET("/:id", handlers.GetWebhook)
			webhooks.GET("", handlers.ListWebhooks)
		}
	}

	return r
}
