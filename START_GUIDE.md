# GoAttack 启动指南

## 概述

本文档提供了 GoAttack 项目的多种启动方式，包括使用 `start.sh` 脚本进行本地部署。

## 快速开始

### 1. 使用 start.sh 脚本（推荐）

`start.sh` 脚本提供了完整的本地部署、编译和启动功能：

```bash
# 给脚本添加执行权限（首次使用）
chmod +x start.sh

# 查看帮助信息
./start.sh help

# 启动本地开发环境（后端 + 前端）
./start.sh start

# 使用 Docker Compose 启动所有服务
./start.sh docker

# 查看服务状态
./start.sh status

# 停止所有服务
./start.sh stop

# 重启所有服务
./start.sh restart
```

### 2. 手动启动方式

#### 后端服务 (GoAttack-Api)

```bash
cd GoAttack-Api/

# 安装依赖
go mod tidy

# 编译并运行
go build -o go-attack .
./go-attack

# 或者直接运行（开发模式）
go run main.go
```

#### 前端服务 (GoAttack-Admin)

```bash
cd GoAttack-Admin/

# 安装依赖（使用 pnpm 或 npm）
pnpm install  # 或 npm install

# 启动开发服务器
pnpm dev      # 或 npm run dev
```

### 3. Docker 部署

```bash
cd GoAttack-Docker/

# 构建并启动所有服务
docker-compose up -d --build

# 查看服务状态
docker-compose ps

# 停止服务
docker-compose down
```

## 环境要求

### 本地开发环境
- **Go**: 1.20+ (后端)
- **Node.js**: 14+ (前端)
- **MySQL**: 8.0+ (数据库)
- **Redis**: (缓存，可选但推荐)
- **npm/pnpm**: (包管理器)

### Docker 环境
- **Docker**: 20.10+
- **Docker Compose**: 2.0+

## 服务端口

| 服务 | 本地开发端口 | Docker 端口 | 说明 |
|------|-------------|-------------|------|
| 前端 | 5173 | 80 | Vue.js 开发服务器 |
| 后端API | 3000 | 3000 | Go Gin 服务器 |
| MySQL | 3306 | 3306 | 数据库 |
| Redis | 6379 | 6379 | 缓存服务 |

## 默认登录凭证

- **用户名**: `admin`
- **密码**: `Qaz@123#`

## 脚本功能详解

### start.sh 脚本功能

#### 1. 环境检查
- 检查 Go、Node.js、npm/pnpm 等依赖
- 检查 MySQL 和 Redis 是否安装
- 检查端口占用情况

#### 2. 目录管理
- 自动创建 `logs/` 目录存放日志文件
- 自动创建 `pids/` 目录存放进程ID文件

#### 3. 服务管理
- **后端服务**: 编译 Go 代码并启动在端口 3000
- **前端服务**: 安装依赖并启动开发服务器在端口 5173
- **健康检查**: 自动检查服务是否成功启动

#### 4. Docker 支持
- 一键启动所有服务（API、前端、MySQL、Redis）
- 自动构建镜像
- 健康检查和服务状态监控

#### 5. 进程管理
- 记录进程ID便于管理
- 支持优雅停止服务
- 支持服务状态查询

## 故障排除

### 常见问题

1. **端口被占用**
   ```bash
   # 查看端口占用
   lsof -i :3000
   lsof -i :5173
   
   # 停止占用端口的进程
   kill -9 <PID>
   ```

2. **依赖安装失败**
   ```bash
   # 清理 Go 模块缓存
   cd GoAttack-Api && go clean -modcache
   
   # 清理 Node.js 依赖
   cd GoAttack-Admin && rm -rf node_modules package-lock.json
   ```

3. **数据库连接问题**
   - 确保 MySQL 服务正在运行
   - 检查数据库配置（默认：root/goattack）
   - 查看后端日志获取详细错误信息

4. **Docker 启动失败**
   ```bash
   # 查看 Docker 日志
   docker-compose logs
   
   # 重新构建镜像
   docker-compose build --no-cache
   ```

### 日志文件

- **后端日志**: `logs/api.log`
- **前端日志**: `logs/admin.log`
- **实时查看日志**: `tail -f logs/api.log`

## 高级配置

### 自定义端口

如果需要修改默认端口，可以编辑以下文件：

1. **后端端口** (`GoAttack-Api/main.go`):
   ```go
   r.Run(":3000")  // 修改这里的端口号
   ```

2. **前端开发端口** (`GoAttack-Admin/package.json`):
   ```json
   "scripts": {
     "dev": "vite --port 5173 --config ./config/vite.config.dev.ts"
   }
   ```

3. **Docker 端口** (`GoAttack-Docker/docker-compose.yml`):
   ```yaml
   ports:
     - "3000:3000"  # 后端
     - "80:80"      # 前端
     - "3306:3306"  # MySQL
     - "6379:6379"  # Redis
   ```

### 环境变量

后端服务支持以下环境变量：

```bash
# MySQL 配置
export MYSQL_HOST=localhost
export MYSQL_PORT=3306
export MYSQL_USER=root
export MYSQL_PASSWORD=goattack
export MYSQL_DB=goattack

# Redis 配置
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=
export REDIS_DB=0
```

## 安全建议

1. **修改默认密码**: 首次使用后立即修改默认密码
2. **限制访问**: 生产环境应配置防火墙规则
3. **定期更新**: 保持依赖包和系统更新
4. **备份数据**: 定期备份数据库重要数据

## 技术支持

如遇到问题，请：

1. 查看相关日志文件
2. 检查服务状态：`./start.sh status`
3. 确保所有依赖已正确安装
4. 参考项目 README.md 文档

---

**注意**: 本工具仅用于合法的安全测试和研究目的。使用前请确保您有合法的授权。