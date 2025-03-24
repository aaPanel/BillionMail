package v1

import "github.com/gogf/gf/v2/frame/g"

type OverviewReq struct {
	g.Meta    `path:"/overview" tags:"Overview" method:"get" summary:"Get overview information"`
	Domains   []string `v:"required" dc:"Domains" dc:"Domains"`
	StartTime int      `v:"required" dc:"StartTime" dc:"StartTime"`
	EndTime   int      `v:"required" dc:"EndTime" dc:"EndTime"`
}

type OverviewRes struct {
	g.Meta `mime:"application/json"`
}
