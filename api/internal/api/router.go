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
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

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
		api.GET("/", handlers.GetRoot)
		api.GET("/test", handlers.GetTest)

		jobs := api.Group("/jobs")
		{
			jobs.GET("/test", handlers.GetJobsTest)
			jobs.POST("/random-text", handlers.CreateRandomTextJob)
		}
	}

	return r
}
