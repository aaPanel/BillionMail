package entity

import (
	"encoding/json"
)

// ContactGroup Entity
type ContactGroup struct {
	Id          int    `json:"id"          dc:"Group ID"`
	Name        string `json:"name"        dc:"Group Name"`
	Description string `json:"description" dc:"Description"`
	CreateTime  int    `json:"create_time" dc:"Create Time"`
	UpdateTime  int    `json:"update_time" dc:"Update Time"`
	Token       string `json:"token"      dc:"Subscription Token"`
	DoubleOptin int    `json:"double_optin" dc:"Double Opt-in Status(0: Single Opt-in 1: Double Opt-in)"`
	WelcomeHtml string `json:"welcome_mail_html" dc:"Welcome Email Html"`
	WelcomeDrag string `json:"welcome_mail_drag" dc:"Welcome Email Drag"`

	WelcomeSubject   string `json:"welcome_subject" dc:"Welcome Email Subject"`
	SendWelcomeEmail int    `json:"send_welcome_email" dc:"Whether to send a welcome email"`
	ConfirmSubject   string `json:"confirm_subject" dc:"Confirmation Email Subject"`

	ConfirmHtml            string `json:"confirm_mail_html" dc:"Confirmation Email Html"`
	ConfirmDrag            string `json:"confirm_mail_drag" dc:"Confirmation Email Drag"`
	SuccessUrl             string `json:"success_url" dc:"Success URL"`
	ConfirmUrl             string `json:"confirm_url" dc:"Confirmation URL"`
	AlreadyUrl             string `json:"already_url" dc:"Already Subscribed URL"`
	SubscribeForm          string `json:"subscribe_form" dc:"Subscription Form HTML"`
	UnsubscribeMailHtml    string `json:"unsubscribe_mail_html" dc:"Unsubscribe Email Html"`
	UnsubscribeMailDrag    string `json:"unsubscribe_mail_drag" dc:"Unsubscribe EmailDrag"`
	UnsubscribeSubject     string `json:"unsubscribe_subject" dc:"Unsubscribe Email Subject"`
	UnsubscribeRedirectUrl string `json:"unsubscribe_redirect_url" dc:"Unsubscribe Success RedirectURL"`
	SendUnsubscribeEmail   int    `json:"send_unsubscribe_email" dc:"Whether to send unsubscribe email"`
}

// Contact Entity
type Contact struct {
	Id           int               `json:"id"          dc:"Contact ID"`
	Email        string            `json:"email"       dc:"Email Address"`
	GroupId      int               `json:"group_id"    dc:"Group ID"`
	Active       int               `json:"active"      dc:"Status(1:Subscribed 0:Unsubscribed)"`
	TaskId       int               `json:"task_id"     dc:"Bulk Mail Task ID"`
	CreateTime   int               `json:"create_time" dc:"Create Time"`
	Status       int               `json:"Status" dc:"1:Confirmed   0:Unconfirmed"`
	Attribs      map[string]string `json:"attribs"`
	LastActiveAt int               `json:"last_active_at" dc:"Last Active At"`
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
	//Etypes          string `json:"etypes"          dc:"Contact Group IDs"`
	TrackOpen       int    `json:"track_open"      dc:"Track Opens"`
	TrackClick      int    `json:"track_click"     dc:"Track Clicks"`
	StartTime       int    `json:"start_time"      dc:"Start Time"`
	CreateTime      int    `json:"create_time"     dc:"Create Time"`
	UpdateTime      int    `json:"update_time"     dc:"Update Time"`
	Remark          string `json:"remark"          dc:"Remark"`
	Active          int    `json:"active"          dc:"Status"`
	AddType         int    `json:"addType"          description:""`
	SendsCount      int    `json:"sendsCount"       description:""`
	DeliveredCount  int    `json:"deliveredCount"   description:""`
	BouncedCount    int    `json:"bouncedCount"     description:""`
	DeferredCount   int    `json:"deferredCount"    description:""`
	StatsUpdateTime int    `json:"statsUpdateTime"  description:""`
	GroupId         int    `json:"group_id"        dc:"Group ID"`
	TagIdsRaw       string `json:"-"               dc:"Tag IDs (JSON string - internal use)" orm:"tag_ids"`
	TagIds          []int  `json:"tag_ids"         dc:"Tag IDs (parsed array)"`
	TagLogic        string `json:"tag_logic"       dc:"Tag Logic (AND/OR/NOT)"`
	UseTagFilter    int    `json:"use_tag_filter"  dc:"Use Tag Filter (0: no, 1: yes)"`
	RotateSenders   int    `json:"rotate_senders"  dc:"Rotate through all domain mailboxes (0: no, 1: yes)"`
}

// MarshalJSON implements custom JSON marshaling to convert TagIdsRaw to TagIds array
func (e *EmailTask) MarshalJSON() ([]byte, error) {
	// Create a temporary struct with all fields
	type Alias EmailTask
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	// Parse TagIdsRaw to TagIds if not already parsed
	if e.TagIdsRaw != "" && len(e.TagIds) == 0 {
		var tagIds []int
		err := json.Unmarshal([]byte(e.TagIdsRaw), &tagIds)
		if err == nil {
			aux.TagIds = tagIds
		}
	}

	return json.Marshal(aux)
}

// AfterFind is called after retrieving data from database
func (e *EmailTask) AfterFind() {
	if e.TagIdsRaw != "" {
		var tagIds []int
		err := json.Unmarshal([]byte(e.TagIdsRaw), &tagIds)
		if err == nil {
			e.TagIds = tagIds
		}
	}
}

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

type ApiTemplates struct {
	Id                int    `json:"id" dc:"id"`
	ApiKey            string `json:"api_key" dc:"api key"`
	ApiName           string `json:"api_name" dc:"api name"`
	TemplateId        int    `json:"template_id" dc:"template id"`
	GroupId           int    `json:"group_id" dc:"group id"`
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

type Tag struct {
	Id         int    `json:"id"          dc:"Tag ID"`
	GroupId    int    `json:"group_id"    dc:"Group ID"`
	GroupName  string `json:"group_name" dc:"Group Name"`
	Name       string `json:"name"       dc:"Tag Name"`
	CreateTime int    `json:"create_time" dc:"Create Time"`
}
