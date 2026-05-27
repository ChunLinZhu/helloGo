// Package log — 日志路由处理
package log

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	appErrors "helloGo/internal/pkg/errors"
	"helloGo/internal/pkg/pagination"
	"helloGo/internal/pkg/response"
)

// Handler 日志路由处理器
type Handler struct {
	service  Service
	logger   *zap.Logger
	validate *validator.Validate
}

// NewHandler 创建日志路由处理器
func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

// List 分页查询日志列表
//
//	@Summary		分页查询日志列表
//	@Description	支持按日志级别筛选和分页
//	@Tags			系统日志
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int		false	"页码（默认1）"
//	@Param			limit	query		int		false	"每页数量（默认10）"
//	@Param			level	query		string	false	"日志级别（info/warn/error）"
//	@Success		200		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Router			/logs [get]
//	@Security		BearerAuth
func (h *Handler) List(c *fiber.Ctx) error {
	page := pagination.GetPagination(c)
	level := c.Query("level", "")

	logs, total, err := h.service.List(page, level)
	if err != nil {
		return appErrors.Internal()
	}

	return response.SuccessPage(c, logs, total, page.Page, page.Limit)
}

// GetByID 按 ID 查询日志
//
//	@Summary		按 ID 查询日志
//	@Description	根据日志 ID 获取日志详细信息
//	@Tags			系统日志
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"日志 ID"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		401	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Router			/logs/{id} [get]
//	@Security		BearerAuth
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少日志 ID")
	}

	l, err := h.service.GetByID(id)
	if err != nil {
		return err
	}

	return response.SuccessOK(c, l)
}

// Create 创建日志
//
//	@Summary		创建日志
//	@Description	手动创建一条日志记录（通常由审计中间件自动写入）
//	@Tags			系统日志
//	@Accept			json
//	@Produce		json
//	@Param			body	body		log.CreateLogRequest	true	"日志信息"
//	@Success		201		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Router			/logs [post]
//	@Security		BearerAuth
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateLogRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	l, err := h.service.Create(&req)
	if err != nil {
		return err
	}

	return response.SuccessCreated(c, l)
}
