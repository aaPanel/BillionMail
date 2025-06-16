package v1

import (
	"billionmail-core/utility/types/api_v1"
	"github.com/gogf/gf/v2/frame/g"
)

// `CREATE TABLE IF NOT EXISTS api_templates (
// 	id SERIAL PRIMARY KEY,
// 	api_key VARCHAR(64) NOT NULL,
// 	api_name VARCHAR(255) NOT NULL,
// 	template_id INTEGER NOT NULL,
// 	subject TEXT NOT NULL,
// 	addresser VARCHAR(320) NOT NULL,
// 	full_name VARCHAR(255),
// 	unsubscribe SMALLINT NOT NULL DEFAULT 0,
// 	track_open SMALLINT NOT NULL DEFAULT 1,
// 	track_click SMALLINT NOT NULL DEFAULT 1,
// 	active SMALLINT NOT NULL DEFAULT 1,
// 	create_time INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
// 	update_time INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
// 	UNIQUE(api_key)
// )`,

// `CREATE TABLE IF NOT EXISTS api_mail_logs (
// 	id SERIAL PRIMARY KEY,
// 	api_id INTEGER NOT NULL,
// 	recipient VARCHAR(320) NOT NULL,
// 	message_id TEXT NOT NULL,
// 	addresser VARCHAR(320) NOT NULL,
// 	send_time INTEGER NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
// )`,

type ApiTemplates struct {
	Id                int    `json:"id" dc:"id"`
	ApiKey            string `json:"api_key" dc:"api key"`
	ApiName           string `json:"api_name" dc:"api name"`
	TemplateId        int    `json:"template_id" dc:"template id"`
	Subject           string `json:"subject" dc:"subject"`
	Addresser         string `json:"addresser" dc:"addresser"`
	FullName          string `json:"full_name" dc:"full name"`
	Unsubscribe       int    `json:"unsubscribe" dc:"unsubscribe"`
	TrackOpen         int    `json:"track_open" dc:"track open"`
	TrackClick        int    `json:"track_click" dc:"track click"`
	Active            int    `json:"active" dc:"active"`
	CreateTime        int    `json:"create_time" dc:"create time"`
	UpdateTime        int    `json:"update_time" dc:"update time"`
	ExpireTime        int    `json:"expire_time" dc:"expire time"`
	LastKeyUpdateTime int    `json:"last_key_update_time" dc:"last key update time"`
}

type ApiMailLogs struct {
	Id        int    `json:"id" dc:"id"`
	ApiId     int    `json:"api_id" dc:"api id"`
	Recipient string `json:"recipient" dc:"recipient"`
	MessageId string `json:"message_id" dc:"message id"`
	Addresser string `json:"addresser" dc:"addresser"`
	SendTime  int    `json:"send_time" dc:"send time"`
}

type ApiTemplatesListReq struct {
	g.Meta        `path:"/batch_mail/api/list" method:"get" tags:"ApiMail" summary:"api list"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Page          int    `json:"page" dc:"page"`
	PageSize      int    `json:"page_size" dc:"page size"`
	Keyword       string `json:"keyword" dc:"Search Keyword"`
	Active        int    `json:"active" dc:"active"`
	StartTime     int    `json:"start_time" dc:"start time"`
	EndTime       int    `json:"end_time" dc:"end time"`
}

type ApiTemplatesInfo struct {
	ApiTemplates
	SendCount        int     `json:"send_count" dc:"send count"`
	SuccessCount     int     `json:"success_count" dc:"success count"`
	FailCount        int     `json:"fail_count" dc:"fail count"`
	OpenRate         float64 `json:"open_rate" dc:"open rate"`
	ClickRate        float64 `json:"click_rate" dc:"click rate"`
	DeliveryRate     float64 `json:"delivery_rate" dc:"delivery rate"`
	BounceRate       float64 `json:"bounce_rate" dc:"bounce rate"`
	UnsubscribeCount int     `json:"unsubscribe_count" dc:"unsubscribe count"`
}

type ApiTemplatesListRes struct {
	api_v1.StandardRes
	Data struct {
		Total int                 `json:"total" dc:"total"`
		List  []*ApiTemplatesInfo `json:"list"  dc:"api templates list"`
	} `json:"data"`
}

type ApiSummaryStats struct {
	TotalSend        int     `json:"total_send" dc:"total send count"`
	AvgDeliveryRate  float64 `json:"avg_delivery_rate" dc:"average delivery rate"`
	AvgOpenRate      float64 `json:"avg_open_rate" dc:"average open rate"`
	AvgClickRate     float64 `json:"avg_click_rate" dc:"average click rate"`
	AvgBounceRate    float64 `json:"avg_bounce_rate" dc:"average bounce rate"`
	AvgUnsubRate     float64 `json:"avg_unsub_rate" dc:"average unsubscribe rate"`
	TotalUnsubscribe int     `json:"total_unsubscribe" dc:"total unsubscribe count"`
}

type ApiOverviewStatsReq struct {
	g.Meta        `path:"/batch_mail/api/overview_stats" method:"get" tags:"ApiMail" summary:"api Overview"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	StartTime     int    `json:"start_time" dc:"start time"`
	EndTime       int    `json:"end_time" dc:"end time"`
}

type ApiOverviewStatsRes struct {
	api_v1.StandardRes
	Data ApiSummaryStats `json:"data"`
}

type ApiTemplatesCreateReq struct {
	g.Meta        `path:"/batch_mail/api/create" method:"post" tags:"ApiMail" summary:"api create"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	ApiName       string `json:"api_name" dc:"api name"`
	TemplateId    int    `json:"template_id" dc:"template id"`
	Subject       string `json:"subject" dc:"subject"`
	Addresser     string `json:"addresser" dc:"addresser"`
	FullName      string `json:"full_name" dc:"full name"`
	Unsubscribe   int    `json:"unsubscribe" dc:"unsubscribe"`
	Active        int    `json:"active" dc:"active"`
	ExpireTime    int    `json:"expire_time" dc:"expire time"` // 0 is a permanently valid unit of seconds
}

type ApiTemplatesCreateRes struct {
	api_v1.StandardRes
}

type ApiTemplatesUpdateReq struct {
	g.Meta        `path:"/batch_mail/api/update" method:"post" tags:"ApiMail" summary:"api update"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	ID            int    `json:"id" dc:"id"`
	ApiName       string `json:"api_name" dc:"api name"`
	TemplateId    int    `json:"template_id" dc:"template id"`
	Subject       string `json:"subject" dc:"subject"`
	Addresser     string `json:"addresser" dc:"addresser"`
	FullName      string `json:"full_name" dc:"full name"`
	Unsubscribe   int    `json:"unsubscribe" dc:"unsubscribe"`
	TrackOpen     int    `json:"track_open" dc:"track open"`
	TrackClick    int    `json:"track_click" dc:"track click"`
	Active        int    `json:"active" dc:"active"`
	ExpireTime    int    `json:"expire_time" dc:"Key expiration time (0 is permanent)"`
	ResetKey      bool   `json:"reset_key" dc:"reset key"`
}

type ApiTemplatesUpdateRes struct {
	api_v1.StandardRes
}

type ApiTemplatesDeleteReq struct {
	g.Meta        `path:"/batch_mail/api/delete" method:"post" tags:"ApiMail" summary:"api delete"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	ID            int    `json:"id" dc:"id"`
}

type ApiTemplatesDeleteRes struct {
	api_v1.StandardRes
}

type ApiMailSendReq struct {
	g.Meta        `path:"/batch_mail/api/send" method:"post" tags:"ApiMail" summary:"call api send mail"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	ApiKey        string `json:"x-api-key" dc:"API Key" in:"header"`
	Addresser     string `json:"addresser" dc:"addresser"`
	Recipient     string `json:"recipient" dc:"recipient"`
}

type ApiMailSendRes struct {
	api_v1.StandardRes
}
