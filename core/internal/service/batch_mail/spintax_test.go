package batch_mail

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// NewSpintaxParser
// ---------------------------------------------------------------------------

func TestNewSpintaxParser(t *testing.T) {
	p := NewSpintaxParser()
	require.NotNil(t, p)
	require.NotNil(t, p.random)
	require.NotNil(t, p.cache)
	require.NotNil(t, p.randPool)
	assert.Equal(t, 100, p.cache.maxSize)
}

// ---------------------------------------------------------------------------
// HasSpintax
// ---------------------------------------------------------------------------

func TestHasSpintax(t *testing.T) {
	p := NewSpintaxParser()

	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{"simple", "{Hello|Hi}", true},
		{"multiple", "{a|b} {c|d}", true},
		{"no braces", "plain text", false},
		{"empty", "", false},
		{"brace no pipe", "{nopipe}", false},
		{"pipe no brace", "a|b", false},
		{"only pipe", "|", false},
		{"only open brace", "{hello", false},
		{"only close brace", "hello}", false},
		{"empty braces with pipe", "{|}", true},
		{"spaces around", "{ hello | world }", true},
		{"spaces with pipe", "{  |  }", true}, // hasValidSpintaxPattern sees pipe in segment
		{"nested braces", "{{a|b}|c}", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, p.HasSpintax(tt.content))
		})
	}
}

// ---------------------------------------------------------------------------
// ParseSpintaxWithSeed (deterministic)
// ---------------------------------------------------------------------------

func TestParseSpintaxWithSeed(t *testing.T) {
	p := NewSpintaxParser()
	const seed int64 = 42

	tests := []struct {
		name       string
		content    string
		wantOneOf  []string // result must be one of these
		wantExact  string   // if non-empty, must equal this (for static content)
	}{
		{
			name:      "simple spintax",
			content:   "{Hello|Hi|Hey}",
			wantOneOf: []string{"Hello", "Hi", "Hey"},
		},
		{
			name:      "multiple spintax",
			content:   "{Hello|Hi} {World|Earth}",
			wantOneOf: nil, // checked below
		},
		{
			name:      "no spintax passthrough",
			content:   "plain text",
			wantExact: "plain text",
		},
		{
			name:      "empty string",
			content:   "",
			wantExact: "",
		},
		{
			name:      "brace no pipe passthrough",
			content:   "{nopipe}",
			wantExact: "{nopipe}",
		},
		{
			name:      "single option",
			content:   "{only}",
			wantExact: "{only}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.ParseSpintaxWithSeed(tt.content, seed)

			if tt.wantExact != "" || (tt.wantOneOf == nil && tt.content != "{Hello|Hi} {World|Earth}") {
				assert.Equal(t, tt.wantExact, result)
				return
			}

			if tt.wantOneOf != nil {
				found := false
				for _, opt := range tt.wantOneOf {
					if result == opt {
						found = true
						break
					}
				}
				assert.True(t, found, "result %q not in %v", result, tt.wantOneOf)
				return
			}

			// multiple spintax: each word must be from respective options
			parts := strings.SplitN(result, " ", 2)
			require.Len(t, parts, 2)
			assert.Contains(t, []string{"Hello", "Hi"}, parts[0])
			assert.Contains(t, []string{"World", "Earth"}, parts[1])
		})
	}
}

func TestParseSpintaxWithSeed_Deterministic(t *testing.T) {
	p := NewSpintaxParser()
	content := "{a|b|c|d|e|f|g|h|i|j}"

	// same seed -> same result every time
	var first string
	for i := 0; i < 20; i++ {
		result := p.ParseSpintaxWithSeed(content, 99)
		if i == 0 {
			first = result
		}
		assert.Equal(t, first, result, "seed=99 should always produce same output")
	}

	// different seed -> may differ (run enough to be near-certain)
	seen := map[string]bool{}
	for seed := int64(0); seed < 100; seed++ {
		seen[p.ParseSpintaxWithSeed(content, seed)] = true
	}
	assert.Greater(t, len(seen), 1, "different seeds should produce different outputs")
}

// ---------------------------------------------------------------------------
// ParseSpintax (random, non-deterministic)
// ---------------------------------------------------------------------------

func TestParseSpintax(t *testing.T) {
	p := NewSpintaxParser()

	t.Run("result is one of the options", func(t *testing.T) {
		options := []string{"a", "b", "c"}
		for i := 0; i < 50; i++ {
			r := p.ParseSpintax("{a|b|c}")
			assert.Contains(t, options, r)
		}
	})

	t.Run("plain text unchanged", func(t *testing.T) {
		assert.Equal(t, "hello world", p.ParseSpintax("hello world"))
	})

	t.Run("empty unchanged", func(t *testing.T) {
		assert.Equal(t, "", p.ParseSpintax(""))
	})
}

// ---------------------------------------------------------------------------
// ParseTemplate + RenderTemplate
// ---------------------------------------------------------------------------

func TestParseTemplate(t *testing.T) {
	p := NewSpintaxParser()

	t.Run("empty content", func(t *testing.T) {
		tmpl := p.ParseTemplate("")
		require.NotNil(t, tmpl)
		assert.Empty(t, tmpl.Parts)
	})

	t.Run("no spintax single static part", func(t *testing.T) {
		tmpl := p.ParseTemplate("just text")
		require.Len(t, tmpl.Parts, 1)
		assert.Equal(t, StaticPart, tmpl.Parts[0].Type)
		assert.Equal(t, "just text", tmpl.Parts[0].Content)
	})

	t.Run("simple spintax", func(t *testing.T) {
		tmpl := p.ParseTemplate("{Hi|Hey}")
		hasSpintax := false
		for _, part := range tmpl.Parts {
			if part.Type == SpintaxPart {
				hasSpintax = true
				assert.ElementsMatch(t, []string{"Hi", "Hey"}, part.Options)
			}
		}
		assert.True(t, hasSpintax)
	})

	t.Run("mixed static and spintax", func(t *testing.T) {
		tmpl := p.ParseTemplate("Dear {friend|colleague}, welcome")
		// should have at least 3 parts: "Dear ", spintax, ", welcome"
		spintaxCount := 0
		staticParts := []string{}
		for _, part := range tmpl.Parts {
			if part.Type == SpintaxPart {
				spintaxCount++
			} else {
				staticParts = append(staticParts, part.Content)
			}
		}
		assert.Equal(t, 1, spintaxCount)
		joined := strings.Join(staticParts, "")
		assert.Contains(t, joined, "Dear ")
		assert.Contains(t, joined, ", welcome")
	})

	t.Run("brace no pipe is static", func(t *testing.T) {
		tmpl := p.ParseTemplate("{nopipe}")
		for _, part := range tmpl.Parts {
			assert.Equal(t, StaticPart, part.Type)
		}
	})
}

func TestRenderTemplate(t *testing.T) {
	p := NewSpintaxParser()

	t.Run("nil template", func(t *testing.T) {
		assert.Equal(t, "", p.RenderTemplate(nil))
	})

	t.Run("empty parts", func(t *testing.T) {
		assert.Equal(t, "", p.RenderTemplate(&SpintaxTemplate{}))
	})

	t.Run("static only", func(t *testing.T) {
		tmpl := &SpintaxTemplate{
			Parts: []TemplatePart{
				{Type: StaticPart, Content: "hello"},
			},
		}
		assert.Equal(t, "hello", p.RenderTemplate(tmpl))
	})

	t.Run("spintax options", func(t *testing.T) {
		tmpl := &SpintaxTemplate{
			Parts: []TemplatePart{
				{Type: SpintaxPart, Options: []string{"x", "y", "z"}},
			},
		}
		for i := 0; i < 30; i++ {
			r := p.RenderTemplate(tmpl)
			assert.Contains(t, []string{"x", "y", "z"}, r)
		}
	})

	t.Run("spintax with empty options slice", func(t *testing.T) {
		tmpl := &SpintaxTemplate{
			Parts: []TemplatePart{
				{Type: SpintaxPart, Options: []string{}},
			},
		}
		assert.Equal(t, "", p.RenderTemplate(tmpl))
	})
}

// ---------------------------------------------------------------------------
// HTML-aware skipping (style/script/tag attributes)
// ---------------------------------------------------------------------------

func TestParseSpintaxWithSeed_HTMLSkipping(t *testing.T) {
	p := NewSpintaxParser()
	const seed int64 = 1

	tests := []struct {
		name        string
		content     string
		mustContain string // the raw spintax-like text must survive inside tags
	}{
		{
			name:        "style tag spintax preserved",
			content:     `<style>{color|background}: red;</style>`,
			mustContain: "{color|background}",
		},
		{
			name:        "script tag spintax preserved",
			content:     `<script>var x = "{a|b}";</script>`,
			mustContain: `{a|b}`,
		},
		{
			name:        "spintax inside html attribute preserved",
			content:     `<div class="{red|blue}">text</div>`,
			mustContain: `{red|blue}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.ParseSpintaxWithSeed(tt.content, seed)
			assert.Contains(t, result, tt.mustContain,
				"spintax inside HTML tags/attributes should not be expanded")
		})
	}

	t.Run("spintax outside tags IS expanded", func(t *testing.T) {
		content := `<p>{Hello|Hi}</p>`
		result := p.ParseSpintaxWithSeed(content, seed)
		assert.NotContains(t, result, "{Hello|Hi}")
		// must contain one of the options wrapped in <p>
		assert.True(t,
			strings.Contains(result, "<p>Hello</p>") || strings.Contains(result, "<p>Hi</p>"),
			"expected expanded spintax, got: %s", result)
	})
}

// ---------------------------------------------------------------------------
// GetSpintaxOptions
// ---------------------------------------------------------------------------

func TestGetSpintaxOptions(t *testing.T) {
	p := NewSpintaxParser()

	tests := []struct {
		name    string
		content string
		want    map[string][]string
	}{
		{
			name:    "single group",
			content: "{Hello|Hi|Hey}",
			want:    map[string][]string{"{Hello|Hi|Hey}": {"Hello", "Hi", "Hey"}},
		},
		{
			name:    "multiple groups",
			content: "{a|b} and {c|d}",
			want: map[string][]string{
				"{a|b}": {"a", "b"},
				"{c|d}": {"c", "d"},
			},
		},
		{
			name:    "no spintax",
			content: "plain text",
			want:    map[string][]string{},
		},
		{
			name:    "empty",
			content: "",
			want:    map[string][]string{},
		},
		{
			name:    "brace no pipe",
			content: "{nopipe}",
			want:    map[string][]string{},
		},
		{
			name:    "options with spaces trimmed",
			content: "{ hello | world }",
			want:    map[string][]string{"{ hello | world }": {"hello", "world"}},
		},
		{
			name:    "empty options filtered",
			content: "{a||b}",
			want:    map[string][]string{"{a||b}": {"a", "b"}},
		},
		{
			name:    "skips CSS in style tag",
			content: `<style>.btn{color|red}</style>{hello|world}`,
			want:    map[string][]string{"{hello|world}": {"hello", "world"}},
		},
		{
			name:    "skips JS in script tag",
			content: `<script>var x={a|b}</script>{yes|no}`,
			want:    map[string][]string{"{yes|no}": {"yes", "no"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.GetSpintaxOptions(tt.content)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// TemplateCache
// ---------------------------------------------------------------------------

func TestTemplateCache_GetSet(t *testing.T) {
	c := NewTemplateCache(5)

	t.Run("get missing returns nil", func(t *testing.T) {
		assert.Nil(t, c.Get("missing"))
	})

	t.Run("set then get", func(t *testing.T) {
		tmpl := &SpintaxTemplate{Hash: "h1"}
		c.Set("h1", tmpl)
		assert.Equal(t, tmpl, c.Get("h1"))
	})

	t.Run("overwrite", func(t *testing.T) {
		tmpl2 := &SpintaxTemplate{Hash: "h1-v2"}
		c.Set("h1", tmpl2)
		assert.Equal(t, tmpl2, c.Get("h1"))
	})
}

func TestTemplateCache_Eviction(t *testing.T) {
	maxSize := 5
	c := NewTemplateCache(maxSize)

	// fill to capacity
	for i := 0; i < maxSize; i++ {
		key := string(rune('a' + i))
		c.Set(key, &SpintaxTemplate{Hash: key})
	}
	assert.Equal(t, maxSize, len(c.templates))

	// one more triggers eviction of maxSize/10 = 0 -> at least 0, but code does maxSize/10
	// with maxSize=5, deleteCount=0, so no eviction. Use maxSize=10.
	c2 := NewTemplateCache(10)
	for i := 0; i < 10; i++ {
		key := string(rune('a' + i))
		c2.Set(key, &SpintaxTemplate{Hash: key})
	}
	assert.Equal(t, 10, len(c2.templates))

	// add one more -> triggers eviction of 10/10=1 entry
	c2.Set("overflow", &SpintaxTemplate{Hash: "overflow"})
	// should have 10 entries (deleted 1, added 1)
	assert.Equal(t, 10, len(c2.templates))
	assert.NotNil(t, c2.Get("overflow"))
}

func TestTemplateCache_ParseTemplate_CacheHit(t *testing.T) {
	p := NewSpintaxParser()
	p.ClearCache()

	content := "{x|y|z}"
	tmpl1 := p.ParseTemplate(content)
	tmpl2 := p.ParseTemplate(content)

	// same pointer -> came from cache
	assert.Same(t, tmpl1, tmpl2)
}

// ---------------------------------------------------------------------------
// ClearCache / GetCacheStats
// ---------------------------------------------------------------------------

func TestClearCache(t *testing.T) {
	p := NewSpintaxParser()
	p.ParseTemplate("{a|b}")
	stats := p.GetCacheStats()
	assert.Greater(t, stats["cache_size"].(int), 0)

	p.ClearCache()
	stats = p.GetCacheStats()
	assert.Equal(t, 0, stats["cache_size"].(int))
}

func TestGetCacheStats(t *testing.T) {
	p := NewSpintaxParser()
	p.ClearCache()

	stats := p.GetCacheStats()
	assert.Equal(t, 0, stats["cache_size"].(int))
	assert.Equal(t, 100, stats["max_cache_size"].(int))
	assert.Equal(t, float64(0), stats["cache_usage"].(float64))

	p.ParseTemplate("{a|b}")
	stats = p.GetCacheStats()
	assert.Equal(t, 1, stats["cache_size"].(int))
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestParseSpintax_EdgeCases(t *testing.T) {
	p := NewSpintaxParser()

	t.Run("only pipes {|}", func(t *testing.T) {
		// {|} has two empty options; after TrimSpace+empty filter -> nil result -> returned as-is
		result := p.ParseSpintaxWithSeed("{|}", 1)
		// both options are empty after trim, parseSpintaxAtPositionWithBuilder returns 0
		assert.Equal(t, "{|}", result)
	})

	t.Run("unclosed brace", func(t *testing.T) {
		result := p.ParseSpintaxWithSeed("{hello|world", 1)
		assert.Equal(t, "{hello|world", result)
	})

	t.Run("extra closing brace", func(t *testing.T) {
		result := p.ParseSpintaxWithSeed("hello}", 1)
		assert.Equal(t, "hello}", result)
	})

	t.Run("adjacent spintax groups", func(t *testing.T) {
		result := p.ParseSpintaxWithSeed("{a|b}{c|d}", 1)
		assert.Len(t, result, 2) // one char from each group
		assert.Contains(t, []string{"a", "b"}, string(result[0]))
		assert.Contains(t, []string{"c", "d"}, string(result[1]))
	})

	t.Run("very long content no spintax", func(t *testing.T) {
		long := strings.Repeat("x", 10000)
		assert.Equal(t, long, p.ParseSpintaxWithSeed(long, 1))
	})

	t.Run("spintax with whitespace options", func(t *testing.T) {
		result := p.ParseSpintaxWithSeed("{ hello | world }", 1)
		assert.Contains(t, []string{"hello", "world"}, result)
	})

	t.Run("single valid option among empties", func(t *testing.T) {
		// {||x||} -> only "x" is valid
		result := p.ParseSpintaxWithSeed("{||x||}", 1)
		assert.Equal(t, "x", result)
	})
}

// ---------------------------------------------------------------------------
// GetSpintaxParser singleton
// ---------------------------------------------------------------------------

func TestGetSpintaxParser(t *testing.T) {
	p1 := GetSpintaxParser()
	p2 := GetSpintaxParser()
	assert.Same(t, p1, p2)
	assert.NotNil(t, p1)
}

// ---------------------------------------------------------------------------
// Concurrency safety
// ---------------------------------------------------------------------------

func TestSpintaxParser_ConcurrentUse(t *testing.T) {
	p := NewSpintaxParser()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = p.ParseSpintax("{a|b|c|d}")
		}()
	}

	wg.Wait() // no race / panic
}

func TestTemplateCache_ConcurrentAccess(t *testing.T) {
	c := NewTemplateCache(50)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(2)
		key := string(rune('a' + (i % 26)))
		go func(k string) {
			defer wg.Done()
			c.Set(k, &SpintaxTemplate{Hash: k})
		}(key)
		go func(k string) {
			defer wg.Done()
			_ = c.Get(k)
		}(key)
	}

	wg.Wait()
}
