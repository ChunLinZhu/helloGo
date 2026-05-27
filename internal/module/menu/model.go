// Package menu — Menu 实体模型
// 菜单表，自引用树结构（parent/children）
package menu

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Menu 菜单模型（自引用树）
type Menu struct {
	ID       string    `gorm:"primaryKey;size:36" json:"id"`
	Name     string    `gorm:"size:128;not null" json:"name"`
	Path     *string   `gorm:"size:256;uniqueIndex" json:"path"`
	Icon     *string   `gorm:"size:128" json:"icon"`
	Order    int       `gorm:"default:0;index" json:"order"`
	ParentID *string   `gorm:"size:36;index" json:"parentId"`
	CreatedAt time.Time `gorm:"autoCreateTime;not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;not null" json:"updatedAt"`

	// 自引用关联
	Parent   *Menu  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Menu `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// TableName GORM 表名
func (Menu) TableName() string {
	return "menus"
}

// BeforeCreate 创建前自动生成 UUID
func (m *Menu) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}
