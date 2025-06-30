package warmup

import (
	"billionmail-core/internal/consts"
	"billionmail-core/internal/model/entity"
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"math"
)

// CampaignWarmupAssociation extends CampaignWarmup with prediction data.
type CampaignWarmupAssociation struct {
	entity.CampaignWarmup
	EstimatedSeconds int64 `json:"estimated_seconds"`
}

// WarmupCampaignService manages the association between warmup processes and marketing campaigns.
type WarmupCampaignService struct {
	ipWarmupService *SenderIpWarmupService
	providerService *SenderIpMailProviderService
}

var insWarmupCampaignService = WarmupCampaignService{
	ipWarmupService: SenderIpWarmup(),
	providerService: SenderIpMailProvider(),
}

// WarmupCampaign returns the singleton instance of the service.
func WarmupCampaign() *WarmupCampaignService {
	return &insWarmupCampaignService
}

// AssociateCampaignWithWarmup associates a marketing task with an IP warmup process.
// If a warmup process for the given IP does not exist, it will be created.
func (s *WarmupCampaignService) AssociateCampaignWithWarmup(ctx context.Context, taskId int64, senderIp string) (*CampaignWarmupAssociation, error) {
	// 1. Get or create the warmup status for the IP to get its ID.
	warmupStatus, err := s.ipWarmupService.InitializeOrGetWarmupStatus(ctx, senderIp)
	if err != nil {
		g.Log().Errorf(ctx, "AssociateCampaignWithWarmup: failed to get warmup status for IP %s: %v", senderIp, err)
		return nil, err
	}
	warmupId := warmupStatus.Id

	// Calculate the estimated sending time
	estimatedSeconds, err := s.CalculateEstimatedTime(ctx, taskId, senderIp)
	if err != nil {
		// Log the error but don't fail the entire association
		g.Log().Warningf(ctx, "AssociateCampaignWithWarmup: failed to calculate estimated time for task %d: %v", taskId, err)
	}

	// 2. Check if an association for this task already exists.
	var existingLink *entity.CampaignWarmup
	err = g.DB().Model("bm_campaign_warmup").Ctx(ctx).Where("task_id", taskId).Scan(&existingLink)
	if err != nil {
		g.Log().Errorf(ctx, "AssociateCampaignWithWarmup: failed to check for existing link for task ID %d: %v", taskId, err)
		return nil, err
	}

	if existingLink != nil {
		if existingLink.WarmupId == warmupId {
			g.Log().Debugf(ctx, "AssociateCampaignWithWarmup: Task ID %d is already associated with warmup ID %d.", taskId, warmupId)
			return &CampaignWarmupAssociation{
				CampaignWarmup:   *existingLink,
				EstimatedSeconds: estimatedSeconds,
			}, nil
		}
		// If the task is associated with a different warmup, update it.
		existingLink.WarmupId = warmupId
		_, err = g.DB().Model("bm_campaign_warmup").Ctx(ctx).Where("id", existingLink.Id).Data(g.Map{
			"warmup_id": warmupId,
		}).Update()
		if err != nil {
			g.Log().Errorf(ctx, "AssociateCampaignWithWarmup: failed to update association for task ID %d to warmup ID %d: %v", taskId, warmupId, err)
			return nil, err
		}
		g.Log().Infof(ctx, "AssociateCampaignWithWarmup: Updated association for task ID %d to new warmup ID %d.", taskId, warmupId)
		return &CampaignWarmupAssociation{
			CampaignWarmup:   *existingLink,
			EstimatedSeconds: estimatedSeconds,
		}, nil
	}

	// 3. Create a new association.
	newLink := &entity.CampaignWarmup{
		TaskId:   taskId,
		WarmupId: warmupId,
	}

	newLink.Id, err = g.DB().Model("bm_campaign_warmup").Ctx(ctx).Data(newLink).OmitEmptyData().InsertAndGetId()
	if err != nil {
		g.Log().Errorf(ctx, "AssociateCampaignWithWarmup: failed to create association for task ID %d and warmup ID %d: %v", taskId, warmupId, err)
		return nil, err
	}

	result := &CampaignWarmupAssociation{
		CampaignWarmup:   *newLink,
		EstimatedSeconds: estimatedSeconds,
	}

	g.Log().Infof(ctx, "Successfully associated task ID %d with warmup ID %d (IP: %s). Estimated time: %d seconds.", taskId, warmupId, senderIp, estimatedSeconds)
	return result, nil
}

// CalculateEstimatedTime calculates the estimated sending time in seconds for a given task and IP.
func (s *WarmupCampaignService) CalculateEstimatedTime(ctx context.Context, taskId int64, senderIp string) (int64, error) {
	// Get the total number of unsent recipients
	unsentCount, err := g.DB().Model("recipient_info").Ctx(ctx).Where("task_id", taskId).Where("is_sent", 0).Count()
	if err != nil {
		return -1, fmt.Errorf("failed to get recipient count: %w", err)
	}

	if unsentCount == 0 {
		return 0, nil
	}

	// Calculate the total hourly sending rate
	totalHourlyRate := 0
	providerGroups := []string{
		consts.MailProviderGroupGmail,
		consts.MailProviderGroupYahoo,
		consts.MailProviderGroupOutlook,
		consts.MailProviderGroupApple,
		consts.MailProviderGroupProton,
		consts.MailProviderGroupZoho,
		consts.MailProviderGroupAmazon,
		consts.MailProviderGroupOther,
	}

	for _, group := range providerGroups {
		dailyLimit, hourlyLimit, err := s.providerService.GetAdjustedSendingLimitsForProvider(ctx, senderIp, group)
		if err == nil {
			totalHourlyRate += int(math.Min(float64(hourlyLimit), float64(dailyLimit)/24.0))
		} else {
			g.Log().Warningf(ctx, "calculateEstimatedTime: unable to get limits for group %s, IP %s: %v", group, senderIp, err)
		}
	}

	if totalHourlyRate == 0 {
		return -1, nil // -1 indicates infinite time
	}

	estimatedSeconds := int64((float64(unsentCount) / float64(totalHourlyRate)) * 3600.0)
	return estimatedSeconds, nil
}

// GetWarmupStatusForCampaign retrieves the warmup status for a given marketing task.
func (s *WarmupCampaignService) GetWarmupStatusForCampaign(ctx context.Context, taskId int64) (*entity.SenderIpWarmup, error) {
	var link *entity.CampaignWarmup
	err := g.DB().Model("bm_campaign_warmup").Ctx(ctx).Where("task_id", taskId).Scan(&link)
	if err != nil {
		g.Log().Errorf(ctx, "GetWarmupStatusForCampaign: failed to find association for task ID %d: %v", taskId, err)
		return nil, err
	}
	if link == nil {
		g.Log().Debugf(ctx, "GetWarmupStatusForCampaign: no warmup association found for task ID %d.", taskId)
		return nil, nil // Or return a specific error like gerror.New("not found")
	}

	var warmupStatus *entity.SenderIpWarmup
	err = g.DB().Model("bm_sender_ip_warmup").Ctx(ctx).Where("id", link.WarmupId).Scan(&warmupStatus)
	if err != nil {
		g.Log().Errorf(ctx, "GetWarmupStatusForCampaign: failed to get warmup status for warmup ID %d (task ID %d): %v", link.WarmupId, taskId, err)
		return nil, err
	}

	return warmupStatus, nil
}
