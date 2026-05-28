// Package auth — AuthService gRPC 服务实现
// 无数据库微服务：通过 gRPC 调用 User Service，使用 Redis 管理会话
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	authv1 "helloGo/gen/go/auth/v1"
	commonv1 "helloGo/gen/go/common/v1"
	userv1 "helloGo/gen/go/user/v1"
	"helloGo/internal/shared/config"
	sharedredis "helloGo/internal/shared/redis"
)

// Service 实现 authv1.AuthServiceServer 接口
type Service struct {
	authv1.UnimplementedAuthServiceServer

	userClient userv1.UserServiceClient
	redis      *sharedredis.Client
	jwt        *JWTService
	login      config.LoginConfig
	logger     *zap.Logger
	conn       *grpc.ClientConn // User Service gRPC 连接
}

// NewService 创建 AuthService
func NewService(
	redisClient *sharedredis.Client,
	jwtConfig config.JWTConfig,
	loginConfig config.LoginConfig,
	userServiceAddr string,
	logger *zap.Logger,
) (*Service, error) {
	// 创建到 User Service 的 gRPC 连接
	conn, err := grpc.NewClient(
		userServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("连接 User Service 失败: %w", err)
	}

	logger.Info("已连接 User Service", zap.String("addr", userServiceAddr))

	return &Service{
		userClient: userv1.NewUserServiceClient(conn),
		redis:      redisClient,
		jwt:        NewJWTService(jwtConfig),
		login:      loginConfig,
		logger:     logger,
		conn:       conn,
	}, nil
}

// Close 关闭 gRPC 连接
func (s *Service) Close() {
	if s.conn != nil {
		s.conn.Close()
	}
}

// Login 用户登录
func (s *Service) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "用户名和密码不能为空")
	}

	// 1. 检查账户是否被锁定
	lockKey := "login:lock:" + req.Username
	locked, _ := s.redis.Exists(ctx, lockKey)
	if locked {
		ttl, _ := s.redis.TTL(ctx, lockKey)
		msg := fmt.Sprintf("账户已被锁定，请 %d 秒后重试", int(ttl.Seconds()))
		return nil, status.Error(codes.PermissionDenied, msg)
	}

	// 2. 调用 User Service 验证密码（跨服务 gRPC 调用）
	verifyResp, err := s.userClient.VerifyPassword(ctx, &userv1.VerifyPasswordRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		// 密码错误或用户不存在 — 记录失败次数
		s.handleLoginFail(ctx, req.Username)
		return nil, status.Error(codes.Unauthenticated, "用户名或密码错误")
	}

	user := verifyResp.User

	// 3. 登录成功 — 清除失败计数
	s.redis.Del(ctx, "login:fail:"+req.Username)

	// 4. 生成 JWT Token
	sessionID := uuid.New().String()
	accessToken, err := s.jwt.GenerateAccessToken(user.Id, user.Username, user.RoleCodes)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "生成 access token 失败: %v", err)
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.Id, sessionID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "生成 refresh token 失败: %v", err)
	}

	// 5. 存储 Session 到 Redis（7 天 TTL）
	sessionKey := fmt.Sprintf("session:%s:%s", user.Id, sessionID)
	if err := s.redis.Set(ctx, sessionKey, refreshToken, 7*24*time.Hour); err != nil {
		s.logger.Warn("存储 Session 失败", zap.Error(err))
	}

	s.logger.Info("用户登录成功",
		zap.String("userId", user.Id),
		zap.String("username", user.Username),
		zap.String("sessionId", sessionID),
	)

	return &authv1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionId:    sessionID,
		ExpiresIn:    s.jwt.ExpiresInSeconds(),
	}, nil
}

// RefreshToken 刷新 access token
func (s *Service) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.LoginResponse, error) {
	if req.RefreshToken == "" || req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token 和 session_id 不能为空")
	}

	// 1. 验证 refresh token
	claims, err := s.jwt.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "refresh token 无效或已过期")
	}

	// 2. 校验 session_id 是否匹配
	if claims.SessionID != req.SessionId {
		return nil, status.Error(codes.Unauthenticated, "session_id 不匹配")
	}

	// 3. 从 Redis 获取存储的 refresh token 并比对
	sessionKey := fmt.Sprintf("session:%s:%s", claims.UserID, claims.SessionID)
	storedToken, err := s.redis.Get(ctx, sessionKey)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "会话不存在或已过期")
	}
	if storedToken != req.RefreshToken {
		return nil, status.Error(codes.Unauthenticated, "refresh token 不匹配")
	}

	// 4. 调用 User Service 获取最新用户信息（含角色）
	userResp, err := s.userClient.GetUser(ctx, &userv1.GetUserRequest{Id: claims.UserID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取用户信息失败: %v", err)
	}

	user := userResp.User

	// 5. 生成新的 Token 对
	newSessionID := uuid.New().String()
	accessToken, err := s.jwt.GenerateAccessToken(user.Id, user.Username, user.RoleCodes)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "生成 access token 失败: %v", err)
	}

	newRefreshToken, err := s.jwt.GenerateRefreshToken(user.Id, newSessionID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "生成 refresh token 失败: %v", err)
	}

	// 6. 更新 Redis Session（删除旧的，创建新的）
	s.redis.Del(ctx, sessionKey)
	newSessionKey := fmt.Sprintf("session:%s:%s", user.Id, newSessionID)
	if err := s.redis.Set(ctx, newSessionKey, newRefreshToken, 7*24*time.Hour); err != nil {
		s.logger.Warn("更新 Session 失败", zap.Error(err))
	}

	s.logger.Info("Token 刷新成功",
		zap.String("userId", user.Id),
		zap.String("username", user.Username),
	)

	return &authv1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		SessionId:    newSessionID,
		ExpiresIn:    s.jwt.ExpiresInSeconds(),
	}, nil
}

// Logout 用户注销
func (s *Service) Logout(ctx context.Context, req *authv1.LogoutRequest) (*commonv1.Empty, error) {
	if req.SessionId == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id 和 user_id 不能为空")
	}

	sessionKey := fmt.Sprintf("session:%s:%s", req.UserId, req.SessionId)
	s.redis.Del(ctx, sessionKey)

	s.logger.Info("用户注销成功",
		zap.String("userId", req.UserId),
		zap.String("sessionId", req.SessionId),
	)

	return &commonv1.Empty{}, nil
}

// VerifyToken 验证 access token（供 Gateway 调用）
func (s *Service) VerifyToken(ctx context.Context, req *authv1.VerifyTokenRequest) (*authv1.VerifyTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token 不能为空")
	}

	claims, err := s.jwt.ValidateToken(req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "token 无效或已过期")
	}

	return &authv1.VerifyTokenResponse{
		UserId:   claims.UserID,
		Username: claims.Username,
		Roles:    claims.Roles,
	}, nil
}

// RequestPasswordReset 请求密码重置
func (s *Service) RequestPasswordReset(ctx context.Context, req *authv1.RequestPasswordResetRequest) (*authv1.RequestPasswordResetResponse, error) {
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "用户名不能为空")
	}

	// 通过 User Service 查找用户
	userResp, err := s.userClient.GetUserByUsername(ctx, &userv1.GetUserByUsernameRequest{
		Username: req.Username,
	})
	if err != nil {
		// 即使用户不存在也返回成功（安全考虑：不暴露用户是否存在）
		return &authv1.RequestPasswordResetResponse{ResetToken: ""}, nil
	}

	user := userResp.User

	// 生成重置令牌
	resetToken := uuid.New().String()
	resetKey := fmt.Sprintf("pwdreset:%s:%s", user.Id, resetToken)

	// 存储到 Redis，15 分钟有效
	if err := s.redis.Set(ctx, resetKey, "1", 15*time.Minute); err != nil {
		return nil, status.Errorf(codes.Internal, "存储重置令牌失败: %v", err)
	}

	s.logger.Info("密码重置请求",
		zap.String("userId", user.Id),
		zap.String("username", user.Username),
	)

	return &authv1.RequestPasswordResetResponse{ResetToken: resetToken}, nil
}

// ConfirmPasswordReset 确认密码重置
func (s *Service) ConfirmPasswordReset(ctx context.Context, req *authv1.ConfirmPasswordResetRequest) (*commonv1.Empty, error) {
	if req.ResetToken == "" || req.NewPassword == "" || req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "重置令牌、新密码和用户名不能为空")
	}

	// 查找用户
	userResp, err := s.userClient.GetUserByUsername(ctx, &userv1.GetUserByUsernameRequest{
		Username: req.Username,
	})
	if err != nil {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}

	user := userResp.User

	// 验证重置令牌
	resetKey := fmt.Sprintf("pwdreset:%s:%s", user.Id, req.ResetToken)
	exists, _ := s.redis.Exists(ctx, resetKey)
	if !exists {
		return nil, status.Error(codes.InvalidArgument, "重置令牌无效或已过期")
	}

	// 调用 User Service 更新密码
	_, err = s.userClient.UpdatePassword(ctx, &userv1.UpdatePasswordRequest{
		Username:    req.Username,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "更新密码失败: %v", err)
	}

	// 删除重置令牌
	s.redis.Del(ctx, resetKey)

	s.logger.Info("密码重置成功",
		zap.String("userId", user.Id),
		zap.String("username", user.Username),
	)

	return &commonv1.Empty{}, nil
}

// UnlockAccount 解锁账户
func (s *Service) UnlockAccount(ctx context.Context, req *authv1.UnlockAccountRequest) (*commonv1.Empty, error) {
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "用户名不能为空")
	}

	// 删除锁定键和失败计数
	s.redis.Del(ctx, "login:lock:"+req.Username, "login:fail:"+req.Username)

	s.logger.Info("账户已解锁", zap.String("username", req.Username))

	return &commonv1.Empty{}, nil
}

// handleLoginFail 处理登录失败（递增失败计数，超过阈值则锁定）
func (s *Service) handleLoginFail(ctx context.Context, username string) {
	failKey := "login:fail:" + username

	// 递增失败计数
	count, err := s.redis.Incr(ctx, failKey)
	if err != nil {
		s.logger.Warn("递增登录失败计数失败", zap.Error(err))
		return
	}

	// 设置/延长 TTL（10 分钟）
	s.redis.Expire(ctx, failKey, 10*time.Minute)

	// 超过最大失败次数则锁定
	if count >= int64(s.login.MaxFails) {
		lockKey := "login:lock:" + username
		lockTTL := time.Duration(s.login.LockTTL) * time.Second
		s.redis.Set(ctx, lockKey, "1", lockTTL)

		s.logger.Warn("账户已锁定",
			zap.String("username", username),
			zap.Int64("failCount", count),
			zap.Int("lockTTL", s.login.LockTTL),
		)
	}
}
