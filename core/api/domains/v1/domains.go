package v1

import (
	"billionmail-core/utility/types/api_v1"
	"github.com/gogf/gf/v2/frame/g"
)

type AddDomainReq struct {
	g.Meta       `path:"/domains/create" tags:"Domain" method:"post" sm:"Add domain" in:"body"`
	Domain       string `json:"domain" v:"required" dc:"Domain"`
	Mailboxes    int    `json:"mailboxes" v:"min:1" dc:"Mailboxes" d:"50"`
	MailboxQuota int    `json:"mailboxQuota" v:"min:1" dc:"MailboxQuota" d:"5242880"`
	Quota        int    `json:"quota" v:"required" dc:"Quota" d:"10485760"`
	RateLimit    int    `json:"rateLimit" v:"min:1" dc:"RateLimit" d:"12"`
}

type AddDomainRes struct {
	api_v1.StandardRes
}

type UpdateDomainReq struct {
	g.Meta       `path:"/domains/update" tags:"Domain" method:"post" sm:"Update domain" in:"body"`
	Domain       string `json:"domain" v:"required" dc:"Domain"`
	Mailboxes    int    `json:"mailboxes" v:"min:1" dc:"Mailboxes" d:"50"`
	MailboxQuota int    `json:"mailboxQuota" v:"min:1" dc:"MailboxQuota" d:"5242880"`
	Quota        int    `json:"quota" v:"required" dc:"Quota" d:"10485760"`
	RateLimit    int    `json:"rateLimit" v:"min:1" dc:"RateLimit" d:"12"`
}

type UpdateDomainRes struct {
	api_v1.StandardRes
}

type DeleteDomainReq struct {
	g.Meta `path:"/domains/delete" tags:"Domain" method:"post" sm:"Delete domain" in:"body"`
	Domain string `json:"domain" v:"required" dc:"Domain"`
}

type DeleteDomainRes struct {
	api_v1.StandardRes
}

type GetDomainReq struct {
	g.Meta   `path:"/domains/list" tags:"Domain" method:"get" sm:"Get domain" in:"query"`
	Keyword  string `json:"keyword" dc:"Keyword"`
	Page     int    `json:"page" dc:"Page" d:"1"`
	PageSize int    `json:"page_size" dc:"Page size" d:"20"`
}

type GetDomainRes struct {
	api_v1.StandardRes
}

type SetSSLReq struct {
	g.Meta      `path:"/domains/set_ssl" tags:"Domain" method:"post" sm:"Set SSL" in:"body"`
	Domain      string `json:"domain" v:"required" dc:"Domain"`
	Certificate string `json:"certificate" v:"required" dc:"Certificate"`
	Key         string `json:"key" v:"required" dc:"Key"`
}

type SetSSLRes struct {
	api_v1.StandardRes
}
