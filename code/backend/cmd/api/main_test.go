package main

import (
	"strings"
	"testing"
)

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
