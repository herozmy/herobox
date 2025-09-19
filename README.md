# HeroBox

一个用于管理 mosdns 和 sing-box 服务的现代化 Web 控制台。

## 功能特性

- 🚀 **服务管理**: 启动、停止、重启、重载 mosdns 和 sing-box 服务
- 📊 **系统监控**: 实时查看系统信息、服务状态、资源使用情况
- 📝 **配置管理**: 在线编辑和备份配置文件
- 📋 **日志查看**: 实时查看和过滤服务日志
- 🔐 **安全认证**: JWT 令牌认证和权限控制
- 🎨 **现代界面**: 基于 Vue 3 + Element Plus 的响应式界面

## 技术栈

### 后端
- **Go 1.21+** - 高性能后端服务
- **Gin** - Web 框架
- **JWT** - 身份认证
- **systemd** - 服务管理

### 前端
- **Vue 3** - 渐进式前端框架
- **Element Plus** - UI 组件库
- **Vite** - 现代化构建工具
- **Pinia** - 状态管理

## 安装部署

### 1. 克隆项目

\`\`\`bash
git clone <repository-url>
cd herobox
\`\`\`

### 2. 快速构建

使用统一构建脚本（推荐）：

\`\`\`bash
# 构建 Linux AMD64 版本
./build.sh amd64

# 构建 Linux ARM64 版本  
./build.sh arm64

# 构建测试版本（本地架构）
./build.sh test

# 查看帮助
./build.sh help
\`\`\`

或使用 Makefile：

\`\`\`bash
# 构建 Linux AMD64 版本
make build

# 构建 Linux ARM64 版本
make build-arm64

# 构建测试版本
make test-build
\`\`\`

### 3. 手动构建

如需手动构建：

\`\`\`bash
# 安装依赖
go mod tidy
cd web && npm install

# 构建前端
cd web && npm run build

# 构建后端
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o herobox main.go
\`\`\`

### 4. 环境配置

创建环境变量配置：

\`\`\`bash
export PORT=8080
export JWT_SECRET="your-secret-key"
export MOSDNS_SERVICE_NAME="mosdns"
export SING_BOX_SERVICE_NAME="sing-box"
export MOSDNS_CONFIG_PATH="/etc/mosdns/config.yaml"
export SING_BOX_CONFIG_PATH="/etc/sing-box/config.json"
export MOSDNS_LOG_PATH="/var/log/mosdns.log"
export SING_BOX_LOG_PATH="/var/log/sing-box.log"
\`\`\`

## 使用说明

### 1. 访问控制台

打开浏览器访问: \`http://localhost:8080\`

> **注意**: 内网环境版本，无需登录认证，直接访问即可使用。

### 2. 服务管理

在"服务管理"页面可以：
- 查看服务运行状态
- 启动/停止/重启服务
- 查看服务资源使用情况

### 3. 配置管理

在"配置管理"页面可以：
- 在线编辑配置文件
- 自动创建配置备份
- 配置文件语法检查

### 4. 日志查看

在"日志查看"页面可以：
- 实时查看服务日志
- 过滤日志内容
- 自动刷新日志
- 下载日志文件

## 开发说明

### 后端开发

\`\`\`bash
# 运行开发模式
go run main.go

# 代码格式化
go fmt ./...

# 运行测试
go test ./...
\`\`\`

### 前端开发

\`\`\`bash
cd web

# 开发模式
npm run dev

# 构建生产版本
npm run build

# 预览生产版本
npm run preview
\`\`\`

### 测试编译

为了方便开发和测试，项目提供了专门的测试编译功能：

\`\`\`bash
# 使用 Makefile
make test-build    # 编译测试版本到 bin/test 目录
make test-clean    # 清理测试编译文件

# 或使用构建脚本
./build-test.sh    # 完整的测试编译过程
\`\`\`

测试编译特点：
- 仅编译 Linux AMD64 平台
- 包含完整的前后端代码
- 自动生成启动脚本和说明文档
- 适合在开发环境快速测试

## API 文档

### 服务管理

- \`GET /api/services\` - 获取所有服务状态
- \`GET /api/services/:name\` - 获取单个服务状态
- \`POST /api/services/:name/action\` - 控制服务操作

### 配置管理

- \`GET /api/config/:service\` - 获取配置文件
- \`PUT /api/config/:service\` - 更新配置文件

### 日志查看

- \`GET /api/logs/:service\` - 获取服务日志

### 仪表板

- \`GET /api/dashboard\` - 获取仪表板数据

## 权限要求

HeroBox 需要以下系统权限：

- 读取/写入配置文件权限
- systemctl 服务管理权限
- 读取日志文件权限
- 创建备份目录权限

建议以具有适当权限的用户运行，或配置 sudo 权限。

## 安全说明

1. **网络访问**: 仅在受信任的内网环境中使用
2. **HTTPS**: 生产环境建议使用 HTTPS
3. **防火墙**: 合理配置防火墙规则，限制访问来源
4. **权限控制**: 确保运行用户具有适当的系统权限

## 故障排除

### 常见问题

1. **服务无法启动**
   - 检查 systemd 服务配置
   - 确认服务文件存在
   - 检查权限设置

2. **配置文件无法保存**
   - 检查文件写入权限
   - 确认备份目录可写
   - 检查磁盘空间

3. **日志无法显示**
   - 检查日志文件路径
   - 确认日志文件读取权限
   - 检查日志文件是否存在

## 贡献指南

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

## 更新日志

### v1.0.0
- 初始版本发布
- 基础服务管理功能
- 配置文件管理
- 日志查看功能
- Web 控制台界面
