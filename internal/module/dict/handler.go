// Package dict — 字典路由处理
package dict

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	appErrors "helloGo/internal/pkg/errors"
	"helloGo/internal/pkg/pagination"
	"helloGo/internal/pkg/response"
)

// Handler 字典路由处理器
type Handler struct {
	service  Service
	logger   *zap.Logger
	validate *validator.Validate
}

// NewHandler 创建字典路由处理器
func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

// List 分页查询字典列表
//
//	@Summary		分页查询字典列表
//	@Description	支持关键词搜索（key/label）和按类型筛选
//	@Tags			字典管理
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int		false	"页码（默认1）"
//	@Param			limit	query		int		false	"每页数量（默认10）"
//	@Param			keyword	query		string	false	"搜索关键词"
//	@Param			type	query		string	false	"字典类型"
//	@Success		200		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Router			/dicts [get]
//	@Security		BearerAuth
func (h *Handler) List(c *fiber.Ctx) error {
	page := pagination.GetPagination(c)
	keyword := c.Query("keyword", "")
	dictType := c.Query("type", "")

	dicts, total, err := h.service.List(page, keyword, dictType)
	if err != nil {
		return appErrors.Internal()
	}

	return response.SuccessPage(c, dicts, total, page.Page, page.Limit)
}

// GetByID 按 ID 查询字典项
//
//	@Summary		按 ID 查询字典项
//	@Description	根据字典项 ID 获取字典项详细信息
//	@Tags			字典管理
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"字典项 ID"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		401	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Router			/dicts/{id} [get]
//	@Security		BearerAuth
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少字典项 ID")
	}

	d, err := h.service.GetByID(id)
	if err != nil {
		return err
	}

	return response.SuccessOK(c, d)
}

// Create 创建字典项
//
//	@Summary		创建字典项
//	@Description	创建新字典项，key + type 组合必须唯一
//	@Tags			字典管理
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dict.CreateDictRequest	true	"字典项信息"
//	@Success		201		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Router			/dicts [post]
//	@Security		BearerAuth
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateDictRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	d, err := h.service.Create(&req)
	if err != nil {
		return err
	}

	return response.SuccessCreated(c, d)
}

// Update 更新字典项
//
//	@Summary		更新字典项
//	@Description	部分更新字典项信息（label/value/type/sort）
//	@Tags			字典管理
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"字典项 ID"
//	@Param			body	body		dict.UpdateDictRequest	true	"更新内容"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Failure		404		{object}	response.Response
//	@Router			/dicts/{id} [patch]
//	@Security		BearerAuth
func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少字典项 ID")
	}

	var req UpdateDictRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	d, err := h.service.Update(id, &req)
	if err != nil {
		return err
	}

	return response.SuccessOK(c, d)
}

// Delete 删除字典项
//
//	@Summary		删除字典项
//	@Description	根据 ID 删除字典项
//	@Tags			字典管理
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"字典项 ID"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		401	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Router			/dicts/{id} [delete]
//	@Security		BearerAuth
func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少字典项 ID")
	}

	if err := h.service.Delete(id); err != nil {
		return err
	}

	return response.SuccessOK(c, fiber.Map{"message": "删除成功"})
}
