package v1

import (
	"billionmail-core/utility/types/api_v1"
	"github.com/gogf/gf/v2/frame/g"
)

type StatusReq struct {
	g.Meta `path:"/status" tags:"DockerApi" method:"get" summary:"get container status"`
	Name   string `json:"name" v:"required|min-length:1" dc:"Name" dc:"Container name"`
}

type StatusRes struct {
	api_v1.StandardRes
}

type RestartReq struct {
	g.Meta `path:"/restart" tags:"DockerApi" method:"post" summary:"restart container"`
	Name   string `json:"name" v:"required|min-length:1" dc:"Name" dc:"Container name"`
}

type RestartRes struct {
	api_v1.StandardRes
}
