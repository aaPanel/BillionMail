package batch_mail

import (
	"billionmail-core/api/batch_mail/v1"
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"time"
)

func (c *ControllerV1) ApiTemplatesList(ctx context.Context, req *v1.ApiTemplatesListReq) (res *v1.ApiTemplatesListRes, err error) {
	res = &v1.ApiTemplatesListRes{}

	// build query conditions
	model := g.DB().Model("api_templates").Safe()

	// add api_name fuzzy search
	if req.Keyword != "" {
		model = model.WhereLike("api_name", "%"+req.Keyword+"%")
	}
	// active filter
	if req.Active != -1 {
		model = model.Where("active", req.Active)
	}

	if req.StartTime > 0 && req.EndTime < 0 {
		req.EndTime = int(time.Now().Unix())
	}

	// get total
	total, err := model.Count()
	if err != nil {
		return nil, err
	}

	// query by page
	var list []*v1.ApiTemplatesInfo
	err = model.Page(req.Page, req.PageSize).
		OrderDesc("id").
		Scan(&list)
	if err != nil {
		return nil, err
	}

	for _, item := range list {
		// build base query
		query := g.DB().Model("api_mail_logs aml")

		query = query.LeftJoin("mailstat_message_ids mi", "aml.message_id=mi.message_id")
		query = query.LeftJoin("mailstat_send_mails sm", "mi.postfix_message_id=sm.postfix_message_id")
		query = query.Where("aml.api_id", item.Id)

		if req.StartTime > 0 {
			query.Where("sm.log_time_millis > ?", req.StartTime*1000-1)
		}

		if req.EndTime > 0 {
			query.Where("sm.log_time_millis < ?", req.EndTime*1000+1)
		}

		// count各项数据
		query.Fields(
			"count(*) as sends",
			"coalesce(sum(case when sm.status='sent' and sm.dsn like '2.%' then 1 else 0 end), 0) as delivered",
			"coalesce(sum(case when sm.status='bounced' then 1 else 0 end), 0) as bounced",
		)

		result, err := query.One()
		if err != nil {
			// g.Log().Error(ctx, "Stats API failed to send data:", err)
			continue
		}

		item.SendCount = result["sends"].Int()
		item.SuccessCount = result["delivered"].Int()
		item.FailCount = result["bounced"].Int()

		// count opened, clicked (use campaign_id directly)
		apiCampaignId := item.Id + 1000000000
		openedCount, _ := g.DB().Model("mailstat_opened").
			Where("campaign_id", apiCampaignId).
			WhereGTE("log_time_millis", req.StartTime*1000).
			WhereLTE("log_time_millis", req.EndTime*1000).
			Fields("count(distinct postfix_message_id) as opened").
			Value()
		clickedCount, _ := g.DB().Model("mailstat_clicked").
			Where("campaign_id", apiCampaignId).
			WhereGTE("log_time_millis", req.StartTime*1000).
			WhereLTE("log_time_millis", req.EndTime*1000).
			Fields("count(distinct postfix_message_id) as clicked").
			Value()

		if item.SendCount > 0 {
			item.DeliveryRate = public.Round(float64(item.SuccessCount)/float64(item.SendCount)*100, 2)
			item.BounceRate = public.Round(float64(item.FailCount)/float64(item.SendCount)*100, 2)
			item.OpenRate = public.Round(float64(openedCount.Int())/float64(item.SendCount)*100, 2)
			item.ClickRate = public.Round(float64(clickedCount.Int())/float64(item.SendCount)*100, 2)
		} else {
			item.DeliveryRate = 0
			item.BounceRate = 0
			item.OpenRate = 0
			item.ClickRate = 0
		}

		// count unsubscribe
		recipients := []string{}
		_, err = g.DB().Model("api_mail_logs").Where("api_id", item.Id).Fields("recipient").Array(&recipients)
		unsubscribeCount := 0
		if len(recipients) > 0 {
			unsubscribeCount, _ = g.DB().Model("bm_contacts").
				Where("email", recipients).
				Where("active", 0).
				WhereGTE("create_time", item.CreateTime).
				Count()
		}
		item.UnsubscribeCount = unsubscribeCount
	}

	res.Data.Total = total
	res.Data.List = list
	res.SetSuccess(public.LangCtx(ctx, "Get API list successfully"))
	return res, nil
}
