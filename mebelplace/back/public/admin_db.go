// admin_db.go — временная страница администрирования PostgreSQL на Go.
// Маршруты:
//   GET  /admin/db                 — список таблиц (public)
//   GET  /admin/db/t/{table}       — первые N строк
//   POST /admin/db/sql             — SQL-консоль (по умолчанию только SELECT)
//
// Защита: Basic Auth (логин/пароль из ENV).
// Подключение к БД: ENV (PGHOST, PGPORT, PGUSER, PGPASSWORD, PGDATABASE).
//
// Быстрый старт ENV (пример):
//   export PGHOST=127.0.0.1 PGPORT=5432 PGUSER=mp_user PGPASSWORD='***' PGDATABASE=mebelplace
//   export ADMIN_DB_USER=admin ADMIN_DB_PASS='сильный_пароль'
//   export ADMIN_DB_ALLOW_WRITE=0   # 1 чтобы разрешить INSERT/UPDATE/DELETE

package main

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib" // драйвер pgx для database/sql
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	db              *sql.DB
	basicUser       = getenv("ADMIN_DB_USER", "admin")
	basicPass       = getenv("ADMIN_DB_PASS", "change_me")
	allowWrite      = getenv("ADMIN_DB_ALLOW_WRITE", "0") == "1"
	rowLimitDefault = 100
)

// очень простой шаблон (тёмный)
var pageTpl = template.Must(template.New("page").Parse(`
<!doctype html>
<html lang="ru">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>{{.Title}}</title>
<style>
  :root{--bg:#0b0b0f;--card:#11131a;--line:#1f2430;--ink:#e5e7eb;--accent:#f59e0b}
  *{box-sizing:border-box} body{margin:0;background:var(--bg);color:var(--ink);font:15px/1.5 system-ui,Inter,Segoe UI}
  a{color:var(--accent);text-decoration:none}
  header{position:sticky;top:0;background:#0d0f15;border-bottom:1px solid var(--line);padding:10px 16px}
  main{max-width:1100px;margin:0 auto;padding:16px}
  .card{background:var(--card);border:1px solid var(--line);border-radius:12px;padding:12px;margin:12px 0}
  table{width:100%;border-collapse:collapse;background:var(--card)}
  th,td{border:1px solid var(--line);padding:6px 10px;vertical-align:top}
  th{background:#161926;text-align:left}
  .muted{opacity:.7}
  .row{display:flex;gap:12px;flex-wrap:wrap}
  input,select,textarea,button{background:#161926;color:var(--ink);border:1px solid var(--line);border-radius:8px;padding:8px 10px}
  button{cursor:pointer}
  code,pre{background:#0f121a;border:1px solid var(--line);border-radius:8px;padding:8px;display:block;overflow:auto}
  .ok{color:#22c55e}.bad{color:#ef4444}
</style>
</head>
<body>
<header>
  <div class="row">
    <div><b>DB Admin</b> — {{.DBName}}@{{.DBHost}}</div>
    <div class="muted">Write: {{if .AllowWrite}}<span class="ok">ENABLED</span>{{else}}<span class="bad">DISABLED</span>{{end}}</div>
    <div><a href="/admin/db">Таблицы</a></div>
  </div>
</header>
<main>
  {{template "content" .}}
</main>
</body>
</html>
`))

var listTpl = template.Must(template.Must(pageTpl.Clone()).New("content").Parse(`
<h2>Таблицы (schema public)</h2>
<div class="card">
  {{if .Tables}}
    <div class="row">
      {{range .Tables}}
        <a href="/admin/db/t/{{.}}" style="padding:8px 12px;border:1px solid var(--line);border-radius:999px;background:#0f121a">{{.}}</a>
      {{end}}
    </div>
  {{else}}
    <div class="muted">Нет таблиц в schema public.</div>
  {{end}}
</div>

<h3>SQL консоль</h3>
<div class="card">
  <form method="post" action="/admin/db/sql">
    <div style="margin-bottom:8px" class="muted">
      {{if .AllowWrite}}
        Разрешены SELECT/INSERT/UPDATE/DELETE (временно).
      {{else}}
        Разрешены только SELECT-запросы.
      {{end}}
    </div>
    <textarea name="sql" rows="6" style="width:100%" placeholder="SELECT * FROM users LIMIT 10;"></textarea>
    <div class="row" style="margin-top:8px">
      <label>Timeout (сек): <input type="number" name="timeout" value="10" min="1" max="60"></label>
      <button type="submit">Выполнить</button>
    </div>
  </form>
  {{if .SQLResult}}
    <h4>Результат</h4>
    {{.SQLResult}}
  {{end}}
  {{if .SQLError}}
    <div class="bad">Ошибка: {{.SQLError}}</div>
  {{end}}
</div>
`))

var tableTpl = template.Must(template.Must(pageTpl.Clone()).New("content").Parse(`
<h2>Таблица: <code>{{.Table}}</code></h2>
<div class="card">
  <form method="get" action="/admin/db/t/{{.Table}}">
    <label>Лимит строк: <input type="number" name="limit" value="{{.Limit}}" min="1" max="10000"></label>
    <button type="submit">Обновить</button>
  </form>
</div>

<div class="card">
  {{if .Rows}}
    <table>
      <tr>
        {{range .Columns}}<th>{{.}}</th>{{end}}
      </tr>
      {{range .Rows}}
        <tr>
          {{range .}}<td>{{.}}</td>{{end}}
        </tr>
      {{end}}
    </table>
    <div class="muted" style="margin-top:6px">Показано {{.Count}} строк.</div>
  {{else}}
    <div class="muted">Нет данных.</div>
  {{end}}
</div>

<h3>SQL консоль (контекст таблицы)</h3>
<div class="card">
  <form method="post" action="/admin/db/sql">
    <textarea name="sql" rows="6" style="width:100%" placeholder="SELECT * FROM {{.Table}} LIMIT 10;"></textarea>
    <div class="row" style="margin-top:8px">
      <label>Timeout (сек): <input type="number" name="timeout" value="10" min="1" max="60"></label>
      <button type="submit">Выполнить</button>
    </div>
  </form>
</div>
`))

func main() {
	// Подключение к БД через ENV (PG* стандартные переменные)
	dsn := buildPgDSN()
	var err error
	db, err = sql.Open("pgx", dsn)
	must(err)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	must(db.Ping())

	mux := http.NewServeMux()
	// Вешаем под /admin/db* с BasicAuth
	mux.Handle("/admin/db", basicAuth(http.HandlerFunc(handleList)))
	mux.Handle("/admin/db/", basicAuth(http.HandlerFunc(routeDB)))

	addr := getenv("ADMIN_DB_ADDR", ":8088") // отдельный порт, чтобы не трогать основной
	log.Printf("DB admin listening on %s, write=%v", addr, allowWrite)
	must(http.ListenAndServe(addr, mux))
}

func routeDB(w http.ResponseWriter, r *http.Request) {
	// /admin/db/t/{table}
	path := strings.TrimPrefix(r.URL.Path, "/admin/db/")
	if path == "" || path == "/" {
		handleList(w, r)
		return
	}
	if strings.HasPrefix(path, "t/") && r.Method == http.MethodGet {
		handleTable(w, r, strings.TrimPrefix(path, "t/"))
		return
	}
	if path == "sql" && r.Method == http.MethodPost {
		handleSQL(w, r)
		return
	}
	http.NotFound(w, r)
}

func handleList(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, `
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
		ORDER BY tablename
	`)
	if err != nil {
		httpError(w, err)
		return
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			httpError(w, err)
			return
		}
		tables = append(tables, t)
	}
	data := map[string]any{
		"Title":      "Таблицы",
		"Tables":     tables,
		"AllowWrite": allowWrite,
		"DBName":     getenv("PGDATABASE", ""),
		"DBHost":     getenv("PGHOST", ""),
	}
	_ = listTpl.Execute(w, data)
}

func handleTable(w http.ResponseWriter, r *http.Request, table string) {
	if !isSafeIdent(table) {
		http.Error(w, "bad table name", http.StatusBadRequest)
		return
	}
	limit := rowLimitDefault
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 10000 {
			limit = n
		}
	}
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	// Забираем данные
	q := `SELECT * FROM "` + table + `" LIMIT $1`
	rows, err := db.QueryContext(ctx, q, limit)
	if err != nil {
		httpError(w, err)
		return
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		httpError(w, err)
		return
	}

	var outRows [][]string
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			httpError(w, err)
			return
		}
		out := make([]string, len(cols))
		for i, v := range vals {
			out[i] = sprintVal(v)
		}
		outRows = append(outRows, out)
	}

	data := map[string]any{
		"Title":      "Таблица: " + table,
		"Table":      table,
		"Columns":    cols,
		"Rows":       outRows,
		"Count":      len(outRows),
		"Limit":      limit,
		"AllowWrite": allowWrite,
		"DBName":     getenv("PGDATABASE", ""),
		"DBHost":     getenv("PGHOST", ""),
	}
	_ = tableTpl.Execute(w, data)
}

func handleSQL(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		httpError(w, err)
		return
	}
	sqlText := strings.TrimSpace(r.Form.Get("sql"))
	timeout := 10
	if v := r.Form.Get("timeout"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= 60 {
			timeout = n
		}
	}
	if sqlText == "" {
		http.Redirect(w, r, "/admin/db", http.StatusSeeOther)
		return
	}

	// Простейшая защита: только один стейтмент, без ';' в середине
	if strings.Count(sqlText, ";") > 1 {
		httpErrorMsg(w, "Разрешён один SQL-стейтмент за раз")
		return
	}
	// Ограничение по длине
	if len(sqlText) > 5000 {
		httpErrorMsg(w, "Слишком длинный запрос")
		return
	}
	// Блокируем DML, если не разрешено
	if !allowWrite && !isSelect(sqlText) {
		httpErrorMsg(w, "Разрешены только SELECT-запросы (включи ADMIN_DB_ALLOW_WRITE=1 для записи)")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeout)*time.Second)
	defer cancel()

	upper := strings.ToUpper(strings.TrimSpace(sqlText))
	if strings.HasPrefix(upper, "SELECT") || strings.HasPrefix(upper, "WITH") {
		rows, err := db.QueryContext(ctx, sqlText)
		if err != nil {
			renderSQL(w, nil, "", err)
			return
		}
		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			renderSQL(w, nil, "", err)
			return
		}
		var data [][]string
		for rows.Next() {
			vals := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				renderSQL(w, nil, "", err)
				return
			}
			out := make([]string, len(cols))
			for i, v := range vals {
				out[i] = sprintVal(v)
			}
			data = append(data, out)
		}
		html := renderTableHTML(cols, data)
		renderSQL(w, html, "", nil)
		return
	}

	// DML (INSERT/UPDATE/DELETE) — если разрешено
	res, err := db.ExecContext(ctx, sqlText)
	if err != nil {
		renderSQL(w, nil, "", err)
		return
	}
	aff, _ := res.RowsAffected()
	msg := map[string]any{"rows_affected": aff}
	b, _ := json.MarshalIndent(msg, "", "  ")
	renderSQL(w, nil, string(b), nil)
}

func renderSQL(w http.ResponseWriter, html template.HTML, jsonText string, err error) {
	var sqlResult template.HTML
	if len(html) > 0 {
		sqlResult = html
	} else if jsonText != "" {
		sqlResult = template.HTML(`<pre><code>` + template.HTMLEscapeString(jsonText) + `</code></pre>`)
	}
	data := map[string]any{
		"Title":      "SQL консоль",
		"SQLResult":  sqlResult,
		"SQLError":   errString(err),
		"AllowWrite": allowWrite,
		"DBName":     getenv("PGDATABASE", ""),
		"DBHost":     getenv("PGHOST", ""),
	}
	_ = listTpl.Execute(w, data)
}

func renderTableHTML(cols []string, data [][]string) template.HTML {
	var b strings.Builder
	b.WriteString(`<table><tr>`)
	for _, c := range cols {
		b.WriteString(`<th>` + template.HTMLEscapeString(c) + `</th>`)
	}
	b.WriteString(`</tr>`)
	for _, row := range data {
		b.WriteString(`<tr>`)
		for _, v := range row {
			b.WriteString(`<td>` + template.HTMLEscapeString(v) + `</td>`)
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</table>`)
	return template.HTML(b.String())
}

func isSelect(s string) bool {
	up := strings.ToUpper(strings.TrimSpace(s))
	return strings.HasPrefix(up, "SELECT") || strings.HasPrefix(up, "WITH")
}

var identRe = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func isSafeIdent(s string) bool {
	return identRe.MatchString(s)
}

func sprintVal(v any) string {
	if v == nil {
		return "NULL"
	}
	switch t := v.(type) {
	case []byte:
		// часто приходит []byte — преобразуем в строку
		if isPrintable(t) && len(t) <= 2000 {
			return string(t)
		}
		return "[bytes]"
	default:
		return toString(v)
	}
}

func isPrintable(b []byte) bool {
	for _, c := range b {
		if c < 9 || (c > 13 && c < 32) {
			return false
		}
	}
	return true
}

func toString(v any) string {
	switch x := v.(type) {
	case time.Time:
		return x.Format(time.RFC3339)
	default:
		return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(fmtSprint(v)), "\r", ""), "\x00", ""))
	}
}

func fmtSprint(v any) string { return strings.TrimSpace(strings.ReplaceAll(strings.TrimPrefix(strings.TrimSuffix(strings.ReplaceAll(strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(strings.ReplaceAll(
	// (минимизируем зависимость от fmt для компактности)
	// но можно просто fmt.Sprintf("%v", v)
	"", "", -1))), "", -1)), "", -1)), "", -1)) }

// ———————————————— утилиты ————————————————

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func buildPgDSN() string {
	// Поддерживает стандартные PG* ENV
	host := getenv("PGHOST", "127.0.0.1")
	port := getenv("PGPORT", "5432")
	user := getenv("PGUSER", "postgres")
	pass := getenv("PGPASSWORD", "")
	dbn  := getenv("PGDATABASE", "postgres")
	ssl  := getenv("PGSSLMODE", "disable")
	// формируем DSN для pgx stdlib:
	// https://pkg.go.dev/github.com/jackc/pgx/v5/stdlib
	return "host=" + host + " port=" + port + " user=" + user + " password=" + pass + " dbname=" + dbn + " sslmode=" + ssl
}

func httpError(w http.ResponseWriter, err error) {
	log.Println("ERR:", err)
	http.Error(w, errString(err), http.StatusInternalServerError)
}

func errString(err error) string {
	if err == nil { return "" }
	return err.Error()
}

func must(err error) {
	if err != nil { log.Fatal(err) }
}

// ———————————————— BasicAuth ————————————————
func basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		u, p, ok := r.BasicAuth()
		if !ok || u != basicUser || p != basicPass {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}