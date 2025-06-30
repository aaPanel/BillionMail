package warmup

import (
	"context"
	"math"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

const (
	// dailyIncreaseFactor is the factor used for sending volume growth after the initial plan.
	// 1.3 means a 30% increase per day.
	dailyIncreaseFactor = 1.3
)

// WarmupStep defines the sending limits for a specific day in the warmup schedule.
type WarmupStep struct {
	DailyLimit  int
	HourlyLimit int
}

// warmupSchedule provides predefined sending volumes for the first few days of warmup.
// This is the base volume for each mail provider group.
var warmupSchedule = []WarmupStep{
	{DailyLimit: 1000, HourlyLimit: 100},   // Day 1
	{DailyLimit: 1500, HourlyLimit: 150},   // Day 2
	{DailyLimit: 2000, HourlyLimit: 200},   // Day 3
	{DailyLimit: 3000, HourlyLimit: 300},   // Day 4
	{DailyLimit: 5000, HourlyLimit: 500},   // Day 5
	{DailyLimit: 7000, HourlyLimit: 700},   // Day 6
	{DailyLimit: 10000, HourlyLimit: 1000}, // Day 7
}

// WarmupDailyUpstairsService manages the daily sending volume during the IP warmup period.
type WarmupDailyUpstairsService struct {
	ipWarmupService *SenderIpWarmupService
}

var insWarmupDailyUpstairsService = WarmupDailyUpstairsService{
	ipWarmupService: SenderIpWarmup(),
}

// WarmupDailyUpstairs returns the singleton instance of the service.
func WarmupDailyUpstairs() *WarmupDailyUpstairsService {
	return &insWarmupDailyUpstairsService
}

// GetSendingLimits calculates the base daily and hourly sending limits for a given IP.
// These limits are based on the number of days the IP has been in the warmup process.
// These are the base limits for each mail provider group and should be further adjusted by a provider-specific scoring factor.
func (s *WarmupDailyUpstairsService) GetSendingLimits(ctx context.Context, senderIp string) (dailyLimit int, hourlyLimit int, err error) {
	ipWarmupStatus, err := s.ipWarmupService.InitializeOrGetWarmupStatus(ctx, senderIp)
	if err != nil {
		g.Log().Errorf(ctx, "GetSendingLimits: Failed to get warmup status for IP %s: %v", senderIp, err)
		return 0, 0, err
	}

	if ipWarmupStatus.BeginTime == 0 {
		g.Log().Debugf(ctx, "GetSendingLimits: IP %s warmup has not started. Limits are 0.", senderIp)
		return 0, 0, nil // Warmup has not started
	}

	// Calculate the current day of the warmup process (1-based index).
	daysSinceWarmupStart := (gtime.Now().Timestamp()-ipWarmupStatus.BeginTime)/(24*60*60) + 1
	if daysSinceWarmupStart <= 0 {
		daysSinceWarmupStart = 1 // If it has started, it is at least day 1.
	}

	scheduleLen := len(warmupSchedule)

	if daysSinceWarmupStart <= int64(scheduleLen) {
		// Within the predefined schedule
		step := warmupSchedule[daysSinceWarmupStart-1]
		dailyLimit = step.DailyLimit
		hourlyLimit = step.HourlyLimit
		g.Log().Debugf(ctx, "GetSendingLimits: IP %s is on day %d of warmup. Using predefined limits: Daily=%d, Hourly=%d", senderIp, daysSinceWarmupStart, dailyLimit, hourlyLimit)
	} else {
		// After the predefined schedule, apply the growth factor
		lastStep := warmupSchedule[scheduleLen-1]
		daysAfterSchedule := daysSinceWarmupStart - int64(scheduleLen)

		dailyLimit = int(float64(lastStep.DailyLimit) * math.Pow(dailyIncreaseFactor, float64(daysAfterSchedule)))
		hourlyLimit = int(float64(lastStep.HourlyLimit) * math.Pow(dailyIncreaseFactor, float64(daysAfterSchedule)))
		g.Log().Debugf(ctx, "GetSendingLimits: IP %s is on day %d of warmup. Calculating limits with growth factor: Daily=%d, Hourly=%d", senderIp, daysSinceWarmupStart, dailyLimit, hourlyLimit)
	}

	return dailyLimit, hourlyLimit, nil
}
