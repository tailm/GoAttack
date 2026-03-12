package alert

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

// Init 初始化预警通知服务
func Init(db *sql.DB) error {
	var initErr error
	serviceOnce.Do(func() {
		serviceInstance = NewService(db)
		
		// 初始化默认预警规则
		if err := initDefaultRules(context.Background(), db); err != nil {
			log.Errorf("初始化默认预警规则失败: %v", err)
			initErr = err
			return
		}
		
		log.Info("预警通知服务初始化完成")
	})
	return initErr
}

// GetService 获取服务实例
func GetService() *Service {
	return serviceInstance
}

// initDefaultRules 初始化默认预警规则
func initDefaultRules(ctx context.Context, db *sql.DB) error {
	// 检查是否已有规则
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM alert_rules").Scan(&count)
	if err != nil {
		return err
	}

	// 如果已有规则，则跳过初始化
	if count > 0 {
		log.Info("预警规则已存在，跳过默认规则初始化")
		return nil
	}

	// 默认预警规则
	defaultRules := []struct {
		name        string
		description string
		ruleType    string
		severity    string
		condition   string
		action      string
		enabled     bool
		priority    int
		tags        []string
		config      map[string]interface{}
	}{
		{
			name:        "高危漏洞预警",
			description: "检测到高危漏洞时发送预警",
			ruleType:    "vulnerability",
			severity:    "critical",
			condition:   "severity == 'critical'",
			action:      "notification",
			enabled:     true,
			priority:    1,
			tags:        []string{"vulnerability", "critical", "alert"},
			config: map[string]interface{}{
				"channels": []string{"system", "email"},
				"template": "高危漏洞预警: {{.Title}}",
			},
		},
		{
			name:        "中危漏洞预警",
			description: "检测到中危漏洞时发送预警",
			ruleType:    "vulnerability",
			severity:    "high",
			condition:   "severity == 'high'",
			action:      "notification",
			enabled:     true,
			priority:    2,
			tags:        []string{"vulnerability", "high", "alert"},
			config: map[string]interface{}{
				"channels": []string{"system"},
				"template": "中危漏洞预警: {{.Title}}",
			},
		},
		{
			name:        "未授权访问预警",
			description: "检测到未授权访问时发送预警",
			ruleType:    "threat",
			severity:    "critical",
			condition:   "type == 'threat' AND severity == 'critical'",
			action:      "notification",
			enabled:     true,
			priority:    3,
			tags:        []string{"threat", "unauthorized", "alert"},
			config: map[string]interface{}{
				"channels": []string{"system", "email", "webhook"},
				"template": "未授权访问预警: {{.Title}}",
			},
		},
		{
			name:        "端口扫描预警",
			description: "检测到端口扫描活动时发送预警",
			ruleType:    "threat",
			severity:    "medium",
			condition:   "type == 'threat' AND target_type == 'host'",
			action:      "notification",
			enabled:     true,
			priority:    4,
			tags:        []string{"threat", "port-scan", "alert"},
			config: map[string]interface{}{
				"channels": []string{"system"},
				"template": "端口扫描预警: {{.Title}}",
			},
		},
		{
			name:        "系统异常预警",
			description: "系统出现异常时发送预警",
			ruleType:    "system",
			severity:    "high",
			condition:   "type == 'system'",
			action:      "notification",
			enabled:     true,
			priority:    5,
			tags:        []string{"system", "error", "alert"},
			config: map[string]interface{}{
				"channels": []string{"system", "email"},
				"template": "系统异常预警: {{.Title}}",
			},
		},
		{
			name:        "Web攻击预警",
			description: "检测到Web攻击时发送预警",
			ruleType:    "threat",
			severity:    "high",
			condition:   "type == 'threat' AND target_type == 'web'",
			action:      "notification",
			enabled:     true,
			priority:    6,
			tags:        []string{"threat", "web-attack", "alert"},
			config: map[string]interface{}{
				"channels": []string{"system", "webhook"},
				"template": "Web攻击预警: {{.Title}}",
			},
		},
		{
			name:        "API安全预警",
			description: "检测到API安全问题时发送预警",
			ruleType:    "threat",
			severity:    "medium",
			condition:   "type == 'threat' AND target_type == 'api'",
			action:      "notification",
			enabled:     true,
			priority:    7,
			tags:        []string{"threat", "api-security", "alert"},
			config: map[string]interface{}{
				"channels": []string{"system"},
				"template": "API安全预警: {{.Title}}",
			},
		},
		{
			name:        "网络攻击预警",
			description: "检测到网络攻击时发送预警",
			ruleType:    "threat",
			severity:    "critical",
			condition:   "type == 'threat' AND target_type == 'network'",
			action:      "notification",
			enabled:     true,
			priority:    8,
			tags:        []string{"threat", "network-attack", "alert"},
			config: map[string]interface{}{
				"channels": []string{"system", "email", "sms"},
				"template": "网络攻击预警: {{.Title}}",
			},
		},
		{
			name:        "自定义规则预警",
			description: "自定义规则触发时发送预警",
			ruleType:    "custom",
			severity:    "medium",
			condition:   "type == 'custom'",
			action:      "notification",
			enabled:     true,
			priority:    9,
			tags:        []string{"custom", "alert"},
			config: map[string]interface{}{
				"channels": []string{"system"},
				"template": "自定义规则预警: {{.Title}}",
			},
		},
		{
			name:        "信息泄露预警",
			description: "检测到信息泄露时发送预警",
			ruleType:    "threat",
			severity:    "high",
			condition:   "type == 'threat' AND severity == 'high'",
			action:      "notification",
			enabled:     true,
			priority:    10,
			tags:        []string{"threat", "information-leak", "alert"},
			config: map[string]interface{}{
				"channels": []string{"system", "email"},
				"template": "信息泄露预警: {{.Title}}",
			},
		},
	}

	// 插入默认规则
	for _, rule := range defaultRules {
		tagsJSON, _ := json.Marshal(rule.tags)
		configJSON, _ := json.Marshal(rule.config)

		_, err := db.ExecContext(ctx, `
			INSERT INTO alert_rules (
				name, description, type, severity, condition, action,
				config, enabled, priority, tags
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (name) DO NOTHING
		`,
			rule.name,
			rule.description,
			rule.ruleType,
			rule.severity,
			rule.condition,
			rule.action,
			string(configJSON),
			rule.enabled,
			rule.priority,
			string(tagsJSON),
		)
		if err != nil {
			log.Errorf("插入默认预警规则失败: %v", err)
			continue
		}
	}

	log.Infof("初始化了 %d 条默认预警规则", len(defaultRules))
	return nil
}