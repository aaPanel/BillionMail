package video_gen

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Annotation describes a single visual annotation on a screenshot.
type Annotation struct {
	Type   AnnotationType
	X, Y   int    // center position for circles, start position for arrows/text
	X2, Y2 int    // end position (arrows only)
	Radius int    // circle radius
	Text   string // text overlay content
	Color  string // color name or hex (default: red)
}

// AnnotationType identifies the kind of annotation.
type AnnotationType int

const (
	AnnotationCircle AnnotationType = iota
	AnnotationArrow
	AnnotationText
)

// AnnotateConfig holds settings for annotating a screenshot.
type AnnotateConfig struct {
	InputPath   string       // source PNG
	OutputPath  string       // annotated output PNG
	Annotations []Annotation // visual annotations to apply
}

// DefaultAnnotations returns standard annotations for each screenshot type
// based on lead scoring signals.
func DefaultAnnotations(screenshotType ScreenshotType, signals map[string]bool) []Annotation {
	switch screenshotType {
	case ScreenshotContact:
		var anns []Annotation
		if signals["no_chat"] {
			anns = append(anns, Annotation{
				Type: AnnotationCircle,
				X:    1700, Y: 400,
				Radius: 80,
				Color:  "red",
			})
			anns = append(anns, Annotation{
				Type: AnnotationText,
				X:    1500, Y: 520,
				Text:  "No live chat or AI receptionist",
				Color: "red",
			})
		}
		if signals["no_online_booking"] {
			anns = append(anns, Annotation{
				Type: AnnotationText,
				X:    100, Y: 950,
				Text:  "No online self-booking available",
				Color: "#FF6600",
			})
		}
		return anns

	case ScreenshotHomepage:
		var anns []Annotation
		if signals["voicemail_after_hrs"] {
			anns = append(anns, Annotation{
				Type: AnnotationText,
				X:    100, Y: 1000,
				Text:  "62% of calls go unanswered during peak hours",
				Color: "red",
			})
		}
		return anns

	case ScreenshotGoogle:
		var anns []Annotation
		if signals["high_spend_low_rating"] {
			anns = append(anns, Annotation{
				Type: AnnotationText,
				X:    100, Y: 950,
				Text:  "High ad spend but room for improvement",
				Color: "#FF6600",
			})
		}
		return anns
	}
	return nil
}

// Annotate applies visual annotations to a screenshot using ImageMagick.
// Requires: convert (ImageMagick) installed.
func Annotate(ctx context.Context, cfg AnnotateConfig) error {
	args := BuildAnnotateArgs(cfg)
	cmd := exec.CommandContext(ctx, "magick", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("imagemagick annotate failed: %w\noutput: %s", err, string(out))
	}
	return nil
}

// BuildAnnotateArgs constructs the ImageMagick CLI arguments for the given config.
// Exported for testing without requiring ImageMagick installed.
func BuildAnnotateArgs(cfg AnnotateConfig) []string {
	args := []string{cfg.InputPath}

	for _, ann := range cfg.Annotations {
		color := ann.Color
		if color == "" {
			color = "red"
		}

		switch ann.Type {
		case AnnotationCircle:
			// Draw a circle outline
			args = append(args,
				"-fill", "none",
				"-stroke", color,
				"-strokewidth", "4",
				"-draw", fmt.Sprintf("circle %d,%d %d,%d",
					ann.X, ann.Y,
					ann.X+ann.Radius, ann.Y),
			)

		case AnnotationArrow:
			// Draw a line with arrowhead
			args = append(args,
				"-fill", color,
				"-stroke", color,
				"-strokewidth", "3",
				"-draw", fmt.Sprintf("line %d,%d %d,%d",
					ann.X, ann.Y, ann.X2, ann.Y2),
			)
			// Arrowhead triangle at endpoint
			ax, ay := ann.X2, ann.Y2
			args = append(args,
				"-fill", color,
				"-stroke", "none",
				"-draw", fmt.Sprintf("polygon %d,%d %d,%d %d,%d",
					ax, ay-8, ax, ay+8, ax+16, ay),
			)

		case AnnotationText:
			// Text with background for readability
			args = append(args,
				"-font", "Helvetica-Bold",
				"-pointsize", "28",
				"-fill", "white",
				"-stroke", "none",
				"-undercolor", fmt.Sprintf("%s80", normalizeColor(color)), // 50% opacity bg
				"-gravity", "NorthWest",
				"-annotate", fmt.Sprintf("+%d+%d", ann.X, ann.Y),
				fmt.Sprintf(" %s ", ann.Text),
			)
		}
	}

	args = append(args, cfg.OutputPath)
	return args
}

// normalizeColor ensures color is in hex format for the undercolor overlay.
func normalizeColor(color string) string {
	if strings.HasPrefix(color, "#") {
		return color
	}
	// Named color → hex mapping for common annotation colors
	named := map[string]string{
		"red":    "#FF0000",
		"orange": "#FF6600",
		"yellow": "#FFCC00",
		"green":  "#00CC00",
		"blue":   "#0066FF",
	}
	if hex, ok := named[strings.ToLower(color)]; ok {
		return hex
	}
	return "#FF0000" // default red
}

// AnnotateAll applies default annotations to all 3 screenshots based on lead signals.
// Returns paths to annotated files (saved as *_annotated.png alongside originals).
func AnnotateAll(ctx context.Context, result *ScreenshotResult, signals map[string]bool) (*ScreenshotResult, error) {
	annotated := &ScreenshotResult{
		Homepage: annotatedPath(result.Homepage),
		Contact:  annotatedPath(result.Contact),
		Google:   annotatedPath(result.Google),
	}

	pairs := []struct {
		typ    ScreenshotType
		input  string
		output string
	}{
		{ScreenshotHomepage, result.Homepage, annotated.Homepage},
		{ScreenshotContact, result.Contact, annotated.Contact},
		{ScreenshotGoogle, result.Google, annotated.Google},
	}

	for _, p := range pairs {
		anns := DefaultAnnotations(p.typ, signals)
		if len(anns) == 0 {
			// No annotations — just copy
			anns = nil
		}
		if err := Annotate(ctx, AnnotateConfig{
			InputPath:   p.input,
			OutputPath:  p.output,
			Annotations: anns,
		}); err != nil {
			return nil, fmt.Errorf("annotate %s: %w", p.typ, err)
		}
	}

	return annotated, nil
}

// annotatedPath derives the annotated output path from the original.
func annotatedPath(original string) string {
	ext := filepath.Ext(original)
	base := strings.TrimSuffix(original, ext)
	return base + "_annotated" + ext
}
