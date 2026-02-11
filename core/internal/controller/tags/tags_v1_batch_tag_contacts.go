package tags

import (
	"billionmail-core/api/tags/v1"
	"billionmail-core/internal/consts"
	"billionmail-core/internal/service/public"
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"strings"
	"time"
)

func (c *ControllerV1) BatchTagContacts(ctx context.Context, req *v1.BatchTagContactsReq) (res *v1.BatchTagContactsRes, err error) {
	res = &v1.BatchTagContactsRes{}

	var groupCount int
	groupCount, err = g.DB().Model("bm_contact_groups").Where("id", req.GroupId).Count()
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to check group")))
		return
	}
	if groupCount == 0 {
		res.SetError(gerror.New(public.LangCtx(ctx, "Group not found")))
		return
	}

	var tags []struct {
		Id      int    `json:"id"`
		Name    string `json:"name"`
		GroupId int    `json:"group_id"`
	}
	err = g.DB().Model("bm_tags").WhereIn("id", req.TagIds).Scan(&tags)
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to check tags")))
		return
	}
	if len(tags) == 0 {
		res.SetError(gerror.New(public.LangCtx(ctx, "No tags found")))
		return
	}
	if len(tags) != len(req.TagIds) {
		res.SetError(gerror.New(public.LangCtx(ctx, "Some tags not found")))
		return
	}

	var tagNames []string
	for _, tag := range tags {
		if tag.GroupId != req.GroupId {
			res.SetError(gerror.New(public.LangCtx(ctx, "Tag '{}' does not belong to this group", tag.Name)))
			return
		}
		tagNames = append(tagNames, tag.Name)
	}

	emailList := parseEmailList(req.Data)
	if len(emailList) == 0 {
		res.SetError(gerror.New(public.LangCtx(ctx, "No valid emails provided")))
		return
	}

	g.Log().Infof(ctx, "BatchTagContacts: Processing %d emails for tags %v (IDs: %v), mode: %d",
		len(emailList), tagNames, req.TagIds, req.MarkInclude)

	var totalProcessedCount, totalSkippedCount, totalErrorCount int

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {

		for _, tagId := range req.TagIds {
			var processedCount, skippedCount, errorCount int
			var err error

			if req.MarkInclude == 1 {
				processedCount, skippedCount, errorCount, err = c.batchAddTags(ctx, tx, req.GroupId, tagId, emailList)
			} else {
				processedCount, skippedCount, errorCount, err = c.batchAddTagsExclude(ctx, tx, req.GroupId, tagId, emailList)
			}

			if err != nil {
				return err
			}

			totalProcessedCount += processedCount
			totalSkippedCount += skippedCount
			totalErrorCount += errorCount
		}
		return nil
	})

	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to process batch tag operation")))
		return
	}

	operation := " Include"
	if req.MarkInclude == 0 {
		operation = " Not included"
	}
	_ = public.WriteLog(ctx, public.LogParams{
		Type: consts.LOGTYPE.Tag,
		Log: "Batch add" + " tags: " + strings.Join(tagNames, ", ") + operation + " for " +
			g.NewVar(totalProcessedCount).String() + " contact-tag pairs successfully",
		Data: map[string]interface{}{
			"tag_ids":      req.TagIds,
			"tag_names":    tagNames,
			"group_id":     req.GroupId,
			"operation":    operation,
			"total_emails": len(emailList),
			"processed":    totalProcessedCount,
			"skipped":      totalSkippedCount,
			"errors":       totalErrorCount,
		},
	})

	successMsg := public.LangCtx(ctx, "Batch tag operation completed")
	//successMsg := public.LangCtx(ctx, "Successfully added {} tags to contacts with {} tag assignments", len(req.TagIds), totalProcessedCount)

	if totalSkippedCount > 0 {
		successMsg += public.LangCtx(ctx, " ( {}  skipped,  {}  errors)", totalSkippedCount, totalErrorCount)
	}

	res.SetSuccess(successMsg)
	return
}

func (c *ControllerV1) batchAddTags(ctx context.Context, tx gdb.TX, groupId, tagId int, emailList []string) (processed, skipped, errors int, err error) {
	const batchSize = 500
	now := time.Now().Unix()

	for i := 0; i < len(emailList); i += batchSize {
		end := i + batchSize
		if end > len(emailList) {
			end = len(emailList)
		}

		currentBatch := emailList[i:end]
		g.Log().Debug(ctx, "Processing batch %d-%d of %d emails", i+1, end, len(emailList))

		var existingContacts []struct {
			Id    int    `json:"id"`
			Email string `json:"email"`
		}
		err = tx.Model("bm_contacts").
			Fields("id, email").
			Where("group_id", groupId).
			WhereIn("email", currentBatch).
			Scan(&existingContacts)
		if err != nil {
			g.Log().Error(ctx, "Failed to query existing contacts: %v", err)
			errors += len(currentBatch)
			continue
		}

		if len(existingContacts) == 0 {
			skipped += len(currentBatch)
			continue
		}

		contactIds := make([]int, len(existingContacts))
		emailToContactId := make(map[string]int)
		for j, contact := range existingContacts {
			contactIds[j] = contact.Id
			emailToContactId[contact.Email] = contact.Id
		}

		var existingTags []struct {
			ContactId int `json:"contact_id"`
		}
		err = tx.Model("bm_contact_tags").
			Fields("contact_id").
			Where("tag_id", tagId).
			WhereIn("contact_id", contactIds).
			Scan(&existingTags)
		if err != nil {
			g.Log().Error(ctx, "Failed to query existing tags: %v", err)
			errors += len(existingContacts)
			continue
		}

		existingTagContactIds := make(map[int]bool)
		for _, tag := range existingTags {
			existingTagContactIds[tag.ContactId] = true
		}

		var insertData []g.Map
		for _, contact := range existingContacts {
			if !existingTagContactIds[contact.Id] {
				insertData = append(insertData, g.Map{
					"contact_id":  contact.Id,
					"tag_id":      tagId,
					"create_time": now,
				})
			}
		}

		if len(insertData) > 0 {

			_, err = tx.Model("bm_contact_tags").InsertIgnore(insertData)
			if err != nil {
				g.Log().Error(ctx, "Failed to insert contact tags: %v", err)
				errors += len(insertData)
				continue
			}
			processed += len(insertData)
		}

		skippedInBatch := len(currentBatch) - len(existingContacts) + len(existingTags)
		skipped += skippedInBatch
	}

	return processed, skipped, errors, nil
}

func (c *ControllerV1) batchAddTagsExclude(ctx context.Context, tx gdb.TX, groupId, tagId int, excludeEmailList []string) (processed, skipped, errors int, err error) {
	const batchSize = 500
	now := time.Now().Unix()

	g.Log().Infof(ctx, "batchAddTagsExclude: Excluding %d emails from tagging", len(excludeEmailList))

	excludeEmailMap := make(map[string]bool)
	for _, email := range excludeEmailList {
		excludeEmailMap[strings.ToLower(email)] = true
	}

	var offset = 0
	for {

		var allContacts []struct {
			Id    int    `json:"id"`
			Email string `json:"email"`
		}
		err = tx.Model("bm_contacts").
			Fields("id, email").
			Where("group_id", groupId).
			Limit(batchSize).
			Offset(offset).
			Scan(&allContacts)
		if err != nil {
			g.Log().Error(ctx, "Failed to query group contacts: %v", err)
			errors += batchSize
			break
		}

		if len(allContacts) == 0 {
			break
		}

		g.Log().Debug(ctx, "Processing batch starting from offset %d, found %d contacts", offset, len(allContacts))

		var targetContacts []struct {
			Id    int    `json:"id"`
			Email string `json:"email"`
		}
		for _, contact := range allContacts {
			if !excludeEmailMap[strings.ToLower(contact.Email)] {
				targetContacts = append(targetContacts, contact)
			} else {
				g.Log().Debug(ctx, "Excluding email: %s", contact.Email)
			}
		}

		if len(targetContacts) == 0 {
			skipped += len(allContacts)
			offset += batchSize
			continue
		}

		targetContactIds := make([]int, len(targetContacts))
		for j, contact := range targetContacts {
			targetContactIds[j] = contact.Id
		}

		var existingTags []struct {
			ContactId int `json:"contact_id"`
		}
		err = tx.Model("bm_contact_tags").
			Fields("contact_id").
			Where("tag_id", tagId).
			WhereIn("contact_id", targetContactIds).
			Scan(&existingTags)
		if err != nil {
			g.Log().Error(ctx, "Failed to query existing tags: %v", err)
			errors += len(targetContacts)
			offset += batchSize
			continue
		}

		existingTagContactIds := make(map[int]bool)
		for _, tag := range existingTags {
			existingTagContactIds[tag.ContactId] = true
		}

		var insertData []g.Map
		for _, contact := range targetContacts {
			if !existingTagContactIds[contact.Id] {
				insertData = append(insertData, g.Map{
					"contact_id":  contact.Id,
					"tag_id":      tagId,
					"create_time": now,
				})
			}
		}

		if len(insertData) > 0 {

			_, err = tx.Model("bm_contact_tags").InsertIgnore(insertData)
			if err != nil {
				g.Log().Error(ctx, "Failed to insert contact tags: %v", err)
				errors += len(insertData)
			} else {
				processed += len(insertData)
				g.Log().Debug(ctx, "Successfully tagged %d contacts in this batch", len(insertData))
			}
		}

		skippedInBatch := len(allContacts) - len(targetContacts) + len(existingTags)
		skipped += skippedInBatch

		if len(allContacts) < batchSize {
			break
		}

		offset += batchSize
	}

	g.Log().Infof(ctx, "batchAddTagsExclude completed: processed=%d, skipped=%d, errors=%d", processed, skipped, errors)
	return processed, skipped, errors, nil
}

func parseEmailList(data string) []string {
	if data == "" {
		return nil
	}

	lines := strings.FieldsFunc(data, func(r rune) bool {
		return r == '\n' || r == '\r' || r == ',' || r == ';'
	})

	var emails []string
	emailMap := make(map[string]bool)

	for _, line := range lines {
		email := strings.ToLower(strings.TrimSpace(line))
		if email != "" && !emailMap[email] {
			if strings.Contains(email, "@") && strings.Contains(email, ".") {
				emails = append(emails, email)
				emailMap[email] = true
			}
		}
	}

	return emails
}
