// Package errors — 业务错误码定义与 i18n 国际化支持
// 错误码枚举 + 中英文消息映射，通过 X-Lang 或 Accept-Language 头选择语言
package errors

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ErrorCode 业务错误码
type ErrorCode string

// 错误码枚举 — 与 helloNest 的 error-codes.ts 保持一致
const (
	CodeUnknown          ErrorCode = "UNKNOWN"            // 未知错误
	CodeHTTPError        ErrorCode = "HTTP_ERROR"         // 请求错误
	CodeInternalError    ErrorCode = "INTERNAL_ERROR"     // 服务器内部错误
	CodeValidationError  ErrorCode = "VALIDATION_ERROR"   // 参数校验失败
	CodeUnauthorized     ErrorCode = "AUTH_UNAUTHORIZED"  // 未认证或认证失效
	CodeForbidden        ErrorCode = "AUTH_FORBIDDEN"     // 无权限访问
	CodeNotFound         ErrorCode = "NOT_FOUND"          // 资源不存在
	CodeAccountLocked    ErrorCode = "ACCOUNT_LOCKED"     // 账户已锁定
	CodeInvalidPassword  ErrorCode = "INVALID_PASSWORD"   // 密码错误
	CodeUserNotFound     ErrorCode = "USER_NOT_FOUND"     // 用户不存在
	CodeUserInactive     ErrorCode = "USER_INACTIVE"      // 用户已禁用
	CodeDuplicateEntry   ErrorCode = "DUPLICATE_ENTRY"    // 数据重复
)

// i18nMessages 各语言的错误码消息映射
var i18nMessages = map[string]map[ErrorCode]string{
	"zh-CN": {
		CodeUnknown:         "未知错误",
		CodeHTTPError:       "请求错误",
		CodeInternalError:   "服务器内部错误",
		CodeValidationError: "参数校验失败",
		CodeUnauthorized:    "未认证或认证失效",
		CodeForbidden:       "无权限访问",
		CodeNotFound:        "资源不存在",
		CodeAccountLocked:   "账户已锁定，请稍后重试",
		CodeInvalidPassword: "用户名或密码错误",
		CodeUserNotFound:    "用户不存在",
		CodeUserInactive:    "用户已禁用",
		CodeDuplicateEntry:  "数据已存在",
	},
	"en-US": {
		CodeUnknown:         "Unknown error",
		CodeHTTPError:       "Request error",
		CodeInternalError:   "Internal server error",
		CodeValidationError: "Validation failed",
		CodeUnauthorized:    "Unauthorized",
		CodeForbidden:       "Forbidden",
		CodeNotFound:        "Resource not found",
		CodeAccountLocked:   "Account locked, please try again later",
		CodeInvalidPassword: "Invalid username or password",
		CodeUserNotFound:    "User not found",
		CodeUserInactive:    "User is inactive",
		CodeDuplicateEntry:  "Duplicate entry",
	},
}

// AppError 业务错误，包含错误码、HTTP 状态码和详细信息
type AppError struct {
	Code       ErrorCode   `json:"code"`
	HTTPStatus int         `json:"statusCode"`
	Message    string      `json:"message"`
	Detail     interface{} `json:"detail,omitempty"` // 可选的错误详情（如字段校验失败列表）
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// New 创建业务错误
func New(code ErrorCode, httpStatus int, detail interface{}) *AppError {
	return &AppError{
		Code:       code,
		HTTPStatus: httpStatus,
		Detail:     detail,
	}
}

// BadRequest 400 参数错误
func BadRequest(detail interface{}) *AppError {
	return New(CodeValidationError, fiber.StatusBadRequest, detail)
}

// Unauthorized 401 未认证
func Unauthorized() *AppError {
	return New(CodeUnauthorized, fiber.StatusUnauthorized, nil)
}

// Forbidden 403 无权限
func Forbidden() *AppError {
	return New(CodeForbidden, fiber.StatusForbidden, nil)
}

// NotFound 404 资源不存在
func NotFound() *AppError {
	return New(CodeNotFound, fiber.StatusNotFound, nil)
}

// Internal 500 服务器内部错误
func Internal() *AppError {
	return New(CodeInternalError, fiber.StatusInternalServerError, nil)
}

// GetMessage 根据语言获取错误消息
func (e *AppError) GetMessage(lang string) string {
	if e.Message != "" {
		return e.Message
	}
	messages, ok := i18nMessages[lang]
	if !ok {
		messages = i18nMessages["zh-CN"] // 默认中文
	}
	if msg, ok := messages[e.Code]; ok {
		return msg
	}
	return i18nMessages["zh-CN"][CodeUnknown]
}

// GetLang 从请求中提取语言偏好
// 优先级：X-Lang 头 > Accept-Language 头 > 默认 zh-CN
func GetLang(c *fiber.Ctx) string {
	// 1. 优先检查 X-Lang 头
	if lang := c.Get("X-Lang"); lang != "" {
		return normalizeLang(lang)
	}
	// 2. 检查 Accept-Language 头
	if lang := c.Get("Accept-Language"); lang != "" {
		return normalizeLang(lang)
	}
	// 3. 默认中文
	return "zh-CN"
}

// normalizeLang 标准化语言代码
func normalizeLang(lang string) string {
	lang = strings.TrimSpace(lang)
	lang = strings.Split(lang, ",")[0] // 取第一个语言
	lang = strings.Split(lang, ";")[0] // 去掉权重
	lang = strings.TrimSpace(lang)

	switch strings.ToLower(lang) {
	case "en", "en-us", "en_us":
		return "en-US"
	default:
		return "zh-CN"
	}
}

// HTTPStatusToErrorCode 将 HTTP 状态码映射为业务错误码
func HTTPStatusToErrorCode(status int) ErrorCode {
	switch {
	case status == fiber.StatusBadRequest:
		return CodeValidationError
	case status == fiber.StatusUnauthorized:
		return CodeUnauthorized
	case status == fiber.StatusForbidden:
		return CodeForbidden
	case status == fiber.StatusNotFound:
		return CodeNotFound
	case status >= 500:
		return CodeInternalError
	default:
		return CodeHTTPError
	}
}
