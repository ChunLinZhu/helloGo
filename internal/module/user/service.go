// Package user — 用户业务逻辑
package user

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"helloGo/internal/pkg/pagination"
)

// Service 用户业务接口
type Service interface {
	Create(req *CreateUserRequest) (*UserResponse, error)
	GetByID(id string) (*UserResponse, error)
	Update(id string, req *UpdateUserRequest) (*UserResponse, error)
	Delete(id string) error
	List(page pagination.Pagination, keyword string) ([]UserResponse, int64, error)
}

// service 用户业务实现
type service struct {
	repo   Repository
	logger *zap.Logger
}

// NewService 创建用户业务层
func NewService(repo Repository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// Create 创建用户
func (s *service) Create(req *CreateUserRequest) (*UserResponse, error) {
	// 检查用户名是否已存在
	existing, _ := s.repo.FindByUsername(req.Username)
	if existing != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "用户名已存在")
	}

	// 哈希密码
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码哈希失败: %w", err)
	}

	// 创建用户
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	user := &User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Email:        req.Email,
		Phone:        req.Phone,
		IsActive:     isActive,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 关联角色
	if len(req.RoleIDs) > 0 {
		if err := s.repo.AssociateRoles(user.ID, req.RoleIDs); err != nil {
			s.logger.Error("关联角色失败", zap.Error(err))
		}
	}

	// 重新查询以获取角色信息
	created, err := s.repo.FindByID(user.ID)
	if err != nil {
		return nil, fmt.Errorf("查询创建的用户失败: %w", err)
	}

	s.logger.Info("用户创建成功",
		zap.String("userId", created.ID),
		zap.String("username", created.Username),
	)

	return ToUserResponse(created), nil
}

// GetByID 按 ID 查询用户
func (s *service) GetByID(id string) (*UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "用户不存在")
	}
	return ToUserResponse(user), nil
}

// Update 更新用户
func (s *service) Update(id string, req *UpdateUserRequest) (*UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "用户不存在")
	}

	// 更新字段
	if req.Email != nil {
		user.Email = req.Email
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.repo.Update(user); err != nil {
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}

	// 更新角色关联
	if req.RoleIDs != nil {
		if err := s.repo.AssociateRoles(user.ID, req.RoleIDs); err != nil {
			s.logger.Error("更新角色关联失败", zap.Error(err))
		}
	}

	// 重新查询以获取最新角色信息
	updated, err := s.repo.FindByID(user.ID)
	if err != nil {
		return nil, fmt.Errorf("查询更新的用户失败: %w", err)
	}

	s.logger.Info("用户更新成功",
		zap.String("userId", updated.ID),
		zap.String("username", updated.Username),
	)

	return ToUserResponse(updated), nil
}

// Delete 删除用户
func (s *service) Delete(id string) error {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "用户不存在")
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	s.logger.Info("用户删除成功",
		zap.String("userId", id),
		zap.String("username", user.Username),
	)

	return nil
}

// List 分页查询用户列表
func (s *service) List(page pagination.Pagination, keyword string) ([]UserResponse, int64, error) {
	users, total, err := s.repo.List(page, keyword)
	if err != nil {
		return nil, 0, fmt.Errorf("查询用户列表失败: %w", err)
	}

	responses := make([]UserResponse, len(users))
	for i, u := range users {
		responses[i] = *ToUserResponse(&u)
	}

	return responses, total, nil
}
