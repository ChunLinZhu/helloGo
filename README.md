# helloGo

基于 Go + Fiber v2 构建的企业级后台管理系统 API，是 helloNest (NestJS) 项目的 Go 实现。

## 特性

- 🔐 **认证授权**：JWT 双令牌（access + refresh）、RBAC 角色权限控制
- 👥 **用户管理**：CRUD、密码加密、账户锁定机制
- 🏢 **组织架构**：菜单树、部门树、字典管理
- 📁 **文件上传**：分片上传、类型校验、自动清理
- 📊 **可观测性**：Prometheus 指标、Swagger API 文档、请求追踪
- 🐳 **容器化**：Docker 多阶段构建、docker-compose 一键部署
- 🧪 **测试覆盖**：单元测试、集成测试、E2E 测试

## 技术栈

| 组件 | 技术选型 |
|------|---------|
| 框架 | Fiber v2 |
| 数据库 | GORM + SQLite/MySQL/PostgreSQL |
| 缓存 | go-redis v9 |
| 认证 | golang-jwt v5 |
| 配置 | Viper |
| 日志 | Zap + Lumberjack |
| 指标 | Prometheus client_golang |
| API 文档 | swaggo/swag |
| 测试 | testify + mock |

## 快速开始

### 环境要求

- Go 1.21+
- Redis（可选，内存模式可运行）
- Docker & Docker Compose（可选）

### 本地开发

```bash
# 1. 克隆项目
git clone https://github.com/your-org/helloGo.git
cd helloGo

# 2. 安装依赖
go mod download

# 3. 复制配置文件
cp configs/.env.development .env

# 4. 启动基础设施（Redis + MySQL + PostgreSQL）
make infra-up

# 5. 运行数据库迁移和种子数据
make seed

# 6. 启动服务（热重载）
make run-dev
```

服务启动后访问：
- API 地址：http://localhost:8000
- Swagger 文档：http://localhost:8000/swagger/index.html
- Prometheus 指标：http://localhost:8000/metrics

---

## 编译

### 开发环境编译

```bash
# 编译到 build 目录
make build

# 输出位置：build/helloGo
```

### 生产环境编译

```bash
# 使用 ldflags 优化二进制体积
CGO_ENABLED=1 go build -ldflags="-s -w" -o build/helloGo ./cmd/server
```

### 交叉编译（Linux ARM64）

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=1 \
  go build -ldflags="-s -w" -o build/helloGo-linux-arm64 ./cmd/server
```

---

## 打包

### Docker 镜像打包

项目提供多阶段构建的 Dockerfile，最终镜像基于 Alpine Linux，体积约 50MB。

```bash
# 构建镜像
docker build -t hellogo:latest .

# 查看镜像
docker images | grep hellogo
```

### 镜像标签规范

```bash
# 语义化版本
docker build -t hellogo:v1.0.0 .
docker build -t hellogo:v1.0.0-rc1 .

# Git commit 哈希
docker build -t hellogo:$(git rev-parse --short HEAD) .

# 推送镜像（示例）
docker tag hellogo:v1.0.0 registry.example.com/hellogo:v1.0.0
docker push registry.example.com/hellogo:v1.0.0
```

---

## 部署

### 方式一：Docker Compose（推荐）

最简单的一键部署方式，包含应用 + 基础设施。

```bash
# 1. 复制生产环境配置
cp configs/.env.production .env

# 2. 修改 .env 中的敏感配置
# - JWT_SECRET：替换为强随机字符串（至少 32 字符）
# - CSRF_SECRET：替换为强随机字符串
# - DB_PASS：数据库密码
# - REDIS_PASS：Redis 密码

# 3. 一键启动
make docker-up

# 或
docker-compose up -d --build
```

**服务清单**：
- `hellogo-app`：应用服务（端口 8000）
- `hellogo-redis`：Redis 缓存（端口 6379）
- `hellogo-mysql`：MySQL 数据库（端口 3306）
- `hellogo-postgres`：PostgreSQL 数据库（端口 5432）

**常用命令**：

```bash
# 查看日志
make docker-logs

# 停止服务
make docker-down

# 仅启动基础设施（本地开发用）
make infra-up

# 重启应用
docker-compose restart app
```

### 方式二：二进制部署

适用于已有数据库和 Redis 的环境。

```bash
# 1. 编译二进制
make build

# 2. 复制配置
cp configs/.env.production .env

# 3. 创建目录
mkdir -p data upload logs

# 4. 启动服务
./build/helloGo

# 或使用 systemd 管理（推荐）
sudo cp scripts/hellogo.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now hellogo
```

**systemd 服务示例**（`scripts/hellogo.service`）：

```ini
[Unit]
Description=helloGo API Server
After=network.target redis.service mysql.service

[Service]
Type=simple
User=hellogo
WorkingDirectory=/opt/hellogo
ExecStart=/opt/hellogo/build/helloGo
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

### 方式三：Kubernetes 部署

```bash
# 1. 创建 ConfigMap（配置文件）
kubectl create configmap hellogo-config --from-env-file=configs/.env.production

# 2. 创建 Secret（敏感配置）
kubectl create secret generic hellogo-secrets \
  --from-literal=JWT_SECRET=$(openssl rand -hex 32) \
  --from-literal=DB_PASS=$(openssl rand -hex 16) \
  --from-literal=REDIS_PASS=$(openssl rand -hex 16)

# 3. 部署应用（示例 Deployment）
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hellogo
spec:
  replicas: 3
  selector:
    matchLabels:
      app: hellogo
  template:
    metadata:
      labels:
        app: hellogo
    spec:
      containers:
      - name: hellogo
        image: registry.example.com/hellogo:v1.0.0
        ports:
        - containerPort: 8000
        envFrom:
        - configMapRef:
            name: hellogo-config
        - secretRef:
            name: hellogo-secrets
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /api/health
            port: 8000
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/health
            port: 8000
          initialDelaySeconds: 5
          periodSeconds: 5
EOF

# 4. 创建 Service
kubectl expose deployment hellogo --type=ClusterIP --port=8000
```

---

## 测试

### 运行全部测试

```bash
# 运行所有测试（单元测试 + 集成测试）
make test

# 查看详细输出
make test-verbose
```

### 单元测试

测试各个模块的核心逻辑，使用 mock 隔离依赖。

```bash
# 仅运行单元测试
make test-unit

# 或直接使用 go test
go test -v ./internal/pkg/... ./internal/module/... ./internal/guard/...
```

**测试覆盖的模块**：
- `internal/pkg/pagination`：分页工具
- `internal/pkg/errors`：错误码处理
- `internal/pkg/response`：响应格式化
- `internal/module/auth`：JWT 工具
- `internal/module/user`：用户服务
- `internal/guard`：JWT 守卫、角色守卫

### 集成测试

使用 SQLite 内存数据库 + Fiber httptest，测试完整请求链路。

```bash
# 运行集成测试
make test-integration

# 或直接使用 go test
go test -v ./internal/integration/...
```

**测试场景**：
- 健康检查
- 用户 CRUD 完整链路
- 认证保护（401）
- 角色保护（403）
- 重复用户名检测
- 请求体校验失败
- 角色 CRUD

### E2E 测试

使用 curl 脚本测试真实运行的服务。

```bash
# 1. 先启动服务
make run

# 2. 在另一个终端运行 E2E 测试
make test-e2e

# 或指定远程服务地址
bash scripts/test_curl.sh https://api.example.com
```

**测试覆盖的端点**：
- 健康检查
- 认证保护
- 登录获取 token
- 用户管理（CRUD）
- 角色管理
- 权限管理
- 菜单管理
- 字典管理
- Prometheus 指标

### 测试覆盖率

```bash
# 生成覆盖率报告
make test-cover

# 打开 HTML 报告
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

**当前覆盖率**：
- `internal/guard`：96.8%
- `internal/pkg/errors`：100%
- `internal/pkg/pagination`：90.9%
- `internal/pkg/response`：92.3%
- `internal/module/user`：34.0%
- `internal/module/auth`：18.8%
- 总体：9.0%（未测试的模块拉低平均值）

---

## 配置说明

### 环境变量（.env）

```bash
# 应用配置
APP_ENV=production          # 环境：development / production
PORT=8000                   # 监听端口

# 数据库
DB_TYPE=mysql               # sqlite / mysql / postgres
DB_HOST=mysql
DB_PORT=3306
DB_USER=hellogo
DB_PASS=hellogo_password
DB_NAME=hellogo

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASS=redis_password

# JWT
JWT_SECRET=your-secret-key  # 至少 32 字符
JWT_EXPIRES=15m             # access token 过期时间
JWT_REFRESH_EXPIRES=7d      # refresh token 过期时间

# CSRF
CSRF_ENABLED=true
CSRF_SECRET=your-csrf-secret

# 限流
THROTTLE_TTL=60
THROTTLE_LIMIT=100

# 可观测性
ENABLE_METRICS=true
SWAGGER_ENABLE=false        # 生产环境建议关闭

# 文件上传
UPLOAD_DEST=./upload
UPLOAD_MAX_SIZE=10485760    # 10MB
UPLOAD_ALLOWED_TYPES=image/jpeg,image/png,application/pdf
UPLOAD_CLEAN_INTERVAL_SEC=3600
UPLOAD_TTL_DAYS=30

# 登录安全
LOGIN_MAX_FAILS=5
LOGIN_LOCK_TTL=600          # 10 分钟
```

### 配置优先级

1. 环境变量
2. `.env.{APP_ENV}`（如 `.env.production`）
3. `.env`
4. 默认值

---

## API 端点

### 认证

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/api/auth/login` | 登录 | ❌ |
| POST | `/api/auth/refresh` | 刷新 token | ❌ |
| POST | `/api/auth/password/request-reset` | 请求重置密码 | ❌ |
| POST | `/api/auth/password/reset` | 重置密码 | ❌ |
| POST | `/api/auth/unlock` | 解锁账户 | ❌ |
| POST | `/api/auth/logout` | 登出 | ✅ |

### 用户管理（需要 admin 角色）

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/users` | 用户列表（分页） |
| GET | `/api/users/:id` | 用户详情 |
| POST | `/api/users` | 创建用户 |
| PATCH | `/api/users/:id` | 更新用户 |
| DELETE | `/api/users/:id` | 删除用户 |

### 其他模块

- **角色管理**：`/api/roles`
- **权限管理**：`/api/permissions`
- **菜单管理**：`/api/menus`
- **部门管理**：`/api/departments`
- **字典管理**：`/api/dicts`
- **日志管理**：`/api/logs`
- **文件上传**：`/api/uploads`

完整 API 文档请访问 Swagger UI：`http://localhost:8000/swagger/index.html`

---

## 常见问题

### 1. 数据库连接失败

```bash
# 检查 MySQL 是否启动
docker-compose ps mysql

# 查看 MySQL 日志
docker-compose logs mysql

# 测试连接
docker-compose exec mysql mysql -uhellogo -phellogo_password hellogo
```

### 2. Redis 连接失败

项目支持 Redis 降级为内存模式，不会影响核心功能运行。

```bash
# 检查 Redis 是否启动
docker-compose ps redis

# 测试连接
docker-compose exec redis redis-cli -a redis_password ping
```

### 3. JWT token 过期

```bash
# 使用 refresh token 获取新的 access token
curl -X POST http://localhost:8000/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refreshToken": "your-refresh-token", "sessionId": "your-session-id"}'
```

### 4. 如何添加新的 API 端点？

1. 在 `internal/module/<模块名>/` 创建 `handler.go`、`service.go`、`repository.go`
2. 在 `handler.go` 中添加 swagger 注解
3. 在 `cmd/server/main.go` 中注册路由
4. 运行 `make swagger` 生成文档
5. 编写单元测试和集成测试

### 5. 如何扩展新的数据库？

在 `internal/database/database.go` 中添加 GORM 驱动：

```go
import (
  "gorm.io/driver/mysql"
  "gorm.io/driver/postgres"
  "gorm.io/driver/sqlite"
)
```

然后在 `configs/.env.production` 中配置 `DB_TYPE` 和连接参数。

---

## 项目结构

```
helloGo/
├── cmd/
│   ├── server/          # 主服务入口
│   └── seed/            # 种子数据脚本
├── configs/
│   ├── .env.example     # 配置模板
│   ├── .env.development # 开发环境配置
│   └── .env.production  # 生产环境配置
├── internal/
│   ├── config/          # 配置加载
│   ├── database/        # 数据库初始化
│   ├── guard/           # JWT 守卫、角色守卫
│   ├── integration/     # 集成测试
│   ├── middleware/      # 中间件（trace、recovery、CSRF、限流、审计、指标）
│   ├── module/          # 业务模块
│   │   ├── auth/        # 认证
│   │   ├── user/        # 用户
│   │   ├── role/        # 角色
│   │   ├── permission/  # 权限
│   │   ├── menu/        # 菜单
│   │   ├── department/  # 部门
│   │   ├── dict/        # 字典
│   │   ├── log/         # 日志
│   │   └── upload/      # 文件上传
│   └── pkg/             # 公共工具
│       ├── errors/      # 错误码
│       ├── pagination/  # 分页
│       ├── redis/       # Redis 客户端
│       └── response/    # 响应格式化
├── scripts/
│   └── test_curl.sh     # E2E 测试脚本
├── Dockerfile           # 多阶段构建
├── docker-compose.yml   # 完整环境编排
├── docker-compose.infra.yml  # 仅基础设施
├── Makefile             # 构建、测试、部署命令
└── README.md            # 本文档
```

---

## License

MIT
