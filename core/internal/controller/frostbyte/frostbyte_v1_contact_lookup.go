package frostbyte

import (
	v1 "billionmail-core/api/frostbyte/v1"
	fb "billionmail-core/internal/service/frostbyte"
	"context"
)

func (c *ControllerV1) ContactLookup(ctx context.Context, req *v1.ContactLookupReq) (res *v1.ContactLookupRes, err error) {
	contact, err := fb.LookupContact(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	res = &v1.ContactLookupRes{}
	res.Success = true
	res.Data = contact
	return
}
