package server

import (
	"github.com/gin-gonic/gin"
	"github.com/seunghoon34/collaborative-coding-platform/internal/handlers"
)

type Server struct {
	router *gin.Engine
}

func NewServer() *Server {
	r := gin.Default()
	s := &Server{
		router: r,
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.GET("/ws/:roomCode", handlers.HandleWebSocket)
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
