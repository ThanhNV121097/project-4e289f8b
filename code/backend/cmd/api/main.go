package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
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

type requestIDKey struct{}

var uuidRE = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

type server struct{ db *sql.DB }
type statusRecorder struct{ http.ResponseWriter; status int; bytes int }
type clientBucket struct{ count int; reset time.Time }
type rateLimiter struct{ mu sync.Mutex; buckets map[string]clientBucket; limit int; window time.Duration }

type task struct {
	ID string `json:"id"`
	Title string `json:"title"`
	Description *string `json:"description"`
	Status string `json:"status"`
	DueDate *string `json:"due_date"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type tasksResponse struct { Tasks []task `json:"tasks"`; NextCursor *string `json:"next_cursor"`; HasMore bool `json:"has_more"` }
type fieldError struct { Field string `json:"field"`; Code string `json:"code"`; Message string `json:"message"` }
type errorResponse struct { Error struct { Code string `json:"code"`; Message string `json:"message"`; Details []fieldError `json:"details"`; RequestID string `json:"request_id"` } `json:"error"` }
type nullableString struct { set bool; value *string }
type patchTaskRequest struct { any bool; Title *string; Description nullableString; Status *string; DueDate nullableString }

func main() { if err := run(); err != nil { log.Fatal(err) } }

func run() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" { return errors.New("DATABASE_URL is required") }
	db, err := sql.Open("pgx", databaseURL)
	if err != nil { return fmt.Errorf("open database: %w", err) }
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second); defer cancel()
	if err := applyMigrations(ctx, db); err != nil { return fmt.Errorf("apply migrations: %w", err) }
	s := &server{db: db}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /api/v1/tasks", s.listTasks)
	mux.HandleFunc("PATCH /api/v1/tasks/{task_id}", s.patchTask)
	handler := withRequestLog(withRateLimit(mux, newRateLimiter()))
	server := &http.Server{Addr: ":" + listenPort(), Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	errCh := make(chan error, 1)
	go func() { log.Printf("listening on %s", server.Addr); errCh <- server.ListenAndServe() }()
	stopCh := make(chan os.Signal, 1); signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
	select { case <-stopCh: ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second); defer cancel(); return server.Shutdown(ctx); case err := <-errCh: if errors.Is(err, http.ErrServerClosed) { return nil }; return err }
}

func newRateLimiter() *rateLimiter {
	if os.Getenv("RATE_LIMIT_DISABLED") == "true" { return nil }
	return &rateLimiter{buckets: map[string]clientBucket{}, limit: 120, window: time.Minute}
}

func withRateLimit(next http.Handler, limiter *rateLimiter) http.Handler {
	if limiter == nil { return next }
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wait := limiter.take(clientIP(r), time.Now())
		if wait > 0 { w.Header().Set("Retry-After", strconv.Itoa(int(wait.Seconds())+1)); writeError(w, r, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests.", nil); return }
		next.ServeHTTP(w, r)
	})
}

func (l *rateLimiter) take(key string, now time.Time) time.Duration {
	l.mu.Lock(); defer l.mu.Unlock()
	b := l.buckets[key]
	if now.After(b.reset) { b = clientBucket{reset: now.Add(l.window)} }
	if b.count >= l.limit { l.buckets[key] = b; return time.Until(b.reset) }
	b.count++; l.buckets[key] = b; return 0
}

func clientIP(r *http.Request) string { host, _, err := net.SplitHostPort(r.RemoteAddr); if err == nil { return host }; return r.RemoteAddr }

func withRequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := requestID(r)
		ctxReq := r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id))
		w.Header().Set("X-Request-Id", id)
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rec, ctxReq)
		log.Printf("request_id=%s method=%s path=%s status=%d duration=%s bytes=%d remote=%s", id, r.Method, r.URL.Path, rec.status, time.Since(start), rec.bytes, clientIP(r))
	})
}
func (r *statusRecorder) WriteHeader(status int) { r.status = status; r.ResponseWriter.WriteHeader(status) }
func (r *statusRecorder) Write(b []byte) (int, error) { n, err := r.ResponseWriter.Write(b); r.bytes += n; return n, err }
func requestID(r *http.Request) string { if id, ok := r.Context().Value(requestIDKey{}).(string); ok && id != "" { return id }; if id := r.Header.Get("X-Request-Id"); id != "" { return id }; return fmt.Sprintf("%d", time.Now().UnixNano()) }

func (s *server) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second); defer cancel()
	if err := s.db.PingContext(ctx); err != nil { writeError(w, r, http.StatusServiceUnavailable, "UNAVAILABLE", "Service unavailable.", nil); return }
	if _, err := s.db.ExecContext(ctx, "SELECT 1"); err != nil { writeError(w, r, http.StatusServiceUnavailable, "UNAVAILABLE", "Service unavailable.", nil); return }
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) listTasks(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > 0 { writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Request body is not allowed.", nil); return }
	q := r.URL.Query(); for key := range q { if key != "limit" && key != "cursor" { writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Unsupported query parameter.", nil); return } }
	if q.Get("cursor") != "" { writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Cursor is invalid.", nil); return }
	limit := 200
	if raw := q.Get("limit"); raw != "" { n, err := strconv.Atoi(raw); if err != nil || n < 1 || n > 200 { writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Limit is invalid.", nil); return }; limit = n }
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second); defer cancel()
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, title, description, status, due_date::text, created_at, updated_at FROM tasks ORDER BY created_at ASC, id ASC LIMIT $1`, limit+1)
	if err != nil { writeError(w, r, http.StatusServiceUnavailable, "UNAVAILABLE", "Service unavailable.", nil); return }
	defer rows.Close()
	items := make([]task, 0)
	for rows.Next() { item, err := scanTask(rows); if err != nil { writeError(w, r, http.StatusInternalServerError, "INTERNAL", "Unexpected server failure.", nil); return }; items = append(items, item) }
	if err := rows.Err(); err != nil { writeError(w, r, http.StatusServiceUnavailable, "UNAVAILABLE", "Service unavailable.", nil); return }
	hasMore := len(items) > limit; if hasMore { items = items[:limit] }
	writeJSON(w, http.StatusOK, tasksResponse{Tasks: items, HasMore: hasMore})
}

func (s *server) patchTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("task_id")
	if !uuidRE.MatchString(id) { writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Task ID is invalid.", []fieldError{{Field: "task_id", Code: "INVALID_UUID", Message: "Task ID must be a UUID."}}); return }
	if !hasJSONContentType(r.Header.Get("Content-Type")) { writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Content-Type must be application/json.", nil); return }
	req, details, bad := decodePatch(w, r); if bad { writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Request body is invalid.", details); return }
	if details = validatePatch(req); len(details) > 0 { writeError(w, r, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "Task fields are invalid.", details); return }
	set := []string{"updated_at = now()"}; args := []any{}
	if req.Title != nil { args = append(args, strings.TrimSpace(*req.Title)); set = append(set, fmt.Sprintf("title = $%d", len(args))) }
	if req.Description.set { args = append(args, req.Description.value); set = append(set, fmt.Sprintf("description = $%d", len(args))) }
	if req.Status != nil { args = append(args, *req.Status); set = append(set, fmt.Sprintf("status = $%d", len(args))) }
	if req.DueDate.set { args = append(args, req.DueDate.value); set = append(set, fmt.Sprintf("due_date = $%d", len(args))) }
	args = append(args, id)
	query := fmt.Sprintf(`UPDATE tasks SET %s WHERE id = $%d RETURNING id::text, title, description, status, due_date::text, created_at, updated_at`, strings.Join(set, ", "), len(args))
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second); defer cancel()
	item, err := scanTask(s.db.QueryRowContext(ctx, query, args...))
	if errors.Is(err, sql.ErrNoRows) { writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Task not found.", nil); return }
	if err != nil { writeError(w, r, http.StatusServiceUnavailable, "UNAVAILABLE", "Service unavailable.", nil); return }
	writeJSON(w, http.StatusOK, item)
}

func hasJSONContentType(v string) bool { mediaType, _, err := mime.ParseMediaType(v); return err == nil && mediaType == "application/json" }

func decodePatch(w http.ResponseWriter, r *http.Request) (patchTaskRequest, []fieldError, bool) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err != nil || len(strings.TrimSpace(string(body))) == 0 { return patchTaskRequest{}, []fieldError{{Field: "body", Code: "MALFORMED_JSON", Message: "Body must be a JSON object."}}, true }
	var top any; dec := json.NewDecoder(strings.NewReader(string(body))); dec.UseNumber()
	if err := dec.Decode(&top); err != nil || dec.Decode(&struct{}{}) != io.EOF { return patchTaskRequest{}, []fieldError{{Field: "body", Code: "MALFORMED_JSON", Message: "Body must be a JSON object."}}, true }
	raw, ok := top.(map[string]any); if !ok { return patchTaskRequest{}, []fieldError{{Field: "body", Code: "MALFORMED_JSON", Message: "Body must be a JSON object."}}, true }
	allowed := map[string]bool{"title": true, "description": true, "status": true, "due_date": true}; var req patchTaskRequest
	for key, value := range raw { if !allowed[key] { return req, []fieldError{{Field: "body", Code: "UNKNOWN_FIELD", Message: "Unknown field is not allowed."}}, true }; req.any = true; switch key {
	case "title": v, ok := value.(string); if !ok { return req, []fieldError{{Field: "body", Code: "MALFORMED_JSON", Message: "Field has wrong JSON type."}}, true }; req.Title = &v
	case "description": req.Description.set = true; if value != nil { v, ok := value.(string); if !ok { return req, []fieldError{{Field: "body", Code: "MALFORMED_JSON", Message: "Field has wrong JSON type."}}, true }; v = strings.TrimSpace(v); if v != "" { req.Description.value = &v } }
	case "status": v, ok := value.(string); if !ok { return req, []fieldError{{Field: "body", Code: "MALFORMED_JSON", Message: "Field has wrong JSON type."}}, true }; req.Status = &v
	case "due_date": req.DueDate.set = true; if value != nil { v, ok := value.(string); if !ok { return req, []fieldError{{Field: "body", Code: "MALFORMED_JSON", Message: "Field has wrong JSON type."}}, true }; v = strings.TrimSpace(v); req.DueDate.value = &v }
	} }
	return req, nil, false
}

func validatePatch(req patchTaskRequest) []fieldError {
	if !req.any { return []fieldError{{Field: "body", Code: "EMPTY_PATCH", Message: "At least one field is required."}} }
	var out []fieldError
	if req.Title != nil { title := strings.TrimSpace(*req.Title); if title == "" { out = append(out, fieldError{Field: "title", Code: "REQUIRED", Message: "Title is required."}) }; if len([]rune(title)) > 120 { out = append(out, fieldError{Field: "title", Code: "TOO_LONG", Message: "Title must be 120 characters or fewer."}) } }
	if req.Description.set && req.Description.value != nil && len([]rune(*req.Description.value)) > 2000 { out = append(out, fieldError{Field: "description", Code: "TOO_LONG", Message: "Description must be 2,000 characters or fewer."}) }
	if req.Status != nil && *req.Status != "todo" && *req.Status != "doing" && *req.Status != "done" { out = append(out, fieldError{Field: "status", Code: "INVALID_ENUM", Message: "Status must be todo, doing, or done."}) }
	if req.DueDate.set && req.DueDate.value != nil && !validDate(*req.DueDate.value) { out = append(out, fieldError{Field: "due_date", Code: "INVALID_DATE", Message: "Due date must be YYYY-MM-DD."}) }
	return out
}

func validDate(v string) bool { if len(v) != 10 { return false }; t, err := time.Parse("2006-01-02", v); return err == nil && t.Format("2006-01-02") == v }
type taskScanner interface{ Scan(dest ...any) error }
func scanTask(row taskScanner) (task, error) { var item task; var description, dueDate sql.NullString; var createdAt, updatedAt time.Time; err := row.Scan(&item.ID, &item.Title, &description, &item.Status, &dueDate, &createdAt, &updatedAt); if err != nil { return item, err }; if description.Valid { item.Description = &description.String }; if dueDate.Valid { item.DueDate = &dueDate.String }; item.CreatedAt = createdAt.UTC().Format(time.RFC3339); item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339); return item, nil }
func writeJSON(w http.ResponseWriter, status int, body any) { w.Header().Set("Content-Type", "application/json; charset=utf-8"); w.WriteHeader(status); _ = json.NewEncoder(w).Encode(body) }
func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string, details []fieldError) { w.Header().Set("Content-Type", "application/json; charset=utf-8"); var body errorResponse; body.Error.Code = code; body.Error.Message = message; body.Error.Details = details; if body.Error.Details == nil { body.Error.Details = []fieldError{} }; body.Error.RequestID = requestID(r); w.WriteHeader(status); _ = json.NewEncoder(w).Encode(body) }
func listenPort() string { if port := os.Getenv("PORT"); port != "" { return port }; if port := os.Getenv("APP_PORT"); port != "" { return port }; return "8080" }

func applyMigrations(ctx context.Context, db *sql.DB) error {
	if err := db.PingContext(ctx); err != nil { return err }
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil { return err }
	entries, err := migrationFS.ReadDir(migrationDir); if err != nil { return err }
	names := make([]string, 0, len(entries)); for _, entry := range entries { if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") { names = append(names, entry.Name()) } }
	sort.Strings(names)
	for _, name := range names { version := strings.TrimSuffix(name, ".up.sql"); if err := applyMigration(ctx, db, version, migrationDir+"/"+name); err != nil { return err } }
	return nil
}
func applyMigration(ctx context.Context, db *sql.DB, version string, path string) error { tx, err := db.BeginTx(ctx, nil); if err != nil { return err }; defer func(){ _ = tx.Rollback() }(); var applied bool; if err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&applied); err != nil { return err }; if applied { return tx.Commit() }; body, err := migrationFS.ReadFile(path); if err != nil { return err }; if _, err := tx.ExecContext(ctx, string(body)); err != nil { return err }; if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING", version); err != nil { return err }; log.Printf("applied migration %s", version); return tx.Commit() }
