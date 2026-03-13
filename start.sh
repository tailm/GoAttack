#!/bin/bash

# GoAttack 本地部署启动脚本
# 功能：编译代码并后台启动前后端服务

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_DIR="$PROJECT_ROOT/GoAttack-Api"
ADMIN_DIR="$PROJECT_ROOT/GoAttack-Admin"
DOCKER_DIR="$PROJECT_ROOT/GoAttack-Docker"

# 日志文件
LOG_DIR="$PROJECT_ROOT/logs"
API_LOG="$LOG_DIR/api.log"
ADMIN_LOG="$LOG_DIR/admin.log"
PID_DIR="$PROJECT_ROOT/pids"
API_PID="$PID_DIR/api.pid"
ADMIN_PID="$PID_DIR/admin.pid"

# 函数：打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 函数：检查命令是否存在
check_command() {
    if ! command -v "$1" &> /dev/null; then
        print_error "命令 '$1' 未找到，请先安装"
        return 1
    fi
    return 0
}

# 函数：检查端口是否被占用
check_port() {
    if lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null 2>&1; then
        print_warning "端口 $1 已被占用"
        return 1
    fi
    return 0
}

# 函数：创建必要的目录
create_directories() {
    mkdir -p "$LOG_DIR"
    mkdir -p "$PID_DIR"
    print_info "创建日志和PID目录"
}

# 函数：检查环境依赖
check_dependencies() {
    print_info "检查环境依赖..."
    
    # 检查Go
    if check_command "go"; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        print_info "Go 版本: $GO_VERSION"
    else
        print_error "请先安装 Go 1.20+"
        exit 1
    fi
    
    # 检查Node.js
    if check_command "node"; then
        NODE_VERSION=$(node --version)
        print_info "Node.js 版本: $NODE_VERSION"
    else
        print_error "请先安装 Node.js"
        exit 1
    fi
    
    # 检查npm/pnpm
    if check_command "pnpm"; then
        print_info "使用 pnpm 作为包管理器"
    elif check_command "npm"; then
        print_info "使用 npm 作为包管理器"
    else
        print_error "请先安装 npm 或 pnpm"
        exit 1
    fi
    
    # 检查PostgreSQL
    if check_command "psql"; then
        print_info "PostgreSQL 已安装"
    else
        print_warning "PostgreSQL 未安装，数据库相关功能可能无法使用"
    fi
    
    # 检查Redis
    if check_command "redis-cli"; then
        print_info "Redis 已安装"
    else
        print_warning "Redis 未安装，缓存相关功能可能无法使用"
    fi
}

# 函数：启动后端服务
start_backend() {
    print_info "启动后端服务..."
    
    cd "$API_DIR"
    
    # 检查端口
    if ! check_port 3000; then
        print_warning "后端端口 3000 已被占用，尝试使用其他端口"
        # 这里可以添加端口切换逻辑
    fi
    
    # 安装Go依赖
    print_info "安装Go依赖..."
    go mod tidy
    
    # 编译Go程序
    print_info "编译Go程序..."
    go build -o go-attack .
    
    # 启动后端服务（后台运行）
    print_info "启动后端服务在端口 3000..."
    nohup ./go-attack > "$API_LOG" 2>&1 &
    echo $! > "$API_PID"
    
    # 等待服务启动
    sleep 3
    
    # 检查服务是否启动成功
    print_info "等待后端服务启动..."
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f http://localhost:3000 > /dev/null 2>&1; then
            print_success "后端服务启动成功 (端口: 3000)"
            return 0
        fi
        
        if [ $attempt -eq 10 ] || [ $attempt -eq 20 ]; then
            print_info "后端服务启动中... (尝试 $attempt/$max_attempts)"
        fi
        
        sleep 1
        attempt=$((attempt + 1))
    done
    
    print_error "后端服务启动失败，请检查日志: $API_LOG"
    tail -20 "$API_LOG"
    return 1
}

# 函数：启动前端服务
start_frontend() {
    print_info "启动前端服务..."
    
    cd "$ADMIN_DIR"
    
    # 检查端口
    if ! check_port 5173; then
        print_warning "前端开发端口 5173 已被占用，尝试使用其他端口"
        # 这里可以添加端口切换逻辑
    fi
    
    # 安装前端依赖
    print_info "安装前端依赖..."
    if command -v pnpm &> /dev/null; then
        pnpm install
    else
        npm install
    fi
    
    # 启动前端开发服务器（后台运行）
    print_info "启动前端开发服务器..."
    if command -v pnpm &> /dev/null; then
        nohup pnpm dev > "$ADMIN_LOG" 2>&1 &
    else
        nohup npm run dev > "$ADMIN_LOG" 2>&1 &
    fi
    echo $! > "$ADMIN_PID"
    
    # 检查服务是否启动成功
    print_info "等待前端服务启动..."
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f http://localhost:5173 > /dev/null 2>&1; then
            print_success "前端服务启动成功 (端口: 5173)"
            return 0
        fi
        
        if [ $attempt -eq 10 ] || [ $attempt -eq 20 ]; then
            print_info "前端服务启动中... (尝试 $attempt/$max_attempts)"
        fi
        
        sleep 1
        attempt=$((attempt + 1))
    done
    
    print_error "前端服务启动失败，请检查日志: $ADMIN_LOG"
    tail -20 "$ADMIN_LOG"
    return 1
}

# 函数：使用Docker启动
start_with_docker() {
    print_info "使用Docker Compose启动..."
    
    if ! check_command "docker"; then
        print_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! check_command "docker-compose"; then
        print_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    cd "$DOCKER_DIR"
    
    # 检查端口冲突
    if ! check_port 3000 || ! check_port 80 || ! check_port 5432 || ! check_port 6379; then
        print_warning "检测到端口冲突，Docker Compose可能无法正常启动"
    fi
    
    # 启动Docker Compose
    print_info "启动Docker Compose服务..."
    docker-compose up -d --build
    
    if [ $? -eq 0 ]; then
        print_success "Docker Compose启动成功"
        print_info "服务访问地址:"
        print_info "  - 前端: http://localhost"
        print_info "  - 后端API: http://localhost:3000"
        print_info "  - PostgreSQL: localhost:5432 (用户: postgres, 密码: goattack, 数据库: goattack)"
        print_info "  - Redis: localhost:6379"
    else
        print_error "Docker Compose启动失败"
        exit 1
    fi
}

# 函数：停止服务
stop_services() {
    print_info "停止所有服务..."
    
    # 停止后端服务
    if [ -f "$API_PID" ]; then
        API_PID_VALUE=$(cat "$API_PID")
        if kill -0 "$API_PID_VALUE" 2>/dev/null; then
            kill "$API_PID_VALUE"
            print_info "后端服务已停止 (PID: $API_PID_VALUE)"
        fi
        rm -f "$API_PID"
    fi
    
    # 停止前端服务
    if [ -f "$ADMIN_PID" ]; then
        ADMIN_PID_VALUE=$(cat "$ADMIN_PID")
        if kill -0 "$ADMIN_PID_VALUE" 2>/dev/null; then
            kill "$ADMIN_PID_VALUE"
            print_info "前端服务已停止 (PID: $ADMIN_PID_VALUE)"
        fi
        rm -f "$ADMIN_PID"
    fi
    
    # 停止Docker服务
    if [ -f "$DOCKER_DIR/docker-compose.yml" ]; then
        cd "$DOCKER_DIR"
        docker-compose down 2>/dev/null && print_info "Docker服务已停止"
    fi
    
    print_success "所有服务已停止"
}

# 函数：查看服务状态
status_services() {
    print_info "服务状态检查..."
    
    echo "=== 本地服务状态 ==="
    
    # 检查后端服务
    if [ -f "$API_PID" ]; then
        API_PID_VALUE=$(cat "$API_PID")
        if kill -0 "$API_PID_VALUE" 2>/dev/null; then
            echo -e "后端服务: ${GREEN}运行中${NC} (PID: $API_PID_VALUE)"
            if curl -s -f http://localhost:3000 > /dev/null 2>&1; then
                echo -e "  端口3000: ${GREEN}可访问${NC}"
            else
                echo -e "  端口3000: ${RED}不可访问${NC}"
            fi
        else
            echo -e "后端服务: ${RED}未运行${NC}"
        fi
    else
        echo -e "后端服务: ${YELLOW}未启动${NC}"
    fi
    
    # 检查前端服务
    if [ -f "$ADMIN_PID" ]; then
        ADMIN_PID_VALUE=$(cat "$ADMIN_PID")
        if kill -0 "$ADMIN_PID_VALUE" 2>/dev/null; then
            echo -e "前端服务: ${GREEN}运行中${NC} (PID: $ADMIN_PID_VALUE)"
            if curl -s -f http://localhost:5173 > /dev/null 2>&1; then
                echo -e "  端口5173: ${GREEN}可访问${NC}"
            else
                echo -e "  端口5173: ${RED}不可访问${NC}"
            fi
        else
            echo -e "前端服务: ${RED}未运行${NC}"
        fi
    else
        echo -e "前端服务: ${YELLOW}未启动${NC}"
    fi
    
    echo ""
    echo "=== Docker服务状态 ==="
    
    # 检查Docker服务
    if command -v docker &> /dev/null && [ -f "$DOCKER_DIR/docker-compose.yml" ]; then
        cd "$DOCKER_DIR"
        if docker-compose ps 2>/dev/null | grep -q "Up"; then
            echo -e "Docker服务: ${GREEN}运行中${NC}"
            docker-compose ps --services | while read service; do
                status=$(docker-compose ps --filter "name=$service" --format "{{.Status}}")
                echo "  - $service: $status"
            done
        else
            echo -e "Docker服务: ${YELLOW}未运行${NC}"
        fi
    else
        echo -e "Docker服务: ${YELLOW}未配置${NC}"
    fi
    
    echo ""
    echo "=== 访问信息 ==="
    echo "前端界面: http://localhost:5173 (本地开发)"
    echo "后端API: http://localhost:3000"
    echo "Docker前端: http://localhost (如果使用Docker部署)"
}

# 函数：显示帮助信息
show_help() {
    echo "GoAttack 部署管理脚本"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  start       启动本地开发环境（先启动后端，再启动前端）"
    echo "  docker      使用Docker Compose启动所有服务"
    echo "  stop        停止所有服务"
    echo "  restart     重启所有服务"
    echo "  status      查看服务状态"
    echo "  help        显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 start    启动本地开发环境"
    echo "  $0 docker   使用Docker启动"
    echo "  $0 status   查看服务状态"
    echo ""
    echo "默认登录凭证:"
    echo "  用户名: admin"
    echo "  密码: Qaz@123#"
}

# 主函数
main() {
    create_directories
    
    case "$1" in
        "start")
            check_dependencies
            start_backend
            start_frontend
            print_success "GoAttack 本地开发环境启动完成！"
            echo ""
            echo "访问地址:"
            echo "  前端: http://localhost:5173"
            echo "  后端API: http://localhost:3000"
            echo ""
            echo "默认登录凭证:"
            echo "  用户名: admin"
            echo "  密码: Qaz@123#"
            echo ""
            echo "查看日志:"
            echo "  后端日志: tail -f $API_LOG"
            echo "  前端日志: tail -f $ADMIN_LOG"
            ;;
        "docker")
            start_with_docker
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            stop_services
            sleep 2
            check_dependencies
            start_backend
            start_frontend
            print_success "服务重启完成"
            ;;
        "status")
            status_services
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            if [ -z "$1" ]; then
                show_help
            else
                print_error "未知命令: $1"
                echo ""
                show_help
                exit 1
            fi
            ;;
    esac
}

# 执行主函数
main "$@"