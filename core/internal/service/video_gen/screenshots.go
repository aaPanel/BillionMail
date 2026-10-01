package video_gen

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ScreenshotType identifies which page/view to capture.
type ScreenshotType string

const (
	ScreenshotHomepage ScreenshotType = "homepage"
	ScreenshotContact  ScreenshotType = "contact"
	ScreenshotGoogle   ScreenshotType = "google"
)

// ScreenshotConfig holds settings for a screenshot capture job.
type ScreenshotConfig struct {
	WebsiteURL   string // prospect's website URL
	BusinessName string // for Google Maps lookup
	OutputDir    string // directory to write PNGs
	Width        int    // viewport width (default 1920)
	Height       int    // viewport height (default 1080)
	Timeout      time.Duration
}

// ScreenshotResult holds paths to captured screenshots.
type ScreenshotResult struct {
	Homepage string // path to homepage screenshot
	Contact  string // path to contact/scheduling page screenshot
	Google   string // path to Google listing screenshot
}

// DefaultScreenshotConfig returns config with sensible defaults.
func DefaultScreenshotConfig(websiteURL, businessName, outputDir string) ScreenshotConfig {
	return ScreenshotConfig{
		WebsiteURL:   websiteURL,
		BusinessName: businessName,
		OutputDir:    outputDir,
		Width:        1920,
		Height:       1080,
		Timeout:      30 * time.Second,
	}
}

// CaptureScreenshots takes 3 screenshots of the prospect's web presence:
// homepage, contact/scheduling page, and Google Maps listing.
// Requires: npx playwright (Node.js + Playwright installed).
func CaptureScreenshots(ctx context.Context, cfg ScreenshotConfig) (*ScreenshotResult, error) {
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	result := &ScreenshotResult{
		Homepage: filepath.Join(cfg.OutputDir, "homepage.png"),
		Contact:  filepath.Join(cfg.OutputDir, "contact.png"),
		Google:   filepath.Join(cfg.OutputDir, "google.png"),
	}

	// Build the Playwright script that captures all 3 screenshots
	script := buildPlaywrightScript(cfg, result)

	timeoutCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "node", "-e", script)
	cmd.Env = append(os.Environ(), "PLAYWRIGHT_BROWSERS_PATH=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("playwright capture failed: %w\noutput: %s", err, string(out))
	}

	return result, nil
}

// buildPlaywrightScript generates a Node.js script that uses Playwright to
// capture the 3 screenshots. Exported for testing command construction.
func buildPlaywrightScript(cfg ScreenshotConfig, result *ScreenshotResult) string {
	contactURL := findContactURL(cfg.WebsiteURL)
	googleURL := buildGoogleMapsURL(cfg.BusinessName)

	return fmt.Sprintf(`
const { chromium } = require('playwright');
(async () => {
  const browser = await chromium.launch({ headless: true });
  const ctx = await browser.newContext({
    viewport: { width: %d, height: %d },
    userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36'
  });

  async function capture(url, path) {
    const page = await ctx.newPage();
    try {
      await page.goto(url, { waitUntil: 'networkidle', timeout: 15000 });
      await page.waitForTimeout(1000);
      await page.screenshot({ path, fullPage: false });
    } catch(e) {
      console.error('Failed: ' + url + ' -> ' + e.message);
    } finally {
      await page.close();
    }
  }

  await Promise.all([
    capture(%q, %q),
    capture(%q, %q),
    capture(%q, %q)
  ]);

  await browser.close();
})();
`,
		cfg.Width, cfg.Height,
		cfg.WebsiteURL, result.Homepage,
		contactURL, result.Contact,
		googleURL, result.Google,
	)
}

// findContactURL attempts to derive the contact/scheduling page URL.
// Common patterns: /contact, /book, /schedule, /appointment.
func findContactURL(baseURL string) string {
	base := strings.TrimRight(baseURL, "/")
	// Default to /contact — the Playwright script handles 404 gracefully
	return base + "/contact"
}

// buildGoogleMapsURL constructs a Google Maps search URL for the business.
func buildGoogleMapsURL(businessName string) string {
	query := strings.ReplaceAll(businessName, " ", "+")
	return fmt.Sprintf("https://www.google.com/maps/search/%s", query)
}

// ScreenshotPaths returns the 3 output file paths for a given output directory.
// Useful for checking existence without running capture.
func ScreenshotPaths(outputDir string) (homepage, contact, google string) {
	return filepath.Join(outputDir, "homepage.png"),
		filepath.Join(outputDir, "contact.png"),
		filepath.Join(outputDir, "google.png")
}
