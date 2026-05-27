// Package role — 角色业务逻辑
package role

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"helloGo/internal/pkg/pagination"
)

// Service 角色业务接口
type Service interface {
	Create(req *CreateRoleRequest) (*RoleResponse, error)
	GetByID(id string) (*RoleResponse, error)
	Update(id string, req *UpdateRoleRequest) (*RoleResponse, error)
	Delete(id string) error
	List(page pagination.Pagination, keyword string) ([]RoleResponse, int64, error)
	AssignPermissions(roleID string, permissionIDs []string) error
}

// service 角色业务实现
type service struct {
	repo   Repository
	logger *zap.Logger
}

// NewService 创建角色业务层
func NewService(repo Repository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// Create 创建角色
func (s *service) Create(req *CreateRoleRequest) (*RoleResponse, error) {
	// 检查代码是否已存在
	existing, _ := s.repo.FindByCode(req.Code)
	if existing != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "角色代码已存在")
	}

	role := &Role{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.repo.Create(role); err != nil {
		return nil, fmt.Errorf("创建角色失败: %w", err)
	}

	s.logger.Info("角色创建成功",
		zap.String("roleId", role.ID),
		zap.String("code", role.Code),
	)

	return ToRoleResponse(role), nil
}

// GetByID 按 ID 查询角色
func (s *service) GetByID(id string) (*RoleResponse, error) {
	role, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "角色不存在")
	}
	return ToRoleResponse(role), nil
}

// Update 更新角色
func (s *service) Update(id string, req *UpdateRoleRequest) (*RoleResponse, error) {
	role, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "角色不存在")
	}

	// 更新字段
	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Description != nil {
		role.Description = req.Description
	}

	if err := s.repo.Update(role); err != nil {
		return nil, fmt.Errorf("更新角色失败: %w", err)
	}

	s.logger.Info("角色更新成功",
		zap.String("roleId", role.ID),
		zap.String("code", role.Code),
	)

	return ToRoleResponse(role), nil
}

// Delete 删除角色
func (s *service) Delete(id string) error {
	role, err := s.repo.FindByID(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "角色不存在")
	}

	// 不允许删除 admin 角色
	if role.Code == "admin" {
		return fiber.NewError(fiber.StatusBadRequest, "不允许删除 admin 角色")
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("删除角色失败: %w", err)
	}

	s.logger.Info("角色删除成功",
		zap.String("roleId", id),
		zap.String("code", role.Code),
	)

	return nil
}

// List 分页查询角色列表
func (s *service) List(page pagination.Pagination, keyword string) ([]RoleResponse, int64, error) {
	roles, total, err := s.repo.List(page, keyword)
	if err != nil {
		return nil, 0, fmt.Errorf("查询角色列表失败: %w", err)
	}

	responses := make([]RoleResponse, len(roles))
	for i, r := range roles {
		responses[i] = *ToRoleResponse(&r)
	}

	return responses, total, nil
}

// AssignPermissions 为角色分配权限
func (s *service) AssignPermissions(roleID string, permissionIDs []string) error {
	_, err := s.repo.FindByID(roleID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "角色不存在")
	}

	if err := s.repo.AssociatePermissions(roleID, permissionIDs); err != nil {
		return fmt.Errorf("分配权限失败: %w", err)
	}

	s.logger.Info("角色权限分配成功",
		zap.String("roleId", roleID),
		zap.Int("permissionCount", len(permissionIDs)),
	)

	return nil
}
