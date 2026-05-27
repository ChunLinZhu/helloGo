// Package permission — 权限数据访问层
package permission

import (
	"gorm.io/gorm"

	"helloGo/internal/pkg/pagination"
)

// Repository 权限数据访问接口
type Repository interface {
	FindByID(id string) (*Permission, error)
	FindByKey(key string) (*Permission, error)
	List(page pagination.Pagination, keyword string) ([]Permission, int64, error)
	Create(perm *Permission) error
	Update(perm *Permission) error
	Delete(id string) error
}

// repository 权限数据访问实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建权限数据访问层
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// FindByID 按 ID 查询权限
func (r *repository) FindByID(id string) (*Permission, error) {
	var perm Permission
	err := r.db.Where("id = ?", id).First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

// FindByKey 按 key 查询权限
func (r *repository) FindByKey(key string) (*Permission, error) {
	var perm Permission
	err := r.db.Where("`key` = ?", key).First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

// List 分页查询权限列表
func (r *repository) List(page pagination.Pagination, keyword string) ([]Permission, int64, error) {
	var perms []Permission
	var total int64

	query := r.db.Model(&Permission{})

	// 关键词搜索（key、描述）
	if keyword != "" {
		keyword = "%" + keyword + "%"
		query = query.Where("`key` LIKE ? OR description LIKE ?", keyword, keyword)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := query.Order("created_at DESC").
		Offset(page.Offset()).
		Limit(page.Limit).
		Find(&perms).Error

	if err != nil {
		return nil, 0, err
	}

	return perms, total, nil
}

// Create 创建权限
func (r *repository) Create(perm *Permission) error {
	return r.db.Create(perm).Error
}

// Update 更新权限
func (r *repository) Update(perm *Permission) error {
	return r.db.Save(perm).Error
}

// Delete 删除权限
func (r *repository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&Permission{}).Error
}
