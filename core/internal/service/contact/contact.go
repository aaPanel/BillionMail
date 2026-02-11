package contact

import (
	v1 "billionmail-core/api/contact/v1"
	"billionmail-core/internal/model/entity"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/gogf/gf/crypto/gmd5"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

func CreateGroup(ctx context.Context, name, description string, double_optin int) (int, error) {
	now := time.Now().Unix()
	token := GfMd5Short()
	data := g.Map{
		"name":         name,
		"description":  description,
		"create_time":  int(now),
		"update_time":  int(now),
		"token":        token,
		"double_optin": double_optin,
	}
	lastInsertId, err := g.DB().Model("bm_contact_groups").Ctx(ctx).Data(data).InsertAndGetId()
	return int(lastInsertId), err
}

func GetGroup(ctx context.Context, id int) (*entity.ContactGroup, error) {
	var group entity.ContactGroup
	err := g.DB().Model("bm_contact_groups").Ctx(ctx).Where("id", id).Scan(&group)
	return &group, err
}

func GetAllGroups(ctx context.Context, keyword string) ([]*v1.ContactGroup, error) {
	model := g.DB().Model("bm_contact_groups").
		Fields("id, name, description, create_time, update_time").
		Order("create_time desc")

	// Add keyword search (group name or description)
	if keyword != "" {
		model = model.WhereLike("name", "%"+keyword+"%").
			WhereOrLike("description", "%"+keyword+"%")
	}

	var groups []*v1.ContactGroup
	err := model.Scan(&groups)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return groups, nil
}

func UpdateGroup(ctx context.Context, id int, data g.Map) error {
	data["update_time"] = time.Now().Unix()
	_, err := g.DB().Model("bm_contact_groups").Ctx(ctx).Data(data).Where("id", id).Update()
	return err
}

func DeleteContactsByGroupId(ctx context.Context, groupId int) error {
	_, err := g.DB().Model("bm_contacts").Ctx(ctx).Where("group_id", groupId).Delete()
	return err
}

func DeleteGroup(ctx context.Context, id int) error {
	_, err := g.DB().Model("bm_contact_groups").Ctx(ctx).Where("id", id).Delete()
	return err
}

func BatchCreateContactsIgnoreDuplicate(ctx context.Context, contacts []*entity.Contact) (int, error) {
	if len(contacts) == 0 {
		return 0, nil
	}

	// Set batch size to avoid PostgreSQL limitations
	const batchSize = 1000

	// Calculate total batches
	totalBatches := (len(contacts) + batchSize - 1) / batchSize
	g.Log().Debug(ctx, "Batch creating %d contacts in %d batches (batch size: %d)",
		len(contacts), totalBatches, batchSize)

	// Process in batches
	totalAffected := 0
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
		var data []g.Map
		for _, contact := range currentBatch {
			data = append(data, g.Map{
				"email":       contact.Email,
				"group_id":    contact.GroupId,
				"active":      contact.Active,
				"task_id":     contact.TaskId,
				"create_time": int(now),
				"attribs":     contact.Attribs,
				"status":      contact.Status,
			})
		}

		// Insert current batch
		result, err := g.DB().Model("bm_contacts").Ctx(ctx).Data(data).InsertIgnore()
		if err != nil {
			g.Log().Error(ctx, "Failed to insert batch %d/%d: %v", i+1, totalBatches, err)
			return totalAffected, err
		}

		// Count affected rows
		affected, err := result.RowsAffected()
		if err != nil {
			g.Log().Debugf(ctx, "Could not get affected rows for batch %d/%d: %v", i+1, totalBatches, err)
		} else {
			totalAffected += int(affected)
			g.Log().Debug(ctx, "Batch %d/%d inserted %d records successfully", i+1, totalBatches, affected)
		}
	}

	g.Log().Info(ctx, "Total %d contacts inserted successfully", totalAffected)
	return totalAffected, nil
}

// BatchCreateContacts creates contacts in batch
func BatchCreateContacts(ctx context.Context, contacts []*entity.Contact) error {
	_, err := BatchCreateContactsIgnoreDuplicate(ctx, contacts)
	return err
}

func GetContactsByGroup(ctx context.Context, groupId int) ([]*entity.Contact, error) {
	var contacts []*entity.Contact
	err := g.DB().Model("bm_contacts").Ctx(ctx).Where("group_id", groupId).Scan(&contacts)
	return contacts, err
}

func CountContactsByGroup(ctx context.Context, groupId int, active int) (int, error) {
	model := g.DB().Model("bm_contacts").Ctx(ctx).Where("group_id", groupId)
	if active >= 0 {
		model = model.Where("active", active)
	}
	return model.Count()
}

func CheckGroupNameExists(ctx context.Context, name string) (bool, error) {
	count, err := g.DB().Model("bm_contact_groups").Ctx(ctx).Where("name", name).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func ContactsGroupWithPage(ctx context.Context, page, pageSize int, keyword string) ([]*entity.ContactGroup, int, error) {

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	keyword = strings.TrimSpace(keyword)

	model := g.DB().Model("bm_contact_groups").Ctx(ctx)

	if keyword != "" {
		model = model.WhereLike("name", "%"+keyword+"%").
			WhereOrLike("description", "%"+keyword+"%")
	}

	total, err := model.Count()
	if err != nil {
		return nil, 0, err
	}

	// limit
	var groups []*entity.ContactGroup
	err = model.Page(page, pageSize).
		OrderDesc("create_time").
		Scan(&groups)

	return groups, total, err
}

func GetContactsWithPage(ctx context.Context, page, pageSize int, groupId int, keyword string, status int) (total int, list []*entity.Contact, err error) {
	model := g.DB().Model("bm_contacts").Safe()

	// Add query conditions
	if groupId > 0 {
		model = model.Where("group_id", groupId)
	}

	if len(keyword) > 0 {
		keywordPattern := "%" + keyword + "%"
		model = model.WhereLike("email", keywordPattern)
	}

	if status != -1 {
		model = model.Where("active", status)
	}

	// Get total with distinct email
	value, err := model.Fields("COUNT(DISTINCT email) as count").Value()
	if err != nil {
		return 0, nil, err
	}
	total = value.Int()

	// If no data, return directly
	if total == 0 {
		return 0, make([]*entity.Contact, 0), nil
	}

	// Construct the query conditions
	conditions := g.Map{}
	if groupId > 0 {
		conditions["group_id"] = groupId
	}
	if status != -1 {
		conditions["active"] = status
	}

	subQuery := g.DB().Model("bm_contacts").
		Fields("DISTINCT ON (email) id,attribs, status")

	if len(conditions) > 0 {
		subQuery = subQuery.Where(conditions)
	}
	if len(keyword) > 0 {
		keywordPattern := "%" + keyword + "%"
		subQuery = subQuery.WhereLike("email", keywordPattern)
	}

	subQuery = subQuery.Order("email, create_time DESC")

	list = make([]*entity.Contact, 0)
	//Where("c.id IN (?)", subQuery).
	err = g.DB().Model("bm_contacts c").
		Where("c.id IN (SELECT id FROM (?))", subQuery).
		Order("c.create_time DESC").
		Page(page, pageSize).
		Scan(&list)
	if err != nil {
		return 0, nil, err
	}

	return total, list, nil
}

// Gets the group information to which the contact belongs
func GetContactGroupsInfo(ctx context.Context, email string, status int) ([]*entity.ContactGroup, error) {
	var groups []*entity.ContactGroup

	err := g.DB().Model("bm_contacts c").
		LeftJoin("bm_contact_groups g", "c.group_id = g.id").
		Fields("DISTINCT g.id, g.name, g.description, g.create_time, g.update_time").
		Where("c.email", email).
		Where("c.active", status).
		Scan(&groups)

	if err != nil {
		return nil, err
	}

	return groups, nil
}

type ContactTrend struct {
	Month            string `json:"month"`
	SubscribeCount   int    `json:"subscribe_count"`
	UnsubscribeCount int    `json:"unsubscribe_count"`
}

func GetContactsTrend(ctx context.Context, startTime, endTime time.Time, groupId int) ([]*ContactTrend, error) {

	var trends []*ContactTrend

	db := g.DB().Model("bm_contacts").
		Fields(
			"to_char(to_timestamp(create_time), 'YYYY-MM') as month",
			"SUM(CASE WHEN active = 1 THEN 1 ELSE 0 END) as subscribe_count",
			"SUM(CASE WHEN active = 0 THEN 1 ELSE 0 END) as unsubscribe_count",
		).
		Where("create_time BETWEEN ? AND ?", startTime.Unix(), endTime.Unix()).
		Group("month").
		Order("month ASC")

	if groupId != 0 {
		db = db.Where("group_id = ?", groupId)
	}

	if err := db.Scan(&trends); err != nil {
		return nil, fmt.Errorf("failed to get contacts trend: %w", err)
	}

	if trends == nil {
		return make([]*ContactTrend, 0), nil
	}

	return trends, nil
}

func UpdateContactsGroups(ctx context.Context, emails []string, status int, newGroupIds []int) (int, error) {

	var existingContacts []struct {
		Email   string            `json:"email"`
		Attribs map[string]string `json:"attribs"`
		Status  int               `json:"status"`
	}

	err := g.DB().Model("bm_contacts").
		Where("email IN(?)", emails).
		Fields("email, attribs, status").
		Group("email, attribs, status").
		Scan(&existingContacts)
	if err != nil {
		return 0, err
	}

	contactInfoMap := make(map[string]struct {
		Attribs map[string]string
		Status  int
	})
	for _, contact := range existingContacts {
		contactInfoMap[contact.Email] = struct {
			Attribs map[string]string
			Status  int
		}{
			Attribs: contact.Attribs,
			Status:  contact.Status,
		}
	}

	// 1. delete old group relations
	_, err = g.DB().Model("bm_contacts").
		Where("email IN(?)", emails).
		Where("active", status).
		Delete()
	if err != nil {
		return 0, err
	}

	// 2. create new group relations
	now := time.Now().Unix()
	var data []g.Map
	for _, email := range emails {

		contactInfo, exists := contactInfoMap[email]
		var attribs map[string]string
		var contactStatus int

		if exists {
			attribs = contactInfo.Attribs
			contactStatus = contactInfo.Status
		}

		for _, groupId := range newGroupIds {
			data = append(data, g.Map{
				"email":       email,
				"group_id":    groupId,
				"active":      status,
				"create_time": now,
				"attribs":     attribs,
				"status":      contactStatus,
			})
		}
	}

	if len(data) == 0 {
		return 0, nil
	}

	result, err := g.DB().Model("bm_contacts").
		Data(data).
		Insert()
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	return int(affected), err
}

// GetGroupContactCount calculates the total number of active contacts in multiple groups
func GetGroupContactCount(ctx context.Context, groupIds []int) (int, error) {
	if len(groupIds) == 0 {
		return 0, nil
	}

	// Query the total number of unique contacts across the specified groups
	var result struct {
		Count int `json:"count"`
	}

	err := g.DB().Model("bm_contacts").
		Fields("COUNT(DISTINCT email) as count").
		WhereIn("group_id", groupIds).
		Where("active", 1). // Only count active contacts
		Scan(&result)

	if err != nil {
		return 0, err
	}

	return result.Count, nil
}

// BatchCreateContactsWithOverwrite
func BatchCreateContactsWithOverwrite(ctx context.Context, contacts []*entity.Contact) (int, error) {
	if len(contacts) == 0 {
		return 0, nil
	}

	const batchSize = 1000
	totalBatches := (len(contacts) + batchSize - 1) / batchSize

	totalSuccessCount := 0
	now := time.Now().Unix()

	for i := 0; i < totalBatches; i++ {
		startIdx := i * batchSize
		endIdx := (i + 1) * batchSize
		if endIdx > len(contacts) {
			endIdx = len(contacts)
		}

		currentBatch := contacts[startIdx:endIdx]

		groupContacts := make(map[int][]*entity.Contact)
		for _, contact := range currentBatch {
			groupContacts[contact.GroupId] = append(groupContacts[contact.GroupId], contact)
		}

		batchSuccessCount := 0
		err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {

			for groupId, contacts := range groupContacts {

				emails := make([]string, len(contacts))
				emailMap := make(map[string]*entity.Contact)
				for j, contact := range contacts {
					emails[j] = contact.Email
					emailMap[contact.Email] = contact
				}

				var existingContacts []*entity.Contact
				err := tx.Model("bm_contacts").
					WhereIn("email", emails).
					Where("group_id", groupId).
					Scan(&existingContacts)
				if err != nil {
					return err
				}

				for _, existing := range existingContacts {
					newContact := emailMap[existing.Email]
					if newContact == nil {
						continue
					}

					updateData := g.Map{
						"active": newContact.Active,
					}

					if len(newContact.Attribs) > 0 {
						attribsBytes, err := json.Marshal(newContact.Attribs)
						if err == nil {
							updateData["attribs"] = string(attribsBytes)
						}
					}

					_, err := tx.Model("bm_contacts").
						Where("id", existing.Id).
						Data(updateData).
						Update()
					if err != nil {
						return err
					}

					batchSuccessCount++
					delete(emailMap, existing.Email)
				}

				// add
				if len(emailMap) > 0 {
					newContacts := make([]map[string]interface{}, 0, len(emailMap))
					for _, contact := range emailMap {
						attribsBytes, err := json.Marshal(contact.Attribs)
						if err != nil {
							continue
						}

						newContacts = append(newContacts, map[string]interface{}{
							"email":       contact.Email,
							"group_id":    groupId,
							"active":      contact.Active,
							"status":      contact.Status,
							"attribs":     string(attribsBytes),
							"create_time": now,
						})
					}

					if len(newContacts) > 0 {
						_, err := tx.Model("bm_contacts").
							Data(newContacts).
							InsertIgnore()
						if err != nil {
							return err
						}
						batchSuccessCount += len(newContacts)
					}
				}
			}
			return nil
		})

		if err != nil {
			return totalSuccessCount, err
		}

		totalSuccessCount += batchSuccessCount

	}

	return totalSuccessCount, nil
}

func GfMd5Short() string {
	str := fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000))
	return gmd5.MustEncryptString(str)[:12]
}

// AddContactToGroup adds a contact to a specified group if it doesn't already exist.
func AddContactToGroup(ctx context.Context, email string, groupId int) (*entity.Contact, error) {
	// Check if the contact already exists in the group
	var existingContact entity.Contact
	err := g.DB().Model("bm_contacts").Ctx(ctx).Where("email", email).Where("group_id", groupId).Scan(&existingContact)
	if err != nil && err != sql.ErrNoRows {
		g.Log().Errorf(ctx, "Failed to check for existing contact %s in group %d: %v", email, groupId, err)
		return nil, err
	}

	// If contact already exists, return it
	if existingContact.Id > 0 {
		return &existingContact, nil
	}

	// If contact does not exist, create a new one
	newContact := &entity.Contact{
		Email:      email,
		GroupId:    groupId,
		Active:     1,
		Status:     1,
		CreateTime: int(time.Now().Unix()),
	}

	result, err := g.DB().Model("bm_contacts").Ctx(ctx).Data(newContact).FieldsEx("id").Insert()
	if err != nil {
		g.Log().Errorf(ctx, "Failed to insert new contact %s into group %d: %v", email, groupId, err)
		return nil, err
	}

	newId, err := result.LastInsertId()
	if err != nil {
		g.Log().Errorf(ctx, "Failed to get last insert ID for contact %s: %v", email, err)
		return nil, err
	}
	newContact.Id = int(newId)

	return newContact, nil
}
