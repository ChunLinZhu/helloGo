// Package permission — Permission Service gRPC 服务器
package permission

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	permissionv1 "helloGo/gen/go/permission/v1"
	"helloGo/internal/shared/interceptor"
)

// Server 管理 Permission Service gRPC 服务器生命周期
type Server struct {
	grpcServer *grpc.Server
	port       int
	logger     *zap.Logger
}

// NewServer 创建 Permission Service gRPC 服务器
func NewServer(port int, svc *Service, logger *zap.Logger) *Server {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.UnaryServerRecovery(logger),
			interceptor.UnaryServerLogger(logger),
		),
	)

	// 注册 PermissionService
	permissionv1.RegisterPermissionServiceServer(grpcServer, svc)

	// 注册反射服务（grpcurl 需要）
	reflection.Register(grpcServer)

	return &Server{
		grpcServer: grpcServer,
		port:       port,
		logger:     logger,
	}
}

// Run 启动 gRPC 服务器，阻塞直到收到退出信号
func (s *Server) Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("监听端口 %d 失败: %w", s.port, err)
	}

	s.logger.Info("Permission Service gRPC 服务器已启动", zap.Int("port", s.port))

	// 优雅关闭：监听 SIGINT/SIGTERM
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit
		s.logger.Info("收到退出信号，正在关闭...", zap.String("signal", sig.String()))
		s.grpcServer.GracefulStop()
	}()

	return s.grpcServer.Serve(lis)
}
