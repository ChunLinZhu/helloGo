# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Status

This project is in the **planning / early implementation phase**. The repo currently contains only planning documents in `docs/` — no Go source code, `go.mod`, or `Makefile` exists yet. When implementing, follow the plans in `docs/` as the source of truth for architecture and conventions.

## Purpose

helloGo is a Go reimplementation of the sibling project **helloNest** (`/Users/zhuchunlin/Documents/github/helloNest`), a NestJS admin backend. The goal is to produce feature-equivalent REST (Phase 1) and gRPC microservice (Phase 2) APIs in Go.

When behavior is ambiguous, refer to helloNest's implementation — especially:
- `src/modules/*` — business logic per domain
- `src/common/*` — guards, interceptors, decorators, DTOs
- `src/database/*` — TypeORM setup, multi-DB support
- `src/main.ts`, `src/app.module.ts` — bootstrap and module wiring

## Planned Tech Stack

| Concern | Choice |
|---|---|
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

## Planned Layout

```
cmd/server/main.go        # HTTP entry
cmd/seed/main.go          # seed data script
internal/
  config/                 # Viper config
  database/               # GORM init, AutoMigrate
  middleware/             # trace, cors, csrf, ratelimit, recovery, audit
  module/{auth,user,role,permission,menu,department,dict,log,upload,health,metrics}/
                          # handler.go, service.go, model.go, dto.go
  guard/                  # jwt, role, permission (Phase 1 middleware wrappers)
  pkg/{response,errors,pagination,redis}/
configs/.env.{example,development,production}
```

Phase 2 restructures this into per-service binaries under `cmd/{gateway,user,auth,permission,biz}/` with shared protobuf definitions in `api/proto/**/v1/*.proto` and generated code in `gen/go/`.

## Planned Commands (from Makefile design)

```
make run          # go run cmd/server/main.go
make build        # go build -o server ./cmd/server
make test         # go test ./...
make test-cover   # go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
make lint         # golangci-lint run ./...
make swagger      # swag init -g cmd/server/main.go
make seed         # go run cmd/seed/main.go
make proto        # bash scripts/gen-proto.sh  (Phase 2)
make infra-up     # docker compose -f docker-compose.infra.yml up -d
make all-up       # docker compose up --build -d
```

Dev workflow: `air` for hot reload (config in `.air.toml`).

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

**Module structure.** Each `internal/module/<name>/` follows: `model.go` (GORM struct) → `repository.go` (DB queries) → `service.go` (business logic, depends on repo) → `handler.go` (Fiber handlers, depends on service) → `dto.go` (request/response structs with `validate` tags).

**Response envelope.** All JSON responses use `{ code, statusCode, message, data, path, timestamp, requestId }`. Error codes are i18n-aware (zh-CN / en-US, selected via `X-Lang` or `Accept-Language`).

**Auth model.** JWT access + refresh tokens; sessions stored in Redis as `session:{userId}:{sessionId}` with 7-day TTL. Login attempts tracked in Redis (`login:fail:{username}`, `login:lock:{username}`). Guard chain order: JWT → Roles → Permissions → Throttle → Audit.

**Self-referencing trees.** `Menu` and `Department` entities use `ParentID *string` + `Parent *Self` + `Children []Self` with GORM's `BelongsTo`/`HasMany`.

## Reference Docs

- `docs/development-plan.md` — Phase 1 (Fiber monolith) plan, full API route table, DB schema, time estimates
- `docs/phase2-grpc-microservices.md` — Phase 2 (gRPC) plan with proto definitions, service split, example code for each layer
- `docs/dev-environment-setup.md` — Ubuntu 22.04 + macOS environment setup (Go, Docker, Redis, MySQL, PG, tooling)

## Language

The user communicates in Chinese. Docs and code comments are in Chinese; identifiers, commit messages, and log keys are in English.
