package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestValidatePatch(t *testing.T) {
	blank := "  "
	tooLongTitle := makeString(121)
	badStatus := "blocked"
	badDate := "2026-02-31"
	tooLongDescription := makeString(2001)

	cases := []struct { name string; req patchTaskRequest; want string }{
		{"empty patch", patchTaskRequest{}, "body:EMPTY_PATCH"},
		{"blank title", patchTaskRequest{any: true, Title: &blank}, "title:REQUIRED"},
		{"long title", patchTaskRequest{any: true, Title: &tooLongTitle}, "title:TOO_LONG"},
		{"long description", patchTaskRequest{any: true, Description: nullableString{set: true, value: &tooLongDescription}}, "description:TOO_LONG"},
		{"bad status", patchTaskRequest{any: true, Status: &badStatus}, "status:INVALID_ENUM"},
		{"bad due date", patchTaskRequest{any: true, DueDate: nullableString{set: true, value: &badDate}}, "due_date:INVALID_DATE"},
	}
	for _, tc := range cases { t.Run(tc.name, func(t *testing.T) { got := validatePatch(tc.req); if len(got) == 0 || got[0].Field+":"+got[0].Code != tc.want { t.Fatalf("got %#v, want %s", got, tc.want) } }) }
}

func TestCursorRoundTrip(t *testing.T) {
	want := listCursor{CreatedAt: "2026-08-17T10:00:00Z", ID: "550e8400-e29b-41d4-a716-446655440000"}
	got, err := decodeCursor(encodeCursor(want))
	if err != nil { t.Fatal(err) }
	if got != want { t.Fatalf("cursor = %#v, want %#v", got, want) }
}

func TestDecodeCursorRejectsBadValues(t *testing.T) {
	badUUID := encodeCursor(listCursor{CreatedAt: "2026-08-17T10:00:00Z", ID: "nope"})
	badTime := encodeCursor(listCursor{CreatedAt: "2026-08-17", ID: "550e8400-e29b-41d4-a716-446655440000"})
	for _, raw := range []string{"not-base64", badUUID, badTime} { if _, err := decodeCursor(raw); err == nil { t.Fatalf("decodeCursor(%q) succeeded, want error", raw) } }
}

func TestHasJSONContentType(t *testing.T) {
	cases := map[string]bool{"application/json": true, "application/json; charset=utf-8": true, "text/plain": false, "application/jsonx": false, "": false}
	for contentType, want := range cases { if got := hasJSONContentType(contentType); got != want { t.Fatalf("hasJSONContentType(%q) = %v, want %v", contentType, got, want) } }
}

func TestDecodePatchRejectsBadBodies(t *testing.T) {
	cases := []struct { name string; body string; want string }{{"malformed", `{"title":`, "body:MALFORMED_JSON"}, {"array", `[]`, "body:MALFORMED_JSON"}, {"unknown field", `{"title":"x","bogus":true}`, "body:UNKNOWN_FIELD"}, {"wrong type", `{"status":3}`, "body:MALFORMED_JSON"}}
	for _, tc := range cases { t.Run(tc.name, func(t *testing.T) { r := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/id", strings.NewReader(tc.body)); w := httptest.NewRecorder(); _, details, bad := decodePatch(w, r); if !bad || len(details) == 0 || details[0].Field+":"+details[0].Code != tc.want { t.Fatalf("bad=%v details=%#v, want %s", bad, details, tc.want) } }) }
}

func TestRateLimiterReturnsRetryAfter(t *testing.T) {
	limiter := &rateLimiter{buckets: map[string]clientBucket{}, limit: 1, window: time.Minute}
	now := time.Now()
	if wait := limiter.take("127.0.0.1", now); wait != 0 { t.Fatalf("first wait = %s, want 0", wait) }
	if wait := limiter.take("127.0.0.1", now); wait <= 0 { t.Fatalf("second wait = %s, want positive", wait) }
}

func TestRequestIDEchoesHeader(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Request-Id", "abc")
	w := httptest.NewRecorder()
	withRequestLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { writeError(w, r, http.StatusInternalServerError, "INTERNAL", "boom", nil) })).ServeHTTP(w, r)
	if got := w.Header().Get("X-Request-Id"); got != "abc" { t.Fatalf("X-Request-Id = %q, want abc", got) }
	var body errorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil { t.Fatal(err) }
	if body.Error.RequestID != "abc" { t.Fatalf("error.request_id = %q, want abc", body.Error.RequestID) }
}

func TestGeneratedRequestIDMatchesHeaderAndError(t *testing.T) {
	w := httptest.NewRecorder()
	withRequestLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { writeError(w, r, http.StatusInternalServerError, "INTERNAL", "boom", nil) })).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	var body errorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil { t.Fatal(err) }
	if got := w.Header().Get("X-Request-Id"); got == "" || got != body.Error.RequestID { t.Fatalf("header request id %q, error request id %q", got, body.Error.RequestID) }
}

func makeString(n int) string { b := make([]byte, n); for i := range b { b[i] = 'x' }; return string(b) }
