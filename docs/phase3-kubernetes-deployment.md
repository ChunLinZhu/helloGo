# 第三阶段开发计划：Kubernetes 部署（minikube + Helm + Lens）

> 基于第二阶段 gRPC 微服务架构，部署到本地 minikube 集群  
> 使用 Helm 管理 K8s 资源，Lens 可视化控制容器，front-end 提供 Web UI  
> 面向 Go 初学者设计，侧重动手实操

---

## 目录

- [0. 总体架构](#0-总体架构)
- [1. 环境准备](#1-环境准备)
- [2. 应用层改造](#2-应用层改造)
- [3. Dockerfile](#3-dockerfile)
- [4. Helm Chart](#4-helm-chart)
- [5. 部署与验证](#5-部署与验证)
- [6. Lens 可视化管理](#6-lens-可视化管理)
- [7. 日常运维操作](#7-日常运维操作)
- [8. Makefile 目标](#8-makefile-目标)
- [9. 常见问题排查](#9-常见问题排查)

---

## 0. 总体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        minikube 集群                             │
│                                                                  │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐       │
│  │  Frontend    │    │   Gateway    │    │    MySQL     │       │
│  │  (nginx)     │    │  (Fiber)     │    │  (StatefulSet)│      │
│  │  :30090      │    │  :30080      │    │  :3306       │       │
│  │  NodePort    │    │  NodePort    │    │  ClusterIP   │       │
│  └──────────────┘    └──────┬───────┘    └──────────────┘       │
│                             │ gRPC                               │
│         ┌───────────────────┼───────────────────┐               │
│         ▼                   ▼                   ▼               │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐        │
│  │ User Service │   │ Auth Service │   │ Perm Service │        │
│  │   :50001     │   │   :50002     │   │   :50003     │        │
│  │  ClusterIP   │   │  ClusterIP   │   │  ClusterIP   │        │
│  └──────────────┘   └──────────────┘   └──────────────┘        │
│         │                                                        │
│  ┌──────────────┐   ┌──────────────┐                           │
│  │  Biz Service │   │    Redis     │                           │
│  │   :50004     │   │  (StatefulSet)│                          │
│  │  ClusterIP   │   │   :6379      │                           │
│  └──────────────┘   └──────────────┘                           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
         ▲                           ▲
         │ HTTP                      │ HTTP
         │                           │
    ┌────┴─────┐               ┌─────┴────┐
    │  浏览器  │               │   Lens   │
    │ 前端界面 │               │ K8s GUI  │
    └──────────┘               └──────────┘
```

**访问方式：**

| 服务 | K8s 内部地址 | 外部访问地址 |
|------|-------------|-------------|
| 前端 | `frontend:80` | `http://$(minikube ip):30090` |
| Gateway | `gateway:8000` | `http://$(minikube ip):30080` |
| User Service | `user-service:50001` | 仅集群内 |
| Auth Service | `auth-service:50002` | 仅集群内 |
| Permission Service | `permission-service:50003` | 仅集群内 |
| Biz Service | `biz-service:50004` | 仅集群内 |
| MySQL | `mysql:3306` | 仅集群内 |
| Redis | `redis:6379` | 仅集群内 |

---

## 1. 环境准备

### 1.1 安装 minikube

```bash
# Ubuntu
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube

# 启动集群（Docker 驱动，分配足够资源）
minikube start \
  --driver=docker \
  --cpus=4 \
  --memory=8192 \
  --disk-size=40g \
  --kubernetes-version=v1.30.0

# 验证
minikube status
kubectl cluster-info
kubectl get nodes
```

### 1.2 安装 kubectl

```bash
# minikube 自带 kubectl，也可单独安装
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install kubectl /usr/local/bin/kubectl

# 启用自动补全（推荐）
echo 'source <(kubectl completion bash)' >> ~/.bashrc
echo 'alias k=kubectl' >> ~/.bashrc
echo 'complete -o default -F __start_kubectl k' >> ~/.bashrc
source ~/.bashrc
```

### 1.3 安装 Helm

```bash
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# 验证
helm version
```

### 1.4 Lens 连接集群（已本地安装）

1. 打开 Lens
2. 左侧 **Catalog → Clusters** → 选择 **minikube**
3. 点击 **Connect**
4. 左上角命名空间筛选器中选择 **hellogo**

> Lens 自动读取 `~/.kube/config`，minikube 启动后会自动出现在集群列表中。

### 1.5 配置 minikube Docker 环境

```bash
# 使用 minikube 内置的 Docker daemon 构建镜像
# 这样不需要推送到远程 registry
eval $(minikube docker-env)

# 验证
docker ps  # 应能看到 minikube 系统容器
```

> **注意**：每次新开终端都需要重新执行 `eval $(minikube docker-env)`

---

## 2. 应用层改造

### 2.1 健康检查端点（所有微服务）

每个 gRPC 微服务新增 HTTP 健康检查端口（8080），供 K8s 探针使用：

```go
// internal/shared/health/health.go
package health

import (
    "net/http"
    "go.uber.org/zap"
)

// StartHealthServer 启动健康检查 HTTP 服务
// 端口 8080，提供 /healthz 和 /readyz
func StartHealthServer(logger *zap.Logger, checks ...CheckFunc) {
    mux := http.NewServeMux()

    // /healthz — 存活探针（服务进程是否运行）
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ok"))
    })

    // /readyz — 就绪探针（依赖服务是否可用）
    mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
        for _, check := range checks {
            if err := check(); err != nil {
                w.WriteHeader(http.StatusServiceUnavailable)
                w.Write([]byte(err.Error()))
                return
            }
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ok"))
    })

    go func() {
        logger.Info("健康检查服务启动", zap.Int("port", 8080))
        if err := http.ListenAndServe(":8080", mux); err != nil {
            logger.Error("健康检查服务失败", zap.Error(err))
        }
    }()
}

// CheckFunc 健康检查函数类型
type CheckFunc func() error

// DBCheck 数据库连接检查
func DBCheck(db interface{ DB() error }) CheckFunc {
    return func() error {
        sqlDB, err := db.DB()
        if err != nil {
            return err
        }
        return sqlDB.Ping()
    }
}

// RedisCheck Redis 连接检查
func RedisCheck(client interface{ Ping() error }) CheckFunc {
    return func() error {
        return client.Ping()
    }
}
```

各微服务 main.go 中启动：

```go
// cmd/user/main.go
health.StartHealthServer(log,
    health.DBCheck(db),
    // 无 Redis 依赖则省略
)
```

### 2.2 服务发现：K8s DNS 替代硬编码地址

在 K8s 中，服务间调用使用 K8s DNS 名称：

```
# 命名空间内：{service-name}:{port}
user-service:50001
auth-service:50002
permission-service:50003
biz-service:50004

# 跨命名空间：{service-name}.{namespace}:{port}
user-service.hellogo.svc.cluster.local:50001
```

**无需修改代码**，通过 Helm values.yaml 配置环境变量即可：

```yaml
# values.yaml
services:
  user:
    addr: "user-service:50001"
  auth:
    addr: "auth-service:50002"
  permission:
    addr: "permission-service:50003"
  biz:
    addr: "biz-service:50004"
```

通过 ConfigMap 注入为环境变量，服务启动时自动读取。

### 2.3 优雅停机

所有微服务 main.go 添加信号监听：

```go
// 优雅停机
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

go func() {
    <-quit
    log.Info("收到停机信号，开始优雅停机...")

    // 1. 停止接收新请求（gRPC GracefulStop）
    grpcServer.GracefulStop()

    // 2. 关闭数据库连接
    sqlDB, _ := db.DB()
    sqlDB.Close()

    // 3. 关闭 Redis
    redisClient.Close()

    log.Info("服务已安全停止")
    os.Exit(0)
}()
```

K8s 会在发送 SIGTERM 后等待 `terminationGracePeriodSeconds`（默认 30s），超时后强制 kill。

### 2.4 配置外部化

**数据库密码、JWT Secret 等敏感信息**通过 K8s Secret 管理：

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: hellogo-secrets
type: Opaque
stringData:
  db-password: "root123456"
  jwt-secret: "k8s-jwt-secret-change-in-prod"
  redis-password: ""
```

**普通配置**通过 ConfigMap 管理：

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: hellogo-config
data:
  APP_ENV: "production"
  DB_TYPE: "mysql"
  DB_HOST: "mysql"
  DB_PORT: "3306"
  DB_NAME: "hellogo"
  DB_USER: "root"
  REDIS_HOST: "redis"
  REDIS_PORT: "6379"
```

---

## 3. Dockerfile

### 3.1 后端微服务（统一 Dockerfile）

```dockerfile
# deploy/docker/Dockerfile.service
# 多阶段构建：编译阶段 + 运行阶段

# ── 阶段一：编译 ──────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

ARG SERVICE_NAME=user

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# 先复制 go.mod/go.sum，利用 Docker 缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/server ./cmd/${SERVICE_NAME}/main.go

# ── 阶段二：运行 ──────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && adduser -D -u 1000 appuser

WORKDIR /app

# 从编译阶段复制二进制
COPY --from=builder /app/server .

# 复制配置文件（configs/ 目录）
COPY configs/ ./configs/

# 非 root 用户运行
USER appuser

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/healthz || exit 1

EXPOSE 50001 8080

ENTRYPOINT ["./server"]
```

### 3.2 前端（nginx 托管静态文件）

```dockerfile
# deploy/docker/Dockerfile.frontend

# ── 阶段一：构建 ──────────────────────────────────────────────
FROM node:20-alpine AS builder

WORKDIR /app

COPY front-end/package.json front-end/package-lock.json ./
RUN npm ci

COPY front-end/ .

# 构建时注入 API 地址（Vite 使用 import.meta.env）
ARG VITE_API_URL=http://localhost:30080
ENV VITE_API_URL=${VITE_API_URL}

RUN npm run build

# ── 阶段二：nginx ─────────────────────────────────────────────
FROM nginx:1.27-alpine

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# nginx 配置（支持 SPA 路由）
COPY deploy/docker/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
```

nginx 配置：

```nginx
# deploy/docker/nginx.conf
server {
    listen 80;
    server_name _;
    root /usr/share/nginx/html;
    index index.html;

    # SPA 路由 fallback
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

### 3.3 构建脚本

```bash
#!/bin/bash
# deploy/docker/build.sh

eval $(minikube docker-env)

# 后端服务
for svc in user auth permission biz gateway; do
    echo "构建 ${svc} 镜像..."
    docker build \
        --build-arg SERVICE_NAME=${svc} \
        -f deploy/docker/Dockerfile.service \
        -t hellogo/${svc}:latest \
        .
done

# 前端
echo "构建 frontend 镜像..."
docker build \
    --build-arg VITE_API_URL=http://$(minikube ip):30080 \
    -f deploy/docker/Dockerfile.frontend \
    -t hellogo/frontend:latest \
    .

echo "构建完成！"
docker images | grep hellogo
```

---

## 4. Helm Chart

### 4.1 Chart 目录结构

```
deploy/helm/hellogo/
├── Chart.yaml
├── values.yaml
├── templates/
│   ├── _helpers.tpl
│   ├── namespace.yaml
│   ├── secrets.yaml
│   ├── configmap.yaml
│   ├── mysql/
│   │   ├── statefulset.yaml
│   │   ├── service.yaml
│   │   └── pvc.yaml
│   ├── redis/
│   │   ├── statefulset.yaml
│   │   └── service.yaml
│   ├── user/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   ├── auth/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   ├── permission/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   ├── biz/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   ├── gateway/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   └── frontend/
│       ├── deployment.yaml
│       └── service.yaml
```

### 4.2 Chart.yaml

```yaml
apiVersion: v2
name: hellogo
description: helloGo 微服务 K8s 部署（minikube）
type: application
version: 0.1.0
appVersion: "1.0.0"
```

### 4.3 values.yaml

```yaml
# ── 全局 ─────────────────────────────────────────────────────
namespace: hellogo
imagePullPolicy: IfNotPresent

# ── 数据库 ────────────────────────────────────────────────────
mysql:
  enabled: true
  image: mysql:8.0
  rootPassword: root123456
  database: hellogo
  storage: 5Gi
  port: 3306

# ── Redis ─────────────────────────────────────────────────────
redis:
  enabled: true
  image: redis:7-alpine
  password: ""
  storage: 1Gi
  port: 6379

# ── 共享配置 ──────────────────────────────────────────────────
config:
  appEnv: production
  dbType: mysql
  dbName: hellogo
  dbUser: root
  jwtSecret: k8s-jwt-secret-change-in-prod
  jwtExpires: 1d
  jwtRefreshExpires: 7d
  loginMaxFails: 5
  loginLockTTL: 600

# ── 微服务 ────────────────────────────────────────────────────
services:
  user:
    image: hellogo/user
    tag: latest
    replicas: 2
    grpcPort: 50001
    healthPort: 8080
    resources:
      requests: { cpu: 100m, memory: 128Mi }
      limits:   { cpu: 500m, memory: 256Mi }

  auth:
    image: hellogo/auth
    tag: latest
    replicas: 2
    grpcPort: 50002
    healthPort: 8080
    resources:
      requests: { cpu: 100m, memory: 128Mi }
      limits:   { cpu: 500m, memory: 256Mi }

  permission:
    image: hellogo/permission
    tag: latest
    replicas: 2
    grpcPort: 50003
    healthPort: 8080
    resources:
      requests: { cpu: 100m, memory: 128Mi }
      limits:   { cpu: 500m, memory: 256Mi }

  biz:
    image: hellogo/biz
    tag: latest
    replicas: 2
    grpcPort: 50004
    healthPort: 8080
    resources:
      requests: { cpu: 100m, memory: 128Mi }
      limits:   { cpu: 500m, memory: 256Mi }

  gateway:
    image: hellogo/gateway
    tag: latest
    replicas: 2
    httpPort: 8000
    healthPort: 8080
    nodePort: 30080
    corsOrigins: "*"
    resources:
      requests: { cpu: 200m, memory: 256Mi }
      limits:   { cpu: 1000m, memory: 512Mi }

# ── 前端 ──────────────────────────────────────────────────────
frontend:
  image: hellogo/frontend
  tag: latest
  replicas: 1
  port: 80
  nodePort: 30090
  resources:
    requests: { cpu: 50m, memory: 64Mi }
    limits:   { cpu: 200m, memory: 128Mi }
```

### 4.4 模板示例

**通用 Deployment 模板（以 user 服务为例）：**

```yaml
# deploy/helm/hellogo/templates/user/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
  namespace: {{ .Values.namespace }}
  labels:
    app: user-service
spec:
  replicas: {{ .Values.services.user.replicas }}
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
    spec:
      terminationGracePeriodSeconds: 30
      containers:
        - name: user-service
          image: "{{ .Values.services.user.image }}:{{ .Values.services.user.tag }}"
          imagePullPolicy: {{ .Values.imagePullPolicy }}
          ports:
            - name: grpc
              containerPort: {{ .Values.services.user.grpcPort }}
            - name: health
              containerPort: {{ .Values.services.user.healthPort }}
          env:
            - name: APP_ENV
              valueFrom:
                configMapKeyRef:
                  name: hellogo-config
                  key: APP_ENV
            - name: DB_TYPE
              valueFrom:
                configMapKeyRef:
                  name: hellogo-config
                  key: DB_TYPE
            - name: DB_HOST
              value: "mysql"
            - name: DB_PORT
              value: "3306"
            - name: DB_NAME
              valueFrom:
                configMapKeyRef:
                  name: hellogo-config
                  key: DB_NAME
            - name: DB_USER
              valueFrom:
                configMapKeyRef:
                  name: hellogo-config
                  key: DB_USER
            - name: DB_PASS
              valueFrom:
                secretKeyRef:
                  name: hellogo-secrets
                  key: db-password
            - name: REDIS_HOST
              value: "redis"
            - name: REDIS_PORT
              value: "6379"
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: hellogo-secrets
                  key: jwt-secret
            - name: GRPC_PORT
              value: "{{ .Values.services.user.grpcPort }}"
            - name: USER_SERVICE_ADDR
              value: "user-service:50001"
            - name: AUTH_SERVICE_ADDR
              value: "auth-service:50002"
            - name: PERMISSION_SERVICE_ADDR
              value: "permission-service:50003"
            - name: BIZ_SERVICE_ADDR
              value: "biz-service:50004"
          # 存活探针
          livenessProbe:
            httpGet:
              path: /healthz
              port: health
            initialDelaySeconds: 15
            periodSeconds: 20
            timeoutSeconds: 3
            failureThreshold: 3
          # 就绪探针
          readinessProbe:
            httpGet:
              path: /readyz
              port: health
            initialDelaySeconds: 5
            periodSeconds: 10
            timeoutSeconds: 3
            failureThreshold: 3
          # 启动探针（给服务足够的初始化时间）
          startupProbe:
            httpGet:
              path: /healthz
              port: health
            initialDelaySeconds: 5
            periodSeconds: 5
            failureThreshold: 10
          resources:
            requests:
              cpu: {{ .Values.services.user.resources.requests.cpu }}
              memory: {{ .Values.services.user.resources.requests.memory }}
            limits:
              cpu: {{ .Values.services.user.resources.limits.cpu }}
              memory: {{ .Values.services.user.resources.limits.memory }}
```

**Gateway Service（NodePort 暴露）：**

```yaml
# deploy/helm/hellogo/templates/gateway/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: gateway
  namespace: {{ .Values.namespace }}
spec:
  type: NodePort
  selector:
    app: gateway
  ports:
    - name: http
      port: {{ .Values.services.gateway.httpPort }}
      targetPort: http
      nodePort: {{ .Values.services.gateway.nodePort }}
```

**Frontend Service（NodePort 暴露）：**

```yaml
# deploy/helm/hellogo/templates/frontend/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: frontend
  namespace: {{ .Values.namespace }}
spec:
  type: NodePort
  selector:
    app: frontend
  ports:
    - name: http
      port: {{ .Values.frontend.port }}
      targetPort: http
      nodePort: {{ .Values.frontend.nodePort }}
```

**MySQL StatefulSet：**

```yaml
# deploy/helm/hellogo/templates/mysql/statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mysql
  namespace: {{ .Values.namespace }}
spec:
  serviceName: mysql
  replicas: 1
  selector:
    matchLabels:
      app: mysql
  template:
    metadata:
      labels:
        app: mysql
    spec:
      containers:
        - name: mysql
          image: {{ .Values.mysql.image }}
          ports:
            - containerPort: {{ .Values.mysql.port }}
          env:
            - name: MYSQL_ROOT_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: hellogo-secrets
                  key: db-password
            - name: MYSQL_DATABASE
              value: {{ .Values.mysql.database }}
          volumeMounts:
            - name: mysql-data
              mountPath: /var/lib/mysql
          resources:
            requests: { cpu: 200m, memory: 512Mi }
            limits:   { cpu: 1000m, memory: 1Gi }
  volumeClaimTemplates:
    - metadata:
        name: mysql-data
      spec:
        accessModes: [ReadWriteOnce]
        resources:
          requests:
            storage: {{ .Values.mysql.storage }}
```

**Redis StatefulSet：**

```yaml
# deploy/helm/hellogo/templates/redis/statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis
  namespace: {{ .Values.namespace }}
spec:
  serviceName: redis
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
        - name: redis
          image: {{ .Values.redis.image }}
          ports:
            - containerPort: {{ .Values.redis.port }}
          volumeMounts:
            - name: redis-data
              mountPath: /data
          resources:
            requests: { cpu: 100m, memory: 128Mi }
            limits:   { cpu: 500m, memory: 256Mi }
  volumeClaimTemplates:
    - metadata:
        name: redis-data
      spec:
        accessModes: [ReadWriteOnce]
        resources:
          requests:
            storage: {{ .Values.redis.storage }}
```

---

## 5. 部署与验证

### 5.1 构建镜像

```bash
# 切换到 minikube Docker 环境
eval $(minikube docker-env)

# 构建所有镜像
bash deploy/docker/build.sh

# 验证镜像
docker images | grep hellogo
# 应看到：hellogo/user, hellogo/auth, hellogo/permission, hellogo/biz, hellogo/gateway, hellogo/frontend
```

### 5.2 Helm 安装

```bash
# 安装（首次部署）
helm install hellogo deploy/helm/hellogo/ \
  --namespace hellogo \
  --create-namespace

# 查看部署状态
helm status hellogo -n hellogo
kubectl get pods -n hellogo
kubectl get svc -n hellogo
```

### 5.3 等待服务就绪

```bash
# 等待所有 Pod Running
kubectl wait --for=condition=Ready pod -l app -n hellogo --timeout=120s

# 查看 Pod 状态
kubectl get pods -n hellogo -o wide

# 预期输出：
# NAME                  READY   STATUS    RESTARTS   AGE
# mysql-0               1/1     Running   0          60s
# redis-0               1/1     Running   0          60s
# user-service-xxx      1/1     Running   0          60s
# auth-service-xxx      1/1     Running   0          60s
# permission-service-xx 1/1     Running   0          60s
# biz-service-xxx       1/1     Running   0          60s
# gateway-xxx           1/1     Running   0          60s
# frontend-xxx          1/1     Running   0          60s
```

### 5.4 数据库初始化

```bash
# 执行数据库迁移（进入 user-service Pod）
kubectl exec -it deploy/user-service -n hellogo -- ./server migrate

# 或者在 MySQL Pod 中手动执行 SQL
kubectl exec -it mysql-0 -n hellogo -- mysql -uroot -proot123456 hellogo
```

### 5.5 验证访问

```bash
# 获取 minikube IP
MINIKUBE_IP=$(minikube ip)

# 测试 Gateway
curl http://${MINIKUBE_IP}:30080/api/health
# {"code":"SUCCESS","data":{"service":"gateway","status":"ok"}}

# 测试登录
curl -X POST http://${MINIKUBE_IP}:30080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 打开前端
echo "前端地址: http://${MINIKUBE_IP}:30090"
# 在浏览器中打开
```

---

## 6. Lens 可视化管理

### 6.1 连接集群

1. 打开 Lens
2. 左侧 "Catalog" → "Clusters" → 选择 "minikube"
3. 点击 "Connect"
4. 在命名空间筛选器中选择 "hellogo"

### 6.2 常用操作

| 操作 | Lens 界面 |
|------|----------|
| 查看 Pod 列表 | Workloads → Pods → 筛选 namespace: hellogo |
| 查看 Pod 日志 | 点击 Pod → Logs 标签 |
| 进入容器 Shell | 点击 Pod → Shell 标签 |
| 查看 Service | Network → Services |
| 查看 ConfigMap | Config → ConfigMaps |
| 查看 Secret | Config → Secrets |
| 扩缩容 | 点击 Deployment → 修改 Replicas |
| 重启 Pod | 点击 Pod → Delete（Deployment 会自动重建）|
| 查看事件 | 点击 Pod → Events 标签 |

### 6.3 资源监控

Lens 内置了资源使用情况展示：
- **Cluster Overview**：CPU / Memory 使用率
- **Pod Metrics**：每个 Pod 的实时资源消耗
- **HPA**（如配置）：自动扩缩容状态

---

## 7. 日常运维操作

### 7.1 查看日志

```bash
# 查看某个服务的日志（实时）
kubectl logs -f deploy/user-service -n hellogo

# 查看最近 100 行
kubectl logs --tail=100 deploy/gateway -n hellogo

# 查看所有服务的日志
kubectl logs -l app -n hellogo --tail=50
```

### 7.2 进入容器调试

```bash
# 进入 user-service 容器
kubectl exec -it deploy/user-service -n hellogo -- sh

# 在容器内测试 gRPC 连接
wget -qO- http://localhost:8080/healthz

# 测试 MySQL 连接（从任意 Pod）
kubectl run debug --rm -it --image=busybox -n hellogo -- sh
# 在 debug Pod 中：
# wget -qO- http://mysql:3306
# nslookup user-service
```

### 7.3 更新镜像

修改代码后，有两种更新方式：

**方式一：快速更新（开发阶段，latest 标签）**

只重建改动的那个服务，然后重启 Pod：

```bash
# 一行命令：构建镜像 + 重启 Pod
make k8s-deploy SVC=user

# 等价于：
make k8s-build-one SVC=user   # 构建 hellogo/user:latest
make k8s-restart SVC=user     # rollout restart 触发重新拉取
```

> 因为 `imagePullPolicy: IfNotPresent` + 标签固定为 `latest`，K8s 不会自动拉取新镜像，必须 `rollout restart`。

**方式二：版本化更新（生产阶段，推荐）**

```bash
# 1. 构建带版本号的镜像
make k8s-build-one SVC=user TAG=v1.1.0

# 2. Helm upgrade 指定新版本（自动滚动更新，不需要 rollout restart）
make k8s-upgrade SVC=user TAG=v1.1.0

# 3. 查看更新状态
kubectl rollout status deploy/user-service -n hellogo
```

> 优点：可回滚（`helm rollback`）、可审计（每个版本有明确标签）。

**前端更新：**

```bash
# 快速更新
make k8s-build-frontend
make k8s-restart SVC=frontend

# 版本化更新
make k8s-build-frontend TAG=v1.1.0
make k8s-upgrade TAG=v1.1.0  # 注意：前端不区分 SVC
```

**全部重建（大版本更新）：**

```bash
make k8s-build       # 重建全部镜像
make k8s-upgrade     # Helm upgrade（tag 不变时需要 rollout restart）
```

### 7.4 回滚

```bash
# 回滚到上一版本
helm rollback hellogo -n hellogo

# 回滚到指定版本
helm rollback hellogo 1 -n hellogo

# 查看回滚历史
helm history hellogo -n hellogo
```

### 7.5 扩缩容

```bash
# 临时扩容 user-service
kubectl scale deploy/user-service --replicas=5 -n hellogo

# 通过 Helm values 永久修改
helm upgrade hellogo deploy/helm/hellogo/ \
  --namespace hellogo \
  --set services.user.replicas=3
```

### 7.6 完全卸载

```bash
# 卸载 Helm release
helm uninstall hellogo -n hellogo

# 删除命名空间（包括 PVC 数据）
kubectl delete namespace hellogo

# 停止 minikube
minikube stop

# 删除 minikube 集群（完全重置）
minikube delete
```

---

## 8. Makefile 目标

| 目标 | 用法 | 说明 |
|------|------|------|
| `k8s-build` | `make k8s-build` | 构建所有镜像（5 后端 + 前端） |
| `k8s-build-one` | `make k8s-build-one SVC=user` | 构建单个服务镜像（latest 标签） |
| `k8s-build-one` | `make k8s-build-one SVC=user TAG=v1.1.0` | 构建带版本号的镜像 |
| `k8s-build-frontend` | `make k8s-build-frontend` | 构建前端镜像 |
| `k8s-deploy` | `make k8s-deploy SVC=user` | **快速部署：构建 + 重启 Pod** |
| `k8s-install` | `make k8s-install` | 首次 Helm 安装 |
| `k8s-upgrade` | `make k8s-upgrade` | Helm 升级（全量） |
| `k8s-upgrade` | `make k8s-upgrade SVC=user TAG=v1.1.0` | Helm 升级单个服务版本 |
| `k8s-uninstall` | `make k8s-uninstall` | 卸载 Helm release |
| `k8s-status` | `make k8s-status` | 查看 Pods + Services + Helm releases |
| `k8s-logs` | `make k8s-logs SVC=user` | 实时查看日志 |
| `k8s-shell` | `make k8s-shell SVC=user` | 进入容器 Shell |
| `k8s-urls` | `make k8s-urls` | 显示前端和 Gateway 访问地址 |
| `k8s-restart` | `make k8s-restart SVC=user` | 滚动重启指定服务 |
| `k8s-rollback` | `make k8s-rollback` | 回滚 Helm release 到上一版本 |

**常用工作流：**

```bash
# 开发阶段：改完代码快速更新某个服务
make k8s-deploy SVC=user

# 生产阶段：版本化发布
make k8s-build-one SVC=user TAG=v1.1.0
make k8s-upgrade SVC=user TAG=v1.1.0

# 出问题了回滚
make k8s-rollback
```
```

---

## 9. 常见问题排查

### 9.1 Pod 启动失败

```bash
# 查看 Pod 事件（最常见排查手段）
kubectl describe pod <pod-name> -n hellogo

# 常见原因：
# - ImagePullBackOff：镜像不存在 → 检查 eval $(minikube docker-env) 是否执行
# - CrashLoopBackOff：服务启动崩溃 → 查看日志 kubectl logs <pod-name> -n hellogo
# - Pending：资源不足 → 检查 minikube 资源：minikube status
```

### 9.2 服务间无法通信

```bash
# 测试 DNS 解析
kubectl run debug --rm -it --image=busybox -n hellogo -- nslookup user-service

# 测试端口连通性
kubectl run debug --rm -it --image=busybox -n hellogo -- \
  wget -qO- http://user-service:50001

# 检查 Service 是否创建
kubectl get svc -n hellogo
kubectl get endpoints -n hellogo
```

### 9.3 数据库连接失败

```bash
# 检查 MySQL Pod 状态
kubectl get pod mysql-0 -n hellogo
kubectl logs mysql-0 -n hellogo

# 手动测试连接
kubectl exec -it mysql-0 -n hellogo -- mysql -uroot -proot123456 -e "SHOW DATABASES"

# 检查 Secret 是否正确
kubectl get secret hellogo-secrets -n hellogo -o jsonpath='{.data.db-password}' | base64 -d
```

### 9.4 NodePort 无法访问

```bash
# 检查 minikube 是否运行
minikube status

# 获取正确的 minikube IP
minikube ip

# 检查 Service 类型
kubectl get svc -n hellogo

# 使用 minikube service 命令直接打开
minikube service gateway -n hellogo
minikube service frontend -n hellogo
```

### 9.5 Helm 安装失败

```bash
# 查看 Helm release 状态
helm status hellogo -n hellogo

# 查看渲染后的 YAML（调试模板）
helm template hellogo deploy/helm/hellogo/ --namespace hellogo

# 强制卸载（卡住时）
helm uninstall hellogo -n hellogo --no-hooks
kubectl delete namespace hellogo --force --grace-period=0
```

### 9.6 前端无法连接 Gateway

```bash
# 检查前端环境变量是否正确注入
kubectl exec -it deploy/frontend -n hellogo -- cat /usr/share/nginx/html/index.html | grep API

# 检查 CORS 配置
kubectl get configmap -n hellogo hellogo-config -o yaml | grep CORS

# 从集群外测试 Gateway
curl http://$(minikube ip):30080/api/health
```

---

## 附录：文件清单

| 文件路径 | 说明 |
|---------|------|
| `deploy/docker/Dockerfile.service` | 后端微服务 Dockerfile |
| `deploy/docker/Dockerfile.frontend` | 前端 Dockerfile |
| `deploy/docker/nginx.conf` | 前端 nginx 配置 |
| `deploy/docker/build.sh` | 镜像构建脚本 |
| `deploy/helm/hellogo/Chart.yaml` | Helm Chart 元信息 |
| `deploy/helm/hellogo/values.yaml` | Helm 配置值 |
| `deploy/helm/hellogo/templates/_helpers.tpl` | Helm 模板辅助函数 |
| `deploy/helm/hellogo/templates/namespace.yaml` | 命名空间 |
| `deploy/helm/hellogo/templates/secrets.yaml` | 敏感配置 Secret |
| `deploy/helm/hellogo/templates/configmap.yaml` | 普通配置 ConfigMap |
| `deploy/helm/hellogo/templates/mysql/*.yaml` | MySQL 资源 |
| `deploy/helm/hellogo/templates/redis/*.yaml` | Redis 资源 |
| `deploy/helm/hellogo/templates/user/*.yaml` | User Service 资源 |
| `deploy/helm/hellogo/templates/auth/*.yaml` | Auth Service 资源 |
| `deploy/helm/hellogo/templates/permission/*.yaml` | Permission Service 资源 |
| `deploy/helm/hellogo/templates/biz/*.yaml` | Biz Service 资源 |
| `deploy/helm/hellogo/templates/gateway/*.yaml` | Gateway 资源 |
| `deploy/helm/hellogo/templates/frontend/*.yaml` | Frontend 资源 |
| `internal/shared/health/health.go` | 健康检查共享包 |
