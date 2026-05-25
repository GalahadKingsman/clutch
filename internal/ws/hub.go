package ws

import (
	"encoding/json"
	"sync"

	"github.com/GalahadKingsman/clutch/internal/models"
	"github.com/gorilla/websocket"
)

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[string]map[*websocket.Conn]struct{})}
}

func (h *Hub) Join(duelID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[duelID] == nil {
		h.rooms[duelID] = make(map[*websocket.Conn]struct{})
	}
	h.rooms[duelID][conn] = struct{}{}
}

func (h *Hub) Leave(duelID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room := h.rooms[duelID]; room != nil {
		delete(room, conn)
		if len(room) == 0 {
			delete(h.rooms, duelID)
		}
	}
}

func (h *Hub) Broadcast(duelID string, msg models.ChatMessage) {
	payload, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.mu.RLock()
	room := h.rooms[duelID]
	conns := make([]*websocket.Conn, 0, len(room))
	for c := range room {
		conns = append(conns, c)
	}
	h.mu.RUnlock()

	for _, c := range conns {
		_ = c.WriteMessage(websocket.TextMessage, payload)
	}
}
