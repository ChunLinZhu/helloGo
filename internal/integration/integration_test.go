// Package integration — 集成测试（SQLite 内存 DB + Fiber httptest）
// 测试完整请求链路：HTTP → 中间件 → handler → service → DB
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"helloGo/internal/database"
	"helloGo/internal/guard"
	"helloGo/internal/module/auth"
	"helloGo/internal/module/permission"
	"helloGo/internal/module/role"
	"helloGo/internal/module/user"
	appErrors "helloGo/internal/pkg/errors"
	"helloGo/internal/pkg/response"
)

// ── 测试应用 ───────────────────────────────────────────────

type testApp struct {
	app      *fiber.App
	db       *gorm.DB
	jwtSvc   *auth.JWTService
	logger   *zap.Logger
}

// setupTestApp 创建集成测试用的 Fiber 应用
func setupTestApp(t *testing.T) *testApp {
	t.Helper()

	logger, _ := zap.NewDevelopment()

	// SQLite 内存数据库
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// AutoMigrate
	require.NoError(t, database.AutoMigrate(db, logger))

	// JWT 服务
	jwtSvc := auth.NewJWTService(auth.JWTConfig{
		Secret:         "integration-test-secret",
		Expires:        "1h",
		RefreshExpires: "7d",
	})

	// JWT 守卫
	jwtGuard := guard.NewJWTGuard(jwtSvc, logger)

	// 创建 Fiber 应用
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if e, ok := err.(*appErrors.AppError); ok {
				msg := e.GetMessage("zh-CN")
				if e.Detail != nil {
					msg = fmt.Sprintf("%v", e.Detail)
				}
				return c.Status(e.HTTPStatus).JSON(response.Response{
					Code:       string(e.Code),
					StatusCode: e.HTTPStatus,
					Message:    msg,
				})
			}
			// Fiber 内置错误类型（fiber.NewError 返回）
			if e, ok := err.(*fiber.Error); ok {
				return c.Status(e.Code).JSON(response.Response{
					Code:       string(appErrors.HTTPStatusToErrorCode(e.Code)),
					StatusCode: e.Code,
					Message:    e.Message,
				})
			}
			return c.Status(500).JSON(response.Response{
				Code:       "INTERNAL",
				StatusCode: 500,
				Message:    err.Error(),
			})
		},
	})

	// 注册路由
	api := app.Group("/api")

	// 用户模块
	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo, logger)
	userHandler := user.NewHandler(userSvc, logger)

	jwtMW := jwtGuard.Middleware()
	_ = jwtMW // 预留：部分路由仅需 JWT 不需 admin 角色
	adminMW := []fiber.Handler{jwtGuard.Middleware(), guard.RequireRoles("admin")}

	api.Get("/users", append(adminMW, userHandler.List)...)
	api.Get("/users/:id", append(adminMW, userHandler.GetByID)...)
	api.Post("/users", append(adminMW, userHandler.Create)...)
	api.Patch("/users/:id", append(adminMW, userHandler.Update)...)
	api.Delete("/users/:id", append(adminMW, userHandler.Delete)...)

	// 角色模块
	roleRepo := role.NewRepository(db)
	roleSvc := role.NewService(roleRepo, logger)
	roleHandler := role.NewHandler(roleSvc, logger)

	api.Get("/roles", append(adminMW, roleHandler.List)...)
	api.Post("/roles", append(adminMW, roleHandler.Create)...)

	// 权限模块
	permRepo := permission.NewRepository(db)
	permSvc := permission.NewService(permRepo, logger)
	permHandler := permission.NewHandler(permSvc, logger)

	api.Get("/permissions", append(adminMW, permHandler.List)...)

	return &testApp{
		app:    app,
		db:     db,
		jwtSvc: jwtSvc,
		logger: logger,
	}
}

// seedUser 在数据库中直接创建用户
func (ta *testApp) seedUser(t *testing.T, username, password string) *user.User {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u := &user.User{
		Username:     username,
		PasswordHash: string(hash),
		IsActive:     true,
	}
	require.NoError(t, ta.db.Create(u).Error)
	return u
}

// seedRole 在数据库中直接创建角色
func (ta *testApp) seedRole(t *testing.T, code, name string) *role.Role {
	t.Helper()
	r := &role.Role{Code: code, Name: name}
	require.NoError(t, ta.db.Create(r).Error)
	return r
}

// getAdminToken 获取 admin token
func (ta *testApp) getAdminToken(t *testing.T) string {
	t.Helper()
	token, err := ta.jwtSvc.GenerateAccessToken("admin-id", "admin", []string{"admin"})
	require.NoError(t, err)
	return token
}

// doRequest 发送 HTTP 请求
func (ta *testApp) doRequest(t *testing.T, method, path string, body interface{}, token string) *http.Response {
	t.Helper()
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := ta.app.Test(req, -1)
	require.NoError(t, err)
	return resp
}

// parseResponse 解析响应体
func parseResponse(t *testing.T, resp *http.Response) response.Response {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var r response.Response
	err = json.Unmarshal(body, &r)
	require.NoError(t, err)
	return r
}

// ── 测试用例 ───────────────────────────────────────────────

func TestHealthCheck(t *testing.T) {
	ta := setupTestApp(t)
	// 手动添加 health 路由
	ta.app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	resp := ta.doRequest(t, "GET", "/api/health", nil, "")
	assert.Equal(t, 200, resp.StatusCode)
}

func TestUserCRUD_FullChain(t *testing.T) {
	ta := setupTestApp(t)
	token := ta.getAdminToken(t)

	// 1. 创建用户
	createBody := map[string]interface{}{
		"username": "alice",
		"password": "password123",
		"email":    "alice@example.com",
	}
	resp := ta.doRequest(t, "POST", "/api/users", createBody, token)
	assert.Equal(t, 201, resp.StatusCode, "创建用户应返回 201")

	r := parseResponse(t, resp)
	assert.Equal(t, "OK", r.Code)

	dataBytes, _ := json.Marshal(r.Data)
	var createdUser map[string]interface{}
	json.Unmarshal(dataBytes, &createdUser)
	userID := createdUser["id"].(string)
	assert.NotEmpty(t, userID)
	assert.Equal(t, "alice", createdUser["username"])

	// 2. 按 ID 查询
	resp = ta.doRequest(t, "GET", "/api/users/"+userID, nil, token)
	assert.Equal(t, 200, resp.StatusCode, "查询用户应返回 200")

	r = parseResponse(t, resp)
	dataBytes, _ = json.Marshal(r.Data)
	var queriedUser map[string]interface{}
	json.Unmarshal(dataBytes, &queriedUser)
	assert.Equal(t, "alice", queriedUser["username"])

	// 3. 更新用户
	updateBody := map[string]interface{}{
		"email": "alice_new@example.com",
	}
	resp = ta.doRequest(t, "PATCH", "/api/users/"+userID, updateBody, token)
	assert.Equal(t, 200, resp.StatusCode, "更新用户应返回 200")

	// 4. 列表查询
	resp = ta.doRequest(t, "GET", "/api/users?page=1&limit=10", nil, token)
	assert.Equal(t, 200, resp.StatusCode, "列表查询应返回 200")

	r = parseResponse(t, resp)
	dataBytes, _ = json.Marshal(r.Data)
	var pageData map[string]interface{}
	json.Unmarshal(dataBytes, &pageData)
	assert.True(t, pageData["total"].(float64) >= 1, "应至少有 1 个用户")

	// 5. 删除用户
	resp = ta.doRequest(t, "DELETE", "/api/users/"+userID, nil, token)
	assert.Equal(t, 200, resp.StatusCode, "删除用户应返回 200")

	// 6. 删除后再查询 → 404
	resp = ta.doRequest(t, "GET", "/api/users/"+userID, nil, token)
	assert.Equal(t, 404, resp.StatusCode, "已删除用户应返回 404")
}

func TestUserEndpoints_RequireAuth(t *testing.T) {
	ta := setupTestApp(t)

	// 无 token 访问 → 401
	resp := ta.doRequest(t, "GET", "/api/users", nil, "")
	assert.Equal(t, 401, resp.StatusCode, "无 token 应返回 401")
}

func TestUserEndpoints_RequireAdminRole(t *testing.T) {
	ta := setupTestApp(t)

	// 非 admin token → 403
	token, _ := ta.jwtSvc.GenerateAccessToken("user-1", "normaluser", []string{"user"})
	resp := ta.doRequest(t, "GET", "/api/users", nil, token)
	assert.Equal(t, 403, resp.StatusCode, "非 admin 应返回 403")
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	ta := setupTestApp(t)
	token := ta.getAdminToken(t)

	body := map[string]interface{}{
		"username": "bob",
		"password": "password123",
	}

	// 第一次创建
	resp := ta.doRequest(t, "POST", "/api/users", body, token)
	assert.Equal(t, 201, resp.StatusCode)

	// 第二次创建同名用户 → 400
	resp = ta.doRequest(t, "POST", "/api/users", body, token)
	assert.Equal(t, 400, resp.StatusCode, "重复用户名应返回 400")
}

func TestCreateUser_ValidationFail(t *testing.T) {
	ta := setupTestApp(t)
	token := ta.getAdminToken(t)

	// 缺少必填字段
	body := map[string]interface{}{
		"username": "a", // 太短（min=2）
	}
	resp := ta.doRequest(t, "POST", "/api/users", body, token)
	assert.Equal(t, 400, resp.StatusCode, "校验失败应返回 400")
}

func TestRoleCRUD(t *testing.T) {
	ta := setupTestApp(t)
	token := ta.getAdminToken(t)

	// 创建角色
	body := map[string]interface{}{
		"code": "editor",
		"name": "编辑",
	}
	resp := ta.doRequest(t, "POST", "/api/roles", body, token)
	assert.Equal(t, 201, resp.StatusCode, "创建角色应返回 201")

	// 查询角色列表
	resp = ta.doRequest(t, "GET", "/api/roles?page=1&limit=10", nil, token)
	assert.Equal(t, 200, resp.StatusCode)

	r := parseResponse(t, resp)
	dataBytes, _ := json.Marshal(r.Data)
	var pageData map[string]interface{}
	json.Unmarshal(dataBytes, &pageData)
	assert.True(t, pageData["total"].(float64) >= 1)
}
