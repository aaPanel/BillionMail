package v1

import (
	"billionmail-core/utility/types/api_v1"
	"github.com/gogf/gf/v2/frame/g"
)

// EmailTemplate defines the email template entity
type EmailTemplate struct {
	Id         int    `json:"id"          description:"Template ID"     orm:"id"`
	TempName   string `json:"temp_name"   description:"Template Name"   orm:"temp_name"`
	AddType    int    `json:"add_type"    description:"Type"           orm:"add_type"`
	Content    string `json:"html_content"     description:"Email Content"   orm:"content"`
	Render     string `json:"drag_data"      description:"Render Data"     orm:"render"`
	CreateTime int    `json:"create_time" description:"Create Time"     orm:"create_time"`
	UpdateTime int    `json:"update_time" description:"Update Time"     orm:"update_time"`
}

// CreateTemplateReq Create template request
type CreateTemplateReq struct {
	g.Meta        `path:"/email_template/create" method:"post" tags:"EmailTemplate" summary:"Create email template"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	TempName      string `json:"temp_name" v:"required" dc:"Template Name"`
	AddType       int    `json:"add_type" v:"required|in:0,1" dc:"Add Type(0:HTML 1:Drag )"`
	Content       string `json:"html_content" dc:"HTML Content(Required when add_type=0,1)"`
	Render        string `json:"drag_data" dc:"Drag and Drop Render Data(Required when add_type=1)"`
}

type CreateTemplateRes struct {
	api_v1.StandardRes
	Data struct {
		Id int `json:"id" dc:"Template ID"`
	} `json:"data"`
}

// DeleteTemplateReq Delete template request
type DeleteTemplateReq struct {
	g.Meta        `path:"/email_template/delete" method:"post" tags:"EmailTemplate" summary:"Delete email template"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Id            int    `json:"id" v:"required" dc:"Template ID"`
}

type DeleteTemplateRes struct {
	api_v1.StandardRes
}

// UpdateTemplateReq Update template request
type UpdateTemplateReq struct {
	g.Meta        `path:"/email_template/update" method:"post" tags:"EmailTemplate" summary:"Update email template"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Id            int    `json:"id" v:"required" dc:"Template ID"`
	TempName      string `json:"temp_name" dc:"Template Name"`
	Content       string `json:"html_content" dc:"HTML Content"`
	Render        string `json:"drag_data" dc:"Render Data"`
}

type UpdateTemplateRes struct {
	api_v1.StandardRes
}

// ListTemplatesReq List templates request
type ListTemplatesReq struct {
	g.Meta        `path:"/email_template/list" method:"get" tags:"EmailTemplate" summary:"Get email template list"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Page          int    `json:"page" v:"required|min:1" dc:"Page Number"`
	PageSize      int    `json:"page_size" v:"required|min:1" dc:"Page Size"`
	Keyword       string `json:"keyword" dc:"Search Keyword(Template Name)"`
	AddType       int    `json:"add_type" dc:"Add Type Filter(0:Upload 1:Drag and Drop -1:All)" default:"-1"`
}

type ListTemplatesRes struct {
	api_v1.StandardRes
	Data struct {
		Total int              `json:"total" dc:"Total Count"`
		List  []*EmailTemplate `json:"list" dc:"Template List"`
	} `json:"data"`
}

// CopyTemplateReq Copy template request
type CopyTemplateReq struct {
	g.Meta        `path:"/email_template/copy" method:"post" tags:"EmailTemplate" summary:"Copy email template"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Id            int    `json:"id" v:"required" dc:"Template ID"`
}

type CopyTemplateRes struct {
	api_v1.StandardRes
}

// GetTemplateReq Req Get template request
type GetTemplateReq struct {
	g.Meta        `path:"/email_template/get" method:"get" tags:"EmailTemplate" summary:"Get email template"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Id            int    `json:"id" v:"required" dc:"Template ID"`
}
type GetTemplateRes struct {
	api_v1.StandardRes
	Data *EmailTemplate `json:"data" dc:"Template Data"`
}

type GetAllTemplatesReq struct {
	g.Meta        `path:"/email_template/all" method:"get" tags:"EmailTemplate" summary:"Get all email templates"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
}
type GetAllTemplatesRes struct {
	api_v1.StandardRes
}
