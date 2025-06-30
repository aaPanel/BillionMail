package batch_mail

import (
	"billionmail-core/internal/service/public"
	"context"
	"database/sql"
	"github.com/gogf/gf/v2/frame/g"

	"billionmail-core/api/batch_mail/v1"
)

func (c *ControllerV1) ApiOverviewStats(ctx context.Context, req *v1.ApiOverviewStatsReq) (res *v1.ApiOverviewStatsRes, err error) {
	res = &v1.ApiOverviewStatsRes{}

	// query all API templates
	apiList := []struct {
		Id         int
		CreateTime int
	}{}
	err = g.DB().Model("api_templates").Fields("id, create_time").Scan(&apiList)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	var totalSend, totalDelivered, totalBounced, totalOpened, totalClicked int

	for _, api := range apiList {

		query := g.DB().Model("api_mail_logs aml")
		query = query.LeftJoin("mailstat_message_ids mi", "aml.message_id=mi.message_id")
		query = query.LeftJoin("mailstat_send_mails sm", "mi.postfix_message_id=sm.postfix_message_id")
		query = query.Where("aml.api_id", api.Id)

		query = query.Where("aml.status", 2)

		if req.StartTime > 0 {
			query.Where("sm.log_time_millis > ?", int64(req.StartTime)*1000-1)
		}
		if req.EndTime > 0 {
			query.Where("sm.log_time_millis < ?", int64(req.EndTime)*1000+1)
		}
		query.Fields(
			"count(*) as sends",
			"coalesce(sum(case when sm.status='sent' and sm.dsn like '2.%' then 1 else 0 end), 0) as delivered",
			"coalesce(sum(case when sm.status='bounced' then 1 else 0 end), 0) as bounced",
		)
		result, _ := query.One()
		sends := result["sends"].Int()
		delivered := result["delivered"].Int()
		bounced := result["bounced"].Int()
		totalSend += sends
		totalDelivered += delivered
		totalBounced += bounced

		// 统计打开、点击
		apiCampaignId := api.Id + 1000000000
		openedCount, _ := g.DB().Model("mailstat_opened").
			Where("campaign_id", apiCampaignId).
			WhereGTE("log_time_millis", int64(req.StartTime)*1000).
			WhereLTE("log_time_millis", int64(req.EndTime)*1000).
			Fields("count(distinct postfix_message_id) as opened").
			Value()
		clickedCount, _ := g.DB().Model("mailstat_clicked").
			Where("campaign_id", apiCampaignId).
			WhereGTE("log_time_millis", int64(req.StartTime)*1000).
			WhereLTE("log_time_millis", int64(req.EndTime)*1000).
			Fields("count(distinct postfix_message_id) as clicked").
			Value()
		totalOpened += openedCount.Int()
		totalClicked += clickedCount.Int()

		//// count unsubscribe
		//recipients := []string{}
		//_, _ = g.DB().Model("api_mail_logs").Where("api_id", api.Id).Fields("recipient,api_id").Array(&recipients)
		//unsubscribeCount := 0
		//if len(recipients) > 0 {
		//	unsubscribeCount, _ = g.DB().Model("bm_contacts").
		//		Where("email", recipients).
		//		Where("active", 0).
		//		WhereGTE("create_time", api.CreateTime).
		//		Count()
		//}
		//totalUnsub += unsubscribeCount
	}

	var avgDeliveryRate, avgOpenRate, avgClickRate, avgBounceRate float64
	if totalSend > 0 {
		avgDeliveryRate = public.Round(float64(totalDelivered)/float64(totalSend)*100, 2)
		avgOpenRate = public.Round(float64(totalOpened)/float64(totalSend)*100, 2)
		avgClickRate = public.Round(float64(totalClicked)/float64(totalSend)*100, 2)
		avgBounceRate = public.Round(float64(totalBounced)/float64(totalSend)*100, 2)
		//avgUnsubRate = public.Round(float64(totalUnsub)/float64(totalSend)*100, 2)
	}
	res.Data = v1.ApiSummaryStats{
		TotalSend:       totalSend,
		AvgDeliveryRate: avgDeliveryRate,
		AvgOpenRate:     avgOpenRate,
		AvgClickRate:    avgClickRate,
		AvgBounceRate:   avgBounceRate,
		//AvgUnsubRate:    avgUnsubRate,
		//TotalUnsubscribe: totalUnsub,
	}
	res.SetSuccess(public.LangCtx(ctx, "Statistics successful"))
	return res, nil
}
