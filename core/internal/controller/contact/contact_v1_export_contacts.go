// core/internal/controller/contact/contact_v1_export_contacts.go
package contact

import (
	"archive/zip"
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
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (c *ControllerV1) ExportContacts(ctx context.Context, req *v1.ExportContactsReq) (res *v1.ExportContactsRes, err error) {
	res = &v1.ExportContactsRes{}

	// Set default value for including inactive contacts
	if req.IncludeInactive == false {
		// Default to true if not explicitly set in request
		req.IncludeInactive = true
	}

	// Create temporary directory
	tempDir := filepath.Join(gfile.MainPkgPath(), "temp", "exports")
	if err = gfile.Mkdir(tempDir); err != nil {
		//fmt.Printf("Failed to create directory: %v, path: %s\n", err, tempDir)
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to create temp directory")))
		return
	}

	if req.ExportType == 1 { // Merged export
		// Use map to handle duplicate contact status
		contactMap := make(map[string]*entity.Contact)

		// Get contacts from all groups
		for _, groupId := range req.GroupIds {
			contacts, err := contact.GetContactsByGroup(ctx, groupId)
			if err != nil {
				continue
			}
			// Filter contacts based on IncludeInactive
			for _, c := range contacts {
				if existing, ok := contactMap[c.Email]; ok {
					// If there is a status conflict (one subscribed, one unsubscribed), use the unsubscribed status
					if existing.Active != c.Active {
						existing.Active = 0
					}
					// If the status is the same, keep the original status
				} else {
					contactMap[c.Email] = c
				}
			}
		}

		// Convert to slice and apply IncludeInactive filter
		var allContacts []*entity.Contact
		for _, c := range contactMap {
			if c.Active == 1 || req.IncludeInactive {
				allContacts = append(allContacts, c)
			}
		}

		// Generate filename
		timestamp := gtime.Now().Format("YmdHis")
		fileName := fmt.Sprintf("contacts_export_%s.%s", timestamp, req.Format)
		filePath := filepath.Join(tempDir, fileName)
		fmt.Println("File path: ", filePath)

		// Export file based on format
		var fileContent []byte
		fileContent, err = exportContactFile1(allContacts, req.Format)
		if err != nil {
			fmt.Println("Export file error: ", err)
			res.Code = 500
			res.SetError(gerror.New(public.LangCtx(ctx, "Failed to export contacts")))
			return
		}

		// Write file
		if err = gfile.PutBytes(filePath, fileContent); err != nil {
			fmt.Println("Write file error: ", err)
			res.Code = 500
			res.SetError(gerror.New(public.LangCtx(ctx, "Failed to save file")))
			return
		}

		res.Data.FileUrl = filePath

	} else { // Separate export
		var files []ExportFile
		totalContacts := 0

		// Create file for each group
		for _, groupId := range req.GroupIds {
			group, err := contact.GetGroup(ctx, groupId)
			if err != nil {
				continue
			}

			contacts, err := contact.GetContactsByGroup(ctx, groupId)
			if err != nil {
				continue
			}

			// Filter contacts based on IncludeInactive
			var filteredContacts []*entity.Contact
			for _, c := range contacts {
				if c.Active == 1 || req.IncludeInactive {
					filteredContacts = append(filteredContacts, c)
				}
			}

			if len(filteredContacts) == 0 {
				continue
			}

			// Generate filename
			fileName := fmt.Sprintf("%s.%s", group.Name, req.Format)
			filePath := filepath.Join(tempDir, fileName)

			// Export file
			fileContent, err := exportContactFile1(filteredContacts, req.Format)
			if err != nil {
				continue
			}

			// Write file
			if err = gfile.PutBytes(filePath, fileContent); err != nil {
				continue
			}

			files = append(files, ExportFile{
				Path:     filePath,
				Name:     fileName,
				Contacts: len(filteredContacts),
			})
			totalContacts += len(filteredContacts)
		}

		if len(files) == 0 {
			// fmt.Println("No contacts found")
			res.Code = 400
			res.SetError(gerror.New(public.LangCtx(ctx, "No contacts to export")))
			return
		}

		// Create ZIP file
		timestamp := gtime.Now().Format("YmdHis")
		zipFileName := fmt.Sprintf("contacts_export_%s.zip", timestamp)
		zipPath := filepath.Join(tempDir, zipFileName)

		if err = createZipFile(files, zipPath); err != nil {
			fmt.Println("Create ZIP file error: ", err)
			res.Code = 500
			res.SetError(gerror.New(public.LangCtx(ctx, "Failed to create zip file")))
			return
		}

		// Clean up temporary files
		for _, file := range files {
			os.Remove(file.Path)
		}

		res.Data.FileUrl = zipPath
	}

	res.SetSuccess(public.LangCtx(ctx, "Contacts exported successfully"))
	return
}

// ExportFile Export file information
type ExportFile struct {
	Path     string // File path
	Name     string // File name
	Contacts int    // Contact count
}

// createZipFile Create ZIP file
func createZipFile(files []ExportFile, zipPath string) error {
	// Create zip file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %v", err)
	}
	defer zipFile.Close()

	// Create zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Iterate through source files
	for _, file := range files {
		// Open source file
		srcFile, err := os.Open(file.Path)
		if err != nil {
			return fmt.Errorf("failed to open source file %s: %v", file.Path, err)
		}
		defer srcFile.Close()

		// Create zip entry
		zipEntry, err := zipWriter.Create(file.Name)
		if err != nil {
			return fmt.Errorf("failed to create zip entry for %s: %v", file.Name, err)
		}

		// Copy file content to zip
		if _, err := io.Copy(zipEntry, srcFile); err != nil {
			return fmt.Errorf("failed to copy file content to zip for %s: %v", file.Name, err)
		}
	}

	return nil
}

// exportContactFile1
func exportContactFile1(contacts []*entity.Contact, format string) ([]byte, error) {
	var content []byte
	switch format {
	case "csv":
		// Create CSV writer
		var csvData [][]string

		// Add CSV headers
		headers := []string{
			"email",       // Email address
			"attributes",  // Attributes
			"active",      // Active status
			"create_time", // Create time
		}
		csvData = append(csvData, headers)

		// Add data rows
		for _, c := range contacts {
			if c.Attribs == nil {
				c.Attribs = make(map[string]string)
			}
			// Convert attributes to JSON string
			attribsJSON, err := json.Marshal(c.Attribs)
			if err != nil {
				attribsJSON = []byte("{}")
			}

			// Format time
			createdAt := ""
			if c.CreateTime != 0 {
				// Convert timestamp to time string
				t := time.Unix(int64(c.CreateTime), 0)
				createdAt = t.Format("2006-01-02 15:04:05")
			}

			row := []string{
				c.Email,
				string(attribsJSON),
				fmt.Sprintf("%d", c.Active),
				createdAt,
			}
			csvData = append(csvData, row)
		}

		// Write CSV data
		buf := &bytes.Buffer{}
		writer := csv.NewWriter(buf)
		err := writer.WriteAll(csvData)
		if err != nil {
			return nil, err
		}
		content = buf.Bytes()

	case "txt":
		var lines []string
		for _, c := range contacts {
			// Convert attributes to JSON string
			attribsJSON, err := json.Marshal(c.Attribs)
			if err != nil {
				attribsJSON = []byte("{}")
			}

			// Format create time
			createdAt := ""
			if c.CreateTime != 0 {
				// Convert timestamp to time string
				t := time.Unix(int64(c.CreateTime), 0)
				createdAt = t.Format("2006-01-02 15:04:05")
			}

			// Format: email,attributes,active,created_at
			line := fmt.Sprintf("%s,%s,%d,%s",
				c.Email,
				string(attribsJSON),
				c.Active,
				createdAt,
			)
			lines = append(lines, line)
		}
		content = []byte(strings.Join(lines, "\n"))

	case "excel":

		return nil, fmt.Errorf("excel format not implemented")

	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	return content, nil
}
