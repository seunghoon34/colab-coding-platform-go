package types

import "sync"

type User struct {
	Username string `json:"username"`
	IsHost   bool   `json:"isHost"`
	Color    string `json:"color"`
}

type Room struct {
	Code    string
	Clients map[*Client]*User
	Mutex   sync.Mutex
}

type Client struct {
	Send     chan []byte
	Username string
	RoomCode string
}

type Message struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}
