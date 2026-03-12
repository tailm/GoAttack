package detection

import (
	"GoAttack/common/log"
	"GoAttack/model"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册检测任务相关路由
func RegisterRoutes(r *gin.RouterGroup) {
	// 初始化数据库连接
	Init()
	
	r.POST("/detection/tasks", CreateDetectionTask)
	r.GET("/detection/tasks", GetDetectionTaskList)
	r.GET("/detection/tasks/:id", GetDetectionTaskDetail)
	r.PUT("/detection/tasks/:id/status", UpdateDetectionTaskStatus)
	r.GET("/detection/tasks/:id/results", GetDetectionResults)
	r.GET("/detection/stats", GetDetectionStats)
}

// CreateDetectionTask 创建检测任务
func CreateDetectionTask(c *gin.Context) {
	var req model.DetectionTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40001,
			"msg":  "请求参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 验证参数
	if len(req.Targets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40002,
			"msg":  "目标不能为空",
			"data": nil,
		})
		return
	}

	if req.DetectionType == "" {
		req.DetectionType = "vulnerability"
	}

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		log.Error("开始事务失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "创建任务失败: " + err.Error(),
			"data": nil,
		})
		return
	}
	defer tx.Rollback()

	// 创建主任务
	taskQuery := `
		INSERT INTO task (name, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	
	var taskID int
	taskName := fmt.Sprintf("漏洞检测任务 - %s", req.Name)
	taskDescription := req.Description
	if taskDescription == "" {
		taskDescription = fmt.Sprintf("对 %d 个目标进行漏洞检测", len(req.Targets))
	}
	
	err = tx.QueryRow(taskQuery, 
		taskName, taskDescription, "pending", time.Now(), time.Now()).Scan(&taskID)
	if err != nil {
		log.Error("创建主任务失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50002,
			"msg":  "创建任务失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 创建检测任务
	detectionTaskQuery := `
		INSERT INTO detection_task (
			task_id, detection_type, config, schedule, enabled, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	
	configJSON, _ := json.Marshal(req.Config)
	scheduleJSON, _ := json.Marshal(req.Schedule)
	
	var detectionTaskID int
	err = tx.QueryRow(detectionTaskQuery,
		taskID, req.DetectionType, configJSON, scheduleJSON, true, time.Now(), time.Now()).Scan(&detectionTaskID)
	if err != nil {
		log.Error("创建检测任务失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50003,
			"msg":  "创建任务失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		log.Error("提交事务失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50004,
			"msg":  "创建任务失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 启动后台检测任务
	go func() {
		log.Info("启动检测任务", "task_id", taskID, "detection_task_id", detectionTaskID)
		// TODO: 实现实际的检测逻辑
	}()

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "检测任务创建成功",
		"data": gin.H{
			"id":                 taskID,
			"name":               taskName,
			"status":             "pending",
			"created_at":         time.Now().Format(time.RFC3339),
			"estimated_duration": 1800, // 预计30分钟
		},
	})
}

// GetDetectionTaskList 获取检测任务列表
func GetDetectionTaskList(c *gin.Context) {
	var query struct {
		Page          int    `form:"page"`
		PageSize      int    `form:"pageSize"`
		Status        string `form:"status"`
		DetectionType string `form:"detection_type"`
		StartDate     string `form:"start_date"`
		EndDate       string `form:"end_date"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40001,
			"msg":  "参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// 构建查询条件
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if query.Status != "" {
		whereClause += fmt.Sprintf(" AND t.status = $%d", argIndex)
		args = append(args, query.Status)
		argIndex++
	}

	if query.DetectionType != "" {
		whereClause += fmt.Sprintf(" AND dt.detection_type = $%d", argIndex)
		args = append(args, query.DetectionType)
		argIndex++
	}

	if query.StartDate != "" {
		whereClause += fmt.Sprintf(" AND t.created_at >= $%d", argIndex)
		args = append(args, query.StartDate)
		argIndex++
	}

	if query.EndDate != "" {
		whereClause += fmt.Sprintf(" AND t.created_at <= $%d", argIndex)
		args = append(args, query.EndDate)
		argIndex++
	}

	// 查询总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM task t
		LEFT JOIN detection_task dt ON t.id = dt.task_id
		%s
	`, whereClause)
	
	var total int
	err := db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		log.Error("查询检测任务总数失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 查询数据
	offset := (query.Page - 1) * query.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT 
			t.id, t.name, t.description, t.status, t.created_at, t.updated_at,
			dt.detection_type,
			COALESCE((
				SELECT COUNT(*) 
				FROM detection_result dr 
				WHERE dr.task_id = t.id
			), 0) as result_count,
			COALESCE((
				SELECT COUNT(*) 
				FROM detection_result dr 
				WHERE dr.task_id = t.id AND dr.risk_level = 'critical'
			), 0) as critical_count,
			COALESCE((
				SELECT COUNT(*) 
				FROM detection_result dr 
				WHERE dr.task_id = t.id AND dr.risk_level = 'high'
			), 0) as high_count,
			COALESCE((
				SELECT COUNT(*) 
				FROM detection_result dr 
				WHERE dr.task_id = t.id AND dr.risk_level = 'medium'
			), 0) as medium_count,
			COALESCE((
				SELECT COUNT(*) 
				FROM detection_result dr 
				WHERE dr.task_id = t.id AND dr.risk_level = 'low'
			), 0) as low_count
		FROM task t
		LEFT JOIN detection_task dt ON t.id = dt.task_id
		%s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, query.PageSize, offset)
	rows, err := db.Query(dataQuery, args...)
	if err != nil {
		log.Error("查询检测任务列表失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}
	defer rows.Close()

	type TaskItem struct {
		ID            int       `json:"id"`
		Name          string    `json:"name"`
		Description   string    `json:"description"`
		DetectionType string    `json:"detection_type"`
		Status        string    `json:"status"`
		ResultCount   int       `json:"result_count"`
		CriticalCount int       `json:"critical_count"`
		HighCount     int       `json:"high_count"`
		MediumCount   int       `json:"medium_count"`
		LowCount      int       `json:"low_count"`
		CreatedAt     time.Time `json:"created_at"`
		UpdatedAt     time.Time `json:"updated_at"`
	}

	var items []TaskItem
	for rows.Next() {
		var item TaskItem
		err := rows.Scan(
			&item.ID, &item.Name, &item.Description, &item.Status,
			&item.CreatedAt, &item.UpdatedAt, &item.DetectionType,
			&item.ResultCount, &item.CriticalCount, &item.HighCount,
			&item.MediumCount, &item.LowCount,
		)
		if err != nil {
			log.Error("扫描检测任务数据失败", "error", err)
			continue
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		log.Error("遍历检测任务数据失败", "error", err)
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
		"data": model.PaginatedResponse{
			Total:    total,
			Page:     query.Page,
			PageSize: query.PageSize,
			Items:    items,
		},
	})
}

// GetDetectionTaskDetail 获取检测任务详情
func GetDetectionTaskDetail(c *gin.Context) {
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

	query := `
		SELECT 
			t.id, t.name, t.description, t.status, t.created_at, t.updated_at,
			dt.detection_type, dt.config, dt.schedule, dt.last_run, dt.next_run, dt.enabled,
			COALESCE((
				SELECT COUNT(*) 
				FROM detection_result dr 
				WHERE dr.task_id = t.id
			), 0) as result_count,
			COALESCE((
				SELECT COUNT(*) 
				FROM detection_result dr 
				WHERE dr.task_id = t.id AND dr.risk_level = 'critical'
			), 0) as critical_count,
			COALESCE((
				SELECT COUNT(*) 
				FROM detection_result dr 
				WHERE dr.task_id = t.id AND dr.risk_level = 'high'
			), 0) as high_count,
			COALESCE((
				SELECT COUNT(*) 
				FROM detection_result dr 
				WHERE dr.task_id = t.id AND dr.risk_level = 'medium'
			), 0) as medium_count,
			COALESCE((
				SELECT COUNT(*) 
				FROM detection_result dr 
				WHERE dr.task_id = t.id AND dr.risk_level = 'low'
			), 0) as low_count
		FROM task t
		LEFT JOIN detection_task dt ON t.id = dt.task_id
		WHERE t.id = $1
	`

	var task struct {
		ID            int                    `json:"id"`
		Name          string                 `json:"name"`
		Description   string                 `json:"description"`
		Status        string                 `json:"status"`
		DetectionType string                 `json:"detection_type"`
		Config        map[string]interface{} `json:"config"`
		Schedule      map[string]interface{} `json:"schedule"`
		LastRun       *time.Time             `json:"last_run"`
		NextRun       *time.Time             `json:"next_run"`
		Enabled       bool                   `json:"enabled"`
		ResultCount   int                    `json:"result_count"`
		CriticalCount int                    `json:"critical_count"`
		HighCount     int                    `json:"high_count"`
		MediumCount   int                    `json:"medium_count"`
		LowCount      int                    `json:"low_count"`
		CreatedAt     time.Time              `json:"created_at"`
		UpdatedAt     time.Time              `json:"updated_at"`
	}

	var configJSON, scheduleJSON []byte
	err = db.QueryRow(query, id).Scan(
		&task.ID, &task.Name, &task.Description, &task.Status,
		&task.CreatedAt, &task.UpdatedAt, &task.DetectionType,
		&configJSON, &scheduleJSON, &task.LastRun, &task.NextRun, &task.Enabled,
		&task.ResultCount, &task.CriticalCount, &task.HighCount,
		&task.MediumCount, &task.LowCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code": 40401,
				"msg":  "检测任务不存在",
				"data": nil,
			})
			return
		}
		log.Error("查询检测任务详情失败", "error", err, "id", id)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 解析JSON字段
	if configJSON != nil {
		json.Unmarshal(configJSON, &task.Config)
	}
	if scheduleJSON != nil {
		json.Unmarshal(scheduleJSON, &task.Schedule)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "success",
		"data": task,
	})
}

// UpdateDetectionTaskStatus 更新检测任务状态
func UpdateDetectionTaskStatus(c *gin.Context) {
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

	var req struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40001,
			"msg":  "请求参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 验证状态值
	validStatuses := []string{"pending", "running", "paused", "completed", "failed", "cancelled"}
	valid := false
	for _, s := range validStatuses {
		if s == req.Status {
			valid = true
			break
		}
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40002,
			"msg":  "无效的状态值",
			"data": nil,
		})
		return
	}

	// 更新任务状态
	query := `
		UPDATE task 
		SET status = $1, updated_at = $2 
		WHERE id = $3 
		RETURNING id, status, updated_at
	`

	var updatedID int
	var status string
	var updatedAt time.Time
	err = db.QueryRow(query, req.Status, time.Now(), id).Scan(&updatedID, &status, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code": 40401,
				"msg":  "检测任务不存在",
				"data": nil,
			})
			return
		}
		log.Error("更新检测任务状态失败", "error", err, "id", id)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "更新失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 记录状态变更日志
	log.Info("检测任务状态更新", "task_id", id, "status", req.Status, "reason", req.Reason)

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "任务状态更新成功",
		"data": gin.H{
			"id":         updatedID,
			"status":     status,
			"updated_at": updatedAt.Format(time.RFC3339),
		},
	})
}

// GetDetectionResults 获取检测结果
func GetDetectionResults(c *gin.Context) {
	idStr := c.Param("id")
	taskID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40001,
			"msg":  "任务ID参数错误",
			"data": nil,
		})
		return
	}

	var query model.DetectionResultQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40001,
			"msg":  "参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// 构建查询条件
	whereClause := "WHERE dr.task_id = $1"
	args := []interface{}{taskID}
	argIndex := 2

	if query.Severity != "" {
		whereClause += fmt.Sprintf(" AND dr.risk_level = $%d", argIndex)
		args = append(args, query.Severity)
		argIndex++
	}

	if query.Status != "" {
		whereClause += fmt.Sprintf(" AND dr.status = $%d", argIndex)
		args = append(args, query.Status)
		argIndex++
	}

	if query.Verified != nil {
		whereClause += fmt.Sprintf(" AND dr.verified = $%d", argIndex)
		args = append(args, *query.Verified)
		argIndex++
	}

	// 查询总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM detection_result dr
		%s
	`, whereClause)
	
	var total int
	err = db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		log.Error("查询检测结果总数失败", "error", err, "task_id", taskID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 查询数据
	offset := (query.Page - 1) * query.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT 
			dr.id, dr.task_id, dr.vuln_intel_id, dr.asset_id, dr.target,
			dr.poc_template_id, dr.status, dr.confidence, dr.evidence,
			dr.request_data, dr.response_data, dr.matched_pattern,
			dr.risk_level, dr.verified, dr.verified_by, dr.verification_notes,
			dr.created_at, dr.updated_at,
			vi.cve_id, vi.title, vi.severity, vi.cvss_score,
			a.ip, a.hostname
		FROM detection_result dr
		LEFT JOIN vulnerability_intelligence vi ON dr.vuln_intel_id = vi.id
		LEFT JOIN asset a ON dr.asset_id = a.id
		%s
		ORDER BY dr.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, query.PageSize, offset)
	rows, err := db.Query(dataQuery, args...)
	if err != nil {
		log.Error("查询检测结果失败", "error", err, "task_id", taskID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}
	defer rows.Close()

	type DetectionResultItem struct {
		ID                int                    `json:"id"`
		TaskID            int                    `json:"task_id"`
		VulnIntelID       *int                   `json:"vuln_intel_id"`
		AssetID           *int                   `json:"asset_id"`
		Target            string                 `json:"target"`
		POCTemplateID     *int                   `json:"poc_template_id"`
		Status            string                 `json:"status"`
		Confidence        *int                   `json:"confidence"`
		Evidence          map[string]interface{} `json:"evidence"`
		RequestData       string                 `json:"request_data"`
		ResponseData      string                 `json:"response_data"`
		MatchedPattern    string                 `json:"matched_pattern"`
		RiskLevel         string                 `json:"risk_level"`
		Verified          bool                   `json:"verified"`
		VerifiedBy        string                 `json:"verified_by"`
		VerificationNotes string                 `json:"verification_notes"`
		CreatedAt         time.Time              `json:"created_at"`
		UpdatedAt         time.Time              `json:"updated_at"`
		Vulnerability     *struct {
			CVEID     string  `json:"cve_id"`
			Title     string  `json:"title"`
			Severity  string  `json:"severity"`
			CVSSScore float64 `json:"cvss_score"`
		} `json:"vulnerability,omitempty"`
		Asset *struct {
			IP       string `json:"ip"`
			Hostname string `json:"hostname"`
		} `json:"asset,omitempty"`
	}

	var items []DetectionResultItem
	for rows.Next() {
		var item DetectionResultItem
		var vulnCVEID, vulnTitle, vulnSeverity sql.NullString
		var vulnCVSSScore sql.NullFloat64
		var assetIP, assetHostname sql.NullString
		var evidenceJSON []byte

		err := rows.Scan(
			&item.ID, &item.TaskID, &item.VulnIntelID, &item.AssetID, &item.Target,
			&item.POCTemplateID, &item.Status, &item.Confidence, &evidenceJSON,
			&item.RequestData, &item.ResponseData, &item.MatchedPattern,
			&item.RiskLevel, &item.Verified, &item.VerifiedBy, &item.VerificationNotes,
			&item.CreatedAt, &item.UpdatedAt,
			&vulnCVEID, &vulnTitle, &vulnSeverity, &vulnCVSSScore,
			&assetIP, &assetHostname,
		)
		if err != nil {
			log.Error("扫描检测结果数据失败", "error", err)
			continue
		}

		// 解析证据JSON
		if evidenceJSON != nil {
			json.Unmarshal(evidenceJSON, &item.Evidence)
		}

		// 设置漏洞信息
		if vulnCVEID.Valid {
			item.Vulnerability = &struct {
				CVEID     string  `json:"cve_id"`
				Title     string  `json:"title"`
				Severity  string  `json:"severity"`
				CVSSScore float64 `json:"cvss_score"`
			}{
				CVEID:     vulnCVEID.String,
				Title:     vulnTitle.String,
				Severity:  vulnSeverity.String,
				CVSSScore: vulnCVSSScore.Float64,
			}
		}

		// 设置资产信息
		if assetIP.Valid {
			item.Asset = &struct {
				IP       string `json:"ip"`
				Hostname string `json:"hostname"`
			}{
				IP:       assetIP.String,
				Hostname: assetHostname.String,
			}
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		log.Error("遍历检测结果数据失败", "error", err)
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
		"data": model.PaginatedResponse{
			Total:    total,
			Page:     query.Page,
			PageSize: query.PageSize,
			Items:    items,
		},
	})
}

// GetDetectionStats 获取检测统计信息
func GetDetectionStats(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "7")
	taskIDStr := c.Query("task_id")

	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 7
	}

	// 构建查询条件
	var whereClause string
	var args []interface{}
	argIndex := 1

	if taskIDStr != "" {
		taskID, err := strconv.Atoi(taskIDStr)
		if err == nil {
			whereClause = "WHERE task_id = $1 AND created_at >= CURRENT_DATE - INTERVAL '1 day' * $2"
			args = append(args, taskID, days)
			argIndex = 3
		} else {
			whereClause = "WHERE created_at >= CURRENT_DATE - INTERVAL '1 day' * $1"
			args = append(args, days)
			argIndex = 2
		}
	} else {
		whereClause = "WHERE created_at >= CURRENT_DATE - INTERVAL '1 day' * $1"
		args = append(args, days)
		argIndex = 2
	}

	// 查询总任务数
	var totalTasks int
	taskQuery := "SELECT COUNT(DISTINCT task_id) FROM detection_result " + whereClause
	err = db.QueryRow(taskQuery, args[:argIndex-1]...).Scan(&totalTasks)
	if err != nil {
		log.Error("查询总任务数失败", "error", err)
		totalTasks = 0
	}

	// 查询总检测数
	var totalDetections int
	detectionQuery := "SELECT COUNT(*) FROM detection_result " + whereClause
	err = db.QueryRow(detectionQuery, args[:argIndex-1]...).Scan(&totalDetections)
	if err != nil {
		log.Error("查询总检测数失败", "error", err)
		totalDetections = 0
	}

	// 按状态统计
	statusQuery := `
		SELECT status, COUNT(*) as count 
		FROM detection_result 
		` + whereClause + `
		GROUP BY status
	`
	statusRows, err := db.Query(statusQuery, args[:argIndex-1]...)
	if err != nil {
		log.Error("按状态统计失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}
	defer statusRows.Close()

	byStatus := make(map[string]int)
	for statusRows.Next() {
		var status string
		var count int
		if err := statusRows.Scan(&status, &count); err != nil {
			log.Error("扫描状态统计失败", "error", err)
			continue
		}
		byStatus[status] = count
	}

	// 按风险等级统计
	riskQuery := `
		SELECT risk_level, COUNT(*) as count 
		FROM detection_result 
		` + whereClause + `
		GROUP BY risk_level
	`
	riskRows, err := db.Query(riskQuery, args[:argIndex-1]...)
	if err != nil {
		log.Error("按风险等级统计失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}
	defer riskRows.Close()

	byRiskLevel := make(map[string]int)
	for riskRows.Next() {
		var riskLevel string
		var count int
		if err := riskRows.Scan(&riskLevel, &count); err != nil {
			log.Error("扫描风险等级统计失败", "error", err)
			continue
		}
		byRiskLevel[riskLevel] = count
	}

	// 查询热门漏洞
	topVulnQuery := `
		SELECT 
			vi.cve_id, vi.title, vi.severity, COUNT(*) as count
		FROM detection_result dr
		LEFT JOIN vulnerability_intelligence vi ON dr.vuln_intel_id = vi.id
		` + whereClause + `
		AND vi.cve_id IS NOT NULL
		GROUP BY vi.cve_id, vi.title, vi.severity
		ORDER BY count DESC
		LIMIT 10
	`
	topVulnRows, err := db.Query(topVulnQuery, args[:argIndex-1]...)
	if err != nil {
		log.Error("查询热门漏洞失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}
	defer topVulnRows.Close()

	var topVulnerabilities []model.TopVulnerability
	for topVulnRows.Next() {
		var vuln model.TopVulnerability
		if err := topVulnRows.Scan(&vuln.CVEID, &vuln.Title, &vuln.Severity, &vuln.Count); err != nil {
			log.Error("扫描热门漏洞失败", "error", err)
			continue
		}
		topVulnerabilities = append(topVulnerabilities, vuln)
	}

	// 查询热门资产
	topAssetQuery := `
		SELECT 
			a.ip, a.hostname, 
			COUNT(*) as vulnerability_count,
			COUNT(CASE WHEN dr.risk_level = 'critical' THEN 1 END) as critical_count
		FROM detection_result dr
		LEFT JOIN asset a ON dr.asset_id = a.id
		` + whereClause + `
		AND a.ip IS NOT NULL
		GROUP BY a.ip, a.hostname
		ORDER BY vulnerability_count DESC
		LIMIT 10
	`
	topAssetRows, err := db.Query(topAssetQuery, args[:argIndex-1]...)
	if err != nil {
		log.Error("查询热门资产失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}
	defer topAssetRows.Close()

	var topAssets []model.TopAsset
	for topAssetRows.Next() {
		var asset model.TopAsset
		if err := topAssetRows.Scan(&asset.IP, &asset.Hostname, &asset.VulnerabilityCount, &asset.CriticalCount); err != nil {
			log.Error("扫描热门资产失败", "error", err)
			continue
		}
		topAssets = append(topAssets, asset)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "success",
		"data": model.DetectionStatsResponse{
			Period:            fmt.Sprintf("%d days", days),
			TotalTasks:        totalTasks,
			TotalDetections:   totalDetections,
			ByStatus:          byStatus,
			ByRiskLevel:       byRiskLevel,
			TopVulnerabilities: topVulnerabilities,
			TopAssets:         topAssets,
		},
	})
}

// db 是全局数据库连接，需要在包初始化时设置


