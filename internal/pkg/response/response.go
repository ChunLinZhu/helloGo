// Package response — 统一 JSON 响应格式
// 所有 API 响应均使用此包封装，保持前后端接口一致性
package response

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// Response 统一响应结构体
type Response struct {
	Code       string      `json:"code"`       // 业务状态码，如 "OK"、"VALIDATION_ERROR"
	StatusCode int         `json:"statusCode"` // HTTP 状态码
	Message    string      `json:"message"`    // 人类可读的消息（支持 i18n）
	Data       interface{} `json:"data"`       // 响应数据（成功时为业务数据，失败时可为 null）
	Path       string      `json:"path"`       // 请求路径
	Timestamp  string      `json:"timestamp"`  // ISO 8601 时间戳
	RequestID  string      `json:"requestId"`  // 请求追踪 ID（来自 X-Trace-Id）
}

// Success 返回成功响应
func Success(c *fiber.Ctx, code string, message string, data interface{}) error {
	return c.JSON(Response{
		Code:       code,
		StatusCode: c.Response().StatusCode(),
		Message:    message,
		Data:       data,
		Path:       c.Path(),
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		RequestID:  getRequestID(c),
	})
}

// SuccessOK 返回 200 成功响应（最常用）
func SuccessOK(c *fiber.Ctx, data interface{}) error {
	c.Status(fiber.StatusOK)
	return Success(c, "OK", "success", data)
}

// SuccessCreated 返回 201 创建成功响应
func SuccessCreated(c *fiber.Ctx, data interface{}) error {
	c.Status(fiber.StatusCreated)
	return Success(c, "OK", "created", data)
}

// SuccessNoContent 返回 204 无内容响应
func SuccessNoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// PageData 分页响应数据
type PageData struct {
	Items      interface{} `json:"items"`      // 数据列表
	Total      int64       `json:"total"`      // 总记录数
	Page       int         `json:"page"`       // 当前页码
	Limit      int         `json:"limit"`      // 每页条数
	TotalPages int         `json:"totalPages"` // 总页数
}

// SuccessPage 返回分页成功响应
func SuccessPage(c *fiber.Ctx, items interface{}, total int64, page, limit int) error {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	return SuccessOK(c, PageData{
		Items:      items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// getRequestID 从 Fiber context 中获取请求追踪 ID
func getRequestID(c *fiber.Ctx) string {
	if id, ok := c.Locals("requestId").(string); ok {
		return id
	}
	return ""
}
