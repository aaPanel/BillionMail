package v1

import "github.com/gogf/gf/v2/frame/g"

type ApplyCertReq struct {
	g.Meta `path:"/ssl/apply_cert" tags:"SSL" method:"post" sm:"ApplyCert" in:"body"`
	Domain string `json:"domain" v:"required|domain"`
}

type ApplyCertRes struct {
	g.Meta `mime:"application/json"`
}

type GetCertListReq struct {
	g.Meta `path:"/ssl/get_cert_list" tags:"SSL" method:"post" sm:"GetCertList" in:"body"`
}

type GetCertListRes struct {
	g.Meta `mime:"application/json"`
}
