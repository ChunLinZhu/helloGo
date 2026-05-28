// API Gateway 入口 — HTTP 网关
// 将 HTTP/JSON 请求转发到后端 gRPC 微服务
package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authv1 "helloGo/gen/go/auth/v1"
	bizv1 "helloGo/gen/go/biz/v1"
	permissionv1 "helloGo/gen/go/permission/v1"
	userv1 "helloGo/gen/go/user/v1"
	"helloGo/internal/gateway"
	"helloGo/internal/shared/config"
	"helloGo/internal/shared/logger"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("gateway", "./configs")
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	log := logger.New(cfg.Service.Env, "gateway")
	defer log.Sync()

	log.Info("API Gateway 启动中",
		zap.Int("httpPort", cfg.Service.HTTPPort),
		zap.String("env", cfg.Service.Env),
	)

	// 3. 连接到后端微服务
	log.Info("连接后端微服务...")

	// User Service
	userConn, err := grpc.NewClient(
		cfg.Services.UserAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal("连接 User Service 失败", zap.Error(err))
	}
	defer userConn.Close()
	log.Info("已连接 User Service", zap.String("addr", cfg.Services.UserAddr))

	// Auth Service
	authConn, err := grpc.NewClient(
		cfg.Services.AuthAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal("连接 Auth Service 失败", zap.Error(err))
	}
	defer authConn.Close()
	log.Info("已连接 Auth Service", zap.String("addr", cfg.Services.AuthAddr))

	// Permission Service
	permConn, err := grpc.NewClient(
		cfg.Services.PermissionAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal("连接 Permission Service 失败", zap.Error(err))
	}
	defer permConn.Close()
	log.Info("已连接 Permission Service", zap.String("addr", cfg.Services.PermissionAddr))

	// Biz Service
	bizConn, err := grpc.NewClient(
		cfg.Services.BizAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal("连接 Biz Service 失败", zap.Error(err))
	}
	defer bizConn.Close()
	log.Info("已连接 Biz Service", zap.String("addr", cfg.Services.BizAddr))

	// 4. 创建 gRPC Client
	conns := &gateway.Connections{
		User:       userv1.NewUserServiceClient(userConn),
		Auth:       authv1.NewAuthServiceClient(authConn),
		Permission: permissionv1.NewPermissionServiceClient(permConn),
		Biz:        bizv1.NewBizServiceClient(bizConn),
	}

	// 5. 创建并启动 Gateway（CORS 从配置文件读取）
	srv := gateway.New(cfg.Service.HTTPPort, conns, cfg.CorsOrigins, log)
	if err := srv.Start(); err != nil {
		log.Fatal("Gateway 启动失败", zap.Error(err))
	}
}
