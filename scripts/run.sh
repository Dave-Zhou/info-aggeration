#!/bin/bash

# Colly 爬虫项目运行脚本
# 用于启动和管理爬虫程序

set -e

# 默认配置
CONFIG_FILE="config/config.yaml"
SITES_CONFIG="config/sites.yaml"
LOG_DIR="data/logs"
OUTPUT_DIR="data/output"
BINARY_PATH="bin/crawler"
PID_FILE="data/crawler.pid"

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

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# 显示帮助信息
show_help() {
    cat << EOF
Colly 爬虫运行脚本

使用方法: 
    $0 [命令] [选项]

命令:
    start       启动爬虫
    stop        停止爬虫
    restart     重启爬虫
    status      查看爬虫状态
    logs        查看日志
    clean       清理数据
    build       构建项目
    help        显示帮助信息

选项:
    -c, --config FILE    指定配置文件 (默认: $CONFIG_FILE)
    -s, --sites FILE     指定站点配置文件 (默认: $SITES_CONFIG)
    -d, --daemon         后台运行模式
    -v, --verbose        详细输出模式
    -t, --test           测试模式（不实际运行）
    --dry-run            干运行模式（显示将要执行的操作）

示例:
    $0 start                           # 启动爬虫
    $0 start -c custom.yaml           # 使用自定义配置启动
    $0 start -d                       # 后台启动
    $0 logs                           # 查看日志
    $0 clean                          # 清理数据
    $0 status                         # 查看状态

EOF
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    # 检查Go环境
    if ! command -v go &> /dev/null; then
        log_error "Go 环境未安装"
        exit 1
    fi
    
    # 检查项目文件
    if [ ! -f "go.mod" ]; then
        log_error "未找到 go.mod 文件，请确保在项目根目录运行此脚本"
        exit 1
    fi
    
    # 检查配置文件
    if [ ! -f "$CONFIG_FILE" ]; then
        log_warn "配置文件 $CONFIG_FILE 不存在，将使用默认配置"
    fi
    
    # 检查二进制文件
    if [ ! -f "$BINARY_PATH" ]; then
        log_warn "二进制文件 $BINARY_PATH 不存在，尝试构建..."
        build_project
    fi
}

# 构建项目
build_project() {
    log_info "构建项目..."
    if [ -f "scripts/build.sh" ]; then
        chmod +x scripts/build.sh
        ./scripts/build.sh
    else
        log_info "使用 go build 构建项目..."
        go build -o "$BINARY_PATH" ./cmd/crawler
    fi
}

# 创建必要的目录
create_directories() {
    log_info "创建必要的目录..."
    mkdir -p "$LOG_DIR"
    mkdir -p "$OUTPUT_DIR"
    mkdir -p "$(dirname "$PID_FILE")"
}

# 启动爬虫
start_crawler() {
    log_info "启动爬虫..."
    
    # 检查是否已经运行
    if [ -f "$PID_FILE" ]; then
        local pid=$(cat "$PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            log_warn "爬虫已经在运行 (PID: $pid)"
            return 0
        else
            log_warn "发现过期的 PID 文件，清理..."
            rm -f "$PID_FILE"
        fi
    fi
    
    # 构建启动命令
    local cmd="./$BINARY_PATH"
    
    if [ -n "$CONFIG_FILE" ] && [ -f "$CONFIG_FILE" ]; then
        cmd="$cmd --config $CONFIG_FILE"
    fi
    
    if [ "$DAEMON_MODE" = "true" ]; then
        log_info "后台模式启动..."
        nohup $cmd > "$LOG_DIR/crawler.out" 2>&1 &
        local pid=$!
        echo $pid > "$PID_FILE"
        log_info "爬虫已启动 (PID: $pid)"
    else
        log_info "前台模式启动..."
        $cmd
    fi
}

# 停止爬虫
stop_crawler() {
    log_info "停止爬虫..."
    
    if [ ! -f "$PID_FILE" ]; then
        log_warn "PID 文件不存在，爬虫可能未运行"
        return 0
    fi
    
    local pid=$(cat "$PID_FILE")
    if kill -0 "$pid" 2>/dev/null; then
        log_info "正在停止爬虫 (PID: $pid)..."
        kill "$pid"
        
        # 等待进程结束
        local count=0
        while kill -0 "$pid" 2>/dev/null && [ $count -lt 10 ]; do
            sleep 1
            count=$((count + 1))
        done
        
        if kill -0 "$pid" 2>/dev/null; then
            log_warn "强制停止爬虫..."
            kill -9 "$pid"
        fi
        
        rm -f "$PID_FILE"
        log_info "爬虫已停止"
    else
        log_warn "爬虫进程不存在，清理 PID 文件..."
        rm -f "$PID_FILE"
    fi
}

# 重启爬虫
restart_crawler() {
    log_info "重启爬虫..."
    stop_crawler
    sleep 2
    start_crawler
}

# 查看爬虫状态
check_status() {
    log_info "检查爬虫状态..."
    
    if [ -f "$PID_FILE" ]; then
        local pid=$(cat "$PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            log_info "爬虫正在运行 (PID: $pid)"
            
            # 显示进程信息
            if command -v ps &> /dev/null; then
                ps -p "$pid" -o pid,ppid,cmd,etime,pcpu,pmem
            fi
            
            # 显示端口占用（如果有）
            if command -v netstat &> /dev/null; then
                netstat -tlnp 2>/dev/null | grep "$pid" || true
            fi
        else
            log_warn "PID 文件存在但进程不在运行"
            rm -f "$PID_FILE"
        fi
    else
        log_info "爬虫未运行"
    fi
}

# 查看日志
view_logs() {
    log_info "查看日志..."
    
    if [ -f "$LOG_DIR/crawler.log" ]; then
        if [ "$FOLLOW_LOGS" = "true" ]; then
            tail -f "$LOG_DIR/crawler.log"
        else
            tail -n 50 "$LOG_DIR/crawler.log"
        fi
    else
        log_warn "日志文件不存在"
    fi
}

# 清理数据
clean_data() {
    log_info "清理数据..."
    
    # 停止爬虫
    if [ -f "$PID_FILE" ]; then
        stop_crawler
    fi
    
    # 清理日志
    if [ -d "$LOG_DIR" ]; then
        log_info "清理日志文件..."
        rm -f "$LOG_DIR"/*.log
        rm -f "$LOG_DIR"/*.out
    fi
    
    # 清理输出数据
    if [ -d "$OUTPUT_DIR" ]; then
        log_info "清理输出数据..."
        rm -f "$OUTPUT_DIR"/*
    fi
    
    # 清理临时文件
    rm -f "$PID_FILE"
    rm -f data/crawler.db
    
    log_info "数据清理完成"
}

# 测试配置
test_config() {
    log_info "测试配置..."
    
    if [ -f "$CONFIG_FILE" ]; then
        log_info "检查配置文件格式..."
        # 这里可以添加YAML格式检查
        if command -v yamllint &> /dev/null; then
            yamllint "$CONFIG_FILE"
        else
            log_warn "yamllint 未安装，跳过配置文件格式检查"
        fi
    fi
    
    if [ -f "$SITES_CONFIG" ]; then
        log_info "检查站点配置文件格式..."
        if command -v yamllint &> /dev/null; then
            yamllint "$SITES_CONFIG"
        fi
    fi
    
    log_info "配置测试完成"
}

# 解析命令行参数
parse_args() {
    COMMAND=""
    DAEMON_MODE="false"
    VERBOSE_MODE="false"
    TEST_MODE="false"
    DRY_RUN="false"
    FOLLOW_LOGS="false"
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            start|stop|restart|status|logs|clean|build|help)
                COMMAND="$1"
                shift
                ;;
            -c|--config)
                CONFIG_FILE="$2"
                shift 2
                ;;
            -s|--sites)
                SITES_CONFIG="$2"
                shift 2
                ;;
            -d|--daemon)
                DAEMON_MODE="true"
                shift
                ;;
            -v|--verbose)
                VERBOSE_MODE="true"
                shift
                ;;
            -t|--test)
                TEST_MODE="true"
                shift
                ;;
            --dry-run)
                DRY_RUN="true"
                shift
                ;;
            -f|--follow)
                FOLLOW_LOGS="true"
                shift
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    if [ -z "$COMMAND" ]; then
        COMMAND="start"
    fi
}

# 主函数
main() {
    parse_args "$@"
    
    if [ "$VERBOSE_MODE" = "true" ]; then
        set -x
    fi
    
    case "$COMMAND" in
        start)
            if [ "$TEST_MODE" = "true" ]; then
                test_config
                exit 0
            fi
            check_dependencies
            create_directories
            start_crawler
            ;;
        stop)
            stop_crawler
            ;;
        restart)
            restart_crawler
            ;;
        status)
            check_status
            ;;
        logs)
            view_logs
            ;;
        clean)
            clean_data
            ;;
        build)
            build_project
            ;;
        help)
            show_help
            ;;
        *)
            log_error "未知命令: $COMMAND"
            show_help
            exit 1
            ;;
    esac
}

# 信号处理
trap 'log_warn "收到中断信号，正在清理..."; stop_crawler; exit 0' INT TERM

# 执行主函数
main "$@" 