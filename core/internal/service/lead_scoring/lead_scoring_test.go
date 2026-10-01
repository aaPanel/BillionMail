package lead_scoring

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScore_Tier1_HighValueLead(t *testing.T) {
	// Multi-provider dental practice running ads, 150 reviews, no chat, voicemail after hours, owner email found
	data := LeadData{
		MultipleProviders: true, // +20
		RunningAds:        true, // +20
		ReviewCount:       150,  // +15
		PremiumServices:   true, // +15
		NoLiveChat:        true, // +20
		VoicemailAfterHrs: true, // +15
		OwnerEmailFound:   true, // +10
	} // = 115

	result := Score(data)

	assert.Equal(t, 115, result.Score)
	assert.Equal(t, 1, result.Tier)
	assert.Equal(t, TagTier1, result.Tag)
	assert.Len(t, result.Signals, 7)
}

func TestScore_Tier1_MinThreshold(t *testing.T) {
	// Exactly 75 points
	data := LeadData{
		MultipleProviders: true, // +20
		RunningAds:        true, // +20
		NoLiveChat:        true, // +20
		PremiumServices:   true, // +15
	} // = 75

	result := Score(data)

	assert.Equal(t, 75, result.Score)
	assert.Equal(t, 1, result.Tier)
	assert.Equal(t, TagTier1, result.Tag)
}

func TestScore_Tier2_MidRange(t *testing.T) {
	// Good signals but missing some key ones
	data := LeadData{
		RunningAds:      true, // +20
		NoLiveChat:      true, // +20
		NoOnlineBooking: true, // +10
	} // = 50

	result := Score(data)

	assert.Equal(t, 50, result.Score)
	assert.Equal(t, 2, result.Tier)
	assert.Equal(t, TagTier2, result.Tag)
}

func TestScore_Tier2_MinThreshold(t *testing.T) {
	data := LeadData{
		RunningAds: true, // +20
		NoLiveChat: true, // +20
	} // = 40

	result := Score(data)

	assert.Equal(t, 40, result.Score)
	assert.Equal(t, 2, result.Tier)
}

func TestScore_Tier3_TooLow(t *testing.T) {
	data := LeadData{
		ReviewCount:     50,   // not enough for signal
		NoOnlineBooking: true, // +10
		ActiveOnSocial:  true, // +5
	} // = 15

	result := Score(data)

	assert.Equal(t, 15, result.Score)
	assert.Equal(t, 3, result.Tier)
	assert.Equal(t, "", result.Tag)
}

func TestScore_ZeroData(t *testing.T) {
	result := Score(LeadData{})

	assert.Equal(t, 0, result.Score)
	assert.Equal(t, 3, result.Tier)
	assert.Empty(t, result.Signals)
}

func TestScore_MaxScore(t *testing.T) {
	data := LeadData{
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

	result := Score(data)

	// 20+20+15+15+10 + 20+15+10+10+10 + 10+5+5 = 165
	assert.Equal(t, 165, result.Score)
	assert.Equal(t, 1, result.Tier)
	assert.Len(t, result.Signals, 13)
}

func TestScore_ReviewThreshold(t *testing.T) {
	tests := []struct {
		name    string
		count   int
		matched bool
	}{
		{"99 reviews — below threshold", 99, false},
		{"100 reviews — at threshold", 100, true},
		{"500 reviews — above threshold", 500, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Score(LeadData{ReviewCount: tt.count})
			hasReviewSignal := false
			for _, s := range result.Signals {
				if s.Name == "high_reviews" {
					hasReviewSignal = true
				}
			}
			assert.Equal(t, tt.matched, hasReviewSignal)
		})
	}
}

func TestToAttribs(t *testing.T) {
	result := ScoreResult{
		Score: 80,
		Tier:  1,
		Tag:   TagTier1,
		Signals: []Signal{
			{Name: "running_ads"},
			{Name: "no_chat"},
			{Name: "owner_email"},
		},
	}

	attribs := result.ToAttribs()

	assert.Equal(t, "80", attribs[AttrLeadScore])
	assert.Equal(t, "1", attribs[AttrLeadTier])
	assert.Equal(t, "running_ads,no_chat,owner_email", attribs[AttrLeadSignal])
}

func TestToAttribs_NoSignals(t *testing.T) {
	result := ScoreResult{Score: 0, Tier: 3}
	attribs := result.ToAttribs()

	assert.Equal(t, "0", attribs[AttrLeadScore])
	assert.Equal(t, "3", attribs[AttrLeadTier])
	assert.Equal(t, "", attribs[AttrLeadSignal])
}

func TestLeadDataFromAttribs(t *testing.T) {
	attribs := map[string]string{
		"multiple_providers":    "true",
		"running_ads":           "1",
		"review_count":          "142",
		"avg_rating":            "4.6",
		"premium_services":      "yes",
		"affluent_zip":          "false",
		"no_live_chat":          "1",
		"voicemail_after_hours": "true",
		"no_online_booking":     "0",
		"slow_form_response":    "1",
		"high_spend_low_rating": "false",
		"owner_email_found":     "true",
		"active_on_social":      "1",
		"in_industry_community": "yes",
	}

	data := LeadDataFromAttribs(attribs)

	assert.True(t, data.MultipleProviders)
	assert.True(t, data.RunningAds)
	assert.Equal(t, 142, data.ReviewCount)
	assert.Equal(t, 4.6, data.AvgRating)
	assert.True(t, data.PremiumServices)
	assert.False(t, data.AffluentZip)
	assert.True(t, data.NoLiveChat)
	assert.True(t, data.VoicemailAfterHrs)
	assert.False(t, data.NoOnlineBooking)
	assert.True(t, data.SlowFormResponse)
	assert.False(t, data.HighSpendLowRating)
	assert.True(t, data.OwnerEmailFound)
	assert.True(t, data.ActiveOnSocial)
	assert.True(t, data.InIndustryCommunity)
}

func TestLeadDataFromAttribs_Empty(t *testing.T) {
	data := LeadDataFromAttribs(map[string]string{})

	assert.False(t, data.MultipleProviders)
	assert.Equal(t, 0, data.ReviewCount)
	assert.Equal(t, 0.0, data.AvgRating)
}

func TestLeadDataFromAttribs_Nil(t *testing.T) {
	data := LeadDataFromAttribs(nil)
	assert.Equal(t, LeadData{}, data)
}

func TestRoundtrip_AttribsToScoreToAttribs(t *testing.T) {
	// Simulate: enrichment data → LeadData → Score → store back in attribs
	enriched := map[string]string{
		"multiple_providers":    "true",
		"running_ads":           "1",
		"review_count":          "150",
		"premium_services":      "true",
		"no_live_chat":          "1",
		"voicemail_after_hours": "true",
		"owner_email_found":     "true",
	}

	data := LeadDataFromAttribs(enriched)
	result := Score(data)
	scoreAttribs := result.ToAttribs()

	require.Equal(t, 1, result.Tier)
	assert.Equal(t, "115", scoreAttribs[AttrLeadScore])
	assert.Equal(t, "1", scoreAttribs[AttrLeadTier])
	assert.Contains(t, scoreAttribs[AttrLeadSignal], "running_ads")
	assert.Contains(t, scoreAttribs[AttrLeadSignal], "no_chat")
}

func TestSignalPoints_SumToMax(t *testing.T) {
	// Verify all signals add up to expected max
	allTrue := LeadData{
		MultipleProviders:   true,
		RunningAds:          true,
		ReviewCount:         100,
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

	result := Score(allTrue)

	expectedMax := 20 + 20 + 15 + 15 + 10 + // revenue
		20 + 15 + 10 + 10 + 10 + // problem
		10 + 5 + 5 // decision-maker
	assert.Equal(t, expectedMax, result.Score)
	assert.Equal(t, 165, expectedMax)
}
