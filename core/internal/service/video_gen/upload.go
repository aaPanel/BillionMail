package video_gen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2Config holds settings for Cloudflare R2 uploads.
type R2Config struct {
	AccountID       string // Cloudflare account ID (env: R2_ACCOUNT_ID)
	AccessKeyID     string // R2 access key ID (env: R2_ACCESS_KEY_ID)
	AccessKeySecret string // R2 access key secret (env: R2_ACCESS_KEY_SECRET)
	BucketName      string // R2 bucket name (env: R2_BUCKET_NAME)
	PublicURL       string // Public URL prefix for the bucket (env: R2_PUBLIC_URL)
}

// UploadResult holds the result of an R2 upload.
type UploadResult struct {
	Key       string // object key in R2
	PublicURL string // full public URL to the uploaded file
}

// DefaultR2Config returns config from environment variables.
func DefaultR2Config() R2Config {
	return R2Config{
		AccountID:       os.Getenv("R2_ACCOUNT_ID"),
		AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		AccessKeySecret: os.Getenv("R2_ACCESS_KEY_SECRET"),
		BucketName:      os.Getenv("R2_BUCKET_NAME"),
		PublicURL:        os.Getenv("R2_PUBLIC_URL"),
	}
}

// R2Endpoint returns the S3-compatible endpoint URL for the given account.
func R2Endpoint(accountID string) string {
	return fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)
}

// BuildR2ObjectKey constructs the object key for a video outreach asset.
// Format: video-outreach/{contactID}/{filename}
func BuildR2ObjectKey(contactID, filename string) string {
	return fmt.Sprintf("video-outreach/%s/%s", contactID, filename)
}

// BuildPublicURL constructs the full public URL for an uploaded object.
func BuildPublicURL(cfg R2Config, key string) string {
	base := strings.TrimRight(cfg.PublicURL, "/")
	return fmt.Sprintf("%s/%s", base, key)
}

// NewR2Client creates an S3 client configured for Cloudflare R2.
func NewR2Client(cfg R2Config) *s3.Client {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: R2Endpoint(cfg.AccountID),
			}, nil
		},
	)

	return s3.NewFromConfig(aws.Config{
		Region:                      "auto",
		Credentials:                 credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.AccessKeySecret, ""),
		EndpointResolverWithOptions: r2Resolver,
	}, func(o *s3.Options) {
		o.UsePathStyle = true
	})
}

// UploadFile uploads a local file to R2 and returns the public URL.
func UploadFile(ctx context.Context, cfg R2Config, localPath, contactID string) (*UploadResult, error) {
	f, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("open file for upload: %w", err)
	}
	defer f.Close()

	filename := filepath.Base(localPath)
	key := BuildR2ObjectKey(contactID, filename)
	contentType := detectContentType(filename)

	client := NewR2Client(cfg)
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(cfg.BucketName),
		Key:         aws.String(key),
		Body:        f,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, fmt.Errorf("r2 upload: %w", err)
	}

	return &UploadResult{
		Key:       key,
		PublicURL: BuildPublicURL(cfg, key),
	}, nil
}

// UploadVideoAssets uploads the video and thumbnail to R2.
// Returns public URLs for both.
func UploadVideoAssets(ctx context.Context, cfg R2Config, videoPath, thumbnailPath, contactID string) (videoURL, thumbURL string, err error) {
	videoResult, err := UploadFile(ctx, cfg, videoPath, contactID)
	if err != nil {
		return "", "", fmt.Errorf("upload video: %w", err)
	}

	thumbResult, err := UploadFile(ctx, cfg, thumbnailPath, contactID)
	if err != nil {
		return "", "", fmt.Errorf("upload thumbnail: %w", err)
	}

	return videoResult.PublicURL, thumbResult.PublicURL, nil
}

// detectContentType returns the MIME type based on file extension.
func detectContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".wav":
		return "audio/wav"
	case ".mp3":
		return "audio/mpeg"
	default:
		return "application/octet-stream"
	}
}

// BuildLandingPageURL constructs the landing page URL with video params.
// Falls back to /landing/video if baseURL is empty.
func BuildLandingPageURL(baseURL, videoURL, thumbnailURL, contactName string) string {
	base := strings.TrimRight(baseURL, "/")
	if base == "" {
		base = "/landing/video"
	}
	return fmt.Sprintf("%s?video=%s&thumb=%s&name=%s",
		base, videoURL, thumbnailURL, contactName)
}
