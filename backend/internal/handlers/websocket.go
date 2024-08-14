package handlers

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/seunghoon34/collaborative-coding-platform/internal/models"
	"github.com/seunghoon34/collaborative-coding-platform/pkg/websocket"
)

func HandleWebSocket(c *gin.Context) {
	conn, err := websocket.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	roomCode := c.Param("roomCode")
	username := c.Query("username")

	client := websocket.NewClient(conn, username, roomCode)

	room := models.GetOrCreateRoom(roomCode)
	models.RegisterClient(room, &client.Client)

	log.Printf("New client connected to room %s: %s", roomCode, username)

	go client.ReadPump(func(message []byte) {
		models.BroadcastMessage(room, message)
	})
	go client.WritePump()
}
