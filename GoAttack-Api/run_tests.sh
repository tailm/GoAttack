#!/bin/bash

# GoAttack 项目测试运行脚本
# 该脚本用于运行项目的所有测试

set -e

echo "=== GoAttack 项目测试运行 ==="
echo

# 检查是否安装了 go
if ! command -v go &> /dev/null; then
    echo "错误: Go 未安装"
    echo "请安装 Go: https://golang.org/dl/"
    exit 1
fi

# 设置测试环境变量
export GO111MODULE=on
export CGO_ENABLED=0

echo "1. 运行单元测试..."
echo "   运行所有包的单元测试"
echo

# 运行所有测试
go test ./... -v -count=1

echo
echo "2. 运行特定包的测试..."
echo

# 运行主要包的测试
echo "=== 运行 common/config 测试 ==="
go test ./common/config -v -count=1
echo

echo "=== 运行 common/log 测试 ==="
go test ./common/log -v -count=1
echo

echo "=== 运行 api 测试 ==="
go test ./api -v -count=1
echo

echo "=== 运行 service/intelligence 测试 ==="
go test ./service/intelligence -v -count=1
echo

echo "=== 运行 service/detection 测试 ==="
go test ./service/detection -v -count=1
echo

echo "3. 运行集成测试..."
echo "   运行主包测试"
echo

# 运行主包测试
go test . -v -count=1

echo
echo "4. 生成测试覆盖率报告..."
echo

# 生成测试覆盖率报告
echo "生成 HTML 覆盖率报告..."
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

echo "生成文本覆盖率报告..."
go tool cover -func=coverage.out

echo
echo "5. 运行基准测试..."
echo "   运行所有基准测试"
echo

# 运行基准测试
go test ./... -bench=. -benchmem

echo
echo "=== 测试完成 ==="
echo
echo "测试报告:"
echo "  - 单元测试: 通过 go test ./... 运行"
echo "  - 覆盖率报告: coverage.html (HTML格式)"
echo "  - 覆盖率报告: coverage.out (原始数据)"
echo "  - 基准测试: 通过 go test -bench=. 运行"
echo
echo "快速命令:"
echo "  ./run_tests.sh              # 运行所有测试"
echo "  go test ./common/config     # 运行特定包测试"
echo "  go test -v -count=1         # 运行详细测试"
echo "  go test -bench=. -benchmem  # 运行基准测试"
echo "  go tool cover -html=coverage.out  # 查看覆盖率报告"
echo
echo "测试目录结构:"
echo "  common/config/config_test.go     # 配置包测试"
echo "  common/log/log_test.go           # 日志包测试"
echo "  api/router_test.go               # API路由测试"
echo "  service/intelligence/service_test.go  # 情报服务测试"
echo "  service/detection/service_test.go     # 检测服务测试"
echo "  main_test.go                     # 主包测试"
echo "  test_helpers.go                  # 测试辅助函数"