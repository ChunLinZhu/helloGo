// Package handler — Auth 路由处理器
package handler

import (
	"github.com/gofiber/fiber/v2"

	authv1 "helloGo/gen/go/auth/v1"
	"helloGo/internal/pkg/response"
)

// AuthHandler 认证路由处理器
type AuthHandler struct {
	client authv1.AuthServiceClient
}

// NewAuthHandler 创建 AuthHandler
func NewAuthHandler(client authv1.AuthServiceClient) *AuthHandler {
	return &AuthHandler{client: client}
}

// Login 用户登录
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req authv1.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"code":       "VALIDATION_ERROR",
			"statusCode": 400,
			"message":    "请求体格式错误",
			"path":       c.Path(),
		})
	}

	resp, err := h.client.Login(c.Context(), &req)
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessOK(c, resp)
}

// RefreshToken 刷新令牌
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req authv1.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"code":       "VALIDATION_ERROR",
			"statusCode": 400,
			"message":    "请求体格式错误",
			"path":       c.Path(),
		})
	}

	resp, err := h.client.RefreshToken(c.Context(), &req)
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessOK(c, resp)
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req authv1.LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"code":       "VALIDATION_ERROR",
			"statusCode": 400,
			"message":    "请求体格式错误",
			"path":       c.Path(),
		})
	}

	_, err := h.client.Logout(c.Context(), &req)
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessOK(c, nil)
}

// RequestPasswordReset 请求密码重置
func (h *AuthHandler) RequestPasswordReset(c *fiber.Ctx) error {
	var req authv1.RequestPasswordResetRequest
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"code":       "VALIDATION_ERROR",
			"statusCode": 400,
			"message":    "请求体格式错误",
			"path":       c.Path(),
		})
	}

	resp, err := h.client.RequestPasswordReset(c.Context(), &req)
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessOK(c, resp)
}

// ConfirmPasswordReset 确认密码重置
func (h *AuthHandler) ConfirmPasswordReset(c *fiber.Ctx) error {
	var req authv1.ConfirmPasswordResetRequest
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"code":       "VALIDATION_ERROR",
			"statusCode": 400,
			"message":    "请求体格式错误",
			"path":       c.Path(),
		})
	}

	_, err := h.client.ConfirmPasswordReset(c.Context(), &req)
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessOK(c, nil)
}

// UnlockAccount 解锁账户
func (h *AuthHandler) UnlockAccount(c *fiber.Ctx) error {
	var req authv1.UnlockAccountRequest
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"code":       "VALIDATION_ERROR",
			"statusCode": 400,
			"message":    "请求体格式错误",
			"path":       c.Path(),
		})
	}

	_, err := h.client.UnlockAccount(c.Context(), &req)
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessOK(c, nil)
}
