package v1

import (
	"billionmail-core/internal/model/entity"
	"billionmail-core/utility/types/api_v1"
	"github.com/gogf/gf/v2/frame/g"
)

type EmailTask struct {
	Id                      int    `json:"id"              dc:"task id"`
	TaskName                string `json:"task_name"       dc:"task name"`
	Addresser               string `json:"addresser"       dc:"addresser"`
	Subject                 string `json:"subject"         dc:"subject"`
	FullName                string `json:"full_name"       dc:"full name"`
	RecipientCount          int    `json:"recipient_count" dc:"recipient count"`
	TaskProcess             int    `json:"task_process"    dc:"task process"`
	Pause                   int    `json:"pause"           dc:"pause"`
	TemplateId              int    `json:"template_id"     dc:"template id"`
	IsRecord                int    `json:"is_record"       dc:"is record"`
	Unsubscribe             int    `json:"unsubscribe"     dc:"unsubscribe"`
	Threads                 int    `json:"threads"         dc:"threads"`
	Etypes                  string `json:"etypes"          dc:"etypes"`
	TrackOpen               int    `json:"track_open"      dc:"track open"`
	TrackClick              int    `json:"track_click"     dc:"track click"`
	StartTime               int    `json:"start_time"      dc:"start time"`
	CreateTime              int    `json:"create_time"     dc:"create time"`
	UpdateTime              int    `json:"update_time"     dc:"update time"`
	Remark                  string `json:"remark"          dc:"remark"`
	Active                  int    `json:"active"          dc:"status"`
	AddType                 int    `json:"add_type"        dc:"add type"`
	EstimatedTimeWithWarmup int64  `json:"estimated_time_with_warmup" dc:"estimated time for warmup (if applicable)"`
}

type GroupInfo struct {
	Id          int    `json:"id"    dc:"group id"`
	Name        string `json:"name"  dc:"group name"`
	Description string `json:"description" dc:"group description"`
	Count       int    `json:"count" dc:"contact count"`
}

type TaskDetail struct {
	EmailTask
	Groups       []*GroupInfo `json:"groups"         dc:"groups"`
	SentCount    int          `json:"sent_count"     dc:"sent count"`
	UnsentCount  int          `json:"unsent_count"   dc:"unsent count"`
	Progress     int          `json:"progress"      dc:"progress"`
	TemplateName string       `json:"template_name"  dc:"template name"`
	SuccessCount int          `json:"success_count"  dc:"success count"`
	ErrorCount   int          `json:"error_count"    dc:"error count"`
	Deferred     int          `json:"deferred"       dc:"deferred count"`
	//Opened       int          `json:"opened"        dc:"opened count"`
	//Clicked      int          `json:"clicked"       dc:"clicked count"`
	//DeliveryRate float64      `json:"delivery_rate" dc:"delivery rate"`
	//BounceRate   float64      `json:"bounce_rate"  dc:"bounce rate"`
	//OpenRate     float64      `json:"open_rate"    dc:"open rate"`
	//ClickRate    float64      `json:"click_rate"   dc:"click rate"`
}

type ListTasksReq struct {
	g.Meta        `path:"/batch_mail/task/list" method:"get" tags:"BatchMail" summary:"get task list"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Page          int    `json:"page" v:"required|min:1" dc:"page"`
	PageSize      int    `json:"page_size" v:"required|min:1" dc:"page size"`
	Keyword       string `json:"keyword" dc:"search keyword"`
	Status        int    `json:"status" dc:"task status" default:"-1"`
}

type ListTasksRes struct {
	api_v1.StandardRes
	Data struct {
		Total int           `json:"total" dc:"total"`
		List  []*TaskDetail `json:"list"  dc:"task list"`
	} `json:"data"`
}

type TaskInfoReq struct {
	g.Meta        `path:"/batch_mail/task/find" method:"get" tags:"BatchMail" summary:"get task info by id"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Id            int    `json:"id" v:"required" dc:"task id"`
}

type TaskInfoRes struct {
	api_v1.StandardRes
	Data *entity.EmailTask `json:"data"`
}

type TaskSummaryStats struct {
	Sends        int     `json:"sends"`
	Delivered    int     `json:"delivered"`
	Bounced      int     `json:"bounced"`
	Opened       int     `json:"opened"`
	Clicked      int     `json:"clicked"`
	DeliveryRate float64 `json:"delivery_rate"`
	BounceRate   float64 `json:"bounce_rate"`
	OpenRate     float64 `json:"open_rate"`
	ClickRate    float64 `json:"click_rate"`
}

type TaskOverviewReq struct {
	g.Meta        `path:"/batch_mail/task/overview" method:"get" tags:"BatchMail" summary:"get task overview"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	StartTime     int    `json:"start_time" dc:"start time"`
	EndTime       int    `json:"end_time" dc:"end time"`
}

type TaskOverviewRes struct {
	api_v1.StandardRes
	Data TaskSummaryStats `json:"data"`
}

type DeleteTaskReq struct {
	g.Meta        `path:"/batch_mail/task/delete" method:"post" tags:"BatchMail" summary:"delete task"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Id            int    `json:"id" v:"required" dc:"task id"`
}

type DeleteTaskRes struct {
	api_v1.StandardRes
}

type CreateTaskReq struct {
	g.Meta        `path:"/batch_mail/task/create" method:"post" tags:"BatchMail" summary:"create task"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Addresser     string `json:"addresser" v:"required" dc:"addresser"`
	Subject       string `json:"subject" v:"required" dc:"subject"`
	FullName      string `json:"full_name" dc:"full name"`
	TemplateId    int    `json:"template_id" v:"required" dc:"template id"`
	GroupIds      []int  `json:"group_ids" v:"required" dc:"group ids"`
	IsRecord      int    `json:"is_record" v:"in:0,1" dc:"is record" default:"1"`
	Unsubscribe   int    `json:"unsubscribe" v:"in:0,1" dc:"unsubscribe" default:"1"`
	Threads       int    `json:"threads" v:"min:0" dc:"threads" default:"5"`
	TrackOpen     int    `json:"track_open" v:"in:0,1" dc:"track open" default:"1"`
	TrackClick    int    `json:"track_click" v:"in:0,1" dc:"track click" default:"1"`
	StartTime     int    `json:"start_time" v:"required" dc:"start time"`
	Warmup        int    `json:"warmup" v:"in:0,1" dc:"warmup" default:"0"`
	Remark        string `json:"remark" dc:"remark"`
}

type CreateTaskRes struct {
	api_v1.StandardRes
	Data struct {
		Id int `json:"id" dc:"task id"`
	} `json:"data"`
}

type TaskSendCountData struct {
	TaskId        int         `json:"task_id"         dc:"Task ID"`
	StartTime     int64       `json:"start_time"      dc:"Start time (unix timestamp)"`
	EndTime       int64       `json:"end_time"        dc:"End time (unix timestamp)"`
	TotalSent     int         `json:"total_sent"      dc:"Total sent count in time range"`
	MaxPerMinute  int         `json:"max_per_minute"  dc:"Maximum emails sent per minute"`
	AvgPerMinute  float64     `json:"avg_per_minute"  dc:"Average emails sent per minute"`
	MinuteDetails interface{} `json:"minute_details"  dc:"Details of emails sent per minute"`
}

// TaskLogItem Task log item
type TaskLogItem struct {
	PostfixMessageId string  `json:"postfix_message_id"`
	Status           string  `json:"status"`
	Recipient        string  `json:"recipient"`
	MailProvider     string  `json:"mail_provider"`
	Delay            float64 `json:"delay"`
	Delays           string  `json:"delays"`
	Dsn              string  `json:"dsn"`
	Relay            string  `json:"relay"`
	Description      string  `json:"description"`
	LogTime          int64   `json:"log_time"`
}

type GetTaskSendCountReq struct {
	g.Meta        `path:"/batch_mail/task/send_count" method:"get" tags:"BatchMail" summary:"get task send count"`
	Authorization string `json:"authorization"  dc:"Authorization" in:"header"`
	TaskId        int    `json:"task_id"        dc:"Task ID"        v:"min:1#Task ID is required" in:"query"`
	StartTime     int64  `json:"start_time"     dc:"Start time (unix timestamp)" in:"query"`
	EndTime       int64  `json:"end_time"       dc:"End time (unix timestamp)"   in:"query"`
}

type GetTaskSendCountRes struct {
	api_v1.StandardRes
	Data *TaskSendCountData `json:"data" dc:"Task send count data"`
}

// Request for task email service provider statistics
type TaskMailProviderStatReq struct {
	g.Meta        `path:"/batch_mail/tracking/mail_provider" method:"get" tags:"MailProviderStat" summary:"Task email service provider statistics"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	TaskId        int    `json:"task_id" v:"required" dc:"Task ID"`
	Status        int    `json:"status" dc:"Status filter, (1:success 0:failure)"`
}

type TaskMailProviderStatRes struct {
	api_v1.StandardRes
}

// GetTaskMailLogsReq Request for task email logs
type GetTaskMailLogsReq struct {
	g.Meta        `path:"/batch_mail/tracking/logs" method:"get" tags:"EmailTracking" summary:"Get task email sending logs"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	TaskId        int    `json:"task_id" v:"required" dc:"Task ID"`
	Status        int    `json:"status" dc:"Status filter, (1:success 0:failure)"`
	Domain        string `json:"domain" dc:"Domain filter"`
	Page          int    `json:"page" dc:"Page number"`
	PageSize      int    `json:"page_size" dc:"Items per page"`
}

type GetTaskMailLogsRes struct {
	api_v1.StandardRes
	Data struct {
		Total int            `json:"total" dc:"Total records"`
		List  []*TaskLogItem `json:"list" dc:"Email log list"`
	} `json:"data" dc:"Data"`
}

type SendTestEmailReq struct {
	g.Meta        `path:"/batch_mail/task/send_test" method:"post" tags:"BatchMail" summary:"Send test email"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Addresser     string `json:"addresser" v:"required" dc:"Addresser"`
	Subject       string `json:"subject" v:"required" dc:"Subject"`
	Recipient     string `json:"recipient" v:"required" dc:"Recipient"`
	TemplateId    int    `json:"template_id" v:"required" dc:"Template ID"`
}

type SendTestEmailRes struct {
	api_v1.StandardRes
}

type TaskStatChartReq struct {
	g.Meta        `path:"/batch_mail/task/stat_chart" method:"get" tags:"BatchMail" summary:"get task stat chart"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	TaskId        int64  `json:"task_id" v:"required" dc:"task id"`
	Domain        string `json:"domain" v:"domain" dc:"domain"`
	StartTime     int64  `json:"start_time" v:"required" dc:"start time"`
	EndTime       int64  `json:"end_time" v:"required" dc:"end time"`
}

type TaskStatChartRes struct {
	api_v1.StandardRes
	Data struct {
		Dashboard       interface{} `json:"dashboard" dc:"dashboard"`
		MailProviders   interface{} `json:"mail_providers" dc:"mail providers"`
		SendMailChart   interface{} `json:"send_mail_chart" dc:"send mail chart"`
		BounceRateChart interface{} `json:"bounce_rate_chart" dc:"bounce rate chart"`
		OpenRateChart   interface{} `json:"open_rate_chart" dc:"open rate chart"`
		ClickRateChart  interface{} `json:"click_rate_chart" dc:"click rate chart"`
	} `json:"data"`
}
