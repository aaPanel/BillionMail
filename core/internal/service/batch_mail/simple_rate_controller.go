package batch_mail

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"sync"
	"time"
)

type SimpleRateController struct {
	maxPerMinute     int           // max send per minute
	sentInLastMinute int           // sent in last minute
	lastResetTime    time.Time     // last reset time
	waitTime         time.Duration // wait time
	mu               sync.Mutex    // mutex

	// add field for more precise rate control
	lastSendTime time.Time // last send time
	sendTokens   int       // available send tokens
	burstLimit   int       // allowed burst send limit
}

// NewSimpleRateController creates a new SimpleRateController with the specified max send rate per minute.
func NewSimpleRateController(maxPerMinute int) *SimpleRateController {
	// parameter validation
	if maxPerMinute <= 0 {
		maxPerMinute = 1000
	}

	waitTime := time.Duration(float64(time.Minute) / float64(maxPerMinute) * 0.9) // reduce 10% wait time

	// ensure min wait time not too small, avoid system overload
	minWaitTime := 2 * time.Millisecond
	if waitTime < minWaitTime {
		waitTime = minWaitTime
	}

	// calculate burst limit, allow more send in short time
	burstLimit := maxPerMinute / 10
	if burstLimit < 10 {
		burstLimit = 10
	}

	g.Log().Debug(context.Background(), "create rate controller: %d per minute, wait time %v, burst limit %d",
		maxPerMinute, waitTime, burstLimit)

	return &SimpleRateController{
		maxPerMinute:  maxPerMinute,
		lastResetTime: time.Now(),
		waitTime:      waitTime,
		lastSendTime:  time.Now().Add(-waitTime), // ensure first send not need wait
		sendTokens:    burstLimit,                // initial tokens equal burst limit
		burstLimit:    burstLimit,
	}
}

// Wait wait for a send interval
func (r *SimpleRateController) Wait(ctx context.Context) error {
	r.mu.Lock()

	// current time and last reset time difference
	now := time.Now()
	timeSinceReset := now.Sub(r.lastResetTime)

	// check if need reset count
	if timeSinceReset >= time.Minute {
		// reset count
		r.sentInLastMinute = 0
		r.lastResetTime = now

		// reset tokens
		r.sendTokens = r.burstLimit
	} else {
		// add tokens based on elapsed time
		elapsedRatio := float64(timeSinceReset) / float64(time.Minute)
		tokensToAdd := int(float64(r.maxPerMinute)*elapsedRatio) - r.sentInLastMinute

		if tokensToAdd > 0 {
			// add tokens, but not exceed burst limit
			r.sendTokens += tokensToAdd
			if r.sendTokens > r.burstLimit {
				r.sendTokens = r.burstLimit
			}
		}
	}

	// if have available tokens, return immediately
	if r.sendTokens > 0 {
		r.sendTokens--
		timeSinceLastSend := now.Sub(r.lastSendTime)
		r.lastSendTime = now
		r.mu.Unlock()

		// if time since last send too short, sleep briefly to avoid system overload
		if timeSinceLastSend < time.Millisecond {
			time.Sleep(time.Millisecond)
		}

		return nil
	}

	// calculate wait time
	nextSendTime := r.lastSendTime.Add(r.waitTime)
	waitTime := nextSendTime.Sub(now)

	// already reached send time, no need wait
	if waitTime <= 0 {
		r.lastSendTime = now
		r.mu.Unlock()
		return nil
	}

	// update last send time
	r.lastSendTime = nextSendTime
	r.mu.Unlock()

	// use more precise timer to wait
	timer := time.NewTimer(waitTime)
	defer timer.Stop()

	select {
	case <-timer.C:
		// wait completed
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// RecordSend record one send
func (r *SimpleRateController) RecordSend() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// check if need reset counteed reset count
	now := time.Now()
	if now.Sub(r.lastResetTime) >= time.Minute {
		r.sentInLastMinute = 1
		r.lastResetTime = now
	} else {
		// add send count
		r.sentInLastMinute++
	}
}

// AdjustRate adjust send rate
func (r *SimpleRateController) AdjustRate() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsedSeconds := now.Sub(r.lastResetTime).Seconds()
	if elapsedSeconds < 5 {
		// collect at least 5 seconds data before adjust
		return
	}

	// projected rate in last minute
	projectedRate := float64(r.sentInLastMinute) / elapsedSeconds * 60
	targetRate := float64(r.maxPerMinute)

	if projectedRate > targetRate*1.2 {
		// send too fast, increase wait time
		r.waitTime = time.Duration(float64(r.waitTime) * 1)
		g.Log().Debug(context.Background(), "send rate too high (%.2f/minute), increase wait time to %v", projectedRate, r.waitTime)
	} else if projectedRate < targetRate*0.8 {
		// send too slow, reduce wait time (increase 20%)
		r.waitTime = time.Duration(float64(r.waitTime) * 0.8)
		g.Log().Debug(context.Background(), "send rate too low (%.2f/minute), reduce wait time to %v", projectedRate, r.waitTime)
	}

	// set min wait time, prevent too fast send
	minWait := time.Duration(2) * time.Millisecond
	if r.waitTime < minWait {
		r.waitTime = minWait
	}
}

// GetCurrentRate get current send rate
func (r *SimpleRateController) GetCurrentRate() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsedSeconds := now.Sub(r.lastResetTime).Seconds()
	if elapsedSeconds <= 0 {
		return 0
	}

	return float64(r.sentInLastMinute) / elapsedSeconds * 60
}

// GetMaxRate get max rate
func (r *SimpleRateController) GetMaxRate() int {
	return r.maxPerMinute
}
