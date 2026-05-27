// Package middleware — CORS 跨域配置中间件
// 根据 CORS_ORIGINS 配置允许的跨域来源，支持 credentials
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORS 跨域中间件
// origins 为逗号分隔的允许来源列表，"*" 表示允许所有
func CORS(origins string) fiber.Handler {
	// 清理空白字符
	allowed := strings.TrimSpace(origins)
	if allowed == "" {
		allowed = "*"
	}

	// Fiber 不允许 AllowCredentials=true 且 AllowOrigins="*"
	// 当配置为 "*" 时自动禁用 credentials
	allowCreds := allowed != "*"

	return cors.New(cors.Config{
		AllowOrigins:     allowed,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Lang,X-Trace-Id,X-CSRF-Token",
		ExposeHeaders:    "X-Trace-Id,X-CSRF-Token",
		AllowCredentials: allowCreds,
		MaxAge:           86400, // 预检请求缓存 24 小时
	})
}
