package models

import (
	"sync"

	"github.com/seunghoon34/collaborative-coding-platform/pkg/types"
)

var (
	rooms = make(map[string]*types.Room)
	mutex sync.Mutex
)

func GetOrCreateRoom(code string) *types.Room {
	mutex.Lock()
	defer mutex.Unlock()

	if room, exists := rooms[code]; exists {
		return room
	}

	room := &types.Room{
		Code:    code,
		Clients: make(map[*types.Client]bool),
	}
	rooms[code] = room
	return room
}

func RegisterClient(room *types.Room, client *types.Client) {
	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	room.Clients[client] = true
}

func UnregisterClient(room *types.Room, client *types.Client) {
	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	if _, ok := room.Clients[client]; ok {
		delete(room.Clients, client)
		close(client.Send)
	}
}

func BroadcastMessage(room *types.Room, message []byte) {
	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	for client := range room.Clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(room.Clients, client)
		}
	}
}
