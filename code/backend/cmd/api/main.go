package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/base64"
	"encoding/hex"
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
const maxBodyBytes = 16 << 10

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

type createTaskRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	DueDate     *string `json:"due_date"`
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
	db          *sql.DB
	limiter     *rateLimiter
	idempotency *idempotencyStore
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

type idempotencyStore struct {
	mu      sync.Mutex
	entries map[string]idempotencyEntry
	now     func() time.Time
}

type idempotencyEntry struct {
	bodyHash string
	task     task
	expires  time.Time
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
	a := app{db: db, limiter: newRateLimiter(), idempotency: newIdempotencyStore()}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", a.health)
	mux.HandleFunc("GET /api/v1/tasks", a.listTasks)
	mux.HandleFunc("POST /api/v1/tasks", a.createTask)
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
	items := make([]task, 0, limit+1)
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, "INTERNAL", "Unexpected server failure.", nil)
			return
		}
		if !isStatus(t.Status) {
			log.Printf("request_id=%s corrupt task id=%s", getRequestID(r), t.ID)
			writeError(w, r, http.StatusInternalServerError, "INTERNAL", "Unexpected server failure.", nil)
			return
		}
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

func (a app) createTask(w http.ResponseWriter, r *http.Request) {
	body, err := readJSONBody(w, r)
	if err != nil {
		return
	}
	key := r.Header.Get("Idempotency-Key")
	bodyHash := hashBody(body)
	if key != "" {
		if saved, ok, sameBody := a.idempotency.get(r.Method + " " + r.URL.Path + " " + key, bodyHash); ok {
			if !sameBody {
				writeError(w, r, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "Idempotency key was reused with a different body.", nil)
				return
			}
			w.Header().Set("Location", "/api/v1/tasks/"+saved.ID)
			writeJSON(w, http.StatusCreated, saved)
			return
		}
	}
	input, details, badRequest := parseCreateBody(body)
	if badRequest {
		writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Request body is invalid.", details)
		return
	}
	if len(details) > 0 {
		writeError(w, r, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "Task could not be created.", details)
		return
	}
	status := "todo"
	if input.Status != nil {
		status = *input.Status
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	row := a.db.QueryRowContext(ctx, `INSERT INTO tasks (title, description, status, due_date) VALUES ($1, $2, $3, $4) RETURNING id::text, title, description, status, due_date::text, created_at, updated_at`, input.Title, input.Description, status, input.DueDate)
	saved, err := scanTask(row)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "UNAVAILABLE", "Database is unavailable.", nil)
		return
	}
	if key != "" {
		a.idempotency.put(r.Method+" "+r.URL.Path+" "+key, bodyHash, saved)
	}
	w.Header().Set("Location", "/api/v1/tasks/"+saved.ID)
	writeJSON(w, http.StatusCreated, saved)
}
func readJSONBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	ct := strings.ToLower(strings.TrimSpace(strings.Split(r.Header.Get("Content-Type"), ";")[0]))
	if ct != "application/json" {
		writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Content-Type must be application/json.", nil)
		return nil, errors.New("bad content type")
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes+1))
	if err != nil || len(body) > maxBodyBytes {
		writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Request body is too large.", nil)
		return nil, errors.New("bad body")
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Request body must be a JSON object.", nil)
		return nil, errors.New("empty body")
	}
	return body, nil
}

func parseCreateBody(body []byte) (createTaskRequest, []fieldError, bool) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return createTaskRequest{}, []fieldError{{Field: "body", Code: "MALFORMED_JSON", Message: "JSON cannot be parsed."}}, true
	}
	if raw == nil {
		return createTaskRequest{}, []fieldError{{Field: "body", Code: "MALFORMED_JSON", Message: "JSON body must be an object."}}, true
	}
	allowed := map[string]bool{"title": true, "description": true, "status": true, "due_date": true}
	for k := range raw {
		if !allowed[k] {
			return createTaskRequest{}, []fieldError{{Field: "body", Code: "UNKNOWN_FIELD", Message: "Unknown field is not supported."}}, true
		}
	}
	var req createTaskRequest
	var details []fieldError
	if v, ok := raw["title"]; !ok {
		details = append(details, fieldError{Field: "title", Code: "REQUIRED", Message: "Title is required."})
	} else if err := json.Unmarshal(v, &req.Title); err != nil {
		return req, nil, true
	} else {
		req.Title = strings.TrimSpace(req.Title)
		if req.Title == "" {
			details = append(details, fieldError{Field: "title", Code: "REQUIRED", Message: "Title is required."})
		} else if len([]rune(req.Title)) > 120 {
			details = append(details, fieldError{Field: "title", Code: "TOO_LONG", Message: "Title must be 120 characters or fewer."})
		}
	}
	if v, ok := raw["description"]; ok {
		if string(v) == "null" {
			req.Description = nil
		} else {
			var desc string
			if err := json.Unmarshal(v, &desc); err != nil {
				return req, nil, true
			}
			desc = strings.TrimSpace(desc)
			if desc != "" {
				req.Description = &desc
			}
			if len([]rune(desc)) > 2000 {
				details = append(details, fieldError{Field: "description", Code: "TOO_LONG", Message: "Description must be 2,000 characters or fewer."})
			}
		}
	}
	if v, ok := raw["status"]; ok {
		var status string
		if err := json.Unmarshal(v, &status); err != nil {
			return req, nil, true
		}
		if !isStatus(status) {
			details = append(details, fieldError{Field: "status", Code: "INVALID_ENUM", Message: "Status must be todo, doing, or done."})
		} else {
			req.Status = &status
		}
	}
	if v, ok := raw["due_date"]; ok {
		if string(v) == "null" {
			req.DueDate = nil
		} else {
			var due string
			if err := json.Unmarshal(v, &due); err != nil {
				return req, nil, true
			}
			if !validDateOnly(due) {
				details = append(details, fieldError{Field: "due_date", Code: "INVALID_DATE", Message: "Due date must be a real YYYY-MM-DD date."})
			} else {
				req.DueDate = &due
			}
		}
	}
	return req, details, false
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(row taskScanner) (task, error) {
	var t task
	var desc, due sql.NullString
	var created, updated time.Time
	if err := row.Scan(&t.ID, &t.Title, &desc, &t.Status, &due, &created, &updated); err != nil {
		return t, err
	}
	if desc.Valid {
		t.Description = &desc.String
	}
	if due.Valid {
		t.DueDate = &due.String
	}
	t.CreatedAt = created.UTC().Format(time.RFC3339Nano)
	t.UpdatedAt = updated.UTC().Format(time.RFC3339Nano)
	return t, nil
}

func isStatus(s string) bool { return s == "todo" || s == "doing" || s == "done" }

func validDateOnly(s string) bool {
	if len(s) != 10 {
		return false
	}
	date, err := time.Parse("2006-01-02", s)
	return err == nil && date.Format("2006-01-02") == s
}

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

func hashBody(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func newIdempotencyStore() *idempotencyStore {
	return &idempotencyStore{entries: map[string]idempotencyEntry{}, now: time.Now}
}

func (s *idempotencyStore) get(key, bodyHash string) (task, bool, bool) {
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[key]
	if !ok {
		return task{}, false, false
	}
	if now.After(entry.expires) {
		delete(s.entries, key)
		return task{}, false, false
	}
	return entry.task, true, entry.bodyHash == bodyHash
}

func (s *idempotencyStore) put(key, bodyHash string, t task) {
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, entry := range s.entries {
		if now.After(entry.expires) {
			delete(s.entries, k)
		}
	}
	s.entries[key] = idempotencyEntry{bodyHash: bodyHash, task: t, expires: now.Add(10 * time.Minute)}
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
