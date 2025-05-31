// core/internal/controller/contact/contact_v1_import_contacts.go
package contact

import (
	"billionmail-core/api/contact/v1"
	"billionmail-core/internal/model/entity"
	"billionmail-core/internal/service/contact"
	"billionmail-core/internal/service/public"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"io"
	"strings"
)

const (
	ImportTypeFile   = 1
	ImportTypePaste  = 2
	DefaultBatchSize = 1000
)

// isCSVFormat Check if the content is in CSV format
func isCSVFormat(firstLine string) bool {
	// 1. Check if the standard CSV header is included
	lowercaseLine := strings.ToLower(firstLine)
	if strings.Contains(lowercaseLine, "email") {
		return true
	}

	// 2. Try to parse the first line with CSV reader
	reader := csv.NewReader(strings.NewReader(firstLine))
	record, err := reader.Read()
	if err != nil {
		return false
	}

	// If it can be parsed into multiple columns and the first column is not a valid email address, it may be a CSV header
	if len(record) > 1 && !strings.Contains(record[0], "@") {
		return true
	}

	return false
}

// parseJSONAttributes Parse JSON attribute string
func parseJSONAttributes(ctx context.Context, attribsStr string) (map[string]string, error) {
	g.Log().Debug(ctx, "Parsing JSON attributes: %s", attribsStr)

	// Handle double quote escaping
	attribsStr = strings.ReplaceAll(attribsStr, "\"\"", "\"")
	// Remove leading and trailing quotes
	attribsStr = strings.Trim(attribsStr, "\"")

	if attribsStr == "{}" {
		return make(map[string]string), nil
	}

	var attribs map[string]interface{}
	if err := json.Unmarshal([]byte(attribsStr), &attribs); err != nil {
		g.Log().Error(ctx, "Failed to unmarshal JSON attributes: %v", err)
		return nil, err
	}

	// Convert all values to strings
	stringAttribs := make(map[string]string)
	for k, v := range attribs {
		switch val := v.(type) {
		case string:
			stringAttribs[k] = val
			g.Log().Debug(ctx, "Processed attribute %s: %s", k, val)
		case float64:
			if float64(int64(val)) == val {
				stringAttribs[k] = fmt.Sprintf("%d", int64(val))
			} else {
				stringAttribs[k] = fmt.Sprintf("%g", val)
			}
		case bool:
			stringAttribs[k] = fmt.Sprintf("%v", val)
		case nil:
			stringAttribs[k] = ""
		default:
			stringAttribs[k] = fmt.Sprintf("%v", val)
		}
	}

	g.Log().Debug(ctx, "Final parsed attributes: %v", stringAttribs)
	return stringAttribs, nil
}

// parseCSVContent Parse CSV content
func parseCSVContent(ctx context.Context, content string) ([]*entity.Contact, error) {
	var contacts []*entity.Contact
	reader := csv.NewReader(bytes.NewReader([]byte(content)))
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Read headers
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %v", err)
	}

	// Find the index of the required columns
	columnIndexes := make(map[string]int)
	for i, header := range headers {
		header = strings.ToLower(strings.TrimSpace(header))
		columnIndexes[header] = i
	}

	// Check the required columns
	if _, ok := columnIndexes["email"]; !ok {
		return nil, fmt.Errorf("required column 'email' not found")
	}

	// Process data rows
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			g.Log().Warning(ctx, "Error reading CSV record: %v", err)
			continue
		}

		email := strings.TrimSpace(record[columnIndexes["email"]])
		if email == "" {
			continue
		}

		contact := &entity.Contact{
			Email:   email,
			Active:  -1,
			Attribs: make(map[string]string),
		}

		// Process active field
		if activeIdx, ok := columnIndexes["active"]; ok && len(record) > activeIdx {
			activeStr := strings.TrimSpace(record[activeIdx])
			if activeStr == "0" {
				contact.Active = 0
			} else if activeStr == "1" {
				contact.Active = 1
			}
		}

		// Process attributes
		if attribsIdx, ok := columnIndexes["attributes"]; ok && len(record) > attribsIdx {
			attribsStr := strings.TrimSpace(record[attribsIdx])
			if attribsStr != "" {
				attribs, err := parseJSONAttributes(ctx, attribsStr)
				if err != nil {
					g.Log().Warning(ctx, "Failed to parse attributes for email %s: %v", email, err)
				} else {
					contact.Attribs = attribs
				}
			}
		}

		contacts = append(contacts, contact)
	}

	return contacts, nil
}

// parseEmailContent parse email content with attributes
func parseEmailContent(ctx context.Context, content string, importType int) []*entity.Contact {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}

	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return nil
	}

	// If it is a CSV format and the import type is file upload
	if isCSVFormat(lines[0]) && importType == ImportTypeFile {
		contacts, err := parseCSVContent(ctx, content)
		if err != nil {
			g.Log().Error(ctx, "Failed to parse CSV content: %v", err)
			return nil
		}
		return contacts
	}

	// Process normal text format
	var contacts []*entity.Contact
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ",", 2)
		email := strings.TrimSpace(parts[0])
		if email == "" {
			continue
		}

		contact := &entity.Contact{
			Email:   email,
			Active:  -1,
			Attribs: make(map[string]string),
		}

		if len(parts) > 1 {
			attribsStr := strings.TrimSpace(parts[1])
			if attribsStr != "" && attribsStr != "null" {
				attribs, err := parseJSONAttributes(ctx, attribsStr)
				if err == nil {
					contact.Attribs = attribs
				}
			}
		}

		contacts = append(contacts, contact)
	}

	return contacts
}

func (c *ControllerV1) ImportContacts(ctx context.Context, req *v1.ImportContactsReq) (res *v1.ImportContactsRes, err error) {
	res = &v1.ImportContactsRes{}

	// Parameter validation
	if err := c.validateImportRequest(ctx, req); err != nil {
		res.Code = 400
		res.SetError(err)
		return res, nil
	}

	var contactList []*entity.Contact

	// Process content based on import type
	switch req.ImportType {
	case ImportTypeFile:
		if req.FileData == "" {
			res.Code = 400
			res.SetError(gerror.New(public.LangCtx(ctx, "File content cannot be empty")))
			return res, nil
		}

		g.Log().Debug(ctx, "Processing file import")
		contactList = parseEmailContent(ctx, req.FileData, req.ImportType)

	case ImportTypePaste:
		if req.Contacts == "" {
			res.Code = 400
			res.SetError(gerror.New(public.LangCtx(ctx, "Contacts content cannot be empty")))
			return res, nil
		}
		g.Log().Debug(ctx, "Processing paste import")
		contactList = parseEmailContent(ctx, req.Contacts, req.ImportType)

	default:
		res.Code = 400
		res.SetError(gerror.New(public.LangCtx(ctx, "Invalid import type")))
		return res, nil
	}

	if len(contactList) == 0 {
		res.Code = 400
		res.SetError(gerror.New(public.LangCtx(ctx, "No valid contacts found")))
		return res, nil
	}

	g.Log().Debug(ctx, "Found %d valid contacts to import", len(contactList))

	// Import contacts to each group
	successCount := 0
	for _, groupId := range req.GroupIds {
		var groupContacts []*entity.Contact
		for _, c := range contactList {
			contactInfo := *c
			contactInfo.GroupId = groupId

			if contactInfo.Active == -1 {
				contactInfo.Active = req.DefaultActive
			}

			contactInfo.Status = req.Status
			groupContacts = append(groupContacts, &contactInfo)

			// Print the attributes of each contact for debugging
			g.Log().Debug(ctx, "Contact to import - Email: %s, Active: %d, Attributes: %v",
				contactInfo.Email,
				contactInfo.Active,
				contactInfo.Attribs)
		}

		if len(groupContacts) > 0 {
			count, importErr := c.importContactsToGroup(ctx, groupContacts, req.Overwrite)
			if importErr != nil {
				res.Code = 500
				res.SetError(gerror.New(public.LangCtx(ctx, "Failed to import contacts: {}", importErr.Error())))
				return res, importErr
			}
			successCount += count
		}
	}

	res.Data.ImportedCount = successCount
	res.SetSuccess(public.LangCtx(ctx, "Successfully imported {} contacts", successCount))
	return res, nil
}

// validateImportRequest Validate import request parameters
func (c *ControllerV1) validateImportRequest(ctx context.Context, req *v1.ImportContactsReq) error {
	if len(req.GroupIds) == 0 {
		return gerror.New(public.LangCtx(ctx, "Group IDs cannot be empty"))
	}

	// Check if the group exists
	for _, groupId := range req.GroupIds {
		_, err := contact.GetGroup(ctx, groupId)
		if err != nil {
			return gerror.New(public.LangCtx(ctx, "Group {} does not exist", groupId))
		}
	}

	return nil
}

// importContactsToGroup Import contacts to the specified group
func (c *ControllerV1) importContactsToGroup(ctx context.Context, contacts []*entity.Contact, overwrite int) (int, error) {
	if overwrite == 1 {
		return contact.BatchCreateContactsWithOverwrite(ctx, contacts)
	}
	return contact.BatchCreateContactsIgnoreDuplicate(ctx, contacts)
}
