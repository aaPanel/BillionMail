package v1

import (
	"billionmail-core/utility/types/api_v1"
	"github.com/gogf/gf/v2/frame/g"
)

type AddMailboxReq struct {
	g.Meta   `path:"/mailbox/create" tags:"MailBox" method:"post" summary:"Create mailbox" in:"body"`
	Domain   string `json:"domain" v:"required|domain" dc:"Domain"`
	Username string `json:"username" v:"required|min-length:1" dc:"Username"`
	Password string `json:"password" v:"required|min-length:1" dc:"Password"`
	Status   int    `json:"status" v:"required" dc:"Status" d:"1"`
	IsAdmin  int    `json:"isAdmin" v:"required" dc:"IsAdmin" d:"0"`
	Quota    int    `json:"quota" v:"required" dc:"Quota" d:"5242880"`
}

type AddMailboxRes struct {
	api_v1.StandardRes
}

type UpdateMailboxReq struct {
	g.Meta   `path:"/mailbox/update" tags:"MailBox" method:"post" summary:"Update mailbox" in:"body"`
	Email    string `json:"email" v:"required|email" dc:"Email"`
	Domain   string `json:"domain" v:"required" dc:"Domain"`
	Username string `json:"username" v:"required" dc:"Username"`
	Password string `json:"password" v:"required" dc:"Password"`
	Status   int    `json:"status" v:"required" dc:"Status" d:"1"`
	IsAdmin  int    `json:"isAdmin" v:"required" dc:"IsAdmin" d:"0"`
	Quota    int    `json:"quota" v:"required" dc:"Quota" d:"5242880"`
}

type UpdateMailboxRes struct {
	api_v1.StandardRes
}

type DeleteMailboxReq struct {
	g.Meta `path:"/mailbox/delete" tags:"MailBox" method:"post" summary:"Delete mailbox" in:"body"`
	Email  string `json:"email" v:"required|email" dc:"Email"`
}

type DeleteMailboxRes struct {
	api_v1.StandardRes
}

type GetMailboxReq struct {
	g.Meta   `path:"/mailbox/list" tags:"MailBox" method:"get" summary:"Get mailbox" in:"query"`
	Keyword  string `json:"keyword" dc:"Keyword"`
	Domain   string `json:"domain" v:"required|domain" dc:"Domain"`
	Page     int    `json:"page" dc:"Page" d:"1"`
	PageSize int    `json:"page_size" dc:"Page size" d:"20"`
}

type GetMailboxRes struct {
	api_v1.StandardRes
}
