package dashboard

import (
	"GoAttack/common/log"
	"GoAttack/common/postgres"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册仪表盘相关路由
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/dashboard/overview", GetOverview)          // 总览统计 (资产数、漏洞数、任务数、指纹数)
	r.GET("/dashboard/vuln-trend", GetVulnTrend)       // 漏洞趋势 (近7天)
	r.GET("/dashboard/vuln-severity", GetVulnSeverity) // 漏洞级别分布
	r.GET("/dashboard/latest-vulns", GetLatestVulns)   // 最新漏洞列表
	r.GET("/dashboard/recent-tasks", GetRecentTasks)   // 最近任务扫描状态
	r.GET("/dashboard/risk-alerts", GetRiskAlerts)     // 风险提醒 (高危漏洞)
}

// refreshDashboardStats 实时重新计算并更新 dashboard 表
// 每次访问仪表盘时调用，确保数据最新
func refreshDashboardStats(username string) error {
	var assetCount, vulnCount, taskCount, fingerprintCount int
	var critical, high, medium, low, info int

	postgres.DB.QueryRow("SELECT COUNT(*) FROM asset").Scan(&assetCount)
	postgres.DB.QueryRow("SELECT COUNT(*) FROM web_fingerprint").Scan(&fingerprintCount)

	// 漏洞各级别统计
	postgres.DB.QueryRow(`
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN severity='critical' THEN 1 ELSE 0 END),
			SUM(CASE WHEN severity='high'     THEN 1 ELSE 0 END),
			SUM(CASE WHEN severity='medium'   THEN 1 ELSE 0 END),
			SUM(CASE WHEN severity='low'      THEN 1 ELSE 0 END),
			SUM(CASE WHEN severity='info'     THEN 1 ELSE 0 END)
		FROM vulnerability`,
	).Scan(&vulnCount, &critical, &high, &medium, &low, &info)

	// 任务统计（按用户隔离）
	if username != "" {
		postgres.DB.QueryRow("SELECT COUNT(*) FROM task WHERE creator = $1", username).Scan(&taskCount)
	} else {
		postgres.DB.QueryRow("SELECT COUNT(*) FROM task").Scan(&taskCount)
	}

	// UPSERT 写入统计表（id=1 单行）- PostgreSQL 使用 INSERT ... ON CONFLICT
	_, err := postgres.DB.Exec(`
		INSERT INTO dashboard
			(id, total_assets, total_vulnerabilities, total_tasks, total_fingerprints,
			 critical_vulns, high_vulns, medium_vulns, low_vulns, info_vulns, updated_at)
		VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (id) DO UPDATE SET
			total_assets          = EXCLUDED.total_assets,
			total_vulnerabilities = EXCLUDED.total_vulnerabilities,
			total_tasks           = EXCLUDED.total_tasks,
			total_fingerprints    = EXCLUDED.total_fingerprints,
			critical_vulns        = EXCLUDED.critical_vulns,
			high_vulns            = EXCLUDED.high_vulns,
			medium_vulns          = EXCLUDED.medium_vulns,
			low_vulns             = EXCLUDED.low_vulns,
			info_vulns            = EXCLUDED.info_vulns,
			updated_at            = NOW()`,
		assetCount, vulnCount, taskCount, fingerprintCount,
		critical, high, medium, low, info,
	)
	return err
}

// GetOverview 获取仪表盘总览统计（先刷新数据库，再读取）
func GetOverview(c *gin.Context) {
	username, _ := c.Get("username")
	user := ""
	if username != nil {
		user = username.(string)
	}

	// 每次访问先刷新统计
	if err := refreshDashboardStats(user); err != nil {
		log.Info("[Dashboard] refreshDashboardStats failed: %v", err)
	}

	// 从统计表读取
	var assets, vulns, tasks, fingerprints int
	postgres.DB.QueryRow(`
		SELECT total_assets, total_vulnerabilities, total_tasks, total_fingerprints
		FROM dashboard WHERE id = 1`,
	).Scan(&assets, &vulns, &tasks, &fingerprints)

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "success",
		"data": gin.H{
			"assets":          assets,
			"vulnerabilities": vulns,
			"tasks":           tasks,
			"fingerprints":    fingerprints,
		},
	})
}

// GetVulnTrend 获取近7天漏洞趋势（实时计算，不走缓存表）
func GetVulnTrend(c *gin.Context) {
	type DayStats struct {
		Date     string `json:"date"`
		Total    int    `json:"total"`
		Critical int    `json:"critical"`
		High     int    `json:"high"`
		Medium   int    `json:"medium"`
		Low      int    `json:"low"`
	}

	result := make([]DayStats, 7)
	now := time.Now()

	for i := 6; i >= 0; i-- {
		d := now.AddDate(0, 0, -i)
		dateStr := d.Format("01-02")
		startOfDay := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)

		var total, critical, high, medium, low int
		postgres.DB.QueryRow(`
			SELECT
				COUNT(*) as total,
				SUM(CASE WHEN severity='critical' THEN 1 ELSE 0 END),
				SUM(CASE WHEN severity='high'     THEN 1 ELSE 0 END),
				SUM(CASE WHEN severity='medium'   THEN 1 ELSE 0 END),
				SUM(CASE WHEN severity='low'      THEN 1 ELSE 0 END)
			FROM vulnerability
			WHERE discovered_at >= $1 AND discovered_at < $2`,
			startOfDay, endOfDay,
		).Scan(&total, &critical, &high, &medium, &low)

		result[6-i] = DayStats{
			Date:     dateStr,
			Total:    total,
			Critical: critical,
			High:     high,
			Medium:   medium,
			Low:      low,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "success",
		"data": result,
	})
}

// GetVulnSeverity 获取全局漏洞级别分布（从已刷新的统计表读取）
func GetVulnSeverity(c *gin.Context) {
	// 自动刷新统计表
	username, _ := c.Get("username")
	user := ""
	if username != nil {
		user = username.(string)
	}
	_ = refreshDashboardStats(user)

	var total, critical, high, medium, low, info int
	postgres.DB.QueryRow(`
		SELECT
			total_vulnerabilities,
			critical_vulns, high_vulns, medium_vulns, low_vulns, info_vulns
		FROM dashboard WHERE id = 1`,
	).Scan(&total, &critical, &high, &medium, &low, &info)

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "success",
		"data": gin.H{
			"total":    total,
			"critical": critical,
			"high":     high,
			"medium":   medium,
			"low":      low,
			"info":     info,
		},
	})
}

// GetLatestVulns 获取最新添加漏洞（按时间降序）
func GetLatestVulns(c *gin.Context) {
	rows, err := postgres.DB.Query(`
		SELECT id, name, severity, COALESCE(cve,''), target, discovered_at
		FROM vulnerability
		ORDER BY discovered_at DESC
		LIMIT 10`)
	if err != nil {
		log.Info("[Dashboard] 获取最新漏洞失败: %v", err)
		c.JSON(http.StatusOK, gin.H{"code": 20000, "msg": "success", "data": []interface{}{}})
		return
	}
	defer rows.Close()

	type VulnItem struct {
		ID           int64     `json:"id"`
		Name         string    `json:"name"`
		Severity     string    `json:"severity"`
		CVE          string    `json:"cve"`
		Target       string    `json:"target"`
		DiscoveredAt time.Time `json:"discovered_at"`
	}

	list := make([]VulnItem, 0)
	for rows.Next() {
		var v VulnItem
		if err := rows.Scan(&v.ID, &v.Name, &v.Severity, &v.CVE, &v.Target, &v.DiscoveredAt); err != nil {
			continue
		}
		list = append(list, v)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "success",
		"data": list,
	})
}

// GetRecentTasks 获取最近3-4个任务的扫描状态
func GetRecentTasks(c *gin.Context) {
	username, _ := c.Get("username")
	user := ""
	if username != nil {
		user = username.(string)
	}

	var rows *sql.Rows
	var err error

	if user != "" {
		rows, err = postgres.DB.Query(`
			SELECT id, name, status, progress, type, created_at
			FROM task
			WHERE creator = $1
			ORDER BY created_at DESC
			LIMIT 4`, user)
	} else {
		rows, err = postgres.DB.Query(`
			SELECT id, name, status, progress, type, created_at
			FROM task
			ORDER BY created_at DESC
			LIMIT 4`)
	}

	if err != nil {
		log.Info("[Dashboard] 获取最近任务失败: %v", err)
		c.JSON(http.StatusOK, gin.H{"code": 20000, "msg": "success", "data": []interface{}{}})
		return
	}
	defer rows.Close()

	type TaskItem struct {
		ID        int64     `json:"id"`
		Name      string    `json:"name"`
		Status    string    `json:"status"`
		Progress  int       `json:"progress"`
		Type      string    `json:"type"`
		CreatedAt time.Time `json:"created_at"`
	}

	list := make([]TaskItem, 0)
	for rows.Next() {
		var t TaskItem
		if err := rows.Scan(&t.ID, &t.Name, &t.Status, &t.Progress, &t.Type, &t.CreatedAt); err != nil {
			continue
		}
		list = append(list, t)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "success",
		"data": list,
	})
}

// GetRiskAlerts 获取风险提醒（高危漏洞，按级别优先）
func GetRiskAlerts(c *gin.Context) {
	rows, err := postgres.DB.Query(`
		SELECT id, name, severity, target, discovered_at
		FROM vulnerability
		WHERE severity IN ('critical', 'high', 'medium', 'low')
		ORDER BY
			CASE severity
				WHEN 'critical' THEN 1
				WHEN 'high' THEN 2
				WHEN 'medium' THEN 3
				WHEN 'low' THEN 4
			END,
			discovered_at DESC
		LIMIT 8`)
	if err != nil {
		log.Info("[Dashboard] 获取风险提醒失败: %v", err)
		c.JSON(http.StatusOK, gin.H{"code": 20000, "msg": "success", "data": []interface{}{}})
		return
	}
	defer rows.Close()

	type RiskItem struct {
		ID           int64     `json:"id"`
		Name         string    `json:"name"`
		Severity     string    `json:"severity"`
		Target       string    `json:"target"`
		DiscoveredAt time.Time `json:"discovered_at"`
	}

	list := make([]RiskItem, 0)
	for rows.Next() {
		var r RiskItem
		if err := rows.Scan(&r.ID, &r.Name, &r.Severity, &r.Target, &r.DiscoveredAt); err != nil {
			continue
		}
		list = append(list, r)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 20000,
		"msg":  "success",
		"data": list,
	})
}
