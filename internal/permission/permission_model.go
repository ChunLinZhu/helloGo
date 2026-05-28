// Package permission — 权限 GORM 模型
package permission

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Permission 权限模型
// 每个权限属于一个角色（通过 RoleID 外键关联）
type Permission struct {
	ID          string    `gorm:"primaryKey;size:36" json:"id"`
	Key         string    `gorm:"size:128;uniqueIndex;not null" json:"key"`
	Description *string   `gorm:"size:256" json:"description"`
	RoleID      string    `gorm:"size:36;index;not null" json:"roleId"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 权限表名
func (Permission) TableName() string {
	return "permissions"
}

// BeforeCreate 创建前自动生成 UUID
func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
