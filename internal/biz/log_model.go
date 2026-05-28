// Package biz — 操作日志 GORM 模型
package biz

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Log 操作日志模型
type Log struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	Level     string    `gorm:"size:16;not null;index:idx_level_created" json:"level"`
	Message   string    `gorm:"size:256;not null" json:"message"`
	Meta      *string   `gorm:"type:text" json:"meta"`
	CreatedAt time.Time `gorm:"autoCreateTime;index:idx_level_created" json:"createdAt"`
}

// TableName 日志表名
func (Log) TableName() string {
	return "logs"
}

// BeforeCreate 创建前自动生成 UUID
func (l *Log) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}
