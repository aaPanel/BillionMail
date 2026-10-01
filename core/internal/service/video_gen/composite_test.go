package video_gen

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultCompositeConfig(t *testing.T) {
	scenes := []Scene{
		{ImagePath: "/tmp/img.png", AudioPath: "/tmp/audio.wav", Duration: 10 * time.Second},
	}
	cfg := DefaultCompositeConfig(scenes, "/tmp/out.mp4")

	assert.Equal(t, 1920, cfg.Width)
	assert.Equal(t, 1080, cfg.Height)
	assert.Equal(t, 30, cfg.FPS)
	assert.Equal(t, 23, cfg.CRF)
	assert.Equal(t, "medium", cfg.Preset)
	assert.Equal(t, 0.5, cfg.TransitionSecs)
	assert.Equal(t, "/tmp/out.mp4", cfg.OutputPath)
	assert.Len(t, cfg.Scenes, 1)
}

func TestBuildFFmpegArgs_EmptyScenes(t *testing.T) {
	cfg := CompositeConfig{Scenes: nil, OutputPath: "/tmp/out.mp4"}
	args := BuildFFmpegArgs(cfg)
	assert.Nil(t, args)
}

func TestBuildFFmpegArgs_SingleScene(t *testing.T) {
	cfg := DefaultCompositeConfig([]Scene{
		{ImagePath: "/tmp/homepage.png", AudioPath: "/tmp/homepage.wav", Duration: 15 * time.Second},
	}, "/tmp/output.mp4")

	args := BuildFFmpegArgs(cfg)
	require.NotNil(t, args)

	joined := strings.Join(args, " ")

	// Should have -y flag
	assert.Equal(t, "-y", args[0])

	// Should have input image and audio
	assert.Contains(t, joined, "-loop 1 -i /tmp/homepage.png")
	assert.Contains(t, joined, "-i /tmp/homepage.wav")

	// Should have filter_complex
	assert.Contains(t, joined, "-filter_complex")

	// Should map video and audio outputs
	assert.Contains(t, joined, "-map [vout]")
	assert.Contains(t, joined, "-map [aout]")

	// Encoding settings
	assert.Contains(t, joined, "-c:v libx264")
	assert.Contains(t, joined, "-crf 23")
	assert.Contains(t, joined, "-preset medium")
	assert.Contains(t, joined, "-c:a aac")
	assert.Contains(t, joined, "-pix_fmt yuv420p")
	assert.Contains(t, joined, "-movflags +faststart")

	// Output path should be last
	assert.Equal(t, "/tmp/output.mp4", args[len(args)-1])
}

func TestBuildFFmpegArgs_ThreeScenes(t *testing.T) {
	scenes := []Scene{
		{ImagePath: "/tmp/homepage.png", AudioPath: "/tmp/h.wav", Duration: 10 * time.Second},
		{ImagePath: "/tmp/contact.png", AudioPath: "/tmp/c.wav", Duration: 15 * time.Second},
		{ImagePath: "/tmp/google.png", AudioPath: "/tmp/g.wav", Duration: 10 * time.Second},
	}
	cfg := DefaultCompositeConfig(scenes, "/tmp/final.mp4")

	args := BuildFFmpegArgs(cfg)
	require.NotNil(t, args)

	joined := strings.Join(args, " ")

	// 3 image inputs + 3 audio inputs = 6 -i flags
	assert.Equal(t, 6, strings.Count(joined, " -i "))

	// Should have concat filter
	assert.Contains(t, joined, "concat=n=3:v=1:a=1")

	// All 3 images referenced
	assert.Contains(t, joined, "/tmp/homepage.png")
	assert.Contains(t, joined, "/tmp/contact.png")
	assert.Contains(t, joined, "/tmp/google.png")
}

func TestBuildFFmpegArgs_WithLipSync(t *testing.T) {
	scenes := []Scene{
		{ImagePath: "/tmp/img.png", AudioPath: "/tmp/audio.wav", Duration: 10 * time.Second},
	}
	cfg := DefaultCompositeConfig(scenes, "/tmp/out.mp4")
	cfg.LipSyncVideo = "/tmp/lipsync.mp4"

	args := BuildFFmpegArgs(cfg)
	require.NotNil(t, args)

	joined := strings.Join(args, " ")

	// Lip sync video should be added as input
	assert.Contains(t, joined, "-i /tmp/lipsync.mp4")

	// PiP overlay filter should be present
	assert.Contains(t, joined, "overlay=")
	assert.Contains(t, joined, "[vpip]")
}

func TestBuildFFmpegArgs_ThreeScenesWithLipSync(t *testing.T) {
	scenes := []Scene{
		{ImagePath: "/tmp/h.png", AudioPath: "/tmp/h.wav", Duration: 10 * time.Second},
		{ImagePath: "/tmp/c.png", AudioPath: "/tmp/c.wav", Duration: 15 * time.Second},
		{ImagePath: "/tmp/g.png", AudioPath: "/tmp/g.wav", Duration: 10 * time.Second},
	}
	cfg := DefaultCompositeConfig(scenes, "/tmp/out.mp4")
	cfg.LipSyncVideo = "/tmp/lip.mp4"

	args := BuildFFmpegArgs(cfg)
	require.NotNil(t, args)

	joined := strings.Join(args, " ")

	// 3 images + 3 audio + 1 lip sync = 7 -i flags
	assert.Equal(t, 7, strings.Count(joined, " -i "))

	// Both concat and PiP
	assert.Contains(t, joined, "concat=n=3")
	assert.Contains(t, joined, "overlay=")
}

func TestBuildFFmpegArgs_CustomCRFAndPreset(t *testing.T) {
	cfg := DefaultCompositeConfig([]Scene{
		{ImagePath: "/tmp/img.png", AudioPath: "/tmp/a.wav"},
	}, "/tmp/out.mp4")
	cfg.CRF = 18
	cfg.Preset = "slow"

	args := BuildFFmpegArgs(cfg)
	joined := strings.Join(args, " ")

	assert.Contains(t, joined, "-crf 18")
	assert.Contains(t, joined, "-preset slow")
}

func TestBuildFFmpegArgs_CustomFPS(t *testing.T) {
	cfg := DefaultCompositeConfig([]Scene{
		{ImagePath: "/tmp/img.png", AudioPath: "/tmp/a.wav"},
	}, "/tmp/out.mp4")
	cfg.FPS = 60

	args := BuildFFmpegArgs(cfg)
	joined := strings.Join(args, " ")

	assert.Contains(t, joined, "-r 60")
	assert.Contains(t, joined, "fps=60")
}

func TestBuildFFmpegArgs_FilterComplexStructure(t *testing.T) {
	scenes := []Scene{
		{ImagePath: "/tmp/a.png", AudioPath: "/tmp/a.wav"},
		{ImagePath: "/tmp/b.png", AudioPath: "/tmp/b.wav"},
	}
	cfg := DefaultCompositeConfig(scenes, "/tmp/out.mp4")

	args := BuildFFmpegArgs(cfg)

	// Find filter_complex arg
	var filterArg string
	for i, arg := range args {
		if arg == "-filter_complex" && i+1 < len(args) {
			filterArg = args[i+1]
			break
		}
	}
	require.NotEmpty(t, filterArg)

	// Should have scale filters for both scenes
	assert.Contains(t, filterArg, "[0:v]scale=1920:1080")
	assert.Contains(t, filterArg, "[2:v]scale=1920:1080")

	// Audio resampling
	assert.Contains(t, filterArg, "[1:a]aresample=44100[a0]")
	assert.Contains(t, filterArg, "[3:a]aresample=44100[a1]")

	// Concat with both scenes
	assert.Contains(t, filterArg, "[v0][a0][v1][a1]concat=n=2:v=1:a=1")

	// Final output labels
	assert.Contains(t, filterArg, "[vout]")
	assert.Contains(t, filterArg, "[aout]")
}

func TestSceneFromScreenshot(t *testing.T) {
	scene := SceneFromScreenshot("/tmp/img.png", "/tmp/audio.wav", 12*time.Second)

	assert.Equal(t, "/tmp/img.png", scene.ImagePath)
	assert.Equal(t, "/tmp/audio.wav", scene.AudioPath)
	assert.Equal(t, 12*time.Second, scene.Duration)
}

func TestScenesFromAnnotated(t *testing.T) {
	result := &ScreenshotResult{
		Homepage: "/tmp/homepage_annotated.png",
		Contact:  "/tmp/contact_annotated.png",
		Google:   "/tmp/google_annotated.png",
	}
	durations := [3]time.Duration{10 * time.Second, 15 * time.Second, 10 * time.Second}

	scenes := ScenesFromAnnotated(result, "/tmp/audio", durations)

	require.Len(t, scenes, 3)

	assert.Equal(t, "/tmp/homepage_annotated.png", scenes[0].ImagePath)
	assert.Equal(t, "/tmp/audio/homepage.wav", scenes[0].AudioPath)
	assert.Equal(t, 10*time.Second, scenes[0].Duration)

	assert.Equal(t, "/tmp/contact_annotated.png", scenes[1].ImagePath)
	assert.Equal(t, "/tmp/audio/contact.wav", scenes[1].AudioPath)
	assert.Equal(t, 15*time.Second, scenes[1].Duration)

	assert.Equal(t, "/tmp/google_annotated.png", scenes[2].ImagePath)
	assert.Equal(t, "/tmp/audio/google.wav", scenes[2].AudioPath)
	assert.Equal(t, 10*time.Second, scenes[2].Duration)
}

func TestBuildFFmpegArgs_PiPPosition(t *testing.T) {
	cfg := DefaultCompositeConfig([]Scene{
		{ImagePath: "/tmp/img.png", AudioPath: "/tmp/a.wav"},
	}, "/tmp/out.mp4")
	cfg.LipSyncVideo = "/tmp/lip.mp4"

	args := BuildFFmpegArgs(cfg)

	var filterArg string
	for i, arg := range args {
		if arg == "-filter_complex" && i+1 < len(args) {
			filterArg = args[i+1]
			break
		}
	}
	require.NotEmpty(t, filterArg)

	// PiP should be 25% size (480x270) at bottom-right with 20px margin
	assert.Contains(t, filterArg, "scale=480:270[pip]")
	// Position: 1920-480-20=1420, 1080-270-20=790
	assert.Contains(t, filterArg, "overlay=1420:790")
}

func TestBuildFFmpegArgs_SingleSceneNoConcat(t *testing.T) {
	cfg := DefaultCompositeConfig([]Scene{
		{ImagePath: "/tmp/img.png", AudioPath: "/tmp/a.wav"},
	}, "/tmp/out.mp4")

	args := BuildFFmpegArgs(cfg)

	var filterArg string
	for i, arg := range args {
		if arg == "-filter_complex" && i+1 < len(args) {
			filterArg = args[i+1]
			break
		}
	}

	// Single scene should NOT use concat
	assert.NotContains(t, filterArg, "concat")
	// Should use copy/acopy for single scene
	assert.Contains(t, filterArg, "copy[vout]")
	assert.Contains(t, filterArg, "acopy[aout]")
}
