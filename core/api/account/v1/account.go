package v1

import (
	"billionmail-core/utility/types/api_v1"
	"github.com/gogf/gf/v2/frame/g"
)

type LoginReq struct {
	g.Meta   `path:"/account/login" tags:"Account" method:"post" sm:"Login" in:"body"`
	Username string `json:"username" v:"required#Lang{Missing username}" dc:"Username"`
	Password string `json:"password" v:"required#Lang{Missing password}" dc:"Password"`
	Language string `json:"language" v:"min-length:1" dc:"Language"`
	Captcha  string `json:"captcha" v:"min-length:1" dc:"Captcha"`
}

type LoginRes struct {
	api_v1.StandardRes
}

type LogoutReq struct {
	g.Meta `path:"/account/logout" tags:"Account" method:"post" sm:"Logout" in:"header"`
}

type LogoutRes struct {
	api_v1.StandardRes
}

type GetAccountsReq struct {
	g.Meta   `path:"/account/get_accounts" tags:"Account" method:"post" sm:"GetAccounts" in:"body"`
	Keyword  string `json:"keyword"  dc:"Keyword"`
	Page     int    `json:"page"    dc:"Page"`
	PageSize int    `json:"page_size"  dc:"PageSize"`
}

type GetAccountsRes struct {
	api_v1.StandardRes
}

type UpdateAccountReq struct {
	g.Meta    `path:"/account/update_account" tags:"Account" method:"post" sm:"UpdateAccount" in:"body"`
	AccountId int    `json:"account_id" v:"required|min:1#Lang{Missing account ID}|Lang{Account ID must be greater than 0}" dc:"Account ID"`
	Password  string `json:"password" v:"required#Lang{Missing password}" dc:"Password"`
	Email     string `json:"email" v:"required#Lang{Missing email}" dc:"Email"`
	Status    int    `json:"status" dc:"Status"`
	GroupId   int    `json:"group_id" dc:"GroupId"`
}

type UpdateAccountRes struct {
	api_v1.StandardRes
}

type DeleteAccountReq struct {
	g.Meta    `path:"/account/delete_account" tags:"Account" method:"post" sm:"DeleteAccount" in:"body"`
	AccountId int `json:"account_id" v:"required|min:1#Lang{Missing account ID}|Lang{Account ID must be greater than 0}" dc:"Account ID"`
}

type DeleteAccountRes struct {
	api_v1.StandardRes
}

type AddAccountReq struct {
	g.Meta   `path:"/account/add_account" tags:"Account" method:"post" sm:"AddAccount" in:"body"`
	Username string `json:"username" v:"required#Lang{Missing username}" dc:"Username"`
	Password string `json:"password" v:"required#Lang{Missing password}" dc:"Password"`
	Email    string `json:"email" v:"required#Lang{Missing email}" dc:"Email"`
	GroupId  int    `json:"group_id" dc:"GroupId"`
	Status   int    `json:"status" dc:"Status"`
	Language string `json:"language" v:"min-length:1" dc:"Language"`
}

type AddAccountRes struct {
	api_v1.StandardRes
}
