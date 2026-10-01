package video_gen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultScriptConfig(t *testing.T) {
	cfg := DefaultScriptConfig()

	assert.Equal(t, defaultScriptModel, cfg.Model)
	// APIKey from env — may be empty in test
}

func TestBuildScriptPrompt_FullInput(t *testing.T) {
	input := ScriptInput{
		BusinessName: "Bright Smile Dental",
		OwnerName:    "Dr. Sarah Johnson",
		WebsiteURL:   "https://brightsmile.com",
		Signals: map[string]bool{
			"no_chat":             true,
			"voicemail_after_hrs": true,
		},
		Attribs: map[string]string{
			"services_offered": "implants,veneers",
			"review_count":     "142",
		},
	}

	prompt := BuildScriptPrompt(input)

	assert.Contains(t, prompt, "Bright Smile Dental")
	assert.Contains(t, prompt, "Dr. Sarah Johnson")
	assert.Contains(t, prompt, "https://brightsmile.com")
	assert.Contains(t, prompt, "Pain points identified")
	assert.Contains(t, prompt, "Additional context")
	assert.Contains(t, prompt, "implants,veneers")
	assert.Contains(t, prompt, "142")
}

func TestBuildScriptPrompt_NoOwnerName(t *testing.T) {
	input := ScriptInput{
		BusinessName: "Test Dental",
		WebsiteURL:   "https://test.com",
		Signals:      map[string]bool{"no_chat": true},
	}

	prompt := BuildScriptPrompt(input)

	assert.Contains(t, prompt, "Test Dental")
	assert.NotContains(t, prompt, "Owner:")
}

func TestBuildScriptPrompt_NoSignals(t *testing.T) {
	input := ScriptInput{
		BusinessName: "Test Dental",
		WebsiteURL:   "https://test.com",
		Signals:      map[string]bool{},
	}

	prompt := BuildScriptPrompt(input)

	assert.NotContains(t, prompt, "Pain points")
}

func TestBuildScriptPrompt_NoAttribs(t *testing.T) {
	input := ScriptInput{
		BusinessName: "Test Dental",
		WebsiteURL:   "https://test.com",
		Signals:      map[string]bool{"no_chat": true},
		Attribs:      map[string]string{},
	}

	prompt := BuildScriptPrompt(input)

	assert.NotContains(t, prompt, "Additional context")
}

func TestBuildScriptPrompt_UnknownSignal(t *testing.T) {
	input := ScriptInput{
		BusinessName: "Test",
		WebsiteURL:   "https://test.com",
		Signals:      map[string]bool{"unknown_signal": true},
	}

	prompt := BuildScriptPrompt(input)

	// Unknown signals should not produce pain point descriptions
	assert.NotContains(t, prompt, "Pain points")
}

func TestBuildScriptPrompt_AllSignals(t *testing.T) {
	input := ScriptInput{
		BusinessName: "Test",
		WebsiteURL:   "https://test.com",
		Signals: map[string]bool{
			"no_chat":               true,
			"no_online_booking":     true,
			"voicemail_after_hrs":   true,
			"high_spend_low_rating": true,
			"slow_form_response":    true,
		},
	}

	prompt := BuildScriptPrompt(input)

	assert.Contains(t, prompt, "Pain points identified")
	// All 5 pain points should appear
	for signal, desc := range painPointDescriptions {
		if input.Signals[signal] {
			assert.Contains(t, prompt, desc, "missing pain point for %s", signal)
		}
	}
}

func TestBuildScriptPrompt_EmptyInput(t *testing.T) {
	input := ScriptInput{}

	prompt := BuildScriptPrompt(input)

	assert.Contains(t, prompt, "Business: ")
	assert.Contains(t, prompt, "Website: ")
	assert.Contains(t, prompt, "Write the video narration script now.")
}

func TestBuildScriptMessages_Structure(t *testing.T) {
	input := ScriptInput{
		BusinessName: "Test",
		WebsiteURL:   "https://test.com",
		Signals:      map[string]bool{"no_chat": true},
	}

	messages := BuildScriptMessages(input)

	require.Len(t, messages, 2)
	assert.Equal(t, "system", messages[0].Role)
	assert.Equal(t, "user", messages[1].Role)
	assert.Contains(t, messages[0].Content, "professional video script writer")
	assert.Contains(t, messages[1].Content, "Test")
}

func TestBuildScriptMessages_SystemPromptRules(t *testing.T) {
	input := ScriptInput{BusinessName: "X", WebsiteURL: "https://x.com"}
	messages := BuildScriptMessages(input)

	sys := messages[0].Content
	assert.Contains(t, sys, "30-45 second")
	assert.Contains(t, sys, "80-120 words")
	assert.Contains(t, sys, "ONLY the script text")
	assert.Contains(t, sys, "NOT salesy")
}

func TestEstimateDuration_EmptyScript(t *testing.T) {
	assert.Equal(t, 0, EstimateDuration(""))
}

func TestEstimateDuration_SingleWord(t *testing.T) {
	assert.Equal(t, 1, EstimateDuration("Hello"))
}

func TestEstimateDuration_NormalScript(t *testing.T) {
	// 100 words ≈ 40 seconds at 150 wpm
	words := make([]string, 100)
	for i := range words {
		words[i] = "word"
	}
	script := strings.Join(words, " ")

	duration := EstimateDuration(script)
	assert.Equal(t, 40, duration)
}

func TestEstimateDuration_ShortScript(t *testing.T) {
	// 10 words = 4 seconds
	duration := EstimateDuration("one two three four five six seven eight nine ten")
	assert.Equal(t, 4, duration)
}

func TestEstimateDuration_LongScript(t *testing.T) {
	// 300 words = 120 seconds (2 min)
	words := make([]string, 300)
	for i := range words {
		words[i] = "word"
	}
	script := strings.Join(words, " ")

	duration := EstimateDuration(script)
	assert.Equal(t, 120, duration)
}

func TestEstimateDuration_WhitespaceOnly(t *testing.T) {
	assert.Equal(t, 0, EstimateDuration("   \t\n  "))
}

func TestPainPointDescriptions_AllKeysExist(t *testing.T) {
	// Verify all signal keys in painPointDescriptions are non-empty
	for key, desc := range painPointDescriptions {
		assert.NotEmpty(t, key, "empty key in painPointDescriptions")
		assert.NotEmpty(t, desc, "empty description for %s", key)
	}
}

func TestBuildScriptPrompt_AttribsOnlyRelevantKeys(t *testing.T) {
	input := ScriptInput{
		BusinessName: "Test",
		WebsiteURL:   "https://test.com",
		Attribs: map[string]string{
			"services_offered": "implants",
			"irrelevant_key":   "should not appear",
			"review_count":     "50",
		},
	}

	prompt := BuildScriptPrompt(input)

	assert.Contains(t, prompt, "implants")
	assert.Contains(t, prompt, "50")
	assert.NotContains(t, prompt, "irrelevant_key")
	assert.NotContains(t, prompt, "should not appear")
}

func TestBuildScriptPrompt_NilMaps(t *testing.T) {
	input := ScriptInput{
		BusinessName: "Test",
		WebsiteURL:   "https://test.com",
		Signals:      nil,
		Attribs:      nil,
	}

	prompt := BuildScriptPrompt(input)

	// Should not panic and should contain basic info
	assert.Contains(t, prompt, "Test")
	assert.NotContains(t, prompt, "Pain points")
	assert.NotContains(t, prompt, "Additional context")
}
