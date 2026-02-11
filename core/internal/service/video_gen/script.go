package video_gen

import (
	"context"
	"fmt"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

const (
	anthropicBaseURL   = "https://api.anthropic.com/v1"
	defaultScriptModel = "claude-sonnet-4-5-20250929"
	maxScriptTokens    = 1024
)

// ScriptConfig holds settings for script generation.
type ScriptConfig struct {
	APIKey  string // Anthropic API key (env: ANTHROPIC_API_KEY)
	Model   string // model ID (default: claude-sonnet-4-5-20250929)
	BaseURL string // optional base URL override (for testing)
}

// scriptBaseURL returns the configured or default Anthropic base URL.
func (cfg ScriptConfig) scriptBaseURL() string {
	if cfg.BaseURL != "" {
		return cfg.BaseURL
	}
	return anthropicBaseURL
}

// ScriptInput contains the data needed to generate a personalized video script.
type ScriptInput struct {
	BusinessName  string            // prospect's business name
	OwnerName     string            // prospect's name (for personalization)
	WebsiteURL    string            // prospect's website
	Signals       map[string]bool   // lead scoring signals (no_chat, voicemail_after_hrs, etc.)
	Attribs       map[string]string // enrichment attribs for extra context
}

// ScriptOutput holds the generated script and metadata.
type ScriptOutput struct {
	Script   string // the video narration script
	Duration int    // estimated duration in seconds (based on word count)
}

// DefaultScriptConfig returns config with API key from env.
func DefaultScriptConfig() ScriptConfig {
	return ScriptConfig{
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
		Model:  defaultScriptModel,
	}
}

// systemPromptTemplate is the system prompt for video script generation.
const systemPromptTemplate = `You are a professional video script writer for personalized cold outreach videos.

Your task: Write a 30-45 second narration script for a personalized video targeting a specific business.

Rules:
- Address the business owner by name if available
- Reference their specific website and business
- Mention 1-2 specific pain points from the signals provided
- Keep tone warm, professional, and consultative (NOT salesy)
- Use short sentences for natural speech rhythm
- End with a soft CTA (e.g., "I'd love to show you how..." or "Would it be worth a quick chat?")
- Output ONLY the script text, no stage directions or formatting
- Target 80-120 words (roughly 30-45 seconds of speech)`

// painPointDescriptions maps signal names to natural-language pain point descriptions.
var painPointDescriptions = map[string]string{
	"no_chat":              "no live chat or AI receptionist on their website",
	"no_online_booking":    "no online self-booking available for customers",
	"voicemail_after_hrs":  "calls going to voicemail during peak hours",
	"high_spend_low_rating": "spending on ads but reviews suggest room for improvement",
	"slow_form_response":   "slow response time on contact form submissions",
}

// BuildScriptPrompt constructs the user prompt for script generation.
// Exported for testing without making API calls.
func BuildScriptPrompt(input ScriptInput) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Business: %s\n", input.BusinessName))
	if input.OwnerName != "" {
		sb.WriteString(fmt.Sprintf("Owner: %s\n", input.OwnerName))
	}
	sb.WriteString(fmt.Sprintf("Website: %s\n", input.WebsiteURL))

	// Pain points from signals
	var painPoints []string
	for signal := range input.Signals {
		if desc, ok := painPointDescriptions[signal]; ok {
			painPoints = append(painPoints, desc)
		}
	}
	if len(painPoints) > 0 {
		sb.WriteString(fmt.Sprintf("\nPain points identified:\n"))
		for _, pp := range painPoints {
			sb.WriteString(fmt.Sprintf("- %s\n", pp))
		}
	}

	// Extra context from attribs
	contextKeys := []string{"services_offered", "review_count", "premium_services"}
	var extras []string
	for _, key := range contextKeys {
		if val, ok := input.Attribs[key]; ok && val != "" {
			extras = append(extras, fmt.Sprintf("%s: %s", key, val))
		}
	}
	if len(extras) > 0 {
		sb.WriteString(fmt.Sprintf("\nAdditional context:\n"))
		for _, e := range extras {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
	}

	sb.WriteString("\nWrite the video narration script now.")
	return sb.String()
}

// BuildScriptMessages constructs the chat messages for the API call.
// Exported for testing.
func BuildScriptMessages(input ScriptInput) []openai.ChatCompletionMessage {
	return []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: systemPromptTemplate},
		{Role: openai.ChatMessageRoleUser, Content: BuildScriptPrompt(input)},
	}
}

// EstimateDuration estimates speech duration from word count.
// Average speaking rate: ~150 words/minute = 2.5 words/second.
func EstimateDuration(script string) int {
	words := len(strings.Fields(script))
	if words == 0 {
		return 0
	}
	seconds := (words * 60) / 150
	if seconds < 1 {
		seconds = 1
	}
	return seconds
}

// GenerateScript calls Claude API to generate a personalized video script.
func GenerateScript(ctx context.Context, cfg ScriptConfig, input ScriptInput) (*ScriptOutput, error) {
	config := openai.DefaultConfig(cfg.APIKey)
	config.BaseURL = cfg.scriptBaseURL()
	client := openai.NewClientWithConfig(config)

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:               cfg.Model,
		Messages:            BuildScriptMessages(input),
		MaxCompletionTokens: maxScriptTokens,
		Temperature:         0.7,
	})
	if err != nil {
		return nil, fmt.Errorf("claude script generation: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("claude returned no choices")
	}

	script := strings.TrimSpace(resp.Choices[0].Message.Content)
	return &ScriptOutput{
		Script:   script,
		Duration: EstimateDuration(script),
	}, nil
}
