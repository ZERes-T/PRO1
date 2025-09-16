package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ToggleFavorite(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		var in struct {
			VideoID int64 `json:"video_id"`
		}
		if err := c.BindJSON(&in); err != nil || in.VideoID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad"})
			return
		}

		// 1) если запись есть — удалим (получим RowsAffected > 0)
		tag, err := db.Exec(c, `DELETE FROM favorites WHERE user_id=$1 AND video_id=$2`, uid, in.VideoID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
			return
		}
		if tag.RowsAffected() > 0 {
			c.JSON(http.StatusOK, gin.H{"ok": true, "favorited": false})
			return
		}

		// 2) иначе — добавим
		_, err = db.Exec(c, `INSERT INTO favorites(user_id, video_id) VALUES($1,$2)`, uid, in.VideoID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "favorited": true})
	}
}

