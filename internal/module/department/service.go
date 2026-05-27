// Package department — 部门业务逻辑
package department

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Service 部门业务接口
type Service interface {
	Create(req *CreateDepartmentRequest) (*DepartmentResponse, error)
	GetByID(id string) (*DepartmentResponse, error)
	Update(id string, req *UpdateDepartmentRequest) (*DepartmentResponse, error)
	Delete(id string) error
	GetTree() ([]*DepartmentResponse, error)
}

// service 部门业务实现
type service struct {
	repo   Repository
	logger *zap.Logger
}

// NewService 创建部门业务层
func NewService(repo Repository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// Create 创建部门
func (s *service) Create(req *CreateDepartmentRequest) (*DepartmentResponse, error) {
	// 如果指定了父节点，检查父节点是否存在
	if req.ParentID != nil && *req.ParentID != "" {
		_, err := s.repo.FindByID(*req.ParentID)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "父部门不存在")
		}
	}

	dept := &Department{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
	}

	if err := s.repo.Create(dept); err != nil {
		return nil, fmt.Errorf("创建部门失败: %w", err)
	}

	s.logger.Info("部门创建成功",
		zap.String("deptId", dept.ID),
		zap.String("name", dept.Name),
	)

	return ToDepartmentResponse(dept), nil
}

// GetByID 按 ID 查询部门
func (s *service) GetByID(id string) (*DepartmentResponse, error) {
	dept, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "部门不存在")
	}
	return ToDepartmentResponse(dept), nil
}

// Update 更新部门
func (s *service) Update(id string, req *UpdateDepartmentRequest) (*DepartmentResponse, error) {
	dept, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "部门不存在")
	}

	// 不允许将自身设为父节点
	if req.ParentID != nil && *req.ParentID == id {
		return nil, fiber.NewError(fiber.StatusBadRequest, "不能将部门设为自身的子部门")
	}

	if req.Name != nil {
		dept.Name = *req.Name
	}
	if req.Description != nil {
		dept.Description = req.Description
	}
	if req.ParentID != nil {
		dept.ParentID = req.ParentID
	}

	if err := s.repo.Update(dept); err != nil {
		return nil, fmt.Errorf("更新部门失败: %w", err)
	}

	s.logger.Info("部门更新成功",
		zap.String("deptId", dept.ID),
		zap.String("name", dept.Name),
	)

	return ToDepartmentResponse(dept), nil
}

// Delete 删除部门
func (s *service) Delete(id string) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "部门不存在")
	}

	// 检查是否有子部门
	hasChildren, err := s.repo.HasChildren(id)
	if err != nil {
		return fmt.Errorf("检查子部门失败: %w", err)
	}
	if hasChildren {
		return fiber.NewError(fiber.StatusBadRequest, "请先删除子部门")
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("删除部门失败: %w", err)
	}

	s.logger.Info("部门删除成功",
		zap.String("deptId", id),
	)

	return nil
}

// GetTree 获取部门树
func (s *service) GetTree() ([]*DepartmentResponse, error) {
	depts, err := s.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("查询部门列表失败: %w", err)
	}

	return BuildDepartmentTree(depts), nil
}
