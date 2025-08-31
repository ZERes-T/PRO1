package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ListNotifications(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		rows, err := db.Query(c, `SELECT id, type, payload, read_at, created_at FROM notifications
			WHERE user_id=$1 ORDER BY created_at DESC LIMIT 100`, uid)
		if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error":"db"}); return }
		defer rows.Close()
		var items []map[string]any
		for rows.Next() {
			var id int64; var typ string; var payload any; var readAt any; var createdAt any
			_ = rows.Scan(&id,&typ,&payload,&readAt,&createdAt)
			items = append(items, gin.H{"id": id, "type": typ, "payload": payload, "read_at": readAt, "created_at": createdAt})
		}
		var unread int64
		_ = db.QueryRow(c, `SELECT count(*) FROM notifications WHERE user_id=$1 AND read_at IS NULL`, uid).Scan(&unread)
		c.JSON(http.StatusOK, gin.H{"items": items, "unread": unread})
	}
}

func MarkNotificationRead(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		var in struct{ ID int64 `json:"id"` }
		if err := c.BindJSON(&in); err != nil || in.ID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error":"bad"}); return
		}
		_, err := db.Exec(c, `UPDATE notifications SET read_at=now() WHERE id=$1 AND user_id=$2 AND read_at IS NULL`, in.ID, uid)
		if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error":"db"}); return }
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func MarkAllNotificationsRead(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		if _, err := db.Exec(c, `UPDATE notifications SET read_at=now() WHERE user_id=$1 AND read_at IS NULL`, uid); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error":"db"}); return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

