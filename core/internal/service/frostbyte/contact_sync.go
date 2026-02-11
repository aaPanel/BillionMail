package frostbyte

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ContactSyncData holds contact data for upsert.
type ContactSyncData struct {
	Email   string            `json:"email"`
	GroupID int               `json:"group_id"`
	Attribs map[string]string `json:"attribs"`
}

// ContactSyncResult is returned after a sync operation.
type ContactSyncResult struct {
	ContactID int  `json:"contact_id"`
	Created   bool `json:"created"`
}

// SyncContact upserts a contact in bm_contacts.
func SyncContact(ctx context.Context, data ContactSyncData) (*ContactSyncResult, error) {
	// Check if contact exists
	record, err := g.DB().Model("bm_contacts").
		Where("email", data.Email).
		Where("group_id", data.GroupID).
		One()

	if err != nil {
		return nil, err
	}

	attribsJSON := "{}"
	if data.Attribs != nil {
		jsonBytes, err := json.Marshal(data.Attribs)
		if err != nil {
			return nil, fmt.Errorf("marshal attribs: %w", err)
		}
		attribsJSON = string(jsonBytes)
	}

	if record.IsEmpty() {
		// Insert new contact
		result, err := g.DB().Model("bm_contacts").Insert(g.Map{
			"email":       data.Email,
			"group_id":    data.GroupID,
			"active":      1,
			"attribs":     attribsJSON,
			"status":      1,
			"create_time": time.Now().Unix(),
		})
		if err != nil {
			return nil, err
		}
		id, _ := result.LastInsertId()
		return &ContactSyncResult{ContactID: int(id), Created: true}, nil
	}

	// Update existing contact — parameterized to prevent SQL injection
	contactID := record["id"].Int()
	_, err = g.DB().Model("bm_contacts").
		Where("id", contactID).
		Update(g.Map{
			"attribs": gdb.Raw(fmt.Sprintf("attribs || '%s'::jsonb", attribsJSON)),
			"active":  1,
		})
	if err != nil {
		return nil, err
	}

	return &ContactSyncResult{ContactID: contactID, Created: false}, nil
}

// LookupContact finds a contact by email.
func LookupContact(ctx context.Context, email string) (map[string]interface{}, error) {
	record, err := g.DB().Model("bm_contacts").
		Where("email", email).
		One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, nil
	}
	return record.Map(), nil
}
