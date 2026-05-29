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

## run-auth: 启动 Auth gRPC 微服务（端口 50002，需先启动 User Service）
run-auth:
	$(GO) run ./cmd/auth/main.go

## run-permission: 启动 Permission gRPC 微服务（端口 50003）
run-permission:
	$(GO) run ./cmd/permission/main.go

## run-biz: 启动 Biz gRPC 微服务（端口 50004）
run-biz:
	$(GO) run ./cmd/biz/main.go

## run-gateway: 启动 API Gateway（端口 8000，需先启动所有微服务）
run-gateway:
	$(GO) run ./cmd/gateway/main.go

# ── Kubernetes / Helm（Phase 3）─────────────────────────

## k8s-build: 构建所有 Docker 镜像（使用 minikube Docker 环境）
k8s-build:
	bash deploy/docker/build.sh

## k8s-build-one: 构建单个服务镜像（用法: make k8s-build-one SVC=user）
k8s-build-one:
	@eval $$(minikube docker-env) && \
	docker build --build-arg SERVICE_NAME=$(SVC) \
		-f deploy/docker/Dockerfile.service \
		-t hellogo/$(SVC):$(or $(TAG),latest) .

## k8s-build-frontend: 构建前端镜像
k8s-build-frontend:
	@eval $$(minikube docker-env) && \
	docker build \
		-f deploy/docker/Dockerfile.frontend \
		-t hellogo/frontend:$(or $(TAG),latest) .

## k8s-install: 首次安装 Helm release
k8s-install:
	helm install hellogo deploy/helm/hellogo/ \
		--namespace hellogo \
		--create-namespace

## k8s-upgrade: 升级 Helm release（可选指定版本: make k8s-upgrade SVC=user TAG=v1.1.0）
k8s-upgrade:
ifdef SVC
	helm upgrade hellogo deploy/helm/hellogo/ \
		--namespace hellogo \
		--set services.$(SVC).tag=$(or $(TAG),latest)
else
	helm upgrade hellogo deploy/helm/hellogo/ \
		--namespace hellogo
endif

## k8s-deploy: 快速部署单个服务（构建 + 重启 Pod，用法: make k8s-deploy SVC=user）
k8s-deploy: k8s-build-one k8s-restart

## k8s-uninstall: 卸载 Helm release
k8s-uninstall:
	helm uninstall hellogo -n hellogo

## k8s-stop: 停止 minikube
k8s-stop:
	minikube stop

## k8s-forward: 重启后恢复端口转发（需要 sudo）
k8s-forward:
	sudo ./scripts/k8s-forward.sh

## k8s-status: 查看部署状态（Pods + Services + Helm releases）
k8s-status:
	@echo "=== Pods ==="
	@kubectl get pods -n hellogo -o wide
	@echo ""
	@echo "=== Services ==="
	@kubectl get svc -n hellogo
	@echo ""
	@echo "=== Helm Releases ==="
	@helm list -n hellogo

## k8s-logs: 查看指定服务日志（用法: make k8s-logs SVC=user）
k8s-logs:
ifeq ($(SVC),frontend)
	kubectl logs -f deploy/frontend -n hellogo --tail=100
else ifeq ($(SVC),gateway)
	kubectl logs -f deploy/gateway -n hellogo --tail=100
else
	kubectl logs -f deploy/$(SVC)-service -n hellogo --tail=100
endif

## k8s-shell: 进入指定服务容器（用法: make k8s-shell SVC=user）
k8s-shell:
ifeq ($(SVC),frontend)
	kubectl exec -it deploy/frontend -n hellogo -- sh
else ifeq ($(SVC),gateway)
	kubectl exec -it deploy/gateway -n hellogo -- sh
else
	kubectl exec -it deploy/$(SVC)-service -n hellogo -- sh
endif

## k8s-urls: 显示所有外部访问地址
k8s-urls:
	@MINIKUBE_IP=$$(minikube ip); \
	echo "前端:    http://$${MINIKUBE_IP}:30090"; \
	echo "Gateway: http://$${MINIKUBE_IP}:30080"; \
	echo "API:     http://$${MINIKUBE_IP}:30080/api/health"

## k8s-restart: 重启指定服务（用法: make k8s-restart SVC=user）
k8s-restart:
ifeq ($(SVC),frontend)
	kubectl rollout restart deploy/frontend -n hellogo
else ifeq ($(SVC),gateway)
	kubectl rollout restart deploy/gateway -n hellogo
else
	kubectl rollout restart deploy/$(SVC)-service -n hellogo
endif

## k8s-rollback: 回滚 Helm release 到上一版本
k8s-rollback:
	helm rollback hellogo -n hellogo

## k8s-seed: 手动执行种子数据（首次部署自动执行，此命令用于重新播种）
k8s-seed:
	kubectl create job hellogo-seed-manual \
		--from=job/hellogo-seed \
		-n hellogo \
		--dry-run=client -o yaml | \
		sed 's/hellogo-seed$$/hellogo-seed-manual/' | \
		kubectl apply -f -
	@echo "种子数据 Job 已创建，查看日志："
	@echo "  kubectl logs -f job/hellogo-seed-manual -n hellogo"

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
        proto proto-install run-user run-auth run-permission run-biz run-gateway \
        k8s-build k8s-build-one k8s-build-frontend k8s-install k8s-upgrade k8s-deploy k8s-uninstall \
        k8s-status k8s-logs k8s-shell k8s-urls k8s-restart k8s-rollback k8s-seed \
        help
