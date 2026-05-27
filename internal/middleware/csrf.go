// Package middleware — CSRF 防护中间件
// 支持两种模式：
//   - header 模式：JWT 签名的 CSRF token（30 分钟有效），通过 X-CSRF-Token 头验证
//   - cookie 模式：使用 Fiber 内置 CSRF 中间件（cookie 双重验证）
package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	appErrors "helloGo/internal/pkg/errors"
)

// CSRFConfig CSRF 配置
type CSRFConfig struct {
	Enabled bool   // 是否启用
	Mode    string // header 或 cookie
	Secret  string // JWT 签名密钥（header 模式）
}

// CSRFTokenResponse CSRF token 响应
type CSRFTokenResponse struct {
	Token     string `json:"csrfToken"`
	ExpiresIn int    `json:"expiresIn"` // 秒
}

// csrfTokenTTL CSRF token 有效期（30 分钟）
const csrfTokenTTL = 30 * time.Minute

// CSRFGenerateToken 生成 CSRF token（header 模式专用）
// 返回 JWT 签名的 token，30 分钟有效
func CSRFGenerateToken(secret string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"t":   now.Unix(),
		"exp": now.Add(csrfTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// CSRFHeaderMiddleware JWT 签名的 CSRF header 验证中间件
// 跳过 GET/HEAD/OPTIONS 请求，其他方法必须携带 X-CSRF-Token 头
func CSRFHeaderMiddleware(secret string, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 安全方法不需要验证
		method := c.Method()
		if method == fiber.MethodGet || method == fiber.MethodHead || method == fiber.MethodOptions {
			return c.Next()
		}

		// 读取 X-CSRF-Token 头
		tokenString := c.Get("X-CSRF-Token")
		if tokenString == "" {
			return appErrors.New(appErrors.CodeForbidden, fiber.StatusForbidden, "缺少 CSRF token")
		}

		// 验证 JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 确保签名算法是 HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusBadRequest, "无效的签名算法")
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			logger.Warn("CSRF 验证失败",
				zap.Error(err),
				zap.String("path", c.Path()),
				zap.String("ip", c.IP()),
			)
			return appErrors.New(appErrors.CodeForbidden, fiber.StatusForbidden, "CSRF token 无效或已过期")
		}

		return c.Next()
	}
}
