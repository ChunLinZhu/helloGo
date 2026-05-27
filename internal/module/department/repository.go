// Package department — 部门数据访问层
package department

import (
	"gorm.io/gorm"
)

// Repository 部门数据访问接口
type Repository interface {
	FindByID(id string) (*Department, error)
	FindAll() ([]Department, error)
	Create(dept *Department) error
	Update(dept *Department) error
	Delete(id string) error
	HasChildren(id string) (bool, error)
}

// repository 部门数据访问实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建部门数据访问层
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// FindByID 按 ID 查询部门
func (r *repository) FindByID(id string) (*Department, error) {
	var dept Department
	err := r.db.Where("id = ?", id).First(&dept).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

// FindAll 查询所有部门
func (r *repository) FindAll() ([]Department, error) {
	var depts []Department
	err := r.db.Order("name ASC").Find(&depts).Error
	return depts, err
}

// Create 创建部门
func (r *repository) Create(dept *Department) error {
	return r.db.Create(dept).Error
}

// Update 更新部门
func (r *repository) Update(dept *Department) error {
	return r.db.Save(dept).Error
}

// Delete 删除部门
func (r *repository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&Department{}).Error
}

// HasChildren 检查部门是否有子部门
func (r *repository) HasChildren(id string) (bool, error) {
	var count int64
	err := r.db.Model(&Department{}).Where("parent_id = ?", id).Count(&count).Error
	return count > 0, err
}
