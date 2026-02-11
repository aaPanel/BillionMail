package video_gen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	cartesiaBaseURL    = "https://api.cartesia.ai"
	cartesiaClonePath  = "/voices/clone"
	cartesiaTTSPath    = "/tts/bytes"
	cartesiaAPIVersion = "2024-06-10"
)

// VoiceConfig holds settings for voice operations.
type VoiceConfig struct {
	APIKey     string             // Cartesia API key (env: CARTESIA_API_KEY)
	OutputDir  string             // directory for audio output
	BaseURL    string             // optional base URL override (for testing)
	HTTPClient *RateLimitedClient // optional rate-limited HTTP client
}

// doHTTP executes an HTTP request using the rate-limited client if configured,
// otherwise falls back to http.DefaultClient.
func (cfg VoiceConfig) doHTTP(req *http.Request) (*http.Response, error) {
	if cfg.HTTPClient != nil {
		return cfg.HTTPClient.Do(req)
	}
	return http.DefaultClient.Do(req)
}

// voiceBaseURL returns the configured or default Cartesia base URL.
func (cfg VoiceConfig) voiceBaseURL() string {
	if cfg.BaseURL != "" {
		return cfg.BaseURL
	}
	return cartesiaBaseURL
}

// VoiceCloneRequest is the payload for Cartesia voice cloning.
type VoiceCloneRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Mode        string `json:"mode"`     // "url" or "clip"
	AudioURL    string `json:"clip"`     // URL to source audio
	Language    string `json:"language"` // e.g. "en"
}

// VoiceCloneResponse holds the cloned voice ID from Cartesia.
type VoiceCloneResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TTSRequest is the payload for Cartesia text-to-speech.
type TTSRequest struct {
	VoiceID      string          `json:"voice_id"`
	Transcript   string          `json:"transcript"`
	ModelID      string          `json:"model_id"`
	OutputFormat TTSOutputFormat `json:"output_format"`
	Language     string          `json:"language"`
}

// TTSOutputFormat specifies the audio output format.
type TTSOutputFormat struct {
	Container  string `json:"container"`   // "wav", "mp3", "raw"
	SampleRate int    `json:"sample_rate"` // e.g. 44100
	Encoding   string `json:"encoding"`    // "pcm_f32le", "pcm_s16le", "pcm_mulaw", "pcm_alaw"
}

// DefaultVoiceConfig returns config with API key from env.
func DefaultVoiceConfig(outputDir string) VoiceConfig {
	return VoiceConfig{
		APIKey:    os.Getenv("CARTESIA_API_KEY"),
		OutputDir: outputDir,
	}
}

// DefaultTTSOutputFormat returns WAV 44.1kHz PCM format.
func DefaultTTSOutputFormat() TTSOutputFormat {
	return TTSOutputFormat{
		Container:  "wav",
		SampleRate: 44100,
		Encoding:   "pcm_f32le",
	}
}

// BuildCloneRequest constructs the HTTP request for voice cloning.
// Exported for testing without making API calls.
func BuildCloneRequest(cfg VoiceConfig, req VoiceCloneRequest) (*http.Request, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal clone request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", cfg.voiceBaseURL()+cartesiaClonePath, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create clone request: %w", err)
	}

	httpReq.Header.Set("X-API-Key", cfg.APIKey)
	httpReq.Header.Set("Cartesia-Version", cartesiaAPIVersion)
	httpReq.Header.Set("Content-Type", "application/json")
	return httpReq, nil
}

// BuildTTSRequest constructs the HTTP request for text-to-speech.
// Exported for testing without making API calls.
func BuildTTSRequest(cfg VoiceConfig, req TTSRequest) (*http.Request, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal TTS request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", cfg.voiceBaseURL()+cartesiaTTSPath, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create TTS request: %w", err)
	}

	httpReq.Header.Set("X-API-Key", cfg.APIKey)
	httpReq.Header.Set("Cartesia-Version", cartesiaAPIVersion)
	httpReq.Header.Set("Content-Type", "application/json")
	return httpReq, nil
}

// CloneVoice creates a cloned voice from an audio sample URL via Cartesia API.
func CloneVoice(ctx context.Context, cfg VoiceConfig, name, audioURL string) (*VoiceCloneResponse, error) {
	req := VoiceCloneRequest{
		Name:        name,
		Description: fmt.Sprintf("Cloned voice for %s", name),
		Mode:        "url",
		AudioURL:    audioURL,
		Language:    "en",
	}

	httpReq, err := BuildCloneRequest(cfg, req)
	if err != nil {
		return nil, err
	}
	httpReq = httpReq.WithContext(ctx)

	resp, err := cfg.doHTTP(httpReq)
	if err != nil {
		return nil, fmt.Errorf("cartesia clone API call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cartesia clone API error %d: %s", resp.StatusCode, string(body))
	}

	var result VoiceCloneResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode clone response: %w", err)
	}
	return &result, nil
}

// TextToSpeech generates audio from text using a Cartesia voice.
// Returns the path to the output WAV file.
func TextToSpeech(ctx context.Context, cfg VoiceConfig, voiceID, transcript, filename string) (string, error) {
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}

	req := TTSRequest{
		VoiceID:      voiceID,
		Transcript:   transcript,
		ModelID:      "sonic-2",
		OutputFormat: DefaultTTSOutputFormat(),
		Language:     "en",
	}

	httpReq, err := BuildTTSRequest(cfg, req)
	if err != nil {
		return "", err
	}
	httpReq = httpReq.WithContext(ctx)

	resp, err := cfg.doHTTP(httpReq)
	if err != nil {
		return "", fmt.Errorf("cartesia TTS API call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("cartesia TTS API error %d: %s", resp.StatusCode, string(body))
	}

	outPath := filepath.Join(cfg.OutputDir, filename)
	f, err := os.Create(outPath)
	if err != nil {
		return "", fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("write audio data: %w", err)
	}

	return outPath, nil
}
