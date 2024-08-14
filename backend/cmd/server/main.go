package main

import (
	"log"

	"github.com/seunghoon34/collaborative-coding-platform/internal/server"
)

func main() {
	s := server.NewServer()
	if err := s.Run(":8080"); err != nil {
		log.Fatal("Failed to run server: ", err)
	}
}
