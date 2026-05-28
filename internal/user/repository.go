// Package user — 用户数据访问层
package user

import (
	"gorm.io/gorm"
)

// Repository 用户数据访问接口
type Repository interface {
	Create(user *User) error
	FindByID(id string) (*User, error)
	FindByUsername(username string) (*User, error)
	Update(user *User) error
	Delete(id string) error
	List(page, limit int, keyword string) ([]User, int64, error)
	AssociateRoles(userID string, roleIDs []string) error
	FindRolesByCodes(codes []string) ([]Role, error)
}

// repository 用户数据访问实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建用户数据访问层
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create 创建用户
func (r *repository) Create(user *User) error {
	return r.db.Create(user).Error
}

// FindByID 按 ID 查询用户（含角色关联）
func (r *repository) FindByID(id string) (*User, error) {
	var user User
	err := r.db.Preload("Roles").Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername 按用户名查询用户
func (r *repository) FindByUsername(username string) (*User, error) {
	var user User
	err := r.db.Preload("Roles").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (r *repository) Update(user *User) error {
	return r.db.Save(user).Error
}

// Delete 删除用户
func (r *repository) Delete(id string) error {
	// 先删除关联的角色
	if err := r.db.Exec("DELETE FROM user_roles WHERE user_id = ?", id).Error; err != nil {
		return err
	}
	// 再删除用户
	return r.db.Where("id = ?", id).Delete(&User{}).Error
}

// List 分页查询用户列表（支持关键词搜索）
func (r *repository) List(page, limit int, keyword string) ([]User, int64, error) {
	var users []User
	var total int64

	query := r.db.Model(&User{})

	// 关键词搜索（用户名、邮箱、手机号）
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR email LIKE ? OR phone LIKE ?",
			like, like, like)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询（含角色关联）
	offset := (page - 1) * limit
	err := query.Preload("Roles").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&users).Error

	return users, total, err
}

// AssociateRoles 关联用户角色（先删后插）
func (r *repository) AssociateRoles(userID string, roleIDs []string) error {
	// 先删除现有关联
	if err := r.db.Exec("DELETE FROM user_roles WHERE user_id = ?", userID).Error; err != nil {
		return err
	}

	if len(roleIDs) == 0 {
		return nil
	}

	// 添加新关联
	for _, roleID := range roleIDs {
		if err := r.db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)",
			userID, roleID).Error; err != nil {
			return err
		}
	}

	return nil
}

// FindRolesByCodes 按角色编码查询角色列表
func (r *repository) FindRolesByCodes(codes []string) ([]Role, error) {
	if len(codes) == 0 {
		return nil, nil
	}
	var roles []Role
	err := r.db.Where("code IN ?", codes).Find(&roles).Error
	return roles, err
}
