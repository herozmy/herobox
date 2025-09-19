#!/bin/bash

# HeroBox 统一构建脚本
# 用法: ./build.sh [amd64|arm64|test|help]

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目信息
PROJECT_NAME="herobox"
VERSION="1.1"
BUILD_TIME=$(date '+%Y-%m-%d %H:%M:%S')
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 函数：打印彩色信息
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

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}    HeroBox 构建脚本 v${VERSION}${NC}"
    echo -e "${BLUE}================================${NC}"
}

# 函数：显示帮助信息
show_help() {
    print_header
    echo
    echo "用法: ./build.sh [选项]"
    echo
    echo "选项:"
    echo "  (无参数)    编译所有平台版本 (AMD64 + ARM64)"
    echo "  amd64       编译 Linux AMD64 版本"
    echo "  arm64       编译 Linux ARM64 版本"
    echo "  test        编译测试版本到 bin/dev 目录"
    echo "  help        显示此帮助信息"
    echo
    echo "示例:"
    echo "  ./build.sh          # 编译所有平台版本"
    echo "  ./build.sh amd64    # 编译 Linux AMD64 版本"
    echo "  ./build.sh arm64    # 编译 Linux ARM64 版本"
    echo "  ./build.sh test     # 编译测试版本"
    echo
}

# 函数：检查依赖
check_dependencies() {
    print_info "检查构建依赖..."
    
    # 检查环境变量，如果设置了MOCK_BUILD则跳过依赖检查
    if [ "$MOCK_BUILD" = "true" ]; then
        print_warning "模拟构建模式，跳过依赖检查"
        return 0
    fi

# 检查 Go
if ! command -v go &> /dev/null; then
        print_error "未安装 Go 语言环境"
        print_info "请访问 https://golang.org/dl/ 下载安装"
        print_info "或者设置 MOCK_BUILD=true 进行模拟构建测试"
    exit 1
fi

# 检查 Node.js
if ! command -v node &> /dev/null; then
        print_error "未安装 Node.js"
        print_info "请访问 https://nodejs.org/ 下载安装"
        print_info "或者设置 MOCK_BUILD=true 进行模拟构建测试"
    exit 1
fi

# 检查 npm
if ! command -v npm &> /dev/null; then
        print_error "未安装 npm"
        print_info "或者设置 MOCK_BUILD=true 进行模拟构建测试"
    exit 1
fi

    print_success "依赖检查通过"
    
    # 显示版本信息
    echo "  Go:      $(go version | awk '{print $3}')"
    echo "  Node.js: $(node --version)"
    echo "  npm:     $(npm --version)"
}

# 函数：构建前端
build_frontend() {
    print_info "构建前端应用..."
    
    # 模拟构建模式
    if [ "$MOCK_BUILD" = "true" ]; then
        print_warning "模拟前端构建..."
        mkdir -p web/dist
        echo '<!DOCTYPE html><html><head><title>HeroBox Test</title></head><body><h1>HeroBox 测试页面</h1></body></html>' > web/dist/index.html
        print_success "模拟前端构建完成"
        return 0
    fi
    
cd web

    # 检查并安装依赖
if [ ! -d "node_modules" ]; then
        print_info "安装前端依赖..."
    npm install
fi

    # 构建前端
    print_info "编译前端代码..."
npm run build
    
    if [ $? -eq 0 ]; then
        print_success "前端构建成功"
    else
        print_error "前端构建失败"
        exit 1
    fi
    
    cd ..
    
    # 检查构建结果
if [ ! -d "web/dist" ]; then
        print_error "前端构建失败，未找到 dist 目录"
    exit 1
fi
}

# 函数：构建后端
build_backend() {
    local goos=$1
    local goarch=$2
    local output_dir=$3
    local binary_name=$4
    
    print_info "构建后端应用 (${goos}/${goarch})..."
    
    # 模拟构建模式
    if [ "$MOCK_BUILD" = "true" ]; then
        print_warning "模拟后端构建..."
        # 创建一个简单的测试二进制文件
        cat > "${output_dir}/${binary_name}" << 'EOF'
#!/bin/bash
echo "HeroBox 模拟运行 - 版本 ${VERSION:-1.1}"
echo "这是一个模拟构建的测试文件"
echo "访问地址: http://localhost:8080"
echo "按 Ctrl+C 退出"
while true; do
    sleep 30
    echo "$(date): HeroBox 模拟服务运行中..."
done
EOF
        chmod +x "${output_dir}/${binary_name}"
        print_success "模拟后端构建完成 (${goos}/${goarch})"
        return 0
    fi
    
    # 设置构建参数
    local ldflags="-w -s -X 'main.Version=${VERSION}' -X 'main.BuildTime=${BUILD_TIME}' -X 'main.CommitHash=${COMMIT_HASH}'"
    
    # 构建
    if [ "$goos" = "local" ]; then
        # 本地构建
        CGO_ENABLED=0 go build -ldflags="$ldflags" -o "${output_dir}/${binary_name}" main.go
    else
        # 交叉编译
        CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch go build -ldflags="$ldflags" -o "${output_dir}/${binary_name}" main.go
    fi
    
    if [ $? -eq 0 ]; then
        print_success "后端构建成功 (${goos}/${goarch})"
    else
        print_error "后端构建失败"
    exit 1
fi
}

# 函数：创建启动脚本
create_start_script() {
    local output_dir=$1
    local arch_info=$2
    
    print_info "创建启动脚本..."
    
    cat > "${output_dir}/start.sh" << 'EOF'
#!/bin/bash

# HeroBox 启动脚本

echo "=== HeroBox 启动脚本 ==="
echo

# 设置环境变量
export PORT=${PORT:-8080}
export ENVIRONMENT=${ENVIRONMENT:-production}
export LOG_LEVEL=${LOG_LEVEL:-info}

# 服务配置
export MOSDNS_SERVICE_NAME=${MOSDNS_SERVICE_NAME:-mosdns}
export SING_BOX_SERVICE_NAME=${SING_BOX_SERVICE_NAME:-sing-box}

# 配置文件路径 - 使用标准路径
export MOSDNS_CONFIG_PATH=${MOSDNS_CONFIG_PATH:-/etc/mosdns/config.yaml}
export SING_BOX_CONFIG_PATH=${SING_BOX_CONFIG_PATH:-/etc/sing-box/config.json}

# 日志文件路径
export MOSDNS_LOG_PATH=${MOSDNS_LOG_PATH:-/var/log/mosdns.log}
export SING_BOX_LOG_PATH=${SING_BOX_LOG_PATH:-/var/log/sing-box.log}

# 高级配置
export BACKUP_DIR=${BACKUP_DIR:-./backups}
export WEB_DIR=${WEB_DIR:-./web}

# 创建必要目录
mkdir -p "$BACKUP_DIR"

echo "环境配置:"
echo "  端口: $PORT"
echo "  Web目录: $WEB_DIR"
echo "  备份目录: $BACKUP_DIR"
echo

echo "启动 HeroBox 服务..."
echo "访问地址: http://localhost:$PORT"
echo "内网环境，无需登录认证"
echo
echo "按 Ctrl+C 停止服务"
echo

# 启动应用
exec ./herobox
EOF

    chmod +x "${output_dir}/start.sh"
    print_success "启动脚本创建完成"
}

# 函数：创建服务配置文件
create_service_file() {
    local output_dir=$1
    
    print_info "创建 systemd 服务配置..."
    
    cat > "${output_dir}/herobox.service" << EOF
[Unit]
Description=HeroBox - mosdns & sing-box Web Management Console
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=/etc/herobox
ExecStart=/etc/herobox/herobox
ExecReload=/bin/kill -HUP \$MAINPID
KillMode=mixed
TimeoutStopSec=5
PrivateTmp=true
Restart=on-failure
RestartSec=10

# 环境变量
Environment=PORT=8080
Environment=ENVIRONMENT=production
Environment=LOG_LEVEL=info
Environment=WEB_DIR=/etc/herobox/web
Environment=BACKUP_DIR=/etc/herobox/backups

[Install]
WantedBy=multi-user.target
EOF

    print_success "服务配置文件创建完成"
}

# 函数：创建安装脚本
create_install_script() {
    local output_dir=$1
    
    print_info "创建安装脚本..."
    
    cat > "${output_dir}/install.sh" << 'EOF'
#!/bin/bash

# HeroBox 安装脚本

set -e

echo "=== HeroBox 安装脚本 ==="
echo

# 检查权限
if [ "$EUID" -ne 0 ]; then
    echo "请使用 root 权限运行此脚本"
    echo "使用方法: sudo ./install.sh"
    exit 1
fi

# 安装目录
INSTALL_DIR="/etc/herobox"
SERVICE_FILE="/etc/systemd/system/herobox.service"

echo "安装目录: $INSTALL_DIR"
echo

# 创建安装目录
echo "1. 创建安装目录..."
mkdir -p "$INSTALL_DIR"

# 复制文件
echo "2. 复制程序文件..."
cp herobox "$INSTALL_DIR/"
cp -r web "$INSTALL_DIR/"
cp start.sh "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/herobox"
chmod +x "$INSTALL_DIR/start.sh"

# 创建备份目录
mkdir -p "$INSTALL_DIR/backups"

# 安装服务
echo "3. 安装 systemd 服务..."
cp herobox.service "$SERVICE_FILE"
systemctl daemon-reload
systemctl enable herobox

echo "4. 设置防火墙（如果需要）..."
if command -v ufw &> /dev/null; then
    echo "检测到 ufw，建议运行: sudo ufw allow 8080"
elif command -v firewall-cmd &> /dev/null; then
    echo "检测到 firewalld，建议运行:"
    echo "  sudo firewall-cmd --permanent --add-port=8080/tcp"
    echo "  sudo firewall-cmd --reload"
fi

echo
echo "安装完成！"
echo
echo "使用方法:"
echo "  启动服务: sudo systemctl start herobox"
echo "  停止服务: sudo systemctl stop herobox"
echo "  查看状态: sudo systemctl status herobox"
echo "  查看日志: sudo journalctl -u herobox -f"
echo
echo "访问地址: http://YOUR_SERVER_IP:8080"
echo "内网环境，无需登录认证"
EOF

    chmod +x "${output_dir}/install.sh"
    print_success "安装脚本创建完成"
}

# 函数：创建构建信息文件
create_build_info() {
    local output_dir=$1
    local arch_info=$2
    
    print_info "创建构建信息文件..."
    
    cat > "${output_dir}/BUILD_INFO.md" << EOF
# HeroBox 构建信息

## 构建详情

- **版本**: ${VERSION}
- **构建时间**: ${BUILD_TIME}
- **Git提交**: ${COMMIT_HASH}
- **目标架构**: ${arch_info}
- **Go版本**: $(if [ "$MOCK_BUILD" = "true" ]; then echo "模拟构建"; else go version | awk '{print $3}'; fi)

## 文件说明

- \`herobox\` - 主程序二进制文件
- \`web/\` - 前端静态文件目录
- \`start.sh\` - 启动脚本
- \`install.sh\` - 系统安装脚本
- \`herobox.service\` - systemd 服务配置文件

## 使用方法

### 直接运行
\`\`\`bash
./start.sh
\`\`\`

### 系统安装
\`\`\`bash
sudo ./install.sh
sudo systemctl start herobox
\`\`\`

## 访问方式

- 默认端口: 8080
- 访问地址: http://localhost:8080
- 内网环境，无需登录认证

## 配置说明

程序通过环境变量进行配置，主要配置项：

- \`PORT\` - 服务端口 (默认: 8080)
- \`MOSDNS_SERVICE_NAME\` - MosDNS 服务名 (默认: mosdns)
- \`SING_BOX_SERVICE_NAME\` - Sing-Box 服务名 (默认: sing-box)
- \`BACKUP_DIR\` - 备份目录 (默认: ./backups)

详细配置请参考 start.sh 脚本中的环境变量设置。
EOF

    print_success "构建信息文件创建完成"
}

# 函数：打包构建结果
create_package() {
    local output_dir=$1
    local arch_info=$2
    
    print_info "创建发布包..."
    
    local package_name="${PROJECT_NAME}-linux-${arch_info}-v${VERSION}.tar.gz"
    
    # 进入输出目录的父目录进行打包
    local dir_name=$(basename "$output_dir")
    local parent_dir=$(dirname "$output_dir")
    
    cd "$parent_dir"
    tar -czf "$package_name" "$dir_name"
    cd - > /dev/null
    
    # 移动到项目根目录
    mv "${parent_dir}/${package_name}" ./
    
    print_success "发布包创建完成: ${package_name}"
    
    # 显示包信息
    local package_size=$(ls -lh "$package_name" | awk '{print $5}')
    echo "  文件大小: $package_size"
    echo "  包含文件: $(tar -tzf "$package_name" | wc -l) 个"
}

# 函数：Linux AMD64 构建
build_linux_amd64() {
    print_header
    print_info "开始构建 Linux AMD64 版本..."
    
    local output_dir="build/linux-amd64"
    local arch_info="amd64"
    
    # 清理并创建输出目录
    rm -rf "$output_dir"
    mkdir -p "$output_dir/web"
    
    # 检查依赖
    check_dependencies
    
    # 构建前端
    build_frontend
    
    # 构建后端
    build_backend "linux" "amd64" "$output_dir" "herobox"
    
    # 复制前端文件
    print_info "复制前端文件..."
    cp -r web/dist/* "$output_dir/web/"
    
    # 创建配置文件
    create_start_script "$output_dir" "$arch_info"
    create_service_file "$output_dir"
    create_install_script "$output_dir"
    create_build_info "$output_dir" "$arch_info"
    
    # 创建发布包
    create_package "$output_dir" "$arch_info"
    
    print_success "Linux AMD64 构建完成！"
    echo
    print_info "构建输出目录: $output_dir"
    print_info "发布包: ${PROJECT_NAME}-linux-${arch_info}-v${VERSION}.tar.gz"
}

# 函数：Linux ARM64 构建
build_linux_arm64() {
    print_header
    print_info "开始构建 Linux ARM64 版本..."
    
    local output_dir="build/linux-arm64"
    local arch_info="arm64"
    
    # 清理并创建输出目录
    rm -rf "$output_dir"
    mkdir -p "$output_dir/web"
    
    # 检查依赖
    check_dependencies
    
    # 构建前端
    build_frontend
    
    # 构建后端
    build_backend "linux" "arm64" "$output_dir" "herobox"
    
    # 复制前端文件
    print_info "复制前端文件..."
    cp -r web/dist/* "$output_dir/web/"
    
    # 创建配置文件
    create_start_script "$output_dir" "$arch_info"
    create_service_file "$output_dir"
    create_install_script "$output_dir"
    create_build_info "$output_dir" "$arch_info"
    
    # 创建发布包
    create_package "$output_dir" "$arch_info"
    
    print_success "Linux ARM64 构建完成！"
    echo
    print_info "构建输出目录: $output_dir"
    print_info "发布包: ${PROJECT_NAME}-linux-${arch_info}-v${VERSION}.tar.gz"
}

# 函数：构建所有平台版本
build_all_platforms() {
    print_header
    print_info "开始构建所有平台版本..."
    
    echo
    print_info "🚀 第1步: 构建 Linux AMD64 版本"
    echo "======================================="
    build_linux_amd64
    
    echo
    print_info "🚀 第2步: 构建 Linux ARM64 版本"
    echo "======================================="
    build_linux_arm64
    
    echo
    print_success "🎉 所有平台构建完成！"
    echo
    print_info "构建输出:"
    echo "  AMD64 版本: build/linux-amd64/"
    echo "  ARM64 版本: build/linux-arm64/"
    echo "  发布包:"
    ls -la *.tar.gz 2>/dev/null | awk '{print "    " $9 " (" $5 ")"}'
}

# 函数：测试构建
build_test() {
    print_header
    print_info "开始构建测试版本..."
    
    local output_dir="bin/dev"
    local arch_info="amd64-dev"
    
    # 清理并创建输出目录
    rm -rf "$output_dir"
    mkdir -p "$output_dir/web"
    
    # 检查依赖
    check_dependencies
    
    # 构建前端
    build_frontend
    
    # 构建后端（AMD64架构）
    build_backend "linux" "amd64" "$output_dir" "herobox"
    
    # 复制前端文件
    print_info "复制前端文件..."
    cp -r web/dist/* "$output_dir/web/"
    
    # 创建简化的启动脚本（测试用）
    print_info "创建测试启动脚本..."
    cat > "${output_dir}/start.sh" << 'EOF'
#!/bin/bash

# HeroBox 测试启动脚本

echo "=== HeroBox 测试版本 ==="
echo

# 设置环境变量
export PORT=${PORT:-8080}
export ENVIRONMENT=development
export LOG_LEVEL=debug

export MOSDNS_SERVICE_NAME=${MOSDNS_SERVICE_NAME:-mosdns}
export SING_BOX_SERVICE_NAME=${SING_BOX_SERVICE_NAME:-sing-box}

# 使用标准配置路径，HeroBox 安装在 /etc/herobox
export MOSDNS_CONFIG_PATH=${MOSDNS_CONFIG_PATH:-/etc/mosdns/config.yaml}
export SING_BOX_CONFIG_PATH=${SING_BOX_CONFIG_PATH:-/etc/sing-box/config.json}

export MOSDNS_LOG_PATH=${MOSDNS_LOG_PATH:-/var/log/mosdns.log}
export SING_BOX_LOG_PATH=${SING_BOX_LOG_PATH:-/var/log/sing-box.log}

export BACKUP_DIR=./backups
export WEB_DIR=./web

mkdir -p ./backups

echo "安装目录: /etc/herobox/"
echo "配置路径: /etc/mosdns/, /etc/sing-box/"
echo "日志路径: /var/log/"
echo "启动 HeroBox 测试服务..."
echo "访问地址: http://localhost:$PORT"
echo "内网环境，无需登录认证"
echo
echo "按 Ctrl+C 停止服务"
echo

exec ./herobox
EOF
    
    chmod +x "${output_dir}/start.sh"
    
    # 复制配置文件到测试目录
    if [ -f "config.json" ]; then
        cp config.json "$output_dir/"
    fi
    
    create_build_info "$output_dir" "$arch_info"
    
    print_success "测试版本构建完成！"
    echo
    print_info "测试目录: $output_dir"
    print_info "目标平台: Linux AMD64"
    print_info "安装路径: /etc/herobox (默认)"
    print_info "启动方法:"
    echo "  cd $output_dir"
    echo "  ./start.sh"
}

# 主函数
main() {
    case "${1:-all}" in
        "")
            build_all_platforms
            ;;
        "all")
            build_all_platforms
            ;;
        "amd64")
            build_linux_amd64
            ;;
        "arm64")
            build_linux_arm64
            ;;
        "test")
            build_test
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "未知参数: $1"
            echo
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"