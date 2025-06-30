package timers

import (
	"billionmail-core/internal/service/batch_mail"
	"billionmail-core/internal/service/collect"
	"billionmail-core/internal/service/domains"
	"billionmail-core/internal/service/fail2ban"
	"billionmail-core/internal/service/mail_boxes"
	"billionmail-core/internal/service/mail_service"
	"billionmail-core/internal/service/maillog_stat"
	"billionmail-core/internal/service/relay"
	"billionmail-core/internal/service/warmup"
	"context"
	"github.com/gogf/gf/os/gtimer"
	"github.com/gogf/gf/v2/frame/g"
	"time"
)

// Start Start initializes and starts all the necessary timers for the service.
func Start(ctx context.Context) (err error) {
	g.Log().Debug(ctx, "Starting timers...")

	// Normalize mailboxes to lowercase
	gtimer.AddOnce(500*time.Millisecond, func() {
		g.Log().Debug(ctx, "Mailboxes normalization started")

		err = mail_boxes.NormalizeMailboxes()

		if err != nil {
			g.Log().Warning(ctx, "NormalizeMailboxes failed: ", err)
			err = nil // Ignore the error to allow the service to continue
		} else {
			g.Log().Debug(ctx, "Mailboxes normalized successfully")
		}
	})

	// Start DNS Records checking
	// Ensure the DNS records are fresh on startup
	gtimer.AddOnce(500*time.Millisecond, func() {
		// repair DKIM signing config
		err = domains.RepairDKIMSigningConfig(ctx)

		if err != nil {
			g.Log().Warning(ctx, "RepairDKIMSigningConfig failed: ", err)
			err = nil
		}

		domains.FreshRecords(ctx)
	})
	// Check DNS Records every 5 minutes
	gtimer.Add(5*time.Minute, func() {
		domains.FreshRecords(ctx)
	})

	// Start maillog analysis
	gtimer.AddOnce(time.Second, func() {
		me := maillog_stat.NewMallogEventHandler("", 1*time.Second)
		me.Start()
	})

	// Start maillog aggregation
	gtimer.AddOnce(time.Second, func() {
		maillog_stat.AggregateMaillogsTask(1 * time.Minute)
	})

	// Start console panel baseurl update timer
	gtimer.AddOnce(100*time.Millisecond, func() {
		domains.UpdateBaseURL(ctx)
	})
	gtimer.Add(5*time.Minute, func() {
		domains.UpdateBaseURL(ctx)
	})

	g.Log().Debug(ctx, "Start timers complete")

	// Collect mail-sent total and mail-relay total
	gtimer.AddOnce(5*time.Second, func() {
		collect.Collect(ctx)
	})
	gtimer.Add(2*time.Hour, func() {
		collect.Collect(ctx)
	})

	// Fix Postfix main configuration and Rspamd DKIM signing config
	gtimer.AddOnce(5*time.Second, func() {
		mail_service.FixPostfixMainConfig(ctx)
		mail_service.FixRspamdDKIMSigningConfig(ctx)
		mail_service.FixDovecotSSLConfig(ctx)
	})

	// ========== Mail task processing: one executor per task ==========
	gtimer.Add(5*time.Second, func() {
		batch_mail.ProcessEmailTasks(ctx)
	})

	// Idle actuators are cleaned every 10 minutes
	gtimer.Add(10*time.Minute, func() {
		batch_mail.CleanupIdleExecutors()
		g.Log().Debug(ctx, "Idle task executors cleanup completed")
	})

	// Test the smtp relay configuration connection status
	gtimer.AddOnce(5*time.Second, func() {
		relay.UpdateRelayStatus(ctx)
	})

	gtimer.Add(5*time.Minute, func() {
		relay.UpdateRelayStatus(ctx)
	})

	gtimer.Add(24*time.Hour, func() {
		domains.AutoRenewSSL(ctx)
	})

	gtimer.Add(1*time.Minute, func() {
		//gtimer.Add(5*time.Second, func() {
		batch_mail.ProcessApiMailQueueWithLock(ctx)
	})

	// Sender IP warmup and mail provider warmup
	gtimer.Add(time.Hour, func() {
		warmup.SenderIpWarmup().PeriodicTask(ctx)
		warmup.SenderIpMailProvider().PeriodicTaskForProviders(ctx)
	})

	// fail2ban access logs detection
	gtimer.AddOnce(800*time.Millisecond, fail2ban.NewAccessLogDetection().Start)

	g.Log().Debug(ctx, "All timers started successfully")
	return nil
}
