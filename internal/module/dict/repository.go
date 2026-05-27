// Package dict — 字典数据访问层
package dict

import (
	"gorm.io/gorm"

	"helloGo/internal/pkg/pagination"
)

// Repository 字典数据访问接口
type Repository interface {
	FindByID(id string) (*Dict, error)
	FindByTypeKey(dictType, key string) (*Dict, error)
	List(page pagination.Pagination, keyword, dictType string) ([]Dict, int64, error)
	Create(dict *Dict) error
	Update(dict *Dict) error
	Delete(id string) error
}

// repository 字典数据访问实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建字典数据访问层
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// FindByID 按 ID 查询字典项
func (r *repository) FindByID(id string) (*Dict, error) {
	var dict Dict
	err := r.db.Where("id = ?", id).First(&dict).Error
	if err != nil {
		return nil, err
	}
	return &dict, nil
}

// FindByTypeKey 按 type+key 查询字典项（复合唯一约束）
func (r *repository) FindByTypeKey(dictType, key string) (*Dict, error) {
	var dict Dict
	err := r.db.Where("`type` = ? AND `key` = ?", dictType, key).First(&dict).Error
	if err != nil {
		return nil, err
	}
	return &dict, nil
}

// List 分页查询字典列表
func (r *repository) List(page pagination.Pagination, keyword, dictType string) ([]Dict, int64, error) {
	var dicts []Dict
	var total int64

	query := r.db.Model(&Dict{})

	// 按类型过滤
	if dictType != "" {
		query = query.Where("`type` = ?", dictType)
	}

	// 关键词搜索（key、value、描述）
	if keyword != "" {
		keyword = "%" + keyword + "%"
		query = query.Where("`key` LIKE ? OR value LIKE ? OR description LIKE ?", keyword, keyword, keyword)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := query.Order("created_at DESC").
		Offset(page.Offset()).
		Limit(page.Limit).
		Find(&dicts).Error

	if err != nil {
		return nil, 0, err
	}

	return dicts, total, nil
}

// Create 创建字典项
func (r *repository) Create(dict *Dict) error {
	return r.db.Create(dict).Error
}

// Update 更新字典项
func (r *repository) Update(dict *Dict) error {
	return r.db.Save(dict).Error
}

// Delete 删除字典项
func (r *repository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&Dict{}).Error
}
