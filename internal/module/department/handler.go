// Package department — 部门路由处理
package department

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	appErrors "helloGo/internal/pkg/errors"
	"helloGo/internal/pkg/response"
)

// Handler 部门路由处理器
type Handler struct {
	service  Service
	logger   *zap.Logger
	validate *validator.Validate
}

// NewHandler 创建部门路由处理器
func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

// GetTree 获取部门树
//
//	@Summary		获取部门树
//	@Description	获取所有部门的树形结构
//	@Tags			部门管理
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response
//	@Failure		401	{object}	response.Response
//	@Router			/departments/tree [get]
//	@Security		BearerAuth
func (h *Handler) GetTree(c *fiber.Ctx) error {
	tree, err := h.service.GetTree()
	if err != nil {
		return appErrors.Internal()
	}

	return response.SuccessOK(c, tree)
}

// GetByID 按 ID 查询部门
//
//	@Summary		按 ID 查询部门
//	@Description	根据部门 ID 获取部门详细信息
//	@Tags			部门管理
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"部门 ID"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		401	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Router			/departments/{id} [get]
//	@Security		BearerAuth
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少部门 ID")
	}

	dept, err := h.service.GetByID(id)
	if err != nil {
		return err
	}

	return response.SuccessOK(c, dept)
}

// Create 创建部门
//
//	@Summary		创建部门
//	@Description	创建新部门，支持父子关系
//	@Tags			部门管理
//	@Accept			json
//	@Produce		json
//	@Param			body	body		department.CreateDepartmentRequest	true	"部门信息"
//	@Success		201		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Router			/departments [post]
//	@Security		BearerAuth
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateDepartmentRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	dept, err := h.service.Create(&req)
	if err != nil {
		return err
	}

	return response.SuccessCreated(c, dept)
}

// Update 更新部门
//
//	@Summary		更新部门
//	@Description	部分更新部门信息（name/sort/parentId）
//	@Tags			部门管理
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"部门 ID"
//	@Param			body	body		department.UpdateDepartmentRequest	true	"更新内容"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Failure		404		{object}	response.Response
//	@Router			/departments/{id} [patch]
//	@Security		BearerAuth
func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少部门 ID")
	}

	var req UpdateDepartmentRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	dept, err := h.service.Update(id, &req)
	if err != nil {
		return err
	}

	return response.SuccessOK(c, dept)
}

// Delete 删除部门
//
//	@Summary		删除部门
//	@Description	根据 ID 删除部门（有子部门或关联用户时不可删除）
//	@Tags			部门管理
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"部门 ID"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		401	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Router			/departments/{id} [delete]
//	@Security		BearerAuth
func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少部门 ID")
	}

	if err := h.service.Delete(id); err != nil {
		return err
	}

	return response.SuccessOK(c, fiber.Map{"message": "删除成功"})
}
