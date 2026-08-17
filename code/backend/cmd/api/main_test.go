package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCreateTaskHandlerPersistsAndReturnsSavedTask(t *testing.T) {
	db := openCreateTaskTestDB(t)
	a := app{db: db, idempotency: newIdempotencyStore()}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(`{"title":" Pay invoice ","description":" Due before Friday ","status":"doing","due_date":"2026-09-30"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	a.createTask(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Location"); got != "/api/v1/tasks/550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("Location=%q", got)
	}
	body := rec.Body.String()
	for _, want := range []string{`"title":"Pay invoice"`, `"description":"Due before Friday"`, `"status":"doing"`, `"due_date":"2026-09-30"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %s in %s", want, body)
		}
	}
	if got := lastCreateExec; got != "Pay invoice|Due before Friday|doing|2026-09-30" {
		t.Fatalf("persisted args=%q", got)
	}
}

func TestParseCreateBodyValidatesAndNormalizes(t *testing.T) {
	desc := "notes"
	body := []byte(`{"title":"  Pay invoice  ","description":"  notes  ","status":"doing","due_date":"2026-09-30"}`)

	got, details, badRequest := parseCreateBody(body)
	if badRequest || len(details) != 0 {
		t.Fatalf("parseCreateBody() badRequest=%v details=%v", badRequest, details)
	}
	if got.Title != "Pay invoice" || got.Description == nil || *got.Description != desc || got.Status == nil || *got.Status != "doing" || got.DueDate == nil || *got.DueDate != "2026-09-30" {
		t.Fatalf("parseCreateBody() = %+v", got)
	}
}

func TestParseCreateBodyRejectsUnknownField(t *testing.T) {
	_, details, badRequest := parseCreateBody([]byte(`{"title":"Pay invoice","typo":"ignored"}`))
	if !badRequest || len(details) != 1 || details[0].Code != "UNKNOWN_FIELD" {
		t.Fatalf("parseCreateBody() badRequest=%v details=%v", badRequest, details)
	}
}

func TestParseCreateBodyRejectsCreateValidationErrors(t *testing.T) {
	_, details, badRequest := parseCreateBody([]byte(`{"title":"   ","description":"ok","status":"later","due_date":"2026-02-30"}`))
	if badRequest {
		t.Fatal("parseCreateBody() returned bad request for validation errors")
	}
	want := map[string]string{"title": "REQUIRED", "status": "INVALID_ENUM", "due_date": "INVALID_DATE"}
	if len(details) != len(want) {
		t.Fatalf("details=%v", details)
	}
	for _, detail := range details {
		if want[detail.Field] != detail.Code {
			t.Fatalf("unexpected detail=%+v all=%v", detail, details)
		}
	}
}

func TestParseCreateBodyAcceptsLimits(t *testing.T) {
	title := strings.Repeat("t", 120)
	desc := strings.Repeat("d", 2000)
	_, details, badRequest := parseCreateBody([]byte(`{"title":"` + title + `","description":"` + desc + `"}`))
	if badRequest || len(details) != 0 {
		t.Fatalf("parseCreateBody() badRequest=%v details=%v", badRequest, details)
	}
}

func TestParseCreateBodyRejectsTooLongFields(t *testing.T) {
	title := strings.Repeat("t", 121)
	desc := strings.Repeat("d", 2001)
	_, details, badRequest := parseCreateBody([]byte(`{"title":"` + title + `","description":"` + desc + `"}`))
	if badRequest {
		t.Fatal("parseCreateBody() returned bad request for validation errors")
	}
	want := map[string]string{"title": "TOO_LONG", "description": "TOO_LONG"}
	if len(details) != len(want) {
		t.Fatalf("details=%v", details)
	}
	for _, detail := range details {
		if want[detail.Field] != detail.Code {
			t.Fatalf("unexpected detail=%+v all=%v", detail, details)
		}
	}
}

var lastCreateExec string

func openCreateTaskTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("create_task_test", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func init() { sql.Register("create_task_test", createTaskTestDriver{}) }

type createTaskTestDriver struct{}
type createTaskTestConn struct{}
type createTaskTestRows struct{ sent bool }

func (createTaskTestDriver) Open(string) (driver.Conn, error) { return createTaskTestConn{}, nil }
func (createTaskTestConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (createTaskTestConn) Close() error { return nil }
func (createTaskTestConn) Begin() (driver.Tx, error) { return nil, driver.ErrSkip }
func (createTaskTestConn) QueryContext(_ context.Context, _ string, args []driver.NamedValue) (driver.Rows, error) {
	parts := make([]string, len(args))
	for i, arg := range args {
		if arg.Value != nil {
			parts[i] = arg.Value.(string)
		}
	}
	lastCreateExec = strings.Join(parts, "|")
	return &createTaskTestRows{}, nil
}
func (createTaskTestRows) Columns() []string {
	return []string{"id", "title", "description", "status", "due_date", "created_at", "updated_at"}
}
func (createTaskTestRows) Close() error { return nil }
func (r *createTaskTestRows) Next(dest []driver.Value) error {
	if r.sent {
		return io.EOF
	}
	r.sent = true
	now := time.Date(2026, 8, 17, 10, 0, 0, 0, time.UTC)
	dest[0] = "550e8400-e29b-41d4-a716-446655440000"
	dest[1] = "Pay invoice"
	dest[2] = "Due before Friday"
	dest[3] = "doing"
	dest[4] = "2026-09-30"
	dest[5] = now
	dest[6] = now
	return nil
}
