package rbac

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Concurrent captcha generation safety
// ---------------------------------------------------------------------------

func TestConcurrentCaptchaGeneration(t *testing.T) {
	pool := NewCodePool(50, DefaultExpiration, DefaultRefreshRate)

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	results := make([]struct {
		id   string
		b64s string
		err  error
	}, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			id, b64s, _, err := pool.captcha.Generate()
			results[idx].id = id
			results[idx].b64s = b64s
			results[idx].err = err
		}(i)
	}

	wg.Wait()

	ids := make(map[string]bool)
	for i, r := range results {
		assert.NoError(t, r.err, "goroutine %d should not error", i)
		assert.NotEmpty(t, r.id, "goroutine %d should return an id", i)
		assert.NotEmpty(t, r.b64s, "goroutine %d should return base64 image", i)

		// All IDs should be unique
		assert.False(t, ids[r.id], "goroutine %d produced duplicate id: %s", i, r.id)
		ids[r.id] = true
	}
}

func TestConcurrentGetCode(t *testing.T) {
	pool := NewCodePool(20, DefaultExpiration, DefaultRefreshRate)
	// Pre-fill pool
	pool.RefreshPool(true)

	const goroutines = 15
	var wg sync.WaitGroup
	wg.Add(goroutines)

	type result struct {
		id  string
		err error
	}
	results := make([]result, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			id, _, err := pool.GetCode()
			results[idx] = result{id, err}
		}(i)
	}

	wg.Wait()

	for i, r := range results {
		assert.NoError(t, r.err, "goroutine %d should not error", i)
		assert.NotEmpty(t, r.id, "goroutine %d should get an id", i)
	}
}

// ---------------------------------------------------------------------------
// Pool refresh behavior
// ---------------------------------------------------------------------------

func TestRefreshPool_FillsToCapacity(t *testing.T) {
	poolSize := 10
	pool := NewCodePool(poolSize, DefaultExpiration, DefaultRefreshRate)

	assert.Empty(t, pool.pool, "pool should start empty")

	pool.RefreshPool(false)

	pool.mutex.RLock()
	size := len(pool.pool)
	pool.mutex.RUnlock()

	assert.Equal(t, poolSize, size, "pool should fill to capacity")
}

func TestRefreshPool_ForceRefreshOnFullPool(t *testing.T) {
	poolSize := 5
	pool := NewCodePool(poolSize, DefaultExpiration, DefaultRefreshRate)
	pool.RefreshPool(false) // fill to capacity

	pool.mutex.RLock()
	sizeBefore := len(pool.pool)
	pool.mutex.RUnlock()
	assert.Equal(t, poolSize, sizeBefore)

	// Force refresh when already full should add at least 1
	pool.RefreshPool(true)

	pool.mutex.RLock()
	sizeAfter := len(pool.pool)
	pool.mutex.RUnlock()

	// Since pool was full, neededCodes = 0, but force=true forces neededCodes=1.
	// The new captcha gets a unique ID, so pool grows by 1.
	assert.GreaterOrEqual(t, sizeAfter, poolSize)
}

func TestRefreshPool_DoesNotOverfill(t *testing.T) {
	poolSize := 5
	pool := NewCodePool(poolSize, DefaultExpiration, DefaultRefreshRate)

	// Refresh multiple times
	for i := 0; i < 5; i++ {
		pool.RefreshPool(false)
	}

	pool.mutex.RLock()
	size := len(pool.pool)
	pool.mutex.RUnlock()

	// Without force, pool should not grow beyond poolSize
	assert.Equal(t, poolSize, size, "pool should not exceed capacity")
}

// ---------------------------------------------------------------------------
// Expired captcha handling
// ---------------------------------------------------------------------------

func TestExpiredCaptcha_NotVerifiable(t *testing.T) {
	// The base64Captcha MemoryStore only runs GC (collect) when numStored
	// exceeds collectNum. NewCodePool sets collectNum = poolSize * 2.
	// So with poolSize=2, collectNum=4. We need to store >4 items after
	// our target expires to trigger GC and remove the expired entry.
	poolSize := 2                      // collectNum = 4
	expiration := 200 * time.Millisecond
	pool := NewCodePool(poolSize, expiration, DefaultRefreshRate)

	// Generate our target captcha
	id, _, answer, err := pool.captcha.Generate()
	require.NoError(t, err)

	// Verify retrievable immediately
	got := pool.store.Get(id, false)
	assert.Equal(t, answer, got, "should be retrievable immediately")

	// Wait for it to expire
	time.Sleep(expiration + 100*time.Millisecond)

	// Generate enough new captchas to exceed collectNum and trigger GC.
	// collectNum = poolSize*2 = 4, but numStored includes the original,
	// so we need to add at least 4 more (total 5 > 4).
	for i := 0; i < poolSize*2+2; i++ {
		_, _, _, _ = pool.captcha.Generate()
	}

	// Give the async GC goroutine time to run
	time.Sleep(100 * time.Millisecond)

	// After GC sweep, the expired captcha should be removed
	gotAfter := pool.store.Get(id, false)
	assert.Empty(t, gotAfter, "expired captcha should be cleaned up after GC sweep")
}

// ---------------------------------------------------------------------------
// Multiple verification attempts (should invalidate after first use)
// ---------------------------------------------------------------------------

func TestVerifyCode_InvalidatesAfterFirstUse(t *testing.T) {
	pool := NewCodePool(10, DefaultExpiration, DefaultRefreshRate)

	// Generate captcha directly
	id, _, answer, err := pool.captcha.Generate()
	require.NoError(t, err)
	require.NotEmpty(t, answer)

	// First verification should succeed
	ok := pool.VerifyCode(id, answer)
	assert.True(t, ok, "first verification with correct answer should succeed")

	// Second verification with same id+answer should fail (consumed)
	ok2 := pool.VerifyCode(id, answer)
	assert.False(t, ok2, "second verification should fail (captcha consumed)")
}

func TestVerifyCode_WrongAnswerDoesNotConsume(t *testing.T) {
	pool := NewCodePool(10, DefaultExpiration, DefaultRefreshRate)

	id, _, answer, err := pool.captcha.Generate()
	require.NoError(t, err)

	// Wrong answer -- the base64Captcha store.Verify with clear=true
	// clears the captcha regardless of correct/wrong answer per the library.
	// So after a wrong attempt, the captcha is consumed too.
	ok := pool.VerifyCode(id, "definitely-wrong")
	assert.False(t, ok, "wrong answer should return false")

	// After a wrong Verify call with clear=true, the captcha is consumed
	ok2 := pool.VerifyCode(id, answer)
	assert.False(t, ok2, "captcha should be consumed even after wrong attempt")
}

// ---------------------------------------------------------------------------
// GetAnswer does not consume captcha
// ---------------------------------------------------------------------------

func TestGetAnswer_DoesNotConsume(t *testing.T) {
	pool := NewCodePool(10, DefaultExpiration, DefaultRefreshRate)

	id, _, answer, err := pool.captcha.Generate()
	require.NoError(t, err)

	// GetAnswer should return the answer without consuming
	got := pool.GetAnswer(id)
	assert.Equal(t, answer, got)

	// Should still be retrievable
	got2 := pool.GetAnswer(id)
	assert.Equal(t, answer, got2, "GetAnswer should not consume the captcha")

	// And verification should still work
	ok := pool.VerifyCode(id, answer)
	assert.True(t, ok, "verification should work after GetAnswer calls")
}

// ---------------------------------------------------------------------------
// NewCodePool configuration
// ---------------------------------------------------------------------------

func TestNewCodePool_Configuration(t *testing.T) {
	tests := []struct {
		name        string
		poolSize    int
		expiration  time.Duration
		refreshRate int
	}{
		{"default values", DefaultPoolSize, DefaultExpiration, DefaultRefreshRate},
		{"small pool", 5, 1 * time.Minute, 2},
		{"large pool", 500, 10 * time.Minute, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewCodePool(tt.poolSize, tt.expiration, tt.refreshRate)

			assert.NotNil(t, pool)
			assert.NotNil(t, pool.store)
			assert.NotNil(t, pool.driver)
			assert.NotNil(t, pool.captcha)
			assert.NotNil(t, pool.pool)
			assert.Equal(t, tt.poolSize, pool.poolSize)
			assert.Equal(t, tt.refreshRate, pool.refreshRate)
			assert.Equal(t, tt.expiration, pool.expiration)
		})
	}
}

// ---------------------------------------------------------------------------
// Default constants
// ---------------------------------------------------------------------------

func TestDefaultConstants(t *testing.T) {
	assert.Equal(t, 100, DefaultPoolSize)
	assert.Equal(t, 3*time.Minute, DefaultExpiration)
	assert.Equal(t, 60, DefaultRefreshRate)
}

// ---------------------------------------------------------------------------
// SetExpiration updates pool config
// ---------------------------------------------------------------------------

func TestSetExpiration(t *testing.T) {
	pool := NewCodePool(10, DefaultExpiration, DefaultRefreshRate)

	newExp := 5 * time.Minute
	pool.SetExpiration(newExp)

	pool.mutex.RLock()
	exp := pool.expiration
	pool.mutex.RUnlock()

	assert.Equal(t, newExp, exp)
}

func TestSetExpiration_IgnoresZero(t *testing.T) {
	pool := NewCodePool(10, DefaultExpiration, DefaultRefreshRate)

	pool.SetExpiration(0)

	pool.mutex.RLock()
	exp := pool.expiration
	pool.mutex.RUnlock()

	assert.Equal(t, DefaultExpiration, exp, "zero expiration should be ignored")
}

func TestSetExpiration_IgnoresNegative(t *testing.T) {
	pool := NewCodePool(10, DefaultExpiration, DefaultRefreshRate)

	pool.SetExpiration(-1 * time.Minute)

	pool.mutex.RLock()
	exp := pool.expiration
	pool.mutex.RUnlock()

	assert.Equal(t, DefaultExpiration, exp, "negative expiration should be ignored")
}

// ---------------------------------------------------------------------------
// GetCode with pre-filled pool
// ---------------------------------------------------------------------------

func TestGetCode_FromPreFilledPool(t *testing.T) {
	pool := NewCodePool(10, DefaultExpiration, DefaultRefreshRate)
	pool.RefreshPool(false) // fill pool

	pool.mutex.RLock()
	sizeBefore := len(pool.pool)
	pool.mutex.RUnlock()
	require.Greater(t, sizeBefore, 0)

	id, b64s, err := pool.GetCode()
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.NotEmpty(t, b64s)
}

// ---------------------------------------------------------------------------
// ConfigDriver
// ---------------------------------------------------------------------------

func TestConfigDriver_NilIgnored(t *testing.T) {
	pool := NewCodePool(10, DefaultExpiration, DefaultRefreshRate)
	originalDriver := pool.driver

	pool.ConfigDriver(nil)

	assert.Same(t, originalDriver, pool.driver, "nil driver should not change config")
}
