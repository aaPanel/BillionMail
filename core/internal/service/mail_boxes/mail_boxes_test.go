package mail_boxes

import (
	"context"
	"sync"
	"testing"
	"time"

	v1 "billionmail-core/api/mail_boxes/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswdEncodeDecodeRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"simple", "hello123"},
		{"empty", ""},
		{"special chars", "p@ss!w0rd#$%^&*()"},
		{"unicode", "密码résumé"},
		{"spaces", "  spaces  "},
		{"long", "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := PasswdEncode(ctx, tt.password)
			decoded, err := PasswdDecode(ctx, encoded)
			require.NoError(t, err)
			assert.Equal(t, tt.password, decoded)
		})
	}
}

func TestPasswdDecodeInvalid(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name  string
		input string
	}{
		{"invalid hex", "xyz"},
		{"odd hex length", "abc"},
		{"valid hex but invalid base64", "deadbeef"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PasswdDecode(ctx, tt.input)
			assert.Error(t, err)
		})
	}
}

func TestGenerateRandomPassword(t *testing.T) {
	tests := []struct {
		name    string
		charset string
		length  int
	}{
		{"default charset", "", 16},
		{"digits only", "0123456789", 8},
		{"single char", "a", 5},
		{"length 1", "", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pw := generateRandomPassword(tt.charset, tt.length)
			assert.Len(t, pw, tt.length)

			if tt.charset != "" {
				for _, c := range pw {
					assert.Contains(t, tt.charset, string(c))
				}
			}
		})
	}
}

func TestShouldSkipAlert(t *testing.T) {
	// Reset cache
	quotaAlertCache = sync.Map{}

	a := quotaAlertTarget{Kind: "mailbox", Target: "user@example.com", Threshold: 90}

	// No cache entry → should not skip
	assert.False(t, shouldSkipAlert(a))

	// Store recent timestamp → should skip
	key := "mailbox|user@example.com|90"
	quotaAlertCache.Store(key, time.Now())
	assert.True(t, shouldSkipAlert(a))

	// Store old timestamp (>24h) → should not skip
	quotaAlertCache.Store(key, time.Now().Add(-25*time.Hour))
	assert.False(t, shouldSkipAlert(a))

	// Different threshold → should not skip
	b := quotaAlertTarget{Kind: "mailbox", Target: "user@example.com", Threshold: 95}
	assert.False(t, shouldSkipAlert(b))

	// Clean up
	quotaAlertCache = sync.Map{}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"zero bytes", 0, "0B"},
		{"bytes", 512, "512B"},
		{"exactly 1KB", 1024, "1.00KB"},
		{"KB range", 1536, "1.50KB"},
		{"exactly 1MB", 1024 * 1024, "1.00MB"},
		{"MB range", 1536 * 1024, "1.50MB"},
		{"exactly 1GB", 1024 * 1024 * 1024, "1.00GB"},
		{"GB range", 1536 * 1024 * 1024, "1.50GB"},
		{"1023 bytes", 1023, "1023B"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatSize(tt.input))
		})
	}
}

// AddImport must refuse a record without a username before touching the database:
// `mailbox.username` is the primary key and an empty string is a legal key value,
// so such a record used to be stored and could then not be deleted from the
// Mailboxes page.
func TestAddImportRejectsMailboxWithoutUsername(t *testing.T) {
	tests := []struct {
		name    string
		mailbox *v1.Mailbox
	}{
		{"empty username", &v1.Mailbox{Password: "pw"}},
		{"blank username", &v1.Mailbox{Username: "   ", Password: "pw", LocalPart: "foo", Domain: "example.com"}},
		{"username with surrounding spaces only", &v1.Mailbox{Username: "\t\n", Password: "pw"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AddImport(context.Background(), tt.mailbox)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "username is empty")
		})
	}
}
