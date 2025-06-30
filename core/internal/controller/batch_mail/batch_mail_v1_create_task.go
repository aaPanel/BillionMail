package batch_mail

import (
	"billionmail-core/internal/service/batch_mail"
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/batch_mail/v1"
	"billionmail-core/internal/service/email_template"
)

func (c *ControllerV1) CreateTask(ctx context.Context, req *v1.CreateTaskReq) (res *v1.CreateTaskRes, err error) {
	res = &v1.CreateTaskRes{}

	// check template
	template, err := email_template.GetTemplate(ctx, req.TemplateId)
	if err != nil {
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to get template")))
		return
	}
	if template == nil {
		res.Code = 400
		res.SetError(gerror.New(public.LangCtx(ctx, "Template not found")))
		return
	}

	addType := 0
	// create task and import recipients
	res.Data.Id, err = batch_mail.CreateTaskWithRecipients(ctx, req, addType)
	if err != nil {
		res.SetError(err)
		return
	}

	res.SetSuccess(public.LangCtx(ctx, "Task created successfully"))
	return
}
