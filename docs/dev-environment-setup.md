# 开发环境搭建指南

> 适用于 helloGo 项目（Go + Fiber），覆盖 Ubuntu 22.04 和 macOS (Mac Mini) 两个平台

---

## 目录

- [1. 环境总览](#1-环境总览)
- [2. Ubuntu 22.04 环境搭建](#2-ubuntu-2204-环境搭建)
- [3. macOS (Mac Mini) 环境搭建](#3-macos-mac-mini-环境搭建)
- [4. Go 项目初始化（通用）](#4-go-项目初始化通用)
- [5. IDE 配置](#5-ide-配置)
- [6. Docker 环境（通用）](#6-docker-环境通用)
- [7. 常见问题排查](#7-常见问题排查)

---

## 1. 环境总览

### 必装工具

| 工具         | 版本要求     | 用途               |
| ------------ | ------------ | ------------------ |
| Go           | >= 1.22      | 语言运行时         |
| Git          | >= 2.30      | 版本管理           |
| Make         | >= 4.0       | 构建自动化         |
| Docker       | >= 24.0      | 容器化运行         |
| Docker Compose | >= 2.20    | 多容器编排         |
| Redis        | >= 7.0       | 缓存 / 会话存储    |
| SQLite       | >= 3.35      | 默认数据库（内置） |

### 可选工具

| 工具           | 用途                      |
| -------------- | ------------------------- |
| MySQL 8        | 生产数据库                |
| PostgreSQL 15  | 生产数据库                |
| golangci-lint  | Go 代码静态检查           |
| swag           | Swagger 文档生成          |
| air            | 热重载开发                |
| curl / httpie  | API 测试                  |
| dbeaver / beekeeper | 数据库 GUI 客户端   |

### 推荐 IDE

- **VS Code** + Go 扩展（免费，跨平台）
- **GoLand**（JetBrains，付费，功能最全）

---

## 2. Ubuntu 22.04 环境搭建

### 2.1 系统更新与基础工具

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装基础构建工具
sudo apt install -y build-essential curl wget git unzip ca-certificates \
    software-properties-common apt-transport-https gnupg lsb-release
```

### 2.2 安装 Go

```bash
# 设置版本号（请根据实际最新版本调整）
GO_VERSION="1.23.4"

# 下载 Go 安装包
wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz

# 删除旧版本，安装新版本
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz

# 清理安装包
rm go${GO_VERSION}.linux-amd64.tar.gz

# 配置环境变量（写入 ~/.bashrc 或 ~/.zshrc）
cat >> ~/.bashrc << 'EOF'

# Go
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
export GO111MODULE=on
export GOPROXY=https://goproxy.cn,direct
EOF

# 使配置生效
source ~/.bashrc

# 验证安装
go version
# 预期输出: go version go1.23.4 linux/amd64
```

> **国内网络提示：** `GOPROXY=https://goproxy.cn,direct` 使用七牛云代理加速模块下载。如不需要可改为 `https://proxy.golang.org,direct`。

### 2.3 安装 Docker

```bash
# 卸载旧版本
sudo apt remove -y docker docker-engine docker.io containerd runc 2>/dev/null

# 添加 Docker 官方 GPG key
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | \
    sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

# 添加 Docker 源
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
  https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# 安装 Docker Engine + Compose
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io \
    docker-buildx-plugin docker-compose-plugin

# 将当前用户加入 docker 组（免 sudo）
sudo usermod -aG docker $USER
newgrp docker

# 验证安装
docker --version
docker compose version
docker run --rm hello-world
```

> **国内镜像加速：** 编辑 `/etc/docker/daemon.json`：
> ```json
> {
>   "registry-mirrors": [
>     "https://mirror.ccs.tencentyun.com",
>     "https://docker.mirrors.ustc.edu.cn"
>   ]
> }
> ```
> 然后 `sudo systemctl restart docker`

### 2.4 安装 Redis

**方式一：apt 安装（开发推荐）**

```bash
sudo apt install -y redis-server

# 配置 systemd 管理
sudo systemctl enable redis-server
sudo systemctl start redis-server

# 验证
redis-cli ping
# 预期输出: PONG
```

**方式二：源码编译（需要最新版本时）**

```bash
REDIS_VERSION="7.2.4"
wget https://download.redis.io/releases/redis-${REDIS_VERSION}.tar.gz
tar xzf redis-${REDIS_VERSION}.tar.gz
cd redis-${REDIS_VERSION}
make -j$(nproc)
sudo make install
cd .. && rm -rf redis-${REDIS_VERSION}*

# 创建 systemd 服务
sudo tee /etc/systemd/system/redis.service << 'EOF'
[Unit]
Description=Redis Server
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/redis-server --daemonize no
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable redis
sudo systemctl start redis

redis-cli ping
```

### 2.5 安装 MySQL 8（可选）

```bash
sudo apt install -y mysql-server

# 安全初始化
sudo mysql_secure_installation

# 配置 root 密码认证
sudo mysql -e "ALTER USER 'root'@'localhost' IDENTIFIED WITH mysql_native_password BY 'your_password';"
sudo mysql -e "FLUSH PRIVILEGES;"

# 验证
mysql -u root -p -e "SELECT VERSION();"

# 允许远程连接（仅开发环境）
sudo sed -i 's/bind-address.*=.*127.0.0.1/bind-address = 0.0.0.0/' /etc/mysql/mysql.conf.d/mysqld.cnf
sudo systemctl restart mysql
```

### 2.6 安装 PostgreSQL 15（可选）

```bash
# 添加 PostgreSQL 官方源
sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -
sudo apt update
sudo apt install -y postgresql-15

# 启动服务
sudo systemctl enable postgresql
sudo systemctl start postgresql

# 设置 postgres 用户密码
sudo -u postgres psql -c "ALTER USER postgres WITH PASSWORD 'your_password';"

# 验证
psql -U postgres -h localhost -c "SELECT version();"
```

### 2.7 安装开发辅助工具

```bash
# golangci-lint（Go 静态分析，最全面的 linter 集合）
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
    sh -s -- -b $(go env GOPATH)/bin v1.61.0

# swag（Swagger 文档生成）
go install github.com/swaggo/swag/cmd/swag@latest

# air（热重载，修改代码自动重启服务）
go install github.com/air-verse/air@latest

# goimports（自动管理 import）
go install golang.org/x/tools/cmd/goimports@latest

# mockgen（生成 mock，用于单元测试）
go install go.uber.org/mock/mockgen@latest

# 验证所有工具
golangci-lint --version
swag --version
air -v
```

### 2.8 Ubuntu 完整验证脚本

```bash
echo "========== 开发环境检查 =========="
echo ""

check() {
    if command -v "$1" &>/dev/null; then
        echo "✅ $1: $(command -v $1)"
        if [ -n "$2" ]; then
            echo "   版本: $($2)"
        fi
    else
        echo "❌ $1: 未安装"
    fi
}

check "go"         "go version"
check "git"        "git --version"
check "make"       "make --version | head -1"
check "docker"     "docker --version"
check "docker"     "docker compose version"  # compose v2
check "redis-cli"  "redis-cli --version"
check "mysql"      "mysql --version"
check "psql"       "psql --version"
check "golangci-lint" "golangci-lint --version"
check "swag"       "swag --version"
check "air"        "air -v"
check "curl"       "curl --version | head -1"

echo ""
echo "========== 服务状态 =========="
systemctl is-active redis-server &>/dev/null && echo "✅ Redis: running" || echo "❌ Redis: stopped"
systemctl is-active mysql &>/dev/null && echo "✅ MySQL: running" || echo "⬜ MySQL: stopped (可选)"
systemctl is-active postgresql &>/dev/null && echo "✅ PostgreSQL: running" || echo "⬜ PostgreSQL: stopped (可选)"
echo ""
echo "========== 环境变量 =========="
echo "GOPATH:   ${GOPATH:-未设置}"
echo "GOPROXY:  ${GOPROXY:-未设置}"
echo "GO111MODULE: ${GO111MODULE:-未设置}"
```

---

## 3. macOS (Mac Mini) 环境搭建

### 3.1 安装 Homebrew（macOS 包管理器）

```bash
# 安装 Homebrew
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Apple Silicon (M1/M2/M3/M4) 需要额外配置 PATH
# 安装脚本完成后会提示执行以下命令：
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
eval "$(/opt/homebrew/bin/brew shellenv)"

# 国内用户可替换为清华源
export HOMEBREW_API_DOMAIN="https://mirrors.tuna.tsinghua.edu.cn/homebrew-bottles/api"
export HOMEBREW_BOTTLE_DOMAIN="https://mirrors.tuna.tsinghua.edu.cn/homebrew-bottles"
export HOMEBREW_BREW_GIT_REMOTE="https://mirrors.tuna.tsinghua.edu.cn/git/homebrew/brew.git"
export HOMEBREW_CORE_GIT_REMOTE="https://mirrors.tuna.tsinghua.edu.cn/git/homebrew/homebrew-core.git"

# 更新 Homebrew
brew update
```

### 3.2 安装 Go

```bash
# 使用 Homebrew 安装最新稳定版
brew install go

# 配置环境变量（写入 ~/.zshrc）
cat >> ~/.zshrc << 'EOF'

# Go
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
export GO111MODULE=on
export GOPROXY=https://goproxy.cn,direct
EOF

source ~/.zshrc

# 验证
go version
go env GOROOT GOPATH
```

> **版本管理替代方案：** 如需管理多个 Go 版本，可用 [gvm](https://github.com/moovweb/gvm)：
> ```bash
> bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
> gvm install go1.23.4
> gvm use go1.23.4 --default
> ```

### 3.3 安装 Docker Desktop

```bash
# 方式一：Homebrew 安装（推荐）
brew install --cask docker

# 方式二：官网下载
# 访问 https://www.docker.com/products/docker-desktop/
# 下载 Apple Silicon (M系列) 或 Intel 版本
```

安装后：

1. 打开 Docker Desktop 应用（首次需授权系统权限）
2. 进入 **Settings → Resources**，建议分配：
   - CPUs: 4+
   - Memory: 4GB+
   - Disk: 30GB+
3. **Settings → Docker Engine**，添加镜像加速（国内用户）：
   ```json
   {
     "registry-mirrors": [
       "https://mirror.ccs.tencentyun.com"
     ]
   }
   ```
4. 验证：
   ```bash
   docker --version
   docker compose version
   docker run --rm hello-world
   ```

### 3.4 安装 Redis

**方式一：Homebrew 安装（开发推荐）**

```bash
brew install redis

# 设置开机自启
brew services start redis

# 验证
redis-cli ping
# 预期输出: PONG

# 管理命令
brew services list        # 查看所有服务状态
brew services stop redis  # 停止
brew services restart redis
```

**方式二：使用 Docker（隔离环境）**

```bash
docker run -d --name redis -p 6379:6379 redis:7-alpine
docker exec -it redis redis-cli ping
```

### 3.5 安装 MySQL 8（可选）

```bash
brew install mysql

# 启动服务
brew services start mysql

# 安全初始化
mysql_secure_installation

# 验证
mysql -u root -p -e "SELECT VERSION();"
```

### 3.6 安装 PostgreSQL 15（可选）

```bash
brew install postgresql@15

# 添加到 PATH（brew 会提示具体路径）
echo 'export PATH="/opt/homebrew/opt/postgresql@15/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# 启动服务
brew services start postgresql@15

# 验证
psql postgres -c "SELECT version();"
```

### 3.7 安装开发辅助工具

```bash
# golangci-lint
brew install golangci-lint

# 或者使用官方脚本（更灵活控制版本）
# curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
#     sh -s -- -b $(go env GOPATH)/bin v1.61.0

# swag（Swagger 文档生成）
go install github.com/swaggo/swag/cmd/swag@latest

# air（热重载）
go install github.com/air-verse/air@latest

# goimports
go install golang.org/x/tools/cmd/goimports@latest

# mockgen（单元测试 mock 生成）
go install go.uber.org/mock/mockgen@latest

# 其他常用工具
brew install jq          # JSON 处理
brew install httpie      # 更友好的 HTTP 客户端（替代 curl）
brew install tree        # 目录树查看

# 验证
golangci-lint --version
swag --version
air -v
```

### 3.8 macOS 完整验证脚本

```bash
echo "========== 开发环境检查 =========="
echo ""

check() {
    if command -v "$1" &>/dev/null; then
        echo "✅ $1: $(command -v $1)"
        if [ -n "$2" ]; then
            echo "   版本: $($2)"
        fi
    else
        echo "❌ $1: 未安装"
    fi
}

check "go"         "go version"
check "git"        "git --version"
check "make"       "make --version | head -1"
check "docker"     "docker --version"
check "docker"     "docker compose version"
check "redis-cli"  "redis-cli --version"
check "mysql"      "mysql --version"
check "psql"       "psql --version"
check "golangci-lint" "golangci-lint --version"
check "swag"       "swag --version"
check "air"        "air -v"
check "curl"       "curl --version | head -1"
check "jq"         "jq --version"
check "http"       "http --version"

echo ""
echo "========== 服务状态 =========="
brew services list 2>/dev/null | grep -E "redis|mysql|postgre" || echo "⬜ brew services 未运行"
echo ""
echo "========== 环境变量 =========="
echo "GOPATH:   ${GOPATH:-未设置}"
echo "GOPROXY:  ${GOPROXY:-未设置}"
echo "GO111MODULE: ${GO111MODULE:-未设置}"
echo ""
echo "========== 系统信息 =========="
echo "芯片: $(uname -m)"
echo "macOS: $(sw_vers -productVersion)"
```

---

## 4. Go 项目初始化（通用）

以下步骤在两个平台上完全一致：

### 4.1 克隆项目 & 安装依赖

```bash
# 克隆项目
git clone <repo-url> helloGo
cd helloGo

# 初始化 Go Module（如果尚未初始化）
go mod init helloGo

# 安装依赖
go mod tidy

# 验证依赖完整性
go mod verify
```

### 4.2 环境变量配置

```bash
# 复制环境配置文件
cp configs/.env.example configs/.env

# 编辑 .env，根据本地环境修改
# 关键配置项：
#   DB_TYPE=sqlite
#   SQLITE_PATH=./data/hello.db
#   REDIS_HOST=127.0.0.1
#   REDIS_PORT=6379
#   JWT_SECRET=your-dev-secret-key-change-in-production
```

### 4.3 启动项目

```bash
# 方式一：直接运行
go run cmd/server/main.go

# 方式二：热重载（开发推荐，代码修改自动重启）
air

# 方式三：Makefile
make run
```

### 4.4 初始化种子数据

```bash
# 运行种子脚本
go run cmd/seed/main.go

# 清除并重建
go run cmd/seed/main.go --purge
```

### 4.5 运行测试

```bash
# 运行所有测试
go test ./...

# 带覆盖率
go test -cover ./...

# 生成覆盖率报告（浏览器打开）
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 4.6 air 热重载配置

在项目根目录创建 `.air.toml`：

```toml
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/server ./cmd/server"
  bin = "tmp/server"
  delay = 1000
  exclude_dir = ["tmp", "vendor", "data", "upload", "docs", "node_modules"]
  exclude_regex = ["_test\\.go"]
  include_ext = ["go", "toml", "yaml", "env"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = true
  stop_on_error = true

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = true
```

---

## 5. IDE 配置

### 5.1 VS Code（推荐）

**安装扩展（两个平台一致）：**

```bash
# 通过命令行安装扩展
code --install-extension golang.go                 # Go 语言支持（必装）
code --install-extension ms-azuretools.vscode-docker  # Docker 支持
code --install-extension redhat.vscode-yaml        # YAML 支持
code --install-extension esbenp.prettier-vscode    # 格式化
code --install-extension eamodio.gitlens           # Git 增强
code --install-extension usernamehw.errorlens      # 错误高亮
```

**VS Code settings.json（Go 相关配置）：**

```json
{
    // Go 工具配置
    "go.useLanguageServer": true,
    "gopls": {
        "ui.semanticTokens": true,
        "formatting.gofumpt": true
    },

    // 保存时自动操作
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
        "source.organizeImports": "explicit"
    },

    // Go 文件格式化
    "[go]": {
        "editor.formatOnSave": true,
        "editor.codeActionsOnSave": {
            "source.organizeImports": "explicit"
        },
        "editor.defaultFormatter": "golang.go"
    },

    // 测试配置
    "go.testFlags": ["-v", "-count=1"],
    "go.coverOnSave": true,
    "go.coverOnTestPackage": true,
    "go.coverageDecorator": {
        "type": "gutter",
        "coveredHighlightColor": "rgba(64,128,64,0.2)",
        "uncoveredHighlightColor": "rgba(128,64,64,0.2)"
    },

    // Lint 配置
    "go.lintTool": "golangci-lint",
    "go.lintFlags": ["--fast"],

    // 代码片段
    "go.toolsManagement.autoUpdate": true
}
```

**安装 Go 工具链（VS Code 首次打开 .go 文件时会提示）：**

在 VS Code 中按 `Cmd+Shift+P`（macOS）或 `Ctrl+Shift+P`（Ubuntu），执行：

```
Go: Install/Update Tools
```

勾选全部工具并安装：`gopls`, `dlv`, `golangci-lint`, `goimports`, `gomodifytags`, `gotests`, `impl`, `fillstruct` 等。

### 5.2 GoLand

**配置步骤：**

1. **Go SDK：** Settings → Go → GOROOT → 选择安装的 Go 路径
2. **GOPATH：** Settings → Go → GOPATH → 设置 Global GOPATH
3. **Go Modules：** Settings → Go → Go Modules → 启用，设置 `GOPROXY=https://goproxy.cn,direct`
4. **Linter：** Settings → Go → Linter → 选择 `golangci-lint`
5. **File Watchers：** Settings → Tools → File Watchers → 添加 `goimports`（保存自动整理 import）
6. **Run Configuration：**
   - Program: `cmd/server/main.go`
   - Working directory: 项目根目录
   - Environment: 加载 `.env` 文件

---

## 6. Docker 环境（通用）

### 6.1 使用 docker-compose 启动全部依赖

项目根目录的 `docker-compose.yml` 会启动所有依赖服务：

```bash
# 启动所有服务
docker compose up -d

# 查看运行状态
docker compose ps

# 查看日志
docker compose logs -f app

# 停止所有服务
docker compose down

# 仅启动依赖（不含 app，用于本地开发）
docker compose up -d redis mysql postgres
```

### 6.2 仅用 Docker 运行 Redis（最轻量的方式）

如果不想装本地 Redis：

```bash
docker run -d \
    --name helloGo-redis \
    -p 6379:6379 \
    -v helloGo-redis-data:/data \
    redis:7-alpine \
    redis-server --appendonly yes

# 验证
docker exec -it helloGo-redis redis-cli ping
```

---

## 7. 常见问题排查

### Q1: `go mod download` 很慢或超时

**原因：** 默认使用 Google 代理，国内网络不稳定。

```bash
# 解决方案：设置国内代理
go env -w GOPROXY=https://goproxy.cn,direct

# 或者使用 Go 官方中国镜像
go env -w GOPROXY=https://goproxy.io,direct

# 验证
go env GOPROXY
```

### Q2: `permission denied` 运行 docker 命令（Ubuntu）

```bash
# 将用户加入 docker 组
sudo usermod -aG docker $USER

# 重新登录或执行
newgrp docker
```

### Q3: Redis 连接被拒绝

```bash
# 检查 Redis 是否运行
redis-cli ping

# Ubuntu: 检查服务状态
sudo systemctl status redis-server

# macOS: 检查 brew 服务
brew services list

# 检查端口占用
lsof -i :6379    # macOS
ss -tlnp | grep 6379  # Ubuntu

# 如果 bind 了 127.0.0.1 但需要远程访问：
# 编辑 redis.conf，修改 bind 0.0.0.0
```

### Q4: Go 版本不匹配

```bash
# 查看当前版本
go version

# 查看 Go 安装路径
which go
go env GOROOT

# Ubuntu: 手动指定版本
sudo rm -rf /usr/local/go
wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz

# macOS: 使用 Homebrew 升级
brew upgrade go
```

### Q5: macOS 上 `port 8000 already in use`

```bash
# 查找占用端口的进程
lsof -i :8000

# 杀掉进程
kill -9 <PID>

# 或修改项目配置使用其他端口
# 在 .env 中设置 PORT=8001
```

### Q6: VS Code 中 Go 工具安装失败

```bash
# 手动安装所有 Go 工具
go install golang.org/x/tools/gopls@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/rogpeppe/godef@latest
go install github.com/ramya-rao-a/go-outline@latest
go install github.com/cweill/gotests/gotests@latest
go install github.com/fatih/gomodifytags@latest
go install github.com/josharian/impl@latest
go install github.com/davidrjenni/reftools/cmd/fillstruct@latest
```

### Q7: 两平台间文件换行符问题（CRLF vs LF）

```bash
# 统一使用 LF（Git 配置）
git config --global core.autocrlf input   # macOS / Linux
git config --global core.eol lf

# 项目根目录添加 .gitattributes
echo "* text=auto eol=lf" > .gitattributes
```

### Q8: `swag init` 找不到命令

```bash
# 确保 GOPATH/bin 在 PATH 中
export PATH=$PATH:$(go env GOPATH)/bin

# 写入 shell 配置
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc   # macOS
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc  # Ubuntu
```

---

## 快速对比：Ubuntu vs macOS 命令差异

| 操作               | Ubuntu 22.04                          | macOS (Mac Mini)                     |
| ------------------ | ------------------------------------- | ------------------------------------ |
| 包管理器           | `apt`                                 | `brew`                               |
| 安装 Go            | 手动下载 tar.gz                       | `brew install go`                    |
| 安装 Docker        | apt 安装 docker-ce                    | `brew install --cask docker`         |
| 安装 Redis         | `apt install redis-server`            | `brew install redis`                 |
| 安装 MySQL         | `apt install mysql-server`            | `brew install mysql`                 |
| 安装 PostgreSQL    | apt + pgdg 源                         | `brew install postgresql@15`         |
| 服务管理           | `systemctl start/stop/status`         | `brew services start/stop/list`      |
| Shell 配置文件     | `~/.bashrc`                           | `~/.zshrc`                           |
| 查看端口占用       | `ss -tlnp` 或 `lsof -i :PORT`        | `lsof -i :PORT`                      |
| Go 默认安装路径    | `/usr/local/go`                       | `/opt/homebrew/opt/go`（Apple Silicon） |
| Docker Compose     | docker-compose-plugin (apt)           | Docker Desktop 内置                  |
| golangci-lint      | curl 脚本安装到 GOPATH/bin            | `brew install golangci-lint`         |
