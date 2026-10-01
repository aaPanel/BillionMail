package batch_mail

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type SpintaxTemplate struct {
	Parts []TemplatePart
	Hash  string
}

type TemplatePart struct {
	Type    PartType
	Content string
	Options []string
}

// PartType
type PartType int

const (
	StaticPart  PartType = iota // Static text
	SpintaxPart                 // Spintax Variant
)

// SpintaxParser Spintax parser with caching
type SpintaxParser struct {
	random   *rand.Rand
	cache    *TemplateCache
	randPool *sync.Pool
}

type TemplateCache struct {
	templates map[string]*SpintaxTemplate
	mutex     sync.RWMutex
	maxSize   int
}

// NewSpintaxParser Create a new Spintax parser with caching
func NewSpintaxParser() *SpintaxParser {
	return &SpintaxParser{
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
		cache:  NewTemplateCache(100),
		randPool: &sync.Pool{
			New: func() interface{} {
				return rand.New(rand.NewSource(time.Now().UnixNano()))
			},
		},
	}
}

func NewTemplateCache(maxSize int) *TemplateCache {
	return &TemplateCache{
		templates: make(map[string]*SpintaxTemplate),
		maxSize:   maxSize,
	}
}

func (p *SpintaxParser) ParseTemplate(content string) *SpintaxTemplate {
	if content == "" {
		return &SpintaxTemplate{Parts: []TemplatePart{}}
	}

	contentHash := p.calculateHash(content)

	template := p.cache.Get(contentHash)
	if template != nil {
		return template
	}

	if !strings.Contains(content, "{") || !strings.Contains(content, "|") || !p.hasValidSpintaxPattern(content) {

		template = &SpintaxTemplate{
			Parts: []TemplatePart{{
				Type:    StaticPart,
				Content: content,
			}},
			Hash: contentHash,
		}
	} else {

		template = p.parseToTemplate(content, contentHash)
	}

	p.cache.Set(contentHash, template)
	return template
}

func (p *SpintaxParser) RenderTemplate(template *SpintaxTemplate) string {
	if template == nil || len(template.Parts) == 0 {
		return ""
	}

	r := p.randPool.Get().(*rand.Rand)
	defer p.randPool.Put(r)

	var result strings.Builder

	estimatedSize := 0
	for _, part := range template.Parts {
		if part.Type == StaticPart {
			estimatedSize += len(part.Content)
		} else {
			avgLen := 0
			for _, option := range part.Options {
				avgLen += len(option)
			}
			if len(part.Options) > 0 {
				estimatedSize += avgLen / len(part.Options)
			}
		}
	}
	result.Grow(estimatedSize + 64)

	for _, part := range template.Parts {
		switch part.Type {
		case StaticPart:
			result.WriteString(part.Content)
		case SpintaxPart:
			if len(part.Options) > 0 {
				selectedIndex := r.Intn(len(part.Options))
				result.WriteString(part.Options[selectedIndex])
			}
		}
	}

	return result.String()
}

// ParseSpintax Parse Spintax syntax with caching optimization (保持向后兼容)
func (p *SpintaxParser) ParseSpintax(content string) string {
	template := p.ParseTemplate(content)
	return p.RenderTemplate(template)
}

func (p *SpintaxParser) calculateHash(content string) string {
	hash := md5.Sum([]byte(content))
	return fmt.Sprintf("%x", hash)
}

func (p *SpintaxParser) parseToTemplate(content string, contentHash string) *SpintaxTemplate {
	template := &SpintaxTemplate{
		Parts: make([]TemplatePart, 0),
		Hash:  contentHash,
	}

	if content == "" {
		return template
	}

	i := 0
	contentLen := len(content)

	for i < contentLen {

		if skipLen := p.getSkipLength(content, i); skipLen > 0 {

			staticContent := content[i : i+skipLen]
			template.Parts = append(template.Parts, TemplatePart{
				Type:    StaticPart,
				Content: staticContent,
			})
			i += skipLen
			continue
		}

		// Search for the Spintax mode
		if content[i] == '{' {
			if spintaxResult := p.parseSpintaxAtPosition(content, i); spintaxResult != nil {
				// Add Spintax
				template.Parts = append(template.Parts, TemplatePart{
					Type:    SpintaxPart,
					Options: spintaxResult.Options,
				})
				i += spintaxResult.Length
				continue
			}
		}

		// Handling ordinary characters
		staticStart := i
		for i < contentLen && content[i] != '{' && content[i] != '<' {
			i++
		}

		if i > staticStart {
			staticContent := content[staticStart:i]
			template.Parts = append(template.Parts, TemplatePart{
				Type:    StaticPart,
				Content: staticContent,
			})
		}

		if i == staticStart {
			template.Parts = append(template.Parts, TemplatePart{
				Type:    StaticPart,
				Content: string(content[i]),
			})
			i++
		}
	}

	return template
}

// SpintaxResult Spintax analysis results
type SpintaxResult struct {
	Options []string
	Length  int
}

func (p *SpintaxParser) parseSpintaxAtPosition(content string, pos int) *SpintaxResult {
	remaining := content[pos:]
	braceCount := 1
	i := 1

	for i < len(remaining) && braceCount > 0 {
		switch remaining[i] {
		case '{':
			braceCount++
		case '}':
			braceCount--
		}
		i++
	}

	if braceCount != 0 {
		return nil
	}

	innerContent := remaining[1 : i-1]

	if !strings.Contains(innerContent, "|") {
		return nil
	}

	options := strings.Split(innerContent, "|")
	validOptions := make([]string, 0, len(options))

	for _, option := range options {
		if trimmed := strings.TrimSpace(option); trimmed != "" {
			validOptions = append(validOptions, trimmed)
		}
	}

	if len(validOptions) == 0 {
		return nil
	}

	return &SpintaxResult{
		Options: validOptions,
		Length:  i,
	}
}

func (c *TemplateCache) Get(hash string) *SpintaxTemplate {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.templates[hash]
}

func (c *TemplateCache) Set(hash string, template *SpintaxTemplate) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.templates) >= c.maxSize {
		deleteCount := c.maxSize / 10
		count := 0
		for key := range c.templates {
			delete(c.templates, key)
			count++
			if count >= deleteCount {
				break
			}
		}
	}

	c.templates[hash] = template
}

// hasValidSpintaxPattern Quickly detect whether it contains an effective Spintax pattern
func (p *SpintaxParser) hasValidSpintaxPattern(content string) bool {
	braceStart := -1
	for i, char := range content {
		if char == '{' {
			braceStart = i
		} else if char == '}' && braceStart >= 0 {
			segment := content[braceStart+1 : i]
			if strings.Contains(segment, "|") && len(strings.TrimSpace(segment)) > 0 {
				return true
			}
			braceStart = -1
		}
	}
	return false
}

// ParseSpintaxWithSeed Parse Spintax with a specified seed (used for testing or reproducible results)
func (p *SpintaxParser) ParseSpintaxWithSeed(content string, seed int64) string {
	r := rand.New(rand.NewSource(seed))
	// Use context-aware parsing with the specified seed
	return p.parseSpintaxWithContext(content, r)
}

// HasSpintax Check if the content contains Spintax syntax
func (p *SpintaxParser) HasSpintax(content string) bool {
	if !strings.Contains(content, "{") || !strings.Contains(content, "|") {
		return false
	}
	return p.hasValidSpintaxPattern(content)
}

// GetSpintaxOptions Get all Spintax options (used for preview).
// Uses context-aware parsing to skip <style>, <script>, and HTML tag interiors.
func (p *SpintaxParser) GetSpintaxOptions(content string) map[string][]string {
	result := make(map[string][]string)
	if content == "" {
		return result
	}

	i := 0
	contentLen := len(content)

	for i < contentLen {
		if skipLen := p.getSkipLength(content, i); skipLen > 0 {
			i += skipLen
			continue
		}

		if content[i] == '{' {
			if spintaxResult := p.parseSpintaxAtPosition(content, i); spintaxResult != nil {
				original := content[i : i+spintaxResult.Length]
				result[original] = spintaxResult.Options
				i += spintaxResult.Length
				continue
			}
		}

		i++
	}

	return result
}

// Global Spintax parser instance
var defaultSpintaxParser = NewSpintaxParser()

// GetSpintaxParser Get the default Spintax parser instance
func GetSpintaxParser() *SpintaxParser {
	return defaultSpintaxParser
}

// parseSpintaxWithContext parses Spintax with context awareness
func (p *SpintaxParser) parseSpintaxWithContext(content string, randGen *rand.Rand) string {
	if content == "" {
		return content
	}

	if !strings.Contains(content, "{") || !strings.Contains(content, "|") {
		return content
	}

	var result strings.Builder
	result.Grow(len(content) + 256)
	i := 0
	contentLen := len(content)

	for i < contentLen {
		if skipLen := p.getSkipLength(content, i); skipLen > 0 {

			result.WriteString(content[i : i+skipLen])
			i += skipLen
			continue
		}

		if content[i] == '{' {
			if spintaxLen := p.parseSpintaxAtPositionWithBuilder(content, i, &result, randGen); spintaxLen > 0 {
				i += spintaxLen
				continue
			}
		}

		result.WriteByte(content[i])
		i++
	}

	return result.String()
}

func (p *SpintaxParser) getSkipLength(content string, pos int) int {
	//   <style>
	if styleStart := p.findTagStart(content, pos, "style"); styleStart >= 0 {
		if styleEnd := strings.Index(content[pos:], "</style>"); styleEnd >= 0 {
			return styleEnd + 8 // 包含 </style>
		}
	}

	//   <script>
	if scriptStart := p.findTagStart(content, pos, "script"); scriptStart >= 0 {
		if scriptEnd := strings.Index(content[pos:], "</script>"); scriptEnd >= 0 {
			return scriptEnd + 9 // 包含 </script>
		}
	}

	//  Inside HTML  < ... >
	if p.isInsideHtmlTag(content, pos) {
		return p.getHtmlTagEndOffset(content, pos)
	}

	return 0
}

func (p *SpintaxParser) findTagStart(content string, pos int, tagName string) int {
	if pos >= len(content) {
		return -1
	}
	// Search for <tagName or <tagName> or <tagName space>
	pattern := "<" + tagName
	if strings.HasPrefix(strings.ToLower(content[pos:]), pattern) {
		nextCharPos := pos + len(pattern)
		if nextCharPos >= len(content) {
			return pos
		}
		nextChar := content[nextCharPos]
		if nextChar == '>' || nextChar == ' ' || nextChar == '\t' || nextChar == '\n' {
			return pos
		}
	}
	return -1
}

// isInsideHtmlTag Check whether the position is within an HTML tag
func (p *SpintaxParser) isInsideHtmlTag(content string, pos int) bool {
	// Search backward for the nearest < or >
	lastOpenTag := strings.LastIndex(content[:pos], "<")
	lastCloseTag := strings.LastIndex(content[:pos], ">")

	// If the open label is closer to the closed label, it indicates that we are within the confines of the label.
	return lastOpenTag > lastCloseTag && lastOpenTag >= 0
}

// getHtmlTagEndOffset Return to the offset from the end of the current HTML tag
func (p *SpintaxParser) getHtmlTagEndOffset(content string, pos int) int {
	if closeTagPos := strings.Index(content[pos:], ">"); closeTagPos >= 0 {
		return closeTagPos + 1
	}
	return len(content) - pos // If not closed, then proceed to the end of the content
}

func (p *SpintaxParser) ClearCache() {
	p.cache.mutex.Lock()
	defer p.cache.mutex.Unlock()
	p.cache.templates = make(map[string]*SpintaxTemplate)
}

func (p *SpintaxParser) GetCacheStats() map[string]interface{} {
	p.cache.mutex.RLock()
	defer p.cache.mutex.RUnlock()

	return map[string]interface{}{
		"cache_size":     len(p.cache.templates),
		"max_cache_size": p.cache.maxSize,
		"cache_usage":    float64(len(p.cache.templates)) / float64(p.cache.maxSize) * 100,
	}
}

// parseSpintaxAtPositionWithBuilder
func (p *SpintaxParser) parseSpintaxAtPositionWithBuilder(content string, pos int, result *strings.Builder, randGen *rand.Rand) int {

	braceCount := 1
	i := pos + 1
	contentLen := len(content)

	for i < contentLen && braceCount > 0 {
		if content[i] == '{' {
			braceCount++
		} else if content[i] == '}' {
			braceCount--
		}
		i++
	}

	if braceCount != 0 {
		return 0
	}

	innerContent := content[pos+1 : i-1]

	if !strings.Contains(innerContent, "|") {
		return 0
	}

	options := strings.Split(innerContent, "|")
	validOptions := make([]string, 0, len(options))

	for _, option := range options {
		option = strings.TrimSpace(option)
		if option != "" {
			validOptions = append(validOptions, option)
		}
	}

	if len(validOptions) == 0 {
		return 0
	}

	selectedIndex := randGen.Intn(len(validOptions))
	result.WriteString(validOptions[selectedIndex])

	return i - pos
}
