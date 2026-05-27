// Package guard — JWT 认证守卫
// 从 Authorization 头提取 Bearer token，验证后将用户信息注入 context
package guard

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"helloGo/internal/module/auth"
	appErrors "helloGo/internal/pkg/errors"
)

// JWTGuard JWT 认证守卫
type JWTGuard struct {
	jwtService *auth.JWTService
	logger     *zap.Logger
}

// NewJWTGuard 创建 JWT 守卫
func NewJWTGuard(jwtService *auth.JWTService, logger *zap.Logger) *JWTGuard {
	return &JWTGuard{
		jwtService: jwtService,
		logger:     logger,
	}
}

// Middleware 返回 JWT 认证中间件
func (g *JWTGuard) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 从 Authorization 头提取 token
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return appErrors.Unauthorized()
		}

		// 检查 Bearer 前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return appErrors.Unauthorized()
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return appErrors.Unauthorized()
		}

		// 验证 token
		claims, err := g.jwtService.ValidateToken(tokenString)
		if err != nil {
			g.logger.Warn("JWT 验证失败", zap.Error(err))
			return appErrors.Unauthorized()
		}

		// 将用户信息注入 context
		c.Locals("userId", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("roles", claims.Roles)

		return c.Next()
	}
}

// RequireRoles 角色守卫中间件
// 检查用户是否拥有指定角色之一
func RequireRoles(roles ...string) fiber.Handler {
	roleSet := make(map[string]bool)
	for _, r := range roles {
		roleSet[r] = true
	}

	return func(c *fiber.Ctx) error {
		userRoles, ok := c.Locals("roles").([]string)
		if !ok {
			return appErrors.Forbidden()
		}

		for _, r := range userRoles {
			if roleSet[r] {
				return c.Next()
			}
		}

		return appErrors.Forbidden()
	}
}

// RequirePermissions 权限守卫中间件（Phase 4 实现）
// 目前返回 Forbidden，后续从 Redis/DB 加载角色权限后实现
func RequirePermissions(perms ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Phase 4 实现权限检查
		// 1. 从 context 获取用户角色
		// 2. 从 Redis 缓存加载角色权限（rolePerms:{roleCode}）
		// 3. 检查用户是否拥有所有需要的权限
		// 4. 支持 :own 后缀的资源级权限
		return appErrors.Forbidden()
	}
}
