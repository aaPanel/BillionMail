package batch_mail

import (
	"billionmail-core/internal/model/entity"
	"billionmail-core/internal/service/batch_mail"
	"billionmail-core/internal/service/public"
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/batch_mail/v1"
)

func (c *ControllerV1) TaskInfo(ctx context.Context, req *v1.TaskInfoReq) (res *v1.TaskInfoRes, err error) {
	res = &v1.TaskInfoRes{}

	var taskInfo *entity.EmailTask
	taskInfo, err = batch_mail.GetTaskInfo(ctx, req.Id)
	if err != nil {
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to get task info: {}", err.Error())))
		return
	}

	if taskInfo == nil || taskInfo.Id == 0 {
		res.Code = 404
		res.SetError(gerror.New(public.LangCtx(ctx, "Task not found: {}", req.Id)))
		return
	}
	res.Data = taskInfo
	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return
}
