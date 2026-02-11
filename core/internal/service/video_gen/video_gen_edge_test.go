package video_gen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Screenshot edge cases ---

func TestDefaultScreenshotConfig_EmptyStrings(t *testing.T) {
	cfg := DefaultScreenshotConfig("", "", "")

	assert.Equal(t, "", cfg.WebsiteURL)
	assert.Equal(t, "", cfg.BusinessName)
	assert.Equal(t, "", cfg.OutputDir)
	assert.Equal(t, 1920, cfg.Width) // defaults still set
	assert.Equal(t, 1080, cfg.Height)
}

func TestScreenshotPaths_EmptyDir(t *testing.T) {
	h, c, g := ScreenshotPaths("")

	assert.Equal(t, "homepage.png", h)
	assert.Equal(t, "contact.png", c)
	assert.Equal(t, "google.png", g)
}

func TestFindContactURL_EmptyBase(t *testing.T) {
	result := findContactURL("")
	assert.Equal(t, "/contact", result)
}

func TestFindContactURL_ProtocolOnly(t *testing.T) {
	result := findContactURL("https://")
	// TrimRight removes trailing "/" so "https://" → "https:" → + "/contact"
	assert.Equal(t, "https:/contact", result) // degenerate but doesn't panic
}

func TestBuildGoogleMapsURL_EmptyName(t *testing.T) {
	result := buildGoogleMapsURL("")
	assert.Equal(t, "https://www.google.com/maps/search/", result)
}

func TestBuildGoogleMapsURL_SpecialChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"ampersand", "A & B Dental", "A+&+B+Dental"},
		{"apostrophe", "Dr. Smith's", "Dr.+Smith's"},
		{"unicode", "Clínica São Paulo", "Clínica+São+Paulo"},
		{"consecutive spaces", "A  B", "A++B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildGoogleMapsURL(tt.input)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestBuildPlaywrightScript_EmptyConfig(t *testing.T) {
	cfg := ScreenshotConfig{}
	result := &ScreenshotResult{}

	script := buildPlaywrightScript(cfg, result)

	// Should still produce valid JS structure
	assert.Contains(t, script, "chromium.launch")
	assert.Contains(t, script, "browser.close")
	assert.Contains(t, script, "width: 0") // zero viewport
}

func TestBuildPlaywrightScript_CustomViewport(t *testing.T) {
	cfg := ScreenshotConfig{
		WebsiteURL:   "https://example.com",
		BusinessName: "Test",
		Width:        1280,
		Height:       720,
	}
	result := &ScreenshotResult{
		Homepage: "/tmp/h.png",
		Contact:  "/tmp/c.png",
		Google:   "/tmp/g.png",
	}

	script := buildPlaywrightScript(cfg, result)

	assert.Contains(t, script, "width: 1280")
	assert.Contains(t, script, "height: 720")
}

func TestBuildPlaywrightScript_URLWithFragmentAndQuery(t *testing.T) {
	cfg := DefaultScreenshotConfig(
		"https://example.com/page?foo=bar&baz=1#section",
		"Test Biz",
		"/tmp/out",
	)
	result := &ScreenshotResult{
		Homepage: "/tmp/out/homepage.png",
		Contact:  "/tmp/out/contact.png",
		Google:   "/tmp/out/google.png",
	}

	script := buildPlaywrightScript(cfg, result)

	// URL should be preserved intact within quotes
	assert.Contains(t, script, "foo=bar&baz=1#section")
}

// --- Annotation edge cases ---

func TestBuildAnnotateArgs_ZeroRadius(t *testing.T) {
	cfg := AnnotateConfig{
		InputPath:  "/tmp/in.png",
		OutputPath: "/tmp/out.png",
		Annotations: []Annotation{
			{Type: AnnotationCircle, X: 100, Y: 100, Radius: 0, Color: "red"},
		},
	}

	args := BuildAnnotateArgs(cfg)

	joined := strings.Join(args, " ")
	// circle with radius 0: both points same
	assert.Contains(t, joined, "circle 100,100 100,100")
}

func TestBuildAnnotateArgs_NegativeCoordinates(t *testing.T) {
	cfg := AnnotateConfig{
		InputPath:  "/tmp/in.png",
		OutputPath: "/tmp/out.png",
		Annotations: []Annotation{
			{Type: AnnotationCircle, X: -10, Y: -20, Radius: 5, Color: "red"},
		},
	}

	args := BuildAnnotateArgs(cfg)

	joined := strings.Join(args, " ")
	assert.Contains(t, joined, "circle -10,-20 -5,-20")
}

func TestBuildAnnotateArgs_EmptyText(t *testing.T) {
	cfg := AnnotateConfig{
		InputPath:  "/tmp/in.png",
		OutputPath: "/tmp/out.png",
		Annotations: []Annotation{
			{Type: AnnotationText, X: 0, Y: 0, Text: "", Color: "red"},
		},
	}

	args := BuildAnnotateArgs(cfg)

	// Should still produce annotate command, just with empty text
	joined := strings.Join(args, " ")
	assert.Contains(t, joined, "-annotate")
}

func TestBuildAnnotateArgs_TextWithSpecialChars(t *testing.T) {
	cfg := AnnotateConfig{
		InputPath:  "/tmp/in.png",
		OutputPath: "/tmp/out.png",
		Annotations: []Annotation{
			{Type: AnnotationText, X: 10, Y: 10, Text: "62% of calls \"missed\"", Color: "red"},
		},
	}

	args := BuildAnnotateArgs(cfg)

	// Verify the text is passed as a separate arg (not inline in -draw)
	found := false
	for _, arg := range args {
		if strings.Contains(arg, "62%") {
			found = true
			break
		}
	}
	assert.True(t, found, "text with special chars should appear in args")
}

func TestBuildAnnotateArgs_ArrowSameStartEnd(t *testing.T) {
	cfg := AnnotateConfig{
		InputPath:  "/tmp/in.png",
		OutputPath: "/tmp/out.png",
		Annotations: []Annotation{
			{Type: AnnotationArrow, X: 50, Y: 50, X2: 50, Y2: 50, Color: "blue"},
		},
	}

	args := BuildAnnotateArgs(cfg)

	joined := strings.Join(args, " ")
	assert.Contains(t, joined, "line 50,50 50,50") // zero-length line
}

func TestBuildAnnotateArgs_AllThreeTypes(t *testing.T) {
	cfg := AnnotateConfig{
		InputPath:  "/tmp/in.png",
		OutputPath: "/tmp/out.png",
		Annotations: []Annotation{
			{Type: AnnotationCircle, X: 100, Y: 100, Radius: 30, Color: "red"},
			{Type: AnnotationArrow, X: 200, Y: 200, X2: 300, Y2: 300, Color: "blue"},
			{Type: AnnotationText, X: 50, Y: 50, Text: "Test", Color: "green"},
		},
	}

	args := BuildAnnotateArgs(cfg)

	joined := strings.Join(args, " ")
	assert.Contains(t, joined, "circle 100,100")
	assert.Contains(t, joined, "line 200,200 300,300")
	assert.Contains(t, joined, "polygon") // arrowhead
	assert.Contains(t, joined, "-annotate")
	assert.Contains(t, joined, "Test")

	// Verify input/output are first/last
	assert.Equal(t, "/tmp/in.png", args[0])
	assert.Equal(t, "/tmp/out.png", args[len(args)-1])
}

func TestBuildAnnotateArgs_CircleArgOrder(t *testing.T) {
	cfg := AnnotateConfig{
		InputPath:  "in.png",
		OutputPath: "out.png",
		Annotations: []Annotation{
			{Type: AnnotationCircle, X: 50, Y: 60, Radius: 25, Color: "red"},
		},
	}

	args := BuildAnnotateArgs(cfg)

	// Expected order: input, -fill none, -stroke color, -strokewidth 4, -draw "circle ...", output
	require.True(t, len(args) >= 8)
	assert.Equal(t, "in.png", args[0])
	assert.Equal(t, "-fill", args[1])
	assert.Equal(t, "none", args[2])
	assert.Equal(t, "-stroke", args[3])
	assert.Equal(t, "red", args[4])
	assert.Equal(t, "-strokewidth", args[5])
	assert.Equal(t, "4", args[6])
	assert.Equal(t, "-draw", args[7])
	assert.Equal(t, "circle 50,60 75,60", args[8])
	assert.Equal(t, "out.png", args[9])
}

func TestBuildAnnotateArgs_TextUndercolorHasAlpha(t *testing.T) {
	cfg := AnnotateConfig{
		InputPath:  "in.png",
		OutputPath: "out.png",
		Annotations: []Annotation{
			{Type: AnnotationText, X: 0, Y: 0, Text: "hi", Color: "red"},
		},
	}

	args := BuildAnnotateArgs(cfg)

	// Find the -undercolor arg
	for i, arg := range args {
		if arg == "-undercolor" && i+1 < len(args) {
			// Should be hex + "80" for alpha
			assert.Equal(t, "#FF000080", args[i+1])
			return
		}
	}
	t.Fatal("expected -undercolor arg")
}

// --- DefaultAnnotations edge cases ---

func TestDefaultAnnotations_NilSignals(t *testing.T) {
	for _, st := range []ScreenshotType{ScreenshotHomepage, ScreenshotContact, ScreenshotGoogle} {
		anns := DefaultAnnotations(st, nil)
		assert.Empty(t, anns, "nil signals should produce no annotations for %s", st)
	}
}

func TestDefaultAnnotations_UnknownScreenshotType(t *testing.T) {
	anns := DefaultAnnotations("unknown_type", map[string]bool{"no_chat": true})
	assert.Nil(t, anns)
}

func TestDefaultAnnotations_IrrelevantSignals(t *testing.T) {
	// Contact page signals shouldn't affect homepage annotations
	signals := map[string]bool{
		"no_chat":           true,
		"no_online_booking": true,
	}
	anns := DefaultAnnotations(ScreenshotHomepage, signals)
	assert.Empty(t, anns) // these signals only affect contact page
}

func TestDefaultAnnotations_AllContactSignals(t *testing.T) {
	signals := map[string]bool{
		"no_chat":           true,
		"no_online_booking": true,
	}
	anns := DefaultAnnotations(ScreenshotContact, signals)

	// 2 for no_chat (circle + text) + 1 for no_online_booking
	assert.Len(t, anns, 3)

	types := make(map[AnnotationType]int)
	for _, a := range anns {
		types[a.Type]++
	}
	assert.Equal(t, 1, types[AnnotationCircle])
	assert.Equal(t, 2, types[AnnotationText])
}

// --- normalizeColor edge cases ---

func TestNormalizeColor_EmptyString(t *testing.T) {
	result := normalizeColor("")
	assert.Equal(t, "#FF0000", result) // fallback to red
}

func TestNormalizeColor_HexVariants(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"#000", "#000"},           // short hex preserved
		{"#AABBCC", "#AABBCC"},     // standard hex
		{"#aabbcc", "#aabbcc"},     // lowercase hex
		{"#FF000080", "#FF000080"}, // hex with alpha
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeColor(tt.input))
		})
	}
}

// --- annotatedPath edge cases ---

func TestAnnotatedPath_NoExtension(t *testing.T) {
	result := annotatedPath("/tmp/screenshot")
	assert.Equal(t, "/tmp/screenshot_annotated", result)
}

func TestAnnotatedPath_DoubleExtension(t *testing.T) {
	result := annotatedPath("/tmp/file.backup.png")
	assert.Equal(t, "/tmp/file.backup_annotated.png", result)
}

func TestAnnotatedPath_HiddenFile(t *testing.T) {
	result := annotatedPath("/tmp/.hidden.png")
	assert.Equal(t, "/tmp/.hidden_annotated.png", result)
}

func TestAnnotatedPath_EmptyPath(t *testing.T) {
	result := annotatedPath("")
	assert.Equal(t, "_annotated", result)
}
