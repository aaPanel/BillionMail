package video_gen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Job status constants
const (
	JobPending    = "pending"
	JobProcessing = "processing"
	JobCompleted  = "completed"
	JobFailed     = "failed"
)

// Max concurrent video generation jobs
const maxConcurrentJobs = 3

// Retry config for external API calls
const (
	maxRetries       = 3
	retryBaseDelay   = 5 * time.Second
	retryMultiplier  = 3
	lipSyncPollDelay = 10 * time.Second
	lipSyncTimeout   = 5 * time.Minute
)

// semaphore limits concurrent pipeline goroutines
var jobSem = make(chan struct{}, maxConcurrentJobs)

// tableOnce ensures bm_video_jobs table is created once
var tableOnce sync.Once

// VideoJob represents a row in bm_video_jobs.
type VideoJob struct {
	ID           int       `json:"id"`
	ContactID    int       `json:"contact_id"`
	ContactEmail string    `json:"contact_email"`
	GroupID      int       `json:"group_id"`
	Status       string    `json:"status"`
	Phase        string    `json:"phase"`
	Error        string    `json:"error"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ensureTable creates bm_video_jobs if it doesn't exist.
func ensureTable(ctx context.Context) {
	tableOnce.Do(func() {
		_, err := g.DB().Exec(ctx, `
			CREATE TABLE IF NOT EXISTS bm_video_jobs (
				id SERIAL PRIMARY KEY,
				contact_id INT NOT NULL,
				contact_email VARCHAR(255) NOT NULL,
				group_id INT NOT NULL,
				status VARCHAR(20) DEFAULT 'pending',
				phase VARCHAR(30) DEFAULT '',
				error TEXT DEFAULT '',
				created_at TIMESTAMP DEFAULT NOW(),
				updated_at TIMESTAMP DEFAULT NOW()
			)
		`)
		if err != nil {
			g.Log().Warning(ctx, "create bm_video_jobs table: ", err)
		}
	})
}

// EnqueueVideoJob inserts a pending job into bm_video_jobs.
func EnqueueVideoJob(ctx context.Context, contactID int, email string, groupID int) (int, error) {
	ensureTable(ctx)

	result, err := g.DB().Model("bm_video_jobs").Ctx(ctx).Insert(g.Map{
		"contact_id":    contactID,
		"contact_email": email,
		"group_id":      groupID,
		"status":        JobPending,
	})
	if err != nil {
		return 0, fmt.Errorf("enqueue video job: %w", err)
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

// ProcessVideoJobs polls for pending jobs and launches pipelines.
// Called by gtimer every 30s.
func ProcessVideoJobs(ctx context.Context) {
	ensureTable(ctx)

	var jobs []VideoJob
	err := g.DB().Model("bm_video_jobs").Ctx(ctx).
		Where("status", JobPending).
		OrderAsc("id").
		Limit(maxConcurrentJobs).
		Scan(&jobs)
	if err != nil {
		g.Log().Warning(ctx, "query pending video jobs: ", err)
		return
	}

	for _, job := range jobs {
		job := job
		select {
		case jobSem <- struct{}{}:
			// Mark processing before launching goroutine
			updateJobStatus(ctx, job.ID, JobProcessing, "screenshot", "")
			go func() {
				defer func() { <-jobSem }()
				RunPipeline(ctx, job) //nolint:errcheck // markFailed logs + persists errors
			}()
		default:
			// Semaphore full, will pick up next poll cycle
			return
		}
	}
}

// updateJobStatus updates the job's status, phase, and error fields.
func updateJobStatus(ctx context.Context, jobID int, status, phase, errMsg string) {
	_, err := g.DB().Model("bm_video_jobs").Ctx(ctx).
		Where("id", jobID).
		Update(g.Map{
			"status":     status,
			"phase":      phase,
			"error":      errMsg,
			"updated_at": time.Now(),
		})
	if err != nil {
		g.Log().Warningf(ctx, "update video job %d: %v", jobID, err)
	}
}

// RunPipeline executes all video generation phases sequentially.
func RunPipeline(ctx context.Context, job VideoJob) error {
	g.Log().Infof(ctx, "video job %d starting pipeline for contact %d <%s>", job.ID, job.ContactID, job.ContactEmail)
	tmpDir := fmt.Sprintf("/tmp/video_gen_%d", job.ID)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		markFailed(ctx, job.ID, "setup", err)
		return err
	}

	// 1. Load contact attribs
	attribs, err := loadContactAttribs(ctx, job.ContactID, job.GroupID)
	if err != nil {
		markFailed(ctx, job.ID, "load_contact", err)
		return err
	}

	websiteURL := attribs["website_url"]
	businessName := attribs["business_name"]
	ownerName := attribs["owner_name"]
	if websiteURL == "" {
		err = fmt.Errorf("contact %d missing website_url", job.ContactID)
		markFailed(ctx, job.ID, "load_contact", err)
		return err
	}

	// Parse signals
	signals := parseSignals(attribs["lead_signals"])

	// 2. Screenshots + Annotate
	updateJobStatus(ctx, job.ID, JobProcessing, "screenshot", "")
	screenshots, err := withRetry(func() (*ScreenshotResult, error) {
		return CaptureScreenshots(ctx, DefaultScreenshotConfig(websiteURL, businessName, tmpDir))
	})
	if err != nil {
		markFailed(ctx, job.ID, "screenshot", err)
		return err
	}

	updateJobStatus(ctx, job.ID, JobProcessing, "annotate", "")
	annotated, err := withRetry(func() (*ScreenshotResult, error) {
		return AnnotateAll(ctx, screenshots, signals)
	})
	if err != nil {
		markFailed(ctx, job.ID, "annotate", err)
		return err
	}

	// Init rate-limited HTTP clients for external APIs
	voiceClient := NewRateLimitedClient("cartesia", 5, 10, 30*time.Second)
	lipSyncClient := NewRateLimitedClient("synclabs", 5, 10, 60*time.Second)

	// 3. Script generation
	updateJobStatus(ctx, job.ID, JobProcessing, "script", "")
	scriptCfg := DefaultScriptConfig()
	scriptOut, err := withRetry(func() (*ScriptOutput, error) {
		return GenerateScript(ctx, scriptCfg, ScriptInput{
			BusinessName: businessName,
			OwnerName:    ownerName,
			WebsiteURL:   websiteURL,
			Signals:      signals,
			Attribs:      attribs,
		})
	})
	if err != nil {
		markFailed(ctx, job.ID, "script", err)
		return err
	}

	// 4. Voice clone + TTS
	updateJobStatus(ctx, job.ID, JobProcessing, "voice", "")
	voiceCfg := DefaultVoiceConfig(tmpDir)
	voiceCfg.HTTPClient = voiceClient
	voiceID, err := resolveVoiceID(ctx, voiceCfg)
	if err != nil {
		markFailed(ctx, job.ID, "voice", err)
		return err
	}

	audioPath, err := withRetry(func() (string, error) {
		return TextToSpeech(ctx, voiceCfg, voiceID, scriptOut.Script, "narration.wav")
	})
	if err != nil {
		markFailed(ctx, job.ID, "voice", err)
		return err
	}

	// 5. Composite video
	updateJobStatus(ctx, job.ID, JobProcessing, "composite", "")
	sceneDuration := time.Duration(scriptOut.Duration) * time.Second / 3
	scenes := []Scene{
		{ImagePath: annotated.Homepage, AudioPath: audioPath, Duration: sceneDuration},
		{ImagePath: annotated.Contact, AudioPath: audioPath, Duration: sceneDuration},
		{ImagePath: annotated.Google, AudioPath: audioPath, Duration: sceneDuration},
	}
	compositeOut := filepath.Join(tmpDir, "composite.mp4")
	_, err = withRetry(func() (*CompositeResult, error) {
		return CompositeVideo(ctx, DefaultCompositeConfig(scenes, compositeOut))
	})
	if err != nil {
		markFailed(ctx, job.ID, "composite", err)
		return err
	}

	// 6. Lip sync
	updateJobStatus(ctx, job.ID, JobProcessing, "lipsync", "")
	lipCfg := DefaultLipSyncConfig(tmpDir)
	lipCfg.HTTPClient = lipSyncClient
	finalVideoPath := compositeOut // default to composite if lip sync skipped

	if lipCfg.APIKey != "" {
		lipJobID, lipErr := withRetry(func() (string, error) {
			return SubmitLipSync(ctx, lipCfg, audioPath, compositeOut)
		})
		if lipErr != nil {
			markFailed(ctx, job.ID, "lipsync", lipErr)
			return lipErr
		}

		lipVideoURL, lipErr := pollLipSync(ctx, lipCfg, lipJobID)
		if lipErr != nil {
			markFailed(ctx, job.ID, "lipsync", lipErr)
			return lipErr
		}

		dlPath, lipErr := withRetry(func() (string, error) {
			return DownloadLipSyncVideo(ctx, lipCfg, lipVideoURL, "lipsync.mp4")
		})
		if lipErr != nil {
			markFailed(ctx, job.ID, "lipsync", lipErr)
			return lipErr
		}
		finalVideoPath = dlPath
	}

	// 7. Thumbnail + Upload
	updateJobStatus(ctx, job.ID, JobProcessing, "upload", "")
	thumbCfg := DefaultThumbnailConfig(annotated.Homepage, tmpDir)
	thumbPath, err := withRetry(func() (string, error) {
		return GenerateThumbnail(ctx, thumbCfg)
	})
	if err != nil {
		markFailed(ctx, job.ID, "upload", err)
		return err
	}

	r2Cfg := DefaultR2Config()
	contactIDStr := fmt.Sprintf("%d", job.ContactID)
	videoURL, thumbURL, err := UploadVideoAssets(ctx, r2Cfg, finalVideoPath, thumbPath, contactIDStr)
	if err != nil {
		markFailed(ctx, job.ID, "upload", err)
		return err
	}

	landingPageURL := BuildLandingPageURL(
		os.Getenv("LANDING_PAGE_BASE_URL"),
		videoURL, thumbURL, ownerName,
	)

	// 8. Update contact attribs
	err = updateContactVideoAttribs(ctx, job.ContactID, job.GroupID, videoURL, thumbURL, landingPageURL)
	if err != nil {
		markFailed(ctx, job.ID, "upload", err)
		return err
	}

	// 9. Done
	updateJobStatus(ctx, job.ID, JobCompleted, "upload", "")
	g.Log().Infof(ctx, "video job %d completed: video=%s landing=%s", job.ID, videoURL, landingPageURL)

	// Cleanup temp dir on success
	os.RemoveAll(tmpDir)

	return nil
}

// markFailed sets job status to failed with error message and logs it.
func markFailed(ctx context.Context, jobID int, phase string, err error) {
	g.Log().Errorf(ctx, "video job %d failed at %s: %v", jobID, phase, err)
	updateJobStatus(ctx, jobID, JobFailed, phase, err.Error())
}

// loadContactAttribs fetches attribs for a contact from DB.
func loadContactAttribs(ctx context.Context, contactID, groupID int) (map[string]string, error) {
	val, err := g.DB().Model("bm_contacts").Ctx(ctx).
		Where("id", contactID).
		Where("group_id", groupID).
		Value("attribs")
	if err != nil {
		return nil, fmt.Errorf("load contact attribs: %w", err)
	}
	if val.IsNil() || val.IsEmpty() {
		return nil, fmt.Errorf("contact %d has no attribs", contactID)
	}

	result := make(map[string]string)
	if err := val.Scan(&result); err != nil {
		return nil, fmt.Errorf("parse contact attribs: %w", err)
	}
	return result, nil
}

// updateContactVideoAttribs merges video URLs into Contact.Attribs.
func updateContactVideoAttribs(ctx context.Context, contactID, groupID int, videoURL, thumbURL, landingURL string) error {
	// Load existing attribs
	attribs, err := loadContactAttribs(ctx, contactID, groupID)
	if err != nil {
		return err
	}

	attribs["video_url"] = videoURL
	attribs["thumbnail_url"] = thumbURL
	attribs["landing_page_url"] = landingURL

	_, err = g.DB().Model("bm_contacts").Ctx(ctx).
		Where("id", contactID).
		Where("group_id", groupID).
		Update(g.Map{"attribs": attribs})
	return err
}

// parseSignals converts comma-separated signal string to map.
func parseSignals(s string) map[string]bool {
	m := make(map[string]bool)
	if s == "" {
		return m
	}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			m[part] = true
		}
	}
	return m
}

// resolveVoiceID determines which Cartesia voice ID to use.
// If VOICE_SAMPLE_URL is set, clones a voice. Otherwise uses VOICE_DEFAULT_ID.
func resolveVoiceID(ctx context.Context, cfg VoiceConfig) (string, error) {
	sampleURL := os.Getenv("VOICE_SAMPLE_URL")
	defaultID := os.Getenv("VOICE_DEFAULT_ID")

	if sampleURL != "" {
		resp, err := withRetry(func() (*VoiceCloneResponse, error) {
			return CloneVoice(ctx, cfg, "sender-voice", sampleURL)
		})
		if err != nil {
			return "", fmt.Errorf("voice clone: %w", err)
		}
		return resp.ID, nil
	}

	if defaultID != "" {
		return defaultID, nil
	}

	return "", fmt.Errorf("no voice configured: set VOICE_SAMPLE_URL or VOICE_DEFAULT_ID")
}

// pollLipSync polls lip sync status until completed or timeout.
func pollLipSync(ctx context.Context, cfg LipSyncConfig, jobID string) (string, error) {
	deadline := time.Now().Add(lipSyncTimeout)
	for {
		if time.Now().After(deadline) {
			return "", fmt.Errorf("lip sync timed out after %v", lipSyncTimeout)
		}

		resp, err := CheckLipSyncStatus(ctx, cfg, jobID)
		if err != nil {
			return "", fmt.Errorf("check lip sync status: %w", err)
		}

		switch resp.Status {
		case "completed":
			if resp.VideoURL == "" {
				return "", fmt.Errorf("lip sync completed but no video URL")
			}
			return resp.VideoURL, nil
		case "failed":
			return "", fmt.Errorf("lip sync failed: %s", resp.Error)
		}

		time.Sleep(lipSyncPollDelay)
	}
}

// withRetry retries a function up to maxRetries times with exponential backoff.
func withRetry[T any](fn func() (T, error)) (T, error) {
	var result T
	var err error
	delay := retryBaseDelay

	for attempt := 0; attempt <= maxRetries; attempt++ {
		result, err = fn()
		if err == nil {
			return result, nil
		}
		if attempt < maxRetries {
			time.Sleep(delay)
			delay *= time.Duration(retryMultiplier)
		}
	}
	return result, err
}
