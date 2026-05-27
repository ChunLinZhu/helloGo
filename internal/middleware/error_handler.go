// Package middleware — 全局错误处理中间件
// 统一捕获 handler 返回的 error，转换为结构化 JSON 响应
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	appErrors "helloGo/internal/pkg/errors"
	"helloGo/internal/pkg/response"
)

// ErrorHandler 全局错误处理中间件
// 放在路由链最末尾，捕获所有未处理的 error
func ErrorHandler(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if err == nil {
			return nil
		}

		lang := appErrors.GetLang(c)

		// ── 业务错误（AppError） ──────────────────────────
		if appErr, ok := err.(*appErrors.AppError); ok {
			c.Status(appErr.HTTPStatus)
			msg := appErr.GetMessage(lang)
			logger.Warn("业务错误",
				zap.String("code", string(appErr.Code)),
				zap.Int("status", appErr.HTTPStatus),
				zap.String("message", msg),
				zap.String("path", c.Path()),
			)
			return response.Success(c, string(appErr.Code), msg, appErr.Detail)
		}

		// ── Fiber 内置错误（如 404、405） ────────────────
		if fiberErr, ok := err.(*fiber.Error); ok {
			code := appErrors.HTTPStatusToErrorCode(fiberErr.Code)
			c.Status(fiberErr.Code)

			// Fiber 404 使用自定义消息
			msg := appErrors.New(code, fiberErr.Code, nil).GetMessage(lang)
			if fiberErr.Code == fiber.StatusNotFound {
				msg = appErrors.New(appErrors.CodeNotFound, 404, nil).GetMessage(lang)
			}

			logger.Warn("HTTP 错误",
				zap.String("code", string(code)),
				zap.Int("status", fiberErr.Code),
				zap.String("message", fiberErr.Message),
				zap.String("path", c.Path()),
			)
			return response.Success(c, string(code), msg, nil)
		}

		// ── 未知错误 — 返回 500 ───────────────────────────
		logger.Error("未预期的错误",
			zap.Error(err),
			zap.String("path", c.Path()),
		)
		c.Status(fiber.StatusInternalServerError)
		appErr := appErrors.Internal()
		return response.Success(c, string(appErr.Code), appErr.GetMessage(lang), nil)
	}
}
