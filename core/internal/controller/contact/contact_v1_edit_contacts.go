package contact

import (
	"billionmail-core/internal/service/contact"
	"billionmail-core/internal/service/public"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/contact/v1"
)

func (c *ControllerV1) EditContacts(ctx context.Context, req *v1.EditContactsReq) (res *v1.EditContactsRes, err error) {
	res = &v1.EditContactsRes{}

	var attribs map[string]string
	if req.Attribs != "" {

		attribsStr := strings.Trim(req.Attribs, "\"")

		attribsStr = strings.ReplaceAll(attribsStr, "\"\"", "\"")

		var rawAttribs map[string]interface{}
		if err := json.Unmarshal([]byte(attribsStr), &rawAttribs); err != nil {
			res.Code = 400
			res.SetError(gerror.New(public.LangCtx(ctx, "Invalid attributes format:{}", err.Error())))
			return res, nil
		}

		attribs = make(map[string]string)
		for k, v := range rawAttribs {
			switch val := v.(type) {
			case string:
				attribs[k] = val
			case float64: // Numbers in JSON parse to float64 by default
				if float64(int64(val)) == val {
					// Integer
					attribs[k] = fmt.Sprintf("%d", int64(val))
				} else {
					// Float
					attribs[k] = fmt.Sprintf("%g", val)
				}
			case bool:
				attribs[k] = fmt.Sprintf("%v", val)
			case nil:
				attribs[k] = ""
			default:
				// Other types try to convert to JSON string
				if bytes, err := json.Marshal(val); err == nil {
					attribs[k] = string(bytes)
				} else {
					attribs[k] = fmt.Sprintf("%v", val)
				}
			}
		}
	}

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {

		if len(req.GroupIds) > 0 {
			for _, groupId := range req.GroupIds {
				if _, err := contact.GetGroup(ctx, groupId); err != nil {
					return gerror.New(public.LangCtx(ctx, "Group {} not found", groupId))
				}
			}
		}

		var existingContact struct {
			Attribs map[string]string `json:"attribs"`
			Status  int               `json:"status"`
		}
		err := g.DB().Model("bm_contacts").
			Where("email", req.Emails).
			Order("id DESC").
			Limit(1).
			Scan(&existingContact)

		var finalAttribs map[string]string
		if attribs != nil {
			// If new attributes are provided, use them directly (without merging)
			finalAttribs = attribs
		} else if err == nil && existingContact.Attribs != nil {
			// If no new attributes are provided, use existing attributes
			finalAttribs = existingContact.Attribs
		} else {

			finalAttribs = make(map[string]string)
		}

		// 3. If group relations need to be updatedrelations need to be updated
		if len(req.GroupIds) > 0 {
			// Delete old contact recordsd contact records
			_, err := g.DB().Model("bm_contacts").
				Where("email", req.Emails).
				Delete()
			if err != nil {
				return gerror.New(public.LangCtx(ctx, "Failed to delete old contact records"))
			}

			// Create new group relations
			now := time.Now().Unix()
			var data []g.Map

			for _, groupId := range req.GroupIds {
				data = append(data, g.Map{
					"email":       req.Emails,
					"group_id":    groupId,
					"active":      req.Active,
					"create_time": now,
					"attribs":     finalAttribs,
					"status":      existingContact.Status,
				})
			}

			if len(data) > 0 {
				_, err := g.DB().Model("bm_contacts").
					Data(data).
					Insert()
				if err != nil {
					return gerror.New(public.LangCtx(ctx, "Failed to create new contact records"))
				}
			}
		}

		return nil
	})

	if err != nil {
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to edit contact: {}", err.Error())))
		return
	}

	res.SetSuccess(public.LangCtx(ctx, "Contact updated successfully"))
	return
}
