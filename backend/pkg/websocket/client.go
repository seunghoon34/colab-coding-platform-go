package models

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Room struct {
	Code        string
	Clients     map[*Client]bool
	CurrentCode string
	mutex       sync.Mutex
}

type Client struct {
	Conn     *websocket.Conn
	Send     chan []byte
	Username string
	RoomCode string
}

type Message struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

var (
	rooms = make(map[string]*Room)
	mutex sync.Mutex
)

func GetOrCreateRoom(code string) *Room {
	mutex.Lock()
	defer mutex.Unlock()

	if room, exists := rooms[code]; exists {
		return room
	}

	room := &Room{
		Code:    code,
		Clients: make(map[*Client]bool),
	}
	rooms[code] = room
	return room
}

func (r *Room) RegisterClient(client *Client, isHost bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.Clients[client] = isHost
}

func (r *Room) UnregisterClient(client *Client) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.Clients[client]; ok {
		delete(r.Clients, client)
		close(client.Send)
	}
}

func (r *Room) BroadcastMessage(message []byte) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for client := range r.Clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(r.Clients, client)
		}
	}
}

func (r *Room) BroadcastUserList() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	users := make([]map[string]interface{}, 0, len(r.Clients))
	for client, isHost := range r.Clients {
		users = append(users, map[string]interface{}{
			"username": client.Username,
			"isHost":   isHost,
		})
	}

	message, _ := json.Marshal(Message{
		Type:    "userList",
		Content: users,
	})

	r.BroadcastMessage(message)
}

func (c *Client) ReadPump(room *Room) {
	defer func() {
		room.UnregisterClient(c)
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("error: %v", err)
			continue
		}

		switch msg.Type {
		case "code":
			room.CurrentCode = msg.Content.(string)
			room.BroadcastMessage(message)
		case "cursor":
			room.BroadcastMessage(message)
		case "getUserList":
			room.BroadcastUserList()
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
