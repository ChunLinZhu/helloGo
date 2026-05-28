# 第二阶段开发计划：gRPC + 微服务架构

> 基于第一阶段 Fiber 单体应用，渐进式拆分为 gRPC 微服务  
> 面向 Go 初学者设计，每个阶段都包含学习目标、示例代码和动手练习

---

## 目录

- [0. 学习路线图](#0-学习路线图)
- [1. gRPC 基础入门](#1-grpc-基础入门)
- [2. 架构设计](#2-架构设计)
- [3. 共享基础层 (pkg + proto)](#3-共享基础层-pkg--proto)
- [4. 用户微服务 (User Service)](#4-用户微服务-user-service)
- [5. 认证微服务 (Auth Service)](#5-认证微服务-auth-service)
- [6. 权限微服务 (Permission Service)](#6-权限微服务-permission-service)
- [7. 业务微服务 (Biz Service)](#7-业务微服务-biz-service)
- [8. API Gateway 网关层](#8-api-gateway-网关层)
- [9. 服务治理](#9-服务治理)
- [10. 可观测性](#10-可观测性)
- [11. 容器编排与部署](#11-容器编排与部署)
- [12. 阶段总结与时间规划](#12-阶段总结与时间规划)

---

## 0. 学习路线图

> 作为 Go 初学者，建议按照以下顺序逐步学习，每完成一个里程碑就验证理解再往下走。

```
┌──────────────────────────────────────────────────────────────────┐
│                      gRPC 微服务学习路线                          │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Milestone 1: 理解 gRPC 是什么                                   │
│  ├── 什么是 Protocol Buffers（.proto 文件）                       │
│  ├── gRPC vs REST 的区别                                         │
│  ├── 四种通信模式：Unary / Server Stream / Client Stream / BiDi  │
│  └── 动手：写一个 Hello World gRPC 服务                          │
│                                                                  │
│  Milestone 2: 理解微服务拆分思路                                  │
│  ├── 单体 → 微服务：为什么要拆？怎么拆？                          │
│  ├── 服务间通信：同步（gRPC）vs 异步（消息队列）                  │
│  ├── API Gateway 模式                                            │
│  └── 动手：画出目标架构图                                        │
│                                                                  │
│  Milestone 3: 逐个实现微服务                                     │
│  ├── User Service（第一个完整 gRPC 服务）                        │
│  ├── Auth Service（JWT + Redis 会话）                            │
│  ├── Permission Service（角色/权限 CRUD）                        │
│  └── Biz Service（菜单/部门/字典/日志）                          │
│                                                                  │
│  Milestone 4: API Gateway                                        │
│  ├── Fiber HTTP → gRPC 转发                                     │
│  ├── JWT 验证在网关层统一处理                                    │
│  └── 动手：用 curl 通过 Gateway 调用后端 gRPC 服务               │
│                                                                  │
│  Milestone 5: 生产化                                             │
│  ├── 服务发现（etcd / Consul）                                   │
│  ├── 链路追踪（OpenTelemetry + Jaeger）                          │
│  ├── Docker Compose 一键启动                                     │
│  └── 动手：完整 E2E 测试                                         │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### 推荐学习资源

| 资源 | 链接 | 说明 |
|------|------|------|
| gRPC 官方文档 | https://grpc.io/docs/languages/go/ | Go gRPC 快速入门 |
| Protocol Buffers 指南 | https://protobuf.dev/ | .proto 语法详解 |
| Go gRPC 示例 | https://github.com/grpc/grpc-go/tree/master/examples | 官方示例代码 |
| 《Go 语言设计与实现》 | https://draveness.me/golang/ | Go 底层原理 |
| 微服务模式 | https://microservices.io/ | 微服务架构模式参考 |

---

## 1. gRPC 基础入门

### 1.1 核心概念（先理解再写代码）

**什么是 gRPC？**

```
┌──────────────┐                    ┌──────────────┐
│   Client     │ ──── gRPC Call ──→ │   Server     │
│              │ ←─── Response ──── │              │
│  (任何语言)   │                    │  (任何语言)   │
└──────────────┘                    └──────────────┘
       │                                   │
       │   .proto 文件定义接口              │
       │   自动生成客户端/服务端代码        │
       │                                   │
       ▼                                   ▼
  Stub 代码                          Stub 代码
  (像调本地函数一样调远程服务)
```

**gRPC vs REST 对比：**

| 特性          | REST (HTTP/JSON)           | gRPC (HTTP/2 + Protobuf)        |
| ------------- | -------------------------- | -------------------------------- |
| 数据格式      | JSON（文本，可读）          | Protobuf（二进制，紧凑）         |
| 性能          | 较慢（JSON 序列化开销）     | 快 3-10 倍                       |
| 接口定义      | OpenAPI/Swagger（可选）     | .proto 文件（强制，即代码即文档） |
| 代码生成      | 需第三方工具                | 官方工具自动生成                 |
| 流式通信      | 不原生支持                  | 原生支持四种流模式               |
| 浏览器直接调用 | ✅ 可以                     | ❌ 需要 grpc-web 代理            |
| 适用场景      | 面向外部/前端               | 微服务间内部通信                 |

**四种通信模式（初学者先掌握 Unary 即可）：**

```
1. Unary（一元）           — 最常用，一问一答
   Client ──Request──→ Server
   Client ←─Response── Server

2. Server Streaming（服务端流）— 服务端持续推送
   Client ──Request──→ Server
   Client ←─Stream1──── Server
   Client ←─Stream2──── Server
   Client ←─Stream3──── Server

3. Client Streaming（客户端流）— 客户端持续发送
   Client ──Stream1──→ Server
   Client ──Stream2──→ Server
   Client ──Stream3──→ Server
   Client ←─Response── Server

4. Bidirectional Streaming（双向流）— 双方同时收发
   Client ←→ Server（持续双向通信，如聊天）
```

### 1.2 Hello World 练习

> 在正式开始项目之前，先用一个小例子理解 gRPC 的工作流程。

**Step 1: 创建练习目录**

```bash
mkdir -p ~/grpc-playground && cd ~/grpc-playground
go mod init grpc-playground
```

**Step 2: 安装 gRPC 工具链**

```bash
# protoc 编译器（Ubuntu）
sudo apt install -y protobuf-compiler
# protoc 编译器（macOS）
brew install protobuf

# Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 验证
protoc --version
protoc-gen-go --version
protoc-gen-go-grpc --version
```

**Step 3: 编写 .proto 文件**

```protobuf
// proto/hello.proto
syntax = "proto3";

package hello;
option go_package = "grpc-playground/proto";

// 定义服务接口
service HelloService {
  // Unary RPC：一问一答
  rpc SayHello (HelloRequest) returns (HelloResponse);
}

// 请求消息
message HelloRequest {
  string name = 1;        // 字段编号，不是默认值
  int32 age = 2;
}

// 响应消息
message HelloResponse {
  string message = 1;
  int64 timestamp = 2;
}
```

**Step 4: 生成 Go 代码**

```bash
# 在项目根目录执行
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/hello.proto

# 生成两个文件：
#   proto/hello.pb.go       ← 消息结构体
#   proto/hello_grpc.pb.go  ← gRPC 服务端/客户端代码
```

**Step 5: 实现 Server**

```go
// server/main.go
package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "time"

    pb "grpc-playground/proto"
    "google.golang.org/grpc"
)

// server 结构体，嵌入生成的 UnimplementedServer
type server struct {
    pb.UnimplementedHelloServiceServer
}

// 实现 SayHello 方法（业务逻辑写在这里）
func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
    log.Printf("收到请求: name=%s, age=%d", req.Name, req.Age)

    return &pb.HelloResponse{
        Message:   fmt.Sprintf("Hello %s, you are %d years old!", req.Name, req.Age),
        Timestamp: time.Now().Unix(),
    }, nil
}

func main() {
    // 监听端口
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("监听失败: %v", err)
    }

    // 创建 gRPC 服务器
    grpcServer := grpc.NewServer()

    // 注册服务
    pb.RegisterHelloServiceServer(grpcServer, &server{})

    log.Println("gRPC Server 启动在 :50051")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("服务启动失败: %v", err)
    }
}
```

**Step 6: 实现 Client**

```go
// client/main.go
package main

import (
    "context"
    "log"
    "time"

    pb "grpc-playground/proto"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    // 连接 gRPC 服务器
    conn, err := grpc.NewClient(
        "localhost:50051",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatalf("连接失败: %v", err)
    }
    defer conn.Close()

    // 创建客户端 stub
    client := pb.NewHelloServiceClient(conn)

    // 调用远程方法（像调本地函数一样！）
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    resp, err := client.SayHello(ctx, &pb.HelloRequest{
        Name: "Go 初学者",
        Age:  1,
    })
    if err != nil {
        log.Fatalf("调用失败: %v", err)
    }

    log.Printf("收到响应: %s (时间戳: %d)", resp.Message, resp.Timestamp)
}
```

**Step 7: 运行测试**

```bash
# 终端 1：启动 Server
go run server/main.go

# 终端 2：启动 Client
go run client/main.go
# 预期输出: 收到响应: Hello Go 初学者, you are 1 years old! (时间戳: 17...)
```

### 1.3 动手练习清单

完成上面的 Hello World 后，尝试以下练习来加深理解：

| # | 练习 | 学习目标 |
|---|------|----------|
| 1 | 给 `HelloRequest` 添加 `repeated string hobbies` 字段 | 理解 repeated（数组）类型 |
| 2 | 新增一个嵌套 message `Address`，在 Request 中引用 | 理解嵌套消息 |
| 3 | 添加 `rpc SayHelloStream` 服务端流模式 | 理解流式通信 |
| 4 | 在 Server 端添加 Interceptor（日志拦截器） | 理解 gRPC 中间件机制 |
| 5 | 让 Server 返回一个 error，Client 处理 | 理解 gRPC 错误处理（Status codes） |

---

## 2. 架构设计

### 2.1 目标架构

```
                        ┌─────────────┐
                        │   Client    │
                        │ (浏览器/App) │
                        └──────┬──────┘
                               │ HTTP/REST (JSON)
                               ▼
                 ┌─────────────────────────┐
                 │     API Gateway         │
                 │   (Fiber HTTP Server)    │
                 │                         │
                 │  • JWT 验证             │
                 │  • 路由转发             │
                 │  • 限流 / CORS          │
                 │  • 请求日志             │
                 │  • Swagger 文档         │
                 └────┬────┬────┬────┬────┘
                      │    │    │    │  gRPC (Protobuf)
           ┌──────────┘    │    │    └──────────┐
           ▼               ▼    ▼               ▼
    ┌────────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
    │ User Svc   │ │ Auth Svc │ │ Perm Svc │ │ Biz Svc  │
    │            │ │          │ │          │ │          │
    │ 用户 CRUD  │ │ 登录     │ │ 角色     │ │ 菜单     │
    │ 部门       │ │ 注册     │ │ 权限     │ │ 部门     │
    │ 上传       │ │ JWT 管理 │ │ 菜单权限 │ │ 字典     │
    │            │ │ 会话管理 │ │          │ │ 日志     │
    └─────┬──────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘
          │              │            │             │
          ▼              ▼            ▼             ▼
    ┌──────────────────────────────────────────────────┐
    │                 共享基础设施                       │
    │  ┌───────┐  ┌────────┐  ┌──────────┐  ┌───────┐ │
    │  │ MySQL │  │ Redis  │  │ etcd     │  │ Jaeger│ │
    │  │/SQLite│  │        │  │(服务发现) │  │(追踪) │ │
    │  └───────┘  └────────┘  └──────────┘  └───────┘ │
    └──────────────────────────────────────────────────┘
```

### 2.2 服务拆分原则

| 原则 | 说明 |
|------|------|
| **单一职责** | 每个服务只做一件事，如 Auth 只管认证 |
| **独立部署** | 每个服务独立编译、独立 Docker 容器 |
| **独立数据库** | 理想情况每个服务自己的 DB（本阶段共享 DB，用 schema 隔离） |
| **通过接口通信** | 服务间只通过 gRPC 接口交互，不直接访问对方数据库 |
| **渐进拆分** | 先从单体中拆出最独立的模块（Auth），逐步扩展 |

> **目录迁移说明：** Phase 1 的 `internal/pkg/`（response、errors、pagination、redis）在 Phase 2 中拆分到两个位置：
> - `internal/shared/` — 微服务共享的**基础设施**代码（config、database、redis、logger、interceptor）
> - 顶层 `pkg/` — 可复用的**公共包**（errors、pagination、response），可被外部项目引用
>
> Phase 1 的 `internal/config/`、`internal/database/` 保留给单体应用（兼容模式），微服务使用 `internal/shared/` 下的新实现。

### 2.3 拆分方案

从第一阶段的单体应用拆分为 4 个微服务 + 1 个网关：

| 服务 | 端口 | 职责 | 拆分理由 |
|------|------|------|----------|
| **API Gateway** | 8000 (HTTP) | HTTP→gRPC 转发、JWT 验证、限流 | 对外统一入口 |
| **User Service** | 50001 (gRPC) | 用户 CRUD、上传 | 用户是核心领域 |
| **Auth Service** | 50002 (gRPC) | 登录、Token 管理、会话 | 认证逻辑独立且可复用 |
| **Permission Service** | 50003 (gRPC) | 角色、权限、菜单权限 | 权限是横切关注点 |
| **Biz Service** | 50004 (gRPC) | 菜单、部门、字典、日志 | 基础业务数据 |

### 2.4 项目目录结构

```
helloGo/
├── api/
│   └── proto/                          # Protobuf 定义（所有服务共享）
│       ├── user/
│       │   └── v1/
│       │       └── user.proto
│       ├── auth/
│       │   └── v1/
│       │       └── auth.proto
│       ├── permission/
│       │   └── v1/
│       │       └── permission.proto
│       ├── biz/
│       │   └── v1/
│       │       └── biz.proto
│       └── common/
│           └── v1/
│               └── common.proto        # 共享消息类型（分页、时间等）
│
├── gen/                                # protoc 生成的代码（不提交到 git）
│   └── go/
│       ├── user/v1/
│       ├── auth/v1/
│       ├── permission/v1/
│       ├── biz/v1/
│       └── common/v1/
│
├── cmd/                                # 每个服务一个入口
│   ├── gateway/
│   │   └── main.go                     # API Gateway 入口
│   ├── user/
│   │   └── main.go                     # User Service 入口
│   ├── auth/
│   │   └── main.go                     # Auth Service 入口
│   ├── permission/
│   │   └── main.go                     # Permission Service 入口
│   ├── biz/
│   │   └── main.go                     # Biz Service 入口
│   └── seed/
│       └── main.go                     # 种子数据
│
├── internal/
│   ├── gateway/                        # 网关层（HTTP → gRPC）
│   │   ├── handler/                    # Fiber HTTP handler
│   │   │   ├── auth_handler.go
│   │   │   ├── user_handler.go
│   │   │   ├── permission_handler.go
│   │   │   └── biz_handler.go
│   │   ├── middleware/                 # 网关中间件
│   │   │   ├── jwt.go
│   │   │   ├── cors.go
│   │   │   ├── ratelimit.go
│   │   │   └── trace.go
│   │   └── router.go
│   │
│   ├── user/                           # User 微服务
│   │   ├── server.go                   # gRPC Server 注册
│   │   ├── service.go                  # 业务逻辑
│   │   ├── model.go                    # GORM 模型
│   │   └── repository.go              # 数据库操作
│   │
│   ├── auth/                           # Auth 微服务
│   │   ├── server.go
│   │   ├── service.go
│   │   ├── jwt.go
│   │   └── session.go                  # Redis 会话管理
│   │   # 注意：Auth Service 没有 model.go / repository.go / dto.go
│   │   # 因为 Auth 不直连数据库，用户查询通过 gRPC 调用 User Service
│   │
│   ├── permission/                     # Permission 微服务
│   │   ├── server.go
│   │   ├── service.go
│   │   ├── role_model.go
│   │   ├── permission_model.go
│   │   ├── menu_model.go
│   │   └── repository.go
│   │
│   ├── biz/                            # Biz 微服务
│   │   ├── server.go
│   │   ├── service.go
│   │   ├── department_model.go
│   │   ├── dict_model.go
│   │   ├── log_model.go
│   │   ├── upload_model.go
│   │   └── repository.go
│   │
│   └── shared/                         # 共享内部代码
│       ├── config/
│       │   └── config.go
│       ├── database/
│       │   └── database.go
│       ├── redis/
│       │   └── redis.go
│       ├── logger/
│       │   └── logger.go
│       └── interceptor/               # gRPC 拦截器
│           ├── auth.go                 # 认证拦截器
│           ├── logging.go             # 日志拦截器
│           ├── recovery.go            # 恢复拦截器
│           └── metrics.go             # 指标拦截器
│
├── pkg/                                # 可复用的公共包
│   ├── errors/
│   │   └── errors.go                   # gRPC 错误码
│   ├── pagination/
│   │   └── pagination.go
│   └── response/
│       └── response.go
│
├── configs/
│   ├── gateway.yaml
│   ├── user.yaml
│   ├── auth.yaml
│   ├── permission.yaml
│   └── biz.yaml
│
├── scripts/
│   ├── gen-proto.sh                    # 一键生成 protobuf 代码
│   └── test_grpc.sh                    # gRPC 测试脚本
│
├── docker-compose.yml
├── docker-compose.infra.yml            # 仅基础设施（redis/mysql/etcd/jaeger）
├── Makefile
├── buf.yaml                            # Buf 配置（可选，替代 protoc）
└── buf.gen.yaml
```

---

## 3. 共享基础层 (pkg + proto)

### 3.1 技术栈补充

在第一阶段基础上新增：

| 组件 | 技术选型 | 说明 |
|------|----------|------|
| gRPC 框架 | `google.golang.org/grpc` | Go gRPC 核心库 |
| Protobuf | `google.golang.org/protobuf` | Protocol Buffers 运行时 |
| gRPC Gateway | `grpc-ecosystem/grpc-gateway` | REST → gRPC 自动转发（可选） |
| gRPC 中间件 | `grpc-ecosystem/go-grpc-middleware` | 拦截器集合 |
| 服务发现 | `go.etcd.io/etcd/client/v3` | etcd 客户端 |
| 链路追踪 | `go.opentelemetry.io/otel` | OpenTelemetry SDK |
| 追踪导出 | Jaeger | 可视化链路追踪 |
| 服务注册 | `google.golang.org/grpc/resolver` | gRPC 内置服务发现接口 |
| 负载均衡 | gRPC 内置 `round_robin` | 客户端负载均衡 |

### 3.2 Protobuf 定义规范

**common.proto — 共享消息类型**

```protobuf
// api/proto/common/v1/common.proto
syntax = "proto3";
package common.v1;
option go_package = "helloGo/gen/go/common/v1;commonv1";

import "google/protobuf/timestamp.proto";

// 分页请求
message PaginationRequest {
  int32 page = 1;      // 页码，从 1 开始
  int32 limit = 2;     // 每页数量
}

// 分页响应
message PaginationResponse {
  int32 page = 1;
  int32 limit = 2;
  int64 total = 3;
  int32 total_pages = 4;
}

// 通用空响应
message Empty {}

// 通用 ID 请求
message IDRequest {
  string id = 1;
}
```

**proto 定义规范（团队约定）：**

```
规范要点：
1. 每个服务一个独立的 package（如 user.v1, auth.v1）
2. 版本号放在 package 中（v1），便于后续 API 演进
3. 消息命名用 PascalCase，字段命名用 snake_case
4. 所有时间字段使用 google.protobuf.Timestamp
5. 所有 ID 使用 string 类型（UUID）
6. 每个 RPC 方法都要写注释（生成文档用）
7. 字段编号 1-15 用一个字节编码，常用字段放前面
```

### 3.3 代码生成脚本

```bash
#!/bin/bash
# scripts/gen-proto.sh

set -e

PROTO_DIR="api/proto"
OUT_DIR="gen/go"

echo "🔨 生成 Protobuf 代码..."

# 清理旧代码
rm -rf ${OUT_DIR}
mkdir -p ${OUT_DIR}

# 遍历所有 .proto 文件
find ${PROTO_DIR} -name "*.proto" | while read -r proto_file; do
    echo "  📦 编译: ${proto_file}"
    protoc \
        --proto_path=${PROTO_DIR} \
        --proto_path=/usr/local/include \
        --go_out=${OUT_DIR} \
        --go_opt=paths=source_relative \
        --go-grpc_out=${OUT_DIR} \
        --go-grpc_opt=paths=source_relative \
        "${proto_file}"
done

echo "✅ 代码生成完成 → ${OUT_DIR}/"
```

Makefile 中添加：

```makefile
.PHONY: proto
proto:
	bash scripts/gen-proto.sh

.PHONY: proto-install
proto-install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 3.4 共享 gRPC 拦截器

> gRPC 的拦截器（Interceptor）等同于 Fiber 的中间件（Middleware）

**日志拦截器示例（初学者重点理解）：**

```go
// internal/shared/interceptor/logging.go
package interceptor

import (
    "context"
    "log"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/status"
)

// UnaryServerLoggingInterceptor 一元服务端日志拦截器
// 功能：记录每个 gRPC 请求的方法、耗时、状态码
func UnaryServerLoggingInterceptor() grpc.UnaryServerInterceptor {
    return func(
        ctx context.Context,
        req interface{},
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (interface{}, error) {
        start := time.Now()

        // 调用实际的 handler
        resp, err := handler(ctx, req)

        // 记录日志
        duration := time.Since(start)
        st, _ := status.FromError(err)

        log.Printf("[gRPC] %s | %s | %v | code=%s",
            info.FullMethod,
            st.Code().String(),
            duration,
            st.Message(),
        )

        return resp, err
    }
}
```

**Recovery 拦截器（防止 panic 导致服务崩溃）：**

```go
// internal/shared/interceptor/recovery.go
package interceptor

import (
    "log"
    "runtime/debug"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func UnaryServerRecoveryInterceptor() grpc.UnaryServerInterceptor {
    return func(
        ctx context.Context,
        req interface{},
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (resp interface{}, err error) {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("[PANIC] %s: %v\n%s", info.FullMethod, r, debug.Stack())
                err = status.Errorf(codes.Internal, "服务内部错误")
            }
        }()
        return handler(ctx, req)
    }
}
```

---

## 4. 用户微服务 (User Service)

> 第一个完整的 gRPC 微服务，从这里开始实战

### 4.1 Proto 定义

```protobuf
// api/proto/user/v1/user.proto
syntax = "proto3";
package user.v1;
option go_package = "helloGo/gen/go/user/v1;userv1";

import "google/protobuf/timestamp.proto";
import "common/v1/common.proto";

// ============ 服务定义 ============
service UserService {
  // 获取用户列表（分页）
  rpc ListUsers (ListUsersRequest) returns (ListUsersResponse);
  // 根据 ID 获取用户
  rpc GetUser (GetUserRequest) returns (UserResponse);
  // 创建用户
  rpc CreateUser (CreateUserRequest) returns (UserResponse);
  // 更新用户
  rpc UpdateUser (UpdateUserRequest) returns (UserResponse);
  // 删除用户
  rpc DeleteUser (DeleteUserRequest) returns (common.v1.Empty);
  // 根据用户名查找（Auth Service 内部调用）
  rpc GetUserByUsername (GetUserByUsernameRequest) returns (UserResponse);
}

// ============ 消息定义 ============

message User {
  string id = 1;
  string username = 2;
  string email = 3;
  string phone = 4;
  bool is_active = 5;
  repeated string role_codes = 6;      // 角色编码列表
  google.protobuf.Timestamp created_at = 7;
  google.protobuf.Timestamp updated_at = 8;
}

// --- 请求 / 响应 ---

message ListUsersRequest {
  common.v1.PaginationRequest pagination = 1;
  string keyword = 2;                   // 搜索关键词（用户名/邮箱）
}

message ListUsersResponse {
  repeated User users = 1;
  common.v1.PaginationResponse pagination = 2;
}

message GetUserRequest {
  string id = 1;
}

message GetUserByUsernameRequest {
  string username = 1;
}

message CreateUserRequest {
  string username = 1;
  string password = 2;
  string email = 3;
  string phone = 4;
  repeated string role_codes = 5;
}

message UpdateUserRequest {
  string id = 1;
  optional string email = 2;
  optional string phone = 3;
  optional bool is_active = 4;
  repeated string role_codes = 5;
}

message DeleteUserRequest {
  string id = 1;
}

message UserResponse {
  User user = 1;
}
```

### 4.2 GORM 模型

```go
// internal/user/model.go
package user

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

type User struct {
    ID           string    `gorm:"type:varchar(36);primaryKey" json:"id"`
    Username     string    `gorm:"type:varchar(64);uniqueIndex" json:"username"`
    PasswordHash string    `gorm:"type:varchar(128)" json:"-"`
    Email        *string   `gorm:"type:varchar(128);index" json:"email"`
    Phone        *string   `gorm:"type:varchar(32);index" json:"phone"`
    IsActive     bool      `gorm:"default:true" json:"is_active"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.ID == "" {
        u.ID = uuid.New().String()
    }
    return nil
}

func (User) TableName() string {
    return "users"
}
```

### 4.3 Repository 层（数据库操作）

```go
// internal/user/repository.go
package user

import (
    "gorm.io/gorm"
)

type Repository struct {
    db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
    return &Repository{db: db}
}

func (r *Repository) List(keyword string, page, limit int) ([]User, int64, error) {
    var users []User
    var total int64

    query := r.db.Model(&User{})
    if keyword != "" {
        like := "%" + keyword + "%"
        query = query.Where("username LIKE ? OR email LIKE ?", like, like)
    }

    query.Count(&total)
    err := query.Offset((page - 1) * limit).Limit(limit).
        Order("created_at DESC").Find(&users).Error

    return users, total, err
}

func (r *Repository) GetByID(id string) (*User, error) {
    var user User
    err := r.db.Where("id = ?", id).First(&user).Error
    return &user, err
}

func (r *Repository) GetByUsername(username string) (*User, error) {
    var user User
    err := r.db.Where("username = ?", username).First(&user).Error
    return &user, err
}

func (r *Repository) Create(user *User) error {
    return r.db.Create(user).Error
}

func (r *Repository) Update(user *User) error {
    return r.db.Save(user).Error
}

func (r *Repository) Delete(id string) error {
    return r.db.Where("id = ?", id).Delete(&User{}).Error
}
```

### 4.4 Service 层（业务逻辑）

```go
// internal/user/service.go
package user

import (
    "context"

    userv1 "helloGo/gen/go/user/v1"
    "golang.org/x/crypto/bcrypt"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
    userv1.UnimplementedUserServiceServer
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

// ListUsers 实现分页查询
func (s *Service) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
    page := int(req.Pagination.Page)
    limit := int(req.Pagination.Limit)
    if page < 1 { page = 1 }
    if limit < 1 { limit = 10 }

    users, total, err := s.repo.List(req.Keyword, page, limit)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "查询失败: %v", err)
    }

    // GORM 模型 → Protobuf 消息
    pbUsers := make([]*userv1.User, len(users))
    for i, u := range users {
        pbUsers[i] = toProtoUser(&u)
    }

    totalPages := int32((total + int64(limit) - 1) / int64(limit))
    return &userv1.ListUsersResponse{
        Users: pbUsers,
        Pagination: &commonv1.PaginationResponse{
            Page:       int32(page),
            Limit:      int32(limit),
            Total:      total,
            TotalPages: totalPages,
        },
    }, nil
}

// CreateUser 创建用户
func (s *Service) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.UserResponse, error) {
    // 校验
    if req.Username == "" || req.Password == "" {
        return nil, status.Error(codes.InvalidArgument, "用户名和密码不能为空")
    }

    // 检查用户名是否已存在
    if existing, _ := s.repo.GetByUsername(req.Username); existing != nil {
        return nil, status.Error(codes.AlreadyExists, "用户名已存在")
    }

    // 加密密码
    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "密码加密失败: %v", err)
    }

    user := &User{
        Username:     req.Username,
        PasswordHash: string(hash),
        Email:        nilIfEmpty(req.Email),
        Phone:        nilIfEmpty(req.Phone),
        IsActive:     true,
    }

    if err := s.repo.Create(user); err != nil {
        return nil, status.Errorf(codes.Internal, "创建失败: %v", err)
    }

    return &userv1.UserResponse{User: toProtoUser(user)}, nil
}

// toProtoUser 将 GORM 模型转换为 Protobuf 消息
func toProtoUser(u *User) *userv1.User {
    email, phone := "", ""
    if u.Email != nil { email = *u.Email }
    if u.Phone != nil { phone = *u.Phone }

    return &userv1.User{
        Id:        u.ID,
        Username:  u.Username,
        Email:     email,
        Phone:     phone,
        IsActive:  u.IsActive,
        CreatedAt: timestamppb.New(u.CreatedAt),
        UpdatedAt: timestamppb.New(u.UpdatedAt),
    }
}

func nilIfEmpty(s string) *string {
    if s == "" { return nil }
    return &s
}
```

### 4.5 gRPC Server 入口

```go
// internal/user/server.go
package user

import (
    "log"
    "net"

    userv1 "helloGo/gen/go/user/v1"
    "helloGo/internal/shared/interceptor"

    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
)

type Server struct {
    grpcServer *grpc.Server
    port       string
}

func NewServer(port string, svc *Service) *Server {
    // 创建 gRPC Server，注册拦截器
    grpcServer := grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            interceptor.UnaryServerRecoveryInterceptor(),
            interceptor.UnaryServerLoggingInterceptor(),
        ),
    )

    // 注册 UserService
    userv1.RegisterUserServiceServer(grpcServer, svc)

    // 注册反射服务（用于 grpcurl 调试）
    reflection.Register(grpcServer)

    return &Server{grpcServer: grpcServer, port: port}
}

func (s *Server) Start() error {
    lis, err := net.Listen("tcp", ":"+s.port)
    if err != nil {
        return err
    }
    log.Printf("[User Service] gRPC 服务启动在 :%s", s.port)
    return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() {
    s.grpcServer.GracefulStop()
}
```

### 4.6 cmd 入口

```go
// cmd/user/main.go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"

    "helloGo/internal/shared/config"
    "helloGo/internal/shared/database"
    "helloGo/internal/user"
)

func main() {
    // 1. 加载配置
    cfg := config.Load("user")

    // 2. 初始化数据库
    db := database.Connect(cfg.Database)

    // 3. 构建依赖链：Repository → Service → Server
    repo := user.NewRepository(db)
    svc := user.NewService(repo)
    srv := user.NewServer(cfg.GRPCPort, svc)

    // 4. 启动服务（goroutine）
    go func() {
        if err := srv.Start(); err != nil {
            log.Fatalf("User Service 启动失败: %v", err)
        }
    }()

    // 5. 优雅关闭
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("[User Service] 正在关闭...")
    srv.Stop()
    log.Println("[User Service] 已关闭")
}
```

### 4.7 测试

**使用 grpcurl 命令行工具测试：**

```bash
# 安装 grpcurl（Ubuntu）
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 安装 grpcurl（macOS）
brew install grpcurl

# 列出所有服务
grpcurl -plaintext localhost:50001 list

# 列出 UserService 的所有方法
grpcurl -plaintext localhost:50001 list user.v1.UserService

# 调用 ListUsers
grpcurl -plaintext -d '{
  "pagination": {"page": 1, "limit": 10}
}' localhost:50001 user.v1.UserService/ListUsers

# 调用 CreateUser
grpcurl -plaintext -d '{
  "username": "testuser",
  "password": "test1234",
  "email": "test@example.com"
}' localhost:50001 user.v1.UserService/CreateUser
```

### 4.8 动手练习

| # | 练习 | 目标 |
|---|------|------|
| 1 | 独立运行 User Service，用 grpcurl 测试全部 5 个 RPC | 熟悉 gRPC 调用流程 |
| 2 | 给 CreateUser 添加字段校验（用户名长度、邮箱格式） | 理解 gRPC 错误处理 |
| 3 | 实现 GetUser / UpdateUser / DeleteUser 的完整逻辑 | 完整 CRUD |
| 4 | 添加 Stream RPC：`WatchUsers` 监听用户变更 | 理解流式通信 |

---

## 5. 认证微服务 (Auth Service)

> **设计决策：** Auth Service 是唯一的"无数据库"微服务。它不直连 MySQL/PostgreSQL，  
> 用户信息查询通过 gRPC 调用 User Service 完成（见 §5.2）。  
> Auth Service 只需要 Redis（会话管理 + 登录锁定）和 JWT 密钥。  
> 因此 Auth Service 没有 model.go / repository.go，与其他服务的标准模块结构不同。

### 5.1 Proto 定义

```protobuf
// api/proto/auth/v1/auth.proto
syntax = "proto3";
package auth.v1;
option go_package = "helloGo/gen/go/auth/v1;authv1";

import "common/v1/common.proto";

service AuthService {
  // 用户登录
  rpc Login (LoginRequest) returns (LoginResponse);
  // 刷新 Token
  rpc RefreshToken (RefreshTokenRequest) returns (LoginResponse);
  // 注销
  rpc Logout (LogoutRequest) returns (common.v1.Empty);
  // 验证 Token（供 Gateway 调用）
  rpc VerifyToken (VerifyTokenRequest) returns (VerifyTokenResponse);
  // 请求密码重置
  rpc RequestPasswordReset (RequestPasswordResetRequest) returns (RequestPasswordResetResponse);
  // 确认密码重置
  rpc ConfirmPasswordReset (ConfirmPasswordResetRequest) returns (common.v1.Empty);
  // 解锁账户
  rpc UnlockAccount (UnlockAccountRequest) returns (common.v1.Empty);
}

message LoginRequest {
  string username = 1;
  string password = 2;
}

message LoginResponse {
  string access_token = 1;
  string refresh_token = 2;
  string session_id = 3;
  int64 expires_in = 4;         // 过期时间（秒）
}

message RefreshTokenRequest {
  string refresh_token = 1;
  string session_id = 2;
}

message LogoutRequest {
  string session_id = 1;
  string user_id = 2;
}

message VerifyTokenRequest {
  string token = 1;
}

message VerifyTokenResponse {
  string user_id = 1;
  string username = 2;
  repeated string roles = 3;
}

message RequestPasswordResetRequest {
  string username = 1;
}

message RequestPasswordResetResponse {
  string reset_token = 1;
}

message ConfirmPasswordResetRequest {
  string reset_token = 1;
  string new_password = 2;
  string user_id = 3;
}

message UnlockAccountRequest {
  string username = 1;
}
```

### 5.2 服务间调用：Auth → User

> Auth Service 需要调用 User Service 来查找用户、验证密码  
> 这就是微服务间 gRPC 调用的第一个实例

```go
// internal/auth/service.go
package auth

import (
    "context"

    authv1 "helloGo/gen/go/auth/v1"
    userv1 "helloGo/gen/go/user/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

type Service struct {
    authv1.UnimplementedAuthServiceServer

    userClient  userv1.UserServiceClient  // 调用 User Service 的客户端
    redis       *redis.Client
    jwtSecret   string
}

func NewService(redisClient *redis.Client, jwtSecret string, userServiceAddr string) (*Service, error) {
    // 创建到 User Service 的 gRPC 连接
    conn, err := grpc.NewClient(
        userServiceAddr,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        return nil, err
    }

    return &Service{
        userClient: userv1.NewUserServiceClient(conn),
        redis:      redisClient,
        jwtSecret:  jwtSecret,
    }, nil
}

// Login 登录流程（调用 User Service 获取用户信息）
func (s *Service) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
    // 1. 检查账户是否被锁定
    lockKey := "login:lock:" + req.Username
    locked, _ := s.redis.Exists(ctx, lockKey).Result()
    if locked > 0 {
        return nil, status.Error(codes.PermissionDenied, "账户已被锁定，请稍后重试")
    }

    // 2. 调用 User Service 获取用户信息（跨服务调用！）
    userResp, err := s.userClient.GetUserByUsername(ctx, &userv1.GetUserByUsernameRequest{
        Username: req.Username,
    })
    if err != nil {
        // 用户不存在
        s.incrementFailCount(ctx, req.Username)
        return nil, status.Error(codes.Unauthenticated, "用户名或密码错误")
    }

    // 3. 验证密码（密码哈希存在 User Service 中，
    //    实际做法：User Service 提供 VerifyPassword RPC）
    // ... bcrypt 校验逻辑 ...

    // 4. 生成 JWT Token
    accessToken, refreshToken, sessionID := s.generateTokens(userResp.User)

    // 5. 存储 Session 到 Redis
    s.storeSession(ctx, userResp.User.Id, sessionID, refreshToken)

    return &authv1.LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        SessionId:    sessionID,
        ExpiresIn:    3600,
    }, nil
}
```

**服务间调用图：**

```
┌──────────┐  gRPC: Login()       ┌──────────┐
│   Auth   │ ──────────────────→  │   User   │
│  Service │                      │  Service │
│          │ ←──────────────────  │          │
│          │  UserResponse        │          │
└──────────┘                      └──────────┘

Auth Service 内部：
  1. Auth.Login() 收到请求
  2. Auth 调用 User.GetUserByUsername() ← gRPC 跨服务调用
  3. User Service 查 DB，返回用户数据
  4. Auth 校验密码，生成 JWT
  5. Auth 返回 Token 给调用方
```

### 5.3 动手练习

| # | 练习 | 目标 |
|---|------|------|
| 1 | 实现 Login + VerifyToken | 理解服务间调用 |
| 2 | 实现 RefreshToken + Logout | 理解 Redis 会话管理 |
| 3 | 实现登录失败锁定逻辑 | 理解 Redis 计数器 |
| 4 | 同时启动 Auth + User，grpcurl 测试 Login | 端到端理解调用链 |

---

## 6. 权限微服务 (Permission Service)

### 6.1 Proto 定义

```protobuf
// api/proto/permission/v1/permission.proto
syntax = "proto3";
package permission.v1;
option go_package = "helloGo/gen/go/permission/v1;permissionv1";

import "common/v1/common.proto";

service PermissionService {
  // --- 角色 ---
  rpc ListRoles (ListRolesRequest) returns (ListRolesResponse);
  rpc CreateRole (CreateRoleRequest) returns (RoleResponse);
  rpc AddPermissionToRole (AddPermissionToRoleRequest) returns (RoleResponse);
  // --- 权限 ---
  rpc ListPermissions (ListPermissionsRequest) returns (ListPermissionsResponse);
  rpc CreatePermission (CreatePermissionRequest) returns (PermissionMsg);
  rpc UpdatePermission (UpdatePermissionRequest) returns (PermissionMsg);
  rpc DeletePermission (DeletePermissionRequest) returns (common.v1.Empty);
  // --- 权限校验（供 Gateway 调用）---
  rpc CheckPermission (CheckPermissionRequest) returns (CheckPermissionResponse);
  rpc GetUserPermissions (GetUserPermissionsRequest) returns (GetUserPermissionsResponse);
  // --- 菜单 ---
  rpc ListMenus (ListMenusRequest) returns (ListMenusResponse);
}

// ... 消息定义省略，结构与 User proto 类似 ...

message CheckPermissionRequest {
  repeated string role_codes = 1;
  string required_permission = 2;
}

message CheckPermissionResponse {
  bool allowed = 1;
}

message GetUserPermissionsRequest {
  repeated string role_codes = 1;
}

message GetUserPermissionsResponse {
  repeated string permissions = 1;  // 所有权限 key 列表
}
```

### 6.2 权限缓存策略

```go
// internal/permission/service.go — 权限缓存

// GetUserPermissions 获取用户权限（优先读缓存）
func (s *Service) GetUserPermissions(ctx context.Context, req *permissionv1.GetUserPermissionsRequest) (*permissionv1.GetUserPermissionsResponse, error) {
    var allPerms []string

    for _, roleCode := range req.RoleCodes {
        cacheKey := "rolePerms:" + roleCode

        // 1. 尝试从 Redis 读取
        cached, err := s.redis.Get(ctx, cacheKey).Result()
        if err == nil {
            // 缓存命中，反序列化
            var perms []string
            json.Unmarshal([]byte(cached), &perms)
            allPerms = append(allPerms, perms...)
            continue
        }

        // 2. 缓存未命中，查数据库
        perms, err := s.repo.GetPermissionsByRoleCode(roleCode)
        if err != nil {
            continue
        }

        // 3. 写入缓存（TTL 300s）
        permKeys := make([]string, len(perms))
        for i, p := range perms {
            permKeys[i] = p.Key
        }
        data, _ := json.Marshal(permKeys)
        s.redis.Set(ctx, cacheKey, string(data), 300*time.Second)

        allPerms = append(allPerms, permKeys...)
    }

    return &permissionv1.GetUserPermissionsResponse{
        Permissions: unique(allPerms),
    }, nil
}
```

---

## 7. 业务微服务 (Biz Service)

### 7.1 Proto 定义

```protobuf
// api/proto/biz/v1/biz.proto
syntax = "proto3";
package biz.v1;
option go_package = "helloGo/gen/go/biz/v1;bizv1";

import "common/v1/common.proto";

service BizService {
  // --- 部门 ---
  rpc ListDepartments (ListDepartmentsRequest) returns (ListDepartmentsResponse);
  rpc CreateDepartment (CreateDepartmentRequest) returns (Department);
  // --- 字典 ---
  rpc ListDicts (ListDictsRequest) returns (ListDictsResponse);
  rpc CreateDict (CreateDictRequest) returns (Dict);
  // --- 日志 ---
  rpc ListLogs (ListLogsRequest) returns (ListLogsResponse);
  rpc CreateLog (CreateLogRequest) returns (Log);
  // --- 上传 ---
  rpc ListUploads (ListUploadsRequest) returns (ListUploadsResponse);
  // --- 健康 ---
  rpc HealthCheck (common.v1.Empty) returns (HealthCheckResponse);
}

// --- 部门消息 ---
message Department {
  string id = 1;
  string name = 2;
  string description = 3;
  string parent_id = 4;
  repeated Department children = 5;
}

// --- 字典消息 ---
message Dict {
  string id = 1;
  string type = 2;
  string key = 3;
  string value = 4;
  string description = 5;
}

// --- 日志消息 ---
message Log {
  string id = 1;
  string level = 2;
  string message = 3;
  string meta = 4;
  string created_at = 5;
}

message HealthCheckResponse {
  string status = 1;
  string service = 2;
  string version = 3;
}

// ... 请求/响应消息省略，与 User 类似 ...
```

### 7.2 日志收集（其他服务调用 Biz 写日志）

```
Audit 日志流向：

  Gateway  ──→  各 gRPC 服务（处理请求）
                    │
                    ├── 拦截器中收集请求信息
                    │
                    └── gRPC 调用 Biz.CreateLog()
                              │
                              └── 写入 logs 表
```

---

## 8. API Gateway 网关层

### 8.1 设计理念

```
浏览器/App 只与 Gateway 通信（HTTP/JSON）
Gateway 负责：
  1. 接收 HTTP 请求
  2. JWT 验证（调用 Auth.VerifyToken）
  3. 权限检查（调用 Permission.CheckPermission）
  4. 将请求转换为 gRPC 调用对应的微服务
  5. 将 gRPC 响应转换为 JSON 返回给客户端

┌────────┐  HTTP/JSON   ┌──────────┐  gRPC   ┌──────────┐
│ Client │ ───────────→ │ Gateway  │ ──────→ │ Service  │
│        │ ←─────────── │          │ ←────── │          │
└────────┘              └──────────┘         └──────────┘
```

### 8.2 Gateway 连接所有微服务

```go
// cmd/gateway/main.go
package main

import (
    "log"

    "helloGo/internal/gateway"
    "helloGo/internal/shared/config"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    cfg := config.Load("gateway")

    // 连接到所有微服务
    connections := map[string]*grpc.ClientConn{}
    services := map[string]string{
        "user":       cfg.Services.UserAddr,       // "localhost:50001"
        "auth":       cfg.Services.AuthAddr,       // "localhost:50002"
        "permission": cfg.Services.PermissionAddr, // "localhost:50003"
        "biz":        cfg.Services.BizAddr,        // "localhost:50004"
    }

    for name, addr := range services {
        conn, err := grpc.NewClient(addr,
            grpc.WithTransportCredentials(insecure.NewCredentials()),
        )
        if err != nil {
            log.Fatalf("连接 %s 服务失败: %v", name, err)
        }
        connections[name] = conn
        log.Printf("✅ 已连接 %s 服务 (%s)", name, addr)
    }

    // 启动 Gateway HTTP 服务器
    gw := gateway.New(cfg.HTTPPort, connections)
    gw.Start()
}
```

### 8.3 HTTP Handler 示例（HTTP → gRPC 转发）

```go
// internal/gateway/handler/user_handler.go
package handler

import (
    userv1 "helloGo/gen/go/user/v1"
    "github.com/gofiber/fiber/v2"
)

type UserHandler struct {
    client userv1.UserServiceClient
}

func NewUserHandler(conn *grpc.ClientConn) *UserHandler {
    return &UserHandler{
        client: userv1.NewUserServiceClient(conn),
    }
}

// ListUsers 处理 GET /api/users
// HTTP 请求 → gRPC 调用 → JSON 响应
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
    // 1. 解析 HTTP 查询参数
    page, _ := strconv.Atoi(c.Query("page", "1"))
    limit, _ := strconv.Atoi(c.Query("limit", "10"))
    keyword := c.Query("keyword")

    // 2. 构造 gRPC 请求
    resp, err := h.client.ListUsers(c.Context(), &userv1.ListUsersRequest{
        Pagination: &commonv1.PaginationRequest{
            Page:  int32(page),
            Limit: int32(limit),
        },
        Keyword: keyword,
    })
    if err != nil {
        return handleGRPCError(c, err)  // gRPC 错误 → HTTP 错误
    }

    // 3. 返回 JSON 响应
    return c.JSON(fiber.Map{
        "code":    "OK",
        "data":    resp.Users,
        "pagination": resp.Pagination,
    })
}

// CreateUser 处理 POST /api/users
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
    // 1. 解析 HTTP Body
    var body struct {
        Username  string   `json:"username"`
        Password  string   `json:"password"`
        Email     string   `json:"email"`
        Phone     string   `json:"phone"`
        RoleCodes []string `json:"role_codes"`
    }
    if err := c.BodyParser(&body); err != nil {
        return c.Status(400).JSON(fiber.Map{"code": "VALIDATION_ERROR"})
    }

    // 2. gRPC 调用
    resp, err := h.client.CreateUser(c.Context(), &userv1.CreateUserRequest{
        Username:  body.Username,
        Password:  body.Password,
        Email:     body.Email,
        Phone:     body.Phone,
        RoleCodes: body.RoleCodes,
    })
    if err != nil {
        return handleGRPCError(c, err)
    }

    return c.Status(201).JSON(fiber.Map{
        "code": "OK",
        "data": resp.User,
    })
}
```

### 8.4 JWT 验证中间件（网关层统一处理）

```go
// internal/gateway/middleware/jwt.go
package middleware

import (
    authv1 "helloGo/gen/go/auth/v1"
    "github.com/gofiber/fiber/v2"
)

// JWTMiddleware 在网关层验证 JWT
// 每个请求都调用 Auth.VerifyToken 验证令牌
func JWTMiddleware(authClient authv1.AuthServiceClient) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // 从 Header 提取 Token
        token := extractBearerToken(c.Get("Authorization"))
        if token == "" {
            return c.Status(401).JSON(fiber.Map{
                "code": "AUTH_UNAUTHORIZED",
                "message": "缺少认证令牌",
            })
        }

        // 调用 Auth Service 验证 Token（gRPC 调用）
        resp, err := authClient.VerifyToken(c.Context(), &authv1.VerifyTokenRequest{
            Token: token,
        })
        if err != nil {
            return c.Status(401).JSON(fiber.Map{
                "code": "AUTH_UNAUTHORIZED",
                "message": "令牌无效或已过期",
            })
        }

        // 将用户信息存入 Fiber Context，后续 handler 可使用
        c.Locals("userID", resp.UserId)
        c.Locals("username", resp.Username)
        c.Locals("roles", resp.Roles)

        return c.Next()
    }
}
```

### 8.5 gRPC 错误 → HTTP 错误映射

```go
// internal/gateway/handler/errors.go
package handler

import (
    "github.com/gofiber/fiber/v2"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// handleGRPCError 将 gRPC 状态码映射为 HTTP 状态码
func handleGRPCError(c *fiber.Ctx, err error) error {
    st, ok := status.FromError(err)
    if !ok {
        return c.Status(500).JSON(fiber.Map{
            "code":    "INTERNAL_ERROR",
            "message": "服务内部错误",
        })
    }

    httpStatus := 500
    switch st.Code() {
    case codes.OK:
        httpStatus = 200
    case codes.InvalidArgument:
        httpStatus = 400
    case codes.NotFound:
        httpStatus = 404
    case codes.AlreadyExists:
        httpStatus = 409
    case codes.PermissionDenied:
        httpStatus = 403
    case codes.Unauthenticated:
        httpStatus = 401
    case codes.ResourceExhausted:
        httpStatus = 429
    case codes.Unavailable:
        httpStatus = 503
    }

    return c.Status(httpStatus).JSON(fiber.Map{
        "code":    st.Code().String(),
        "message": st.Message(),
    })
}
```

### 8.6 路由注册

```go
// internal/gateway/router.go
package gateway

import (
    "helloGo/internal/gateway/handler"
    "helloGo/internal/gateway/middleware"

    authv1 "helloGo/gen/go/auth/v1"
    userv1 "helloGo/gen/go/user/v1"
    permissionv1 "helloGo/gen/go/permission/v1"
    bizv1 "helloGo/gen/go/biz/v1"

    "github.com/gofiber/fiber/v2"
    "google.golang.org/grpc"
)

func setupRoutes(app *fiber.App, conns map[string]*grpc.ClientConn) {
    // 创建各服务的 gRPC 客户端
    authClient := authv1.NewAuthServiceClient(conns["auth"])
    userClient := userv1.NewUserServiceClient(conns["user"])
    permClient := permissionv1.NewPermissionServiceClient(conns["permission"])
    bizClient := bizv1.NewBizServiceClient(conns["biz"])

    // 创建 Handler
    authHandler := handler.NewAuthHandler(authClient)
    userHandler := handler.NewUserHandler(conns["user"])
    permHandler := handler.NewPermHandler(conns["permission"])
    bizHandler := handler.NewBizHandler(conns["biz"])

    // ---- Public 路由（不需要 JWT）----
    public := app.Group("/api")
    public.Post("/auth/login", authHandler.Login)
    public.Post("/auth/refresh", authHandler.RefreshToken)
    public.Post("/auth/password/request-reset", authHandler.RequestPasswordReset)
    public.Post("/auth/password/reset", authHandler.ConfirmPasswordReset)
    public.Post("/auth/unlock", authHandler.UnlockAccount)
    public.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "ok"})
    })

    // ---- 需要认证的路由 ----
    api := app.Group("/api", middleware.JWTMiddleware(authClient))
    api.Post("/auth/logout", authHandler.Logout)

    // Users
    api.Get("/users", userHandler.ListUsers)
    api.Get("/users/:id", userHandler.GetUser)
    api.Post("/users", userHandler.CreateUser)
    api.Patch("/users/:id", userHandler.UpdateUser)
    api.Delete("/users/:id", userHandler.DeleteUser)

    // Roles & Permissions
    api.Get("/roles", permHandler.ListRoles)
    api.Post("/roles", permHandler.CreateRole)
    api.Get("/permissions", permHandler.ListPermissions)
    api.Post("/permissions", permHandler.CreatePermission)

    // Menus / Departments / Dicts / Logs
    api.Get("/menus", bizHandler.ListMenus)
    api.Get("/departments", bizHandler.ListDepartments)
    api.Get("/dicts", bizHandler.ListDicts)
    api.Get("/logs", bizHandler.ListLogs)

    // Uploads
    api.Post("/uploads", bizHandler.Upload)
    api.Get("/uploads", bizHandler.ListUploads)
}
```

---

## 9. 服务治理

### 9.1 服务发现（etcd）

> 当服务有多个实例时，需要服务发现来动态找到可用实例

```
没有服务发现：
  Gateway 硬编码连接 user:50001
  → User 扩容到 3 个实例后，Gateway 不知道其他 2 个的地址

有服务发现（etcd）：
  User 启动时 → 向 etcd 注册自己的地址
  Gateway 连接时 → 从 etcd 查询所有 User 实例地址
  User 实例下线 → etcd 自动摘除（TTL 过期）

┌──────────┐    注册    ┌──────┐    查询    ┌──────────┐
│ User Svc │ ────────→  │ etcd │  ←─────── │ Gateway  │
│ :50001   │            │      │           │          │
└──────────┘            └──────┘           └──────────┘
┌──────────┐    注册    ┌──────┐
│ User Svc │ ────────→  │ etcd │
│ :50002   │            │      │
└──────────┘            └──────┘
```

**服务注册实现：**

```go
// internal/shared/registry/etcd.go
package registry

import (
    "context"
    "fmt"
    "time"

    clientv3 "go.etcd.io/etcd/client/v3"
)

type Registry struct {
    client *clientv3.Client
    ttl    int64 // 租约 TTL（秒）
}

// Register 注册服务到 etcd
func (r *Registry) Register(ctx context.Context, serviceName, addr string) error {
    // 创建带 TTL 的租约
    lease, err := r.client.Grant(ctx, r.ttl)
    if err != nil {
        return err
    }

    // 注册 key: /services/{serviceName}/{addr}
    key := fmt.Sprintf("/services/%s/%s", serviceName, addr)
    _, err = r.client.Put(ctx, key, addr, clientv3.WithLease(lease.ID))
    if err != nil {
        return err
    }

    // 自动续约（防止服务还活着但 key 过期）
    ch, err := r.client.KeepAlive(ctx, lease.ID)
    if err != nil {
        return err
    }

    go func() {
        for range ch {
            // 续约成功，无需处理
        }
    }()

    return nil
}

// Discover 发现服务实例
func (r *Registry) Discover(ctx context.Context, serviceName string) ([]string, error) {
    prefix := fmt.Sprintf("/services/%s/", serviceName)
    resp, err := r.client.Get(ctx, prefix, clientv3.WithPrefix())
    if err != nil {
        return nil, err
    }

    addrs := make([]string, 0, len(resp.Kvs))
    for _, kv := range resp.Kvs {
        addrs = append(addrs, string(kv.Value))
    }
    return addrs, nil
}
```

### 9.2 负载均衡

```go
// 使用 gRPC 内置的 round_robin 负载均衡
// Gateway 连接时使用 "dns:///service-name" 格式

conn, err := grpc.NewClient(
    "dns:///user-service:50001",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithDefaultServiceConfig(`{
        "loadBalancingPolicy": "round_robin"
    }`),
)
```

### 9.3 熔断与重试

```go
// 使用 grpc 重试策略
import "google.golang.org/grpc"

var retryPolicy = `{
    "methodConfig": [{
        "name": [{"service": "user.v1.UserService"}],
        "retryPolicy": {
            "maxAttempts": 3,
            "initialBackoff": "0.1s",
            "maxBackoff": "1s",
            "backoffMultiplier": 2.0,
            "retryableStatusCodes": ["UNAVAILABLE"]
        }
    }]
}`

conn, err := grpc.NewClient(addr,
    grpc.WithDefaultServiceConfig(retryPolicy),
)
```

### 9.4 动手练习

| # | 练习 | 目标 |
|---|------|------|
| 1 | 启动 etcd（Docker），手动注册/发现服务 | 理解服务发现原理 |
| 2 | 启动 2 个 User Service 实例（不同端口） | 理解多实例部署 |
| 3 | Gateway 通过 etcd 发现 User 并负载均衡 | 理解完整链路 |

---

## 10. 可观测性

### 10.1 链路追踪（OpenTelemetry + Jaeger）

> 微服务调用链长了之后，一个请求可能经过 3-4 个服务，出问题很难排查  
> 链路追踪给每个请求分配唯一 TraceID，记录完整的调用链

```
请求调用链可视化：

TraceID: abc-123
├── [Gateway] POST /api/users (50ms)
│   ├── [Auth.VerifyToken] (5ms)
│   ├── [Permission.CheckPermission] (3ms)
│   └── [User.CreateUser] (40ms)
│       └── [DB INSERT] (15ms)
│
│ Total: 50ms
```

**Jaeger 启动（Docker）：**

```yaml
# docker-compose.infra.yml 中添加
jaeger:
  image: jaegertracing/all-in-one:1.53
  ports:
    - "16686:16686"   # Jaeger UI
    - "4318:4318"     # OTLP HTTP
```

**在 gRPC 拦截器中注入 TraceID：**

```go
// internal/shared/interceptor/tracing.go
// 使用 OpenTelemetry 的 gRPC 拦截器

import (
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

// 在创建 gRPC Server 时添加
grpcServer := grpc.NewServer(
    grpc.StatsHandler(otelgrpc.NewServerHandler()),
)

// 在创建 gRPC Client 时添加
conn, err := grpc.NewClient(addr,
    grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
)
```

### 10.2 gRPC 指标（Prometheus）

```go
// internal/shared/interceptor/metrics.go

import (
    "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
    "github.com/prometheus/client_golang/prometheus"
)

// 创建 gRPC Prometheus 指标
var grpcMetrics = prometheus.NewServerMetrics(
    prometheus.WithServerHandlingTimeHistogram(),
)

func init() {
    prometheus.MustRegister(grpcMetrics)
}
```

### 10.3 健康检查（gRPC 标准）

```protobuf
// 使用 gRPC 官方健康检查协议
import "google.golang.org/grpc/health/grpc_health_v1";

// 每个微服务都注册健康检查
grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
```

---

## 11. 容器编排与部署

### 11.1 各服务 Dockerfile（统一模板）

```dockerfile
# 每个微服务使用相同的 Dockerfile，通过 build arg 区分
# 例如: docker build --build-arg SERVICE=user -f Dockerfile -t helloGo-user .

ARG SERVICE

FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/${SERVICE}

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /server .
COPY --from=builder /app/configs ./configs
EXPOSE 50001
CMD ["./server"]
```

### 11.2 docker-compose.yml（完整微服务环境）

```yaml
services:
  # ====== 基础设施 ======
  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]

  mysql:
    image: mysql:8
    environment:
      MYSQL_ROOT_PASSWORD: root123
      MYSQL_DATABASE: hellogo
    ports: ["3306:3306"]
    volumes: ["mysql-data:/var/lib/mysql"]

  etcd:
    image: quay.io/coreos/etcd:v3.5.12
    command:
      - etcd
      - --listen-client-urls=http://0.0.0.0:2379
      - --advertise-client-urls=http://etcd:2379
    ports: ["2379:2379"]

  jaeger:
    image: jaegertracing/all-in-one:1.53
    ports:
      - "16686:16686"  # UI
      - "4318:4318"    # OTLP

  # ====== 微服务 ======
  user-service:
    build:
      context: .
      args: { SERVICE: user }
    ports: ["50001:50001"]
    depends_on: [mysql, redis, etcd]
    environment:
      DB_TYPE: mysql
      DB_HOST: mysql
      REDIS_HOST: redis
      ETCD_ENDPOINTS: etcd:2379
      GRPC_PORT: "50001"

  auth-service:
    build:
      context: .
      args: { SERVICE: auth }
    ports: ["50002:50002"]
    depends_on: [redis, etcd, user-service]
    environment:
      REDIS_HOST: redis
      ETCD_ENDPOINTS: etcd:2379
      USER_SERVICE_ADDR: user-service:50001
      GRPC_PORT: "50002"

  permission-service:
    build:
      context: .
      args: { SERVICE: permission }
    ports: ["50003:50003"]
    depends_on: [mysql, redis, etcd]
    environment:
      DB_TYPE: mysql
      DB_HOST: mysql
      REDIS_HOST: redis
      ETCD_ENDPOINTS: etcd:2379
      GRPC_PORT: "50003"

  biz-service:
    build:
      context: .
      args: { SERVICE: biz }
    ports: ["50004:50004"]
    depends_on: [mysql, redis, etcd]
    environment:
      DB_TYPE: mysql
      DB_HOST: mysql
      REDIS_HOST: redis
      ETCD_ENDPOINTS: etcd:2379
      GRPC_PORT: "50004"

  gateway:
    build:
      context: .
      args: { SERVICE: gateway }
    ports: ["8000:8000"]
    depends_on: [user-service, auth-service, permission-service, biz-service]
    environment:
      USER_SERVICE_ADDR: user-service:50001
      AUTH_SERVICE_ADDR: auth-service:50002
      PERMISSION_SERVICE_ADDR: permission-service:50003
      BIZ_SERVICE_ADDR: biz-service:50004
      HTTP_PORT: "8000"

volumes:
  mysql-data:
```

### 11.3 一键启动

```bash
# 构建并启动全部服务
docker compose up --build -d

# 查看状态
docker compose ps

# 查看日志
docker compose logs -f gateway

# 测试
curl http://localhost:8000/api/health
curl -X POST http://localhost:8000/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}'

# 停止全部
docker compose down
```

### 11.4 Makefile 更新

```makefile
# ===== Proto =====
.PHONY: proto
proto:
	bash scripts/gen-proto.sh

# ===== 各服务启动（本地开发） =====
.PHONY: run-gateway
run-gateway:
	go run cmd/gateway/main.go

.PHONY: run-user
run-user:
	go run cmd/user/main.go

.PHONY: run-auth
run-auth:
	go run cmd/auth/main.go

.PHONY: run-permission
run-permission:
	go run cmd/permission/main.go

.PHONY: run-biz
run-biz:
	go run cmd/biz/main.go

# ===== 开发环境（只启动基础设施） =====
.PHONY: infra-up
infra-up:
	docker compose -f docker-compose.infra.yml up -d

.PHONY: infra-down
infra-down:
	docker compose -f docker-compose.infra.yml down

# ===== 全部服务 =====
.PHONY: all-up
all-up:
	docker compose up --build -d

.PHONY: all-down
all-down:
	docker compose down

# ===== 测试 =====
.PHONY: test
test:
	go test ./...

.PHONY: test-grpc
test-grpc:
	bash scripts/test_grpc.sh

# ===== 工具 =====
.PHONY: seed
seed:
	go run cmd/seed/main.go

.PHONY: lint
lint:
	golangci-lint run ./...
```

---

## 12. 阶段总结与时间规划

### 开发时间估算（初学者视角）

| 阶段 | 内容 | 学习目标 | 预计时间 |
|------|------|----------|----------|
| §1 gRPC 基础 | Hello World + 练习题 | 理解 Protobuf / gRPC 通信模式 | 2-3 天 |
| §2 架构设计 | 确定拆分方案 | 理解微服务架构模式 | 1 天 |
| §3 共享基础 | proto 定义 + 拦截器 + 代码生成 | 掌握 protoc 工具链 | 2 天 |
| §4 User Service | 第一个完整 gRPC 服务 | gRPC CRUD 全流程 | 3-4 天 |
| §5 Auth Service | 认证 + 服务间调用 | 理解服务间 gRPC 调用 | 3-4 天 |
| §6 Permission Service | 角色/权限 + 缓存 | 理解权限模型 + Redis 缓存 | 2-3 天 |
| §7 Biz Service | 菜单/部门/字典/日志 | 熟练 gRPC 开发 | 2-3 天 |
| §8 API Gateway | HTTP→gRPC 转发 | 理解网关模式 | 3-4 天 |
| §9 服务治理 | etcd + 负载均衡 + 熔断 | 理解生产级微服务 | 3-4 天 |
| §10 可观测性 | 链路追踪 + 指标 | OpenTelemetry 实践 | 2 天 |
| §11 容器化 | Docker + Compose | 容器编排 | 1-2 天 |
| **合计** | | | **24-33 天** |

### 建议的学习节奏

```
Week 1: gRPC 基础（§1）+ 架构设计（§2）+ 共享层（§3）
         └→ 里程碑：能独立写 .proto 并生成代码

Week 2: User Service（§4）
         └→ 里程碑：grpcurl 能完成用户 CRUD

Week 3: Auth Service（§5）+ Permission Service（§6）
         └→ 里程碑：两个服务能互相调用

Week 4: Biz Service（§7）+ Gateway（§8）
         └→ 里程碑：curl 通过 Gateway 调用所有服务

Week 5: 服务治理（§9）+ 可观测性（§10）+ 容器化（§11）
         └→ 里程碑：docker compose up 一键启动全部
```

### 每完成一个服务的自检清单

- [ ] .proto 文件定义完整，`make proto` 生成代码无错误
- [ ] gRPC Server 能独立启动
- [ ] `grpcurl` 能调用所有 RPC 方法
- [ ] 错误处理正确（返回合适的 gRPC status code）
- [ ] 日志拦截器正常工作
- [ ] 单元测试覆盖 Service 层核心逻辑
- [ ] Dockerfile 构建成功
- [ ] docker-compose 中可以正常启动
