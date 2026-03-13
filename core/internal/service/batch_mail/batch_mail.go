package batch_mail

import (
	"billionmail-core/api/batch_mail/v1"
	"billionmail-core/internal/model/entity"
	"billionmail-core/internal/service/maillog_stat"
	"billionmail-core/internal/service/public"
	"billionmail-core/internal/service/warmup"
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/gvalid"
)

var (
	updateTaskStatsRunning int32 // 0: not running, 1: running
)

type CreateTaskArgs struct {
	Addresser   string `json:"addresser"`   // Sender email address
	Subject     string `json:"subject"`     // Email subject
	FullName    string `json:"full_name"`   // FullName of Sender
	TemplateId  int    `json:"template_id"` // Template ID
	IsRecord    int    `json:"is_record"`   // Landing to outbox
	Unsubscribe int    `json:"unsubscribe"` // Enable unsubscribe link
	Threads     int    `json:"threads"`     // Number of threads to use
	TrackOpen   int    `json:"track_open"`  // Track email opens
	TrackClick  int    `json:"track_click"` // Track email clicks
	//Etypes      string `json:"etypes"`      // Email types (e.g., group IDs)
	Remark    string `json:"remark"`     // Task remark
	StartTime int    `json:"start_time"` // Scheduled start time
	Warmup    int    `json:"warmup"`     // Warmup campaign association
	AddType   int    `json:"add_type"`   // Add type (0: normal)
	GroupId   int    `json:"group_id"`   // Groups to unsubscribe from
	TagIds    []int  `json:"tag_ids"`    // Tag IDs for filtering contacts
	TagLogic  string `json:"tag_logic"`  // Tag logic (AND/OR)
}

// ============= task related operations =============

// GetTasksWithPage get task list (pagination)
func GetTasksWithPage(ctx context.Context, page, pageSize int, keyword string, status int) (total int, list []*v1.EmailTask, err error) {
	// default pagination parameters
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	model := g.DB().Model("email_tasks").Safe()

	// add query conditions
	if keyword != "" {
		model = model.WhereLike("task_name", "%"+keyword+"%").
			WhereOrLike("subject", "%"+keyword+"%")
	}
	if status != -1 {
		model = model.Where("task_process", status)
	}

	// get total
	total, err = model.Count()
	if err != nil {
		return 0, nil, err
	}

	// pagination query - include tag_ids field and map it to TagIdsRaw
	list = make([]*v1.EmailTask, 0)
	err = model.Page(page, pageSize).
		Fields("*, tag_ids as TagIdsRaw").
		Order("create_time DESC").
		Scan(&list)

	if err == nil {
		serverIP, _ := public.GetServerIP()

		if serverIP != "" {
			// Filter associated warmup campaigns
			ids := make([]int, 0)
			m := make(map[int]*v1.EmailTask)
			for _, v := range list {
				// Calculate estimated time for warmup
				ids = append(ids, v.Id)
				v.EstimatedTimeWithWarmup = -1 // default -1 means no warmup
				m[v.Id] = v
			}

			var vals []gdb.Value
			vals, _ = g.DB().Ctx(ctx).Model("bm_campaign_warmup").WhereIn("task_id", ids).Array("task_id")

			for _, val := range vals {
				m[val.Int()].EstimatedTimeWithWarmup, _ = warmup.WarmupCampaign().CalculateEstimatedTime(ctx, val.Int64(), serverIP)
			}
		}
	}

	return total, list, err
}

// DeleteTask
func DeleteTask(ctx context.Context, id int) error {
	// delete task before removing task executor
	RemoveTaskExecutor(id)

	_, err := g.DB().Model("email_tasks").
		Where("id", id).
		Delete()
	return err
}

// CreateTask create task
func CreateTask(ctx context.Context, args CreateTaskArgs) (int, error) {

	now := time.Now().Unix()
	// generate task name
	taskName := "task_" + gconv.String(now)

	// Convert tag IDs to JSON string for database storage
	tagIdsJson := ""
	if len(args.TagIds) > 0 {
		tagIdsJson = gconv.String(args.TagIds)
	}

	result, err := g.DB().Model("email_tasks").Insert(g.Map{
		"task_name":       taskName,
		"addresser":       args.Addresser,
		"subject":         args.Subject,
		"full_name":       args.FullName,
		"recipient_count": 0, // initial 0, update later
		"task_process":    0,
		"pause":           0,
		"template_id":     args.TemplateId,
		"is_record":       args.IsRecord,
		"unsubscribe":     args.Unsubscribe,
		"threads":         args.Threads,
		//"etypes":          args.Etypes,
		"track_open":  args.TrackOpen,
		"track_click": args.TrackClick,
		"start_time":  args.StartTime,
		"create_time": now,
		"update_time": now,
		"active":      1,
		"remark":      args.Remark,
		"add_type":    args.AddType,
		"group_id":    args.GroupId,
		"tag_ids":     tagIdsJson,
		"tag_logic":   args.TagLogic,
	})
	if err != nil {
		g.Log().Debug(ctx, "Failed to create campaign:", err.Error())
		return 0, err
	}

	id, err := result.LastInsertId()

	if err == nil && args.Warmup == 1 {
		// link to sender ip warmup
		serverIP, _ := public.GetServerIP()

		if serverIP != "" {
			_, err = warmup.WarmupCampaign().AssociateCampaignWithWarmup(ctx, id, serverIP)

			if err != nil {
				g.Log().Warning(ctx, "Failed to associate campaign with warmup for task ID %d: %v", id, err)
				err = nil
			}
		}
	}

	g.Log().Debugf(ctx, "CreateTask: Task ID %d is created.", id)

	return int(id), err
}

// UpdateRecipientCount update recipient count
func UpdateRecipientCount(ctx context.Context, taskId, count int) error {
	_, err := g.DB().Model("email_tasks").
		Where("id", taskId).
		Data(g.Map{"recipient_count": count}).
		Update()
	return err
}

// ============= recipient related operations =============

// GetSentCount get sent count
func GetSentCount(ctx context.Context, taskId int) (int, error) {
	return g.DB().Model("recipient_info").
		Where("task_id", taskId).
		Where("is_sent", 1).
		Count()
}

// ImportRecipients import recipient information with batch processing
func ImportRecipients(ctx context.Context, taskId int, contacts []*entity.Contact) error {
	if len(contacts) == 0 {
		return nil
	}

	// Set batch size to avoid PostgreSQL limitations
	const batchSize = 1000

	totalBatches := (len(contacts) + batchSize - 1) / batchSize

	totalImported := 0
	now := time.Now().Unix()

	for i := 0; i < totalBatches; i++ {
		// Calculate start and end indices for current batch
		startIdx := i * batchSize
		endIdx := (i + 1) * batchSize
		if endIdx > len(contacts) {
			endIdx = len(contacts)
		}

		// Get current batch
		currentBatch := contacts[startIdx:endIdx]

		// Prepare data for current batch
		values := make([]g.Map, len(currentBatch))
		for j, contact := range currentBatch {
			values[j] = g.Map{
				"task_id":     taskId,
				"recipient":   contact.Email,
				"is_sent":     0,
				"sent_time":   0,
				"message_id":  "",
				"create_time": now,
			}
		}

		// Insert current batch
		result, err := g.DB().Model("recipient_info").InsertIgnore(values)
		if err != nil {
			g.Log().Error(ctx, "Failed to import recipient batch %d/%d for task %d: %v",
				i+1, totalBatches, taskId, err)
			return fmt.Errorf("failed to import recipients (batch %d/%d): %w", i+1, totalBatches, err)
		}

		// Count affected rows
		affected, err := result.RowsAffected()
		if err != nil {
			g.Log().Debugf(ctx, "Could not get affected rows for batch %d/%d: %v", i+1, totalBatches, err)
		} else {
			totalImported += int(affected)
		}
	}

	g.Log().Info(ctx, "Task %d: Total %d recipients imported successfully", taskId, totalImported)
	return nil
}

// ImportRecipientsTx
func ImportRecipientsTx(ctx context.Context, tx gdb.TX, taskId int, contacts []*entity.Contact) error {
	if len(contacts) == 0 {
		return nil
	}
	const batchSize = 1000
	totalBatches := (len(contacts) + batchSize - 1) / batchSize
	now := time.Now().Unix()
	for i := 0; i < totalBatches; i++ {
		startIdx := i * batchSize
		endIdx := (i + 1) * batchSize
		if endIdx > len(contacts) {
			endIdx = len(contacts)
		}
		currentBatch := contacts[startIdx:endIdx]
		values := make([]g.Map, len(currentBatch))
		for j, contact := range currentBatch {
			values[j] = g.Map{
				"task_id":     taskId,
				"recipient":   contact.Email,
				"is_sent":     0,
				"sent_time":   0,
				"message_id":  "",
				"create_time": now,
			}
		}
		result, err := tx.Ctx(ctx).Model("recipient_info").InsertIgnore(values)
		if err != nil {
			return fmt.Errorf("failed to import recipients (batch %d/%d): %w", i+1, totalBatches, err)
		}
		if _, err = result.RowsAffected(); err != nil {
			g.Log().Debugf(ctx, "ImportRecipientsTx: could not get affected rows for batch %d/%d: %v", i+1, totalBatches, err)
		}
	}
	return nil
}

// ============= contact group related operations =============

// GetGroupInfo get group info
func GetGroupInfo(ctx context.Context, groupId int) (*v1.GroupInfo, error) {
	// use struct map instead of gdb.Record
	var result struct {
		Id    int    `json:"id"`
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	err := g.DB().Model("bm_contact_groups cg").
		LeftJoin("bm_contacts c", "cg.id = c.group_id").
		Fields("cg.id, cg.name, COUNT(CASE WHEN c.active = 1 THEN 1 END) as count").
		Where("cg.id", groupId).
		Group("cg.id, cg.name").
		Scan(&result)

	if err != nil {
		return nil, err
	}

	return &v1.GroupInfo{
		Id:    result.Id,
		Name:  result.Name,
		Count: result.Count,
	}, nil
}

// GetActiveContacts get active contacts in group
func GetActiveContacts(ctx context.Context, groupId int) ([]*entity.Contact, error) {
	var contacts []*entity.Contact
	err := g.DB().Model("bm_contacts").
		Where("group_id", groupId).
		Where("active", 1).
		Where("status", 1). // Confirmed email address
		Scan(&contacts)
	return contacts, err
}

type ContactFilter struct {
	GroupId  int
	TagIds   []int
	TagLogic string // "AND" or "OR"
}

func GetFilteredContacts(ctx context.Context, filter ContactFilter) ([]*entity.Contact, error) {
	var contacts []*entity.Contact

	model := g.DB().Model("bm_contacts c").
		Where("c.active", 1).
		Where("c.status", 1). // Confirmed email address
		Fields("c.*")

	if filter.GroupId > 0 {
		model = model.Where("c.group_id", filter.GroupId)
	}

	if len(filter.TagIds) > 0 {
		if filter.TagLogic == "AND" {
			// AND
			for i, tagId := range filter.TagIds {
				alias := g.NewVar("ct" + g.NewVar(i).String()).String()
				model = model.InnerJoin(
					"bm_contact_tags "+alias,
					"c.id = "+alias+".contact_id AND "+alias+".tag_id = "+g.NewVar(tagId).String(),
				)
			}
		} else if filter.TagLogic == "OR" {
			// OR
			var inValues []string
			for _, tagId := range filter.TagIds {

				inValues = append(inValues, strconv.Itoa(tagId))
			}

			subQuery := fmt.Sprintf(
				"(SELECT DISTINCT contact_id FROM bm_contact_tags WHERE tag_id IN (%s)) ct",
				strings.Join(inValues, ","),
			)

			model = model.InnerJoin(
				subQuery,
				"c.id = ct.contact_id",
			)
		} else if filter.TagLogic == "NOT" {
			// NOT: contact must not have any of the selected tags
			subQuery := g.DB().Model("bm_contact_tags").
				Fields("DISTINCT contact_id").
				WhereIn("tag_id", filter.TagIds)
			model = model.WhereNotIn("c.id", subQuery)
		}
	}

	err := model.Scan(&contacts)
	if err != nil {
		g.Log().Error(ctx, "Failed to get filtered contacts: %v", err)
		return nil, err
	}

	return contacts, nil
}

// ============= business logic combination =============

func CreateTaskWithRecipients(ctx context.Context, req *v1.CreateTaskReq, addType int) (int, error) {
	var taskId int
	var err error
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {

		now := time.Now().Unix()
		taskName := fmt.Sprintf("task_%d", now)
		var tagIdsJson string
		if len(req.TagIds) > 0 {
			tagIdsJson = gconv.String(req.TagIds)
		}

		res, e := tx.Ctx(ctx).Model("email_tasks").Insert(g.Map{
			"task_name":       taskName,
			"addresser":       req.Addresser,
			"subject":         req.Subject,
			"full_name":       req.FullName,
			"recipient_count": 0,
			"task_process":    0,
			"pause":           0,
			"template_id":     req.TemplateId,
			"is_record":       req.IsRecord,
			"unsubscribe":     req.Unsubscribe,
			"threads":         req.Threads,
			"track_open":      req.TrackOpen,
			"track_click":     req.TrackClick,
			"rotate_senders":  req.RotateSenders,
			"start_time":      req.StartTime,
			"create_time":     now,
			"update_time":     now,
			"active":          1,
			"remark":          req.Remark,
			"add_type":        addType,
			"group_id":        req.GroupId,
			"tag_ids":         tagIdsJson,
			"tag_logic":       req.TagLogic,
		})
		if e != nil {
			return gerror.New(public.LangCtx(ctx, "Failed to create task {}", e.Error()))
		}
		id64, _ := res.LastInsertId()
		taskId = int(id64)

		var abnormalRecipients []struct {
			Recipient string `json:"recipient"`
			Count     int    `json:"count"`
		}
		if e = tx.Model("abnormal_recipient").Where("count >= ?", 3).Fields("recipient, count").Scan(&abnormalRecipients); e != nil {
			g.Log().Debugf(ctx, "Failed to get the exception recipient list: %v", e)
		}
		abnormalMap := make(map[string]int, len(abnormalRecipients))
		for _, ar := range abnormalRecipients {
			abnormalMap[ar.Recipient] = ar.Count
		}

		filter := ContactFilter{GroupId: req.GroupId, TagIds: req.TagIds, TagLogic: req.TagLogic}
		var contacts []*entity.Contact
		if len(req.TagIds) > 0 {
			contacts, err = GetFilteredContacts(ctx, filter)
			if err != nil {
				return gerror.New(public.LangCtx(ctx, "Failed to get filtered contacts: {}", err.Error()))
			}
		} else {
			if req.GroupId <= 0 {
				return gerror.New(public.LangCtx(ctx, "Group ID is required when not using tag filter"))
			}
			contacts, err = GetActiveContacts(ctx, req.GroupId)
			if err != nil {
				return gerror.New(public.LangCtx(ctx, "Failed to get contacts for group {}: {}", req.GroupId, err.Error()))
			}
		}
		if len(contacts) == 0 {
			if len(req.TagIds) > 0 {
				return gerror.New(public.LangCtx(ctx, "No contacts found matching the tag filter criteria"))
			}
			return gerror.New(public.LangCtx(ctx, "No contacts found in group {}", req.GroupId))
		}

		filteredContacts := make([]*entity.Contact, 0, len(contacts))
		for _, contact := range contacts {
			if _, exists := abnormalMap[contact.Email]; exists {
				continue
			}
			if e := gvalid.New().Rules("email").Data(contact.Email).Run(ctx); e != nil {
				g.Log().Debug(ctx, "Skip erroneous email addresses:", contact.Email, " Error:", e)
				continue
			}
			filteredContacts = append(filteredContacts, contact)
		}
		if len(filteredContacts) == 0 {
			return gerror.New(public.LangCtx(ctx, "No valid contacts found after filtering"))
		}

		if err = ImportRecipientsTx(ctx, tx, taskId, filteredContacts); err != nil {
			return gerror.New(public.LangCtx(ctx, "Failed to import recipients: {}", err.Error()))
		}

		if actualCount, e2 := GetActualRecipientCount(ctx, taskId); e2 == nil && actualCount > 0 {
			if _, e2 = tx.Ctx(ctx).Model("email_tasks").Where("id", taskId).Data(g.Map{"recipient_count": actualCount}).Update(); e2 != nil {
				return gerror.New(public.LangCtx(ctx, "Failed to update recipient count for task {}: {}", taskId, e2.Error()))
			}
		}

		if req.Warmup == 1 {
			if serverIP, _ := public.GetServerIP(); serverIP != "" {
				if _, e2 := warmup.WarmupCampaign().AssociateCampaignWithWarmup(ctx, int64(taskId), serverIP); e2 != nil {
					g.Log().Warning(ctx, "Failed to associate campaign with warmup for task ID %d: %v", taskId, e2)
				}
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return taskId, nil
}

// GetTaskInfo get task info
func GetTaskInfo(ctx context.Context, taskId int) (*entity.EmailTask, error) {
	var task entity.EmailTask
	err := g.DB().Model("email_tasks").
		Where("id", taskId).
		Fields("*, tag_ids as TagIdsRaw").
		Scan(&task)

	if err != nil {
		return nil, fmt.Errorf("failed to get task info: %w", err)
	}

	return &task, nil
}

// UpdateTaskPauseStatus update task pause status
func UpdateTaskPauseStatus(ctx context.Context, taskId int, isPaused bool) error {
	pauseValue := 0
	processValue := 1

	if isPaused {
		pauseValue = 1
		processValue = 3
	}

	_, err := g.DB().Model("email_tasks").
		Where("id", taskId).
		Data(g.Map{
			"pause":        pauseValue,
			"task_process": processValue,
		}).
		Update()

	if err != nil {
		return fmt.Errorf("failed to update task pause status: %w", err)
	}

	g.Log().Info(ctx, "Updated task %d pause status: isPaused=%v, task_process=%d", taskId, isPaused, processValue)
	return nil
}

// UpdateTaskProcessStatus update task process status
func UpdateTaskProcessStatus(ctx context.Context, taskId int, status int) error {
	_, err := g.DB().Model("email_tasks").
		Where("id", taskId).
		Data(g.Map{"task_process": status}).
		Update()

	if err != nil {
		return fmt.Errorf("failed to update task process status: %w", err)
	}

	return nil
}

// GetTaskSendingStats get task sending stats (success count and failed count)
func GetTaskSendingStats(ctx context.Context, taskID int) (int, int, error) {
	if taskID <= 0 {
		return 0, 0, nil
	}

	// query success count
	successQuery := g.DB().Model("mailstat_send_mails sm")
	successQuery = successQuery.LeftJoin("mailstat_message_ids mid", "sm.postfix_message_id=mid.postfix_message_id")
	successQuery = successQuery.LeftJoin("recipient_info ri", "mid.message_id=ri.message_id")
	successQuery = successQuery.Where("ri.task_id", taskID)
	successQuery = successQuery.Where("sm.status", "sent")
	successQuery = successQuery.Where("sm.dsn LIKE '2.%'")
	successCount, err := successQuery.Count()
	if err != nil {
		// fmt.Println("query success count failed:", err)
		return 0, 0, err
	}

	// query failed count bounced
	failedQuery := g.DB().Model("mailstat_send_mails sm")
	failedQuery = failedQuery.LeftJoin("mailstat_message_ids mid", "sm.postfix_message_id=mid.postfix_message_id")
	failedQuery = failedQuery.LeftJoin("recipient_info ri", "mid.message_id=ri.message_id")
	failedQuery = failedQuery.Where("ri.task_id", taskID)
	failedQuery = failedQuery.Where("sm.status = 'bounced'")
	//failedQuery = failedQuery.Where("sm.status = 'deferred'")
	failedCount, err := failedQuery.Count()
	if err != nil {
		// fmt.Println("query failed count failed:", err)
		return successCount, 0, err
	}
	// fmt.Println("success count:", successCount, "failed count:", failedCount)
	return successCount, failedCount, nil
}

func GetActualRecipientCount(ctx context.Context, taskId int) (int, error) {

	count, err := g.DB().Model("recipient_info").
		Where("task_id", taskId).
		Count()

	if err != nil {
		return 0, fmt.Errorf("failed to get recipient count: %w", err)
	}

	return count, nil
}

// Regularly update the linked query data of marketing tasks
func UpdateTaskJoinMailstat(ctx context.Context) {

	if !atomic.CompareAndSwapInt32(&updateTaskStatsRunning, 0, 1) {
		g.Log().Info(ctx, "UpdateTaskJoinMailstat is already running, skipping this run.")
		return
	}
	defer atomic.StoreInt32(&updateTaskStatsRunning, 0)

	// Search for tasks that need to be updated
	// 1. Tasks that are currently in progress or have been paused (task_process = 1 or 3)
	// 2. Tasks that have been completed but have sent fewer messages than the number of recipients (task_process = 2 and sends_count < recipient_count)
	// 3. Tasks that have never been updated (stats_update_time = 0)
	var tasksToUpdate []*entity.EmailTask
	err := g.DB().Model("email_tasks").
		Where("task_process IN (?)", []int{1, 3}).
		WhereOr("task_process = 2 AND stats_update_time > ?", time.Now().Unix()-36*3600).
		WhereOr("task_process = 2 AND sends_count < recipient_count*0.99").
		WhereOr("stats_update_time", 0).
		Order("id DESC").
		Scan(&tasksToUpdate)
	if err != nil {
		g.Log().Errorf(ctx, "UpdateTaskJoinMailstat: failed to get tasks to update: %v", err)
		return
	}

	if len(tasksToUpdate) == 0 {
		g.Log().Debug(ctx, "UpdateTaskJoinMailstat: no tasks to update.")
		return
	}

	g.Log().Infof(ctx, "UpdateTaskJoinMailstat: found %d tasks to update.", len(tasksToUpdate))

	for _, task := range tasksToUpdate {
		//g.Log().Warningf(ctx, "start %d: %s", task.Id, task.TaskName)

		stats := getSingleTaskStats(ctx, int64(task.Id))
		if stats == nil {
			//g.Log().Warningf(ctx, "UpdateTaskJoinMailstat: could not get stats for task %d, skipping.", task.Id)
			continue
		}

		_, err := g.DB().Model("email_tasks").
			Where("id", task.Id).
			Data(g.Map{
				"sends_count":       stats["sends"],
				"delivered_count":   stats["delivered"],
				"bounced_count":     stats["bounced"],
				"deferred_count":    stats["deferred"],
				"stats_update_time": time.Now().Unix(),
			}).
			Update()

		if err != nil {
			g.Log().Errorf(ctx, "UpdateTaskJoinMailstat: failed to update stats for task %d: %v", task.Id, err)

		} else {
			g.Log().Debugf(ctx, "UpdateTaskJoinMailstat: successfully updated stats for task %d.", task.Id)
		}
	}

	return
}

// getSingleTaskStats Obtain the statistical data for a single task
func getSingleTaskStats(ctx context.Context, taskId int64) map[string]interface{} {

	overview := maillog_stat.NewOverview()
	stats := overview.OverviewDashboard(taskId, "", 0, 0)

	if stats == nil {
		g.Log().Errorf(ctx, "getSingleTaskStats: OverviewDashboard returned nil for task %d", taskId)
		return nil
	}

	requiredKeys := []string{"sends", "delivered", "bounced", "deferred"}
	for _, key := range requiredKeys {
		if _, ok := stats[key]; !ok {
			//g.Log().Warningf(ctx, "getSingleTaskStats: 任务 %d is missing key '%s'", taskId, key)

			stats[key] = 0
		}
	}

	return stats
}

// GetTagsByIds get tags by IDs
func GetTagsByIds(ctx context.Context, tagIds []int) ([]v1.TagInfo, error) {
	if len(tagIds) == 0 {
		return []v1.TagInfo{}, nil
	}

	var tags []v1.TagInfo
	err := g.DB().Model("bm_tags").
		WhereIn("id", tagIds).
		Fields("id, name, create_time").
		Scan(&tags)

	if err != nil {
		return nil, err
	}

	return tags, nil
}

// GetTaskTagIds get tag IDs from task tag_ids field (JSON array)
func GetTaskTagIds(tagIdsStr string) []int {
	if tagIdsStr == "" {
		return []int{}
	}

	var tagIds []int
	if err := gconv.Scan(tagIdsStr, &tagIds); err != nil {
		g.Log().Warningf(context.Background(), "Failed to parse tag_ids: %s, error: %v", tagIdsStr, err)
		return []int{}
	}

	return tagIds
}

// GetGroupsByIds get groups by IDs
func GetGroupsByIds(ctx context.Context, groupIds []int) (map[int]string, error) {
	if len(groupIds) == 0 {
		return make(map[int]string), nil
	}

	var groups []struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}

	err := g.DB().Model("bm_contact_groups").
		WhereIn("id", groupIds).
		Fields("id, name").
		Scan(&groups)

	if err != nil {
		return nil, err
	}

	// Convert to map for quick lookup
	groupMap := make(map[int]string)
	for _, group := range groups {
		groupMap[group.Id] = group.Name
	}

	return groupMap, nil
}
