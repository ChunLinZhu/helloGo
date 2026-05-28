// Package permission — PermissionService gRPC 服务实现
// 管理角色、权限、菜单，提供权限校验和 Redis 缓存
package permission

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "helloGo/gen/go/common/v1"
	permissionv1 "helloGo/gen/go/permission/v1"
	sharedredis "helloGo/internal/shared/redis"
)

// 权限缓存配置
const (
	permCachePrefix = "rolePerms:"
	permCacheTTL    = 300 * time.Second
)

// Service 实现 permissionv1.PermissionServiceServer 接口
type Service struct {
	permissionv1.UnimplementedPermissionServiceServer
	repo   Repository
	redis  *sharedredis.Client
	logger *zap.Logger
}

// NewService 创建 PermissionService
func NewService(repo Repository, redis *sharedredis.Client, logger *zap.Logger) *Service {
	return &Service{repo: repo, redis: redis, logger: logger}
}

// ── 角色管理 ──────────────────────────────────────────────────

// ListRoles 分页查询角色列表
func (s *Service) ListRoles(ctx context.Context, req *permissionv1.ListRolesRequest) (*permissionv1.ListRolesResponse, error) {
	page, limit := normalizePagination(req.Pagination)

	roles, total, err := s.repo.ListRoles(page, limit, req.Keyword)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询角色列表失败: %v", err)
	}

	pbRoles := make([]*permissionv1.Role, len(roles))
	for i := range roles {
		pbRoles[i] = toProtoRole(&roles[i])
	}

	return &permissionv1.ListRolesResponse{
		Roles:      pbRoles,
		Pagination: buildPaginationResponse(page, limit, total),
	}, nil
}

// CreateRole 创建角色
func (s *Service) CreateRole(ctx context.Context, req *permissionv1.CreateRoleRequest) (*permissionv1.RoleResponse, error) {
	if req.Code == "" || req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "角色代码和名称不能为空")
	}

	// 检查代码唯一性
	if existing, _ := s.repo.FindRoleByCode(req.Code); existing != nil {
		return nil, status.Error(codes.AlreadyExists, "角色代码已存在")
	}

	role := &Role{
		Code:        req.Code,
		Name:        req.Name,
		Description: nilIfEmpty(req.Description),
	}

	if err := s.repo.CreateRole(role); err != nil {
		return nil, status.Errorf(codes.Internal, "创建角色失败: %v", err)
	}

	s.logger.Info("角色创建成功", zap.String("code", role.Code), zap.String("id", role.ID))

	// 重新查询以加载关联
	created, _ := s.repo.FindRoleByID(role.ID)
	return &permissionv1.RoleResponse{Role: toProtoRole(created)}, nil
}

// AddPermissionToRole 为角色分配权限（全量替换）
func (s *Service) AddPermissionToRole(ctx context.Context, req *permissionv1.AddPermissionToRoleRequest) (*permissionv1.RoleResponse, error) {
	if req.RoleId == "" {
		return nil, status.Error(codes.InvalidArgument, "角色 ID 不能为空")
	}

	// 验证角色存在
	role, err := s.repo.FindRoleByID(req.RoleId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "角色不存在")
	}

	// 全量替换权限关联
	if err := s.repo.AssociatePermissions(req.RoleId, req.PermissionIds); err != nil {
		return nil, status.Errorf(codes.Internal, "分配权限失败: %v", err)
	}

	// 清除 Redis 缓存
	s.invalidateRolePermCache(ctx, role.Code)

	s.logger.Info("角色权限分配成功",
		zap.String("roleId", req.RoleId),
		zap.Int("permCount", len(req.PermissionIds)),
	)

	// 重新查询以获取最新关联
	updated, _ := s.repo.FindRoleByID(req.RoleId)
	return &permissionv1.RoleResponse{Role: toProtoRole(updated)}, nil
}

// ── 权限管理 ──────────────────────────────────────────────────

// ListPermissions 分页查询权限列表
func (s *Service) ListPermissions(ctx context.Context, req *permissionv1.ListPermissionsRequest) (*permissionv1.ListPermissionsResponse, error) {
	page, limit := normalizePagination(req.Pagination)

	perms, total, err := s.repo.ListPermissions(page, limit, req.Keyword)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询权限列表失败: %v", err)
	}

	pbPerms := make([]*permissionv1.PermissionMsg, len(perms))
	for i := range perms {
		pbPerms[i] = toProtoPermission(&perms[i])
	}

	return &permissionv1.ListPermissionsResponse{
		Permissions: pbPerms,
		Pagination:  buildPaginationResponse(page, limit, total),
	}, nil
}

// CreatePermission 创建权限
func (s *Service) CreatePermission(ctx context.Context, req *permissionv1.CreatePermissionRequest) (*permissionv1.PermissionMsg, error) {
	if req.Key == "" || req.RoleId == "" {
		return nil, status.Error(codes.InvalidArgument, "权限标识和角色 ID 不能为空")
	}

	perm := &Permission{
		Key:         req.Key,
		Description: nilIfEmpty(req.Description),
		RoleID:      req.RoleId,
	}

	if err := s.repo.CreatePermission(perm); err != nil {
		// GORM unique constraint violation
		return nil, status.Errorf(codes.Internal, "创建权限失败（key 可能已存在）: %v", err)
	}

	// 清除该角色的权限缓存
	if role, err := s.repo.FindRoleByID(req.RoleId); err == nil {
		s.invalidateRolePermCache(ctx, role.Code)
	}

	s.logger.Info("权限创建成功", zap.String("key", perm.Key), zap.String("id", perm.ID))

	created, _ := s.repo.FindPermissionByID(perm.ID)
	return toProtoPermission(created), nil
}

// UpdatePermission 更新权限
func (s *Service) UpdatePermission(ctx context.Context, req *permissionv1.UpdatePermissionRequest) (*permissionv1.PermissionMsg, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "权限 ID 不能为空")
	}

	perm, err := s.repo.FindPermissionByID(req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "权限不存在")
	}

	oldRoleID := perm.RoleID

	if req.Key != "" {
		perm.Key = req.Key
	}
	if req.Description != "" {
		perm.Description = nilIfEmpty(req.Description)
	}
	if req.RoleId != "" {
		perm.RoleID = req.RoleId
	}

	if err := s.repo.UpdatePermission(perm); err != nil {
		return nil, status.Errorf(codes.Internal, "更新权限失败: %v", err)
	}

	// 清除旧角色和新角色的缓存
	if oldRole, err := s.repo.FindRoleByID(oldRoleID); err == nil {
		s.invalidateRolePermCache(ctx, oldRole.Code)
	}
	if perm.RoleID != oldRoleID {
		if newRole, err := s.repo.FindRoleByID(perm.RoleID); err == nil {
			s.invalidateRolePermCache(ctx, newRole.Code)
		}
	}

	s.logger.Info("权限更新成功", zap.String("key", perm.Key), zap.String("id", perm.ID))

	updated, _ := s.repo.FindPermissionByID(perm.ID)
	return toProtoPermission(updated), nil
}

// DeletePermission 删除权限
func (s *Service) DeletePermission(ctx context.Context, req *permissionv1.DeletePermissionRequest) (*commonv1.Empty, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "权限 ID 不能为空")
	}

	perm, err := s.repo.FindPermissionByID(req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "权限不存在")
	}

	// 记录角色 ID 以便清除缓存
	roleID := perm.RoleID

	if err := s.repo.DeletePermission(req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "删除权限失败: %v", err)
	}

	// 清除角色缓存
	if role, err := s.repo.FindRoleByID(roleID); err == nil {
		s.invalidateRolePermCache(ctx, role.Code)
	}

	s.logger.Info("权限删除成功", zap.String("key", perm.Key), zap.String("id", perm.ID))

	return &commonv1.Empty{}, nil
}

// ── 权限校验 ──────────────────────────────────────────────────

// CheckPermission 检查角色是否拥有指定权限（Gateway 调用）
func (s *Service) CheckPermission(ctx context.Context, req *permissionv1.CheckPermissionRequest) (*permissionv1.CheckPermissionResponse, error) {
	if len(req.RoleCodes) == 0 || req.RequiredPermission == "" {
		return &permissionv1.CheckPermissionResponse{Allowed: false}, nil
	}

	// 获取用户所有权限
	permResp, err := s.GetUserPermissions(ctx, &permissionv1.GetUserPermissionsRequest{
		RoleCodes: req.RoleCodes,
	})
	if err != nil {
		return &permissionv1.CheckPermissionResponse{Allowed: false}, nil
	}

	// 检查是否包含所需权限
	for _, perm := range permResp.Permissions {
		if perm == req.RequiredPermission {
			return &permissionv1.CheckPermissionResponse{Allowed: true}, nil
		}
	}

	return &permissionv1.CheckPermissionResponse{Allowed: false}, nil
}

// GetUserPermissions 获取用户所有权限（按角色聚合，带 Redis 缓存）
func (s *Service) GetUserPermissions(ctx context.Context, req *permissionv1.GetUserPermissionsRequest) (*permissionv1.GetUserPermissionsResponse, error) {
	if len(req.RoleCodes) == 0 {
		return &permissionv1.GetUserPermissionsResponse{Permissions: []string{}}, nil
	}

	var allPerms []string

	for _, roleCode := range req.RoleCodes {
		cacheKey := permCachePrefix + roleCode

		// 1. 尝试从 Redis 读取
		cached, err := s.redis.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var perms []string
			if json.Unmarshal([]byte(cached), &perms) == nil {
				allPerms = append(allPerms, perms...)
				continue
			}
		}

		// 2. 缓存未命中，查数据库
		role, err := s.repo.FindRoleByCode(roleCode)
		if err != nil {
			s.logger.Warn("角色不存在，跳过",
				zap.String("roleCode", roleCode),
				zap.Error(err),
			)
			continue
		}

		// 3. 提取权限 key 列表
		permKeys := make([]string, len(role.Permissions))
		for i, p := range role.Permissions {
			permKeys[i] = p.Key
		}

		// 4. 写入 Redis 缓存（TTL 300s）
		data, _ := json.Marshal(permKeys)
		if err := s.redis.Set(ctx, cacheKey, string(data), permCacheTTL); err != nil {
			s.logger.Warn("写入权限缓存失败",
				zap.String("roleCode", roleCode),
				zap.Error(err),
			)
		}

		allPerms = append(allPerms, permKeys...)
	}

	// 5. 去重
	return &permissionv1.GetUserPermissionsResponse{
		Permissions: uniqueStrings(allPerms),
	}, nil
}

// ── 菜单 ──────────────────────────────────────────────────────

// ListMenus 返回完整菜单树
func (s *Service) ListMenus(ctx context.Context, req *permissionv1.ListMenusRequest) (*permissionv1.ListMenusResponse, error) {
	menus, err := s.repo.FindAllMenus()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询菜单列表失败: %v", err)
	}

	tree := buildMenuTree(menus)

	return &permissionv1.ListMenusResponse{
		Menus: tree,
	}, nil
}

// ── 辅助函数 ──────────────────────────────────────────────────

// invalidateRolePermCache 清除角色权限 Redis 缓存
func (s *Service) invalidateRolePermCache(ctx context.Context, roleCode string) {
	cacheKey := permCachePrefix + roleCode
	if err := s.redis.Del(ctx, cacheKey); err != nil {
		s.logger.Warn("清除权限缓存失败",
			zap.String("roleCode", roleCode),
			zap.Error(err),
		)
	}
}

// toProtoRole 将 GORM Role 转为 proto Role
func toProtoRole(r *Role) *permissionv1.Role {
	desc := ""
	if r.Description != nil {
		desc = *r.Description
	}

	permKeys := make([]string, len(r.Permissions))
	for i, p := range r.Permissions {
		permKeys[i] = p.Key
	}

	return &permissionv1.Role{
		Id:             r.ID,
		Code:           r.Code,
		Name:           r.Name,
		Description:    desc,
		PermissionKeys: permKeys,
		CreatedAt:      timestamppb.New(r.CreatedAt),
		UpdatedAt:      timestamppb.New(r.UpdatedAt),
	}
}

// toProtoPermission 将 GORM Permission 转为 proto PermissionMsg
func toProtoPermission(p *Permission) *permissionv1.PermissionMsg {
	desc := ""
	if p.Description != nil {
		desc = *p.Description
	}

	return &permissionv1.PermissionMsg{
		Id:          p.ID,
		Key:         p.Key,
		Description: desc,
		RoleId:      p.RoleID,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}

// toProtoMenu 将 GORM Menu 转为 proto Menu（不含子菜单）
func toProtoMenu(m *Menu) *permissionv1.Menu {
	path, icon, parentID := "", "", ""
	if m.Path != nil {
		path = *m.Path
	}
	if m.Icon != nil {
		icon = *m.Icon
	}
	if m.ParentID != nil {
		parentID = *m.ParentID
	}

	return &permissionv1.Menu{
		Id:       m.ID,
		Name:     m.Name,
		Path:     path,
		Icon:     icon,
		Order:    int32(m.Order),
		ParentId: parentID,
	}
}

// buildMenuTree 将平面菜单列表构建为树结构
func buildMenuTree(menus []Menu) []*permissionv1.Menu {
	// 第一遍：创建所有节点
	nodeMap := make(map[string]*permissionv1.Menu, len(menus))
	for i := range menus {
		pb := toProtoMenu(&menus[i])
		pb.Children = []*permissionv1.Menu{}
		nodeMap[pb.Id] = pb
	}

	// 第二遍：链接父子关系
	var roots []*permissionv1.Menu
	for i := range menus {
		node := nodeMap[menus[i].ID]
		if menus[i].ParentID != nil && *menus[i].ParentID != "" {
			if parent, ok := nodeMap[*menus[i].ParentID]; ok {
				parent.Children = append(parent.Children, node)
				continue
			}
		}
		// 无父节点或父节点不存在 → 作为根节点
		roots = append(roots, node)
	}

	return roots
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

// uniqueStrings 去重字符串切片
func uniqueStrings(ss []string) []string {
	seen := make(map[string]struct{}, len(ss))
	result := make([]string, 0, len(ss))
	for _, s := range ss {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}
	return result
}

// nilIfEmpty 空字符串返回 nil，否则返回指针
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
