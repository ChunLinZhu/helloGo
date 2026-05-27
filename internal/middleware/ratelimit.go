// Package middleware — 限流中间件
// 使用 Fiber 内置 limiter 中间件，支持 Redis 或内存存储后端
package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"go.uber.org/zap"

	appErrors "helloGo/internal/pkg/errors"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	TTL   int // 限流时间窗口（秒）
	Limit int // 窗口内最大请求数
}

// RateLimiter 创建限流中间件
// 使用 Fiber 内置 limiter，以 IP 为 key 进行限流
func RateLimiter(cfg RateLimitConfig, logger *zap.Logger) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        cfg.Limit,
		Expiration: time.Duration(cfg.TTL) * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			// 优先使用用户 ID，否则使用 IP
			if userID, ok := c.Locals("userId").(string); ok && userID != "" {
				return "rl:user:" + userID
			}
			return "rl:ip:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			logger.Warn("限流触发",
				zap.String("ip", c.IP()),
				zap.String("path", c.Path()),
			)
			return appErrors.New(appErrors.CodeHTTPError, fiber.StatusTooManyRequests, "请求过于频繁，请稍后重试")
		},
		// 跳过健康检查和静态资源
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
	})
}
