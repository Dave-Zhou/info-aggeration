#!/bin/bash

# Web服务器启动脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    # 检查Go是否安装
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装，请先安装 Go"
        exit 1
    fi
    
    log_info "Go 版本: $(go version)"
    
    # 检查MySQL是否运行
    if ! pgrep -x "mysqld" > /dev/null; then
        log_warn "MySQL 服务未运行，请启动 MySQL 服务"
    fi
}

# 初始化数据库
init_database() {
    log_info "初始化数据库..."
    
    # 这里可以添加数据库初始化脚本
    # mysql -u root -p < scripts/init.sql
    
    log_info "数据库初始化完成"
}

# 构建前端
build_frontend() {
    log_info "构建前端..."
    
    cd web
    
    # 检查Node.js
    if command -v npm &> /dev/null; then
        log_info "安装前端依赖..."
        npm install
        
        log_info "构建前端..."
        npm run build
    else
        log_warn "Node.js 未安装，跳过前端构建"
        log_warn "请手动安装 Node.js 并运行 'npm install && npm run build'"
    fi
    
    cd ..
}

# 构建后端
build_backend() {
    log_info "构建后端..."
    
    # 下载依赖
    go mod download
    
    # 构建Web服务器
    go build -o bin/webserver ./cmd/webserver
    
    # 构建爬虫
    go build -o bin/crawler ./cmd/crawler
    
    log_info "后端构建完成"
}

# 启动服务
start_services() {
    log_info "启动服务..."
    
    # 创建必要的目录
    mkdir -p data/logs
    mkdir -p data/output
    
    # 启动Web服务器
    log_info "启动Web服务器..."
    ./bin/webserver --config config/config.yaml &
    WEB_PID=$!
    
    log_info "Web服务器已启动，PID: $WEB_PID"
    log_info "访问地址: http://localhost:8080"
    
    # 保存PID
    echo $WEB_PID > data/webserver.pid
    
    # 等待用户中断
    trap 'kill $WEB_PID; exit' INT
    wait $WEB_PID
}

# 主函数
main() {
    log_info "启动爬虫Web管理系统..."
    
    check_dependencies
    init_database
    build_frontend
    build_backend
    start_services
}

# 执行主函数
main "$@" 