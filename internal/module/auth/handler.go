// Package auth — 认证路由处理函数
// 提供登录、刷新、登出、密码重置等 HTTP 端点
package auth

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	appErrors "helloGo/internal/pkg/errors"
	"helloGo/internal/pkg/response"
)

// AuthHandler 认证路由处理器
type AuthHandler struct {
	service  AuthService
	logger   *zap.Logger
	validate *validator.Validate
}

// NewAuthHandler 创建认证路由处理器
func NewAuthHandler(service AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

// RegisterRoutes 注册认证路由
func (h *AuthHandler) RegisterRoutes(api fiber.Router) {
	auth := api.Group("/auth")

	// 公开路由（无需认证）
	auth.Post("/login", h.Login)
	auth.Post("/refresh", h.Refresh)
	auth.Post("/password/request-reset", h.RequestPasswordReset)
	auth.Post("/password/reset", h.ResetPassword)
	auth.Post("/unlock", h.Unlock)

	// 需要认证的路由（在 main.go 中通过 jwtGuard 中间件保护）
	auth.Post("/logout", h.Logout)
}

// Login 用户登录
//
//	@Summary		用户登录
//	@Description	验证用户名密码，返回 JWT access token 和 refresh token
//	@Tags			认证
//	@Accept			json
//	@Produce		json
//	@Param			body	body		auth.LoginRequest	true	"登录信息"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	resp, err := h.service.Login(c.UserContext(), req.Username, req.Password)
	if err != nil {
		return appErrors.New(appErrors.CodeUnauthorized, fiber.StatusUnauthorized, err.Error())
	}

	return response.SuccessOK(c, resp)
}

// Refresh 刷新令牌
//
//	@Summary		刷新令牌
//	@Description	使用 refresh token 获取新的 access token
//	@Tags			认证
//	@Accept			json
//	@Produce		json
//	@Param			body	body		auth.RefreshRequest	true	"刷新信息"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Router			/auth/refresh [post]
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	resp, err := h.service.Refresh(c.UserContext(), req.RefreshToken, req.SessionID)
	if err != nil {
		return appErrors.New(appErrors.CodeUnauthorized, fiber.StatusUnauthorized, err.Error())
	}

	return response.SuccessOK(c, resp)
}

// Logout 登出
//
//	@Summary		登出
//	@Description	使当前会话失效，从 Redis 中删除 session
//	@Tags			认证
//	@Accept			json
//	@Produce		json
//	@Param			body	body		auth.LogoutRequest	true	"登出信息"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Router			/auth/logout [post]
//	@Security		BearerAuth
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	// 从 JWT 中间件注入的 context 获取 userID
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return appErrors.Unauthorized()
	}

	if err := h.service.Logout(c.UserContext(), userID, req.SessionID); err != nil {
		return appErrors.Internal()
	}

	return response.SuccessOK(c, fiber.Map{"message": "登出成功"})
}

// RequestPasswordReset 请求密码重置
//
//	@Summary		请求密码重置
//	@Description	生成密码重置 token（实际场景中会发送邮件）
//	@Tags			认证
//	@Accept			json
//	@Produce		json
//	@Param			body	body		auth.RequestPasswordResetRequest	true	"用户名"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		404		{object}	response.Response
//	@Router			/auth/password/request-reset [post]
func (h *AuthHandler) RequestPasswordReset(c *fiber.Ctx) error {
	var req RequestPasswordResetRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	resp, err := h.service.RequestPasswordReset(c.UserContext(), req.Username)
	if err != nil {
		return appErrors.New(appErrors.CodeUserNotFound, fiber.StatusNotFound, err.Error())
	}

	return response.SuccessOK(c, resp)
}

// ResetPassword 重置密码
//
//	@Summary		重置密码
//	@Description	使用重置 token 设置新密码
//	@Tags			认证
//	@Accept			json
//	@Produce		json
//	@Param			body	body		auth.ResetPasswordRequest	true	"重置信息"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Router			/auth/password/reset [post]
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	if err := h.service.ResetPassword(c.UserContext(), req.Username, req.NewPassword, req.Token); err != nil {
		return appErrors.New(appErrors.CodeValidationError, fiber.StatusBadRequest, err.Error())
	}

	return response.SuccessOK(c, fiber.Map{"message": "密码重置成功"})
}

// Unlock 解锁账户
//
//	@Summary		解锁账户
//	@Description	手动解锁因连续登录失败而被锁定的账户
//	@Tags			认证
//	@Accept			json
//	@Produce		json
//	@Param			body	body		auth.UnlockRequest	true	"用户名"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Router			/auth/unlock [post]
func (h *AuthHandler) Unlock(c *fiber.Ctx) error {
	var req UnlockRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	if err := h.service.UnlockUser(c.UserContext(), req.Username); err != nil {
		return appErrors.Internal()
	}

	return response.SuccessOK(c, fiber.Map{"message": "账户解锁成功"})
}
