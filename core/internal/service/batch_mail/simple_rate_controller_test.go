package batch_mail

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// NewSimpleRateController
// ---------------------------------------------------------------------------

func TestNewSimpleRateController(t *testing.T) {
	tests := []struct {
		name         string
		maxPerMinute int
		wantMax      int
		wantBurst    int
	}{
		{"valid rate 1000", 1000, 1000, 100},        // 1000/10 = 100
		{"valid rate 60", 60, 60, 10},                // 60/10 = 6, but min 10
		{"valid rate 50", 50, 50, 10},                // 50/10 = 5, but min 10
		{"valid rate 200", 200, 200, 20},             // 200/10 = 20
		{"zero defaults to 1000", 0, 1000, 100},      // invalid -> 1000
		{"negative defaults to 1000", -1, 1000, 100}, // invalid -> 1000
		{"very small rate 1", 1, 1, 10},              // 1/10 = 0, min 10
		{"large rate 10000", 10000, 10000, 1000},     // 10000/10 = 1000
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := NewSimpleRateController(tt.maxPerMinute)
			require.NotNil(t, rc)
			assert.Equal(t, tt.wantMax, rc.maxPerMinute)
			assert.Equal(t, tt.wantBurst, rc.burstLimit)
			assert.Equal(t, tt.wantBurst, rc.sendTokens, "initial tokens should equal burst limit")
			assert.Equal(t, 0, rc.sentInLastMinute)
		})
	}
}

func TestNewSimpleRateController_WaitTimeMinimum(t *testing.T) {
	// very high rate -> waitTime should be clamped to >= 2ms
	rc := NewSimpleRateController(999999)
	assert.GreaterOrEqual(t, rc.waitTime, 2*time.Millisecond)
}

// ---------------------------------------------------------------------------
// GetMaxRate
// ---------------------------------------------------------------------------

func TestGetMaxRate(t *testing.T) {
	tests := []struct {
		name         string
		maxPerMinute int
		want         int
	}{
		{"normal", 500, 500},
		{"zero defaults", 0, 1000},
		{"negative defaults", -5, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := NewSimpleRateController(tt.maxPerMinute)
			assert.Equal(t, tt.want, rc.GetMaxRate())
		})
	}
}

// ---------------------------------------------------------------------------
// GetCurrentRate
// ---------------------------------------------------------------------------

func TestGetCurrentRate_NoSends(t *testing.T) {
	rc := NewSimpleRateController(1000)
	// no sends recorded -> rate ~ 0
	rate := rc.GetCurrentRate()
	assert.InDelta(t, 0, rate, 1.0, "rate should be ~0 with no sends")
}

func TestGetCurrentRate_AfterSends(t *testing.T) {
	rc := NewSimpleRateController(1000)
	for i := 0; i < 10; i++ {
		rc.RecordSend()
	}
	rate := rc.GetCurrentRate()
	assert.Greater(t, rate, float64(0), "rate should be positive after sends")
}

// ---------------------------------------------------------------------------
// RecordSend
// ---------------------------------------------------------------------------

func TestRecordSend_Increments(t *testing.T) {
	rc := NewSimpleRateController(1000)
	assert.Equal(t, 0, rc.sentInLastMinute)

	rc.RecordSend()
	assert.Equal(t, 1, rc.sentInLastMinute)

	rc.RecordSend()
	assert.Equal(t, 2, rc.sentInLastMinute)

	rc.RecordSend()
	assert.Equal(t, 3, rc.sentInLastMinute)
}

func TestRecordSend_ResetsAfterOneMinute(t *testing.T) {
	rc := NewSimpleRateController(1000)

	rc.RecordSend()
	rc.RecordSend()
	assert.Equal(t, 2, rc.sentInLastMinute)

	// simulate time passing: set lastResetTime to > 1 minute ago
	rc.mu.Lock()
	rc.lastResetTime = time.Now().Add(-2 * time.Minute)
	rc.mu.Unlock()

	rc.RecordSend() // should reset counter then set to 1
	assert.Equal(t, 1, rc.sentInLastMinute)
}

// ---------------------------------------------------------------------------
// Wait
// ---------------------------------------------------------------------------

func TestWait_CancelledContext(t *testing.T) {
	rc := NewSimpleRateController(1) // 1 per minute -> long waits

	// exhaust burst tokens
	rc.mu.Lock()
	rc.sendTokens = 0
	rc.lastSendTime = time.Now() // just sent
	rc.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := rc.Wait(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestWait_WithTokensAvailable(t *testing.T) {
	rc := NewSimpleRateController(10000)

	ctx := context.Background()
	start := time.Now()
	err := rc.Wait(ctx)
	elapsed := time.Since(start)

	assert.NoError(t, err)
	// with burst tokens available, should return very quickly
	assert.Less(t, elapsed, 50*time.Millisecond)
}

func TestWait_DeadlineExceeded(t *testing.T) {
	rc := NewSimpleRateController(1) // very slow rate

	rc.mu.Lock()
	rc.sendTokens = 0
	rc.lastSendTime = time.Now()
	rc.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	err := rc.Wait(ctx)
	assert.Error(t, err)
}

func TestWait_TokenDecrements(t *testing.T) {
	rc := NewSimpleRateController(1000)
	initialTokens := rc.sendTokens

	ctx := context.Background()
	_ = rc.Wait(ctx)

	// tokens should have decremented (or time-based reset happened)
	rc.mu.Lock()
	assert.Less(t, rc.sendTokens, initialTokens)
	rc.mu.Unlock()
}

// ---------------------------------------------------------------------------
// AdjustRate
// ---------------------------------------------------------------------------

func TestAdjustRate_SkipsIfLessThan5Seconds(t *testing.T) {
	rc := NewSimpleRateController(1000)
	originalWait := rc.waitTime

	// lastResetTime is now -> elapsed < 5s
	rc.AdjustRate()
	assert.Equal(t, originalWait, rc.waitTime, "should not adjust if <5s elapsed")
}

func TestAdjustRate_ReducesWaitWhenTooSlow(t *testing.T) {
	rc := NewSimpleRateController(1000)

	rc.mu.Lock()
	rc.lastResetTime = time.Now().Add(-10 * time.Second) // 10s ago
	rc.sentInLastMinute = 1                               // very few sends -> too slow
	originalWait := rc.waitTime
	rc.mu.Unlock()

	rc.AdjustRate()

	rc.mu.Lock()
	defer rc.mu.Unlock()
	// projected rate = 1/10*60 = 6, target=1000, 6 < 800 -> reduce wait by 0.8
	assert.Less(t, rc.waitTime, originalWait)
}

func TestAdjustRate_MinimumWaitTime(t *testing.T) {
	rc := NewSimpleRateController(1000)

	rc.mu.Lock()
	rc.lastResetTime = time.Now().Add(-10 * time.Second)
	rc.sentInLastMinute = 1
	rc.waitTime = 1 * time.Millisecond // already below minimum
	rc.mu.Unlock()

	rc.AdjustRate()

	rc.mu.Lock()
	defer rc.mu.Unlock()
	assert.GreaterOrEqual(t, rc.waitTime, 2*time.Millisecond)
}

func TestAdjustRate_HighRate(t *testing.T) {
	rc := NewSimpleRateController(100)

	rc.mu.Lock()
	rc.lastResetTime = time.Now().Add(-10 * time.Second)
	rc.sentInLastMinute = 500 // projected = 500/10*60 = 3000, target=100 -> too fast
	originalWait := rc.waitTime
	rc.mu.Unlock()

	rc.AdjustRate()

	rc.mu.Lock()
	defer rc.mu.Unlock()
	// code multiplies by 1 when too fast (no-op effectively), so waitTime stays same
	assert.Equal(t, originalWait, rc.waitTime)
}

// ---------------------------------------------------------------------------
// Concurrency
// ---------------------------------------------------------------------------

func TestSimpleRateController_ConcurrentRecordSend(t *testing.T) {
	rc := NewSimpleRateController(10000)
	var wg sync.WaitGroup

	n := 100
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rc.RecordSend()
		}()
	}
	wg.Wait()

	// all sends should be counted (no race)
	assert.Equal(t, n, rc.sentInLastMinute)
}

func TestSimpleRateController_ConcurrentWait(t *testing.T) {
	rc := NewSimpleRateController(100000)
	var wg sync.WaitGroup

	ctx := context.Background()
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = rc.Wait(ctx)
		}()
	}
	wg.Wait() // should not panic or deadlock
}

// ---------------------------------------------------------------------------
// Wait resets tokens after minute elapses
// ---------------------------------------------------------------------------

func TestWait_ResetsTokensAfterMinute(t *testing.T) {
	rc := NewSimpleRateController(1000)

	rc.mu.Lock()
	rc.sendTokens = 0
	rc.sentInLastMinute = 999
	rc.lastResetTime = time.Now().Add(-2 * time.Minute)
	rc.lastSendTime = time.Now().Add(-2 * time.Minute)
	rc.mu.Unlock()

	ctx := context.Background()
	err := rc.Wait(ctx)
	assert.NoError(t, err)

	rc.mu.Lock()
	defer rc.mu.Unlock()
	assert.Equal(t, 0, rc.sentInLastMinute, "counter should reset after minute")
}

// ---------------------------------------------------------------------------
// Burst limit edge cases
// ---------------------------------------------------------------------------

func TestBurstLimit_MinimumTen(t *testing.T) {
	for _, rate := range []int{1, 5, 10, 50, 99} {
		rc := NewSimpleRateController(rate)
		assert.GreaterOrEqual(t, rc.burstLimit, 10,
			"burst limit should be at least 10 for rate=%d", rate)
	}
}

func TestBurstLimit_TenthOfRate(t *testing.T) {
	for _, rate := range []int{100, 200, 500, 1000, 5000} {
		rc := NewSimpleRateController(rate)
		assert.Equal(t, rate/10, rc.burstLimit,
			"burst limit should be rate/10 for rate=%d", rate)
	}
}
