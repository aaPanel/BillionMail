package public

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRound(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		precision int
		expected  float64
	}{
		{name: "zero", value: 0, precision: 2, expected: 0},
		{name: "integer value", value: 5.0, precision: 2, expected: 5.0},
		{name: "round down", value: 1.234, precision: 2, expected: 1.23},
		{name: "round up", value: 1.235, precision: 2, expected: 1.24},
		{name: "precision 0", value: 1.5, precision: 0, expected: 2},
		{name: "precision 1", value: 1.55, precision: 1, expected: 1.6},
		{name: "precision 3", value: 1.2345, precision: 3, expected: 1.235},
		{name: "negative value", value: -1.235, precision: 2, expected: -1.24},
		{name: "100 percent rate", value: 100.0, precision: 2, expected: 100.0},
		{name: "typical delivery rate", value: 98.765, precision: 2, expected: 98.77},
		{name: "typical bounce rate", value: 1.234, precision: 2, expected: 1.23},
		{name: "very small", value: 0.001, precision: 2, expected: 0},
		{name: "very small with precision 3", value: 0.001, precision: 3, expected: 0.001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Round(tt.value, tt.precision)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeUTF8(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "valid ASCII", input: "hello world", expected: "hello world"},
		{name: "valid UTF-8 Chinese", input: "hello world", expected: "hello world"},
		{name: "empty string", input: "", expected: ""},
		{name: "valid emoji", input: "hello ", expected: "hello "},
		{name: "valid multibyte", input: "cafe\u0301", expected: "cafe\u0301"},
		{name: "with invalid bytes", input: "hello\x80world", expected: "helloworld"},
		{name: "only invalid bytes", input: "\x80\x81\x82", expected: ""},
		{name: "mixed valid and invalid", input: "abc\xffdef", expected: "abcdef"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeUTF8(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeIPChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "valid IPv4", input: "192.168.1.1", expected: "192.168.1.1"},
		{name: "valid IPv6 chars", input: "::1", expected: "::1"},
		{name: "with letters", input: "abc123", expected: "abc123"},
		{name: "remove uppercase", input: "ABC", expected: ""},
		{name: "remove special chars", input: "192.168.1.1!", expected: "192.168.1.1"},
		{name: "empty string", input: "", expected: ""},
		{name: "only special chars", input: "!@#$%", expected: ""},
		{name: "mixed with spaces", input: "192. 168.1.1", expected: "192.168.1.1"},
		{name: "colon for IPv6", input: "fe80::1", expected: "fe80::1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeIPChars(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAddUnsubscribeButton(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		containsUnsub   bool
		containsBody    bool
		containsHtml    bool
		shouldAddButton bool
	}{
		{
			name:            "already has unsubscribe link",
			content:         `<body>Hello {{ UnsubscribeURL . }}</body>`,
			containsUnsub:   true,
			shouldAddButton: false,
		},
		{
			name:            "has body tag",
			content:         `<body>Hello World</body>`,
			containsBody:    true,
			shouldAddButton: true,
		},
		{
			name:            "has html tag no body",
			content:         `<html>Hello World</html>`,
			containsHtml:    true,
			shouldAddButton: true,
		},
		{
			name:            "plain text no tags",
			content:         `Hello World`,
			shouldAddButton: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddUnsubscribeButton(tt.content)

			if tt.containsUnsub {
				// Should return as-is
				assert.Equal(t, tt.content, result)
			} else {
				assert.Contains(t, result, "{{ UnsubscribeURL . }}")
				assert.Contains(t, result, "Unsubscribe")
			}
		})
	}
}

func TestAddUnsubscribeButton_PlacesBeforeClosingBody(t *testing.T) {
	content := `<!DOCTYPE html><html><body><p>Hello</p></body></html>`
	result := AddUnsubscribeButton(content)

	// The unsubscribe button should be placed before </body>
	assert.Contains(t, result, "Unsubscribe</a></div></body>")
}

func TestAddUnsubscribeButton_PlacesBeforeClosingHtml(t *testing.T) {
	content := `<html><p>Hello</p></html>`
	result := AddUnsubscribeButton(content)

	// Should be placed before </html> when no </body> tag exists
	assert.Contains(t, result, "Unsubscribe</a></div></html>")
}

func TestAddUnsubscribeButton_WrapsPlainText(t *testing.T) {
	content := `Just some plain text`
	result := AddUnsubscribeButton(content)

	assert.Contains(t, result, "<!DOCTYPE html>")
	assert.Contains(t, result, "<body>")
	assert.Contains(t, result, "Just some plain text")
	assert.Contains(t, result, "Unsubscribe")
}
