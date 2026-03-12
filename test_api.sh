#!/bin/bash

echo "测试漏洞情报收集与检测系统API接口"
echo "======================================"

# 获取token
echo "1. 获取认证token..."
TOKEN=$(curl -s -X POST http://localhost:3000/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Qaz@123#"}' | jq -r '.data.token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "错误: 无法获取token"
    exit 1
fi

echo "Token获取成功: ${TOKEN:0:50}..."

echo ""
echo "2. 测试漏洞情报列表API..."
curl -s -X GET "http://localhost:3000/api/vulnerability/intelligence?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

echo ""
echo "3. 测试漏洞情报统计API..."
curl -s -X GET "http://localhost:3000/api/vulnerability/intelligence/stats" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

echo ""
echo "4. 测试检测任务列表API..."
curl -s -X GET "http://localhost:3000/api/detection/tasks?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

echo ""
echo "5. 测试预警列表API..."
curl -s -X GET "http://localhost:3000/api/alerts?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

echo ""
echo "6. 测试漏洞情报源配置API..."
curl -s -X GET "http://localhost:3000/api/config/intelligence-sources" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

echo ""
echo "7. 测试检测规则配置API..."
curl -s -X GET "http://localhost:3000/api/config/detection-rules" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

echo ""
echo "8. 测试预警配置API..."
curl -s -X GET "http://localhost:3000/api/config/alert-configs" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

echo ""
echo "9. 测试系统健康检查API..."
curl -s -X GET "http://localhost:3000/api/stats/health" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

echo ""
echo "API测试完成！"