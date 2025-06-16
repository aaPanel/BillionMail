// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package email_template

import (
	"context"

	"billionmail-core/api/email_template/v1"
)

type IEmailTemplateV1 interface {
	CreateTemplate(ctx context.Context, req *v1.CreateTemplateReq) (res *v1.CreateTemplateRes, err error)
	DeleteTemplate(ctx context.Context, req *v1.DeleteTemplateReq) (res *v1.DeleteTemplateRes, err error)
	UpdateTemplate(ctx context.Context, req *v1.UpdateTemplateReq) (res *v1.UpdateTemplateRes, err error)
	ListTemplates(ctx context.Context, req *v1.ListTemplatesReq) (res *v1.ListTemplatesRes, err error)
	CopyTemplate(ctx context.Context, req *v1.CopyTemplateReq) (res *v1.CopyTemplateRes, err error)
	GetTemplate(ctx context.Context, req *v1.GetTemplateReq) (res *v1.GetTemplateRes, err error)
	GetAllTemplates(ctx context.Context, req *v1.GetAllTemplatesReq) (res *v1.GetAllTemplatesRes, err error)
	CheckEmailContent(ctx context.Context, req *v1.CheckEmailContentReq) (res *v1.CheckEmailContentRes, err error)
}
