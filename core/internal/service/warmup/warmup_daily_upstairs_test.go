package warmup

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWarmupSchedule_Values(t *testing.T) {
	// Verify the predefined warmup schedule is correctly defined
	expected := []WarmupStep{
		{DailyLimit: 1000, HourlyLimit: 100},
		{DailyLimit: 1500, HourlyLimit: 150},
		{DailyLimit: 2000, HourlyLimit: 200},
		{DailyLimit: 3000, HourlyLimit: 300},
		{DailyLimit: 5000, HourlyLimit: 500},
		{DailyLimit: 7000, HourlyLimit: 700},
		{DailyLimit: 10000, HourlyLimit: 1000},
	}

	assert.Equal(t, len(expected), len(warmupSchedule))

	for i, step := range warmupSchedule {
		assert.Equal(t, expected[i].DailyLimit, step.DailyLimit, "day %d daily limit", i+1)
		assert.Equal(t, expected[i].HourlyLimit, step.HourlyLimit, "day %d hourly limit", i+1)
	}
}

func TestWarmupSchedule_MonotonicallyIncreasing(t *testing.T) {
	for i := 1; i < len(warmupSchedule); i++ {
		assert.Greater(t, warmupSchedule[i].DailyLimit, warmupSchedule[i-1].DailyLimit,
			"daily limit should increase from day %d to day %d", i, i+1)
		assert.Greater(t, warmupSchedule[i].HourlyLimit, warmupSchedule[i-1].HourlyLimit,
			"hourly limit should increase from day %d to day %d", i, i+1)
	}
}

func TestWarmupSchedule_HourlyLimitRatio(t *testing.T) {
	for i, step := range warmupSchedule {
		ratio := float64(step.HourlyLimit) / float64(step.DailyLimit)
		assert.InDelta(t, 0.1, ratio, 0.01,
			"day %d: hourly limit should be ~10%% of daily limit", i+1)
	}
}

func TestDailyIncreaseFactor(t *testing.T) {
	assert.Equal(t, 1.3, dailyIncreaseFactor)
}

func TestGrowthAfterSchedule(t *testing.T) {
	// Test the math.Pow growth formula used in GetSendingLimits
	// after the predefined schedule ends.
	lastStep := warmupSchedule[len(warmupSchedule)-1]

	tests := []struct {
		name             string
		daysAfterSchedule int
		expectedDaily    int
		expectedHourly   int
	}{
		{
			name:              "day 1 after schedule",
			daysAfterSchedule: 1,
			expectedDaily:     int(float64(lastStep.DailyLimit) * math.Pow(dailyIncreaseFactor, 1)),
			expectedHourly:    int(float64(lastStep.HourlyLimit) * math.Pow(dailyIncreaseFactor, 1)),
		},
		{
			name:              "day 2 after schedule",
			daysAfterSchedule: 2,
			expectedDaily:     int(float64(lastStep.DailyLimit) * math.Pow(dailyIncreaseFactor, 2)),
			expectedHourly:    int(float64(lastStep.HourlyLimit) * math.Pow(dailyIncreaseFactor, 2)),
		},
		{
			name:              "day 5 after schedule",
			daysAfterSchedule: 5,
			expectedDaily:     int(float64(lastStep.DailyLimit) * math.Pow(dailyIncreaseFactor, 5)),
			expectedHourly:    int(float64(lastStep.HourlyLimit) * math.Pow(dailyIncreaseFactor, 5)),
		},
		{
			name:              "day 10 after schedule",
			daysAfterSchedule: 10,
			expectedDaily:     int(float64(lastStep.DailyLimit) * math.Pow(dailyIncreaseFactor, 10)),
			expectedHourly:    int(float64(lastStep.HourlyLimit) * math.Pow(dailyIncreaseFactor, 10)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dailyLimit := int(float64(lastStep.DailyLimit) * math.Pow(dailyIncreaseFactor, float64(tt.daysAfterSchedule)))
			hourlyLimit := int(float64(lastStep.HourlyLimit) * math.Pow(dailyIncreaseFactor, float64(tt.daysAfterSchedule)))

			assert.Equal(t, tt.expectedDaily, dailyLimit)
			assert.Equal(t, tt.expectedHourly, hourlyLimit)
			assert.Greater(t, dailyLimit, lastStep.DailyLimit,
				"growth day %d should exceed schedule last step", tt.daysAfterSchedule)
			assert.Greater(t, hourlyLimit, lastStep.HourlyLimit,
				"growth day %d should exceed schedule last step", tt.daysAfterSchedule)
		})
	}
}

func TestGrowthAfterSchedule_Compounding(t *testing.T) {
	lastStep := warmupSchedule[len(warmupSchedule)-1]
	prevDaily := lastStep.DailyLimit

	for day := 1; day <= 20; day++ {
		daily := int(float64(lastStep.DailyLimit) * math.Pow(dailyIncreaseFactor, float64(day)))
		assert.Greater(t, daily, prevDaily,
			"day %d after schedule should be greater than day %d", day, day-1)
		prevDaily = daily
	}
}

func TestWarmupConstants(t *testing.T) {
	assert.Equal(t, 45, defaultWarmupPeriodDays)
	assert.Equal(t, 20, minScoreForSending)
	assert.Equal(t, 20, minSendVolumeForFullScoring)
}

func TestWarmupStep_Struct(t *testing.T) {
	step := WarmupStep{
		DailyLimit:  5000,
		HourlyLimit: 500,
	}
	assert.Equal(t, 5000, step.DailyLimit)
	assert.Equal(t, 500, step.HourlyLimit)
}
