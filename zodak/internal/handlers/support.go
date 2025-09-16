package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// POST /api/support/tickets {subject}
func CreateTicket(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		var in struct{ Subject string `json:"subject"` }
		if err := c.BindJSON(&in); err != nil || len(in.Subject) < 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error":"bad"}); return
		}
		var id int64
		if err := db.QueryRow(c, `INSERT INTO support_tickets(user_id,subject) VALUES($1,$2) RETURNING id`, uid, in.Subject).Scan(&id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error":"db"}); return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "ticket_id": id})
	}
}

// GET /api/support/tickets
func ListTickets(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		rows, err := db.Query(c, `SELECT id, subject, status, created_at FROM support_tickets WHERE user_id=$1 ORDER BY created_at DESC LIMIT 100`, uid)
		if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error":"db"}); return }
		defer rows.Close()
		type T struct { ID int64; Subject, Status string; CreatedAt any }
		var list []gin.H
		for rows.Next() {
			var t T; _ = rows.Scan(&t.ID, &t.Subject, &t.Status, &t.CreatedAt)
			list = append(list, gin.H{"id": t.ID, "subject": t.Subject, "status": t.Status, "created_at": t.CreatedAt})
		}
		c.JSON(http.StatusOK, gin.H{"items": list})
	}
}

// POST /api/support/messages {ticket_id, text}
func AddSupportMessage(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		var in struct{ TicketID int64 `json:"ticket_id"`; Text string `json:"text"` }
		if err := c.BindJSON(&in); err != nil || in.TicketID <= 0 || len(in.Text) < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error":"bad"}); return
		}
		var owner int64
		if err := db.QueryRow(c, `SELECT user_id FROM support_tickets WHERE id=$1`, in.TicketID).Scan(&owner); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error":"not_found"}); return
		}
		if owner != uid {
			c.JSON(http.StatusForbidden, gin.H{"error":"forbidden"}); return
		}
		if _, err := db.Exec(c, `INSERT INTO support_messages(ticket_id,user_id,text) VALUES($1,$2,$3)`, in.TicketID, uid, in.Text); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error":"db"}); return
		}
		_, _ = db.Exec(c, `UPDATE support_tickets SET status='pending' WHERE id=$1 AND status='open'`, in.TicketID)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// GET /api/support/messages?ticket_id=...
func ListSupportMessages(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		tid := c.Query("ticket_id")
		var owner int64
		if err := db.QueryRow(c, `SELECT user_id FROM support_tickets WHERE id=$1`, tid).Scan(&owner); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error":"not_found"}); return
		}
		if owner != uid { c.JSON(http.StatusForbidden, gin.H{"error":"forbidden"}); return }

		rows, err := db.Query(c, `SELECT id, user_id, text, created_at FROM support_messages WHERE ticket_id=$1 ORDER BY id ASC LIMIT 500`, tid)
		if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error":"db"}); return }
		defer rows.Close()
		type M struct { ID int64; UserID int64; Text string; CreatedAt any }
		var list []gin.H
		for rows.Next() {
			var m M; _ = rows.Scan(&m.ID,&m.UserID,&m.Text,&m.CreatedAt)
			list = append(list, gin.H{"id": m.ID, "user_id": m.UserID, "text": m.Text, "created_at": m.CreatedAt})
		}
		c.JSON(http.StatusOK, gin.H{"items": list})
	}
}

