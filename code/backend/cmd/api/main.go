package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.up.sql
var migrationFS embed.FS

const migrationDir = "migrations"

type task struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Status      string  `json:"status"`
	DueDate     *string `json:"due_date"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type tasksResponse struct {
	Tasks      []task  `json:"tasks"`
	NextCursor *string `json:"next_cursor"`
	HasMore    bool    `json:"has_more"`
}

type fieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error struct {
		Code      string       `json:"code"`
		Message   string       `json:"message"`
		Details   []fieldError `json:"details"`
		RequestID string       `json:"request_id"`
	} `json:"error"`
}

type app struct {
	db      *sql.DB
	limiter *rateLimiter
}

type requestIDKey struct{}

type rateLimiter struct {
	mu      sync.Mutex
	clients map[string]*bucket
	now     func() time.Time
}

type bucket struct {
	tokens float64
	seen   time.Time
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := applyMigrations(ctx, db); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	a := app{db: db, limiter: newRateLimiter()}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", a.health)
	mux.HandleFunc("GET /api/v1/tasks", a.listTasks)
	server := &http.Server{Addr: ":" + listenPort(), Handler: requestID(a.rateLimit(mux)), ReadHeaderTimeout: 5 * time.Second}
	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", server.Addr)
		errCh <- server.ListenAndServe()
	}()
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
	select {
	case sig := <-stopCh:
		log.Printf("shutdown signal: %s", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(ctx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func listenPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	if port := os.Getenv("APP_PORT"); port != "" {
		return port
	}
	return "8080"
}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = strconv.FormatInt(time.Now().UnixNano(), 36)
		}
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
	})
}

func getRequestID(r *http.Request) string {
	if id, ok := r.Context().Value(requestIDKey{}).(string); ok {
		return id
	}
	return "unknown"
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{clients: map[string]*bucket{}, now: time.Now}
}

func (a app) rateLimit(next http.Handler) http.Handler {
	if os.Getenv("RATE_LIMIT_DISABLED") == "true" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.limiter.allow(clientIP(r)) {
			w.Header().Set("Retry-After", "1")
			writeError(w, r, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests. Retry later.", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *rateLimiter) allow(key string) bool {
	const burst = 30.0
	const refillPerSecond = 2.0
	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()
	b := l.clients[key]
	if b == nil {
		l.clients[key] = &bucket{tokens: burst - 1, seen: now}
		return true
	}
	elapsed := now.Sub(b.seen).Seconds()
	b.tokens = min(burst, b.tokens+elapsed*refillPerSecond)
	b.seen = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

func (a app) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := a.db.PingContext(ctx); err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "UNAVAILABLE", "Database is unavailable.", nil)
		return
	}
	if _, err := a.db.ExecContext(ctx, "SELECT 1"); err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "UNAVAILABLE", "Database is unavailable.", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a app) listTasks(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > 0 {
		writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Request body is not supported.", nil)
		return
	}
	query := r.URL.Query()
	for k := range query {
		if k != "limit" && k != "cursor" {
			writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Unsupported query parameter.", nil)
			return
		}
	}
	limit := 200
	if v := query.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 200 || len(query["limit"]) != 1 {
			writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Limit must be between 1 and 200.", nil)
			return
		}
		limit = n
	}
	var afterTime, afterID string
	if c := query.Get("cursor"); c != "" {
		if len(query["cursor"]) != 1 {
			writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Cursor must be provided once.", nil)
			return
		}
		b, err := base64.RawURLEncoding.DecodeString(c)
		parts := strings.Split(string(b), ",")
		if err != nil || len(parts) != 2 || !validUUID(parts[1]) {
			writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Cursor is malformed.", nil)
			return
		}
		if _, err := time.Parse(time.RFC3339Nano, parts[0]); err != nil {
			writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Cursor is malformed.", nil)
			return
		}
		afterTime, afterID = parts[0], parts[1]
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	var rows *sql.Rows
	var err error
	if afterTime == "" {
		rows, err = a.db.QueryContext(ctx, `SELECT id::text, title, description, status, due_date::text, created_at, updated_at FROM tasks ORDER BY created_at ASC, id ASC LIMIT $1`, limit+1)
	} else {
		rows, err = a.db.QueryContext(ctx, `SELECT id::text, title, description, status, due_date::text, created_at, updated_at FROM tasks WHERE (created_at, id) > ($1::timestamptz, $2::uuid) ORDER BY created_at ASC, id ASC LIMIT $3`, afterTime, afterID, limit+1)
	}
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "UNAVAILABLE", "Database is unavailable.", nil)
		return
	}
	defer rows.Close()
	items := make([]task, 0, limit)
	for rows.Next() {
		var t task
		var desc, due sql.NullString
		var created, updated time.Time
		if err := rows.Scan(&t.ID, &t.Title, &desc, &t.Status, &due, &created, &updated); err != nil {
			writeError(w, r, http.StatusInternalServerError, "INTERNAL", "Unexpected server failure.", nil)
			return
		}
		if !isStatus(t.Status) {
			log.Printf("request_id=%s corrupt task id=%s", getRequestID(r), t.ID)
			writeError(w, r, http.StatusInternalServerError, "INTERNAL", "Unexpected server failure.", nil)
			return
		}
		if desc.Valid {
			t.Description = &desc.String
		}
		if due.Valid {
			t.DueDate = &due.String
		}
		t.CreatedAt = created.UTC().Format(time.RFC3339Nano)
		t.UpdatedAt = updated.UTC().Format(time.RFC3339Nano)
		items = append(items, t)
	}
	if err := rows.Err(); err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "UNAVAILABLE", "Database is unavailable.", nil)
		return
	}
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	var next *string
	if hasMore {
		last := items[len(items)-1]
		cursor := base64.RawURLEncoding.EncodeToString([]byte(last.CreatedAt + "," + last.ID))
		next = &cursor
	}
	writeJSON(w, http.StatusOK, tasksResponse{Tasks: items, NextCursor: next, HasMore: hasMore})
}

func isStatus(s string) bool { return s == "todo" || s == "doing" || s == "done" }

func validUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, r := range s {
		switch i {
		case 8, 13, 18, 23:
			if r != '-' {
				return false
			}
		default:
			if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F') {
				return false
			}
		}
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string, details []fieldError) {
	if details == nil {
		details = []fieldError{}
	}
	var out errorResponse
	out.Error.Code = code
	out.Error.Message = message
	out.Error.Details = details
	out.Error.RequestID = getRequestID(r)
	writeJSON(w, status, out)
}

func applyMigrations(ctx context.Context, db *sql.DB) error {
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	entries, err := migrationFS.ReadDir(migrationDir)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		version := strings.TrimSuffix(name, ".up.sql")
		if err := applyMigration(ctx, db, version, migrationDir+"/"+name); err != nil {
			return err
		}
	}
	return nil
}

func applyMigration(ctx context.Context, db *sql.DB, version string, path string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var exists bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'schema_migrations')`).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		var applied bool
		err = tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&applied)
		if err != nil {
			return err
		}
		if applied {
			return tx.Commit()
		}
	}
	body, err := migrationFS.ReadFile(path)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, string(body)); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING", version); err != nil {
		return err
	}
	log.Printf("applied migration %s", version)
	return tx.Commit()
}
