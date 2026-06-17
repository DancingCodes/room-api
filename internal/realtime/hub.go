package realtime

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type Event struct {
	Type   string `json:"type"`
	RoomID uint64 `json:"room_id"`
	Data   any    `json:"data"`
}

type Client struct {
	RoomID uint64
	UserID uint64
	Conn   *websocket.Conn
}

type Hub struct {
	mu    sync.RWMutex
	rooms map[uint64]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[uint64]map[*Client]struct{})}
}

func (h *Hub) Add(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[client.RoomID] == nil {
		h.rooms[client.RoomID] = make(map[*Client]struct{})
	}
	h.rooms[client.RoomID][client] = struct{}{}
}

func (h *Hub) Remove(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients := h.rooms[client.RoomID]
	if clients == nil {
		return
	}
	delete(clients, client)
	if len(clients) == 0 {
		delete(h.rooms, client.RoomID)
	}
}

func (h *Hub) Broadcast(roomID uint64, event Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	clients := make([]*Client, 0, len(h.rooms[roomID]))
	for client := range h.rooms[roomID] {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		if err := client.Conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			_ = client.Conn.Close()
			h.Remove(client)
		}
	}
}
