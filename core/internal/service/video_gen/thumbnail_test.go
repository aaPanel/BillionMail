package video_gen

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- DefaultThumbnailConfig ---

func TestDefaultThumbnailConfig(t *testing.T) {
	cfg := DefaultThumbnailConfig("/tmp/screenshots/homepage.png", "/tmp/output")

	assert.Equal(t, "/tmp/screenshots/homepage.png", cfg.InputPath)
	assert.Equal(t, "/tmp/output/thumbnail.png", cfg.OutputPath)
	assert.Equal(t, 640, cfg.Width)
	assert.Equal(t, 360, cfg.Height)
}

func TestDefaultThumbnailConfig_EmptyInputPath(t *testing.T) {
	cfg := DefaultThumbnailConfig("", "/tmp/output")
	assert.Empty(t, cfg.InputPath)
	assert.Equal(t, "/tmp/output/thumbnail.png", cfg.OutputPath)
}

func TestDefaultThumbnailConfig_EmptyOutputDir(t *testing.T) {
	cfg := DefaultThumbnailConfig("/tmp/input.png", "")
	assert.Equal(t, "thumbnail.png", cfg.OutputPath)
}

func TestDefaultThumbnailConfig_PathWithSpaces(t *testing.T) {
	cfg := DefaultThumbnailConfig("/tmp/my screenshots/home.png", "/tmp/my output")
	assert.Equal(t, "/tmp/my screenshots/home.png", cfg.InputPath)
	assert.Equal(t, "/tmp/my output/thumbnail.png", cfg.OutputPath)
}

// --- BuildThumbnailArgs ---

func TestBuildThumbnailArgs(t *testing.T) {
	tests := []struct {
		name string
		cfg  ThumbnailConfig
		want []string
	}{
		{
			"defaults",
			ThumbnailConfig{
				InputPath:  "/tmp/input.png",
				OutputPath: "/tmp/thumb.png",
				Width:      640,
				Height:     360,
			},
			[]string{"/tmp/input.png", "-resize", "640x360!", "-quality", "90", "/tmp/thumb.png"},
		},
		{
			"custom dimensions",
			ThumbnailConfig{
				InputPath:  "/tmp/input.png",
				OutputPath: "/tmp/thumb.png",
				Width:      320,
				Height:     180,
			},
			[]string{"/tmp/input.png", "-resize", "320x180!", "-quality", "90", "/tmp/thumb.png"},
		},
		{
			"zero dimensions use defaults",
			ThumbnailConfig{
				InputPath:  "/tmp/input.png",
				OutputPath: "/tmp/thumb.png",
				Width:      0,
				Height:     0,
			},
			[]string{"/tmp/input.png", "-resize", "640x360!", "-quality", "90", "/tmp/thumb.png"},
		},
		{
			"only width zero",
			ThumbnailConfig{
				InputPath:  "/tmp/input.png",
				OutputPath: "/tmp/thumb.png",
				Width:      0,
				Height:     180,
			},
			[]string{"/tmp/input.png", "-resize", "640x180!", "-quality", "90", "/tmp/thumb.png"},
		},
		{
			"only height zero",
			ThumbnailConfig{
				InputPath:  "/tmp/input.png",
				OutputPath: "/tmp/thumb.png",
				Width:      320,
				Height:     0,
			},
			[]string{"/tmp/input.png", "-resize", "320x360!", "-quality", "90", "/tmp/thumb.png"},
		},
		{
			"very large dimensions",
			ThumbnailConfig{
				InputPath:  "/tmp/input.png",
				OutputPath: "/tmp/thumb.png",
				Width:      3840,
				Height:     2160,
			},
			[]string{"/tmp/input.png", "-resize", "3840x2160!", "-quality", "90", "/tmp/thumb.png"},
		},
		{
			"1x1 pixel",
			ThumbnailConfig{
				InputPath:  "/tmp/input.png",
				OutputPath: "/tmp/thumb.png",
				Width:      1,
				Height:     1,
			},
			[]string{"/tmp/input.png", "-resize", "1x1!", "-quality", "90", "/tmp/thumb.png"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildThumbnailArgs(tt.cfg)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildThumbnailArgs_InputOutputPreserved(t *testing.T) {
	cfg := ThumbnailConfig{
		InputPath:  "/path/with spaces/input.png",
		OutputPath: "/path/with spaces/thumb.png",
		Width:      640,
		Height:     360,
	}
	args := BuildThumbnailArgs(cfg)

	assert.Equal(t, "/path/with spaces/input.png", args[0])
	assert.Equal(t, "/path/with spaces/thumb.png", args[len(args)-1])
}

func TestBuildThumbnailArgs_EmptyPaths(t *testing.T) {
	cfg := ThumbnailConfig{
		InputPath:  "",
		OutputPath: "",
		Width:      640,
		Height:     360,
	}
	args := BuildThumbnailArgs(cfg)

	assert.Equal(t, "", args[0])
	assert.Equal(t, "", args[len(args)-1])
	assert.Len(t, args, 6) // input, -resize, WxH!, -quality, 90, output
}

func TestBuildThumbnailArgs_ConsistentArgCount(t *testing.T) {
	// All configs should produce exactly 6 args
	configs := []ThumbnailConfig{
		{InputPath: "a", OutputPath: "b", Width: 100, Height: 100},
		{InputPath: "a", OutputPath: "b", Width: 0, Height: 0},
		{InputPath: "", OutputPath: "", Width: 1, Height: 1},
	}
	for _, cfg := range configs {
		args := BuildThumbnailArgs(cfg)
		assert.Len(t, args, 6)
	}
}

func TestBuildThumbnailArgs_ForceResize(t *testing.T) {
	// The ! suffix forces exact dimensions (ignores aspect ratio)
	cfg := ThumbnailConfig{
		InputPath:  "/tmp/input.png",
		OutputPath: "/tmp/thumb.png",
		Width:      640,
		Height:     360,
	}
	args := BuildThumbnailArgs(cfg)
	assert.Equal(t, "640x360!", args[2])
	assert.Contains(t, args[2], "!")
}

// --- GenerateThumbnail error paths ---

func TestGenerateThumbnail_NonexistentInput(t *testing.T) {
	cfg := ThumbnailConfig{
		InputPath:  "/nonexistent/input.png",
		OutputPath: t.TempDir() + "/thumb.png",
		Width:      640,
		Height:     360,
	}

	_, err := GenerateThumbnail(context.Background(), cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "imagemagick thumbnail failed")
}

func TestGenerateThumbnail_CancelledContext(t *testing.T) {
	cfg := ThumbnailConfig{
		InputPath:  "/nonexistent/input.png",
		OutputPath: t.TempDir() + "/thumb.png",
		Width:      640,
		Height:     360,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := GenerateThumbnail(ctx, cfg)
	assert.Error(t, err)
}

// --- ThumbnailConfig struct ---

func TestThumbnailConfig_ZeroValue(t *testing.T) {
	var cfg ThumbnailConfig
	assert.Empty(t, cfg.InputPath)
	assert.Empty(t, cfg.OutputPath)
	assert.Zero(t, cfg.Width)
	assert.Zero(t, cfg.Height)
}
