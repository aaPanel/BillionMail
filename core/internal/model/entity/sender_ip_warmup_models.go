package entity

// SenderIpWarmup bm_sender_ip_warmup
type SenderIpWarmup struct {
	Id               int64  `json:"id"`
	SenderIp         string `json:"sender_ip"`
	Period           int    `json:"period"`
	Score            int    `json:"score"`
	LastEvaluateTime int64  `json:"last_evaluate_time"`
	Progress         int    `json:"progress"`
	ReWarmups        int    `json:"re_warmups"`
	BeginTime        int64  `json:"begin_time"`
	EndTime          int64  `json:"end_time"`
	CreateTime       int64  `json:"create_time"`
	UpdateTime       int64  `json:"update_time"`
}

// SenderIpMailProvider bm_sender_ip_mail_provider
type SenderIpMailProvider struct {
	Id               int64  `json:"id"`
	SenderIp         string `json:"sender_ip"`
	MailProvider     string `json:"mail_provider"`
	Score            int    `json:"score"`
	LastEvaluateTime int64  `json:"last_evaluate_time"`
	CreateTime       int64  `json:"create_time"`
	UpdateTime       int64  `json:"update_time"`
}

// CampaignWarmup bm_campaign_warmup
type CampaignWarmup struct {
	Id       int64 `json:"id"`
	TaskId   int64 `json:"task_id"`
	WarmupId int64 `json:"warmup_id"`
}
