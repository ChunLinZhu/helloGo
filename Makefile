# ============================================================
# helloGo Makefile — 构建、运行、测试、代码质量命令
# ============================================================

# ── 变量 ──────────────────────────────────────────────────
APP_NAME    := helloGo
CMD_DIR     := ./cmd/server
SEED_CMD    := ./cmd/seed
BUILD_DIR   := ./build
BINARY      := $(BUILD_DIR)/$(APP_NAME)
GO          := go
GOLANGCI    := golangci-lint

# ── 构建 ──────────────────────────────────────────────────

## build: 编译服务端二进制文件
build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BINARY) $(CMD_DIR)

## run: 启动开发服务器（go run）
run:
	$(GO) run $(CMD_DIR)/main.go

## run-dev: 使用 air 热重载启动（需要先安装 air）
run-dev:
	air

# ── 测试 ──────────────────────────────────────────────────

## test: 运行所有测试
test:
	$(GO) test ./... -count=1

## test-cover: 运行测试并生成覆盖率报告
test-cover:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告: coverage.html"

## test-verbose: 详细模式运行测试（禁用缓存）
test-verbose:
	$(GO) test -v -count=1 ./...

## test-unit: 仅运行单元测试（排除集成测试）
test-unit:
	$(GO) test -count=1 ./internal/pkg/... ./internal/module/... ./internal/guard/...

## test-integration: 仅运行集成测试
test-integration:
	$(GO) test -v -count=1 ./internal/integration/...

## test-e2e: 运行 E2E curl 测试（需先启动服务）
test-e2e:
	bash scripts/test_curl.sh

# ── 代码质量 ──────────────────────────────────────────────

## lint: 运行 golangci-lint 静态检查
lint:
	$(GOLANGCI) run ./...

## fmt: 格式化代码（goimports 自动整理 import）
fmt:
	goimports -w .

# ── 文档 ──────────────────────────────────────────────────

## swagger: 生成 Swagger API 文档
swagger:
	swag init -g $(CMD_DIR)/main.go --parseDependency --parseInternal

# ── 数据库 ────────────────────────────────────────────────

## seed: 运行种子数据脚本
seed:
	$(GO) run $(SEED_CMD)/main.go

## seed-purge: 清除现有数据后重新播种
seed-purge:
	$(GO) run $(SEED_CMD)/main.go --purge

# ── Docker ────────────────────────────────────────────────

## docker-up: 启动所有服务（应用 + 基础设施）
docker-up:
	docker-compose up --build -d

## docker-down: 停止所有服务
docker-down:
	docker-compose down

## all-up: 启动完整环境（等同 docker-up）
all-up:
	docker-compose up --build -d

## infra-up: 仅启动基础设施（Redis、MySQL、PostgreSQL）
infra-up:
	docker-compose -f docker-compose.infra.yml up -d

## infra-down: 停止基础设施
infra-down:
	docker-compose -f docker-compose.infra.yml down

## docker-logs: 查看应用日志
docker-logs:
	docker-compose logs -f app

# ── 清理 ──────────────────────────────────────────────────

## clean: 清除构建产物和覆盖率文件
clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html

# ── 前端 ──────────────────────────────────────────────────

## frontend-install: 安装前端依赖
frontend-install:
	cd front-end && npm install

## frontend-dev: 启动前端开发服务器（端口 9003）
frontend-dev:
	cd front-end && npm run dev

## frontend-build: 构建前端生产版本
frontend-build:
	cd front-end && npm run build

# ── Proto（Phase 2）──────────────────────────────────────

## proto: 生成 Protobuf Go 代码
proto:
	bash scripts/gen-proto.sh

## proto-install: 安装 protoc Go 插件（首次使用需执行）
proto-install:
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# ── 微服务（Phase 2）─────────────────────────────────────

## run-user: 启动 User gRPC 微服务（端口 50001）
run-user:
	$(GO) run ./cmd/user/main.go

# ── 帮助 ──────────────────────────────────────────────────

## help: 显示所有可用命令
help:
	@echo "用法: make <target>"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

.PHONY: build run run-dev test test-cover test-verbose test-unit test-integration test-e2e \
        lint fmt swagger seed seed-purge \
        docker-up docker-down all-up infra-up infra-down docker-logs clean \
        frontend-install frontend-dev frontend-build \
        proto proto-install run-user help
