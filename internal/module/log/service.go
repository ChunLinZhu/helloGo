// Package log — 日志业务逻辑
package log

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"helloGo/internal/pkg/pagination"
)

// Service 日志业务接口
type Service interface {
	Create(req *CreateLogRequest) (*LogResponse, error)
	GetByID(id string) (*LogResponse, error)
	List(page pagination.Pagination, level string) ([]LogResponse, int64, error)
}

// service 日志业务实现
type service struct {
	repo   Repository
	logger *zap.Logger
}

// NewService 创建日志业务层
func NewService(repo Repository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// Create 创建日志
func (s *service) Create(req *CreateLogRequest) (*LogResponse, error) {
	l := &Log{
		Level:   req.Level,
		Message: req.Message,
		Meta:    req.Meta,
	}

	if err := s.repo.Create(l); err != nil {
		return nil, fmt.Errorf("创建日志失败: %w", err)
	}

	return ToLogResponse(l), nil
}

// GetByID 按 ID 查询日志
func (s *service) GetByID(id string) (*LogResponse, error) {
	l, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "日志不存在")
	}
	return ToLogResponse(l), nil
}

// List 分页查询日志列表
func (s *service) List(page pagination.Pagination, level string) ([]LogResponse, int64, error) {
	logs, total, err := s.repo.List(page, level)
	if err != nil {
		return nil, 0, fmt.Errorf("查询日志列表失败: %w", err)
	}

	responses := make([]LogResponse, len(logs))
	for i, l := range logs {
		responses[i] = *ToLogResponse(&l)
	}

	return responses, total, nil
}
