package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/seunghoon34/collaborative-coding-platform/internal/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Adjust this for production!
	},
}

func HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}

	roomCode := c.Param("roomCode")
	username := c.Query("username")

	log.Printf("New WebSocket connection: Room: %s, Username: %s", roomCode, username)

	client := &models.Client{
		Conn:     conn,
		Username: username,
		Send:     make(chan []byte, 256),
	}

	room := models.GetOrCreateRoom(roomCode)
	room.RegisterClient(client)

	log.Printf("Client %s registered in room %s", username, roomCode)

	// Send initial user list
	room.BroadcastUserList()

	go client.WritePump()
	go client.ReadPump(room)

	// Send a ping message every 5 seconds
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()
}
