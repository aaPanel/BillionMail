package batch_mail

import (
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"time"

	"billionmail-core/api/batch_mail/v1"
)

func (c *ControllerV1) TaskOverview(ctx context.Context, req *v1.TaskOverviewReq) (res *v1.TaskOverviewRes, err error) {
	res = &v1.TaskOverviewRes{}

	startTime := req.StartTime
	endTime := req.EndTime
	if startTime > 0 && endTime < 0 {
		endTime = int(time.Now().Unix())
	}

	query := g.DB().Model("mailstat_send_mails")
	if startTime > 0 {
		query = query.Where("log_time_millis > ?", startTime*1000-1)
	}
	if endTime > 0 {
		query = query.Where("log_time_millis < ?", endTime*1000+1)
	}
	query = query.Fields(
		"count(*) as sends",
		"coalesce(sum(case when status='sent' and dsn like '2.%' then 1 else 0 end), 0) as delivered",
		"coalesce(sum(case when status='bounced' then 1 else 0 end), 0) as bounced",
	)
	result, err := query.One()
	if err != nil {
		return nil, err
	}
	sends := result["sends"].Int()
	delivered := result["delivered"].Int()
	bounced := result["bounced"].Int()

	openedCount, _ := g.DB().Model("mailstat_opened").
		WhereGTE("log_time_millis", startTime*1000).
		WhereLTE("log_time_millis", endTime*1000).
		Fields("count(distinct postfix_message_id) as opened").
		Value()
	clickedCount, _ := g.DB().Model("mailstat_clicked").
		WhereGTE("log_time_millis", startTime*1000).
		WhereLTE("log_time_millis", endTime*1000).
		Fields("count(distinct postfix_message_id) as clicked").
		Value()

	var deliveryRate, bounceRate, openRate, clickRate float64
	if sends > 0 {
		deliveryRate = public.Round(float64(delivered)/float64(sends)*100, 2)
		bounceRate = public.Round(float64(bounced)/float64(sends)*100, 2)
		openRate = public.Round(float64(openedCount.Int())/float64(sends)*100, 2)
		clickRate = public.Round(float64(clickedCount.Int())/float64(sends)*100, 2)
	}

	res.Data.Sends = sends
	res.Data.Delivered = delivered
	res.Data.Bounced = bounced
	res.Data.Opened = openedCount.Int()
	res.Data.Clicked = clickedCount.Int()
	res.Data.DeliveryRate = deliveryRate
	res.Data.BounceRate = bounceRate
	res.Data.OpenRate = openRate
	res.Data.ClickRate = clickRate

	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return res, nil
}
