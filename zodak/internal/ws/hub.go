package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu    sync.RWMutex
	conns map[int64]*websocket.Conn
}

func NewHub() *Hub { return &Hub{conns: map[int64]*websocket.Conn{}} }

func (h *Hub) Add(userID int64, c *websocket.Conn) {
	h.mu.Lock(); defer h.mu.Unlock()
	if old, ok := h.conns[userID]; ok { old.Close() }
	h.conns[userID] = c
}

func (h *Hub) Remove(userID int64) {
	h.mu.Lock(); defer h.mu.Unlock()
	delete(h.conns, userID)
}

func (h *Hub) Send(userID int64, msg any) {
	h.mu.RLock(); conn := h.conns[userID]; h.mu.RUnlock()
	if conn != nil { _ = conn.WriteJSON(msg) }
}

