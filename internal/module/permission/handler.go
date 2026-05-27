// Package permission — 权限路由处理
package permission

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	appErrors "helloGo/internal/pkg/errors"
	"helloGo/internal/pkg/pagination"
	"helloGo/internal/pkg/response"
)

// Handler 权限路由处理器
type Handler struct {
	service  Service
	logger   *zap.Logger
	validate *validator.Validate
}

// NewHandler 创建权限路由处理器
func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

// List 分页查询权限列表
//
//	@Summary		分页查询权限列表
//	@Description	支持关键词搜索（key/description）和分页
//	@Tags			权限管理
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int	false	"页码（默认1）"
//	@Param			limit	query		int	false	"每页数量（默认10）"
//	@Param			keyword	query		string	false	"搜索关键词"
//	@Success		200		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Router			/permissions [get]
//	@Security		BearerAuth
func (h *Handler) List(c *fiber.Ctx) error {
	page := pagination.GetPagination(c)
	keyword := c.Query("keyword", "")

	perms, total, err := h.service.List(page, keyword)
	if err != nil {
		return appErrors.Internal()
	}

	return response.SuccessPage(c, perms, total, page.Page, page.Limit)
}

// GetByID 按 ID 查询权限
//
//	@Summary		按 ID 查询权限
//	@Description	根据权限 ID 获取权限详细信息
//	@Tags			权限管理
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"权限 ID"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		401	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Router			/permissions/{id} [get]
//	@Security		BearerAuth
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少权限 ID")
	}

	perm, err := h.service.GetByID(id)
	if err != nil {
		return err
	}

	return response.SuccessOK(c, perm)
}

// Create 创建权限
//
//	@Summary		创建权限
//	@Description	创建新权限，key 必须唯一
//	@Tags			权限管理
//	@Accept			json
//	@Produce		json
//	@Param			body	body		permission.CreatePermissionRequest	true	"权限信息"
//	@Success		201		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Router			/permissions [post]
//	@Security		BearerAuth
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreatePermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	perm, err := h.service.Create(&req)
	if err != nil {
		return err
	}

	return response.SuccessCreated(c, perm)
}

// Update 更新权限
//
//	@Summary		更新权限
//	@Description	部分更新权限信息（key/description/roleId）
//	@Tags			权限管理
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"权限 ID"
//	@Param			body	body		permission.UpdatePermissionRequest	true	"更新内容"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Failure		404		{object}	response.Response
//	@Router			/permissions/{id} [patch]
//	@Security		BearerAuth
func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少权限 ID")
	}

	var req UpdatePermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	perm, err := h.service.Update(id, &req)
	if err != nil {
		return err
	}

	return response.SuccessOK(c, perm)
}

// Delete 删除权限
//
//	@Summary		删除权限
//	@Description	根据 ID 删除权限
//	@Tags			权限管理
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"权限 ID"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		401	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Router			/permissions/{id} [delete]
//	@Security		BearerAuth
func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少权限 ID")
	}

	if err := h.service.Delete(id); err != nil {
		return err
	}

	return response.SuccessOK(c, fiber.Map{"message": "删除成功"})
}
