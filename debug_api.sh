#!/bin/bash

echo "调试API接口..."
echo "================"

# 等待服务启动
sleep 3

# 获取token
echo "1. 获取认证token..."
TOKEN=$(curl -s -X POST http://localhost:3000/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Qaz@123#"}' | jq -r '.data.token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "错误: 无法获取token"
    exit 1
fi

echo "Token: ${TOKEN:0:50}..."

echo ""
echo "2. 测试漏洞情报列表API（带详细输出）..."
curl -v -X GET "http://localhost:3000/api/vulnerability/intelligence?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN" 2>&1 | grep -A20 "> GET\|< HTTP\|{" | head -30

echo ""
echo "3. 检查数据库连接..."
PGPASSWORD=tailm123 psql -h 127.0.0.1 -p 5432 -U zwj -d goattack -c "\dt" | grep -E "(vulnerability|detection|alert|intelligence)"

echo ""
echo "4. 检查表结构..."
PGPASSWORD=tailm123 psql -h 127.0.0.1 -p 5432 -U zwj -d goattack -c "\d vulnerability_intelligence" | head -20

echo ""
echo "5. 检查数据库初始化数据..."
PGPASSWORD=tailm123 psql -h 127.0.0.1 -p 5432 -U zwj -d goattack -c "SELECT COUNT(*) as count, 'vulnerability_intelligence' as table FROM vulnerability_intelligence UNION ALL SELECT COUNT(*), 'intelligence_source' FROM intelligence_source UNION ALL SELECT COUNT(*), 'detection_rule' FROM detection_rule UNION ALL SELECT COUNT(*), 'alert_config' FROM alert_config;"