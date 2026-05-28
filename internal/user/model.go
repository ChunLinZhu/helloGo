// Package user — 用户微服务 GORM 模型
// 自包含定义，不依赖 Phase 1 的 internal/module/role
package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           string    `gorm:"primaryKey;size:36" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:128;not null" json:"-"`
	Email        *string   `gorm:"size:128;index" json:"email"`
	Phone        *string   `gorm:"size:32;index" json:"phone"`
	IsActive     bool      `gorm:"default:true;not null" json:"isActive"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// 关联关系 — 使用本地 Role 定义
	Roles []Role `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// Role 最小化角色模型，仅用于 user_roles JOIN 查询
// 不拥有 roles 表的完整业务逻辑（由 Permission Service 管理）
type Role struct {
	ID   string `gorm:"primaryKey;size:36" json:"id"`
	Code string `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Name string `gorm:"size:128;not null" json:"name"`
}

// TableName 角色表名
func (Role) TableName() string {
	return "roles"
}

// TableName 用户表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate 创建前自动生成 UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}
