# 第一阶段开发计划：Fiber 单体应用

> 基于 helloNest (NestJS) 项目，使用 Go + Fiber 构建等效后端 API

## 技术栈

| 组件         | 技术选型                          | 说明                                     |
| ------------ | --------------------------------- | ---------------------------------------- |
| Web 框架     | [Fiber](https://gofiber.io/) v2   | 高性能 HTTP 框架，基于 fasthttp          |
| ORM          | [GORM](https://gorm.io/)          | Go 主流 ORM，支持多数据库                |
| 数据库       | SQLite（默认）/ MySQL / PostgreSQL | 与原项目保持一致                         |
| 缓存         | [go-redis](https://redis.uptrace.dev/) v9 | Redis 客户端                      |
| 认证         | [golang-jwt](https://github.com/golang-jwt/jwt) v5 | JWT 令牌生成与验证            |
| 密码加密     | `golang.org/x/crypto/bcrypt`      | bcrypt 哈希                              |
| 配置管理     | [Viper](https://github.com/spf13/viper) | 支持 .env / YAML / 环境变量          |
| 日志         | [Zap](https://github.com/uber-go/zap) + [Lumberjack](https://github.com/natefinch/lumberjack) | 结构化日志 + 日志轮转 |
| 参数校验     | [Validator](https://github.com/go-playground/validator) v10 | Fiber 内置集成                  |
| Prometheus   | [prometheus/client_golang](https://github.com/prometheus/client_golang) | 指标采集与暴露        |
| API 文档     | [swaggo/swag](https://github.com/swaggo/swag) + fiber-swagger | Swagger 自动生成               |
| UUID         | [google/uuid](https://github.com/google/uuid) | UUID v4 生成                    |
| 文件上传     | Fiber 内置 `multipart`            | 磁盘存储 + 分块上传                      |
| 容器化       | Docker + Docker Compose           | 多阶段构建                               |

## 项目结构

```
helloGo/
├── cmd/
│   ├── server/
│   │   └── main.go                 # 入口文件
│   └── seed/
│       └── main.go                 # 种子数据脚本
├── internal/
│   ├── config/
│   │   └── config.go               # Viper 配置加载
│   ├── database/
│   │   ├── database.go             # GORM 初始化（多数据库）
│   │   └── migration.go            # AutoMigrate
│   ├── middleware/
│   │   ├── trace.go                # X-Trace-Id 中间件
│   │   ├── cors.go                 # CORS 配置
│   │   ├── csrf.go                 # CSRF 防护
│   │   ├── ratelimit.go            # 限流中间件
│   │   ├── recovery.go             # Panic 恢复
│   │   ├── request_logger.go       # 请求日志
│   │   ├── error_handler.go        # 全局错误处理（AppError/FiberError/未知错误）
│   │   ├── metrics.go              # Prometheus 指标中间件
│   │   └── audit.go                # 审计日志持久化（写入数据库）
│   ├── module/
│   │   ├── auth/
│   │   │   ├── handler.go          # 路由处理函数
│   │   │   ├── service.go          # 业务逻辑
│   │   │   ├── jwt.go              # JWT 工具
│   │   │   └── dto.go              # 请求/响应结构体
│   │   ├── user/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── model.go            # User 实体
│   │   │   └── dto.go
│   │   ├── role/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── model.go            # Role 实体
│   │   │   └── dto.go
│   │   ├── permission/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── model.go            # Permission 实体
│   │   │   └── dto.go
│   │   ├── menu/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── model.go            # Menu 实体（自引用树）
│   │   ├── department/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── model.go            # Department 实体（自引用树）
│   │   ├── dict/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── model.go            # Dict 实体
│   │   ├── log/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── model.go            # Log 实体
│   │   ├── upload/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── model.go            # Upload 实体
│   │   │   └── cleanup.go          # 定时清理
│   │   ├── health/                  # 健康检查（目录保留，逻辑在 main.go 中实现）
│   │   └── metrics/                 # Prometheus 指标（目录保留，逻辑在 main.go 中实现）
│   ├── guard/
│   │   ├── jwt_guard.go            # JWT 认证守卫
│   │   ├── role_guard.go           # 角色守卫
│   │   └── permission_guard.go     # 权限守卫
│   └── pkg/
│       ├── response/
│       │   └── response.go         # 统一响应格式
│       ├── errors/
│       │   └── errors.go           # 错误码定义 + i18n
│       ├── pagination/
│       │   └── pagination.go       # 分页工具
│       └── redis/
│           └── redis.go            # Redis 客户端封装
├── configs/
│   ├── .env.example
│   ├── .env.development
│   └── .env.production
├── upload/                          # 上传文件存储目录
├── data/                            # SQLite 数据文件
├── docs/                            # 文档目录
├── scripts/
│   └── test_curl.sh
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

## 开发阶段

---

### Phase 1: 项目初始化与基础设施

**目标：** 搭建项目骨架，完成配置、数据库连接、基础中间件

#### 1.1 项目初始化

- [ ] `go mod init helloGo`
- [ ] 安装核心依赖：fiber, gorm, viper, zap, go-redis, golang-jwt, bcrypt, uuid, validator
- [ ] 创建目录结构
- [ ] 编写 Makefile（build / run / test / lint / swagger 命令）

#### 1.2 配置管理（`internal/config/config.go`）

- [ ] 使用 Viper 加载配置，优先级：环境变量 > `.env.{NODE_ENV}` > `.env`
- [ ] 定义 Config 结构体，涵盖所有配置项：

| 分类       | 配置项                                                  |
| ---------- | ------------------------------------------------------- |
| 通用       | `APP_ENV`, `PORT`                                       |
| 数据库     | `DB_TYPE`, `SQLITE_PATH`, `DB_HOST/PORT/USER/PASS/NAME` |
| Redis      | `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASS`                |
| JWT        | `JWT_SECRET`, `JWT_EXPIRES`, `JWT_REFRESH_EXPIRES`      |
| 安全       | `CSRF_ENABLED`, `CSRF_MODE`, `CSRF_SECRET`, `CORS_ORIGINS` |
| 限流       | `THROTTLE_TTL`, `THROTTLE_LIMIT`                        |
| 指标       | `ENABLE_METRICS`                                        |
| Swagger    | `SWAGGER_ENABLE`                                        |
| 上传       | `UPLOAD_DEST`, `UPLOAD_MAX_SIZE`, `UPLOAD_ALLOWED_TYPES` |
| 登录安全   | `LOGIN_MAX_FAILS`, `LOGIN_LOCK_TTL`                     |

#### 1.3 数据库连接（`internal/database/`）

- [ ] GORM 初始化，根据 `DB_TYPE` 切换 SQLite / MySQL / PostgreSQL
- [ ] 启动时自动创建数据库（MySQL: `CREATE DATABASE IF NOT EXISTS`，PostgreSQL: 检查 `pg_database` 后创建）
- [ ] AutoMigrate 注册所有模型
- [ ] 连接池配置（MaxOpenConns, MaxIdleConns, ConnMaxLifetime）

#### 1.4 Redis 连接（`internal/pkg/redis/`）

- [ ] go-redis v9 客户端初始化
- [ ] 封装 `Get`, `Set`（带 TTL）, `Del`, `Ping` 方法
- [ ] 连接失败时降级为内存缓存（sync.Map）

#### 1.5 基础中间件

- [ ] **Trace 中间件** — 读取或生成 `X-Trace-Id`（UUID），注入请求上下文和响应头
- [ ] **Recovery 中间件** — 捕获 panic，返回 500 结构化响应
- [ ] **CORS 中间件** — 根据 `CORS_ORIGINS` 配置允许来源，启用 credentials
- [ ] **请求日志中间件** — 记录每个请求的 method, path, status, duration, traceId, user

#### 1.6 统一响应与错误处理

- [ ] 统一 JSON 响应格式：

```json
{
  "code": "OK",
  "statusCode": 200,
  "message": "success",
  "data": {},
  "path": "/api/users",
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}
```

- [ ] 错误码枚举（ErrorCode）+ i18n 支持（zh-CN / en-US），通过 `X-Lang` 或 `Accept-Language` 头选择
- [ ] 全局错误处理中间件，统一捕获并格式化错误

#### 1.7 分页工具

- [ ] `Pagination` 结构体：`Page`（默认 1）、`Limit`（默认 10）
- [ ] 从 query string 自动绑定并校验
- [ ] GORM Scope 封装：`Paginate(page, limit)` 返回 `Offset` + `Limit`

#### 1.8 启动入口（`cmd/server/main.go`）

- [ ] 加载配置 → 初始化数据库 → 初始化 Redis → 注册路由 → 启动 Fiber
- [ ] 优雅关闭（`os.Signal` + `app.Shutdown()`）

---

### Phase 2: 数据模型与种子数据

**目标：** 定义所有 GORM 模型，完成种子数据脚本

#### 2.1 GORM 模型定义

**User（`internal/module/user/model.go`）**

| 字段         | 类型         | 约束                      |
| ------------ | ------------ | ------------------------- |
| ID           | string (UUID) | PrimaryKey               |
| Username     | string(64)   | UniqueIndex               |
| PasswordHash | string(128)  |                           |
| Email        | *string(128) | Index                     |
| Phone        | *string(32)  | Index                     |
| IsActive     | bool         | Default: true             |
| CreatedAt    | time.Time    | AutoCreateTime            |
| UpdatedAt    | time.Time    | AutoUpdateTime            |
| Roles        | []Role       | ManyToMany (user_roles)   |

**Role（`internal/module/role/model.go`）**

| 字段        | 类型         | 约束                       |
| ----------- | ------------ | -------------------------- |
| ID          | string (UUID) | PrimaryKey                |
| Code        | string(64)   | UniqueIndex                |
| Name        | string(128)  |                            |
| Description | *string(255) |                            |
| Permissions | []Permission | HasMany                    |
| Users       | []User       | ManyToMany (user_roles)    |

**Permission（`internal/module/permission/model.go`）**

| 字段        | 类型         | 约束                       |
| ----------- | ------------ | -------------------------- |
| ID          | string (UUID) | PrimaryKey                |
| Key         | string(128)  | UniqueIndex                |
| Description | *string(256) |                            |
| RoleID      | string       | ForeignKey → roles         |

**Menu（`internal/module/menu/model.go`）** — 自引用树

| 字段     | 类型         | 约束                       |
| -------- | ------------ | -------------------------- |
| ID       | string (UUID) | PrimaryKey                |
| Name     | string(128)  |                            |
| Path     | *string(256) | UniqueIndex                |
| Icon     | *string(128) |                            |
| Order    | int          | Default: 0, Index          |
| ParentID | *string      | ForeignKey → menus         |
| Parent   | *Menu        | BelongsTo                  |
| Children | []Menu       | HasMany                    |

**Department（`internal/module/department/model.go`）** — 自引用树

| 字段        | 类型         | 约束                       |
| ----------- | ------------ | -------------------------- |
| ID          | string (UUID) | PrimaryKey                |
| Name        | string(128)  |                            |
| Description | *string(256) |                            |
| ParentID    | *string      | ForeignKey → departments   |
| Parent      | *Department  | BelongsTo                  |
| Children    | []Department | HasMany                    |

**Dict（`internal/module/dict/model.go`）**

| 字段        | 类型         | 约束                       |
| ----------- | ------------ | -------------------------- |
| ID          | string (UUID) | PrimaryKey                |
| Type        | string(128)  | CompositeUnique (type,key) |
| Key         | string(128)  |                            |
| Value       | string(256)  |                            |
| Description | *string(255) |                            |

**Log（`internal/module/log/model.go`）**

| 字段      | 类型         | 约束                       |
| --------- | ------------ | -------------------------- |
| ID        | string (UUID) | PrimaryKey                |
| Level     | string(16)   | info/warn/error/debug      |
| Message   | string(256)  |                            |
| Meta      | *string      | Type: text                 |
| CreatedAt | time.Time    | CompositeIndex (level, created_at) |

**Upload（`internal/module/upload/model.go`）**

| 字段         | 类型         | 约束             |
| ------------ | ------------ | ---------------- |
| ID           | string (UUID) | PrimaryKey      |
| Filename     | string       |                  |
| OriginalName | string       |                  |
| Mimetype     | string       |                  |
| Size         | int64        |                  |
| Path         | string       |                  |
| CreatedAt    | time.Time    | AutoCreateTime   |

#### 2.2 种子数据（`cmd/seed/main.go`）

- [ ] 可选 `--purge` 参数清除现有数据
- [ ] 创建 admin 角色 + `user:create`, `role:manage` 权限
- [ ] 创建 admin 用户（密码 `admin123`），关联 admin 角色
- [ ] 批量创建 100 个角色 / 权限 / 用户 / 菜单 / 部门 / 字典 / 日志

---

### Phase 3: 认证与授权

**目标：** 实现完整的 JWT 认证 + 角色/权限授权体系

#### 3.1 JWT 工具（`internal/module/auth/jwt.go`）

- [ ] `GenerateAccessToken(userID, username, roles)` — 生成 access token
- [ ] `GenerateRefreshToken(userID, sessionID)` — 生成 refresh token（`typ: "refresh"`）
- [ ] `ValidateToken(tokenString)` — 解析并验证 JWT claims
- [ ] Token payload：`{ sub, username, roles, exp }`

#### 3.2 认证处理（`internal/module/auth/`）

| 端点                              | 方法 | 访问     | 说明                    |
| --------------------------------- | ---- | -------- | ----------------------- |
| `POST /api/auth/login`            | POST | Public   | 用户名/密码登录         |
| `POST /api/auth/refresh`          | POST | Public   | 刷新 access token       |
| `POST /api/auth/logout`           | POST | Auth     | 注销会话                |
| `POST /api/auth/password/request-reset` | POST | Public | 请求密码重置令牌  |
| `POST /api/auth/password/reset`   | POST | Public   | 确认密码重置            |
| `POST /api/auth/unlock`           | POST | Public   | 解锁账户                |

**登录安全逻辑：**
1. 检查 Redis 锁：`login:lock:{username}`
2. 验证 bcrypt 密码
3. 失败计数：`login:fail:{username}`（TTL 10min），超过 `LOGIN_MAX_FAILS` 次后锁定 `LOGIN_LOCK_TTL` 秒
4. 成功后清除计数，生成 token 对，存储 session 到 Redis（`session:{userId}:{sessionId}`，TTL 7 天）
5. 返回 `{ accessToken, refreshToken, sessionId }`

#### 3.3 认证守卫（`internal/guard/`）

**JWT Guard（`jwt_guard.go`）**
- [ ] 从 `Authorization: Bearer <token>` 提取 token
- [ ] 解析 JWT，将用户信息注入 Fiber context（`c.Locals`）
- [ ] 支持 `Public` 标记跳过认证（通过路由 metadata 或 path 白名单）

**Role Guard（`role_guard.go`）**
- [ ] 读取路由所需角色（通过 `c.Route().Name` 或自定义 handler wrapper）
- [ ] 比对 JWT payload 中的 roles

**Permission Guard（`permission_guard.go`）**
- [ ] 读取路由所需权限 key
- [ ] 从 Redis 缓存加载角色权限（`rolePerms:{roleCode}`，TTL 300s），缓存未命中则查 DB
- [ ] 支持 `:own` 后缀的资源级权限（仅允许操作自己的资源）

#### 3.4 守卫集成方式

Fiber 没有 NestJS 的装饰器机制，采用以下方式实现：

```go
// 路由级守卫 — 通过中间件链组合
auth := app.Group("/api", jwtGuard)             // 全局 JWT
auth.Get("/users", requireRoles("admin"), requirePerms("user:list"), listUsers)

// Public 路由 — 不加 jwtGuard
app.Post("/api/auth/login", login)
```

---

### Phase 4: 业务模块 CRUD

**目标：** 实现所有业务模块的增删改查 API

#### 4.1 Users 模块

| 端点                  | 方法  | 权限                     | 说明           |
| --------------------- | ----- | ------------------------ | -------------- |
| `GET /api/users`      | GET   | admin                    | 分页查询用户   |
| `GET /api/users/:id`  | GET   | admin                    | 按 ID 查询     |
| `POST /api/users`     | POST  | admin + `user:create`    | 创建用户       |
| `PATCH /api/users/:id`| PATCH | admin                    | 更新用户       |
| `DELETE /api/users/:id`|DELETE| admin                    | 删除用户       |

#### 4.2 Roles 模块

| 端点                        | 方法 | 权限                     | 说明             |
| --------------------------- | ---- | ------------------------ | ---------------- |
| `GET /api/roles`            | GET  | admin                    | 分页查询角色     |
| `POST /api/roles`           | POST | admin + `role:manage`    | 创建角色         |
| `POST /api/roles/permission`| POST | admin + `role:manage`    | 为角色添加权限   |

#### 4.3 Permissions 模块

| 端点                          | 方法   | 权限  | 说明           |
| ----------------------------- | ------ | ----- | -------------- |
| `GET /api/permissions`        | GET    | admin | 分页查询       |
| `POST /api/permissions`       | POST   | admin | 创建权限       |
| `PATCH /api/permissions/:id`  | PATCH  | admin | 更新权限       |
| `DELETE /api/permissions/:id` | DELETE | admin | 删除权限       |
| `POST /api/permissions/cache/evict` | POST | admin | 清除权限缓存 |

#### 4.4 Menus 模块

| 端点           | 方法 | 权限  | 说明                         |
| -------------- | ---- | ----- | ---------------------------- |
| `GET /api/menus` | GET | admin | 查询菜单树（含 parent/children） |

#### 4.5 Departments 模块

| 端点                  | 方法 | 权限  | 说明                       |
| --------------------- | ---- | ----- | -------------------------- |
| `GET /api/departments` | GET | admin | 查询部门树（含 parent/children） |

#### 4.6 Dicts 模块

| 端点            | 方法 | 权限  | 说明           |
| --------------- | ---- | ----- | -------------- |
| `GET /api/dicts` | GET | admin | 分页查询字典   |

#### 4.7 Logs 模块

| 端点            | 方法 | 权限  | 说明           |
| --------------- | ---- | ----- | -------------- |
| `GET /api/logs`  | GET | admin | 分页查询日志   |
| `POST /api/logs` | POST| admin | 创建日志       |

---

### Phase 5: 文件上传

**目标：** 实现文件上传（含分块上传）和自动清理

#### 5.1 Upload API

| 端点                    | 方法 | 访问 | 说明               |
| ----------------------- | ---- | ---- | ------------------ |
| `POST /api/uploads`     | POST | Auth | 上传文件（max 5MB）|
| `GET /api/uploads`      | GET  | Auth | 查询上传记录       |
| `POST /api/uploads/chunk` | POST | Auth | 上传单个分块     |
| `POST /api/uploads/merge` | POST | Auth | 合并分块         |

#### 5.2 功能要点

- [ ] 磁盘存储，目录由 `UPLOAD_DEST` 配置
- [ ] 文件大小限制：`UPLOAD_MAX_SIZE`（默认 5MB）
- [ ] MIME 类型白名单：`UPLOAD_ALLOWED_TYPES`（默认 jpeg, png, pdf）
- [ ] Magic number 校验（读取文件头验证真实类型）
- [ ] 分块上传：chunk 存临时目录 → merge 合并 → 清理临时文件
- [ ] 定时清理：每 `UPLOAD_CLEAN_INTERVAL_SEC` 秒，删除超过 `UPLOAD_TTL_DAYS` 天的文件和记录

#### 5.3 静态文件服务

- [ ] Fiber `Static` 中间件提供 `/upload/` 目录的静态文件访问

---

### Phase 6: 安全与防护

**目标：** 实现 CSRF、限流、审计等安全机制

#### 6.1 CSRF 防护（`internal/middleware/csrf.go`）

- [ ] **Header 模式**：生成 JWT 签名的 CSRF token（30 分钟有效），通过 `GET /api/csrf-token` 获取
- [ ] **Cookie 模式**：使用 Fiber 内置 CSRF 中间件
- [ ] 非 GET/HEAD/OPTIONS 请求验证 `X-CSRF-Token` 头

#### 6.2 限流（`internal/middleware/ratelimit.go`）

- [ ] 使用 Fiber 内置 `limiter` 中间件
- [ ] 配置：`THROTTLE_TTL`（默认 60s），`THROTTLE_LIMIT`（默认 100 次）
- [ ] 支持 Redis 作为限流存储后端

#### 6.3 审计拦截器（`internal/middleware/request_logger.go`）

- [ ] 记录每个请求到 logs 表，meta 包含：traceId, method, path, status, duration, userID, username
- [ ] 同时写入 Prometheus 指标

---

### Phase 7: 运维与可观测性

**目标：** 健康检查、Prometheus 指标、Swagger 文档

#### 7.1 Health 检查

| 端点                   | 方法 | 访问 | 说明                          |
| ---------------------- | ---- | ---- | ----------------------------- |
| `GET /api/health`      | GET  | Public | 存活检查：`{ "status": "ok" }` |
| `GET /api/health/ready`| GET  | Public | 就绪检查：验证 DB + Redis 连接 |

#### 7.2 Prometheus 指标

- [ ] 启用条件：`ENABLE_METRICS=true`
- [ ] 自定义指标：
  - `http_requests_total` — Counter（labels: method, path, status）
  - `http_request_duration_ms` — Histogram（buckets: 50, 100, 200, 500, 1000, 2000, 5000）
- [ ] Go runtime 默认指标（goroutine, GC, memory）
- [ ] 暴露端点：`GET /api/metrics`

#### 7.3 Swagger 文档

- [ ] 使用 swaggo 注解所有 handler
- [ ] 启用条件：`SWAGGER_ENABLE=true`
- [ ] 访问地址：`/docs`

---

### Phase 8: 容器化与部署

**目标：** Docker 多阶段构建 + docker-compose 完整环境

#### 8.1 Dockerfile（多阶段构建）

```dockerfile
# Stage 1: Build
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/server

# Stage 2: Runtime
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/configs ./configs
EXPOSE 8000
CMD ["./server"]
```

#### 8.2 docker-compose.yml

| 服务     | 镜像              | 端口  |
| -------- | ----------------- | ----- |
| app      | 自构建            | 8000  |
| redis    | redis:7-alpine    | 6379  |
| mysql    | mysql:8           | 3306  |
| postgres | postgres:15-alpine| 5432  |

#### 8.3 Makefile 命令

```makefile
build       # go build
run         # go run cmd/server/main.go
test        # go test ./...
lint        # golangci-lint run
swagger     # swag init
seed        # go run cmd/seed/main.go
docker-up   # docker-compose up -d
docker-down # docker-compose down
migrate     # gorm AutoMigrate
```

---

### Phase 9: 测试

**目标：** 单元测试 + 集成测试 + E2E 测试

#### 9.1 单元测试

- [ ] 每个 service 层编写单元测试（mock DB + Redis）
- [ ] JWT 工具测试
- [ ] 权限守卫测试
- [ ] 分页工具测试

#### 9.2 集成测试

- [ ] 使用 `httptest` + SQLite 内存数据库
- [ ] 测试完整请求链路（中间件 → 守卫 → handler → service → DB）

#### 9.3 E2E 测试

- [ ] 使用 docker-compose 启动完整环境
- [ ] Shell 脚本自动化测试（`scripts/test_curl.sh`）

---

## 开发优先级与时间估算

| 阶段   | 内容                   | 优先级 | 预计时间 |
| ------ | ---------------------- | ------ | -------- |
| Phase 1 | 项目初始化与基础设施  | P0     | 2 天     |
| Phase 2 | 数据模型与种子数据    | P0     | 1 天     |
| Phase 3 | 认证与授权            | P0     | 2-3 天   |
| Phase 4 | 业务模块 CRUD         | P0     | 3-4 天   |
| Phase 5 | 文件上传              | P1     | 1-2 天   |
| Phase 6 | 安全与防护            | P1     | 1-2 天   |
| Phase 7 | 运维与可观测性        | P1     | 1 天     |
| Phase 8 | 容器化与部署          | P2     | 1 天     |
| Phase 9 | 测试                  | P2     | 2-3 天   |
| **合计** |                       |        | **14-19 天** |

## Go 与 NestJS 关键差异适配

| NestJS 特性           | Go + Fiber 替代方案                              |
| --------------------- | ------------------------------------------------ |
| 装饰器 `@Public()`    | 路由分组 + 中间件链（Public 路由不加 jwtGuard）  |
| 装饰器 `@Roles()`     | 自定义 middleware wrapper `requireRoles(...)`    |
| 装饰器 `@Permissions()`| 自定义 middleware wrapper `requirePerms(...)`   |
| DI 容器（IoC）        | 构造函数注入 + 接口抽象（Service 接口 + 实现）   |
| Module 系统           | Go package + `RegisterRoutes(app, deps)` 函数    |
| Pipe（ValidationPipe）| Fiber 内置 `BodyParser` + `validator` tag 校验   |
| Interceptor           | Fiber 中间件（Next 前后分别处理）                |
| Exception Filter      | Fiber 全局错误处理中间件 `app.Use(errorHandler)` |
| Guard                 | Fiber 中间件链                                   |
| TypeORM               | GORM                                             |
| Passport JWT          | golang-jwt + 自定义中间件                        |

## API 路由总览

```
Public:
  POST   /api/auth/login
  POST   /api/auth/refresh
  POST   /api/auth/password/request-reset
  POST   /api/auth/password/reset
  POST   /api/auth/unlock
  GET    /api/csrf-token
  GET    /api/health
  GET    /api/health/ready
  GET    /api/metrics

Authenticated (JWT):
  POST   /api/auth/logout

  GET    /api/dicts
  GET    /api/dicts/:id
  GET    /api/logs
  GET    /api/logs/:id
  GET    /api/menus/tree
  GET    /api/departments/tree

  POST   /api/uploads
  POST   /api/uploads/chunk
  POST   /api/uploads/merge
  GET    /api/uploads
  GET    /api/uploads/:id

Admin:
  GET    /api/users
  GET    /api/users/:id
  POST   /api/users
  PATCH  /api/users/:id
  DELETE /api/users/:id

  GET    /api/roles
  GET    /api/roles/:id
  POST   /api/roles
  PATCH  /api/roles/:id
  DELETE /api/roles/:id
  POST   /api/roles/:id/permissions

  GET    /api/permissions
  GET    /api/permissions/:id
  POST   /api/permissions
  PATCH  /api/permissions/:id
  DELETE /api/permissions/:id

  GET    /api/menus/:id
  POST   /api/menus
  PATCH  /api/menus/:id
  DELETE /api/menus/:id

  GET    /api/departments/:id
  POST   /api/departments
  PATCH  /api/departments/:id
  DELETE /api/departments/:id

  POST   /api/dicts
  PATCH  /api/dicts/:id
  DELETE /api/dicts/:id

  POST   /api/logs
  DELETE /api/uploads/:id
```
