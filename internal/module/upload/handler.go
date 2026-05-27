// Package upload — 上传路由处理
package upload

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	appErrors "helloGo/internal/pkg/errors"
	"helloGo/internal/pkg/pagination"
	"helloGo/internal/pkg/response"
)

// Handler 上传路由处理器
type Handler struct {
	service  Service
	logger   *zap.Logger
	validate *validator.Validate
}

// NewHandler 创建上传路由处理器
func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

// Upload 上传文件
//
//	@Summary		上传文件
//	@Description	单文件上传，使用 multipart/form-data，字段名 "file"
//	@Tags			文件上传
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"上传文件"
//	@Success		201		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Router			/uploads [post]
//	@Security		BearerAuth
func (h *Handler) Upload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return appErrors.BadRequest("缺少文件字段 (file)")
	}

	result, err := h.service.Upload(file)
	if err != nil {
		return err
	}

	return response.SuccessCreated(c, result)
}

// UploadChunk 上传分片
//
//	@Summary		上传分片
//	@Description	分片上传，使用 multipart/form-data，字段名 "file"，附带分片参数
//	@Tags			文件上传
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file		formData	file	true	"分片文件"
//	@Param			uploadId	formData	string	true	"上传 ID"
//	@Param			chunkIndex	formData	int		true	"分片索引"
//	@Success		200			{object}	response.Response
//	@Failure		400			{object}	response.Response
//	@Failure		401			{object}	response.Response
//	@Router			/uploads/chunk [post]
//	@Security		BearerAuth
func (h *Handler) UploadChunk(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return appErrors.BadRequest("缺少文件字段 (file)")
	}

	var req ChunkUploadRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求参数解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	if err := h.service.UploadChunk(&req, file); err != nil {
		return err
	}

	return response.SuccessOK(c, fiber.Map{"message": "分片上传成功"})
}

// MergeChunks 合并分片
//
//	@Summary		合并分片
//	@Description	所有分片上传完成后，调用此接口合并为完整文件
//	@Tags			文件上传
//	@Accept			json
//	@Produce		json
//	@Param			body	body		upload.MergeRequest	true	"合并参数"
//	@Success		201		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Router			/uploads/merge [post]
//	@Security		BearerAuth
func (h *Handler) MergeChunks(c *fiber.Ctx) error {
	var req MergeRequest
	if err := c.BodyParser(&req); err != nil {
		return appErrors.BadRequest("请求体解析失败")
	}

	if err := h.validate.Struct(req); err != nil {
		return appErrors.BadRequest(err.Error())
	}

	result, err := h.service.MergeChunks(&req)
	if err != nil {
		return err
	}

	return response.SuccessCreated(c, result)
}

// List 分页查询上传记录
//
//	@Summary		分页查询上传记录
//	@Description	获取已上传文件列表，支持分页
//	@Tags			文件上传
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int	false	"页码（默认1）"
//	@Param			limit	query		int	false	"每页数量（默认10）"
//	@Success		200		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Router			/uploads [get]
//	@Security		BearerAuth
func (h *Handler) List(c *fiber.Ctx) error {
	page := pagination.GetPagination(c)

	uploads, total, err := h.service.List(page)
	if err != nil {
		return appErrors.Internal()
	}

	return response.SuccessPage(c, uploads, total, page.Page, page.Limit)
}

// GetByID 按 ID 查询上传记录
//
//	@Summary		按 ID 查询上传记录
//	@Description	根据上传记录 ID 获取文件详细信息
//	@Tags			文件上传
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"上传记录 ID"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		401	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Router			/uploads/{id} [get]
//	@Security		BearerAuth
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少上传记录 ID")
	}

	upload, err := h.service.GetByID(id)
	if err != nil {
		return err
	}

	return response.SuccessOK(c, upload)
}

// Delete 删除上传记录
//
//	@Summary		删除上传记录
//	@Description	根据 ID 删除上传记录及对应文件
//	@Tags			文件上传
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"上传记录 ID"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		401	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Router			/uploads/{id} [delete]
//	@Security		BearerAuth
func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return appErrors.BadRequest("缺少上传记录 ID")
	}

	if err := h.service.Delete(id); err != nil {
		return err
	}

	return response.SuccessOK(c, fiber.Map{"message": "删除成功"})
}
