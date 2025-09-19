.PHONY: dev build frontend backend clean install help test-build test-clean

# 变量定义
BINARY_NAME=herobox
FRONTEND_DIR=web
BUILD_DIR=herobox-deploy
TEST_DIR=bin/dev

# 默认目标
help:
	@echo "HeroBox 构建工具"
	@echo ""
	@echo "可用命令:"
	@echo "  dev         - 启动开发模式"
	@echo "  build       - 构建 Linux AMD64 版本"
	@echo "  build-arm64 - 构建 Linux ARM64 版本"
	@echo "  frontend    - 仅构建前端"
	@echo "  backend     - 仅构建后端"
	@echo "  test-build  - 构建测试版本"
	@echo "  test-clean  - 清理测试文件"
	@echo "  clean       - 清理构建文件"
	@echo "  install     - 安装依赖"
	@echo "  help        - 显示此帮助信息"
	@echo ""
	@echo "构建脚本用法:"
	@echo "  ./build.sh        - 构建所有平台版本 (AMD64 + ARM64)"
	@echo "  ./build.sh amd64  - 构建 Linux AMD64 版本"
	@echo "  ./build.sh arm64  - 构建 Linux ARM64 版本"
	@echo "  ./build.sh test   - 构建测试版本 (到 bin/dev)"
	@echo "  ./build.sh help   - 显示构建脚本帮助"

# 安装依赖
install:
	@echo "安装后端依赖..."
	go mod tidy
	@echo "安装前端依赖..."
	cd $(FRONTEND_DIR) && npm install

# 开发模式
dev:
	@echo "启动开发模式..."
	@echo "后端将运行在 :8080"
	@echo "前端将运行在 :3000"
	@echo "按 Ctrl+C 停止服务"
	@make -j2 dev-backend dev-frontend

dev-backend:
	go run main.go

dev-frontend:
	cd $(FRONTEND_DIR) && npm run dev

# 构建前端
frontend:
	@echo "构建前端..."
	cd $(FRONTEND_DIR) && npm run build

# 构建后端
backend:
	@echo "构建后端..."
	go mod tidy
	CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BINARY_NAME) main.go

# 完整构建 (Linux AMD64)
build:
	@echo "构建 Linux AMD64 版本..."
	./build.sh amd64

# 构建 ARM64 版本
build-arm64:
	@echo "构建 Linux ARM64 版本..."
	./build.sh arm64

# 清理
clean:
	@echo "清理构建文件..."
	rm -f $(BINARY_NAME)
	rm -rf build/
	rm -rf $(TEST_DIR)
	rm -rf $(FRONTEND_DIR)/dist
	rm -rf *.tar.gz

# 快速启动（仅后端）
run: backend
	@echo "启动 HeroBox..."
	./$(BINARY_NAME)

# 测试
test:
	@echo "运行测试..."
	go test ./...

# 格式化代码
fmt:
	@echo "格式化 Go 代码..."
	go fmt ./...
	@echo "格式化前端代码..."
	cd $(FRONTEND_DIR) && npm run lint --fix 2>/dev/null || true

# 测试编译
test-build:
	@echo "构建测试版本..."
	./build.sh test

# 清理测试编译文件
test-clean:
	@echo "清理测试编译文件..."
	rm -rf $(TEST_DIR)
