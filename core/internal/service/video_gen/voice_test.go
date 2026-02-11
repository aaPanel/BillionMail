package video_gen

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultVoiceConfig(t *testing.T) {
	cfg := DefaultVoiceConfig("/tmp/audio")

	assert.Equal(t, "/tmp/audio", cfg.OutputDir)
	// APIKey from env — may be empty in test, that's fine
}

func TestDefaultTTSOutputFormat(t *testing.T) {
	fmt := DefaultTTSOutputFormat()

	assert.Equal(t, "wav", fmt.Container)
	assert.Equal(t, 44100, fmt.SampleRate)
	assert.Equal(t, "pcm_f32le", fmt.Encoding)
}

func TestBuildCloneRequest_Headers(t *testing.T) {
	cfg := VoiceConfig{APIKey: "test-key-123", OutputDir: "/tmp"}
	req := VoiceCloneRequest{
		Name:     "Dr. Smith",
		Mode:     "url",
		AudioURL: "https://example.com/voice.mp3",
		Language: "en",
	}

	httpReq, err := BuildCloneRequest(cfg, req)
	require.NoError(t, err)

	assert.Equal(t, "POST", httpReq.Method)
	assert.Equal(t, cartesiaBaseURL+cartesiaClonePath, httpReq.URL.String())
	assert.Equal(t, "test-key-123", httpReq.Header.Get("X-API-Key"))
	assert.Equal(t, cartesiaAPIVersion, httpReq.Header.Get("Cartesia-Version"))
	assert.Equal(t, "application/json", httpReq.Header.Get("Content-Type"))
}

func TestBuildCloneRequest_Body(t *testing.T) {
	cfg := VoiceConfig{APIKey: "k", OutputDir: "/tmp"}
	req := VoiceCloneRequest{
		Name:        "Test Voice",
		Description: "A test voice",
		Mode:        "url",
		AudioURL:    "https://example.com/sample.wav",
		Language:    "en",
	}

	httpReq, err := BuildCloneRequest(cfg, req)
	require.NoError(t, err)

	body, err := io.ReadAll(httpReq.Body)
	require.NoError(t, err)

	var parsed VoiceCloneRequest
	err = json.Unmarshal(body, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "Test Voice", parsed.Name)
	assert.Equal(t, "url", parsed.Mode)
	assert.Equal(t, "https://example.com/sample.wav", parsed.AudioURL)
	assert.Equal(t, "en", parsed.Language)
}

func TestBuildTTSRequest_Headers(t *testing.T) {
	cfg := VoiceConfig{APIKey: "tts-key", OutputDir: "/tmp"}
	req := TTSRequest{
		VoiceID:      "voice-123",
		Transcript:   "Hello world",
		ModelID:      "sonic-2",
		OutputFormat: DefaultTTSOutputFormat(),
		Language:     "en",
	}

	httpReq, err := BuildTTSRequest(cfg, req)
	require.NoError(t, err)

	assert.Equal(t, "POST", httpReq.Method)
	assert.Equal(t, cartesiaBaseURL+cartesiaTTSPath, httpReq.URL.String())
	assert.Equal(t, "tts-key", httpReq.Header.Get("X-API-Key"))
	assert.Equal(t, cartesiaAPIVersion, httpReq.Header.Get("Cartesia-Version"))
}

func TestBuildTTSRequest_Body(t *testing.T) {
	cfg := VoiceConfig{APIKey: "k", OutputDir: "/tmp"}
	req := TTSRequest{
		VoiceID:      "v-abc",
		Transcript:   "This is a test script for the video.",
		ModelID:      "sonic-2",
		OutputFormat: DefaultTTSOutputFormat(),
		Language:     "en",
	}

	httpReq, err := BuildTTSRequest(cfg, req)
	require.NoError(t, err)

	body, err := io.ReadAll(httpReq.Body)
	require.NoError(t, err)

	var parsed TTSRequest
	err = json.Unmarshal(body, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "v-abc", parsed.VoiceID)
	assert.Equal(t, "sonic-2", parsed.ModelID)
	assert.Contains(t, parsed.Transcript, "test script")
	assert.Equal(t, "wav", parsed.OutputFormat.Container)
	assert.Equal(t, 44100, parsed.OutputFormat.SampleRate)
}

func TestBuildCloneRequest_EmptyAPIKey(t *testing.T) {
	cfg := VoiceConfig{APIKey: "", OutputDir: "/tmp"}
	req := VoiceCloneRequest{Name: "test", Mode: "url", AudioURL: "https://x.com/a.mp3"}

	httpReq, err := BuildCloneRequest(cfg, req)
	require.NoError(t, err)

	// Empty key is set — API will reject, but request builds fine
	assert.Equal(t, "", httpReq.Header.Get("X-API-Key"))
}

func TestBuildTTSRequest_EmptyTranscript(t *testing.T) {
	cfg := VoiceConfig{APIKey: "k", OutputDir: "/tmp"}
	req := TTSRequest{
		VoiceID:      "v-1",
		Transcript:   "",
		ModelID:      "sonic-2",
		OutputFormat: DefaultTTSOutputFormat(),
	}

	httpReq, err := BuildTTSRequest(cfg, req)
	require.NoError(t, err)

	body, err := io.ReadAll(httpReq.Body)
	require.NoError(t, err)

	var parsed TTSRequest
	require.NoError(t, json.Unmarshal(body, &parsed))
	assert.Equal(t, "", parsed.Transcript)
}

func TestBuildTTSRequest_LongTranscript(t *testing.T) {
	cfg := VoiceConfig{APIKey: "k", OutputDir: "/tmp"}
	longText := ""
	for i := 0; i < 500; i++ {
		longText += "This is a long sentence for testing purposes. "
	}

	req := TTSRequest{
		VoiceID:      "v-1",
		Transcript:   longText,
		ModelID:      "sonic-2",
		OutputFormat: DefaultTTSOutputFormat(),
	}

	httpReq, err := BuildTTSRequest(cfg, req)
	require.NoError(t, err)

	body, err := io.ReadAll(httpReq.Body)
	require.NoError(t, err)

	var parsed TTSRequest
	require.NoError(t, json.Unmarshal(body, &parsed))
	assert.Equal(t, longText, parsed.Transcript)
}

func TestBuildCloneRequest_SpecialCharsInName(t *testing.T) {
	cfg := VoiceConfig{APIKey: "k", OutputDir: "/tmp"}
	req := VoiceCloneRequest{
		Name:     `Dr. O'Brien & Associates "Best"`,
		Mode:     "url",
		AudioURL: "https://example.com/voice.mp3",
	}

	httpReq, err := BuildCloneRequest(cfg, req)
	require.NoError(t, err)

	body, err := io.ReadAll(httpReq.Body)
	require.NoError(t, err)

	var parsed VoiceCloneRequest
	require.NoError(t, json.Unmarshal(body, &parsed))
	assert.Equal(t, `Dr. O'Brien & Associates "Best"`, parsed.Name)
}
