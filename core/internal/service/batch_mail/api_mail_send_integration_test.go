package batch_mail

import (
	"billionmail-core/internal/model/entity"
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Constants validation
// ---------------------------------------------------------------------------

func TestApiMailSendConstants(t *testing.T) {
	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"StatusPending", StatusPending, 0},
		{"StatusSuccess", StatusSuccess, 2},
		{"StatusFailed", StatusFailed, 3},
		{"BatchSize", BatchSize, 1000},
		{"WorkerCount", WorkerCount, 20},
		{"QueryTimeout", QueryTimeout, 15},
		{"SendTimeout", SendTimeout, 15},
		{"LockKey", LockKey, "api_mail_queue_lock"},
		{"LockTimeout", LockTimeout, 120},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.got)
		})
	}
}

// ---------------------------------------------------------------------------
// CacheData: in-memory cache operations
// ---------------------------------------------------------------------------

func TestCacheData_EmptyInitialization(t *testing.T) {
	cache := &CacheData{
		ApiTemplates:   make(map[int]entity.ApiTemplates),
		EmailTemplates: make(map[int]entity.EmailTemplate),
		Contacts:       make(map[string]entity.Contact),
	}

	assert.Empty(t, cache.ApiTemplates)
	assert.Empty(t, cache.EmailTemplates)
	assert.Empty(t, cache.Contacts)
}

func TestCacheData_StoreAndRetrieve(t *testing.T) {
	cache := &CacheData{
		ApiTemplates:   make(map[int]entity.ApiTemplates),
		EmailTemplates: make(map[int]entity.EmailTemplate),
		Contacts:       make(map[string]entity.Contact),
	}

	// Store API template
	apiTmpl := entity.ApiTemplates{
		Id:         1,
		ApiName:    "test-api",
		TemplateId: 10,
		Subject:    "Test Subject",
		Addresser:  "sender@example.com",
		TrackOpen:  1,
		TrackClick: 1,
	}
	cache.ApiTemplates[1] = apiTmpl

	// Store email template
	emailTmpl := entity.EmailTemplate{
		Id:       10,
		TempName: "Test Template",
		Content:  "<html><body>Hello</body></html>",
	}
	cache.EmailTemplates[10] = emailTmpl

	// Store contact
	contact := entity.Contact{
		Id:    100,
		Email: "user@example.com",
	}
	cache.Contacts["user@example.com"] = contact

	// Retrieve and verify
	t.Run("api template", func(t *testing.T) {
		got, ok := cache.ApiTemplates[1]
		require.True(t, ok)
		assert.Equal(t, "test-api", got.ApiName)
		assert.Equal(t, 10, got.TemplateId)
	})

	t.Run("email template", func(t *testing.T) {
		got, ok := cache.EmailTemplates[10]
		require.True(t, ok)
		assert.Equal(t, "Test Template", got.TempName)
	})

	t.Run("contact", func(t *testing.T) {
		got, ok := cache.Contacts["user@example.com"]
		require.True(t, ok)
		assert.Equal(t, 100, got.Id)
	})

	t.Run("missing key returns zero value", func(t *testing.T) {
		_, ok := cache.ApiTemplates[999]
		assert.False(t, ok)
	})
}

func TestCacheData_OverwriteExistingEntry(t *testing.T) {
	cache := &CacheData{
		ApiTemplates:   make(map[int]entity.ApiTemplates),
		EmailTemplates: make(map[int]entity.EmailTemplate),
		Contacts:       make(map[string]entity.Contact),
	}

	cache.ApiTemplates[1] = entity.ApiTemplates{ApiName: "v1"}
	cache.ApiTemplates[1] = entity.ApiTemplates{ApiName: "v2"}

	assert.Equal(t, "v2", cache.ApiTemplates[1].ApiName)
}

// ---------------------------------------------------------------------------
// ApiMailLog structure
// ---------------------------------------------------------------------------

func TestApiMailLog_Structure(t *testing.T) {
	tests := []struct {
		name string
		log  ApiMailLog
	}{
		{
			name: "basic log entry",
			log: ApiMailLog{
				Id:        1,
				ApiId:     10,
				Recipient: "user@example.com",
				Addresser: "sender@example.com",
				MessageId: "abc123",
				Attribs:   map[string]string{"name": "John"},
			},
		},
		{
			name: "nil attribs",
			log: ApiMailLog{
				Id:        2,
				ApiId:     20,
				Recipient: "other@example.com",
				Addresser: "sender@example.com",
				MessageId: "def456",
				Attribs:   nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.log.Recipient)
			assert.NotEmpty(t, tt.log.Addresser)
		})
	}
}

// ---------------------------------------------------------------------------
// WorkerPool: Start and context cancellation
// ---------------------------------------------------------------------------

func TestWorkerPool_StartAndCancel(t *testing.T) {
	cache := &CacheData{
		ApiTemplates:   make(map[int]entity.ApiTemplates),
		EmailTemplates: make(map[int]entity.EmailTemplate),
		Contacts:       make(map[string]entity.Contact),
	}

	pool := &WorkerPool{
		workers: 3,
		jobs:    make(chan ApiMailLog, 10),
		cache:   cache,
	}

	ctx, cancel := context.WithCancel(context.Background())
	pool.Start(ctx)

	// Cancel to stop workers
	cancel()

	// Close jobs channel to let workers exit
	close(pool.jobs)

	// Wait for workers to finish
	pool.wg.Wait()

	// If we get here without deadlock, the test passes
}

func TestWorkerPool_WorkersReadFromJobsChannel(t *testing.T) {
	cache := &CacheData{
		ApiTemplates:   make(map[int]entity.ApiTemplates),
		EmailTemplates: make(map[int]entity.EmailTemplate),
		Contacts:       make(map[string]entity.Contact),
	}

	pool := &WorkerPool{
		workers: 1,
		jobs:    make(chan ApiMailLog, 5),
		cache:   cache,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool.Start(ctx)

	// Workers started, cancel to stop them
	cancel()
	close(pool.jobs)
	pool.wg.Wait()
}

// ---------------------------------------------------------------------------
// WorkerPool: multiple workers process concurrently
// ---------------------------------------------------------------------------

func TestWorkerPool_MultipleWorkersStartAndStop(t *testing.T) {
	tests := []struct {
		name    string
		workers int
	}{
		{"single worker", 1},
		{"five workers", 5},
		{"twenty workers", 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := &CacheData{
				ApiTemplates:   make(map[int]entity.ApiTemplates),
				EmailTemplates: make(map[int]entity.EmailTemplate),
				Contacts:       make(map[string]entity.Contact),
			}

			pool := &WorkerPool{
				workers: tt.workers,
				jobs:    make(chan ApiMailLog, 100),
				cache:   cache,
			}

			ctx, cancel := context.WithCancel(context.Background())
			pool.Start(ctx)

			cancel()
			close(pool.jobs)
			pool.wg.Wait() // no deadlock = success
		})
	}
}

// ---------------------------------------------------------------------------
// CacheData: concurrent read/write safety (map access pattern)
// ---------------------------------------------------------------------------

func TestCacheData_ConcurrentReadWrite(t *testing.T) {
	// Note: the production code does NOT use mutexes for CacheData.
	// This test verifies the pattern of "partition by goroutine" works
	// when each goroutine accesses its own key range.

	cache := &CacheData{
		ApiTemplates:   make(map[int]entity.ApiTemplates),
		EmailTemplates: make(map[int]entity.EmailTemplate),
		Contacts:       make(map[string]entity.Contact),
	}

	// Pre-populate to avoid concurrent map write
	for i := 0; i < 10; i++ {
		cache.ApiTemplates[i] = entity.ApiTemplates{Id: i}
	}

	var wg sync.WaitGroup
	// Concurrent reads are safe on Go maps
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			got, ok := cache.ApiTemplates[id]
			assert.True(t, ok)
			assert.Equal(t, id, got.Id)
		}(i)
	}
	wg.Wait()
}

// ---------------------------------------------------------------------------
// Addresser grouping pattern (from ProcessApiMailQueue)
// ---------------------------------------------------------------------------

func TestAddresserGrouping(t *testing.T) {
	logs := []ApiMailLog{
		{Id: 1, Addresser: "a@example.com", Recipient: "user1@test.com"},
		{Id: 2, Addresser: "a@example.com", Recipient: "user2@test.com"},
		{Id: 3, Addresser: "b@example.com", Recipient: "user3@test.com"},
		{Id: 4, Addresser: "a@example.com", Recipient: "user4@test.com"},
		{Id: 5, Addresser: "c@example.com", Recipient: "user5@test.com"},
	}

	addresserMap := make(map[string][]ApiMailLog)
	for _, log := range logs {
		addresserMap[log.Addresser] = append(addresserMap[log.Addresser], log)
	}

	assert.Len(t, addresserMap, 3)
	assert.Len(t, addresserMap["a@example.com"], 3)
	assert.Len(t, addresserMap["b@example.com"], 1)
	assert.Len(t, addresserMap["c@example.com"], 1)
}

// ---------------------------------------------------------------------------
// Batch splitting pattern
// ---------------------------------------------------------------------------

func TestBatchSplitting(t *testing.T) {
	tests := []struct {
		name       string
		totalItems int
		workers    int
		wantBatches int
	}{
		{"10 items / 5 workers", 10, 5, 5},
		{"7 items / 3 workers", 7, 3, 3},
		{"1 item / 5 workers", 1, 5, 1},
		{"0 items / 5 workers", 0, 5, 0},
		{"100 items / 5 workers", 100, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batchSize := 0
			if tt.workers > 0 {
				batchSize = (tt.totalItems + tt.workers - 1) / tt.workers
			}

			batches := 0
			for i := 0; i < tt.workers; i++ {
				start := i * batchSize
				if start >= tt.totalItems {
					break
				}
				end := start + batchSize
				if end > tt.totalItems {
					end = tt.totalItems
				}
				if start < end {
					batches++
				}
			}

			assert.Equal(t, tt.wantBatches, batches)
		})
	}
}
