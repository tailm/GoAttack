package intelligence

import (
	"GoAttack/common/log"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Service 漏洞情报服务
type Service struct {
	db        *sql.DB
	collector *Collector
	scheduler *Scheduler
}

// NewService 创建新的服务
func NewService(db *sql.DB) *Service {
	collector := NewCollector(db)
	scheduler := NewScheduler(db)
	return &Service{
		db:        db,
		collector: collector,
		scheduler: scheduler,
	}
}

// StartScheduler 启动调度器
func (s *Service) StartScheduler(ctx context.Context) error {
	return s.scheduler.Start(ctx)
}

// StopScheduler 停止调度器
func (s *Service) StopScheduler() {
	s.scheduler.Stop()
}

// GetSources 获取情报源列表
func (s *Service) GetSources(ctx context.Context, page, pageSize int, search string) ([]SourceConfig, int, error) {
	query := `
		SELECT id, name, url, type, enabled, sync_interval, config, 
		       last_sync, last_sync_status, last_sync_message,
		       created_at, updated_at
		FROM intelligence_source
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if search != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR url ILIKE $%d)", argIndex, argIndex)
		args = append(args, "%"+search+"%")
		argIndex++
	}

	query += " ORDER BY id"

	// 分页
	if pageSize > 0 {
		offset := (page - 1) * pageSize
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, pageSize, offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询情报源失败: %v", err)
	}
	defer rows.Close()

	var sources []SourceConfig
	for rows.Next() {
		var source SourceConfig
		var configJSON []byte
		var lastSync *time.Time
		var lastSyncStatus, lastSyncMessage *string
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&source.ID,
			&source.Name,
			&source.URL,
			&source.Type,
			&source.Enabled,
			&source.SyncInterval,
			&configJSON,
			&lastSync,
			&lastSyncStatus,
			&lastSyncMessage,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			log.Errorf("扫描情报源行失败: %v", err)
			continue
		}

		// 解析配置
		if len(configJSON) > 0 {
			var config map[string]interface{}
			if err := json.Unmarshal(configJSON, &config); err != nil {
				log.Errorf("解析情报源配置失败: %v", err)
			} else {
				source.Config = config
			}
		}

		source.LastSync = lastSync
		if lastSyncStatus != nil {
			source.LastSyncStatus = *lastSyncStatus
		}

		sources = append(sources, source)
	}

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM intelligence_source"
	if search != "" {
		countQuery += " WHERE name ILIKE $1 OR url ILIKE $1"
	}

	var total int
	if search != "" {
		err = s.db.QueryRowContext(ctx, countQuery, "%"+search+"%").Scan(&total)
	} else {
		err = s.db.QueryRowContext(ctx, countQuery).Scan(&total)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %v", err)
	}

	return sources, total, nil
}

// GetSourceByID 根据ID获取情报源
func (s *Service) GetSourceByID(ctx context.Context, id int) (*SourceConfig, error) {
	var source SourceConfig
	var configJSON []byte
	var lastSync *time.Time
	var lastSyncStatus, lastSyncMessage *string
	var createdAt, updatedAt time.Time

	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, url, type, enabled, sync_interval, config,
		       last_sync, last_sync_status, last_sync_message,
		       created_at, updated_at
		FROM intelligence_source
		WHERE id = $1
	`, id).Scan(
		&source.ID,
		&source.Name,
		&source.URL,
		&source.Type,
		&source.Enabled,
		&source.SyncInterval,
		&configJSON,
		&lastSync,
		&lastSyncStatus,
		&lastSyncMessage,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("查询情报源失败: %v", err)
	}

	// 解析配置
	if len(configJSON) > 0 {
		var config map[string]interface{}
		if err := json.Unmarshal(configJSON, &config); err != nil {
			log.Errorf("解析情报源配置失败: %v", err)
		} else {
			source.Config = config
		}
	}

	source.LastSync = lastSync
	if lastSyncStatus != nil {
		source.LastSyncStatus = *lastSyncStatus
	}

	return &source, nil
}

// CreateSource 创建情报源
func (s *Service) CreateSource(ctx context.Context, source *SourceConfig) error {
	configJSON, err := json.Marshal(source.Config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO intelligence_source (
			name, url, type, enabled, sync_interval, config
		) VALUES ($1, $2, $3, $4, $5, $6)
	`,
		source.Name,
		source.URL,
		source.Type,
		source.Enabled,
		source.SyncInterval,
		configJSON,
	)
	if err != nil {
		return fmt.Errorf("创建情报源失败: %v", err)
	}

	// 如果启用，添加到调度器
	if source.Enabled {
		s.scheduler.addJob(*source)
	}

	return nil
}

// UpdateSource 更新情报源
func (s *Service) UpdateSource(ctx context.Context, id int, updates map[string]interface{}) error {
	// 获取当前配置
	currentSource, err := s.GetSourceByID(ctx, id)
	if err != nil {
		return err
	}

	// 构建更新语句
	query := "UPDATE intelligence_source SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{}
	argIndex := 1

	if name, ok := updates["name"].(string); ok && name != "" {
		query += fmt.Sprintf(", name = $%d", argIndex)
		args = append(args, name)
		argIndex++
	}

	if url, ok := updates["url"].(string); ok && url != "" {
		query += fmt.Sprintf(", url = $%d", argIndex)
		args = append(args, url)
		argIndex++
	}

	if enabled, ok := updates["enabled"].(bool); ok {
		query += fmt.Sprintf(", enabled = $%d", argIndex)
		args = append(args, enabled)
		argIndex++
	}

	if syncInterval, ok := updates["sync_interval"].(int); ok && syncInterval > 0 {
		query += fmt.Sprintf(", sync_interval = $%d", argIndex)
		args = append(args, syncInterval)
		argIndex++
	}

	if config, ok := updates["config"].(map[string]interface{}); ok {
		configJSON, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("序列化配置失败: %v", err)
		}
		query += fmt.Sprintf(", config = $%d", argIndex)
		args = append(args, configJSON)
		argIndex++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, id)

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("更新情报源失败: %v", err)
	}

	// 更新调度器中的任务
	if enabled, ok := updates["enabled"].(bool); ok {
		if enabled && !currentSource.Enabled {
			// 从禁用变为启用，添加任务
			updatedSource, err := s.GetSourceByID(ctx, id)
			if err == nil {
				s.scheduler.addJob(*updatedSource)
			}
		} else if !enabled && currentSource.Enabled {
			// 从启用变为禁用，移除任务
			s.scheduler.removeJob(id)
		}
	} else if _, ok := updates["sync_interval"].(int); ok {
		// 同步间隔变化，更新任务
		s.scheduler.removeJob(id)
		if currentSource.Enabled {
			updatedSource, err := s.GetSourceByID(ctx, id)
			if err == nil {
				s.scheduler.addJob(*updatedSource)
			}
		}
	}

	return nil
}

// DeleteSource 删除情报源
func (s *Service) DeleteSource(ctx context.Context, id int) error {
	// 先从调度器中移除任务
	s.scheduler.removeJob(id)

	// 从数据库中删除
	_, err := s.db.ExecContext(ctx, "DELETE FROM intelligence_source WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("删除情报源失败: %v", err)
	}

	return nil
}

// SyncSource 同步指定情报源
func (s *Service) SyncSource(ctx context.Context, id int) error {
	return s.scheduler.TriggerManualSync(ctx, id)
}

// GetJobStatus 获取任务状态
func (s *Service) GetJobStatus() map[int]JobStatus {
	return s.scheduler.GetJobStatus()
}

// GetVulnerabilities 获取漏洞情报列表
func (s *Service) GetVulnerabilities(ctx context.Context, query VulnerabilityQuery) ([]VulnerabilityData, int, error) {
	sqlQuery := `
		SELECT id, source, source_id, title, description, severity, cve_id,
		       cvss_score, cvss_vector, affected_products, references,
		       published_date, last_modified_date, raw_data,
		       created_at, updated_at
		FROM vulnerability_intelligence
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if query.Search != "" {
		sqlQuery += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d OR cve_id ILIKE $%d)", 
			argIndex, argIndex, argIndex)
		args = append(args, "%"+query.Search+"%")
		argIndex++
	}

	if query.Source != "" {
		sqlQuery += fmt.Sprintf(" AND source = $%d", argIndex)
		args = append(args, query.Source)
		argIndex++
	}

	if query.Severity != "" {
		sqlQuery += fmt.Sprintf(" AND severity = $%d", argIndex)
		args = append(args, query.Severity)
		argIndex++
	}

	if query.CVEID != "" {
		sqlQuery += fmt.Sprintf(" AND cve_id ILIKE $%d", argIndex)
		args = append(args, "%"+query.CVEID+"%")
		argIndex++
	}

	if query.StartDate != nil {
		sqlQuery += fmt.Sprintf(" AND published_date >= $%d", argIndex)
		args = append(args, query.StartDate)
		argIndex++
	}

	if query.EndDate != nil {
		sqlQuery += fmt.Sprintf(" AND published_date <= $%d", argIndex)
		args = append(args, query.EndDate)
		argIndex++
	}

	// 排序
	switch query.SortBy {
	case "published_date":
		sqlQuery += " ORDER BY published_date"
	case "cvss_score":
		sqlQuery += " ORDER BY cvss_score"
	case "severity":
		sqlQuery += " ORDER BY CASE severity WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 WHEN 'low' THEN 4 ELSE 5 END"
	default:
		sqlQuery += " ORDER BY published_date DESC"
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
		return nil, 0, fmt.Errorf("查询漏洞情报失败: %v", err)
	}
	defer rows.Close()

	var vulnerabilities []VulnerabilityData
	for rows.Next() {
		var vuln VulnerabilityData
		var affectedProductsJSON, referencesJSON, rawDataJSON []byte
		var publishedDate, lastModifiedDate *time.Time
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&vuln.ID,
			&vuln.Source,
			&vuln.SourceID,
			&vuln.Title,
			&vuln.Description,
			&vuln.Severity,
			&vuln.CVEID,
			&vuln.CVSSScore,
			&vuln.CVSSVector,
			&affectedProductsJSON,
			&referencesJSON,
			&publishedDate,
			&lastModifiedDate,
			&rawDataJSON,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			log.Errorf("扫描漏洞情报行失败: %v", err)
			continue
		}

		// 解析JSON字段
		if len(affectedProductsJSON) > 0 {
			json.Unmarshal(affectedProductsJSON, &vuln.AffectedProducts)
		}
		if len(referencesJSON) > 0 {
			json.Unmarshal(referencesJSON, &vuln.References)
		}
		if len(rawDataJSON) > 0 {
			json.Unmarshal(rawDataJSON, &vuln.RawData)
		}

		vuln.PublishedDate = publishedDate
		vuln.LastModifiedDate = lastModifiedDate

		vulnerabilities = append(vulnerabilities, vuln)
	}

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM vulnerability_intelligence WHERE 1=1"
	countArgs := []interface{}{}
	countArgIndex := 1

	if query.Search != "" {
		countQuery += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d OR cve_id ILIKE $%d)", 
			countArgIndex, countArgIndex, countArgIndex)
		countArgs = append(countArgs, "%"+query.Search+"%")
		countArgIndex++
	}

	if query.Source != "" {
		countQuery += fmt.Sprintf(" AND source = $%d", countArgIndex)
		countArgs = append(countArgs, query.Source)
		countArgIndex++
	}

	if query.Severity != "" {
		countQuery += fmt.Sprintf(" AND severity = $%d", countArgIndex)
		countArgs = append(countArgs, query.Severity)
		countArgIndex++
	}

	if query.CVEID != "" {
		countQuery += fmt.Sprintf(" AND cve_id ILIKE $%d", countArgIndex)
		countArgs = append(countArgs, "%"+query.CVEID+"%")
		countArgIndex++
	}

	if query.StartDate != nil {
		countQuery += fmt.Sprintf(" AND published_date >= $%d", countArgIndex)
		countArgs = append(countArgs, query.StartDate)
		countArgIndex++
	}

	if query.EndDate != nil {
		countQuery += fmt.Sprintf(" AND published_date <= $%d", countArgIndex)
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

	return vulnerabilities, total, nil
}

// VulnerabilityQuery 漏洞查询参数
type VulnerabilityQuery struct {
	Search    string
	Source    string
	Severity  string
	CVEID     string
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	PageSize  int
	SortBy    string
	SortOrder string
}

// GetVulnerabilityStats 获取漏洞统计信息
func (s *Service) GetVulnerabilityStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总漏洞数
	var totalCount int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vulnerability_intelligence").Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("查询总漏洞数失败: %v", err)
	}
	stats["total"] = totalCount

	// 按严重程度统计
	severityStats := make(map[string]int)
	rows, err := s.db.QueryContext(ctx, `
		SELECT severity, COUNT(*) 
		FROM vulnerability_intelligence 
		GROUP BY severity
	`)
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

	// 按来源统计
	sourceStats := make(map[string]int)
	rows, err = s.db.QueryContext(ctx, `
		SELECT source, COUNT(*) 
		FROM vulnerability_intelligence 
		GROUP BY source
	`)
	if err != nil {
		return nil, fmt.Errorf("查询来源统计失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var source string
		var count int
		if err := rows.Scan(&source, &count); err == nil {
			sourceStats[source] = count
		}
	}
	stats["by_source"] = sourceStats

	// 最近7天新增漏洞数
	recentStats := make([]map[string]interface{}, 0)
	rows, err = s.db.QueryContext(ctx, `
		SELECT DATE(published_date) as date, COUNT(*) as count
		FROM vulnerability_intelligence
		WHERE published_date >= CURRENT_DATE - INTERVAL '7 days'
		GROUP BY DATE(published_date)
		ORDER BY date DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("查询最近统计失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var date time.Time
		var count int
		if err := rows.Scan(&date, &count); err == nil {
			recentStats = append(recentStats, map[string]interface{}{
				"date":  date.Format("2006-01-02"),
				"count": count,
			})
		}
	}
	stats["recent"] = recentStats

	return stats, nil
}