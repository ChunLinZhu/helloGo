// Package middleware — Panic 恢复中间件
// 捕获 handler 中的 panic，返回 500 结构化响应，避免服务崩溃
package middleware

import (
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"helloGo/internal/pkg/errors"
	"helloGo/internal/pkg/response"
)

// Recovery panic 恢复中间件
func Recovery(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				// 记录 panic 堆栈到日志
				logger.Error("服务 panic",
					zap.Any("error", r),
					zap.String("stack", string(debug.Stack())),
					zap.String("path", c.Path()),
					zap.String("method", c.Method()),
				)

				// 返回 500 结构化响应
				c.Status(fiber.StatusInternalServerError)
				appErr := errors.Internal()
				lang := errors.GetLang(c)
				_ = response.Success(c, string(appErr.Code), appErr.GetMessage(lang), nil)
			}
		}()
		return c.Next()
	}
}
