package router

import (
	"time"
	"zodak/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// New создаёт роутер с CORS и логами
func New(corsOrigin string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{corsOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// логируем задержку
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		_ = time.Since(start) // можно писать в логи
	})

	return r
}

// AuthGroup создаёт защищённую группу API
func AuthGroup(r *gin.Engine, jwtSecret string) *gin.RouterGroup {
	g := r.Group("/api")
	g.Use(middleware.JWT(jwtSecret))
	return g
}

