// Package role — Role 实体模型
// 角色表，支持多对多用户关联 + 一对多权限关联
package role

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"helloGo/internal/module/permission"
)

// Role 角色模型
type Role struct {
	ID          string                   `gorm:"primaryKey;size:36" json:"id"`
	Code        string                   `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Name        string                   `gorm:"size:128;not null" json:"name"`
	Description *string                  `gorm:"size:255" json:"description"`
	CreatedAt   time.Time                `gorm:"autoCreateTime;not null" json:"createdAt"`
	UpdatedAt   time.Time                `gorm:"autoUpdateTime;not null" json:"updatedAt"`

	// 关联关系
	Permissions []permission.Permission `gorm:"foreignKey:RoleID" json:"permissions,omitempty"`
}

// TableName GORM 表名
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
