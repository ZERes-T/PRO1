package middleware

import (
	"net/http"
	"strings"

	"zodak/internal/auth"

	"github.com/gin-gonic/gin"
)

func JWT(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(strings.ToLower(h), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no token"})
			return
		}
		token := strings.TrimSpace(h[7:])
		claims, err := auth.Parse(secret, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("uid", claims.UserID)
		c.Next()
	}
}

