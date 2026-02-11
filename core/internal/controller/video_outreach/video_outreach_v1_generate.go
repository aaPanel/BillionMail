package video_outreach

import (
	v1 "billionmail-core/api/video_outreach/v1"
	"billionmail-core/internal/model/entity"
	"billionmail-core/internal/service/video_gen"
	vo "billionmail-core/internal/service/video_outreach"
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

func (c *ControllerV1) GenerateVideo(ctx context.Context, req *v1.GenerateVideoReq) (res *v1.GenerateVideoRes, err error) {
	res = &v1.GenerateVideoRes{}

	// Look up the contact
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

	// Check eligibility
	if contact.Attribs == nil {
		res.SetError(gerror.New("Contact has no lead scoring data"))
		return
	}

	if !vo.IsVideoEligible(contact.Attribs) {
		if contact.Attribs["lead_tier"] != "tier_1" {
			res.SetError(gerror.New("Contact is not tier_1"))
			return
		}
		// Tier 1 but no assets yet — enqueue job
		_, enqErr := video_gen.EnqueueVideoJob(ctx, contact.Id, req.ContactEmail, req.GroupID)
		if enqErr != nil {
			res.SetError(gerror.Newf("Failed to enqueue video job: %v", enqErr))
			return
		}
		res.Data.ContactEmail = req.ContactEmail
		res.Data.Status = "queued"
		res.SetSuccess("Video generation queued")
		return
	}

	// Already has video assets
	res.Data.ContactEmail = req.ContactEmail
	res.Data.Status = "completed"
	res.SetSuccess("Video already generated")
	return
}
