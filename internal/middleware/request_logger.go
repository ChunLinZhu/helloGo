// Package middleware — 请求审计日志中间件
// 记录每个请求的 method, path, status, duration, traceId 到 Zap 日志
package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RequestLogger 请求审计日志中间件
func RequestLogger(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// 执行后续处理
		err := c.Next()

		// 计算请求耗时
		duration := time.Since(start)

		// 获取请求追踪 ID
		requestID, _ := c.Locals("requestId").(string)

		// 获取用户信息（JWT 中间件注入后生效，Phase 1 暂时为空）
		userID, _ := c.Locals("userId").(string)
		username, _ := c.Locals("username").(string)

		// 结构化日志
		fields := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("duration", duration),
			zap.String("traceId", requestID),
			zap.String("ip", c.IP()),
		}

		// 有用户信息时附加
		if userID != "" {
			fields = append(fields, zap.String("userId", userID))
			fields = append(fields, zap.String("username", username))
		}

		// 根据状态码选择日志级别
		status := c.Response().StatusCode()
		switch {
		case status >= 500:
			logger.Error("请求完成", fields...)
		case status >= 400:
			logger.Warn("请求完成", fields...)
		default:
			logger.Info("请求完成", fields...)
		}

		return err
	}
}
