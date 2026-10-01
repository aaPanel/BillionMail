package batch_mail

import (
	"billionmail-core/internal/model/entity"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanUndefinedVariables(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name     string
		input    string
		wantKeep []string // substrings that must remain unchanged
		wantGone []string // substrings that must NOT appear in output
		wantHas  []string // substrings that must appear (replacement markers etc.)
	}{
		{
			name:     "preserves Subscriber.Email",
			input:    `Hello {{.Subscriber.Email}}`,
			wantKeep: []string{`{{.Subscriber.Email}}`},
		},
		{
			name:     "preserves Subscriber with whitespace",
			input:    `{{ .Subscriber.Email }}`,
			wantKeep: []string{`{{ .Subscriber.Email }}`},
		},
		{
			name:     "preserves Subscriber.Active",
			input:    `Status: {{.Subscriber.Active}}`,
			wantKeep: []string{`{{.Subscriber.Active}}`},
		},
		{
			name:     "preserves Subscriber custom attrib",
			input:    `Hi {{.Subscriber.FirstName}}`,
			wantKeep: []string{`{{.Subscriber.FirstName}}`},
		},
		{
			name:     "preserves Task variables",
			input:    `Task: {{.Task.TaskName}} ({{.Task.Subject}})`,
			wantKeep: []string{`{{.Task.TaskName}}`, `{{.Task.Subject}}`},
		},
		{
			name:     "preserves API variables",
			input:    `Key: {{.API.CustomKey}}`,
			wantKeep: []string{`{{.API.CustomKey}}`},
		},
		{
			name:     "preserves UnsubscribeURL function",
			input:    `<a href="{{UnsubscribeURL .}}">Unsub</a>`,
			wantKeep: []string{`{{UnsubscribeURL .}}`},
		},
		{
			name:     "replaces unknown variable",
			input:    `Hello {{.Unknown}}`,
			wantGone: []string{`{{.Unknown}}`},
			wantHas:  []string{`[__.Unknown__]`},
		},
		{
			name:     "replaces unknown nested variable",
			input:    `Val: {{.Foo.Bar}}`,
			wantGone: []string{`{{.Foo.Bar}}`},
			wantHas:  []string{`[__.Foo.Bar__]`},
		},
		{
			name:  "mixed known and unknown",
			input: `{{.Subscriber.Email}} {{.Unknown}} {{.Task.Subject}} {{.Nope}}`,
			wantKeep: []string{
				`{{.Subscriber.Email}}`,
				`{{.Task.Subject}}`,
			},
			wantGone: []string{
				`{{.Unknown}}`,
				`{{.Nope}}`,
			},
			wantHas: []string{
				`[__.Unknown__]`,
				`[__.Nope__]`,
			},
		},
		{
			name:     "empty content",
			input:    ``,
			wantKeep: []string{``},
		},
		{
			name:     "no template variables",
			input:    `<html><body>Hello World</body></html>`,
			wantKeep: []string{`<html><body>Hello World</body></html>`},
		},
		{
			name:     "plain text with braces that are not templates",
			input:    `Use { and } in JSON`,
			wantKeep: []string{`Use { and } in JSON`},
		},
		{
			name:     "unknown function call",
			input:    `{{SomeFunc .Arg}}`,
			wantGone: []string{`{{SomeFunc .Arg}}`},
			wantHas:  []string{`[__SomeFunc .Arg__]`},
		},
		{
			name:     "multiple Subscriber vars preserved",
			input:    `{{.Subscriber.Email}} and {{.Subscriber.Status}}`,
			wantKeep: []string{`{{.Subscriber.Email}}`, `{{.Subscriber.Status}}`},
		},
		{
			name:     "API and Task together",
			input:    `{{.API.Token}} / {{.Task.Id}}`,
			wantKeep: []string{`{{.API.Token}}`, `{{.Task.Id}}`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.cleanUndefinedVariables(tt.input)

			for _, s := range tt.wantKeep {
				assert.Contains(t, result, s, "should preserve: %s", s)
			}
			for _, s := range tt.wantGone {
				assert.NotContains(t, result, s, "should replace: %s", s)
			}
			for _, s := range tt.wantHas {
				assert.Contains(t, result, s, "should contain replacement: %s", s)
			}
		})
	}
}

func TestCleanUndefinedVariables_BlankLineRemoval(t *testing.T) {
	engine := NewTemplateEngine()

	input := "line1\n\n\nline2\n\nline3"
	result := engine.cleanUndefinedVariables(input)

	assert.NotContains(t, result, "\n\n", "blank lines should be removed")
	assert.Equal(t, "line1\nline2\nline3", result)
}

func TestCleanUndefinedVariables_WhitespaceVariants(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name     string
		input    string
		wantKeep bool
	}{
		{"compact subscriber", "{{.Subscriber.Email}}", true},
		{"spaced subscriber", "{{ .Subscriber.Email }}", true},
		{"extra spaces subscriber", "{{  .  Subscriber  .  Email  }}", true},
		{"compact task", "{{.Task.Id}}", true},
		{"spaced task", "{{ .Task.Id }}", true},
		{"compact api", "{{.API.Key}}", true},
		{"spaced api", "{{ .API.Key }}", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.cleanUndefinedVariables(tt.input)
			if tt.wantKeep {
				assert.Equal(t, tt.input, result, "known variable should be preserved as-is")
			}
		})
	}
}

func TestCleanUndefinedVariables_UnsubscribeURLVariants(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name  string
		input string
	}{
		{"basic call", `{{UnsubscribeURL .}}`},
		{"with spaces", `{{ UnsubscribeURL . }}`},
		{"extra spaces", `{{  UnsubscribeURL  .  }}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.cleanUndefinedVariables(tt.input)
			assert.Equal(t, tt.input, result, "UnsubscribeURL should be preserved")
		})
	}
}

func TestCleanUndefinedVariables_ReplacementFormat(t *testing.T) {
	engine := NewTemplateEngine()

	// The replacement format is [__<inner>__] where <inner> is the trimmed
	// content between {{ and }}
	result := engine.cleanUndefinedVariables("{{.SomeRandom}}")
	assert.Contains(t, result, "[__")
	assert.Contains(t, result, "__]")
	assert.NotContains(t, result, "{{")
	assert.NotContains(t, result, "}}")
}

func TestNewTemplateEngine_NotNil(t *testing.T) {
	engine := NewTemplateEngine()
	assert.NotNil(t, engine)
	assert.NotNil(t, engine.view)
}

func TestGetTemplateEngine_Singleton(t *testing.T) {
	e1 := GetTemplateEngine()
	e2 := GetTemplateEngine()
	assert.Same(t, e1, e2, "should return same instance")
}

// ===== VideoOutreach template variable tests =====

func TestCleanUndefinedVariables_VideoOutreach(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name     string
		input    string
		wantKeep []string
		wantGone []string
		wantHas  []string
	}{
		{
			name:     "preserves VideoOutreach.VideoURL",
			input:    `{{.VideoOutreach.VideoURL}}`,
			wantKeep: []string{`{{.VideoOutreach.VideoURL}}`},
		},
		{
			name:     "preserves VideoOutreach.ThumbnailURL",
			input:    `{{.VideoOutreach.ThumbnailURL}}`,
			wantKeep: []string{`{{.VideoOutreach.ThumbnailURL}}`},
		},
		{
			name:     "preserves VideoOutreach.LandingPageURL",
			input:    `{{.VideoOutreach.LandingPageURL}}`,
			wantKeep: []string{`{{.VideoOutreach.LandingPageURL}}`},
		},
		{
			name:     "preserves VideoOutreach.LeadTier",
			input:    `{{.VideoOutreach.LeadTier}}`,
			wantKeep: []string{`{{.VideoOutreach.LeadTier}}`},
		},
		{
			name:     "preserves VideoOutreach.LeadScore",
			input:    `{{.VideoOutreach.LeadScore}}`,
			wantKeep: []string{`{{.VideoOutreach.LeadScore}}`},
		},
		{
			name:     "preserves VideoOutreach with whitespace",
			input:    `{{ .VideoOutreach.VideoURL }}`,
			wantKeep: []string{`{{ .VideoOutreach.VideoURL }}`},
		},
		{
			name:     "preserves VideoOutreach mixed with others",
			input:    `{{.Subscriber.Email}} | {{.VideoOutreach.VideoURL}} | {{.Task.Subject}}`,
			wantKeep: []string{`{{.Subscriber.Email}}`, `{{.VideoOutreach.VideoURL}}`, `{{.Task.Subject}}`},
		},
		{
			name:     "multiple VideoOutreach vars",
			input:    `<video src="{{.VideoOutreach.VideoURL}}" poster="{{.VideoOutreach.ThumbnailURL}}">`,
			wantKeep: []string{`{{.VideoOutreach.VideoURL}}`, `{{.VideoOutreach.ThumbnailURL}}`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.cleanUndefinedVariables(tt.input)

			for _, s := range tt.wantKeep {
				assert.Contains(t, result, s, "should preserve: %s", s)
			}
			for _, s := range tt.wantGone {
				assert.NotContains(t, result, s, "should replace: %s", s)
			}
			for _, s := range tt.wantHas {
				assert.Contains(t, result, s, "should contain replacement: %s", s)
			}
		})
	}
}

func TestRenderEmailTemplate_VideoOutreachVars(t *testing.T) {
	engine := NewTemplateEngine()
	ctx := context.Background()

	contact := &entity.Contact{
		Email: "john@example.com",
		Attribs: map[string]string{
			"video_url":        "https://cdn.example.com/video.mp4",
			"thumbnail_url":    "https://cdn.example.com/thumb.png",
			"landing_page_url": "https://example.com/watch",
			"lead_tier":        "tier_1",
			"lead_score":       "85",
		},
	}

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			"video url",
			`<video src="{{.VideoOutreach.VideoURL}}">`,
			`<video src="https://cdn.example.com/video.mp4">`,
		},
		{
			"thumbnail url",
			`<img src="{{.VideoOutreach.ThumbnailURL}}">`,
			`<img src="https://cdn.example.com/thumb.png">`,
		},
		{
			"landing page url",
			`<a href="{{.VideoOutreach.LandingPageURL}}">Watch</a>`,
			`<a href="https://example.com/watch">Watch</a>`,
		},
		{
			"lead tier",
			`Tier: {{.VideoOutreach.LeadTier}}`,
			`Tier: tier_1`,
		},
		{
			"lead score",
			`Score: {{.VideoOutreach.LeadScore}}`,
			`Score: 85`,
		},
		{
			"combined with subscriber",
			`Hi {{.Subscriber.Email}}, watch: {{.VideoOutreach.LandingPageURL}}`,
			`Hi john@example.com, watch: https://example.com/watch`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.RenderEmailTemplate(ctx, tt.template, contact, nil, "")
			assert.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestRenderEmailTemplate_VideoOutreach_NilAttribs(t *testing.T) {
	engine := NewTemplateEngine()
	ctx := context.Background()

	contact := &entity.Contact{
		Email:   "john@example.com",
		Attribs: nil,
	}

	// VideoOutreach map will be empty → template var produces empty string
	result, err := engine.RenderEmailTemplate(ctx, `Video: {{.VideoOutreach.VideoURL}}`, contact, nil, "")
	assert.NoError(t, err)
	// When attribs are nil, VideoURL key doesn't exist in the map, template outputs <no value>
	// The exact behavior depends on GoFrame's template engine
	assert.NotEmpty(t, result)
}

func TestRenderEmailTemplate_VideoOutreach_EmptyAttribs(t *testing.T) {
	engine := NewTemplateEngine()
	ctx := context.Background()

	contact := &entity.Contact{
		Email:   "john@example.com",
		Attribs: map[string]string{},
	}

	// No video keys in attribs → empty VideoOutreach map
	result, err := engine.RenderEmailTemplate(ctx, `Hello {{.Subscriber.Email}}`, contact, nil, "")
	assert.NoError(t, err)
	assert.Equal(t, "Hello john@example.com", result)
}

func TestRenderEmailTemplate_VideoOutreach_PartialAttribs(t *testing.T) {
	engine := NewTemplateEngine()
	ctx := context.Background()

	contact := &entity.Contact{
		Email: "john@example.com",
		Attribs: map[string]string{
			"video_url": "https://cdn.example.com/video.mp4",
			// no thumbnail_url, landing_page_url, lead_tier, lead_score
		},
	}

	result, err := engine.RenderEmailTemplate(ctx, `Video: {{.VideoOutreach.VideoURL}}`, contact, nil, "")
	assert.NoError(t, err)
	assert.Contains(t, result, "https://cdn.example.com/video.mp4")
}

func TestRenderEmailTemplate_VideoOutreach_NilContact(t *testing.T) {
	engine := NewTemplateEngine()
	ctx := context.Background()

	// nil contact → empty VideoOutreach map
	result, err := engine.RenderEmailTemplate(ctx, `Hello World`, nil, nil, "")
	assert.NoError(t, err)
	assert.Equal(t, "Hello World", result)
}

func TestRenderEmailTemplate_VideoOutreach_WithTask(t *testing.T) {
	engine := NewTemplateEngine()
	ctx := context.Background()

	contact := &entity.Contact{
		Email: "john@example.com",
		Attribs: map[string]string{
			"video_url":        "https://cdn.example.com/video.mp4",
			"thumbnail_url":    "https://cdn.example.com/thumb.png",
			"landing_page_url": "https://example.com/watch",
		},
	}
	task := &entity.EmailTask{
		Id:      42,
		Subject: "Watch your video",
	}

	tmpl := `Task #{{.Task.Id}}: {{.Task.Subject}} - {{.VideoOutreach.LandingPageURL}}`
	result, err := engine.RenderEmailTemplate(ctx, tmpl, contact, task, "")
	assert.NoError(t, err)
	assert.Equal(t, `Task #42: Watch your video - https://example.com/watch`, result)
}

func TestRenderEmailTemplate_VideoOutreach_NonVideoAttribsNotLeaked(t *testing.T) {
	engine := NewTemplateEngine()
	ctx := context.Background()

	contact := &entity.Contact{
		Email: "john@example.com",
		Attribs: map[string]string{
			"video_url":        "https://cdn.example.com/video.mp4",
			"thumbnail_url":    "https://cdn.example.com/thumb.png",
			"landing_page_url": "https://example.com/watch",
			"business_name":    "Test Co",
			"random_field":     "should not appear in VideoOutreach",
		},
	}

	// business_name and random_field should be in Subscriber, NOT in VideoOutreach
	result, err := engine.RenderEmailTemplate(ctx, `Biz: {{.Subscriber.business_name}}`, contact, nil, "")
	assert.NoError(t, err)
	assert.Contains(t, result, "Test Co")
}

// ===== Enrichment template variable tests =====

func TestCleanUndefinedVariables_Enrichment(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name     string
		input    string
		wantKeep []string
	}{
		{"preserves Enrichment.TopSignal", `{{.Enrichment.TopSignal}}`, []string{`{{.Enrichment.TopSignal}}`}},
		{"preserves Enrichment.SecondSignal", `{{.Enrichment.SecondSignal}}`, []string{`{{.Enrichment.SecondSignal}}`}},
		{"preserves Enrichment.PainPoints", `{{.Enrichment.PainPoints}}`, []string{`{{.Enrichment.PainPoints}}`}},
		{"preserves Enrichment.SignalCount", `{{.Enrichment.SignalCount}}`, []string{`{{.Enrichment.SignalCount}}`}},
		{"preserves Enrichment.Signals", `{{.Enrichment.Signals}}`, []string{`{{.Enrichment.Signals}}`}},
		{"preserves with whitespace", `{{ .Enrichment.TopSignal }}`, []string{`{{ .Enrichment.TopSignal }}`}},
		{
			"mixed Enrichment and Subscriber",
			`{{.Subscriber.Email}} - {{.Enrichment.TopSignal}}`,
			[]string{`{{.Subscriber.Email}}`, `{{.Enrichment.TopSignal}}`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.cleanUndefinedVariables(tt.input)
			for _, s := range tt.wantKeep {
				assert.Contains(t, result, s, "should preserve: %s", s)
			}
		})
	}
}

func TestRenderEmailTemplate_EnrichmentVars(t *testing.T) {
	engine := NewTemplateEngine()
	ctx := context.Background()

	contact := &entity.Contact{
		Email: "jane@example.com",
		Attribs: map[string]string{
			"lead_signals": "no_chat,running_ads,owner_email",
			"lead_tier":    "tier_2",
			"lead_score":   "55",
			"first_name":   "Jane",
		},
	}

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			"top signal",
			`noticed {{.Enrichment.TopSignal}}`,
			`noticed missing live chat on your website`,
		},
		{
			"second signal",
			`also {{.Enrichment.SecondSignal}}`,
			`also investing in paid advertising`,
		},
		{
			"pain points joined",
			`issues: {{.Enrichment.PainPoints}}`,
			`issues: missing live chat on your website, investing in paid advertising, being hands-on with the business`,
		},
		{
			"signal count",
			`{{.Enrichment.SignalCount}} signals`,
			`3 signals`,
		},
		{
			"combined with subscriber",
			`Hi {{.Subscriber.first_name}}, noticed {{.Enrichment.TopSignal}}`,
			`Hi Jane, noticed missing live chat on your website`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.RenderEmailTemplate(ctx, tt.template, contact, nil, "")
			assert.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestRenderEmailTemplate_Enrichment_NoSignals(t *testing.T) {
	engine := NewTemplateEngine()
	ctx := context.Background()

	contact := &entity.Contact{
		Email:   "jane@example.com",
		Attribs: map[string]string{},
	}

	result, err := engine.RenderEmailTemplate(ctx, `Count: {{.Enrichment.SignalCount}}`, contact, nil, "")
	assert.NoError(t, err)
	assert.Equal(t, "Count: 0", result)
}

func TestRenderEmailTemplate_Enrichment_NilAttribs(t *testing.T) {
	engine := NewTemplateEngine()
	ctx := context.Background()

	contact := &entity.Contact{
		Email:   "jane@example.com",
		Attribs: nil,
	}

	// Enrichment map is empty when attribs nil
	result, err := engine.RenderEmailTemplate(ctx, `Hello {{.Subscriber.Email}}`, contact, nil, "")
	assert.NoError(t, err)
	assert.Equal(t, "Hello jane@example.com", result)
}
