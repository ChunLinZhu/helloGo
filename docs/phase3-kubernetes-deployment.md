# 第三阶段开发计划：Kubernetes 部署

> 基于第二阶段 gRPC 微服务架构，从 Docker Compose 本地部署进阶到 Kubernetes 生产级部署  
> 面向 Go 初学者设计，逐步掌握容器编排、服务治理与生产化运维

---

## 目录

- [0. 学习路线图](#0-学习路线图)
- [1. Kubernetes 基础入门](#1-kubernetes-基础入门)
- [2. 架构演进：Compose → K8s](#2-架构演进compose--k8s)
- [3. 应用层改造](#3-应用层改造)
- [4. Dockerfile 生产化](#4-dockerfile-生产化)
- [5. K8s 资源清单（基础篇）](#5-k8s-资源清单基础篇)
- [6. 配置管理与密钥](#6-配置管理与密钥)
- [7. 服务发现：etcd → K8s DNS](#7-服务发现etcd--k8s-dns)
- [8. 入口与流量管理](#8-入口与流量管理)
- [9. 持久化存储](#9-持久化存储)
- [10. Helm Chart 打包](#10-helm-chart-打包)
- [11. CI/CD 流水线](#11-cicd-流水线)
- [12. 可观测性（K8s 原生方案）](#12-可观测性k8s-原生方案)
- [13. 生产加固](#13-生产加固)
- [14. 多环境策略](#14-多环境策略)
- [15. 阶段总结与时间规划](#15-阶段总结与时间规划)

---

## 0. 学习路线图

> 第二阶段用 Docker Compose 在单机上运行全部微服务。第三阶段的目标是把这套服务部署到 K8s 集群，  
> 获得自动扩缩容、滚动更新、自愈、配置管理等生产级能力。

```
┌──────────────────────────────────────────────────────────────────────┐
│                     Kubernetes 部署学习路线                           │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Milestone 1: 理解 K8s 核心概念                                      │
│  ├── Pod / Deployment / Service / ConfigMap / Secret                 │
│  ├── kubectl 基本操作                                                │
│  ├── 本地集群：minikube / kind / k3d                                 │
│  └── 动手：在 minikube 上部署一个 Nginx Pod                         │
│                                                                      │
│  Milestone 2: 应用改造                                               │
│  ├── 健康检查探针（Liveness / Readiness / Startup）                   │
│  ├── 优雅停机（PreStop hook + signal 处理）                          │
│  ├── 配置外部化（环境变量 → ConfigMap / Secret）                     │
│  ├── Dockerfile 安全加固（非 root、只读文件系统）                    │
│  └── 动手：改造 User Service 并在 K8s 上运行                        │
│                                                                      │
│  Milestone 3: 完整部署                                               │
│  ├── 编写全部微服务的 K8s 清单（Manifests）                          │
│  ├── 服务发现从 etcd 迁移到 K8s DNS                                  │
│  ├── Ingress 配置 HTTP 入口                                          │
│  ├── ConfigMap / Secret 管理配置                                     │
│  └── 动手：kubectl apply 一键部署全部服务                            │
│                                                                      │
│  Milestone 4: Helm + CI/CD                                           │
│  ├── 用 Helm Chart 模板化 K8s 清单                                   │
│  ├── GitHub Actions 自动构建镜像 + 推送                              │
│  ├── 自动部署到 K8s 集群                                             │
│  └── 动手：提交代码后自动部署到 dev 环境                             │
│                                                                      │
│  Milestone 5: 生产化                                                 │
│  ├── HPA 自动扩缩容                                                 │
│  ├── NetworkPolicy 网络隔离                                          │
│  ├── PodDisruptionBudget 保障可用性                                  │
│  ├── Prometheus + Grafana 监控                                       │
│  └── 动手：压测触发自动扩容 + 查看 Grafana 面板                     │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

### 前置条件

| 条件 | 说明 |
|------|------|
| 完成第二阶段 | gRPC 微服务 + Docker Compose 能正常运行 |
| Docker 基础 | 理解 Dockerfile、镜像、容器 |
| Linux 基础 | 熟悉命令行、网络、存储基本概念 |
| Go 开发环境 | Go 1.23+、kubectl、helm、minikube/kind |

### 推荐学习资源

| 资源 | 链接 | 说明 |
|------|------|------|
| K8s 官方文档 | https://kubernetes.io/zh-cn/docs/ | 中文文档，概念详解 |
| K8s 交互式教程 | https://kubernetes.io/docs/tutorials/ | 浏览器内实操 |
| Helm 官方文档 | https://helm.sh/docs/ | Chart 开发指南 |
| Kubernetes the Hard Way | https://github.com/kelseyhightower/kubernetes-the-hard-way | 深入理解 K8s 原理 |
| 12-Factor App | https://12factor.net/zh_cn/ | 云原生应用方法论 |

---

## 1. Kubernetes 基础入门

### 1.1 为什么需要 K8s？

> 第二阶段用 Docker Compose 可以在单机上运行全部服务，但在生产环境中会遇到很多问题：

| 问题 | Docker Compose | Kubernetes |
|------|---------------|------------|
| 单机故障 | 全部服务宕机 | 自动在其他节点重新调度 |
| 流量突增 | 手动扩缩容 | HPA 自动扩缩 |
| 滚动更新 | 需要手动编排 | 内置滚动更新策略 |
| 多环境部署 | 复制多份 compose 文件 | Kustomize / Helm values |
| 配置管理 | 环境变量硬编码 | ConfigMap + Secret |
| 服务发现 | 需要 etcd 等外部组件 | 内置 DNS |
| 健康检查 | 简单重启 | Liveness / Readiness / Startup 探针 |
| 负载均衡 | 简单端口映射 | Service + Ingress |

### 1.2 核心概念速览

```
Kubernetes 集群
├── Control Plane（控制面）
│   ├── API Server          — 所有操作的入口（REST API）
│   ├── etcd                — 集群状态存储（K8s 自己的 etcd，不是我们的服务发现）
│   ├── Scheduler           — 决定 Pod 运行在哪个 Node
│   └── Controller Manager  — 维持期望状态（如 Deployment 副本数）
│
└── Worker Nodes（工作节点）
    ├── kubelet             — 管理 Pod 生命周期
    ├── kube-proxy          — 实现 Service 负载均衡
    └── Container Runtime   — 运行容器（containerd / CRI-O）

核心资源对象：
┌─────────────┬────────────────────────────────────────────┐
│ 资源         │ 作用                                       │
├─────────────┼────────────────────────────────────────────┤
│ Pod         │ 最小调度单元，包含 1+ 个容器                │
│ Deployment  │ 管理 Pod 副本数、滚动更新、回滚             │
│ Service     │ 为一组 Pod 提供稳定的 IP + DNS 名称          │
│ ConfigMap   │ 存储非敏感配置（键值对 / 文件）             │
│ Secret      │ 存储敏感数据（密码 / Token / 证书）         │
│ Ingress     │ HTTP 路由规则，将外部流量引入集群          │
│ PVC         │ 持久化存储声明                              │
│ HPA         │ 根据指标自动调整 Pod 副本数                 │
│ Namespace   │ 逻辑隔离（类似文件夹）                      │
└─────────────┴────────────────────────────────────────────┘
```

### 1.3 本地开发环境搭建

> 在本地跑 K8s 集群，推荐以下工具（选一个即可）：

| 工具 | 平台 | 说明 |
|------|------|------|
| **minikube** | Linux / macOS / Windows | 最成熟，支持多种驱动（Docker / VirtualBox / KVM） |
| **kind** | 全平台 | 用 Docker 容器模拟 K8s 节点，启动快 |
| **k3d** | 全平台 | 在 Docker 中运行 k3s，轻量级 |
| **Docker Desktop** | macOS / Windows | 内置 K8s，一键开启 |

**使用 minikube（推荐）：**

```bash
# 安装 minikube（Ubuntu）
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube

# 安装 kubectl
sudo apt install kubectl
# 或
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install kubectl /usr/local/bin/kubectl

# 启动集群（使用 Docker 驱动）
minikube start --driver=docker --cpus=4 --memory=8192

# 验证
kubectl cluster-info
kubectl get nodes
# NAME       STATUS   ROLES           AGE   VERSION
# minikube   Ready    control-plane   30s   v1.29.0

# 开启辅助工具
minikube addons enable ingress
minikube addons enable metrics-server
minikube addons enable dashboard
```

**使用 kind（轻量替代）：**

```bash
# 安装 kind
go install sigs.k8s.io/kind@latest

# 创建多节点集群（模拟真实环境）
cat > kind-config.yaml << 'EOF'
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    kubeadmConfigPatches:
      - |
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: "ingress-ready=true"
    extraPortMappings:
      - containerPort: 80
        hostPort: 80
      - containerPort: 443
        hostPort: 443
  - role: worker
  - role: worker
EOF

kind create cluster --name hellogo --config kind-config.yaml
```

### 1.4 kubectl 常用命令

```bash
# ===== 查看资源 =====
kubectl get pods                    # 查看 Pod
kubectl get pods -o wide            # 更多信息（IP、Node）
kubectl get deployments             # 查看 Deployment
kubectl get services                # 查看 Service
kubectl get all                     # 查看全部资源
kubectl get events --sort-by=.metadata.creationTimestamp  # 查看事件

# ===== 描述与调试 =====
kubectl describe pod <pod-name>     # 查看 Pod 详情（事件、状态）
kubectl logs <pod-name>             # 查看日志
kubectl logs -f <pod-name>          # 实时跟踪日志
kubectl logs <pod-name> --previous  # 查看上一次崩溃的日志
kubectl exec -it <pod-name> -- sh   # 进入容器

# ===== 部署与管理 =====
kubectl apply -f deployment.yaml    # 创建/更新资源
kubectl delete -f deployment.yaml   # 删除资源
kubectl rollout status deployment/<name>    # 查看滚动更新状态
kubectl rollout undo deployment/<name>      # 回滚到上一版本
kubectl scale deployment/<name> --replicas=5  # 手动扩缩容

# ===== 端口转发（调试用） =====
kubectl port-forward svc/gateway 8000:8000  # 本地 8000 → 集群 gateway:8000
```

### 1.5 动手练习

| # | 练习 | 目标 |
|---|------|------|
| 1 | `minikube start` 启动本地集群 | 搭建 K8s 开发环境 |
| 2 | `kubectl run nginx --image=nginx` 并查看 Pod 状态 | 理解 Pod 生命周期 |
| 3 | 创建 Deployment + Service，通过 port-forward 访问 | 理解 Deployment 和 Service |
| 4 | 修改 Deployment 镜像版本，观察滚动更新过程 | 理解滚动更新策略 |
| 5 | 删除一个 Pod，观察自动重建 | 理解自愈机制 |

---

## 2. 架构演进：Compose → K8s

### 2.1 第二阶段架构回顾（Docker Compose）

```
┌─────────────────────────────────────────────────────┐
│                Docker Compose 单机                   │
│                                                      │
│  ┌──────────┐                                       │
│  │ Gateway  │ :8000  ← curl / 浏览器               │
│  └────┬─────┘                                       │
│       │ gRPC                                         │
│  ┌────┴─────────────────────────┐                   │
│  │  ┌──────┐ ┌──────┐ ┌──────┐ │                   │
│  │  │ User │ │ Auth │ │ Perm │ │ Biz │              │
│  │  │:50001│ │:50002│ │:50003│ │:50004│             │
│  │  └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘             │
│  └─────┼────────┼────────┼────────┼─────┘           │
│        ▼        ▼        ▼        ▼                  │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌────────┐             │
│  │MySQL │ │Redis │ │ etcd │ │ Jaeger │              │
│  └──────┘ └──────┘ └──────┘ └────────┘              │
└─────────────────────────────────────────────────────┘

问题：
✗ 单点故障（一台机器挂了全部挂）
✗ 无法水平扩展
✗ 没有自动扩缩容
✗ 更新需要手动编排
✗ 服务发现依赖 etcd（额外运维）
```

### 2.2 第三阶段目标架构（Kubernetes）

```
                          ┌─────────────┐
                          │   Client    │
                          │ (浏览器/App) │
                          └──────┬──────┘
                                 │ HTTPS
                                 ▼
                    ┌────────────────────────┐
                    │      Ingress           │  ← TLS 终止、路由规则
                    │  (Nginx / Traefik)     │
                    └────────────┬───────────┘
                                 │
                    ┌────────────┴───────────┐
                    │   Gateway Service      │  ← ClusterIP + 负载均衡
                    │   (Fiber HTTP)         │
                    │   Deployment: 2+ Pod   │
                    └───┬───────┬──────┬─────┘
                        │       │      │  gRPC (K8s DNS 服务发现)
              ┌─────────┘       │      └─────────┐
              ▼                 ▼                 ▼
     ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
     │ User Service │ │ Auth Service │ │  Biz Service │ ...
     │ Deployment   │ │ Deployment   │ │  Deployment  │
     │ 2+ Pod       │ │ 2+ Pod       │ │  2+ Pod      │
     │ HPA: 2~10    │ │ HPA: 2~5     │ │  HPA: 2~5    │
     └──────┬───────┘ └──────┬───────┘ └──────┬───────┘
            │                 │                 │
            ▼                 ▼                 ▼
     ┌──────────────────────────────────────────────────┐
     │                  K8s 内置能力                      │
     │  ┌────────┐ ┌──────────┐ ┌──────────┐            │
     │  │ DNS    │ │ConfigMap │ │ Secret   │            │
     │  │(服务发现)│ │(配置管理) │ │(密钥管理) │           │
     │  └────────┘ └──────────┘ └──────────┘            │
     └──────────────────────────────────────────────────┘
            │                 │
            ▼                 ▼
     ┌──────────────────────────────────────────────────┐
     │              外部基础设施（托管服务）               │
     │  ┌──────────────┐  ┌──────────────┐              │
     │  │ Cloud SQL    │  │ ElastiCache  │              │
     │  │ (MySQL 托管)  │  │ (Redis 托管)  │             │
     │  └──────────────┘  └──────────────┘              │
     │  ┌──────────────┐  ┌──────────────┐              │
     │  │ Prometheus   │  │ Grafana      │              │
     │  │ (监控)        │  │ (面板)        │             │
     │  └──────────────┘  └──────────────┘              │
     └──────────────────────────────────────────────────┘

优势：
✓ 高可用（Pod 故障自动重建）
✓ 自动扩缩容（HPA 根据 CPU/QPS 自动调整）
✓ 滚动更新（零停机发布）
✓ 服务发现用 K8s DNS（无需 etcd）
✓ 配置与密钥统一管理
✓ 网络隔离（NetworkPolicy）
```

### 2.3 关键变化对比

| 维度 | 第二阶段（Compose） | 第三阶段（K8s） |
|------|---------------------|-----------------|
| 编排工具 | docker-compose | kubectl + Helm |
| 服务发现 | etcd + TTL 注册 | K8s 内置 DNS |
| 负载均衡 | Docker 端口映射 | Service (ClusterIP) |
| 配置管理 | 环境变量 | ConfigMap + Secret |
| 健康检查 | 无 / 简单 restart | Liveness + Readiness + Startup 探针 |
| 存储 | Docker named volume | PVC + StorageClass |
| 扩缩容 | 手动 | HPA 自动扩缩 |
| 网络 | Docker bridge 网络 | CNI + NetworkPolicy |
| TLS | 无 / 手动证书 | cert-manager 自动签发 |
| 日志 | docker compose logs | 容器 stdout → 日志收集 |
| 监控 | Prometheus（手动配置） | Prometheus Operator + Grafana |
| CI/CD | 无 | GitHub Actions → 镜像构建 → 自动部署 |
| 多环境 | 多份 compose 文件 | Kustomize / Helm values |

### 2.4 项目目录结构（新增部分）

```
helloGo/
├── ...（Phase 1 & 2 的代码保持不变）
│
├── deploy/                              # 第三阶段新增
│   ├── k8s/                             # K8s 原生清单
│   │   ├── base/                        # Kustomize base（各环境共享）
│   │   │   ├── namespace.yaml
│   │   │   ├── configmap.yaml
│   │   │   ├── secret.yaml
│   │   │   ├── gateway/
│   │   │   │   ├── deployment.yaml
│   │   │   │   ├── service.yaml
│   │   │   │   └── kustomization.yaml
│   │   │   ├── user-service/
│   │   │   │   ├── deployment.yaml
│   │   │   │   ├── service.yaml
│   │   │   │   └── kustomization.yaml
│   │   │   ├── auth-service/
│   │   │   ├── permission-service/
│   │   │   ├── biz-service/
│   │   │   └── kustomization.yaml       # base 入口
│   │   │
│   │   └── overlays/                    # Kustomize overlays（各环境差异）
│   │       ├── dev/
│   │       │   ├── kustomization.yaml
│   │       │   └── patches/
│   │       │       └── replica-count.yaml
│   │       ├── staging/
│   │       │   ├── kustomization.yaml
│   │       │   └── patches/
│   │       └── production/
│   │           ├── kustomization.yaml
│   │           └── patches/
│   │               ├── hpa.yaml
│   │               └── resource-limits.yaml
│   │
│   ├── helm/                            # Helm Chart（推荐方式）
│   │   └── hellogo/
│   │       ├── Chart.yaml
│   │       ├── values.yaml              # 默认值（dev 环境）
│   │       ├── values-staging.yaml
│   │       ├── values-production.yaml
│   │       └── templates/
│   │           ├── _helpers.tpl
│   │           ├── namespace.yaml
│   │           ├── configmap.yaml
│   │           ├── secret.yaml
│   │           ├── gateway/
│   │           │   ├── deployment.yaml
│   │           │   ├── service.yaml
│   │           │   ├── hpa.yaml
│   │           │   └── ingress.yaml
│   │           ├── user-service/
│   │           │   ├── deployment.yaml
│   │           │   ├── service.yaml
│   │           │   └── hpa.yaml
│   │           ├── auth-service/
│   │           ├── permission-service/
│   │           ├── biz-service/
│   │           ├── NOTES.txt
│   │           └── tests/
│   │               └── test-connection.yaml
│   │
│   └── docker/                          # Docker 相关文件
│       ├── Dockerfile                   # 多阶段构建（复用 Phase 2 并改进）
│       └── .dockerignore
│
├── .github/                             # CI/CD
│   └── workflows/
│       ├── ci.yml                       # 测试 + 构建
│       ├── cd-dev.yml                   # 自动部署到 dev
│       └── cd-production.yml            # 手动审批部署到 prod
│
└── scripts/
    ├── k8s-deploy.sh                    # K8s 部署脚本
    ├── k8s-rollback.sh                  # 回滚脚本
    └── k8s-health-check.sh             # 健康检查脚本
```

---

## 3. 应用层改造

> 在把应用部署到 K8s 之前，需要先改造应用本身，让它"K8s-friendly"。

### 3.1 健康检查探针

> K8s 通过探针（Probe）判断 Pod 的状态，决定是否需要重启或将流量路由到该 Pod。  
> 三种探针缺一不可：

```
┌───────────────────────────────────────────────────────────┐
│                    K8s 探针机制                             │
├─────────────────┬─────────────────────────────────────────┤
│ StartupProbe    │ 应用启动慢时使用（如数据库迁移）         │
│                 │ 成功之前，Liveness/Readiness 不会启动    │
│                 │ 防止启动慢的应用被反复杀死               │
├─────────────────┼─────────────────────────────────────────┤
│ LivenessProbe   │ 应用是否"活着"                          │
│                 │ 失败 → K8s 重启 Pod（kill + restart）   │
│                 │ 检测死锁、无限循环等不可恢复错误         │
├─────────────────┼─────────────────────────────────────────┤
│ ReadinessProbe  │ 应用是否"准备好接收流量"                │
│                 │ 失败 → 从 Service Endpoints 中移除      │
│                 │ 检测 DB 连接、缓存预热等暂时性问题       │
└─────────────────┴─────────────────────────────────────────┘
```

**改造 gRPC 服务的健康检查端点：**

```go
// internal/shared/health/health.go
// 所有微服务共享的健康检查模块

package health

import (
    "sync/atomic"

    "github.com/gofiber/fiber/v2"
    "gorm.io/gorm"
)

// Checker 健康检查器
type Checker struct {
    db      *gorm.DB
    ready   atomic.Bool
    started atomic.Bool
}

func NewChecker(db *gorm.DB) *Checker {
    return &Checker{db: db}
}

// SetReady 标记服务就绪（在启动完成后调用）
func (c *Checker) SetReady(ready bool) {
    c.ready.Store(ready)
    c.started.Store(true)
}

// RegisterRoutes 注册健康检查路由（在 Fiber HTTP 端口上暴露）
func (c *Checker) RegisterRoutes(app *fiber.App) {
    health := app.Group("/healthz")

    // 存活探针：应用进程是否存活
    // 不依赖外部服务，只要进程在就返回 OK
    // K8s 失败时会重启 Pod
    health.Get("/liveness", func(ctx *fiber.Ctx) error {
        return ctx.JSON(fiber.Map{
            "status": "ok",
        })
    })

    // 就绪探针：是否可以接收流量
    // 检查 DB 连接 + Redis 连接
    // K8s 失败时会从 Service Endpoints 移除该 Pod
    health.Get("/readiness", func(ctx *fiber.Ctx) error {
        if !c.ready.Load() {
            return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
                "status": "not_ready",
                "reason": "service is still initializing",
            })
        }

        // 检查数据库连接
        sqlDB, err := c.db.DB()
        if err != nil {
            return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
                "status":  "not_ready",
                "reason":  "database connection failed",
                "details": err.Error(),
            })
        }
        if err := sqlDB.Ping(); err != nil {
            return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
                "status":  "not_ready",
                "reason":  "database ping failed",
                "details": err.Error(),
            })
        }

        return ctx.JSON(fiber.Map{
            "status": "ready",
        })
    })

    // 启动探针：应用是否完成启动
    // 用于启动慢的服务（如需要加载大量数据）
    health.Get("/startup", func(ctx *fiber.Ctx) error {
        if !c.started.Load() {
            return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
                "status": "starting",
            })
        }
        return ctx.JSON(fiber.Map{
            "status": "started",
        })
    })
}
```

**在每个微服务的 main.go 中使用：**

```go
// cmd/user/main.go
func main() {
    // ... 初始化 DB、Redis、gRPC Server ...

    // 创建健康检查器
    checker := health.NewChecker(db)

    // Fiber HTTP Server（健康检查 + 指标暴露）
    httpApp := fiber.New()
    checker.RegisterRoutes(httpApp)

    // Prometheus 指标端点
    httpApp.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

    // 启动 HTTP Server（健康检查端口，默认 8080）
    go httpApp.Listen(":8080")

    // 启动 gRPC Server
    go func() {
        lis, _ := net.Listen("tcp", ":50001")
        grpcServer.Serve(lis)
    }()

    // 标记服务就绪
    checker.SetReady(true)
    logger.Info("User Service 已就绪，开始接收流量")

    // 等待退出信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // 优雅停机（详见 3.2）
    checker.SetReady(false)
    gracefulShutdown(grpcServer, httpApp)
}
```

### 3.2 优雅停机

> K8s 在终止 Pod 时会发送 SIGTERM 信号，默认等待 30 秒（`terminationGracePeriodSeconds`）。  
> 应用需要在这段时间内完成：停止接收新请求 → 处理完已有请求 → 关闭连接。

```go
// internal/shared/shutdown/shutdown.go

package shutdown

import (
    "context"
    "time"

    "github.com/gofiber/fiber/v2"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/health"
)

// GracefulShutdown 优雅关闭所有服务
func GracefulShutdown(
    logger *zap.Logger,
    grpcServer *grpc.Server,
    httpApp *fiber.App,
    healthServer *health.Server,
    timeout time.Duration,
) {
    logger.Info("收到终止信号，开始优雅停机...")

    // 1. 立即标记不健康（从 K8s Service Endpoints 中移除）
    //    这样新的流量不会再路由到这个 Pod
    healthServer.SetServingStatus("", healthpb2.HealthCheckResponse_NOT_SERVING)

    // 2. 等待一小段时间，让 K8s 更新 Endpoints
    //    因为 Endpoints 更新有延迟，直接关闭可能会丢失少量请求
    logger.Info("等待 K8s Endpoints 更新...")
    time.Sleep(5 * time.Second)

    // 3. 停止 gRPC Server（GracefulStop 会等待已有请求处理完）
    logger.Info("停止 gRPC Server（等待已有请求完成）...")
    done := make(chan struct{})
    go func() {
        grpcServer.GracefulStop()
        close(done)
    }()

    // 4. 停止 HTTP Server
    shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    if err := httpApp.ShutdownWithContext(shutdownCtx); err != nil {
        logger.Error("HTTP Server 关闭出错", zap.Error(err))
    }

    // 5. 等待 gRPC 关闭（最多等 timeout）
    select {
    case <-done:
        logger.Info("gRPC Server 已优雅关闭")
    case <-shutdownCtx.Done():
        logger.Warn("gRPC Server 关闭超时，强制关闭")
        grpcServer.Stop()
    }

    logger.Info("所有服务已关闭")
}
```

**在 Deployment 中配合 PreStop Hook：**

```yaml
# K8s 终止 Pod 的流程：
# 1. Pod 标记为 Terminating
# 2. 同时执行：
#    a. PreStop Hook（如果配置了的话）
#    b. kube-proxy 更新 iptables 规则（不再转发新流量）
# 3. PreStop 完成后，发送 SIGTERM
# 4. 等待 terminationGracePeriodSeconds（默认 30s）
# 5. 超时则发送 SIGKILL 强杀

# PreStop Hook 的作用：
# 在 SIGTERM 之前先等几秒，让 kube-proxy 有时间更新路由规则
# 否则 SIGTERM 和路由更新是并行的，可能会丢失少量请求
```

### 3.3 配置外部化

> 遵循 12-Factor App 原则：配置从环境变量注入，不硬编码在代码中。  
> 在 K8s 中，环境变量来自 ConfigMap 和 Secret。

**改造配置加载逻辑（兼容本地开发 + K8s）：**

```go
// internal/shared/config/config.go
// 改造后的配置加载，复用 Phase 1 的嵌套结构体风格
// 优先级：环境变量 > .env 文件 > 默认值

package config

import (
    "fmt"

    "github.com/spf13/viper"
)

type Config struct {
    App      AppConfig      `mapstructure:",squash"`
    Database DatabaseConfig `mapstructure:",squash"`
    Redis    RedisConfig    `mapstructure:",squash"`
    JWT      JWTConfig      `mapstructure:",squash"`
    Security SecurityConfig `mapstructure:",squash"`
    Login    LoginConfig    `mapstructure:",squash"`
    Upload   UploadConfig   `mapstructure:",squash"`

    // 服务发现（Phase 2 新增）
    // K8s 环境下使用 DNS 地址，不需要 etcd
    ServiceDiscovery string `mapstructure:"SERVICE_DISCOVERY"` // "etcd" 或 "dns"
    EtcdEndpoints    string `mapstructure:"ETCD_ENDPOINTS"`

    // 下游服务地址（K8s DNS 格式：service-name.namespace.svc.cluster.local）
    UserServiceAddr       string `mapstructure:"USER_SERVICE_ADDR"`
    AuthServiceAddr       string `mapstructure:"AUTH_SERVICE_ADDR"`
    PermissionServiceAddr string `mapstructure:"PERMISSION_SERVICE_ADDR"`
    BizServiceAddr        string `mapstructure:"BIZ_SERVICE_ADDR"`

    // 可观测性
    OtelEnabled    bool   `mapstructure:"OTEL_ENABLED"`
    OtelEndpoint   string `mapstructure:"OTEL_ENDPOINT"`
    MetricsEnabled bool   `mapstructure:"ENABLE_METRICS"`
}

type AppConfig struct {
    Env  string `mapstructure:"APP_ENV"`
    Port int    `mapstructure:"PORT"`       // Gateway HTTP 端口
}

type DatabaseConfig struct {
    Type     string `mapstructure:"DB_TYPE"`
    MySQL    MySQLConfig
    Postgres PostgresConfig
    SQLite   SQLiteConfig
}

type MySQLConfig struct {
    Host     string `mapstructure:"DB_HOST"`
    Port     int    `mapstructure:"DB_PORT"`
    User     string `mapstructure:"DB_USER"`
    Password string `mapstructure:"DB_PASS"`
    Name     string `mapstructure:"DB_NAME"`
}

type PostgresConfig struct {
    Host     string `mapstructure:"PG_HOST"`
    Port     int    `mapstructure:"PG_PORT"`
    User     string `mapstructure:"PG_USER"`
    Password string `mapstructure:"PG_PASS"`
    Name     string `mapstructure:"PG_DB"`
}

type SQLiteConfig struct {
    Path string `mapstructure:"SQLITE_PATH"`
}

type RedisConfig struct {
    Host     string `mapstructure:"REDIS_HOST"`
    Port     int    `mapstructure:"REDIS_PORT"`
    Password string `mapstructure:"REDIS_PASS"`
}

type JWTConfig struct {
    Secret         string `mapstructure:"JWT_SECRET"`
    Expires        string `mapstructure:"JWT_EXPIRES"`         // "1d", "2h" 等
    RefreshExpires string `mapstructure:"JWT_REFRESH_EXPIRES"` // "7d" 等
}

type SecurityConfig struct {
    CSRFEnabled bool   `mapstructure:"CSRF_ENABLED"`
    CSRFMode    string `mapstructure:"CSRF_MODE"`
    CSRFSecret  string `mapstructure:"CSRF_SECRET"`
    CORSOrigins string `mapstructure:"CORS_ORIGINS"`
}

type LoginConfig struct {
    MaxFails int `mapstructure:"LOGIN_MAX_FAILS"`
    LockTTL  int `mapstructure:"LOGIN_LOCK_TTL"`
}

type UploadConfig struct {
    Dest            string `mapstructure:"UPLOAD_DEST"`
    MaxSize         int    `mapstructure:"UPLOAD_MAX_SIZE"`
    AllowedTypes    string `mapstructure:"UPLOAD_ALLOWED_TYPES"`
    CleanInterval   int    `mapstructure:"UPLOAD_CLEAN_INTERVAL_SEC"`
    TTLDays         int    `mapstructure:"UPLOAD_TTL_DAYS"`
}

func Load() (*Config, error) {
    v := viper.New()
    v.SetConfigName(".env")
    v.SetConfigType("env")
    v.AddConfigPath(".")
    v.AddConfigPath("./configs")

    // 自动绑定环境变量（K8s ConfigMap/Secret 注入的环境变量）
    v.AutomaticEnv()

    // 设置默认值
    v.SetDefault("APP_ENV", "development")
    v.SetDefault("PORT", 8000)
    v.SetDefault("DB_TYPE", "sqlite")
    v.SetDefault("SERVICE_DISCOVERY", "dns")   // K8s 默认用 DNS
    v.SetDefault("OTEL_ENABLED", false)
    v.SetDefault("ENABLE_METRICS", true)

    // 尝试读取 .env 文件（本地开发用，K8s 中不存在该文件）
    _ = v.ReadInConfig()

    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
```

### 3.4 服务发现的配置切换

```go
// internal/shared/discovery/discovery.go
// 支持两种服务发现模式：etcd（本地开发）和 dns（K8s 环境）

package discovery

import (
    "context"
    "fmt"
)

// Discoverer 服务发现接口
type Discoverer interface {
    // Register 注册当前服务实例
    Register(ctx context.Context, serviceName, addr string) error
    // Deregister 注销当前服务实例
    Deregister(ctx context.Context, serviceName, addr string) error
    // Discover 获取服务实例地址列表
    Discover(ctx context.Context, serviceName string) ([]string, error)
    // GetAddr 获取单个服务地址（用于 gRPC 连接）
    GetAddr(ctx context.Context, serviceName string) (string, error)
}

// NewDiscoverer 根据配置创建服务发现实现
func NewDiscoverer(mode string, cfg *config.Config) (Discoverer, error) {
    switch mode {
    case "etcd":
        // 第二阶段的方式：使用 etcd + TTL 注册
        return NewEtcdDiscoverer(cfg.EtcdEndpoints)
    case "dns":
        // 第三阶段的方式：使用 K8s DNS
        return NewDNSDiscoverer(), nil
    default:
        return nil, fmt.Errorf("unknown discovery mode: %s", mode)
    }
}
```

```go
// internal/shared/discovery/dns.go
// K8s DNS 服务发现

package discovery

import (
    "context"
    "fmt"
    "os"
)

// DNSDiscoverer 使用 K8s DNS 进行服务发现
// 无需注册/注销，K8s Service 自动管理 DNS 记录
type DNSDiscoverer struct {
    namespace string
}

func NewDNSDiscoverer() *DNSDiscoverer {
    // 从 K8s 自动注入的文件中读取 namespace
    ns := "default"
    if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
        ns = string(data)
    }
    return &DNSDiscoverer{namespace: ns}
}

// Register 在 DNS 模式下是空操作（K8s 自动管理）
func (d *DNSDiscoverer) Register(_ context.Context, _, _ string) error {
    return nil
}

// Deregister 在 DNS 模式下是空操作
func (d *DNSDiscoverer) Deregister(_ context.Context, _, _ string) error {
    return nil
}

// Discover 返回 K8s Service 的 DNS 地址
// K8s 的 Service 会自动负载均衡到后端的 Pod
func (d *DNSDiscoverer) Discover(_ context.Context, serviceName string) ([]string, error) {
    addr := fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, d.namespace)
    return []string{addr}, nil
}

// GetAddr 返回可直接用于 gRPC 连接的地址
func (d *DNSDiscoverer) GetAddr(_ context.Context, serviceName string) (string, error) {
    return fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, d.namespace), nil
}
```

### 3.5 动手练习

| # | 练习 | 目标 |
|---|------|------|
| 1 | 给 User Service 添加三个健康检查端点 | 理解探针的作用 |
| 2 | 在 main.go 中实现优雅停机逻辑 | 理解 SIGTERM 处理流程 |
| 3 | 改造配置加载，支持从环境变量注入所有配置 | 理解 12-Factor 配置原则 |
| 4 | 实现 DNS 服务发现，替换 etcd | 理解 K8s DNS 服务发现 |

---

## 4. Dockerfile 生产化

> 第二阶段的 Dockerfile 可以用于 Compose 部署，但在 K8s 生产环境需要做安全加固。

### 4.1 改进后的 Dockerfile

```dockerfile
# deploy/docker/Dockerfile
# 多阶段构建 + 安全加固
# 用法: docker build --build-arg SERVICE=user -f deploy/docker/Dockerfile -t hellogo-user .

ARG SERVICE
ARG GO_VERSION=1.23
ARG ALPINE_VERSION=3.20

# ===== Stage 1: 构建 =====
FROM golang:${GO_VERSION}-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

# 先复制依赖文件，利用 Docker 缓存层
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -X main.version=$(git describe --tags --always)" \
    -o /server ./cmd/${SERVICE}

# ===== Stage 2: 运行 =====
FROM alpine:${ALPINE_VERSION}

# 安全加固：安装必要包
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    # curl 用于健康检查探针（K8s 也可以使用 exec 探针代替）
    curl \
    # 创建非 root 用户
    && addgroup -S appgroup \
    && adduser -S appuser -G appgroup -h /app

WORKDIR /app

# 复制二进制和配置
COPY --from=builder /server .
COPY --from=builder /build/configs ./configs

# 安全加固：设置文件权限
RUN chown -R appuser:appgroup /app

# 安全加固：切换到非 root 用户
USER appuser

# 健康检查端口（HTTP 探针端口）
EXPOSE 8080
# gRPC 端口（由 build arg 决定，这里只声明）
EXPOSE 50001

# 安全加固：声明健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/healthz/liveness || exit 1

# 启动
ENTRYPOINT ["./server"]
```

### 4.2 Dockerfile 改进点对照

| 维度 | Phase 2 Dockerfile | Phase 3 Dockerfile | 原因 |
|------|--------------------|--------------------|------|
| 基础镜像 | alpine:3.19 | alpine:3.20（可配置） | 使用更新的安全补丁 |
| 用户 | root（默认） | appuser（非 root） | K8s PodSecurity 要求 |
| 构建标记 | 无 | `-ldflags="-s -w"` | 减小二进制体积 |
| 版本注入 | 无 | `-X main.version=` | 运行时可查版本 |
| 健康检查 | 无 | HEALTHCHECK 指令 | Docker 级别的健康检查 |
| 文件权限 | root 所有 | appuser 所有 | 安全最小权限原则 |
| 时区 | 有 tzdata | 有 tzdata | 保持日志时间正确 |
| 依赖缓存 | 全量 COPY | 先 go.mod 再 COPY 源码 | 利用 Docker layer cache |

### 4.3 多架构构建（可选）

> 如果 K8s 节点有不同架构（如 ARM + AMD64），需要构建多架构镜像。

```bash
# 使用 docker buildx 构建多架构镜像
docker buildx create --name hellogo-builder --use

# 构建并推送多架构镜像
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --build-arg SERVICE=user \
    -t registry.example.com/hellogo-user:latest \
    -t registry.example.com/hellogo-user:v1.0.0 \
    --push \
    -f deploy/docker/Dockerfile .
```

### 4.4 .dockerignore

```
# deploy/docker/.dockerignore
.git
.github
.vscode
.idea
*.md
docs/
front-end/
scripts/
deploy/
build/
coverage/
*.test
*.out
.env.*
!.env.example
docker-compose*.yml
.air.toml
Makefile
```

---

## 5. K8s 资源清单（基础篇）

> 以下清单是部署 helloGo 微服务到 K8s 所需的核心资源。  
> 先手写原生 YAML 理解概念，再用 Helm 模板化（§10）。

### 5.1 Namespace（命名空间）

```yaml
# deploy/k8s/base/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: hellogo
  labels:
    app.kubernetes.io/name: hellogo
    app.kubernetes.io/part-of: hellogo
```

```bash
kubectl apply -f deploy/k8s/base/namespace.yaml
# 后续所有资源都部署在 hellogo namespace 中
kubectl config set-context --current --namespace=hellogo
```

### 5.2 Deployment（部署）

> Deployment 管理 Pod 的副本数、更新策略和回滚。每个微服务一个 Deployment。

**User Service Deployment 示例：**

```yaml
# deploy/k8s/base/user-service/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
  namespace: hellogo
  labels:
    app.kubernetes.io/name: user-service
    app.kubernetes.io/part-of: hellogo
    app.kubernetes.io/component: backend
spec:
  replicas: 2
  # 滚动更新策略
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1          # 更新时最多多出 1 个 Pod
      maxUnavailable: 0    # 更新时不允许有 Pod 不可用（零停机）
  selector:
    matchLabels:
      app.kubernetes.io/name: user-service
  template:
    metadata:
      labels:
        app.kubernetes.io/name: user-service
        app.kubernetes.io/part-of: hellogo
        app.kubernetes.io/component: backend
        # Prometheus 自动发现指标端点
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      # 优雅停机等待时间
      terminationGracePeriodSeconds: 45

      # 安全上下文（Pod 级别）
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        runAsGroup: 1000
        fsGroup: 1000

      # 初始化容器（可选：等待数据库就绪）
      initContainers:
        - name: wait-for-db
          image: busybox:1.36
          command:
            - sh
            - -c
            - |
              until nc -z $DB_HOST $DB_PORT; do
                echo "等待数据库就绪..."
                sleep 2
              done
              echo "数据库已就绪"
          env:
            - name: DB_HOST
              valueFrom:
                configMapKeyRef:
                  name: hellogo-config
                  key: DB_HOST
            - name: DB_PORT
              valueFrom:
                configMapKeyRef:
                  name: hellogo-config
                  key: DB_PORT

      containers:
        - name: user-service
          image: registry.example.com/hellogo-user:latest
          imagePullPolicy: Always

          # 环境变量（从 ConfigMap 和 Secret 注入）
          envFrom:
            - configMapRef:
                name: hellogo-config
            - secretRef:
                name: hellogo-secrets

          # 资源限制（必须设置，HPA 依赖此值）
          resources:
            requests:
              cpu: 100m        # 最小请求 0.1 核
              memory: 128Mi    # 最小请求 128MB
            limits:
              cpu: 500m        # 最大使用 0.5 核
              memory: 256Mi    # 最大使用 256MB

          # 端口声明
          ports:
            - name: grpc
              containerPort: 50001
              protocol: TCP
            - name: http
              containerPort: 8080
              protocol: TCP

          # 启动探针：启动慢时使用，成功前不执行其他探针
          startupProbe:
            httpGet:
              path: /healthz/startup
              port: http
            initialDelaySeconds: 5
            periodSeconds: 3
            failureThreshold: 10   # 最多等 30s（3s × 10 次）

          # 存活探针：检测进程是否存活
          livenessProbe:
            httpGet:
              path: /healthz/liveness
              port: http
            initialDelaySeconds: 0   # startupProbe 成功后立即开始
            periodSeconds: 10
            timeoutSeconds: 3
            failureThreshold: 3      # 连续 3 次失败则重启 Pod

          # 就绪探针：检测是否可以接收流量
          readinessProbe:
            httpGet:
              path: /healthz/readiness
              port: http
            initialDelaySeconds: 0
            periodSeconds: 5
            timeoutSeconds: 3
            failureThreshold: 3      # 连续 3 次失败则从 Endpoints 移除

          # 安全上下文（容器级别）
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop: ["ALL"]

          # 生命周期钩子
          lifecycle:
            preStop:
              exec:
                # 等待 K8s 更新 Endpoints 路由规则
                # 避免 SIGTERM 和路由更新并行导致丢请求
                command: ["sh", "-c", "sleep 5"]

          # 只读文件系统的可写挂载
          volumeMounts:
            - name: tmp
              mountPath: /tmp

      volumes:
        - name: tmp
          emptyDir: {}
```

### 5.3 Service（服务）

> Service 为一组 Pod 提供稳定的 IP 地址和 DNS 名称。  
> K8s 的 Service 自带负载均衡（kube-proxy 实现），替代了 Phase 2 中 gRPC 客户端的 round_robin。

**gRPC 服务使用 Headless Service：**

```yaml
# deploy/k8s/base/user-service/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: user-service
  namespace: hellogo
  labels:
    app.kubernetes.io/name: user-service
    app.kubernetes.io/part-of: hellogo
spec:
  # Headless Service（clusterIP: None）
  # gRPC 需要直连到每个 Pod IP，不能用 ClusterIP 的 L4 负载均衡
  # Headless Service 让 DNS 返回所有 Pod IP（A 记录列表）
  # gRPC 客户端可以自己实现负载均衡（round_robin）
  clusterIP: None
  selector:
    app.kubernetes.io/name: user-service
  ports:
    - name: grpc
      port: 50001
      targetPort: grpc
      protocol: TCP
    - name: http
      port: 8080
      targetPort: http
      protocol: TCP
```

> **为什么 gRPC 用 Headless Service？**  
> 普通 ClusterIP Service 通过 kube-proxy 做 L4 负载均衡（iptables/IPVS），  
> 但 gRPC 使用 HTTP/2 长连接，所有请求会复用同一条 TCP 连接，导致负载不均匀。  
> Headless Service 让 gRPC 客户端直接拿到所有 Pod IP，自己做 round_robin 负载均衡。

**Gateway 使用 ClusterIP Service（HTTP 短连接，普通负载均衡即可）：**

```yaml
# deploy/k8s/base/gateway/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: gateway
  namespace: hellogo
  labels:
    app.kubernetes.io/name: gateway
    app.kubernetes.io/part-of: hellogo
spec:
  type: ClusterIP
  selector:
    app.kubernetes.io/name: gateway
  ports:
    - name: http
      port: 8000
      targetPort: http
      protocol: TCP
```

### 5.4 全部服务 Deployment + Service 一览

| 服务 | Deployment | Service 类型 | gRPC 端口 | HTTP 端口 |
|------|-----------|-------------|-----------|-----------|
| Gateway | gateway | ClusterIP | — | 8000 |
| User Service | user-service | Headless | 50001 | 8080 |
| Auth Service | auth-service | Headless | 50002 | 8080 |
| Permission Service | permission-service | Headless | 50003 | 8080 |
| Biz Service | biz-service | Headless | 50004 | 8080 |

**其他服务的 Deployment 与 User Service 结构相同，仅修改：**
- `name`、`labels`、`image`
- gRPC 端口号
- 环境变量（下游服务地址）
- 资源限制（按服务特点调整）

---

## 6. 配置管理与密钥

### 6.1 ConfigMap（非敏感配置）

```yaml
# deploy/k8s/base/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: hellogo-config
  namespace: hellogo
  labels:
    app.kubernetes.io/part-of: hellogo
data:
  # ===== 通用 =====
  APP_ENV: "production"
  PORT: "8000"
  SERVICE_DISCOVERY: "dns"

  # ===== 数据库 =====
  DB_TYPE: "mysql"
  DB_HOST: "mysql.hellogo.svc.cluster.local"
  DB_PORT: "3306"
  DB_NAME: "hellogo"

  # ===== Redis =====
  REDIS_HOST: "redis.hellogo.svc.cluster.local"
  REDIS_PORT: "6379"

  # ===== JWT =====
  JWT_EXPIRES: "1d"
  JWT_REFRESH_EXPIRES: "7d"

  # ===== 安全 =====
  CSRF_ENABLED: "true"
  CSRF_MODE: "header"
  CORS_ORIGINS: "https://app.hellogo.com"

  # ===== 限流 =====
  THROTTLE_TTL: "60"
  THROTTLE_LIMIT: "100"

  # ===== 登录安全 =====
  LOGIN_MAX_FAILS: "5"
  LOGIN_LOCK_TTL: "600"

  # ===== 上传 =====
  UPLOAD_DEST: "/app/upload"
  UPLOAD_MAX_SIZE: "10485760"
  UPLOAD_ALLOWED_TYPES: "image/jpeg,image/png,application/pdf"
  UPLOAD_CLEAN_INTERVAL_SEC: "3600"
  UPLOAD_TTL_DAYS: "30"

  # ===== 服务地址（K8s DNS 格式） =====
  USER_SERVICE_ADDR: "user-service.hellogo.svc.cluster.local:50001"
  AUTH_SERVICE_ADDR: "auth-service.hellogo.svc.cluster.local:50002"
  PERMISSION_SERVICE_ADDR: "permission-service.hellogo.svc.cluster.local:50003"
  BIZ_SERVICE_ADDR: "biz-service.hellogo.svc.cluster.local:50004"

  # ===== 可观测性 =====
  OTEL_ENABLED: "true"
  OTEL_ENDPOINT: "jaeger-collector.hellogo.svc.cluster.local:4318"
  ENABLE_METRICS: "true"
```

### 6.2 Secret（敏感数据）

```yaml
# deploy/k8s/base/secret.yaml
# 注意：生产环境中不要将 Secret 明文提交到 Git！
# 使用 Sealed Secrets、External Secrets Operator 或 Vault 管理。
apiVersion: v1
kind: Secret
metadata:
  name: hellogo-secrets
  namespace: hellogo
  labels:
    app.kubernetes.io/part-of: hellogo
type: Opaque
# 值是 base64 编码的（不是加密！仅避免明文暴露）
# echo -n "your-value" | base64
data:
  DB_USER: "aGVsbG9nbw=="           # hellogo
  DB_PASS: "eW91ci1kYi1wYXNz"       # your-db-pass
  REDIS_PASS: ""                    # 空密码
  JWT_SECRET: "eW91ci1qd3Qtc2VjcmV0LWtleS1jaGFuZ2UtbWU="  # your-jwt-secret-key-change-me
  CSRF_SECRET: "Y2hhbmdlLW1l"       # change-me
```

**生产环境推荐使用 Sealed Secrets：**

```bash
# 安装 Sealed Secrets Controller
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.26.0/controller.yaml

# 安装 kubeseal CLI
go install github.com/bitnami-labs/sealed-secrets/cmd/kubeseal@latest

# 加密 Secret（加密后的 SealedSecret 可以安全提交到 Git）
cat secret.yaml | kubeseal \
    --controller-namespace kube-system \
    --controller-name sealed-secrets-controller \
    --format yaml \
    > sealed-secret.yaml

# sealed-secret.yaml 可以安全提交到 Git
# K8s 集群中的 Sealed Secrets Controller 会自动解密为普通 Secret
```

### 6.3 环境变量注入方式

```yaml
# 在 Deployment 中注入配置（推荐方式）
envFrom:
  # 全部 ConfigMap 键值对注入为环境变量
  - configMapRef:
      name: hellogo-config
  # 全部 Secret 键值对注入为环境变量
  - secretRef:
      name: hellogo-secrets

# 也可以单独注入某个键
env:
  - name: DB_PASS
    valueFrom:
      secretKeyRef:
        name: hellogo-secrets
        key: DB_PASS
  - name: POD_NAME
    valueFrom:
      fieldRef:
        fieldPath: metadata.name       # 注入 Pod 名称（日志中使用）
  - name: POD_IP
    valueFrom:
      fieldRef:
        fieldPath: status.podIP        # 注入 Pod IP
```

---

## 7. 服务发现：etcd → K8s DNS

> 第二阶段使用 etcd + TTL 做服务注册与发现。在 K8s 环境中，可以完全去掉 etcd，  
> 利用 K8s 内置的 DNS 服务（CoreDNS）实现服务发现。

### 7.1 两种模式对比

```
第二阶段：etcd 服务发现
┌──────────┐    注册 /services/user/10.0.0.1:50001    ┌──────┐
│ User Svc │ ────────────────────────────────────────→ │ etcd │
└──────────┘                                           └──┬───┘
                                                          │
┌──────────┐    查询 /services/user/*                     │
│ Gateway  │ ────────────────────────────────────────────→│
└──────────┘    返回 [10.0.0.1:50001, 10.0.0.2:50001]    │
                                                          │
问题：etcd 本身需要运维（高可用、备份、监控）


第三阶段：K8s DNS 服务发现
┌──────────┐    K8s 自动注册                              ┌──────────┐
│ User Svc │ ────────────────────────────────────────────→│ CoreDNS  │
│ Pod 1    │                                              │          │
│ Pod 2    │                                              │          │
│ Pod 3    │                                              └────┬─────┘
└──────────┘                                                   │
                                                               │
┌──────────┐    DNS 查询 user-service.hellogo.svc.cluster.local│
│ Gateway  │ ─────────────────────────────────────────────────→│
└──────────┘    返回所有 Pod IP（A 记录列表）                   │

优势：无需额外组件，K8s 自动管理 DNS 记录
```

### 7.2 K8s DNS 解析规则

```
# K8s DNS 解析格式：
<service-name>.<namespace>.svc.cluster.local

# 示例（hellogo 命名空间内）：
user-service.hellogo.svc.cluster.local       → User Service 所有 Pod IP
auth-service.hellogo.svc.cluster.local        → Auth Service 所有 Pod IP
gateway.hellogo.svc.cluster.local             → Gateway 所有 Pod IP

# 同命名空间内可以简写：
user-service                                  → 同 user-service.hellogo.svc.cluster.local
user-service.hellogo                          → 同上

# Headless Service（clusterIP: None）返回所有 Pod IP：
# dig user-service.hellogo.svc.cluster.local
# ;; ANSWER SECTION:
# user-service.hellogo.svc.cluster.local. 5 IN A 10.244.0.15
# user-service.hellogo.svc.cluster.local. 5 IN A 10.244.0.16
# user-service.hellogo.svc.cluster.local. 5 IN A 10.244.1.12

# Pod 级别的 DNS（StatefulSet 才有）：
<pod-name>.<headless-service>.<namespace>.svc.cluster.local
```

### 7.3 Gateway 连接下游服务

```go
// internal/gateway/clients.go
// Gateway 连接下游 gRPC 服务的代码改造

package gateway

import (
    "context"

    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

// ServiceClients 管理所有下游 gRPC 连接
type ServiceClients struct {
    UserConn       *grpc.ClientConn
    AuthConn       *grpc.ClientConn
    PermissionConn *grpc.ClientConn
    BizConn        *grpc.ClientConn
}

// NewServiceClients 创建到下游服务的 gRPC 连接
func NewServiceClients(cfg *config.Config, logger *zap.Logger) (*ServiceClients, error) {
    // gRPC 服务配置（重试 + 负载均衡）
    serviceConfig := `{
        "loadBalancingPolicy": "round_robin",
        "methodConfig": [{
            "name": [{}],
            "retryPolicy": {
                "maxAttempts": 3,
                "initialBackoff": "0.1s",
                "maxBackoff": "1s",
                "backoffMultiplier": 2.0,
                "retryableStatusCodes": ["UNAVAILABLE"]
            }
        }]
    }`

    dialOpts := []grpc.DialOption{
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithDefaultServiceConfig(serviceConfig),
    }

    // K8s DNS 模式：直接用 DNS 地址
    // 例如: dns:///user-service.hellogo.svc.cluster.local:50001
    userConn, err := grpc.NewClient(
        "dns:///"+cfg.UserServiceAddr,
        dialOpts...,
    )
    if err != nil {
        return nil, err
    }

    authConn, err := grpc.NewClient(
        "dns:///"+cfg.AuthServiceAddr,
        dialOpts...,
    )
    if err != nil {
        return nil, err
    }

    // ... 其他服务类似 ...

    return &ServiceClients{
        UserConn:       userConn,
        AuthConn:       authConn,
        PermissionConn: permissionConn,
        BizConn:        bizConn,
    }, nil
}
```

### 7.4 迁移步骤

```
迁移流程（渐进式，可回退）：

Step 1: 代码层面支持双模式
├── 添加 SERVICE_DISCOVERY 配置项（"etcd" 或 "dns"）
├── 实现 DNSDiscoverer（§3.4 中的代码）
├── 保留 EtcdDiscoverer 不删除
└── 验证：本地开发继续用 etcd 模式正常运行

Step 2: K8s 环境验证 DNS
├── 部署服务到 K8s（先用 etcd 模式 + K8s 内 etcd）
├── 切换到 DNS 模式
├── 验证服务间调用正常
└── 验证 Headless Service DNS 解析正确

Step 3: 清理
├── 从 docker-compose.yml 中移除 etcd 服务
├── 删除 etcd 相关代码
└── 更新文档
```

---

## 8. 入口与流量管理

### 8.1 Ingress（HTTP 入口）

> Ingress 将集群外部的 HTTP/HTTPS 流量引入集群内部服务。  
> 相当于 K8s 版的 Nginx 反向代理。

```yaml
# deploy/k8s/base/gateway/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hellogo-ingress
  namespace: hellogo
  labels:
    app.kubernetes.io/part-of: hellogo
  annotations:
    # Nginx Ingress Controller 配置
    nginx.ingress.kubernetes.io/rewrite-target: /
    # 请求体大小限制（文件上传）
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
    # 请求超时
    nginx.ingress.kubernetes.io/proxy-read-timeout: "60"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "60"
    # CORS
    nginx.ingress.kubernetes.io/enable-cors: "true"
    nginx.ingress.kubernetes.io/cors-allow-origin: "http://localhost:9003"
    # 限流（每个 IP 每秒 10 个请求）
    nginx.ingress.kubernetes.io/limit-rps: "10"
    nginx.ingress.kubernetes.io/limit-connections: "20"
spec:
  ingressClassName: nginx
  rules:
    # 生产环境使用真实域名
    - host: api.hellogo.local
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: gateway
                port:
                  number: 8000
    # Swagger 文档（可选）
    - host: docs.hellogo.local
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: gateway
                port:
                  number: 8000
```

### 8.2 TLS 证书（自动签发）

> 使用 cert-manager 自动从 Let's Encrypt 签发免费 TLS 证书。

```yaml
# 安装 cert-manager
# kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml

# 创建 Let's Encrypt Issuer
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: letsencrypt-prod-key
    solvers:
      - http01:
          ingress:
            class: nginx
---
# 在 Ingress 中启用自动 TLS
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hellogo-ingress
  namespace: hellogo
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - api.hellogo.local
      secretName: hellogo-tls     # cert-manager 自动创建此 Secret
  rules:
    - host: api.hellogo.local
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: gateway
                port:
                  number: 8000
```

### 8.3 本地开发环境入口

```bash
# minikube 环境
minikube addons enable ingress

# 配置本地域名解析
echo "$(minikube ip) api.hellogo.local" | sudo tee -a /etc/hosts

# kind 环境
# kind 创建时已映射 80/443 端口到宿主机
echo "127.0.0.1 api.hellogo.local" | sudo tee -a /etc/hosts

# 验证
curl http://api.hellogo.local/api/health
```

---

## 9. 持久化存储

### 9.1 存储策略

> 生产环境中，数据库和缓存推荐使用云厂商的托管服务（RDS / ElastiCache）。  
> 开发/测试环境可以在 K8s 内部署，使用 PVC 持久化数据。

| 组件 | 开发/测试环境 | 生产环境 |
|------|-------------|----------|
| MySQL | StatefulSet + PVC | Cloud SQL / RDS / PolarDB |
| Redis | StatefulSet + PVC | ElastiCache / Redis Cloud |
| etcd | 不再需要（用 K8s DNS） | 不再需要 |

### 9.2 MySQL StatefulSet（开发环境用）

```yaml
# deploy/k8s/base/mysql-statefulset.yaml
# 仅用于开发/测试环境，生产环境使用托管数据库
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mysql
  namespace: hellogo
spec:
  serviceName: mysql
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: mysql
  template:
    metadata:
      labels:
        app.kubernetes.io/name: mysql
    spec:
      containers:
        - name: mysql
          image: mysql:8.0
          ports:
            - containerPort: 3306
              name: mysql
          env:
            - name: MYSQL_ROOT_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: hellogo-secrets
                  key: MYSQL_ROOT_PASSWORD
            - name: MYSQL_DATABASE
              value: hellogo
          resources:
            requests:
              cpu: 250m
              memory: 512Mi
            limits:
              cpu: "1"
              memory: 1Gi
          volumeMounts:
            - name: mysql-data
              mountPath: /var/lib/mysql
            - name: mysql-initdb
              mountPath: /docker-entrypoint-initdb.d
          # MySQL 存活检查
          livenessProbe:
            exec:
              command: ["mysqladmin", "ping", "-h", "localhost"]
            initialDelaySeconds: 30
            periodSeconds: 10
          # MySQL 就绪检查
          readinessProbe:
            exec:
              command:
                - sh
                - -c
                - "mysql -u root -p$MYSQL_ROOT_PASSWORD -e 'SELECT 1'"
            initialDelaySeconds: 10
            periodSeconds: 5
      volumes:
        - name: mysql-initdb
          configMap:
            name: mysql-initdb
  # PVC 模板（每个 Pod 自动创建独立 PVC）
  volumeClaimTemplates:
    - metadata:
        name: mysql-data
      spec:
        accessModes: ["ReadWriteOnce"]
        storageClassName: standard    # minikube 默认 StorageClass
        resources:
          requests:
            storage: 5Gi
---
# MySQL Headless Service
apiVersion: v1
kind: Service
metadata:
  name: mysql
  namespace: hellogo
spec:
  clusterIP: None
  selector:
    app.kubernetes.io/name: mysql
  ports:
    - port: 3306
      targetPort: mysql
```

### 9.3 Redis StatefulSet（开发环境用）

```yaml
# deploy/k8s/base/redis-statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis
  namespace: hellogo
spec:
  serviceName: redis
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: redis
  template:
    metadata:
      labels:
        app.kubernetes.io/name: redis
    spec:
      containers:
        - name: redis
          image: redis:7-alpine
          command: ["redis-server", "--appendonly", "yes"]
          ports:
            - containerPort: 6379
              name: redis
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 250m
              memory: 256Mi
          volumeMounts:
            - name: redis-data
              mountPath: /data
          livenessProbe:
            exec:
              command: ["redis-cli", "ping"]
            initialDelaySeconds: 10
            periodSeconds: 10
          readinessProbe:
            exec:
              command: ["redis-cli", "ping"]
            initialDelaySeconds: 5
            periodSeconds: 5
  volumeClaimTemplates:
    - metadata:
        name: redis-data
      spec:
        accessModes: ["ReadWriteOnce"]
        storageClassName: standard
        resources:
          requests:
            storage: 1Gi
---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: hellogo
spec:
  clusterIP: None
  selector:
    app.kubernetes.io/name: redis
  ports:
    - port: 6379
      targetPort: redis
```

### 9.4 数据库迁移 Job

```yaml
# deploy/k8s/base/db-migrate-job.yaml
# K8s Job：一次性任务，完成后自动退出
# 在应用 Deployment 之前运行，确保数据库 schema 是最新的
apiVersion: batch/v1
kind: Job
metadata:
  name: db-migrate
  namespace: hellogo
spec:
  backoffLimit: 3
  template:
    spec:
      restartPolicy: Never
      containers:
        - name: migrate
          image: registry.example.com/hellogo-user:latest
          command: ["./server", "--migrate-only"]
          envFrom:
            - configMapRef:
                name: hellogo-config
            - secretRef:
                name: hellogo-secrets
```

---

## 10. Helm Chart 打包

> Helm 是 K8s 的包管理器（类似 Go 的 `go mod`、Node 的 `npm`）。  
> 用模板（template）+ 值文件（values）来管理不同环境的差异。

### 10.1 Chart.yaml

```yaml
# deploy/helm/hellogo/Chart.yaml
apiVersion: v2
name: hellogo
description: helloGo 微服务管理平台 Helm Chart
type: application
version: 0.1.0          # Chart 版本
appVersion: "1.0.0"     # 应用版本

keywords:
  - go
  - grpc
  - microservices
  - admin

maintainers:
  - name: hellogo-team

dependencies: []
# 如果需要依赖其他 Chart（如 mysql-operator）：
# dependencies:
#   - name: mysql
#     version: "9.x.x"
#     repository: "https://charts.bitnami.com/bitnami"
#     condition: mysql.enabled
```

### 10.2 values.yaml（默认值 — dev 环境）

```yaml
# deploy/helm/hellogo/values.yaml
# 默认值，适用于开发环境

# ===== 全局配置 =====
global:
  imageRegistry: registry.example.com
  imagePullPolicy: Always
  namespace: hellogo

# ===== 通用配置（ConfigMap） =====
config:
  appEnv: development
  logLevel: debug
  logFormat: text
  serviceDiscovery: dns
  dbType: mysql
  dbHost: mysql.hellogo.svc.cluster.local
  dbPort: "3306"
  dbName: hellogo
  redisHost: redis.hellogo.svc.cluster.local
  redisPort: "6379"
  jwtExpires: "1d"
  jwtRefreshExpires: "7d"
  throttleTTL: "60"
  throttleLimit: "100"
  csrfEnabled: "false"
  csrfMode: "header"
  loginMaxFails: "5"
  loginLockTTL: "600"
  port: "8000"
  otelEnabled: "false"
  enableMetrics: "true"
  uploadMaxSize: "10485760"
  uploadAllowedTypes: "image/jpeg,image/png,application/pdf"

# ===== 密钥（Secret） =====
secrets:
  dbUser: hellogo
  dbPassword: changeme
  redisPassword: ""
  jwtSecret: changeme-in-production

# ===== Gateway =====
gateway:
  enabled: true
  replicaCount: 1
  image:
    repository: hellogo-gateway
    tag: latest
  service:
    type: ClusterIP
    port: 8000
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 256Mi
  autoscaling:
    enabled: false
    minReplicas: 2
    maxReplicas: 10
    targetCPUUtilizationPercentage: 70

# ===== User Service =====
userService:
  enabled: true
  replicaCount: 1
  image:
    repository: hellogo-user
    tag: latest
  grpcPort: 50001
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 256Mi
  autoscaling:
    enabled: false

# ===== Auth Service =====
authService:
  enabled: true
  replicaCount: 1
  image:
    repository: hellogo-auth
    tag: latest
  grpcPort: 50002
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 256Mi
  autoscaling:
    enabled: false

# ===== Permission Service =====
permissionService:
  enabled: true
  replicaCount: 1
  image:
    repository: hellogo-permission
    tag: latest
  grpcPort: 50003
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 256Mi
  autoscaling:
    enabled: false

# ===== Biz Service =====
bizService:
  enabled: true
  replicaCount: 1
  image:
    repository: hellogo-biz
    tag: latest
  grpcPort: 50004
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 256Mi
  autoscaling:
    enabled: false

# ===== Ingress =====
ingress:
  enabled: true
  className: nginx
  annotations:
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
  hosts:
    - host: api.hellogo.local
      paths:
        - path: /
          pathType: Prefix
  tls: []
  # 生产环境启用 TLS：
  # tls:
  #   - secretName: hellogo-tls
  #     hosts:
  #       - api.hellogo.local

# ===== MySQL（开发环境内置） =====
mysql:
  enabled: true    # 生产环境设为 false，使用托管数据库
  storage: 5Gi
  storageClass: standard

# ===== Redis（开发环境内置） =====
redis:
  enabled: true    # 生产环境设为 false，使用托管 Redis
  storage: 1Gi
  storageClass: standard
```

### 10.3 模板文件示例

```yaml
# deploy/helm/hellogo/templates/_helpers.tpl
{{/*
通用标签
*/}}
{{- define "hellogo.labels" -}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: hellogo
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end }}

{{/*
通用选择器标签
*/}}
{{- define "hellogo.selectorLabels" -}}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
```

```yaml
# deploy/helm/hellogo/templates/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-config
  namespace: {{ .Values.global.namespace }}
  labels:
    {{- include "hellogo.labels" . | nindent 4 }}
data:
  APP_ENV: {{ .Values.config.appEnv | quote }}
  PORT: {{ .Values.config.port | quote }}
  SERVICE_DISCOVERY: {{ .Values.config.serviceDiscovery | quote }}
  DB_TYPE: {{ .Values.config.dbType | quote }}
  DB_HOST: {{ .Values.config.dbHost | quote }}
  DB_PORT: {{ .Values.config.dbPort | quote }}
  DB_NAME: {{ .Values.config.dbName | quote }}
  REDIS_HOST: {{ .Values.config.redisHost | quote }}
  REDIS_PORT: {{ .Values.config.redisPort | quote }}
  JWT_EXPIRES: {{ .Values.config.jwtExpires | quote }}
  JWT_REFRESH_EXPIRES: {{ .Values.config.jwtRefreshExpires | quote }}
  CSRF_ENABLED: {{ .Values.config.csrfEnabled | quote }}
  CSRF_MODE: {{ .Values.config.csrfMode | quote }}
  THROTTLE_TTL: {{ .Values.config.throttleTTL | quote }}
  THROTTLE_LIMIT: {{ .Values.config.throttleLimit | quote }}
  LOGIN_MAX_FAILS: {{ .Values.config.loginMaxFails | quote }}
  LOGIN_LOCK_TTL: {{ .Values.config.loginLockTTL | quote }}
  USER_SERVICE_ADDR: "user-service.{{ .Values.global.namespace }}.svc.cluster.local:{{ .Values.userService.grpcPort }}"
  AUTH_SERVICE_ADDR: "auth-service.{{ .Values.global.namespace }}.svc.cluster.local:{{ .Values.authService.grpcPort }}"
  PERMISSION_SERVICE_ADDR: "permission-service.{{ .Values.global.namespace }}.svc.cluster.local:{{ .Values.permissionService.grpcPort }}"
  BIZ_SERVICE_ADDR: "biz-service.{{ .Values.global.namespace }}.svc.cluster.local:{{ .Values.bizService.grpcPort }}"
  OTEL_ENABLED: {{ .Values.config.otelEnabled | quote }}
  ENABLE_METRICS: {{ .Values.config.enableMetrics | quote }}
```

```yaml
# deploy/helm/hellogo/templates/gateway/deployment.yaml
{{- if .Values.gateway.enabled }}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}-gateway
  namespace: {{ .Values.global.namespace }}
  labels:
    app.kubernetes.io/name: gateway
    {{- include "hellogo.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.gateway.replicaCount }}
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app.kubernetes.io/name: gateway
      {{- include "hellogo.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: gateway
        {{- include "hellogo.selectorLabels" . | nindent 8 }}
    spec:
      terminationGracePeriodSeconds: 45
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
      containers:
        - name: gateway
          image: "{{ .Values.global.imageRegistry }}/{{ .Values.gateway.image.repository }}:{{ .Values.gateway.image.tag }}"
          imagePullPolicy: {{ .Values.global.imagePullPolicy }}
          envFrom:
            - configMapRef:
                name: {{ .Release.Name }}-config
            - secretRef:
                name: {{ .Release.Name }}-secrets
          ports:
            - name: http
              containerPort: 8000
            - name: health
              containerPort: 8080
          resources:
            {{- toYaml .Values.gateway.resources | nindent 12 }}
          # 健康检查探针（与 §5.2 一致，使用 /healthz/* 端点）
          startupProbe:
            httpGet:
              path: /healthz/startup
              port: health
            initialDelaySeconds: 5
            periodSeconds: 3
            failureThreshold: 10
          livenessProbe:
            httpGet:
              path: /healthz/liveness
              port: health
            initialDelaySeconds: 0
            periodSeconds: 10
            timeoutSeconds: 3
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /healthz/readiness
              port: health
            initialDelaySeconds: 0
            periodSeconds: 5
            timeoutSeconds: 3
            failureThreshold: 3
          lifecycle:
            preStop:
              exec:
                command: ["sh", "-c", "sleep 5"]
{{- end }}
```

```yaml
# deploy/helm/hellogo/templates/gateway/hpa.yaml
{{- if and .Values.gateway.enabled .Values.gateway.autoscaling.enabled }}
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{ .Release.Name }}-gateway
  namespace: {{ .Values.global.namespace }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ .Release.Name }}-gateway
  minReplicas: {{ .Values.gateway.autoscaling.minReplicas }}
  maxReplicas: {{ .Values.gateway.autoscaling.maxReplicas }}
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: {{ .Values.gateway.autoscaling.targetCPUUtilizationPercentage }}
{{- end }}
```

### 10.4 环境 Values 覆盖

```yaml
# deploy/helm/hellogo/values-production.yaml
# 生产环境覆盖项

global:
  imagePullPolicy: IfNotPresent

config:
  appEnv: production
  logLevel: warn
  logFormat: json
  csrfEnabled: "true"
  otelEnabled: "true"

# 生产环境不使用集群内数据库
mysql:
  enabled: false

redis:
  enabled: false

# 使用外部数据库地址
config:
  dbHost: "your-rds-endpoint.rds.amazonaws.com"
  redisHost: "your-elasticache-endpoint.cache.amazonaws.com"

# 所有服务增加副本数和资源
gateway:
  replicaCount: 3
  resources:
    requests:
      cpu: 500m
      memory: 512Mi
    limits:
      cpu: "2"
      memory: 1Gi
  autoscaling:
    enabled: true
    minReplicas: 3
    maxReplicas: 20
    targetCPUUtilizationPercentage: 60

userService:
  replicaCount: 3
  resources:
    requests:
      cpu: 250m
      memory: 256Mi
    limits:
      cpu: "1"
      memory: 512Mi
  autoscaling:
    enabled: true
    minReplicas: 3
    maxReplicas: 10

# ... 其他服务类似 ...

ingress:
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  tls:
    - secretName: hellogo-tls
      hosts:
        - api.hellogo.com
```

### 10.5 Helm 操作命令

```bash
# ===== 安装 =====
# 开发环境
helm install hellogo ./deploy/helm/hellogo \
    --namespace hellogo --create-namespace

# 生产环境（使用生产 values）
helm install hellogo ./deploy/helm/hellogo \
    --namespace hellogo --create-namespace \
    -f ./deploy/helm/hellogo/values-production.yaml

# ===== 升级 =====
helm upgrade hellogo ./deploy/helm/hellogo \
    --namespace hellogo \
    -f ./deploy/helm/hellogo/values-production.yaml

# ===== 回滚 =====
helm rollback hellogo 1 --namespace hellogo    # 回滚到版本 1
helm history hellogo --namespace hellogo       # 查看版本历史

# ===== 卸载 =====
helm uninstall hellogo --namespace hellogo

# ===== 调试 =====
helm template hellogo ./deploy/helm/hellogo    # 渲染模板（不部署，查看生成的 YAML）
helm lint ./deploy/helm/hellogo                # 检查 Chart 语法
helm diff upgrade hellogo ./deploy/helm/hellogo  # 查看变更（需安装 helm-diff 插件）
```

### 10.6 Helm 测试

```yaml
# deploy/helm/hellogo/templates/tests/test-connection.yaml
# Helm 内置测试：helm test hellogo
apiVersion: v1
kind: Pod
metadata:
  name: {{ .Release.Name }}-test-connection
  namespace: {{ .Values.global.namespace }}
  annotations:
    "helm.sh/hook": test
spec:
  containers:
    - name: test
      image: busybox:1.36
      command:
        - sh
        - -c
        - |
          # 测试 Gateway 可达
          wget -q --spider http://gateway:8000/api/health || exit 1
          echo "Gateway 健康检查通过"

          # 测试 User Service DNS 解析
          nslookup user-service || exit 1
          echo "User Service DNS 解析通过"
  restartPolicy: Never
```

---

## 11. CI/CD 流水线

### 11.1 流水线架构

```
代码提交
  │
  ▼
┌─────────────────────────────────────────────────────┐
│  CI Pipeline（每次 push / PR 触发）                   │
│                                                      │
│  ┌──────┐   ┌──────┐   ┌───────┐   ┌──────────┐   │
│  │ lint │──→│ test │──→│ build │──→│ push img │   │
│  └──────┘   └──────┘   └───────┘   └──────────┘   │
│                                                      │
│  golangci     go test     go build    docker push    │
│  -lint       -cover      多服务       到 Registry    │
└─────────────────────────────────────────────────────┘
  │
  ▼ （merge 到 main 分支时）
┌─────────────────────────────────────────────────────┐
│  CD Pipeline                                         │
│                                                      │
│  ┌──────────┐   ┌──────────────┐   ┌────────────┐  │
│  │ helm     │──→│ kubectl/helm │──→│ 健康检查   │  │
│  │ package  │   │ upgrade      │   │ 验证       │  │
│  └──────────┘   └──────────────┘   └────────────┘  │
│                                                      │
│  打包 Chart    部署到 K8s      验证部署成功           │
└─────────────────────────────────────────────────────┘
```

### 11.2 GitHub Actions 完整流水线

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  GO_VERSION: "1.23"
  REGISTRY: ghcr.io
  IMAGE_PREFIX: ${{ github.repository }}

jobs:
  # ===== 代码质量 =====
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest

  # ===== 单元测试 =====
  test:
    runs-on: ubuntu-latest
    services:
      redis:
        image: redis:7-alpine
        ports: ["6379:6379"]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
      - name: 运行测试
        run: go test ./... -coverprofile=coverage.out -count=1
      - name: 检查覆盖率
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
          echo "覆盖率: $coverage"
          # 覆盖率不低于 60%
          echo "$coverage" | awk '{gsub(/%/,""); if ($1 < 60) exit 1}'
      - uses: actions/upload-artifact@v4
        with:
          name: coverage
          path: coverage.out

  # ===== 构建 Docker 镜像 =====
  build:
    needs: [lint, test]
    runs-on: ubuntu-latest
    if: github.event_name == 'push'
    strategy:
      matrix:
        service: [gateway, user, auth, permission, biz]
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v4

      - name: 登录容器仓库
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: 提取版本信息
        id: meta
        run: |
          SHA_SHORT=$(echo ${{ github.sha }} | cut -c1-7)
          echo "sha_short=$SHA_SHORT" >> $GITHUB_OUTPUT
          echo "image_name=${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}-${{ matrix.service }}" >> $GITHUB_OUTPUT

      - name: 构建并推送镜像
        uses: docker/build-push-action@v5
        with:
          context: .
          file: deploy/docker/Dockerfile
          build-args: SERVICE=${{ matrix.service }}
          push: true
          tags: |
            ${{ steps.meta.outputs.image_name }}:latest
            ${{ steps.meta.outputs.image_name }}:${{ steps.meta.outputs.sha_short }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

      - name: 镜像安全扫描（Trivy）
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: "${{ steps.meta.outputs.image_name }}:${{ steps.meta.outputs.sha_short }}"
          format: table
          severity: CRITICAL,HIGH
```

```yaml
# .github/workflows/cd-dev.yml
name: CD - Dev

on:
  workflow_run:
    workflows: ["CI"]
    types: [completed]
    branches: [develop]

env:
  REGISTRY: ghcr.io
  IMAGE_PREFIX: ${{ github.repository }}

jobs:
  deploy-dev:
    if: ${{ github.event.workflow_run.conclusion == 'success' }}
    runs-on: ubuntu-latest
    environment: development
    steps:
      - uses: actions/checkout@v4

      - name: 配置 kubectl
        uses: azure/setup-kubectl@v4

      - name: 配置 kubeconfig
        run: |
          mkdir -p ~/.kube
          echo "${{ secrets.KUBECONFIG_DEV }}" | base64 -d > ~/.kube/config

      - name: 安装 Helm
        uses: azure/setup-helm@v4

      - name: 提取镜像标签
        id: meta
        run: |
          SHA_SHORT=$(echo ${{ github.sha }} | cut -c1-7)
          echo "tag=$SHA_SHORT" >> $GITHUB_OUTPUT

      - name: Helm 部署到 dev 环境
        run: |
          helm upgrade --install hellogo ./deploy/helm/hellogo \
            --namespace hellogo-dev --create-namespace \
            --set global.imageRegistry=${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }} \
            --set gateway.image.tag=${{ steps.meta.outputs.tag }} \
            --set userService.image.tag=${{ steps.meta.outputs.tag }} \
            --set authService.image.tag=${{ steps.meta.outputs.tag }} \
            --set permissionService.image.tag=${{ steps.meta.outputs.tag }} \
            --set bizService.image.tag=${{ steps.meta.outputs.tag }} \
            --wait --timeout 5m

      - name: 验证部署
        run: |
          kubectl -n hellogo-dev rollout status deployment/hellogo-gateway --timeout=120s
          kubectl -n hellogo-dev rollout status deployment/hellogo-user-service --timeout=120s
          # 运行 Helm 测试
          helm test hellogo --namespace hellogo-dev
```

```yaml
# .github/workflows/cd-production.yml
name: CD - Production

on:
  push:
    tags: ["v*"]    # 只在打 tag 时触发

jobs:
  deploy-production:
    runs-on: ubuntu-latest
    environment:
      name: production
      url: https://api.hellogo.com
    steps:
      - uses: actions/checkout@v4

      - name: 配置 kubectl
        uses: azure/setup-kubectl@v4

      - name: 配置 kubeconfig
        run: |
          mkdir -p ~/.kube
          echo "${{ secrets.KUBECONFIG_PROD }}" | base64 -d > ~/.kube/config

      - name: 安装 Helm
        uses: azure/setup-helm@v4

      - name: 提取版本号
        id: meta
        run: echo "version=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT

      - name: Helm 部署到生产环境
        run: |
          helm upgrade --install hellogo ./deploy/helm/hellogo \
            --namespace hellogo --create-namespace \
            -f ./deploy/helm/hellogo/values-production.yaml \
            --set global.imageRegistry=${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }} \
            --set gateway.image.tag=${{ steps.meta.outputs.version }} \
            --set userService.image.tag=${{ steps.meta.outputs.version }} \
            --set authService.image.tag=${{ steps.meta.outputs.version }} \
            --set permissionService.image.tag=${{ steps.meta.outputs.version }} \
            --set bizService.image.tag=${{ steps.meta.outputs.version }} \
            --wait --timeout 10m

      - name: 验证生产部署
        run: |
          kubectl -n hellogo rollout status deployment/hellogo-gateway --timeout=300s
          # 冒烟测试
          curl -sf https://api.hellogo.com/api/health || exit 1

      - name: 部署失败自动回滚
        if: failure()
        run: |
          helm rollback hellogo --namespace hellogo
          echo "部署失败，已自动回滚"
```

### 11.3 Makefile 集成

```makefile
# Makefile 新增 K8s 相关命令

# ===== K8s 部署 =====
KUBE_NS ?= hellogo
HELM_CHART := deploy/helm/hellogo

.PHONY: k8s-deploy-dev
k8s-deploy-dev:
	helm upgrade --install hellogo $(HELM_CHART) \
		--namespace $(KUBE_NS)-dev --create-namespace \
		--wait --timeout 5m

.PHONY: k8s-deploy-prod
k8s-deploy-prod:
	helm upgrade --install hellogo $(HELM_CHART) \
		--namespace $(KUBE_NS) --create-namespace \
		-f $(HELM_CHART)/values-production.yaml \
		--wait --timeout 10m

.PHONY: k8s-rollback
k8s-rollback:
	helm rollback hellogo --namespace $(KUBE_NS)

.PHONY: k8s-status
k8s-status:
	kubectl -n $(KUBE_NS) get all
	kubectl -n $(KUBE_NS) get hpa

.PHONY: k8s-logs
k8s-logs:
	@echo "用法: make k8s-logs SVC=gateway"
	kubectl -n $(KUBE_NS) logs -f deployment/hellogo-$(SVC)

.PHONY: k8s-test
k8s-test:
	helm test hellogo --namespace $(KUBE_NS)

.PHONY: k8s-port-forward
k8s-port-forward:
	kubectl -n $(KUBE_NS) port-forward svc/gateway 8000:8000

# ===== Docker 镜像 =====
REGISTRY ?= registry.example.com
VERSION ?= latest

.PHONY: docker-build-all
docker-build-all:
	@for svc in gateway user auth permission biz; do \
		echo "构建 $$svc ..."; \
		docker build --build-arg SERVICE=$$svc \
			-t $(REGISTRY)/hellogo-$$svc:$(VERSION) \
			-f deploy/docker/Dockerfile .; \
	done

.PHONY: docker-push-all
docker-push-all:
	@for svc in gateway user auth permission biz; do \
		echo "推送 $$svc ..."; \
		docker push $(REGISTRY)/hellogo-$$svc:$(VERSION); \
	done
```

---

## 12. 可观测性（K8s 原生方案）

### 12.1 监控架构

```
┌────────────────────────────────────────────────────────────────┐
│                     可观测性架构                                 │
│                                                                 │
│  ┌──────────────────┐    ┌─────────────────────┐               │
│  │   微服务 Pod     │    │   Prometheus         │               │
│  │                  │    │   (指标采集)          │               │
│  │  /metrics ───────┼────│→ 每 15s 拉取一次    │               │
│  │  (Prometheus     │    │                      │               │
│  │   client_golang) │    │  告警规则 → AlertManager             │
│  └──────────────────┘    └──────────┬───────────┘               │
│                                      │                          │
│                                      ▼                          │
│  ┌──────────────────┐    ┌─────────────────────┐               │
│  │  Grafana         │    │  Jaeger              │               │
│  │  (可视化面板)     │←───│  (链路追踪)          │               │
│  │                  │    │                      │               │
│  │  • QPS 面板      │    │  • TraceID 查询      │               │
│  │  • 延迟面板      │    │  • 调用链可视化       │               │
│  │  • 错误率面板    │    │  • Span 分析          │               │
│  │  • 资源使用面板  │    │                      │               │
│  └──────────────────┘    └─────────────────────┘               │
└────────────────────────────────────────────────────────────────┘
```

### 12.2 Prometheus 指标暴露

> Phase 1 已经集成了 prometheus/client_golang。在 K8s 中，通过 Pod 注解让 Prometheus 自动发现。

```go
// internal/shared/metrics/metrics.go
// 所有微服务共享的指标定义

package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // gRPC 请求总数
    GRPCRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "grpc_requests_total",
            Help: "gRPC 请求总数",
        },
        []string{"service", "method", "status"},
    )

    // gRPC 请求延迟
    GRPCRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "grpc_request_duration_seconds",
            Help:    "gRPC 请求延迟（秒）",
            Buckets: prometheus.DefBuckets,
        },
        []string{"service", "method"},
    )

    // gRPC 活跃连接数
    GRPCActiveConnections = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "grpc_active_connections",
            Help: "gRPC 活跃连接数",
        },
        []string{"service"},
    )

    // 业务指标：登录尝试
    LoginAttempts = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "login_attempts_total",
            Help: "登录尝试总数",
        },
        []string{"status"},  // "success" / "failure" / "locked"
    )
)
```

**在 Pod 注解中配置 Prometheus 自动发现：**

```yaml
# 在 Deployment 的 Pod template 中添加注解
metadata:
  annotations:
    prometheus.io/scrape: "true"     # 允许 Prometheus 抓取
    prometheus.io/port: "8080"       # 指标端口
    prometheus.io/path: "/metrics"   # 指标路径
```

### 12.3 Prometheus + Grafana 部署（kube-prometheus-stack）

```bash
# 安装 kube-prometheus-stack（包含 Prometheus + Grafana + AlertManager）
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm install kube-prometheus prometheus-community/kube-prometheus-stack \
    --namespace monitoring --create-namespace \
    --set grafana.adminPassword=admin123 \
    --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false

# 访问 Grafana（端口转发）
kubectl port-forward -n monitoring svc/kube-prometheus-grafana 3000:80
# 浏览器打开 http://localhost:3000
# 用户名: admin  密码: admin123
```

### 12.4 Grafana Dashboard 关键面板

```
推荐的 Grafana 面板：

1. 服务概览（Overview）
   ├── QPS（每秒请求数）— 按服务分组
   ├── 错误率（Error Rate）— 5xx / 总请求
   ├── P50 / P95 / P99 延迟
   └── 活跃连接数

2. gRPC 详情
   ├── 各 RPC 方法的请求量
   ├── 各 RPC 方法的延迟分布
   ├── 错误码分布（UNAVAILABLE / INTERNAL / ...）
   └── 重试次数

3. 基础设施
   ├── Pod CPU / 内存使用率
   ├── Pod 重启次数
   ├── HPA 副本数变化
   └── PVC 使用率

4. 业务指标
   ├── 登录成功/失败趋势
   ├── 活跃用户数
   └── 文件上传量
```

### 12.5 告警规则

```yaml
# 关键告警规则示例
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: hellogo-alerts
  namespace: hellogo
spec:
  groups:
    - name: hellogo.rules
      rules:
        # 服务不可用
        - alert: ServiceDown
          expr: up{namespace="hellogo"} == 0
          for: 1m
          labels:
            severity: critical
          annotations:
            summary: "{{ $labels.pod }} 不可用"

        # 错误率过高
        - alert: HighErrorRate
          expr: |
            sum(rate(grpc_requests_total{status!="OK"}[5m]))
            / sum(rate(grpc_requests_total[5m])) > 0.05
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "gRPC 错误率超过 5%"

        # Pod 频繁重启
        - alert: PodCrashLooping
          expr: |
            increase(kube_pod_container_status_restarts_total{namespace="hellogo"}[1h]) > 3
          for: 10m
          labels:
            severity: warning
          annotations:
            summary: "{{ $labels.pod }} 1小时内重启超过 3 次"

        # HPA 达到最大副本数
        - alert: HPAMaxedOut
          expr: |
            kube_horizontalpodautoscaler_status_current_replicas
            == kube_horizontalpodautoscaler_spec_max_replicas
          for: 15m
          labels:
            severity: warning
          annotations:
            summary: "{{ $labels.horizontalpodautoscaler }} 已达到最大副本数"
```

---

## 13. 生产加固

### 13.1 HPA 自动扩缩容

> HPA（Horizontal Pod Autoscaler）根据 CPU / 内存 / 自定义指标自动调整 Pod 副本数。

```yaml
# deploy/k8s/overlays/production/patches/hpa.yaml
# Gateway HPA（对外入口，按 QPS 扩缩）
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: gateway
  namespace: hellogo
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: gateway
  minReplicas: 3
  maxReplicas: 20
  metrics:
    # CPU 利用率
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 60
    # 内存利用率
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 70
  behavior:
    # 扩容策略：快速扩容
    scaleUp:
      stabilizationWindowSeconds: 30     # 30s 内持续高负载才扩
      policies:
        - type: Pods
          value: 4                       # 每次最多扩 4 个
          periodSeconds: 60
    # 缩容策略：缓慢缩容（避免反复震荡）
    scaleDown:
      stabilizationWindowSeconds: 300    # 5 分钟内持续低负载才缩
      policies:
        - type: Pods
          value: 1                       # 每次只缩 1 个
          periodSeconds: 120
```

### 13.2 PodDisruptionBudget（PDB）

> PDB 保障在维护操作（如节点升级、驱逐）期间，始终有最小数量的 Pod 可用。

```yaml
# 每个微服务都需要 PDB
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: gateway-pdb
  namespace: hellogo
spec:
  # 始终保证至少 50% 的 Pod 可用
  maxUnavailable: "50%"
  selector:
    matchLabels:
      app.kubernetes.io/name: gateway
---
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: user-service-pdb
  namespace: hellogo
spec:
  # 始终保证至少有 2 个 Pod 可用
  minAvailable: 2
  selector:
    matchLabels:
      app.kubernetes.io/name: user-service
```

### 13.3 NetworkPolicy（网络隔离）

> 默认情况下，K8s 集群内所有 Pod 可以互相通信。NetworkPolicy 限制网络访问范围。

```yaml
# 只允许 Gateway 访问后端 gRPC 服务
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: backend-grpc-only-from-gateway
  namespace: hellogo
spec:
  podSelector:
    matchExpressions:
      - key: app.kubernetes.io/component
        operator: In
        values: ["backend"]
  policyTypes:
    - Ingress
  ingress:
    # 只允许来自 Gateway Pod 的 gRPC 流量
    - from:
        - podSelector:
            matchLabels:
              app.kubernetes.io/name: gateway
      ports:
        - port: grpc
          protocol: TCP
    # 允许 Prometheus 抓取指标
    - from:
        - namespaceSelector:
            matchLabels:
              name: monitoring
      ports:
        - port: http
          protocol: TCP
---
# Gateway 只允许来自 Ingress Controller 的流量
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: gateway-only-from-ingress
  namespace: hellogo
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: gateway
  policyTypes:
    - Ingress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: ingress-nginx
      ports:
        - port: http
          protocol: TCP
---
# 只允许后端服务访问 MySQL 和 Redis
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: db-only-from-backend
  namespace: hellogo
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: mysql
  policyTypes:
    - Ingress
  ingress:
    - from:
        - podSelector:
            matchExpressions:
              - key: app.kubernetes.io/component
                operator: In
                values: ["backend"]
      ports:
        - port: 3306
          protocol: TCP
```

### 13.4 资源配额（ResourceQuota）

```yaml
# 限制 hellogo 命名空间的总资源使用量
apiVersion: v1
kind: ResourceQuota
metadata:
  name: hellogo-quota
  namespace: hellogo
spec:
  hard:
    requests.cpu: "4"
    requests.memory: 8Gi
    limits.cpu: "8"
    limits.memory: 16Gi
    pods: "50"
    services: "20"
    persistentvolumeclaims: "10"
```

### 13.5 安全检查清单

```
生产环境安全检查清单：

容器安全
  ✓ 非 root 用户运行（runAsNonRoot: true）
  ✓ 只读根文件系统（readOnlyRootFilesystem: true）
  ✓ 禁止特权提升（allowPrivilegeEscalation: false）
  ✓ 删除所有 Linux capabilities（drop: ["ALL"]）
  ✓ 使用 distroless 或最小基础镜像
  ✓ 镜像安全扫描（Trivy）

网络安全
  ✓ NetworkPolicy 限制 Pod 间通信
  ✓ Ingress TLS（HTTPS）
  ✓ 内部 gRPC 可选 mTLS（Istio / Linkerd）

密钥管理
  ✓ 使用 Sealed Secrets / External Secrets / Vault
  ✓ Secret 不以明文提交到 Git
  ✓ 定期轮换密钥（JWT Secret、DB 密码）

访问控制
  ✓ RBAC 最小权限原则
  ✓ ServiceAccount 专用（每个微服务一个）
  ✓ 禁止使用 default ServiceAccount

镜像安全
  ✓ 固定镜像版本（不使用 :latest）
  ✓ 使用私有 Registry
  ✓ imagePullPolicy: IfNotPresent（生产环境）
```

---

## 14. 多环境策略

### 14.1 环境拓扑

```
┌───────────────────────────────────────────────────────────────┐
│                        多环境策略                              │
├───────────┬─────────────┬──────────────┬──────────────────────┤
│  环境      │ Namespace   │ K8s 集群     │ 用途                 │
├───────────┼─────────────┼──────────────┼──────────────────────┤
│ 本地开发   │ —           │ minikube/kind│ 开发者个人调试        │
│ dev       │ hellogo-dev │ 共享 dev 集群 │ 联调测试              │
│ staging   │ hellogo-stg │ 独立集群      │ 预发布验证            │
│ production│ hellogo     │ 生产集群      │ 线上服务              │
└───────────┴─────────────┴──────────────┴──────────────────────┘
```

### 14.2 Kustomize Overlay 方式

```yaml
# deploy/k8s/overlays/dev/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: hellogo-dev

resources:
  - ../../base

# dev 环境的覆盖项
patches:
  # 减少副本数
  - target:
      kind: Deployment
    patch: |
      - op: replace
        path: /spec/replicas
        value: 1

  # 降低资源限制
  - target:
      kind: Deployment
    patch: |
      - op: replace
        path: /spec/template/spec/containers/0/resources/requests/cpu
        value: 50m
      - op: replace
        path: /spec/template/spec/containers/0/resources/requests/memory
        value: 64Mi
```

```yaml
# deploy/k8s/overlays/production/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: hellogo

resources:
  - ../../base
  - patches/hpa.yaml
  - patches/pdb.yaml
  - patches/network-policy.yaml

# 生产环境使用固定版本镜像
images:
  - name: registry.example.com/hellogo-gateway
    newTag: v1.0.0
  - name: registry.example.com/hellogo-user
    newTag: v1.0.0
  # ... 其他服务

patches:
  - target:
      kind: Deployment
    patch: |
      - op: replace
        path: /spec/replicas
        value: 3
```

### 14.3 Helm Values 方式（推荐）

```bash
# 本地开发（minikube）
helm install hellogo ./deploy/helm/hellogo \
    --namespace hellogo-dev --create-namespace

# dev 环境
helm upgrade --install hellogo ./deploy/helm/hellogo \
    --namespace hellogo-dev \
    --set global.imageRegistry=ghcr.io/your-org \
    --set gateway.image.tag=abc1234

# staging 环境
helm upgrade --install hellogo ./deploy/helm/hellogo \
    --namespace hellogo-stg \
    -f ./deploy/helm/hellogo/values-staging.yaml

# production 环境
helm upgrade --install hellogo ./deploy/helm/hellogo \
    --namespace hellogo \
    -f ./deploy/helm/hellogo/values-production.yaml
```

### 14.4 环境差异对照表

| 配置项 | 本地开发 | dev | staging | production |
|--------|---------|-----|---------|------------|
| 副本数 | 1 | 1 | 2 | 3+ |
| HPA | 关 | 关 | 开 | 开 |
| DB 来源 | 集群内 MySQL | 集群内 MySQL | 托管 RDS | 托管 RDS |
| Redis 来源 | 集群内 Redis | 集群内 Redis | ElastiCache | ElastiCache |
| 日志级别 | debug | debug | info | warn |
| 日志格式 | text | json | json | json |
| TLS | 无 | 无 | Let's Encrypt | Let's Encrypt |
| CORS | localhost:9003 | dev.example.com | stg.example.com | example.com |
| 限流 | 关 | 开（宽松） | 开 | 开（严格） |
| 镜像标签 | :latest | :sha-xxxx | :v1.0.0-rc1 | :v1.0.0 |
| 网络策略 | 无 | 基础 | 完整 | 完整 |

---

## 15. 阶段总结与时间规划

### 开发时间估算（初学者视角）

| 阶段 | 内容 | 学习目标 | 预计时间 |
|------|------|----------|----------|
| §1 K8s 基础 | 概念 + minikube + kubectl | 理解 Pod/Deployment/Service | 3-4 天 |
| §2 架构演进 | Compose vs K8s 对比 | 理解架构差异 | 1 天 |
| §3 应用改造 | 探针 + 优雅停机 + 配置外部化 | 掌握云原生应用设计 | 3-4 天 |
| §4 Dockerfile | 安全加固 + 多架构 | 生产级容器镜像 | 1-2 天 |
| §5 K8s 清单 | Deployment + Service + Namespace | 手写 YAML 理解概念 | 2-3 天 |
| §6 配置管理 | ConfigMap + Secret | 配置与密钥管理 | 1-2 天 |
| §7 服务发现 | etcd → K8s DNS | 理解 K8s DNS 机制 | 2-3 天 |
| §8 入口管理 | Ingress + TLS | HTTP 流量路由 | 1-2 天 |
| §9 持久化 | PVC + StatefulSet | 理解 K8s 存储 | 1-2 天 |
| §10 Helm | Chart 开发 + values | 模板化部署 | 2-3 天 |
| §11 CI/CD | GitHub Actions 流水线 | 自动化构建部署 | 2-3 天 |
| §12 可观测性 | Prometheus + Grafana | K8s 监控方案 | 2-3 天 |
| §13 生产加固 | HPA + PDB + NetworkPolicy | 生产级安全与弹性 | 2-3 天 |
| §14 多环境 | Kustomize / Helm overlays | 多环境管理 | 1-2 天 |
| **合计** | | | **24-36 天** |

### 建议的学习节奏

```
Week 1: K8s 基础（§1）+ 架构演进（§2）+ 应用改造（§3）
         └→ 里程碑：理解 K8s 核心概念，应用支持健康检查

Week 2: Dockerfile（§4）+ K8s 清单（§5）+ 配置管理（§6）
         └→ 里程碑：手写 YAML 在 minikube 上运行单个服务

Week 3: 服务发现（§7）+ 入口管理（§8）+ 持久化（§9）
         └→ 里程碑：全部微服务在 K8s 上运行，curl 通过 Ingress 调用

Week 4: Helm（§10）+ CI/CD（§11）
         └→ 里程碑：helm install 一键部署，Git push 自动部署到 dev

Week 5: 可观测性（§12）+ 生产加固（§13）+ 多环境（§14）
         └→ 里程碑：Grafana 面板 + HPA 自动扩容 + 多环境部署
```

### 每完成一个阶段的自检清单

- [ ] 所有微服务在 minikube 上正常运行
- [ ] 健康检查探针全部通过（`kubectl describe pod` 无重启）
- [ ] `curl` 通过 Ingress 访问 Gateway 成功
- [ ] 服务间 gRPC 调用通过 K8s DNS 正常工作
- [ ] ConfigMap / Secret 正确注入，无硬编码配置
- [ ] Helm Chart 可以一键安装/卸载
- [ ] CI/CD 流水线：push 代码后自动部署到 dev
- [ ] Grafana 面板显示所有服务的指标
- [ ] HPA 压测时自动扩容，空闲时自动缩容
- [ ] NetworkPolicy 正确隔离了网络访问
- [ ] Pod 重启后自动恢复（验证自愈能力）
- [ ] 滚动更新零停机（验证 rolling update 策略）

### 三阶段全景回顾

```
Phase 1: Fiber 单体应用（已完成）
├── 用户/角色/权限/菜单/部门/字典/日志/上传
├── JWT 认证 + RBAC 鉴权
├── SQLite / MySQL / PostgreSQL
└── Docker Compose 部署

Phase 2: gRPC 微服务（规划中）
├── 拆分为 4 个 gRPC 服务 + 1 个 API Gateway
├── etcd 服务发现 + gRPC 负载均衡
├── OpenTelemetry 链路追踪
└── Docker Compose 全栈部署

Phase 3: Kubernetes 部署（本文档）
├── K8s 原生服务发现（DNS）
├── Helm Chart 打包 + CI/CD
├── HPA 自动扩缩容 + 滚动更新
├── Prometheus + Grafana 监控
├── NetworkPolicy + 安全加固
└── 多环境（dev / staging / production）
```
