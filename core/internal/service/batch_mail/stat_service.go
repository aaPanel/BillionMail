package batch_mail

import (
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"strconv"
	"time"
)

type TaskStatService struct{}

func NewTaskStatService() *TaskStatService {
	return &TaskStatService{}
}

// buildBaseQuery
func (s *TaskStatService) buildBaseQuery(taskId int64, domain string, startTime, endTime int64) *gdb.Model {

	query := g.DB().Model("mailstat_send_mails sm")

	query.LeftJoin("mailstat_senders s", "sm.postfix_message_id=s.postfix_message_id")
	query.Where("s.postfix_message_id is not null")

	query.LeftJoin("mailstat_message_ids mi", "sm.postfix_message_id=mi.postfix_message_id")
	query.LeftJoin("recipient_info ri", "mi.message_id=ri.message_id")

	query.Where("ri.task_id = ?", taskId)

	if domain != "" {
		query.Where("sm.recipient LIKE ?", "%@"+domain)
	}

	if startTime > 0 {
		query.Where("sm.log_time_millis > ?", startTime*1000-1)
	}

	if endTime > 0 {
		query.Where("sm.log_time_millis < ?", endTime*1000+1)
	}

	return query
}

// GetTaskStatChart
func (s *TaskStatService) GetTaskStatChart(taskId int64, domain string, startTime, endTime int64) map[string]interface{} {
	startTime, endTime = s.filterAndPrepareTimeSection(startTime, endTime)

	return map[string]interface{}{
		"dashboard":         s.getTaskDashboard(taskId, domain, startTime, endTime),
		"mail_providers":    s.getTaskMailProviders(taskId, domain, startTime, endTime),
		"send_mail_chart":   s.getTaskSendMailChart(taskId, domain, startTime, endTime),
		"bounce_rate_chart": s.getTaskBounceRateChart(taskId, domain, startTime, endTime),
		"open_rate_chart":   s.getTaskOpenRateChart(taskId, domain, startTime, endTime),
		"click_rate_chart":  s.getTaskClickRateChart(taskId, domain, startTime, endTime),
	}
}

// filterAndPrepareTimeSection
func (s *TaskStatService) filterAndPrepareTimeSection(startTime, endTime int64) (int64, int64) {
	if startTime > 0 && endTime < 0 {
		endTime = time.Now().Unix()
	}

	if startTime > endTime {
		panic(public.Lang("end_time must greater than start_time"))
	}

	return startTime, endTime
}

// prepareChartData
func (s *TaskStatService) prepareChartData(startTime, endTime int64) (string, string) {
	columnType := "daily"
	secs := endTime - startTime
	// xAxisField := "to_char(to_timestamp(sm.log_time_millis / 1000), 'YYYY-MM-DD') as x"
	xAxisField := "EXTRACT(EPOCH FROM to_timestamp(sm.log_time_millis / 1000))::bigint as x"

	if secs < 86400 {
		columnType = "hourly"
		// xAxisField = "to_char(to_timestamp(sm.log_time_millis / 1000), 'HH24')::integer as x"
		xAxisField = "EXTRACT(EPOCH FROM date_trunc('hour', to_timestamp(sm.log_time_millis / 1000)))::bigint as x"
	}

	return columnType, xAxisField
}

// getTaskDashboard
func (s *TaskStatService) getTaskDashboard(taskId int64, domain string, startTime, endTime int64) map[string]interface{} {
	query := s.buildBaseQuery(taskId, domain, startTime, endTime)

	query.Fields(
		"count(*) as sends",
		"coalesce(sum(case when status='sent' and dsn like '2.%' then 1 else 0 end), 0) as delivered",
		"coalesce(sum(case when status='bounced' then 1 else 0 end), 0) as bounced",
	)

	result, err := query.One()
	if err != nil {
		g.Log().Error(context.Background(), err)
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

// fillChartData
func (s *TaskStatService) fillChartData(data []map[string]interface{}, fillItem map[string]interface{}, fillType, fillKey string, startTime, endTime int64) []map[string]interface{} {
	switch fillType {
	case "daily":
		return s.fillChartDataDaily(data, fillItem, fillKey, startTime, endTime)
	case "hourly":
		return s.fillChartDataHourly(data, fillItem, fillKey)
	default:
		return data
	}
}

// fillChartDataHourly
func (s *TaskStatService) fillChartDataHourly(data []map[string]interface{}, fillItem map[string]interface{}, fillKey string) []map[string]interface{} {
	result := make([]map[string]interface{}, 24)
	dataMap := make(map[int]map[string]interface{})

	for _, item := range data {
		hour := item[fillKey].(int)
		dataMap[hour] = item
	}

	for i := 0; i < 24; i++ {
		if item, exists := dataMap[i]; exists {
			result[i] = item
		} else {
			item := make(map[string]interface{}, len(fillItem))
			for k, v := range fillItem {
				item[k] = v
			}
			item[fillKey] = i
			result[i] = item
		}
	}

	return result
}

// fillChartDataDaily
func (s *TaskStatService) fillChartDataDaily(data []map[string]interface{}, fillItem map[string]interface{}, fillKey string, startTime, endTime int64) []map[string]interface{} {
	if startTime < 0 || endTime < 0 {
		return data
	}

	if startTime > endTime {
		return data
	}

	// 使用开始日期的0点
	// startTime = time.Unix(startTime, 0).Truncate(24 * time.Hour).Unix()

	m := make(map[int64]map[string]interface{}, len(data))
	for _, item := range data {
		// 确保类型转换正确
		var timestamp int64
		switch x := item[fillKey].(type) {
		case int64:
			timestamp = x
		case float64:
			timestamp = int64(x)
		case string:
			t, err := strconv.ParseInt(x, 10, 64)
			if err != nil {
				continue
			}
			timestamp = t
		default:
			continue
		}
		m[timestamp] = item
	}

	lst := make([]map[string]interface{}, 0, (endTime-startTime)/86400+1)
	for t := startTime; t <= endTime; t += 86400 {
		if item, ok := m[t]; ok {
			lst = append(lst, item)
			continue
		}

		item := make(map[string]interface{}, len(fillItem))
		for k, v := range fillItem {
			item[k] = v
		}
		item[fillKey] = t
		lst = append(lst, item)
	}

	return lst

}

// getTaskBounceRateChart
func (s *TaskStatService) getTaskBounceRateChart(taskId int64, domain string, startTime, endTime int64) map[string]interface{} {
	query := s.buildBaseQuery(taskId, domain, startTime, endTime)
	columnType, xAxisField := s.prepareChartData(startTime, endTime)

	query.Fields(
		xAxisField,
		"case when count(*) > 0 then round(1.0 * coalesce(sum(case when status='bounced' then 1 else 0 end), 0) / count(*) * 100.0, 2) else 0.0 end as bounce_rate",
	)

	query.Group("x")
	query.Order("x")

	results, err := query.All()
	if err != nil {
		return nil
	}

	fillItem := map[string]interface{}{
		"bounce_rate": 0.0,
	}

	return map[string]interface{}{
		"column_type": columnType,
		"data":        s.fillChartData(results.List(), fillItem, columnType, "x", startTime, endTime),
	}
}

// getTaskOpenRateChart
func (s *TaskStatService) getTaskOpenRateChart(taskId int64, domain string, startTime, endTime int64) map[string]interface{} {
	query := s.buildBaseQuery(taskId, domain, startTime, endTime)
	query.LeftJoin("mailstat_opened o", "sm.postfix_message_id=o.postfix_message_id")

	columnType, xAxisField := s.prepareChartData(startTime, endTime)

	query.Fields(
		xAxisField,
		"case when count(*) > 0 then round(1.0 * count(distinct o.postfix_message_id) / count(*) * 100.0, 2) else 0.0 end as open_rate",
	)

	query.Group("x")
	query.Order("x")

	results, err := query.All()
	if err != nil {
		return nil
	}
	fillItem := map[string]interface{}{
		"open_rate": 0.0,
	}

	return map[string]interface{}{
		"column_type": columnType,
		"data":        s.fillChartData(results.List(), fillItem, columnType, "x", startTime, endTime),
	}
}

// getTaskClickRateChart
func (s *TaskStatService) getTaskClickRateChart(taskId int64, domain string, startTime, endTime int64) map[string]interface{} {
	query := s.buildBaseQuery(taskId, domain, startTime, endTime)
	query.LeftJoin("mailstat_clicked c", "sm.postfix_message_id=c.postfix_message_id")

	columnType, xAxisField := s.prepareChartData(startTime, endTime)

	query.Fields(
		xAxisField,
		"case when count(*) > 0 then round(1.0 * count(distinct c.postfix_message_id) / count(*) * 100.0, 2) else 0.0 end as click_rate",
	)

	query.Group("x")
	query.Order("x")

	results, err := query.All()
	if err != nil {
		return nil
	}
	fillItem := map[string]interface{}{
		"click_rate": 0.0,
	}
	return map[string]interface{}{
		"column_type": columnType,
		"data":        s.fillChartData(results.List(), fillItem, columnType, "x", startTime, endTime),
	}
}

// getTaskMailProviders
func (s *TaskStatService) getTaskMailProviders(taskId int64, domain string, startTime, endTime int64) []map[string]interface{} {
	query := s.buildBaseQuery(taskId, domain, startTime, endTime)
	query.LeftJoin("mailstat_opened o", "sm.postfix_message_id=o.postfix_message_id")
	query.LeftJoin("mailstat_clicked c", "sm.postfix_message_id=c.postfix_message_id")

	query.Fields(
		"sm.mail_provider",
		"count(*) as sends",
		"coalesce(sum(case when status='sent' and dsn like '2.%' then 1 else 0 end), 0) as delivered",
		"count(distinct o.postfix_message_id) as opened",
		"count(distinct c.postfix_message_id) as clicked",
		"coalesce(sum(case when status='bounced' then 1 else 0 end), 0) as bounced",
	)

	query.Group("sm.mail_provider")

	results, err := query.All()
	if err != nil {
		return nil
	}

	providers := make([]map[string]interface{}, 0)
	for _, result := range results {
		provider := result.Map()
		sends := result["sends"].Int()
		if sends > 0 {
			provider["delivery_rate"] = public.Round(float64(result["delivered"].Int())/float64(sends)*100, 2)
			provider["bounce_rate"] = public.Round(float64(result["bounced"].Int())/float64(sends)*100, 2)
			provider["open_rate"] = public.Round(float64(result["opened"].Int())/float64(sends)*100, 2)
			provider["click_rate"] = public.Round(float64(result["clicked"].Int())/float64(sends)*100, 2)
		}
		providers = append(providers, provider)
	}

	return providers
}

// getTaskSendMailChart
func (s *TaskStatService) getTaskSendMailChart(taskId int64, domain string, startTime, endTime int64) map[string]interface{} {
	query := s.buildBaseQuery(taskId, domain, startTime, endTime)
	columnType, xAxisField := s.prepareChartData(startTime, endTime)

	query.Fields(
		xAxisField,
		"count(*) as sends",
		"coalesce(sum(case when status='sent' and dsn like '2.%' then 1 else 0 end), 0) as delivered",
		"count(*) - coalesce(sum(case when status='sent' and dsn like '2.%' then 1 else 0 end), 0) as failed",
	)

	query.Group("x")
	query.Order("x")

	results, err := query.All()
	if err != nil {
		return nil
	}

	fillItem := map[string]interface{}{
		"sends":     0,
		"delivered": 0,
		"failed":    0,
	}

	return map[string]interface{}{
		"column_type": columnType,
		"dashboard":   s.getTaskSendMailDashboard(taskId, domain, startTime, endTime),
		"data":        s.fillChartData(results.List(), fillItem, columnType, "x", startTime, endTime),
	}
}

// getTaskSendMailDashboard 获取发送邮件仪表盘数据
func (s *TaskStatService) getTaskSendMailDashboard(taskId int64, domain string, startTime, endTime int64) map[string]interface{} {
	query := s.buildBaseQuery(taskId, domain, startTime, endTime)

	query.Fields(
		"count(*) as sends",
		"coalesce(sum(case when status='sent' and dsn like '2.%' then 1 else 0 end), 0) as delivered",
		"count(*) - coalesce(sum(case when status='sent' and dsn like '2.%' then 1 else 0 end), 0) as failed",
	)

	aggregate := map[string]interface{}{
		"sends":     0,
		"delivered": 0,
		"failed":    0,
	}

	result, err := query.One()
	if err != nil {
		g.Log().Error(context.Background(), err)
		return aggregate
	}

	for k, v := range result {
		if _, ok := aggregate[k]; !ok {
			continue
		}
		aggregate[k] = v.Int()
	}

	if aggregate["sends"].(int) > 0 {
		aggregate["delivery_rate"] = public.Round(float64(aggregate["delivered"].(int))/float64(aggregate["sends"].(int))*100, 2)
		aggregate["failure_rate"] = public.Round(float64(aggregate["failed"].(int))/float64(aggregate["sends"].(int))*100, 2)
	} else {
		aggregate["delivery_rate"] = 0.0
		aggregate["failure_rate"] = 0.0
	}

	return aggregate
}
