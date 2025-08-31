package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FeedItem struct {
	Type string      `json:"type"` // "video" | "ad"
	Data interface{} `json:"data"`
}

func FeedChunked(fetch func(limit int) []FeedItem) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		w := c.Writer
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("["))

		items := fetch(20)
		enc := json.NewEncoder(w)
		for i, it := range items {
			if i > 0 { _, _ = w.Write([]byte(",")) }
			_ = enc.Encode(it)
			if f, ok := w.(http.Flusher); ok { f.Flush() }
		}
		_, _ = w.Write([]byte("]"))
	}
}

