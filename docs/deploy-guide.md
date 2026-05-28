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

```bash
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

### 方式 A：快速更新（开发阶段）

修改代码后，一行命令完成更新：

```bash
# 后端服务（以 user 为例）
make k8s-deploy SVC=user

# 前端
make k8s-build-frontend && make k8s-restart SVC=frontend
```

原理：构建新镜像（覆盖 `latest` 标签）→ `rollout restart` 强制 Pod 重新拉取。

可用服务名：`user`、`auth`、`permission`、`biz`、`gateway`

### 方式 B：版本化更新（正式发布）

```bash
# 1. 构建带版本号的镜像
make k8s-build-one SVC=user TAG=v1.1.0

# 2. Helm 升级（自动滚动更新）
make k8s-upgrade SVC=user TAG=v1.1.0
```

优点：每个版本有明确标签，支持 `make k8s-rollback` 回滚。

### 方式 C：全部重建（大版本更新）

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
