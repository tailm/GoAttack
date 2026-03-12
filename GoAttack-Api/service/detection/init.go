package detection

import (
	"GoAttack/common/log"
	"context"
	"database/sql"
	"encoding/json"
	"sync"
)

var (
	serviceInstance *Service
	serviceOnce     sync.Once
)

// Init 初始化检测引擎服务
func Init(db *sql.DB) error {
	var initErr error
	serviceOnce.Do(func() {
		serviceInstance = NewService(db)
		
		// 初始化默认检测规则
		if err := initDefaultRules(context.Background(), db); err != nil {
			log.Errorf("初始化默认检测规则失败: %v", err)
			initErr = err
			return
		}
		
		log.Info("检测引擎服务初始化完成")
	})
	return initErr
}

// GetService 获取服务实例
func GetService() *Service {
	return serviceInstance
}

// initDefaultRules 初始化默认检测规则
func initDefaultRules(ctx context.Context, db *sql.DB) error {
	// 检查是否已有规则
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM detection_rules").Scan(&count)
	if err != nil {
		return err
	}

	// 如果已有规则，则跳过初始化
	if count > 0 {
		log.Info("检测规则已存在，跳过默认规则初始化")
		return nil
	}

	// 默认检测规则
	defaultRules := []struct {
		name        string
		description string
		ruleType    string
		condition   string
		action      string
		severity    string
		enabled     bool
		priority    int
		tags        []string
	}{
		{
			name:        "SSH弱密码检测",
			description: "检测SSH服务是否存在弱密码",
			ruleType:    "service",
			condition:   "service == 'ssh'",
			action:      "alert",
			severity:    "high",
			enabled:     true,
			priority:    10,
			tags:        []string{"ssh", "authentication", "brute-force"},
		},
		{
			name:        "MySQL未授权访问",
			description: "检测MySQL服务是否存在未授权访问漏洞",
			ruleType:    "service",
			condition:   "service == 'mysql'",
			action:      "alert",
			severity:    "critical",
			enabled:     true,
			priority:    5,
			tags:        []string{"mysql", "database", "unauthorized"},
		},
		{
			name:        "Redis未授权访问",
			description: "检测Redis服务是否存在未授权访问漏洞",
			ruleType:    "service",
			condition:   "service == 'redis'",
			action:      "alert",
			severity:    "critical",
			enabled:     true,
			priority:    5,
			tags:        []string{"redis", "database", "unauthorized"},
		},
		{
			name:        "FTP匿名登录",
			description: "检测FTP服务是否允许匿名登录",
			ruleType:    "service",
			condition:   "service == 'ftp'",
			action:      "alert",
			severity:    "medium",
			enabled:     true,
			priority:    20,
			tags:        []string{"ftp", "anonymous", "file-transfer"},
		},
		{
			name:        "HTTP目录遍历",
			description: "检测Web服务是否存在目录遍历漏洞",
			ruleType:    "web",
			condition:   "web.framework == 'Apache' || web.server == 'nginx'",
			action:      "alert",
			severity:    "medium",
			enabled:     true,
			priority:    15,
			tags:        []string{"http", "web", "directory-traversal"},
		},
		{
			name:        "Telnet服务开放",
			description: "检测Telnet服务是否开放（不安全协议）",
			ruleType:    "port",
			condition:   "port == 23",
			action:      "alert",
			severity:    "high",
			enabled:     true,
			priority:    8,
			tags:        []string{"telnet", "insecure", "protocol"},
		},
		{
			name:        "SMBv1协议检测",
			description: "检测SMBv1协议使用（存在永恒之蓝漏洞风险）",
			ruleType:    "service",
			condition:   "service == 'smb' && version < '2.0'",
			action:      "alert",
			severity:    "critical",
			enabled:     true,
			priority:    3,
			tags:        []string{"smb", "eternalblue", "windows"},
		},
		{
			name:        "RDP服务开放",
			description: "检测RDP服务是否开放（远程桌面协议）",
			ruleType:    "port",
			condition:   "port == 3389",
			action:      "alert",
			severity:    "medium",
			enabled:     true,
			priority:    12,
			tags:        []string{"rdp", "remote-desktop", "windows"},
		},
		{
			name:        "Elasticsearch未授权访问",
			description: "检测Elasticsearch服务是否存在未授权访问漏洞",
			ruleType:    "service",
			condition:   "service == 'elasticsearch'",
			action:      "alert",
			severity:    "critical",
			enabled:     true,
			priority:    6,
			tags:        []string{"elasticsearch", "search", "unauthorized"},
		},
		{
			name:        "MongoDB未授权访问",
			description: "检测MongoDB服务是否存在未授权访问漏洞",
			ruleType:    "service",
			condition:   "service == 'mongodb'",
			action:      "alert",
			severity:    "critical",
			enabled:     true,
			priority:    7,
			tags:        []string{"mongodb", "database", "unauthorized"},
		},
	}

	// 插入默认规则
	for _, rule := range defaultRules {
		tagsJSON, _ := json.Marshal(rule.tags)
		configJSON, _ := json.Marshal(map[string]interface{}{
			"auto_enabled": true,
			"category":     "default",
		})

		_, err := db.ExecContext(ctx, `
			INSERT INTO detection_rules (
				name, description, type, condition, action, severity,
				enabled, priority, tags, config
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (name) DO NOTHING
		`,
			rule.name,
			rule.description,
			rule.ruleType,
			rule.condition,
			rule.action,
			rule.severity,
			rule.enabled,
			rule.priority,
			string(tagsJSON),
			string(configJSON),
		)
		if err != nil {
			log.Errorf("插入默认规则失败: %v", err)
			continue
		}
	}

	log.Infof("初始化了 %d 条默认检测规则", len(defaultRules))
	return nil
}