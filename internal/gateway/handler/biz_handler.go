// Package handler — Biz 路由处理器
package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	bizv1 "helloGo/gen/go/biz/v1"
	commonv1 "helloGo/gen/go/common/v1"
	"helloGo/internal/pkg/response"
)

// BizHandler 业务路由处理器
type BizHandler struct {
	client bizv1.BizServiceClient
}

// NewBizHandler 创建 BizHandler
func NewBizHandler(client bizv1.BizServiceClient) *BizHandler {
	return &BizHandler{client: client}
}

// ListDepartments 查询部门树
func (h *BizHandler) ListDepartments(c *fiber.Ctx) error {
	resp, err := h.client.ListDepartments(c.Context(), &bizv1.ListDepartmentsRequest{})
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessOK(c, resp.Departments)
}

// CreateDepartment 创建部门
func (h *BizHandler) CreateDepartment(c *fiber.Ctx) error {
	var req bizv1.CreateDepartmentRequest
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"code":       "VALIDATION_ERROR",
			"statusCode": 400,
			"message":    "请求体格式错误",
			"path":       c.Path(),
		})
	}

	resp, err := h.client.CreateDepartment(c.Context(), &req)
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessCreated(c, resp)
}

// ListDicts 查询字典列表
func (h *BizHandler) ListDicts(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	keyword := c.Query("keyword")
	dictType := c.Query("type")

	resp, err := h.client.ListDicts(c.Context(), &bizv1.ListDictsRequest{
		Pagination: &commonv1.PaginationRequest{
			Page:  int32(page),
			Limit: int32(limit),
		},
		Keyword: keyword,
		Type:    dictType,
	})
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessPage(c, resp.Dicts, resp.Pagination.Total,
		int(resp.Pagination.Page), int(resp.Pagination.Limit))
}

// CreateDict 创建字典
func (h *BizHandler) CreateDict(c *fiber.Ctx) error {
	var req bizv1.CreateDictRequest
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"code":       "VALIDATION_ERROR",
			"statusCode": 400,
			"message":    "请求体格式错误",
			"path":       c.Path(),
		})
	}

	resp, err := h.client.CreateDict(c.Context(), &req)
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessCreated(c, resp)
}

// ListLogs 查询日志列表
func (h *BizHandler) ListLogs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	level := c.Query("level")

	resp, err := h.client.ListLogs(c.Context(), &bizv1.ListLogsRequest{
		Pagination: &commonv1.PaginationRequest{
			Page:  int32(page),
			Limit: int32(limit),
		},
		Level: level,
	})
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessPage(c, resp.Logs, resp.Pagination.Total,
		int(resp.Pagination.Page), int(resp.Pagination.Limit))
}

// ListUploads 查询上传文件列表
func (h *BizHandler) ListUploads(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	resp, err := h.client.ListUploads(c.Context(), &bizv1.ListUploadsRequest{
		Pagination: &commonv1.PaginationRequest{
			Page:  int32(page),
			Limit: int32(limit),
		},
	})
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessPage(c, resp.Uploads, resp.Pagination.Total,
		int(resp.Pagination.Page), int(resp.Pagination.Limit))
}

// DeleteUpload 删除上传文件
func (h *BizHandler) DeleteUpload(c *fiber.Ctx) error {
	id := c.Params("id")

	_, err := h.client.DeleteUpload(c.Context(), &bizv1.DeleteUploadRequest{
		Id: id,
	})
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessNoContent(c)
}

// UploadFile 普通文件上传（存根 — 微服务架构下文件上传需独立设计）
func (h *BizHandler) UploadFile(c *fiber.Ctx) error {
	c.Status(fiber.StatusNotImplemented)
	return c.JSON(fiber.Map{
		"code":       "NOT_IMPLEMENTED",
		"statusCode": 501,
		"message":    "文件上传功能在微服务架构下需要独立设计，暂不支持",
		"path":       c.Path(),
	})
}

// UploadChunk 分片上传（存根）
func (h *BizHandler) UploadChunk(c *fiber.Ctx) error {
	c.Status(fiber.StatusNotImplemented)
	return c.JSON(fiber.Map{
		"code":       "NOT_IMPLEMENTED",
		"statusCode": 501,
		"message":    "分片上传功能在微服务架构下需要独立设计，暂不支持",
		"path":       c.Path(),
	})
}

// MergeChunks 合并分片（存根）
func (h *BizHandler) MergeChunks(c *fiber.Ctx) error {
	c.Status(fiber.StatusNotImplemented)
	return c.JSON(fiber.Map{
		"code":       "NOT_IMPLEMENTED",
		"statusCode": 501,
		"message":    "分片合并功能在微服务架构下需要独立设计，暂不支持",
		"path":       c.Path(),
	})
}
