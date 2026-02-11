package lead_scoring

import (
	"fmt"
	"strconv"
	"strings"
)

// Tier thresholds
const (
	Tier1Threshold = 75 // Video email
	Tier2Threshold = 40 // Text email
	// Below Tier2Threshold = skip
)

// Tier labels applied as contact tags
const (
	TagTier1 = "tier_1"
	TagTier2 = "tier_2"
)

// Attrib keys stored on Contact.Attribs
const (
	AttrLeadScore  = "lead_score"
	AttrLeadTier   = "lead_tier"
	AttrLeadSignal = "lead_signals" // comma-separated signal names
)

// Signal represents a single scoring signal with its points and source.
type Signal struct {
	Name   string // machine-readable identifier
	Label  string // human-readable description
	Points int
}

// LeadData holds scraped/enriched data about a prospect used for scoring.
type LeadData struct {
	// Revenue & capacity
	MultipleProviders bool // multiple providers/locations listed
	RunningAds        bool // Google/Meta PPC detected
	ReviewCount       int  // Google review volume
	AvgRating         float64
	PremiumServices   bool // implants, Invisalign, veneers, Botox, fillers
	AffluentZip       bool // high median income area

	// Problem signals
	NoLiveChat         bool // no live chat or AI receptionist
	VoicemailAfterHrs  bool // phone goes to voicemail after hours
	NoOnlineBooking    bool // no self-service scheduling
	SlowFormResponse   bool // form response > 10 min
	HighSpendLowRating bool // high ad spend + 3.5-4.2 star rating

	// Decision-maker access
	OwnerEmailFound     bool // personal email (not info@)
	ActiveOnSocial      bool // owner posts on LinkedIn
	InIndustryCommunity bool // DentalTown, etc.
}

// ScoreResult contains the computed score, tier, and matched signals.
type ScoreResult struct {
	Score   int
	Tier    int      // 1, 2, or 3
	Tag     string   // tier_1, tier_2, or ""
	Signals []Signal // which signals matched
}

// Score evaluates a LeadData and returns the score, tier, and matched signals.
func Score(data LeadData) ScoreResult {
	var matched []Signal
	total := 0

	check := func(cond bool, s Signal) {
		if cond {
			matched = append(matched, s)
			total += s.Points
		}
	}

	// Revenue & capacity signals
	check(data.MultipleProviders, Signal{"multiple_providers", "Multiple providers/locations", 20})
	check(data.RunningAds, Signal{"running_ads", "Running Google/Meta ads", 20})
	check(data.ReviewCount >= 100, Signal{"high_reviews", fmt.Sprintf("%d Google reviews", data.ReviewCount), 15})
	check(data.PremiumServices, Signal{"premium_services", "Premium services listed", 15})
	check(data.AffluentZip, Signal{"affluent_zip", "Affluent zip code", 10})

	// Problem signals
	check(data.NoLiveChat, Signal{"no_chat", "No live chat/AI receptionist", 20})
	check(data.VoicemailAfterHrs, Signal{"voicemail_after_hrs", "Voicemail after hours", 15})
	check(data.NoOnlineBooking, Signal{"no_online_booking", "No online self-booking", 10})
	check(data.SlowFormResponse, Signal{"slow_form", "Slow form response (>10 min)", 10})
	check(data.HighSpendLowRating, Signal{"high_spend_low_rating", "High ad spend, mediocre reviews", 10})

	// Decision-maker access
	check(data.OwnerEmailFound, Signal{"owner_email", "Owner email found", 10})
	check(data.ActiveOnSocial, Signal{"active_social", "Active on social media", 5})
	check(data.InIndustryCommunity, Signal{"industry_community", "Listed in industry community", 5})

	tier := 3
	tag := ""
	if total >= Tier1Threshold {
		tier = 1
		tag = TagTier1
	} else if total >= Tier2Threshold {
		tier = 2
		tag = TagTier2
	}

	return ScoreResult{
		Score:   total,
		Tier:    tier,
		Tag:     tag,
		Signals: matched,
	}
}

// ToAttribs converts a ScoreResult into a map suitable for Contact.Attribs merge.
func (r ScoreResult) ToAttribs() map[string]string {
	names := make([]string, len(r.Signals))
	for i, s := range r.Signals {
		names[i] = s.Name
	}
	m := map[string]string{
		AttrLeadScore:  strconv.Itoa(r.Score),
		AttrLeadTier:   strconv.Itoa(r.Tier),
		AttrLeadSignal: strings.Join(names, ","),
	}
	return m
}

// LeadDataFromAttribs constructs a LeadData from Contact.Attribs populated by
// FrostByte scraping. This bridges enrichment data → scoring input.
func LeadDataFromAttribs(a map[string]string) LeadData {
	boolVal := func(key string) bool {
		v, ok := a[key]
		return ok && (v == "1" || v == "true" || v == "yes")
	}
	intVal := func(key string) int {
		v, _ := strconv.Atoi(a[key])
		return v
	}
	floatVal := func(key string) float64 {
		v, _ := strconv.ParseFloat(a[key], 64)
		return v
	}

	return LeadData{
		MultipleProviders:   boolVal("multiple_providers"),
		RunningAds:          boolVal("running_ads"),
		ReviewCount:         intVal("review_count"),
		AvgRating:           floatVal("avg_rating"),
		PremiumServices:     boolVal("premium_services"),
		AffluentZip:         boolVal("affluent_zip"),
		NoLiveChat:          boolVal("no_live_chat"),
		VoicemailAfterHrs:   boolVal("voicemail_after_hours"),
		NoOnlineBooking:     boolVal("no_online_booking"),
		SlowFormResponse:    boolVal("slow_form_response"),
		HighSpendLowRating:  boolVal("high_spend_low_rating"),
		OwnerEmailFound:     boolVal("owner_email_found"),
		ActiveOnSocial:      boolVal("active_on_social"),
		InIndustryCommunity: boolVal("in_industry_community"),
	}
}
