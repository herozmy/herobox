# HeroBox 构建指南

## 统一构建脚本说明

HeroBox 现在使用统一的 `build.sh` 脚本来处理所有构建任务，支持多架构编译和测试构建。

## 🚀 快速开始

### 查看帮助
```bash
./build.sh help
```

### 构建生产版本

#### 构建所有平台（推荐）
```bash
./build.sh          # 构建 AMD64 + ARM64 两个平台
```

#### 单独构建特定平台
```bash
./build.sh amd64    # 仅构建 Linux AMD64 版本
./build.sh arm64    # 仅构建 Linux ARM64 版本
```

### 构建测试版本
```bash
./build.sh test     # 构建测试版本到 bin/dev 目录 (AMD64)
```

## 📋 构建输出

### 生产版本构建输出

每次生产构建会创建以下结构：

```
build/
├── linux-amd64/          # AMD64 版本
│   ├── herobox           # 主程序
│   ├── web/              # 前端文件
│   ├── start.sh          # 启动脚本
│   ├── install.sh        # 系统安装脚本
│   ├── herobox.service   # systemd 服务文件
│   └── BUILD_INFO.md     # 构建信息
└── linux-arm64/          # ARM64 版本
    └── ...               # 相同结构
```

同时会生成发布包：
- `herobox-linux-amd64-v1.1.tar.gz`
- `herobox-linux-arm64-v1.1.tar.gz`

### 测试版本构建输出

测试构建输出到 `bin/dev/` 目录：

```
bin/dev/
├── herobox           # 主程序（Linux AMD64）
├── web/              # 前端文件
├── start.sh          # 测试启动脚本
├── config.json       # 配置文件样例
└── BUILD_INFO.md     # 构建信息
```

**路径说明**:
- **HeroBox 安装目录**: `/etc/herobox/` (程序安装位置)
- **服务配置目录**: `/etc/mosdns/`, `/etc/sing-box/` (标准配置路径)
- **日志目录**: `/var/log/` (标准日志路径)

## 🔧 构建脚本特性

### 依赖检查
- 自动检查 Go、Node.js、npm 环境
- 显示版本信息

### 智能构建
- 自动检测并安装前端依赖
- 优化的构建参数（-w -s 减小体积）
- 版本信息注入

### 完整打包
- 自动复制前端静态文件
- 生成启动脚本和服务配置
- 创建系统安装脚本
- 打包为 tar.gz 发布包

### 构建信息
- 记录构建时间、版本、Git 提交
- 生成详细的构建信息文件

## 🛠️ Makefile 集成

构建脚本已与 Makefile 集成：

```bash
# 查看所有可用命令
make help

# 构建命令
make build          # Linux AMD64 版本
make build-arm64    # Linux ARM64 版本
make test-build     # 测试版本

# 开发命令
make dev            # 启动开发环境
make frontend       # 仅构建前端
make backend        # 仅构建后端

# 清理命令
make clean          # 清理构建文件
make test-clean     # 清理测试文件
```

## 📦 部署使用

### 生产部署

1. 解压发布包：
```bash
tar -xzf herobox-linux-amd64-v1.1.tar.gz
cd linux-amd64
```

2. 系统安装（推荐）：
```bash
sudo ./install.sh
sudo systemctl start herobox
```

3. 或直接运行：
```bash
./start.sh
```

### 测试使用

```bash
cd bin/test
./start.sh
```

## 🔍 故障排除

### 构建失败

1. **依赖检查失败**
   - 确保已安装 Go 1.21+
   - 确保已安装 Node.js 18+
   - 确保网络连接正常

2. **前端构建失败**
   - 删除 `web/node_modules` 重新安装
   - 检查 npm 源设置

3. **后端构建失败**
   - 运行 `go mod tidy` 清理依赖
   - 检查 Go 版本兼容性

### 权限问题

确保构建脚本有执行权限：
```bash
chmod +x build.sh
```

## 📝 环境变量

构建脚本支持以下环境变量：

- `VERSION` - 覆盖默认版本号
- `BUILD_TIME` - 覆盖构建时间
- `PROJECT_NAME` - 覆盖项目名称

示例：
```bash
VERSION=2.0 ./build.sh amd64
```

## 🆕 版本说明

### v1.1 新特性

- 统一构建脚本，支持多架构
- 自动化打包和发布包生成
- 完整的系统安装脚本
- 构建信息记录和版本管理
- Makefile 集成和简化

### 与旧版本区别

- 移除了 `build-linux.sh` 和 `build-test.sh`
- 新增多架构支持（AMD64/ARM64）
- 改进了打包和部署流程
- 统一了构建接口和参数
