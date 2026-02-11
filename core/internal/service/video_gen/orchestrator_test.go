package video_gen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSignals(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect map[string]bool
	}{
		{"empty", "", map[string]bool{}},
		{"single", "no_chat", map[string]bool{"no_chat": true}},
		{"multiple", "no_chat,running_ads,owner_email", map[string]bool{
			"no_chat":     true,
			"running_ads": true,
			"owner_email": true,
		}},
		{"with spaces", " no_chat , running_ads ", map[string]bool{
			"no_chat":     true,
			"running_ads": true,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSignals(tt.input)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestJobStatusConstants(t *testing.T) {
	assert.Equal(t, "pending", JobPending)
	assert.Equal(t, "processing", JobProcessing)
	assert.Equal(t, "completed", JobCompleted)
	assert.Equal(t, "failed", JobFailed)
}

func TestWithRetry_ImmediateSuccess(t *testing.T) {
	calls := 0
	result, err := withRetry(func() (string, error) {
		calls++
		return "ok", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "ok", result)
	assert.Equal(t, 1, calls)
}

func TestWithRetry_EventualSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping retry test in short mode")
	}

	calls := 0
	result, err := withRetry(func() (string, error) {
		calls++
		if calls < 3 {
			return "", fmt.Errorf("transient error")
		}
		return "recovered", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "recovered", result)
	assert.Equal(t, 3, calls)
}

func TestWithRetry_AllFail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping retry test in short mode")
	}

	calls := 0
	_, err := withRetry(func() (string, error) {
		calls++
		return "", fmt.Errorf("permanent error")
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permanent error")
	assert.Equal(t, maxRetries+1, calls)
}
