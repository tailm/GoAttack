package intelligence

import (
	"GoAttack/common/log"
	"GoAttack/model"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册漏洞情报相关路由
func RegisterRoutes(r *gin.RouterGroup) {
	// 初始化数据库连接
	Init()
	
	r.GET("/vulnerability/intelligence", GetVulnerabilityIntelligenceList)
	r.GET("/vulnerability/intelligence/:id", GetVulnerabilityIntelligenceDetail)
	r.POST("/vulnerability/intelligence/sync", SyncVulnerabilityIntelligence)
	r.GET("/vulnerability/intelligence/sync/:sync_id", GetSyncStatus)
	r.GET("/vulnerability/intelligence/stats", GetVulnerabilityStats)
	r.GET("/vulnerability/intelligence/search", SearchVulnerabilityIntelligence)
}



// GetVulnerabilityIntelligenceList 获取漏洞情报列表
func GetVulnerabilityIntelligenceList(c *gin.Context) {
	var query model.VulnerabilityIntelligenceQuery
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
	if query.SortBy == "" {
		query.SortBy = "published_at"
	}
	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}

	// 构建查询条件
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if query.Search != "" {
		whereClause += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d OR cve_id ILIKE $%d)", argIndex, argIndex, argIndex)
		args = append(args, "%"+query.Search+"%")
		argIndex++
	}

	if query.Severity != "" {
		whereClause += fmt.Sprintf(" AND severity = $%d", argIndex)
		args = append(args, query.Severity)
		argIndex++
	}

	if query.Source != "" {
		whereClause += fmt.Sprintf(" AND source = $%d", argIndex)
		args = append(args, query.Source)
		argIndex++
	}

	if query.CVEID != "" {
		whereClause += fmt.Sprintf(" AND cve_id = $%d", argIndex)
		args = append(args, query.CVEID)
		argIndex++
	}

	if query.StartDate != "" {
		whereClause += fmt.Sprintf(" AND published_at >= $%d", argIndex)
		args = append(args, query.StartDate)
		argIndex++
	}

	if query.EndDate != "" {
		whereClause += fmt.Sprintf(" AND published_at <= $%d", argIndex)
		args = append(args, query.EndDate)
		argIndex++
	}

	if query.Is0Day != nil {
		whereClause += fmt.Sprintf(" AND is_0day = $%d", argIndex)
		args = append(args, *query.Is0Day)
		argIndex++
	}

	if query.ExploitAvailable != nil {
		whereClause += fmt.Sprintf(" AND exploit_available = $%d", argIndex)
		args = append(args, *query.ExploitAvailable)
		argIndex++
	}

	// 构建排序
	orderClause := fmt.Sprintf("ORDER BY %s %s", query.SortBy, query.SortOrder)

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM vulnerability_intelligence %s", whereClause)
	var total int
	err := db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		log.Error("查询漏洞情报总数失败", "error", err)
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
		SELECT id, cve_id, cnvd_id, cnnvd_id, title, description, severity, 
		       cvss_score, cvss_vector, affected_products, affected_versions, 
		       vuln_references, exploit_available, poc_available, published_at, 
		       last_modified, source, is_0day, created_at, updated_at
		FROM vulnerability_intelligence 
		%s 
		%s 
		LIMIT $%d OFFSET $%d`,
		whereClause, orderClause, argIndex, argIndex+1)

	args = append(args, query.PageSize, offset)
	rows, err := db.Query(dataQuery, args...)
	if err != nil {
		log.Error("查询漏洞情报列表失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}
	defer rows.Close()

	var items []model.VulnerabilityIntelligence
	for rows.Next() {
		var item model.VulnerabilityIntelligence
		err := rows.Scan(
			&item.ID, &item.CVEID, &item.CNVDID, &item.CNNVDID, &item.Title,
			&item.Description, &item.Severity, &item.CVSSScore, &item.CVSSVector,
			&item.AffectedProducts, &item.AffectedVersions, &item.References,
			&item.ExploitAvailable, &item.POCAvailable, &item.PublishedAt,
			&item.LastModified, &item.Source, &item.Is0Day, &item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			log.Error("扫描漏洞情报数据失败", "error", err)
			continue
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		log.Error("遍历漏洞情报数据失败", "error", err)
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

// GetVulnerabilityIntelligenceDetail 获取漏洞情报详情
func GetVulnerabilityIntelligenceDetail(c *gin.Context) {
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
		SELECT id, cve_id, cnvd_id, cnnvd_id, title, description, severity, 
		       cvss_score, cvss_vector, affected_products, affected_versions, 
		       references, exploit_available, poc_available, published_at, 
		       last_modified, source, is_0day, created_at, updated_at
		FROM vulnerability_intelligence 
		WHERE id = $1
	`

	var item model.VulnerabilityIntelligence
	err = db.QueryRow(query, id).Scan(
		&item.ID, &item.CVEID, &item.CNVDID, &item.CNNVDID, &item.Title,
		&item.Description, &item.Severity, &item.CVSSScore, &item.CVSSVector,
		&item.AffectedProducts, &item.AffectedVersions, &item.References,
		&item.ExploitAvailable, &item.POCAvailable, &item.PublishedAt,
		&item.LastModified, &item.Source, &item.Is0Day, &item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"code": 40401,
				"msg":  "漏洞情报不存在",
				"data": nil,
			})
			return
		}
		log.Error("查询漏洞情报详情失败", "error", err, "id", id)
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
		"data": item,
	})
}

// SyncVulnerabilityIntelligence 手动同步漏洞情报
func SyncVulnerabilityIntelligence(c *gin.Context) {
	var req model.SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40001,
			"msg":  "请求参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 生成同步任务ID
	syncID := fmt.Sprintf("sync_%d", time.Now().UnixNano())

	// 这里应该启动一个后台任务来执行同步
	// 为了简化，我们先返回成功响应
	go func() {
		// 在实际实现中，这里应该调用同步服务
		log.Info("开始同步漏洞情报", "sync_id", syncID, "source", req.Source, "force", req.Force)
		// TODO: 实现实际的同步逻辑
	}()

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "同步任务已启动",
		"data": gin.H{
			"sync_id":        syncID,
			"status":         "pending",
			"estimated_time": 300, // 预计5分钟
		},
	})
}

// GetSyncStatus 获取同步状态
func GetSyncStatus(c *gin.Context) {
	syncID := c.Param("sync_id")
	if syncID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40001,
			"msg":  "同步ID不能为空",
			"data": nil,
		})
		return
	}

	// 这里应该从数据库或缓存中获取同步状态
	// 为了简化，我们返回模拟数据
	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "success",
		"data": gin.H{
			"sync_id":          syncID,
			"source":           "nvd",
			"status":           "completed",
			"progress":         100,
			"total_count":      1000,
			"processed_count":  1000,
			"new_count":        120,
			"updated_count":    530,
			"error_count":      0,
			"start_time":       time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			"estimated_end_time": time.Now().Format(time.RFC3339),
		},
	})
}

// GetVulnerabilityStats 获取漏洞统计信息
func GetVulnerabilityStats(c *gin.Context) {
	period := c.DefaultQuery("period", "month")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// 构建时间条件
	var timeCondition string
	var timeArgs []interface{}

	if startDate != "" && endDate != "" {
		timeCondition = "WHERE published_at >= $1 AND published_at <= $2"
		timeArgs = append(timeArgs, startDate, endDate)
	} else {
		// 根据period设置默认时间范围
		now := time.Now()
		switch period {
		case "day":
			startDate = now.AddDate(0, 0, -1).Format("2006-01-02")
			timeCondition = "WHERE published_at >= $1"
			timeArgs = append(timeArgs, startDate)
		case "week":
			startDate = now.AddDate(0, 0, -7).Format("2006-01-02")
			timeCondition = "WHERE published_at >= $1"
			timeArgs = append(timeArgs, startDate)
		case "month":
			startDate = now.AddDate(0, -1, 0).Format("2006-01-02")
			timeCondition = "WHERE published_at >= $1"
			timeArgs = append(timeArgs, startDate)
		case "year":
			startDate = now.AddDate(-1, 0, 0).Format("2006-01-02")
			timeCondition = "WHERE published_at >= $1"
			timeArgs = append(timeArgs, startDate)
		default:
			// 默认查询所有
			timeCondition = ""
		}
	}

	// 查询总数
	totalQuery := "SELECT COUNT(*) FROM vulnerability_intelligence"
	if timeCondition != "" {
		totalQuery += " " + timeCondition
	}
	
	var totalCount int
	err := db.QueryRow(totalQuery, timeArgs...).Scan(&totalCount)
	if err != nil {
		log.Error("查询漏洞总数失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 查询今日数量
	today := time.Now().Format("2006-01-02")
	var todayCount int
	err = db.QueryRow(
		"SELECT COUNT(*) FROM vulnerability_intelligence WHERE DATE(created_at) = $1", 
		today).Scan(&todayCount)
	if err != nil {
		log.Error("查询今日漏洞数量失败", "error", err)
		todayCount = 0
	}

	// 查询本周数量
	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday())+1).Format("2006-01-02")
	var weekCount int
	err = db.QueryRow(
		"SELECT COUNT(*) FROM vulnerability_intelligence WHERE created_at >= $1", 
		weekStart).Scan(&weekCount)
	if err != nil {
		log.Error("查询本周漏洞数量失败", "error", err)
		weekCount = 0
	}

	// 查询本月数量
	monthStart := time.Now().Format("2006-01") + "-01"
	var monthCount int
	err = db.QueryRow(
		"SELECT COUNT(*) FROM vulnerability_intelligence WHERE created_at >= $1", 
		monthStart).Scan(&monthCount)
	if err != nil {
		log.Error("查询本月漏洞数量失败", "error", err)
		monthCount = 0
	}

	// 按严重程度统计
	severityQuery := `
		SELECT severity, COUNT(*) as count 
		FROM vulnerability_intelligence 
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
	severityRows, err := db.Query(severityQuery)
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

	// 按数据源统计
	sourceQuery := `
		SELECT source, COUNT(*) as count 
		FROM vulnerability_intelligence 
		GROUP BY source 
		ORDER BY count DESC
	`
	sourceRows, err := db.Query(sourceQuery)
	if err != nil {
		log.Error("按数据源统计失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "查询失败: " + err.Error(),
			"data": nil,
		})
		return
	}
	defer sourceRows.Close()

	bySource := make(map[string]int)
	for sourceRows.Next() {
		var source string
		var count int
		if err := sourceRows.Scan(&source, &count); err != nil {
			log.Error("扫描数据源统计失败", "error", err)
			continue
		}
		bySource[source] = count
	}

	// 查询趋势数据
	trendQuery := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as total,
			COUNT(CASE WHEN severity = 'critical' THEN 1 END) as critical,
			COUNT(CASE WHEN severity = 'high' THEN 1 END) as high,
			COUNT(CASE WHEN severity = 'medium' THEN 1 END) as medium,
			COUNT(CASE WHEN severity = 'low' THEN 1 END) as low,
			COUNT(CASE WHEN is_0day THEN 1 END) as zero_day,
			COUNT(CASE WHEN exploit_available THEN 1 END) as with_exploit
		FROM vulnerability_intelligence
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

	var trend []model.VulnerabilityTrend
	for trendRows.Next() {
		var item model.VulnerabilityTrend
		if err := trendRows.Scan(
			&item.Date, &item.Total, &item.Critical, &item.High,
			&item.Medium, &item.Low, &item.ZeroDay, &item.WithExploit,
		); err != nil {
			log.Error("扫描趋势数据失败", "error", err)
			continue
		}
		trend = append(trend, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "success",
		"data": model.StatsResponse{
			TotalCount:  totalCount,
			TodayCount:  todayCount,
			WeekCount:   weekCount,
			MonthCount:  monthCount,
			BySeverity:  bySeverity,
			BySource:    bySource,
			Trend:       trend,
		},
	})
}

// SearchVulnerabilityIntelligence 搜索漏洞情报
func SearchVulnerabilityIntelligence(c *gin.Context) {
	query := c.Query("q")
	fields := c.QueryArray("fields")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 40001,
			"msg":  "搜索关键词不能为空",
			"data": nil,
		})
		return
	}

	// 设置默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 构建搜索条件
	searchFields := []string{"title", "description", "cve_id", "cnvd_id", "cnnvd_id"}
	if len(fields) > 0 {
		searchFields = fields
	}

	// 构建搜索条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	for _, field := range searchFields {
		conditions = append(conditions, fmt.Sprintf("%s ILIKE $%d", field, argIndex))
		args = append(args, "%"+query+"%")
		argIndex++
	}

	whereClause := "WHERE " + strings.Join(conditions, " OR ")

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM vulnerability_intelligence %s", whereClause)
	var total int
	err := db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		log.Error("搜索漏洞情报总数失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "搜索失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 查询数据
	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf(`
		SELECT id, cve_id, title, severity, cvss_score, source, published_at
		FROM vulnerability_intelligence 
		%s 
		ORDER BY published_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argIndex, argIndex+1)

	args = append(args, pageSize, offset)
	rows, err := db.Query(dataQuery, args...)
	if err != nil {
		log.Error("搜索漏洞情报失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "搜索失败: " + err.Error(),
			"data": nil,
		})
		return
	}
	defer rows.Close()

	type SearchResult struct {
		ID         int       `json:"id"`
		CVEID      string    `json:"cve_id"`
		Title      string    `json:"title"`
		Severity   string    `json:"severity"`
		CVSSScore  float64   `json:"cvss_score"`
		Source     string    `json:"source"`
		PublishedAt time.Time `json:"published_at"`
		Highlight  map[string][]string `json:"highlight,omitempty"`
	}

	var items []SearchResult
	for rows.Next() {
		var item SearchResult
		err := rows.Scan(
			&item.ID, &item.CVEID, &item.Title, &item.Severity,
			&item.CVSSScore, &item.Source, &item.PublishedAt,
		)
		if err != nil {
			log.Error("扫描搜索结果失败", "error", err)
			continue
		}

		// 简单的高亮处理
		item.Highlight = make(map[string][]string)
		if strings.Contains(strings.ToLower(item.Title), strings.ToLower(query)) {
			item.Highlight["title"] = []string{strings.Replace(item.Title, query, "<em>"+query+"</em>", -1)}
		}
		if strings.Contains(strings.ToLower(item.CVEID), strings.ToLower(query)) {
			item.Highlight["cve_id"] = []string{strings.Replace(item.CVEID, query, "<em>"+query+"</em>", -1)}
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		log.Error("遍历搜索结果失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 50001,
			"msg":  "搜索失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "success",
		"data": model.PaginatedResponse{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			Items:    items,
		},
	})
}

// db 是全局数据库连接，需要在包初始化时设置


