# CLAUDE.md

helloGo 是 helloNest（`/home/zhuchunlin/github/helloNest`，NestJS 管理后台）的 Go 重写版。行为不确定时参考 helloNest 实现。

## 常用命令

```bash
# 构建 & 运行
make run / run-dev / build

# 测试
make test / test-unit / test-integration / test-e2e / test-cover
go test -run TestFuncName ./path/to/pkg/...   # 单个测试

# 代码质量
make lint / fmt / swagger

# 数据库
make seed / seed-purge

# Docker
make docker-up / infra-up

# 微服务
make run-gateway / run-user / run-auth / run-permission / run-biz
make proto    # 生成 protobuf 代码

# K8s
make k8s-build / k8s-install / k8s-upgrade / k8s-deploy SVC=user
make k8s-status / k8s-logs SVC=user / k8s-urls

# 前端 (front-end/)
make frontend-install / frontend-dev / frontend-build
```

## 技术栈

Go ≥1.26 · Fiber v2 · GORM · SQLite/MySQL/PG · Redis · JWT v5 · Viper · Zap · gRPC · Protobuf · Prometheus · Swagger

## 项目结构

```
cmd/
  server/       # Phase 1 单体入口
  gateway/      # Phase 2 API Gateway (:8000)
  user/         # User Service (:50001)
  auth/         # Auth Service (:50002)
  permission/   # Permission Service (:50003)
  biz/          # Biz Service (:50004)
  seed/         # 种子数据（Phase 1）
  seed-v2/      # 种子数据（Phase 2 微服务版）
internal/
  module/{auth,user,role,permission,menu,department,dict,log,upload}/  # Phase 1 模块
  config/ database/ middleware/ guard/ pkg/ integration/               # Phase 1 基础设施
  shared/{config,database,redis,logger,interceptor,health}/            # Phase 2 共享
  {user,auth,permission,biz,gateway}/                                  # Phase 2 服务
api/proto/{service}/v1/*.proto   # Protobuf 定义
gen/go/                          # Protobuf 生成代码（gitignored）
configs/                         # .env.{example,development,production} + 各服务 .env
front-end/                       # React 测试面板（Vite + Zustand，端口 9003）
deploy/                          # Docker + Helm 部署
```

## 架构要点

**模块分层（Phase 1）：** 每个 `internal/module/<name>/` 按 model → repository → service → handler → dto 分层。`auth` 模块无 model/repository（使用 Redis + JWT）。

**NestJS → Go 映射：** `@Public()` → JWT 中间件外的路由 · `@Roles()` → `requireRoles()` · `@Permissions()` → `requirePerms()` · Guard/Interceptor → Fiber 中间件 · DI → 接口注入 · Module → `RegisterRoutes(app, deps)`。

**响应格式：** 统一 `{ code, statusCode, message, data, path, timestamp, requestId }`，错误码支持 i18n（zh-CN/en-US，`X-Lang` 头）。

**认证：** JWT access + refresh token，Redis 存储 session（`session:{userId}:{sessionId}`，7 天 TTL），登录失败追踪（`login:fail/lock:{username}`）。中间件链：Recovery → Trace → CORS → RequestLogger → ErrorHandler → CSRF → RateLimiter → Metrics。

**自引用树：** Menu、Department 使用 `ParentID *string` + GORM `BelongsTo`/`HasMany`。

**Redis 降级：** Redis 连接失败时自动降级为 `sync.Map` 内存缓存。

**数据库自动创建：** 启动时自动创建数据库（MySQL/PG/SQLite 均支持），无需手动建库。

**配置优先级：** 环境变量 > `.env.{APP_ENV}` > `.env` > 默认值。

## 微服务架构（Phase 2）

Gateway(:8000) → 4 个 gRPC 服务。Gateway 通过 `Auth.VerifyToken` 验证 JWT，手动映射 gRPC 响应为 camelCase。Auth Service 不连数据库（通过 gRPC 调用 User Service）。各服务默认 SQLite（`./data/{service}.db`），可配置 MySQL/PG。服务发现用 K8s DNS。

## K8s 部署（Phase 3）

minikube + Helm + Lens。Frontend/Gateway 用 NodePort（30090/30080），内部服务用 ClusterIP，MySQL/Redis 用 StatefulSet。所有服务实现 `/healthz` + `/readyz`（8080 端口）。

## 语言

用户用中文交流。文档和代码注释用中文，标识符、commit message、日志 key 用英文。
