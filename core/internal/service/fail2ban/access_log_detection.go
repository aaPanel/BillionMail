package fail2ban

import (
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfsnotify"
	"os/exec"
	"path/filepath"
	"time"
)

type AccessLogDetection struct {
	logPath string
	ch      chan struct{}
}

// NewAccessLogDetection creates a new AccessLogDetection instance with the specified log path.
func NewAccessLogDetection() *AccessLogDetection {
	return &AccessLogDetection{
		logPath: public.AbsPath("logs"),
		ch:      make(chan struct{}, 2),
	}
}

// Start begins the detection process for access logs.
func (ald *AccessLogDetection) Start() {
	_, err := gfsnotify.Add(ald.logPath, func(event *gfsnotify.Event) {

		if event.IsCreate() {
			if event.Path == filepath.Join(ald.logPath, time.Now().Format("access-20060102.log")) {
				g.Log().Debug(context.Background(), "MallogEventHandler: access log file created, triggering detection")
				ald.ch <- struct{}{}
			}
		}
	})

	if err != nil {
		g.Log().Error(context.Background(), "AccessLogDetection: error adding file watcher", err)
		return
	}

	g.Log().Debug(context.Background(), "AccessLogDetection: start event")

	for {
		select {
		case <-ald.ch:
			g.Log().Debug(context.Background(), "AccessLogDetection: timer triggered, restarting fail2ban service")

			if !public.FileExists("/restart_fail2ban.sh") {
				g.Log().Warning(context.Background(), "AccessLogDetection: /restart_fail2ban.sh script does not exist, skipping restart")
				continue
			}

			cmd := exec.Command("bash", "/restart_fail2ban.sh")
			if res, err := cmd.CombinedOutput(); err != nil {
				g.Log().Warning(context.Background(), "AccessLogDetection: error restarting fail2ban service", err, string(res))
			} else {
				g.Log().Debug(context.Background(), "AccessLogDetection: restarting fail2ban service", string(res))
			}
		}
	}
}
