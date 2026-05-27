# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Purpose

helloGo is a Go reimplementation of the sibling project **helloNest** (`/home/zhuchunlin/github/helloNest`), a NestJS admin backend. Phase 1 (Fiber monolith) is fully implemented. Phase 2 (gRPC microservices) is planned but not yet started.

When behavior is ambiguous, refer to helloNest's implementation.

## Tech Stack

| Concern | Choice |
|---|---|
| Go version | >= 1.22 |
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

## Project Layout

```
cmd/server/main.go        # HTTP entry（路由注册、中间件链）
cmd/seed/main.go          # seed data script
internal/
  config/                 # Viper config
  database/               # GORM init, AutoMigrate
  middleware/             # trace, cors, csrf, ratelimit, recovery, audit, metrics
  module/{auth,user,role,permission,menu,department,dict,log,upload}/
  guard/                  # jwt guard, role guard
  pkg/{response,errors,pagination,redis}/
  integration/            # 集成测试（SQLite 内存 DB + Fiber httptest）
front-end/                # React 前端测试面板
  src/lib/axios.ts        # HTTP 客户端（拦截器注入 token/CSRF）
  src/stores/app.ts       # Zustand store（token/sessionId/apiUrl）
  src/pages/*.tsx         # 11 个 API 测试页面
configs/.env.{example,development,production}
```

Phase 2 (未开始) will restructure into per-service binaries under `cmd/{gateway,user,auth,permission,biz}/` with shared protobuf definitions.

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

**Response envelope.** All JSON responses use `{ code, statusCode, message, data, path, timestamp, requestId }`. Error codes are i18n-aware (zh-CN / en-US, selected via `X-Lang` or `Accept-Language`).

**Auth model.** JWT access + refresh tokens; sessions stored in Redis as `session:{userId}:{sessionId}` with 7-day TTL. Login attempts tracked in Redis (`login:fail:{username}` with 10-min TTL, `login:lock:{username}`). Guard chain order: JWT → Roles → Permissions → Throttle → Audit.

**Self-referencing trees.** `Menu` and `Department` entities use `ParentID *string` + `Parent *Self` + `Children []Self` with GORM's `BelongsTo`/`HasMany`.

**Redis fallback.** If Redis connection fails at startup, the app degrades to an in-memory cache (`sync.Map`) rather than crashing. This is intentional for dev environments.

**Config priority.** Viper loads config in this order: env vars > `.env.{APP_ENV}` > `.env` > defaults. Key env vars: `APP_ENV`, `PORT`, `DB_TYPE`, `SQLITE_PATH`, `REDIS_HOST/PORT/PASS`, `JWT_SECRET`, `JWT_EXPIRES`, `CORS_ORIGINS`, `THROTTLE_TTL/LIMIT`, `ENABLE_METRICS`, `SWAGGER_ENABLE`, `CSRF_ENABLED`, `UPLOAD_DEST/MAX_SIZE/ALLOWED_TYPES`, `LOGIN_MAX_FAILS`, `LOGIN_LOCK_TTL`.

## Phase 2: gRPC Microservices

Phase 2 splits the monolith into 4 gRPC services + 1 API Gateway:

| Service | Port | Responsibility |
|---|---|---|
| API Gateway | 8000 (HTTP) | HTTP→gRPC forwarding, JWT verification, rate limiting |
| User Service | 50001 (gRPC) | User CRUD, file upload |
| Auth Service | 50002 (gRPC) | Login, JWT token management, Redis sessions |
| Permission Service | 50003 (gRPC) | Roles, permissions, menu permissions |
| Biz Service | 50004 (gRPC) | Menus, departments, dicts, logs |

Gateway validates JWT via `Auth.VerifyToken` gRPC call and checks permissions via `Permission.CheckPermission`. Proto definitions live in `api/proto/{service}/v1/*.proto`; generated code goes to `gen/go/` (gitignored). Service discovery uses etcd with TTL-based registration; load balancing uses gRPC's built-in `round_robin`. Observability via OpenTelemetry + Jaeger for distributed tracing.

## Testing

- Unit tests: mock DB + Redis per service layer (`mockgen` for generating mocks)
- Integration tests: `httptest` + SQLite in-memory DB for full request chain
- E2E tests: docker-compose full stack + shell scripts (`scripts/test_curl.sh`)
- gRPC tests: `grpcurl -plaintext localhost:{port} {service}/{method}`

## Reference Docs

- `docs/development-plan.md` — Phase 1 (Fiber monolith) plan, full API route table, DB schema, time estimates
- `docs/phase2-grpc-microservices.md` — Phase 2 (gRPC) plan with proto definitions, service split, example code for each layer
- `docs/dev-environment-setup.md` — Ubuntu 22.04 + macOS environment setup (Go, Docker, Redis, MySQL, PG, tooling)

## Language

The user communicates in Chinese. Docs and code comments are in Chinese; identifiers, commit messages, and log keys are in English.
