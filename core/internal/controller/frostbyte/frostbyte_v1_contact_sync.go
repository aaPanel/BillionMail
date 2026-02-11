package frostbyte

import (
	v1 "billionmail-core/api/frostbyte/v1"
	fb "billionmail-core/internal/service/frostbyte"
	"context"
)

func (c *ControllerV1) ContactSync(ctx context.Context, req *v1.ContactSyncReq) (res *v1.ContactSyncRes, err error) {
	result, err := fb.SyncContact(ctx, fb.ContactSyncData{
		Email:   req.Email,
		GroupID: req.GroupID,
		Attribs: req.Attribs,
	})
	if err != nil {
		return nil, err
	}

	res = &v1.ContactSyncRes{}
	res.Success = true
	res.Data.ContactID = result.ContactID
	res.Data.Created = result.Created
	return
}
