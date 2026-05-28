// Auth Service 入口 — gRPC 认证微服务
// 无数据库，通过 gRPC 调用 User Service，使用 Redis 管理会话
package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"

	"helloGo/internal/auth"
	"helloGo/internal/shared/config"
	"helloGo/internal/shared/health"
	"helloGo/internal/shared/logger"
	sharedredis "helloGo/internal/shared/redis"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("auth", "./configs")
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	log := logger.New(cfg.Service.Env, "auth")
	defer log.Sync()

	log.Info("Auth Service 启动中",
		zap.Int("grpcPort", cfg.Service.GRPCPort),
		zap.String("env", cfg.Service.Env),
		zap.String("userService", cfg.Services.UserAddr),
	)

	// 3. 初始化 Redis（Auth Service 不需要数据库）
	redisClient := sharedredis.New(cfg.Redis, log)
	defer redisClient.Close()

	// 3.5 启动健康检查服务（K8s 探针，端口 8080）
	var healthChecks []health.CheckFunc
	if redisClient.IsRedis() {
		healthChecks = append(healthChecks, health.RedisCheck(redisClient.Ping))
	}
	healthSrv := health.NewServer(8080, log, healthChecks...)
	healthSrv.Start()
	defer healthSrv.Stop()

	// 4. 创建 Auth Service（内部连接 User Service）
	svc, err := auth.NewService(
		redisClient,
		cfg.JWT,
		cfg.Login,
		cfg.Services.UserAddr,
		log,
	)
	if err != nil {
		log.Fatal("Auth Service 初始化失败", zap.Error(err))
	}
	defer svc.Close()

	// 5. 创建并启动 gRPC 服务器
	srv := auth.NewServer(cfg.Service.GRPCPort, svc, log)

	if err := srv.Run(); err != nil {
		log.Fatal("gRPC 服务器异常退出", zap.Error(err))
	}

	log.Info("Auth Service 已停止")
}
