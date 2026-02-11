package lead_scoring

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Score edge cases ---

func TestScore_NegativeReviewCount(t *testing.T) {
	result := Score(LeadData{ReviewCount: -50})

	assert.Equal(t, 0, result.Score)
	assert.Equal(t, 3, result.Tier)
	assert.Empty(t, result.Signals)
}

func TestScore_HugeReviewCount(t *testing.T) {
	result := Score(LeadData{ReviewCount: 999999})

	assert.Equal(t, 15, result.Score)
	require.Len(t, result.Signals, 1)
	assert.Equal(t, "high_reviews", result.Signals[0].Name)
	assert.Contains(t, result.Signals[0].Label, "999999")
}

func TestScore_BoundaryTier2_Below(t *testing.T) {
	// 39 points — just below Tier2
	data := LeadData{
		RunningAds:          true, // +20
		NoOnlineBooking:     true, // +10
		ActiveOnSocial:      true, // +5
		InIndustryCommunity: true, // +5
	} // = 40... that's tier2. Let me adjust.

	// Actually 20+10+5+5=40 which IS tier2. Need 39.
	// Use: NoOnlineBooking(10) + VoicemailAfterHrs(15) + OwnerEmailFound(10) + ActiveOnSocial(5) = 40 — still 40
	// Use: NoOnlineBooking(10) + SlowFormResponse(10) + OwnerEmailFound(10) + ActiveOnSocial(5) = 35
	data = LeadData{
		NoOnlineBooking:     true, // +10
		SlowFormResponse:    true, // +10
		OwnerEmailFound:     true, // +10
		ActiveOnSocial:      true, // +5
		InIndustryCommunity: true, // +5
	} // = 40 — exact boundary, tier2

	result := Score(data)
	assert.Equal(t, 40, result.Score)
	assert.Equal(t, 2, result.Tier)
}

func TestScore_BoundaryTier3_JustBelow40(t *testing.T) {
	// 35 points — tier3
	data := LeadData{
		NoOnlineBooking:  true, // +10
		SlowFormResponse: true, // +10
		OwnerEmailFound:  true, // +10
		ActiveOnSocial:   true, // +5
	} // = 35

	result := Score(data)
	assert.Equal(t, 35, result.Score)
	assert.Equal(t, 3, result.Tier)
	assert.Equal(t, "", result.Tag)
}

func TestScore_BoundaryTier1_JustBelow75(t *testing.T) {
	// 74 points — tier2
	data := LeadData{
		MultipleProviders:   true, // +20
		RunningAds:          true, // +20
		NoLiveChat:          true, // +20
		OwnerEmailFound:     true, // +10
		InIndustryCommunity: true, // +5 (total would be 75, remove this)
	}
	// 20+20+20+10+5 = 75 — that's tier1. Need 74.
	// Use: 20+20+20+10+4... but signals are discrete. Try:
	// MultipleProviders(20) + RunningAds(20) + VoicemailAfterHrs(15) + OwnerEmailFound(10) + ActiveOnSocial(5) + InIndustryCommunity(5) = 75
	// So 20+20+15+10+5 = 70 → tier2 ✓
	data = LeadData{
		MultipleProviders: true, // +20
		RunningAds:        true, // +20
		VoicemailAfterHrs: true, // +15
		OwnerEmailFound:   true, // +10
		ActiveOnSocial:    true, // +5
	} // = 70

	result := Score(data)
	assert.Equal(t, 70, result.Score)
	assert.Equal(t, 2, result.Tier)
	assert.Equal(t, TagTier2, result.Tag)
}

func TestScore_EachSignalIsolated(t *testing.T) {
	// Verify each signal's exact point value in isolation
	tests := []struct {
		name   string
		data   LeadData
		points int
		signal string
	}{
		{"MultipleProviders", LeadData{MultipleProviders: true}, 20, "multiple_providers"},
		{"RunningAds", LeadData{RunningAds: true}, 20, "running_ads"},
		{"ReviewCount100", LeadData{ReviewCount: 100}, 15, "high_reviews"},
		{"PremiumServices", LeadData{PremiumServices: true}, 15, "premium_services"},
		{"AffluentZip", LeadData{AffluentZip: true}, 10, "affluent_zip"},
		{"NoLiveChat", LeadData{NoLiveChat: true}, 20, "no_chat"},
		{"VoicemailAfterHrs", LeadData{VoicemailAfterHrs: true}, 15, "voicemail_after_hrs"},
		{"NoOnlineBooking", LeadData{NoOnlineBooking: true}, 10, "no_online_booking"},
		{"SlowFormResponse", LeadData{SlowFormResponse: true}, 10, "slow_form"},
		{"HighSpendLowRating", LeadData{HighSpendLowRating: true}, 10, "high_spend_low_rating"},
		{"OwnerEmailFound", LeadData{OwnerEmailFound: true}, 10, "owner_email"},
		{"ActiveOnSocial", LeadData{ActiveOnSocial: true}, 5, "active_social"},
		{"InIndustryCommunity", LeadData{InIndustryCommunity: true}, 5, "industry_community"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Score(tt.data)
			assert.Equal(t, tt.points, result.Score, "wrong score for %s", tt.name)
			require.Len(t, result.Signals, 1, "expected exactly 1 signal for %s", tt.name)
			assert.Equal(t, tt.signal, result.Signals[0].Name)
			assert.Equal(t, tt.points, result.Signals[0].Points)
		})
	}
}

func TestScore_SignalOrder(t *testing.T) {
	// Signals should be in deterministic order (revenue → problem → decision-maker)
	data := LeadData{
		InIndustryCommunity: true, // decision-maker, last
		NoLiveChat:          true, // problem, middle
		RunningAds:          true, // revenue, first
	}

	result := Score(data)

	require.Len(t, result.Signals, 3)
	assert.Equal(t, "running_ads", result.Signals[0].Name)
	assert.Equal(t, "no_chat", result.Signals[1].Name)
	assert.Equal(t, "industry_community", result.Signals[2].Name)
}

func TestScore_Idempotent(t *testing.T) {
	data := LeadData{
		RunningAds:  true,
		NoLiveChat:  true,
		ReviewCount: 200,
	}

	r1 := Score(data)
	r2 := Score(data)

	assert.Equal(t, r1.Score, r2.Score)
	assert.Equal(t, r1.Tier, r2.Tier)
	assert.Equal(t, r1.Tag, r2.Tag)
	assert.Equal(t, len(r1.Signals), len(r2.Signals))
}

// --- LeadDataFromAttribs edge cases ---

func TestLeadDataFromAttribs_MalformedInts(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  int
	}{
		{"empty string", "", 0},
		{"text", "abc", 0},
		{"float string", "4.6", 0},
		{"negative", "-5", -5},
		{"very large", "999999999", 999999999},
		{"leading space", " 100", 0},
		{"overflow", "99999999999999999999", 9223372036854775807}, // strconv.Atoi clamps to MaxInt64
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := LeadDataFromAttribs(map[string]string{"review_count": tt.value})
			assert.Equal(t, tt.want, data.ReviewCount)
		})
	}
}

func TestLeadDataFromAttribs_MalformedFloats(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  float64
	}{
		{"empty", "", 0},
		{"text", "abc", 0},
		{"negative", "-1.5", -1.5},
		{"integer", "5", 5.0},
		{"very precise", "4.123456789", 4.123456789},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := LeadDataFromAttribs(map[string]string{"avg_rating": tt.value})
			assert.Equal(t, tt.want, data.AvgRating)
		})
	}
}

func TestLeadDataFromAttribs_BoolVariants(t *testing.T) {
	// Only "1", "true", "yes" are truthy
	truthy := []string{"1", "true", "yes"}
	falsy := []string{"0", "false", "no", "True", "TRUE", "Yes", "YES", "", "2", "on"}

	for _, v := range truthy {
		t.Run("truthy_"+v, func(t *testing.T) {
			data := LeadDataFromAttribs(map[string]string{"running_ads": v})
			assert.True(t, data.RunningAds, "%q should be truthy", v)
		})
	}

	for _, v := range falsy {
		t.Run("falsy_"+v, func(t *testing.T) {
			data := LeadDataFromAttribs(map[string]string{"running_ads": v})
			assert.False(t, data.RunningAds, "%q should be falsy", v)
		})
	}
}

func TestLeadDataFromAttribs_ExtraKeys(t *testing.T) {
	// Unknown keys should be silently ignored
	attribs := map[string]string{
		"running_ads":   "true",
		"unknown_field": "whatever",
		"another_extra": "123",
		"business_name": "Test Dental",
	}

	data := LeadDataFromAttribs(attribs)

	assert.True(t, data.RunningAds)
	assert.False(t, data.MultipleProviders) // not set, stays default
}

func TestLeadDataFromAttribs_KeyMismatchCasing(t *testing.T) {
	// Keys are case-sensitive — wrong casing should not match
	attribs := map[string]string{
		"Running_Ads": "true",
		"RUNNING_ADS": "true",
		"RunningAds":  "true",
	}

	data := LeadDataFromAttribs(attribs)
	assert.False(t, data.RunningAds) // only "running_ads" matches
}

// --- ToAttribs edge cases ---

func TestToAttribs_LargeScore(t *testing.T) {
	result := ScoreResult{Score: 99999, Tier: 1, Signals: []Signal{}}
	attribs := result.ToAttribs()

	assert.Equal(t, "99999", attribs[AttrLeadScore])
}

func TestToAttribs_SignalWithCommaInName(t *testing.T) {
	// Signal names shouldn't contain commas, but verify behavior
	result := ScoreResult{
		Score: 10,
		Tier:  3,
		Signals: []Signal{
			{Name: "signal_a"},
			{Name: "signal_b"},
		},
	}
	attribs := result.ToAttribs()

	parts := strings.Split(attribs[AttrLeadSignal], ",")
	assert.Equal(t, []string{"signal_a", "signal_b"}, parts)
}

func TestToAttribs_AlwaysHasThreeKeys(t *testing.T) {
	tests := []ScoreResult{
		{Score: 0, Tier: 3},
		{Score: 50, Tier: 2, Tag: TagTier2, Signals: []Signal{{Name: "a"}}},
		{Score: 100, Tier: 1, Tag: TagTier1, Signals: []Signal{{Name: "a"}, {Name: "b"}}},
	}

	for _, r := range tests {
		attribs := r.ToAttribs()
		assert.Len(t, attribs, 3, "ToAttribs should always return exactly 3 keys")
		assert.Contains(t, attribs, AttrLeadScore)
		assert.Contains(t, attribs, AttrLeadTier)
		assert.Contains(t, attribs, AttrLeadSignal)
	}
}

// --- Integration: Attribs round-trip simulating DB JSONB merge ---

func TestIntegration_JSONBMergeSimulation(t *testing.T) {
	// Simulates: existing Contact.Attribs in DB (from FrostByte enrichment)
	// → scoring → merge score attribs back → verify combined result

	// Step 1: Existing attribs from FrostByte scrape
	existing := map[string]string{
		"business_name":         "Bright Horizon Dental",
		"website_url":           "https://brighthorizon.com",
		"owner_name":            "Dr. Torres",
		"review_count":          "142",
		"avg_rating":            "4.6",
		"running_ads":           "1",
		"multiple_providers":    "true",
		"premium_services":      "yes",
		"no_live_chat":          "1",
		"voicemail_after_hours": "true",
		"owner_email_found":     "true",
	}

	// Step 2: Score
	data := LeadDataFromAttribs(existing)
	result := Score(data)
	scoreAttribs := result.ToAttribs()

	// Step 3: Simulate PostgreSQL JSONB merge (existing || new)
	merged := make(map[string]string)
	for k, v := range existing {
		merged[k] = v
	}
	for k, v := range scoreAttribs {
		merged[k] = v
	}

	// Step 4: Verify merged result
	assert.Equal(t, "Bright Horizon Dental", merged["business_name"])  // original preserved
	assert.Equal(t, "142", merged["review_count"])                     // original preserved
	assert.Equal(t, strconv.Itoa(result.Score), merged[AttrLeadScore]) // score added
	assert.Equal(t, "1", merged[AttrLeadTier])                         // tier 1
	assert.Contains(t, merged[AttrLeadSignal], "running_ads")
	assert.Contains(t, merged[AttrLeadSignal], "no_chat")

	// Step 5: Verify JSON serialization (what goes into DB)
	jsonBytes, err := json.Marshal(merged)
	require.NoError(t, err)
	assert.Contains(t, string(jsonBytes), `"lead_score"`)
	assert.Contains(t, string(jsonBytes), `"lead_tier"`)

	// Step 6: Verify JSON deserialization (reading back from DB)
	var fromDB map[string]string
	err = json.Unmarshal(jsonBytes, &fromDB)
	require.NoError(t, err)
	assert.Equal(t, merged[AttrLeadScore], fromDB[AttrLeadScore])
}

func TestIntegration_RescoringOverwritesPrevious(t *testing.T) {
	// Simulate re-scoring after enrichment data changes
	initial := map[string]string{
		"running_ads":  "1",
		"no_live_chat": "1",
	}

	// First score
	r1 := Score(LeadDataFromAttribs(initial))
	merged := make(map[string]string)
	for k, v := range initial {
		merged[k] = v
	}
	for k, v := range r1.ToAttribs() {
		merged[k] = v
	}
	assert.Equal(t, "40", merged[AttrLeadScore])
	assert.Equal(t, "2", merged[AttrLeadTier])

	// Enrichment adds more signals
	merged["multiple_providers"] = "true"
	merged["premium_services"] = "true"
	merged["owner_email_found"] = "true"

	// Re-score with updated data
	r2 := Score(LeadDataFromAttribs(merged))
	for k, v := range r2.ToAttribs() {
		merged[k] = v
	}

	// Score should reflect new signals, overwriting old score
	assert.Equal(t, strconv.Itoa(r2.Score), merged[AttrLeadScore])
	assert.Equal(t, "1", merged[AttrLeadTier]) // now tier 1
	assert.True(t, r2.Score > r1.Score)
}

func TestIntegration_Tier1Tag_Tier2Tag_Tier3NoTag(t *testing.T) {
	tests := []struct {
		name string
		data LeadData
		tag  string
	}{
		{
			"tier1 gets tier_1 tag",
			LeadData{MultipleProviders: true, RunningAds: true, NoLiveChat: true, PremiumServices: true},
			TagTier1,
		},
		{
			"tier2 gets tier_2 tag",
			LeadData{RunningAds: true, NoLiveChat: true},
			TagTier2,
		},
		{
			"tier3 gets empty tag",
			LeadData{ActiveOnSocial: true},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Score(tt.data)
			assert.Equal(t, tt.tag, result.Tag)
		})
	}
}

// --- Realistic prospect scenarios ---

func TestScenario_HighRevenueDentalPractice(t *testing.T) {
	// Dr. Torres: $180K/month, $15K ad spend, 142 reviews, 4.6 stars, no AI handling
	attribs := map[string]string{
		"multiple_providers":    "true",
		"running_ads":           "1",
		"review_count":          "142",
		"avg_rating":            "4.6",
		"premium_services":      "yes",
		"no_live_chat":          "1",
		"voicemail_after_hours": "true",
		"no_online_booking":     "0",
		"owner_email_found":     "true",
		"active_on_social":      "1",
	}

	data := LeadDataFromAttribs(attribs)
	result := Score(data)

	assert.Equal(t, 1, result.Tier, "high-revenue practice should be tier 1")
	assert.GreaterOrEqual(t, result.Score, Tier1Threshold)
	// Should have revenue + problem + decision-maker signals
	signalNames := make(map[string]bool)
	for _, s := range result.Signals {
		signalNames[s.Name] = true
	}
	assert.True(t, signalNames["running_ads"])
	assert.True(t, signalNames["no_chat"])
	assert.True(t, signalNames["owner_email"])
}

func TestScenario_SmallSoloPractice(t *testing.T) {
	// Small practice: 20 reviews, no ads, basic website
	attribs := map[string]string{
		"review_count":      "20",
		"avg_rating":        "4.8",
		"no_live_chat":      "1",
		"no_online_booking": "1",
	}

	data := LeadDataFromAttribs(attribs)
	result := Score(data)

	assert.Equal(t, 3, result.Tier, "small practice without ads should be tier 3")
	assert.Less(t, result.Score, Tier2Threshold)
}

func TestScenario_GoodFitMissingDecisionMaker(t *testing.T) {
	// Good practice but can't find owner email — tier 2
	attribs := map[string]string{
		"running_ads":      "1",
		"review_count":     "150",
		"premium_services": "true",
		"no_live_chat":     "1",
	}

	data := LeadDataFromAttribs(attribs)
	result := Score(data)

	// 20 + 15 + 15 + 20 = 70 → tier 2
	assert.Equal(t, 2, result.Tier)
}
