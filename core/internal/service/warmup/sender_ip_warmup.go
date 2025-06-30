package warmup

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/database/gdb"
	"math"
	"strings"
	"time"

	"billionmail-core/internal/model/entity"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	defaultWarmupPeriodDays     = 45
	minScoreForSending          = 20                 // IPs with a score below this will have their sending paused or sent at a very low rate
	evaluationInterval          = 1 * time.Hour      // The frequency of evaluating scores and progress (example)
	reWarmupCheckInterval       = 24 * time.Hour     // The frequency of checking if re-warmup is needed (example)
	defaultEvaluationWindow     = 7 * 24 * time.Hour // The time window for evaluation data, e.g., the last 7 days
	minSendVolumeForFullScoring = 20                 // The minimum sending volume required for a full score calculation
)

type SenderIpWarmupService struct{}

var insSenderIpWarmupService = SenderIpWarmupService{}

func SenderIpWarmup() *SenderIpWarmupService {
	return &insSenderIpWarmupService
}

// getSenderIdentitiesForIp gets the associated sender identities (e.g., email addresses or domains) for a given sender IP.
func (s *SenderIpWarmupService) getSenderIdentitiesForIp(ctx context.Context, senderIp string) (domains []string, err error) {
	// Get the sender identities associated with this IP.
	var vals []gdb.Value
	vals, err = g.DB().Ctx(ctx).Model("domain").Where("active = 1").Fields("domain").Array("domain")

	if err != nil {
		err = fmt.Errorf("getSenderIdentitiesForIp err: %v", err)
		return
	}

	for _, v := range vals {
		domains = append(domains, v.String())
	}

	g.Log().Debug(ctx, "Domains for IP", senderIp, domains)

	return // return empty, let the scoring logic handle the absence of data
}

// InitializeOrGetWarmupStatus initializes or retrieves the warmup status of an IP.
func (s *SenderIpWarmupService) InitializeOrGetWarmupStatus(ctx context.Context, senderIp string) (record *entity.SenderIpWarmup, err error) {
	err = g.DB().Model("bm_sender_ip_warmup").Ctx(ctx).Where("sender_ip", senderIp).Scan(&record)
	if err != nil {
		g.Log().Errorf(ctx, "Failed to query SenderIpWarmup for IP %s: %v", senderIp, err)
		return
	}
	if record == nil {
		now := gtime.Now()
		record = &entity.SenderIpWarmup{
			SenderIp:         senderIp,
			Period:           defaultWarmupPeriodDays,
			Score:            40, // Initial score for a new IP (e.g., 40)
			LastEvaluateTime: now.Timestamp(),
			Progress:         0,
			ReWarmups:        0,
			BeginTime:        now.Timestamp(), // Warmup start time
			EndTime:          now.AddDate(0, 0, defaultWarmupPeriodDays).Timestamp(),
			CreateTime:       now.Timestamp(),
			UpdateTime:       now.Timestamp(),
		}
		record.Id, err = g.DB().Model("bm_sender_ip_warmup").Ctx(ctx).Data(record).OmitEmptyData().InsertAndGetId()
		if err != nil {
			g.Log().Errorf(ctx, "Failed to insert SenderIpWarmup for IP %s: %v", senderIp, err)
			return
		}
		g.Log().Infof(ctx, "Initialized warmup status for new IP: %s", senderIp)
	}
	return
}

// EvaluateIpScore evaluates and updates the score of an IP.
func (s *SenderIpWarmupService) EvaluateIpScore(ctx context.Context, senderIp string) (score int, err error) {
	var warmupStatus *entity.SenderIpWarmup
	warmupStatus, err = s.InitializeOrGetWarmupStatus(ctx, senderIp)
	if err != nil {
		return
	}

	evaluationWindowStartMillis := gtime.New(warmupStatus.LastEvaluateTime).Add(-defaultEvaluationWindow).TimestampMilli()
	if warmupStatus.CreateTime == warmupStatus.LastEvaluateTime || gtime.New(warmupStatus.LastEvaluateTime).Before(gtime.New(warmupStatus.BeginTime)) { // After the first evaluation or a reset
		evaluationWindowStartMillis = gtime.New(warmupStatus.BeginTime).TimestampMilli()
	}
	if gtime.Now().TimestampMilli()-evaluationWindowStartMillis > defaultEvaluationWindow.Milliseconds()*2 { // Prevent the window from being too large
		evaluationWindowStartMillis = gtime.Now().Add(-defaultEvaluationWindow).TimestampMilli()
	}

	senderIdentities, err := s.getSenderIdentitiesForIp(ctx, senderIp)
	if err != nil {
		g.Log().Errorf(ctx, "EvaluateIpScore: Failed to get sender identities for IP %s: %v", senderIp, err)
		// Even if identities cannot be fetched, try to update the evaluation time to avoid getting stuck
		_, _ = g.DB().Model("bm_sender_ip_warmup").Ctx(ctx).Data(g.Map{
			"last_evaluate_time": gtime.Now().Timestamp(),
		}).Update()
		return warmupStatus.Score, err
	}

	relevantPmids := make([]string, 0)
	if len(senderIdentities) > 0 {
		var vals []gdb.Value
		// Note: This query assumes that mailstat_senders.sender stores email addresses.
		// If it's a domain, the query condition needs to be adjusted to LIKE '%@domain.com'.
		likes := make([]string, len(senderIdentities))
		params := make([]interface{}, len(senderIdentities))
		for i, identity := range senderIdentities {
			// Adjust according to the format of identity. If it's a full email address, use sender = ?; if it's a domain, use sender LIKE '%@domain'.
			if strings.Contains(identity, "@") {
				likes[i] = "sender = ?"
				params[i] = identity
			} else {
				likes[i] = "sender LIKE ?"
				params[i] = "%@" + identity
			}
		}
		vals, err = g.DB().Ctx(ctx).Model("mailstat_senders").
			Fields("DISTINCT postfix_message_id").
			Where(strings.Join(likes, " OR "), params...).
			Where("log_time_millis >= ?", evaluationWindowStartMillis).
			Array()
		if err != nil {
			g.Log().Errorf(ctx, "EvaluateIpScore: Failed to get postfix_message_ids for IP %s (identities: %v): %v", senderIp, senderIdentities, err)
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
	deferredSends := 0 // status='deferred' from mailstat_send_mails
	openedCount := 0   // Needs to be obtained from mailstat_opened
	clickedCount := 0  // Needs to be obtained from mailstat_clicked

	if len(relevantPmids) > 0 {
		// 1. Get sending statistics (mailstat_send_mails)
		type SendMailStatItem struct {
			Status string `json:"status"`
			Dsn    string `json:"dsn"`
			Count  int    `json:"count"`
		}
		var sendMailStats []SendMailStatItem
		err = g.DB().Model("mailstat_send_mails").
			Ctx(ctx).
			WhereIn("postfix_message_id", relevantPmids).
			Where("log_time_millis >= ?", evaluationWindowStartMillis).
			Group("status, dsn").
			Fields("status, dsn, COUNT(*) AS count").
			Scan(&sendMailStats)
		if err != nil {
			g.Log().Errorf(ctx, "EvaluateIpScore: Failed to get send mail stats for IP %s (pmids: %d): %v", senderIp, len(relevantPmids), err)
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
							softBounces += stat.Count // Others are classified as soft bounces
						}
					} else {
						softBounces += stat.Count // Bounces without a DSN are treated as soft bounces
					}
				case "deferred":
					deferredSends += stat.Count
				}
			}
		}
		// Get openedCount and clickedCount (requires user interaction data)
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

	finishedWarmup := false
	currentScore := float64(warmupStatus.Score) // Adjust starting from the current score

	if totalSent == 0 && len(relevantPmids) == 0 && len(senderIdentities) == 0 {
		// If the IP is not configured with any sender identities, do not adjust the score, only update the evaluation time
		g.Log().Infof(ctx, "EvaluateIpScore: IP %s has no configured sender identities. Score remains %d.", senderIp, warmupStatus.Score)
	} else if totalSent < minSendVolumeForFullScoring && totalSent > 0 { // Low sending volume
		if hardBounces > 0 {
			currentScore -= 1.5 // Low volume but with hard bounces, heavy penalty
		} else if softBounces > 0 {
			currentScore -= 0.7 // Low volume with soft bounces
		} else if successfulSends > 0 {
			currentScore += 0.2 // Low volume but with successes
		} else { // Low volume and no successes (could be all deferred or other issues)
			currentScore -= 0.5
		}
	} else if totalSent >= minSendVolumeForFullScoring { // Sufficient sending volume for a more comprehensive scoring
		if totalSent >= 70000 {
			finishedWarmup = true
		}

		successRate := float64(successfulSends) / float64(totalSent)
		hardBounceRate := float64(hardBounces) / float64(totalSent)
		softBounceRate := float64(softBounces) / float64(totalSent)
		deferredRate := float64(deferredSends) / float64(totalSent)

		// Adjust based on success rate
		currentScore += math.Max(0, successRate-0.8) * 30 // 80% success rate is the baseline, fluctuation affects 30 points

		// Adjust based on bounce rate
		currentScore -= hardBounceRate * 10 // Penalty coefficient for hard bounces is 10
		currentScore -= softBounceRate * 5  // Penalty coefficient for soft bounces is 5
		currentScore -= deferredRate * 2    // Penalty coefficient for deferred is 2

		// Adjust based on open and click rates (requires openedCount, clickedCount data)
		if successfulSends > 0 {
			openRate := float64(openedCount) / float64(successfulSends)
			currentScore += openRate * 20 // 10% open rate baseline, affects 20 points (no penalty mechanism for now)
			if openedCount > 0 {
				clickRate := float64(clickedCount) / float64(openedCount)
				currentScore += clickRate * 15 // 2% click rate (based on opens) baseline, affects 15 points (no penalty mechanism for now)
			}
		}
	} else { // totalSent == 0 (but there are relevant pmids or identities, indicating no sending within the window)
		if warmupStatus.Progress > 65 { // High progress but no sending
			if warmupStatus.Score < 50 {
				currentScore -= 0.3 // If the score is already not high and there is continuous no sending, decrease slowly
			} else if warmupStatus.Score > 70 {
				currentScore -= 0.2 // If the score is high but there is continuous no sending, decrease very slowly
			}
		}
	}

	finalScore := int(math.Max(0, math.Min(100, currentScore)))

	curTime := gtime.Now().Timestamp()

	updateData := g.Map{
		"last_evaluate_time": curTime,
		"score":              finalScore,
		"update_time":        curTime,
	}

	if finishedWarmup {
		updateData["period"] = 0 // Set period to 0 to indicate warmup is complete
		updateData["end_time"] = curTime
		updateData["progress"] = 100 // Set progress to 100% if warmup is complete
	}

	_, err = g.DB().
		Model("bm_sender_ip_warmup").
		Ctx(ctx).
		Where("id", warmupStatus.Id).
		Data(updateData).
		Update()

	if err != nil {
		g.Log().Errorf(ctx, "EvaluateIpScore: Failed to update score for IP %s: %v", senderIp, err)
		return warmupStatus.Score, err // Return the old score and the error
	}
	g.Log().Debugf(ctx, "Evaluated score for IP %s: %d (TotalSent: %d, HardBounce: %d, SoftBounce: %d, Deferred: %d)", senderIp, finalScore, totalSent, hardBounces, softBounces, deferredSends)
	return finalScore, nil
}

// UpdateWarmupPeriod dynamically adjusts the total warmup period.
func (s *SenderIpWarmupService) UpdateWarmupPeriod(ctx context.Context, senderIp string) error {
	record, err := s.InitializeOrGetWarmupStatus(ctx, senderIp)
	if err != nil {
		return err
	}

	newPeriod := defaultWarmupPeriodDays
	if record.Score < 40 { // Low score
		newPeriod = defaultWarmupPeriodDays + 15 + (record.ReWarmups * 5) // Extend, and increase based on the number of re-warmups
	} else if record.Score > 80 && record.Progress > 50 { // Good performance and more than halfway through
		newPeriod = record.Period - 10
		if newPeriod < 21 {
			newPeriod = 21
		} // Minimum warmup period of 21 days
	}
	// Ensure the new period is not less than a minimum value, e.g., 7 days
	if newPeriod < 7 {
		newPeriod = 7
	}

	if newPeriod != record.Period {
		newEndTime := gtime.New(record.BeginTime).AddDate(0, 0, newPeriod).Timestamp()
		_, err = g.DB().Model("bm_sender_ip_warmup").Ctx(ctx).Data(g.Map{
			"period":      newPeriod,
			"end_time":    newEndTime,
			"update_time": gtime.Now().Timestamp(),
		}).Where("id", record.Id).Update()
		if err != nil {
			g.Log().Errorf(ctx, "UpdateWarmupPeriod: Failed for IP %s: %v", senderIp, err)
			return err
		}
		g.Log().Infof(ctx, "Updated warmup period for IP %s to %d days. New end time: %s", senderIp, newPeriod, gtime.New(newEndTime).String())
	}
	return nil
}

// UpdateWarmupProgress updates the warmup progress of an IP.
func (s *SenderIpWarmupService) UpdateWarmupProgress(ctx context.Context, senderIp string) error {
	record, err := s.InitializeOrGetWarmupStatus(ctx, senderIp)
	if err != nil {
		return err
	}

	if record.BeginTime == 0 || record.Period <= 0 {
		g.Log().Debugf(ctx, "UpdateWarmupProgress: IP %s has no begin time or invalid period. Progress not updated.", senderIp)
		return nil // Not started or period is invalid
	}

	daysPassed := (gtime.Now().Timestamp() - record.BeginTime) / (24 * 60 * 60)
	if daysPassed < 0 {
		daysPassed = 0
	} // Prevent BeginTime from being in the future

	progress := int((float64(daysPassed) / float64(record.Period)) * 100)

	// Adjust progress based on score: a low score may cause progress to stagnate or slightly regress
	if record.Score < minScoreForSending && progress > record.Progress && record.Progress > 10 { // Extremely low score and had progress before
		progress = record.Progress - 5 // Progress regresses
	} else if record.Score < 40 && progress > record.Progress { // Poor score
		progress = record.Progress + int(math.Max(0, (float64(progress-record.Progress)*0.3))) // Slow progress
	} else if record.Score >= 70 && progress > record.Progress { // Good score
		progress = record.Progress + int(math.Max(0, (float64(progress-record.Progress)*0.5))) // Accelerated progress
	}

	finalProgress := int(math.Max(0, math.Min(100, float64(progress))))

	if finalProgress != record.Progress {
		_, err = g.DB().Model("bm_sender_ip_warmup").Ctx(ctx).Data(g.Map{
			"progress":    finalProgress,
			"update_time": gtime.Now().Timestamp(),
		}).Where("id", record.Id).Update()
		if err != nil {
			g.Log().Errorf(ctx, "UpdateWarmupProgress: Failed for IP %s: %v", senderIp, err)
			return err
		}
		g.Log().Debugf(ctx, "Updated warmup progress for IP %s to %d%% (Days passed: %d, Period: %d, Score: %d)", senderIp, finalProgress, daysPassed, record.Period, record.Score)
	}
	return nil
}

// CheckAndTriggerReWarmup checks if a re-warmup is necessary.
func (s *SenderIpWarmupService) CheckAndTriggerReWarmup(ctx context.Context, senderIp string) error {
	record, err := s.InitializeOrGetWarmupStatus(ctx, senderIp)
	if err != nil {
		return err
	}

	needsReWarmup := false
	// Condition 1: Warmup completed (100% progress), but the current score has dropped significantly (e.g., below 40)
	if record.Progress >= 100 && record.Score < 40 {
		needsReWarmup = true
		g.Log().Infof(ctx, "IP %s (score: %d, progress: %d%%) may need re-warmup due to score drop.", senderIp, record.Score, record.Progress)
	}

	if needsReWarmup {
		now := gtime.Now()
		newPeriod := defaultWarmupPeriodDays + 15 // The re-warmup period can be slightly longer
		_, err = g.DB().Model("bm_sender_ip_warmup").Ctx(ctx).Data(g.Map{
			"progress":           0,
			"re_warmups":         record.ReWarmups + 1,
			"period":             newPeriod,
			"begin_time":         now.Timestamp(),
			"end_time":           now.AddDate(0, 0, newPeriod).Timestamp(),
			"last_evaluate_time": now.Timestamp(),
			"update_time":        now.Timestamp(),
			// "score": 30, // Optional: Reset the score to a lower initial value upon re-warmup
		}).Where("id", record.Id).Update()
		if err != nil {
			g.Log().Errorf(ctx, "CheckAndTriggerReWarmup: Failed to reset IP %s for re-warmup: %v", senderIp, err)
			return err
		}
		g.Log().Infof(ctx, "IP %s has been reset for re-warmup. New period: %d days.", senderIp, newPeriod)
	}
	return nil
}

// GetSendingRateFactor returns a sending rate adjustment factor (0.0 - 1.0) based on the score.
func (s *SenderIpWarmupService) GetSendingRateFactor(ctx context.Context, senderIp string) (factor float64, err error) {
	record, err := s.InitializeOrGetWarmupStatus(ctx, senderIp)
	if err != nil || record == nil {
		g.Log().Warningf(ctx, "GetSendingRateFactor: IP %s status unknown or error (%v), defaulting to very cautious rate.", senderIp, err)
		return 0.05, err // Status unknown, very low rate
	}

	// Warmup not complete, and score is very low, or warmup has just begun
	if record.Progress < 100 {
		//if record.Score < minScoreForSending {
		//	return 0.0, nil // Pause sending
		//}
		// In the early stages of warmup, the rate should be very low and gradually increase with progress and score
		// This is a simplified step-wise increase example
		baseProgressFactor := float64(record.Progress) / 100.0
		scoreFactor := (float64(record.Score) - float64(minScoreForSending)) / (100.0 - float64(minScoreForSending))
		if scoreFactor < 0 {
			scoreFactor = 0.05 // Ensure the score factor is not negative, default to a very cautious rate
		}

		// Example: a smoother warmup rate = (base progress rate * 0.3) + (score performance rate * 0.7)
		// The rate should not exceed the upper limit corresponding to the current progress
		rate := baseProgressFactor*0.3 + scoreFactor*0.7

		if rate < 0.01 && record.Progress < 10 {
			rate = 0.01
		} // Minimum startup rate
		if rate > baseProgressFactor {
			rate = baseProgressFactor
		} // The rate does not exceed the limit allowed by the current progress
		if rate > 1.0 {
			rate = 1.0
		}
		if rate < 0 {
			rate = 0.05 // Ensure the rate is not negative, default to a very cautious rate
		}
		g.Log().Debugf(ctx, "GetSendingRateFactor: IP %s (Progress: %d%%, Score: %d) -> In Warmup, Rate Factor: %.2f", senderIp, record.Progress, record.Score, rate)
		return rate, nil
	}

	// Rate adjustment after warmup is complete
	if record.Score < minScoreForSending {
		g.Log().Debugf(ctx, "GetSendingRateFactor: IP %s (Progress: %d%%, Score: %d) -> Paused (Score too low)", senderIp, record.Progress, record.Score)
		return 0.0, nil // Pause sending
	}

	if record.Score >= 85 {
		factor = 1.0
		return
	} else if record.Score >= 70 {
		factor = 0.8
		return
	} else if record.Score >= 50 {
		factor = 0.5
		return
	} else if record.Score >= minScoreForSending {
		factor = 0.2
		return
	}

	g.Log().Debugf(ctx, "GetSendingRateFactor: IP %s (Progress: %d%%, Score: %d) -> Default Rate Factor: 0.1 (Should be covered by brackets)", senderIp, record.Progress, record.Score)
	return 0.1, nil // Default to a lower rate, should be covered by the conditions above
}

// PeriodicTask is a background periodic task to update the warmup status of all relevant IPs.
func (s *SenderIpWarmupService) PeriodicTask(ctx context.Context) {
	g.Log().Info(ctx, "Sender IP Warmup Periodic Task: Starting...")
	var activeIPs []*entity.SenderIpWarmup
	err := g.DB().Model("bm_sender_ip_warmup").Ctx(ctx).Scan(&activeIPs) // Get all recorded IPs
	if err != nil {
		g.Log().Errorf(ctx, "PeriodicTask: Failed to get active IPs from SenderIpWarmup: %v", err)
		return
	}
	if len(activeIPs) == 0 {
		g.Log().Info(ctx, "PeriodicTask: No sender IPs found in warmup table.")
		return
	}

	for _, ipRecord := range activeIPs {
		senderIp := ipRecord.SenderIp
		g.Log().Debugf(ctx, "PeriodicTask: Processing IP %s", senderIp)

		// 1. Evaluate score (most frequent)
		if gtime.Now().After(gtime.New(ipRecord.LastEvaluateTime).Add(evaluationInterval)) {
			_, err := s.EvaluateIpScore(ctx, senderIp)
			if err != nil {
				g.Log().Errorf(ctx, "PeriodicTask: Failed to evaluate score for IP %s: %v", senderIp, err)
			}
		} else {
			g.Log().Debugf(ctx, "PeriodicTask: Skipping score evaluation for IP %s, last evaluated at %s", senderIp, gtime.New(ipRecord.LastEvaluateTime))
		}

		// Re-fetch the record as the score may have been updated
		currentIpStatus, getErr := s.InitializeOrGetWarmupStatus(ctx, senderIp)
		if getErr != nil {
			g.Log().Errorf(ctx, "PeriodicTask: Failed to get updated status for IP %s after eval: %v", senderIp, getErr)
			continue
		}

		// 2. Update warmup period (can be less frequent)
		// Simple handling: try to update the warmup period and progress after each score evaluation
		err = s.UpdateWarmupPeriod(ctx, senderIp)
		if err != nil {
			g.Log().Errorf(ctx, "PeriodicTask: Failed to update warmup period for IP %s: %v", senderIp, err)
		}

		// 3. Update progress (can be less frequent)
		err = s.UpdateWarmupProgress(ctx, senderIp)
		if err != nil {
			g.Log().Errorf(ctx, "PeriodicTask: Failed to update warmup progress for IP %s: %v", senderIp, err)
		}

		// 4. Check if re-warmup is needed (can be less frequent, e.g., once a day)
		// For simplicity, check every time; the frequency can be adjusted in practice
		// Ensure that the score and progress are up-to-date before checking
		if currentIpStatus.Progress >= 100 || (gtime.Now().Unix()-currentIpStatus.UpdateTime > gconv.Int64(reWarmupCheckInterval.Seconds()/2)) { // Progress is 100% or it has been a while since the last update
			err = s.CheckAndTriggerReWarmup(ctx, senderIp)
			if err != nil {
				g.Log().Errorf(ctx, "PeriodicTask: Failed to check re-warmup for IP %s: %v", senderIp, err)
			}
		}
	}
	g.Log().Info(ctx, "Sender IP Warmup Periodic Task: Finished.")
}
