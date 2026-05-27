// Package middleware — 审计日志中间件
// 将每个请求记录持久化到 logs 数据库表，meta 字段包含请求详情
package middleware

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	logModel "helloGo/internal/module/log"
)

// AuditMeta 审计日志元数据
type AuditMeta struct {
	TraceID  string  `json:"traceId"`
	Method   string  `json:"method"`
	Path     string  `json:"path"`
	Status   int     `json:"status"`
	Duration float64 `json:"duration"` // 毫秒
	UserID   string  `json:"userId,omitempty"`
	Username string  `json:"username,omitempty"`
	IP       string  `json:"ip"`
}

// AuditLogger 审计日志中间件
// 将每个请求写入 logs 数据库表，同时保持 Zap 日志输出
func AuditLogger(logger *zap.Logger, logRepo logModel.Repository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// 执行后续处理
		err := c.Next()

		// 计算耗时
		duration := time.Since(start)

		// 获取请求追踪 ID
		requestID, _ := c.Locals("requestId").(string)

		// 获取用户信息（JWT 中间件注入后生效）
		userID, _ := c.Locals("userId").(string)
		username, _ := c.Locals("username").(string)

		status := c.Response().StatusCode()

		// 构建 meta JSON
		meta := AuditMeta{
			TraceID:  requestID,
			Method:   c.Method(),
			Path:     c.Path(),
			Status:   status,
			Duration: float64(duration.Microseconds()) / 1000.0,
			UserID:   userID,
			Username: username,
			IP:       c.IP(),
		}

		metaJSON, _ := json.Marshal(meta)
		metaStr := string(metaJSON)

		// 根据状态码确定日志级别
		level := "info"
		switch {
		case status >= 500:
			level = "error"
		case status >= 400:
			level = "warn"
		}

		// 异步写入数据库（不阻塞响应）
		go func() {
			logEntry := &logModel.Log{
				Level:   level,
				Message: meta.Method + " " + meta.Path,
				Meta:    &metaStr,
			}
			if createErr := logRepo.Create(logEntry); createErr != nil {
				// 审计写入失败不影响请求，仅记录 Zap 错误
				logger.Error("审计日志写入失败",
					zap.Error(createErr),
					zap.String("path", meta.Path),
				)
			}
		}()

		return err
	}
}
