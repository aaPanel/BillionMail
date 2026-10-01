package mail_service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessForwardUsers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"newline split", "a@x.com\nb@x.com", []string{"a@x.com", "b@x.com"}},
		{"escaped newline split", `a@x.com\nb@x.com`, []string{"a@x.com", "b@x.com"}},
		{"trims whitespace", " a@x.com \n b@x.com ", []string{"a@x.com", "b@x.com"}},
		{"filters empty", "a@x.com\n\nb@x.com\n", []string{"a@x.com", "b@x.com"}},
		{"single address", "a@x.com", []string{"a@x.com"}},
		{"empty string", "", []string{}},
		{"only whitespace", "  \n  ", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProcessForwardUsers(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsDomain(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid domain", "example.com", true},
		{"subdomain", "mail.example.com", true},
		{"hyphen", "my-domain.co.uk", true},
		{"IP address", "192.168.1.1", false},
		{"empty", "", false},
		{"no tld", "localhost", false},
		{"single char tld", "example.a", false},
		{"starts with hyphen", "-example.com", false},
		{"email address", "user@example.com", false},
		{"special chars", "exam!ple.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsDomain(tt.input))
		})
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"email address", "user@example.com", "example.com"},
		{"@-prefixed domain", "@example.com", "example.com"},
		{"bare domain", "example.com", "example.com"},
		{"multiple @", "a@b@c.com", "b@c.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ExtractDomain(tt.input))
		})
	}
}
