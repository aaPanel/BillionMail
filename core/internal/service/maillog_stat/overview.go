package maillog_stat

import (
	"billionmail-core/internal/consts"
	docker "billionmail-core/internal/service/dockerapi"
	"billionmail-core/internal/service/public"
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

// Overview maillog data overview structure
type Overview struct{}

// NewOverview new overview instance
func NewOverview() *Overview {
	return &Overview{}
}

// filterAndPrepareTimeSection filter and provide time section
func (o *Overview) filterAndPrepareTimeSection(startTime, endTime int64) (int64, int64) {
	if startTime > 0 && endTime < 0 {
		endTime = time.Now().Unix()
	}

	if startTime > endTime {
		panic(public.Lang("end_time must greater than start_time"))
	}

	// Maximum time range is 1 year
	if endTime-startTime > 31622400 {
		startTime = endTime - 31622400 // 1 year
	}

	return startTime, endTime
}

// buildBaseQuery build basic query
func (o *Overview) buildBaseQuery(campaignID int64, domain string, startTime, endTime int64) *gdb.Model {
	subQuery := "SELECT * FROM mailstat_send_mails WHERE true"

	if startTime > 0 {
		subQuery += fmt.Sprintf(" AND log_time_millis > %d", startTime*1000)
	}

	if endTime > 0 {
		subQuery += fmt.Sprintf(" AND log_time_millis < %d", endTime*1000)
	}

	query := g.DB().Model("(" + subQuery + ") sm")
	query.LeftJoin("mailstat_senders s", "sm.postfix_message_id=s.postfix_message_id")
	query.Where("s.postfix_message_id is not null")

	if campaignID > 0 {
		query.InnerJoin("mailstat_message_ids mi", "sm.postfix_message_id=mi.postfix_message_id")
		query.InnerJoin(fmt.Sprintf("(SELECT message_id FROM recipient_info WHERE task_id = %d) r", campaignID), "mi.message_id=r.message_id")
		// query.Where("r.task_id = ?", campaignID)
	}

	if domain != "" {
		query.Where("s.sender LIKE ?", "%@"+domain)
	}

	//if startTime > 0 {
	//	query.Where("sm.log_time_millis > ?", startTime*1000-1)
	//}
	//
	//if endTime > 0 {
	//	query.Where("sm.log_time_millis < ?", endTime*1000+1)
	//}

	return query
}

// Overview the maillog
func (o *Overview) Overview(campaignID int64, domain string, startTime, endTime int64) map[string]interface{} {
	startTime, endTime = o.filterAndPrepareTimeSection(startTime, endTime)

	return map[string]interface{}{
		"dashboard":         o.overviewDashboard(campaignID, domain, startTime, endTime),
		"mail_providers":    o.overviewProviders(campaignID, domain, startTime, endTime),
		"send_mail_chart":   o.chartSendMail(campaignID, domain, startTime, endTime),
		"bounce_rate_chart": o.chartBounceRate(campaignID, domain, startTime, endTime),
		"open_rate_chart":   o.chartOpenRate(campaignID, domain, startTime, endTime),
		"click_rate_chart":  o.chartClickRate(campaignID, domain, startTime, endTime),
	}
}

// overviewDashboard dashboard data
func (o *Overview) overviewDashboard(campaignID int64, domain string, startTime, endTime int64) map[string]interface{} {
	query := o.buildBaseQuery(campaignID, domain, startTime, endTime)

	query.LeftJoin(`LATERAL(
	SELECT id
	FROM mailstat_opened
	WHERE sm.postfix_message_id = postfix_message_id
	LIMIT 1
) as o`, "true")

	query.LeftJoin(`LATERAL(
	SELECT id
	FROM mailstat_clicked
	WHERE sm.postfix_message_id = postfix_message_id
	LIMIT 1
) as c`, "true")

	query.Fields("count(*) as sends")
	query.Fields("coalesce(sum(case when status='sent' and dsn like '2.%' then 1 else 0 end), 0) as delivered")
	query.Fields("count(o.id) as opened")
	query.Fields("count(c.id) as clicked")
	query.Fields("coalesce(sum(case when status='bounced' then 1 else 0 end), 0) as bounced")

	aggregate := map[string]interface{}{
		"sends":         0,
		"delivered":     0,
		"opened":        0,
		"clicked":       0,
		"bounced":       0,
		"delayed_queue": 0,
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
		aggregate["bounce_rate"] = public.Round(float64(aggregate["bounced"].(int))/float64(aggregate["sends"].(int))*100, 2)
	} else {
		aggregate["delivery_rate"] = 0
		aggregate["bounce_rate"] = 0
	}
	if aggregate["delivered"].(int) > 0 {
		aggregate["open_rate"] = public.Round(float64(aggregate["opened"].(int))/float64(aggregate["delivered"].(int))*100, 2)
		aggregate["click_rate"] = public.Round(float64(aggregate["clicked"].(int))/float64(aggregate["delivered"].(int))*100, 2)
	} else {
		aggregate["open_rate"] = 0
		aggregate["click_rate"] = 0
	}

	// delayed_queue
	aggregate["delayed_queue"] = 0
	delayedCount, err := o.getPostfixDeferredQueueCount(context.Background())
	if err == nil {
		aggregate["delayed_queue"] = delayedCount
	}

	return aggregate
}

// OverviewDashboard dashboard data
func (o *Overview) OverviewDashboard(campaignID int64, domain string, startTime, endTime int64) map[string]interface{} {
	aggregate := o.overviewDashboard(campaignID, domain, startTime, endTime)
	return aggregate
}

// Obtain the number of deferred queues with cache
func (o *Overview) getPostfixDeferredQueueCount(ctx context.Context) (int, error) {
	cacheKey := "postfix_deferred_queue_count"
	v := public.GetCache(cacheKey)
	if v != nil {
		if num, ok := v.(int); ok {
			return num, nil
		} else {
			return 0, nil
		}
	}

	dk, err := docker.NewDockerAPI()
	if err != nil {
		return 0, err
	}
	defer dk.Close()

	cmd := []string{"postqueue", "-j"}
	result, err := dk.ExecCommandByName(ctx, consts.SERVICES.Postfix, cmd, "root")
	if err != nil || result == nil || result.ExitCode != 0 {
		return 0, err
	}

	lines := strings.Split(result.Output, "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var item struct {
			QueueName string `json:"queue_name"`
		}
		if err := gjson.Unmarshal([]byte(line), &item); err != nil {
			continue
		}
		if item.QueueName == "deferred" {
			count++
		}
	}

	public.SetCache(cacheKey, count, 120)
	return count, nil
}

// overviewProviders statistic mail provider
func (o *Overview) overviewProviders(campaignID int64, domain string, startTime, endTime int64) []map[string]interface{} {
	startTime, endTime = o.filterAndPrepareTimeSection(startTime, endTime)

	query := o.buildBaseQuery(campaignID, domain, startTime, endTime)

	query.LeftJoin(`LATERAL(
	SELECT id
	FROM mailstat_opened
	WHERE sm.postfix_message_id = postfix_message_id
	LIMIT 1
) as o`, "true")

	query.LeftJoin(`LATERAL(
	SELECT id
	FROM mailstat_clicked
	WHERE sm.postfix_message_id = postfix_message_id
	LIMIT 1
) as c`, "true")

	query.Fields("sm.mail_provider")
	query.Fields("count(*) as sends")
	query.Fields("coalesce(sum(case when status='sent' and dsn like '2.%' then 1 else 0 end), 0) as delivered")
	query.Fields("count(o.id) as opened")
	query.Fields("count(c.id) as clicked")
	query.Fields("coalesce(sum(case when status='bounced' then 1 else 0 end), 0) as bounced")

	query.Group("sm.mail_provider")

	lst := make([]map[string]interface{}, 0)

	results, err := query.All()
	if err != nil {
		g.Log().Error(context.Background(), err)
		return lst
	}

	aggregate := map[string]interface{}{
		"sends":     0,
		"delivered": 0,
		"opened":    0,
		"clicked":   0,
		"bounced":   0,
	}

	m := make(map[string]map[string]interface{})

	for _, result := range results {
		provider := result["mail_provider"].String()
		if _, ok := m[provider]; !ok {
			m[provider] = make(map[string]interface{})
			for k, v := range aggregate {
				m[provider][k] = v
			}
		}

		for k, v := range result {
			if _, ok := aggregate[k]; !ok {
				continue
			}
			m[provider][k] = v.Int()
		}
	}

	for k, item := range m {
		item["mail_provider"] = k
		if item["sends"].(int) > 0 {
			item["delivery_rate"] = public.Round(float64(item["delivered"].(int))/float64(item["sends"].(int))*100, 2)
			item["bounce_rate"] = public.Round(float64(item["bounced"].(int))/float64(item["sends"].(int))*100, 2)
			item["open_rate"] = public.Round(float64(item["opened"].(int))/float64(item["sends"].(int))*100, 2)
			item["click_rate"] = public.Round(float64(item["clicked"].(int))/float64(item["sends"].(int))*100, 2)
		} else {
			item["delivery_rate"] = 0
			item["bounce_rate"] = 0
			item["open_rate"] = 0
			item["click_rate"] = 0
		}
		lst = append(lst, item)
	}

	// descending by sends
	sort.Slice(lst, func(i, j int) bool {
		return lst[i]["sends"].(int) > lst[j]["sends"].(int)
	})

	return lst
}

func (o *Overview) fillChartData(data []map[string]interface{}, fillItem map[string]interface{}, fillType, fillKey string, startTime, endTime int64) []map[string]interface{} {
	switch fillType {
	case "monthly":
		return o.fillChartDataMonthly(data, fillItem, fillKey, startTime, endTime)
	case "daily":
		return o.fillChartDataDaily(data, fillItem, fillKey, startTime, endTime)
	case "hourly":
		return o.fillChartDataHourly(data, fillItem, fillKey)
	default:
		return data
	}
}

func (o *Overview) fillChartDataHourly(data []map[string]interface{}, fillItem map[string]interface{}, fillKey string) []map[string]interface{} {
	for _, item := range data {
		item[fillKey] = gconv.Int(item[fillKey])
	}

	if len(data) == 24 {
		return data
	}

	m := make(map[int]map[string]interface{})
	for _, item := range data {
		m[gconv.Int(item[fillKey])] = item
	}

	lst := make([]map[string]interface{}, 0, 24)
	for i := 0; i < 24; i++ {
		if item, ok := m[i]; ok {
			lst = append(lst, item)
			continue
		}

		item := make(map[string]interface{}, len(fillItem))
		for k, v := range fillItem {
			item[k] = v
		}
		item[fillKey] = i
		lst = append(lst, item)
	}

	return lst
}

func (o *Overview) fillChartDataDaily(data []map[string]interface{}, fillItem map[string]interface{}, fillKey string, startTime, endTime int64) []map[string]interface{} {
	if startTime < 0 || endTime < 0 {
		return data
	}

	if startTime > endTime {
		return data
	}

	m := make(map[string]map[string]interface{}, len(data))
	for _, item := range data {
		m[gconv.String(item[fillKey])] = item
	}

	lst := make([]map[string]interface{}, 0, (endTime-startTime)/86400+1)
	for i := startTime; i <= endTime; i += 86400 {
		dayDateObj := time.Unix(i, 0)
		dayDate := dayDateObj.Format("2006-01-02")
		dayTime := time.Date(dayDateObj.Year(), dayDateObj.Month(), dayDateObj.Day(), 0, 0, 0, 0, time.Local).Unix()

		if item, ok := m[dayDate]; ok {
			item[fillKey] = dayTime
			lst = append(lst, item)
			continue
		}

		item := make(map[string]interface{}, len(fillItem))
		for k, v := range fillItem {
			item[k] = v
		}
		item[fillKey] = dayTime
		lst = append(lst, item)
	}

	return lst
}

func (o *Overview) fillChartDataMonthly(data []map[string]interface{}, fillItem map[string]interface{}, fillKey string, startTime, endTime int64) []map[string]interface{} {
	if startTime < 0 || endTime < 0 {
		return data
	}

	if startTime > endTime {
		return data
	}

	m := make(map[string]map[string]interface{}, len(data))
	for _, item := range data {
		m[gconv.String(item[fillKey])] = item
	}

	lst := make([]map[string]interface{}, 0, (endTime-startTime)/2592000+1)
	for i := startTime; i <= endTime; i += 2592000 {
		monthDateObj := time.Unix(i, 0)
		monthDate := monthDateObj.Format("2006-01")
		monthTime := time.Date(monthDateObj.Year(), monthDateObj.Month(), 1, 0, 0, 0, 0, time.Local).Unix()

		if item, ok := m[monthDate]; ok {
			item[fillKey] = monthTime
			lst = append(lst, item)
			continue
		}

		item := make(map[string]interface{}, len(fillItem))
		for k, v := range fillItem {
			item[k] = v
		}
		item[fillKey] = monthTime
		lst = append(lst, item)
	}

	return lst
}

// sendMailDashboard dashboard data for send mail
func (o *Overview) sendMailDashboard(campaignID int64, domain string, startTime, endTime int64) map[string]interface{} {
	startTime, endTime = o.filterAndPrepareTimeSection(startTime, endTime)

	query := o.buildBaseQuery(campaignID, domain, startTime, endTime)

	query.Fields("count(*) as sends")
	query.Fields("coalesce(sum(case when status='sent' and dsn like '2.%' then 1 else 0 end), 0) as delivered")
	query.Fields("count(*) - coalesce(sum(case when status='sent' and dsn like '2.%' then 1 else 0 end), 0) as failed")

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
		aggregate[k] = v.Int()
	}

	if aggregate["sends"].(int) > 0 {
		aggregate["delivery_rate"] = public.Round(float64(aggregate["delivered"].(int))/float64(aggregate["sends"].(int))*100, 2)
		aggregate["failure_rate"] = public.Round(float64(aggregate["failed"].(int))/float64(aggregate["sends"].(int))*100, 2)
	} else {
		aggregate["delivery_rate"] = 0
		aggregate["failure_rate"] = 0
	}

	return aggregate
}

func (o *Overview) prepareChartData(startTime, endTime int64) (string, string) {
	columnType := "daily"
	secs := endTime - startTime
	xAxisField := "TO_CHAR(TO_TIMESTAMP(sm.log_time_millis / 1000), 'YYYY-MM-DD') as x"

	if secs < 86400 {
		columnType = "hourly"
		xAxisField = "TO_CHAR(TO_TIMESTAMP(sm.log_time_millis / 1000), 'HH24')::integer as x"
	}

	if secs >= 2592000 {
		columnType = "monthly"
		xAxisField = "TO_CHAR(TO_TIMESTAMP(sm.log_time_millis / 1000), 'YYYY-MM') as x"
	}

	return columnType, xAxisField
}

func (o *Overview) chartSendMail(campaignID int64, domain string, startTime, endTime int64) map[string]interface{} {
	startTime, endTime = o.filterAndPrepareTimeSection(startTime, endTime)

	query := o.buildBaseQuery(campaignID, domain, startTime, endTime)

	columnType, xAxisField := o.prepareChartData(startTime, endTime)

	query.Fields(xAxisField)
	query.Fields("count(*) as sends")
	query.Fields("coalesce(sum(case when status='sent' and dsn like '2.%' then 1 else 0 end), 0) as delivered")
	query.Fields("count(*) - coalesce(sum(case when status='sent' and dsn like '2.%' then 1 else 0 end), 0) as failed")

	query.Group("x")
	query.Order("x")

	rs, err := query.All()
	if err != nil {
		g.Log().Error(context.Background(), err)
		return nil
	}

	results := make([]map[string]interface{}, 0, len(rs))
	for _, item := range rs {
		results = append(results, item.Map())
	}

	fillItem := map[string]interface{}{
		"sends":     0,
		"delivered": 0,
		"failed":    0,
	}

	return map[string]interface{}{
		"column_type": columnType,
		"dashboard":   o.sendMailDashboard(campaignID, domain, startTime, endTime),
		"data":        o.fillChartData(results, fillItem, columnType, "x", startTime, endTime),
	}
}

func (o *Overview) chartBounceRate(campaignID int64, domain string, startTime, endTime int64) map[string]interface{} {
	startTime, endTime = o.filterAndPrepareTimeSection(startTime, endTime)

	query := o.buildBaseQuery(campaignID, domain, startTime, endTime)

	columnType, xAxisField := o.prepareChartData(startTime, endTime)

	query.Fields(xAxisField)
	query.Fields(gdb.Raw("case when count(*) > 0 then round(1.0 * coalesce(sum(case when status='bounced' then 1 else 0 end), 0) / count(*) * 100.0, 2) else 0.0 end as bounce_rate"))

	query.Group("x")
	query.Order("x")

	rs, err := query.All()
	if err != nil {
		g.Log().Error(context.Background(), err)
		return nil
	}

	results := make([]map[string]interface{}, 0, len(rs))
	for _, item := range rs {
		results = append(results, item.Map())
	}

	fillItem := map[string]interface{}{
		"bounce_rate": 0.0,
	}

	return map[string]interface{}{
		"column_type": columnType,
		"data":        o.fillChartData(results, fillItem, columnType, "x", startTime, endTime),
	}
}

func (o *Overview) chartOpenRate(campaignID int64, domain string, startTime, endTime int64) map[string]interface{} {
	startTime, endTime = o.filterAndPrepareTimeSection(startTime, endTime)

	query := o.buildBaseQuery(campaignID, domain, startTime, endTime)

	columnType, xAxisField := o.prepareChartData(startTime, endTime)

	query.LeftJoin(`LATERAL(
	SELECT id
	FROM mailstat_opened
	WHERE sm.postfix_message_id = postfix_message_id
	LIMIT 1
) as o`, "true")

	query.Fields(xAxisField)
	query.Fields("case when coalesce(sum(case when status='sent' and dsn like '2.%' then 1 else 0 end), 0) > 0 then round(1.0 * count(o.id) / coalesce(sum(case when status='sent' and dsn like '2.%' then 1 else 0 end), 0) * 100, 2) else 0.0 end as open_rate")

	query.Group("x")

	rs, err := query.All()
	if err != nil {
		g.Log().Error(context.Background(), err)
		return nil
	}

	results := make([]map[string]interface{}, 0, len(rs))
	for _, item := range rs {
		results = append(results, item.Map())
	}

	fillItem := map[string]interface{}{
		"open_rate": 0.0,
	}

	return map[string]interface{}{
		"column_type": columnType,
		"data":        o.fillChartData(results, fillItem, columnType, "x", startTime, endTime),
	}
}

func (o *Overview) chartClickRate(campaignID int64, domain string, startTime, endTime int64) map[string]interface{} {
	startTime, endTime = o.filterAndPrepareTimeSection(startTime, endTime)

	query := o.buildBaseQuery(campaignID, domain, startTime, endTime)

	columnType, xAxisField := o.prepareChartData(startTime, endTime)

	query.LeftJoin(`LATERAL(
	SELECT id
	FROM mailstat_opened
	WHERE sm.postfix_message_id = postfix_message_id
	LIMIT 1
) as o`, "true")

	query.LeftJoin(`LATERAL(
	SELECT id
	FROM mailstat_clicked
	WHERE sm.postfix_message_id = postfix_message_id
	LIMIT 1
) as c`, "true")

	query.Fields(xAxisField)
	query.Fields("case when count(o.id) > 0 then round(1.0 * count(c.id) / count(o.id) * 100, 2) else 0.0 end as click_rate")

	query.Group("x")

	rs, err := query.All()
	if err != nil {
		g.Log().Error(context.Background(), err)
		return nil
	}

	results := make([]map[string]interface{}, 0, len(rs))
	for _, item := range rs {
		results = append(results, item.Map())
	}

	fillItem := map[string]interface{}{
		"click_rate": 0.0,
	}

	return map[string]interface{}{
		"column_type": columnType,
		"data":        o.fillChartData(results, fillItem, columnType, "x", startTime, endTime),
	}
}

// FailedList failed list
func (o *Overview) FailedList(campaignID int64, domain string, startTime, endTime int64) []map[string]interface{} {
	startTime, endTime = o.filterAndPrepareTimeSection(startTime, endTime)

	query := o.buildBaseQuery(campaignID, domain, startTime, endTime)

	query.LeftJoin(`LATERAL(
	SELECT id, dsn, delay, delays, relay, description
	FROM mailstat_deferred_mails
	WHERE sm.postfix_message_id = postfix_message_id
	ORDER BY id DESC
	LIMIT 1
) as d`, "true")

	query.Fields("sm.postfix_message_id")
	query.Fields("s.sender")
	query.Fields("sm.recipient")
	query.Fields("sm.log_time")
	query.Fields("sm.status")
	query.Fields("coalesce(d.dsn, sm.dsn) as dsn")
	query.Fields("coalesce(d.delay, sm.delay) as delay")
	query.Fields("coalesce(d.delays, sm.delays) as delays")
	query.Fields("coalesce(d.relay, sm.relay) as relay")
	query.Fields("coalesce(d.description, sm.description) as description")

	query.Where("sm.status != ?", "sent")

	query.OrderDesc("sm.log_time_millis")

	results, err := query.All()
	if err != nil {
		g.Log().Error(context.Background(), err)
		return nil
	}

	lst := make([]map[string]interface{}, 0, len(results))
	for _, item := range results {
		lst = append(lst, item.Map())
	}

	return lst
}

// description FailedListBounced
func (o *Overview) FailedListBounced(campaignID int64, domain string, startTime, endTime int64) []map[string]interface{} {

	startTime, endTime = o.filterAndPrepareTimeSection(startTime, endTime)

	query := o.buildBaseQuery(campaignID, domain, startTime, endTime)

	query.Fields("sm.recipient")
	query.Fields("sm.status")
	query.Fields("sm.description")
	query.Fields("sm.log_time")

	query.Where("sm.status = ?", "bounced")

	query.OrderDesc("sm.log_time_millis")

	results, err := query.All()
	if err != nil {
		g.Log().Error(context.Background(), err)
		return nil
	}

	lst := make([]map[string]interface{}, 0, len(results))
	for _, item := range results {
		lst = append(lst, item.Map())
	}

	return lst
}
