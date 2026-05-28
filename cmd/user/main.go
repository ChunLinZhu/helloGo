// User Service 入口 — gRPC 微服务
package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"

	"helloGo/internal/shared/config"
	"helloGo/internal/shared/database"
	"helloGo/internal/shared/health"
	"helloGo/internal/shared/logger"
	"helloGo/internal/user"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("user", "./configs")
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	log := logger.New(cfg.Service.Env, "user")
	defer log.Sync()

	log.Info("User Service 启动中",
		zap.Int("grpcPort", cfg.Service.GRPCPort),
		zap.String("env", cfg.Service.Env),
	)

	// 3. 初始化数据库
	db, err := database.Connect(cfg.Database, cfg.Service.Env, log)
	if err != nil {
		log.Fatal("数据库初始化失败", zap.Error(err))
	}

	// 3.5 启动健康检查服务（K8s 探针，端口 8080）
	sqlDB, _ := db.DB()
	healthSrv := health.NewServer(8080, log, health.DBCheck(sqlDB))
	healthSrv.Start()
	defer healthSrv.Stop()

	// 4. 自动迁移 User Service 模型
	if err := database.AutoMigrate(db, log, &user.User{}, &user.Role{}); err != nil {
		log.Fatal("数据库迁移失败", zap.Error(err))
	}

	// 5. 构建依赖链：Repository → Service → Server
	repo := user.NewRepository(db)
	svc := user.NewService(repo, log)
	srv := user.NewServer(cfg.Service.GRPCPort, svc, log)

	// 6. 启动 gRPC 服务器（阻塞直到退出信号）
	if err := srv.Run(); err != nil {
		log.Fatal("gRPC 服务器异常退出", zap.Error(err))
	}

	log.Info("User Service 已停止")
}
