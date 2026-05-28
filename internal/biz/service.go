// Package biz — BizService gRPC 服务实现
// 管理部门、字典、日志、上传文件
package biz

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "helloGo/gen/go/common/v1"
	bizv1 "helloGo/gen/go/biz/v1"
)

// Service 实现 bizv1.BizServiceServer 接口
type Service struct {
	bizv1.UnimplementedBizServiceServer
	repo   Repository
	logger *zap.Logger
}

// NewService 创建 BizService
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

// ── 部门 ──────────────────────────────────────────────────────

// ListDepartments 返回完整部门树
func (s *Service) ListDepartments(ctx context.Context, req *bizv1.ListDepartmentsRequest) (*bizv1.ListDepartmentsResponse, error) {
	depts, err := s.repo.FindAllDepartments()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询部门列表失败: %v", err)
	}

	tree := buildDepartmentTree(depts)

	return &bizv1.ListDepartmentsResponse{
		Departments: tree,
	}, nil
}

// CreateDepartment 创建部门
func (s *Service) CreateDepartment(ctx context.Context, req *bizv1.CreateDepartmentRequest) (*bizv1.Department, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "部门名称不能为空")
	}

	dept := &Department{
		Name:        req.Name,
		Description: nilIfEmpty(req.Description),
		ParentID:    nilIfEmpty(req.ParentId),
	}

	if err := s.repo.CreateDepartment(dept); err != nil {
		return nil, status.Errorf(codes.Internal, "创建部门失败: %v", err)
	}

	s.logger.Info("部门创建成功", zap.String("name", dept.Name), zap.String("id", dept.ID))

	return toProtoDepartment(dept, nil), nil
}

// ── 字典 ──────────────────────────────────────────────────────

// ListDicts 分页查询字典
func (s *Service) ListDicts(ctx context.Context, req *bizv1.ListDictsRequest) (*bizv1.ListDictsResponse, error) {
	page, limit := normalizePagination(req.Pagination)

	dicts, total, err := s.repo.ListDicts(page, limit, req.Keyword, req.Type)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询字典列表失败: %v", err)
	}

	pbDicts := make([]*bizv1.Dict, len(dicts))
	for i := range dicts {
		pbDicts[i] = toProtoDict(&dicts[i])
	}

	return &bizv1.ListDictsResponse{
		Dicts:      pbDicts,
		Pagination: buildPaginationResponse(page, limit, total),
	}, nil
}

// CreateDict 创建字典项
func (s *Service) CreateDict(ctx context.Context, req *bizv1.CreateDictRequest) (*bizv1.Dict, error) {
	if req.Type == "" || req.Key == "" || req.Value == "" {
		return nil, status.Error(codes.InvalidArgument, "字典类型、键和值不能为空")
	}

	dict := &Dict{
		Type:        req.Type,
		Key:         req.Key,
		Value:       req.Value,
		Description: nilIfEmpty(req.Description),
	}

	if err := s.repo.CreateDict(dict); err != nil {
		return nil, status.Errorf(codes.Internal, "创建字典失败（type+key 可能已存在）: %v", err)
	}

	s.logger.Info("字典创建成功", zap.String("type", dict.Type), zap.String("key", dict.Key))

	return toProtoDict(dict), nil
}

// ── 日志 ──────────────────────────────────────────────────────

// ListLogs 分页查询日志
func (s *Service) ListLogs(ctx context.Context, req *bizv1.ListLogsRequest) (*bizv1.ListLogsResponse, error) {
	page, limit := normalizePagination(req.Pagination)

	logs, total, err := s.repo.ListLogs(page, limit, req.Level)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询日志列表失败: %v", err)
	}

	pbLogs := make([]*bizv1.Log, len(logs))
	for i := range logs {
		pbLogs[i] = toProtoLog(&logs[i])
	}

	return &bizv1.ListLogsResponse{
		Logs:       pbLogs,
		Pagination: buildPaginationResponse(page, limit, total),
	}, nil
}

// CreateLog 创建日志（供其他服务调用）
func (s *Service) CreateLog(ctx context.Context, req *bizv1.CreateLogRequest) (*bizv1.Log, error) {
	if req.Level == "" || req.Message == "" {
		return nil, status.Error(codes.InvalidArgument, "日志级别和消息不能为空")
	}

	log := &Log{
		Level:   req.Level,
		Message: req.Message,
		Meta:    nilIfEmpty(req.Meta),
	}

	if err := s.repo.CreateLog(log); err != nil {
		return nil, status.Errorf(codes.Internal, "创建日志失败: %v", err)
	}

	return toProtoLog(log), nil
}

// ── 上传 ──────────────────────────────────────────────────────

// ListUploads 分页查询上传文件
func (s *Service) ListUploads(ctx context.Context, req *bizv1.ListUploadsRequest) (*bizv1.ListUploadsResponse, error) {
	page, limit := normalizePagination(req.Pagination)

	uploads, total, err := s.repo.ListUploads(page, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询上传列表失败: %v", err)
	}

	pbUploads := make([]*bizv1.Upload, len(uploads))
	for i := range uploads {
		pbUploads[i] = toProtoUpload(&uploads[i])
	}

	return &bizv1.ListUploadsResponse{
		Uploads:    pbUploads,
		Pagination: buildPaginationResponse(page, limit, total),
	}, nil
}

// ── 健康检查 ──────────────────────────────────────────────────

// HealthCheck 健康检查
func (s *Service) HealthCheck(ctx context.Context, req *commonv1.Empty) (*bizv1.HealthCheckResponse, error) {
	return &bizv1.HealthCheckResponse{
		Status:  "ok",
		Service: "biz",
		Version: "1.0.0",
	}, nil
}

// ── 辅助函数 ──────────────────────────────────────────────────

// toProtoDepartment 将 GORM Department 转为 proto Department（不含子部门）
func toProtoDepartment(d *Department, children []*bizv1.Department) *bizv1.Department {
	desc, parentID := "", ""
	if d.Description != nil {
		desc = *d.Description
	}
	if d.ParentID != nil {
		parentID = *d.ParentID
	}

	return &bizv1.Department{
		Id:          d.ID,
		Name:        d.Name,
		Description: desc,
		ParentId:    parentID,
		Children:    children,
		CreatedAt:   timestamppb.New(d.CreatedAt),
		UpdatedAt:   timestamppb.New(d.UpdatedAt),
	}
}

// buildDepartmentTree 将平面部门列表构建为树结构
func buildDepartmentTree(depts []Department) []*bizv1.Department {
	nodeMap := make(map[string]*bizv1.Department, len(depts))
	for i := range depts {
		pb := toProtoDepartment(&depts[i], []*bizv1.Department{})
		nodeMap[pb.Id] = pb
	}

	var roots []*bizv1.Department
	for i := range depts {
		node := nodeMap[depts[i].ID]
		if depts[i].ParentID != nil && *depts[i].ParentID != "" {
			if parent, ok := nodeMap[*depts[i].ParentID]; ok {
				parent.Children = append(parent.Children, node)
				continue
			}
		}
		roots = append(roots, node)
	}

	return roots
}

// toProtoDict 将 GORM Dict 转为 proto Dict
func toProtoDict(d *Dict) *bizv1.Dict {
	desc := ""
	if d.Description != nil {
		desc = *d.Description
	}
	return &bizv1.Dict{
		Id:          d.ID,
		Type:        d.Type,
		Key:         d.Key,
		Value:       d.Value,
		Description: desc,
		CreatedAt:   timestamppb.New(d.CreatedAt),
		UpdatedAt:   timestamppb.New(d.UpdatedAt),
	}
}

// toProtoLog 将 GORM Log 转为 proto Log
func toProtoLog(l *Log) *bizv1.Log {
	meta := ""
	if l.Meta != nil {
		meta = *l.Meta
	}
	return &bizv1.Log{
		Id:        l.ID,
		Level:     l.Level,
		Message:   l.Message,
		Meta:      meta,
		CreatedAt: timestamppb.New(l.CreatedAt),
	}
}

// toProtoUpload 将 GORM Upload 转为 proto Upload
func toProtoUpload(u *Upload) *bizv1.Upload {
	return &bizv1.Upload{
		Id:           u.ID,
		Filename:     u.Filename,
		OriginalName: u.OriginalName,
		Mimetype:     u.Mimetype,
		Size:         u.Size,
		Path:         u.Path,
		CreatedAt:    timestamppb.New(u.CreatedAt),
	}
}

// normalizePagination 标准化分页参数
func normalizePagination(p *commonv1.PaginationRequest) (int, int) {
	page := int(p.GetPage())
	limit := int(p.GetLimit())
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

// buildPaginationResponse 构建分页响应
func buildPaginationResponse(page, limit int, total int64) *commonv1.PaginationResponse {
	totalPages := int32(0)
	if limit > 0 {
		totalPages = int32((int(total) + limit - 1) / limit)
	}
	return &commonv1.PaginationResponse{
		Page:       int32(page),
		Limit:      int32(limit),
		Total:      total,
		TotalPages: totalPages,
	}
}

// nilIfEmpty 空字符串返回 nil
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
