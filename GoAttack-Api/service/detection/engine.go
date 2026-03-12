package detection

import (
	"GoAttack/common/log"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Engine 漏洞检测引擎
type Engine struct {
	db *sql.DB
}

// NewEngine 创建新的检测引擎
func NewEngine(db *sql.DB) *Engine {
	return &Engine{
		db: db,
	}
}

// DetectionTask 检测任务
type DetectionTask struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Targets     []string               `json:"targets"`
	TargetType  string                 `json:"target_type"` // host, network, web, api
	ScanType    string                 `json:"scan_type"`   // full, quick, custom
	Rules       []DetectionRule        `json:"rules"`
	Config      map[string]interface{} `json:"config"`
	Status      string                 `json:"status"`      // pending, running, completed, failed
	Progress    int                    `json:"progress"`    // 0-100
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at"`
	Results     []DetectionResult      `json:"results"`
}

// DetectionRule 检测规则
type DetectionRule struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`        // port, service, version, cve, custom
	Condition   string                 `json:"condition"`   // 检测条件表达式
	Action      string                 `json:"action"`      // alert, block, log, report
	Severity    string                 `json:"severity"`    // critical, high, medium, low
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	Tags        []string               `json:"tags"`
	Config      map[string]interface{} `json:"config"`
}

// DetectionResult 检测结果
type DetectionResult struct {
	ID          int                    `json:"id"`
	TaskID      int                    `json:"task_id"`
	Target      string                 `json:"target"`
	RuleID      int                    `json:"rule_id"`
	RuleName    string                 `json:"rule_name"`
	VulnerabilityID *int               `json:"vulnerability_id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`
	Evidence    map[string]interface{} `json:"evidence"`
	Status      string                 `json:"status"`      // detected, false_positive, fixed
	CreatedAt   time.Time              `json:"created_at"`
}

// StartTask 启动检测任务
func (e *Engine) StartTask(ctx context.Context, taskID int) error {
	log.Infof("启动检测任务: %d", taskID)

	// 获取任务信息
	task, err := e.getTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("获取任务失败: %v", err)
	}

	// 更新任务状态为运行中
	if err := e.updateTaskStatus(ctx, taskID, "running", 0); err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}

	// 启动异步检测
	go e.executeTask(ctx, task)

	return nil
}

// executeTask 执行检测任务
func (e *Engine) executeTask(ctx context.Context, task DetectionTask) {
	log.Infof("开始执行检测任务: %s (ID: %d)", task.Name, task.ID)

	// 记录开始时间
	now := time.Now()
	if err := e.updateTaskStartedAt(ctx, task.ID, now); err != nil {
		log.Errorf("更新任务开始时间失败: %v", err)
	}

	totalTargets := len(task.Targets)
	completedTargets := 0

	// 遍历所有目标
	for _, target := range task.Targets {
		select {
		case <-ctx.Done():
			log.Infof("检测任务被取消: %s (ID: %d)", task.Name, task.ID)
			_ = e.updateTaskStatus(ctx, task.ID, "failed", completedTargets*100/totalTargets)
			return
		default:
			// 对每个目标执行检测
			results := e.detectTarget(ctx, target, task)
			
			// 保存检测结果
			if err := e.saveResults(ctx, task.ID, target, results); err != nil {
				log.Errorf("保存检测结果失败: %v", err)
			}

			completedTargets++
			progress := completedTargets * 100 / totalTargets
			
			// 更新进度
			if err := e.updateTaskProgress(ctx, task.ID, progress); err != nil {
				log.Errorf("更新任务进度失败: %v", err)
			}
		}
	}

	// 标记任务完成
	if err := e.updateTaskStatus(ctx, task.ID, "completed", 100); err != nil {
		log.Errorf("更新任务状态失败: %v", err)
	}

	// 记录完成时间
	completedAt := time.Now()
	if err := e.updateTaskCompletedAt(ctx, task.ID, completedAt); err != nil {
		log.Errorf("更新任务完成时间失败: %v", err)
	}

	log.Infof("检测任务完成: %s (ID: %d)", task.Name, task.ID)
}

// detectTarget 检测单个目标
func (e *Engine) detectTarget(ctx context.Context, target string, task DetectionTask) []DetectionResult {
	var results []DetectionResult

	// 根据目标类型执行不同的检测
	switch task.TargetType {
	case "host":
		results = e.detectHost(ctx, target, task)
	case "network":
		results = e.detectNetwork(ctx, target, task)
	case "web":
		results = e.detectWeb(ctx, target, task)
	case "api":
		results = e.detectAPI(ctx, target, task)
	default:
		log.Warnf("未知的目标类型: %s", task.TargetType)
	}

	return results
}

// detectHost 检测主机
func (e *Engine) detectHost(ctx context.Context, target string, task DetectionTask) []DetectionResult {
	log.Debugf("检测主机: %s", target)
	var results []DetectionResult

	// 获取启用的规则
	rules, err := e.getEnabledRules(ctx)
	if err != nil {
		log.Errorf("获取检测规则失败: %v", err)
		return results
	}

	// 执行端口扫描
	openPorts := e.scanPorts(ctx, target)
	
	// 对每个开放端口执行服务检测
	for _, port := range openPorts {
		serviceInfo := e.detectService(ctx, target, port)
		
		// 应用规则
		for _, rule := range rules {
			if e.matchRule(rule, target, port, serviceInfo) {
				result := DetectionResult{
					TaskID:      task.ID,
					Target:      target,
					RuleID:      rule.ID,
					RuleName:    rule.Name,
					Title:       fmt.Sprintf("%s 漏洞检测", rule.Name),
					Description: rule.Description,
					Severity:    rule.Severity,
					Evidence: map[string]interface{}{
						"target":      target,
						"port":        port,
						"service":     serviceInfo,
						"rule_type":   rule.Type,
						"rule_config": rule.Config,
					},
					Status:    "detected",
					CreatedAt: time.Now(),
				}
				results = append(results, result)
			}
		}
	}

	return results
}

// detectNetwork 检测网络
func (e *Engine) detectNetwork(ctx context.Context, target string, task DetectionTask) []DetectionResult {
	log.Debugf("检测网络: %s", target)
	// 网络范围检测实现
	return []DetectionResult{}
}

// detectWeb 检测Web应用
func (e *Engine) detectWeb(ctx context.Context, target string, task DetectionTask) []DetectionResult {
	log.Debugf("检测Web应用: %s", target)
	var results []DetectionResult

	// 获取启用的规则
	rules, err := e.getEnabledRules(ctx)
	if err != nil {
		log.Errorf("获取检测规则失败: %v", err)
		return results
	}

	// 执行Web扫描
	webInfo := e.scanWeb(ctx, target)
	
	// 应用规则
	for _, rule := range rules {
		if e.matchWebRule(rule, target, webInfo) {
			result := DetectionResult{
				TaskID:      task.ID,
				Target:      target,
				RuleID:      rule.ID,
				RuleName:    rule.Name,
				Title:       fmt.Sprintf("%s Web漏洞检测", rule.Name),
				Description: rule.Description,
				Severity:    rule.Severity,
				Evidence: map[string]interface{}{
					"target":      target,
					"web_info":    webInfo,
					"rule_type":   rule.Type,
					"rule_config": rule.Config,
				},
				Status:    "detected",
				CreatedAt: time.Now(),
			}
			results = append(results, result)
		}
	}

	return results
}

// detectAPI 检测API
func (e *Engine) detectAPI(ctx context.Context, target string, task DetectionTask) []DetectionResult {
	log.Debugf("检测API: %s", target)
	// API安全检测实现
	return []DetectionResult{}
}

// scanPorts 扫描端口
func (e *Engine) scanPorts(ctx context.Context, target string) []int {
	// 这里可以集成现有的端口扫描功能
	// 暂时返回模拟数据
	return []int{22, 80, 443, 3306, 5432, 8080}
}

// detectService 检测服务信息
func (e *Engine) detectService(ctx context.Context, target string, port int) map[string]interface{} {
	// 这里可以集成现有的服务识别功能
	// 暂时返回模拟数据
	services := map[int]map[string]interface{}{
		22:   {"name": "ssh", "version": "OpenSSH 8.9", "banner": "SSH-2.0-OpenSSH_8.9"},
		80:   {"name": "http", "version": "nginx/1.18.0", "banner": "Server: nginx/1.18.0"},
		443:  {"name": "https", "version": "nginx/1.18.0", "banner": "Server: nginx/1.18.0"},
		3306: {"name": "mysql", "version": "8.0.33", "banner": "8.0.33"},
		5432: {"name": "postgresql", "version": "14.0", "banner": "PostgreSQL 14.0"},
		8080: {"name": "http", "version": "Apache Tomcat/9.0", "banner": "Apache Tomcat/9.0"},
	}

	if service, ok := services[port]; ok {
		return service
	}
	return map[string]interface{}{"name": "unknown", "version": "unknown"}
}

// scanWeb 扫描Web应用
func (e *Engine) scanWeb(ctx context.Context, target string) map[string]interface{} {
	// 这里可以集成现有的Web指纹识别功能
	// 暂时返回模拟数据
	return map[string]interface{}{
		"server":       "nginx/1.18.0",
		"framework":    "Vue.js",
		"cms":          "WordPress",
		"version":      "6.0",
		"technologies": []string{"JavaScript", "PHP", "MySQL"},
		"headers": map[string]string{
			"Server":         "nginx/1.18.0",
			"X-Powered-By":   "PHP/8.1",
			"Content-Type":   "text/html; charset=UTF-8",
		},
	}
}

// matchRule 匹配规则
func (e *Engine) matchRule(rule DetectionRule, target string, port int, serviceInfo map[string]interface{}) bool {
	// 根据规则类型进行匹配
	switch rule.Type {
	case "port":
		return e.matchPortRule(rule, port)
	case "service":
		return e.matchServiceRule(rule, serviceInfo)
	case "version":
		return e.matchVersionRule(rule, serviceInfo)
	case "cve":
		return e.matchCVERule(rule, target, serviceInfo)
	case "custom":
		return e.matchCustomRule(rule, target, port, serviceInfo)
	default:
		return false
	}
}

// matchPortRule 匹配端口规则
func (e *Engine) matchPortRule(rule DetectionRule, port int) bool {
	// 解析条件，例如: "port == 22 || port == 3389"
	// 这里简化实现，实际应该解析条件表达式
	if strings.Contains(rule.Condition, fmt.Sprintf("port == %d", port)) {
		return true
	}
	return false
}

// matchServiceRule 匹配服务规则
func (e *Engine) matchServiceRule(rule DetectionRule, serviceInfo map[string]interface{}) bool {
	serviceName, ok := serviceInfo["name"].(string)
	if !ok {
		return false
	}

	// 解析条件，例如: "service == 'mysql' || service == 'postgresql'"
	if strings.Contains(rule.Condition, fmt.Sprintf("service == '%s'", serviceName)) {
		return true
	}
	return false
}

// matchVersionRule 匹配版本规则
func (e *Engine) matchVersionRule(rule DetectionRule, serviceInfo map[string]interface{}) bool {
	version, ok := serviceInfo["version"].(string)
	if !ok {
		return false
	}

	// 解析条件，例如: "version < '8.0'"
	// 这里简化实现，实际应该解析版本比较表达式
	if strings.Contains(rule.Condition, version) {
		return true
	}
	return false
}

// matchCVERule 匹配CVE规则
func (e *Engine) matchCVERule(rule DetectionRule, target string, serviceInfo map[string]interface{}) bool {
	// 查询漏洞情报数据库，检查目标是否存在相关CVE漏洞
	// 这里简化实现
	return false
}

// matchWebRule 匹配Web规则
func (e *Engine) matchWebRule(rule DetectionRule, target string, webInfo map[string]interface{}) bool {
	// 根据Web信息匹配规则
	// 这里简化实现
	return false
}

// matchCustomRule 匹配自定义规则
func (e *Engine) matchCustomRule(rule DetectionRule, target string, port int, serviceInfo map[string]interface{}) bool {
	// 执行自定义规则条件
	// 这里简化实现
	return false
}

// getEnabledRules 获取启用的规则
func (e *Engine) getEnabledRules(ctx context.Context) ([]DetectionRule, error) {
	rows, err := e.db.QueryContext(ctx, `
		SELECT id, name, description, type, condition, action, severity, 
		       enabled, priority, tags, config
		FROM detection_rules
		WHERE enabled = true
		ORDER BY priority ASC, id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("查询检测规则失败: %v", err)
	}
	defer rows.Close()

	var rules []DetectionRule
	for rows.Next() {
		var rule DetectionRule
		var tagsJSON, configJSON []byte

		err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.Description,
			&rule.Type,
			&rule.Condition,
			&rule.Action,
			&rule.Severity,
			&rule.Enabled,
			&rule.Priority,
			&tagsJSON,
			&configJSON,
		)
		if err != nil {
			log.Errorf("扫描检测规则行失败: %v", err)
			continue
		}

		// 解析JSON字段
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &rule.Tags)
		}
		if len(configJSON) > 0 {
			json.Unmarshal(configJSON, &rule.Config)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// getTaskByID 根据ID获取任务
func (e *Engine) getTaskByID(ctx context.Context, taskID int) (DetectionTask, error) {
	var task DetectionTask
	var targetsJSON, rulesJSON, configJSON []byte
	var createdAt time.Time
	var startedAt, completedAt sql.NullTime

	err := e.db.QueryRowContext(ctx, `
		SELECT id, name, description, targets, target_type, scan_type, 
		       rules, config, status, progress, created_by, created_at,
		       started_at, completed_at
		FROM detection_tasks
		WHERE id = $1
	`, taskID).Scan(
		&task.ID,
		&task.Name,
		&task.Description,
		&targetsJSON,
		&task.TargetType,
		&task.ScanType,
		&rulesJSON,
		&configJSON,
		&task.Status,
		&task.Progress,
		&task.CreatedBy,
		&createdAt,
		&startedAt,
		&completedAt,
	)
	if err != nil {
		return task, fmt.Errorf("查询检测任务失败: %v", err)
	}

	// 解析JSON字段
	if len(targetsJSON) > 0 {
		json.Unmarshal(targetsJSON, &task.Targets)
	}
	if len(rulesJSON) > 0 {
		json.Unmarshal(rulesJSON, &task.Rules)
	}
	if len(configJSON) > 0 {
		json.Unmarshal(configJSON, &task.Config)
	}

	task.CreatedAt = createdAt
	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	return task, nil
}

// updateTaskStatus 更新任务状态
func (e *Engine) updateTaskStatus(ctx context.Context, taskID int, status string, progress int) error {
	_, err := e.db.ExecContext(ctx, `
		UPDATE detection_tasks 
		SET status = $1, progress = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`, status, progress, taskID)
	return err
}

// updateTaskProgress 更新任务进度
func (e *Engine) updateTaskProgress(ctx context.Context, taskID int, progress int) error {
	_, err := e.db.ExecContext(ctx, `
		UPDATE detection_tasks 
		SET progress = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, progress, taskID)
	return err
}

// updateTaskStartedAt 更新任务开始时间
func (e *Engine) updateTaskStartedAt(ctx context.Context, taskID int, startedAt time.Time) error {
	_, err := e.db.ExecContext(ctx, `
		UPDATE detection_tasks 
		SET started_at = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, startedAt, taskID)
	return err
}

// updateTaskCompletedAt 更新任务完成时间
func (e *Engine) updateTaskCompletedAt(ctx context.Context, taskID int, completedAt time.Time) error {
	_, err := e.db.ExecContext(ctx, `
		UPDATE detection_tasks 
		SET completed_at = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, completedAt, taskID)
	return err
}

// saveResults 保存检测结果
func (e *Engine) saveResults(ctx context.Context, taskID int, target string, results []DetectionResult) error {
	if len(results) == 0 {
		return nil
	}

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	for _, result := range results {
		evidenceJSON, _ := json.Marshal(result.Evidence)
		
		_, err := tx.ExecContext(ctx, `
			INSERT INTO detection_results (
				task_id, target, rule_id, rule_name, vulnerability_id,
				title, description, severity, evidence, status
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`,
			taskID,
			target,
			result.RuleID,
			result.RuleName,
			result.VulnerabilityID,
			result.Title,
			result.Description,
			result.Severity,
			string(evidenceJSON),
			result.Status,
		)
		if err != nil {
			log.Errorf("保存检测结果失败: %v", err)
			continue
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	log.Infof("保存了 %d 条检测结果", len(results))
	return nil
}

// GetTaskResults 获取任务结果
func (e *Engine) GetTaskResults(ctx context.Context, taskID int) ([]DetectionResult, error) {
	rows, err := e.db.QueryContext(ctx, `
		SELECT id, task_id, target, rule_id, rule_name, vulnerability_id,
		       title, description, severity, evidence, status, created_at
		FROM detection_results
		WHERE task_id = $1
		ORDER BY severity DESC, created_at DESC
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("查询检测结果失败: %v", err)
	}
	defer rows.Close()

	var results []DetectionResult
	for rows.Next() {
		var result DetectionResult
		var evidenceJSON []byte
		var vulnerabilityID sql.NullInt64
		var createdAt time.Time

		err := rows.Scan(
			&result.ID,
			&result.TaskID,
			&result.Target,
			&result.RuleID,
			&result.RuleName,
			&vulnerabilityID,
			&result.Title,
			&result.Description,
			&result.Severity,
			&evidenceJSON,
			&result.Status,
			&createdAt,
		)
		if err != nil {
			log.Errorf("扫描检测结果行失败: %v", err)
			continue
		}

		// 解析JSON字段
		if len(evidenceJSON) > 0 {
			json.Unmarshal(evidenceJSON, &result.Evidence)
		}

		if vulnerabilityID.Valid {
			id := int(vulnerabilityID.Int64)
			result.VulnerabilityID = &id
		}

		result.CreatedAt = createdAt
		results = append(results, result)
	}

	return results, nil
}

// GetTaskStats 获取任务统计信息
func (e *Engine) GetTaskStats(ctx context.Context, taskID int) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取任务基本信息
	var taskName, status string
	var progress int
	err := e.db.QueryRowContext(ctx, `
		SELECT name, status, progress
		FROM detection_tasks
		WHERE id = $1
	`, taskID).Scan(&taskName, &status, &progress)
	if err != nil {
		return nil, fmt.Errorf("查询任务信息失败: %v", err)
	}

	stats["task_name"] = taskName
	stats["status"] = status
	stats["progress"] = progress

	// 按严重程度统计结果
	severityStats := make(map[string]int)
	rows, err := e.db.QueryContext(ctx, `
		SELECT severity, COUNT(*) 
		FROM detection_results
		WHERE task_id = $1
		GROUP BY severity
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("查询严重程度统计失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var severity string
		var count int
		if err := rows.Scan(&severity, &count); err == nil {
			severityStats[severity] = count
		}
	}
	stats["by_severity"] = severityStats

	// 按规则统计结果
	ruleStats := make(map[string]int)
	rows, err = e.db.QueryContext(ctx, `
		SELECT rule_name, COUNT(*) 
		FROM detection_results
		WHERE task_id = $1
		GROUP BY rule_name
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("查询规则统计失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ruleName string
		var count int
		if err := rows.Scan(&ruleName, &count); err == nil {
			ruleStats[ruleName] = count
		}
	}
	stats["by_rule"] = ruleStats

	// 按目标统计结果
	targetStats := make(map[string]int)
	rows, err = e.db.QueryContext(ctx, `
		SELECT target, COUNT(*) 
		FROM detection_results
		WHERE task_id = $1
		GROUP BY target
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("查询目标统计失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var target string
		var count int
		if err := rows.Scan(&target, &count); err == nil {
			targetStats[target] = count
		}
	}
	stats["by_target"] = targetStats

	// 总结果数
	var total int
	err = e.db.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM detection_results
		WHERE task_id = $1
	`, taskID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("查询总数失败: %v", err)
	}
	stats["total"] = total

	return stats, nil
}