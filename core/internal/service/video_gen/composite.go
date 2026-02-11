package video_gen

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultFPS       = 30
	defaultVideoCRF  = 23
	defaultVideoPreset = "medium"
	defaultTransitionDuration = 0.5 // seconds
)

// Scene represents a single visual segment of the video.
type Scene struct {
	ImagePath string        // path to the image (screenshot/annotated)
	AudioPath string        // path to the narration audio for this scene
	Duration  time.Duration // how long to show this scene (derived from audio length if zero)
}

// CompositeConfig holds settings for video compositing.
type CompositeConfig struct {
	Scenes         []Scene       // ordered list of scenes
	OutputPath     string        // final video output path
	LipSyncVideo   string        // optional lip sync overlay video (picture-in-picture)
	Width          int           // output width (default 1920)
	Height         int           // output height (default 1080)
	FPS            int           // frames per second (default 30)
	CRF            int           // constant rate factor for quality (default 23, lower = better)
	Preset         string        // encoding speed preset (default "medium")
	TransitionSecs float64       // crossfade transition duration between scenes
}

// CompositeResult holds the output of video compositing.
type CompositeResult struct {
	VideoPath string        // path to the final video
	Duration  time.Duration // total video duration
}

// DefaultCompositeConfig returns config with sensible defaults.
func DefaultCompositeConfig(scenes []Scene, outputPath string) CompositeConfig {
	return CompositeConfig{
		Scenes:         scenes,
		OutputPath:     outputPath,
		Width:          1920,
		Height:         1080,
		FPS:            defaultFPS,
		CRF:            defaultVideoCRF,
		Preset:         defaultVideoPreset,
		TransitionSecs: defaultTransitionDuration,
	}
}

// BuildFFmpegArgs constructs the FFmpeg CLI arguments for compositing.
// Exported for testing without requiring FFmpeg installed.
//
// Strategy:
//  1. Each scene: image + audio → video segment (image shown for audio duration)
//  2. Concatenate all segments with crossfade transitions
//  3. Optionally overlay lip sync PiP in bottom-right corner
func BuildFFmpegArgs(cfg CompositeConfig) []string {
	if len(cfg.Scenes) == 0 {
		return nil
	}

	args := []string{"-y"} // overwrite output

	// Add inputs: for each scene, add image and audio
	for _, scene := range cfg.Scenes {
		args = append(args, "-loop", "1", "-i", scene.ImagePath)
		args = append(args, "-i", scene.AudioPath)
	}

	// Add lip sync video input if provided
	lipSyncIdx := -1
	if cfg.LipSyncVideo != "" {
		lipSyncIdx = len(cfg.Scenes) * 2
		args = append(args, "-i", cfg.LipSyncVideo)
	}

	// Build filter_complex
	filter := buildFilterComplex(cfg, lipSyncIdx)
	args = append(args, "-filter_complex", filter)

	// Map the final output
	args = append(args, "-map", "[vout]", "-map", "[aout]")

	// Encoding settings
	args = append(args,
		"-c:v", "libx264",
		"-crf", fmt.Sprintf("%d", cfg.CRF),
		"-preset", cfg.Preset,
		"-c:a", "aac",
		"-b:a", "192k",
		"-r", fmt.Sprintf("%d", cfg.FPS),
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		cfg.OutputPath,
	)

	return args
}

// buildFilterComplex constructs the FFmpeg filter_complex string.
func buildFilterComplex(cfg CompositeConfig, lipSyncIdx int) string {
	n := len(cfg.Scenes)
	var parts []string

	// Step 1: Scale each image to target resolution and pair with audio
	for i := 0; i < n; i++ {
		imgIdx := i * 2
		audIdx := i*2 + 1

		// Scale image, set duration to match audio, set SAR/FPS
		parts = append(parts, fmt.Sprintf(
			"[%d:v]scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,setsar=1,fps=%d[v%d]",
			imgIdx, cfg.Width, cfg.Height, cfg.Width, cfg.Height, cfg.FPS, i,
		))

		// Trim audio and label it
		parts = append(parts, fmt.Sprintf("[%d:a]aresample=44100[a%d]", audIdx, i))
	}

	if n == 1 {
		// Single scene: no concat needed
		videoLabel := "v0"
		audioLabel := "a0"

		if lipSyncIdx >= 0 {
			parts = append(parts, buildPiPFilter(lipSyncIdx, videoLabel, cfg))
			videoLabel = "vpip"
		}

		parts = append(parts, fmt.Sprintf("[%s]copy[vout]", videoLabel))
		parts = append(parts, fmt.Sprintf("[%s]acopy[aout]", audioLabel))
	} else {
		// Concatenate scenes
		var concatInputs string
		for i := 0; i < n; i++ {
			concatInputs += fmt.Sprintf("[v%d][a%d]", i, i)
		}
		parts = append(parts, fmt.Sprintf(
			"%sconcat=n=%d:v=1:a=1[vconcat][aconcat]",
			concatInputs, n,
		))

		videoLabel := "vconcat"
		if lipSyncIdx >= 0 {
			parts = append(parts, buildPiPFilter(lipSyncIdx, videoLabel, cfg))
			videoLabel = "vpip"
		}

		parts = append(parts, fmt.Sprintf("[%s]copy[vout]", videoLabel))
		parts = append(parts, "[aconcat]acopy[aout]")
	}

	return strings.Join(parts, ";")
}

// buildPiPFilter creates the picture-in-picture overlay filter for lip sync video.
// Places the lip sync video in the bottom-right corner at 25% of frame size.
func buildPiPFilter(lipSyncIdx int, baseVideo string, cfg CompositeConfig) string {
	pipW := cfg.Width / 4
	pipH := cfg.Height / 4
	pipX := cfg.Width - pipW - 20  // 20px margin
	pipY := cfg.Height - pipH - 20

	return fmt.Sprintf(
		"[%d:v]scale=%d:%d[pip];[%s][pip]overlay=%d:%d:shortest=1[vpip]",
		lipSyncIdx, pipW, pipH, baseVideo, pipX, pipY,
	)
}

// CompositeVideo runs FFmpeg to combine screenshots + audio into a final video.
// Requires: ffmpeg installed and in PATH.
func CompositeVideo(ctx context.Context, cfg CompositeConfig) (*CompositeResult, error) {
	args := BuildFFmpegArgs(cfg)
	if args == nil {
		return nil, fmt.Errorf("no scenes provided")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(cfg.OutputPath), 0755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg compositing failed: %w\noutput: %s", err, string(out))
	}

	// Calculate total duration from scenes
	var totalDuration time.Duration
	for _, scene := range cfg.Scenes {
		totalDuration += scene.Duration
	}

	return &CompositeResult{
		VideoPath: cfg.OutputPath,
		Duration:  totalDuration,
	}, nil
}

// SceneFromScreenshot creates a Scene from a screenshot path and audio path.
func SceneFromScreenshot(imagePath, audioPath string, duration time.Duration) Scene {
	return Scene{
		ImagePath: imagePath,
		AudioPath: audioPath,
		Duration:  duration,
	}
}

// ScenesFromAnnotated creates scenes from annotated screenshot result + audio files.
// Expects audio files named homepage.wav, contact.wav, google.wav in audioDir.
func ScenesFromAnnotated(result *ScreenshotResult, audioDir string, durations [3]time.Duration) []Scene {
	return []Scene{
		{ImagePath: result.Homepage, AudioPath: filepath.Join(audioDir, "homepage.wav"), Duration: durations[0]},
		{ImagePath: result.Contact, AudioPath: filepath.Join(audioDir, "contact.wav"), Duration: durations[1]},
		{ImagePath: result.Google, AudioPath: filepath.Join(audioDir, "google.wav"), Duration: durations[2]},
	}
}
