package batch_mail

import (
	"billionmail-core/internal/model/entity"
	"billionmail-core/internal/service/batch_mail"
	"billionmail-core/internal/service/domains"
	"billionmail-core/internal/service/mail_service"
	"strings"

	"billionmail-core/internal/service/public"
	"context"
	"fmt"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"billionmail-core/api/batch_mail/v1"
)

func (c *ControllerV1) SendTestEmail(ctx context.Context, req *v1.SendTestEmailReq) (res *v1.SendTestEmailRes, err error) {

	res = &v1.SendTestEmailRes{}
	sender, err := mail_service.NewEmailSenderWithLocal(req.Addresser)
	if err != nil {
		res.Code = 400
		res.SetError(gerror.New(public.LangCtx(ctx, "create email sender failed: : {}", err.Error())))
		return
	}
	defer sender.Close()

	var template entity.EmailTemplate

	err = g.DB().Model("email_templates").
		Where("id", req.TemplateId).
		Scan(&template)

	if err != nil {
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "failed to get template {}", err)))
		return
	}

	if !strings.Contains(template.Content, "{{ UnsubscribeURL . }}") {
		template.Content = public.AddUnsubscribeButton(template.Content)
	}

	content := strings.ReplaceAll(template.Content, "__UNSUBSCRIBE_URL__", "{{ UnsubscribeURL . }}")

	jwtToken, _ := batch_mail.GenerateUnsubscribeJWT(req.Recipient, req.TemplateId, 0)
	domain := domains.GetBaseURL()
	unsubscribeURL := fmt.Sprintf("%s/api/unsubscribe", domain)
	groupURL := fmt.Sprintf("%s/api/unsubscribe/user_group", domain)
	unsubscribeJumpURL := fmt.Sprintf("%s/unsubscribe.html?jwt=%s&email=%s&url_type=%s&url_unsubscribe=%s",
		domain, jwtToken, req.Recipient, groupURL, unsubscribeURL)

	var contact entity.Contact
	err = g.DB().Model("bm_contacts").Where("email", req.Recipient).Scan(&contact)
	if err != nil {
		g.Log().Error(ctx, "Failed to get contact: %v", err)
	}

	engine := batch_mail.GetTemplateEngine()
	personalizedContent, err := engine.RenderEmailTemplate(ctx, content, &contact, nil, unsubscribeJumpURL)
	if err != nil {
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "failed to render email content: {}", err)))
		return
	}

	personalizedSubject, err := engine.RenderEmailTemplate(ctx, req.Subject, &contact, nil, unsubscribeJumpURL)
	if err != nil {
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "failed to render email subject: {}", err)))
		return
	}

	message := mail_service.NewMessage(personalizedSubject, personalizedContent)
	message.SetMessageID(sender.GenerateMessageID())

	// send email
	err = sender.Send(message, []string{req.Recipient})
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "send email to {} failed: {}", req.Recipient, err)))
		return
	}

	res.SetSuccess(public.LangCtx(ctx, "send email successfully"))
	return
}
