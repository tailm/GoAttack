package postgres

import (
	"GoAttack/model"
	"database/sql"
	"encoding/json"
	"fmt"
)

// ============================================
// 端口资产管理模块
// 说明: 处理端口扫描结果的增删改查操作
// ============================================

// SaveAssetPort 保存端口资产信息
func SaveAssetPort(port *model.AssetPort) error {
	query := `
	INSERT INTO asset_port (
		task_id, asset_id, ip, port, protocol, state,
		service_name, service_product, service_version, service_extra_info,
		service_hostname, service_os_type, service_device_type, service_confidence,
		banner, fingerprint_method, raw_response, cpes, scripts
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	ON CONFLICT (ip, port, task_id) DO UPDATE SET
		state = EXCLUDED.state,
		service_name = EXCLUDED.service_name,
		service_product = EXCLUDED.service_product,
		service_version = EXCLUDED.service_version,
		service_extra_info = EXCLUDED.service_extra_info,
		service_hostname = EXCLUDED.service_hostname,
		service_os_type = EXCLUDED.service_os_type,
		service_device_type = EXCLUDED.service_device_type,
		service_confidence = EXCLUDED.service_confidence,
		banner = EXCLUDED.banner,
		fingerprint_method = EXCLUDED.fingerprint_method,
		raw_response = EXCLUDED.raw_response,
		cpes = EXCLUDED.cpes,
		scripts = EXCLUDED.scripts,
		last_seen = CURRENT_TIMESTAMP
	`

	// 转换CPEs为JSON
	cpesJSON, err := json.Marshal(port.CPEs)
	if err != nil {
		cpesJSON = []byte("[]")
	}

	// 转换Scripts为JSON
	scriptsJSON, err := json.Marshal(port.Scripts)
	if err != nil {
		scriptsJSON = []byte("{}")
	}

	_, err = DB.Exec(query,
		port.TaskID, port.AssetID, port.IP, port.Port, port.Protocol, port.State,
		port.ServiceName, port.ServiceProduct, port.ServiceVersion, port.ServiceExtraInfo,
		port.ServiceHostname, port.ServiceOSType, port.ServiceDeviceType, port.ServiceConfidence,
		port.Banner, port.FingerprintMethod, port.RawResponse, string(cpesJSON), string(scriptsJSON),
	)

	return err
}

// GetAssetPortsByTaskID 根据任务ID获取端口列表
func GetAssetPortsByTaskID(taskID int) (*sql.Rows, error) {
	query := `
	SELECT 
		id, task_id, asset_id, ip, port, protocol, state,
		service_name, service_product, service_version, service_extra_info,
		service_hostname, service_os_type, service_device_type, service_confidence,
		banner, fingerprint_method, raw_response, cpes, scripts,
		discovered_at, last_seen
	FROM asset_port
	WHERE task_id = $1
	ORDER BY ip, port`

	return DB.Query(query, taskID)
}

// GetAssetPortsByIP 根据IP获取端口列表
func GetAssetPortsByIP(ip string) (*sql.Rows, error) {
	query := `
	SELECT 
		id, task_id, asset_id, ip, port, protocol, state,
		service_name, service_product, service_version, service_extra_info,
		service_hostname, service_os_type, service_device_type, service_confidence,
		banner, fingerprint_method, raw_response, cpes, scripts,
		discovered_at, last_seen
	FROM asset_port
	WHERE ip = $1 AND state = 'open'
	ORDER BY last_seen DESC, port`

	return DB.Query(query, ip)
}

// GetAssetPortByID 根据ID获取单个端口信息
func GetAssetPortByID(id int64) (*model.AssetPort, error) {
	query := `
	SELECT 
		id, task_id, asset_id, ip, port, protocol, state,
		service_name, service_product, service_version, service_extra_info,
		service_hostname, service_os_type, service_device_type, service_confidence,
		banner, fingerprint_method, raw_response, cpes, scripts,
		discovered_at, last_seen
	FROM asset_port
	WHERE id = $1`

	port := &model.AssetPort{}
	var cpesJSON, scriptsJSON sql.NullString

	err := DB.QueryRow(query, id).Scan(
		&port.ID, &port.TaskID, &port.AssetID, &port.IP, &port.Port, &port.Protocol, &port.State,
		&port.ServiceName, &port.ServiceProduct, &port.ServiceVersion, &port.ServiceExtraInfo,
		&port.ServiceHostname, &port.ServiceOSType, &port.ServiceDeviceType, &port.ServiceConfidence,
		&port.Banner, &port.FingerprintMethod, &port.RawResponse, &cpesJSON, &scriptsJSON,
		&port.DiscoveredAt, &port.LastSeen,
	)

	if err != nil {
		return nil, err
	}

	// 解析JSON字段
	if cpesJSON.Valid && cpesJSON.String != "" {
		json.Unmarshal([]byte(cpesJSON.String), &port.CPEs)
	}

	if scriptsJSON.Valid && scriptsJSON.String != "" {
		json.Unmarshal([]byte(scriptsJSON.String), &port.Scripts)
	}

	return port, nil
}

// GetAssetPortSummary 获取端口资产统计摘要
func GetAssetPortSummary(taskID int) (*model.AssetPortSummary, error) {
	summary := &model.AssetPortSummary{
		ServiceCounts: make(map[string]int),
		TopPorts:      make([]model.PortCount, 0),
		TopServices:   make([]model.ServiceCount, 0),
	}

	// 1. 统计总端口数和开放端口数
	countQuery := `
	SELECT 
		COUNT(*) as total_ports,
		SUM(CASE WHEN state = 'open' THEN 1 ELSE 0 END) as open_ports,
		MAX(discovered_at) as latest_discovery
	FROM asset_port
	WHERE task_id = $1`

	err := DB.QueryRow(countQuery, taskID).Scan(
		&summary.TotalPorts,
		&summary.OpenPorts,
		&summary.LatestDiscovery,
	)
	if err != nil {
		return nil, err
	}

	// 2. 统计各服务数量
	serviceQuery := `
	SELECT service_name, COUNT(*) as count
	FROM asset_port
	WHERE task_id = $1 AND service_name IS NOT NULL AND service_name != ''
	GROUP BY service_name`

	rows, err := DB.Query(serviceQuery, taskID)
	if err != nil {
		return summary, err
	}
	defer rows.Close()

	for rows.Next() {
		var service string
		var count int
		if err := rows.Scan(&service, &count); err == nil {
			summary.ServiceCounts[service] = count
		}
	}

	// 3. Top 10 常见端口
	portQuery := `
	SELECT port, COUNT(*) as count
	FROM asset_port
	WHERE task_id = $1 AND state = 'open'
	GROUP BY port
	ORDER BY count DESC
	LIMIT 10`

	rows, err = DB.Query(portQuery, taskID)
	if err != nil {
		return summary, err
	}
	defer rows.Close()

	for rows.Next() {
		var pc model.PortCount
		if err := rows.Scan(&pc.Port, &pc.Count); err == nil {
			summary.TopPorts = append(summary.TopPorts, pc)
		}
	}

	// 4. Top 10 常见服务
	topServiceQuery := `
	SELECT service_name, COUNT(*) as count
	FROM asset_port
	WHERE task_id = $1 AND service_name IS NOT NULL AND service_name != ''
	GROUP BY service_name
	ORDER BY count DESC
	LIMIT 10`

	rows, err = DB.Query(topServiceQuery, taskID)
	if err != nil {
		return summary, err
	}
	defer rows.Close()

	for rows.Next() {
		var sc model.ServiceCount
		if err := rows.Scan(&sc.Service, &sc.Count); err == nil {
			summary.TopServices = append(summary.TopServices, sc)
		}
	}

	return summary, nil
}

// DeleteAssetPortsByTaskID 删除任务的所有端口记录
func DeleteAssetPortsByTaskID(taskID int) error {
	_, err := DB.Exec("DELETE FROM asset_port WHERE task_id = $1", taskID)
	return err
}

// DeleteAssetPortByID 删除指定端口记录
func DeleteAssetPortByID(id int64) error {
	_, err := DB.Exec("DELETE FROM asset_port WHERE id = $1", id)
	return err
}

// SearchAssetPorts 搜索端口资产（支持多条件查询）
func SearchAssetPorts(filters map[string]interface{}, page, pageSize int) (*sql.Rows, int, error) {
	whereClause := "WHERE 1=1"
	args := make([]interface{}, 0)
	argIndex := 1

	// IP筛选
	if ip, ok := filters["ip"].(string); ok && ip != "" {
		whereClause += fmt.Sprintf(" AND ip = $%d", argIndex)
		args = append(args, ip)
		argIndex++
	}

	// 端口筛选
	if port, ok := filters["port"].(int); ok && port > 0 {
		whereClause += fmt.Sprintf(" AND port = $%d", argIndex)
		args = append(args, port)
		argIndex++
	}

	// 服务筛选
	if service, ok := filters["service"].(string); ok && service != "" {
		whereClause += fmt.Sprintf(" AND service_name = $%d", argIndex)
		args = append(args, service)
		argIndex++
	}

	// 状态筛选
	if state, ok := filters["state"].(string); ok && state != "" {
		whereClause += fmt.Sprintf(" AND state = $%d", argIndex)
		args = append(args, state)
		argIndex++
	}

	// 任务ID筛选
	if taskID, ok := filters["task_id"].(int); ok && taskID > 0 {
		whereClause += fmt.Sprintf(" AND task_id = $%d", argIndex)
		args = append(args, taskID)
		argIndex++
	}

	// 查询总数
	var total int
	countQuery := "SELECT COUNT(*) FROM asset_port " + whereClause
	err := DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	offset := (page - 1) * pageSize
	query := `SELECT 
		id, task_id, asset_id, ip, port, protocol, state,
		service_name, service_product, service_version, service_extra_info,
		service_hostname, service_os_type, service_device_type, service_confidence,
		banner, fingerprint_method, raw_response, cpes, scripts,
		discovered_at, last_seen
	FROM asset_port ` + whereClause + fmt.Sprintf(" ORDER BY last_seen DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	rows, err := DB.Query(query, args...)
	return rows, total, err
}
