// =====================================================
// Define the standard response format for the API.
// =====================================================

package api_v1

import "github.com/gogf/gf/v2/frame/g"

// StandardRes response format
type StandardRes struct {
	g.Meta   `mime:"application/json"`
	Success  bool        `json:"success"`
	Code     int         `json:"code"`
	Msg      string      `json:"msg" dc:"Success message"`
	ErrorMsg string      `json:"error_msg" dc:"Error message"`
	Data     interface{} `json:"data"`
}

// StandardDownloadRes download response format
type StandardDownloadRes struct {
	g.Meta   `mime:"application/octet-stream"`
	Filename string `json:"filename" dc:"Full path of the downloaded file"`
}

// StandardPagination pagination data structure
type StandardPagination struct {
	Total int `json:"total" dc:"count all"`
	List  int `json:"list" dc:"data list for pagination"`
}
