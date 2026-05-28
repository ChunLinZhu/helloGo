// Package middleware — Gateway JWT 中间件
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	authv1 "helloGo/gen/go/auth/v1"
)

// JWTMiddleware JWT 验证中间件
// 调用 Auth Service 的 VerifyToken RPC 验证令牌
func JWTMiddleware(authClient authv1.AuthServiceClient, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. 提取 Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			c.Status(fiber.StatusUnauthorized)
			return c.JSON(fiber.Map{
				"code":       "UNAUTHORIZED",
				"statusCode": 401,
				"message":    "缺少认证令牌",
				"path":       c.Path(),
			})
		}

		// 2. 验证 Bearer 前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.Status(fiber.StatusUnauthorized)
			return c.JSON(fiber.Map{
				"code":       "UNAUTHORIZED",
				"statusCode": 401,
				"message":    "认证令牌格式错误",
				"path":       c.Path(),
			})
		}

		// 3. 提取 token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.Status(fiber.StatusUnauthorized)
			return c.JSON(fiber.Map{
				"code":       "UNAUTHORIZED",
				"statusCode": 401,
				"message":    "认证令牌为空",
				"path":       c.Path(),
			})
		}

		// 4. 调用 Auth Service 验证 token
		resp, err := authClient.VerifyToken(c.Context(), &authv1.VerifyTokenRequest{
			Token: token,
		})
		if err != nil {
			logger.Warn("JWT 验证失败",
				zap.Error(err),
				zap.String("path", c.Path()),
			)
			c.Status(fiber.StatusUnauthorized)
			return c.JSON(fiber.Map{
				"code":       "UNAUTHORIZED",
				"statusCode": 401,
				"message":    "令牌无效或已过期",
				"path":       c.Path(),
			})
		}

		// 5. 将用户信息注入到 Fiber context
		c.Locals("userId", resp.UserId)
		c.Locals("username", resp.Username)
		c.Locals("roles", resp.Roles)

		// 6. 继续执行后续处理器
		return c.Next()
	}
}
