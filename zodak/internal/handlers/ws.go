package handlers

import (
	"zodak/internal/ws"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgr = websocket.Upgrader{ CheckOrigin: func(r *http.Request) bool { return true } }

func WS(h *ws.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		conn, err := upgr.Upgrade(c.Writer, c.Request, nil)
		if err != nil { return }
		h.Add(uid, conn)
		defer func(){ h.Remove(uid); conn.Close() }()
		for {
			var msg map[string]any
			if err := conn.ReadJSON(&msg); err != nil { break }
			_ = conn.WriteJSON(gin.H{"type":"pong"})
		}
	}
}

