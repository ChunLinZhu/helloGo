#!/bin/bash
# 构建所有 Docker 镜像（使用 minikube Docker 环境）
# 用法: bash deploy/docker/build.sh

set -e

echo "=== 切换到 minikube Docker 环境 ==="
eval $(minikube docker-env)

echo ""
echo "=== 构建后端微服务镜像 ==="
for svc in user auth permission biz gateway; do
    echo ">>> 构建 hellogo/${svc}:latest ..."
    docker build \
        --build-arg SERVICE_NAME=${svc} \
        -f deploy/docker/Dockerfile.service \
        -t hellogo/${svc}:latest \
        .
done

echo ""
echo "=== 构建前端镜像 ==="
MINIKUBE_IP=$(minikube ip)
echo ">>> 构建 hellogo/frontend:latest (API_URL=http://${MINIKUBE_IP}:30080) ..."
docker build \
    --build-arg VITE_API_URL=http://${MINIKUBE_IP}:30080 \
    -f deploy/docker/Dockerfile.frontend \
    -t hellogo/frontend:latest \
    .

echo ""
echo "=== 构建完成 ==="
docker images | grep hellogo
echo ""
echo "下一步: make k8s-install"
