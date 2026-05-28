// Package user — User 实体模型
// 用户表，支持多对多角色关联（通过 user_roles 中间表）
package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"helloGo/internal/module/role"
)

// User 用户模型
type User struct {
	ID           string     `gorm:"primaryKey;size:36" json:"id"`
	Username     string     `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string     `gorm:"size:128;not null" json:"-"` // json:"-" 防止密码哈希泄露
	Email        *string    `gorm:"size:128;index" json:"email"`
	Phone        *string    `gorm:"size:32;index" json:"phone"`
	IsActive     bool       `gorm:"default:true;not null" json:"isActive"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`

	// 关联关系
	Roles []role.Role `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// TableName GORM 表名
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
