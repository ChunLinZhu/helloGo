// Package user — 用户路由处理
package user

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	appErrors "helloGo/internal/pkg/errors"
	"helloGo/internal/pkg/pagination"
	"helloGo/internal/pkg/response"
)

// Handler 用户路由处理器
type Handler struct {
	service  Service
	logger   *zap.Logger
	validate *validator.Validate
}

// NewHandler 创建用户路由处理器
func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

// List 分页查询用户列表
//
//	@Summary		分页查询用户列表
//	@Description	支持关键词搜索（username/email/phone）和分页
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int	false	"页码（默认1）"
//	@Param			limit	query		int	false	"每页数量（默认10）"
//	@Param			keyword	query		string	false	"搜索关键词"
//	@Success		200		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Router			/users [get]
//	@Security		BearerAuth
func (h *Handler) List(c *fiber.Ctx) error {
	page := pagination.GetPagination(c)
	keyword := c.Query("keyword", "")

	users, total, err := h.service.List(page, keyword)
	if err != nil {
		return appErrors.Internal()
	}

	return response.SuccessPage(c, users, total, page.Page, page.Limit)
}

// GetByID 按 ID 查询用户
//
//	@Summary		按 ID 查询用户
//	@Description	根据用户 ID 获取用户详细信息
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"用户 ID"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		401	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Router			/users/{id} [get]
//	@Security		BearerAuth
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少用户 ID")
	}

	user, err := h.service.GetByID(id)
	if err != nil {
		return err
	}

	return response.SuccessOK(c, user)
}

// Create 创建用户
//
//	@Summary		创建用户
//	@Description	创建新用户，密码将被 bcrypt 加密存储
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Param			body	body		user.CreateUserRequest	true	"用户信息"
//	@Success		201		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Router			/users [post]
//	@Security		BearerAuth
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	user, err := h.service.Create(&req)
	if err != nil {
		return err
	}

	return response.SuccessCreated(c, user)
}

// Update 更新用户
//
//	@Summary		更新用户
//	@Description	部分更新用户信息，支持更新 email/phone/isActive/roleIds
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"用户 ID"
//	@Param			body	body		user.UpdateUserRequest	true	"更新内容"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Failure		404		{object}	response.Response
//	@Router			/users/{id} [patch]
//	@Security		BearerAuth
func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少用户 ID")
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	user, err := h.service.Update(id, &req)
	if err != nil {
		return err
	}

	return response.SuccessOK(c, user)
}

// Delete 删除用户
//
//	@Summary		删除用户
//	@Description	根据 ID 删除用户
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"用户 ID"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		401	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Router			/users/{id} [delete]
//	@Security		BearerAuth
func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少用户 ID")
	}

	if err := h.service.Delete(id); err != nil {
		return err
	}

	return response.SuccessOK(c, fiber.Map{"message": "删除成功"})
}
