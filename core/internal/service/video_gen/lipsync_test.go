package video_gen

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultLipSyncConfig(t *testing.T) {
	cfg := DefaultLipSyncConfig("/tmp/lipsync")

	assert.Equal(t, "/tmp/lipsync", cfg.OutputDir)
}

func TestBuildLipSyncRequest_Headers(t *testing.T) {
	cfg := LipSyncConfig{APIKey: "sync-key-123", OutputDir: "/tmp"}
	req := LipSyncRequest{
		AudioURL: "https://example.com/audio.wav",
		VideoURL: "https://example.com/talking-head.mp4",
	}

	httpReq, err := BuildLipSyncRequest(cfg, req)
	require.NoError(t, err)

	assert.Equal(t, "POST", httpReq.Method)
	assert.Equal(t, lipSyncBaseURL+lipSyncPath, httpReq.URL.String())
	assert.Equal(t, "sync-key-123", httpReq.Header.Get("x-api-key"))
	assert.Equal(t, "application/json", httpReq.Header.Get("Content-Type"))
}

func TestBuildLipSyncRequest_Body(t *testing.T) {
	cfg := LipSyncConfig{APIKey: "k", OutputDir: "/tmp"}
	req := LipSyncRequest{
		AudioURL:       "https://example.com/audio.wav",
		VideoURL:       "https://example.com/video.mp4",
		SynergizeAudio: true,
	}

	httpReq, err := BuildLipSyncRequest(cfg, req)
	require.NoError(t, err)

	body, err := io.ReadAll(httpReq.Body)
	require.NoError(t, err)

	var parsed LipSyncRequest
	require.NoError(t, json.Unmarshal(body, &parsed))

	assert.Equal(t, "https://example.com/audio.wav", parsed.AudioURL)
	assert.Equal(t, "https://example.com/video.mp4", parsed.VideoURL)
	assert.Equal(t, "sync-1.7.1-beta", parsed.Model) // default model applied
	assert.True(t, parsed.SynergizeAudio)
}

func TestBuildLipSyncRequest_CustomModel(t *testing.T) {
	cfg := LipSyncConfig{APIKey: "k", OutputDir: "/tmp"}
	req := LipSyncRequest{
		AudioURL: "https://example.com/audio.wav",
		VideoURL: "https://example.com/video.mp4",
		Model:    "sync-2.0",
	}

	httpReq, err := BuildLipSyncRequest(cfg, req)
	require.NoError(t, err)

	body, err := io.ReadAll(httpReq.Body)
	require.NoError(t, err)

	var parsed LipSyncRequest
	require.NoError(t, json.Unmarshal(body, &parsed))

	assert.Equal(t, "sync-2.0", parsed.Model) // custom model preserved
}

func TestBuildLipSyncRequest_EmptyAPIKey(t *testing.T) {
	cfg := LipSyncConfig{APIKey: "", OutputDir: "/tmp"}
	req := LipSyncRequest{AudioURL: "https://x.com/a.wav", VideoURL: "https://x.com/v.mp4"}

	httpReq, err := BuildLipSyncRequest(cfg, req)
	require.NoError(t, err)

	assert.Equal(t, "", httpReq.Header.Get("x-api-key"))
}

func TestBuildLipSyncStatusRequest_URL(t *testing.T) {
	cfg := LipSyncConfig{APIKey: "key-456", OutputDir: "/tmp"}

	httpReq, err := BuildLipSyncStatusRequest(cfg, "job-abc-123")
	require.NoError(t, err)

	assert.Equal(t, "GET", httpReq.Method)
	assert.Equal(t, lipSyncBaseURL+lipSyncPath+"/job-abc-123", httpReq.URL.String())
	assert.Equal(t, "key-456", httpReq.Header.Get("x-api-key"))
}

func TestBuildLipSyncStatusRequest_EmptyJobID(t *testing.T) {
	cfg := LipSyncConfig{APIKey: "k", OutputDir: "/tmp"}

	httpReq, err := BuildLipSyncStatusRequest(cfg, "")
	require.NoError(t, err)

	// Empty job ID: URL ends with /lipsync/
	assert.Equal(t, lipSyncBaseURL+lipSyncPath+"/", httpReq.URL.String())
}

func TestBuildLipSyncRequest_EmptyURLs(t *testing.T) {
	cfg := LipSyncConfig{APIKey: "k", OutputDir: "/tmp"}
	req := LipSyncRequest{AudioURL: "", VideoURL: ""}

	httpReq, err := BuildLipSyncRequest(cfg, req)
	require.NoError(t, err)

	body, err := io.ReadAll(httpReq.Body)
	require.NoError(t, err)

	var parsed LipSyncRequest
	require.NoError(t, json.Unmarshal(body, &parsed))

	assert.Equal(t, "", parsed.AudioURL)
	assert.Equal(t, "", parsed.VideoURL)
}

func TestBuildLipSyncRequest_SpecialCharsInURL(t *testing.T) {
	cfg := LipSyncConfig{APIKey: "k", OutputDir: "/tmp"}
	req := LipSyncRequest{
		AudioURL: "https://example.com/audio?token=abc&foo=bar%20baz",
		VideoURL: "https://example.com/video#section",
	}

	httpReq, err := BuildLipSyncRequest(cfg, req)
	require.NoError(t, err)

	body, err := io.ReadAll(httpReq.Body)
	require.NoError(t, err)

	var parsed LipSyncRequest
	require.NoError(t, json.Unmarshal(body, &parsed))

	assert.Equal(t, "https://example.com/audio?token=abc&foo=bar%20baz", parsed.AudioURL)
	assert.Equal(t, "https://example.com/video#section", parsed.VideoURL)
}

func TestLipSyncResponse_JSONParsing(t *testing.T) {
	raw := `{"id":"job-123","status":"completed","video_url":"https://cdn.example.com/out.mp4","error":""}`

	var resp LipSyncResponse
	err := json.Unmarshal([]byte(raw), &resp)
	require.NoError(t, err)

	assert.Equal(t, "job-123", resp.ID)
	assert.Equal(t, "completed", resp.Status)
	assert.Equal(t, "https://cdn.example.com/out.mp4", resp.VideoURL)
	assert.Empty(t, resp.Error)
}

func TestLipSyncResponse_ErrorState(t *testing.T) {
	raw := `{"id":"job-456","status":"failed","video_url":"","error":"audio too short"}`

	var resp LipSyncResponse
	err := json.Unmarshal([]byte(raw), &resp)
	require.NoError(t, err)

	assert.Equal(t, "failed", resp.Status)
	assert.Empty(t, resp.VideoURL)
	assert.Equal(t, "audio too short", resp.Error)
}
