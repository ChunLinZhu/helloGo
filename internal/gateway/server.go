// Package gateway — API Gateway 服务器
package gateway

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	authv1 "helloGo/gen/go/auth/v1"
	bizv1 "helloGo/gen/go/biz/v1"
	permissionv1 "helloGo/gen/go/permission/v1"
	userv1 "helloGo/gen/go/user/v1"
	"helloGo/internal/gateway/handler"
	gwmiddleware "helloGo/internal/gateway/middleware"
	"helloGo/internal/middleware"
	"helloGo/internal/pkg/response"
)

// Server API Gateway 服务器
type Server struct {
	app    *fiber.App
	port   int
	logger *zap.Logger
}

// Connections gRPC 连接集合
type Connections struct {
	User       userv1.UserServiceClient
	Auth       authv1.AuthServiceClient
	Permission permissionv1.PermissionServiceClient
	Biz        bizv1.BizServiceClient
}

// New 创建 API Gateway 服务器
func New(port int, conns *Connections, corsOrigins string, logger *zap.Logger) *Server {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			c.Status(code)
			return c.JSON(fiber.Map{
				"code":       "ERROR",
				"statusCode": code,
				"message":    err.Error(),
				"path":       c.Path(),
			})
		},
	})

	// ── 注册全局中间件 ─────────────────────────────────
	app.Use(middleware.Recovery(logger))
	app.Use(middleware.Trace())
	app.Use(middleware.CORS(corsOrigins))
	app.Use(middleware.RequestLogger(logger))

	// ── 创建处理器 ─────────────────────────────────────
	authHandler := handler.NewAuthHandler(conns.Auth)
	userHandler := handler.NewUserHandler(conns.User)
	permHandler := handler.NewPermissionHandler(conns.Permission)
	bizHandler := handler.NewBizHandler(conns.Biz)

	// ── JWT 中间件 ─────────────────────────────────────
	jwtMW := gwmiddleware.JWTMiddleware(conns.Auth, logger)

	// ── 健康检查（公开）────────────────────────────────
	app.Get("/api/health", func(c *fiber.Ctx) error {
		return response.SuccessOK(c, fiber.Map{
			"service": "gateway",
			"status":  "ok",
		})
	})

	app.Get("/api/health/ready", func(c *fiber.Ctx) error {
		return response.SuccessOK(c, fiber.Map{
			"service": "gateway",
			"status":  "ready",
		})
	})

	// ── CSRF Token（公开，兼容前端）──────────────────
	app.Get("/api/csrf-token", func(c *fiber.Ctx) error {
		// 微服务架构下 CSRF 可由 Gateway 统一生成
		// 前端仅需一个 token 值，此处返回随机字符串
		token := generateCSRFToken()
		return response.SuccessOK(c, fiber.Map{
			"csrfToken": token,
		})
	})

	// ── Metrics（公开，Prometheus 格式）────────────────
	app.Get("/api/metrics", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/plain; charset=utf-8")
		return c.SendString("# Gateway metrics endpoint (stub)\n# Integrate with Prometheus client in production\n")
	})

	// ── Auth 路由（公开）──────────────────────────────
	auth := app.Group("/api/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)
	auth.Post("/password/request-reset", authHandler.RequestPasswordReset)
	auth.Post("/password/reset", authHandler.ConfirmPasswordReset)
	auth.Post("/unlock", authHandler.UnlockAccount)

	// ── Auth 路由（受保护）────────────────────────────
	auth.Use(jwtMW)
	auth.Post("/logout", authHandler.Logout)

	// ── User 路由（受保护）────────────────────────────
	users := app.Group("/api/users", jwtMW)
	users.Get("/", userHandler.ListUsers)
	users.Get("/:id", userHandler.GetUser)
	users.Post("/", userHandler.CreateUser)
	users.Patch("/:id", userHandler.UpdateUser)
	users.Delete("/:id", userHandler.DeleteUser)

	// ── Permission 路由（受保护）──────────────────────
	roles := app.Group("/api/roles", jwtMW)
	roles.Get("/", permHandler.ListRoles)
	roles.Post("/", permHandler.CreateRole)
	roles.Get("/:id", permHandler.GetRole)
	roles.Delete("/:id", permHandler.DeleteRole)
	roles.Post("/:id/permissions", permHandler.AddPermissionToRole)

	permissions := app.Group("/api/permissions", jwtMW)
	permissions.Get("/", permHandler.ListPermissions)
	permissions.Post("/", permHandler.CreatePermission)
	permissions.Get("/:id", permHandler.GetPermission)
	permissions.Patch("/:id", permHandler.UpdatePermission)
	permissions.Delete("/:id", permHandler.DeletePermission)

	menus := app.Group("/api/menus", jwtMW)
	menus.Get("/", permHandler.ListMenus)
	menus.Get("/tree", permHandler.ListMenus) // 前端使用 /api/menus/tree

	// ── Biz 路由（受保护）────────────────────────────
	departments := app.Group("/api/departments", jwtMW)
	departments.Get("/", bizHandler.ListDepartments)
	departments.Get("/tree", bizHandler.ListDepartments) // 前端使用 /api/departments/tree
	departments.Post("/", bizHandler.CreateDepartment)

	dicts := app.Group("/api/dicts", jwtMW)
	dicts.Get("/", bizHandler.ListDicts)
	dicts.Post("/", bizHandler.CreateDict)

	logs := app.Group("/api/logs", jwtMW)
	logs.Get("/", bizHandler.ListLogs)

	uploads := app.Group("/api/uploads", jwtMW)
	uploads.Get("/", bizHandler.ListUploads)
	uploads.Post("/", bizHandler.UploadFile)
	uploads.Post("/chunk", bizHandler.UploadChunk)
	uploads.Post("/merge", bizHandler.MergeChunks)
	uploads.Delete("/:id", bizHandler.DeleteUpload)

	return &Server{
		app:    app,
		port:   port,
		logger: logger,
	}
}

// Start 启动 Gateway 服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	s.logger.Info("API Gateway 启动中",
		zap.Int("port", s.port),
		zap.String("addr", addr),
	)
	return s.app.Listen(addr)
}

// generateCSRFToken 生成随机 CSRF Token
func generateCSRFToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
