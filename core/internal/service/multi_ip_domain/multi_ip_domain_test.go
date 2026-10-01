package multi_ip_domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidDomain(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid domain", "example.com", true},
		{"subdomain", "mail.example.com", true},
		{"hyphenated", "my-domain.com", true},
		{"numeric", "123.456", true},
		{"empty", "", false},
		{"special chars", "exam!ple.com", false},
		{"spaces", "example .com", false},
		{"underscore", "ex_ample.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isValidDomain(tt.input))
		})
	}
}

func TestIsValidIP(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid IPv4", "192.168.1.1", true},
		{"loopback", "127.0.0.1", true},
		{"all zeros", "0.0.0.0", true},
		{"max octets", "255.255.255.255", true},
		{"empty", "", false},
		{"too few octets", "192.168.1", false},
		{"too many octets", "192.168.1.1.1", false},
		{"letters", "abc.def.ghi.jkl", false},
		{"IPv6", "::1", false},
		{"four digit octet", "1234.1.1.1", false}, // regex limits to 1-3 digits per octet
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isValidIP(tt.input))
		})
	}
}
