// Package handler — Gateway HTTP 处理器
package handler

import (
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleGRPCError 将 gRPC 错误转换为 HTTP 响应
func HandleGRPCError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	// 尝试从错误中提取 gRPC 状态
	st, ok := status.FromError(err)
	if !ok {
		// 非 gRPC 错误，返回 500
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"code":       "INTERNAL_ERROR",
			"statusCode": 500,
			"message":    err.Error(),
			"path":       c.Path(),
		})
	}

	// 根据 gRPC 状态码映射 HTTP 状态码
	var httpStatus int
	var code string

	switch st.Code() {
	case codes.OK:
		return nil
	case codes.InvalidArgument:
		httpStatus = fiber.StatusBadRequest
		code = "INVALID_ARGUMENT"
	case codes.NotFound:
		httpStatus = fiber.StatusNotFound
		code = "NOT_FOUND"
	case codes.AlreadyExists:
		httpStatus = fiber.StatusConflict
		code = "ALREADY_EXISTS"
	case codes.PermissionDenied:
		httpStatus = fiber.StatusForbidden
		code = "FORBIDDEN"
	case codes.Unauthenticated:
		httpStatus = fiber.StatusUnauthorized
		code = "UNAUTHORIZED"
	case codes.ResourceExhausted:
		httpStatus = fiber.StatusTooManyRequests
		code = "RATE_LIMITED"
	case codes.Unavailable:
		httpStatus = fiber.StatusServiceUnavailable
		code = "SERVICE_UNAVAILABLE"
	case codes.Internal:
		httpStatus = fiber.StatusInternalServerError
		code = "INTERNAL_ERROR"
	default:
		httpStatus = fiber.StatusInternalServerError
		code = "UNKNOWN_ERROR"
	}

	c.Status(httpStatus)
	return c.JSON(fiber.Map{
		"code":       code,
		"statusCode": httpStatus,
		"message":    st.Message(),
		"path":       c.Path(),
	})
}
