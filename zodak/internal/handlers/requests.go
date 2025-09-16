package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateRequest(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		var in struct{ Title, Body string }
		if err := c.BindJSON(&in); err != nil || len(in.Title) < 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error":"invalid"}); return
		}
		_, err := db.Exec(c, `INSERT INTO requests(author_id,title,body,status,created_at) VALUES($1,$2,$3,'new',NOW())`,
			uid, in.Title, in.Body)
		if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error":"db"}); return }
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func ListRequests(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		rows, err := db.Query(c, `SELECT id,title,status,created_at FROM requests WHERE author_id=$1 ORDER BY id DESC LIMIT 50`, uid)
		if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error":"db"}); return }
		defer rows.Close()
		type item struct{ ID int64; Title, Status string; CreatedAt time.Time }
		var list []item
		for rows.Next() {
			var it item
			_ = rows.Scan(&it.ID,&it.Title,&it.Status,&it.CreatedAt)
			list = append(list,it)
		}
		c.JSON(http.StatusOK, gin.H{"items": list})
	}
}
