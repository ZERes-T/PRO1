package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RequireRole(db *pgxpool.Pool, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		var exists bool
		err := db.QueryRow(c, `SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id=$1 AND role=$2)`, uid, role).Scan(&exists)
		if err != nil || !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error":"forbidden"})
			return
		}
		c.Next()
	}
}

