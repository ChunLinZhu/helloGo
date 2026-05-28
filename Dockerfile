# ============================================================
# helloGo Dockerfile — 多阶段构建
# ============================================================

# ── Stage 1: 编译 ──────────────────────────────────────────
FROM golang:1.26-alpine AS builder

# CGO 编译需要 gcc（mattn/go-sqlite3 依赖）
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制全部源码
COPY . .

# 生成 Swagger 文档（docs/ 已被 .dockerignore 排除）
RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init -g cmd/server/main.go --parseDependency --parseInternal

# 编译二进制文件
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o server ./cmd/server

# ── Stage 2: 运行 ──────────────────────────────────────────
FROM alpine:3.19

# ca-certificates: HTTPS 证书
# tzdata: 时区支持
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为亚洲/上海
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从 builder 复制二进制和配置文件
COPY --from=builder /app/server .
COPY --from=builder /app/configs ./configs

# 创建数据和上传目录
RUN mkdir -p data upload logs

EXPOSE 8000

CMD ["./server"]
