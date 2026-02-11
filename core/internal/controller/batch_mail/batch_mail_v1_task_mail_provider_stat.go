package batch_mail

import (
	"billionmail-core/internal/service/batch_mail"
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/batch_mail/v1"
)

func (c *ControllerV1) TaskMailProviderStat(ctx context.Context, req *v1.TaskMailProviderStatReq) (res *v1.TaskMailProviderStatRes, err error) {
	g.Log().Printf(ctx, "get task mail provider stat, TaskId: %d, Status: %d",
		req.TaskId, req.Status)

	res = &v1.TaskMailProviderStatRes{}

	// validate task info
	taskInfo, err := batch_mail.GetTaskInfo(ctx, req.TaskId)
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

	domainStats, err := calculateDomainStats(ctx, req.TaskId, req.Status)
	if err != nil {
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to calculate domain stats: {}", err.Error())))
		return
	}

	res.Data = domainStats
	res.SetSuccess(public.LangCtx(ctx, "Success"))

	return
}

func calculateDomainStats(ctx context.Context, taskId int, status int) ([]map[string]interface{}, error) {
	// build basic query
	query := g.DB().Model("mailstat_send_mails sm")
	query = query.LeftJoin("mailstat_message_ids mi", "sm.postfix_message_id=mi.postfix_message_id")
	query = query.LeftJoin("recipient_info ri", "mi.message_id=ri.message_id")
	query = query.Where("ri.task_id", taskId)

	// add filter conditions based on status
	if status == 1 {
		query = query.Where("sm.status = 'sent' AND sm.dsn LIKE '2.%'")
	} else if status == 0 {
		query = query.Where("sm.status = 'bounced'")
	}

	// calculate the number of emails for each recipient domain
	var results []struct {
		Domain string `json:"domain"`
		Count  int    `json:"count"`
	}

	// extract domain from recipient field and count
	err := query.Fields("SUBSTRING(ri.recipient FROM POSITION('@' IN ri.recipient) + 1) as domain, COUNT(*) as count").Group("domain").Order("count DESC").Scan(&results)

	if err != nil {
		return nil, gerror.Newf("Failed to calculate domain stats: %v", err)
	}

	domainStats := make([]map[string]interface{}, 0, len(results))
	for _, result := range results {
		if result.Domain != "" {
			domainStats = append(domainStats, map[string]interface{}{
				"domain": result.Domain,
				"count":  result.Count,
			})
		}
	}

	return domainStats, nil
}
