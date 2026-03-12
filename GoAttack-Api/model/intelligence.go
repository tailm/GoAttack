package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// VulnerabilityIntelligence 漏洞情报
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
	AffectedProducts StringArray `json:"affected_products" db:"affected_products"`
	AffectedVersions StringArray `json:"affected_versions" db:"affected_versions"`
	References       StringArray `json:"references" db:"references"`
	ExploitAvailable bool      `json:"exploit_available" db:"exploit_available"`
	POCAvailable     bool      `json:"poc_available" db:"poc_available"`
	PublishedAt      time.Time `json:"published_at" db:"published_at"`
	LastModified     time.Time `json:"last_modified" db:"last_modified"`
	Source           string    `json:"source" db:"source"`
	Is0Day           bool      `json:"is_0day" db:"is_0day"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// IntelligenceSource 漏洞情报源
type IntelligenceSource struct {
	ID                int       `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	URL               string    `json:"url" db:"url"`
	Type              string    `json:"type" db:"type"`
	Enabled           bool      `json:"enabled" db:"enabled"`
	SyncInterval      int       `json:"sync_interval" db:"sync_interval"`
	LastSync          *time.Time `json:"last_sync" db:"last_sync"`
	LastSyncStatus    string    `json:"last_sync_status" db:"last_sync_status"`
	LastSyncError     string    `json:"last_sync_error" db:"last_sync_error"`
	Config            JSONMapInterface   `json:"config" db:"config"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// DetectionResult 检测结果
type DetectionResult struct {
	ID                int       `json:"id" db:"id"`
	TaskID            int       `json:"task_id" db:"task_id"`
	VulnIntelID       *int      `json:"vuln_intel_id" db:"vuln_intel_id"`
	AssetID           *int      `json:"asset_id" db:"asset_id"`
	Target            string    `json:"target" db:"target"`
	POCTemplateID     *int      `json:"poc_template_id" db:"poc_template_id"`
	Status            string    `json:"status" db:"status"`
	Confidence        *int      `json:"confidence" db:"confidence"`
	Evidence          JSONMapInterface   `json:"evidence" db:"evidence"`
	RequestData       string    `json:"request_data" db:"request_data"`
	ResponseData      string    `json:"response_data" db:"response_data"`
	MatchedPattern    string    `json:"matched_pattern" db:"matched_pattern"`
	RiskLevel         string    `json:"risk_level" db:"risk_level"`
	Verified          bool      `json:"verified" db:"verified"`
	VerifiedBy        string    `json:"verified_by" db:"verified_by"`
	VerificationNotes string    `json:"verification_notes" db:"verification_notes"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// DetectionRule 检测规则
type DetectionRule struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        string    `json:"type" db:"type"`
	Condition   string    `json:"condition" db:"condition"`
	Action      string    `json:"action" db:"action"`
	Severity    string    `json:"severity" db:"severity"`
	Enabled     bool      `json:"enabled" db:"enabled"`
	Priority    int       `json:"priority" db:"priority"`
	Tags        StringArray `json:"tags" db:"tags"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// AlertConfig 预警配置
type AlertConfig struct {
	ID                   int       `json:"id" db:"id"`
	Name                 string    `json:"name" db:"name"`
	Description          string    `json:"description" db:"description"`
	SeverityFilter       StringArray `json:"severity_filter" db:"severity_filter"`
	SourceFilter         StringArray `json:"source_filter" db:"source_filter"`
	Enabled              bool      `json:"enabled" db:"enabled"`
	NotificationChannels JSONMapInterface   `json:"notification_channels" db:"notification_channels"`
	Schedule             JSONMapInterface   `json:"schedule" db:"schedule"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// SyncLog 同步日志
type SyncLog struct {
	ID            int        `json:"id" db:"id"`
	SourceID      *int       `json:"source_id" db:"source_id"`
	SourceName    string     `json:"source_name" db:"source_name"`
	StartTime     time.Time  `json:"start_time" db:"start_time"`
	EndTime       *time.Time `json:"end_time" db:"end_time"`
	Status        string     `json:"status" db:"status"`
	TotalCount    int        `json:"total_count" db:"total_count"`
	NewCount      int        `json:"new_count" db:"new_count"`
	UpdatedCount  int        `json:"updated_count" db:"updated_count"`
	ErrorCount    int        `json:"error_count" db:"error_count"`
	ErrorMessage  string     `json:"error_message" db:"error_message"`
	Duration      *int       `json:"duration" db:"duration"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// DetectionTask 检测任务
type DetectionTask struct {
	ID         int       `json:"id" db:"id"`
	TaskID     int       `json:"task_id" db:"task_id"`
	DetectionType string `json:"detection_type" db:"detection_type"`
	Config     JSONMapInterface   `json:"config" db:"config"`
	Schedule   JSONMapInterface   `json:"schedule" db:"schedule"`
	LastRun    *time.Time `json:"last_run" db:"last_run"`
	NextRun    *time.Time `json:"next_run" db:"next_run"`
	Enabled    bool      `json:"enabled" db:"enabled"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// MatchRule 匹配规则
type MatchRule struct {
	ID            int       `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	Description   string    `json:"description" db:"description"`
	ProductPattern string   `json:"product_pattern" db:"product_pattern"`
	VersionPattern string   `json:"version_pattern" db:"version_pattern"`
	CPEPattern    string    `json:"cpe_pattern" db:"cpe_pattern"`
	CVEIDs        StringArray `json:"cve_ids" db:"cve_ids"`
	Severity      string    `json:"severity" db:"severity"`
	Action        string    `json:"action" db:"action"`
	Enabled       bool      `json:"enabled" db:"enabled"`
	Priority      int       `json:"priority" db:"priority"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// AssetDetail 资产详情
type AssetDetail struct {
	ID            int       `json:"id" db:"id"`
	AssetID       int       `json:"asset_id" db:"asset_id"`
	OS            string    `json:"os" db:"os"`
	OSVersion     string    `json:"os_version" db:"os_version"`
	KernelVersion string    `json:"kernel_version" db:"kernel_version"`
	Architecture  string    `json:"architecture" db:"architecture"`
	Vendor        string    `json:"vendor" db:"vendor"`
	Product       string    `json:"product" db:"product"`
	Version       string    `json:"version" db:"version"`
	CPE           string    `json:"cpe" db:"cpe"`
	Fingerprints  JSONMapInterface   `json:"fingerprints" db:"fingerprints"`
	SoftwareList  JSONMapInterface   `json:"software_list" db:"software_list"`
	OpenPorts     JSONMapInterface   `json:"open_ports" db:"open_ports"`
	Services      JSONMapInterface   `json:"services" db:"services"`
	Vulnerabilities JSONMapInterface `json:"vulnerabilities" db:"vulnerabilities"`
	LastScanned   *time.Time `json:"last_scanned" db:"last_scanned"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// Alert 预警
type Alert struct {
	ID          int        `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Severity    string     `json:"severity" db:"severity"`
	Type        string     `json:"type" db:"type"`
	Source      string     `json:"source" db:"source"`
	Status      string     `json:"status" db:"status"`
	Data        JSONMapInterface    `json:"data" db:"data"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ReadAt      *time.Time `json:"read_at" db:"read_at"`
	ResolvedAt  *time.Time `json:"resolved_at" db:"resolved_at"`
}

// JSONMapInterface 用于处理JSONB字段（interface{}类型）
type JSONMapInterface map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONMapInterface) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONMapInterface) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	
	return json.Unmarshal(b, j)
}

// StringArray 用于处理字符串数组字段
type StringArray []string

// Value 实现 driver.Valuer 接口
func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan 实现 sql.Scanner 接口
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	
	return json.Unmarshal(b, s)
}

// VulnerabilityIntelligenceQuery 漏洞情报查询参数
type VulnerabilityIntelligenceQuery struct {
	Page            int      `form:"page" json:"page"`
	PageSize        int      `form:"pageSize" json:"pageSize"`
	Search          string   `form:"search" json:"search"`
	Severity        string   `form:"severity" json:"severity"`
	Source          string   `form:"source" json:"source"`
	CVEID           string   `form:"cve_id" json:"cve_id"`
	StartDate       string   `form:"start_date" json:"start_date"`
	EndDate         string   `form:"end_date" json:"end_date"`
	Is0Day          *bool    `form:"is_0day" json:"is_0day"`
	ExploitAvailable *bool   `form:"exploit_available" json:"exploit_available"`
	SortBy          string   `form:"sort_by" json:"sort_by"`
	SortOrder       string   `form:"sort_order" json:"sort_order"`
}

// DetectionResultQuery 检测结果查询参数
type DetectionResultQuery struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"pageSize" json:"pageSize"`
	Severity string `form:"severity" json:"severity"`
	Status   string `form:"status" json:"status"`
	Verified *bool  `form:"verified" json:"verified"`
	TaskID   int    `form:"task_id" json:"task_id"`
}

// AlertQuery 预警查询参数
type AlertQuery struct {
	Page      int       `form:"page" json:"page"`
	PageSize  int       `form:"pageSize" json:"pageSize"`
	Severity  string    `form:"severity" json:"severity"`
	Status    string    `form:"status" json:"status"`
	StartDate string    `form:"start_date" json:"start_date"`
	EndDate   string    `form:"end_date" json:"end_date"`
}

// SyncRequest 同步请求
type SyncRequest struct {
	Source string `json:"source" binding:"omitempty"`
	Force  bool   `json:"force"`
}

// DetectionTaskRequest 检测任务请求
type DetectionTaskRequest struct {
	Name          string                 `json:"name" binding:"required"`
	Description   string                 `json:"description"`
	Targets       []string               `json:"targets" binding:"required"`
	DetectionType string                 `json:"detection_type" binding:"required"`
	Config        map[string]interface{} `json:"config"`
	Schedule      map[string]interface{} `json:"schedule"`
	Notifications map[string]interface{} `json:"notifications"`
}

// AlertRequest 预警请求
type AlertRequest struct {
	Title       string                 `json:"title" binding:"required"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity" binding:"required"`
	Type        string                 `json:"type" binding:"required"`
	Data        map[string]interface{} `json:"data"`
	Channels    []string               `json:"channels"`
	Recipients  []string               `json:"recipients"`
}

// DetectionRuleRequest 检测规则请求
type DetectionRuleRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Type        string   `json:"type" binding:"required"`
	Condition   string   `json:"condition" binding:"required"`
	Action      string   `json:"action" binding:"required"`
	Severity    string   `json:"severity" binding:"required"`
	Enabled     bool     `json:"enabled"`
	Priority    int      `json:"priority"`
	Tags        []string `json:"tags"`
}

// IntelligenceSourceRequest 情报源请求
type IntelligenceSourceRequest struct {
	Name         string                 `json:"name" binding:"required"`
	URL          string                 `json:"url" binding:"required"`
	Type         string                 `json:"type" binding:"required"`
	Enabled      bool                   `json:"enabled"`
	SyncInterval int                    `json:"sync_interval"`
	Config       map[string]interface{} `json:"config"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Items    interface{} `json:"items"`
}

// StatsResponse 统计响应
type StatsResponse struct {
	TotalCount     int                    `json:"total_count"`
	TodayCount     int                    `json:"today_count"`
	WeekCount      int                    `json:"week_count"`
	MonthCount     int                    `json:"month_count"`
	BySeverity     map[string]int         `json:"by_severity"`
	BySource       map[string]int         `json:"by_source"`
	Trend          []VulnerabilityTrend   `json:"trend"`
}

// VulnerabilityTrend 漏洞趋势
type VulnerabilityTrend struct {
	Date          string `json:"date"`
	Total         int    `json:"total"`
	Critical      int    `json:"critical"`
	High          int    `json:"high"`
	Medium        int    `json:"medium"`
	Low           int    `json:"low"`
	ZeroDay       int    `json:"zero_day"`
	WithExploit   int    `json:"with_exploit"`
}

// DetectionStatsResponse 检测统计响应
type DetectionStatsResponse struct {
	Period            string                     `json:"period"`
	TotalTasks        int                        `json:"total_tasks"`
	TotalDetections   int                        `json:"total_detections"`
	ByStatus          map[string]int             `json:"by_status"`
	ByRiskLevel       map[string]int             `json:"by_risk_level"`
	TopVulnerabilities []TopVulnerability        `json:"top_vulnerabilities"`
	TopAssets         []TopAsset                 `json:"top_assets"`
}

// TopVulnerability 热门漏洞
type TopVulnerability struct {
	CVEID    string  `json:"cve_id"`
	Title    string  `json:"title"`
	Count    int     `json:"count"`
	Severity string  `json:"severity"`
}

// TopAsset 热门资产
type TopAsset struct {
	IP                string `json:"ip"`
	Hostname          string `json:"hostname"`
	VulnerabilityCount int    `json:"vulnerability_count"`
	CriticalCount     int    `json:"critical_count"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string                 `json:"status"`
	Components map[string]ComponentHealth `json:"components"`
	Metrics   map[string]interface{} `json:"metrics"`
	LastUpdated time.Time            `json:"last_updated"`
}

// ComponentHealth 组件健康状态
type ComponentHealth struct {
	Status  string      `json:"status"`
	Latency int         `json:"latency,omitempty"`
	Details interface{} `json:"details,omitempty"`
}