package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidatePatch(t *testing.T) {
	blank := "  "
	tooLongTitle := makeString(121)
	badStatus := "blocked"
	badDate := "2026-02-31"
	tooLongDescription := makeString(2001)

	cases := []struct {
		name string
		req  patchTaskRequest
		want string
	}{
		{"empty patch", patchTaskRequest{}, "body:EMPTY_PATCH"},
		{"blank title", patchTaskRequest{any: true, Title: &blank}, "title:REQUIRED"},
		{"long title", patchTaskRequest{any: true, Title: &tooLongTitle}, "title:TOO_LONG"},
		{"long description", patchTaskRequest{any: true, Description: nullableString{set: true, value: &tooLongDescription}}, "description:TOO_LONG"},
		{"bad status", patchTaskRequest{any: true, Status: &badStatus}, "status:INVALID_ENUM"},
		{"bad due date", patchTaskRequest{any: true, DueDate: nullableString{set: true, value: &badDate}}, "due_date:INVALID_DATE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := validatePatch(tc.req)
			if len(got) == 0 || got[0].Field+":"+got[0].Code != tc.want {
				t.Fatalf("got %#v, want %s", got, tc.want)
			}
		})
	}
}

func TestHasJSONContentType(t *testing.T) {
	cases := map[string]bool{
		"application/json":                 true,
		"application/json; charset=utf-8": true,
		"text/plain":                       false,
		"application/jsonx":                false,
		"":                                 false,
	}
	for contentType, want := range cases {
		if got := hasJSONContentType(contentType); got != want {
			t.Fatalf("hasJSONContentType(%q) = %v, want %v", contentType, got, want)
		}
	}
}

func TestDecodePatchRejectsBadBodies(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{"malformed", `{"title":`, "body:MALFORMED_JSON"},
		{"array", `[]`, "body:MALFORMED_JSON"},
		{"unknown field", `{"title":"x","bogus":true}`, "body:UNKNOWN_FIELD"},
		{"wrong type", `{"status":3}`, "body:MALFORMED_JSON"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/id", strings.NewReader(tc.body))
			w := httptest.NewRecorder()
			_, details, bad := decodePatch(w, r)
			if !bad || len(details) == 0 || details[0].Field+":"+details[0].Code != tc.want {
				t.Fatalf("bad=%v details=%#v, want %s", bad, details, tc.want)
			}
		})
	}
}

func makeString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'x'
	}
	return string(b)
}
