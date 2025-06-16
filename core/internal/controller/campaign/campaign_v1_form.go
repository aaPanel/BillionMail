package campaign

import (
	"billionmail-core/internal/model/entity"
	"billionmail-core/internal/service/contact"
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/grand"
	"time"

	"billionmail-core/api/campaign/v1"
)

func (c *ControllerV1) Form(ctx context.Context, req *v1.FormReq) (res *v1.FormRes, err error) {
	r := g.RequestFromCtx(ctx)

	if r.Method == "POST" {
		// Verify Csrf token if needed
		if req.Token == "" {
			r.Response.WriteStatus(400, "CSRF token not provided")
			return
		}

		var v *g.Var
		if v, err = r.Session.Get("csrf_token"); err != nil {
			r.Response.WriteStatus(400, "CSRF token not valid")
			return
		}

		if v.String() != req.Token {
			r.Response.WriteStatus(400, "CSRF token not valid")
			return
		}

		// Clear CSRF token from session
		if err = r.Session.Remove("csrf_token"); err != nil {
			r.Response.WriteStatus(500, "Failed to clear CSRF token")
			return
		}

		contacts := make([]*entity.Contact, 0, len(req.GroupIds))

		for _, groupId := range req.GroupIds {
			contacts = append(contacts, &entity.Contact{
				Email:      req.Email,
				GroupId:    groupId,
				Active:     1,
				CreateTime: int(time.Now().Unix()),
				Attribs: map[string]string{
					"Username": req.Name,
				},
			})
		}

		cnt, _ := contact.BatchCreateContactsIgnoreDuplicate(ctx, contacts)

		if cnt > 0 {
			// TODO : Send confirm email to the user
		}

		err = r.Response.WriteTpl("subscription_success.html", g.Map{
			"EmailAddress": req.Email,
		})
		return
	}

	// Generate CSRF token
	csrfToken := grand.S(16)

	// Store CSRF token in session
	if err = r.Session.Set("csrf_token", csrfToken); err != nil {
		csrfToken = ""
	}

	groups, _ := contact.GetAllGroups(ctx, "")

	err = r.Response.WriteTpl("subscription_form.html", g.Map{
		"Groups":    groups,
		"CSRFToken": csrfToken,
	})
	return
}
