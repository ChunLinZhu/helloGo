// Package permission — 权限业务逻辑
package permission

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"helloGo/internal/pkg/pagination"
)

// Service 权限业务接口
type Service interface {
	Create(req *CreatePermissionRequest) (*PermissionResponse, error)
	GetByID(id string) (*PermissionResponse, error)
	Update(id string, req *UpdatePermissionRequest) (*PermissionResponse, error)
	Delete(id string) error
	List(page pagination.Pagination, keyword string) ([]PermissionResponse, int64, error)
}

// service 权限业务实现
type service struct {
	repo   Repository
	logger *zap.Logger
}

// NewService 创建权限业务层
func NewService(repo Repository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// Create 创建权限
func (s *service) Create(req *CreatePermissionRequest) (*PermissionResponse, error) {
	// 检查 key 是否已存在
	existing, _ := s.repo.FindByKey(req.Key)
	if existing != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "权限 key 已存在")
	}

	perm := &Permission{
		Key:         req.Key,
		Description: req.Description,
		RoleID:      req.RoleID,
	}

	if err := s.repo.Create(perm); err != nil {
		return nil, fmt.Errorf("创建权限失败: %w", err)
	}

	s.logger.Info("权限创建成功",
		zap.String("permissionId", perm.ID),
		zap.String("key", perm.Key),
	)

	return ToPermissionResponse(perm), nil
}

// GetByID 按 ID 查询权限
func (s *service) GetByID(id string) (*PermissionResponse, error) {
	perm, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "权限不存在")
	}
	return ToPermissionResponse(perm), nil
}

// Update 更新权限
func (s *service) Update(id string, req *UpdatePermissionRequest) (*PermissionResponse, error) {
	perm, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "权限不存在")
	}

	// 如果要更新 key，检查是否与其他权限冲突
	if req.Key != nil && *req.Key != perm.Key {
		existing, _ := s.repo.FindByKey(*req.Key)
		if existing != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "权限 key 已存在")
		}
		perm.Key = *req.Key
	}

	if req.Description != nil {
		perm.Description = req.Description
	}
	if req.RoleID != nil {
		perm.RoleID = *req.RoleID
	}

	if err := s.repo.Update(perm); err != nil {
		return nil, fmt.Errorf("更新权限失败: %w", err)
	}

	s.logger.Info("权限更新成功",
		zap.String("permissionId", perm.ID),
		zap.String("key", perm.Key),
	)

	return ToPermissionResponse(perm), nil
}

// Delete 删除权限
func (s *service) Delete(id string) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "权限不存在")
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("删除权限失败: %w", err)
	}

	s.logger.Info("权限删除成功",
		zap.String("permissionId", id),
	)

	return nil
}

// List 分页查询权限列表
func (s *service) List(page pagination.Pagination, keyword string) ([]PermissionResponse, int64, error) {
	perms, total, err := s.repo.List(page, keyword)
	if err != nil {
		return nil, 0, fmt.Errorf("查询权限列表失败: %w", err)
	}

	responses := make([]PermissionResponse, len(perms))
	for i, p := range perms {
		responses[i] = *ToPermissionResponse(&p)
	}

	return responses, total, nil
}
