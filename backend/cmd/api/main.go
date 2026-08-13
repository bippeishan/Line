package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type config struct {
	port        string
	databaseURL string
}

type pattern struct {
	ID              int64  `json:"id"`
	Title           string `json:"title"`
	Category        string `json:"category"`
	Difficulty      string `json:"difficulty"`
	Yarn            string `json:"yarn"`
	HookSize        string `json:"hookSize"`
	DurationMinutes int    `json:"durationMinutes"`
}

func main() {
	cfg := config{port: env("PORT", "8080"), databaseURL: os.Getenv("DATABASE_URL")}
	var db *sql.DB
	if cfg.databaseURL != "" {
		var err error
		db, err = sql.Open("mysql", cfg.databaseURL)
		if err != nil {
			log.Fatalf("open database: %v", err)
		}
		defer db.Close()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", healthHandler(db))
	mux.HandleFunc("GET /api/patterns", patternsHandler(db))

	server := &http.Server{Addr: ":" + cfg.port, Handler: cors(mux), ReadHeaderTimeout: 5 * time.Second}
	log.Printf("Line API listening on http://localhost:%s", cfg.port)
	log.Fatal(server.ListenAndServe())
}

func healthHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "ok"
		code := http.StatusOK
		if db == nil {
			status, code = "database_not_configured", http.StatusServiceUnavailable
		} else if err := db.PingContext(r.Context()); err != nil {
			status, code = "database_unavailable", http.StatusServiceUnavailable
		}
		writeJSON(w, code, map[string]string{"status": status, "service": "line-api"})
	}
}

func patternsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database is not configured"})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		rows, err := db.QueryContext(ctx, `SELECT id, title, category, difficulty, yarn, hook_size, duration_minutes FROM patterns ORDER BY created_at DESC LIMIT 24`)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not load patterns"})
			return
		}
		defer rows.Close()
		items := make([]pattern, 0)
		for rows.Next() {
			var item pattern
			if err := rows.Scan(&item.ID, &item.Title, &item.Category, &item.Difficulty, &item.Yarn, &item.HookSize, &item.DurationMinutes); err != nil {
				writeJSON(w, 500, map[string]string{"error": "could not read pattern"})
				return
			}
			items = append(items, item)
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
	}
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
