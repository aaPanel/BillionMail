package timers

import (
	"billionmail-core/internal/service/batch_mail"
	"billionmail-core/internal/service/domains"
	"billionmail-core/internal/service/maillog_stat"
	"billionmail-core/internal/service/relay"
	"context"
	"github.com/gogf/gf/os/gtimer"
	"github.com/gogf/gf/v2/frame/g"
	"time"
)

// Start 启动所有定时任务 (主入口函数)
func Start(ctx context.Context) (err error) {
	g.Log().Debug(ctx, "Starting timers...")

	// Start DNS Records checking
	// Ensure the DNS records are fresh on startup
	gtimer.AddOnce(5*time.Second, func() {
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
	gtimer.AddOnce(5*time.Second, func() {
		me := maillog_stat.NewMallogEventHandler("", 1*time.Second)
		me.Start()
	})

	// Start maillog aggregation
	gtimer.AddOnce(5*time.Second, func() {
		maillog_stat.AggregateMaillogsTask(1 * time.Minute)
	})

	// Start console panel baseurl update timer
	gtimer.AddOnce(5*time.Second, func() {
		domains.UpdateBaseURL(ctx)
	})
	gtimer.Add(5*time.Minute, func() {
		domains.UpdateBaseURL(ctx)
	})

	g.Log().Debug(ctx, "Start timers complete")

	gtimer.AddOnce(5*time.Second, func() {
		me := maillog_stat.NewMallogEventHandler("", 1*time.Second)
		me.Start()
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

	g.Log().Debug(ctx, "All timers started successfully")
	return nil
}
