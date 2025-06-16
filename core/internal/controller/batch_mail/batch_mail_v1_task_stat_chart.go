package batch_mail

import (
	"billionmail-core/internal/service/batch_mail"
	"billionmail-core/internal/service/public"
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/batch_mail/v1"
)

func (c *ControllerV1) TaskStatChart(ctx context.Context, req *v1.TaskStatChartReq) (res *v1.TaskStatChartRes, err error) {
	res = &v1.TaskStatChartRes{}

	taskInfo, err := batch_mail.GetTaskInfo(ctx, int(req.TaskId))
	if err != nil {
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to get task information: {}", err.Error())))
		return
	}

	if taskInfo == nil || taskInfo.Id == 0 {
		res.Code = 404
		res.SetError(gerror.New(public.LangCtx(ctx, "Task not found: {}", req.TaskId)))
		return
	}

	statService := batch_mail.NewTaskStatService()

	chartData := statService.GetTaskStatChart(req.TaskId, req.Domain, req.StartTime, req.EndTime)

	res.Data.Dashboard = chartData["dashboard"]
	res.Data.MailProviders = chartData["mail_providers"]
	res.Data.SendMailChart = chartData["send_mail_chart"]
	res.Data.BounceRateChart = chartData["bounce_rate_chart"]
	res.Data.OpenRateChart = chartData["open_rate_chart"]
	res.Data.ClickRateChart = chartData["click_rate_chart"]

	res.SetSuccess(public.LangCtx(ctx, "Task statistics retrieved successfully"))
	return
}
