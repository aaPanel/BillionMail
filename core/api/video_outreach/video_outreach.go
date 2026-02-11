package video_outreach

import (
	"context"

	v1 "billionmail-core/api/video_outreach/v1"
)

type IVideoOutreachV1 interface {
	GenerateVideo(ctx context.Context, req *v1.GenerateVideoReq) (res *v1.GenerateVideoRes, err error)
	GetVideoStatus(ctx context.Context, req *v1.GetVideoStatusReq) (res *v1.GetVideoStatusRes, err error)
	ListVideoOutreach(ctx context.Context, req *v1.ListVideoOutreachReq) (res *v1.ListVideoOutreachRes, err error)
}
