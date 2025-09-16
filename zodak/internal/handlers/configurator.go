package handlers

import (
	"encoding/json"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type QuoteInput struct {
	Kind     string `json:"kind"`       // "wardrobe"
	Wmm      int    `json:"w_mm"`       // ширина, мм
	Hmm      int    `json:"h_mm"`       // высота, мм
	Dmm      int    `json:"d_mm"`       // глубина, мм
	Material string `json:"material"`   // ldsp | mdf | plywood
	Facade   string `json:"facade"`     // none | mirror | glass | mdf_paint
	Drawers  int    `json:"drawers"`
	Shelves  int    `json:"shelves"`
	Doors    int    `json:"doors"`      // 1..4
	Corner   bool   `json:"corner"`
	Lighting bool   `json:"lighting"`
	Save     bool   `json:"save"`       // сохранить в БД
}

func clamp(x, lo, hi int) int {
	if x < lo { return lo }
	if x > hi { return hi }
	return x
}

func calcPrice(in QuoteInput) (float64, map[string]float64) {
	// БАЗОВЫЕ СТАВКИ (условные, RUB)
	materialPerM2 := map[string]float64{
		"ldsp":    1200,  // ЛДСП
		"mdf":     1800,
		"plywood": 2000,
	}
	facadePerM2 := map[string]float64{
		"none":      0,
		"mirror":    3500,
		"glass":     3000,
		"mdf_paint": 4000,
	}
	edgePerM := 60
	hwPerDrawer := 400
	hwPerDoor := 1200
	hwPerShelf := 150
	optLighting := 1500

	// геометрия (мм → м)
	w := float64(in.Wmm) / 1000.0
	h := float64(in.Hmm) / 1000.0
	d := float64(in.Dmm) / 1000.0

	// приближённая площадь панелей (м^2) — короб + задник, коэффициент 0.6 для "не всё из панелей"
	panelArea := 2*(w*h + w*d + h*d) * 0.6

	// фасад — площадь фронта
	facadeArea := w * h

	// длина кромки по периметру (условно): 2*(w+h) * 2 (верх/низ), в метрах
	edgeLen := 2*(w+h) * 2

	mat := materialPerM2[in.Material]
	if mat == 0 { mat = materialPerM2["ldsp"] }
	fac := facadePerM2[in.Facade]

	carcass := panelArea * mat
	if in.Corner {
		carcass *= 1.10 // угловой — немного дороже
	}

	facade := facadeArea * fac
	edge := edgeLen * float64(edgePerM)

	hardware := float64(in.Drawers*hwPerDrawer + in.Doors*hwPerDoor + in.Shelves*hwPerShelf)
	options := 0.0
	if in.Lighting { options += float64(optLighting) }

	subtotal := carcass + facade + edge + hardware + options
	// минималка, чтобы не улетать вниз
	if subtotal < 12000 {
		subtotal = 12000
	}

	out := map[string]float64{
		"carcass": carcass,
		"facade":  facade,
		"edge":    edge,
		"hardware": hardware,
		"options": options,
		"subtotal": subtotal,
	}
	total := math.Round(subtotal*100) / 100
	return total, out
}

// POST /api/configurator/quote  (JWT)
// body: QuoteInput; если Save=true — сохраняем в БД, возвращаем id.
func CreateQuote(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		var in QuoteInput
		if err := c.BindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":"bad"}); return
		}
		// валидация/клампы
		in.Kind = "wardrobe"
		in.Wmm = clamp(in.Wmm, 500, 5000)
		in.Hmm = clamp(in.Hmm, 500, 3000)
		in.Dmm = clamp(in.Dmm, 300, 1200)
		if in.Doors <= 0 || in.Doors > 6 { in.Doors = 2 }
		if in.Drawers < 0 { in.Drawers = 0 }
		if in.Shelves < 0 { in.Shelves = 0 }
		if in.Material == "" { in.Material = "ldsp" }
		if in.Facade == "" { in.Facade = "none" }

		total, breakdown := calcPrice(in)
		resp := gin.H{
			"ok": true,
			"total": total,
			"currency": "RUB",
			"breakdown": breakdown,
		}
		if !in.Save {
			c.JSON(http.StatusOK, resp)
			return
		}

		blob, _ := json.Marshal(breakdown)
		var id int64
		err := db.QueryRow(c, `
			INSERT INTO quotes(user_id, kind, w_mm, h_mm, d_mm, material, facade, drawers, shelves, doors, corner, lighting, total, breakdown, status)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14::jsonb,'draft')
			RETURNING id
		`, uid, in.Kind, in.Wmm, in.Hmm, in.Dmm, in.Material, in.Facade, in.Drawers, in.Shelves, in.Doors, in.Corner, in.Lighting, total, string(blob)).Scan(&id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error":"db"})
			return
		}
		resp["id"] = id
		c.JSON(http.StatusOK, resp)
	}
}

// GET /api/configurator/quotes  (JWT) -> твои последние 50
func ListQuotes(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		rows, err := db.Query(c, `
			SELECT id, kind, w_mm, h_mm, d_mm, material, facade, drawers, shelves, doors, corner, lighting, total, status, created_at
			FROM quotes WHERE user_id=$1 ORDER BY created_at DESC LIMIT 50
		`, uid)
		if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error":"db"}); return }
		defer rows.Close()
		type item struct {
			ID int64 `json:"id"`
			Kind string `json:"kind"`
			Wmm, Hmm, Dmm int `json:"w_mm","h_mm","d_mm"`
			Material, Facade string `json:"material","facade"`
			Drawers, Shelves, Doors int `json:"drawers","shelves","doors"`
			Corner, Lighting bool `json:"corner","lighting"`
			Total float64 `json:"total"`
			Status string `json:"status"`
			CreatedAt string `json:"created_at"`
		}
		var list []map[string]any
		for rows.Next() {
			var it item
			var createdAt any
			_ = rows.Scan(&it.ID,&it.Kind,&it.Wmm,&it.Hmm,&it.Dmm,&it.Material,&it.Facade,&it.Drawers,&it.Shelves,&it.Doors,&it.Corner,&it.Lighting,&it.Total,&it.Status,&createdAt)
			list = append(list, gin.H{
				"id": it.ID, "kind": it.Kind, "w_mm": it.Wmm, "h_mm": it.Hmm, "d_mm": it.Dmm,
				"material": it.Material, "facade": it.Facade,
				"drawers": it.Drawers, "shelves": it.Shelves, "doors": it.Doors,
				"corner": it.Corner, "lighting": it.Lighting,
				"total": it.Total, "status": it.Status, "created_at": createdAt,
			})
		}
		c.JSON(http.StatusOK, gin.H{"items": list})
	}
}

// GET /api/configurator/quote?id=...  (JWT) -> с расшифровкой breakdown
func GetQuote(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		id := c.Query("id")
		if id == "" { c.JSON(http.StatusBadRequest, gin.H{"error":"bad"}); return }
		var (
			userID int64
			kind string
			wmm, hmm, dmm int
			material, facade string
			drawers, shelves, doors int
			corner, lighting bool
			total float64
			status string
			breakdownStr string
		)
		err := db.QueryRow(c, `
			SELECT user_id, kind, w_mm, h_mm, d_mm, material, facade, drawers, shelves, doors, corner, lighting, total, status, breakdown::text
			FROM quotes WHERE id=$1
		`, id).Scan(&userID,&kind,&wmm,&hmm,&dmm,&material,&facade,&drawers,&shelves,&doors,&corner,&lighting,&total,&status,&breakdownStr)
		if err != nil { c.JSON(http.StatusNotFound, gin.H{"error":"not_found"}); return }
		if userID != uid { c.JSON(http.StatusForbidden, gin.H{"error":"forbidden"}); return }
		var breakdown map[string]float64
		_ = json.Unmarshal([]byte(breakdownStr), &breakdown)
		c.JSON(http.StatusOK, gin.H{
			"id": id, "kind": kind, "w_mm": wmm, "h_mm": hmm, "d_mm": dmm,
			"material": material, "facade": facade,
			"drawers": drawers, "shelves": shelves, "doors": doors,
			"corner": corner, "lighting": lighting,
			"total": total, "currency":"RUB",
			"breakdown": breakdown,
		})
	}
}

