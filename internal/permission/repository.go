// Package permission — 权限数据访问层
package permission

import (
	"gorm.io/gorm"
)

// Repository 权限数据访问接口
type Repository interface {
	// 角色
	CreateRole(role *Role) error
	FindRoleByID(id string) (*Role, error)
	FindRoleByCode(code string) (*Role, error)
	ListRoles(page, limit int, keyword string) ([]Role, int64, error)
	AssociatePermissions(roleID string, permissionIDs []string) error

	// 权限
	CreatePermission(perm *Permission) error
	FindPermissionByID(id string) (*Permission, error)
	UpdatePermission(perm *Permission) error
	DeletePermission(id string) error
	ListPermissions(page, limit int, keyword string) ([]Permission, int64, error)
	GetPermissionsByRoleID(roleID string) ([]Permission, error)

	// 菜单
	FindAllMenus() ([]Menu, error)
}

// repository 权限数据访问实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建权限数据访问层
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// ── 角色 ──────────────────────────────────────────────────────

// CreateRole 创建角色
func (r *repository) CreateRole(role *Role) error {
	return r.db.Create(role).Error
}

// FindRoleByID 按 ID 查询角色（含权限关联）
func (r *repository) FindRoleByID(id string) (*Role, error) {
	var role Role
	err := r.db.Preload("Permissions").Where("id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// FindRoleByCode 按代码查询角色（含权限关联）
func (r *repository) FindRoleByCode(code string) (*Role, error) {
	var role Role
	err := r.db.Preload("Permissions").Where("code = ?", code).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// ListRoles 分页查询角色列表（支持关键词搜索）
func (r *repository) ListRoles(page, limit int, keyword string) ([]Role, int64, error) {
	var roles []Role
	var total int64

	query := r.db.Model(&Role{})

	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("code LIKE ? OR name LIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Preload("Permissions").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&roles).Error

	return roles, total, err
}

// AssociatePermissions 为角色分配权限（全量替换）
// 先将角色现有权限的 role_id 清空，再将新权限的 role_id 设为该角色
func (r *repository) AssociatePermissions(roleID string, permissionIDs []string) error {
	// 清除角色现有权限关联
	if err := r.db.Model(&Permission{}).
		Where("role_id = ?", roleID).
		Update("role_id", "").Error; err != nil {
		return err
	}

	if len(permissionIDs) == 0 {
		return nil
	}

	// 设置新权限关联
	return r.db.Model(&Permission{}).
		Where("id IN ?", permissionIDs).
		Update("role_id", roleID).Error
}

// ── 权限 ──────────────────────────────────────────────────────

// CreatePermission 创建权限
func (r *repository) CreatePermission(perm *Permission) error {
	return r.db.Create(perm).Error
}

// FindPermissionByID 按 ID 查询权限
func (r *repository) FindPermissionByID(id string) (*Permission, error) {
	var perm Permission
	err := r.db.Where("id = ?", id).First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

// UpdatePermission 更新权限
func (r *repository) UpdatePermission(perm *Permission) error {
	return r.db.Save(perm).Error
}

// DeletePermission 删除权限
func (r *repository) DeletePermission(id string) error {
	return r.db.Where("id = ?", id).Delete(&Permission{}).Error
}

// ListPermissions 分页查询权限列表（支持关键词搜索）
func (r *repository) ListPermissions(page, limit int, keyword string) ([]Permission, int64, error) {
	var perms []Permission
	var total int64

	query := r.db.Model(&Permission{})

	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("key LIKE ? OR description LIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&perms).Error

	return perms, total, err
}

// GetPermissionsByRoleID 按角色 ID 查询所有权限
func (r *repository) GetPermissionsByRoleID(roleID string) ([]Permission, error) {
	var perms []Permission
	err := r.db.Where("role_id = ?", roleID).Order("key ASC").Find(&perms).Error
	return perms, err
}

// ── 菜单 ──────────────────────────────────────────────────────

// FindAllMenus 查询所有菜单（按 order 排序）
func (r *repository) FindAllMenus() ([]Menu, error) {
	var menus []Menu
	err := r.db.Order("`order` ASC").Find(&menus).Error
	return menus, err
}
