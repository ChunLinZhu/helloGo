// Package menu — 菜单路由处理
package menu

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	appErrors "helloGo/internal/pkg/errors"
	"helloGo/internal/pkg/response"
)

// Handler 菜单路由处理器
type Handler struct {
	service  Service
	logger   *zap.Logger
	validate *validator.Validate
}

// NewHandler 创建菜单路由处理器
func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

// GetTree 获取菜单树
//
//	@Summary		获取菜单树
//	@Description	获取所有菜单的树形结构
//	@Tags			菜单管理
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response
//	@Failure		401	{object}	response.Response
//	@Router			/menus/tree [get]
//	@Security		BearerAuth
func (h *Handler) GetTree(c *fiber.Ctx) error {
	tree, err := h.service.GetTree()
	if err != nil {
		return appErrors.Internal()
	}

	return response.SuccessOK(c, tree)
}

// GetByID 按 ID 查询菜单
//
//	@Summary		按 ID 查询菜单
//	@Description	根据菜单 ID 获取菜单详细信息
//	@Tags			菜单管理
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"菜单 ID"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		401	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Router			/menus/{id} [get]
//	@Security		BearerAuth
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少菜单 ID")
	}

	menu, err := h.service.GetByID(id)
	if err != nil {
		return err
	}

	return response.SuccessOK(c, menu)
}

// Create 创建菜单
//
//	@Summary		创建菜单
//	@Description	创建新菜单，支持父子关系
//	@Tags			菜单管理
//	@Accept			json
//	@Produce		json
//	@Param			body	body		menu.CreateMenuRequest	true	"菜单信息"
//	@Success		201		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Router			/menus [post]
//	@Security		BearerAuth
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateMenuRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	menu, err := h.service.Create(&req)
	if err != nil {
		return err
	}

	return response.SuccessCreated(c, menu)
}

// Update 更新菜单
//
//	@Summary		更新菜单
//	@Description	部分更新菜单信息（name/path/icon/sort/parentId）
//	@Tags			菜单管理
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"菜单 ID"
//	@Param			body	body		menu.UpdateMenuRequest	true	"更新内容"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Failure		404		{object}	response.Response
//	@Router			/menus/{id} [patch]
//	@Security		BearerAuth
func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少菜单 ID")
	}

	var req UpdateMenuRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	menu, err := h.service.Update(id, &req)
	if err != nil {
		return err
	}

	return response.SuccessOK(c, menu)
}

// Delete 删除菜单
//
//	@Summary		删除菜单
//	@Description	根据 ID 删除菜单（有子菜单时不可删除）
//	@Tags			菜单管理
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"菜单 ID"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		401	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Router			/menus/{id} [delete]
//	@Security		BearerAuth
func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少菜单 ID")
	}

	if err := h.service.Delete(id); err != nil {
		return err
	}

	return response.SuccessOK(c, fiber.Map{"message": "删除成功"})
}
