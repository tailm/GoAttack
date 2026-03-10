package postgres

import (
	"database/sql"
	"fmt"
	"time"
)

// ============================================
// 资产管理模块
// 说明: 处理资产的增删改查操作
// ============================================

// CreateOrUpdateAsset 创建或更新资产
func CreateOrUpdateAsset(value, assetType string, isAlive bool) (int64, error) {
	now := time.Now()

	// 先尝试查找已存在的资产
	var id int64
	err := DB.QueryRow("SELECT id FROM asset WHERE value = $1", value).Scan(&id)

	if err == sql.ErrNoRows {
		// 资产不存在，创建新资产
		err := DB.QueryRow(
			`INSERT INTO asset (value, asset_type, is_alive, first_seen, last_seen) 
			 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			value, assetType, isAlive, now, now,
		).Scan(&id)
		if err != nil {
			return 0, err
		}
		return id, nil
	} else if err != nil {
		return 0, err
	}

	// 资产已存在，更新信息
	_, err = DB.Exec(
		`UPDATE asset SET is_alive = $1, last_seen = $2, asset_type = $3 WHERE id = $4`,
		isAlive, now, assetType, id,
	)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetOrCreateAsset 获取或创建资产（CreateOrUpdateAsset的别名）
// 如果资产不存在则创建，存在则返回ID
func GetOrCreateAsset(value, assetType string) (int64, error) {
	return CreateOrUpdateAsset(value, assetType, false)
}

// GetAssetByValue 根据值获取资产
func GetAssetByValue(value string) (*sql.Row, error) {
	query := "SELECT id, value, asset_type, is_alive, first_seen, last_seen FROM asset WHERE value = $1"
	return DB.QueryRow(query, value), nil
}

// GetAssetByID 根据ID获取资产
func GetAssetByID(id int64) (*sql.Row, error) {
	query := "SELECT id, value, asset_type, is_alive, first_seen, last_seen FROM asset WHERE id = $1"
	return DB.QueryRow(query, id), nil
}

// GetAllAssets 获取所有资产列表（支持分页）
func GetAllAssets(page, pageSize int, assetType string, isAlive *bool) (*sql.Rows, int, error) {
	whereClause := "WHERE 1=1"
	args := make([]interface{}, 0)
	argIndex := 1

	if assetType != "" {
		whereClause += fmt.Sprintf(" AND asset_type = $%d", argIndex)
		args = append(args, assetType)
		argIndex++
	}

	if isAlive != nil {
		whereClause += fmt.Sprintf(" AND is_alive = $%d", argIndex)
		args = append(args, *isAlive)
		argIndex++
	}

	// 查询总数
	var total int
	countQuery := "SELECT COUNT(*) FROM asset " + whereClause
	err := DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	offset := (page - 1) * pageSize
	query := `SELECT id, value, asset_type, is_alive, first_seen, last_seen 
			  FROM asset ` + whereClause + ` ORDER BY last_seen DESC LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)
	args = append(args, pageSize, offset)

	rows, err := DB.Query(query, args...)
	return rows, total, err
}

// DeleteAsset 删除资产
func DeleteAsset(id int64) error {
	_, err := DB.Exec("DELETE FROM asset WHERE id = $1", id)
	return err
}

// ============================================
// 资产扫描结果管理
// ============================================

// SaveAssetScanResult 保存资产扫描结果
func SaveAssetScanResult(taskID int, assetID int64, scanType, status, result string) error {
	_, err := DB.Exec(
		`INSERT INTO asset_scan_result (task_id, asset_id, scan_type, status, result, scanned_at) 
		 VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)`,
		taskID, assetID, scanType, status, result,
	)
	return err
}

// GetAssetScanResultsByTaskID 根据任务ID获取扫描结果
func GetAssetScanResultsByTaskID(taskID int) (*sql.Rows, error) {
	query := `
	SELECT asr.id, asr.task_id, asr.asset_id, asr.scan_type, asr.status, asr.result, asr.scanned_at,
	       a.value, a.asset_type, a.is_alive
	FROM asset_scan_result asr
	LEFT JOIN asset a ON asr.asset_id = a.id
	WHERE asr.task_id = $1
	ORDER BY asr.scanned_at DESC`

	return DB.Query(query, taskID)
}

// GetAssetScanResultByID 根据ID获取单个扫描结果
func GetAssetScanResultByID(id int64) (*sql.Row, error) {
	query := `
	SELECT asr.id, asr.task_id, asr.asset_id, asr.scan_type, asr.status, asr.result, asr.scanned_at,
	       a.value, a.asset_type, a.is_alive
	FROM asset_scan_result asr
	LEFT JOIN asset a ON asr.asset_id = a.id
	WHERE asr.id = $1`

	return DB.QueryRow(query, id), nil
}

// DeleteAssetScanResultsByTaskID 删除任务的所有扫描结果
func DeleteAssetScanResultsByTaskID(taskID int) error {
	_, err := DB.Exec("DELETE FROM asset_scan_result WHERE task_id = $1", taskID)
	return err
}

// GetAssetScanStats 获取资产扫描统计信息
func GetAssetScanStats(taskID int) (map[string]int, error) {
	stats := make(map[string]int)

	query := `
	SELECT 
		COUNT(*) as total,
		SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success,
		SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed,
		SUM(CASE WHEN a.is_alive = TRUE THEN 1 ELSE 0 END) as alive
	FROM asset_scan_result asr
	LEFT JOIN asset a ON asr.asset_id = a.id
	WHERE asr.task_id = $1`

	var total, success, failed, alive int
	err := DB.QueryRow(query, taskID).Scan(&total, &success, &failed, &alive)
	if err != nil {
		return stats, err
	}

	stats["total"] = total
	stats["success"] = success
	stats["failed"] = failed
	stats["alive"] = alive

	return stats, nil
}
