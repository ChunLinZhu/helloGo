// Package role — 角色路由处理
package role

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	appErrors "helloGo/internal/pkg/errors"
	"helloGo/internal/pkg/pagination"
	"helloGo/internal/pkg/response"
)

// Handler 角色路由处理器
type Handler struct {
	service  Service
	logger   *zap.Logger
	validate *validator.Validate
}

// NewHandler 创建角色路由处理器
func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

// List 分页查询角色列表
//
//	@Summary		分页查询角色列表
//	@Description	支持关键词搜索（code/name）和分页
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int	false	"页码（默认1）"
//	@Param			limit	query		int	false	"每页数量（默认10）"
//	@Param			keyword	query		string	false	"搜索关键词"
//	@Success		200		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Router			/roles [get]
//	@Security		BearerAuth
func (h *Handler) List(c *fiber.Ctx) error {
	page := pagination.GetPagination(c)
	keyword := c.Query("keyword", "")

	roles, total, err := h.service.List(page, keyword)
	if err != nil {
		return appErrors.Internal()
	}

	return response.SuccessPage(c, roles, total, page.Page, page.Limit)
}

// GetByID 按 ID 查询角色
//
//	@Summary		按 ID 查询角色
//	@Description	根据角色 ID 获取角色详细信息（含权限列表）
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"角色 ID"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		401	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Router			/roles/{id} [get]
//	@Security		BearerAuth
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少角色 ID")
	}

	role, err := h.service.GetByID(id)
	if err != nil {
		return err
	}

	return response.SuccessOK(c, role)
}

// Create 创建角色
//
//	@Summary		创建角色
//	@Description	创建新角色，code 必须唯一
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Param			body	body		role.CreateRoleRequest	true	"角色信息"
//	@Success		201		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Router			/roles [post]
//	@Security		BearerAuth
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	role, err := h.service.Create(&req)
	if err != nil {
		return err
	}

	return response.SuccessCreated(c, role)
}

// Update 更新角色
//
//	@Summary		更新角色
//	@Description	部分更新角色信息（name/description）
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"角色 ID"
//	@Param			body	body		role.UpdateRoleRequest	true	"更新内容"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Failure		404		{object}	response.Response
//	@Router			/roles/{id} [patch]
//	@Security		BearerAuth
func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少角色 ID")
	}

	var req UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	role, err := h.service.Update(id, &req)
	if err != nil {
		return err
	}

	return response.SuccessOK(c, role)
}

// Delete 删除角色
//
//	@Summary		删除角色
//	@Description	根据 ID 删除角色（admin 角色不可删除）
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"角色 ID"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		401	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Router			/roles/{id} [delete]
//	@Security		BearerAuth
func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少角色 ID")
	}

	if err := h.service.Delete(id); err != nil {
		return err
	}

	return response.SuccessOK(c, fiber.Map{"message": "删除成功"})
}

// AssignPermissions 为角色分配权限
//
//	@Summary		为角色分配权限
//	@Description	全量替换角色的权限列表（先清除旧权限，再关联新权限）
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"角色 ID"
//	@Param			body	body		role.AssignPermissionsRequest	true	"权限 ID 列表"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Failure		404		{object}	response.Response
//	@Router			/roles/{id}/permissions [post]
//	@Security		BearerAuth
func (h *Handler) AssignPermissions(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少角色 ID")
	}

	var req AssignPermissionsRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	if err := h.service.AssignPermissions(id, req.PermissionIDs); err != nil {
		return err
	}

	return response.SuccessOK(c, fiber.Map{"message": "权限分配成功"})
}
