# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Purpose

helloGo is a Go reimplementation of the sibling project **helloNest** (`/home/zhuchunlin/github/helloNest`), a NestJS admin backend. Phase 1 (Fiber monolith) and Phase 2 (gRPC microservices + API Gateway) are fully implemented. Phase 3 (K8s deployment) is in progress.

When behavior is ambiguous, refer to helloNest's implementation.

## Tech Stack

| Concern | Choice |
|---|---|
| Go version | >= 1.26.3 |
| HTTP framework | Fiber v2 |
| ORM | GORM |
| Databases | SQLite (default) / MySQL / PostgreSQL |
| Cache | go-redis v9 |
| Auth | golang-jwt v5 + bcrypt |
| Config | Viper (.env + YAML + env vars) |
| Logging | Zap + Lumberjack |
| Metrics | prometheus/client_golang |
| API docs | swaggo + fiber-swagger |
| gRPC (Phase 2) | google.golang.org/grpc + protoc |
| K8s (Phase 3) | minikube + Helm + Lens |

## Build & Run Commands

Backend:

```
make run          # go run cmd/server/main.go
make run-dev      # air 热重载（需先安装 air：go install github.com/air-verse/air@latest）
make build        # go build -o build/helloGo ./cmd/server
make test         # go test ./... -count=1
make test-unit    # 仅单元测试
make test-integration  # 仅集成测试（SQLite 内存 DB + Fiber httptest）
make test-e2e     # curl 脚本测试（需先启动服务）
make test-cover   # 覆盖率报告 → coverage.html
make lint         # golangci-lint run ./...
make swagger      # swag init -g cmd/server/main.go --parseDependency --parseInternal
make seed         # 种子数据
make seed-purge   # 清除后重新播种
make docker-up    # docker-compose 完整环境
make infra-up     # 仅基础设施（Redis + MySQL + PostgreSQL）
```

前端测试面板（`front-end/`）：

```
cd front-end && npm install    # 安装依赖
cd front-end && npm run dev    # 开发服务器（端口 9003）
cd front-end && npm run build  # 生产构建
make frontend-install          # 从根目录安装前端依赖
make frontend-dev              # 从根目录启动前端
```

Single test:
```
go test ./internal/module/user/...                        # one package
go test -run TestCreateUser ./internal/module/user/...    # one function
go test -v -count=1 ./internal/module/auth/...            # verbose, no cache
```

Dev workflow: `air` for hot reload (config in `.air.toml`).

Phase 2 微服务：

```
make run-user        # 启动 User Service (gRPC :50001)
make run-auth        # 启动 Auth Service (gRPC :50002)
make run-permission  # 启动 Permission Service (gRPC :50003)
make run-biz         # 启动 Biz Service (gRPC :50004)
make run-gateway     # 启动 API Gateway (HTTP :8000)
make proto-gen       # 生成 protobuf Go 代码
```

## Project Layout

```
cmd/server/main.go        # Phase 1 HTTP entry（路由注册、中间件链）
cmd/seed/main.go          # seed data script
cmd/gateway/main.go       # Phase 2 API Gateway entry
cmd/user/main.go          # Phase 2 User Service entry
cmd/auth/main.go          # Phase 2 Auth Service entry
cmd/permission/main.go    # Phase 2 Permission Service entry
cmd/biz/main.go           # Phase 2 Biz Service entry
api/proto/                # protobuf 定义
  {service}/v1/*.proto
gen/go/                   # protobuf 生成的 Go 代码（gitignored）
internal/
  config/                 # Phase 1 Viper config
  database/               # Phase 1 GORM init, AutoMigrate
  middleware/             # trace, cors, csrf, ratelimit, recovery, request_logger, error_handler, audit, metrics
  module/{auth,user,role,permission,menu,department,dict,log,upload}/  # Phase 1 模块
  guard/                  # Phase 1 jwt guard, role guard
  pkg/{response,errors,pagination,redis}/  # Phase 1 共享工具
  integration/            # 集成测试（SQLite 内存 DB + Fiber httptest）
  shared/                 # Phase 2 共享（config, database, redis, logger, interceptor）
  user/                   # Phase 2 User Service（model, repository, service）
  auth/                   # Phase 2 Auth Service（service, session）
  permission/             # Phase 2 Permission Service（model, repository, service）
  biz/                    # Phase 2 Biz Service（model, repository, service）
  gateway/                # Phase 2 API Gateway（server, handler, middleware）
front-end/                # React 前端测试面板
  src/lib/axios.ts        # HTTP 客户端（拦截器注入 token/CSRF）
  src/stores/app.ts       # Zustand store（token/sessionId/apiUrl）
  src/pages/*.tsx         # 11 个 API 测试页面
configs/
  .env.{example,development,production}  # Phase 1 配置
  {service}.env           # Phase 2 各服务配置（user.env, auth.env 等）
deploy/                   # Phase 3 部署（Docker + Helm，待实现）
```

Phase 2 (已完成) restructured into per-service binaries under `cmd/{gateway,user,auth,permission,biz}/` with shared protobuf definitions in `api/proto/`. Each service has its own package under `internal/{user,auth,permission,biz,gateway}/` and shared utilities in `internal/shared/`.

Phase 3 (进行中) deploys the microservices to minikube with Helm charts. Frontend (`front-end/`) is served via nginx. Lens is used for K8s GUI management. See `docs/phase3-kubernetes-deployment.md` for details.

## Key Architectural Conventions

**NestJS → Go mapping.** Fiber has no decorators or DI container. Apply these substitutions consistently:
- `@Public()` → routes outside the JWT middleware group
- `@Roles(...)` → `requireRoles(...)` middleware in the route chain
- `@Permissions(...)` → `requirePerms(...)` middleware
- NestJS `Guard` → Fiber middleware
- NestJS `Interceptor` → Fiber middleware (call `c.Next()`, then post-process)
- NestJS `ExceptionFilter` → global error handler via `app.Use(errorHandler)`
- NestJS DI → constructor injection with interfaces (`Service` interface + concrete impl)
- NestJS `Module` → Go package exposing `RegisterRoutes(app, deps)`

**Module structure.** Each `internal/module/<name>/` follows a strict layering:
1. `model.go` — GORM struct with UUID primary key (`BeforeCreate` hook generates `uuid.New().String()`)
2. `repository.go` — DB queries (receives `*gorm.DB`)
3. `service.go` — business logic (depends on repo interface)
4. `handler.go` — Fiber handlers (depends on service interface)
5. `dto.go` — request/response structs with `validate` tags

Exception: `auth` module has no `model.go`/`repository.go` — it uses Redis for sessions and JWT, not a database.

**Response envelope.** All JSON responses use `{ code, statusCode, message, data, path, timestamp, requestId }`. Error codes are i18n-aware (zh-CN / en-US, selected via `X-Lang` or `Accept-Language`).

**Auth model.** JWT access + refresh tokens; sessions stored in Redis as `session:{userId}:{sessionId}` with 7-day TTL. Login attempts tracked in Redis (`login:fail:{username}` with 10-min TTL, `login:lock:{username}`). Middleware chain in `main.go`: `Recovery → Trace → CORS → RequestLogger → ErrorHandler → CSRF → RateLimiter → Metrics`. Route-level guards: `jwtMW` (JWT only) for read-only endpoints, `adminMW` (JWT + RequireRoles("admin")) for write operations.

**Self-referencing trees.** `Menu` and `Department` entities use `ParentID *string` + `Parent *Self` + `Children []Self` with GORM's `BelongsTo`/`HasMany`.

**Redis fallback.** If Redis connection fails at startup, the app degrades to an in-memory cache (`sync.Map`) rather than crashing. This is intentional for dev environments.

**Config priority.** Viper loads config in this order: env vars > `.env.{APP_ENV}` > `.env` > defaults. Key env vars: `APP_ENV`, `PORT`, `DB_TYPE`, `SQLITE_PATH`, `DB_HOST/PORT/USER/PASS/NAME`, `PG_HOST/PORT/USER/PASS/DB`, `REDIS_HOST/PORT/PASS`, `JWT_SECRET`, `JWT_EXPIRES`, `JWT_REFRESH_EXPIRES`, `CORS_ORIGINS`, `THROTTLE_TTL/LIMIT`, `ENABLE_METRICS`, `SWAGGER_ENABLE`, `CSRF_ENABLED/MODE/SECRET`, `UPLOAD_DEST/MAX_SIZE/ALLOWED_TYPES/CLEAN_INTERVAL_SEC/TTL_DAYS`, `LOGIN_MAX_FAILS`, `LOGIN_LOCK_TTL`.

**Database auto-creation.** On startup, `internal/database/database.go` automatically creates the database if it doesn't exist:
- **MySQL**: `CREATE DATABASE IF NOT EXISTS hellogo` (connects without DB name first)
- **PostgreSQL**: connects to `postgres` DB, checks `pg_database`, creates if missing
- **SQLite**: GORM auto-creates the file

This means you don't need to manually create the database before running the app.

## Phase 2: gRPC Microservices (已完成)

Phase 2 splits the monolith into 4 gRPC services + 1 API Gateway:

| Service | Port | Responsibility |
|---|---|---|
| API Gateway | 8000 (HTTP) | HTTP→gRPC forwarding, JWT verification, CORS |
| User Service | 50001 (gRPC) | User CRUD, file upload |
| Auth Service | 50002 (gRPC) | Login, JWT token management, Redis sessions |
| Permission Service | 50003 (gRPC) | Roles, permissions, menus |
| Biz Service | 50004 (gRPC) | Departments, dicts, logs, uploads |

Gateway validates JWT via `Auth.VerifyToken` gRPC call. Proto definitions live in `api/proto/{service}/v1/*.proto`; generated code goes to `gen/go/` (gitignored). Each service uses SQLite by default (`./data/{service}.db`), configurable to MySQL/PostgreSQL via `DB_TYPE` in `{service}.env`.

**Key design decisions:**
- Auth Service does not connect to any database — user queries go through gRPC calls to User Service.
- Phase 1's `internal/pkg/` (response, errors, pagination, redis) migrates to `internal/shared/` (config, database, redis, logger, interceptor) for Phase 2. Phase 1's `internal/config/` and `internal/database/` are preserved for backward compatibility.
- Gateway handlers manually map gRPC response fields to camelCase for frontend compatibility (proto generates snake_case JSON tags).
- Service addresses configured via `internal/shared/config` with `USER_SERVICE_ADDR`, `AUTH_SERVICE_ADDR`, etc.

## Phase 3: Kubernetes Deployment (进行中)

Deploy to **minikube** (local K8s), managed by **Helm**, visualized with **Lens**.

| Component | K8s Resource | External Access |
|---|---|---|
| Frontend (nginx) | Deployment + NodePort Service | `http://$(minikube ip):30090` |
| Gateway | Deployment + NodePort Service | `http://$(minikube ip):30080` |
| User/Auth/Permission/Biz | Deployment + ClusterIP Service | 仅集群内 gRPC |
| MySQL | StatefulSet + ClusterIP Service | 仅集群内 |
| Redis | StatefulSet + ClusterIP Service | 仅集群内 |

所有服务需实现 8080 端口健康检查（`/healthz`, `/readyz`），供 K8s 探针使用。服务发现使用 K8s DNS（无需 etcd）。

## Testing

- Unit tests: mock DB + Redis per service layer (`mockgen` for generating mocks)
- Integration tests: `httptest` + SQLite in-memory DB for full request chain
- E2E tests: docker-compose full stack + shell scripts (`scripts/test_curl.sh`)
- gRPC tests: `grpcurl -plaintext localhost:{port} {service}/{method}`

## Reference Docs

- `docs/phase0-dev-environment-setup.md` — Ubuntu 22.04 + macOS environment setup (Go, Docker, Redis, MySQL, PG, tooling)
- `docs/phase1-fiber-monolith.md` — Phase 1 (Fiber monolith) plan, full API route table, DB schema, time estimates
- `docs/phase2-grpc-microservices.md` — Phase 2 (gRPC) plan with proto definitions, service split, example code for each layer
- `docs/phase3-kubernetes-deployment.md` — Phase 3 (K8s) plan: minikube + Helm + Lens 部署方案

## Language

The user communicates in Chinese. Docs and code comments are in Chinese; identifiers, commit messages, and log keys are in English.
