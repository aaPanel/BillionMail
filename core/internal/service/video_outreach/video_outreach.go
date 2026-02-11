package video_outreach

import (
	"billionmail-core/internal/service/lead_scoring"
	"os"
	"strconv"
	"strings"
)

// TemplateSelection holds the chosen template ID and type for a contact.
type TemplateSelection struct {
	TemplateID int
	Type       string // "video" or "text" or "skip"
	Tier       string // "tier_1", "tier_2", or ""
}

// VideoOutreachConfig holds template IDs for each tier.
type VideoOutreachConfig struct {
	VideoTemplateID int // template ID for tier_1 (video email)
	TextTemplateID  int // template ID for tier_2 (text email)
}

// DefaultVideoOutreachConfig returns config from env vars VIDEO_TEMPLATE_ID and TEXT_TEMPLATE_ID.
func DefaultVideoOutreachConfig() VideoOutreachConfig {
	vid, _ := strconv.Atoi(os.Getenv("VIDEO_TEMPLATE_ID"))
	txt, _ := strconv.Atoi(os.Getenv("TEXT_TEMPLATE_ID"))
	return VideoOutreachConfig{
		VideoTemplateID: vid,
		TextTemplateID:  txt,
	}
}

// SelectTemplate picks the right template based on the contact's lead tier.
// Returns the template selection with type and tier info.
func SelectTemplate(cfg VideoOutreachConfig, attribs map[string]string) TemplateSelection {
	tier := attribs[lead_scoring.AttrLeadTier]

	switch tier {
	case lead_scoring.TagTier1:
		return TemplateSelection{
			TemplateID: cfg.VideoTemplateID,
			Type:       "video",
			Tier:       tier,
		}
	case lead_scoring.TagTier2:
		return TemplateSelection{
			TemplateID: cfg.TextTemplateID,
			Type:       "text",
			Tier:       tier,
		}
	default:
		return TemplateSelection{
			Type: "skip",
		}
	}
}

// VideoAttribKeys are the Contact.Attribs keys set by the video pipeline.
const (
	AttrVideoURL       = "video_url"
	AttrThumbnailURL   = "thumbnail_url"
	AttrLandingPageURL = "landing_page_url"
)

// HasVideoAssets checks if a contact has all required video assets in attribs.
func HasVideoAssets(attribs map[string]string) bool {
	return attribs[AttrVideoURL] != "" &&
		attribs[AttrThumbnailURL] != "" &&
		attribs[AttrLandingPageURL] != ""
}

// IsVideoEligible checks if a contact is tier_1 and has video assets ready.
func IsVideoEligible(attribs map[string]string) bool {
	return attribs[lead_scoring.AttrLeadTier] == lead_scoring.TagTier1 && HasVideoAssets(attribs)
}

// SignalCopy maps machine-readable signal names to human-readable pain point
// descriptions suitable for use in tier_2 outreach email copy.
var SignalCopy = map[string]string{
	"multiple_providers":    "managing multiple locations",
	"running_ads":           "investing in paid advertising",
	"high_reviews":          "a strong online reputation",
	"premium_services":      "offering premium services",
	"affluent_zip":          "serving an affluent market",
	"no_chat":               "missing live chat on your website",
	"voicemail_after_hrs":   "calls going to voicemail after hours",
	"no_online_booking":     "no online self-scheduling for clients",
	"slow_form":             "slow response times on contact forms",
	"high_spend_low_rating": "high ad spend with room to improve reviews",
	"owner_email":           "being hands-on with the business",
	"active_social":         "being active on social media",
	"industry_community":    "being involved in industry communities",
}

// SignalsToCopy converts a comma-separated signal string into a slice of
// human-readable descriptions using SignalCopy.
func SignalsToCopy(signals string) []string {
	if signals == "" {
		return nil
	}
	parts := strings.Split(signals, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if copy, ok := SignalCopy[p]; ok {
			result = append(result, copy)
		}
	}
	return result
}
