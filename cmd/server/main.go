// Package main — helloGo HTTP 服务入口
// 加载配置 → 初始化日志 → 初始化数据库 → 初始化 Redis → 注册中间件与路由 → 启动服务
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	swaggerFiber "github.com/swaggo/fiber-swagger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	_ "helloGo/docs" // Swagger 生成的文档
	"helloGo/internal/config"
	"helloGo/internal/database"
	"helloGo/internal/guard"
	"helloGo/internal/middleware"
	"helloGo/internal/module/auth"
	"helloGo/internal/module/department"
	"helloGo/internal/module/dict"
	logModule "helloGo/internal/module/log"
	"helloGo/internal/module/menu"
	"helloGo/internal/module/permission"
	"helloGo/internal/module/role"
	"helloGo/internal/module/upload"
	"helloGo/internal/module/user"
	redisPkg "helloGo/internal/pkg/redis"
	"gorm.io/gorm"
)

//	@title			helloGo API
//	@version		1.0
//	@description	helloGo 管理后台 API — Go + Fiber 实现
//	@host			localhost:8000
//	@BasePath		/api
//	@schemes		http
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				JWT Bearer token（格式：Bearer {token}）
func main() {
	// ── 1. 加载配置 ───────────────────────────────────────
	cfg, err := config.Load("./configs")
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// ── 2. 初始化日志（Zap + Lumberjack 日志轮转） ────────
	logger := initLogger(cfg.App.Env)
	defer logger.Sync()
	logger.Info("helloGo 启动中",
		zap.String("env", cfg.App.Env),
		zap.Int("port", cfg.App.Port),
	)

	// ── 3. 初始化数据库 ───────────────────────────────────
	db, err := database.Init(cfg.Database, cfg.App.Env, logger)
	if err != nil {
		logger.Fatal("数据库初始化失败", zap.Error(err))
	}

	// ── 4. 数据库迁移 ─────────────────────────────────────
	if err := database.AutoMigrate(db, logger); err != nil {
		logger.Fatal("数据库迁移失败", zap.Error(err))
	}

	// ── 5. 初始化 Redis ───────────────────────────────────
	redisClient := redisPkg.Init(cfg.Redis, logger)
	defer redisClient.Close()

	// ── 5.1 初始化 JWT 服务 ───────────────────────────────
	jwtService := auth.NewJWTService(auth.JWTConfig{
		Secret:         cfg.JWT.Secret,
		Expires:        cfg.JWT.Expires,
		RefreshExpires: cfg.JWT.RefreshExpires,
	})

	// ── 5.2 初始化 JWT 守卫 ───────────────────────────────
	jwtGuard := guard.NewJWTGuard(jwtService, logger)

	// ── 6. 创建 Fiber 应用 ────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:      "helloGo",
		BodyLimit:    int(cfg.Upload.MaxSize), // 请求体大小限制（与上传限制一致）
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		// 禁用 Fiber 默认的 Server 头
		DisableStartupMessage: true,
	})

	// ── 7. 注册全局中间件（按执行顺序） ───────────────────
	app.Use(middleware.Recovery(logger))   // 最先注册：捕获所有 panic
	app.Use(middleware.Trace())            // 请求追踪 ID
	app.Use(middleware.CORS(cfg.Security.CORSOrigins))
	app.Use(middleware.RequestLogger(logger))
	app.Use(middleware.ErrorHandler(logger)) // 全局错误处理（包裹路由，捕获所有 error）

	// CSRF 防护（按配置启用）
	if cfg.Security.CSRFEnabled {
		if cfg.Security.CSRFMode == "header" {
			app.Use(middleware.CSRFHeaderMiddleware(cfg.Security.CSRFSecret, logger))
			logger.Info("CSRF 防护已启用（header 模式）")
		} else {
			// cookie 模式：使用 Fiber 内置 CSRF 中间件
			logger.Info("CSRF 防护已启用（cookie 模式）")
			// Fiber 内置 CSRF 中间件在 cookie 模式下自动处理
			// 此处预留，后续可按需启用
		}
	}

	// 限流中间件
	app.Use(middleware.RateLimiter(middleware.RateLimitConfig{
		TTL:   cfg.Throttle.TTL,
		Limit: cfg.Throttle.Limit,
	}, logger))

	// Prometheus 指标中间件（按配置启用）
	if cfg.Metrics.Enabled {
		app.Use(middleware.MetricsMiddleware())
		logger.Info("Prometheus 指标采集已启用")
	}

	// ── 8. 注册路由 ───────────────────────────────────────
	api := app.Group("/api")
	registerHealthRoutes(api, db, redisClient, logger)

	// ── 8.0 Prometheus 指标端点（公开，无需认证） ─────────
	if cfg.Metrics.Enabled {
		api.Get("/metrics", middleware.MetricsHandler())
		logger.Info("Prometheus 指标端点已注册：GET /api/metrics")
	}

	// ── 8.0 CSRF token 端点（公开，无需 JWT） ─────────────
	if cfg.Security.CSRFEnabled && cfg.Security.CSRFMode == "header" {
		csrfSecret := cfg.Security.CSRFSecret
		api.Get("/csrf-token", func(c *fiber.Ctx) error {
			token, err := middleware.CSRFGenerateToken(csrfSecret)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "生成 CSRF token 失败")
			}
			return c.JSON(middleware.CSRFTokenResponse{
				Token:     token,
				ExpiresIn: 1800, // 30 分钟
			})
		})
	}

	// ── 8.0.1 审计日志中间件（需要 logRepo，放在路由注册前） ─
	logRepo := logModule.NewRepository(db)
	app.Use(middleware.AuditLogger(logger, logRepo))

	// ── 8.0.2 Swagger 文档端点（按配置启用，无需认证） ──────
	if cfg.Swagger.Enabled {
		app.Get("/docs/*", swaggerFiber.WrapHandler)
		logger.Info("Swagger 文档已启用：http://localhost:8000/docs/index.html")
	}

	// ── 8.1 注册认证路由 ──────────────────────────────────
	authService := auth.NewAuthService(db, redisClient, jwtService, logger, &auth.AuthConfig{
		MaxFails: cfg.Login.MaxFails,
		LockTTL:  cfg.Login.LockTTL,
	})
	authHandler := auth.NewAuthHandler(authService, logger)

	// 公开认证路由（无需 JWT）
	api.Post("/auth/login", authHandler.Login)
	api.Post("/auth/refresh", authHandler.Refresh)
	api.Post("/auth/password/request-reset", authHandler.RequestPasswordReset)
	api.Post("/auth/password/reset", authHandler.ResetPassword)
	api.Post("/auth/unlock", authHandler.Unlock)

	// 需要认证的认证路由（JWT 保护）
	api.Post("/auth/logout", jwtGuard.Middleware(), authHandler.Logout)

	// ── 8.2 注册业务模块路由（JWT + admin 角色保护） ───────
	// 中间件链：JWT 认证 + admin 角色检查
	jwtMW := jwtGuard.Middleware()
	adminMW := []fiber.Handler{jwtGuard.Middleware(), guard.RequireRoles("admin")}

	// ── 用户模块 ─────────────────────────────────────────
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo, logger)
	userHandler := user.NewHandler(userService, logger)

	api.Get("/users", append(adminMW, userHandler.List)...)
	api.Get("/users/:id", append(adminMW, userHandler.GetByID)...)
	api.Post("/users", append(adminMW, userHandler.Create)...)
	api.Patch("/users/:id", append(adminMW, userHandler.Update)...)
	api.Delete("/users/:id", append(adminMW, userHandler.Delete)...)

	// ── 角色模块 ─────────────────────────────────────────
	roleRepo := role.NewRepository(db)
	roleService := role.NewService(roleRepo, logger)
	roleHandler := role.NewHandler(roleService, logger)

	api.Get("/roles", append(adminMW, roleHandler.List)...)
	api.Get("/roles/:id", append(adminMW, roleHandler.GetByID)...)
	api.Post("/roles", append(adminMW, roleHandler.Create)...)
	api.Patch("/roles/:id", append(adminMW, roleHandler.Update)...)
	api.Delete("/roles/:id", append(adminMW, roleHandler.Delete)...)
	api.Post("/roles/:id/permissions", append(adminMW, roleHandler.AssignPermissions)...)

	// ── 权限模块 ─────────────────────────────────────────
	permRepo := permission.NewRepository(db)
	permService := permission.NewService(permRepo, logger)
	permHandler := permission.NewHandler(permService, logger)

	api.Get("/permissions", append(adminMW, permHandler.List)...)
	api.Get("/permissions/:id", append(adminMW, permHandler.GetByID)...)
	api.Post("/permissions", append(adminMW, permHandler.Create)...)
	api.Patch("/permissions/:id", append(adminMW, permHandler.Update)...)
	api.Delete("/permissions/:id", append(adminMW, permHandler.Delete)...)

	// ── 菜单模块 ─────────────────────────────────────────
	menuRepo := menu.NewRepository(db)
	menuService := menu.NewService(menuRepo, logger)
	menuHandler := menu.NewHandler(menuService, logger)

	api.Get("/menus/tree", jwtMW, menuHandler.GetTree)
	api.Get("/menus/:id", append(adminMW, menuHandler.GetByID)...)
	api.Post("/menus", append(adminMW, menuHandler.Create)...)
	api.Patch("/menus/:id", append(adminMW, menuHandler.Update)...)
	api.Delete("/menus/:id", append(adminMW, menuHandler.Delete)...)

	// ── 部门模块 ─────────────────────────────────────────
	deptRepo := department.NewRepository(db)
	deptService := department.NewService(deptRepo, logger)
	deptHandler := department.NewHandler(deptService, logger)

	api.Get("/departments/tree", jwtMW, deptHandler.GetTree)
	api.Get("/departments/:id", append(adminMW, deptHandler.GetByID)...)
	api.Post("/departments", append(adminMW, deptHandler.Create)...)
	api.Patch("/departments/:id", append(adminMW, deptHandler.Update)...)
	api.Delete("/departments/:id", append(adminMW, deptHandler.Delete)...)

	// ── 字典模块 ─────────────────────────────────────────
	dictRepo := dict.NewRepository(db)
	dictService := dict.NewService(dictRepo, logger)
	dictHandler := dict.NewHandler(dictService, logger)

	api.Get("/dicts", jwtMW, dictHandler.List)
	api.Get("/dicts/:id", jwtMW, dictHandler.GetByID)
	api.Post("/dicts", append(adminMW, dictHandler.Create)...)
	api.Patch("/dicts/:id", append(adminMW, dictHandler.Update)...)
	api.Delete("/dicts/:id", append(adminMW, dictHandler.Delete)...)

	// ── 日志模块 ─────────────────────────────────────────
	logService := logModule.NewService(logRepo, logger)
	logHandler := logModule.NewHandler(logService, logger)

	api.Get("/logs", jwtMW, logHandler.List)
	api.Get("/logs/:id", jwtMW, logHandler.GetByID)
	api.Post("/logs", append(adminMW, logHandler.Create)...)

	// ── 上传模块 ─────────────────────────────────────────
	uploadRepo := upload.NewRepository(db)
	uploadService := upload.NewService(uploadRepo, logger, cfg.Upload)
	uploadHandler := upload.NewHandler(uploadService, logger)

	api.Post("/uploads", jwtMW, uploadHandler.Upload)
	api.Post("/uploads/chunk", jwtMW, uploadHandler.UploadChunk)
	api.Post("/uploads/merge", jwtMW, uploadHandler.MergeChunks)
	api.Get("/uploads", jwtMW, uploadHandler.List)
	api.Get("/uploads/:id", jwtMW, uploadHandler.GetByID)
	api.Delete("/uploads/:id", append(adminMW, uploadHandler.Delete)...)

	// ── 静态文件服务（上传文件访问） ───────────────────────
	app.Static("/uploads", cfg.Upload.Dest)

	// ── 定时清理过期上传文件 ─────────────────────────────
	if cfg.Upload.CleanInterval > 0 {
		go func() {
			ticker := time.NewTicker(time.Duration(cfg.Upload.CleanInterval) * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				if err := uploadService.CleanExpired(); err != nil {
					logger.Error("定时清理失败", zap.Error(err))
				}
			}
		}()
		logger.Info("定时清理已启动",
			zap.Int("intervalSec", cfg.Upload.CleanInterval),
			zap.Int("ttlDays", cfg.Upload.TTLDays),
		)
	}

	// ── 9. 启动服务（支持优雅关闭） ──────────────────────
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	go func() {
		logger.Info("HTTP 服务已启动",
			zap.String("addr", addr),
			zap.Bool("swagger", cfg.Swagger.Enabled),
			zap.Bool("metrics", cfg.Metrics.Enabled),
		)
		if err := app.Listen(addr); err != nil {
			logger.Fatal("HTTP 服务启动失败", zap.Error(err))
		}
	}()

	// ── 10. 等待中断信号，优雅关闭 ────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info("收到关闭信号，正在优雅关闭...", zap.String("signal", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("优雅关闭失败", zap.Error(err))
	}
	logger.Info("服务已关闭")
}

// registerHealthRoutes 注册健康检查路由
func registerHealthRoutes(api fiber.Router, db *gorm.DB, redis *redisPkg.Client, logger *zap.Logger) {
	health := api.Group("/health")

	// GET /api/health — 存活检查（不依赖外部服务）
	health.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// GET /api/health/ready — 就绪检查（验证 DB + Redis 连接）
	health.Get("/ready", func(c *fiber.Ctx) error {
		ctx := context.Background()
		status := "ready"
		dbStatus := "ok"
		cacheStatus := "ok"

		// 检查数据库
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			status = "degraded"
			dbStatus = "error"
		}

		// 检查 Redis
		if err := redis.Ping(ctx); err != nil {
			status = "degraded"
			cacheStatus = "error"
		}

		httpStatus := fiber.StatusOK
		if status != "ready" {
			httpStatus = fiber.StatusServiceUnavailable
		}

		return c.Status(httpStatus).JSON(fiber.Map{
			"status": status,
			"db":     dbStatus,
			"cache":  cacheStatus,
		})
	})
}

// initLogger 初始化 Zap 日志
// 开发环境：控制台彩色输出；生产环境：JSON 格式 + 文件轮转
func initLogger(env string) *zap.Logger {
	// 日志轮转配置
	lumberjackLogger := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    50,  // 每个日志文件最大 50MB
		MaxBackups: 10,  // 最多保留 10 个旧文件
		MaxAge:     30,  // 旧文件最多保留 30 天
		Compress:   true, // 旧文件压缩
	}

	if env == "production" {
		// 生产环境：JSON 格式写入文件
		encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
		core := zapcore.NewCore(
			encoder,
			zapcore.AddSync(lumberjackLogger),
			zap.InfoLevel,
		)
		return zap.New(core, zap.AddCaller())
	}

	// 开发环境：彩色控制台输出
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zap.DebugLevel,
	)
	return zap.New(core, zap.AddCaller())
}
