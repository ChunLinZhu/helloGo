// Package handler — Permission 路由处理器
package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	commonv1 "helloGo/gen/go/common/v1"
	permissionv1 "helloGo/gen/go/permission/v1"
	"helloGo/internal/pkg/response"
)

// PermissionHandler 权限路由处理器
type PermissionHandler struct {
	client permissionv1.PermissionServiceClient
}

// NewPermissionHandler 创建 PermissionHandler
func NewPermissionHandler(client permissionv1.PermissionServiceClient) *PermissionHandler {
	return &PermissionHandler{client: client}
}

// ListRoles 查询角色列表
func (h *PermissionHandler) ListRoles(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	keyword := c.Query("keyword")

	resp, err := h.client.ListRoles(c.Context(), &permissionv1.ListRolesRequest{
		Pagination: &commonv1.PaginationRequest{
			Page:  int32(page),
			Limit: int32(limit),
		},
		Keyword: keyword,
	})
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessPage(c, resp.Roles, resp.Pagination.Total,
		int(resp.Pagination.Page), int(resp.Pagination.Limit))
}

// CreateRole 创建角色
func (h *PermissionHandler) CreateRole(c *fiber.Ctx) error {
	var req permissionv1.CreateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"code":       "VALIDATION_ERROR",
			"statusCode": 400,
			"message":    "请求体格式错误",
			"path":       c.Path(),
		})
	}

	resp, err := h.client.CreateRole(c.Context(), &req)
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessCreated(c, resp.Role)
}

// AddPermissionToRole 为角色添加权限
func (h *PermissionHandler) AddPermissionToRole(c *fiber.Ctx) error {
	roleId := c.Params("id")

	var req permissionv1.AddPermissionToRoleRequest
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"code":       "VALIDATION_ERROR",
			"statusCode": 400,
			"message":    "请求体格式错误",
			"path":       c.Path(),
		})
	}

	req.RoleId = roleId
	resp, err := h.client.AddPermissionToRole(c.Context(), &req)
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessOK(c, resp.Role)
}

// ListPermissions 查询权限列表
func (h *PermissionHandler) ListPermissions(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	keyword := c.Query("keyword")

	resp, err := h.client.ListPermissions(c.Context(), &permissionv1.ListPermissionsRequest{
		Pagination: &commonv1.PaginationRequest{
			Page:  int32(page),
			Limit: int32(limit),
		},
		Keyword: keyword,
	})
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessPage(c, resp.Permissions, resp.Pagination.Total,
		int(resp.Pagination.Page), int(resp.Pagination.Limit))
}

// CreatePermission 创建权限
func (h *PermissionHandler) CreatePermission(c *fiber.Ctx) error {
	var req permissionv1.CreatePermissionRequest
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"code":       "VALIDATION_ERROR",
			"statusCode": 400,
			"message":    "请求体格式错误",
			"path":       c.Path(),
		})
	}

	resp, err := h.client.CreatePermission(c.Context(), &req)
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessCreated(c, resp)
}

// UpdatePermission 更新权限
func (h *PermissionHandler) UpdatePermission(c *fiber.Ctx) error {
	id := c.Params("id")

	var req permissionv1.UpdatePermissionRequest
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"code":       "VALIDATION_ERROR",
			"statusCode": 400,
			"message":    "请求体格式错误",
			"path":       c.Path(),
		})
	}

	req.Id = id
	resp, err := h.client.UpdatePermission(c.Context(), &req)
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessOK(c, resp)
}

// DeletePermission 删除权限
func (h *PermissionHandler) DeletePermission(c *fiber.Ctx) error {
	id := c.Params("id")

	_, err := h.client.DeletePermission(c.Context(), &permissionv1.DeletePermissionRequest{
		Id: id,
	})
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessNoContent(c)
}

// ListMenus 查询菜单树
func (h *PermissionHandler) ListMenus(c *fiber.Ctx) error {
	resp, err := h.client.ListMenus(c.Context(), &permissionv1.ListMenusRequest{})
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessOK(c, resp.Menus)
}
