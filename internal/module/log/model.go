// Package log — Log 实体模型
// 日志表，记录系统运行日志（info/warn/error/debug）
package log

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Log 日志模型
type Log struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	Level     string    `gorm:"size:16;not null;index:idx_level_created" json:"level"` // info/warn/error/debug
	Message   string    `gorm:"size:256;not null" json:"message"`
	Meta      *string   `gorm:"type:text" json:"meta"` // JSON 格式的元数据
	CreatedAt time.Time `gorm:"autoCreateTime;index:idx_level_created" json:"createdAt"`
}

// TableName GORM 表名
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
