// Package guard — JWT 守卫中间件单元测试
package guard

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"helloGo/internal/module/auth"
	appErrors "helloGo/internal/pkg/errors"
)

func newTestJWTGuard() (*JWTGuard, *auth.JWTService) {
	jwtSvc := auth.NewJWTService(auth.JWTConfig{
		Secret:         "test-secret",
		Expires:        "1h",
		RefreshExpires: "7d",
	})
	logger, _ := zap.NewDevelopment()
	return NewJWTGuard(jwtSvc, logger), jwtSvc
}

// newTestApp 创建带错误处理的测试 Fiber 应用
func newTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if e, ok := err.(*appErrors.AppError); ok {
				return c.Status(e.HTTPStatus).JSON(fiber.Map{"code": e.Code})
			}
			return c.Status(500).JSON(fiber.Map{"code": "INTERNAL"})
		},
	})
}

// ── JWT 中间件测试 ─────────────────────────────────────────

func TestJWTMiddleware_NoHeader(t *testing.T) {
	guard, _ := newTestJWTGuard()
	app := newTestApp()
	app.Get("/protected", guard.Middleware(), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestJWTMiddleware_MissingBearer(t *testing.T) {
	guard, _ := newTestJWTGuard()
	app := newTestApp()
	app.Get("/protected", guard.Middleware(), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "some-token")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	guard, _ := newTestJWTGuard()
	app := newTestApp()
	app.Get("/protected", guard.Middleware(), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	guard, jwtSvc := newTestJWTGuard()

	token, err := jwtSvc.GenerateAccessToken("user-1", "alice", []string{"admin"})
	require.NoError(t, err)

	var capturedUserID, capturedUsername string
	var capturedRoles []string

	app := fiber.New()
	app.Get("/protected", guard.Middleware(), func(c *fiber.Ctx) error {
		capturedUserID = c.Locals("userId").(string)
		capturedUsername = c.Locals("username").(string)
		capturedRoles = c.Locals("roles").([]string)
		return c.JSON(fiber.Map{"ok": true})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "user-1", capturedUserID)
	assert.Equal(t, "alice", capturedUsername)
	assert.Equal(t, []string{"admin"}, capturedRoles)
}

// ── RequireRoles 中间件测试 ────────────────────────────────

func TestRequireRoles_Allowed(t *testing.T) {
	app := fiber.New()

	app.Get("/admin", func(c *fiber.Ctx) error {
		c.Locals("roles", []string{"admin", "user"})
		return c.Next()
	}, RequireRoles("admin"), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRequireRoles_Forbidden(t *testing.T) {
	app := newTestApp()

	app.Get("/admin", func(c *fiber.Ctx) error {
		c.Locals("roles", []string{"user"})
		return c.Next()
	}, RequireRoles("admin"), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRequireRoles_NoRoles(t *testing.T) {
	app := newTestApp()

	app.Get("/admin", RequireRoles("admin"), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRequireRoles_MultipleAllowed(t *testing.T) {
	app := fiber.New()

	app.Get("/staff", func(c *fiber.Ctx) error {
		c.Locals("roles", []string{"editor"})
		return c.Next()
	}, RequireRoles("admin", "editor"), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/staff", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// ── RequirePermissions 中间件测试 ──────────────────────────

func TestRequirePermissions_AlwaysForbidden(t *testing.T) {
	app := newTestApp()

	app.Get("/api", RequirePermissions("users:read"), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/api", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
