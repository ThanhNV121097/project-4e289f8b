package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code      string        `json:"code"`
	Message   string        `json:"message"`
	Details   []errorDetail `json:"details"`
	RequestID string        `json:"request_id"`
}

type errorDetail struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type clientBucket struct {
	windowStart time.Time
	count       int
}

type rateLimiter struct {
	mu       sync.Mutex
	disabled bool
	limit    int
	window   time.Duration
	clients  map[string]clientBucket
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

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler(db))
	mux.HandleFunc("GET /api/v1/tasks", listTasksHandler(db))
	mux.HandleFunc("DELETE /api/v1/tasks/{task_id}", deleteTaskHandler(db))

	addr := ":" + listenPort()
	server := &http.Server{
		Addr:              addr,
		Handler:           withRequestID(newRateLimiter().middleware(mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", addr)
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

func healthHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			writeError(w, r, http.StatusServiceUnavailable, "UNAVAILABLE", "Database unavailable.", nil)
			return
		}
		if _, err := db.ExecContext(ctx, "SELECT 1"); err != nil {
			writeError(w, r, http.StatusServiceUnavailable, "UNAVAILABLE", "Database unavailable.", nil)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func listTasksHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if hasNonEmptyBody(w, r) {
			writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "GET /api/v1/tasks does not accept a request body.", nil)
			return
		}
		query := r.URL.Query()
		for key := range query {
			if key != "limit" && key != "cursor" {
				writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Unsupported query parameter.", nil)
				return
			}
		}
		limit := 200
		if raw := query.Get("limit"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 1 || parsed > 200 {
				writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Limit must be an integer from 1 to 200.", nil)
				return
			}
			limit = parsed
		}
		cursorCreated, cursorID, hasCursor, ok := parseCursor(query.Get("cursor"))
		if !ok {
			writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Cursor is malformed.", nil)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		items, err := listTasks(ctx, db, limit+1, cursorCreated, cursorID, hasCursor)
		if err != nil {
			status, code := dbError(err)
			writeError(w, r, status, code, "Could not load tasks.", nil)
			return
		}
		hasMore := len(items) > limit
		if hasMore {
			items = items[:limit]
		}
		var next *string
		if hasMore {
			last := items[len(items)-1]
			encoded := encodeCursor(last.CreatedAt, last.ID)
			next = &encoded
		}
		writeJSON(w, http.StatusOK, tasksResponse{Tasks: items, NextCursor: next, HasMore: hasMore})
	}
}

func deleteTaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("task_id")
		if !isUUID(id) {
			writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Task ID must be a UUID.", []errorDetail{{Field: "task_id", Code: "INVALID_UUID", Message: "Task ID must be a UUID."}})
			return
		}
		if hasNonEmptyBody(w, r) {
			writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "DELETE /api/v1/tasks/{task_id} does not accept a request body.", nil)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		result, err := db.ExecContext(ctx, "DELETE FROM tasks WHERE id = $1", id)
		if err != nil {
			status, code := dbError(err)
			writeError(w, r, status, code, "Could not delete task.", nil)
			return
		}
		rows, err := result.RowsAffected()
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, "INTERNAL", "Could not delete task.", nil)
			return
		}
		if rows == 0 {
			writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Task was not found.", nil)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func hasNonEmptyBody(w http.ResponseWriter, r *http.Request) bool {
	if r.Body == nil || r.Body == http.NoBody || r.ContentLength == 0 {
		return false
	}
	if r.ContentLength > 0 {
		return true
	}
	body := http.MaxBytesReader(w, r.Body, 1)
	buf := make([]byte, 1)
	n, err := body.Read(buf)
	if n > 0 {
		return true
	}
	return err != nil && !errors.Is(err, io.EOF)
}

func listTasks(ctx context.Context, db *sql.DB, limit int, cursorCreated time.Time, cursorID string, hasCursor bool) ([]task, error) {
	query := `SELECT id::text, title, description, status, due_date, created_at, updated_at FROM tasks ORDER BY created_at ASC, id ASC LIMIT $1`
	args := []any{limit}
	if hasCursor {
		query = `SELECT id::text, title, description, status, due_date, created_at, updated_at FROM tasks WHERE (created_at, id) > ($1, $2::uuid) ORDER BY created_at ASC, id ASC LIMIT $3`
		args = []any{cursorCreated, cursorID, limit}
	}
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []task{}
	for rows.Next() {
		var item task
		var description sql.NullString
		var dueDate sql.NullTime
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.Title, &description, &item.Status, &dueDate, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if item.Status != "todo" && item.Status != "doing" && item.Status != "done" {
			return nil, errors.New("invalid task status from database")
		}
		if description.Valid {
			item.Description = &description.String
		}
		if dueDate.Valid {
			formatted := dueDate.Time.Format("2006-01-02")
			item.DueDate = &formatted
		}
		item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
		items = append(items, item)
	}
	return items, rows.Err()
}

func parseCursor(raw string) (time.Time, string, bool, bool) {
	if raw == "" {
		return time.Time{}, "", false, true
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return time.Time{}, "", false, false
	}
	parts := strings.Split(string(decoded), ",")
	if len(parts) != 2 || !isUUID(parts[1]) {
		return time.Time{}, "", false, false
	}
	createdAt, err := time.Parse(time.RFC3339, parts[0])
	if err != nil {
		return time.Time{}, "", false, false
	}
	return createdAt, parts[1], true, true
}

func encodeCursor(createdAt string, id string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(createdAt + "," + id))
}

func dbError(err error) (int, string) {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return http.StatusServiceUnavailable, "UNAVAILABLE"
	}
	return http.StatusInternalServerError, "INTERNAL"
}

func newRateLimiter() *rateLimiter {
	disabled, _ := strconv.ParseBool(os.Getenv("RATE_LIMIT_DISABLED"))
	return &rateLimiter{disabled: disabled, limit: 120, window: time.Minute, clients: map[string]clientBucket{}}
}

func (l *rateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if l.allow(clientIP(r)) {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Retry-After", "60")
		writeError(w, r, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests. Try again later.", nil)
	})
}

func (l *rateLimiter) allow(client string) bool {
	if l.disabled {
		return true
	}
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	bucket := l.clients[client]
	if now.Sub(bucket.windowStart) >= l.window {
		l.clients[client] = clientBucket{windowStart: now, count: 1}
		return true
	}
	if bucket.count >= l.limit {
		return false
	}
	bucket.count++
	l.clients[client] = bucket
	return true
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || host == "" {
		return r.RemoteAddr
	}
	return host
}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
		}
		w.Header().Set("X-Request-Id", requestID)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, requestID)))
	})
}

type requestIDKey struct{}

func requestID(r *http.Request) string {
	id, _ := r.Context().Value(requestIDKey{}).(string)
	if id == "" {
		return "unknown"
	}
	return id
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code string, message string, details []errorDetail) {
	if details == nil {
		details = []errorDetail{}
	}
	writeJSON(w, status, errorResponse{Error: apiError{Code: code, Message: message, Details: details, RequestID: requestID(r)}})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func isUUID(value string) bool {
	if len(value) != 36 {
		return false
	}
	for i, c := range value {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
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
	err = tx.QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'schema_migrations'
	)`).Scan(&exists)
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
