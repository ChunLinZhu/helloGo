// Package log — 日志数据访问层
package log

import (
	"gorm.io/gorm"

	"helloGo/internal/pkg/pagination"
)

// Repository 日志数据访问接口
type Repository interface {
	FindByID(id string) (*Log, error)
	List(page pagination.Pagination, level string) ([]Log, int64, error)
	Create(log *Log) error
}

// repository 日志数据访问实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建日志数据访问层
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// FindByID 按 ID 查询日志
func (r *repository) FindByID(id string) (*Log, error) {
	var l Log
	err := r.db.Where("id = ?", id).First(&l).Error
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// List 分页查询日志列表
func (r *repository) List(page pagination.Pagination, level string) ([]Log, int64, error) {
	var logs []Log
	var total int64

	query := r.db.Model(&Log{})

	// 按级别过滤
	if level != "" {
		query = query.Where("level = ?", level)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询（按创建时间倒序）
	err := query.Order("created_at DESC").
		Offset(page.Offset()).
		Limit(page.Limit).
		Find(&logs).Error

	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// Create 创建日志
func (r *repository) Create(l *Log) error {
	return r.db.Create(l).Error
}
