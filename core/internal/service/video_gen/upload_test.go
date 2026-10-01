package video_gen

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- DefaultR2Config ---

func TestDefaultR2Config(t *testing.T) {
	cfg := DefaultR2Config()
	assert.IsType(t, R2Config{}, cfg)
}

func TestDefaultR2Config_ReadsEnvVars(t *testing.T) {
	t.Setenv("R2_ACCOUNT_ID", "test-acct")
	t.Setenv("R2_ACCESS_KEY_ID", "test-key")
	t.Setenv("R2_ACCESS_KEY_SECRET", "test-secret")
	t.Setenv("R2_BUCKET_NAME", "test-bucket")
	t.Setenv("R2_PUBLIC_URL", "https://cdn.test.com")

	cfg := DefaultR2Config()
	assert.Equal(t, "test-acct", cfg.AccountID)
	assert.Equal(t, "test-key", cfg.AccessKeyID)
	assert.Equal(t, "test-secret", cfg.AccessKeySecret)
	assert.Equal(t, "test-bucket", cfg.BucketName)
	assert.Equal(t, "https://cdn.test.com", cfg.PublicURL)
}

func TestDefaultR2Config_EmptyWhenNoEnv(t *testing.T) {
	t.Setenv("R2_ACCOUNT_ID", "")
	t.Setenv("R2_ACCESS_KEY_ID", "")
	t.Setenv("R2_ACCESS_KEY_SECRET", "")
	t.Setenv("R2_BUCKET_NAME", "")
	t.Setenv("R2_PUBLIC_URL", "")

	cfg := DefaultR2Config()
	assert.Empty(t, cfg.AccountID)
	assert.Empty(t, cfg.AccessKeyID)
	assert.Empty(t, cfg.AccessKeySecret)
	assert.Empty(t, cfg.BucketName)
	assert.Empty(t, cfg.PublicURL)
}

// --- R2Endpoint ---

func TestR2Endpoint(t *testing.T) {
	tests := []struct {
		name      string
		accountID string
		want      string
	}{
		{"standard", "abc123", "https://abc123.r2.cloudflarestorage.com"},
		{"empty", "", "https://.r2.cloudflarestorage.com"},
		{"with hyphens", "my-account-id", "https://my-account-id.r2.cloudflarestorage.com"},
		{"numeric", "123456", "https://123456.r2.cloudflarestorage.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, R2Endpoint(tt.accountID))
		})
	}
}

// --- BuildR2ObjectKey ---

func TestBuildR2ObjectKey(t *testing.T) {
	tests := []struct {
		name      string
		contactID string
		filename  string
		want      string
	}{
		{"video", "contact-123", "video.mp4", "video-outreach/contact-123/video.mp4"},
		{"thumbnail", "contact-456", "thumb.png", "video-outreach/contact-456/thumb.png"},
		{"nested", "abc", "subdir/file.mp4", "video-outreach/abc/subdir/file.mp4"},
		{"empty contact ID", "", "video.mp4", "video-outreach//video.mp4"},
		{"empty filename", "contact-1", "", "video-outreach/contact-1/"},
		{"both empty", "", "", "video-outreach//"},
		{"special chars", "contact@123", "file (1).mp4", "video-outreach/contact@123/file (1).mp4"},
		{"unicode", "контакт", "видео.mp4", "video-outreach/контакт/видео.mp4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, BuildR2ObjectKey(tt.contactID, tt.filename))
		})
	}
}

// --- BuildPublicURL ---

func TestBuildPublicURL(t *testing.T) {
	tests := []struct {
		name string
		cfg  R2Config
		key  string
		want string
	}{
		{
			"trailing slash",
			R2Config{PublicURL: "https://cdn.example.com/"},
			"video-outreach/c1/video.mp4",
			"https://cdn.example.com/video-outreach/c1/video.mp4",
		},
		{
			"no trailing slash",
			R2Config{PublicURL: "https://cdn.example.com"},
			"video-outreach/c1/thumb.png",
			"https://cdn.example.com/video-outreach/c1/thumb.png",
		},
		{
			"multiple trailing slashes",
			R2Config{PublicURL: "https://cdn.example.com///"},
			"key",
			"https://cdn.example.com/key",
		},
		{
			"empty public URL",
			R2Config{PublicURL: ""},
			"video-outreach/c1/video.mp4",
			"/video-outreach/c1/video.mp4",
		},
		{
			"empty key",
			R2Config{PublicURL: "https://cdn.example.com"},
			"",
			"https://cdn.example.com/",
		},
		{
			"both empty",
			R2Config{PublicURL: ""},
			"",
			"/",
		},
		{
			"with path prefix",
			R2Config{PublicURL: "https://cdn.example.com/assets"},
			"video-outreach/c1/video.mp4",
			"https://cdn.example.com/assets/video-outreach/c1/video.mp4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, BuildPublicURL(tt.cfg, tt.key))
		})
	}
}

// --- detectContentType ---

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"video.mp4", "video/mp4"},
		{"video.webm", "video/webm"},
		{"image.png", "image/png"},
		{"image.jpg", "image/jpeg"},
		{"image.jpeg", "image/jpeg"},
		{"image.gif", "image/gif"},
		{"audio.wav", "audio/wav"},
		{"audio.mp3", "audio/mpeg"},
		{"unknown.xyz", "application/octet-stream"},
		{"VIDEO.MP4", "video/mp4"},        // case insensitive
		{"Image.PNG", "image/png"},         // case insensitive
		{"file.JPEG", "image/jpeg"},        // case insensitive
		{"no_extension", "application/octet-stream"},
		{"", "application/octet-stream"},   // empty filename
		{".mp4", "video/mp4"},             // only extension
		{"path/to/video.mp4", "video/mp4"}, // with path
		{"file.tar.gz", "application/octet-stream"}, // compound extension
		{"file.mp4.bak", "application/octet-stream"}, // non-media final ext
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			assert.Equal(t, tt.want, detectContentType(tt.filename))
		})
	}
}

// --- BuildLandingPageURL ---

func TestBuildLandingPageURL(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		videoURL     string
		thumbnailURL string
		contactName  string
		want         string
	}{
		{
			"standard",
			"https://douro-digital-agency.vercel.app/book-a-call",
			"https://cdn.example.com/video.mp4",
			"https://cdn.example.com/thumb.png",
			"John",
			"https://douro-digital-agency.vercel.app/book-a-call?video=https://cdn.example.com/video.mp4&thumb=https://cdn.example.com/thumb.png&name=John",
		},
		{
			"trailing slash",
			"https://example.com/page/",
			"https://cdn.example.com/v.mp4",
			"https://cdn.example.com/t.png",
			"Jane",
			"https://example.com/page?video=https://cdn.example.com/v.mp4&thumb=https://cdn.example.com/t.png&name=Jane",
		},
		{
			"empty name",
			"https://example.com",
			"https://cdn.example.com/v.mp4",
			"https://cdn.example.com/t.png",
			"",
			"https://example.com?video=https://cdn.example.com/v.mp4&thumb=https://cdn.example.com/t.png&name=",
		},
		{
			"all empty",
			"",
			"",
			"",
			"",
			"/landing/video?video=&thumb=&name=",
		},
		{
			"name with spaces",
			"https://example.com",
			"https://cdn.example.com/v.mp4",
			"https://cdn.example.com/t.png",
			"John Doe",
			"https://example.com?video=https://cdn.example.com/v.mp4&thumb=https://cdn.example.com/t.png&name=John Doe",
		},
		{
			"name with special chars",
			"https://example.com",
			"https://cdn.example.com/v.mp4",
			"https://cdn.example.com/t.png",
			"José García",
			"https://example.com?video=https://cdn.example.com/v.mp4&thumb=https://cdn.example.com/t.png&name=José García",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildLandingPageURL(tt.baseURL, tt.videoURL, tt.thumbnailURL, tt.contactName)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- NewR2Client ---

func TestNewR2Client_NotNil(t *testing.T) {
	cfg := R2Config{
		AccountID:       "test-account",
		AccessKeyID:     "test-key",
		AccessKeySecret: "test-secret",
		BucketName:      "test-bucket",
		PublicURL:       "https://cdn.example.com",
	}
	client := NewR2Client(cfg)
	assert.NotNil(t, client)
}

func TestNewR2Client_EmptyConfig(t *testing.T) {
	client := NewR2Client(R2Config{})
	assert.NotNil(t, client, "should create client even with empty config")
}

// --- UploadFile error paths ---

func TestUploadFile_NonexistentFile(t *testing.T) {
	cfg := R2Config{
		AccountID:       "test",
		AccessKeyID:     "key",
		AccessKeySecret: "secret",
		BucketName:      "bucket",
		PublicURL:       "https://cdn.test.com",
	}

	result, err := UploadFile(context.Background(), cfg, "/nonexistent/path/video.mp4", "contact-1")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "open file for upload")
}

func TestUploadFile_EmptyPath(t *testing.T) {
	cfg := R2Config{
		AccountID:       "test",
		AccessKeyID:     "key",
		AccessKeySecret: "secret",
		BucketName:      "bucket",
		PublicURL:       "https://cdn.test.com",
	}

	result, err := UploadFile(context.Background(), cfg, "", "contact-1")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUploadVideoAssets_FirstFileMissing(t *testing.T) {
	cfg := R2Config{
		AccountID:       "test",
		AccessKeyID:     "key",
		AccessKeySecret: "secret",
		BucketName:      "bucket",
		PublicURL:       "https://cdn.test.com",
	}

	videoURL, thumbURL, err := UploadVideoAssets(
		context.Background(), cfg,
		"/nonexistent/video.mp4", "/nonexistent/thumb.png", "contact-1",
	)
	assert.Error(t, err)
	assert.Empty(t, videoURL)
	assert.Empty(t, thumbURL)
	assert.Contains(t, err.Error(), "upload video")
}

func TestUploadVideoAssets_SecondFileMissing(t *testing.T) {
	// Create a temp file for video but not thumbnail
	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "video.mp4")
	require.NoError(t, os.WriteFile(videoPath, []byte("fake video data"), 0644))

	cfg := R2Config{
		AccountID:       "test",
		AccessKeyID:     "key",
		AccessKeySecret: "secret",
		BucketName:      "bucket",
		PublicURL:       "https://cdn.test.com",
	}

	// This will fail on the R2 API call (no real server), but tests the flow
	_, _, err := UploadVideoAssets(
		context.Background(), cfg,
		videoPath, "/nonexistent/thumb.png", "contact-1",
	)
	assert.Error(t, err)
}

func TestUploadFile_CancelledContext(t *testing.T) {
	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "video.mp4")
	require.NoError(t, os.WriteFile(videoPath, []byte("fake"), 0644))

	cfg := R2Config{
		AccountID:       "test",
		AccessKeyID:     "key",
		AccessKeySecret: "secret",
		BucketName:      "bucket",
		PublicURL:       "https://cdn.test.com",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := UploadFile(ctx, cfg, videoPath, "contact-1")
	assert.Error(t, err)
}

// --- UploadResult struct ---

func TestUploadResult_ZeroValue(t *testing.T) {
	var result UploadResult
	assert.Empty(t, result.Key)
	assert.Empty(t, result.PublicURL)
}

// --- R2Config struct ---

func TestR2Config_ZeroValue(t *testing.T) {
	var cfg R2Config
	assert.Empty(t, cfg.AccountID)
	assert.Empty(t, cfg.AccessKeyID)
	assert.Empty(t, cfg.AccessKeySecret)
	assert.Empty(t, cfg.BucketName)
	assert.Empty(t, cfg.PublicURL)
}
