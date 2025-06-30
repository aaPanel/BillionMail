package warmup

import (
	"billionmail-core/internal/service/public"
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

const (
	// cacheTTLForLimits is the cache time for sending limits.
	// We cache for a short period to avoid hitting the database on every request,
	// while still allowing the system to pick up changes in scores and progress relatively quickly.
	cacheTTLForLimits = 5 * time.Minute

	// counterExpireInDay is the expiration time for the daily counter.
	// Set to 2 days.
	counterExpireInDay = 24 * time.Hour
)

// tokenBucketLua is an atomic script for the token bucket rate limiter.
// It refills tokens based on elapsed time and consumes one if available.
const tokenBucketLua = `
-- KEYS[1] - The key for the token bucket hash
-- ARGV[1] - The capacity of the bucket
-- ARGV[2] - The refill rate (tokens per second)
-- ARGV[3] - The current timestamp (seconds)
-- ARGV[4] - The number of tokens to consume

local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])

local bucket_info = redis.call('HGETALL', key)
local last_tokens = capacity
local last_refill_time = now

if #bucket_info > 0 then
    for i = 1, #bucket_info, 2 do
        if bucket_info[i] == 'tokens' then
            last_tokens = tonumber(bucket_info[i+1])
        elseif bucket_info[i] == 'last_refill_time' then
            last_refill_time = tonumber(bucket_info[i+1])
        end
    end
end

local time_passed = now - last_refill_time
local new_tokens = time_passed * rate

local current_tokens = math.min(capacity, last_tokens + new_tokens)

if current_tokens >= requested then
    current_tokens = current_tokens - requested
    redis.call('HSET', key, 'tokens', current_tokens, 'last_refill_time', now)
    -- Set an expiration for safety, e.g., twice the time it takes to fill an empty bucket
    local expiry = 86400 -- Default 24 hours
    if rate > 0 then
        expiry = math.ceil(capacity / rate * 2)
    end
    redis.call('EXPIRE', key, expiry)
    return 1
else
    -- Not enough tokens, don't update anything, just return 0
    return 0
end
`

// RateLimiterService provides rate limiting functionality for IP warmup.
// It uses Redis for high-performance, distributed counting and caching.
type RateLimiterService struct {
	providerService *SenderIpMailProviderService
}

var insRateLimiterService = RateLimiterService{
	providerService: SenderIpMailProvider(),
}

// RateLimiter returns the singleton instance of RateLimiterService.
func RateLimiter() *RateLimiterService {
	return &insRateLimiterService
}

// Allow checks if sending is allowed for the given sender IP and recipient.
// It uses a hybrid approach:
// 1. Daily limit: using Redis's variable window counter to ensure the absolute daily quota is not exceeded.
// 2. Hourly limit: using the token bucket algorithm to smooth the sending rate and allow short-term bursts.
func (s *RateLimiterService) Allow(ctx context.Context, senderIp string, recipientEmail string) (allow bool, waits int, err error) {
	// Default wait time is 300 seconds (5 minutes), if not exceeded limit, it will be 0.
	waits = 300

	// 1. Determine the mail service provider group from the recipient email.
	mailProviderGroup := public.GetMailProviderGroup(recipientEmail)

	// 2. Get the sending limits (daily and hourly) for this IP and service provider group.
	var dailyLimit, hourlyLimit int
	dailyLimit, hourlyLimit, err = s.getCachedLimitsForProvider(ctx, senderIp, mailProviderGroup)
	if err != nil {
		g.Log().Errorf(ctx, "RateLimiter: Failed to get limits for IP %s, Group %s: %v", senderIp, mailProviderGroup, err)
		return
	}

	if dailyLimit <= 0 || hourlyLimit <= 0 {
		g.Log().Debugf(ctx, "RateLimiter: Sending denied for IP %s, Group %s because limits are zero (Daily: %d, Hourly: %d)", senderIp, mailProviderGroup, dailyLimit, hourlyLimit)
		return
	}

	// 3. Check the daily limit (variable window counter)
	var dailyCount int64
	now := time.Now()
	dailyKey := fmt.Sprintf("warmup:vw:d:%s:%s", senderIp, mailProviderGroup)

	dailyCount, err = g.Redis().Incr(ctx, dailyKey)
	if err != nil {
		g.Log().Errorf(ctx, "RateLimiter: Redis INCR failed for daily key %s: %v", dailyKey, err)
		return
	}
	if dailyCount == 1 {
		// New daily counter, set expiration time
		_, _ = g.Redis().Expire(ctx, dailyKey, int64(counterExpireInDay.Seconds()))
	}

	if dailyCount > int64(dailyLimit) {
		g.Log().Debugf(ctx, "RateLimiter: Daily limit exceeded for IP %s, Group %s. Count: %d, Limit: %d", senderIp, mailProviderGroup, dailyCount, dailyLimit)
		// Note: We don't roll back the counter here, because the exceeded requests have indeed occurred.
		// Keeping the counter in an exceeded state prevents subsequent requests.
		return
	}

	// 4. Check the hourly limit (token bucket)
	hourlyKey := fmt.Sprintf("warmup:tb:h:%s:%s", senderIp, mailProviderGroup)
	capacity := float64(hourlyLimit)
	rate := capacity / 3600.0 // Tokens replenished per second

	res, err := g.Redis().Eval(ctx, tokenBucketLua, 1, []string{hourlyKey}, []interface{}{
		capacity,
		rate,
		now.Unix(),
		1, // Consume 1 token
	})
	if err != nil {
		g.Log().Errorf(ctx, "RateLimiter: Token bucket script failed for IP %s, Group %s: %v", senderIp, mailProviderGroup, err)
		// Script failed, status unknown, roll back daily count and deny
		_, _ = g.Redis().Decr(ctx, dailyKey)
		return
	}

	allowed := res.Int() == 1
	if !allowed {
		g.Log().Debugf(ctx, "RateLimiter: Hourly rate limit exceeded (token bucket) for IP %s, Group %s.", senderIp, mailProviderGroup)
		// Not enough tokens in the bucket, roll back daily count
		_, _ = g.Redis().Decr(ctx, dailyKey)
		// Calculate the wait time/seconds for the next available token
		waits = int(float64(1)/rate) + 1
		return
	}

	// 5. If all checks pass, allow sending.
	allow = true
	waits = 0 // No wait time, request is allowed

	g.Log().Debugf(ctx, "RateLimiter: Allow send for IP %s, Group %s. Daily count: %d/%d", senderIp, mailProviderGroup, dailyCount, dailyLimit)
	return
}

// getCachedLimitsForProvider retrieves the sending limits, utilizing cache to avoid frequent database calls.
func (s *RateLimiterService) getCachedLimitsForProvider(ctx context.Context, senderIp, mailProviderGroup string) (dailyLimit, hourlyLimit int, err error) {
	cacheKey := fmt.Sprintf("warmup:limits:%s:%s", senderIp, mailProviderGroup)

	// First, try to get it from the cache.
	cachedVal, err := g.Redis().Get(ctx, cacheKey)
	if err == nil && !cachedVal.IsNil() {
		// Cache hit
		limitsJson := gjson.New(cachedVal.String())
		dailyLimit = limitsJson.Get("d").Int()
		hourlyLimit = limitsJson.Get("h").Int()
		g.Log().Debugf(ctx, "RateLimiter: Cache HIT for limits on IP %s, Group %s. Daily: %d, Hourly: %d", senderIp, mailProviderGroup, dailyLimit, hourlyLimit)
		return
	}

	// Cache miss, fetch from service.
	g.Log().Debugf(ctx, "RateLimiter: Cache MISS for limits on IP %s, Group %s. Fetching from service.", senderIp, mailProviderGroup)
	dailyLimit, hourlyLimit, err = s.providerService.GetAdjustedSendingLimitsForProvider(ctx, senderIp, mailProviderGroup)
	if err != nil {
		return 0, 0, err
	}

	// Store in cache for future requests.
	limits := g.Map{
		"d": dailyLimit,
		"h": hourlyLimit,
	}
	limitsStr, _ := gjson.New(limits).ToJsonString()
	err = g.Redis().SetEX(ctx, cacheKey, limitsStr, int64(cacheTTLForLimits.Seconds()))
	if err != nil {
		// Log the error, but don't fail the operation, as we have already fetched the limits.
		g.Log().Warningf(ctx, "RateLimiter: Failed to set cache for limits on IP %s, Group %s: %v", senderIp, mailProviderGroup, err)
	}

	return
}
