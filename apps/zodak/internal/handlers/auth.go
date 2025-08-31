package handlers

import (
	"net/http"
	"strings"
	"time"
	"log"
	"zodak/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func Register(pgxPool *pgxpool.Pool, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in struct{ Email, Name, Password string }
		if err := c.BindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
			return
		}
		in.Email = strings.TrimSpace(strings.ToLower(in.Email))
		if !strings.Contains(in.Email, "@") || len(in.Password) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid creds"})
			return
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		_, err := pgxPool.Exec(c,
			`INSERT INTO users(email, name, pass_hash) VALUES($1,$2,$3)
             ON CONFLICT (email) DO NOTHING`,
			in.Email, strings.TrimSpace(in.Name), string(hash),
		)
		if err != nil {
			log.Println("[Register] DB error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func Login(pgxPool *pgxpool.Pool, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in struct{ Email, Password string }
		if err := c.BindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
			return
		}
		var id int64
		var hash string
		err := pgxPool.QueryRow(c,
			`SELECT id, pass_hash FROM users WHERE email=$1`,
			strings.ToLower(in.Email),
		).Scan(&id, &hash)
		if err != nil {
			log.Println("[Login] DB error:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong creds"})
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)) != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong creds"})
			return
		}
		tok, _ := auth.Sign(jwtSecret, id, 24*time.Hour)
		c.JSON(http.StatusOK, gin.H{"token": tok})
	}
}

