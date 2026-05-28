// Package biz — 字典 GORM 模型
package biz

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Dict 字典模型（type + key 组合唯一）
type Dict struct {
	ID          string    `gorm:"primaryKey;size:36" json:"id"`
	Type        string    `gorm:"size:128;not null;uniqueIndex:idx_type_key" json:"type"`
	Key         string    `gorm:"size:128;not null;uniqueIndex:idx_type_key" json:"key"`
	Value       string    `gorm:"size:256;not null" json:"value"`
	Description *string   `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 字典表名
func (Dict) TableName() string {
	return "dicts"
}

// BeforeCreate 创建前自动生成 UUID
func (d *Dict) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}
