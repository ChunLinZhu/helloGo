// Biz Service 入口 — gRPC 业务微服务
// 管理部门、字典、日志、上传文件
package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"

	"helloGo/internal/biz"
	"helloGo/internal/shared/config"
	"helloGo/internal/shared/database"
	"helloGo/internal/shared/logger"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("biz", "./configs")
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	log := logger.New(cfg.Service.Env, "biz")
	defer log.Sync()

	log.Info("Biz Service 启动中",
		zap.Int("grpcPort", cfg.Service.GRPCPort),
		zap.String("env", cfg.Service.Env),
	)

	// 3. 初始化数据库
	db, err := database.Connect(cfg.Database, cfg.Service.Env, log)
	if err != nil {
		log.Fatal("数据库初始化失败", zap.Error(err))
	}

	// 4. 自动迁移 Biz Service 模型
	if err := database.AutoMigrate(db, log,
		&biz.Department{},
		&biz.Dict{},
		&biz.Log{},
		&biz.Upload{},
	); err != nil {
		log.Fatal("数据库迁移失败", zap.Error(err))
	}

	// 5. 构建依赖链：Repository → Service → Server
	repo := biz.NewRepository(db)
	svc := biz.NewService(repo, log)
	srv := biz.NewServer(cfg.Service.GRPCPort, svc, log)

	// 6. 启动 gRPC 服务器（阻塞直到退出信号）
	if err := srv.Run(); err != nil {
		log.Fatal("gRPC 服务器异常退出", zap.Error(err))
	}

	log.Info("Biz Service 已停止")
}
