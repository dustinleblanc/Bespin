package api

import (
	"github.com/dustinleblanc/go-bespin-api/internal/queue"
	"github.com/dustinleblanc/go-bespin-api/internal/webhook"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// NewRouter creates a new router with all routes configured
func NewRouter(jobQueue queue.Queue, webhookService *webhook.Service) *gin.Engine {
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Event-Type", "X-Signature"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Create handlers
	handlers := NewHandlers(jobQueue, webhookService)

	// API routes
	api := router.Group("/api")
	{
		// Random text generation
		api.GET("/random-text", handlers.HandleRandomText)
		api.GET("/jobs/:id", handlers.HandleGetJobResult)

		// Webhooks
		api.POST("/webhooks/:source", handlers.HandleWebhook)

		// WebSocket
		api.GET("/ws", handlers.HandleWebSocket)
	}

	return router
}
