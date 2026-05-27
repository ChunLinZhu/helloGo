// Package dict — 字典业务逻辑
package dict

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"helloGo/internal/pkg/pagination"
)

// Service 字典业务接口
type Service interface {
	Create(req *CreateDictRequest) (*DictResponse, error)
	GetByID(id string) (*DictResponse, error)
	Update(id string, req *UpdateDictRequest) (*DictResponse, error)
	Delete(id string) error
	List(page pagination.Pagination, keyword, dictType string) ([]DictResponse, int64, error)
}

// service 字典业务实现
type service struct {
	repo   Repository
	logger *zap.Logger
}

// NewService 创建字典业务层
func NewService(repo Repository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// Create 创建字典项
func (s *service) Create(req *CreateDictRequest) (*DictResponse, error) {
	// 检查 type+key 是否已存在
	existing, _ := s.repo.FindByTypeKey(req.Type, req.Key)
	if existing != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "该类型下的 key 已存在")
	}

	d := &Dict{
		Type:        req.Type,
		Key:         req.Key,
		Value:       req.Value,
		Description: req.Description,
	}

	if err := s.repo.Create(d); err != nil {
		return nil, fmt.Errorf("创建字典项失败: %w", err)
	}

	s.logger.Info("字典项创建成功",
		zap.String("dictId", d.ID),
		zap.String("type", d.Type),
		zap.String("key", d.Key),
	)

	return ToDictResponse(d), nil
}

// GetByID 按 ID 查询字典项
func (s *service) GetByID(id string) (*DictResponse, error) {
	d, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "字典项不存在")
	}
	return ToDictResponse(d), nil
}

// Update 更新字典项
func (s *service) Update(id string, req *UpdateDictRequest) (*DictResponse, error) {
	d, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "字典项不存在")
	}

	// 如果要更新 type 或 key，检查是否与其他项冲突
	newType := d.Type
	newKey := d.Key
	if req.Type != nil {
		newType = *req.Type
	}
	if req.Key != nil {
		newKey = *req.Key
	}
	if newType != d.Type || newKey != d.Key {
		existing, _ := s.repo.FindByTypeKey(newType, newKey)
		if existing != nil && existing.ID != id {
			return nil, fiber.NewError(fiber.StatusBadRequest, "该类型下的 key 已存在")
		}
		d.Type = newType
		d.Key = newKey
	}

	if req.Value != nil {
		d.Value = *req.Value
	}
	if req.Description != nil {
		d.Description = req.Description
	}

	if err := s.repo.Update(d); err != nil {
		return nil, fmt.Errorf("更新字典项失败: %w", err)
	}

	s.logger.Info("字典项更新成功",
		zap.String("dictId", d.ID),
		zap.String("type", d.Type),
		zap.String("key", d.Key),
	)

	return ToDictResponse(d), nil
}

// Delete 删除字典项
func (s *service) Delete(id string) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "字典项不存在")
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("删除字典项失败: %w", err)
	}

	s.logger.Info("字典项删除成功",
		zap.String("dictId", id),
	)

	return nil
}

// List 分页查询字典列表
func (s *service) List(page pagination.Pagination, keyword, dictType string) ([]DictResponse, int64, error) {
	dicts, total, err := s.repo.List(page, keyword, dictType)
	if err != nil {
		return nil, 0, fmt.Errorf("查询字典列表失败: %w", err)
	}

	responses := make([]DictResponse, len(dicts))
	for i, d := range dicts {
		responses[i] = *ToDictResponse(&d)
	}

	return responses, total, nil
}
