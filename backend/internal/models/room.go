package models

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Room struct {
	Code    string
	Clients map[*Client]bool
	mutex   sync.Mutex
}

type Client struct {
	Conn     *websocket.Conn
	Username string
	Send     chan []byte
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
		log.Printf("Existing room found: %s", code)
		return room
	}

	log.Printf("Creating new room: %s", code)
	room := &Room{
		Code:    code,
		Clients: make(map[*Client]bool),
	}
	rooms[code] = room
	return room
}

func (r *Room) RegisterClient(client *Client) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.Clients[client] = true
	log.Printf("Client registered: %s in room %s", client.Username, r.Code)
}

func (r *Room) UnregisterClient(client *Client) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.Clients[client]; ok {
		delete(r.Clients, client)
		close(client.Send)
		log.Printf("Client unregistered: %s from room %s", client.Username, r.Code)
	}
}

func (r *Room) BroadcastMessage(message []byte) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for client := range r.Clients {
		err := client.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Error broadcasting message to client: %v", err)
			delete(r.Clients, client)
			client.Conn.Close()
		}
	}
}

func (r *Room) BroadcastUserList() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	users := make([]string, 0, len(r.Clients))
	for client := range r.Clients {
		users = append(users, client.Username)
	}

	message, err := json.Marshal(Message{
		Type:    "userList",
		Content: users,
	})
	if err != nil {
		log.Printf("Error marshaling user list: %v", err)
		return
	}

	log.Printf("Broadcasting user list: %v", users)
	r.BroadcastMessage(message)
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				log.Printf("Send channel closed for client %s", c.Username)
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("Error getting next writer for client %s: %v", c.Username, err)
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				log.Printf("Error closing writer for client %s: %v", c.Username, err)
				return
			}
			log.Printf("Message written to client %s: %s", c.Username, string(message))
		}
	}
}

func (c *Client) ReadPump(room *Room) {
	defer func() {
		room.UnregisterClient(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)

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
			room.BroadcastMessage(message)
		case "getUserList":
			room.BroadcastUserList()
		default:
			log.Printf("unknown message type: %s", msg.Type)
		}
	}
}
