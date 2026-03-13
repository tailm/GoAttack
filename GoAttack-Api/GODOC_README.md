# GoAttack 项目文档

本文档介绍如何使用 godoc 查看和生成 GoAttack 项目的代码文档。

## 已添加文档的包

### 1. 主包 (main)
- **文件**: `main.go`
- **功能**: 应用程序入口点，初始化所有系统组件
- **文档内容**:
  - 包描述和系统概述
  - main 函数说明
  - 环境变量配置说明
  - 启动流程说明

### 2. API 包 (api)
- **文件**: `api/router.go`
- **功能**: HTTP 路由配置和中间件设置
- **文档内容**:
  - 包描述和功能概述
  - 路由结构说明
  - 认证机制说明
  - 支持的 API 模块列表
  - SetupRouter 函数详细说明

### 3. 配置包 (common/config)
- **文件**: `common/config/config.go`
- **功能**: 系统配置管理
- **文档内容**:
  - 包描述和配置优先级
  - PostgreSQL 配置变量说明
  - Redis 配置变量说明
  - 辅助函数文档 (getEnv, getEnvInt)

### 4. 日志包 (common/log)
- **文件**: `common/log/log.go`
- **功能**: 日志记录系统
- **文档内容**:
  - 包描述和特性说明
  - 日志级别定义和说明
  - Logger 结构体文档
  - 所有日志函数文档 (InitLogger, Info, Warn, Error, Fatal, Close 等)

### 5. 检测服务 (service/detection)
- **文件**: `service/detection/service.go`
- **功能**: 漏洞检测引擎服务
- **文档内容**:
  - 包描述和功能概述
  - Service 结构体文档
  - NewService 函数文档

### 6. 情报服务 (service/intelligence)
- **文件**: `service/intelligence/service.go`
- **功能**: 漏洞情报收集和管理服务
- **文档内容**:
  - 包描述和功能概述
  - Service 结构体文档
  - NewService 函数文档

## 查看文档

### 方法1: 使用 godoc 命令行工具

```bash
# 查看特定包的文档
godoc GoAttack/api
godoc GoAttack/common/config
godoc GoAttack/common/log
godoc GoAttack/service/detection
godoc GoAttack/service/intelligence

# 查看特定函数的文档
godoc GoAttack/common/log Info
godoc GoAttack/api SetupRouter
```

### 方法2: 启动本地文档服务器

```bash
# 启动 godoc 服务器
godoc -http=:6060

# 然后在浏览器中访问
# http://localhost:6060
```

### 方法3: 使用提供的脚本

```bash
# 运行文档生成脚本
./generate_docs.sh

# 脚本会：
# 1. 启动本地文档服务器 (http://localhost:6060)
# 2. 在终端显示主要包的文档摘要
```

## 文档生成脚本

项目根目录下提供了 `generate_docs.sh` 脚本，可以方便地生成和查看文档：

```bash
# 使脚本可执行
chmod +x generate_docs.sh

# 运行脚本
./generate_docs.sh
```

脚本功能：
- 检查 godoc 是否安装
- 启动本地文档服务器
- 在终端显示各包的文档摘要
- 提供访问链接和说明

## 文档编写规范

### 包文档
每个包的开头应该有包级别的文档注释，格式如下：

```go
// Package [包名] [简要描述]
//
// [详细描述]
//   - 功能1
//   - 功能2
//   - 功能3
//
// [其他说明]
//   - 使用示例
//   - 注意事项
package [包名]
```

### 函数文档
每个导出的函数应该有详细的文档注释，格式如下：

```go
// [函数名] [简要描述]
//
// [详细描述]
// 参数：
//   - param1: 参数1说明
//   - param2: 参数2说明
//
// 返回值：
//   type: 返回值说明
//
// 示例：
//   result := FunctionName(param1, param2)
func FunctionName(param1 type1, param2 type2) returnType {
    // 函数实现
}
```

### 类型和常量文档
导出的类型和常量应该有文档注释：

```go
// [类型名] [简要描述]
//
// [详细描述]
// 字段：
//   - Field1: 字段1说明
//   - Field2: 字段2说明
type TypeName struct {
    Field1 type1
    Field2 type2
}

// [常量名] [简要描述]
const ConstantName = value
```

## 最佳实践

1. **及时更新文档**: 代码修改后，及时更新相关文档
2. **保持一致性**: 使用统一的文档格式和风格
3. **提供示例**: 重要的函数和类型应该提供使用示例
4. **说明参数和返回值**: 详细说明每个参数和返回值的含义
5. **标注注意事项**: 重要的使用限制和注意事项应该明确说明

## 后续工作

建议继续为以下包添加文档：
- `common/postgres`: 数据库操作
- `common/redis`: 缓存操作
- `api/*`: 各个 API 模块
- `model/*`: 数据模型
- `util/*`: 工具函数

通过完善的文档，可以提高代码的可维护性和团队协作效率。