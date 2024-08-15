package handlers

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"

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

	client := &models.Client{
		Conn:     conn,
		Username: username,
	}

	room := models.GetRoomManager().GetOrCreateRoom(roomCode)
	room.RegisterClient(client)

	log.Printf("New client %s connected to room %s", username, roomCode)

	defer room.UnregisterClient(client)

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var message map[string]interface{}
		if err := json.Unmarshal(p, &message); err != nil {
			log.Println("Error unmarshaling message:", err)
			continue
		}

		switch message["type"] {
		case "code":
			room.BroadcastMessage(p)
		case "language":
			room.BroadcastMessage(p)
		case "chat":
			chatMessage := models.ChatMessage{
				Type:     "chat",
				Username: username,
				Content:  message["content"].(string),
			}
			chatJSON, _ := json.Marshal(chatMessage)
			room.BroadcastMessage(chatJSON)
		default:
			log.Printf("Unknown message type: %v", message["type"])
		}
	}
}

func generateRoomCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	code := make([]byte, 6)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

func CreateRoom(c *gin.Context) {
	var request struct {
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if request.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}
	roomCode := generateRoomCode()
	models.GetOrCreateRoom(roomCode)
	c.JSON(http.StatusOK, gin.H{"roomCode": roomCode})
}
