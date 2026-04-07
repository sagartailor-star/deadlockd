package api

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients   map[*websocket.Conn]bool
	Broadcast chan []byte
	mu        sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:   make(map[*websocket.Conn]bool),
		Broadcast: make(chan []byte, 256),
	}
}

func (h *Hub) Register(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()
}

func (h *Hub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	if _, ok := h.clients[conn]; ok {
		delete(h.clients, conn)
		conn.Close()
	}
	h.mu.Unlock()
}

func (h *Hub) Run() {
	for msg := range h.Broadcast {
		h.mu.RLock()
		targets := make([]*websocket.Conn, 0, len(h.clients))
		for conn := range h.clients {
			targets = append(targets, conn)
		}
		h.mu.RUnlock()

		for _, conn := range targets {
			err := conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				h.Unregister(conn)
			}
		}
	}
}
