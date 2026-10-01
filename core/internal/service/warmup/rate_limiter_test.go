package warmup

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// The rate_limiter.go file is almost entirely Redis/DB-dependent.
// The only pure-logic we can test without infrastructure are the constants
// and the Lua script string validity. We also verify the singleton wiring
// and the data structures.

func TestConstants(t *testing.T) {
	tests := []struct {
		name  string
		value time.Duration
		want  time.Duration
	}{
		{"cacheTTLForLimits is 5 minutes", cacheTTLForLimits, 5 * time.Minute},
		{"counterExpireInDay is 24 hours", counterExpireInDay, 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.value)
		})
	}
}

func TestTokenBucketLuaScript_NotEmpty(t *testing.T) {
	assert.NotEmpty(t, tokenBucketLua, "Lua script should not be empty")
}

func TestTokenBucketLuaScript_ContainsKeyStructure(t *testing.T) {
	// Verify the script references expected KEYS and ARGV
	tests := []struct {
		name    string
		contain string
	}{
		{"references KEYS[1]", "KEYS[1]"},
		{"references ARGV[1] (capacity)", "ARGV[1]"},
		{"references ARGV[2] (rate)", "ARGV[2]"},
		{"references ARGV[3] (timestamp)", "ARGV[3]"},
		{"references ARGV[4] (requested tokens)", "ARGV[4]"},
		{"uses HGETALL", "HGETALL"},
		{"uses HSET", "HSET"},
		{"uses EXPIRE", "EXPIRE"},
		{"returns 1 on success", "return 1"},
		{"returns 0 on failure", "return 0"},
		{"tracks last_refill_time", "last_refill_time"},
		{"tracks tokens", "'tokens'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, tokenBucketLua, tt.contain)
		})
	}
}

func TestTokenBucketLuaScript_Logic(t *testing.T) {
	// The script should:
	// 1. Calculate time_passed and new_tokens
	// 2. Cap tokens at capacity
	// 3. Only consume if enough tokens
	// 4. Set a safety expiry
	tests := []struct {
		name    string
		contain string
	}{
		{"calculates time passed", "time_passed"},
		{"calculates new tokens", "new_tokens"},
		{"caps at capacity with math.min", "math.min"},
		{"checks if enough tokens", "current_tokens >= requested"},
		{"sets safety expiry", "expiry"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, tokenBucketLua, tt.contain)
		})
	}
}

func TestRateLimiterSingleton(t *testing.T) {
	rl1 := RateLimiter()
	rl2 := RateLimiter()
	assert.Same(t, rl1, rl2, "should return same singleton instance")
}

func TestRateLimiterServiceHasProviderService(t *testing.T) {
	rl := RateLimiter()
	assert.NotNil(t, rl.providerService, "providerService should not be nil")
}

// ---------------------------------------------------------------------------
// Hourly rate formula verification
// ---------------------------------------------------------------------------

func TestHourlyRateCalculation(t *testing.T) {
	// From the Allow() method: rate = capacity / 3600.0
	// And waits = int(1/rate) + 1
	tests := []struct {
		name          string
		hourlyLimit   int
		wantRateApprx float64
		wantWaits     int
	}{
		{
			name:          "100 per hour",
			hourlyLimit:   100,
			wantRateApprx: 100.0 / 3600.0,
			wantWaits:     int(1.0/(100.0/3600.0)) + 1, // 37
		},
		{
			name:          "3600 per hour (1/sec)",
			hourlyLimit:   3600,
			wantRateApprx: 1.0,
			wantWaits:     int(1.0/1.0) + 1, // 2
		},
		{
			name:          "10 per hour",
			hourlyLimit:   10,
			wantRateApprx: 10.0 / 3600.0,
			wantWaits:     int(1.0/(10.0/3600.0)) + 1, // 361
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capacity := float64(tt.hourlyLimit)
			rate := capacity / 3600.0
			waits := int(1.0/rate) + 1

			assert.InDelta(t, tt.wantRateApprx, rate, 0.0001)
			assert.Equal(t, tt.wantWaits, waits)
		})
	}
}

// ---------------------------------------------------------------------------
// Key format verification
// ---------------------------------------------------------------------------

func TestKeyFormats(t *testing.T) {
	// Verify the key format strings used in Allow() are consistent
	// These are computed inline in Allow(), so we just verify the patterns
	tests := []struct {
		name   string
		format string
		ip     string
		group  string
		want   string
	}{
		{
			name:   "daily key",
			format: "warmup:vw:d:%s:%s",
			ip:     "1.2.3.4",
			group:  "google",
			want:   "warmup:vw:d:1.2.3.4:google",
		},
		{
			name:   "hourly token bucket key",
			format: "warmup:tb:h:%s:%s",
			ip:     "10.0.0.1",
			group:  "outlook",
			want:   "warmup:tb:h:10.0.0.1:outlook",
		},
		{
			name:   "cache key",
			format: "warmup:limits:%s:%s",
			ip:     "192.168.1.1",
			group:  "yahoo",
			want:   "warmup:limits:192.168.1.1:yahoo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fmt.Sprintf(tt.format, tt.ip, tt.group)
			assert.Equal(t, tt.want, result)
		})
	}
}
