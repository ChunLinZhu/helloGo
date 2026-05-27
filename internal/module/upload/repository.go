// Package upload — 上传数据访问层
package upload

import (
	"time"

	"gorm.io/gorm"

	"helloGo/internal/pkg/pagination"
)

// Repository 上传数据访问接口
type Repository interface {
	FindByID(id string) (*Upload, error)
	List(page pagination.Pagination) ([]Upload, int64, error)
	Create(upload *Upload) error
	Delete(id string) error
	DeleteExpired(before time.Time) ([]Upload, error)
}

// repository 上传数据访问实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建上传数据访问层
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// FindByID 按 ID 查询上传记录
func (r *repository) FindByID(id string) (*Upload, error) {
	var upload Upload
	err := r.db.Where("id = ?", id).First(&upload).Error
	if err != nil {
		return nil, err
	}
	return &upload, nil
}

// List 分页查询上传记录
func (r *repository) List(page pagination.Pagination) ([]Upload, int64, error) {
	var uploads []Upload
	var total int64

	query := r.db.Model(&Upload{})

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询（按创建时间倒序）
	err := query.Order("created_at DESC").
		Offset(page.Offset()).
		Limit(page.Limit).
		Find(&uploads).Error

	if err != nil {
		return nil, 0, err
	}

	return uploads, total, nil
}

// Create 创建上传记录
func (r *repository) Create(upload *Upload) error {
	return r.db.Create(upload).Error
}

// Delete 删除上传记录
func (r *repository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&Upload{}).Error
}

// DeleteExpired 删除过期上传记录，返回被删除的记录列表（用于清理磁盘文件）
func (r *repository) DeleteExpired(before time.Time) ([]Upload, error) {
	var uploads []Upload
	if err := r.db.Where("created_at < ?", before).Find(&uploads).Error; err != nil {
		return nil, err
	}
	if len(uploads) == 0 {
		return nil, nil
	}
	if err := r.db.Where("created_at < ?", before).Delete(&Upload{}).Error; err != nil {
		return nil, err
	}
	return uploads, nil
}
