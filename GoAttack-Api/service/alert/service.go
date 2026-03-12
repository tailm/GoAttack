package alert

import (
	"GoAttack/common/log"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Service 预警通知服务
type Service struct {
	db *sql.DB
}

// NewService 创建新的预警通知服务
func NewService(db *sql.DB) *Service {
	return &Service{
		db: db,
	}
}

// CreateAlert 创建预警
func (s *Service) CreateAlert(ctx context.Context, req CreateAlertRequest) (int, error) {
	log.Infof("创建预警: %s", req.Title)

	// 验证请求参数
	if err := s.validateAlertRequest(req); err != nil {
		return 0, err
	}

	// 序列化JSON字段
	detailsJSON, err := json.Marshal(req.Details)
	if err != nil {
		return 0, fmt.Errorf("序列化详细信息失败: %v", err)
	}

	metadataJSON, err := json.Marshal(req.Metadata)
	if err != nil {
		return 0, fmt.Errorf("序列化元数据失败: %v", err)
	}

	// 插入预警
	var alertID int
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO alerts (
			title, description, type, severity, status, source, source_id,
			target, target_type, details, metadata, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`,
		req.Title,
		req.Description,
		req.Type,
		req.Severity,
		"pending",
		req.Source,
		req.SourceID,
		req.Target,
		req.TargetType,
		string(detailsJSON),
		string(metadataJSON),
		req.CreatedBy,
	).Scan(&alertID)

	if err != nil {
		return 0, fmt.Errorf("创建预警失败: %v", err)
	}

	// 触发通知
	go s.triggerNotifications(ctx, alertID)

	log.Infof("预警创建成功: %s (ID: %d)", req.Title, alertID)
	return alertID, nil
}

// GetAlert 获取预警详情
func (s *Service) GetAlert(ctx context.Context, alertID int) (*Alert, error) {
	var alert Alert
	var detailsJSON, metadataJSON []byte
	var createdAt, updatedAt time.Time
	var resolvedAt, closedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, `
		SELECT id, title, description, type, severity, status, source, source_id,
		       target, target_type, details, metadata, created_by, assigned_to,
		       created_at, updated_at, resolved_at, closed_at
		FROM alerts
		WHERE id = $1
	`, alertID).Scan(
		&alert.ID,
		&alert.Title,
		&alert.Description,
		&alert.Type,
		&alert.Severity,
		&alert.Status,
		&alert.Source,
		&alert.SourceID,
		&alert.Target,
		&alert.TargetType,
		&detailsJSON,
		&metadataJSON,
		&alert.CreatedBy,
		&alert.AssignedTo,
		&createdAt,
		&updatedAt,
		&resolvedAt,
		&closedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("预警不存在: %d", alertID)
		}
		return nil, fmt.Errorf("查询预警失败: %v", err)
	}

	// 解析JSON字段
	if len(detailsJSON) > 0 {
		json.Unmarshal(detailsJSON, &alert.Details)
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &alert.Metadata)
	}

	alert.CreatedAt = createdAt
	alert.UpdatedAt = updatedAt
	if resolvedAt.Valid {
		alert.ResolvedAt = &resolvedAt.Time
	}
	if closedAt.Valid {
		alert.ClosedAt = &closedAt.Time
	}

	return &alert, nil
}

// GetAlerts 获取预警列表
func (s *Service) GetAlerts(ctx context.Context, query AlertQuery) ([]Alert, int, error) {
	sqlQuery := `
		SELECT id, title, description, type, severity, status, source, source_id,
		       target, target_type, details, metadata, created_by, assigned_to,
		       created_at, updated_at, resolved_at, closed_at
		FROM alerts
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if query.Search != "" {
		sqlQuery += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex)
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

	if query.Status != "" {
		sqlQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, query.Status)
		argIndex++
	}

	if query.Source != "" {
		sqlQuery += fmt.Sprintf(" AND source = $%d", argIndex)
		args = append(args, query.Source)
		argIndex++
	}

	if query.Target != "" {
		sqlQuery += fmt.Sprintf(" AND target ILIKE $%d", argIndex)
		args = append(args, "%"+query.Target+"%")
		argIndex++
	}

	if query.TargetType != "" {
		sqlQuery += fmt.Sprintf(" AND target_type = $%d", argIndex)
		args = append(args, query.TargetType)
		argIndex++
	}

	if query.CreatedBy != "" {
		sqlQuery += fmt.Sprintf(" AND created_by = $%d", argIndex)
		args = append(args, query.CreatedBy)
		argIndex++
	}

	if query.AssignedTo != "" {
		sqlQuery += fmt.Sprintf(" AND assigned_to = $%d", argIndex)
		args = append(args, query.AssignedTo)
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
	case "severity":
		sqlQuery += " ORDER BY CASE severity WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 WHEN 'low' THEN 4 ELSE 5 END"
	case "created_at":
		sqlQuery += " ORDER BY created_at"
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
		return nil, 0, fmt.Errorf("查询预警列表失败: %v", err)
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var alert Alert
		var detailsJSON, metadataJSON []byte
		var createdAt, updatedAt time.Time
		var resolvedAt, closedAt sql.NullTime

		err := rows.Scan(
			&alert.ID,
			&alert.Title,
			&alert.Description,
			&alert.Type,
			&alert.Severity,
			&alert.Status,
			&alert.Source,
			&alert.SourceID,
			&alert.Target,
			&alert.TargetType,
			&detailsJSON,
			&metadataJSON,
			&alert.CreatedBy,
			&alert.AssignedTo,
			&createdAt,
			&updatedAt,
			&resolvedAt,
			&closedAt,
		)
		if err != nil {
			log.Errorf("扫描预警行失败: %v", err)
			continue
		}

		// 解析JSON字段
		if len(detailsJSON) > 0 {
			json.Unmarshal(detailsJSON, &alert.Details)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &alert.Metadata)
		}

		alert.CreatedAt = createdAt
		alert.UpdatedAt = updatedAt
		if resolvedAt.Valid {
			alert.ResolvedAt = &resolvedAt.Time
		}
		if closedAt.Valid {
			alert.ClosedAt = &closedAt.Time
		}

		alerts = append(alerts, alert)
	}

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM alerts WHERE 1=1"
	countArgs := []interface{}{}
	countArgIndex := 1

	if query.Search != "" {
		countQuery += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", countArgIndex, countArgIndex)
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

	if query.Status != "" {
		countQuery += fmt.Sprintf(" AND status = $%d", countArgIndex)
		countArgs = append(countArgs, query.Status)
		countArgIndex++
	}

	if query.Source != "" {
		countQuery += fmt.Sprintf(" AND source = $%d", countArgIndex)
		countArgs = append(countArgs, query.Source)
		countArgIndex++
	}

	if query.Target != "" {
		countQuery += fmt.Sprintf(" AND target ILIKE $%d", countArgIndex)
		countArgs = append(countArgs, "%"+query.Target+"%")
		countArgIndex++
	}

	if query.TargetType != "" {
		countQuery += fmt.Sprintf(" AND target_type = $%d", countArgIndex)
		countArgs = append(countArgs, query.TargetType)
		countArgIndex++
	}

	if query.CreatedBy != "" {
		countQuery += fmt.Sprintf(" AND created_by = $%d", countArgIndex)
		countArgs = append(countArgs, query.CreatedBy)
		countArgIndex++
	}

	if query.AssignedTo != "" {
		countQuery += fmt.Sprintf(" AND assigned_to = $%d", countArgIndex)
		countArgs = append(countArgs, query.AssignedTo)
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

	return alerts, total, nil
}

// UpdateAlert 更新预警
func (s *Service) UpdateAlert(ctx context.Context, alertID int, updates UpdateAlertRequest) error {
	// 构建更新语句
	query := "UPDATE alerts SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{}
	argIndex := 1

	if updates.Title != nil && *updates.Title != "" {
		query += fmt.Sprintf(", title = $%d", argIndex)
		args = append(args, *updates.Title)
		argIndex++
	}

	if updates.Description != nil {
		query += fmt.Sprintf(", description = $%d", argIndex)
		args = append(args, *updates.Description)
		argIndex++
	}

	if updates.Status != nil && *updates.Status != "" {
		query += fmt.Sprintf(", status = $%d", argIndex)
		args = append(args, *updates.Status)
		argIndex++

		// 如果是解决状态，设置解决时间
		if *updates.Status == "resolved" {
			query += fmt.Sprintf(", resolved_at = CURRENT_TIMESTAMP")
		} else if *updates.Status == "closed" {
			query += fmt.Sprintf(", closed_at = CURRENT_TIMESTAMP")
		}
	}

	if updates.AssignedTo != nil {
		query += fmt.Sprintf(", assigned_to = $%d", argIndex)
		args = append(args, *updates.AssignedTo)
		argIndex++
	}

	if updates.Details != nil {
		detailsJSON, err := json.Marshal(updates.Details)
		if err != nil {
			return fmt.Errorf("序列化详细信息失败: %v", err)
		}
		query += fmt.Sprintf(", details = $%d", argIndex)
		args = append(args, string(detailsJSON))
		argIndex++
	}

	if updates.Metadata != nil {
		metadataJSON, err := json.Marshal(updates.Metadata)
		if err != nil {
			return fmt.Errorf("序列化元数据失败: %v", err)
		}
		query += fmt.Sprintf(", metadata = $%d", argIndex)
		args = append(args, string(metadataJSON))
		argIndex++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, alertID)

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("更新预警失败: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("预警不存在: %d", alertID)
	}

	log.Infof("预警更新成功: %d", alertID)
	return nil
}

// DeleteAlert 删除预警
func (s *Service) DeleteAlert(ctx context.Context, alertID int) error {
	// 先删除相关通知
	_, err := s.db.ExecContext(ctx, "DELETE FROM notifications WHERE alert_id = $1", alertID)
	if err != nil {
		return fmt.Errorf("删除通知记录失败: %v", err)
	}

	// 再删除预警
	result, err := s.db.ExecContext(ctx, "DELETE FROM alerts WHERE id = $1", alertID)
	if err != nil {
		return fmt.Errorf("删除预警失败: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("预警不存在: %d", alertID)
	}

	log.Infof("预警删除成功: %d", alertID)
	return nil
}

// GetAlertStats 获取预警统计
func (s *Service) GetAlertStats(ctx context.Context) (*AlertStats, error) {
	stats := &AlertStats{
		BySeverity: make(map[string]int),
		ByType:     make(map[string]int),
		ByStatus:   make(map[string]int),
		BySource:   make(map[string]int),
	}

	// 获取总数
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM alerts").Scan(&stats.Total)
	if err != nil {
		return nil, fmt.Errorf("获取预警总数失败: %v", err)
	}

	// 按严重程度统计
	rows, err := s.db.QueryContext(ctx, "SELECT severity, COUNT(*) FROM alerts GROUP BY severity")
	if err != nil {
		return nil, fmt.Errorf("按严重程度统计失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var severity string
		var count int
		if err := rows.Scan(&severity, &count); err == nil {
			stats.BySeverity[severity] = count
		}
	}

	// 按类型统计
	rows, err = s.db.QueryContext(ctx, "SELECT type, COUNT(*) FROM alerts GROUP BY type")
	if err != nil {
		return nil, fmt.Errorf("按类型统计失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var alertType string
		var count int
		if err := rows.Scan(&alertType, &count); err == nil {
			stats.ByType[alertType] = count
		}
	}

	// 按状态统计
	rows, err = s.db.QueryContext(ctx, "SELECT status, COUNT(*) FROM alerts GROUP BY status")
	if err != nil {
		return nil, fmt.Errorf("按状态统计失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err == nil {
			stats.ByStatus[status] = count
		}
	}

	// 按来源统计
	rows, err = s.db.QueryContext(ctx, "SELECT source, COUNT(*) FROM alerts GROUP BY source")
	if err != nil {
		return nil, fmt.Errorf("按来源统计失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var source string
		var count int
		if err := rows.Scan(&source, &count); err == nil {
			stats.BySource[source] = count
		}
	}

	// 获取最近预警
	recentAlerts, _, err := s.GetAlerts(ctx, AlertQuery{
		Page:     1,
		PageSize: 10,
		SortBy:   "created_at",
		SortOrder: "desc",
	})
	if err != nil {
		return nil, fmt.Errorf("获取最近预警失败: %v", err)
	}
	stats.RecentAlerts = recentAlerts

	return stats, nil
}

// CreateRule 创建预警规则
func (s *Service) CreateRule(ctx context.Context, req CreateRuleRequest) (int, error) {
	log.Infof("创建预警规则: %s", req.Name)

	// 验证规则参数
	if err := s.validateRuleRequest(req); err != nil {
		return 0, err
	}

	// 序列化JSON字段
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return 0, fmt.Errorf("序列化配置失败: %v", err)
	}

	tagsJSON, err := json.Marshal(req.Tags)
	if err != nil {
		return 0, fmt.Errorf("序列化标签失败: %v", err)
	}

	// 插入规则
	var ruleID int
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO alert_rules (
			name, description, type, severity, condition, action,
			config, enabled, priority, tags
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`,
		req.Name,
		req.Description,
		req.Type,
		req.Severity,
		req.Condition,
		req.Action,
		string(configJSON),
		req.Enabled,
		req.Priority,
		string(tagsJSON),
	).Scan(&ruleID)

	if err != nil {
		return 0, fmt.Errorf("创建预警规则失败: %v", err)
	}

	log.Infof("预警规则创建成功: %s (ID: %d)", req.Name, ruleID)
	return ruleID, nil
}

// GetRules 获取预警规则列表
func (s *Service) GetRules(ctx context.Context, query RuleQuery) ([]AlertRule, int, error) {
	sqlQuery := `
		SELECT id, name, description, type, severity, condition, action,
		       config, enabled, priority, tags, created_at, updated_at
		FROM alert_rules
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
		return nil, 0, fmt.Errorf("查询预警规则失败: %v", err)
	}
	defer rows.Close()

	var rules []AlertRule
	for rows.Next() {
		var rule AlertRule
		var configJSON, tagsJSON []byte
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.Description,
			&rule.Type,
			&rule.Severity,
			&rule.Condition,
			&rule.Action,
			&configJSON,
			&rule.Enabled,
			&rule.Priority,
			&tagsJSON,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			log.Errorf("扫描预警规则行失败: %v", err)
			continue
		}

		// 解析JSON字段
		if len(configJSON) > 0 {
			json.Unmarshal(configJSON, &rule.Config)
		}
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &rule.Tags)
		}

		rule.CreatedAt = createdAt
		rule.UpdatedAt = updatedAt

		rules = append(rules, rule)
	}

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM alert_rules WHERE 1=1"
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

// UpdateRule 更新预警规则
func (s *Service) UpdateRule(ctx context.Context, ruleID int, updates map[string]interface{}) error {
	// 构建更新语句
	query := "UPDATE alert_rules SET updated_at = CURRENT_TIMESTAMP"
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

	if severity, ok := updates["severity"].(string); ok && severity != "" {
		query += fmt.Sprintf(", severity = $%d", argIndex)
		args = append(args, severity)
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

	if config, ok := updates["config"].(map[string]interface{}); ok {
		configJSON, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("序列化配置失败: %v", err)
		}
		query += fmt.Sprintf(", config = $%d", argIndex)
		args = append(args, string(configJSON))
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

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, ruleID)

	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("更新预警规则失败: %v", err)
	}

	log.Infof("预警规则更新成功: %d", ruleID)
	return nil
}

// DeleteRule 删除预警规则
func (s *Service) DeleteRule(ctx context.Context, ruleID int) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM alert_rules WHERE id = $1", ruleID)
	if err != nil {
		return fmt.Errorf("删除预警规则失败: %v", err)
	}

	log.Infof("预警规则删除成功: %d", ruleID)
	return nil
}

// triggerNotifications 触发通知
func (s *Service) triggerNotifications(ctx context.Context, alertID int) {
	// 获取预警详情
	alert, err := s.GetAlert(ctx, alertID)
	if err != nil {
		log.Errorf("获取预警详情失败: %v", err)
		return
	}

	// 获取启用的规则
	rules, _, err := s.GetRules(ctx, RuleQuery{
		Enabled: &[]bool{true}[0],
		Page:    1,
		PageSize: 100,
	})
	if err != nil {
		log.Errorf("获取预警规则失败: %v", err)
		return
	}

	// 检查每个规则是否匹配
	for _, rule := range rules {
		if s.matchRule(alert, &rule) {
			s.sendNotification(ctx, alert, &rule)
		}
	}
}

// matchRule 检查规则是否匹配
func (s *Service) matchRule(alert *Alert, rule *AlertRule) bool {
	// 这里实现规则匹配逻辑
	// 可以根据规则的条件表达式进行匹配
	// 目前先简单实现：如果预警类型和严重程度匹配规则，则触发
	if alert.Type == rule.Type && alert.Severity == rule.Severity {
		return true
	}
	return false
}

// sendNotification 发送通知
func (s *Service) sendNotification(ctx context.Context, alert *Alert, rule *AlertRule) {
	// 创建通知记录
	content := map[string]interface{}{
		"alert_id":   alert.ID,
		"alert_title": alert.Title,
		"alert_type": alert.Type,
		"alert_severity": alert.Severity,
		"rule_id":    rule.ID,
		"rule_name":  rule.Name,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	contentJSON, err := json.Marshal(content)
	if err != nil {
		log.Errorf("序列化通知内容失败: %v", err)
		return
	}

	// 根据规则动作发送通知
	switch rule.Action {
	case "email":
		s.sendEmailNotification(ctx, alert, rule, content)
	case "webhook":
		s.sendWebhookNotification(ctx, alert, rule, content)
	case "sms":
		s.sendSMSNotification(ctx, alert, rule, content)
	case "notification":
		s.sendSystemNotification(ctx, alert, rule, content)
	default:
		log.Warnf("未知的通知动作: %s", rule.Action)
	}

	// 记录通知
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO notifications (
			alert_id, type, status, recipient, content
		) VALUES ($1, $2, $3, $4, $5)
	`,
		alert.ID,
		rule.Action,
		"sent",
		"system",
		string(contentJSON),
	)
	if err != nil {
		log.Errorf("记录通知失败: %v", err)
	}
}

// sendEmailNotification 发送邮件通知
func (s *Service) sendEmailNotification(ctx context.Context, alert *Alert, rule *AlertRule, content map[string]interface{}) {
	// TODO: 实现邮件发送逻辑
	log.Infof("发送邮件通知: 预警ID=%d, 规则ID=%d", alert.ID, rule.ID)
}

// sendWebhookNotification 发送Webhook通知
func (s *Service) sendWebhookNotification(ctx context.Context, alert *Alert, rule *AlertRule, content map[string]interface{}) {
	// TODO: 实现Webhook发送逻辑
	log.Infof("发送Webhook通知: 预警ID=%d, 规则ID=%d", alert.ID, rule.ID)
}

// sendSMSNotification 发送短信通知
func (s *Service) sendSMSNotification(ctx context.Context, alert *Alert, rule *AlertRule, content map[string]interface{}) {
	// TODO: 实现短信发送逻辑
	log.Infof("发送短信通知: 预警ID=%d, 规则ID=%d", alert.ID, rule.ID)
}

// sendSystemNotification 发送系统通知
func (s *Service) sendSystemNotification(ctx context.Context, alert *Alert, rule *AlertRule, content map[string]interface{}) {
	// TODO: 实现系统通知逻辑
	log.Infof("发送系统通知: 预警ID=%d, 规则ID=%d", alert.ID, rule.ID)
}

// validateAlertRequest 验证预警请求
func (s *Service) validateAlertRequest(req CreateAlertRequest) error {
	if req.Title == "" {
		return fmt.Errorf("预警标题不能为空")
	}
	if req.Type == "" {
		return fmt.Errorf("预警类型不能为空")
	}
	if req.Severity == "" {
		return fmt.Errorf("预警严重程度不能为空")
	}
	if req.Source == "" {
		return fmt.Errorf("预警来源不能为空")
	}
	if req.CreatedBy == "" {
		return fmt.Errorf("创建者不能为空")
	}
	return nil
}

// validateRuleRequest 验证规则请求
func (s *Service) validateRuleRequest(req CreateRuleRequest) error {
	if req.Name == "" {
		return fmt.Errorf("规则名称不能为空")
	}
	if req.Type == "" {
		return fmt.Errorf("规则类型不能为空")
	}
	if req.Severity == "" {
		return fmt.Errorf("规则严重程度不能为空")
	}
	if req.Condition == "" {
		return fmt.Errorf("规则条件不能为空")
	}
	if req.Action == "" {
		return fmt.Errorf("规则动作不能为空")
	}
	return nil
}