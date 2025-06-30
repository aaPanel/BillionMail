package batch_mail

import (
	"billionmail-core/internal/model/entity"
	"billionmail-core/internal/service/domains"
	"billionmail-core/internal/service/mail_service"
	"billionmail-core/internal/service/maillog_stat"
	"billionmail-core/internal/service/public"
	"billionmail-core/internal/service/warmup"
	"context"
	"errors"
	"fmt"
	"github.com/gogf/gf/util/grand"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/panjf2000/ants/v2"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// global task executor map
	taskExecutors      = make(map[int]*TaskExecutor)
	taskExecutorsMutex sync.RWMutex

	// global rate limiter
	//globalLimiter = rate.NewLimiter(rate.Limit(5000), 100)
)

// GetTaskExecutor get task executor
func GetTaskExecutor(taskId int) *TaskExecutor {
	taskExecutorsMutex.RLock()
	defer taskExecutorsMutex.RUnlock()
	return taskExecutors[taskId]
}

// GetOrCreateTaskExecutor get or create task executor
func GetOrCreateTaskExecutor(ctx context.Context, taskId int) *TaskExecutor {
	taskExecutorsMutex.RLock()
	executor, exists := taskExecutors[taskId]
	taskExecutorsMutex.RUnlock()

	if !exists {
		taskExecutorsMutex.Lock()
		defer taskExecutorsMutex.Unlock()
		// double check, avoid race condition
		if executor, exists = taskExecutors[taskId]; !exists {
			executor = NewTaskExecutor(ctx)
			taskExecutors[taskId] = executor
		}
	}
	return executor
}

// RegisterTaskExecutor register task executor
func RegisterTaskExecutor(taskId int, executor *TaskExecutor) {
	taskExecutorsMutex.Lock()
	defer taskExecutorsMutex.Unlock()
	taskExecutors[taskId] = executor
}

// RemoveTaskExecutor remove task executor
func RemoveTaskExecutor(taskId int) {
	taskExecutorsMutex.Lock()
	defer taskExecutorsMutex.Unlock()

	if executor, exists := taskExecutors[taskId]; exists {
		// stop all operations of the executor
		executor.Stop()
		delete(taskExecutors, taskId)
	}
}

// CleanupIdleExecutors cleanup idle executors
func CleanupIdleExecutors() {
	taskExecutorsMutex.Lock()
	defer taskExecutorsMutex.Unlock()

	now := time.Now()
	for id, executor := range taskExecutors {
		if !executor.IsRunning() && now.Sub(executor.lastActivity) > 30*time.Minute {
			executor.Stop()
			delete(taskExecutors, id)
		}
	}
}

// ProcessEmailTasks
func ProcessEmailTasks(ctx context.Context) {
	// get pending tasks
	var tasks []*entity.EmailTask
	err := g.DB().Model("email_tasks").
		Where("task_process IN (0,1)").              // not started or running
		Where("pause", 0).                           // not paused
		Where("start_time <= ?", time.Now().Unix()). // start time has arrived
		Order("id ASC").
		Scan(&tasks)

	if err != nil {
		g.Log().Error(ctx, "Failed to get pending email tasks: %v", err)
		return
	}

	if len(tasks) == 0 {
		return
	}

	//g.Log().Debug(ctx, "Found %d pending email tasks", len(tasks))

	// process each task
	for _, task := range tasks {
		// check if task already has executor and is running
		executor := GetTaskExecutor(task.Id)
		if executor != nil && executor.IsRunning() {
			continue // skip running task
		}

		// create new executor
		newCtx := gctx.New()
		executor = NewTaskExecutor(newCtx)
		RegisterTaskExecutor(task.Id, executor)

		// start task processing
		go func(taskId int) {
			if err := executor.ProcessTask(newCtx); err != nil {
				g.Log().Error(newCtx, "Error processing task %d: %v", taskId, err)
			}
		}(task.Id)
	}
}

// TaskExecutor task executor
type TaskExecutor struct {
	// context and cancel function
	ctx    context.Context
	cancel context.CancelFunc

	// running status
	isRunning    atomic.Bool
	isPaused     atomic.Bool
	lastActivity time.Time

	// rate controller
	rateController *SimpleRateController

	// worker pool
	pool *ants.Pool
	wg   sync.WaitGroup

	// metrics
	sentCount   atomic.Int64
	failedCount atomic.Int64
	startTime   time.Time

	// pause/resume control
	pauseChan  chan struct{}
	resumeChan chan struct{}
}

// SendResult send result
type SendResult struct {
	RecipientID int
	Success     bool
	MessageID   string
	Error       error
}

// NewTaskExecutor create task executor
func NewTaskExecutor(ctx context.Context) *TaskExecutor {
	taskCtx, cancel := context.WithCancel(ctx)

	// Set server IP in context
	serverIP, _ := public.GetServerIP()

	taskCtx = context.WithValue(taskCtx, "serverIP", serverIP)

	executor := &TaskExecutor{
		ctx:            taskCtx,
		cancel:         cancel,
		lastActivity:   time.Now(),
		startTime:      time.Now(),
		pauseChan:      make(chan struct{}, 1),
		resumeChan:     make(chan struct{}, 1),
		rateController: NewSimpleRateController(1000),
	}

	return executor
}

// IsRunning check if task is running
func (e *TaskExecutor) IsRunning() bool {
	return e.isRunning.Load()
}

// IsPaused check if task is paused
func (e *TaskExecutor) IsPaused() bool {
	return e.isPaused.Load()
}

// ProcessTask
func (e *TaskExecutor) ProcessTask(ctx context.Context) error {
	// prevent duplicate running
	if !e.isRunning.CompareAndSwap(false, true) {
		return errors.New("task executor is already running")
	}

	defer e.isRunning.Store(false)

	// update activity time
	e.lastActivity = time.Now()

	// get task id
	taskId, err := e.getTaskIdFromContext(ctx)
	if err != nil {
		g.Log().Error(ctx, "failed to get task id: %v", err)
		return err
	}

	// get task info
	task, err := GetTaskInfo(ctx, taskId)
	if err != nil {
		return fmt.Errorf("failed to get task info: %w", err)
	}

	if task == nil || task.Id == 0 {
		return fmt.Errorf("task %d not found", taskId)
	}

	// check if task should run
	if task.TaskProcess == 2 { // completed
		return nil
	}

	currentTaskInfo, err := GetTaskInfo(ctx, taskId)
	if err == nil && currentTaskInfo != nil && currentTaskInfo.TaskProcess == 2 {

		return nil
	}

	// set pause status
	if task.Pause == 1 {
		e.isPaused.Store(true)
	}

	// check campaign warmup association
	warmupAssociated := false
	if warmupStat, _ := warmup.WarmupCampaign().GetWarmupStatusForCampaign(ctx, int64(taskId)); warmupStat != nil {
		warmupAssociated = true
	}

	e.ctx = context.WithValue(e.ctx, "warmupAssociated", warmupAssociated)

	// configure rate controller
	e.configureRateController(task)

	// get template info
	template, err := e.getTemplateInfo(ctx, task.TemplateId)
	if err != nil {
		g.Log().Error(ctx, "failed to get template: %v", err)
		return fmt.Errorf("failed to get template: %w", err)
	}

	// process email content
	emailContent := e.processEmailContent(ctx, template.Content, task)

	// create worker pool
	poolSize := task.Threads
	if poolSize <= 0 {
		poolSize = 6
	}

	// detailed record thread parameters
	g.Log().Info(ctx, "task %d: create worker pool, size: %d", task.Id, poolSize)

	// increase worker pool options, improve efficiency
	e.pool, err = ants.NewPool(poolSize,
		ants.WithPreAlloc(true),
		ants.WithPanicHandler(func(p interface{}) {
			g.Log().Error(ctx, "Worker panic: %v", p)
		}),
		ants.WithMaxBlockingTasks(poolSize*100), // allow more waiting tasks
		ants.WithNonblocking(false))             // blocking submit can improve stability

	if err != nil {
		g.Log().Error(ctx, "failed to create worker pool: %v", err)
		return fmt.Errorf("failed to create worker pool: %w", err)
	}

	defer e.pool.Release()

	// update task status to running
	if task.TaskProcess == 0 {
		if err := UpdateTaskProcessStatus(ctx, task.Id, 1); err != nil {
			g.Log().Error(ctx, "failed to update task status: %v", err)
			return fmt.Errorf("failed to update task status: %w", err)
		}
	}

	// start time
	startTime := time.Now()
	//g.Log().Info(ctx, "task %d: start processing, start time: %s", task.Id, startTime.Format("2006-01-02 15:04:05"))

	// process task
	if err := e.processTaskRecipients(ctx, task, emailContent); err != nil {
		g.Log().Error(ctx, "failed to process task: %v", err)
		if errors.Is(err, context.Canceled) {
			g.Log().Info(ctx, "task %d is canceled", task.Id)
			return nil
		}
		return err
	}

	// end time and duration
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	sentCount := e.sentCount.Load()

	// calculate average send rate
	var avgRate float64
	if duration.Seconds() > 0 {
		avgRate = float64(sentCount) / duration.Seconds() * 60
	}

	summaryMsg := fmt.Sprintf("task %d: processing completed, end time: %s, total duration: %.2f minutes, total sent: %d, average rate: %.1f emails/minute",
		task.Id, endTime.Format("2006-01-02 15:04:05"),
		duration.Minutes(), sentCount, avgRate)
	g.Log().Info(ctx, summaryMsg)

	currentTask, err := GetTaskInfo(ctx, taskId)
	if err != nil {
		errMsg := fmt.Sprintf("failed to get current task status: %v", err)
		g.Log().Error(ctx, errMsg)
		return fmt.Errorf("failed to get current task status: %w", err)
	}

	if currentTask.TaskProcess == 2 {
		return nil
	}

	completed, err := e.isTaskComplete(ctx, task.Id)
	if err != nil {
		errMsg := fmt.Sprintf("failed to check completion: %v", err)
		g.Log().Error(ctx, errMsg)
		return fmt.Errorf("failed to check completion: %w", err)
	}

	currentTask, err = GetTaskInfo(ctx, taskId)
	if err == nil && currentTask != nil {
		if currentTask.TaskProcess == 2 {

			return nil
		}
	}

	if completed {
		if err := UpdateTaskProcessStatus(ctx, task.Id, 2); err != nil {
			return fmt.Errorf("failed to update task status: %w", err)
		}
		completeMsg := fmt.Sprintf("task %d is successfully marked as completed", task.Id)
		g.Log().Info(ctx, completeMsg)
	}

	return nil
}

// Stop stop task executor
func (e *TaskExecutor) Stop() {
	if e.cancel != nil {
		e.cancel()
	}

	// wait for all work to complete
	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	// wait for 3 seconds
	select {
	case <-done:
		// work is completed
	case <-time.After(3 * time.Second):
		// timeout, force stop
	}

	// release worker pool
	if e.pool != nil {
		e.pool.Release()
	}

	e.isRunning.Store(false)
}

// PauseTask
func (e *TaskExecutor) PauseTask(taskId int) error {
	// if already paused, return immediately
	if e.isPaused.Load() {
		return nil
	}

	// set pause status
	e.isPaused.Store(true)

	// update database status
	if err := UpdateTaskPauseStatus(context.Background(), taskId, true); err != nil {
		e.isPaused.Store(false) // restore status
		return fmt.Errorf("failed to update task pause status: %w", err)
	}

	//g.Log().Info(context.Background(), "Task %d paused successfully", taskId)
	return nil
}

// ResumeTask
func (e *TaskExecutor) ResumeTask(taskId int) error {
	// if not paused, return immediately
	if !e.isPaused.Load() {
		return nil
	}

	// restore running status
	e.isPaused.Store(false)

	// send resume signal
	select {
	case e.resumeChan <- struct{}{}:
		// successfully send resume signal
	default:
		// channel may be full, recreate
		e.resumeChan = make(chan struct{}, 1)
		e.resumeChan <- struct{}{}
	}

	// update database status
	if err := UpdateTaskPauseStatus(context.Background(), taskId, false); err != nil {
		e.isPaused.Store(true) // restore status
		return fmt.Errorf("failed to update task resume status: %w", err)
	}

	//g.Log().Info(context.Background(), "Task %d resumed successfully", taskId)
	return nil
}

// getTaskIdFromContext
func (e *TaskExecutor) getTaskIdFromContext(ctx context.Context) (int, error) {
	for id, executor := range taskExecutors {
		if executor == e {
			return id, nil
		}
	}
	return 0, errors.New("task id not found in context")
}

// configureRateController
func (e *TaskExecutor) configureRateController(task *entity.EmailTask) {
	maxPerMinute := task.Threads * 20 * 60
	if maxPerMinute <= 0 {
		maxPerMinute = 1000
	}
	g.Log().Info(context.Background(), "task %d: initialize send rate - max %d emails per minute, threads: %d",
		task.Id, maxPerMinute, task.Threads)
	e.rateController = NewSimpleRateController(maxPerMinute)
}

// processTaskRecipients
func (e *TaskExecutor) processTaskRecipients(ctx context.Context, task *entity.EmailTask, emailContent string) error {
	const batchSize = 50
	var lastId = 0

	// add performance monitoring timer
	statsTicker := time.NewTicker(15 * time.Second)
	defer statsTicker.Stop()

	// record last check time and sent count
	lastCheckTime := time.Now()
	lastSentCount := int64(0)

	// start statistics goroutine
	go func() {
		for {
			select {
			case <-statsTicker.C:
				currentTime := time.Now()
				currentSent := e.sentCount.Load()

				// calculate interval send rate
				elapsedSeconds := currentTime.Sub(lastCheckTime).Seconds()
				sentInInterval := currentSent - lastSentCount
				ratePerMinute := float64(0)
				if elapsedSeconds > 0 {
					ratePerMinute = float64(sentInInterval) / elapsedSeconds * 60
				}

				// get task id
				taskId, _ := e.getTaskIdFromContext(ctx)

				infoMsg := fmt.Sprintf("task %d: performance stats - %.1f seconds sent %d emails, rate: %.1f emails/minute, goroutine pool usage: %d/%d",
					taskId, elapsedSeconds, sentInInterval, ratePerMinute,
					e.pool.Running(), e.pool.Cap())
				g.Log().Info(ctx, infoMsg)

				// update baseline value
				lastCheckTime = currentTime
				lastSentCount = currentSent

			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		// check if context is canceled
		select {
		case <-ctx.Done():
			g.Log().Info(ctx, "context canceled, stop task execution:", ctx.Err())
			return ctx.Err()
		default:
		}

		// check pause status
		if e.isPaused.Load() {
			g.Log().Debug(ctx, "Task %d is paused, waiting for resume signal", task.Id)

			// wait for resume signal
			select {
			case <-e.resumeChan:
				g.Log().Debug(ctx, "Task %d received resume signal", task.Id)
			case <-ctx.Done():
				g.Log().Info(ctx, "context canceled, stop task execution:", ctx.Err())
				return ctx.Err()
			}
		}

		// get a batch of recipients to send
		recipients, err := e.getNextRecipientBatch(ctx, task.Id, lastId, batchSize)
		if err != nil {
			return fmt.Errorf("failed to get recipients: %w", err)
		}

		// no more recipients, exit loop
		if len(recipients) == 0 {
			break
		}

		// record batch size
		//g.Log().Debug(ctx, "task %d: got %d recipients to send", task.Id, len(recipients))

		// update last id
		lastId = recipients[len(recipients)-1].Id

		// process this batch of recipients
		if err := e.processRecipientBatch(ctx, task, recipients, emailContent); err != nil {
			return err
		}

		// adjust send rate
		e.rateController.AdjustRate()
	}

	// wait for all tasks to complete
	//g.Log().Info(ctx, "task %d: all recipients processed, waiting for remaining send tasks to complete...", task.Id)
	e.wg.Wait()
	g.Log().Info(ctx, "task %d: all send tasks completed", task.Id)
	return nil
}

// getNextRecipientBatch
func (e *TaskExecutor) getNextRecipientBatch(ctx context.Context, taskId, lastId, batchSize int) ([]*entity.RecipientInfo, error) {
	var recipients []*entity.RecipientInfo

	// use ID paging, more efficient
	err := g.DB().Model("recipient_info").
		Where("task_id", taskId).
		Where("is_sent", 0).
		Where("sent_time = 0 OR sent_time < ?", time.Now().Unix()). // not sent yet
		Where("id > ?", lastId).
		Order("id ASC").
		Limit(batchSize).
		Scan(&recipients)

	return recipients, err
}

// processRecipientBatch
func (e *TaskExecutor) processRecipientBatch(ctx context.Context, task *entity.EmailTask, recipients []*entity.RecipientInfo, emailContent string) error {
	// create result channel, buffer size same as recipient count
	resultChan := make(chan *SendResult, len(recipients))

	// create wait group to track all send tasks
	var sendWg sync.WaitGroup

	// create a mutex and flag to control channel closure
	var mu sync.Mutex
	channelClosed := false

	// create a safe send function
	safeSend := func(result *SendResult) {
		mu.Lock()
		defer mu.Unlock()
		if !channelClosed {
			select {
			case resultChan <- result:
				// successfully sent
			case <-ctx.Done():
				// context canceled, no more send
			}
		}
	}

	// safe close channel function
	safeClose := func() {
		mu.Lock()
		defer mu.Unlock()
		if !channelClosed {
			channelClosed = true
			close(resultChan)
		}
	}

	updates := make(map[int]int)

	// submit send task for each recipient
	for _, recipient := range recipients {
		// check again if paused or canceled
		if e.isPaused.Load() {
			select {
			case <-e.resumeChan:
				// resumed
			case <-ctx.Done():
				safeClose() // safe close channel
				return ctx.Err()
			}
		}

		select {
		case <-ctx.Done():
			safeClose()
			return ctx.Err()
		default:
		}

		// wait for rate control
		if err := e.rateController.Wait(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				safeClose() // safe close channel
				return err
			}
			// record error but continue
			g.Log().Warning(ctx, "Rate limit wait error: %v", err)

		}

		// check if recipient is allowed to send with warmup
		if warmupAssociated, ok := e.ctx.Value("warmupAssociated").(bool); ok && warmupAssociated {
			if allow, waits, _ := warmup.RateLimiter().Allow(ctx, e.ctx.Value("serverIP").(string), public.GetMailProviderGroup(recipient.Recipient)); !allow {
				if waits > 0 {
					updates[recipient.Id] = waits * 2
				}
				// rate limit exceeded, skip this recipient
				g.Log().Warningf(ctx, "Rate limit exceeded for recipient %d, wait for %d seconds after retry, skipping", recipient.Id, waits)
				continue
			}
		}

		// create recipient copy to avoid closure problem
		recipientBak := recipient

		// add wait count
		e.wg.Add(1)
		sendWg.Add(1)

		// submit to worker pool
		err := e.pool.Submit(func() {
			defer e.wg.Done()
			defer sendWg.Done()
			// print task id
			//g.Log().Debug(ctx, "current task id", task.Id, "sender-", task.Addresser, "recipient-", recipientBak.Recipient)
			// personalize content
			personalized, _ := e.personalizeEmail(ctx, emailContent, task, recipientBak)
			//personalized := emailContent

			// send email
			result := e.sendEmail(ctx, task, recipientBak, personalized)

			// use sendEmailMock (Don't use in production)
			// result := e.sendEmailMock(ctx, task, recipientBak, personalized)

			// record send
			e.rateController.RecordSend()

			// update stats
			if result.Success {
				e.sentCount.Add(1)
			} else {
				e.failedCount.Add(1)
			}

			// safe send result
			safeSend(result)
		})

		if err != nil {
			e.wg.Done()   // reduce wait count
			sendWg.Done() // reduce send wait count

			// create failed result
			failResult := &SendResult{
				RecipientID: recipient.Id,
				Success:     false,
				Error:       fmt.Errorf("failed to submit to worker pool: %w", err),
			}

			// safe send result
			safeSend(failResult)
		}
	}

	if len(updates) > 0 {
		curTime := int(time.Now().Unix())
		data := make([]map[string]interface{}, 0, len(updates))
		i := 0
		for id, waits := range updates {
			data = append(data, g.Map{
				"id":         id,
				"task_id":    0,
				"recipient":  "",
				"message_id": "",
				"sent_time":  curTime + (waits * ((i % 10) + 1)),
			})
			i++
		}
		_, _ = g.DB().Ctx(ctx).Model("recipient_info").Data(data).OnConflict("id").OnDuplicate(g.Map{
			"sent_time": gdb.Raw("excluded.sent_time"),
		}).Save()
	}

	// all tasks submitted, start result processing and channel closure goroutine
	resultsDone := make(chan struct{})

	// start goroutine to close channel
	go func() {
		// wait for all send tasks to complete or context canceled
		sendDone := make(chan struct{})
		go func() {
			sendWg.Wait()
			close(sendDone)
		}()

		select {
		case <-sendDone:
			// all send tasks completed, safe close channel
			safeClose()
		case <-ctx.Done():
			// context canceled, safe close channel
			safeClose()
		}
	}()

	// start result processing goroutine
	go func() {
		e.processSendResults(ctx, resultChan)
		close(resultsDone)
	}()

	// wait for result processing to complete or context canceled
	select {
	case <-resultsDone:
		// result processing completed
		return nil
	case <-ctx.Done():
		// context canceled
		return ctx.Err()
	}
}

// processSendResults
func (e *TaskExecutor) processSendResults(ctx context.Context, resultChan <-chan *SendResult) {
	const batchSize = 50
	const flushInterval = 200 * time.Millisecond

	successResults := make([]*SendResult, 0, batchSize)
	failedIDs := make([]int, 0, batchSize)

	// create ticker to flush results
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	// flush function
	flushUpdates := func() {
		if len(successResults) == 0 && len(failedIDs) == 0 {
			return
		}

		// process success records
		if len(successResults) > 0 {
			// prepare batch update
			now := time.Now().Unix()

			// batch update recipient status
			// prepare batch update SQL
			if len(successResults) > 0 {
				// method 2: use SQL batch update
				ids := make([]interface{}, 0, len(successResults))
				messageIds := make(map[int]string, len(successResults))

				for _, result := range successResults {
					ids = append(ids, result.RecipientID)
					messageIds[result.RecipientID] = result.MessageID
				}

				// step 1: batch update is sent and sent time
				_, err := g.DB().Model("recipient_info").
					WhereIn("id", ids).
					Data(g.Map{
						"is_sent":   1,
						"sent_time": now,
					}).
					Update()

				if err != nil {
					g.Log().Error(ctx, "batch update recipient status failed: %v", err)
				} else {
					// step 2: update each recipient's message ID
					for id, messageID := range messageIds {
						// remove message_id external < >
						messageID = strings.Trim(messageID, "<>")
						_, err := g.DB().Model("recipient_info").
							Where("id", id).
							Data(g.Map{"message_id": messageID}).
							Update()

						if err != nil {
							g.Log().Error(ctx, "update recipient(ID:%d) message ID failed: %v", id, err)
						}
					}

					g.Log().Debug(ctx, "successfully batch updated %d recipient status", len(successResults))
				}
			}

			// clear success results
			successResults = successResults[:0]
		}

		// clear failed records
		if len(failedIDs) > 0 {
			failedIDs = failedIDs[:0]
		}
	}

	// main loop
	for {
		select {
		case result, ok := <-resultChan:
			if !ok {
				// channel closed, process remaining results
				flushUpdates()
				return
			}

			if result.Success {
				successResults = append(successResults, result)
			} else {
				failedIDs = append(failedIDs, result.RecipientID)

				g.Log().Warning(ctx, "send email to recipient %d failed: %v",
					result.RecipientID, result.Error)
			}

			// reach batch processing size, flush
			if len(successResults)+len(failedIDs) >= batchSize {
				flushUpdates()
			}

		case <-ticker.C:
			// flush periodically
			flushUpdates()

		case <-ctx.Done():
			// context canceled, process remaining results
			flushUpdates()
			return
		}
	}
}

// getTemplateInfo get template info
func (e *TaskExecutor) getTemplateInfo(ctx context.Context, templateId int) (*entity.EmailTemplate, error) {
	var template entity.EmailTemplate

	err := g.DB().Model("email_templates").
		Where("id", templateId).
		Scan(&template)

	if err != nil {
		return nil, err
	}

	if template.Id == 0 {
		return nil, fmt.Errorf("template %d not found", templateId)
	}

	return &template, nil
}

// processEmailContent
func (e *TaskExecutor) processEmailContent(ctx context.Context, content string, task *entity.EmailTask) string {
	// process unsubscribe link
	if task.Unsubscribe == 1 {
		// __UNSUBSCRIBE_URL__  {{ UnsubscribeURL }}
		if !strings.Contains(content, "__UNSUBSCRIBE_URL__") && !strings.Contains(content, "{{ UnsubscribeURL . }}") {
			content = public.AddUnsubscribeButton(content)
		}

		content = strings.ReplaceAll(content, "__UNSUBSCRIBE_URL__", "{{ UnsubscribeURL . }}")
	}
	return content
}

// personalizeEmail personalize email content
func (e *TaskExecutor) personalizeEmail(ctx context.Context, content string, task *entity.EmailTask, recipient *entity.RecipientInfo) (string, string) {

	var contact entity.Contact
	err := g.DB().Model("bm_contacts").Where("email", recipient.Recipient).Scan(&contact)
	if err != nil {
		g.Log().Error(ctx, "get contact info failed: %v", err)
	}

	var emailtask entity.EmailTask
	err = g.DB().Model("email_tasks").Where("id", task.Id).Scan(&emailtask)
	if err != nil {
		g.Log().Error(ctx, "get task info failed: %v", err)
		emailtask = *task
	}

	// Unsubscribe
	var renderedContent, renderedSubject string
	engine := GetTemplateEngine()

	if task.Unsubscribe == 1 {
		domain := domains.GetBaseURLBySender(task.Addresser)
		unsubscribeURL := fmt.Sprintf("%s/api/unsubscribe", domain)
		groupURL := fmt.Sprintf("%s/api/unsubscribe/user_group", domain)

		jwtToken, err := GenerateUnsubscribeJWT(
			recipient.Recipient,
			task.TemplateId,
			task.Id,
		)
		if err != nil {
			g.Log().Error(ctx, "generate unsubscribe JWT failed: %v", err)
			jwtToken = ""
		}

		// generate unsubscribe URL
		unsubscribeJumpURL := fmt.Sprintf("%s/unsubscribe.html?jwt=%s&email=%s&url_type=%s&url_unsubscribe=%s",
			domain, jwtToken, recipient.Recipient, groupURL, unsubscribeURL)

		// render email content
		renderedContent, err = engine.RenderEmailTemplate(ctx, content, &contact, &emailtask, unsubscribeJumpURL)
		if err != nil {
			g.Log().Error(ctx, "render email content failed: %v", err)
			renderedContent = content
		}

		// render email subject
		renderedSubject, err = engine.RenderEmailTemplate(ctx, emailtask.Subject, &contact, &emailtask, unsubscribeJumpURL)
		if err != nil {
			g.Log().Error(ctx, "render email subject failed: %v", err)
			renderedSubject = emailtask.Subject
		}
	} else {
		// if unsubscribe is not enabled, render email content
		renderedContent, err = engine.RenderEmailTemplate(ctx, content, &contact, &emailtask, "")
		if err != nil {
			g.Log().Error(ctx, "render email content failed: %v", err)
			renderedContent = content
		}

		// render email subject
		renderedSubject, err = engine.RenderEmailTemplate(ctx, emailtask.Subject, &contact, &emailtask, "")
		if err != nil {
			g.Log().Error(ctx, "render email subject failed: %v", err)
			renderedSubject = emailtask.Subject
		}
	}

	return renderedContent, renderedSubject
}

// sendEmail send email
func (e *TaskExecutor) sendEmail(ctx context.Context, task *entity.EmailTask, recipient *entity.RecipientInfo, content string) *SendResult {
	// check if context is canceled
	select {
	case <-ctx.Done():
		return &SendResult{
			RecipientID: recipient.Id,
			Success:     false,
			Error:       ctx.Err(),
		}
	default:
		// continue execution
	}

	// get rendered content and subject
	renderedContent, renderedSubject := e.personalizeEmail(ctx, content, task, recipient)

	sender, err := mail_service.NewEmailSenderWithLocal(task.Addresser)
	if err != nil {
		g.Log().Error(ctx, "create email sender failed: %v", err)
		return &SendResult{
			RecipientID: recipient.Id,
			Success:     false,
			Error:       fmt.Errorf("create email sender failed: %w", err),
		}
	}
	defer sender.Close()
	// set message ID
	messageID := sender.GenerateMessageID()

	//Tracking emails
	baseURL := domains.GetBaseURLBySender(task.Addresser)
	mail_tracker := maillog_stat.NewMailTracker(renderedContent, task.Id, messageID, recipient.Recipient, baseURL)
	mail_tracker.TrackLinks()
	mail_tracker.AppendTrackingPixel()
	renderedContent = mail_tracker.GetHTML()

	// create email message with rendered subject
	message := mail_service.NewMessage(renderedSubject, renderedContent)
	message.SetMessageID(messageID)

	// set sender display name
	if task.FullName != "" {
		message.SetRealName(task.FullName)
	}

	// send email
	err = sender.Send(message, []string{recipient.Recipient})
	if err != nil {
		g.Log().Error(ctx, "send email to %s failed: %v", recipient.Recipient, err)
		return &SendResult{
			RecipientID: recipient.Id,
			Success:     false,
			Error:       fmt.Errorf("send email failed: %w", err),
		}
	}

	return &SendResult{
		RecipientID: recipient.Id,
		MessageID:   messageID,
		Success:     true,
		Error:       nil,
	}
}

// sendEmailMock simulates sending an email and records it in the database.
func (e *TaskExecutor) sendEmailMock(ctx context.Context, task *entity.EmailTask, recipient *entity.RecipientInfo, content string) *SendResult {
	// Check if the context is canceled
	select {
	case <-ctx.Done():
		return &SendResult{
			RecipientID: recipient.Id,
			Success:     false,
			Error:       ctx.Err(),
		}
	default:
		// Continue execution
	}

	// Get the rendered content and subject
	renderedContent, renderedSubject := e.personalizeEmail(ctx, content, task, recipient)

	sender, err := mail_service.NewEmailSenderWithLocal(task.Addresser)
	if err != nil {
		g.Log().Error(ctx, "Failed to create email sender: %v", err)
		return &SendResult{
			RecipientID: recipient.Id,
			Success:     false,
			Error:       fmt.Errorf("failed to create email sender: %w", err),
		}
	}
	defer sender.Close()
	// Set message ID
	messageID := sender.GenerateMessageID()

	// Track email
	baseURL := domains.GetBaseURLBySender(task.Addresser)
	mail_tracker := maillog_stat.NewMailTracker(renderedContent, task.Id, messageID, recipient.Recipient, baseURL)
	mail_tracker.TrackLinks()
	mail_tracker.AppendTrackingPixel()
	renderedContent = mail_tracker.GetHTML()

	// Create email message with rendered subject
	message := mail_service.NewMessage(renderedSubject, renderedContent)
	message.SetMessageID(messageID)

	// Set sender display name
	if task.FullName != "" {
		message.SetRealName(task.FullName)
	}

	// We will create a log entry and save it, instead of sending.
	// This simulates a successful send.
	postfixMessageID := strings.ToUpper("TEST_" + grand.S(11))
	nowMillis := time.Now().UnixMilli()

	// 1. Create MailSender record
	senderRecord := &maillog_stat.MailSender{
		MailRecord: maillog_stat.MailRecord{
			PostfixMessageID: postfixMessageID,
			LogTimeMillis:    nowMillis,
		},
		Sender: task.Addresser,
		Size:   int64(len(renderedContent)),
	}
	_, err = g.DB().Model("mailstat_senders").InsertIgnore(senderRecord)
	if err != nil {
		g.Log().Warningf(ctx, "sendEmailMock: failed to insert mailstat_senders: %v", err)
	}

	// 2. Create MailMessageID record
	messageIDRecord := &maillog_stat.MailMessageID{
		MailRecord: maillog_stat.MailRecord{
			PostfixMessageID: postfixMessageID,
			LogTimeMillis:    nowMillis,
		},
		MessageID: strings.Trim(messageID, "<>"),
	}
	_, err = g.DB().Model("mailstat_message_ids").InsertIgnore(messageIDRecord)
	if err != nil {
		g.Log().Warningf(ctx, "sendEmailMock: failed to insert mailstat_message_ids: %v", err)
	}

	// 3. Create MailSendRecord record
	sendRecord := &maillog_stat.MailSendRecord{
		MailRecord: maillog_stat.MailRecord{
			PostfixMessageID: postfixMessageID,
			LogTimeMillis:    nowMillis,
		},
		Recipient:    recipient.Recipient,
		MailProvider: public.GetMailProviderGroup(recipient.Recipient),
		Status:       "sent",
		Delay:        0.1,              // Mock value
		Delays:       "0/0/0.1/0",      // Mock value
		Dsn:          "2.0.0",          // Mock value for successful send
		Relay:        "mock.relay.com", // Mock value
		Description:  "250 2.0.0 OK",   // Mock value
	}

	_, err = g.DB().Model("mailstat_send_mails").Data(sendRecord).Insert()
	if err != nil {
		g.Log().Errorf(ctx, "sendEmailMock: failed to insert mailstat_send_mails: %v", err)
		return &SendResult{
			RecipientID: recipient.Id,
			Success:     false,
			Error:       fmt.Errorf("sendEmailMock: failed to save record: %w", err),
		}
	}

	// 4. Simulate email open and click
	// Simulate a 50% open rate
	if grand.Intn(100) < 50 {
		openTimeMillis := nowMillis + int64(grand.Intn(3600*1000)) // Simulate opening within 1 hour of sending
		_, err = g.DB().Model("mailstat_opened").Insert(g.Map{
			"campaign_id":        task.Id,
			"log_time_millis":    openTimeMillis,
			"recipient":          recipient.Recipient,
			"message_id":         strings.Trim(messageID, "<>"),
			"postfix_message_id": postfixMessageID,
		})
		if err != nil {
			g.Log().Warningf(ctx, "sendEmailMock: failed to insert mailstat_opened: %v", err)
		}

		// If opened, simulate a 20% click rate
		if grand.Intn(100) < 20 {
			clickTimeMillis := openTimeMillis + int64(grand.Intn(600*1000)) // Simulate clicking within 10 minutes of opening
			_, err = g.DB().Model("mailstat_clicked").Insert(g.Map{
				"campaign_id":        task.Id,
				"log_time_millis":    clickTimeMillis,
				"recipient":          recipient.Recipient,
				"message_id":         strings.Trim(messageID, "<>"),
				"postfix_message_id": postfixMessageID,
			})
			if err != nil {
				g.Log().Warningf(ctx, "sendEmailMock: failed to insert mailstat_clicked: %v", err)
			}
		}
	}

	return &SendResult{
		RecipientID: recipient.Id,
		MessageID:   messageID,
		Success:     true,
		Error:       nil,
	}
}

// isTaskComplete check if task is complete
func (e *TaskExecutor) isTaskComplete(ctx context.Context, taskId int) (bool, error) {
	type CountResult struct {
		TotalCount int `json:"total_count"`
		SentCount  int `json:"sent_count"`
	}

	var result CountResult

	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {

		err := tx.Model("recipient_info").
			Fields("COUNT(1) as total_count, SUM(CASE WHEN is_sent = 1 THEN 1 ELSE 0 END) as sent_count").
			Where("task_id", taskId).
			Scan(&result)
		return err
	})

	if err != nil {
		return false, err
	}

	// if there are no recipients, task is not complete
	if result.TotalCount == 0 {
		return false, nil
	}

	// if sent count equals or exceeds total count, task is complete
	return result.SentCount >= result.TotalCount, nil
}

// GetMetrics get execution metrics
func (e *TaskExecutor) GetMetrics() map[string]interface{} {
	duration := time.Since(e.startTime).Seconds()
	sent := e.sentCount.Load()
	failed := e.failedCount.Load()
	total := sent + failed

	var successRate float64

	if total > 0 {
		successRate = float64(sent) / float64(total)
	}

	return map[string]interface{}{
		"sent_count":    sent,
		"failed_count":  failed,
		"total_count":   total,
		"success_rate":  successRate,
		"current_speed": e.rateController.GetCurrentRate(),
		"max_rate":      e.rateController.GetMaxRate(),
		"duration_sec":  duration,
	}
}

func (e *TaskExecutor) UpdateTaskThreads(taskId int, threads int) error {
	// parameter validation
	if threads <= 0 {
		return fmt.Errorf("threads must be greater than zero")
	}

	if threads > 100 {
		return fmt.Errorf("threads must be less than 100")
	}

	// get task info
	task, err := GetTaskInfo(context.Background(), taskId)
	if err != nil {
		return fmt.Errorf("get task info failed: %w", err)
	}

	if task == nil || task.Id == 0 {
		return fmt.Errorf("task %d not found", taskId)
	}

	// record current pool status
	var oldPoolSize int
	var runningWorkers int
	if e.pool != nil {
		oldPoolSize = e.pool.Cap()
		runningWorkers = e.pool.Running()
	}

	// new threads
	newThreads := threads
	// calculate new rate limit - 20 emails per thread per second
	targetSendPerThreadPerSecond := 20
	newRate := newThreads * targetSendPerThreadPerSecond * 60

	// create new rate controller
	e.rateController = NewSimpleRateController(newRate)

	// if task is running, adjust pool size
	if e.pool != nil && e.IsRunning() {
		// if new pool size is not equal to current pool size, create new pool
		if newThreads != oldPoolSize {
			// if request to decrease capacity, but current running number is close to new capacity, output warning
			if newThreads < oldPoolSize && runningWorkers > int(float64(newThreads)*0.8) {
				warningMsg := fmt.Sprintf("task %d: request pool size (%d) is less than current running workers (%d), may cause task queue",
					taskId, newThreads, runningWorkers)
				g.Log().Warning(context.Background(), warningMsg)
			}

			// create new pool
			newPool, err := ants.NewPool(newThreads,
				ants.WithPreAlloc(true),
				ants.WithPanicHandler(func(p interface{}) {
					g.Log().Error(context.Background(), "Worker panic: %v", p)
				}),
				ants.WithMaxBlockingTasks(newThreads*200),
				ants.WithNonblocking(false))

			if err != nil {
				g.Log().Error(context.Background(), "task %d: create new pool failed: %v", taskId, err)
				// even if creating new pool failed, we will still update rate controller
			} else {
				// get old pool reference
				oldPool := e.pool

				// replace with new pool
				e.pool = newPool

				// safely close old pool
				go func(pool *ants.Pool, oldSize int, oldRunning int) {
					// calculate wait time - adjust dynamically based on current active workers
					waitTime := 1 * time.Second
					if oldRunning > 0 {
						// add 1 second for every 10 active workers, minimum 1 second, maximum 10 seconds
						waitSecs := 1 + oldRunning/10
						if waitSecs > 10 {
							waitSecs = 10
						}
						waitTime = time.Duration(waitSecs) * time.Second
					}

					waitInfoMsg := fmt.Sprintf("task %d: wait %d seconds to release old pool (running: %d/%d)",
						taskId, int(waitTime.Seconds()), oldRunning, oldSize)
					g.Log().Info(context.Background(), waitInfoMsg)

					time.Sleep(waitTime)
					pool.Release()
					releaseMsg := fmt.Sprintf("task %d: old pool released", taskId)
					g.Log().Info(context.Background(), releaseMsg)
				}(oldPool, oldPoolSize, runningWorkers)
			}
		} else {
			keepMsg := fmt.Sprintf("task %d: pool size keep unchanged (%d), only adjust rate controller", taskId, oldPoolSize)
			g.Log().Info(context.Background(), keepMsg)
		}
	}

	// update threads in database
	_, err = g.DB().Model("email_tasks").
		Where("id", taskId).
		Data(g.Map{"threads": newThreads}).
		Update()

	if err != nil {

		return fmt.Errorf("task %d: update database threads failed: %w", taskId, err)
	}

	return nil
}
