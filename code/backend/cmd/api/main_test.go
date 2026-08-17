package main

import "testing"

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
			if len(got) == 0 || got[0].Field+":"+got[0].Code != tc.want { t.Fatalf("got %#v, want %s", got, tc.want) }
		})
	}
}

func makeString(n int) string { b := make([]byte, n); for i := range b { b[i] = 'x' }; return string(b) }
