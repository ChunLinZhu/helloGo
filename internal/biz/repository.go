// Package biz — 业务数据访问层
package biz

import (
	"gorm.io/gorm"
)

// Repository 业务数据访问接口
type Repository interface {
	// 部门
	CreateDepartment(dept *Department) error
	FindAllDepartments() ([]Department, error)

	// 字典
	CreateDict(dict *Dict) error
	ListDicts(page, limit int, keyword, dictType string) ([]Dict, int64, error)

	// 日志
	CreateLog(log *Log) error
	ListLogs(page, limit int, level string) ([]Log, int64, error)

	// 上传
	ListUploads(page, limit int) ([]Upload, int64, error)
	DeleteUpload(id string) error
}

// repository 业务数据访问实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建业务数据访问层
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// ── 部门 ──────────────────────────────────────────────────────

// CreateDepartment 创建部门
func (r *repository) CreateDepartment(dept *Department) error {
	return r.db.Create(dept).Error
}

// FindAllDepartments 查询所有部门
func (r *repository) FindAllDepartments() ([]Department, error) {
	var depts []Department
	err := r.db.Order("name ASC").Find(&depts).Error
	return depts, err
}

// ── 字典 ──────────────────────────────────────────────────────

// CreateDict 创建字典项
func (r *repository) CreateDict(dict *Dict) error {
	return r.db.Create(dict).Error
}

// ListDicts 分页查询字典列表
func (r *repository) ListDicts(page, limit int, keyword, dictType string) ([]Dict, int64, error) {
	var dicts []Dict
	var total int64

	query := r.db.Model(&Dict{})

	if dictType != "" {
		query = query.Where("type = ?", dictType)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("key LIKE ? OR value LIKE ? OR description LIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("type ASC, key ASC").
		Offset(offset).
		Limit(limit).
		Find(&dicts).Error

	return dicts, total, err
}

// ── 日志 ──────────────────────────────────────────────────────

// CreateLog 创建日志
func (r *repository) CreateLog(log *Log) error {
	return r.db.Create(log).Error
}

// ListLogs 分页查询日志
func (r *repository) ListLogs(page, limit int, level string) ([]Log, int64, error) {
	var logs []Log
	var total int64

	query := r.db.Model(&Log{})

	if level != "" {
		query = query.Where("level = ?", level)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error

	return logs, total, err
}

// ── 上传 ──────────────────────────────────────────────────────

// ListUploads 分页查询上传文件
func (r *repository) ListUploads(page, limit int) ([]Upload, int64, error) {
	var uploads []Upload
	var total int64

	query := r.db.Model(&Upload{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&uploads).Error

	return uploads, total, err
}

// DeleteUpload 删除上传文件记录
func (r *repository) DeleteUpload(id string) error {
	return r.db.Where("id = ?", id).Delete(&Upload{}).Error
}
