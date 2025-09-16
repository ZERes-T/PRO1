package handlers

import "github.com/gin-gonic/gin"

// GET /legal/terms
func Terms(c *gin.Context) {
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(200, `Пользовательское соглашение (MVP)
1. Вы — автор загружаемого контента и соблюдаете закон.
2. Мы можем модерировать материалы и блокировать нарушения.
3. Персональные данные — по политике конфиденциальности.
4. Используя сервис, вы соглашаетесь с данными условиями.
Версия: MVP`)
}

