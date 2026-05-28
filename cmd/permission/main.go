// Permission Service 入口 — gRPC 权限微服务
// 管理角色、权限、菜单，提供权限校验和 Redis 缓存
package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"

	"helloGo/internal/permission"
	"helloGo/internal/shared/config"
	"helloGo/internal/shared/database"
	"helloGo/internal/shared/health"
	"helloGo/internal/shared/logger"
	sharedredis "helloGo/internal/shared/redis"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("permission", "./configs")
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	log := logger.New(cfg.Service.Env, "permission")
	defer log.Sync()

	log.Info("Permission Service 启动中",
		zap.Int("grpcPort", cfg.Service.GRPCPort),
		zap.String("env", cfg.Service.Env),
	)

	// 3. 初始化数据库
	db, err := database.Connect(cfg.Database, cfg.Service.Env, log)
	if err != nil {
		log.Fatal("数据库初始化失败", zap.Error(err))
	}

	// 4. 自动迁移 Permission Service 模型
	if err := database.AutoMigrate(db, log,
		&permission.Role{},
		&permission.Permission{},
		&permission.Menu{},
	); err != nil {
		log.Fatal("数据库迁移失败", zap.Error(err))
	}

	// 5. 初始化 Redis（权限缓存）
	redisClient := sharedredis.New(cfg.Redis, log)
	defer redisClient.Close()

	// 5.5 启动健康检查服务（K8s 探针，端口 8080）
	sqlDB, _ := db.DB()
	healthChecks := []health.CheckFunc{health.DBCheck(sqlDB)}
	if redisClient.IsRedis() {
		healthChecks = append(healthChecks, health.RedisCheck(redisClient.Ping))
	}
	healthSrv := health.NewServer(8080, log, healthChecks...)
	healthSrv.Start()
	defer healthSrv.Stop()

	// 6. 构建依赖链：Repository → Service → Server
	repo := permission.NewRepository(db)
	svc := permission.NewService(repo, redisClient, log)
	srv := permission.NewServer(cfg.Service.GRPCPort, svc, log)

	// 7. 启动 gRPC 服务器（阻塞直到退出信号）
	if err := srv.Run(); err != nil {
		log.Fatal("gRPC 服务器异常退出", zap.Error(err))
	}

	log.Info("Permission Service 已停止")
}
