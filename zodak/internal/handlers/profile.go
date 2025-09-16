package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Me(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		var email, name string
		if err := db.QueryRow(c, `SELECT email,name FROM users WHERE id=$1`, uid).Scan(&email,&name); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error":"not_found"}); return
		}
		c.JSON(http.StatusOK, gin.H{"id": uid, "email": email, "name": name})
	}
}

func UpdateMe(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		var in struct{ Name string `json:"name"` }
		if err := c.BindJSON(&in); err != nil || len(in.Name) < 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error":"bad"}); return
		}
		if _, err := db.Exec(c, `UPDATE users SET name=$1 WHERE id=$2`, in.Name, uid); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error":"db"}); return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

