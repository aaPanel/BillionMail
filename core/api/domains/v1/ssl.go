package v1

import (
	"billionmail-core/utility/types/api_v1"
	"github.com/gogf/gf/v2/frame/g"
)

type ApplyCertReq struct {
	g.Meta        `path:"/ssl/apply_cert" tags:"SSL" method:"post" sm:"ApplyCert" in:"body"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Domain        string `json:"domain" v:"required|domain"`
}

type ApplyCertRes struct {
	api_v1.StandardRes
}

type GetCertListReq struct {
	g.Meta        `path:"/ssl/get_cert_list" tags:"SSL" method:"post" sm:"GetCertList" in:"body"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
}

type GetCertListRes struct {
	api_v1.StandardRes
}

type ConsoleApplyCertReq struct {
	g.Meta        `path:"/ssl/console_apply_cert" tags:"SSL" method:"post" sm:"ConsoleApplyCert" in:"body"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
}
type ConsoleApplyCertRes struct {
	api_v1.StandardRes
}
