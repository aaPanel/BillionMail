package rbac

import (
	"billionmail-core/internal/model"
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type roleService struct{}

func newRoleService() *roleService {
	return &roleService{}
}

// GetList gets role list with pagination
func (s *roleService) GetList(ctx context.Context, page, pageSize int, name string, status int) ([]model.Role, int, error) {
	var roles []model.Role
	var total int

	query := g.DB().Model("role")
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status > 0 {
		query = query.Where("status = ?", status)
	}

	// Get total count
	count, err := query.Count()
	if err != nil {
		return nil, 0, err
	}
	total = count

	// Get paginated data
	err = query.Page(page, pageSize).Scan(&roles)
	if err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// GetAll gets all roles
func (s *roleService) GetAll(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	err := g.DB().Model("role").Scan(&roles)
	return roles, err
}

// GetById gets role details by ID
func (s *roleService) GetById(ctx context.Context, roleId int64) (*model.Role, error) {
	var role model.Role
	err := g.DB().Model("role").Where("id = ?", roleId).Scan(&role)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// Create creates a new role
func (s *roleService) Create(ctx context.Context, name, description string, status int) (int64, error) {
	result, err := g.DB().Model("role").Insert(g.Map{
		"role_name":   name,
		"description": description,
		"status":      status,
		"create_time": time.Now().Unix(),
		"update_time": time.Now().Unix(),
	})
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Update updates role information
func (s *roleService) Update(ctx context.Context, roleId int64, name, description string, status int) error {
	_, err := g.DB().Model("role").Data(g.Map{
		"name":        name,
		"description": description,
		"status":      status,
		"update_time": time.Now(),
	}).Where("id = ?", roleId).Update()
	return err
}

// Delete deletes a role
func (s *roleService) Delete(ctx context.Context, roleId int64) error {
	_, err := g.DB().Model("role").Where("id = ?", roleId).Delete()
	return err
}

// BindPermissions binds permissions to a role
func (s *roleService) BindPermissions(ctx context.Context, roleId int64, permissionIds []int64) error {
	// Delete existing permissions
	_, err := g.DB().Model("role_permission").Where("role_id = ?", roleId).Delete()
	if err != nil {
		return err
	}

	// Add new permissions
	for _, permId := range permissionIds {
		_, err = g.DB().Model("role_permission").Data(g.Map{
			"role_id":       roleId,
			"permission_id": permId,
			"create_time":   time.Now(),
		}).Insert()
		if err != nil {
			return err
		}
	}

	return nil
}

// GetPermissions gets permissions assigned to a role
func (s *roleService) GetPermissions(ctx context.Context, roleId int64) ([]model.Permission, error) {
	var permissions []model.Permission
	err := g.DB().Model("permission").
		LeftJoin("role_permission", "permission.id=role_permission.permission_id").
		Where("role_permission.role_id = ?", roleId).
		Scan(&permissions)
	return permissions, err
}

// NameExists checks if role name exists
func (s *roleService) NameExists(ctx context.Context, name string) (bool, error) {
	count, err := g.DB().Model("role").Where("name = ?", name).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// AssignPermission assigns a single permission to a role
func (s *roleService) AssignPermission(ctx context.Context, roleId int64, permissionId int64) error {
	_, err := g.DB().Model("role_permission").Data(g.Map{
		"role_id":       roleId,
		"permission_id": permissionId,
		"create_time":   time.Now(),
	}).Insert()
	return err
}

// ClearPermissions clears all permissions assigned to a role
func (s *roleService) ClearPermissions(ctx context.Context, roleId int64) error {
	_, err := g.DB().Model("role_permission").Where("role_id = ?", roleId).Delete()
	return err
}

// HasAccounts checks if role has associated accounts
func (s *roleService) HasAccounts(ctx context.Context, roleId int64) (bool, error) {
	count, err := g.DB().Model("account_role").Where("role_id = ?", roleId).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetRolePermissions gets role permissions (alias for GetPermissions)
func (s *roleService) GetRolePermissions(ctx context.Context, roleId int64) ([]model.Permission, error) {
	return s.GetPermissions(ctx, roleId)
}
