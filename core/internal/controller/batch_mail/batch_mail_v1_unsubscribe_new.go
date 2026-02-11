package batch_mail

import (
	"billionmail-core/api/batch_mail/v1"
	"billionmail-core/internal/controller/subscribe_list"
	"billionmail-core/internal/model/entity"
	"billionmail-core/internal/service/batch_mail"
	"billionmail-core/internal/service/contact_activity"
	"billionmail-core/internal/service/domains"
	"billionmail-core/internal/service/public"
	"context"
	"fmt"
	"github.com/gogf/gf/os/gtimer"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"strings"
	"time"
)

func (c *ControllerV1) UnsubscribeNew(ctx context.Context, req *v1.UnsubscribeNewReq) (res *v1.UnsubscribeNewRes, err error) {
	res = &v1.UnsubscribeNewRes{}

	claims, err := batch_mail.ParseUnsubscribeJWT(req.Jwt)
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Invalid token")))
		return
	}

	g.Log().Info(ctx, "UnsubscribeNew request - Email: %s, Template: %d, Task: %d, GroupId: %d",
		claims.Email, claims.TemplateId, claims.TaskId, claims.GroupId)

	if claims.Email == "" {
		res.SetError(gerror.New(public.LangCtx(ctx, "Email address is required")))
		return
	}

	if claims.GroupId <= 0 {
		res.SetError(gerror.New(public.LangCtx(ctx, "Invalid group information")))
		return
	}

	var groupInfo entity.ContactGroup

	err = g.DB().Model("bm_contact_groups").
		Where("id", claims.GroupId).
		Scan(&groupInfo)
	if err != nil {
		g.Log().Error(ctx, "Failed to get group info: %v", err)
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to get group information")))
		return
	}

	// Update contact activity when user accesses unsubscribe page (user interaction)
	contact_activity.UpdateActivityByEmailAndGroup(claims.Email, claims.GroupId)

	if groupInfo.Id == 0 {
		res.SetError(gerror.New(public.LangCtx(ctx, "Group not found")))
		return
	}

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {

		result, err := tx.Model("bm_contacts").
			Where("email", claims.Email).
			Where("group_id", claims.GroupId).
			Data(g.Map{"active": 0}).
			Update()

		if err != nil {
			g.Log().Error(ctx, "Failed to update contact status: %v", err)
			return err
		}

		affected, _ := result.RowsAffected()
		if affected == 0 {
			g.Log().Warning(ctx, "No contact record found for email: %s in group: %d", claims.Email, claims.GroupId)
		}

		_, err = tx.Model("unsubscribe_records").InsertIgnore(g.Map{
			"email":            claims.Email,
			"group_id":         claims.GroupId,
			"template_id":      claims.TemplateId,
			"task_id":          claims.TaskId,
			"unsubscribe_time": time.Now().Unix(),
		})

		if err != nil {
			g.Log().Error(ctx, "Failed to record unsubscribe: %v", err)
			return err
		}

		return nil
	})

	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to process unsubscribe request")))
		return
	}

	var senderEmail string
	if claims.TaskId > 0 {
		val, _ := g.DB().Model("email_tasks").Where("id", claims.TaskId).Value("addresser")
		senderEmail = val.String()
	}
	hostUrl := domains.GetBaseURLBySender(senderEmail)

	if groupInfo.SendUnsubscribeEmail == 1 {
		if groupInfo.UnsubscribeMailHtml == "" {
			//groupInfo.UnsubscribeMailHtml, _ = readTemplateUnsubscribeFiles()
			groupInfo.UnsubscribeMailHtml, _ = subscribe_list.GetDefaultTemplate(3)
		}
		if groupInfo.UnsubscribeSubject == "" {
			groupInfo.UnsubscribeSubject = "You're unsubscribed"
		}

		if strings.Contains(groupInfo.UnsubscribeMailHtml, "{{ ResubscribeLink . }}") {

			confirmToken := subscribe_list.GenerateConfirmToken(claims.Email, groupInfo.Token)
			subscribeURL := fmt.Sprintf("%s/api/subscribe/confirm?token=%s&email=%s", hostUrl, confirmToken, claims.Email)
			groupInfo.UnsubscribeMailHtml = strings.ReplaceAll(groupInfo.UnsubscribeMailHtml, "{{ ResubscribeLink . }}", subscribeURL)
		}

		gtimer.AddOnce(500*time.Millisecond, func() {
			err = subscribe_list.SendMail(ctx, groupInfo.UnsubscribeMailHtml, claims.Email, groupInfo.UnsubscribeSubject, "")
			if err != nil {
				g.Log().Error(ctx, "Failed to send welcome email: {}", err)
				return
			}
		})
	}

	res.Data = v1.UnsubscribeResult{
		Email:       claims.Email,
		GroupName:   groupInfo.Name,
		RedirectUrl: groupInfo.UnsubscribeRedirectUrl,
	}

	res.SetSuccess(public.LangCtx(ctx, "You have been successfully unsubscribed"))
	return
}
