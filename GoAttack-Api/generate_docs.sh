#!/bin/bash

# GoAttack 项目文档生成脚本
# 该脚本用于生成项目的 godoc 文档

set -e

echo "=== GoAttack 项目文档生成 ==="
echo

# 检查是否安装了 godoc
if ! command -v godoc &> /dev/null; then
    echo "错误: godoc 未安装"
    echo "请安装 Go 工具链: https://golang.org/dl/"
    exit 1
fi

# 设置 GOPROXY 为国内镜像（可选）
export GOPROXY=https://goproxy.cn,direct

echo "1. 生成 HTML 文档..."
echo "   文档将在 http://localhost:6060 提供"
echo "   按 Ctrl+C 停止服务"
echo
echo "   访问以下链接查看文档："
echo "   - 主包: http://localhost:6060/pkg/GoAttack/"
echo "   - API 包: http://localhost:6060/pkg/GoAttack/api/"
echo "   - 配置包: http://localhost:6060/pkg/GoAttack/common/config/"
echo "   - 日志包: http://localhost:6060/pkg/GoAttack/common/log/"
echo "   - 检测服务: http://localhost:6060/pkg/GoAttack/service/detection/"
echo "   - 情报服务: http://localhost:6060/pkg/GoAttack/service/intelligence/"
echo

# 启动 godoc 服务器
godoc -http=:6060

echo
echo "2. 生成命令行文档..."
echo
echo "=== 主包文档 ==="
echo
godoc GoAttack 2>/dev/null || echo "主包文档生成失败"

echo
echo "=== API 包文档 ==="
echo
godoc GoAttack/api 2>/dev/null || echo "API 包文档生成失败"

echo
echo "=== 配置包文档 ==="
echo
godoc GoAttack/common/config 2>/dev/null || echo "配置包文档生成失败"

echo
echo "=== 日志包文档 ==="
echo
godoc GoAttack/common/log 2>/dev/null || echo "日志包文档生成失败"

echo
echo "=== 检测服务文档 ==="
echo
godoc GoAttack/service/detection 2>/dev/null || echo "检测服务文档生成失败"

echo
echo "=== 情报服务文档 ==="
echo
godoc GoAttack/service/intelligence 2>/dev/null || echo "情报服务文档生成失败"

echo
echo "=== 文档生成完成 ==="
echo
echo "使用说明："
echo "1. 运行 ./generate_docs.sh 启动文档服务器"
echo "2. 在浏览器中访问 http://localhost:6060"
echo "3. 按 Ctrl+C 停止服务器"
echo
echo "快速查看特定包的文档："
echo "  godoc GoAttack/api"
echo "  godoc GoAttack/common/config"
echo "  godoc GoAttack/service/detection"