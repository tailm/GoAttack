# 漏洞情报收集与检测系统API设计文档

## 1. 概述

本文档定义了漏洞情报收集与检测系统的RESTful API接口规范。所有API均使用JSON格式进行数据交换，并遵循统一的响应格式。

### 1.1 基础信息
- **基础URL**: `/api`
- **认证方式**: JWT Token (Bearer Token)
- **响应格式**: JSON
- **编码**: UTF-8

### 1.2 统一响应格式
```json
{
  "code": 20000,
  "msg": "success",
  "data": {}
}
```

### 1.3 状态码说明
| 状态码 | 说明 |
|--------|------|
| 20000 | 成功 |
| 40000 | 客户端错误 |
| 40100 | 未授权 |
| 40300 | 禁止访问 |
| 40400 | 资源不存在 |
| 50000 | 服务器内部错误 |

## 2. 漏洞情报管理API

### 2.1 获取漏洞情报列表
**GET** `/api/vulnerability/intelligence`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| page | integer | 否 | 页码，默认1 |
| pageSize | integer | 否 | 每页数量，默认20 |
| search | string | 否 | 搜索关键词 |
| severity | string | 否 | 严重程度过滤 |
| source | string | 否 | 数据源过滤 |
| cve_id | string | 否 | CVE ID过滤 |
| start_date | string | 否 | 开始日期 (YYYY-MM-DD) |
| end_date | string | 否 | 结束日期 (YYYY-MM-DD) |
| is_0day | boolean | 否 | 是否0day漏洞 |
| exploit_available | boolean | 否 | 是否有公开利用 |
| sort_by | string | 否 | 排序字段 |
| sort_order | string | 否 | 排序顺序 (asc/desc) |

**响应示例**:
```json
{
  "code": 20000,
  "msg": "success",
  "data": {
    "total": 150,
    "page": 1,
    "pageSize": 20,
    "items": [
      {
        "id": 1,
        "cve_id": "CVE-2024-12345",
        "cnvd_id": "CNVD-2024-12345",
        "cnnvd_id": "CNNVD-202404-1234",
        "title": "Apache HTTP Server 远程代码执行漏洞",
        "description": "Apache HTTP Server 2.4.58及之前版本存在远程代码执行漏洞...",
        "severity": "critical",
        "cvss_score": 9.8,
        "cvss_vector": "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H",
        "affected_products": ["Apache HTTP Server"],
        "affected_versions": ["<=2.4.58"],
        "references": [
          "https://httpd.apache.org/security/vulnerabilities_24.html",
          "https://nvd.nist.gov/vuln/detail/CVE-2024-12345"
        ],
        "exploit_available": true,
        "poc_available": true,
        "published_at": "2024-01-15T10:30:00Z",
        "last_modified": "2024-01-16T14:20:00Z",
        "source": "nvd",
        "is_0day": false,
        "created_at": "2024-01-16T08:00:00Z",
        "updated_at": "2024-01-16T08:00:00Z"
      }
    ]
  }
}
```

### 2.2 获取漏洞详情
**GET** `/api/vulnerability/intelligence/:id`

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | 漏洞ID |

**响应示例**:
```json
{
  "code": 20000,
  "msg": "success",
  "data": {
    "id": 1,
    "cve_id": "CVE-2024-12345",
    "cnvd_id": "CNVD-2024-12345",
    "cnnvd_id": "CNNVD-202404-1234",
    "title": "Apache HTTP Server 远程代码执行漏洞",
    "description": "详细漏洞描述...",
    "severity": "critical",
    "cvss_score": 9.8,
    "cvss_vector": "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H",
    "affected_products": ["Apache HTTP Server"],
    "affected_versions": ["<=2.4.58"],
    "references": [
      {
        "url": "https://httpd.apache.org/security/vulnerabilities_24.html",
        "type": "vendor"
      },
      {
        "url": "https://nvd.nist.gov/vuln/detail/CVE-2024-12345",
        "type": "nvd"
      }
    ],
    "exploit_available": true,
    "poc_available": true,
    "published_at": "2024-01-15T10:30:00Z",
    "last_modified": "2024-01-16T14:20:00Z",
    "source": "nvd",
    "is_0day": false,
    "technical_details": {
      "attack_vector": "NETWORK",
      "attack_complexity": "LOW",
      "privileges_required": "NONE",
      "user_interaction": "NONE",
      "scope": "UNCHANGED",
      "confidentiality_impact": "HIGH",
      "integrity_impact": "HIGH",
      "availability_impact": "HIGH"
    },
    "created_at": "2024-01-16T08:00:00Z",
    "updated_at": "2024-01-16T08:00:00Z"
  }
}
```

### 2.3 手动同步漏洞情报
**POST** `/api/vulnerability/intelligence/sync`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| source | string | 否 | 数据源名称，不传则同步所有 |
| force | boolean | 否 | 是否强制同步，默认false |

**请求体**:
```json
{
  "source": "nvd",
  "force": false
}
```

**响应示例**:
```json
{
  "code": 20000,
  "msg": "同步任务已启动",
  "data": {
    "sync_id": "sync_1234567890",
    "status": "pending",
    "estimated_time": 300
  }
}
```

### 2.4 获取同步状态
**GET** `/api/vulnerability/intelligence/sync/:sync_id`

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| sync_id | string | 是 | 同步任务ID |

**响应示例**:
```json
{
  "code": 20000,
  "msg": "success",
  "data": {
    "sync_id": "sync_1234567890",
    "source": "nvd",
    "status": "running",
    "progress": 65,
    "total_count": 1000,
    "processed_count": 650,
    "new_count": 120,
    "updated_count": 530,
    "error_count": 0,
    "start_time": "2024-01-16T10:00:00Z",
    "estimated_end_time": "2024-01-16T10:05:00Z"
  }
}
```

### 2.5 获取漏洞统计信息
**GET** `/api/vulnerability/intelligence/stats`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| period | string | 否 | 统计周期 (day/week/month/year) |
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |

**响应示例**:
```json
{
  "code": 20000,
  "msg": "success",
  "data": {
    "total_count": 15000,
    "today_count": 45,
    "week_count": 320,
    "month_count": 1500,
    "by_severity": {
      "critical": 150,
      "high": 850,
      "medium": 7500,
      "low": 6500
    },
    "by_source": {
      "nvd": 8000,
      "avd": 3000,
      "cnnvd": 2500,
      "cnvd": 1500
    },
    "trend": [
      {
        "date": "2024-01-01",
        "total": 42,
        "critical": 2,
        "high": 8,
        "medium": 20,
        "low": 12
      }
    ]
  }
}
```

### 2.6 搜索漏洞情报
**GET** `/api/vulnerability/intelligence/search`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| q | string | 是 | 搜索关键词 |
| fields | array | 否 | 搜索字段 (title,description,cve_id等) |
| page | integer | 否 | 页码 |
| pageSize | integer | 否 | 每页数量 |

**响应示例**:
```json
{
  "code": 20000,
  "msg": "success",
  "data": {
    "total": 25,
    "page": 1,
    "pageSize": 20,
    "items": [
      {
        "id": 123,
        "cve_id": "CVE-2024-12345",
        "title": "Apache HTTP Server 远程代码执行漏洞",
        "severity": "critical",
        "cvss_score": 9.8,
        "source": "nvd",
        "published_at": "2024-01-15T10:30:00Z",
        "highlight": {
          "title": ["<em>Apache</em> HTTP Server 远程代码执行漏洞"],
          "description": ["<em>Apache</em> HTTP Server 2.4.58及之前版本存在..."]
        }
      }
    ]
  }
}
```

## 3. 检测任务管理API

### 3.1 创建检测任务
**POST** `/api/detection/tasks`

**请求体**:
```json
{
  "name": "Web服务器漏洞检测",
  "description": "检测所有Web服务器的漏洞",
  "targets": ["192.168.1.0/24", "example.com"],
  "detection_type": "vulnerability",
  "config": {
    "scan_depth": "full",
    "port_range": "1-65535",
    "timeout": 30,
    "concurrency": 10,
    "rules": ["high_risk", "web_vuln", "database_vuln"]
  },
  "schedule": {
    "type": "once",
    "start_time": "2024-01-16T14:00:00Z"
  },
  "notifications": {
    "email": ["admin@example.com"],
    "webhook": "https://hooks.example.com/vuln-alert"
  }
}
```

**响应示例**:
```json
{
  "code": 20000,
  "msg": "检测任务创建成功",
  "data": {
    "id": 1001,
    "name": "Web服务器漏洞检测",
    "status": "pending",
    "created_at": "2024-01-16T10:00:00Z",
    "estimated_duration": 1800
  }
}
```

### 3.2 获取检测任务列表
**GET** `/api/detection/tasks`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| page | integer | 否 | 页码 |
| pageSize | integer | 否 | 每页数量 |
| status | string | 否 | 任务状态过滤 |
| detection_type | string | 否 | 检测类型过滤 |
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |

**响应示例**:
```json
{
  "code": 20000,
  "msg": "success",
  "data": {
    "total": 50,
    "page": 1,
    "pageSize": 20,
    "items": [
      {
        "id": 1001,
        "name": "Web服务器漏洞检测",
        "description": "检测所有Web服务器的漏洞",
        "detection_type": "vulnerability",
        "status": "running",
        "progress": 65,
        "target_count": 254,
        "scanned_count": 165,
        "vulnerability_count": 12,
        "start_time": "2024-01-16T10:00:00Z",
        "estimated_end_time": "2024-01-16T10:30:00Z",
        "created_at": "2024-01-16T09:45:00Z"
      }
    ]
  }
}
```

### 3.3 获取检测任务详情
**GET** `/api/detection/tasks/:id`

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | 任务ID |

**响应示例**:
```json
{
  "code": 20000,
  "msg": "success",
  "data": {
    "id": 1001,
    "name": "Web服务器漏洞检测",
    "description": "检测所有Web服务器的漏洞",
    "targets": ["192.168.1.0/24", "example.com"],
    "detection_type": "vulnerability",
    "config": {
      "scan_depth": "full",
      "port_range": "1-65535",
      "timeout": 30,
      "concurrency": 10,
      "rules": ["high_risk", "web_vuln", "database_vuln"]
    },
    "schedule": {
      "type": "once",
      "start_time": "2024-01-16T14:00:00Z"
    },
    "status": "running",
    "progress": 65,
    "target_count": 254,
    "scanned_count": 165,
    "vulnerability_count": 12,
    "critical_count": 2,
    "high_count": 5,
    "medium_count": 3,
    "low_count": 2,
    "start_time": "2024-01-16T10:00:00Z",
    "estimated_end_time": "2024-01-16T10:30:00Z",
    "created_at": "2024-01-16T09:45:00Z",
    "updated_at": "2024-01-16T10:15:00Z"
  }
}
```

### 3.4 更新任务状态
**PUT** `/api/detection/tasks/:id/status`

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | 任务ID |

**请求体**:
```json
{
  "status": "paused",
  "reason": "手动暂停"
}
```

**响应示例**:
```json
{
  "code": 20000,
  "msg": "任务状态更新成功",
  "data": {
    "id": 1001,
    "status": "paused",
    "updated_at": "2024-01-16T10:20:00Z"
  }
}
```

### 3.5 获取检测结果
**GET** `/api/detection/tasks/:id/results`

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | 任务ID |

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| page | integer | 否 | 页码 |
| pageSize | integer | 否 | 每页数量 |
| severity | string | 否 | 严重程度过滤 |
| status | string | 否 | 状态过滤 |
| verified | boolean | 否 | 是否已验证 |

**响应示例**:
```json
{
  "code": 20000,
  "msg": "success",
  "data": {
    "total": 12,
    "page": 1,
    "pageSize": 20,
    "items": [
      {
        "id": 5001,
        "task_id": 1001,
        "vuln_intel_id": 123,
        "asset_id": 2001,
        "target": "192.168.1.100:80",
        "vulnerability": {
          "cve_id": "CVE-2024-12345",
          "title": "Apache HTTP Server 远程代码执行漏洞",
          "severity": "critical",
          "cvss_score": 9.8
        },
        "asset": {
          "ip": "192.168.1.100",
          "hostname": "web01.example.com",
          "service": "Apache HTTP Server",
          "version": "2.4.58"
        },
        "status": "detected",
        "confidence": 95,
        "risk_level": "critical",
        "verified": false,
        "evidence": {
          "request": "GET / HTTP/1.1",
          "response": "HTTP/1.1 200 OK\nServer: Apache/2.4.58",
          "matched_pattern": "Server: Apache/2\\.4\\.58"
        },
        "created_at": "2024-01-16T10:05:00Z"
      }
    ]
  }
}
```

## 4. 预警通知API

### 4.1 获取预警列表
**GET** `/api/alerts`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| page | integer | 否 | 页码 |
| pageSize | integer | 否 | 每页数量 |
| severity | string | 否 | 严重程度过滤 |
| status | string | 否 | 状态过滤 (unread/read/resolved) |
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |

**响应示例**:
```json
{
  "code": 20000,
  "msg": "success",
  "data": {
    "total": 45,
    "page": 1,
    "pageSize": 20,
    "items": [
      {
        "id": 3001,
        "title": "发现高危漏洞 CVE-2024-12345",
        "description": "在资产 192.168.1.100 上发现Apache HTTP Server远程代码执行漏洞",
        "severity": "critical",
        "type": "vulnerability",
        "source": "detection",
        "status": "unread",
        "data": {
          "vulnerability_id": 123,
          "asset_id": 2001,
          "detection_id": 5001,
          "cve_id": "CVE-2024-12345",
          "cvss_score": 9.8
        },
        "created_at": "2024-01-16T10:05:00Z",
        "read_at": null,
        "resolved_at": null
      }
    ]
  }
}
```

### 4.2 创建预警
**POST** `/api/alerts`

**请求体**:
```json
{
  "title": "手动创建预警",
  "description": "测试预警通知",
  "severity": "high",
  "type": "manual",
  "data": {
    "message": "这是一个测试预警"
  },
  "channels": ["email", "webhook"],
  "recipients": ["admin@example.com"]
}
```

**响应示例**:
```json
{
  "code": 20000,
  "msg": "预警创建成功",
  "data": {
    "id": 3002,
    "title": "手动创建预警",
    "status": "pending",
    "created_at": "2024-01-16T10:30:00Z"
  }
}
```

### 4.3 更新预警状态
**PUT** `/api/alerts/:id/status`

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | 预警ID |

**请求体**:
```json
{
  "status": "read",
  "notes": "已查看"
}
```

**响应示例**:
```json
{
  "code": 20000,
  "msg": "预警状态更新成功",
  "data": {
    "id": 3001,
    "status": "read",
    "read_at": "2024-01-16T10:35:00Z",
    "updated_at": "2024-01-16T10:35:00Z"
  }
}
```

### 4.4 获取预警统计
**GET** `/api/alerts/stats`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| period | string | 否 | 统计周期 |
| start_date | string | 否 | 开始日期 |
| end_date | string | 否 | 结束日期 |

**响应示例**:
```json
{
  "code": 20000,
  "msg": "success",
  "data": {
    "total_count": 150,
    "unread_count": 25,
    "by_severity": {
      "critical": 15,
      "high": 45,
      "medium": 60,
      "low": 30
    },
    "by_type": {
      "vulnerability": 120,
      "system": 20,
      "manual": 10
    },
    "by_status": {
      "unread": 25,
      "read": 100,
      "resolved": 25
    },
    "trend": [
      {
        "date": "2024-01-01",
        "total": 8,
        "critical": 1,
        "high": 3,
        "medium": 3,
        "low": 1
      }
    ]
  }
}
```

### 4.5 订阅预警
**POST** `/api/alerts/subscribe`

**请求体**:
```json
{
  "user_id": 1,
  "channels": ["email", "webhook", "dingtalk"],
  "severities": ["critical", "high"],
  "types": ["vulnerability", "system"],
  "schedule": "realtime"
}
```

**响应示例**:
```json
{
  "code": 20000,
  "msg": "预警订阅成功",
  "data": {
    "subscription_id": "sub_1234567890",
    "user_id": 1,
    "channels": ["email", "webhook", "dingtalk"],
    "created_at": "2024-01-16T10:40:00Z"
  }
}
```

## 5. 系统配置API

### 5.1 获取漏洞情报源配置
**GET** `/api/config/intelligence-sources`

**响应示例**:
```json
{
  "code": 20000,
  "msg": "success",
  "data": [
    {
      "id": 1,
      "name": "阿里云漏洞库 (AVD)",
      "url": "https://avd.aliyun.com/api",
      "type": "avd",
      "enabled": true,
      "sync_interval": 3600,
      "last_sync": "2024-01-16T08:00:00Z",
      "last_sync_status": "success",
      "config": {
        "api_version": "v1",
        "rate_limit": 10
      },
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-16T08:00:00Z"
    }
  ]
}
```

### 5.2 更新漏洞情报源配置
**PUT** `/api/config/intelligence-sources/:id`

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | 配置ID |

**请求体**:
```json
{
  "enabled": false,
  "sync_interval": 7200,
  "config": {
    "api_key": "new-api-key",
    "rate_limit": 5
  }
}
```

**响应示例**:
```json
{
  "code": 20000,
  "msg": "配置更新成功",
  "data": {
    "id": 1,
    "enabled": false,
    "sync_interval": 7200,
    "updated_at": "2024-01-16T11:00:00Z"
  }
}
```

### 5.3 获取检测规则
**GET** `/api/config/detection-rules`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| enabled | boolean | 否 | 是否启用过滤 |
| type | string | 否 | 规则类型过滤 |

**响应示例**:
```json
{
  "code": 20000,
  "msg": "success",
  "data": [
    {
      "id": 1,
      "name": "高危CVE检测",
      "description": "检测CVSS评分>=7.0的高危CVE漏洞",
      "type": "cve",
      "condition": "cvss_score >= 7.0",
      "action": "alert",
      "severity": "high",
      "enabled": true,
      "priority": 10,
      "tags": ["cve", "high_risk"],
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-16T09:00:00Z"
    }
  ]
}
```

### 5.4 创建检测规则
**POST** `/api/config/detection-rules`

**请求体**:
```json
{
  "name": "自定义Web漏洞检测",
  "description": "检测特定Web应用的漏洞",
  "type": "custom",
  "condition": "product = 'WordPress' AND version < '6.0'",
  "action": "report",
  "severity": "medium",
  "enabled": true,
  "priority": 5,
  "tags": ["web", "wordpress"]
}
```

**响应示例**:
```json
{
  "code": 20000,
  "msg": "检测规则创建成功",
  "data": {
    "id": 6,
    "name": "自定义Web漏洞检测",
    "created_at": "2024-01-16T11:05:00Z"
  }
}
```

### 5.5 获取预警配置
**GET** `/api/config/alert-configs`

**响应示例**:
```json
{
  "code": 20000,
  "msg": "success",
  "data": [
    {
      "id": 1,
      "name": "高危漏洞预警",
      "description": "高危及以上漏洞实时预警",
      "severity_filter": ["critical", "high"],
      "source_filter": ["avd", "nvd", "cnnvd", "cnvd"],
      "enabled": true,
      "notification_channels": {
        "email": true,
        "webhook": true,
        "dingtalk": false
      },
      "schedule": {
        "type": "realtime",
        "interval": 0
      },
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-16T09:00:00Z"
    }
  ]
}
```

## 6. 数据统计API

### 6.1 获取漏洞趋势统计
**GET** `/api/stats/vulnerability-trend`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| days | integer | 否 | 统计天数，默认30 |
| group_by | string | 否 | 分组方式 (day/week/month) |

**响应示例**:
```json
{
  "code": 20000,
  "msg": "success",
  "data": {
    "period": "30 days",
    "total": 1500,
    "trend": [
      {
        "date": "2024-01-01",
        "total": 42,
        "critical": 2,
        "high": 8,
        "medium": 20,
        "low": 12,
        "zero_day": 1,
        "with_exploit": 5
      }
    ],
    "by_source": {
      "nvd": 800,
      "avd": 300,
      "cnnvd": 250,
      "cnvd": 150
    },
    "by_severity": {
      "critical": 50,
      "high": 250,
      "medium": 700,
      "low": 500
    }
  }
}
```

### 6.2 获取检测结果统计
**GET** `/api/stats/detection-results`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| days | integer | 否 | 统计天数，默认7 |
| task_id | integer | 否 | 任务ID过滤 |

**响应示例**:
```json
{
  "code": 20000,
  "msg": "success",
  "data": {
    "period": "7 days",
    "total_tasks": 15,
    "total_detections": 1200,
    "by_status": {
      "detected": 850,
      "verified": 300,
      "false_positive": 50
    },
    "by_risk_level": {
      "critical": 25,
      "high": 150,
      "medium": 500,
      "low": 525
    },
    "top_vulnerabilities": [
      {
        "cve_id": "CVE-2024-12345",
        "title": "Apache HTTP Server 远程代码执行漏洞",
        "count": 45,
        "severity": "critical"
      }
    ],
    "top_assets": [
      {
        "ip": "192.168.1.100",
        "hostname": "web01.example.com",
        "vulnerability_count": 12,
        "critical_count": 2
      }
    ]
  }
}
```

### 6.3 获取系统健康状态
**GET** `/api/stats/health`

**响应示例**:
```json
{
  "code": 20000,
  "msg": "success",
  "data": {
    "status": "healthy",
    "components": {
      "database": {
        "status": "healthy",
        "latency": 15
      },
      "redis": {
        "status": "healthy",
        "latency": 2
      },
      "scheduler": {
        "status": "healthy",
        "running_jobs": 3
      },
      "scanner": {
        "status": "healthy",
        "active_scans": 2
      }
    },
    "metrics": {
      "vulnerability_count": 15000,
      "detection_count": 1200,
      "alert_count": 45,
      "uptime": 86400,
      "memory_usage": 65.5,
      "cpu_usage": 42.3
    },
    "last_updated": "2024-01-16T11:00:00Z"
  }
}
```

## 7. 错误码说明

### 7.1 通用错误码
| 错误码 | 说明 | HTTP状态码 |
|--------|------|------------|
| 40001 | 参数错误 | 400 |
| 40002 | 请求体格式错误 | 400 |
| 40101 | 未授权访问 | 401 |
| 40301 | 权限不足 | 403 |
| 40401 | 资源不存在 | 404 |
| 50001 | 服务器内部错误 | 500 |
| 50002 | 数据库错误 | 500 |
| 50003 | 外部服务错误 | 500 |

### 7.2 业务错误码
| 错误码 | 说明 | HTTP状态码 |
|--------|------|------------|
| 60001 | 漏洞情报源配置错误 | 400 |
| 60002 | 数据同步失败 | 500 |
| 60003 | 检测任务创建失败 | 400 |
| 60004 | 检测任务不存在 | 404 |
| 60005 | 检测任务已存在 | 409 |
| 60006 | 检测规则语法错误 | 400 |
| 60007 | 预警配置错误 | 400 |
| 60008 | 数据格式错误 | 400 |

## 8. 数据模型

### 8.1 漏洞情报 (VulnerabilityIntelligence)
```go
type VulnerabilityIntelligence struct {
    ID               int       `json:"id" db:"id"`
    CVEID            string    `json:"cve_id" db:"cve_id"`
    CNVDID           string    `json:"cnvd_id" db:"cnvd_id"`
    CNNVDID          string    `json:"cnnvd_id" db:"cnnvd_id"`
    Title            string    `json:"title" db:"title"`
    Description      string    `json:"description" db:"description"`
    Severity         string    `json:"severity" db:"severity"`
    CVSSScore        float64   `json:"cvss_score" db:"cvss_score"`
    CVSSVector       string    `json:"cvss_vector" db:"cvss_vector"`
    AffectedProducts []string  `json:"affected_products" db:"affected_products"`
    AffectedVersions []string  `json:"affected_versions" db:"affected_versions"`
    References       []string  `json:"references" db:"references"`
    ExploitAvailable bool      `json:"exploit_available" db:"exploit_available"`
    POCAvailable     bool      `json:"poc_available" db:"poc_available"`
    PublishedAt      time.Time `json:"published_at" db:"published_at"`
    LastModified     time.Time `json:"last_modified" db:"last_modified"`
    Source           string    `json:"source" db:"source"`
    Is0Day           bool      `json:"is_0day" db:"is_0day"`
    CreatedAt        time.Time `json:"created_at" db:"created_at"`
    UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}
```

### 8.2 检测结果 (DetectionResult)
```go
type DetectionResult struct {
    ID             int                    `json:"id" db:"id"`
    TaskID         int                    `json:"task_id" db:"task_id"`
    VulnIntelID    int                    `json:"vuln_intel_id" db:"vuln_intel_id"`
    AssetID        int                    `json:"asset_id" db:"asset_id"`
    Target         string                 `json:"target" db:"target"`
    POCTemplateID  int                    `json:"poc_template_id" db:"poc_template_id"`
    Status         string                 `json:"status" db:"status"`
    Confidence     int                    `json:"confidence" db:"confidence"`
    Evidence       map[string]interface{} `json:"evidence" db:"evidence"`
    RequestData    string                 `json:"request_data" db:"request_data"`
    ResponseData   string                 `json:"response_data" db:"response_data"`
    MatchedPattern string                 `json:"matched_pattern" db:"matched_pattern"`
    RiskLevel      string                 `json:"risk_level" db:"risk_level"`
    Verified       bool                   `json:"verified" db:"verified"`
    VerifiedBy     string                 `json:"verified_by" db:"verified_by"`
    VerificationNotes string              `json:"verification_notes" db:"verification_notes"`
    CreatedAt      time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}
```

### 8.3 预警 (Alert)
```go
type Alert struct {
    ID          int                    `json:"id" db:"id"`
    Title       string                 `json:"title" db:"title"`
    Description string                 `json:"description" db:"description"`
    Severity    string                 `json:"severity" db:"severity"`
    Type        string                 `json:"type" db:"type"`
    Source      string                 `json:"source" db:"source"`
    Status      string                 `json:"status" db:"status"`
    Data        map[string]interface{} `json:"data" db:"data"`
    CreatedAt   time.Time              `json:"created_at" db:"created_at"`
    ReadAt      *time.Time             `json:"read_at" db:"read_at"`
    ResolvedAt  *time.Time             `json:"resolved_at" db:"resolved_at"`
}
```

## 9. 安全考虑

### 9.1 认证和授权
- 所有API需要JWT Token认证
- 基于角色的访问控制 (RBAC)
- API密钥管理
- 请求频率限制

### 9.2 数据安全
- 敏感数据加密存储
- 数据传输使用HTTPS
- 输入验证和过滤
- SQL注入防护

### 9.3 审计日志
- 所有API调用记录日志
- 敏感操作记录详细日志
- 日志脱敏处理
- 日志归档和保留

## 10. 性能优化

### 10.1 缓存策略
- 热点数据使用Redis缓存
- 查询结果缓存
- 缓存失效策略
- 分布式缓存支持

### 10.2 数据库优化
- 合理使用索引
- 查询优化
- 分页查询
- 读写分离

### 10.3 异步处理
- 大数据量操作异步处理
- 消息队列解耦
- 批量操作优化
- 并发控制

---

**版本**: v1.0  
**创建时间**: 2026-03-12  
**更新记录**:  
- v1.0: 初始版本，定义完整的API接口规范