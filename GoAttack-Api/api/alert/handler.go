package alert

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

// RegisterRoutes 注册预警通知相关路由
func RegisterRoutes(r *gin.RouterGroup) {
	// 初始化数据库连接
	Init()
	
	r.GET("/alerts", GetAlertList)
	r.POST("/alerts", CreateAlert)
	r.PUT("/alerts/:id/status", UpdateAlertStatus)
	r.GET("/alerts/stats", GetAlertStats)
	r.POST("/alerts/subscribe", SubscribeAlert)
}

// GetAlertList 获取预警列表
func GetAlertList(c *gin.Context) {
	var query model.AlertQuery
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

	if query.Severity != "" {
		whereClause += fmt.Sprintf(" AND severity = $%d", argIndex)
		args = append(args, query.Severity)
		argIndex++
	}

	if query.Status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, query.Status)
		argIndex++
	}

	if query.StartDate != "" {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, query.StartDate)
		argIndex++
	}

	if query.EndDate != "" {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, query.EndDate)
		argIndex++
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM alert %s", whereClause)
	var total int
	err := db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		log.Error("查询预警总数失败", "error", err)
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
		SELECT id, title, description, severity, type, source, status, 
		       data, created_at, read_at, resolved_at
		FROM alert 
		%s 
		ORDER BY 
			CASE severity 
				WHEN 'critical' THEN 1
				WHEN 'high' THEN 2
				WHEN 'medium' THEN 3
				WHEN 'low' THEN 4
				ELSE 5
			END,
			created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, len(args)+1, len(args)+2)

	args = append(args, query.PageSize, offset)
	rows, err := db.Query(dataQuery, args...)
	if err != nil {
		log.Error("查询预警列表失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}
	defer rows.Close()

	var items []model.Alert
	for rows.Next() {
		var item model.Alert
		var dataJSON []byte
		var readAt, resolvedAt sql.NullTime

		err := rows.Scan(
			&item.ID, &item.Title, &item.Description, &item.Severity,
			&item.Type, &item.Source, &item.Status, &dataJSON,
			&item.CreatedAt, &readAt, &resolvedAt,
		)
		if err != nil {
			log.Error("扫描预警数据失败", "error", err)
			continue
		}

		// 解析JSON字段
		if dataJSON != nil {
			json.Unmarshal(dataJSON, &item.Data)
		}

		// 处理可为空的时间字段
		if readAt.Valid {
			item.ReadAt = &readAt.Time
		}
		if resolvedAt.Valid {
			item.ResolvedAt = &resolvedAt.Time
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		log.Error("遍历预警数据失败", "error", err)
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

// CreateAlert 创建预警
func CreateAlert(c *gin.Context) {
	var req model.AlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40001,
			"msg":  "请求参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 验证参数
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40002,
			"msg":  "预警标题不能为空",
			"data": nil,
		})
		return
	}

	if req.Severity == "" {
		req.Severity = "medium"
	}

	if req.Type == "" {
		req.Type = "system"
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
			"code": 40003,
			"msg":  "无效的严重程度",
			"data": nil,
		})
		return
	}

	// 验证类型
	validTypes := []string{"vulnerability", "system", "manual", "other"}
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
			"msg":  "无效的预警类型",
			"data": nil,
		})
		return
	}

	// 插入预警记录
	query := `
		INSERT INTO alert (
			title, description, severity, type, source, status, 
			data, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, title, status, created_at
	`

	dataJSON, _ := json.Marshal(req.Data)
	source := "manual"
	if req.Type == "vulnerability" {
		source = "detection"
	}

	var alertID int
	var title, status string
	var createdAt time.Time

	err := db.QueryRow(query,
		req.Title, req.Description, req.Severity, req.Type,
		source, "unread", dataJSON, time.Now(),
	).Scan(&alertID, &title, &status, &createdAt)

	if err != nil {
		log.Error("创建预警失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "创建预警失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 发送通知（异步）
	go func() {
		log.Info("发送预警通知", "alert_id", alertID, "title", title, "severity", req.Severity)
		// TODO: 实现通知发送逻辑
	}()

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "预警创建成功",
		"data": gin.H{
			"id":         alertID,
			"title":      title,
			"status":     status,
			"created_at": createdAt.Format(time.RFC3339),
		},
	})
}

// UpdateAlertStatus 更新预警状态
func UpdateAlertStatus(c *gin.Context) {
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
		Notes  string `json:"notes"`
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
	validStatuses := []string{"unread", "read", "resolved", "archived"}
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

	// 更新预警状态
	var query string
	var args []interface{}

	now := time.Now()
	if req.Status == "read" {
		query = `
			UPDATE alert 
			SET status = $1, read_at = $2, updated_at = $3 
			WHERE id = $4 
			RETURNING id, status, read_at, updated_at
		`
		args = []interface{}{req.Status, now, now, id}
	} else if req.Status == "resolved" {
		query = `
			UPDATE alert 
			SET status = $1, resolved_at = $2, updated_at = $3 
			WHERE id = $4 
			RETURNING id, status, resolved_at, updated_at
		`
		args = []interface{}{req.Status, now, now, id}
	} else {
		query = `
			UPDATE alert 
			SET status = $1, updated_at = $2 
			WHERE id = $3 
			RETURNING id, status, updated_at
		`
		args = []interface{}{req.Status, now, id}
	}

	var alertID int
	var status string
	var updatedAt time.Time
	var timestamp *time.Time

	if req.Status == "read" {
		var readAt time.Time
		err = db.QueryRow(query, args...).Scan(&alertID, &status, &readAt, &updatedAt)
		timestamp = &readAt
	} else if req.Status == "resolved" {
		var resolvedAt time.Time
		err = db.QueryRow(query, args...).Scan(&alertID, &status, &resolvedAt, &updatedAt)
		timestamp = &resolvedAt
	} else {
		err = db.QueryRow(query, args...).Scan(&alertID, &status, &updatedAt)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code": 40401,
				"msg":  "预警不存在",
				"data": nil,
			})
			return
		}
		log.Error("更新预警状态失败", "error", err, "id", id)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "更新失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	response := gin.H{
		"id":         alertID,
		"status":     status,
		"updated_at": updatedAt.Format(time.RFC3339),
	}

	if timestamp != nil {
		response[req.Status+"_at"] = timestamp.Format(time.RFC3339)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "预警状态更新成功",
		"data": response,
	})
}

// GetAlertStats 获取预警统计
func GetAlertStats(c *gin.Context) {
	period := c.DefaultQuery("period", "month")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// 构建时间条件
	var timeCondition string
	var timeArgs []interface{}

	if startDate != "" && endDate != "" {
		timeCondition = "WHERE created_at >= $1 AND created_at <= $2"
		timeArgs = append(timeArgs, startDate, endDate)
	} else {
		// 根据period设置默认时间范围
		now := time.Now()
		switch period {
		case "day":
			startDate = now.AddDate(0, 0, -1).Format("2006-01-02")
			timeCondition = "WHERE created_at >= $1"
			timeArgs = append(timeArgs, startDate)
		case "week":
			startDate = now.AddDate(0, 0, -7).Format("2006-01-02")
			timeCondition = "WHERE created_at >= $1"
			timeArgs = append(timeArgs, startDate)
		case "month":
			startDate = now.AddDate(0, -1, 0).Format("2006-01-02")
			timeCondition = "WHERE created_at >= $1"
			timeArgs = append(timeArgs, startDate)
		case "year":
			startDate = now.AddDate(-1, 0, 0).Format("2006-01-02")
			timeCondition = "WHERE created_at >= $1"
			timeArgs = append(timeArgs, startDate)
		default:
			// 默认查询所有
			timeCondition = ""
		}
	}

	// 查询总数
	totalQuery := "SELECT COUNT(*) FROM alert"
	if timeCondition != "" {
		totalQuery += " " + timeCondition
	}
	
	var totalCount int
	err := db.QueryRow(totalQuery, timeArgs...).Scan(&totalCount)
	if err != nil {
		log.Error("查询预警总数失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 查询未读数量
	unreadQuery := "SELECT COUNT(*) FROM alert WHERE status = 'unread'"
	if timeCondition != "" {
		unreadQuery += " AND " + timeCondition[6:] // 去掉 "WHERE "
	}
	
	var unreadCount int
	err = db.QueryRow(unreadQuery, timeArgs...).Scan(&unreadCount)
	if err != nil {
		log.Error("查询未读预警数量失败", "error", err)
		unreadCount = 0
	}

	// 按严重程度统计
	severityQuery := `
		SELECT severity, COUNT(*) as count 
		FROM alert 
		` + timeCondition + `
		GROUP BY severity 
		ORDER BY 
			CASE severity 
				WHEN 'critical' THEN 1
				WHEN 'high' THEN 2
				WHEN 'medium' THEN 3
				WHEN 'low' THEN 4
				ELSE 5
			END
	`
	severityRows, err := db.Query(severityQuery, timeArgs...)
	if err != nil {
		log.Error("按严重程度统计失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}
	defer severityRows.Close()

	bySeverity := make(map[string]int)
	for severityRows.Next() {
		var severity string
		var count int
		if err := severityRows.Scan(&severity, &count); err != nil {
			log.Error("扫描严重程度统计失败", "error", err)
			continue
		}
		bySeverity[severity] = count
	}

	// 按类型统计
	typeQuery := `
		SELECT type, COUNT(*) as count 
		FROM alert 
		` + timeCondition + `
		GROUP BY type 
		ORDER BY count DESC
	`
	typeRows, err := db.Query(typeQuery, timeArgs...)
	if err != nil {
		log.Error("按类型统计失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}
	defer typeRows.Close()

	byType := make(map[string]int)
	for typeRows.Next() {
		var alertType string
		var count int
		if err := typeRows.Scan(&alertType, &count); err != nil {
			log.Error("扫描类型统计失败", "error", err)
			continue
		}
		byType[alertType] = count
	}

	// 按状态统计
	statusQuery := `
		SELECT status, COUNT(*) as count 
		FROM alert 
		` + timeCondition + `
		GROUP BY status 
		ORDER BY 
			CASE status 
				WHEN 'unread' THEN 1
				WHEN 'read' THEN 2
				WHEN 'resolved' THEN 3
				WHEN 'archived' THEN 4
				ELSE 5
			END
	`
	statusRows, err := db.Query(statusQuery, timeArgs...)
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

	// 查询趋势数据
	trendQuery := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as total,
			COUNT(CASE WHEN severity = 'critical' THEN 1 END) as critical,
			COUNT(CASE WHEN severity = 'high' THEN 1 END) as high,
			COUNT(CASE WHEN severity = 'medium' THEN 1 END) as medium,
			COUNT(CASE WHEN severity = 'low' THEN 1 END) as low
		FROM alert
		WHERE created_at >= CURRENT_DATE - INTERVAL '30 days'
		GROUP BY DATE(created_at)
		ORDER BY date DESC
		LIMIT 30
	`
	trendRows, err := db.Query(trendQuery)
	if err != nil {
		log.Error("查询趋势数据失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}
	defer trendRows.Close()

	type AlertTrend struct {
		Date     string `json:"date"`
		Total    int    `json:"total"`
		Critical int    `json:"critical"`
		High     int    `json:"high"`
		Medium   int    `json:"medium"`
		Low      int    `json:"low"`
	}

	var trend []AlertTrend
	for trendRows.Next() {
		var item AlertTrend
		if err := trendRows.Scan(
			&item.Date, &item.Total, &item.Critical, &item.High,
			&item.Medium, &item.Low,
		); err != nil {
			log.Error("扫描趋势数据失败", "error", err)
			continue
		}
		trend = append(trend, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "success",
		"data": gin.H{
			"total_count":   totalCount,
			"unread_count":  unreadCount,
			"by_severity":   bySeverity,
			"by_type":       byType,
			"by_status":     byStatus,
			"trend":         trend,
		},
	})
}

// SubscribeAlert 订阅预警
func SubscribeAlert(c *gin.Context) {
	var req struct {
		UserID    int      `json:"user_id" binding:"required"`
		Channels  []string `json:"channels" binding:"required"`
		Severities []string `json:"severities" binding:"required"`
		Types     []string `json:"types" binding:"required"`
		Schedule  string   `json:"schedule" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40001,
			"msg":  "请求参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 验证渠道
	validChannels := []string{"email", "webhook", "dingtalk", "wechat", "sms"}
	for _, channel := range req.Channels {
		valid := false
		for _, vc := range validChannels {
			if vc == channel {
				valid = true
				break
			}
		}
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 40002,
				"msg":  "无效的通知渠道: " + channel,
				"data": nil,
			})
			return
		}
	}

	// 验证严重程度
	validSeverities := []string{"critical", "high", "medium", "low", "info"}
	for _, severity := range req.Severities {
		valid := false
		for _, vs := range validSeverities {
			if vs == severity {
				valid = true
				break
			}
		}
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 40003,
				"msg":  "无效的严重程度: " + severity,
				"data": nil,
			})
			return
		}
	}

	// 验证类型
	validTypes := []string{"vulnerability", "system", "manual", "other"}
	for _, alertType := range req.Types {
		valid := false
		for _, vt := range validTypes {
			if vt == alertType {
				valid = true
				break
			}
		}
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 40004,
				"msg":  "无效的预警类型: " + alertType,
				"data": nil,
			})
			return
		}
	}

	// 验证计划
	validSchedules := []string{"realtime", "hourly", "daily", "weekly", "monthly"}
	validSchedule := false
	for _, vs := range validSchedules {
		if vs == req.Schedule {
			validSchedule = true
			break
		}
	}
	if !validSchedule {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40005,
			"msg":  "无效的订阅计划",
			"data": nil,
		})
		return
	}

	// 生成订阅ID
	subscriptionID := fmt.Sprintf("sub_%d_%d", req.UserID, time.Now().UnixNano())

	// 这里应该将订阅信息保存到数据库
	// 为了简化，我们先返回成功响应
	log.Info("用户订阅预警", 
		"user_id", req.UserID, 
		"subscription_id", subscriptionID,
		"channels", req.Channels,
		"severities", req.Severities,
		"types", req.Types,
		"schedule", req.Schedule,
	)

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "预警订阅成功",
		"data": gin.H{
			"subscription_id": subscriptionID,
			"user_id":         req.UserID,
			"channels":        req.Channels,
			"created_at":      time.Now().Format(time.RFC3339),
		},
	})
}

// db 是全局数据库连接，需要在包初始化时设置


