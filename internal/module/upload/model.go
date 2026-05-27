// Package upload — Upload 实体模型
// 上传文件记录表，跟踪所有上传文件的元数据
package upload

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Upload 上传文件模型
type Upload struct {
	ID           string    `gorm:"primaryKey;size:36" json:"id"`
	Filename     string    `gorm:"size:256;not null" json:"filename"`         // 存储文件名（UUID）
	OriginalName string    `gorm:"size:256;not null" json:"originalName"`     // 原始文件名
	Mimetype     string    `gorm:"size:128;not null" json:"mimetype"`         // MIME 类型
	Size         int64     `gorm:"not null" json:"size"`                      // 文件大小（字节）
	Path         string    `gorm:"size:512;not null" json:"path"`             // 存储路径
	CreatedAt    time.Time `gorm:"autoCreateTime;not null" json:"createdAt"`
}

// TableName GORM 表名
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
