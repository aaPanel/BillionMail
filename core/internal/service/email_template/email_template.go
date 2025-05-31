package email_template

import (
	"billionmail-core/api/email_template/v1"
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"time"
)

// CheckTemplateNameExists
func CheckTemplateNameExists(ctx context.Context, name string) (bool, error) {
	count, err := g.DB().Model("email_templates").
		Ctx(ctx).
		Where("temp_name", name).
		Count()
	return count > 0, err
}

// CreateTemplate
func CreateTemplate(ctx context.Context, name string, addType int, content, render string) (int, error) {
	now := time.Now().Unix()
	result, err := g.DB().Model("email_templates").
		Ctx(ctx).
		Insert(g.Map{
			"temp_name":   name,
			"add_type":    addType,
			"content":     content,
			"render":      render,
			"create_time": now,
			"update_time": now,
		})
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

// DeleteTemplate
func DeleteTemplate(ctx context.Context, id int) error {
	_, err := g.DB().Model("email_templates").
		Ctx(ctx).
		Where("id", id).
		Delete()
	return err
}

// GetTemplate
func GetTemplate(ctx context.Context, id int) (*v1.EmailTemplate, error) {
	var template *v1.EmailTemplate
	err := g.DB().Model("email_templates").
		Ctx(ctx).
		Where("id", id).
		Scan(&template)
	return template, err
}

// UpdateTemplate
func UpdateTemplate(ctx context.Context, id int, name, content, render string) error {
	data := g.Map{
		"update_time": time.Now().Unix(),
	}
	if name != "" {
		data["temp_name"] = name
	}
	if content != "" {
		data["content"] = content
	}
	if render != "" {
		data["render"] = render
	}

	_, err := g.DB().Model("email_templates").
		Ctx(ctx).
		Where("id", id).
		Data(data).
		Update()
	return err
}

// GetTemplatesWithPage
func GetTemplatesWithPage(ctx context.Context, page, pageSize int, keyword string, addType int) (total int, list []*v1.EmailTemplate, err error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	model := g.DB().Model("email_templates").
		Ctx(ctx).
		Safe()

	if keyword != "" {
		model = model.WhereLike("temp_name", "%"+keyword+"%")
	}
	if addType != -1 {
		model = model.Where("add_type", addType)
	}

	// Get total
	total, err = model.Count()
	if err != nil {
		return 0, nil, err
	}
	//selectFields := "id, temp_name, add_type, create_time, update_time"

	// Pagination query
	list = make([]*v1.EmailTemplate, 0)
	err = model.Page(page, pageSize).
		Order("create_time DESC").
		Scan(&list)

	return total, list, err
}

func GetTemplatesByID(ctx context.Context, id int) (*v1.EmailTemplate, error) {
	var template *v1.EmailTemplate
	err := g.DB().Model("email_templates").
		Ctx(ctx).
		Where("id", id).
		Scan(&template)
	return template, err
}

// GetTemplatesAll  id name
func GetTemplatesAll(ctx context.Context) ([]*v1.EmailTemplate, error) {
	var templates []*v1.EmailTemplate
	//selectFields := "id, temp_name"

	err := g.DB().Model("email_templates").
		Ctx(ctx).
		//Fields(selectFields).
		Order("create_time DESC").
		Scan(&templates)
	return templates, err
}
