// Package menu — 菜单数据访问层
package menu

import (
	"gorm.io/gorm"
)

// Repository 菜单数据访问接口
type Repository interface {
	FindByID(id string) (*Menu, error)
	FindAll() ([]Menu, error)
	Create(menu *Menu) error
	Update(menu *Menu) error
	Delete(id string) error
	HasChildren(id string) (bool, error)
}

// repository 菜单数据访问实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建菜单数据访问层
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// FindByID 按 ID 查询菜单
func (r *repository) FindByID(id string) (*Menu, error) {
	var menu Menu
	err := r.db.Where("id = ?", id).First(&menu).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

// FindAll 查询所有菜单（按 order 排序）
func (r *repository) FindAll() ([]Menu, error) {
	var menus []Menu
	err := r.db.Order("`order` ASC").Find(&menus).Error
	return menus, err
}

// Create 创建菜单
func (r *repository) Create(menu *Menu) error {
	return r.db.Create(menu).Error
}

// Update 更新菜单
func (r *repository) Update(menu *Menu) error {
	return r.db.Save(menu).Error
}

// Delete 删除菜单
func (r *repository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&Menu{}).Error
}

// HasChildren 检查菜单是否有子菜单
func (r *repository) HasChildren(id string) (bool, error) {
	var count int64
	err := r.db.Model(&Menu{}).Where("parent_id = ?", id).Count(&count).Error
	return count > 0, err
}
