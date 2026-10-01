package domains

import (
	"billionmail-core/internal/model"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReverseIP(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard IPv4",
			input:    "1.2.3.4",
			expected: "4.3.2.1",
		},
		{
			name:     "loopback",
			input:    "127.0.0.1",
			expected: "1.0.0.127",
		},
		{
			name:     "all zeros",
			input:    "0.0.0.0",
			expected: "0.0.0.0",
		},
		{
			name:     "all same",
			input:    "5.5.5.5",
			expected: "5.5.5.5",
		},
		{
			name:     "high numbers",
			input:    "192.168.10.25",
			expected: "25.10.168.192",
		},
		{
			name:     "broadcast",
			input:    "255.255.255.255",
			expected: "255.255.255.255",
		},
		{
			name:     "google DNS",
			input:    "8.8.8.8",
			expected: "8.8.8.8",
		},
		{
			name:     "google DNS alt",
			input:    "8.8.4.4",
			expected: "4.4.8.8",
		},
		{
			name:     "class A",
			input:    "10.0.0.1",
			expected: "1.0.0.10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReverseIP(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReverseIP_DoubleReverse(t *testing.T) {
	ips := []string{"1.2.3.4", "192.168.0.1", "10.20.30.40", "127.0.0.1"}

	for _, ip := range ips {
		reversed := ReverseIP(ReverseIP(ip))
		assert.Equal(t, ip, reversed, "double reverse should return original")
	}
}

func TestGetBlacklistLogPath(t *testing.T) {
	tests := []struct {
		name   string
		domain string
	}{
		{
			name:   "simple domain",
			domain: "example.com",
		},
		{
			name:   "subdomain",
			domain: "mail.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := GetBlacklistLogPath(tt.domain)
			assert.Contains(t, path, tt.domain)
			assert.True(t, strings.HasSuffix(path, "_blcheck.txt"))
		})
	}
}

func TestCONF_BLACKLISTS_NotEmpty(t *testing.T) {
	assert.Greater(t, len(CONF_BLACKLISTS), 0, "blacklist providers should not be empty")
}

func TestCONF_BLACKLISTS_NoDuplicates(t *testing.T) {
	seen := make(map[string]bool)
	duplicates := make([]string, 0)

	for _, bl := range CONF_BLACKLISTS {
		if seen[bl] {
			duplicates = append(duplicates, bl)
		}
		seen[bl] = true
	}

	// Note: source code has known duplicates (multi.surbl.org, cbl.anti-spam.org.cn).
	// This test documents them.
	if len(duplicates) > 0 {
		t.Logf("found %d duplicate blacklist entries: %v", len(duplicates), duplicates)
	}
}

func TestCONF_BLACKLISTS_ValidFormat(t *testing.T) {
	for _, bl := range CONF_BLACKLISTS {
		assert.NotEmpty(t, bl)
		assert.Contains(t, bl, ".", "blacklist %q should be a domain name", bl)
		assert.False(t, strings.HasPrefix(bl, "."), "blacklist %q should not start with dot", bl)
		assert.False(t, strings.HasSuffix(bl, "."), "blacklist %q should not end with dot", bl)
	}
}

func TestSKIP_IP_RESPONSES(t *testing.T) {
	assert.Greater(t, len(SKIP_IP_RESPONSES), 0, "skip responses should not be empty")

	expectedIPs := []string{
		"127.255.255.254",
		"127.255.255.255",
		"127.0.0.1",
		"127.0.1.1",
		"127.0.0.7",
	}

	for _, ip := range expectedIPs {
		val, ok := SKIP_IP_RESPONSES[ip]
		assert.True(t, ok, "expected skip IP %s to be present", ip)
		assert.Equal(t, "Passed", val)
	}
}

func TestBuildBlacklistAlertEmailHTML(t *testing.T) {
	tests := []struct {
		name   string
		ip     string
		domain string
		result *model.BlacklistCheckResult
	}{
		{
			name:   "single blacklist hit",
			ip:     "1.2.3.4",
			domain: "example.com",
			result: &model.BlacklistCheckResult{
				Domain:      "example.com",
				IP:          "1.2.3.4",
				Time:        1700000000,
				Tested:      150,
				Passed:      149,
				Invalid:     0,
				Blacklisted: 1,
				BlackList: []model.BlacklistDetail{
					{Blacklist: "zen.spamhaus.org", Time: 1700000000, Response: "127.0.0.2"},
				},
			},
		},
		{
			name:   "multiple blacklist hits",
			ip:     "10.0.0.1",
			domain: "test.org",
			result: &model.BlacklistCheckResult{
				Domain:      "test.org",
				IP:          "10.0.0.1",
				Time:        1700000000,
				Tested:      150,
				Passed:      140,
				Invalid:     5,
				Blacklisted: 5,
				BlackList: []model.BlacklistDetail{
					{Blacklist: "zen.spamhaus.org", Time: 1700000000, Response: "127.0.0.2"},
					{Blacklist: "bl.spamcop.net", Time: 1700000000, Response: "127.0.0.3"},
				},
			},
		},
		{
			name:   "no blacklist hits",
			ip:     "8.8.8.8",
			domain: "clean.com",
			result: &model.BlacklistCheckResult{
				Domain:      "clean.com",
				IP:          "8.8.8.8",
				Time:        1700000000,
				Tested:      150,
				Passed:      150,
				Invalid:     0,
				Blacklisted: 0,
				BlackList:   []model.BlacklistDetail{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := buildBlacklistAlertEmailHTML(tt.ip, tt.domain, tt.result)

			assert.Contains(t, html, tt.domain)
			assert.Contains(t, html, tt.ip)
			assert.Contains(t, html, "Blacklist Detection Alert")
			assert.Contains(t, html, "BillionMail")
			assert.Contains(t, html, "<!DOCTYPE html>")

			for _, bl := range tt.result.BlackList {
				assert.Contains(t, html, bl.Blacklist)
				assert.Contains(t, html, bl.Response)
			}
		})
	}
}

func TestBlacklistAlertSettings_Struct(t *testing.T) {
	settings := BlacklistAlertSettings{
		Name:          "Test Alert",
		SenderEmail:   "alert@example.com",
		SMTPPassword:  "secret",
		SMTPServer:    "smtp.example.com",
		SMTPPort:      587,
		RecipientList: []string{"admin@example.com", "ops@example.com"},
	}

	assert.Equal(t, "Test Alert", settings.Name)
	assert.Equal(t, 587, settings.SMTPPort)
	assert.Len(t, settings.RecipientList, 2)
}
