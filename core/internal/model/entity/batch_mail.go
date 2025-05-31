package entity

// ContactGroup Entity
type ContactGroup struct {
	Id          int    `json:"id"          dc:"Group ID"`
	Name        string `json:"name"        dc:"Group Name"`
	Description string `json:"description" dc:"Description"`
	CreateTime  int    `json:"create_time" dc:"Create Time"`
	UpdateTime  int    `json:"update_time" dc:"Update Time"`
}

// Contact Entity
type Contact struct {
	Id         int               `json:"id"          dc:"Contact ID"`
	Email      string            `json:"email"       dc:"Email Address"`
	GroupId    int               `json:"group_id"    dc:"Group ID"`
	Active     int               `json:"active"      dc:"Status(1:Subscribed 0:Unsubscribed)"`
	TaskId     int               `json:"task_id"     dc:"Bulk Mail Task ID"`
	CreateTime int               `json:"create_time" dc:"Create Time"`
	Status     int               `json:"Status" dc:"1:Confirmed   0:Unconfirmed"`
	Attribs    map[string]string `json:"attribs"`
}

// EmailTemplate Entity
type EmailTemplate struct {
	Id         int    `json:"id"          description:"Template ID"     orm:"id"`
	TempName   string `json:"temp_name"   description:"Template Name"   orm:"temp_name"`
	AddType    int    `json:"add_type"    description:"Type"           orm:"add_type"`
	Content    string `json:"content"     description:"Email Content"   orm:"content"`
	Render     string `json:"render"      description:"Render Data"     orm:"render"`
	CreateTime int    `json:"create_time" description:"Create Time"     orm:"create_time"`
	UpdateTime int    `json:"update_time" description:"Update Time"     orm:"update_time"`
}

// EmailTask Entity
type EmailTask struct {
	Id             int    `json:"id"              dc:"Task ID"`
	TaskName       string `json:"task_name"       dc:"Task Name"`
	Addresser      string `json:"addresser"       dc:"Sender"`
	Subject        string `json:"subject"         dc:"Subject"`
	FullName       string `json:"full_name"       dc:"Sender Name"`
	RecipientCount int    `json:"recipient_count" dc:"Recipient Count"`
	TaskProcess    int    `json:"task_process"    dc:"Task Status"`
	Pause          int    `json:"pause"           dc:"Pause Status"`
	TemplateId     int    `json:"template_id"     dc:"Template ID"`
	IsRecord       int    `json:"is_record"       dc:"Record to Outbox"`
	Unsubscribe    int    `json:"unsubscribe"     dc:"Allow Unsubscribe"`
	Threads        int    `json:"threads"         dc:"Thread Count"`
	Etypes         string `json:"etypes"          dc:"Contact Group IDs"`
	TrackOpen      int    `json:"track_open"      dc:"Track Opens"`
	TrackClick     int    `json:"track_click"     dc:"Track Clicks"`
	StartTime      int    `json:"start_time"      dc:"Start Time"`
	CreateTime     int    `json:"create_time"     dc:"Create Time"`
	UpdateTime     int    `json:"update_time"     dc:"Update Time"`
	Remark         string `json:"remark"          dc:"Remark"`
	Active         int    `json:"active"          dc:"Status"`
}

// RecipientInfo Entity
type RecipientInfo struct {
	Id         int    `json:"id"          dc:"Recipient ID"`
	TaskId     int    `json:"task_id"     dc:"Task ID"`
	Recipient  string `json:"recipient"   dc:"Recipient Email"`
	IsSent     int    `json:"is_sent"     dc:"Send Status"`
	SentTime   int    `json:"sent_time"   dc:"Send Time"`
	MessageId  string `json:"message_id"  dc:"Email Message-ID"`
	CreateTime int    `json:"create_time" dc:"Create Time"`
}

type AbnormalRecipient struct {
	Id          int    `json:"id"          dc:"Group ID"`
	Recipient   string `json:"recipient"   dc:"Recipient Email"`
	CreateTime  int    `json:"create_time" dc:"Create Time"`
	Description string `json:"description" dc:"Description"`
	Count       int    `json:"count"       dc:"Count"`
	AddType     int    `json:"add_type"    dc:"Add Type"`
	// AddType 1: Manually added, 2: Automatically scanned, 3: Manually scanned
}

type MailTemplateContext struct {
	Subscriber     *Contact
	Task           *EmailTask
	UnsubscribeURL string
}
