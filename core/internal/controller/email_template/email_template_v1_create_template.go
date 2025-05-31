package email_template

import (
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/email_template/v1"
	"billionmail-core/internal/service/email_template"
)

func (c *ControllerV1) CreateTemplate(ctx context.Context, req *v1.CreateTemplateReq) (res *v1.CreateTemplateRes, err error) {
	res = &v1.CreateTemplateRes{}

	// Check if template name exists
	exists, err := email_template.CheckTemplateNameExists(ctx, req.TempName)
	if err != nil {
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to check template name")))
		return
	}
	if exists {
		res.Code = 400
		res.SetError(gerror.New(public.LangCtx(ctx, "Template name already exists")))
		return
	}

	var content, render string
	if req.AddType == 0 { // upload
		if req.Content == "" {
			res.Code = 400
			res.SetError(gerror.New(public.LangCtx(ctx, "File data is required for upload type")))
			return
		}

		content = req.Content

	} else { // Drag to generate
		if req.Content == "" || req.Render == "" {
			res.Code = 400
			res.SetError(gerror.New(public.LangCtx(ctx, "Content and render data are required for drag type")))
			return
		}
		content = req.Content
		render = req.Render
	}

	// Create template
	id, err := email_template.CreateTemplate(ctx, req.TempName, req.AddType, content, render)
	if err != nil {
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to create template")))
		return
	}

	res.Data.Id = id
	res.SetSuccess(public.LangCtx(ctx, "Template created successfully"))
	return
}
