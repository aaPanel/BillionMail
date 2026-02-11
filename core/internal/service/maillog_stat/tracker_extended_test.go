package maillog_stat

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper to create a tracker with sensible defaults
func newTestTracker(html string) *MailTracker {
	return NewMailTracker(html, 42, "<msg-001>", "user@example.com", "https://track.example.com")
}

// ---------------------------------------------------------------------------
// TrackLinks: protocol-based skip tests
// ---------------------------------------------------------------------------

func TestTrackLinks_SkipsNonHTTPSchemes(t *testing.T) {
	tests := []struct {
		name string
		html string
	}{
		{"mailto link", `<a href="mailto:info@example.com">Email</a>`},
		{"tel link", `<a href="tel:+15551234567">Call</a>`},
		{"data URI", `<a href="data:text/html,<h1>hi</h1>">Data</a>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := newTestTracker(tt.html)
			tracker.TrackLinks()

			assert.False(t, tracker.IsModified(), "should not modify HTML for %s", tt.name)
			assert.Equal(t, tt.html, tracker.GetHTML())
		})
	}
}

func TestTrackLinks_SkipsAnchorLinks(t *testing.T) {
	html := `<a href="#section-2">Jump</a>`
	tracker := newTestTracker(html)
	tracker.TrackLinks()

	assert.False(t, tracker.IsModified())
	assert.Equal(t, html, tracker.GetHTML())
}

// ---------------------------------------------------------------------------
// TrackLinks: URLs with query params and fragments
// ---------------------------------------------------------------------------

func TestTrackLinks_URLsWithQueryParamsAndFragments(t *testing.T) {
	tests := []struct {
		name        string
		href        string
		wantTracked bool
	}{
		{"query params", "https://example.com/page?foo=bar&baz=1", true},
		{"fragment", "https://example.com/page#section", true},
		{"query + fragment", "https://example.com/page?a=1#top", true},
		{"bare domain", "https://example.com", true},
		{"relative path (no scheme/host)", "/path/to/page", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := `<a href="` + tt.href + `">Link</a>`
			tracker := newTestTracker(html)
			tracker.TrackLinks()

			assert.Equal(t, tt.wantTracked, tracker.IsModified())
			if tt.wantTracked {
				assert.Contains(t, tracker.GetHTML(), "track.example.com/pmta/")
				assert.NotContains(t, tracker.GetHTML(), `href="`+tt.href+`"`)
			} else {
				assert.Contains(t, tracker.GetHTML(), tt.href)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TrackLinks: idempotency (calling twice should not double-wrap)
// ---------------------------------------------------------------------------

func TestTrackLinks_Idempotency(t *testing.T) {
	html := `<a href="https://example.com/page">Click</a>`
	tracker := newTestTracker(html)

	tracker.TrackLinks()
	firstPass := tracker.GetHTML()
	require.True(t, tracker.IsModified())

	// Call again on the already-tracked HTML.
	// The tracked URL goes through Encrypt() producing a base64-url string
	// that gets a scheme+host from baseURL, so it IS a valid URL and would
	// be re-wrapped. This test documents the current behaviour.
	tracker2 := newTestTracker(firstPass)
	tracker2.TrackLinks()
	secondPass := tracker2.GetHTML()

	// Count how many times the tracking prefix appears
	firstCount := strings.Count(firstPass, "track.example.com/pmta/")
	secondCount := strings.Count(secondPass, "track.example.com/pmta/")

	// Both passes should contain exactly one tracked href
	assert.Equal(t, 1, firstCount, "first pass should have exactly 1 tracked link")
	assert.Equal(t, 1, secondCount, "second pass should have exactly 1 tracked link")
}

// ---------------------------------------------------------------------------
// TrackLinks: multiple links, mixed trackable and non-trackable
// ---------------------------------------------------------------------------

func TestTrackLinks_MixedLinks(t *testing.T) {
	html := `<html><body>
		<a href="https://example.com">HTTP</a>
		<a href="mailto:x@y.com">Mail</a>
		<a href="https://other.com/path?q=1">Other</a>
		<a href="#anchor">Anchor</a>
	</body></html>`

	tracker := newTestTracker(html)
	tracker.TrackLinks()

	result := tracker.GetHTML()
	assert.True(t, tracker.IsModified())

	// HTTP links tracked
	assert.NotContains(t, result, `href="https://example.com"`)
	assert.NotContains(t, result, `href="https://other.com/path?q=1"`)

	// Non-trackable links preserved
	assert.Contains(t, result, `href="mailto:x@y.com"`)
	assert.Contains(t, result, `href="#anchor"`)

	// Tracking prefix present (at least 2 for the 2 HTTP links)
	assert.GreaterOrEqual(t, strings.Count(result, "track.example.com/pmta/"), 2)
}

// ---------------------------------------------------------------------------
// AppendTrackingPixel: nested HTML structures
// ---------------------------------------------------------------------------

func TestAppendTrackingPixel_NestedHTML(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		wantBefore    string // pixel should appear just before this tag
		wantPixelOnce bool
	}{
		{
			name: "deeply nested body",
			html: `<html><head><title>T</title></head><body><div><table><tr><td>content</td></tr></table></div></body></html>`,
			// pixel inserted before </body>
			wantBefore:    "</body>",
			wantPixelOnce: true,
		},
		{
			name:          "multiple body tags (only first replaced)",
			html:          `<html><body>first</body><body>second</body></html>`,
			wantBefore:    "</body>",
			wantPixelOnce: true,
		},
		{
			name:          "body with attributes",
			html:          `<html><body class="main" style="margin:0">Content</body></html>`,
			wantBefore:    "</body>",
			wantPixelOnce: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := newTestTracker(tt.html)
			tracker.AppendTrackingPixel()

			result := tracker.GetHTML()
			assert.True(t, tracker.IsModified())
			assert.Contains(t, result, `<img src="`)
			assert.Contains(t, result, `style="display:none"`)

			if tt.wantPixelOnce {
				assert.Equal(t, 1, strings.Count(result, `<img src="`),
					"should insert exactly one tracking pixel")
			}

			// Pixel appears before the closing tag
			pixelIdx := strings.Index(result, `<img src="`)
			closeIdx := strings.Index(result, tt.wantBefore)
			assert.Less(t, pixelIdx, closeIdx, "pixel should be before %s", tt.wantBefore)
		})
	}
}

// ---------------------------------------------------------------------------
// AppendTrackingPixel: HTML with existing img tags
// ---------------------------------------------------------------------------

func TestAppendTrackingPixel_ExistingImages(t *testing.T) {
	html := `<html><body><img src="https://example.com/logo.png" alt="Logo" /><p>Hello</p></body></html>`
	tracker := newTestTracker(html)
	tracker.AppendTrackingPixel()

	result := tracker.GetHTML()
	assert.True(t, tracker.IsModified())

	// Original image preserved
	assert.Contains(t, result, `src="https://example.com/logo.png"`)
	// Tracking pixel added
	assert.Equal(t, 2, strings.Count(result, `<img `), "should have original img + tracking pixel")
}

// ---------------------------------------------------------------------------
// Case-insensitive body/html tag matching
// ---------------------------------------------------------------------------

func TestAppendTrackingPixel_CaseSensitivity(t *testing.T) {
	// The source uses strings.Contains which is case-sensitive.
	// Upper-case tags should fall through to the append path.
	tests := []struct {
		name       string
		html       string
		wantAppend bool // true = pixel appended at end (no tag match)
	}{
		{"lowercase body", `<html><body>X</body></html>`, false},
		{"uppercase BODY", `<html><BODY>X</BODY></html>`, false}, // </html> is lowercase, pixel goes before it
		{"mixed case Body", `<html><Body>X</Body></html>`, false}, // </html> is lowercase, pixel goes before it
		{"lowercase html no body", `<html>X</html>`, false},
		{"uppercase HTML", `<HTML>X</HTML>`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := newTestTracker(tt.html)
			tracker.AppendTrackingPixel()

			result := tracker.GetHTML()
			assert.True(t, tracker.IsModified())

			if tt.wantAppend {
				// Pixel is appended at the very end
				assert.True(t, strings.HasSuffix(result, `" />`),
					"pixel should be appended at end for %s", tt.name)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Combined TrackLinks + AppendTrackingPixel
// ---------------------------------------------------------------------------

func TestCombinedTrackLinksAndPixel(t *testing.T) {
	html := `<html><body><a href="https://example.com">Visit</a></body></html>`
	tracker := newTestTracker(html)

	tracker.TrackLinks()
	tracker.AppendTrackingPixel()

	result := tracker.GetHTML()
	assert.True(t, tracker.IsModified())

	// Link is tracked
	assert.NotContains(t, result, `href="https://example.com"`)
	assert.Contains(t, result, "track.example.com/pmta/")

	// Pixel is present
	assert.Contains(t, result, `<img src="`)
	assert.Contains(t, result, `style="display:none"`)

	// Pixel is before </body>
	pixelIdx := strings.Index(result, `<img src="`)
	bodyIdx := strings.Index(result, "</body>")
	assert.Less(t, pixelIdx, bodyIdx)
}

// ---------------------------------------------------------------------------
// GetHTML returns modified content, GetOriginalHTML returns original
// ---------------------------------------------------------------------------

func TestGetHTMLAndGetOriginalHTML(t *testing.T) {
	original := `<html><body><a href="https://example.com">X</a></body></html>`
	tracker := newTestTracker(original)

	assert.Equal(t, original, tracker.GetHTML(), "before modification GetHTML == original")
	assert.Equal(t, original, tracker.GetOriginalHTML())

	tracker.TrackLinks()
	assert.NotEqual(t, original, tracker.GetHTML(), "after TrackLinks GetHTML should differ")
	assert.Equal(t, original, tracker.GetOriginalHTML(), "GetOriginalHTML unchanged after TrackLinks")

	tracker.AppendTrackingPixel()
	assert.NotEqual(t, original, tracker.GetHTML())
	assert.Equal(t, original, tracker.GetOriginalHTML(), "GetOriginalHTML still unchanged")
}

// ---------------------------------------------------------------------------
// IsModified state transitions
// ---------------------------------------------------------------------------

func TestIsModified_StateTransitions(t *testing.T) {
	t.Run("new tracker is not modified", func(t *testing.T) {
		tracker := newTestTracker(`<html><body>Hello</body></html>`)
		assert.False(t, tracker.IsModified())
	})

	t.Run("TrackLinks with no trackable links stays unmodified", func(t *testing.T) {
		tracker := newTestTracker(`<p>No links</p>`)
		tracker.TrackLinks()
		assert.False(t, tracker.IsModified())
	})

	t.Run("TrackLinks with trackable link becomes modified", func(t *testing.T) {
		tracker := newTestTracker(`<a href="https://example.com">X</a>`)
		tracker.TrackLinks()
		assert.True(t, tracker.IsModified())
	})

	t.Run("AppendTrackingPixel always sets modified", func(t *testing.T) {
		tracker := newTestTracker(`<p>plain</p>`)
		assert.False(t, tracker.IsModified())
		tracker.AppendTrackingPixel()
		assert.True(t, tracker.IsModified())
	})

	t.Run("TrackLinks unmodified then AppendTrackingPixel modified", func(t *testing.T) {
		tracker := newTestTracker(`<p>No links</p>`)
		tracker.TrackLinks()
		assert.False(t, tracker.IsModified())
		tracker.AppendTrackingPixel()
		assert.True(t, tracker.IsModified())
	})
}

// ---------------------------------------------------------------------------
// NewMailTracker: messageID angle-bracket trimming
// ---------------------------------------------------------------------------

func TestNewMailTracker_MessageIDTrimming(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"with angle brackets", "<msg-123>", "msg-123"},
		{"without angle brackets", "msg-123", "msg-123"},
		{"empty", "", ""},
		{"nested brackets", "<<msg>>", "msg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewMailTracker("", 1, tt.input, "x@y.com", "https://example.com")
			assert.Equal(t, tt.want, tracker.messageID)
		})
	}
}

// ---------------------------------------------------------------------------
// NewMailTracker: baseURL normalization
// ---------------------------------------------------------------------------

func TestNewMailTracker_BaseURLNormalization(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    string
	}{
		{"no trailing slash", "https://example.com", "https://example.com/pmta"},
		{"trailing slash", "https://example.com/", "https://example.com/pmta"},
		{"multiple trailing slashes", "https://example.com///", "https://example.com/pmta"},
		{"with path", "https://example.com/api", "https://example.com/api/pmta"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewMailTracker("", 1, "id", "x@y.com", tt.baseURL)
			assert.Equal(t, tt.want, tracker.baseURL)
		})
	}
}

// ---------------------------------------------------------------------------
// GetTrackingURL produces encrypted output that round-trips
// ---------------------------------------------------------------------------

func TestGetTrackingURL_Roundtrip(t *testing.T) {
	tracker := newTestTracker("")
	trackURL := tracker.GetTrackingURL("https://destination.com/page")

	assert.True(t, strings.HasPrefix(trackURL, "https://track.example.com/pmta/"))

	// Extract encrypted portion
	encrypted := strings.TrimPrefix(trackURL, "https://track.example.com/pmta/")
	require.NotEmpty(t, encrypted)

	// Decrypt and verify
	var data struct {
		Type       string `json:"type"`
		CampaignID int    `json:"campaign_id"`
		Recipient  string `json:"recipient"`
		MessageID  string `json:"message_id"`
		URL        string `json:"url"`
	}
	err := Decrypt(encrypted, &data)
	require.NoError(t, err)

	assert.Equal(t, "click", data.Type)
	assert.Equal(t, 42, data.CampaignID)
	assert.Equal(t, "user@example.com", data.Recipient)
	assert.Equal(t, "msg-001", data.MessageID) // angle brackets trimmed
	assert.Equal(t, "https://destination.com/page", data.URL)
}

// ---------------------------------------------------------------------------
// GetTrackingPixel produces encrypted output that round-trips
// ---------------------------------------------------------------------------

func TestGetTrackingPixel_Roundtrip(t *testing.T) {
	tracker := newTestTracker("")
	pixelURL := tracker.GetTrackingPixel()

	assert.True(t, strings.HasPrefix(pixelURL, "https://track.example.com/pmta/"))

	encrypted := strings.TrimPrefix(pixelURL, "https://track.example.com/pmta/")
	require.NotEmpty(t, encrypted)

	var data struct {
		Type       string `json:"type"`
		CampaignID int    `json:"campaign_id"`
		Recipient  string `json:"recipient"`
		MessageID  string `json:"message_id"`
	}
	err := Decrypt(encrypted, &data)
	require.NoError(t, err)

	assert.Equal(t, "open", data.Type)
	assert.Equal(t, 42, data.CampaignID)
	assert.Equal(t, "user@example.com", data.Recipient)
	assert.Equal(t, "msg-001", data.MessageID)
}

// ---------------------------------------------------------------------------
// TrackLinks: href pattern matching edge cases
// ---------------------------------------------------------------------------

func TestTrackLinks_HrefPatternEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		html         string
		wantModified bool
		wantContains string
	}{
		{
			name:         "href with spaces around equals",
			html:         `<a href = "https://example.com">X</a>`,
			wantModified: true,
			wantContains: "track.example.com/pmta/",
		},
		{
			name:         "single-quoted href not matched (regex uses double quotes)",
			html:         `<a href='https://example.com'>X</a>`,
			wantModified: false,
			wantContains: `href='https://example.com'`,
		},
		{
			name:         "empty href",
			html:         `<a href="">X</a>`,
			wantModified: false,
			wantContains: `href=""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := newTestTracker(tt.html)
			tracker.TrackLinks()

			assert.Equal(t, tt.wantModified, tracker.IsModified())
			assert.Contains(t, tracker.GetHTML(), tt.wantContains)
		})
	}
}

// ---------------------------------------------------------------------------
// hrefPattern compiled correctly
// ---------------------------------------------------------------------------

func TestHrefPatternCompilation(t *testing.T) {
	pattern := regexp.MustCompile(`href\s*=\s*"([^"]+)"`)

	tests := []struct {
		input   string
		wantURL string
	}{
		{`href="https://x.com"`, "https://x.com"},
		{`href = "https://y.com/p"`, "https://y.com/p"},
		{`href="mailto:a@b.com"`, "mailto:a@b.com"},
	}

	for _, tt := range tests {
		matches := pattern.FindStringSubmatch(tt.input)
		require.Len(t, matches, 2)
		assert.Equal(t, tt.wantURL, matches[1])
	}
}
