package types

import "sync"

type Room struct {
	Code    string
	Clients map[*Client]bool
	Mutex   sync.Mutex
}

type Client struct {
	Send     chan []byte
	Username string
	RoomCode string
}
