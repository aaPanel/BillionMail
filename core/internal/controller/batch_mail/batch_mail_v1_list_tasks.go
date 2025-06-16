package batch_mail

import (
	"billionmail-core/api/batch_mail/v1"
	"billionmail-core/internal/service/batch_mail"
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

func (c *ControllerV1) ListTasks(ctx context.Context, req *v1.ListTasksReq) (res *v1.ListTasksRes, err error) {

	res = &v1.ListTasksRes{}

	total, tasks, err := batch_mail.GetTasksWithPage(ctx, req.Page, req.PageSize, req.Keyword, req.Status)
	if err != nil {
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to get task list {}", err.Error())))
		return
	}

	task_list := make([]*v1.TaskDetail, len(tasks))
	for i, task := range tasks {
		detail := &v1.TaskDetail{
			EmailTask: *task,
		}

		stats := GetTaskStats(ctx, int64(task.Id))
		if stats != nil {

			detail.SentCount = stats["sends"].(int)
			detail.SuccessCount = stats["delivered"].(int)
			detail.Opened = stats["opened"].(int)
			detail.Clicked = stats["clicked"].(int)
			detail.ErrorCount = stats["bounced"].(int)
			detail.DeliveryRate = stats["delivery_rate"].(float64)
			detail.BounceRate = stats["bounce_rate"].(float64)
			detail.OpenRate = stats["open_rate"].(float64)
			detail.ClickRate = stats["click_rate"].(float64)
		}

		sentCount := detail.SentCount

		//sentCount, _ := batch_mail.GetSentCount(ctx, task.Id)
		//detail.SentCount = sentCount

		// Fix: Check if recipient_count is valid
		if task.RecipientCount <= 0 {

			actualCount, err := batch_mail.GetActualRecipientCount(ctx, task.Id)
			if err == nil && actualCount > 0 {
				// Update the recipient count in the database
				_ = batch_mail.UpdateRecipientCount(ctx, task.Id, actualCount)
				detail.RecipientCount = actualCount
				task.RecipientCount = actualCount
			}
		}

		// Now calculate unsent_count with the corrected recipient_count
		if task.RecipientCount >= sentCount {
			detail.UnsentCount = task.RecipientCount - sentCount
		} else {
			// If still inconsistent, set unsent count to 0
			detail.UnsentCount = 0
			g.Log().Warning(ctx, "Task %d has inconsistent data: recipient_count(%d) < sent_count(%d)",
				task.Id, task.RecipientCount, sentCount)
		}

		// Calculate progress with safeguards against division by zero
		if task.RecipientCount > 0 {
			detail.Progress = int(float64(sentCount) / float64(task.RecipientCount) * 100)
			// Ensure progress is between 0 and 100
			if detail.Progress > 100 {
				detail.Progress = 100
			}
		} else {
			// Default progress for tasks with no recipients
			detail.Progress = 0
		}

		// Get success and error counts
		//successCount, errorCount, err := batch_mail.GetTaskSendingStats(ctx, task.Id)
		//if err == nil {
		//	detail.SuccessCount = successCount
		//	detail.ErrorCount = errorCount
		//} else {
		//	detail.SuccessCount = 0
		//	detail.ErrorCount = 0
		//	g.Log().Error(ctx, "Failed to get sending statistics: ", err)
		//}

		task_list[i] = detail
	}

	res.Data.Total = total
	res.Data.List = task_list
	res.SetSuccess(public.LangCtx(ctx, "Task list retrieved successfully"))
	return
}

func GetTaskStats(ctx context.Context, taskId int64) map[string]interface{} {
	query := g.DB().Model("mailstat_send_mails sm")

	query = query.LeftJoin("mailstat_message_ids mi", "sm.postfix_message_id=mi.postfix_message_id")
	query = query.LeftJoin("recipient_info ri", "mi.message_id=ri.message_id")

	query = query.Where("ri.task_id = ?", taskId)

	query.Fields(
		"count(*) as sends",
		"coalesce(sum(case when sm.status='sent' and sm.dsn like '2.%' then 1 else 0 end), 0) as delivered", // success count
		"coalesce(sum(case when sm.status='bounced' then 1 else 0 end), 0) as bounced",                      // bounced count
	)

	result, err := query.One()
	if err != nil {
		g.Log().Error(ctx, err)
		return nil
	}

	sends := result["sends"].Int()
	delivered := result["delivered"].Int()
	bounced := result["bounced"].Int()

	// 通过campaign_id查打开和点击
	campaignId := int(taskId)
	openedCount, _ := g.DB().Model("mailstat_opened").
		Where("campaign_id", campaignId).
		Fields("count(distinct postfix_message_id) as opened").
		Value()
	clickedCount, _ := g.DB().Model("mailstat_clicked").
		Where("campaign_id", campaignId).
		Fields("count(distinct postfix_message_id) as clicked").
		Value()

	stats := map[string]interface{}{
		"sends":     sends,
		"delivered": delivered,
		"opened":    openedCount.Int(),
		"clicked":   clickedCount.Int(),
		"bounced":   bounced,
	}

	if sends > 0 {
		stats["delivery_rate"] = public.Round(float64(delivered)/float64(sends)*100, 2)
		stats["bounce_rate"] = public.Round(float64(bounced)/float64(sends)*100, 2)
		stats["open_rate"] = public.Round(float64(stats["opened"].(int))/float64(sends)*100, 2)
		stats["click_rate"] = public.Round(float64(stats["clicked"].(int))/float64(sends)*100, 2)
	} else {
		stats["delivery_rate"] = 0.0
		stats["bounce_rate"] = 0.0
		stats["open_rate"] = 0.0
		stats["click_rate"] = 0.0
	}

	return stats
}
