package v1

import (
	"billionmail-core/utility/types/api_v1"
	"github.com/gogf/gf/v2/frame/g"
)

type Alias struct {
	Address    string `json:"address"     dc:"Alias address"`
	Goto       string `json:"goto"        dc:"Forwarding target"`
	Domain     string `json:"domain"      dc:"Domain name"`
	CreateTime int64  `json:"create_time" dc:"Creation time"`
	UpdateTime int64  `json:"update_time" dc:"Update time"`
	Active     int    `json:"active"      dc:"Status: 1-enabled, 0-disabled"`
}

type GetMailForwardListReq struct {
	g.Meta    `path:"/mail_forward/list" method:"get" summary:"get mail forward list"`
	Domain    string `json:"domain"  desc:"domain filter"`
	Page      int    `json:"page" v:"min:1" desc:"page num" default:"1"`
	PageSize  int    `json:"page_size" v:"min:1" desc:"page size" default:"20"`
	SearchKey string `json:"search_key" desc:"search key, can search address"`
}

type GetMailForwardListRes struct {
	api_v1.StandardRes
	Data struct {
		Total int      `json:"total" desc:"total records"`
		List  []*Alias `json:"list" desc:"mail forward list"`
	} `json:"data"`
}

type AddMailForwardReq struct {
	g.Meta  `path:"/mail_forward/add" method:"post" summary:"add mail forward"`
	Address string `json:"address" v:"required|email" desc:"forwarded email address, e.g. user@example.com"`
	Goto    string `json:"goto" v:"required" desc:"forward target address, multiple addresses separated by newline"`
	Active  int    `json:"active" v:"in:0,1" desc:"status: 1-enabled, 0-disabled" default:"1"`
}

type AddMailForwardRes struct {
	api_v1.StandardRes
}

type EditMailForwardReq struct {
	g.Meta  `path:"/mail_forward/edit" method:"post" summary:"edit mail forward"`
	Address string `json:"address" v:"required" desc:"forwarded email address"`
	Goto    string `json:"goto" desc:"forward target address, multiple addresses separated by newline. not set means only modify status"`
	Active  int    `json:"active" v:"in:0,1" desc:"status: 1-enabled, 0-disabled"`
}

type EditMailForwardRes struct {
	api_v1.StandardRes
}

type DeleteMailForwardReq struct {
	g.Meta  `path:"/mail_forward/delete" method:"post" summary:"delete mail forward"`
	Address string `json:"address" v:"required" desc:"forwarded email address"`
}

type DeleteMailForwardRes struct {
	api_v1.StandardRes
}
