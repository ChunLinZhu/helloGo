#!/bin/bash
# Proto 代码生成脚本
# 遍历 api/proto/ 下所有 .proto 文件，生成 Go 代码到 gen/go/
set -e

PROTO_DIR="api/proto"
OUT_DIR="gen/go"

# 确保 protoc 和插件在 PATH 中
export PATH="$HOME/.local/bin:$HOME/go/bin:/usr/local/go/bin:$PATH"

echo "=== Proto 代码生成 ==="
echo "  protoc: $(which protoc)"
echo "  protoc-gen-go: $(which protoc-gen-go)"
echo "  protoc-gen-go-grpc: $(which protoc-gen-go-grpc)"
echo ""

# 清理并重建输出目录
rm -rf ${OUT_DIR}
mkdir -p ${OUT_DIR}

# 编译每个 .proto 文件
find ${PROTO_DIR} -name "*.proto" | sort | while read -r proto_file; do
    echo "  编译: ${proto_file}"
    protoc \
        --proto_path=${PROTO_DIR} \
        --proto_path=$HOME/.local/include \
        --go_out=${OUT_DIR} \
        --go_opt=paths=source_relative \
        --go-grpc_out=${OUT_DIR} \
        --go-grpc_opt=paths=source_relative \
        "${proto_file}"
done

echo ""
echo "=== 生成完成 ==="
echo "输出目录: ${OUT_DIR}/"
find ${OUT_DIR} -name "*.go" | sort | while read -r f; do
    echo "  $f"
done
