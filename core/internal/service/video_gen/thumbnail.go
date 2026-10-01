package video_gen

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	defaultThumbWidth  = 640
	defaultThumbHeight = 360
)

// ThumbnailConfig holds settings for thumbnail generation.
type ThumbnailConfig struct {
	InputPath  string // source image (first screenshot)
	OutputPath string // thumbnail output path
	Width      int    // thumbnail width (default 640)
	Height     int    // thumbnail height (default 360)
}

// DefaultThumbnailConfig returns config with standard dimensions.
func DefaultThumbnailConfig(inputPath, outputDir string) ThumbnailConfig {
	return ThumbnailConfig{
		InputPath:  inputPath,
		OutputPath: filepath.Join(outputDir, "thumbnail.png"),
		Width:      defaultThumbWidth,
		Height:     defaultThumbHeight,
	}
}

// BuildThumbnailArgs constructs the ImageMagick CLI arguments for thumbnail generation.
// Exported for testing without requiring ImageMagick installed.
func BuildThumbnailArgs(cfg ThumbnailConfig) []string {
	w := cfg.Width
	if w == 0 {
		w = defaultThumbWidth
	}
	h := cfg.Height
	if h == 0 {
		h = defaultThumbHeight
	}

	return []string{
		cfg.InputPath,
		"-resize", fmt.Sprintf("%dx%d!", w, h),
		"-quality", "90",
		cfg.OutputPath,
	}
}

// GenerateThumbnail creates a thumbnail from a screenshot using ImageMagick.
// Requires: magick (ImageMagick) installed.
func GenerateThumbnail(ctx context.Context, cfg ThumbnailConfig) (string, error) {
	if err := os.MkdirAll(filepath.Dir(cfg.OutputPath), 0755); err != nil {
		return "", fmt.Errorf("create thumbnail dir: %w", err)
	}

	args := BuildThumbnailArgs(cfg)
	cmd := exec.CommandContext(ctx, "magick", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("imagemagick thumbnail failed: %w\noutput: %s", err, string(out))
	}

	return cfg.OutputPath, nil
}
