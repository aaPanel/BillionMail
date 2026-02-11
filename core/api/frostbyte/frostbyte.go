package frostbyte

import (
	"context"

	v1 "billionmail-core/api/frostbyte/v1"
)

type IFrostbyteV1 interface {
	ContactSync(ctx context.Context, req *v1.ContactSyncReq) (res *v1.ContactSyncRes, err error)
	ContactLookup(ctx context.Context, req *v1.ContactLookupReq) (res *v1.ContactLookupRes, err error)
}
