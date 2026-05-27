// Package auth — 认证业务逻辑
// 登录、刷新令牌、登出、密码重置、账户解锁
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"helloGo/internal/module/user"
	redisPkg "helloGo/internal/pkg/redis"
)

// AuthService 认证服务接口
type AuthService interface {
	Login(ctx context.Context, username, password string) (*LoginResponse, error)
	Refresh(ctx context.Context, refreshToken, sessionID string) (*RefreshResponse, error)
	Logout(ctx context.Context, userID, sessionID string) error
	RequestPasswordReset(ctx context.Context, username string) (*RequestPasswordResetResponse, error)
	ResetPassword(ctx context.Context, username, newPassword, token string) error
	UnlockUser(ctx context.Context, username string) error
}

// authService 认证服务实现
type authService struct {
	db      *gorm.DB
	redis   *redisPkg.Client
	jwt     *JWTService
	logger  *zap.Logger
	config  *AuthConfig
}

// AuthConfig 认证配置
type AuthConfig struct {
	MaxFails int // 最大连续失败次数
	LockTTL  int // 账户锁定时长（秒）
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB, redis *redisPkg.Client, jwt *JWTService, logger *zap.Logger, config *AuthConfig) AuthService {
	return &authService{
		db:     db,
		redis:  redis,
		jwt:    jwt,
		logger: logger,
		config: config,
	}
}

// Login 用户登录
func (s *authService) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
	// 1. 检查账户是否被锁定
	lockKey := fmt.Sprintf("login:lock:%s", username)
	if locked, _ := s.redis.Get(ctx, lockKey); locked != "" {
		ttl, _ := s.redis.TTL(ctx, lockKey)
		return nil, fmt.Errorf("账户已锁定，请在 %v 后重试", ttl.Round(time.Minute))
	}

	// 2. 查询用户（预加载角色）
	var u user.User
	if err := s.db.Preload("Roles").Where("username = ?", username).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, s.handleLoginFail(ctx, username, "用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 3. 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, s.handleLoginFail(ctx, username, "密码错误")
	}

	// 4. 检查用户是否激活
	if !u.IsActive {
		return nil, fmt.Errorf("用户已被禁用")
	}

	// 5. 登录成功，清除失败计数
	failKey := fmt.Sprintf("login:fail:%s", username)
	s.redis.Del(ctx, failKey)

	// 6. 生成令牌
	roles := make([]string, len(u.Roles))
	for i, r := range u.Roles {
		roles[i] = r.Code
	}

	sessionID := uuid.New().String()

	accessToken, err := s.jwt.GenerateAccessToken(u.ID, u.Username, roles)
	if err != nil {
		return nil, fmt.Errorf("生成 access token 失败: %w", err)
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(u.ID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("生成 refresh token 失败: %w", err)
	}

	// 7. 存储 session 到 Redis（7天过期）
	sessionKey := fmt.Sprintf("session:%s:%s", u.ID, sessionID)
	if err := s.redis.Set(ctx, sessionKey, refreshToken, 7*24*time.Hour); err != nil {
		return nil, fmt.Errorf("存储 session 失败: %w", err)
	}

	s.logger.Info("用户登录成功",
		zap.String("userId", u.ID),
		zap.String("username", u.Username),
	)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionID:    sessionID,
	}, nil
}

// Refresh 刷新令牌
func (s *authService) Refresh(ctx context.Context, refreshToken, sessionID string) (*RefreshResponse, error) {
	// 1. 验证 refresh token
	claims, err := s.jwt.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("无效的 refresh token: %w", err)
	}

	// 2. 验证 session ID
	if claims.SessionID != sessionID {
		return nil, fmt.Errorf("session ID 不匹配")
	}

	// 3. 验证 Redis 中的 session
	sessionKey := fmt.Sprintf("session:%s:%s", claims.UserID, sessionID)
	storedToken, err := s.redis.Get(ctx, sessionKey)
	if err != nil || storedToken != refreshToken {
		return nil, fmt.Errorf("session 不存在或已过期")
	}

	// 4. 查询用户最新信息
	var u user.User
	if err := s.db.Preload("Roles").Where("id = ?", claims.UserID).First(&u).Error; err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 5. 生成新令牌
	roles := make([]string, len(u.Roles))
	for i, r := range u.Roles {
		roles[i] = r.Code
	}

	newAccessToken, err := s.jwt.GenerateAccessToken(u.ID, u.Username, roles)
	if err != nil {
		return nil, fmt.Errorf("生成 access token 失败: %w", err)
	}

	newRefreshToken, err := s.jwt.GenerateRefreshToken(u.ID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("生成 refresh token 失败: %w", err)
	}

	// 6. 更新 Redis session
	if err := s.redis.Set(ctx, sessionKey, newRefreshToken, 7*24*time.Hour); err != nil {
		return nil, fmt.Errorf("更新 session 失败: %w", err)
	}

	s.logger.Info("令牌刷新成功",
		zap.String("userId", u.ID),
		zap.String("username", u.Username),
	)

	return &RefreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		SessionID:    sessionID,
	}, nil
}

// Logout 登出
func (s *authService) Logout(ctx context.Context, userID, sessionID string) error {
	sessionKey := fmt.Sprintf("session:%s:%s", userID, sessionID)
	if err := s.redis.Del(ctx, sessionKey); err != nil {
		return fmt.Errorf("删除 session 失败: %w", err)
	}

	s.logger.Info("用户登出成功", zap.String("userId", userID))
	return nil
}

// RequestPasswordReset 请求密码重置
func (s *authService) RequestPasswordReset(ctx context.Context, username string) (*RequestPasswordResetResponse, error) {
	// 1. 查询用户
	var u user.User
	if err := s.db.Where("username = ?", username).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 2. 生成重置令牌
	token := uuid.New().String()

	// 3. 存储到 Redis（15分钟过期）
	resetKey := fmt.Sprintf("pwdreset:%s:%s", u.ID, token)
	if err := s.redis.Set(ctx, resetKey, "1", 15*time.Minute); err != nil {
		return nil, fmt.Errorf("存储重置令牌失败: %w", err)
	}

	s.logger.Info("密码重置请求", zap.String("userId", u.ID))

	return &RequestPasswordResetResponse{
		Token: token,
	}, nil
}

// ResetPassword 重置密码
func (s *authService) ResetPassword(ctx context.Context, username, newPassword, token string) error {
	// 1. 查询用户
	var u user.User
	if err := s.db.Where("username = ?", username).First(&u).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 2. 验证重置令牌
	resetKey := fmt.Sprintf("pwdreset:%s:%s", u.ID, token)
	if val, _ := s.redis.Get(ctx, resetKey); val == "" {
		return fmt.Errorf("重置令牌无效或已过期")
	}

	// 3. 哈希新密码
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码哈希失败: %w", err)
	}

	// 4. 更新密码
	if err := s.db.Model(&u).Update("password_hash", string(hash)).Error; err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	// 5. 删除重置令牌
	s.redis.Del(ctx, resetKey)

	s.logger.Info("密码重置成功", zap.String("userId", u.ID))
	return nil
}

// UnlockUser 解锁账户
func (s *authService) UnlockUser(ctx context.Context, username string) error {
	lockKey := fmt.Sprintf("login:lock:%s", username)
	failKey := fmt.Sprintf("login:fail:%s", username)

	if err := s.redis.Del(ctx, lockKey, failKey); err != nil {
		return fmt.Errorf("解锁失败: %w", err)
	}

	s.logger.Info("账户解锁成功", zap.String("username", username))
	return nil
}

// handleLoginFail 处理登录失败
func (s *authService) handleLoginFail(ctx context.Context, username, reason string) error {
	failKey := fmt.Sprintf("login:fail:%s", username)
	lockKey := fmt.Sprintf("login:lock:%s", username)

	// 递增失败计数
	count, err := s.redis.Incr(ctx, failKey)
	if err != nil {
		s.logger.Error("递增失败计数失败", zap.Error(err))
	}

	// 设置过期时间（10分钟）
	s.redis.Expire(ctx, failKey, 10*time.Minute)

	// 检查是否超过最大失败次数
	if count >= int64(s.config.MaxFails) {
		// 锁定账户
		s.redis.Set(ctx, lockKey, "1", time.Duration(s.config.LockTTL)*time.Second)
		s.logger.Warn("账户已锁定",
			zap.String("username", username),
			zap.Int64("failCount", count),
		)
		return fmt.Errorf("登录失败次数过多，账户已锁定 %d 秒", s.config.LockTTL)
	}

	s.logger.Warn("登录失败",
		zap.String("username", username),
		zap.String("reason", reason),
		zap.Int64("failCount", count),
	)

	return fmt.Errorf("用户名或密码错误（剩余 %d 次尝试）", s.config.MaxFails-int(count))
}
