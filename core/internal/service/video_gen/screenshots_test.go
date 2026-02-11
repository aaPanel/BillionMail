package video_gen

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultScreenshotConfig(t *testing.T) {
	cfg := DefaultScreenshotConfig("https://example.com", "Example Dental", "/tmp/out")

	assert.Equal(t, "https://example.com", cfg.WebsiteURL)
	assert.Equal(t, "Example Dental", cfg.BusinessName)
	assert.Equal(t, "/tmp/out", cfg.OutputDir)
	assert.Equal(t, 1920, cfg.Width)
	assert.Equal(t, 1080, cfg.Height)
}

func TestScreenshotPaths(t *testing.T) {
	h, c, g := ScreenshotPaths("/tmp/prospect_123")

	assert.Equal(t, "/tmp/prospect_123/homepage.png", h)
	assert.Equal(t, "/tmp/prospect_123/contact.png", c)
	assert.Equal(t, "/tmp/prospect_123/google.png", g)
}

func TestBuildPlaywrightScript(t *testing.T) {
	cfg := DefaultScreenshotConfig("https://brighthorizon.com", "Bright Horizon Dental", "/tmp/out")
	result := &ScreenshotResult{
		Homepage: "/tmp/out/homepage.png",
		Contact:  "/tmp/out/contact.png",
		Google:   "/tmp/out/google.png",
	}

	script := buildPlaywrightScript(cfg, result)

	// Verify script contains key elements
	assert.Contains(t, script, "chromium.launch")
	assert.Contains(t, script, "width: 1920")
	assert.Contains(t, script, "height: 1080")
	assert.Contains(t, script, "https://brighthorizon.com")
	assert.Contains(t, script, "brighthorizon.com/contact")
	assert.Contains(t, script, "google.com/maps/search/Bright+Horizon+Dental")
	assert.Contains(t, script, "Promise.all")
	assert.Contains(t, script, "browser.close")
}

func TestBuildPlaywrightScript_WrapsURLsInQuotes(t *testing.T) {
	cfg := DefaultScreenshotConfig("https://example.com/path?q=1&b=2", "Test Business", "/tmp/out")
	result := &ScreenshotResult{
		Homepage: "/tmp/out/homepage.png",
		Contact:  "/tmp/out/contact.png",
		Google:   "/tmp/out/google.png",
	}

	script := buildPlaywrightScript(cfg, result)

	// %q wraps URLs in double quotes, preserving special chars safely
	assert.Contains(t, script, `"https://example.com/path?q=1&b=2"`)
	assert.Contains(t, script, "chromium.launch")
}

func TestFindContactURL(t *testing.T) {
	tests := []struct {
		base     string
		expected string
	}{
		{"https://example.com", "https://example.com/contact"},
		{"https://example.com/", "https://example.com/contact"},
		{"https://sub.example.com", "https://sub.example.com/contact"},
	}

	for _, tt := range tests {
		t.Run(tt.base, func(t *testing.T) {
			assert.Equal(t, tt.expected, findContactURL(tt.base))
		})
	}
}

func TestBuildGoogleMapsURL(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"Bright Horizon Dental", "https://www.google.com/maps/search/Bright+Horizon+Dental"},
		{"DrSmith", "https://www.google.com/maps/search/DrSmith"},
		{"A B C", "https://www.google.com/maps/search/A+B+C"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, buildGoogleMapsURL(tt.name))
		})
	}
}

// --- Annotation tests ---

func TestBuildAnnotateArgs_Circle(t *testing.T) {
	cfg := AnnotateConfig{
		InputPath:  "/tmp/in.png",
		OutputPath: "/tmp/out.png",
		Annotations: []Annotation{
			{Type: AnnotationCircle, X: 100, Y: 200, Radius: 50, Color: "red"},
		},
	}

	args := BuildAnnotateArgs(cfg)

	assert.Equal(t, "/tmp/in.png", args[0])
	assert.Equal(t, "/tmp/out.png", args[len(args)-1])
	joined := strings.Join(args, " ")
	assert.Contains(t, joined, "-stroke red")
	assert.Contains(t, joined, "circle 100,200 150,200")
}

func TestBuildAnnotateArgs_Text(t *testing.T) {
	cfg := AnnotateConfig{
		InputPath:  "/tmp/in.png",
		OutputPath: "/tmp/out.png",
		Annotations: []Annotation{
			{Type: AnnotationText, X: 50, Y: 900, Text: "No live chat", Color: "red"},
		},
	}

	args := BuildAnnotateArgs(cfg)

	joined := strings.Join(args, " ")
	assert.Contains(t, joined, "-pointsize 28")
	assert.Contains(t, joined, "+50+900")
	assert.Contains(t, joined, "No live chat")
}

func TestBuildAnnotateArgs_Arrow(t *testing.T) {
	cfg := AnnotateConfig{
		InputPath:  "/tmp/in.png",
		OutputPath: "/tmp/out.png",
		Annotations: []Annotation{
			{Type: AnnotationArrow, X: 100, Y: 100, X2: 300, Y2: 100, Color: "blue"},
		},
	}

	args := BuildAnnotateArgs(cfg)

	joined := strings.Join(args, " ")
	assert.Contains(t, joined, "line 100,100 300,100")
	assert.Contains(t, joined, "polygon") // arrowhead
}

func TestBuildAnnotateArgs_DefaultColor(t *testing.T) {
	cfg := AnnotateConfig{
		InputPath:  "/tmp/in.png",
		OutputPath: "/tmp/out.png",
		Annotations: []Annotation{
			{Type: AnnotationCircle, X: 10, Y: 10, Radius: 5},
		},
	}

	args := BuildAnnotateArgs(cfg)

	joined := strings.Join(args, " ")
	assert.Contains(t, joined, "-stroke red")
}

func TestBuildAnnotateArgs_NoAnnotations(t *testing.T) {
	cfg := AnnotateConfig{
		InputPath:  "/tmp/in.png",
		OutputPath: "/tmp/out.png",
	}

	args := BuildAnnotateArgs(cfg)

	// Just input + output, no draw commands
	assert.Equal(t, []string{"/tmp/in.png", "/tmp/out.png"}, args)
}

func TestBuildAnnotateArgs_MultipleAnnotations(t *testing.T) {
	cfg := AnnotateConfig{
		InputPath:  "/tmp/in.png",
		OutputPath: "/tmp/out.png",
		Annotations: []Annotation{
			{Type: AnnotationCircle, X: 100, Y: 200, Radius: 50, Color: "red"},
			{Type: AnnotationText, X: 50, Y: 900, Text: "Missing chat", Color: "orange"},
		},
	}

	args := BuildAnnotateArgs(cfg)

	joined := strings.Join(args, " ")
	assert.Contains(t, joined, "circle 100,200")
	assert.Contains(t, joined, "Missing chat")
}

func TestDefaultAnnotations_Contact_NoChat(t *testing.T) {
	signals := map[string]bool{"no_chat": true}
	anns := DefaultAnnotations(ScreenshotContact, signals)

	require.Len(t, anns, 2) // circle + text
	assert.Equal(t, AnnotationCircle, anns[0].Type)
	assert.Equal(t, AnnotationText, anns[1].Type)
	assert.Contains(t, anns[1].Text, "No live chat")
}

func TestDefaultAnnotations_Contact_NoBooking(t *testing.T) {
	signals := map[string]bool{"no_online_booking": true}
	anns := DefaultAnnotations(ScreenshotContact, signals)

	require.Len(t, anns, 1)
	assert.Contains(t, anns[0].Text, "No online self-booking")
}

func TestDefaultAnnotations_Contact_BothSignals(t *testing.T) {
	signals := map[string]bool{"no_chat": true, "no_online_booking": true}
	anns := DefaultAnnotations(ScreenshotContact, signals)

	assert.Len(t, anns, 3) // circle + text + booking text
}

func TestDefaultAnnotations_Homepage_Voicemail(t *testing.T) {
	signals := map[string]bool{"voicemail_after_hrs": true}
	anns := DefaultAnnotations(ScreenshotHomepage, signals)

	require.Len(t, anns, 1)
	assert.Contains(t, anns[0].Text, "62%")
}

func TestDefaultAnnotations_Google_HighSpendLowRating(t *testing.T) {
	signals := map[string]bool{"high_spend_low_rating": true}
	anns := DefaultAnnotations(ScreenshotGoogle, signals)

	require.Len(t, anns, 1)
	assert.Contains(t, anns[0].Text, "ad spend")
}

func TestDefaultAnnotations_NoSignals(t *testing.T) {
	for _, st := range []ScreenshotType{ScreenshotHomepage, ScreenshotContact, ScreenshotGoogle} {
		anns := DefaultAnnotations(st, map[string]bool{})
		assert.Empty(t, anns, "expected no annotations for %s with no signals", st)
	}
}

func TestNormalizeColor(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"red", "#FF0000"},
		{"Red", "#FF0000"},
		{"#FF6600", "#FF6600"},
		{"orange", "#FF6600"},
		{"blue", "#0066FF"},
		{"unknown", "#FF0000"}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeColor(tt.input))
		})
	}
}

func TestAnnotatedPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/tmp/out/homepage.png", "/tmp/out/homepage_annotated.png"},
		{"/tmp/out/contact.png", "/tmp/out/contact_annotated.png"},
		{"screenshot.jpg", "screenshot_annotated.jpg"},
	}

	for _, tt := range tests {
		t.Run(filepath.Base(tt.input), func(t *testing.T) {
			assert.Equal(t, tt.expected, annotatedPath(tt.input))
		})
	}
}
