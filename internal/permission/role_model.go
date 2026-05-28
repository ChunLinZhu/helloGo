// Package permission — 权限微服务 GORM 模型
// 角色模型，自包含定义，不依赖 Phase 1 的 internal/module/role
package permission

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role 角色模型
type Role struct {
	ID          string    `gorm:"primaryKey;size:36" json:"id"`
	Code        string    `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Description *string   `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// 关联关系 — 角色拥有多个权限
	Permissions []Permission `gorm:"foreignKey:RoleID" json:"permissions,omitempty"`
}

// TableName 角色表名
func (Role) TableName() string {
	return "roles"
}

// BeforeCreate 创建前自动生成 UUID
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}
