package warmup

import (
	"billionmail-core/internal/consts"
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"math"
	"strings"
	"time"

	"billionmail-core/internal/model/entity"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

const (
	minScoreForSendingProvider      = 25                 // If the score is below this value, the (IP, provider group) combination will be suspended.
	evaluationIntervalProvider      = 2 * time.Hour      // The evaluation frequency for provider group scores.
	defaultEvaluationWindowProvider = 3 * 24 * time.Hour // The data window for provider group scoring, e.g., the last 3 days.
	minSendVolumeForProviderScoring = 15                 // The minimum sending volume required to calculate a full score for a provider group.
	initialProviderScore            = 40                 // The initial score for an (IP, provider group).
)

// SenderIpMailProviderService struct.
type SenderIpMailProviderService struct {
	ipWarmupService      *SenderIpWarmupService      // Depends on the overall IP warmup service.
	dailyUpstairsService *WarmupDailyUpstairsService // Depends on the daily sending volume service.
}

var insSenderIpMailProviderService = SenderIpMailProviderService{
	ipWarmupService:      SenderIpWarmup(),      // Get an instance of SenderIpWarmupService.
	dailyUpstairsService: WarmupDailyUpstairs(), // Get an instance of WarmupDailyUpstairsService.
}

// SenderIpMailProvider returns the service instance.
func SenderIpMailProvider() *SenderIpMailProviderService {
	return &insSenderIpMailProviderService
}

// getMailProviderGroup returns the corresponding service provider group name based on the mail_provider field in the database (usually a domain name).
func (s *SenderIpMailProviderService) getMailProviderGroup(dbMailProvider string) string {
	return public.GetMailProviderGroup(dbMailProvider)
}

// InitializeOrGetProviderStatus initializes or retrieves the scoring status for a specified IP and mail provider group.
func (s *SenderIpMailProviderService) InitializeOrGetProviderStatus(ctx context.Context, senderIp string, mailProviderGroup string) (record *entity.SenderIpMailProvider, err error) {
	err = g.DB().Model("bm_sender_ip_mail_provider").Ctx(ctx).
		Where("sender_ip", senderIp).
		Where("mail_provider", mailProviderGroup).
		Scan(&record)
	if err != nil {
		g.Log().Errorf(ctx, "Failed to query SenderIpMailProvider for IP %s, Group %s: %v", senderIp, mailProviderGroup, err)
		return
	}

	if record == nil {
		now := gtime.Now()
		record = &entity.SenderIpMailProvider{
			SenderIp:         senderIp,
			MailProvider:     mailProviderGroup,
			Score:            initialProviderScore,
			LastEvaluateTime: now.Timestamp(), // When first created, the evaluation time is set to the current time to avoid immediate re-evaluation.
			CreateTime:       now.Timestamp(),
			UpdateTime:       now.Timestamp(),
		}
		var insertId int64
		insertId, err = g.DB().Model("bm_sender_ip_mail_provider").Ctx(ctx).Data(record).OmitEmptyData().InsertAndGetId()
		if err != nil {
			g.Log().Errorf(ctx, "Failed to insert SenderIpMailProvider for IP %s, Group %s: %v", senderIp, mailProviderGroup, err)
			return
		}
		record.Id = insertId
		g.Log().Infof(ctx, "Initialized SenderIpMailProvider status for IP %s, Group %s", senderIp, mailProviderGroup)
	}
	return
}

// EvaluateScoreForProvider evaluates and updates the score for a specified IP and mail provider group.
func (s *SenderIpMailProviderService) EvaluateScoreForProvider(ctx context.Context, senderIp string, mailProviderGroup string) (score int, err error) {
	providerStatus, err := s.InitializeOrGetProviderStatus(ctx, senderIp, mailProviderGroup)
	if err != nil {
		return 0, err
	}

	// Use the getSenderIdentitiesForIp method from SenderIpWarmupService.
	// Note: This method needs to be exported from SenderIpWarmupService or provided as a public utility function.
	senderIdentities, err := s.ipWarmupService.getSenderIdentitiesForIp(ctx, senderIp) // Assume it can be called
	if err != nil {
		g.Log().Errorf(ctx, "EvaluateScoreForProvider: Failed to get sender identities for IP %s: %v", senderIp, err)
		// Update the evaluation time to avoid getting stuck.
		_, _ = g.DB().Model("bm_sender_ip_mail_provider").Ctx(ctx).
			Where("id", providerStatus.Id).
			Data(g.Map{"last_evaluate_time": gtime.Now().Timestamp()}).Update()
		return providerStatus.Score, err
	}

	evaluationWindowStartMillis := gtime.New(providerStatus.LastEvaluateTime).Add(-defaultEvaluationWindowProvider).TimestampMilli()
	// If it's the first evaluation (LastEvaluateTime == CreateTime), calculate from CreateTime, or from a reasonable earliest time.
	if providerStatus.LastEvaluateTime == providerStatus.CreateTime {
		evaluationWindowStartMillis = gtime.New(providerStatus.CreateTime).Add(-defaultEvaluationWindowProvider).TimestampMilli()
		// Prevent the window from being too early, ensure it starts at least from the creation time.
		if evaluationWindowStartMillis < gtime.New(providerStatus.CreateTime).Add(-2*defaultEvaluationWindowProvider).TimestampMilli() {
			evaluationWindowStartMillis = gtime.New(providerStatus.CreateTime).TimestampMilli()
		}
	}
	if gtime.Now().TimestampMilli()-evaluationWindowStartMillis > defaultEvaluationWindowProvider.Milliseconds()*2 { // Prevent the window from being too large.
		evaluationWindowStartMillis = gtime.Now().Add(-defaultEvaluationWindowProvider).TimestampMilli()
	}

	relevantPmids := make([]string, 0)
	if len(senderIdentities) > 0 {
		var vals []gdb.Value
		// Assume here that senderIdentities are domains or patterns that can be used for LIKE matching.
		likes := make([]string, len(senderIdentities))
		params := make([]interface{}, len(senderIdentities)+1)
		for i, identity := range senderIdentities {
			// Adjust according to the format of identity. If it's a full email address, use sender = ?; if it's a domain, use sender LIKE '%@domain'.
			if strings.Contains(identity, "@") {
				likes[i] = "s.sender = ?"
				params[i] = identity
			} else {
				likes[i] = "s.sender LIKE ?"
				params[i] = "%@" + identity
			}
		}
		params[len(senderIdentities)] = evaluationWindowStartMillis
		vals, err = g.DB().Ctx(ctx).Model("mailstat_senders s").
			InnerJoin("mailstat_send_mails sm", "s.postfix_message_id=sm.postfix_message_id").
			Fields("DISTINCT s.postfix_message_id").
			Where(strings.Join(likes, " OR "), params[:len(senderIdentities)]...).
			Where("s.log_time_millis >= ?", evaluationWindowStartMillis).
			Where("sm.mail_provider", mailProviderGroup).
			Array()
		if err != nil {
			g.Log().Errorf(ctx, "EvaluateScoreForProvider: Failed to get postfix_message_ids for IP %s (identities: %v): %v", senderIp, senderIdentities, err)
		}
		for _, v := range vals {
			if pmid := v.String(); pmid != "" {
				relevantPmids = append(relevantPmids, pmid)
			}
		}
	}

	totalSent := 0
	successfulSends := 0
	hardBounces := 0
	softBounces := 0
	deferredSends := 0
	openedCount := 0
	clickedCount := 0
	// complaintCount := 0 // TODO: Complaint rate

	if len(relevantPmids) > 0 {
		// 1. Get sending statistics (mailstat_send_mails)
		type SendMailStatItemWithProvider struct {
			Status string `json:"status"`
			Dsn    string `json:"dsn"`
			Count  int    `json:"count"`
		}
		var sendMailStats []SendMailStatItemWithProvider
		err = g.DB().Ctx(ctx).Model("mailstat_send_mails").
			Fields("status", "dsn", "COUNT(*) as count").
			WhereIn("postfix_message_id", relevantPmids).
			Where("log_time_millis >= ?", evaluationWindowStartMillis).
			Group("status", "dsn").
			Scan(&sendMailStats)

		if err != nil {
			g.Log().Errorf(ctx, "EvaluateScoreForProvider: Failed to get send mail stats for IP %s, Group %s: %v", senderIp, mailProviderGroup, err)
		} else {
			for _, stat := range sendMailStats {
				totalSent += stat.Count
				switch strings.ToLower(stat.Status) {
				case "sent":
					successfulSends += stat.Count
				case "bounced":
					if len(stat.Dsn) > 0 {
						switch stat.Dsn[0] {
						case '5':
							hardBounces += stat.Count
						case '4':
							softBounces += stat.Count
						default:
							softBounces += stat.Count
						}
					} else {
						softBounces += stat.Count
					}
				case "deferred":
					deferredSends += stat.Count
				}
			}
		}

		// 2. Get open and click statistics (needs to be grouped by recipient domain)
		// Get openedCount and clickedCount (requires user participation data)
		openedCount, _ = g.DB().Model("mailstat_opened").
			Ctx(ctx).
			WhereIn("postfix_message_id", relevantPmids).
			Where("log_time_millis >= ?", evaluationWindowStartMillis).
			Fields("DISTINCT postfix_message_id").
			Count()
		clickedCount, _ = g.DB().Model("mailstat_clicked").
			Ctx(ctx).
			WhereIn("postfix_message_id", relevantPmids).
			Where("log_time_millis >= ?", evaluationWindowStartMillis).
			Fields("DISTINCT postfix_message_id").
			Count()
	}

	currentScore := float64(providerStatus.Score)
	if totalSent == 0 && len(relevantPmids) == 0 && len(senderIdentities) == 0 && mailProviderGroup == consts.MailProviderGroupOther {
		// For the "Other" category, if there are no specific identities or traffic, the score may remain unchanged or decrease slowly.
		// currentScore -= 0.1
		g.Log().Infof(ctx, "EvaluateIpScore: IP %s Group %s has no configured sender identities. Score remains %d.", senderIp, providerStatus.MailProvider, providerStatus.Score)
	} else if totalSent < minSendVolumeForProviderScoring && totalSent > 0 {
		switch {
		case hardBounces > 0:
			currentScore -= 0.2 // Hard bounce penalty
		case softBounces > 0:
			currentScore -= 0.1 // Soft bounce penalty
		case successfulSends > 0:
			currentScore += 0.3 // Reward for successful sending
		default:
			currentScore -= 0.8 // Penalty for no sending records
		}
	} else if totalSent >= minSendVolumeForProviderScoring {
		successRate := float64(successfulSends) / float64(totalSent)
		hardBounceRate := float64(hardBounces) / float64(totalSent)
		softBounceRate := float64(softBounces) / float64(totalSent)
		deferredRate := float64(deferredSends) / float64(totalSent) // deferredSends comes from overall statistics and needs adjustment

		currentScore += math.Max(0, successRate-0.85) * 40 // 85% success rate benchmark, affects 40 points
		currentScore -= hardBounceRate * 1.5               // Higher penalty coefficient for hard bounces
		currentScore -= softBounceRate * 0.7
		currentScore -= deferredRate * 0.25 // Latency also affects provider score

		if successfulSends > 0 {
			openRate := float64(openedCount) / float64(successfulSends)
			currentScore += openRate * 25 // 12% open rate benchmark
			if openedCount > 0 {
				clickRate := float64(clickedCount) / float64(openedCount)
				currentScore += clickRate * 20 // 2.5% click rate benchmark
			}
		}
		// TODO: complaintRate (complaintCount / successfulSends)
	} else { // totalSent == 0 for this provider group
		var warmupStatus *entity.SenderIpWarmup
		warmupStatus, err = s.ipWarmupService.InitializeOrGetWarmupStatus(ctx, senderIp)
		if err == nil && warmupStatus != nil && warmupStatus.Progress > 65 { // Warmup progress exceeds 65%
			switch {
			case providerStatus.Score < 50:
				currentScore -= 1.5 // When the score is below 50, the penalty is heavier
			case providerStatus.Score > 70:
				currentScore -= 0.5 // When the score is above 70, the penalty is lighter
			}
		}
	}

	finalScore := int(math.Max(0, math.Min(100, currentScore)))
	_, err = g.DB().Model("bm_sender_ip_mail_provider").Ctx(ctx).
		Where("id", providerStatus.Id).
		Data(g.Map{
			"score":              finalScore,
			"last_evaluate_time": gtime.Now().Timestamp(),
			"update_time":        gtime.Now().Timestamp(),
		}).Update()

	if err != nil {
		g.Log().Errorf(ctx, "EvaluateScoreForProvider: Failed to update score for IP %s, Group %s: %v", senderIp, mailProviderGroup, err)
		return providerStatus.Score, err
	}
	g.Log().Debugf(ctx, "Evaluated score for IP %s, Group %s: %d (Sent: %d, Success: %d, HB: %d, SB: %d, Open: %d, Click: %d)", senderIp, mailProviderGroup, finalScore, totalSent, successfulSends, hardBounces, softBounces, openedCount, clickedCount)
	return finalScore, nil
}

// GetSendingRateFactorForProvider calculates the sending rate health factor based on the score of a specific provider group.
// This factor (0.0 - 1.0) reflects the reputation health of the IP with that provider.
func (s *SenderIpMailProviderService) GetSendingRateFactorForProvider(ctx context.Context, senderIp string, mailProviderGroup string) (rateFactor float64, err error) {
	providerStatus, err := s.InitializeOrGetProviderStatus(ctx, senderIp, mailProviderGroup)
	if err != nil {
		g.Log().Warningf(ctx, "GetSendingRateFactorForProvider: IP %s, Group %s status unknown: %v. Returning low factor.", senderIp, mailProviderGroup, err)
		return 0.05, err // Unable to get provider group status, return a very low factor.
	}

	minScore := float64(minScoreForSendingProvider)
	score := float64(providerStatus.Score)

	//if score < minScore {
	//	g.Log().Debugf(ctx, "GetSendingRateFactorForProvider: IP %s, Group %s PAUSED (Score %d < %d)", senderIp, mailProviderGroup, providerStatus.Score, int(minScore))
	//	return 0.0, nil // Score is too low, pause sending to this provider group.
	//}

	// Linearly map the score from minScore-100 to a rate factor of 0.1-1.0.
	// Even at the lowest allowed score, a 10% rate is maintained.
	factor := 0.1 + (score-minScore)/(100.0-minScore)*0.9
	rateFactor = math.Max(0.05, math.Min(1, factor)) // Ensure the factor is between 0.05 and 1.

	g.Log().Debugf(ctx, "GetSendingRateFactorForProvider: IP %s, Group %s (Prov Score %d) -> Rate Factor: %.3f", senderIp, mailProviderGroup, providerStatus.Score, rateFactor)
	return rateFactor, nil
}

// GetAdjustedSendingLimitsForProvider calculates the final sending limits for a specific IP and mail provider group.
// It combines the daily base sending volume of the IP warmup with the reputation health factor of the specific provider.
func (s *SenderIpMailProviderService) GetAdjustedSendingLimitsForProvider(ctx context.Context, senderIp string, mailProviderGroup string) (adjustedDailyLimit int, adjustedHourlyLimit int, err error) {
	// 1. Get the base sending volume from the daily increment service.
	baseDailyLimit, baseHourlyLimit, err := s.dailyUpstairsService.GetSendingLimits(ctx, senderIp)
	if err != nil {
		g.Log().Errorf(ctx, "GetAdjustedSendingLimitsForProvider: Failed to get base limits for IP %s: %v", senderIp, err)
		return 0, 0, err
	}

	// 2. Get the reputation health factor for the specific mail provider group.
	rateFactor, err := s.GetSendingRateFactorForProvider(ctx, senderIp, mailProviderGroup)
	if err != nil {
		g.Log().Errorf(ctx, "GetAdjustedSendingLimitsForProvider: Failed to get rate factor for IP %s, Group %s: %v", senderIp, mailProviderGroup, err)
		// Even if fetching the factor fails, a very low value based on the base volume can be returned instead of completely blocking.
		return int(float64(baseDailyLimit) * 0.05), int(float64(baseHourlyLimit) * 0.05), err
	}

	// 3. Apply the factor to the base volume to calculate the final adjusted limits.
	// Use Ceil to ensure that even a small factor multiplied by the base volume results in at least 1 (if the result is greater than 0).
	adjustedDailyLimit = int(math.Ceil(float64(baseDailyLimit) * rateFactor))
	adjustedHourlyLimit = int(math.Ceil(float64(baseHourlyLimit) * rateFactor))

	g.Log().Infof(ctx, "Adjusted Limits for IP %s, Group %s: Daily=%d, Hourly=%d (Base: D=%d, H=%d; Factor: %.2f)", senderIp, mailProviderGroup, adjustedDailyLimit, adjustedHourlyLimit, baseDailyLimit, baseHourlyLimit, rateFactor)

	return adjustedDailyLimit, adjustedHourlyLimit, nil
}

// PeriodicTaskForProviders is a background periodic task that updates the scores of all IPs for each provider group.
func (s *SenderIpMailProviderService) PeriodicTaskForProviders(ctx context.Context) {
	g.Log().Info(ctx, "Sender IP Mail Provider Periodic Task: Starting...")
	var allIpWarmupRecords []*entity.SenderIpWarmup
	// Get all IPs that are warming up or have been warmed up.
	err := g.DB().Model("bm_sender_ip_warmup").Ctx(ctx).Scan(&allIpWarmupRecords)
	if err != nil {
		g.Log().Errorf(ctx, "PeriodicTaskForProviders: Failed to get all IPs from bm_sender_ip_warmup: %v", err)
		return
	}
	if len(allIpWarmupRecords) == 0 {
		g.Log().Info(ctx, "PeriodicTaskForProviders: No sender IPs found in warmup table.")
		return
	}

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

	for _, ipRecord := range allIpWarmupRecords {
		senderIp := ipRecord.SenderIp
		for _, groupName := range providerGroups {
			providerStatus, err := s.InitializeOrGetProviderStatus(ctx, senderIp, groupName)
			if err != nil {
				// The error has been logged in InitializeOrGetProviderStatus, continue to the next one here.
				continue
			}

			// Check if it's time for evaluation.
			if gtime.Now().After(gtime.New(providerStatus.LastEvaluateTime).Add(evaluationIntervalProvider)) {
				g.Log().Debugf(ctx, "PeriodicTaskForProviders: Evaluating score for IP %s, Group %s", senderIp, groupName)
				_, evalErr := s.EvaluateScoreForProvider(ctx, senderIp, groupName)
				if evalErr != nil {
					g.Log().Errorf(ctx, "PeriodicTaskForProviders: Failed to evaluate score for IP %s, Group %s: %v", senderIp, groupName, evalErr)
				}
			} else {
				g.Log().Debugf(ctx, "PeriodicTaskForProviders: Skipping score evaluation for IP %s, Group %s (Last eval: %s)", senderIp, groupName, gtime.New(providerStatus.LastEvaluateTime))
			}
		}
	}
	g.Log().Info(ctx, "Sender IP Mail Provider Periodic Task: Finished.")
}
