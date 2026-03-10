package postgres

import (
	"database/sql"
	"fmt"
)

// ============================================
// 任务管理模块
// 说明: 处理扫描任务的增删改查、状态更新等操作
// ============================================

// CreateTask 创建扫描任务
func CreateTask(name, target, taskType, creator, description, options string) (int64, error) {
	var id int64
	err := DB.QueryRow(
		"INSERT INTO task (name, target, type, creator, description, options, created_at) VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP) RETURNING id",
		name, target, taskType, creator, description, options,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetTaskByID 根据ID查询任务
func GetTaskByID(taskID int) (*sql.Row, error) {
	query := `
	SELECT id, name, target, type, status, progress, creator, description, 
	       options, created_at, updated_at, started_at, completed_at
	FROM task WHERE id = $1`
	return DB.QueryRow(query, taskID), nil
}

// GetTaskList 获取任务列表（支持分页和筛选）
func GetTaskList(page, pageSize int, creator, status, name, taskType string) (*sql.Rows, int, error) {
	// 构建查询条件
	whereClause := "WHERE 1=1"
	args := make([]interface{}, 0)
	argIndex := 1

	if creator != "" {
		whereClause += fmt.Sprintf(" AND creator = $%d", argIndex)
		args = append(args, creator)
		argIndex++
	}
	if status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}
	if name != "" {
		whereClause += fmt.Sprintf(" AND name LIKE $%d", argIndex)
		args = append(args, "%"+name+"%")
		argIndex++
	}
	if taskType != "" {
		whereClause += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, taskType)
		argIndex++
	}

	// 查询总数
	var total int
	countQuery := "SELECT COUNT(*) FROM task " + whereClause
	err := DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	offset := (page - 1) * pageSize
	query := `
	SELECT id, name, target, type, status, progress, creator, description, 
	       options, created_at, updated_at, started_at, completed_at
	FROM task ` + whereClause + ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)
	args = append(args, pageSize, offset)

	rows, err := DB.Query(query, args...)
	return rows, total, err
}

// GetPendingScheduledTasks 获取所有处于 pending 状态的任务
func GetPendingScheduledTasks() (*sql.Rows, error) {
	query := `
	SELECT id, name, target, type, status, progress, creator, description, 
	       options, created_at, updated_at, started_at, completed_at
	FROM task WHERE status = 'pending'`
	return DB.Query(query)
}

// UpdateTaskStatus 更新任务状态
func UpdateTaskStatus(taskID int, status string, progress int) error {
	_, err := DB.Exec(
		"UPDATE task SET status = $1, progress = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3",
		status, progress, taskID,
	)
	return err
}

// UpdateTaskProgress 更新任务进度和状态（不再更新 result 字段）
// 扫描结果现在存储在 asset_scan_result 表中
func UpdateTaskProgress(taskID int, status string, progress int) error {
	query := `
	UPDATE task 
	SET status = $1, progress = $2, updated_at = CURRENT_TIMESTAMP,
	    created_at = CASE WHEN $3 = 'running' AND $4 = 0 THEN CURRENT_TIMESTAMP ELSE created_at END,
	    started_at = CASE WHEN $5 = 'running' AND $6 = 0 THEN CURRENT_TIMESTAMP ELSE started_at END,
	    completed_at = CASE WHEN $7 = 'running' AND $8 = 0 THEN NULL WHEN $9 IN ('completed', 'failed', 'stopped') THEN CURRENT_TIMESTAMP ELSE completed_at END
	WHERE id = $10`

	_, err := DB.Exec(query, status, progress, status, progress, status, progress, status, progress, status, taskID)
	return err
}

// UpdateTaskResult 保持向后兼容（废弃，请使用 UpdateTaskProgress）
func UpdateTaskResult(taskID int, status string, progress int, _ string) error {
	return UpdateTaskProgress(taskID, status, progress)
}

// DeleteTask 删除任务（级联删除相关漏洞）
func DeleteTask(taskID int) error {
	_, err := DB.Exec("DELETE FROM task WHERE id = $1", taskID)
	return err
}

// GetTaskStats 获取任务统计信息
func GetTaskStats(creator string) (map[string]int, error) {
	stats := make(map[string]int)

	query := `
	SELECT 
		COUNT(*) as total,
		SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
		SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) as running,
		SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed,
		SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed
	FROM task WHERE creator = $1`

	var total, pending, running, completed, failed int
	err := DB.QueryRow(query, creator).Scan(&total, &pending, &running, &completed, &failed)
	if err != nil {
		return stats, err
	}

	stats["total"] = total
	stats["pending"] = pending
	stats["running"] = running
	stats["completed"] = completed
	stats["failed"] = failed

	return stats, nil
}

// ============================================
// 任务筛选辅助函数
// ============================================

// GetTasksByCreatorWithFilter 根据创建者获取任务列表（支持筛选和分页）
func GetTasksByCreatorWithFilter(creator string, limit, offset int, name, status, taskType string) (*sql.Rows, error) {
	query := `SELECT id, name, target, type, status, progress, creator, description, 
			  options, created_at, updated_at, started_at, completed_at 
			  FROM task WHERE creator = $1`

	args := []interface{}{creator}
	argIndex := 2

	// 添加名称筛选
	if name != "" {
		query += fmt.Sprintf(" AND name LIKE $%d", argIndex)
		args = append(args, "%"+name+"%")
		argIndex++
	}

	// 添加状态筛选
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	// 添加类型筛选
	if taskType != "" {
		query += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, taskType)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	return DB.Query(query, args...)
}

// CountTasksByCreatorWithFilter 统计用户的任务数量（支持筛选）
func CountTasksByCreatorWithFilter(creator, name, status, taskType string) (int, error) {
	query := "SELECT COUNT(*) FROM task WHERE creator = $1"
	args := []interface{}{creator}
	argIndex := 2

	// 添加名称筛选
	if name != "" {
		query += fmt.Sprintf(" AND name LIKE $%d", argIndex)
		args = append(args, "%"+name+"%")
		argIndex++
	}

	// 添加状态筛选
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	// 添加类型筛选
	if taskType != "" {
		query += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, taskType)
		argIndex++
	}

	var count int
	err := DB.QueryRow(query, args...).Scan(&count)
	return count, err
}

// ClearTaskResults 清除任务的所有扫描结果（用于重新扫描）
func ClearTaskResults(taskID int) error {
	// 开启事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败: %v", err)
	}
	defer tx.Rollback()

	// 1. 删除 asset_scan_result 表中该任务的所有记录
	_, err = tx.Exec("DELETE FROM asset_scan_result WHERE task_id = $1", taskID)
	if err != nil {
		return fmt.Errorf("删除扫描结果失败: %v", err)
	}

	// 2. 删除 asset_port 表中该任务的所有记录
	_, err = tx.Exec("DELETE FROM asset_port WHERE task_id = $1", taskID)
	if err != nil {
		return fmt.Errorf("删除端口信息失败: %v", err)
	}

	// 3. 删除 web_fingerprint 表中该任务的所有记录
	// 注意：web_fingerprint表通过asset_port_id关联到asset_port，再通过asset关联到task
	// 需要先获取该任务的所有资产ID
	var assetIDs []int
	rows, err := tx.Query("SELECT id FROM asset WHERE task_id = $1", taskID)
	if err != nil {
		return fmt.Errorf("查询任务资产失败: %v", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var assetID int
		if err := rows.Scan(&assetID); err != nil {
			continue
		}
		assetIDs = append(assetIDs, assetID)
	}
	
	// 删除关联的web_fingerprint记录
	if len(assetIDs) > 0 {
		// 构建IN查询
		query := "DELETE FROM web_fingerprint WHERE asset_port_id IN (SELECT id FROM asset_port WHERE asset_id = ANY($1))"
		_, err = tx.Exec(query, assetIDs)
		if err != nil {
			return fmt.Errorf("删除Web指纹失败: %v", err)
		}
	}

	// 4. 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}
