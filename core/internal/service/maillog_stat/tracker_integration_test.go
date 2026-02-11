package maillog_stat

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Full pipeline: NewMailTracker -> TrackLinks -> AppendTrackingPixel -> GetHTML
// Tests the complete integration of tracker operations.
// ---------------------------------------------------------------------------

func TestTrackerPipeline_FullHTMLEmail(t *testing.T) {
	html := `<html>
<head><title>Newsletter</title></head>
<body>
<h1>Welcome!</h1>
<p>Visit <a href="https://example.com/offer">our offer</a></p>
<p>Read <a href="https://blog.example.com/post/1">our blog</a></p>
<p>Contact <a href="mailto:info@example.com">us</a></p>
<p><a href="#footer">Jump to footer</a></p>
</body>
</html>`

	tracker := NewMailTracker(html, 99, "<pipeline-msg@test.com>", "reader@inbox.com", "https://track.mydomain.com")

	// Step 1: Track links
	tracker.TrackLinks()
	afterLinks := tracker.GetHTML()

	// HTTP links should be tracked
	assert.NotContains(t, afterLinks, `href="https://example.com/offer"`)
	assert.NotContains(t, afterLinks, `href="https://blog.example.com/post/1"`)

	// Non-trackable links should remain
	assert.Contains(t, afterLinks, `href="mailto:info@example.com"`)
	assert.Contains(t, afterLinks, `href="#footer"`)

	// Tracking prefix should appear twice (for the 2 HTTP links)
	assert.Equal(t, 2, strings.Count(afterLinks, "track.mydomain.com/pmta/"))

	// Step 2: Append tracking pixel
	tracker.AppendTrackingPixel()
	final := tracker.GetHTML()

	// Pixel should be present
	assert.Contains(t, final, `<img src="`)
	assert.Contains(t, final, `style="display:none"`)

	// Pixel is before </body>
	pixelIdx := strings.Index(final, `<img src="`)
	bodyIdx := strings.Index(final, "</body>")
	assert.Less(t, pixelIdx, bodyIdx)

	// Original should be unchanged
	assert.Equal(t, html, tracker.GetOriginalHTML())
	assert.True(t, tracker.IsModified())
}

// ---------------------------------------------------------------------------
// Pipeline with decrypt verification: end-to-end data integrity
// ---------------------------------------------------------------------------

func TestTrackerPipeline_EncryptDecryptRoundtrip(t *testing.T) {
	html := `<html><body><a href="https://dest.example.com/page?utm=abc">Click</a></body></html>`
	tracker := NewMailTracker(html, 777, "<roundtrip-msg>", "user@dest.com", "https://t.example.com")

	tracker.TrackLinks()
	tracker.AppendTrackingPixel()

	result := tracker.GetHTML()

	// Extract all encrypted portions from tracking URLs
	prefix := "https://t.example.com/pmta/"
	idx := 0
	encryptedParts := make([]string, 0)
	for {
		pos := strings.Index(result[idx:], prefix)
		if pos < 0 {
			break
		}
		start := idx + pos + len(prefix)
		// Find the end of the encrypted portion (either " or space)
		end := strings.IndexAny(result[start:], `" `)
		if end < 0 {
			break
		}
		encryptedParts = append(encryptedParts, result[start:start+end])
		idx = start + end
	}

	require.GreaterOrEqual(t, len(encryptedParts), 2, "should have at least click URL + pixel URL")

	// Verify each encrypted part decrypts correctly
	for i, enc := range encryptedParts {
		var data struct {
			Type       string `json:"type"`
			CampaignID int    `json:"campaign_id"`
			Recipient  string `json:"recipient"`
			MessageID  string `json:"message_id"`
			URL        string `json:"url"`
		}
		err := Decrypt(enc, &data)
		require.NoError(t, err, "decrypt failed for part %d", i)

		assert.Equal(t, 777, data.CampaignID)
		assert.Equal(t, "user@dest.com", data.Recipient)
		assert.Equal(t, "roundtrip-msg", data.MessageID)

		switch data.Type {
		case "click":
			assert.Equal(t, "https://dest.example.com/page?utm=abc", data.URL)
		case "open":
			assert.Empty(t, data.URL)
		default:
			t.Errorf("unexpected type: %s", data.Type)
		}
	}
}

// ---------------------------------------------------------------------------
// Multiple trackers on same HTML: independence
// ---------------------------------------------------------------------------

func TestMultipleTrackersIndependence(t *testing.T) {
	html := `<html><body><a href="https://example.com">Link</a></body></html>`

	tracker1 := NewMailTracker(html, 1, "msg-1", "user1@test.com", "https://t1.example.com")
	tracker2 := NewMailTracker(html, 2, "msg-2", "user2@test.com", "https://t2.example.com")

	tracker1.TrackLinks()
	tracker2.TrackLinks()

	result1 := tracker1.GetHTML()
	result2 := tracker2.GetHTML()

	// Each tracker should use its own baseURL
	assert.Contains(t, result1, "t1.example.com/pmta/")
	assert.NotContains(t, result1, "t2.example.com/pmta/")

	assert.Contains(t, result2, "t2.example.com/pmta/")
	assert.NotContains(t, result2, "t1.example.com/pmta/")

	// Decrypt and verify different campaign IDs
	prefix1 := "https://t1.example.com/pmta/"
	pos1 := strings.Index(result1, prefix1)
	enc1End := strings.Index(result1[pos1+len(prefix1):], `"`)
	enc1 := result1[pos1+len(prefix1) : pos1+len(prefix1)+enc1End]

	prefix2 := "https://t2.example.com/pmta/"
	pos2 := strings.Index(result2, prefix2)
	enc2End := strings.Index(result2[pos2+len(prefix2):], `"`)
	enc2 := result2[pos2+len(prefix2) : pos2+len(prefix2)+enc2End]

	var data1, data2 struct {
		CampaignID int    `json:"campaign_id"`
		Recipient  string `json:"recipient"`
	}
	require.NoError(t, Decrypt(enc1, &data1))
	require.NoError(t, Decrypt(enc2, &data2))

	assert.Equal(t, 1, data1.CampaignID)
	assert.Equal(t, "user1@test.com", data1.Recipient)
	assert.Equal(t, 2, data2.CampaignID)
	assert.Equal(t, "user2@test.com", data2.Recipient)
}

// ---------------------------------------------------------------------------
// Concurrent tracker creation and processing
// ---------------------------------------------------------------------------

func TestTrackerConcurrentProcessing(t *testing.T) {
	html := `<html><body><a href="https://example.com">Link</a></body></html>`

	var wg sync.WaitGroup
	results := make([]string, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			tracker := NewMailTracker(html, idx, "<msg>", "u@x.com", "https://t.example.com")
			tracker.TrackLinks()
			tracker.AppendTrackingPixel()
			results[idx] = tracker.GetHTML()
		}(i)
	}

	wg.Wait()

	for i, r := range results {
		assert.NotEmpty(t, r, "result %d should not be empty", i)
		assert.Contains(t, r, "t.example.com/pmta/")
		assert.Contains(t, r, `<img src="`)
	}
}

// ---------------------------------------------------------------------------
// Pipeline order: pixel then links vs links then pixel
// ---------------------------------------------------------------------------

func TestTrackerPipeline_OrderMatters(t *testing.T) {
	html := `<html><body><a href="https://example.com">Link</a></body></html>`

	// Order 1: links first, then pixel
	t1 := NewMailTracker(html, 1, "m1", "u@x.com", "https://t.example.com")
	t1.TrackLinks()
	t1.AppendTrackingPixel()
	result1 := t1.GetHTML()

	// Order 2: pixel first, then links
	t2 := NewMailTracker(html, 1, "m1", "u@x.com", "https://t.example.com")
	t2.AppendTrackingPixel()
	t2.TrackLinks()
	result2 := t2.GetHTML()

	// Both should have tracking link and pixel
	assert.Contains(t, result1, "t.example.com/pmta/")
	assert.Contains(t, result1, `<img src="`)

	assert.Contains(t, result2, "t.example.com/pmta/")
	assert.Contains(t, result2, `<img src="`)

	// The pixel URL should NOT be tracked as a click link in order 2
	// because img src is not matched by href pattern
	clickCount1 := strings.Count(result1, "t.example.com/pmta/")
	clickCount2 := strings.Count(result2, "t.example.com/pmta/")

	// Order 1: 1 click link + 1 pixel = 2
	// Order 2: 1 pixel + 1 click link = 2 (pixel img src not matched by href regex)
	assert.Equal(t, clickCount1, clickCount2,
		"both orders should produce same number of tracking URLs")
}

// ---------------------------------------------------------------------------
// Large HTML email with many links
// ---------------------------------------------------------------------------

func TestTrackerPipeline_ManyLinks(t *testing.T) {
	var builder strings.Builder
	builder.WriteString("<html><body>")
	numLinks := 50
	for i := 0; i < numLinks; i++ {
		builder.WriteString(`<a href="https://example.com/page/`)
		builder.WriteString(strings.Repeat("x", i+1))
		builder.WriteString(`">Link</a>`)
	}
	builder.WriteString("</body></html>")

	html := builder.String()
	tracker := NewMailTracker(html, 42, "msg", "u@x.com", "https://t.example.com")

	tracker.TrackLinks()
	tracker.AppendTrackingPixel()

	result := tracker.GetHTML()

	// All links should be tracked
	trackCount := strings.Count(result, "t.example.com/pmta/")
	// numLinks click links + 1 pixel = numLinks+1
	assert.Equal(t, numLinks+1, trackCount)
}

// ---------------------------------------------------------------------------
// Empty HTML edge case
// ---------------------------------------------------------------------------

func TestTrackerPipeline_EmptyHTML(t *testing.T) {
	tracker := NewMailTracker("", 1, "m", "u@x.com", "https://t.example.com")

	tracker.TrackLinks()
	assert.False(t, tracker.IsModified())

	tracker.AppendTrackingPixel()
	assert.True(t, tracker.IsModified())

	result := tracker.GetHTML()
	assert.Contains(t, result, `<img src="`)
}

// ---------------------------------------------------------------------------
// API template ID offset pattern (used in api_mail_send.go)
// ---------------------------------------------------------------------------

func TestTrackerWithApiTemplateIdOffset(t *testing.T) {
	// API template IDs are offset by 1 billion to avoid conflict with campaign IDs
	apiTemplateId := 5 + 1000000000
	tracker := NewMailTracker(
		`<html><body><a href="https://example.com">Link</a></body></html>`,
		apiTemplateId,
		"<api-msg>",
		"user@example.com",
		"https://track.example.com",
	)

	tracker.TrackLinks()
	tracker.AppendTrackingPixel()

	// Verify campaign ID is preserved through encrypt/decrypt
	result := tracker.GetHTML()
	prefix := "https://track.example.com/pmta/"
	pos := strings.Index(result, prefix)
	require.Greater(t, pos, -1)

	encEnd := strings.Index(result[pos+len(prefix):], `"`)
	enc := result[pos+len(prefix) : pos+len(prefix)+encEnd]

	var data struct {
		CampaignID int `json:"campaign_id"`
	}
	require.NoError(t, Decrypt(enc, &data))
	assert.Equal(t, apiTemplateId, data.CampaignID)
}

// ---------------------------------------------------------------------------
// GetTrackingURL and GetTrackingPixel produce different types
// ---------------------------------------------------------------------------

func TestTrackingURLvsPixel_DifferentTypes(t *testing.T) {
	tracker := NewMailTracker("", 10, "msg-id", "user@test.com", "https://t.example.com")

	clickURL := tracker.GetTrackingURL("https://dest.com")
	pixelURL := tracker.GetTrackingPixel()

	// Both should have the base prefix
	assert.True(t, strings.HasPrefix(clickURL, "https://t.example.com/pmta/"))
	assert.True(t, strings.HasPrefix(pixelURL, "https://t.example.com/pmta/"))

	// They should be different
	assert.NotEqual(t, clickURL, pixelURL)

	// Decrypt and verify types
	clickEnc := strings.TrimPrefix(clickURL, "https://t.example.com/pmta/")
	pixelEnc := strings.TrimPrefix(pixelURL, "https://t.example.com/pmta/")

	var clickData, pixelData struct {
		Type string `json:"type"`
	}
	require.NoError(t, Decrypt(clickEnc, &clickData))
	require.NoError(t, Decrypt(pixelEnc, &pixelData))

	assert.Equal(t, "click", clickData.Type)
	assert.Equal(t, "open", pixelData.Type)
}
