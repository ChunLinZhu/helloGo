// Package handler — User 路由处理器
package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	commonv1 "helloGo/gen/go/common/v1"
	userv1 "helloGo/gen/go/user/v1"
	"helloGo/internal/pkg/response"
)

// UserHandler 用户路由处理器
type UserHandler struct {
	client userv1.UserServiceClient
}

// NewUserHandler 创建 UserHandler
func NewUserHandler(client userv1.UserServiceClient) *UserHandler {
	return &UserHandler{client: client}
}

// ListUsers 查询用户列表
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	keyword := c.Query("keyword")

	resp, err := h.client.ListUsers(c.Context(), &userv1.ListUsersRequest{
		Pagination: &commonv1.PaginationRequest{
			Page:  int32(page),
			Limit: int32(limit),
		},
		Keyword: keyword,
	})
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessPage(c, resp.Users, resp.Pagination.Total,
		int(resp.Pagination.Page), int(resp.Pagination.Limit))
}

// GetUser 查询单个用户
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.client.GetUser(c.Context(), &userv1.GetUserRequest{
		Id: id,
	})
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessOK(c, resp.User)
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req userv1.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"code":       "VALIDATION_ERROR",
			"statusCode": 400,
			"message":    "请求体格式错误",
			"path":       c.Path(),
		})
	}

	resp, err := h.client.CreateUser(c.Context(), &req)
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessCreated(c, resp.User)
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")

	var req userv1.UpdateUserRequest
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
	resp, err := h.client.UpdateUser(c.Context(), &req)
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessOK(c, resp.User)
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	_, err := h.client.DeleteUser(c.Context(), &userv1.DeleteUserRequest{
		Id: id,
	})
	if err != nil {
		return HandleGRPCError(c, err)
	}

	return response.SuccessNoContent(c)
}
