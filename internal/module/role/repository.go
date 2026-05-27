// Package role — 角色数据访问层
package role

import (
	"helloGo/internal/pkg/pagination"
	"gorm.io/gorm"
)

// Repository 角色数据访问接口
type Repository interface {
	FindByID(id string) (*Role, error)
	FindByCode(code string) (*Role, error)
	List(page pagination.Pagination, keyword string) ([]Role, int64, error)
	Create(role *Role) error
	Update(role *Role) error
	Delete(id string) error
	AssociatePermissions(roleID string, permissionIDs []string) error
}

// repository 角色数据访问实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建角色数据访问层
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// FindByID 按 ID 查询角色（含权限）
func (r *repository) FindByID(id string) (*Role, error) {
	var role Role
	// 手动加载权限
	err := r.db.Where("id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}

	// 查询该角色的权限
	r.db.Where("role_id = ?", id).Find(&role.Permissions)

	return &role, nil
}

// FindByCode 按代码查询角色
func (r *repository) FindByCode(code string) (*Role, error) {
	var role Role
	err := r.db.Where("code = ?", code).First(&role).Error
	if err != nil {
		return nil, err
	}

	// 查询该角色的权限
	r.db.Where("role_id = ?", role.ID).Find(&role.Permissions)

	return &role, nil
}

// List 分页查询角色列表
func (r *repository) List(page pagination.Pagination, keyword string) ([]Role, int64, error) {
	var roles []Role
	var total int64

	query := r.db.Model(&Role{})

	// 关键词搜索（代码、名称）
	if keyword != "" {
		keyword = "%" + keyword + "%"
		query = query.Where("code LIKE ? OR name LIKE ?", keyword, keyword)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := query.Order("created_at DESC").
		Offset(page.Offset()).
		Limit(page.Limit).
		Find(&roles).Error

	if err != nil {
		return nil, 0, err
	}

	// 为每个角色加载权限
	for i := range roles {
		r.db.Where("role_id = ?", roles[i].ID).Find(&roles[i].Permissions)
	}

	return roles, total, err
}

// Create 创建角色
func (r *repository) Create(role *Role) error {
	return r.db.Create(role).Error
}

// Update 更新角色
func (r *repository) Update(role *Role) error {
	return r.db.Save(role).Error
}

// Delete 删除角色
func (r *repository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&Role{}).Error
}

// AssociatePermissions 关联角色权限
func (r *repository) AssociatePermissions(roleID string, permissionIDs []string) error {
	// 先将该角色的所有权限解除关联
	if err := r.db.Exec("UPDATE permissions SET role_id = '' WHERE role_id = ?", roleID).Error; err != nil {
		return err
	}

	// 如果没有新权限，直接返回
	if len(permissionIDs) == 0 {
		return nil
	}

	// 将新权限关联到该角色
	return r.db.Exec("UPDATE permissions SET role_id = ? WHERE id IN ?", roleID, permissionIDs).Error
}
