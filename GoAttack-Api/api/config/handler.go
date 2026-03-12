package config

import (
	"GoAttack/common/log"
	"GoAttack/model"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册系统配置相关路由
func RegisterRoutes(r *gin.RouterGroup) {
	// 初始化数据库连接
	Init()
	
	r.GET("/config/intelligence-sources", GetIntelligenceSources)
	r.PUT("/config/intelligence-sources/:id", UpdateIntelligenceSource)
	r.GET("/config/detection-rules", GetDetectionRules)
	r.POST("/config/detection-rules", CreateDetectionRule)
	r.GET("/config/alert-configs", GetAlertConfigs)
	r.GET("/stats/health", GetHealthStatus)
}

// GetIntelligenceSources 获取漏洞情报源配置
func GetIntelligenceSources(c *gin.Context) {
	query := `
		SELECT id, name, url, type, enabled, sync_interval, 
		       last_sync, last_sync_status, last_sync_error, 
		       config, created_at, updated_at
		FROM intelligence_source
		ORDER BY enabled DESC, name ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Error("查询漏洞情报源配置失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}
	defer rows.Close()

	var sources []model.IntelligenceSource
	for rows.Next() {
		var source model.IntelligenceSource
		var configJSON []byte
		var lastSync sql.NullTime

		err := rows.Scan(
			&source.ID, &source.Name, &source.URL, &source.Type,
			&source.Enabled, &source.SyncInterval, &lastSync,
			&source.LastSyncStatus, &source.LastSyncError,
			&configJSON, &source.CreatedAt, &source.UpdatedAt,
		)
		if err != nil {
			log.Error("扫描漏洞情报源配置失败", "error", err)
			continue
		}

		// 解析JSON字段
		if configJSON != nil {
			json.Unmarshal(configJSON, &source.Config)
		}

		// 处理可为空的时间字段
		if lastSync.Valid {
			source.LastSync = &lastSync.Time
		}

		sources = append(sources, source)
	}

	if err = rows.Err(); err != nil {
		log.Error("遍历漏洞情报源配置失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "success",
		"data": sources,
	})
}

// UpdateIntelligenceSource 更新漏洞情报源配置
func UpdateIntelligenceSource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40001,
			"msg":  "ID参数错误",
			"data": nil,
		})
		return
	}

	var req model.IntelligenceSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40001,
			"msg":  "请求参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 构建更新字段
	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != "" {
		updateFields = append(updateFields, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, req.Name)
		argIndex++
	}

	if req.URL != "" {
		updateFields = append(updateFields, fmt.Sprintf("url = $%d", argIndex))
		args = append(args, req.URL)
		argIndex++
	}

	if req.Type != "" {
		updateFields = append(updateFields, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, req.Type)
		argIndex++
	}

	updateFields = append(updateFields, fmt.Sprintf("enabled = $%d", argIndex))
	args = append(args, req.Enabled)
	argIndex++

	if req.SyncInterval > 0 {
		updateFields = append(updateFields, fmt.Sprintf("sync_interval = $%d", argIndex))
		args = append(args, req.SyncInterval)
		argIndex++
	}

	if req.Config != nil {
		configJSON, _ := json.Marshal(req.Config)
		updateFields = append(updateFields, fmt.Sprintf("config = $%d", argIndex))
		args = append(args, configJSON)
		argIndex++
	}

	// 添加更新时间
	updateFields = append(updateFields, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// 添加ID条件
	args = append(args, id)

	// 构建更新语句
	updateQuery := fmt.Sprintf(`
		UPDATE intelligence_source 
		SET %s 
		WHERE id = $%d 
		RETURNING id, name, enabled, sync_interval, updated_at
	`, strings.Join(updateFields, ", "), argIndex)

	var sourceID int
	var name string
	var enabled bool
	var syncInterval int
	var updatedAt time.Time

	err = db.QueryRow(updateQuery, args...).Scan(
		&sourceID, &name, &enabled, &syncInterval, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code": 40401,
				"msg":  "漏洞情报源不存在",
				"data": nil,
			})
			return
		}
		log.Error("更新漏洞情报源配置失败", "error", err, "id", id)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "更新失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "配置更新成功",
		"data": gin.H{
			"id":            sourceID,
			"name":          name,
			"enabled":       enabled,
			"sync_interval": syncInterval,
			"updated_at":    updatedAt.Format(time.RFC3339),
		},
	})
}

// GetDetectionRules 获取检测规则
func GetDetectionRules(c *gin.Context) {
	enabled := c.Query("enabled")
	ruleType := c.Query("type")

	// 构建查询条件
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if enabled != "" {
		enabledBool, err := strconv.ParseBool(enabled)
		if err == nil {
			whereClause += fmt.Sprintf(" AND enabled = $%d", argIndex)
			args = append(args, enabledBool)
			argIndex++
		}
	}

	if ruleType != "" {
		whereClause += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, ruleType)
		argIndex++
	}

	query := fmt.Sprintf(`
		SELECT id, name, description, type, condition, action, 
		       severity, enabled, priority, tags, created_by, 
		       created_at, updated_at
		FROM detection_rule
		%s
		ORDER BY priority DESC, created_at DESC
	`, whereClause)

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Error("查询检测规则失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}
	defer rows.Close()

	var rules []model.DetectionRule
	for rows.Next() {
		var rule model.DetectionRule
		var tagsJSON []byte

		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.Type,
			&rule.Condition, &rule.Action, &rule.Severity,
			&rule.Enabled, &rule.Priority, &tagsJSON,
			&rule.CreatedBy, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			log.Error("扫描检测规则失败", "error", err)
			continue
		}

		// 解析JSON字段
		if tagsJSON != nil {
			json.Unmarshal(tagsJSON, &rule.Tags)
		}

		rules = append(rules, rule)
	}

	if err = rows.Err(); err != nil {
		log.Error("遍历检测规则失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "success",
		"data": rules,
	})
}

// CreateDetectionRule 创建检测规则
func CreateDetectionRule(c *gin.Context) {
	var req model.DetectionRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40001,
			"msg":  "请求参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 验证参数
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40002,
			"msg":  "规则名称不能为空",
			"data": nil,
		})
		return
	}

	if req.Type == "" {
		req.Type = "custom"
	}

	if req.Condition == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40003,
			"msg":  "规则条件不能为空",
			"data": nil,
		})
		return
	}

	if req.Action == "" {
		req.Action = "alert"
	}

	if req.Severity == "" {
		req.Severity = "medium"
	}

	// 验证类型
	validTypes := []string{"port", "service", "version", "cve", "custom"}
	validType := false
	for _, t := range validTypes {
		if t == req.Type {
			validType = true
			break
		}
	}
	if !validType {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40004,
			"msg":  "无效的规则类型",
			"data": nil,
		})
		return
	}

	// 验证动作
	validActions := []string{"alert", "block", "log", "report"}
	validAction := false
	for _, a := range validActions {
		if a == req.Action {
			validAction = true
			break
		}
	}
	if !validAction {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40005,
			"msg":  "无效的规则动作",
			"data": nil,
		})
		return
	}

	// 验证严重程度
	validSeverities := []string{"critical", "high", "medium", "low", "info"}
	validSeverity := false
	for _, s := range validSeverities {
		if s == req.Severity {
			validSeverity = true
			break
		}
	}
	if !validSeverity {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40006,
			"msg":  "无效的严重程度",
			"data": nil,
		})
		return
	}

	// 设置默认值
	if req.Priority <= 0 {
		req.Priority = 0
	}

	// 插入检测规则
	query := `
		INSERT INTO detection_rule (
			name, description, type, condition, action, severity,
			enabled, priority, tags, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, name, created_at
	`

	tagsJSON, _ := json.Marshal(req.Tags)
	createdBy := "system" // 这里应该从上下文中获取当前用户
	now := time.Now()

	var ruleID int
	var ruleName string
	var createdAt time.Time

	err := db.QueryRow(query,
		req.Name, req.Description, req.Type, req.Condition,
		req.Action, req.Severity, req.Enabled, req.Priority,
		tagsJSON, createdBy, now, now,
	).Scan(&ruleID, &ruleName, &createdAt)

	if err != nil {
		log.Error("创建检测规则失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "创建失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "检测规则创建成功",
		"data": gin.H{
			"id":         ruleID,
			"name":       ruleName,
			"created_at": createdAt.Format(time.RFC3339),
		},
	})
}

// GetAlertConfigs 获取预警配置
func GetAlertConfigs(c *gin.Context) {
	query := `
		SELECT id, name, description, severity_filter, source_filter,
		       enabled, notification_channels, schedule,
		       created_at, updated_at
		FROM alert_config
		ORDER BY enabled DESC, name ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Error("查询预警配置失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}
	defer rows.Close()

	var configs []model.AlertConfig
	for rows.Next() {
		var config model.AlertConfig
		var severityFilterJSON, sourceFilterJSON, channelsJSON, scheduleJSON []byte

		err := rows.Scan(
			&config.ID, &config.Name, &config.Description,
			&severityFilterJSON, &sourceFilterJSON, &config.Enabled,
			&channelsJSON, &scheduleJSON, &config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			log.Error("扫描预警配置失败", "error", err)
			continue
		}

		// 解析JSON字段
		if severityFilterJSON != nil {
			json.Unmarshal(severityFilterJSON, &config.SeverityFilter)
		}
		if sourceFilterJSON != nil {
			json.Unmarshal(sourceFilterJSON, &config.SourceFilter)
		}
		if channelsJSON != nil {
			json.Unmarshal(channelsJSON, &config.NotificationChannels)
		}
		if scheduleJSON != nil {
			json.Unmarshal(scheduleJSON, &config.Schedule)
		}

		configs = append(configs, config)
	}

	if err = rows.Err(); err != nil {
		log.Error("遍历预警配置失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "success",
		"data": configs,
	})
}

// GetHealthStatus 获取系统健康状态
func GetHealthStatus(c *gin.Context) {
	// 检查数据库连接
	dbHealth := checkDatabaseHealth()
	
	// 检查Redis连接
	redisHealth := checkRedisHealth()
	
	// 检查调度器状态
	schedulerHealth := checkSchedulerHealth()
	
	// 检查扫描器状态
	scannerHealth := checkScannerHealth()
	
	// 获取系统指标
	metrics := getSystemMetrics()

	// 确定整体状态
	overallStatus := "healthy"
	if dbHealth.Status != "healthy" || redisHealth.Status != "healthy" {
		overallStatus = "degraded"
	}
	if dbHealth.Status == "unhealthy" || redisHealth.Status == "unhealthy" {
		overallStatus = "unhealthy"
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "success",
		"data": model.HealthResponse{
			Status: overallStatus,
			Components: map[string]model.ComponentHealth{
				"database":  dbHealth,
				"redis":     redisHealth,
				"scheduler": schedulerHealth,
				"scanner":   scannerHealth,
			},
			Metrics:     metrics,
			LastUpdated: time.Now(),
		},
	})
}

// checkDatabaseHealth 检查数据库健康状态
func checkDatabaseHealth() model.ComponentHealth {
	start := time.Now()
	
	// 检查数据库连接
	var result string
	err := db.QueryRow("SELECT 'OK'").Scan(&result)
	
	latency := int(time.Since(start).Milliseconds())
	
	if err != nil {
		log.Error("数据库健康检查失败", "error", err)
		return model.ComponentHealth{
			Status:  "unhealthy",
			Latency: latency,
			Details: gin.H{"error": err.Error()},
		}
	}

	// 检查表数量
	var tableCount int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = 'public'
	`).Scan(&tableCount)

	if err != nil {
		log.Error("检查数据库表数量失败", "error", err)
		return model.ComponentHealth{
			Status:  "degraded",
			Latency: latency,
			Details: gin.H{"error": err.Error()},
		}
	}

	return model.ComponentHealth{
		Status:  "healthy",
		Latency: latency,
		Details: gin.H{
			"table_count": tableCount,
			"connection":  "established",
		},
	}
}

// checkRedisHealth 检查Redis健康状态
func checkRedisHealth() model.ComponentHealth {
	// 这里需要实现Redis健康检查
	// 暂时返回模拟数据
	return model.ComponentHealth{
		Status:  "healthy",
		Latency: 2,
		Details: gin.H{
			"connection": "established",
			"memory_used": "256MB",
			"connected_clients": 5,
		},
	}
}

// checkSchedulerHealth 检查调度器健康状态
func checkSchedulerHealth() model.ComponentHealth {
	// 这里需要实现调度器健康检查
	// 暂时返回模拟数据
	return model.ComponentHealth{
		Status:  "healthy",
		Latency: 0,
		Details: gin.H{
			"running_jobs": 3,
			"queued_jobs":  0,
			"last_run":     time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
		},
	}
}

// checkScannerHealth 检查扫描器健康状态
func checkScannerHealth() model.ComponentHealth {
	// 这里需要实现扫描器健康检查
	// 暂时返回模拟数据
	return model.ComponentHealth{
		Status:  "healthy",
		Latency: 0,
		Details: gin.H{
			"active_scans": 2,
			"total_scans":  150,
			"last_scan":    time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
		},
	}
}

// getSystemMetrics 获取系统指标
func getSystemMetrics() map[string]interface{} {
	// 查询漏洞数量
	var vulnerabilityCount int
	err := db.QueryRow("SELECT COUNT(*) FROM vulnerability_intelligence").Scan(&vulnerabilityCount)
	if err != nil {
		vulnerabilityCount = 0
	}

	// 查询检测数量
	var detectionCount int
	err = db.QueryRow("SELECT COUNT(*) FROM detection_result").Scan(&detectionCount)
	if err != nil {
		detectionCount = 0
	}

	// 查询预警数量
	var alertCount int
	err = db.QueryRow("SELECT COUNT(*) FROM alert WHERE status = 'unread'").Scan(&alertCount)
	if err != nil {
		alertCount = 0
	}

	// 这里可以添加更多的系统指标
	// 例如：内存使用率、CPU使用率、磁盘使用率等

	return map[string]interface{}{
		"vulnerability_count": vulnerabilityCount,
		"detection_count":     detectionCount,
		"alert_count":         alertCount,
		"uptime":              int64(86400), // 模拟24小时运行时间
		"memory_usage":        65.5,         // 模拟内存使用率
		"cpu_usage":           42.3,         // 模拟CPU使用率
	}
}

