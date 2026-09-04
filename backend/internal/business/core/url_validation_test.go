package core

import (
	"errors"
	"testing"
)

func TestValidateBaseURL(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"valid https", "https://gitlab.example.com", "https://gitlab.example.com", false},
		{"valid http", "http://localhost:8090", "http://localhost:8090", false},
		{"trims trailing slash", "https://gitlab.example.com/", "https://gitlab.example.com", false},
		{"trims whitespace", "  https://gitlab.example.com  ", "https://gitlab.example.com", false},
		{"empty", "", "", true},
		{"whitespace only", "   ", "", true},
		{"missing scheme", "gitlab.example.com", "", true},
		{"disallowed scheme", "ftp://gitlab.example.com", "", true},
		{"javascript scheme", "javascript:alert(1)", "", true},
		{"scheme without host", "https://", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := validateBaseURL(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("validateBaseURL(%q) expected error, got nil", tc.input)
				}
				if !errors.Is(err, ErrInvalidBaseURL) {
					t.Fatalf("validateBaseURL(%q) error = %v, want wrapped ErrInvalidBaseURL", tc.input, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("validateBaseURL(%q) unexpected error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("validateBaseURL(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
