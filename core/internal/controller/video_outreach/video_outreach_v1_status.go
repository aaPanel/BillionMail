package video_outreach

import (
	v1 "billionmail-core/api/video_outreach/v1"
	"billionmail-core/internal/model/entity"
	vo "billionmail-core/internal/service/video_outreach"
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

func (c *ControllerV1) GetVideoStatus(ctx context.Context, req *v1.GetVideoStatusReq) (res *v1.GetVideoStatusRes, err error) {
	res = &v1.GetVideoStatusRes{}

	var contact entity.Contact
	err = g.DB().Model("bm_contacts").
		Where("email", req.ContactEmail).
		Where("group_id", req.GroupID).
		OrderDesc("create_time").
		Limit(1).
		Scan(&contact)
	if err != nil {
		res.SetError(gerror.New("Failed to look up contact"))
		return
	}
	if contact.Id == 0 {
		res.SetError(gerror.New("Contact not found"))
		return
	}

	status := &v1.VideoStatus{
		ContactEmail: contact.Email,
	}

	if contact.Attribs != nil {
		status.LeadTier = contact.Attribs["lead_tier"]
		status.LeadScore = contact.Attribs["lead_score"]
		status.VideoURL = contact.Attribs[vo.AttrVideoURL]
		status.ThumbnailURL = contact.Attribs[vo.AttrThumbnailURL]
		status.LandingPageURL = contact.Attribs[vo.AttrLandingPageURL]

		if vo.HasVideoAssets(contact.Attribs) {
			status.Status = "ready"
		} else if status.LeadTier == "tier_1" || status.LeadTier == "tier_2" {
			status.Status = "pending"
		} else {
			status.Status = "not_eligible"
		}
	} else {
		status.Status = "not_eligible"
	}

	res.Data = status
	res.SetSuccess("Video status retrieved")
	return
}
