// Package dict — Dict 实体模型
// 字典表，用于存储键值对配置（如状态码、类型枚举等）
package dict

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Dict 字典模型
type Dict struct {
	ID          string    `gorm:"primaryKey;size:36" json:"id"`
	Type        string    `gorm:"size:128;not null;uniqueIndex:idx_type_key" json:"type"`
	Key         string    `gorm:"size:128;not null;uniqueIndex:idx_type_key" json:"key"`
	Value       string    `gorm:"size:256;not null" json:"value"`
	Description *string   `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName GORM 表名
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
