package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"zodak/internal/cache"
	"zodak/internal/config"
	"zodak/internal/db"
	"zodak/internal/handlers"
	"zodak/internal/router"
	"zodak/internal/ws"
	"zodak/internal/middleware" // 👈 вот этого не хватало

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.FromEnv()
	pg := db.MustPool(cfg.PGDsn)
	if err := pg.Ping(context.Background()); err != nil {
		log.Fatal("PG ping failed:", err)
	}
	defer pg.Close()

	_ = cache.New(cfg.RedisAddr)
	hub := ws.NewHub()
	r := router.New(cfg.CORSOrigin)

	// публичные
	r.GET("/health", handlers.Health)
	r.POST("/api/auth/register", handlers.Register(pg, cfg.JWTSecret))
	r.POST("/api/auth/login", handlers.Login(pg, cfg.JWTSecret))
	r.GET("/legal/terms", handlers.Terms)

	// лента (демо-чанки с рекламой 1/5)
	r.GET("/api/media", handlers.ListMedia)
	r.GET("/api/feed", handlers.FeedChunked(func(limit int) []handlers.FeedItem {
		items := make([]handlers.FeedItem, 0, limit)
		for i := 0; i < limit; i++ {
			if (i+1)%5 == 0 {
				items = append(items, handlers.FeedItem{Type: "ad", Data: gin.H{"slot": "feed-5"}})
			} else {
				items = append(items, handlers.FeedItem{Type: "video", Data: gin.H{"id": i + 1, "title": "Demo"}})
			}
		}
		return items
	}))

	// приватные (JWT)
	auth := router.AuthGroup(r, cfg.JWTSecret)
	auth.POST("/favorites/toggle", handlers.ToggleFavorite(pg))
	auth.POST("/requests", handlers.CreateRequest(pg))
	auth.GET("/requests", handlers.ListRequests(pg))
	auth.GET("/ws", handlers.WS(hub))

	// мессенджер
	auth.GET("/dialogs", handlers.ListDialogs(pg))
	auth.GET("/messages", handlers.ListMessages(pg))
	auth.POST("/messages/send", handlers.SendMessage(pg, hub))
	auth.POST("/messages/read", handlers.MarkRead(pg))

	// ЛК
	auth.GET("/me", handlers.Me(pg))
	auth.PUT("/me", handlers.UpdateMe(pg))

	// Конфигуратор
	auth.POST("/configurator/quote", handlers.CreateQuote(pg))
	auth.GET("/configurator/quotes", handlers.ListQuotes(pg))
	auth.GET("/configurator/quote", handlers.GetQuote(pg))

	// Саппорт
	auth.POST("/support/tickets", handlers.CreateTicket(pg))
	auth.GET("/support/tickets", handlers.ListTickets(pg))
	auth.POST("/support/messages", handlers.AddSupportMessage(pg))
	auth.GET("/support/messages", handlers.ListSupportMessages(pg))

	// Нотификации
	auth.GET("/notifications", handlers.ListNotifications(pg))
	auth.POST("/notifications/read", handlers.MarkNotificationRead(pg))
	auth.POST("/notifications/read_all", handlers.MarkAllNotificationsRead(pg))

	// 🔐 АДМИНКА — тут. Защищена ролями через middleware.RequireRole(pg, "admin")
	admin := auth.Group("/admin", middleware.RequireRole(pg, "admin"))
	admin.POST("/roles/grant", handlers.GrantRole(pg))
	admin.POST("/roles/revoke", handlers.RevokeRole(pg))
	admin.GET("/roles", handlers.ListRoles(pg))

	srv := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: r,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 0,
	}
	log.Println("API on :" + cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
