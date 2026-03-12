package detection

import (
	"GoAttack/common/log"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Service 检测服务
type Service struct {
	db    *sql.DB
	engine *Engine
}

// NewService 创建新的检测服务
func NewService(db *sql.DB) *Service {
	engine := NewEngine(db)
	return &Service{
		db:     db,
		engine: engine,
	}
}

// CreateTask 创建检测任务
func (s *Service) CreateTask(ctx context.Context, task CreateTaskRequest) (int, error) {
	log.Infof("创建检测任务: %s", task.Name)

	// 验证任务参数
	if err := s.validateTask(task); err != nil {
		return 0, err
	}

	// 序列化JSON字段
	targetsJSON, err := json.Marshal(task.Targets)
	if err != nil {
		return 0, fmt.Errorf("序列化目标列表失败: %v", err)
	}

	rulesJSON, err := json.Marshal(task.Rules)
	if err != nil {
		return 0, fmt.Errorf("序列化规则列表失败: %v", err)
	}

	configJSON, err := json.Marshal(task.Config)
	if err != nil {
		return 0, fmt.Errorf("序列化配置失败: %v", err)
	}

	// 插入任务
	var taskID int
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO detection_tasks (
			name, description, targets, target_type, scan_type,
			rules, config, status, progress, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`,
		task.Name,
		task.Description,
		string(targetsJSON),
		task.TargetType,
		task.ScanType,
		string(rulesJSON),
		string(configJSON),
		"pending",
		0,
		task.CreatedBy,
	).Scan(&taskID)

	if err != nil {
		return 0, fmt.Errorf("创建检测任务失败: %v", err)
	}

	log.Infof("检测任务创建成功: %s (ID: %d)", task.Name, taskID)
	return taskID, nil
}

// GetTask 获取任务详情
func (s *Service) GetTask(ctx context.Context, taskID int) (*DetectionTask, error) {
	task, err := s.engine.getTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// 获取任务结果
	results, err := s.engine.GetTaskResults(ctx, taskID)
	if err != nil {
		log.Errorf("获取任务结果失败: %v", err)
	} else {
		task.Results = results
	}

	return &task, nil
}

// GetTasks 获取任务列表
func (s *Service) GetTasks(ctx context.Context, query TaskQuery) ([]DetectionTask, int, error) {
	sqlQuery := `
		SELECT id, name, description, targets, target_type, scan_type,
		       rules, config, status, progress, created_by, created_at,
		       started_at, completed_at
		FROM detection_tasks
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if query.Search != "" {
		sqlQuery += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex)
		args = append(args, "%"+query.Search+"%")
		argIndex++
	}

	if query.Status != "" {
		sqlQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, query.Status)
		argIndex++
	}

	if query.CreatedBy != "" {
		sqlQuery += fmt.Sprintf(" AND created_by = $%d", argIndex)
		args = append(args, query.CreatedBy)
		argIndex++
	}

	if query.StartDate != nil {
		sqlQuery += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, query.StartDate)
		argIndex++
	}

	if query.EndDate != nil {
		sqlQuery += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, query.EndDate)
		argIndex++
	}

	// 排序
	switch query.SortBy {
	case "created_at":
		sqlQuery += " ORDER BY created_at"
	case "progress":
		sqlQuery += " ORDER BY progress"
	case "status":
		sqlQuery += " ORDER BY status"
	default:
		sqlQuery += " ORDER BY created_at DESC"
	}

	if query.SortOrder == "asc" {
		sqlQuery += " ASC"
	} else {
		sqlQuery += " DESC"
	}

	// 分页
	if query.PageSize > 0 {
		offset := (query.Page - 1) * query.PageSize
		sqlQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, query.PageSize, offset)
	}

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询检测任务失败: %v", err)
	}
	defer rows.Close()

	var tasks []DetectionTask
	for rows.Next() {
		var task DetectionTask
		var targetsJSON, rulesJSON, configJSON []byte
		var createdAt time.Time
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(
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
			log.Errorf("扫描检测任务行失败: %v", err)
			continue
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

		tasks = append(tasks, task)
	}

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM detection_tasks WHERE 1=1"
	countArgs := []interface{}{}
	countArgIndex := 1

	if query.Search != "" {
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", countArgIndex, countArgIndex)
		countArgs = append(countArgs, "%"+query.Search+"%")
		countArgIndex++
	}

	if query.Status != "" {
		countQuery += fmt.Sprintf(" AND status = $%d", countArgIndex)
		countArgs = append(countArgs, query.Status)
		countArgIndex++
	}

	if query.CreatedBy != "" {
		countQuery += fmt.Sprintf(" AND created_by = $%d", countArgIndex)
		countArgs = append(countArgs, query.CreatedBy)
		countArgIndex++
	}

	if query.StartDate != nil {
		countQuery += fmt.Sprintf(" AND created_at >= $%d", countArgIndex)
		countArgs = append(countArgs, query.StartDate)
		countArgIndex++
	}

	if query.EndDate != nil {
		countQuery += fmt.Sprintf(" AND created_at <= $%d", countArgIndex)
		countArgs = append(countArgs, query.EndDate)
		countArgIndex++
	}

	var total int
	if len(countArgs) > 0 {
		err = s.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	} else {
		err = s.db.QueryRowContext(ctx, countQuery).Scan(&total)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %v", err)
	}

	return tasks, total, nil
}

// StartTask 启动检测任务
func (s *Service) StartTask(ctx context.Context, taskID int) error {
	return s.engine.StartTask(ctx, taskID)
}

// StopTask 停止检测任务
func (s *Service) StopTask(ctx context.Context, taskID int) error {
	// 更新任务状态为停止
	_, err := s.db.ExecContext(ctx, `
		UPDATE detection_tasks 
		SET status = 'stopped', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status = 'running'
	`, taskID)
	
	if err != nil {
		return fmt.Errorf("停止任务失败: %v", err)
	}

	log.Infof("检测任务已停止: %d", taskID)
	return nil
}

// DeleteTask 删除检测任务
func (s *Service) DeleteTask(ctx context.Context, taskID int) error {
	// 先删除相关结果
	_, err := s.db.ExecContext(ctx, "DELETE FROM detection_results WHERE task_id = $1", taskID)
	if err != nil {
		return fmt.Errorf("删除检测结果失败: %v", err)
	}

	// 再删除任务
	_, err = s.db.ExecContext(ctx, "DELETE FROM detection_tasks WHERE id = $1", taskID)
	if err != nil {
		return fmt.Errorf("删除检测任务失败: %v", err)
	}

	log.Infof("检测任务已删除: %d", taskID)
	return nil
}

// GetTaskResults 获取任务结果
func (s *Service) GetTaskResults(ctx context.Context, taskID int, query ResultQuery) ([]DetectionResult, int, error) {
	// 先验证任务是否存在
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM detection_tasks WHERE id = $1)", taskID).Scan(&exists)
	if err != nil || !exists {
		return nil, 0, fmt.Errorf("任务不存在: %d", taskID)
	}

	// 获取结果
	results, err := s.engine.GetTaskResults(ctx, taskID)
	if err != nil {
		return nil, 0, err
	}

	// 应用过滤
	var filteredResults []DetectionResult
	for _, result := range results {
		if query.Severity != "" && result.Severity != query.Severity {
			continue
		}
		if query.Status != "" && result.Status != query.Status {
			continue
		}
		filteredResults = append(filteredResults, result)
	}

	// 分页
	total := len(filteredResults)
	start := (query.Page - 1) * query.PageSize
	end := start + query.PageSize

	if start >= total {
		return []DetectionResult{}, total, nil
	}
	if end > total {
		end = total
	}

	return filteredResults[start:end], total, nil
}

// GetTaskStats 获取任务统计信息
func (s *Service) GetTaskStats(ctx context.Context, taskID int) (map[string]interface{}, error) {
	return s.engine.GetTaskStats(ctx, taskID)
}

// GetDetectionRules 获取检测规则
func (s *Service) GetDetectionRules(ctx context.Context, query RuleQuery) ([]DetectionRule, int, error) {
	sqlQuery := `
		SELECT id, name, description, type, condition, action, severity,
		       enabled, priority, tags, config, created_at, updated_at
		FROM detection_rules
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if query.Search != "" {
		sqlQuery += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex)
		args = append(args, "%"+query.Search+"%")
		argIndex++
	}

	if query.Type != "" {
		sqlQuery += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, query.Type)
		argIndex++
	}

	if query.Severity != "" {
		sqlQuery += fmt.Sprintf(" AND severity = $%d", argIndex)
		args = append(args, query.Severity)
		argIndex++
	}

	if query.Enabled != nil {
		sqlQuery += fmt.Sprintf(" AND enabled = $%d", argIndex)
		args = append(args, *query.Enabled)
		argIndex++
	}

	// 排序
	switch query.SortBy {
	case "priority":
		sqlQuery += " ORDER BY priority"
	case "severity":
		sqlQuery += " ORDER BY CASE severity WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 WHEN 'low' THEN 4 ELSE 5 END"
	default:
		sqlQuery += " ORDER BY created_at DESC"
	}

	if query.SortOrder == "asc" {
		sqlQuery += " ASC"
	} else {
		sqlQuery += " DESC"
	}

	// 分页
	if query.PageSize > 0 {
		offset := (query.Page - 1) * query.PageSize
		sqlQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, query.PageSize, offset)
	}

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询检测规则失败: %v", err)
	}
	defer rows.Close()

	var rules []DetectionRule
	for rows.Next() {
		var rule DetectionRule
		var tagsJSON, configJSON []byte
		var createdAt, updatedAt time.Time

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
			&createdAt,
			&updatedAt,
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

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM detection_rules WHERE 1=1"
	countArgs := []interface{}{}
	countArgIndex := 1

	if query.Search != "" {
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", countArgIndex, countArgIndex)
		countArgs = append(countArgs, "%"+query.Search+"%")
		countArgIndex++
	}

	if query.Type != "" {
		countQuery += fmt.Sprintf(" AND type = $%d", countArgIndex)
		countArgs = append(countArgs, query.Type)
		countArgIndex++
	}

	if query.Severity != "" {
		countQuery += fmt.Sprintf(" AND severity = $%d", countArgIndex)
		countArgs = append(countArgs, query.Severity)
		countArgIndex++
	}

	if query.Enabled != nil {
		countQuery += fmt.Sprintf(" AND enabled = $%d", countArgIndex)
		countArgs = append(countArgs, *query.Enabled)
		countArgIndex++
	}

	var total int
	if len(countArgs) > 0 {
		err = s.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	} else {
		err = s.db.QueryRowContext(ctx, countQuery).Scan(&total)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %v", err)
	}

	return rules, total, nil
}

// CreateDetectionRule 创建检测规则
func (s *Service) CreateDetectionRule(ctx context.Context, rule CreateRuleRequest) (int, error) {
	// 验证规则参数
	if err := s.validateRule(rule); err != nil {
		return 0, err
	}

	// 序列化JSON字段
	tagsJSON, err := json.Marshal(rule.Tags)
	if err != nil {
		return 0, fmt.Errorf("序列化标签失败: %v", err)
	}

	configJSON, err := json.Marshal(rule.Config)
	if err != nil {
		return 0, fmt.Errorf("序列化配置失败: %v", err)
	}

	// 插入规则
	var ruleID int
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO detection_rules (
			name, description, type, condition, action, severity,
			enabled, priority, tags, config
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`,
		rule.Name,
		rule.Description,
		rule.Type,
		rule.Condition,
		rule.Action,
		rule.Severity,
		rule.Enabled,
		rule.Priority,
		string(tagsJSON),
		string(configJSON),
	).Scan(&ruleID)

	if err != nil {
		return 0, fmt.Errorf("创建检测规则失败: %v", err)
	}

	log.Infof("检测规则创建成功: %s (ID: %d)", rule.Name, ruleID)
	return ruleID, nil
}

// UpdateDetectionRule 更新检测规则
func (s *Service) UpdateDetectionRule(ctx context.Context, ruleID int, updates map[string]interface{}) error {
	// 构建更新语句
	query := "UPDATE detection_rules SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{}
	argIndex := 1

	if name, ok := updates["name"].(string); ok && name != "" {
		query += fmt.Sprintf(", name = $%d", argIndex)
		args = append(args, name)
		argIndex++
	}

	if description, ok := updates["description"].(string); ok {
		query += fmt.Sprintf(", description = $%d", argIndex)
		args = append(args, description)
		argIndex++
	}

	if ruleType, ok := updates["type"].(string); ok && ruleType != "" {
		query += fmt.Sprintf(", type = $%d", argIndex)
		args = append(args, ruleType)
		argIndex++
	}

	if condition, ok := updates["condition"].(string); ok && condition != "" {
		query += fmt.Sprintf(", condition = $%d", argIndex)
		args = append(args, condition)
		argIndex++
	}

	if action, ok := updates["action"].(string); ok && action != "" {
		query += fmt.Sprintf(", action = $%d", argIndex)
		args = append(args, action)
		argIndex++
	}

	if severity, ok := updates["severity"].(string); ok && severity != "" {
		query += fmt.Sprintf(", severity = $%d", argIndex)
		args = append(args, severity)
		argIndex++
	}

	if enabled, ok := updates["enabled"].(bool); ok {
		query += fmt.Sprintf(", enabled = $%d", argIndex)
		args = append(args, enabled)
		argIndex++
	}

	if priority, ok := updates["priority"].(int); ok {
		query += fmt.Sprintf(", priority = $%d", argIndex)
		args = append(args, priority)
		argIndex++
	}

	if tags, ok := updates["tags"].([]string); ok {
		tagsJSON, err := json.Marshal(tags)
		if err != nil {
			return fmt.Errorf("序列化标签失败: %v", err)
		}
		query += fmt.Sprintf(", tags = $%d", argIndex)
		args = append(args, string(tagsJSON))
		argIndex++
	}

	if config, ok := updates["config"].(map[string]interface{}); ok {
		configJSON, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("序列化配置失败: %v", err)
		}
		query += fmt.Sprintf(", config = $%d", argIndex)
		args = append(args, string(configJSON))
		argIndex++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, ruleID)

	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("更新检测规则失败: %v", err)
	}

	log.Infof("检测规则更新成功: %d", ruleID)
	return nil
}

// DeleteDetectionRule 删除检测规则
func (s *Service) DeleteDetectionRule(ctx context.Context, ruleID int) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM detection_rules WHERE id = $1", ruleID)
	if err != nil {
		return fmt.Errorf("删除检测规则失败: %v", err)
	}

	log.Infof("检测规则删除成功: %d", ruleID)
	return nil
}

// 验证任务参数
func (s *Service) validateTask(task CreateTaskRequest) error {
	if task.Name == "" {
		return fmt.Errorf("任务名称不能为空")
	}
	if len(task.Targets) == 0 {
		return fmt.Errorf("目标列表不能为空")
	}
	if task.TargetType == "" {
		return fmt.Errorf("目标类型不能为空")
	}
	if task.ScanType == "" {
		return fmt.Errorf("扫描类型不能为空")
	}
	if task.CreatedBy == "" {
		return fmt.Errorf("创建者不能为空")
	}
	return nil
}

// 验证规则参数
func (s *Service) validateRule(rule CreateRuleRequest) error {
	if rule.Name == "" {
		return fmt.Errorf("规则名称不能为空")
	}
	if rule.Type == "" {
		return fmt.Errorf("规则类型不能为空")
	}
	if rule.Condition == "" {
		return fmt.Errorf("检测条件不能为空")
	}
	if rule.Action == "" {
		return fmt.Errorf("动作不能为空")
	}
	if rule.Severity == "" {
		return fmt.Errorf("严重程度不能为空")
	}
	return nil
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Targets     []string               `json:"targets"`
	TargetType  string                 `json:"target_type"`
	ScanType    string                 `json:"scan_type"`
	Rules       []DetectionRule        `json:"rules"`
	Config      map[string]interface{} `json:"config"`
	CreatedBy   string                 `json:"created_by"`
}

// TaskQuery 任务查询参数
type TaskQuery struct {
	Search    string
	Status    string
	CreatedBy string
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	PageSize  int
	SortBy    string
	SortOrder string
}

// ResultQuery 结果查询参数
type ResultQuery struct {
	Severity string
	Status   string
	Page     int
	PageSize int
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

// CreateRuleRequest 创建规则请求
type CreateRuleRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Condition   string                 `json:"condition"`
	Action      string                 `json:"action"`
	Severity    string                 `json:"severity"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	Tags        []string               `json:"tags"`
	Config      map[string]interface{} `json:"config"`
}