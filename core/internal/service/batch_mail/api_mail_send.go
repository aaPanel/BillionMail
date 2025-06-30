// core/internal/service/batch_mail/api_mail_send.go
package batch_mail

import (
	"billionmail-core/internal/model/entity"
	"billionmail-core/internal/service/domains"
	"billionmail-core/internal/service/mail_service"
	"billionmail-core/internal/service/maillog_stat"
	"billionmail-core/internal/service/public"
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"strings"
	"sync"
	"time"
)

const (
	StatusPending = 0
	StatusSuccess = 2
	StatusFailed  = 3

	BatchSize    = 100 // Number of items to process per batch
	WorkerCount  = 20  // Number of worker goroutines
	QueryTimeout = 15  // Query timeout duration (seconds)
	SendTimeout  = 15  // Send timeout duration (seconds)
	LockKey      = "api_mail_queue_lock"
	LockTimeout  = 60 // Lock timeout duration (seconds)
)

type ApiMailLog struct {
	Id        int64
	ApiId     int
	Recipient string
	Addresser string
	MessageId string
}

// Cache data structure
type CacheData struct {
	ApiTemplates   map[int]entity.ApiTemplates
	EmailTemplates map[int]entity.EmailTemplate
	Contacts       map[string]entity.Contact
}

// Worker pool structure
type WorkerPool struct {
	workers int
	jobs    chan ApiMailLog
	wg      sync.WaitGroup
	cache   *CacheData
}

// Create a new worker pool
func NewWorkerPool(workers int, cache *CacheData) *WorkerPool {
	return &WorkerPool{
		workers: workers,
		jobs:    make(chan ApiMailLog, BatchSize),
		cache:   cache,
	}
}

// Start the worker pool
func (p *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(ctx)
	}
}

// Worker goroutine
func (p *WorkerPool) worker(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case log, ok := <-p.jobs:
			if !ok {
				return
			}
			p.processMail(ctx, log)
		}
	}
}

// Process a single email
func (p *WorkerPool) processMail(ctx context.Context, log ApiMailLog) {
	ctx, cancel := context.WithTimeout(ctx, SendTimeout*time.Second)
	defer cancel()

	// Get data from cache, if not exist, fetch from database
	apiTemplate, ok := p.cache.ApiTemplates[log.ApiId]
	if !ok {
		err := g.DB().Model("api_templates").
			Where("id", log.ApiId).
			Ctx(ctx).
			Scan(&apiTemplate)
		if err != nil {
			updateLogStatus(ctx, log.Id, StatusFailed, fmt.Sprintf("Failed to get API template: %v", err))
			return
		}
		p.cache.ApiTemplates[log.ApiId] = apiTemplate
	}

	emailTemplate, ok := p.cache.EmailTemplates[apiTemplate.TemplateId]
	if !ok {
		err := g.DB().Model("email_templates").
			Where("id", apiTemplate.TemplateId).
			Ctx(ctx).
			Scan(&emailTemplate)
		if err != nil {
			updateLogStatus(ctx, log.Id, StatusFailed, fmt.Sprintf("Failed to get email template: %v", err))
			return
		}
		p.cache.EmailTemplates[apiTemplate.TemplateId] = emailTemplate
	}

	contact, ok := p.cache.Contacts[log.Recipient]
	if !ok {
		err := g.DB().Model("bm_contacts").
			Where("email", log.Recipient).
			Ctx(ctx).
			Scan(&contact)
		if err != nil {
			updateLogStatus(ctx, log.Id, StatusFailed, fmt.Sprintf("Failed to get contact info: %v", err))
			return
		}
		p.cache.Contacts[log.Recipient] = contact
	}

	// Process email content and subject
	content, subject := processMailContentAndSubject(ctx, emailTemplate.Content, apiTemplate.Subject, &apiTemplate, contact, log)

	// Send email
	err := sendApiMail(ctx, &apiTemplate, subject, content, log)
	if err != nil {
		updateLogStatus(ctx, log.Id, StatusFailed, fmt.Sprintf("Failed to send email: %v", err))
		return
	}

	updateLogStatus(ctx, log.Id, StatusSuccess, "")
}

// Queue processing with distributed lock
func ProcessApiMailQueueWithLock(ctx context.Context) {
	// Attempt to acquire the distributed lock
	locked, err := g.Redis().SetNX(ctx, LockKey, "1")
	if err != nil {
		g.Log().Error(ctx, "Failed to acquire lock:", err)
		return
	}
	if !locked {
		g.Log().Info(ctx, "Another instance is processing the mail queue")
		return
	}

	// Set expiration time
	g.Redis().Expire(ctx, LockKey, int64(LockTimeout))

	defer func() {
		g.Redis().Del(ctx, LockKey)
	}()

	// Process the email queue
	ProcessApiMailQueue(ctx)
}

// Batch process emails
func ProcessApiMailQueue(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout*time.Second)
	defer cancel()

	// Process all pending emails in batche
	offset := 0
	for {
		// Get a batch of pending emails
		var mailLogs []ApiMailLog
		err := g.DB().Model("api_mail_logs").
			Where("status = ?", StatusPending).
			Order("id ASC").
			Offset(offset).
			Limit(BatchSize).
			Ctx(ctx).
			Scan(&mailLogs)

		if err != nil || len(mailLogs) == 0 {
			break
		}

		// Preload data
		cache, err := preloadData(ctx, mailLogs)
		if err != nil {
			g.Log().Error(ctx, "Failed to preload data:", err)
			break
		}

		pool := NewWorkerPool(WorkerCount, cache)
		pool.Start(ctx)

		// Add emails to the worker pool
		for _, log := range mailLogs {
			select {
			case <-ctx.Done():
				return
			case pool.jobs <- log:
			}
		}

		// Close the worker pool and wait for completion
		close(pool.jobs)
		pool.wg.Wait()

		offset += len(mailLogs)
		if len(mailLogs) < BatchSize {
			break
		}
	}
}

func preloadData(ctx context.Context, logs []ApiMailLog) (*CacheData, error) {
	// Collect all necessary IDs
	apiIds := make([]int, 0, len(logs))
	recipientEmails := make([]string, 0, len(logs))
	for _, log := range logs {
		apiIds = append(apiIds, log.ApiId)
		recipientEmails = append(recipientEmails, log.Recipient)
	}
	// Batch query API templates
	var apiTemplates []entity.ApiTemplates
	err := g.DB().Model("api_templates").
		WhereIn("id", apiIds).
		Ctx(ctx).
		Scan(&apiTemplates)
	if err != nil {
		return nil, fmt.Errorf("failed to load API templates: %v", err)
	}
	// Collect template IDs
	templateIds := make([]int, 0, len(apiTemplates))
	for _, t := range apiTemplates {
		templateIds = append(templateIds, t.TemplateId)
	}
	// Batch query email templates
	var emailTemplates []entity.EmailTemplate
	err = g.DB().Model("email_templates").
		WhereIn("id", templateIds).
		Ctx(ctx).
		Scan(&emailTemplates)
	if err != nil {
		return nil, fmt.Errorf("failed to load email templates: %v", err)
	}
	// Batch query contacts
	var contacts []entity.Contact
	err = g.DB().Model("bm_contacts").
		WhereIn("email", recipientEmails).
		Ctx(ctx).
		Scan(&contacts)
	if err != nil {
		return nil, fmt.Errorf("failed to load contacts: %v", err)
	}
	// Build cache
	cache := &CacheData{
		ApiTemplates:   make(map[int]entity.ApiTemplates),
		EmailTemplates: make(map[int]entity.EmailTemplate),
		Contacts:       make(map[string]entity.Contact),
	}

	for _, t := range apiTemplates {
		cache.ApiTemplates[t.Id] = t
	}
	for _, t := range emailTemplates {
		cache.EmailTemplates[t.Id] = t
	}
	for _, c := range contacts {
		cache.Contacts[c.Email] = c
	}

	return cache, nil
}

// send email
func sendApiMail(ctx context.Context, apiTemplate *entity.ApiTemplates, subject string, content string, log ApiMailLog) error {

	// create email sender
	sender, err := mail_service.NewEmailSenderWithLocal(log.Addresser)
	if err != nil {
		updateLogStatus(ctx, log.Id, 3, err.Error())
		return err
	}
	defer sender.Close()

	// generate message ID
	messageId := "<" + log.MessageId + ">"

	// add 1 billion to prevent conflict with marketing task id
	baseURL := domains.GetBaseURLBySender(log.Addresser)
	apiTemplate_id := apiTemplate.Id + 1000000000
	mailTracker := maillog_stat.NewMailTracker(content, apiTemplate_id, messageId, log.Recipient, baseURL)
	mailTracker.TrackLinks()
	mailTracker.AppendTrackingPixel()
	content = mailTracker.GetHTML()

	// create email message
	message := mail_service.NewMessage(subject, content)
	message.SetMessageID(messageId)

	// set sender display name
	if apiTemplate.FullName != "" {
		message.SetRealName(apiTemplate.FullName)
	}
	// send email

	err = sender.Send(message, []string{log.Recipient})
	if err != nil {
		errorMessage := public.LangCtx(ctx, "Failed to send email: {}", err.Error())
		updateLogStatus(ctx, log.Id, 3, errorMessage)
		return err
	}
	updateLogStatus(ctx, log.Id, 2, "")
	return nil
}

func updateLogStatus(ctx context.Context, id int64, status int, errorMsg string) {
	_, err := g.DB().Model("api_mail_logs").
		Where("id", id).
		Update(g.Map{
			"status":        status,
			"error_message": errorMsg,
			"send_time":     time.Now().Unix(),
		})
	if err != nil {
		g.Log().Error(ctx, "Failed to update mail log status:", err)
	}
}

// process mail content and subject
func processMailContentAndSubject(ctx context.Context, content, subject string, apiTemplate *entity.ApiTemplates, contact entity.Contact, log ApiMailLog) (string, string) {
	// unsubscribe link processing
	if apiTemplate.Unsubscribe == 1 {
		if !strings.Contains(content, "__UNSUBSCRIBE_URL__") && !strings.Contains(content, "{{ UnsubscribeURL . }}") {
			content = public.AddUnsubscribeButton(content)
		}
		domain := domains.GetBaseURLBySender(log.Addresser)
		unsubscribeURL := fmt.Sprintf("%s/api/unsubscribe", domain)
		groupURL := fmt.Sprintf("%s/api/unsubscribe/user_group", domain)
		jwtToken, _ := GenerateUnsubscribeJWT(log.Recipient, apiTemplate.TemplateId, apiTemplate.Id)
		unsubscribeJumpURL := fmt.Sprintf("%s/unsubscribe.html?jwt=%s&email=%s&url_type=%s&url_unsubscribe=%s", domain, jwtToken, log.Recipient, groupURL, unsubscribeURL)

		if contact.Id > 0 {
			engine := GetTemplateEngine()
			renderedContent, err := engine.RenderEmailTemplate(ctx, content, &contact, nil, unsubscribeJumpURL)
			if err == nil {
				content = renderedContent
			}
			renderedSubject, err := engine.RenderEmailTemplate(ctx, subject, &contact, nil, unsubscribeJumpURL)
			if err == nil {
				subject = renderedSubject
			}
		} else {
			content = strings.ReplaceAll(content, "{{ UnsubscribeURL . }}", unsubscribeJumpURL)
		}
	} else if contact.Id > 0 {
		engine := GetTemplateEngine()
		renderedContent, err := engine.RenderEmailTemplate(ctx, content, &contact, nil, "")
		if err == nil {
			content = renderedContent
		}
		renderedSubject, err := engine.RenderEmailTemplate(ctx, subject, &contact, nil, "")
		if err == nil {
			subject = renderedSubject
		}
	}
	return content, subject
}
