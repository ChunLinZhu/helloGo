// Package main — Phase 2 种子数据脚本
// 创建 admin 角色/用户/权限 + 批量测试数据
// 使用 internal/shared 包，兼容 Phase 2 微服务架构
// 用法: go run cmd/seed-v2/main.go [--purge]
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"helloGo/internal/biz"
	"helloGo/internal/permission"
	"helloGo/internal/shared/config"
	"helloGo/internal/shared/database"
	"helloGo/internal/user"
)

var purgeFlag = flag.Bool("purge", false, "清除现有数据后重新播种")

func main() {
	flag.Parse()

	// ── 1. 加载配置 ───────────────────────────────────────
	cfg, err := config.Load("seed", "./configs")
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// ── 2. 初始化日志 ─────────────────────────────────────
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// ── 3. 初始化数据库 ───────────────────────────────────
	db, err := database.Connect(cfg.Database, cfg.Service.Env, logger)
	if err != nil {
		logger.Fatal("数据库初始化失败", zap.Error(err))
	}

	// ── 4. 数据库迁移（确保表结构存在）────────────────────
	if err := database.AutoMigrate(db, logger,
		&user.User{}, &user.Role{},
		&permission.Role{}, &permission.Permission{}, &permission.Menu{},
		&biz.Department{}, &biz.Dict{}, &biz.Log{}, &biz.Upload{},
	); err != nil {
		logger.Fatal("数据库迁移失败", zap.Error(err))
	}

	// ── 5. 清除数据（--purge 参数）──────────────────────
	if *purgeFlag {
		logger.Info("正在清除现有数据...")
		// 按依赖顺序删除：关联表 → 主表
		db.Exec("DELETE FROM user_roles")
		db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&user.User{})
		db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&permission.Permission{})
		db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&permission.Role{})
		db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&permission.Menu{})
		db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&biz.Department{})
		db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&biz.Dict{})
		db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&biz.Log{})
		db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&biz.Upload{})
		logger.Info("数据清除完成")
	}

	// ── 6. 创建 admin 角色 + 权限 ─────────────────────────
	logger.Info("创建 admin 角色和权限...")
	adminRole := createAdminRole(db, logger)

	// ── 7. 创建 admin 用户 ────────────────────────────────
	logger.Info("创建 admin 用户...")
	createAdminUser(db, adminRole, logger)

	// ── 8. 批量创建测试数据 ───────────────────────────────
	logger.Info("批量创建测试数据...")
	batchCreateTestData(db, logger)

	logger.Info("种子数据播种完成！")
}

// createAdminRole 创建 admin 角色及权限
func createAdminRole(db *gorm.DB, logger *zap.Logger) *permission.Role {
	// 创建权限
	perms := []permission.Permission{
		{ID: uuid.New().String(), Key: "user:create", Description: strPtr("创建用户")},
		{ID: uuid.New().String(), Key: "role:manage", Description: strPtr("管理角色")},
	}

	adminRole := &permission.Role{
		ID:          uuid.New().String(),
		Code:        "admin",
		Name:        "管理员",
		Description: strPtr("系统管理员，拥有所有权限"),
		Permissions: perms,
	}

	// 检查是否已存在
	var count int64
	db.Model(&permission.Role{}).Where("code = ?", "admin").Count(&count)
	if count > 0 {
		logger.Info("admin 角色已存在，跳过创建")
		db.Where("code = ?", "admin").Preload("Permissions").First(adminRole)
		return adminRole
	}

	if err := db.Create(adminRole).Error; err != nil {
		logger.Fatal("创建 admin 角色失败", zap.Error(err))
	}
	logger.Info("admin 角色创建成功", zap.String("id", adminRole.ID))
	return adminRole
}

// createAdminUser 创建 admin 用户
func createAdminUser(db *gorm.DB, adminRole *permission.Role, logger *zap.Logger) {
	// 检查是否已存在
	var count int64
	db.Model(&user.User{}).Where("username = ?", "admin").Count(&count)
	if count > 0 {
		logger.Info("admin 用户已存在，跳过创建")
		return
	}

	// 密码哈希（admin123）
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		logger.Fatal("密码哈希失败", zap.Error(err))
	}

	// User Service 的 Role 模型（最小化，仅用于 user_roles JOIN）
	userRole := user.Role{
		ID:   adminRole.ID,
		Code: adminRole.Code,
		Name: adminRole.Name,
	}

	// 确保 User Service 的 roles 表也有这条记录
	db.Where("id = ?", userRole.ID).FirstOrCreate(&userRole)

	adminUser := &user.User{
		ID:           uuid.New().String(),
		Username:     "admin",
		PasswordHash: string(hash),
		Email:        strPtr("admin@example.com"),
		Phone:        strPtr("13800138000"),
		IsActive:     true,
		Roles:        []user.Role{userRole},
	}

	if err := db.Create(adminUser).Error; err != nil {
		logger.Fatal("创建 admin 用户失败", zap.Error(err))
	}
	logger.Info("admin 用户创建成功", zap.String("id", adminUser.ID))
}

// batchCreateTestData 批量创建测试数据
func batchCreateTestData(db *gorm.DB, logger *zap.Logger) {
	// ── 批量创建角色（10 个）────────────────────────────
	logger.Info("批量创建角色...")
	roles := make([]permission.Role, 0, 10)
	for i := 1; i <= 10; i++ {
		roles = append(roles, permission.Role{
			ID:          uuid.New().String(),
			Code:        fmt.Sprintf("role_%d", i),
			Name:        fmt.Sprintf("角色 %d", i),
			Description: strPtr(fmt.Sprintf("测试角色 %d", i)),
		})
	}
	db.Create(&roles)

	// ── 批量创建权限（每个角色 2 个权限，共 20 个）──────
	logger.Info("批量创建权限...")
	perms := make([]permission.Permission, 0, 20)
	for _, r := range roles {
		perms = append(perms, permission.Permission{
			ID:          uuid.New().String(),
			Key:         fmt.Sprintf("%s:read", r.Code),
			Description: strPtr(fmt.Sprintf("读取 %s", r.Name)),
			RoleID:      r.ID,
		})
		perms = append(perms, permission.Permission{
			ID:          uuid.New().String(),
			Key:         fmt.Sprintf("%s:write", r.Code),
			Description: strPtr(fmt.Sprintf("写入 %s", r.Name)),
			RoleID:      r.ID,
		})
	}
	db.Create(&perms)

	// ── 批量创建用户（100 个）───────────────────────────
	logger.Info("批量创建用户...")
	users := make([]user.User, 0, 100)
	for i := 1; i <= 100; i++ {
		hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		users = append(users, user.User{
			ID:           uuid.New().String(),
			Username:     fmt.Sprintf("user%d", i),
			PasswordHash: string(hash),
			Email:        strPtr(fmt.Sprintf("user%d@example.com", i)),
			Phone:        strPtr(fmt.Sprintf("13800138%03d", i)),
			IsActive:     i%10 != 0, // 每 10 个用户有 1 个禁用
		})
	}
	db.Create(&users)

	// 为前 50 个用户分配 role_1
	if len(roles) > 0 && len(users) > 0 {
		userRoles := make([]user.Role, 0, 50)
		for i := 0; i < 50 && i < len(users); i++ {
			userRoles = append(userRoles, user.Role{
				ID:   roles[0].ID,
				Code: roles[0].Code,
				Name: roles[0].Name,
			})
			db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)",
				users[i].ID, roles[0].ID)
		}
	}

	// ── 批量创建菜单（树结构，共 20 个）─────────────────
	logger.Info("批量创建菜单...")
	menus := make([]permission.Menu, 0, 20)
	// 顶级菜单（5 个）
	for i := 1; i <= 5; i++ {
		menus = append(menus, permission.Menu{
			ID:    uuid.New().String(),
			Name:  fmt.Sprintf("菜单 %d", i),
			Path:  strPtr(fmt.Sprintf("/menu%d", i)),
			Icon:  strPtr(fmt.Sprintf("icon-%d", i)),
			Order: i,
		})
	}
	db.Create(&menus)
	// 子菜单（每个顶级菜单 3 个子菜单）
	for i, parent := range menus {
		for j := 1; j <= 3; j++ {
			childID := uuid.New().String()
			db.Create(&permission.Menu{
				ID:       childID,
				Name:     fmt.Sprintf("%s - 子菜单 %d", parent.Name, j),
				Path:     strPtr(fmt.Sprintf("%s/sub%d", *parent.Path, j)),
				Icon:     strPtr(fmt.Sprintf("sub-icon-%d", j)),
				Order:    j,
				ParentID: &parent.ID,
			})
			if i >= 4 { // 只创建前 5 个顶级菜单的子菜单
				break
			}
		}
	}

	// ── 批量创建部门（树结构，共 15 个）─────────────────
	logger.Info("批量创建部门...")
	// 顶级部门（3 个）
	var topDepts []biz.Department
	for i := 1; i <= 3; i++ {
		dept := biz.Department{
			ID:          uuid.New().String(),
			Name:        fmt.Sprintf("部门 %d", i),
			Description: strPtr(fmt.Sprintf("顶级部门 %d", i)),
		}
		db.Create(&dept)
		topDepts = append(topDepts, dept)
	}
	// 子部门（每个顶级部门 4 个子部门）
	for _, parent := range topDepts {
		for j := 1; j <= 4; j++ {
			db.Create(&biz.Department{
				ID:          uuid.New().String(),
				Name:        fmt.Sprintf("%s - 子部门 %d", parent.Name, j),
				Description: strPtr(fmt.Sprintf("子部门 %d", j)),
				ParentID:    &parent.ID,
			})
		}
	}

	// ── 批量创建字典（50 个）────────────────────────────
	logger.Info("批量创建字典...")
	dicts := make([]biz.Dict, 0, 50)
	types := []string{"status", "gender", "level", "category"}
	for i := 1; i <= 50; i++ {
		dicts = append(dicts, biz.Dict{
			ID:          uuid.New().String(),
			Type:        types[i%len(types)],
			Key:         fmt.Sprintf("key_%d", i),
			Value:       fmt.Sprintf("value_%d", i),
			Description: strPtr(fmt.Sprintf("字典项 %d", i)),
		})
	}
	db.Create(&dicts)

	// ── 批量创建日志（100 个）───────────────────────────
	logger.Info("批量创建日志...")
	logs := make([]biz.Log, 0, 100)
	levels := []string{"info", "warn", "error", "debug"}
	for i := 1; i <= 100; i++ {
		logs = append(logs, biz.Log{
			ID:        uuid.New().String(),
			Level:     levels[i%len(levels)],
			Message:   fmt.Sprintf("测试日志消息 %d", i),
			Meta:      strPtr(fmt.Sprintf(`{"index": %d, "timestamp": "%s"}`, i, time.Now().Format(time.RFC3339))),
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Minute),
		})
	}
	db.Create(&logs)

	logger.Info("批量测试数据创建完成")
}

// strPtr 字符串指针辅助函数
func strPtr(s string) *string {
	return &s
}
