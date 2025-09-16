package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GrantRole(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in struct{ UserID int64 `json:"user_id"`; Role string `json:"role"` }
		if err := c.BindJSON(&in); err != nil || in.UserID <= 0 || in.Role == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error":"bad"}); return
		}
		in.Role = strings.ToLower(in.Role)
		_, err := db.Exec(c, `INSERT INTO user_roles(user_id,role) VALUES($1,$2) ON CONFLICT DO NOTHING`, in.UserID, in.Role)
		if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error":"db"}); return }
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func RevokeRole(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in struct{ UserID int64 `json:"user_id"`; Role string `json:"role"` }
		if err := c.BindJSON(&in); err != nil || in.UserID <= 0 || in.Role == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error":"bad"}); return
		}
		_, err := db.Exec(c, `DELETE FROM user_roles WHERE user_id=$1 AND role=$2`, in.UserID, strings.ToLower(in.Role))
		if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error":"db"}); return }
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func ListRoles(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidStr := c.Query("user_id")
		if uidStr == "" { c.JSON(http.StatusBadRequest, gin.H{"error":"bad"}); return }
		rows, err := db.Query(c, `SELECT role, granted_at FROM user_roles WHERE user_id=$1 ORDER BY role`, uidStr)
		if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error":"db"}); return }
		defer rows.Close()
		var list []map[string]any
		for rows.Next() {
			var role string; var when any
			_ = rows.Scan(&role, &when)
			list = append(list, gin.H{"role": role, "granted_at": when})
		}
		c.JSON(http.StatusOK, gin.H{"items": list})
	}
}

