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
	lipSyncBaseURL = "https://api.synclabs.so"
	lipSyncPath    = "/lipsync"
)

// LipSyncConfig holds settings for the lip sync API.
type LipSyncConfig struct {
	APIKey     string             // Sync Labs API key (env: SYNCLABS_API_KEY)
	OutputDir  string             // directory to save output video
	BaseURL    string             // optional base URL override (for testing)
	HTTPClient *RateLimitedClient // optional rate-limited HTTP client
}

// doHTTP executes an HTTP request using the rate-limited client if configured,
// otherwise falls back to http.DefaultClient.
func (cfg LipSyncConfig) doHTTP(req *http.Request) (*http.Response, error) {
	if cfg.HTTPClient != nil {
		return cfg.HTTPClient.Do(req)
	}
	return http.DefaultClient.Do(req)
}

// lipSyncBase returns the configured or default Sync Labs base URL.
func (cfg LipSyncConfig) lipSyncBase() string {
	if cfg.BaseURL != "" {
		return cfg.BaseURL
	}
	return lipSyncBaseURL
}

// LipSyncRequest is the payload for the lip sync API.
type LipSyncRequest struct {
	AudioURL       string `json:"audio_url"`       // URL to the narration audio
	VideoURL       string `json:"video_url"`       // URL to the base video (talking head template)
	Model          string `json:"model"`           // model to use (default: "sync-1.7.1-beta")
	SynergizeAudio bool   `json:"synergize_audio"` // enhance audio sync
}

// LipSyncResponse holds the API response.
type LipSyncResponse struct {
	ID       string `json:"id"`
	Status   string `json:"status"`    // "pending", "processing", "completed", "failed"
	VideoURL string `json:"video_url"` // URL to the output video (when completed)
	Error    string `json:"error"`
}

// DefaultLipSyncConfig returns config with API key from env.
func DefaultLipSyncConfig(outputDir string) LipSyncConfig {
	return LipSyncConfig{
		APIKey:    os.Getenv("SYNCLABS_API_KEY"),
		OutputDir: outputDir,
	}
}

// BuildLipSyncRequest constructs the HTTP request for lip sync generation.
// Exported for testing without making API calls.
func BuildLipSyncRequest(cfg LipSyncConfig, req LipSyncRequest) (*http.Request, error) {
	if req.Model == "" {
		req.Model = "sync-1.7.1-beta"
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal lipsync request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", cfg.lipSyncBase()+lipSyncPath, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create lipsync request: %w", err)
	}

	httpReq.Header.Set("x-api-key", cfg.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")
	return httpReq, nil
}

// BuildLipSyncStatusRequest constructs the HTTP request to check job status.
// Exported for testing.
func BuildLipSyncStatusRequest(cfg LipSyncConfig, jobID string) (*http.Request, error) {
	url := fmt.Sprintf("%s%s/%s", cfg.lipSyncBase(), lipSyncPath, jobID)
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create status request: %w", err)
	}

	httpReq.Header.Set("x-api-key", cfg.APIKey)
	return httpReq, nil
}

// SubmitLipSync submits a lip sync job and returns the job ID.
func SubmitLipSync(ctx context.Context, cfg LipSyncConfig, audioURL, videoURL string) (string, error) {
	req := LipSyncRequest{
		AudioURL:       audioURL,
		VideoURL:       videoURL,
		SynergizeAudio: true,
	}

	httpReq, err := BuildLipSyncRequest(cfg, req)
	if err != nil {
		return "", err
	}
	httpReq = httpReq.WithContext(ctx)

	resp, err := cfg.doHTTP(httpReq)
	if err != nil {
		return "", fmt.Errorf("lipsync API call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("lipsync API error %d: %s", resp.StatusCode, string(body))
	}

	var result LipSyncResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode lipsync response: %w", err)
	}
	return result.ID, nil
}

// CheckLipSyncStatus checks the status of a lip sync job.
func CheckLipSyncStatus(ctx context.Context, cfg LipSyncConfig, jobID string) (*LipSyncResponse, error) {
	httpReq, err := BuildLipSyncStatusRequest(cfg, jobID)
	if err != nil {
		return nil, err
	}
	httpReq = httpReq.WithContext(ctx)

	resp, err := cfg.doHTTP(httpReq)
	if err != nil {
		return nil, fmt.Errorf("lipsync status API call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("lipsync status API error %d: %s", resp.StatusCode, string(body))
	}

	var result LipSyncResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode lipsync status: %w", err)
	}
	return &result, nil
}

// DownloadLipSyncVideo downloads the completed lip sync video to the output directory.
func DownloadLipSyncVideo(ctx context.Context, cfg LipSyncConfig, videoURL, filename string) (string, error) {
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", videoURL, nil)
	if err != nil {
		return "", fmt.Errorf("create download request: %w", err)
	}

	resp, err := cfg.doHTTP(req)
	if err != nil {
		return "", fmt.Errorf("download lipsync video: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download error %d", resp.StatusCode)
	}

	outPath := filepath.Join(cfg.OutputDir, filename)
	f, err := os.Create(outPath)
	if err != nil {
		return "", fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("write video data: %w", err)
	}

	return outPath, nil
}
