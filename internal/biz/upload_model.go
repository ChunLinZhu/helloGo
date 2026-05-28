// Package biz — 上传文件 GORM 模型
package biz

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Upload 上传文件记录模型
type Upload struct {
	ID           string    `gorm:"primaryKey;size:36" json:"id"`
	Filename     string    `gorm:"size:256;not null" json:"filename"`
	OriginalName string    `gorm:"size:256;not null" json:"originalName"`
	Mimetype     string    `gorm:"size:128;not null" json:"mimetype"`
	Size         int64     `gorm:"not null" json:"size"`
	Path         string    `gorm:"size:512;not null" json:"path"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName 上传表名
func (Upload) TableName() string {
	return "uploads"
}

// BeforeCreate 创建前自动生成 UUID
func (u *Upload) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}
