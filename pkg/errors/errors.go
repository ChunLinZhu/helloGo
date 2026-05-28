// Package errors — 框架无关的业务错误码
// 从 Phase 1 的 internal/pkg/errors 改造，移除 Fiber 依赖
// 新增 gRPC 状态码转换函数
package errors

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorCode 业务错误码
type ErrorCode string

// 错误码枚举 — 与 Phase 1 保持一致
const (
	CodeUnknown         ErrorCode = "UNKNOWN"
	CodeHTTPError       ErrorCode = "HTTP_ERROR"
	CodeInternalError   ErrorCode = "INTERNAL_ERROR"
	CodeValidationError ErrorCode = "VALIDATION_ERROR"
	CodeUnauthorized    ErrorCode = "AUTH_UNAUTHORIZED"
	CodeForbidden       ErrorCode = "AUTH_FORBIDDEN"
	CodeNotFound        ErrorCode = "NOT_FOUND"
	CodeAccountLocked   ErrorCode = "ACCOUNT_LOCKED"
	CodeInvalidPassword ErrorCode = "INVALID_PASSWORD"
	CodeUserNotFound    ErrorCode = "USER_NOT_FOUND"
	CodeUserInactive    ErrorCode = "USER_INACTIVE"
	CodeDuplicateEntry  ErrorCode = "DUPLICATE_ENTRY"
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

// AppError 业务错误
type AppError struct {
	Code       ErrorCode   `json:"code"`
	HTTPStatus int         `json:"statusCode"`
	Message    string      `json:"message"`
	Detail     interface{} `json:"detail,omitempty"`
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
	err := New(CodeValidationError, 400, detail)
	if s, ok := detail.(string); ok {
		err.Message = s
	}
	return err
}

// Unauthorized 401 未认证
func Unauthorized() *AppError {
	return New(CodeUnauthorized, 401, nil)
}

// Forbidden 403 无权限
func Forbidden() *AppError {
	return New(CodeForbidden, 403, nil)
}

// NotFound 404 资源不存在
func NotFound() *AppError {
	return New(CodeNotFound, 404, nil)
}

// Internal 500 服务器内部错误
func Internal() *AppError {
	return New(CodeInternalError, 500, nil)
}

// GetMessage 根据语言获取错误消息
func (e *AppError) GetMessage(lang string) string {
	if e.Message != "" {
		return e.Message
	}
	messages, ok := i18nMessages[lang]
	if !ok {
		messages = i18nMessages["zh-CN"]
	}
	if msg, ok := messages[e.Code]; ok {
		return msg
	}
	return i18nMessages["zh-CN"][CodeUnknown]
}

// NormalizeLang 标准化语言码（导出供 gRPC metadata 使用）
func NormalizeLang(lang string) string {
	lang = strings.TrimSpace(lang)
	lang = strings.Split(lang, ",")[0]
	lang = strings.Split(lang, ";")[0]
	lang = strings.TrimSpace(lang)

	switch strings.ToLower(lang) {
	case "en", "en-us", "en_us":
		return "en-US"
	default:
		return "zh-CN"
	}
}

// ToGRPCStatus 将 AppError 转换为 gRPC status error
func ToGRPCStatus(e *AppError) error {
	var code codes.Code
	switch e.Code {
	case CodeValidationError:
		code = codes.InvalidArgument
	case CodeUnauthorized:
		code = codes.Unauthenticated
	case CodeForbidden:
		code = codes.PermissionDenied
	case CodeNotFound, CodeUserNotFound:
		code = codes.NotFound
	case CodeDuplicateEntry:
		code = codes.AlreadyExists
	case CodeAccountLocked:
		code = codes.FailedPrecondition
	case CodeInvalidPassword:
		code = codes.InvalidArgument
	case CodeUserInactive:
		code = codes.FailedPrecondition
	default:
		code = codes.Internal
	}
	msg := e.Message
	if msg == "" {
		msg = e.GetMessage("zh-CN")
	}
	return status.Error(code, msg)
}

// FromGRPCStatus 将 gRPC status error 转换回 AppError（Gateway 层使用）
func FromGRPCStatus(err error) *AppError {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return &AppError{Code: CodeInternalError, HTTPStatus: 500, Message: err.Error()}
	}

	var appCode ErrorCode
	var httpStatus int
	switch st.Code() {
	case codes.InvalidArgument:
		appCode = CodeValidationError
		httpStatus = 400
	case codes.Unauthenticated:
		appCode = CodeUnauthorized
		httpStatus = 401
	case codes.PermissionDenied:
		appCode = CodeForbidden
		httpStatus = 403
	case codes.NotFound:
		appCode = CodeNotFound
		httpStatus = 404
	case codes.AlreadyExists:
		appCode = CodeDuplicateEntry
		httpStatus = 409
	case codes.FailedPrecondition:
		appCode = CodeAccountLocked
		httpStatus = 400
	default:
		appCode = CodeInternalError
		httpStatus = 500
	}

	return &AppError{
		Code:       appCode,
		HTTPStatus: httpStatus,
		Message:    st.Message(),
	}
}
