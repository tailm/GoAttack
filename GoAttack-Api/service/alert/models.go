package alert

import (
	"time"
)

// Alert 预警通知
type Alert struct {
	ID          int                    `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`        // 预警类型：vulnerability, threat, system, custom
	Severity    string                 `json:"severity"`    // 严重程度：critical, high, medium, low, info
	Status      string                 `json:"status"`      // 状态：pending, processing, resolved, closed
	Source      string                 `json:"source"`      // 来源：detection, intelligence, manual, system
	SourceID    string                 `json:"source_id"`   // 来源ID
	Target      string                 `json:"target"`      // 目标
	TargetType  string                 `json:"target_type"` // 目标类型：host, network, web, api
	Details     map[string]interface{} `json:"details"`     // 详细信息
	Metadata    map[string]interface{} `json:"metadata"`    // 元数据
	CreatedBy   string                 `json:"created_by"`   // 创建者
	AssignedTo  string                 `json:"assigned_to"`  // 分配给
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ResolvedAt  *time.Time             `json:"resolved_at"`
	ClosedAt    *time.Time             `json:"closed_at"`
}

// AlertRule 预警规则
type AlertRule struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`        // 规则类型：vulnerability, threat, system, custom
	Severity    string                 `json:"severity"`    // 触发严重程度
	Condition   string                 `json:"condition"`   // 触发条件
	Action      string                 `json:"action"`      // 触发动作：email, webhook, sms, notification
	Config      map[string]interface{} `json:"config"`      // 动作配置
	Enabled     bool                   `json:"enabled"`     // 是否启用
	Priority    int                    `json:"priority"`    // 优先级
	Tags        []string               `json:"tags"`        // 标签
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Notification 通知记录
type Notification struct {
	ID        int                    `json:"id"`
	AlertID   int                    `json:"alert_id"`
	Type      string                 `json:"type"`      // 通知类型：email, webhook, sms, notification
	Status    string                 `json:"status"`    // 状态：pending, sent, failed, delivered
	Recipient string                 `json:"recipient"` // 接收者
	Content   map[string]interface{} `json:"content"`   // 通知内容
	Error     string                 `json:"error"`     // 错误信息
	SentAt    *time.Time             `json:"sent_at"`
	CreatedAt time.Time              `json:"created_at"`
}

// AlertQuery 预警查询参数
type AlertQuery struct {
	Search     string
	Type       string
	Severity   string
	Status     string
	Source     string
	Target     string
	TargetType string
	CreatedBy  string
	AssignedTo string
	StartDate  *time.Time
	EndDate    *time.Time
	Page       int
	PageSize   int
	SortBy     string
	SortOrder  string
}

// AlertStats 预警统计
type AlertStats struct {
	Total        int            `json:"total"`
	BySeverity   map[string]int `json:"by_severity"`
	ByType       map[string]int `json:"by_type"`
	ByStatus     map[string]int `json:"by_status"`
	BySource     map[string]int `json:"by_source"`
	RecentAlerts []Alert        `json:"recent_alerts"`
}

// CreateAlertRequest 创建预警请求
type CreateAlertRequest struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Source      string                 `json:"source"`
	SourceID    string                 `json:"source_id"`
	Target      string                 `json:"target"`
	TargetType  string                 `json:"target_type"`
	Details     map[string]interface{} `json:"details"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedBy   string                 `json:"created_by"`
}

// UpdateAlertRequest 更新预警请求
type UpdateAlertRequest struct {
	Title       *string                `json:"title,omitempty"`
	Description *string                `json:"description,omitempty"`
	Status      *string                `json:"status,omitempty"`
	AssignedTo  *string                `json:"assigned_to,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CreateRuleRequest 创建规则请求
type CreateRuleRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Condition   string                 `json:"condition"`
	Action      string                 `json:"action"`
	Config      map[string]interface{} `json:"config"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	Tags        []string               `json:"tags"`
}

// RuleQuery 规则查询参数
type RuleQuery struct {
	Search   string
	Type     string
	Severity string
	Enabled  *bool
	Page     int
	PageSize int
	SortBy   string
	SortOrder string
}

// NotificationQuery 通知查询参数
type NotificationQuery struct {
	AlertID   int
	Type      string
	Status    string
	Recipient string
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	PageSize  int
	SortBy    string
	SortOrder string
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	Email    EmailConfig    `json:"email"`
	Webhook  WebhookConfig  `json:"webhook"`
	SMS      SMSConfig      `json:"sms"`
	Slack    SlackConfig    `json:"slack"`
	DingTalk DingTalkConfig `json:"dingtalk"`
	WeChat   WeChatConfig   `json:"wechat"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	Enabled  bool   `json:"enabled"`
	SMTPHost string `json:"smtp_host"`
	SMTPPort int    `json:"smtp_port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	SSL      bool   `json:"ssl"`
}

// WebhookConfig Webhook配置
type WebhookConfig struct {
	Enabled bool     `json:"enabled"`
	URLs    []string `json:"urls"`
	Headers map[string]string `json:"headers"`
}

// SMSConfig 短信配置
type SMSConfig struct {
	Enabled    bool   `json:"enabled"`
	Provider   string `json:"provider"`
	APIKey     string `json:"api_key"`
	APISecret  string `json:"api_secret"`
	SignName   string `json:"sign_name"`
	TemplateID string `json:"template_id"`
}

// SlackConfig Slack配置
type SlackConfig struct {
	Enabled   bool   `json:"enabled"`
	WebhookURL string `json:"webhook_url"`
	Channel   string `json:"channel"`
	Username  string `json:"username"`
}

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	Enabled   bool   `json:"enabled"`
	WebhookURL string `json:"webhook_url"`
	Secret    string `json:"secret"`
}

// WeChatConfig 微信配置
type WeChatConfig struct {
	Enabled   bool   `json:"enabled"`
	CorpID    string `json:"corp_id"`
	AgentID   int    `json:"agent_id"`
	Secret    string `json:"secret"`
	ToUser    string `json:"to_user"`
	ToParty   string `json:"to_party"`
	ToTag     string `json:"to_tag"`
}