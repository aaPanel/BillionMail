package contact

import (
	"billionmail-core/api/contact/v1"
	"billionmail-core/internal/service/contact"
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) CreateGroup(ctx context.Context, req *v1.CreateGroupReq) (res *v1.CreateGroupRes, err error) {
	res = &v1.CreateGroupRes{}

	var groupId int
	switch req.CreateType {
	case 1, 2: // create new group
		// check if group name exists
		exists, err := contact.CheckGroupNameExists(ctx, req.Name)
		if err != nil {
			res.Code = 500
			res.SetError(gerror.New(public.LangCtx(ctx, "Failed to check group name existence {}", err.Error())))
			return res, err
		}
		if exists {
			res.Code = 400
			res.SetError(gerror.New(public.LangCtx(ctx, "The group name already exists")))
			return res, err
		}

		// create new group
		groupId, err = contact.CreateGroup(ctx, req.Name, req.Description)
		if err != nil {
			res.Code = 500
			res.SetError(gerror.New(public.LangCtx(ctx, "Failed to create group {}", err.Error())))
			return res, err
		}

	case 3: // import to existing group
		// validate group id
		if req.GroupId <= 0 {
			res.Code = 400
			res.SetError(gerror.New(public.LangCtx(ctx, "Group ID is required for importing to existing group")))
			return
		}

		// check if group exists
		_, err = contact.GetGroup(ctx, req.GroupId)
		if err != nil {
			res.Code = 400
			res.SetError(gerror.New(public.LangCtx(ctx, "Group {} does not exist", req.GroupId)))
			return
		}
		groupId = req.GroupId
	}

	// if type 2 or 3, need to handle file import
	if req.CreateType == 2 || req.CreateType == 3 {
		if req.FileData == "" {
			//Allow 2 to create an empty group
			if req.CreateType == 2 {
				res.SetSuccess(public.LangCtx(ctx, "Group created successfully"))
				return
			}
			res.Code = 400
			res.SetError(gerror.New(public.LangCtx(ctx, "File content cannot be empty")))
			return
		}

		// parse contacts from file content
		contacts := parseEmailContent(ctx, req.FileData, 1)
		if len(contacts) == 0 {
			res.Code = 400
			res.SetError(gerror.New(public.LangCtx(ctx, "No valid contacts found in file")))
			return
		}

		// set group id
		for _, contactInfo := range contacts {
			contactInfo.GroupId = groupId
			contactInfo.Status = req.Status
		}

		err = contact.BatchCreateContacts(ctx, contacts)
		if err != nil {
			res.Code = 500
			res.SetError(gerror.New(public.LangCtx(ctx, "Failed to import contacts {}", err.Error())))
			return
		}
	}

	// set success message according to different situations
	switch req.CreateType {
	case 1:
		res.SetSuccess(public.LangCtx(ctx, "Group created successfully"))
	case 2:
		res.SetSuccess(public.LangCtx(ctx, "Group created and contacts imported successfully"))
	case 3:
		res.SetSuccess(public.LangCtx(ctx, "Contacts imported to existing group successfully"))
	}

	return
}
