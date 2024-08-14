package main

import (
	"net/http"
	"time"

	"github.com/seunghoon34/collaborative-coding-platform/internal/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/ws/:roomCode", handlers.HandleWebSocket)
	r.POST("/execute", gin.HandlerFunc(func(c *gin.Context) {
		// Set a timeout for code execution
		c.Request = c.Request.WithContext(c.Request.Context())
		done := make(chan struct{})
		go func() {
			handlers.ExecuteCode(c)
			close(done)
		}()
		select {
		case <-done:
			return
		case <-time.After(30 * time.Second):
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "Code execution timed out"})
		}
	}))

	r.Run(":8080")
}
