package contact

import (
	"billionmail-core/internal/service/contact"
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"billionmail-core/api/contact/v1"
)

func (c *ControllerV1) ListContacts(ctx context.Context, req *v1.ListContactsReq) (res *v1.ListContactsRes, err error) {
	res = &v1.ListContactsRes{}

	if req.Page < 1 {
		req.Page = DefaultPage
	}
	if req.PageSize < 1 || req.PageSize > MaxPageSize {
		req.PageSize = DefaultPageSize
	}
	if req.Status < 0 || req.Status > 1 {
		req.Status = 1 // Default to active
	}

	// Get contacts list
	total, list, err := contact.GetContactsWithPage(ctx, req.Page, req.PageSize, req.GroupId, req.Keyword, req.Status)
	if err != nil {
		g.Log().Errorf(ctx, "Failed to get contacts list: %v, group ID: %d, keyword: %s, status: %d", err, req.GroupId, req.Keyword, req.Status)
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to get contacts list: {}", err.Error())))
		return
	}

	apiList := make([]*v1.Contact, 0)
	emailMap := make(map[string]*v1.Contact)

	// filter out duplicate emails
	for _, contactOne := range list {
		if _, exists := emailMap[contactOne.Email]; !exists {
			contactInfo := &v1.Contact{
				Id:         contactOne.Id,
				Email:      contactOne.Email,
				GroupId:    contactOne.GroupId,
				Active:     contactOne.Active,
				TaskId:     contactOne.TaskId,
				CreateTime: contactOne.CreateTime,
				Groups:     make([]v1.GroupInfo, 0),
				Status:     contactOne.Status,
				Attribs:    contactOne.Attribs,
			}
			emailMap[contactOne.Email] = contactInfo
			apiList = append(apiList, contactInfo)
		}
	}

	// Get groups for each contact
	for email, contactInfo := range emailMap {
		groups, err := contact.GetContactGroupsInfo(ctx, email, req.Status)
		if err != nil {
			g.Log().Warningf(ctx, "Failed to get groups for contact %s: %v", email, err)
			continue
		}

		// Add group information to the contact
		for _, group := range groups {
			contactInfo.Groups = append(contactInfo.Groups, v1.GroupInfo{
				Id:   group.Id,
				Name: group.Name,
			})
		}
	}

	res.Data.Total = total
	res.Data.List = apiList
	res.SetSuccess(public.LangCtx(ctx, "Get contacts list successfully"))
	return
}
