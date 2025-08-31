package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"zodak/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Dialog struct {
	ID        int64     `json:"id"`
	AID       int64     `json:"a_id"`
	BID       int64     `json:"b_id"`
	UpdatedAt time.Time `json:"updated_at"`
}
type Msg struct {
	ID        int64      `json:"id"`
	DialogID  int64      `json:"dialog_id"`
	SenderID  int64      `json:"sender_id"`
	Text      string     `json:"text"`
	CreatedAt time.Time  `json:"created_at"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
}

// ensureDialog: ищем диалог (a<=b). Если нет — создаём. ВАЖНО: ctx не nil!
func ensureDialog(ctx context.Context, db *pgxpool.Pool, uid, peer int64) (int64, error) {
	a, b := uid, peer
	if a > b {
		a, b = b, a
	}
	var id int64
	err := db.QueryRow(ctx, `SELECT id FROM dialogs WHERE a_id=$1 AND b_id=$2`, a, b).Scan(&id)
	if err == nil {
		return id, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		err = db.QueryRow(ctx, `INSERT INTO dialogs(a_id,b_id) VALUES($1,$2) RETURNING id`, a, b).Scan(&id)
		return id, err
	}
	return 0, err
}

// POST /api/messages/send {peer_id, text}
func SendMessage(db *pgxpool.Pool, hub *ws.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		var in struct {
			PeerID int64  `json:"peer_id"`
			Text   string `json:"text"`
		}
		if err := c.BindJSON(&in); err != nil || in.PeerID <= 0 || len(in.Text) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad"})
			return
		}
		if in.PeerID == uid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "self"})
			return
		}
		// передаём КОНТЕКСТ из gin
		did, err := ensureDialog(c, db, uid, in.PeerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
			return
		}

		var mid int64
		var ts time.Time
		err = db.QueryRow(c,
			`INSERT INTO messages(dialog_id, sender_id, text) VALUES($1,$2,$3) RETURNING id, created_at`,
			did, uid, in.Text,
		).Scan(&mid, &ts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
			return
		}

		_, _ = db.Exec(c, `UPDATE dialogs SET updated_at=now() WHERE id=$1`, did)

		payload := gin.H{
			"id": mid, "dialog_id": did, "sender_id": uid, "text": in.Text, "created_at": ts,
		}
		// пушим получателю
		hub.Send(in.PeerID, gin.H{"type": "message.new", "payload": payload})
		c.JSON(http.StatusOK, gin.H{"ok": true, "message": payload})
		// оффлайн-нотификация получателю
		_, _ = db.Exec(c, `
		  INSERT INTO notifications(user_id, type, payload)
		  VALUES($1,'message', jsonb_build_object('dialog_id',$2,'message_id',$3,'sender_id',$4,'text',substr($5,1,140)))
		`, in.PeerID, did, mid, uid, in.Text)
	
		// пуш нотифки по WS (если онлайн)
		hub.Send(in.PeerID, gin.H{"type":"notification.new", "payload": gin.H{
		  "type":"message", "dialog_id": did, "message_id": mid, "sender_id": uid,
		}})

	}
}

// GET /api/dialogs
func ListDialogs(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		rows, err := db.Query(c, `
			SELECT id,a_id,b_id,updated_at FROM dialogs
			WHERE a_id=$1 OR b_id=$1
			ORDER BY updated_at DESC
			LIMIT 50
		`, uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
			return
		}
		defer rows.Close()
		var list []Dialog
		for rows.Next() {
			var d Dialog
			_ = rows.Scan(&d.ID, &d.AID, &d.BID, &d.UpdatedAt)
			list = append(list, d)
		}
		c.JSON(http.StatusOK, gin.H{"items": list})
	}
}

// GET /api/messages?dialog_id=...
func ListMessages(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		did := c.Query("dialog_id")
		if did == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad"})
			return
		}
		var a, b int64
		if err := db.QueryRow(c, `SELECT a_id,b_id FROM dialogs WHERE id=$1`, did).Scan(&a, &b); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		if uid != a && uid != b {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		rows, err := db.Query(c, `
			SELECT id,dialog_id,sender_id,text,created_at,read_at
			FROM messages WHERE dialog_id=$1 ORDER BY id ASC LIMIT 500
		`, did)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
			return
		}
		defer rows.Close()
		var list []Msg
		for rows.Next() {
			var m Msg
			_ = rows.Scan(&m.ID, &m.DialogID, &m.SenderID, &m.Text, &m.CreatedAt, &m.ReadAt)
			list = append(list, m)
		}
		c.JSON(http.StatusOK, gin.H{"items": list})
	}
}

// POST /api/messages/read {dialog_id}
func MarkRead(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		var in struct {
			DialogID int64 `json:"dialog_id"`
		}
		if err := c.BindJSON(&in); err != nil || in.DialogID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad"})
			return
		}
		var a, b int64
		if err := db.QueryRow(c, `SELECT a_id,b_id FROM dialogs WHERE id=$1`, in.DialogID).Scan(&a, &b); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		if uid != a && uid != b {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		_, err := db.Exec(c, `
			UPDATE messages SET read_at=now()
			WHERE dialog_id=$1 AND sender_id <> $2 AND read_at IS NULL
		`, in.DialogID, uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

