#!/bin/bash

# Colly 爬虫项目构建脚本
# 用于编译和构建项目

set -e

echo "开始构建 Colly 爬虫项目..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "错误: Go 环境未安装，请先安装 Go 1.18 或更高版本"
    exit 1
fi

# 显示Go版本
echo "Go 版本: $(go version)"

# 检查项目目录
if [ ! -f "go.mod" ]; then
    echo "错误: 未找到 go.mod 文件，请确保在项目根目录运行此脚本"
    exit 1
fi

# 设置环境变量
export CGO_ENABLED=1
export GOOS=$(go env GOOS)
export GOARCH=$(go env GOARCH)

echo "构建目标: ${GOOS}/${GOARCH}"

# 创建必要的目录
echo "创建必要的目录..."
mkdir -p data/output
mkdir -p data/logs
mkdir -p bin

# 下载依赖
echo "下载依赖包..."
go mod download

# 整理依赖
echo "整理依赖..."
go mod tidy

# 检查代码格式
echo "检查代码格式..."
if command -v gofmt &> /dev/null; then
    UNFORMATTED=$(gofmt -l .)
    if [ -n "$UNFORMATTED" ]; then
        echo "警告: 以下文件需要格式化:"
        echo "$UNFORMATTED"
        echo "运行 'gofmt -w .' 来格式化代码"
    fi
fi

# 运行静态检查
echo "运行静态检查..."
if command -v go vet &> /dev/null; then
    go vet ./...
else
    echo "警告: go vet 不可用，跳过静态检查"
fi

# 构建主程序
echo "构建主程序..."
BUILD_TIME=$(date -u +%Y%m%d.%H%M%S)
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION=$(go version | cut -d' ' -f3)

# 设置构建参数
LDFLAGS="-w -s"
LDFLAGS="$LDFLAGS -X main.BuildTime=${BUILD_TIME}"
LDFLAGS="$LDFLAGS -X main.GitCommit=${GIT_COMMIT}"
LDFLAGS="$LDFLAGS -X main.GoVersion=${GO_VERSION}"

# 构建不同平台的二进制文件
echo "构建 Linux 版本..."
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o bin/crawler-linux-amd64 ./cmd/crawler

echo "构建 Windows 版本..."
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o bin/crawler-windows-amd64.exe ./cmd/crawler

echo "构建 macOS 版本..."
GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o bin/crawler-darwin-amd64 ./cmd/crawler

# 构建当前平台版本
echo "构建当前平台版本..."
go build -ldflags "$LDFLAGS" -o bin/crawler ./cmd/crawler

# 运行测试
echo "运行测试..."
if [ -d "tests" ] || find . -name "*_test.go" | grep -q .; then
    go test -v ./...
else
    echo "未找到测试文件，跳过测试"
fi

# 检查构建结果
echo "检查构建结果..."
if [ -f "bin/crawler" ]; then
    echo "✅ 构建成功: bin/crawler"
    echo "文件大小: $(du -h bin/crawler | cut -f1)"
else
    echo "❌ 构建失败"
    exit 1
fi

# 显示所有构建的文件
echo "所有构建文件:"
ls -la bin/

# 创建配置文件（如果不存在）
if [ ! -f "config/config.yaml" ]; then
    echo "创建默认配置文件..."
    # 这里可以创建默认配置文件
fi

# 设置执行权限
chmod +x bin/crawler*

echo "构建完成！"
echo ""
echo "使用方法:"
echo "  ./bin/crawler                    # 运行爬虫"
echo "  ./bin/crawler --help             # 查看帮助"
echo "  ./bin/crawler --config custom.yaml  # 使用自定义配置"
echo ""
echo "构建信息:"
echo "  构建时间: ${BUILD_TIME}"
echo "  Git 提交: ${GIT_COMMIT}"
echo "  Go 版本: ${GO_VERSION}"
echo "  目标平台: ${GOOS}/${GOARCH}" 