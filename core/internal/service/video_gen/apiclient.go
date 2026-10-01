package video_gen

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sony/gobreaker/v2"
	"golang.org/x/time/rate"
)

// RateLimitedClient wraps http.Client with per-API rate limiter + circuit breaker.
type RateLimitedClient struct {
	client  *http.Client
	limiter *rate.Limiter
	cb      *gobreaker.CircuitBreaker[*http.Response]
}

// NewRateLimitedClient creates a client with rate limiting and circuit breaking.
// rps: requests per second, burst: max burst size, timeout: HTTP request timeout.
func NewRateLimitedClient(name string, rps float64, burst int, timeout time.Duration) *RateLimitedClient {
	return &RateLimitedClient{
		client:  &http.Client{Timeout: timeout},
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
		cb: gobreaker.NewCircuitBreaker[*http.Response](gobreaker.Settings{
			Name:        name,
			MaxRequests: 2,                // half-open: allow 2 probe requests
			Interval:    60 * time.Second, // rolling window for failure counting
			Timeout:     30 * time.Second, // time in open state before half-open
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures >= 5
			},
		}),
	}
}

// Do executes an HTTP request with rate limiting and circuit breaking.
// Blocks until rate limiter allows, then executes through circuit breaker.
func (c *RateLimitedClient) Do(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter: %w", err)
	}

	resp, err := c.cb.Execute(func() (*http.Response, error) {
		return c.client.Do(req)
	})
	if err != nil {
		return nil, fmt.Errorf("circuit breaker [%s]: %w", c.cb.Name(), err)
	}
	return resp, nil
}
