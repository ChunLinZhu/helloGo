# 部署指南：minikube + Helm

## 前置条件

| 工具 | 安装验证 |
|------|---------|
| minikube | `minikube version` |
| kubectl | `kubectl version --client` |
| helm | `helm version` |
| docker | `docker version` |

---

## 一、首次部署

### 1. 启动 minikube

```bash
minikube start \
  --driver=docker \
  --cpus=4 \
  --memory=8192 \
  --disk-size=40g
```

验证：

```bash◊
minikube status
kubectl get nodes
# 应看到 minikube 节点 Ready
```

### 2. 构建镜像

```bash
make k8s-build
```

构建内容：
- 后端：`hellogo/user:latest`、`hellogo/auth:latest`、`hellogo/permission:latest`、`hellogo/biz:latest`、`hellogo/gateway:latest`
- 前端：`hellogo/frontend:latest`（自动注入 `VITE_API_URL=http://$(minikube ip):30080`）

验证：

```bash
docker images | grep hellogo
```

### 3. Helm 安装

```bash
make k8s-install
```

自动执行：
1. 创建 `hellogo` 命名空间
2. 部署 MySQL + Redis（StatefulSet）
3. 执行 `db-init-job`（创建 `hellogo` 数据库）
4. 部署 5 个微服务（各自等待 MySQL/Redis 就绪后启动）
5. 部署 Gateway（NodePort 30080）+ Frontend（NodePort 30090）
6. 各服务启动时自动建表（AutoMigrate）
7. 执行 `seed-job`（创建 admin 用户 + 测试数据）

### 4. 等待就绪

```bash
# 查看 Pod 状态，等所有 Pod 变为 Running
watch kubectl get pods -n hellogo
```

预期结果：

```
NAME                  READY   STATUS    RESTARTS   AGE
mysql-0               1/1     Running   0          60s
redis-0               1/1     Running   0          60s
user-service-xxx      1/1     Running   0          45s
auth-service-xxx      1/1     Running   0          45s
permission-service-xx 1/1     Running   0          45s
biz-service-xxx       1/1     Running   0          45s
gateway-xxx           1/1     Running   0          45s
frontend-xxx          1/1     Running   0          45s
```

### 5. 验证访问

```bash
# 查看访问地址
make k8s-urls
# 输出：
# 前端:    http://192.168.49.2:30090
# Gateway: http://192.168.49.2:30080

# 测试 API
curl http://$(minikube ip):30080/api/health
# {"code":"SUCCESS","data":{"service":"gateway","status":"ok"}}
```

浏览器打开前端地址即可使用。

### 6. 种子数据

**首次部署时自动执行：**
- `seed-job` 作为 Helm post-install hook 自动运行
- 等待 user-service、permission-service、biz-service 健康检查通过
- 创建 admin 用户（`admin` / `admin123`）
- 创建 10 个角色 + 20 个权限 + 100 个测试用户 + 菜单/部门/字典/日志

**⚠️ 注意：`helm upgrade` 不会重新执行 seed-job**

这是设计意图——避免覆盖用户手动修改的数据。如需重新播种：

```bash
make k8s-seed
```

查看 seed 日志：

```bash
kubectl logs -f job/hellogo-seed-manual -n hellogo
```

---

## 二、代码更新后重新部署

### 核心概念

K8s 更新有两种场景，本质区别在于**是否需要修改 Helm 配置**：

| 场景 | 变更内容 | 使用命令 | 原理 |
|------|---------|---------|------|
| **只改了代码** | 镜像内容变了，`tag: latest` 没变 | `make k8s-deploy SVC=xxx` | 构建新镜像 + `kubectl rollout restart` 强制重建 Pod |
| **改了配置** | `values.yaml` 里的副本数、CPU、端口等 | `helm upgrade` 或 `make k8s-upgrade` | Helm 对比 YAML 差异，触发滚动更新 |
| **代码 + 配置都改了** | 两者兼有 | 先 build 再 upgrade | 两者结合 |

---

### 方式 A：快速更新（开发阶段推荐）

适用于：**只修改了代码，没有改 `values.yaml`**

```bash
# 后端服务（以 user 为例）— 构建镜像 + 重启 Pod，一步完成
make k8s-deploy SVC=user

# 前端
make k8s-build-frontend && make k8s-restart SVC=frontend
```

**原理：**
1. `make k8s-build-one SVC=user` — 构建新镜像，覆盖 `hellogo/user:latest` 标签
2. `make k8s-restart SVC=user` — 执行 `kubectl rollout restart`，在 Pod 模板里注入时间戳，强制 K8s 重建 Pod 并拉取最新镜像

> **为什么不能只用 `helm upgrade`？**
> 因为 `tag` 还是 `latest`，Helm 对比 YAML 后发现没有任何变化，不会触发更新。`rollout restart` 通过修改 Pod 模板的注解来"欺骗" Deployment 控制器重新拉取镜像。

可用服务名：`user`、`auth`、`permission`、`biz`、`gateway`、`frontend`

---

### 方式 B：版本化更新（正式发布）

适用于：**打版本号 tag，需要可追溯、可回滚**

```bash
# 1. 构建带版本号的镜像
make k8s-build-one SVC=user TAG=v1.1.0

# 2. Helm 升级（自动滚动更新）
make k8s-upgrade SVC=user TAG=v1.1.0
```

**原理：**
- `TAG=v1.1.0` 改变了 Deployment 的镜像引用（`hellogo/user:latest` → `hellogo/user:v1.1.0`）
- Helm 检测到 YAML 变化，自动触发滚动更新
- 每个版本有明确标签，支持 `make k8s-rollback` 回滚

---

### 方式 C：修改 Helm 配置后更新

适用于：**修改了 `values.yaml`（副本数、资源限制、端口、环境变量等）**

```bash
# 直接执行 helm upgrade，Helm 会对比差异并应用
helm upgrade hellogo deploy/helm/hellogo/ --namespace hellogo

# 或者通过 Makefile（等价）
make k8s-upgrade
```

常见配置变更场景：

| 修改内容 | 文件位置 | 示例 |
|---------|---------|------|
| 副本数 | `values.yaml` → `services.xxx.replicas` | 2 → 1（省资源） |
| CPU/内存限制 | `values.yaml` → `services.xxx.resources` | 500m → 1000m |
| 环境变量 | `values.yaml` → `config.xxx` | 修改 CORS 地址 |
| Service 端口 | `values.yaml` → `services.xxx.grpcPort` | 修改端口 |

---

### 方式 D：全部重建（大版本更新）

```bash
# 重建所有镜像
make k8s-build

# 重启所有服务
kubectl rollout restart deploy -n hellogo
```

---

## 三、回滚

```bash
# 回滚到上一 Helm 版本
make k8s-rollback

# 查看版本历史
helm history hellogo -n hellogo
```

> **注意：** `rollout restart` 不会产生 Helm revision，不能用 `helm rollback` 回滚。如果代码有问题，重新 build + deploy 即可。

---

## 四、常用运维命令

```bash
make k8s-status                    # 查看 Pods + Services + Helm
make k8s-logs SVC=user             # 实时查看日志
make k8s-shell SVC=user            # 进入容器 Shell
make k8s-restart SVC=user          # 重启某个服务
make k8s-urls                      # 显示访问地址
```

---

## 五、完全卸载

```bash
make k8s-uninstall                 # 卸载 Helm release
kubectl delete namespace hellogo   # 删除命名空间和数据
minikube stop                      # 停止 minikube
```

---

## 六、常见问题排查

### 1. `make k8s-build` 报错：go.mod requires go >= 1.26.3

**现象：**
```
> [builder 5/7] RUN go mod download:
0.162 go: go.mod requires go >= 1.26.3 (running go 1.23.12; GOTOOLCHAIN=local)
```

**原因：** `deploy/docker/Dockerfile.seed` 中的 Go 基础镜像版本低于 `go.mod` 要求的版本。

**修复：** 确保所有 Dockerfile 的 `FROM golang:` 版本一致：
```dockerfile
FROM golang:1.26-alpine AS builder
```

---

### 2. `make k8s-install` 报错：namespace already exists

**现象：**
```
Error: INSTALLATION FAILED: namespaces "hellogo" already exists
```

**原因：** Helm Chart 的 `namespace.yaml` 模板尝试创建命名空间，与 `--create-namespace` 参数冲突。

**修复：** 删除 `deploy/helm/hellogo/templates/namespace.yaml`，依赖 `--create-namespace` 参数（Helm 标准做法）。

**注意：** 如果命名空间处于 `Terminating` 状态，需要等待删除完成后再执行 install。

---

### 3. seed-job 超时：timed out waiting for the condition

**现象：**
```
Error: INSTALLATION FAILED: failed post-install: timed out waiting for the condition
```
seed Pod 卡在 `Init:1/4`，日志显示 `wget: error getting response`。

**原因：** 微服务的 Service 只暴露了 gRPC 端口（50001-50004），没有暴露健康检查端口（8080）。seed-job 的 init container 通过 Service DNS 访问 `user-service:8080/healthz`，自然不通。

**修复：** 在每个微服务的 `service.yaml` 中添加 health 端口：
```yaml
spec:
  ports:
    - name: grpc
      port: {{ .Values.services.user.grpcPort }}
      targetPort: grpc
    - name: health
      port: {{ .Values.services.user.healthPort }}
      targetPort: health
```

**排查命令：**
```bash
# 查看 Service 暴露的端口
kubectl get svc -n hellogo

# 从 Pod 内部测试健康检查
kubectl exec -n hellogo redis-0 -- wget -qO- http://user-service:8080/healthz
```

---

### 4. 容器日志报错：permission denied

**现象：**
```
write error: can't make directories for new logfile: mkdir logs: permission denied
```

**原因：** Dockerfile 中容器以 `appuser`（UID 1000）运行，但 `/app/logs` 目录不存在或无写权限。

**修复：** 在 Dockerfile 中创建 logs 目录并赋权：
```dockerfile
RUN mkdir -p ./data ./logs && chown appuser:appuser ./data ./logs
```

---

### 5. 宿主机 IP 无法访问（局域网）

**现象：** 本机 `curl http://10.0.0.100:30090` 正常，局域网其他电脑访问超时。

**原因分析：**

minikube 运行在 Docker 网络中（如 192.168.49.2），需要通过 iptables 转发。涉及三条链：

| 链 | 作用 | 流量来源 |
|---|------|---------|
| `PREROUTING` | DNAT 目标地址 | 外部/局域网来的流量 |
| `OUTPUT` | DNAT 目标地址 | 本机进程访问宿主机 IP |
| `FORWARD` | 允许/拒绝转发 | PREROUTING DNAT 后的包 |

**修复步骤：**

```bash
MINIKUBE_IP=$(minikube ip)

# 1. PREROUTING — 外部流量转发
sudo iptables -t nat -A PREROUTING -d 10.0.0.100 -p tcp --dport 30080 -j DNAT --to-destination $MINIKUBE_IP:30080
sudo iptables -t nat -A PREROUTING -d 10.0.0.100 -p tcp --dport 30090 -j DNAT --to-destination $MINIKUBE_IP:30090

# 2. OUTPUT — 本机访问宿主机 IP
sudo iptables -t nat -A OUTPUT -d 10.0.0.100 -p tcp --dport 30080 -j DNAT --to-destination $MINIKUBE_IP:30080
sudo iptables -t nat -A OUTPUT -d 10.0.0.100 -p tcp --dport 30090 -j DNAT --to-destination $MINIKUBE_IP:30090

# 3. POSTROUTING — 源地址伪装（回程路由）
sudo iptables -t nat -A POSTROUTING -d $MINIKUBE_IP -p tcp -j MASQUERADE

# 4. FORWARD — 允许转发到 minikube（⚠️ 必须用 -I 插入到最前面）
sudo iptables -I FORWARD 1 -d $MINIKUBE_IP -p tcp --dport 30080 -j ACCEPT
sudo iptables -I FORWARD 2 -d $MINIKUBE_IP -p tcp --dport 30090 -j ACCEPT
sudo iptables -I FORWARD 3 -s $MINIKUBE_IP -j ACCEPT

# 5. 持久化
sudo apt-get install -y iptables-persistent
sudo netfilter-persistent save
```

> **⚠️ 关键：FORWARD 规则必须用 `-I FORWARD 1`（插入到位置 1），不能用 `-A`（追加）。**
> Docker 的 `DOCKER-FORWARD` 链默认策略为 `DROP`，用 `-A` 追加的规则排在后面，包还没到就被 DROP 了。

**验证：**
```bash
# 查看 FORWARD 链（确认我们的规则在最前面）
sudo iptables -L FORWARD -n --line-numbers

# 查看 NAT 规则
sudo iptables -t nat -L PREROUTING -n --line-numbers
sudo iptables -t nat -L POSTROUTING -n --line-numbers

# 确认 IP 转发已开启
cat /proc/sys/net/ipv4/ip_forward   # 应为 1
```

---

### 6. 前端 API 请求指向 minikube IP

**现象：** 浏览器开发者工具看到请求地址为 `http://192.168.49.2:30080/api/csrf-token`。

**原因：** 前端构建时 `VITE_API_URL` 被写死为 minikube IP，外部访问时该地址不可达。

**修复：** 让 nginx 反向代理 API 请求，前端使用相对路径。

1. **nginx.conf 添加 `/api/` 反向代理：**
```nginx
location /api/ {
    proxy_pass http://gateway:8000/api/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
}
```

2. **前端 `apiUrl` 默认改为空（相对路径）：**
```ts
apiUrl: import.meta.env.VITE_API_URL || '',
```

3. **Makefile 去掉 `VITE_API_URL` 构建参数：**
```makefile
k8s-build-frontend:
	@eval $$(minikube docker-env) && \
	docker build -f deploy/docker/Dockerfile.frontend \
		-t hellogo/frontend:$(or $(TAG),latest) .
```

4. **重新构建并重启：**
```bash
make k8s-build-frontend && make k8s-restart SVC=frontend
```

**效果：**
| | 修改前 | 修改后 |
|---|---|---|
| API 请求 | `http://192.168.49.2:30080/api/xxx` | `/api/xxx`（相对路径） |
| 链路 | 浏览器 → minikube IP | 浏览器 → nginx → Gateway（集群内 DNS） |

---

### 7. Makefile 命令报错：deployment not found

**现象：**
```
make k8s-restart SVC=frontend
Error from server (NotFound): deployments.apps "frontend-service" not found
```

**原因：** Makefile 固定拼接 `$(SVC)-service`，但 `frontend` 和 `gateway` 的 Deployment 名称没有 `-service` 后缀。

**修复：** 对 `frontend`/`gateway` 做特殊处理：
```makefile
k8s-restart:
ifeq ($(SVC),frontend)
	kubectl rollout restart deploy/frontend -n hellogo
else ifeq ($(SVC),gateway)
	kubectl rollout restart deploy/gateway -n hellogo
else
	kubectl rollout restart deploy/$(SVC)-service -n hellogo
endif
```

同样的修复适用于 `k8s-logs` 和 `k8s-shell`。

---

### 8. 访问速度慢

**原因分析：**

| 因素 | 开发模式 | K8s 部署 |
|------|---------|---------|
| 前端请求 | 浏览器 → Gateway（直连） | 浏览器 → nginx Pod → Gateway Pod |
| 服务调用 | 本机 gRPC（localhost） | 跨 Pod gRPC（集群网络 + DNS） |
| CPU | 直接用宿主机 | minikube VM，多 Pod 共享 |

**优化措施：**

1. **减少副本数**（开发环境不需要高可用）：
```yaml
# values.yaml
services:
  user:
    replicas: 1   # 从 2 改为 1
```

2. **提高 CPU 限制**（默认 500m = 0.5 核，可能不够）：
```yaml
services:
  user:
    resources:
      limits:
        cpu: 1000m   # 从 500m 改为 1000m
```

3. **应用变更：**
```bash
helm upgrade hellogo deploy/helm/hellogo/ --namespace hellogo
```

---

### 9. 部署问题速查表

| 症状 | 排查命令 | 可能原因 |
|------|---------|---------|
| Pod 卡在 `Init:x/y` | `kubectl logs <pod> -c <init-container>` | 等待的服务未就绪 |
| Pod `CrashLoopBackOff` | `kubectl logs <pod> --previous` | 应用启动崩溃 |
| Service 不通 | `kubectl exec ... -- wget <svc>:<port>` | Service 缺少端口 |
| 外部访问超时 | `sudo iptables -L FORWARD -n` | FORWARD 规则被 DROP |
| `helm install` 失败 | `kubectl get ns` | 命名空间残留 |
| 前端 API 404 | 浏览器开发者工具 → Network | VITE_API_URL 配置错误 |
| 构建失败 | `docker logs` | Go 版本不匹配 |
| 响应慢 | `kubectl top pods -n hellogo` | CPU/内存限制太低 |
