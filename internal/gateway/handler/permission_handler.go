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

// GetRole 按 ID 查询角色
func (h *PermissionHandler) GetRole(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.client.GetRole(c.Context(), &permissionv1.GetRoleRequest{
		Id: id,
	})
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessOK(c, resp.Role)
}

// DeleteRole 删除角色
func (h *PermissionHandler) DeleteRole(c *fiber.Ctx) error {
	id := c.Params("id")

	_, err := h.client.DeleteRole(c.Context(), &permissionv1.DeleteRoleRequest{
		Id: id,
	})
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessNoContent(c)
}

// AddPermissionToRole 为角色添加权限
// 前端发送 { permissionCodes: [...] }，转为 gRPC 的 permission_keys
func (h *PermissionHandler) AddPermissionToRole(c *fiber.Ctx) error {
	roleId := c.Params("id")

	// 前端发送 { permissionCodes: ["user:create", "user:view", ...] }
	var body struct {
		PermissionCodes []string `json:"permissionCodes"`
	}
	if err := c.BodyParser(&body); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"code":       "VALIDATION_ERROR",
			"statusCode": 400,
			"message":    "请求体格式错误",
			"path":       c.Path(),
		})
	}

	resp, err := h.client.AddPermissionToRole(c.Context(), &permissionv1.AddPermissionToRoleRequest{
		RoleId:         roleId,
		PermissionKeys: body.PermissionCodes,
	})
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
// 前端发送 { code, name, description }，映射为 gRPC 的 { key, description }
func (h *PermissionHandler) CreatePermission(c *fiber.Ctx) error {
	// 前端字段: code → key, name → 合并到 description
	var body struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&body); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"code":       "VALIDATION_ERROR",
			"statusCode": 400,
			"message":    "请求体格式错误",
			"path":       c.Path(),
		})
	}

	// code 映射为 key，name 作为 description 的一部分
	desc := body.Description
	if desc == "" && body.Name != "" {
		desc = body.Name
	}

	resp, err := h.client.CreatePermission(c.Context(), &permissionv1.CreatePermissionRequest{
		Key:         body.Code,
		Description: desc,
	})
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessCreated(c, resp)
}

// GetPermission 按 ID 查询权限
func (h *PermissionHandler) GetPermission(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.client.GetPermission(c.Context(), &permissionv1.GetPermissionRequest{
		Id: id,
	})
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessOK(c, resp)
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
