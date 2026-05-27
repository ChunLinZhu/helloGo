// Package menu — 菜单业务逻辑
package menu

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Service 菜单业务接口
type Service interface {
	Create(req *CreateMenuRequest) (*MenuResponse, error)
	GetByID(id string) (*MenuResponse, error)
	Update(id string, req *UpdateMenuRequest) (*MenuResponse, error)
	Delete(id string) error
	GetTree() ([]*MenuResponse, error)
}

// service 菜单业务实现
type service struct {
	repo   Repository
	logger *zap.Logger
}

// NewService 创建菜单业务层
func NewService(repo Repository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// Create 创建菜单
func (s *service) Create(req *CreateMenuRequest) (*MenuResponse, error) {
	// 如果指定了父节点，检查父节点是否存在
	if req.ParentID != nil && *req.ParentID != "" {
		_, err := s.repo.FindByID(*req.ParentID)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "父菜单不存在")
		}
	}

	menu := &Menu{
		Name:     req.Name,
		Path:     req.Path,
		Icon:     req.Icon,
		Order:    req.Order,
		ParentID: req.ParentID,
	}

	if err := s.repo.Create(menu); err != nil {
		return nil, fmt.Errorf("创建菜单失败: %w", err)
	}

	s.logger.Info("菜单创建成功",
		zap.String("menuId", menu.ID),
		zap.String("name", menu.Name),
	)

	return ToMenuResponse(menu), nil
}

// GetByID 按 ID 查询菜单
func (s *service) GetByID(id string) (*MenuResponse, error) {
	menu, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "菜单不存在")
	}
	return ToMenuResponse(menu), nil
}

// Update 更新菜单
func (s *service) Update(id string, req *UpdateMenuRequest) (*MenuResponse, error) {
	menu, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "菜单不存在")
	}

	// 不允许将自身设为父节点
	if req.ParentID != nil && *req.ParentID == id {
		return nil, fiber.NewError(fiber.StatusBadRequest, "不能将菜单设为自身的子菜单")
	}

	if req.Name != nil {
		menu.Name = *req.Name
	}
	if req.Path != nil {
		menu.Path = req.Path
	}
	if req.Icon != nil {
		menu.Icon = req.Icon
	}
	if req.Order != nil {
		menu.Order = *req.Order
	}
	if req.ParentID != nil {
		menu.ParentID = req.ParentID
	}

	if err := s.repo.Update(menu); err != nil {
		return nil, fmt.Errorf("更新菜单失败: %w", err)
	}

	s.logger.Info("菜单更新成功",
		zap.String("menuId", menu.ID),
		zap.String("name", menu.Name),
	)

	return ToMenuResponse(menu), nil
}

// Delete 删除菜单
func (s *service) Delete(id string) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "菜单不存在")
	}

	// 检查是否有子菜单
	hasChildren, err := s.repo.HasChildren(id)
	if err != nil {
		return fmt.Errorf("检查子菜单失败: %w", err)
	}
	if hasChildren {
		return fiber.NewError(fiber.StatusBadRequest, "请先删除子菜单")
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("删除菜单失败: %w", err)
	}

	s.logger.Info("菜单删除成功",
		zap.String("menuId", id),
	)

	return nil
}

// GetTree 获取菜单树
func (s *service) GetTree() ([]*MenuResponse, error) {
	menus, err := s.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("查询菜单列表失败: %w", err)
	}

	return BuildMenuTree(menus), nil
}
