#!/bin/bash
# 重启后恢复 minikube 端口转发（iptables DNAT 方式）
# 用法: ./scripts/k8s-forward.sh [宿主机IP]
# 默认宿主机 IP: 10.0.0.100
# 注: iptables 操作会自动 sudo，minikube 以当前用户运行

set -e

HOST_IP="${1:-10.0.0.100}"
PORTS=(30080 30090)

echo "==> 检查 minikube 状态..."
if ! minikube status &>/dev/null; then
    echo "==> 启动 minikube..."
    minikube start
fi

MINIKUBE_IP=$(minikube ip)
echo "==> minikube IP: $MINIKUBE_IP"
echo "==> 宿主机 IP: $HOST_IP"
echo "==> 转发端口: ${PORTS[*]}"

echo ""
echo "==> 清理旧规则..."
for PORT in "${PORTS[@]}"; do
    # 删除 PREROUTING 规则
    while sudo iptables -t nat -D PREROUTING -d "$HOST_IP" -p tcp --dport "$PORT" -j DNAT --to-destination "$(sudo iptables -t nat -L PREROUTING -n | grep "$HOST_IP.*dpt:$PORT" | awk '{print $NF}' | sed 's/to://')" 2>/dev/null; do :; done

    # 删除 OUTPUT 规则
    while sudo iptables -t nat -D OUTPUT -d "$HOST_IP" -p tcp --dport "$PORT" -j DNAT --to-destination "$(sudo iptables -t nat -L OUTPUT -n | grep "$HOST_IP.*dpt:$PORT" | awk '{print $NF}' | sed 's/to://')" 2>/dev/null; do :; done

    # 删除 FORWARD 规则
    while sudo iptables -D FORWARD -d "$MINIKUBE_IP" -p tcp --dport "$PORT" -j ACCEPT 2>/dev/null; do :; done
done

# 删除旧的 MASQUERADE 和反向 FORWARD
while sudo iptables -t nat -D POSTROUTING -d "$MINIKUBE_IP" -p tcp -j MASQUERADE 2>/dev/null; do :; done
while sudo iptables -D FORWARD -s "$MINIKUBE_IP" -j ACCEPT 2>/dev/null; do :; done

echo "==> 添加新规则..."
for PORT in "${PORTS[@]}"; do
    # PREROUTING — 外部流量转发
    sudo iptables -t nat -A PREROUTING -d "$HOST_IP" -p tcp --dport "$PORT" -j DNAT --to-destination "$MINIKUBE_IP:$PORT"

    # OUTPUT — 本机访问宿主机 IP
    sudo iptables -t nat -A OUTPUT -d "$HOST_IP" -p tcp --dport "$PORT" -j DNAT --to-destination "$MINIKUBE_IP:$PORT"

    # FORWARD — 允许转发到 minikube（插入到最前面）
    sudo iptables -I FORWARD 1 -d "$MINIKUBE_IP" -p tcp --dport "$PORT" -j ACCEPT
done

# POSTROUTING — 源地址伪装
sudo iptables -t nat -A POSTROUTING -d "$MINIKUBE_IP" -p tcp -j MASQUERADE

# FORWARD — 允许 minikube 回程
sudo iptables -I FORWARD 1 -s "$MINIKUBE_IP" -j ACCEPT

echo "==> 保存规则..."
sudo netfilter-persistent save

echo ""
echo "✓ 完成！"
echo "  前端:    http://$HOST_IP:30090"
echo "  Gateway: http://$HOST_IP:30080"
echo ""
echo "==> 验证 Pod 状态..."
kubectl get pods -n hellogo
