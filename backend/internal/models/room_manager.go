package models

import (
	"sync"
)

var (
	roomManager *RoomManager
	once        sync.Once
)

type RoomManager struct {
	rooms map[string]*Room
	mutex sync.RWMutex
}

func GetRoomManager() *RoomManager {
	once.Do(func() {
		roomManager = &RoomManager{
			rooms: make(map[string]*Room),
		}
	})
	return roomManager
}

func (rm *RoomManager) GetOrCreateRoom(code string) *Room {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if room, exists := rm.rooms[code]; exists {
		return room
	}

	room := &Room{
		Code:    code,
		Clients: make(map[*Client]bool),
	}
	rm.rooms[code] = room
	return room
}

func (rm *RoomManager) BroadcastToRoom(roomCode string, message []byte) {
	rm.mutex.RLock()
	room, exists := rm.rooms[roomCode]
	rm.mutex.RUnlock()

	if exists {
		room.BroadcastMessage(message)
	}
}
