// Package middleware — X-Trace-Id 请求追踪中间件
// 从请求头读取或生成 UUID 作为 traceId，注入 context 和响应头
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Trace 请求追踪中间件
// 优先使用客户端传入的 X-Trace-Id，否则生成新的 UUID
func Trace() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 优先读取客户端传入的 traceId
		traceID := c.Get("X-Trace-Id")
		if traceID == "" {
			// 也兼容 x-request-id 头（helloNest 使用的名称）
			traceID = c.Get("x-request-id")
		}
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// 注入 Fiber context（后续中间件和 handler 可通过 c.Locals 获取）
		c.Locals("requestId", traceID)

		// 设置响应头，方便前端 / 客户端关联日志
		c.Set("X-Trace-Id", traceID)

		return c.Next()
	}
}
