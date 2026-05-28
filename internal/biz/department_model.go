// Package biz — 部门 GORM 模型（自引用树结构）
package biz

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Department 部门模型
type Department struct {
	ID          string    `gorm:"primaryKey;size:36" json:"id"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Description *string   `gorm:"size:256" json:"description"`
	ParentID    *string   `gorm:"size:36;index" json:"parentId"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// 自引用关联
	Parent   *Department  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Department `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// TableName 部门表名
func (Department) TableName() string {
	return "departments"
}

// BeforeCreate 创建前自动生成 UUID
func (d *Department) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}
