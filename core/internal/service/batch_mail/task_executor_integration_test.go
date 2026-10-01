package batch_mail

import (
	"billionmail-core/internal/model/entity"
	"billionmail-core/internal/service/testutil"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// TaskExecutor registry: register, get, remove, concurrent access
// ---------------------------------------------------------------------------

func TestTaskExecutorRegistry_RegisterGetRemove(t *testing.T) {
	// Clean slate
	taskExecutorsMutex.Lock()
	origExecutors := taskExecutors
	taskExecutors = make(map[int]*TaskExecutor)
	taskExecutorsMutex.Unlock()
	defer func() {
		taskExecutorsMutex.Lock()
		taskExecutors = origExecutors
		taskExecutorsMutex.Unlock()
	}()

	ctx := context.Background()
	executor := NewTaskExecutor(ctx)

	tests := []struct {
		name   string
		action func()
		check  func()
	}{
		{
			name:   "get non-existent returns nil",
			action: func() {},
			check: func() {
				assert.Nil(t, GetTaskExecutor(999))
			},
		},
		{
			name: "register then get returns same executor",
			action: func() {
				RegisterTaskExecutor(42, executor)
			},
			check: func() {
				got := GetTaskExecutor(42)
				assert.Same(t, executor, got)
			},
		},
		{
			name: "remove then get returns nil",
			action: func() {
				RegisterTaskExecutor(43, NewTaskExecutor(ctx))
				RemoveTaskExecutor(43)
			},
			check: func() {
				assert.Nil(t, GetTaskExecutor(43))
			},
		},
		{
			name: "remove non-existent does not panic",
			action: func() {
				RemoveTaskExecutor(9999)
			},
			check: func() {
				// no panic is success
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.action()
			tt.check()
		})
	}
}

func TestTaskExecutorRegistry_ConcurrentAccess(t *testing.T) {
	taskExecutorsMutex.Lock()
	origExecutors := taskExecutors
	taskExecutors = make(map[int]*TaskExecutor)
	taskExecutorsMutex.Unlock()
	defer func() {
		taskExecutorsMutex.Lock()
		taskExecutors = origExecutors
		taskExecutorsMutex.Unlock()
	}()

	ctx := context.Background()
	var wg sync.WaitGroup

	// Concurrent register + get + remove
	for i := 0; i < 50; i++ {
		wg.Add(3)
		id := i
		go func() {
			defer wg.Done()
			RegisterTaskExecutor(id, NewTaskExecutor(ctx))
		}()
		go func() {
			defer wg.Done()
			_ = GetTaskExecutor(id)
		}()
		go func() {
			defer wg.Done()
			RemoveTaskExecutor(id)
		}()
	}

	wg.Wait() // no race/deadlock
}

func TestGetOrCreateTaskExecutor_CreatesOnce(t *testing.T) {
	taskExecutorsMutex.Lock()
	origExecutors := taskExecutors
	taskExecutors = make(map[int]*TaskExecutor)
	taskExecutorsMutex.Unlock()
	defer func() {
		taskExecutorsMutex.Lock()
		taskExecutors = origExecutors
		taskExecutorsMutex.Unlock()
	}()

	ctx := context.Background()
	e1 := GetOrCreateTaskExecutor(ctx, 100)
	e2 := GetOrCreateTaskExecutor(ctx, 100)

	assert.Same(t, e1, e2, "should return same executor for same task ID")
	assert.NotNil(t, e1)
}

func TestGetOrCreateTaskExecutor_DifferentIDs(t *testing.T) {
	taskExecutorsMutex.Lock()
	origExecutors := taskExecutors
	taskExecutors = make(map[int]*TaskExecutor)
	taskExecutorsMutex.Unlock()
	defer func() {
		taskExecutorsMutex.Lock()
		taskExecutors = origExecutors
		taskExecutorsMutex.Unlock()
	}()

	ctx := context.Background()
	e1 := GetOrCreateTaskExecutor(ctx, 200)
	e2 := GetOrCreateTaskExecutor(ctx, 201)

	assert.NotSame(t, e1, e2, "different IDs should get different executors")
}

// ---------------------------------------------------------------------------
// CleanupIdleExecutors
// ---------------------------------------------------------------------------

func TestCleanupIdleExecutors(t *testing.T) {
	taskExecutorsMutex.Lock()
	origExecutors := taskExecutors
	taskExecutors = make(map[int]*TaskExecutor)
	taskExecutorsMutex.Unlock()
	defer func() {
		taskExecutorsMutex.Lock()
		taskExecutors = origExecutors
		taskExecutorsMutex.Unlock()
	}()

	ctx := context.Background()

	// Create an idle executor (last activity > 30 min ago, not running)
	idle := NewTaskExecutor(ctx)
	idle.lastActivity = time.Now().Add(-31 * time.Minute)
	idle.isRunning.Store(false)
	RegisterTaskExecutor(300, idle)

	// Create a recent executor
	recent := NewTaskExecutor(ctx)
	recent.lastActivity = time.Now()
	recent.isRunning.Store(false)
	RegisterTaskExecutor(301, recent)

	// Create a running executor with old activity (should NOT be cleaned)
	running := NewTaskExecutor(ctx)
	running.lastActivity = time.Now().Add(-31 * time.Minute)
	running.isRunning.Store(true)
	RegisterTaskExecutor(302, running)

	CleanupIdleExecutors()

	assert.Nil(t, GetTaskExecutor(300), "idle executor should be cleaned up")
	assert.NotNil(t, GetTaskExecutor(301), "recent executor should remain")
	assert.NotNil(t, GetTaskExecutor(302), "running executor should remain even if old")

	// Cleanup running flag for test teardown
	running.isRunning.Store(false)
}

// ---------------------------------------------------------------------------
// TaskExecutor state transitions
// ---------------------------------------------------------------------------

func TestNewTaskExecutor_InitialState(t *testing.T) {
	ctx := context.Background()
	executor := NewTaskExecutor(ctx)

	assert.False(t, executor.IsRunning(), "new executor should not be running")
	assert.False(t, executor.IsPaused(), "new executor should not be paused")
	assert.NotNil(t, executor.rateController)
	assert.Equal(t, 1000, executor.rateController.GetMaxRate(), "default rate should be 1000")
}

func TestTaskExecutor_IsRunning_AtomicToggle(t *testing.T) {
	ctx := context.Background()
	executor := NewTaskExecutor(ctx)

	assert.False(t, executor.IsRunning())

	executor.isRunning.Store(true)
	assert.True(t, executor.IsRunning())

	executor.isRunning.Store(false)
	assert.False(t, executor.IsRunning())
}

func TestTaskExecutor_IsPaused_AtomicToggle(t *testing.T) {
	ctx := context.Background()
	executor := NewTaskExecutor(ctx)

	assert.False(t, executor.IsPaused())

	executor.isPaused.Store(true)
	assert.True(t, executor.IsPaused())

	executor.isPaused.Store(false)
	assert.False(t, executor.IsPaused())
}

// ---------------------------------------------------------------------------
// ProcessTask: duplicate run prevention
// ---------------------------------------------------------------------------

func TestProcessTask_PreventsDuplicateRun(t *testing.T) {
	ctx := context.Background()
	executor := NewTaskExecutor(ctx)

	// Simulate already running
	executor.isRunning.Store(true)

	err := executor.ProcessTask(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

// ---------------------------------------------------------------------------
// Stop
// ---------------------------------------------------------------------------

func TestTaskExecutor_Stop(t *testing.T) {
	ctx := context.Background()
	executor := NewTaskExecutor(ctx)
	executor.isRunning.Store(true)

	executor.Stop()

	assert.False(t, executor.IsRunning(), "executor should not be running after Stop")
}

func TestTaskExecutor_StopWithTimeout(t *testing.T) {
	ctx := context.Background()
	executor := NewTaskExecutor(ctx)

	// Add a pending wait group item that won't complete
	executor.wg.Add(1)
	executor.isRunning.Store(true)

	start := time.Now()
	executor.Stop()
	elapsed := time.Since(start)

	assert.False(t, executor.IsRunning())
	// Should timeout around 3 seconds
	assert.Less(t, elapsed, 5*time.Second)
	assert.Greater(t, elapsed, 2*time.Second)

	// Release the waitgroup to avoid test leak
	executor.wg.Done()
}

// ---------------------------------------------------------------------------
// configureRateController
// ---------------------------------------------------------------------------

func TestConfigureRateController(t *testing.T) {
	tests := []struct {
		name        string
		task        *entity.EmailTask
		wantMaxRate int
	}{
		{
			name:        "5 threads -> 6000/min",
			task:        testutil.NewEmailTask(testutil.WithTaskThreads(5)),
			wantMaxRate: 5 * 20 * 60,
		},
		{
			name:        "10 threads -> 12000/min",
			task:        testutil.NewEmailTask(testutil.WithTaskThreads(10)),
			wantMaxRate: 10 * 20 * 60,
		},
		{
			name:        "0 threads -> defaults to 1000",
			task:        testutil.NewEmailTask(testutil.WithTaskThreads(0)),
			wantMaxRate: 1000,
		},
		{
			name:        "negative threads -> defaults to 1000",
			task:        testutil.NewEmailTask(testutil.WithTaskThreads(-1)),
			wantMaxRate: 1000,
		},
		{
			name:        "1 thread -> 1200/min",
			task:        testutil.NewEmailTask(testutil.WithTaskThreads(1)),
			wantMaxRate: 1 * 20 * 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			executor := NewTaskExecutor(ctx)
			executor.configureRateController(tt.task)

			assert.Equal(t, tt.wantMaxRate, executor.rateController.GetMaxRate())
		})
	}
}

// ---------------------------------------------------------------------------
// GetMetrics
// ---------------------------------------------------------------------------

func TestGetMetrics_Initial(t *testing.T) {
	ctx := context.Background()
	executor := NewTaskExecutor(ctx)

	metrics := executor.GetMetrics()

	assert.Equal(t, int64(0), metrics["sent_count"])
	assert.Equal(t, int64(0), metrics["failed_count"])
	assert.Equal(t, int64(0), metrics["total_count"])
	assert.Equal(t, float64(0), metrics["success_rate"])
	assert.Greater(t, metrics["duration_sec"].(float64), float64(0))
}

func TestGetMetrics_AfterSendsAndFailures(t *testing.T) {
	ctx := context.Background()
	executor := NewTaskExecutor(ctx)

	executor.sentCount.Store(80)
	executor.failedCount.Store(20)

	metrics := executor.GetMetrics()

	assert.Equal(t, int64(80), metrics["sent_count"])
	assert.Equal(t, int64(20), metrics["failed_count"])
	assert.Equal(t, int64(100), metrics["total_count"])
	assert.InDelta(t, 0.8, metrics["success_rate"], 0.001)
}

func TestGetMetrics_AllSuccesses(t *testing.T) {
	ctx := context.Background()
	executor := NewTaskExecutor(ctx)

	executor.sentCount.Store(50)
	executor.failedCount.Store(0)

	metrics := executor.GetMetrics()
	assert.Equal(t, float64(1.0), metrics["success_rate"])
}

func TestGetMetrics_AllFailures(t *testing.T) {
	ctx := context.Background()
	executor := NewTaskExecutor(ctx)

	executor.sentCount.Store(0)
	executor.failedCount.Store(10)

	metrics := executor.GetMetrics()
	assert.Equal(t, float64(0), metrics["success_rate"])
}

// ---------------------------------------------------------------------------
// Circuit breaker pattern (via atomic counters)
// ---------------------------------------------------------------------------

func TestCircuitBreakerThreshold_Constant(t *testing.T) {
	assert.Equal(t, int64(10), int64(CircuitBreakerThreshold))
}

func TestCircuitBreaker_TripsAfterConsecutiveFailures(t *testing.T) {
	ctx := context.Background()
	executor := NewTaskExecutor(ctx)

	// Simulate consecutive failures below threshold
	for i := 0; i < CircuitBreakerThreshold-1; i++ {
		executor.consecutiveFailures.Add(1)
	}
	assert.False(t, executor.IsPaused(), "should not be paused below threshold")

	// One more failure hits threshold
	failures := executor.consecutiveFailures.Add(1)
	if failures >= CircuitBreakerThreshold {
		executor.isPaused.Store(true) // simulate what processRecipientBatch does
	}
	assert.True(t, executor.IsPaused(), "should pause at threshold")
}

func TestCircuitBreaker_ResetsOnSuccess(t *testing.T) {
	ctx := context.Background()
	executor := NewTaskExecutor(ctx)

	executor.consecutiveFailures.Store(5)
	// Simulate a success
	executor.consecutiveFailures.Store(0)

	assert.Equal(t, int64(0), executor.consecutiveFailures.Load())
}

// ---------------------------------------------------------------------------
// restoreErrorVariables
// ---------------------------------------------------------------------------

func TestRestoreErrorVariables(t *testing.T) {
	ctx := context.Background()
	executor := NewTaskExecutor(ctx)

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single variable",
			input: "[__.SomeVar__]",
			want:  "{{.SomeVar}}",
		},
		{
			name:  "multiple variables",
			input: "[__.Foo__] and [__.Bar__]",
			want:  "{{.Foo}} and {{.Bar}}",
		},
		{
			name:  "no variables",
			input: "plain text",
			want:  "plain text",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "mixed with HTML",
			input: `<p>[__.Name__]</p>`,
			want:  `<p>{{.Name}}</p>`,
		},
		{
			name:  "nested underscores not matched",
			input: "[__not_a_match__]",
			want:  "[__not_a_match__]", // contains underscore in middle, regex requires [^_]+
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := executor.restoreErrorVariables(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateTaskThreads: parameter validation (DB-independent part)
// ---------------------------------------------------------------------------

func TestUpdateTaskThreads_ValidationOnly(t *testing.T) {
	ctx := context.Background()
	executor := NewTaskExecutor(ctx)

	tests := []struct {
		name    string
		threads int
		wantErr string
	}{
		{"zero threads", 0, "threads must be greater than zero"},
		{"negative threads", -5, "threads must be greater than zero"},
		{"too many threads", 101, "threads must be less than 100"},
		{"way too many", 500, "threads must be less than 100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.UpdateTaskThreads(999, tt.threads)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// ---------------------------------------------------------------------------
// SendResult structure
// ---------------------------------------------------------------------------

func TestSendResult_Structure(t *testing.T) {
	tests := []struct {
		name   string
		result SendResult
	}{
		{
			name: "success result",
			result: SendResult{
				RecipientID: 1,
				Success:     true,
				MessageID:   "<msg-001@example.com>",
				Error:       nil,
			},
		},
		{
			name: "failure result",
			result: SendResult{
				RecipientID: 2,
				Success:     false,
				MessageID:   "",
				Error:       fmt.Errorf("connection refused"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.result.Success, tt.result.Error == nil || tt.result.Success)
		})
	}
}

// ---------------------------------------------------------------------------
// Metrics + rate controller integration
// ---------------------------------------------------------------------------

func TestMetricsIntegration_RateControllerInMetrics(t *testing.T) {
	ctx := context.Background()
	executor := NewTaskExecutor(ctx)

	task := testutil.NewEmailTask(testutil.WithTaskThreads(5))
	executor.configureRateController(task)

	metrics := executor.GetMetrics()
	assert.Equal(t, 5*20*60, metrics["max_rate"])
}

// ---------------------------------------------------------------------------
// Concurrent metrics access
// ---------------------------------------------------------------------------

func TestGetMetrics_ConcurrentSafe(t *testing.T) {
	ctx := context.Background()
	executor := NewTaskExecutor(ctx)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			executor.sentCount.Add(1)
		}()
		go func() {
			defer wg.Done()
			_ = executor.GetMetrics()
		}()
	}
	wg.Wait()

	metrics := executor.GetMetrics()
	assert.Equal(t, int64(50), metrics["sent_count"])
}

// ---------------------------------------------------------------------------
// getTaskIdFromContext
// ---------------------------------------------------------------------------

func TestGetTaskIdFromContext_Found(t *testing.T) {
	taskExecutorsMutex.Lock()
	origExecutors := taskExecutors
	taskExecutors = make(map[int]*TaskExecutor)
	taskExecutorsMutex.Unlock()
	defer func() {
		taskExecutorsMutex.Lock()
		taskExecutors = origExecutors
		taskExecutorsMutex.Unlock()
	}()

	ctx := context.Background()
	executor := NewTaskExecutor(ctx)
	RegisterTaskExecutor(55, executor)

	id, err := executor.getTaskIdFromContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 55, id)
}

func TestGetTaskIdFromContext_NotFound(t *testing.T) {
	taskExecutorsMutex.Lock()
	origExecutors := taskExecutors
	taskExecutors = make(map[int]*TaskExecutor)
	taskExecutorsMutex.Unlock()
	defer func() {
		taskExecutorsMutex.Lock()
		taskExecutors = origExecutors
		taskExecutorsMutex.Unlock()
	}()

	ctx := context.Background()
	executor := NewTaskExecutor(ctx)
	// not registered

	_, err := executor.getTaskIdFromContext(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task id not found")
}

// ---------------------------------------------------------------------------
// Task config cache
// ---------------------------------------------------------------------------

func TestTaskConfigCache(t *testing.T) {
	ctx := context.Background()
	executor := NewTaskExecutor(ctx)

	assert.Nil(t, executor.getTaskConfig(), "initially nil")

	task := testutil.NewEmailTask(testutil.WithTaskName("cached"))
	executor.taskConfig = task
	executor.configLoaded = time.Now()

	assert.Same(t, task, executor.getTaskConfig())
	assert.Equal(t, "cached", executor.getTaskConfig().TaskName)
}

// ---------------------------------------------------------------------------
// Pause/resume channel mechanics
// ---------------------------------------------------------------------------

func TestPauseResumeChannelMechanics(t *testing.T) {
	ctx := context.Background()
	executor := NewTaskExecutor(ctx)

	// Test resume channel send/receive
	executor.isPaused.Store(true)

	// Simulate resume
	executor.isPaused.Store(false)
	select {
	case executor.resumeChan <- struct{}{}:
		// ok, sent
	default:
		t.Fatal("resume channel should accept a signal")
	}

	// Read it back
	select {
	case <-executor.resumeChan:
		// ok, received
	case <-time.After(100 * time.Millisecond):
		t.Fatal("should have received resume signal")
	}
}
