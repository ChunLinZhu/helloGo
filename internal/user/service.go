// Package user — UserService gRPC 服务实现
package user

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "helloGo/gen/go/common/v1"
	userv1 "helloGo/gen/go/user/v1"
)

// Service 实现 userv1.UserServiceServer 接口
type Service struct {
	userv1.UnimplementedUserServiceServer
	repo   Repository
	logger *zap.Logger
}

// NewService 创建 UserService
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

// ListUsers 分页查询用户列表
func (s *Service) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	page := int(req.Pagination.GetPage())
	limit := int(req.Pagination.GetLimit())
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	users, total, err := s.repo.List(page, limit, req.Keyword)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询用户列表失败: %v", err)
	}

	pbUsers := make([]*userv1.User, len(users))
	for i := range users {
		pbUsers[i] = toProtoUser(&users[i])
	}

	totalPages := int32(0)
	if limit > 0 {
		totalPages = int32((int(total) + limit - 1) / limit)
	}

	return &userv1.ListUsersResponse{
		Users: pbUsers,
		Pagination: &commonv1.PaginationResponse{
			Page:       int32(page),
			Limit:      int32(limit),
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetUser 按 ID 查询用户
func (s *Service) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.UserResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "用户 ID 不能为空")
	}

	user, err := s.repo.FindByID(req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}

	return &userv1.UserResponse{User: toProtoUser(user)}, nil
}

// GetUserByUsername 按用户名查询用户（Auth Service 登录时调用）
func (s *Service) GetUserByUsername(ctx context.Context, req *userv1.GetUserByUsernameRequest) (*userv1.UserResponse, error) {
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "用户名不能为空")
	}

	user, err := s.repo.FindByUsername(req.Username)
	if err != nil {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}

	return &userv1.UserResponse{User: toProtoUser(user)}, nil
}

// CreateUser 创建新用户
func (s *Service) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.UserResponse, error) {
	// 参数校验
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "用户名不能为空")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "密码不能为空")
	}

	// 检查用户名唯一性
	existing, _ := s.repo.FindByUsername(req.Username)
	if existing != nil {
		return nil, status.Error(codes.AlreadyExists, "用户名已存在")
	}

	// 密码哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "密码哈希失败: %v", err)
	}

	// 创建用户
	user := &User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Email:        nilIfEmpty(req.Email),
		Phone:        nilIfEmpty(req.Phone),
		IsActive:     true, // CreateUser 默认启用
	}

	if err := s.repo.Create(user); err != nil {
		return nil, status.Errorf(codes.Internal, "创建用户失败: %v", err)
	}

	// 关联角色（通过 role_codes 解析为 role IDs）
	if len(req.RoleCodes) > 0 {
		roles, err := s.repo.FindRolesByCodes(req.RoleCodes)
		if err != nil {
			s.logger.Warn("查询角色失败", zap.Error(err))
		} else {
			roleIDs := make([]string, len(roles))
			for i, r := range roles {
				roleIDs[i] = r.ID
			}
			if err := s.repo.AssociateRoles(user.ID, roleIDs); err != nil {
				s.logger.Error("关联角色失败", zap.Error(err))
			}
		}
	}

	// 重新查询以获取完整信息（含角色）
	created, err := s.repo.FindByID(user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询创建的用户失败: %v", err)
	}

	s.logger.Info("用户创建成功",
		zap.String("userId", created.ID),
		zap.String("username", created.Username),
	)

	return &userv1.UserResponse{User: toProtoUser(created)}, nil
}

// UpdateUser 更新用户信息
func (s *Service) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UserResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "用户 ID 不能为空")
	}

	user, err := s.repo.FindByID(req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}

	// 更新可选字段
	if req.Email != nil {
		user.Email = nilIfEmpty(*req.Email)
	}
	if req.Phone != nil {
		user.Phone = nilIfEmpty(*req.Phone)
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.repo.Update(user); err != nil {
		return nil, status.Errorf(codes.Internal, "更新用户失败: %v", err)
	}

	// 更新角色关联（仅在 role_codes 非空时）
	if len(req.RoleCodes) > 0 {
		roles, err := s.repo.FindRolesByCodes(req.RoleCodes)
		if err != nil {
			s.logger.Warn("查询角色失败", zap.Error(err))
		} else {
			roleIDs := make([]string, len(roles))
			for i, r := range roles {
				roleIDs[i] = r.ID
			}
			if err := s.repo.AssociateRoles(user.ID, roleIDs); err != nil {
				s.logger.Error("更新角色关联失败", zap.Error(err))
			}
		}
	}

	// 重新查询以获取最新信息
	updated, err := s.repo.FindByID(user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询更新的用户失败: %v", err)
	}

	s.logger.Info("用户更新成功",
		zap.String("userId", updated.ID),
		zap.String("username", updated.Username),
	)

	return &userv1.UserResponse{User: toProtoUser(updated)}, nil
}

// DeleteUser 删除用户
func (s *Service) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*commonv1.Empty, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "用户 ID 不能为空")
	}

	user, err := s.repo.FindByID(req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}

	if err := s.repo.Delete(req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "删除用户失败: %v", err)
	}

	s.logger.Info("用户删除成功",
		zap.String("userId", req.Id),
		zap.String("username", user.Username),
	)

	return &commonv1.Empty{}, nil
}

// toProtoUser 将 GORM 模型转为 proto 消息
func toProtoUser(u *User) *userv1.User {
	email, phone := "", ""
	if u.Email != nil {
		email = *u.Email
	}
	if u.Phone != nil {
		phone = *u.Phone
	}

	roleCodes := make([]string, len(u.Roles))
	for i, r := range u.Roles {
		roleCodes[i] = r.Code
	}

	return &userv1.User{
		Id:         u.ID,
		Username:   u.Username,
		Email:      email,
		Phone:      phone,
		IsActive:   u.IsActive,
		RoleCodes:  roleCodes,
		CreatedAt:  timestamppb.New(u.CreatedAt),
		UpdatedAt:  timestamppb.New(u.UpdatedAt),
	}
}

// nilIfEmpty 空字符串转 nil，非空转 *string
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
