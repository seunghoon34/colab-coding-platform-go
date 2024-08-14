package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/seunghoon34/collaborative-coding-platform/pkg/types"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Adjust this for production!
	},
}

type Client struct {
	types.Client
	conn *websocket.Conn
}

func NewClient(conn *websocket.Conn, username, roomCode string) *Client {
	return &Client{
		Client: types.Client{
			Send:     make(chan []byte, 256),
			Username: username,
			RoomCode: roomCode,
		},
		conn: conn,
	}
}

func (c *Client) ReadPump(broadcastMessage func([]byte)) {
	defer func() {
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		broadcastMessage(message)
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
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

func (c *Client) Close() {
	close(c.Send)
}
