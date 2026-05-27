#!/usr/bin/env bash
# ============================================================
# helloGo E2E 测试脚本 — 使用 curl 测试 API 端点
# 前提：服务已通过 docker-compose 或 go run 启动在 localhost:8000
# 用法：bash scripts/test_curl.sh [BASE_URL]
# ============================================================

set -e

BASE_URL="${1:-http://localhost:8000}"
API="${BASE_URL}/api"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASS=0
FAIL=0
TOTAL=0

# ── 辅助函数 ──────────────────────────────────────────────

pass() {
    PASS=$((PASS + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "  ${GREEN}✓${NC} $1"
}

fail() {
    FAIL=$((FAIL + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "  ${RED}✗${NC} $1 (期望: $2, 实际: $3)"
}

check_status() {
    local desc="$1"
    local expected="$2"
    local actual="$3"
    if [ "$actual" -eq "$expected" ]; then
        pass "$desc"
    else
        fail "$desc" "$expected" "$actual"
    fi
}

# 发送请求并获取状态码
do_request() {
    local method="$1"
    local url="$2"
    local data="$3"
    local token="$4"

    local args=(-s -o /dev/null -w '%{http_code}' -X "$method")
    args+=(-H "Content-Type: application/json")

    if [ -n "$token" ]; then
        args+=(-H "Authorization: Bearer ${token}")
    fi

    if [ -n "$data" ]; then
        args+=(-d "$data")
    fi

    curl "${args[@]}" "$url" 2>/dev/null
}

# 发送请求并获取响应体
do_request_body() {
    local method="$1"
    local url="$2"
    local data="$3"
    local token="$4"

    local args=(-s -X "$method")
    args+=(-H "Content-Type: application/json")

    if [ -n "$token" ]; then
        args+=(-H "Authorization: Bearer ${token}")
    fi

    if [ -n "$data" ]; then
        args+=(-d "$data")
    fi

    curl "${args[@]}" "$url" 2>/dev/null
}

echo ""
echo "============================================================"
echo " helloGo E2E 测试 — ${BASE_URL}"
echo "============================================================"
echo ""

# ── 1. 健康检查 ──────────────────────────────────────────

echo -e "${YELLOW}[1] 健康检查${NC}"

STATUS=$(do_request GET "${API}/health" "" "")
check_status "GET /api/health → 200" 200 "$STATUS"

# ── 2. 认证（无 token 应拒绝） ───────────────────────────

echo ""
echo -e "${YELLOW}[2] 认证保护${NC}"

STATUS=$(do_request GET "${API}/users" "" "")
check_status "GET /api/users 无 token → 401" 401 "$STATUS"

# ── 3. 登录获取 token ────────────────────────────────────

echo ""
echo -e "${YELLOW}[3] 登录认证${NC}"

# 先尝试用 seed 数据登录（需要先运行 make seed）
LOGIN_BODY='{"username":"admin","password":"admin123"}'
LOGIN_RESP=$(do_request_body POST "${API}/auth/login" "$LOGIN_BODY" "")

# 提取 token（使用 python3 解析 JSON）
ACCESS_TOKEN=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('accessToken',''))" 2>/dev/null || echo "")

if [ -n "$ACCESS_TOKEN" ] && [ "$ACCESS_TOKEN" != "" ]; then
    pass "POST /auth/login 获取 token"

    # ── 4. 用户 CRUD ─────────────────────────────────────
    echo ""
    echo -e "${YELLOW}[4] 用户管理${NC}"

    # 查询用户列表
    STATUS=$(do_request GET "${API}/users?page=1&limit=10" "" "$ACCESS_TOKEN")
    check_status "GET /api/users → 200" 200 "$STATUS"

    # 创建用户
    CREATE_BODY='{"username":"e2e_test_user","password":"Test@123456"}'
    STATUS=$(do_request POST "${API}/users" "$CREATE_BODY" "$ACCESS_TOKEN")
    check_status "POST /api/users → 201" 201 "$STATUS"

    # 获取创建的用户 ID
    CREATE_RESP=$(do_request_body POST "${API}/users" '{"username":"e2e_test_user_2","password":"Test@123456"}' "$ACCESS_TOKEN")
    USER_ID=$(echo "$CREATE_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('id',''))" 2>/dev/null || echo "")

    if [ -n "$USER_ID" ]; then
        # 按 ID 查询
        STATUS=$(do_request GET "${API}/users/${USER_ID}" "" "$ACCESS_TOKEN")
        check_status "GET /api/users/:id → 200" 200 "$STATUS"

        # 更新用户
        UPDATE_BODY='{"email":"e2e@test.com"}'
        STATUS=$(do_request PATCH "${API}/users/${USER_ID}" "$UPDATE_BODY" "$ACCESS_TOKEN")
        check_status "PATCH /api/users/:id → 200" 200 "$STATUS"

        # 删除用户
        STATUS=$(do_request DELETE "${API}/users/${USER_ID}" "" "$ACCESS_TOKEN")
        check_status "DELETE /api/users/:id → 200" 200 "$STATUS"
    fi

    # 清理第一个测试用户
    # 获取列表找到 e2e_test_user 的 ID
    LIST_RESP=$(do_request_body GET "${API}/users?page=1&limit=100" "" "$ACCESS_TOKEN")
    E2E_ID=$(echo "$LIST_RESP" | python3 -c "
import sys,json
d=json.load(sys.stdin)
items=d.get('data',{}).get('items',[])
for item in items:
    if item.get('username')=='e2e_test_user':
        print(item.get('id',''))
        break
" 2>/dev/null || echo "")
    if [ -n "$E2E_ID" ]; then
        do_request DELETE "${API}/users/${E2E_ID}" "" "$ACCESS_TOKEN" > /dev/null 2>&1
    fi

    # ── 5. 角色管理 ──────────────────────────────────────
    echo ""
    echo -e "${YELLOW}[5] 角色管理${NC}"

    STATUS=$(do_request GET "${API}/roles?page=1&limit=10" "" "$ACCESS_TOKEN")
    check_status "GET /api/roles → 200" 200 "$STATUS"

    # ── 6. 权限管理 ──────────────────────────────────────
    echo ""
    echo -e "${YELLOW}[6] 权限管理${NC}"

    STATUS=$(do_request GET "${API}/permissions?page=1&limit=10" "" "$ACCESS_TOKEN")
    check_status "GET /api/permissions → 200" 200 "$STATUS"

    # ── 7. 菜单管理 ──────────────────────────────────────
    echo ""
    echo -e "${YELLOW}[7] 菜单管理${NC}"

    STATUS=$(do_request GET "${API}/menus/tree" "" "$ACCESS_TOKEN")
    check_status "GET /api/menus/tree → 200" 200 "$STATUS"

    # ── 8. 字典管理 ──────────────────────────────────────
    echo ""
    echo -e "${YELLOW}[8] 字典管理${NC}"

    STATUS=$(do_request GET "${API}/dicts?page=1&limit=10" "" "$ACCESS_TOKEN")
    check_status "GET /api/dicts → 200" 200 "$STATUS"

    # ── 9. 指标端点 ──────────────────────────────────────
    echo ""
    echo -e "${YELLOW}[9] 可观测性${NC}"

    STATUS=$(do_request GET "${API}/metrics" "" "")
    check_status "GET /api/metrics → 200" 200 "$STATUS"

else
    fail "POST /auth/login 获取 token" "200" "登录失败（请先运行 make seed）"
    echo ""
    echo -e "  ${YELLOW}提示: 请先运行 'make seed' 创建种子数据后再执行 E2E 测试${NC}"
fi

# ── 汇总 ──────────────────────────────────────────────────

echo ""
echo "============================================================"
echo -e " 结果: ${GREEN}${PASS} 通过${NC}, ${RED}${FAIL} 失败${NC}, 共 ${TOTAL} 项"
echo "============================================================"
echo ""

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
