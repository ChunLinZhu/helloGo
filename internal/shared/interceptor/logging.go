// Package interceptor — gRPC 共享拦截器
package interceptor

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerLogger 返回 gRPC 一元服务器日志拦截器
// 记录每个请求的 method、status code、duration
func UnaryServerLogger(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)
		st, _ := status.FromError(err)

		logger.Info("gRPC 请求",
			zap.String("method", info.FullMethod),
			zap.String("code", st.Code().String()),
			zap.Duration("duration", duration),
		)
		return resp, err
	}
}
