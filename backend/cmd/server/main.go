package main

import (
	"github.com/seunghoon34/collaborative-coding-platform/internal/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	r := gin.Default()

	// Add CORS middleware
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"} // Adjust this to match your frontend URL
	config.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	r.Use(cors.New(config))

	r.POST("/create-room", handlers.CreateRoom)
	r.GET("/ws/:roomCode", handlers.HandleWebSocket)
	r.POST("/execute", handlers.ExecuteCode)

	r.Run("0.0.0.0:8080")
}
