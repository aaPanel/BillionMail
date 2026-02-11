package video_gen

import (
	"strings"
	"testing"

	"billionmail-core/internal/service/lead_scoring"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests: lead scoring signals → screenshot annotations → ImageMagick args.
// Tests the full pipeline without external deps (no Playwright, no ImageMagick, no DB).

func TestIntegration_ScoringSignalsDriveAnnotations(t *testing.T) {
	// Simulate FrostByte enrichment data for a high-value dental practice
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

	// Step 1: Score the lead
	data := lead_scoring.LeadDataFromAttribs(enriched)
	result := lead_scoring.Score(data)

	require.Equal(t, 1, result.Tier, "should be tier 1")

	// Step 2: Convert scoring signals to annotation signal map
	signalMap := make(map[string]bool)
	for _, s := range result.Signals {
		signalMap[s.Name] = true
	}

	// Step 3: Verify annotations are generated for matched signals
	contactAnns := DefaultAnnotations(ScreenshotContact, signalMap)
	assert.NotEmpty(t, contactAnns, "contact page should have annotations for no_chat signal")

	homepageAnns := DefaultAnnotations(ScreenshotHomepage, signalMap)
	assert.NotEmpty(t, homepageAnns, "homepage should have annotations for voicemail signal")

	// Step 4: Verify the annotation args produce valid ImageMagick commands
	for _, ann := range contactAnns {
		cfg := AnnotateConfig{
			InputPath:   "/tmp/contact.png",
			OutputPath:  "/tmp/contact_annotated.png",
			Annotations: []Annotation{ann},
		}
		args := BuildAnnotateArgs(cfg)
		assert.Equal(t, "/tmp/contact.png", args[0])
		assert.Equal(t, "/tmp/contact_annotated.png", args[len(args)-1])
		assert.True(t, len(args) > 2, "should have annotation commands between input/output")
	}
}

func TestIntegration_Tier2LeadMinimalAnnotations(t *testing.T) {
	// Tier 2 lead: has ads + some problem signals but fewer annotation triggers
	enriched := map[string]string{
		"running_ads":  "1",
		"no_live_chat": "1",
		"review_count": "30", // below 100, no high_reviews signal
	}

	data := lead_scoring.LeadDataFromAttribs(enriched)
	result := lead_scoring.Score(data)

	assert.Equal(t, 2, result.Tier)

	signalMap := make(map[string]bool)
	for _, s := range result.Signals {
		signalMap[s.Name] = true
	}

	// no_chat signal should trigger contact page annotations
	contactAnns := DefaultAnnotations(ScreenshotContact, signalMap)
	assert.NotEmpty(t, contactAnns)

	// no voicemail signal → homepage should be empty
	homepageAnns := DefaultAnnotations(ScreenshotHomepage, signalMap)
	assert.Empty(t, homepageAnns)

	// no high_spend_low_rating → google should be empty
	googleAnns := DefaultAnnotations(ScreenshotGoogle, signalMap)
	assert.Empty(t, googleAnns)
}

func TestIntegration_Tier3LeadNoAnnotations(t *testing.T) {
	// Tier 3 lead: minimal data, no problem signals that map to annotations
	enriched := map[string]string{
		"active_on_social": "1",
		"review_count":     "10",
	}

	data := lead_scoring.LeadDataFromAttribs(enriched)
	result := lead_scoring.Score(data)

	assert.Equal(t, 3, result.Tier)

	signalMap := make(map[string]bool)
	for _, s := range result.Signals {
		signalMap[s.Name] = true
	}

	// No visual annotation signals triggered
	for _, st := range []ScreenshotType{ScreenshotHomepage, ScreenshotContact, ScreenshotGoogle} {
		anns := DefaultAnnotations(st, signalMap)
		assert.Empty(t, anns, "tier 3 lead should have no annotations for %s", st)
	}
}

func TestIntegration_FullPipeline_AttribsToAnnotatedArgs(t *testing.T) {
	// End-to-end: raw attribs → score → screenshot config → annotations → magick args
	enriched := map[string]string{
		"business_name":         "Smile Studio",
		"website_url":           "https://smilestudio.com",
		"running_ads":           "1",
		"multiple_providers":    "true",
		"no_live_chat":          "1",
		"voicemail_after_hours": "true",
		"high_spend_low_rating": "true",
		"premium_services":      "true",
		"owner_email_found":     "true",
	}

	// Score
	data := lead_scoring.LeadDataFromAttribs(enriched)
	result := lead_scoring.Score(data)
	require.Equal(t, 1, result.Tier)

	// Screenshot config
	cfg := DefaultScreenshotConfig(
		enriched["website_url"],
		enriched["business_name"],
		"/tmp/prospect_42",
	)
	assert.Equal(t, "https://smilestudio.com", cfg.WebsiteURL)

	// Signal map from scoring
	signalMap := make(map[string]bool)
	for _, s := range result.Signals {
		signalMap[s.Name] = true
	}

	// All 3 screenshot types should have annotations
	screenshotResult := &ScreenshotResult{
		Homepage: "/tmp/prospect_42/homepage.png",
		Contact:  "/tmp/prospect_42/contact.png",
		Google:   "/tmp/prospect_42/google.png",
	}

	// Homepage: voicemail signal
	homeAnns := DefaultAnnotations(ScreenshotHomepage, signalMap)
	require.NotEmpty(t, homeAnns)
	homeArgs := BuildAnnotateArgs(AnnotateConfig{
		InputPath:   screenshotResult.Homepage,
		OutputPath:  annotatedPath(screenshotResult.Homepage),
		Annotations: homeAnns,
	})
	homeJoined := strings.Join(homeArgs, " ")
	assert.Contains(t, homeJoined, "62%")

	// Contact: no_chat signal
	contactAnns := DefaultAnnotations(ScreenshotContact, signalMap)
	require.NotEmpty(t, contactAnns)
	contactArgs := BuildAnnotateArgs(AnnotateConfig{
		InputPath:   screenshotResult.Contact,
		OutputPath:  annotatedPath(screenshotResult.Contact),
		Annotations: contactAnns,
	})
	contactJoined := strings.Join(contactArgs, " ")
	assert.Contains(t, contactJoined, "circle")
	assert.Contains(t, contactJoined, "No live chat")

	// Google: high_spend_low_rating signal
	googleAnns := DefaultAnnotations(ScreenshotGoogle, signalMap)
	require.NotEmpty(t, googleAnns)
	googleArgs := BuildAnnotateArgs(AnnotateConfig{
		InputPath:   screenshotResult.Google,
		OutputPath:  annotatedPath(screenshotResult.Google),
		Annotations: googleAnns,
	})
	googleJoined := strings.Join(googleArgs, " ")
	assert.Contains(t, googleJoined, "ad spend")
}

func TestIntegration_ScoreAttribsMergePreservesEnrichment(t *testing.T) {
	// Verify that scoring attribs don't clobber enrichment data
	enriched := map[string]string{
		"business_name":    "Test Dental",
		"website_url":      "https://test.com",
		"running_ads":      "1",
		"no_live_chat":     "1",
		"review_count":     "50",
		"owner_name":       "Dr. Test",
		"services_offered": "implants,veneers",
	}

	data := lead_scoring.LeadDataFromAttribs(enriched)
	result := lead_scoring.Score(data)
	scoreAttribs := result.ToAttribs()

	// Merge (simulating JSONB ||)
	merged := make(map[string]string)
	for k, v := range enriched {
		merged[k] = v
	}
	for k, v := range scoreAttribs {
		merged[k] = v
	}

	// Original enrichment data preserved
	assert.Equal(t, "Test Dental", merged["business_name"])
	assert.Equal(t, "https://test.com", merged["website_url"])
	assert.Equal(t, "Dr. Test", merged["owner_name"])
	assert.Equal(t, "implants,veneers", merged["services_offered"])

	// Score attribs added
	assert.NotEmpty(t, merged[lead_scoring.AttrLeadScore])
	assert.NotEmpty(t, merged[lead_scoring.AttrLeadTier])
	assert.NotEmpty(t, merged[lead_scoring.AttrLeadSignal])

	// Can re-parse the merged attribs for screenshots
	cfg := DefaultScreenshotConfig(
		merged["website_url"],
		merged["business_name"],
		"/tmp/out",
	)
	assert.Equal(t, "https://test.com", cfg.WebsiteURL)
	assert.Equal(t, "Test Dental", cfg.BusinessName)
}

func TestIntegration_SignalNameConsistency(t *testing.T) {
	// Verify that signal names from Score() match the keys DefaultAnnotations() checks
	annotationSignalKeys := []string{
		"no_chat",
		"no_online_booking",
		"voicemail_after_hrs",
		"high_spend_low_rating",
	}

	// Score with all signals active
	allData := lead_scoring.LeadData{
		MultipleProviders:   true,
		RunningAds:          true,
		ReviewCount:         200,
		PremiumServices:     true,
		AffluentZip:         true,
		NoLiveChat:          true,
		VoicemailAfterHrs:   true,
		NoOnlineBooking:     true,
		SlowFormResponse:    true,
		HighSpendLowRating:  true,
		OwnerEmailFound:     true,
		ActiveOnSocial:      true,
		InIndustryCommunity: true,
	}
	result := lead_scoring.Score(allData)

	scoreSignalNames := make(map[string]bool)
	for _, s := range result.Signals {
		scoreSignalNames[s.Name] = true
	}

	// Every annotation signal key must exist in scoring output
	for _, key := range annotationSignalKeys {
		assert.True(t, scoreSignalNames[key],
			"annotation checks signal %q but Score() produces no signal with that name", key)
	}
}

func TestIntegration_ScreenshotConfigFromAttribs(t *testing.T) {
	// Verify screenshot config can be built from the same attribs used for scoring
	attribs := map[string]string{
		"business_name": "Pacific Dental Arts",
		"website_url":   "https://pacificdentalarts.com",
		"running_ads":   "1",
		"no_live_chat":  "1",
	}

	cfg := DefaultScreenshotConfig(attribs["website_url"], attribs["business_name"], "/tmp/out")

	script := buildPlaywrightScript(cfg, &ScreenshotResult{
		Homepage: "/tmp/out/homepage.png",
		Contact:  "/tmp/out/contact.png",
		Google:   "/tmp/out/google.png",
	})

	assert.Contains(t, script, "pacificdentalarts.com")
	assert.Contains(t, script, "pacificdentalarts.com/contact")
	assert.Contains(t, script, "Pacific+Dental+Arts")
}

func TestIntegration_EmptyAttribsProduceNoAnnotations(t *testing.T) {
	// Empty enrichment → zero score → no annotation signals
	data := lead_scoring.LeadDataFromAttribs(map[string]string{})
	result := lead_scoring.Score(data)

	assert.Equal(t, 0, result.Score)
	assert.Empty(t, result.Signals)

	signalMap := make(map[string]bool)
	for _, s := range result.Signals {
		signalMap[s.Name] = true
	}

	for _, st := range []ScreenshotType{ScreenshotHomepage, ScreenshotContact, ScreenshotGoogle} {
		anns := DefaultAnnotations(st, signalMap)
		assert.Empty(t, anns)
	}
}
