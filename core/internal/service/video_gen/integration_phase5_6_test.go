package video_gen

import (
	"context"
	"strings"
	"testing"

	"billionmail-core/internal/service/lead_scoring"
	vo "billionmail-core/internal/service/video_outreach"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests for Phase 5 (R2 Upload + Thumbnail) and Phase 6 (BillionMail Integration).
// Tests the full pipeline without external deps (no R2, no ImageMagick, no DB).

// --- Scoring → Template Selection → Upload URL Pipeline ---

func TestIntegration_Tier1_FullVideoOutreachPipeline(t *testing.T) {
	// Simulate a tier_1 contact with all enrichment data
	enriched := map[string]string{
		"business_name":         "Bright Horizon Dental",
		"website_url":           "https://brighthorizon.com",
		"multiple_providers":    "true",
		"running_ads":           "1",
		"review_count":          "142",
		"premium_services":      "yes",
		"no_live_chat":          "1",
		"voicemail_after_hours": "true",
		"no_online_booking":     "1",
		"owner_email_found":     "true",
	}

	// Step 1: Score → tier_1
	data := lead_scoring.LeadDataFromAttribs(enriched)
	result := lead_scoring.Score(data)
	require.Equal(t, 1, result.Tier)

	// Step 2: Merge score attribs into contact
	// Note: ToAttribs() stores numeric tier ("1"), but the video_outreach
	// service expects tag strings ("tier_1"). Use result.Tag for lead_tier.
	scoreAttribs := result.ToAttribs()
	for k, v := range scoreAttribs {
		enriched[k] = v
	}
	// Override with tag string (real pipeline uses result.Tag for this field)
	enriched[lead_scoring.AttrLeadTier] = result.Tag
	assert.Equal(t, "tier_1", enriched[lead_scoring.AttrLeadTier])

	// Step 3: Template selection → video
	cfg := vo.VideoOutreachConfig{VideoTemplateID: 10, TextTemplateID: 20}
	sel := vo.SelectTemplate(cfg, enriched)
	assert.Equal(t, "video", sel.Type)
	assert.Equal(t, 10, sel.TemplateID)

	// Step 4: Simulate video generation → add asset URLs to attribs
	r2Cfg := R2Config{
		AccountID: "acct123",
		PublicURL: "https://cdn.example.com",
	}

	videoKey := BuildR2ObjectKey("contact-42", "video.mp4")
	thumbKey := BuildR2ObjectKey("contact-42", "thumbnail.png")
	videoURL := BuildPublicURL(r2Cfg, videoKey)
	thumbURL := BuildPublicURL(r2Cfg, thumbKey)

	enriched[vo.AttrVideoURL] = videoURL
	enriched[vo.AttrThumbnailURL] = thumbURL

	// Step 5: Build landing page URL
	landingURL := BuildLandingPageURL(
		"https://douro-digital-agency.vercel.app/book-a-call",
		videoURL, thumbURL, "Dr. Bright",
	)
	enriched[vo.AttrLandingPageURL] = landingURL

	// Step 6: Verify the contact is now video-eligible
	assert.True(t, vo.IsVideoEligible(enriched))
	assert.True(t, vo.HasVideoAssets(enriched))

	// Step 7: Verify URLs are well-formed
	assert.Contains(t, videoURL, "video-outreach/contact-42/video.mp4")
	assert.Contains(t, thumbURL, "video-outreach/contact-42/thumbnail.png")
	assert.Contains(t, landingURL, "video=")
	assert.Contains(t, landingURL, "thumb=")
	assert.Contains(t, landingURL, "name=Dr. Bright")
}

func TestIntegration_Tier2_TextOnlyPipeline(t *testing.T) {
	// Tier 2 lead: gets text email, no video
	enriched := map[string]string{
		"business_name": "Small Clinic",
		"running_ads":   "1",
		"no_live_chat":  "1",
		"review_count":  "30",
	}

	data := lead_scoring.LeadDataFromAttribs(enriched)
	result := lead_scoring.Score(data)
	assert.Equal(t, 2, result.Tier)

	scoreAttribs := result.ToAttribs()
	for k, v := range scoreAttribs {
		enriched[k] = v
	}
	enriched[lead_scoring.AttrLeadTier] = result.Tag

	// Template selection → text
	cfg := vo.VideoOutreachConfig{VideoTemplateID: 10, TextTemplateID: 20}
	sel := vo.SelectTemplate(cfg, enriched)
	assert.Equal(t, "text", sel.Type)
	assert.Equal(t, 20, sel.TemplateID)

	// NOT video-eligible (no video assets)
	assert.False(t, vo.IsVideoEligible(enriched))
	assert.False(t, vo.HasVideoAssets(enriched))
}

func TestIntegration_BelowThreshold_SkipPipeline(t *testing.T) {
	// Below scoring threshold: should skip
	enriched := map[string]string{
		"active_on_social": "1",
		"review_count":     "5",
	}

	data := lead_scoring.LeadDataFromAttribs(enriched)
	result := lead_scoring.Score(data)
	assert.Equal(t, 3, result.Tier)

	scoreAttribs := result.ToAttribs()
	for k, v := range scoreAttribs {
		enriched[k] = v
	}
	enriched[lead_scoring.AttrLeadTier] = result.Tag // "" for tier 3

	cfg := vo.VideoOutreachConfig{VideoTemplateID: 10, TextTemplateID: 20}
	sel := vo.SelectTemplate(cfg, enriched)
	assert.Equal(t, "skip", sel.Type)
	assert.Zero(t, sel.TemplateID)
}

// --- Thumbnail + Upload URL Building ---

func TestIntegration_ThumbnailConfig_FeedsIntoUploadKey(t *testing.T) {
	// Thumbnail generation produces a file, upload builds key from it
	thumbCfg := DefaultThumbnailConfig("/tmp/screenshots/homepage.png", "/tmp/output")
	assert.Equal(t, "/tmp/output/thumbnail.png", thumbCfg.OutputPath)

	// Upload key from thumbnail output path
	key := BuildR2ObjectKey("contact-99", "thumbnail.png")
	assert.Equal(t, "video-outreach/contact-99/thumbnail.png", key)

	r2Cfg := R2Config{PublicURL: "https://cdn.example.com"}
	url := BuildPublicURL(r2Cfg, key)
	assert.Equal(t, "https://cdn.example.com/video-outreach/contact-99/thumbnail.png", url)
}

func TestIntegration_ThumbnailArgs_ValidForImageMagick(t *testing.T) {
	cfg := DefaultThumbnailConfig("/tmp/input.png", "/tmp/output")
	args := BuildThumbnailArgs(cfg)

	// First arg is input, last is output
	assert.Equal(t, "/tmp/input.png", args[0])
	assert.Equal(t, "/tmp/output/thumbnail.png", args[len(args)-1])

	// Contains -resize with valid geometry
	joined := strings.Join(args, " ")
	assert.Contains(t, joined, "-resize")
	assert.Contains(t, joined, "640x360!")
	assert.Contains(t, joined, "-quality")
	assert.Contains(t, joined, "90")
}

// --- Upload URL Consistency ---

func TestIntegration_UploadURLs_ConsistentFormat(t *testing.T) {
	r2Cfg := R2Config{
		AccountID: "test-acct",
		PublicURL: "https://cdn.test.com",
	}

	contacts := []string{"contact-1", "contact-2", "contact-3"}
	for _, cid := range contacts {
		videoKey := BuildR2ObjectKey(cid, "video.mp4")
		thumbKey := BuildR2ObjectKey(cid, "thumbnail.png")

		videoURL := BuildPublicURL(r2Cfg, videoKey)
		thumbURL := BuildPublicURL(r2Cfg, thumbKey)

		assert.True(t, strings.HasPrefix(videoURL, "https://cdn.test.com/"))
		assert.True(t, strings.HasPrefix(thumbURL, "https://cdn.test.com/"))
		assert.Contains(t, videoURL, cid)
		assert.Contains(t, thumbURL, cid)
		assert.True(t, strings.HasSuffix(videoURL, ".mp4"))
		assert.True(t, strings.HasSuffix(thumbURL, ".png"))
	}
}

// --- Landing Page URL Integration ---

func TestIntegration_LandingPageURL_ContainsAllParams(t *testing.T) {
	r2Cfg := R2Config{PublicURL: "https://cdn.test.com"}

	videoURL := BuildPublicURL(r2Cfg, BuildR2ObjectKey("c1", "video.mp4"))
	thumbURL := BuildPublicURL(r2Cfg, BuildR2ObjectKey("c1", "thumb.png"))

	landingURL := BuildLandingPageURL(
		"https://douro-digital-agency.vercel.app/book-a-call",
		videoURL, thumbURL, "Dr. Smith",
	)

	assert.Contains(t, landingURL, "douro-digital-agency.vercel.app")
	assert.Contains(t, landingURL, "video=https://cdn.test.com/video-outreach/c1/video.mp4")
	assert.Contains(t, landingURL, "thumb=https://cdn.test.com/video-outreach/c1/thumb.png")
	assert.Contains(t, landingURL, "name=Dr. Smith")
}

// --- UploadFile Error Handling ---

func TestIntegration_UploadFile_NonexistentPath_ReturnsWrappedError(t *testing.T) {
	cfg := R2Config{
		AccountID:       "test",
		AccessKeyID:     "key",
		AccessKeySecret: "secret",
		BucketName:      "bucket",
		PublicURL:       "https://cdn.test.com",
	}

	_, err := UploadFile(context.Background(), cfg, "/nonexistent/video.mp4", "c1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "open file for upload")
}

func TestIntegration_UploadVideoAssets_BothMissing(t *testing.T) {
	cfg := R2Config{
		AccountID:       "test",
		AccessKeyID:     "key",
		AccessKeySecret: "secret",
		BucketName:      "bucket",
		PublicURL:       "https://cdn.test.com",
	}

	videoURL, thumbURL, err := UploadVideoAssets(
		context.Background(), cfg,
		"/nonexistent/video.mp4", "/nonexistent/thumb.png", "c1",
	)
	assert.Error(t, err)
	assert.Empty(t, videoURL)
	assert.Empty(t, thumbURL)
}

// --- Attrib Key Consistency ---

func TestIntegration_AttribKeys_MatchBetweenServices(t *testing.T) {
	// video_outreach constants must match what template_render.go expects
	assert.Equal(t, "video_url", vo.AttrVideoURL)
	assert.Equal(t, "thumbnail_url", vo.AttrThumbnailURL)
	assert.Equal(t, "landing_page_url", vo.AttrLandingPageURL)

	// lead_scoring constants must match what video_outreach uses
	assert.Equal(t, "lead_tier", lead_scoring.AttrLeadTier)
	assert.Equal(t, "lead_score", lead_scoring.AttrLeadScore)
}

// --- Empty/Null Safety ---

func TestIntegration_EmptyAttribs_NoVideoAssets(t *testing.T) {
	attribs := map[string]string{}

	assert.False(t, vo.HasVideoAssets(attribs))
	assert.False(t, vo.IsVideoEligible(attribs))

	cfg := vo.VideoOutreachConfig{VideoTemplateID: 10, TextTemplateID: 20}
	sel := vo.SelectTemplate(cfg, attribs)
	assert.Equal(t, "skip", sel.Type)
}

func TestIntegration_Tier1_BeforeVideoGeneration(t *testing.T) {
	// Contact is tier_1 but video hasn't been generated yet
	attribs := map[string]string{
		"lead_tier":  "tier_1",
		"lead_score": "85",
	}

	// Should select video template
	cfg := vo.VideoOutreachConfig{VideoTemplateID: 10, TextTemplateID: 20}
	sel := vo.SelectTemplate(cfg, attribs)
	assert.Equal(t, "video", sel.Type)

	// But NOT yet eligible (no video assets)
	assert.False(t, vo.IsVideoEligible(attribs))
	assert.False(t, vo.HasVideoAssets(attribs))
}

func TestIntegration_Tier1_AfterVideoGeneration(t *testing.T) {
	// Contact is tier_1 and video has been generated
	attribs := map[string]string{
		"lead_tier":        "tier_1",
		"lead_score":       "85",
		"video_url":        "https://cdn.example.com/video.mp4",
		"thumbnail_url":    "https://cdn.example.com/thumb.png",
		"landing_page_url": "https://example.com/watch",
	}

	assert.True(t, vo.IsVideoEligible(attribs))
	assert.True(t, vo.HasVideoAssets(attribs))
}

// --- Content Type Detection Feeds Into Upload ---

func TestIntegration_ContentType_ForVideoAndThumbnail(t *testing.T) {
	// Video file should be detected as video/mp4
	assert.Equal(t, "video/mp4", detectContentType("video.mp4"))
	assert.Equal(t, "image/png", detectContentType("thumbnail.png"))
	assert.Equal(t, "image/jpeg", detectContentType("thumbnail.jpg"))
}

// --- R2 Client Configuration ---

func TestIntegration_R2Client_EndpointCorrect(t *testing.T) {
	endpoint := R2Endpoint("myaccount123")
	assert.Equal(t, "https://myaccount123.r2.cloudflarestorage.com", endpoint)
	assert.True(t, strings.HasPrefix(endpoint, "https://"))
}

func TestIntegration_R2Config_FromEnv(t *testing.T) {
	t.Setenv("R2_ACCOUNT_ID", "test-acct")
	t.Setenv("R2_ACCESS_KEY_ID", "AKIATEST")
	t.Setenv("R2_ACCESS_KEY_SECRET", "secret123")
	t.Setenv("R2_BUCKET_NAME", "video-assets")
	t.Setenv("R2_PUBLIC_URL", "https://cdn.test.com")

	cfg := DefaultR2Config()

	client := NewR2Client(cfg)
	assert.NotNil(t, client)

	endpoint := R2Endpoint(cfg.AccountID)
	assert.Equal(t, "https://test-acct.r2.cloudflarestorage.com", endpoint)

	key := BuildR2ObjectKey("contact-1", "video.mp4")
	url := BuildPublicURL(cfg, key)
	assert.Equal(t, "https://cdn.test.com/video-outreach/contact-1/video.mp4", url)
}
