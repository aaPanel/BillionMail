package v1

import (
	"billionmail-core/utility/types/api_v1"

	"github.com/gogf/gf/v2/frame/g"
)

// ContactSyncReq upserts a contact with attribs.
type ContactSyncReq struct {
	g.Meta        `path:"/frostbyte/contacts/sync" method:"post" tags:"FrostByte" summary:"Sync a contact from FrostByte"`
	Authorization string            `json:"authorization" dc:"Authorization" in:"header"`
	Email         string            `json:"email" v:"required|email" dc:"Contact email"`
	GroupID       int               `json:"group_id" v:"required|min:1" dc:"Contact group ID"`
	Attribs       map[string]string `json:"attribs" dc:"Contact attributes"`
}

type ContactSyncRes struct {
	api_v1.StandardRes
	Data struct {
		ContactID int  `json:"contact_id" dc:"Contact ID"`
		Created   bool `json:"created" dc:"Whether the contact was newly created"`
	} `json:"data"`
}

// ContactLookupReq looks up a contact by email.
type ContactLookupReq struct {
	g.Meta        `path:"/frostbyte/contacts/lookup" method:"get" tags:"FrostByte" summary:"Lookup a contact by email"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Email         string `json:"email" v:"required|email" dc:"Contact email" in:"query"`
}

type ContactLookupRes struct {
	api_v1.StandardRes
	Data interface{} `json:"data" dc:"Contact data"`
}
