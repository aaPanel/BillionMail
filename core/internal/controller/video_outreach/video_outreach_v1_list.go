package video_outreach

import (
	v1 "billionmail-core/api/video_outreach/v1"
	"billionmail-core/internal/model/entity"
	vo "billionmail-core/internal/service/video_outreach"
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

func (c *ControllerV1) ListVideoOutreach(ctx context.Context, req *v1.ListVideoOutreachReq) (res *v1.ListVideoOutreachRes, err error) {
	res = &v1.ListVideoOutreachRes{}

	q := g.DB().Model("bm_contacts")

	if req.GroupID > 0 {
		q = q.Where("group_id", req.GroupID)
	}

	// Filter by tier via JSONB attribs (cast to text for LIKE since GoFrame escapes ?)
	if req.Tier != "" {
		q = q.Where("attribs::text LIKE ?", `%"lead_tier":"`+req.Tier+`"%`)
	} else {
		// Only show contacts that have lead scoring data
		q = q.Where("attribs::text LIKE ?", `%"lead_tier"%`)
	}

	// Count total
	total, err := q.Count()
	if err != nil {
		g.Log().Error(ctx, "count video outreach contacts: %v", err)
		total = 0
	}

	// Fetch page
	var contacts []entity.Contact
	err = q.OrderDesc("create_time").
		Page(req.Page, req.PageSize).
		Scan(&contacts)
	if err != nil {
		g.Log().Error(ctx, "list video outreach contacts: %v", err)
		res.SetSuccess("Listed video outreach contacts")
		return
	}

	list := make([]*v1.VideoStatus, 0, len(contacts))
	for _, c := range contacts {
		status := &v1.VideoStatus{
			ContactEmail: c.Email,
		}
		if c.Attribs != nil {
			status.LeadTier = c.Attribs["lead_tier"]
			status.LeadScore = c.Attribs["lead_score"]
			status.VideoURL = c.Attribs[vo.AttrVideoURL]
			status.ThumbnailURL = c.Attribs[vo.AttrThumbnailURL]
			status.LandingPageURL = c.Attribs[vo.AttrLandingPageURL]

			if vo.HasVideoAssets(c.Attribs) {
				status.Status = "ready"
			} else if status.LeadTier == "tier_1" || status.LeadTier == "tier_2" {
				status.Status = "pending"
			} else {
				status.Status = "not_eligible"
			}
		}
		list = append(list, status)
	}

	res.Data.Total = total
	res.Data.List = list
	res.SetSuccess("Listed video outreach contacts")
	return
}
