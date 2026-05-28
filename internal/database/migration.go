// Package database — GORM AutoMigrate 自动建表
// 注册所有 GORM 模型，启动时自动创建/更新表结构
package database

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"helloGo/internal/module/department"
	"helloGo/internal/module/dict"
	logModel "helloGo/internal/module/log"
	"helloGo/internal/module/menu"
	"helloGo/internal/module/permission"
	"helloGo/internal/module/role"
	"helloGo/internal/module/upload"
	"helloGo/internal/module/user"
)

// AutoMigrate 注册所有模型并自动迁移
// 按依赖顺序注册：独立表 → 有外键依赖的表
func AutoMigrate(db *gorm.DB, log *zap.Logger) error {
	models := []interface{}{
		// 1. 独立表（无外键依赖）
		&dict.Dict{},
		&logModel.Log{},
		&upload.Upload{},

		// 2. 自引用树（parent/children）
		&menu.Menu{},
		&department.Department{},

		// 3. 角色（无外键；Permission 和 User 均依赖此表，必须先创建）
		&role.Role{},

		// 4. 权限（role_id 外键 → roles 表）
		&permission.Permission{},

		// 5. 用户（通过 user_roles 中间表多对多关联 roles）
		&user.User{},
	}

	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	log.Info("数据库迁移完成", zap.Int("models", len(models)))
	return nil
}
