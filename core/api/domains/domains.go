// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package domains

import (
	"context"

	"billionmail-core/api/domains/v1"
)

type IDomainsV1 interface {
	AddDomain(ctx context.Context, req *v1.AddDomainReq) (res *v1.AddDomainRes, err error)
	UpdateDomain(ctx context.Context, req *v1.UpdateDomainReq) (res *v1.UpdateDomainRes, err error)
	DeleteDomain(ctx context.Context, req *v1.DeleteDomainReq) (res *v1.DeleteDomainRes, err error)
	GetDomain(ctx context.Context, req *v1.GetDomainReq) (res *v1.GetDomainRes, err error)
	GetDomainAll(ctx context.Context, req *v1.GetDomainAllReq) (res *v1.GetDomainAllRes, err error)
	FreshDNSRecords(ctx context.Context, req *v1.FreshDNSRecordsReq) (res *v1.FreshDNSRecordsRes, err error)
	SetSSL(ctx context.Context, req *v1.SetSSLReq) (res *v1.SetSSLRes, err error)
	GetSSL(ctx context.Context, req *v1.GetSSLReq) (res *v1.GetSSLRes, err error)
	ApplyCert(ctx context.Context, req *v1.ApplyCertReq) (res *v1.ApplyCertRes, err error)
	GetCertList(ctx context.Context, req *v1.GetCertListReq) (res *v1.GetCertListRes, err error)
	ConsoleApplyCert(ctx context.Context, req *v1.ConsoleApplyCertReq) (res *v1.ConsoleApplyCertRes, err error)
}
