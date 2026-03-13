package intelligence

import (
	"GoAttack/common/log"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// Collector 漏洞情报收集器
type Collector struct {
	client *resty.Client
	db     *sql.DB
}

// NewCollector 创建新的收集器
func NewCollector(db *sql.DB) *Collector {
	client := resty.New().
		SetTimeout(30 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(2 * time.Second).
		SetRetryMaxWaitTime(10 * time.Second).
		SetHeader("User-Agent", "GoAttack-Vulnerability-Intelligence-Collector/1.0")

	return &Collector{
		client: client,
		db:     db,
	}
}

// SourceConfig 情报源配置
type SourceConfig struct {
	ID           int                    `json:"id"`
	Name         string                 `json:"name"`
	URL          string                 `json:"url"`
	Type         string                 `json:"type"` // avd, nvd, cnnvd, cnvd, exploitdb, securityfocus, custom
	Enabled      bool                   `json:"enabled"`
	SyncInterval int                    `json:"sync_interval"` // 同步间隔（秒）
	Config       map[string]interface{} `json:"config"`
	LastSync     *time.Time             `json:"last_sync"`
	LastSyncStatus string               `json:"last_sync_status"` // success, failed, pending
}

// VulnerabilityData 漏洞数据
type VulnerabilityData struct {
	ID               string                 `json:"id"`
	Source           string                 `json:"source"`
	SourceID         string                 `json:"source_id"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	Severity         string                 `json:"severity"` // critical, high, medium, low
	CVEID            string                 `json:"cve_id"`
	CVSSScore        float64                `json:"cvss_score"`
	CVSSVector       string                 `json:"cvss_vector"`
	AffectedProducts []string               `json:"affected_products"`
	References       []string               `json:"references"`
	PublishedDate    *time.Time             `json:"published_date"`
	LastModifiedDate *time.Time             `json:"last_modified_date"`
	RawData          map[string]interface{} `json:"raw_data"`
}

// CollectFromSource 从指定源收集漏洞情报
func (c *Collector) CollectFromSource(ctx context.Context, source SourceConfig) ([]VulnerabilityData, error) {
	log.Infof("开始从源 %s (%s) 收集漏洞情报", source.Name, source.Type)

	switch source.Type {
	case "avd":
		return c.collectFromAVD(ctx, source)
	case "nvd":
		return c.collectFromNVD(ctx, source)
	case "cnnvd":
		return c.collectFromCNNVD(ctx, source)
	case "cnvd":
		return c.collectFromCNVD(ctx, source)
	case "exploitdb":
		return c.collectFromExploitDB(ctx, source)
	case "securityfocus":
		return c.collectFromSecurityFocus(ctx, source)
	case "custom":
		return c.collectFromCustom(ctx, source)
	default:
		return nil, fmt.Errorf("不支持的情报源类型: %s", source.Type)
	}
}

// collectFromAVD 从阿里云漏洞库收集
func (c *Collector) collectFromAVD(ctx context.Context, source SourceConfig) ([]VulnerabilityData, error) {
	log.Infof("从阿里云漏洞库收集漏洞情报: %s", source.URL)

	// 这里实现AVD API调用逻辑
	// 由于AVD API可能需要认证，这里先返回模拟数据
	return c.mockAVDData(), nil
}

// collectFromNVD 从NVD收集
func (c *Collector) collectFromNVD(ctx context.Context, source SourceConfig) ([]VulnerabilityData, error) {
	log.Infof("从NVD收集漏洞情报: %s", source.URL)

	// NVD API调用
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		Get(source.URL)

	if err != nil {
		return nil, fmt.Errorf("NVD API请求失败: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("NVD API返回错误状态码: %d", resp.StatusCode())
	}

	var nvdResponse map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &nvdResponse); err != nil {
		return nil, fmt.Errorf("解析NVD响应失败: %v", err)
	}

	return c.parseNVDResponse(nvdResponse), nil
}

// collectFromCNNVD 从CNNVD收集
func (c *Collector) collectFromCNNVD(ctx context.Context, source SourceConfig) ([]VulnerabilityData, error) {
	log.Infof("从CNNVD收集漏洞情报: %s", source.URL)
	// 实现CNNVD API调用
	return c.mockCNNVDData(), nil
}

// collectFromCNVD 从CNVD收集
func (c *Collector) collectFromCNVD(ctx context.Context, source SourceConfig) ([]VulnerabilityData, error) {
	log.Infof("从CNVD收集漏洞情报: %s", source.URL)
	// 实现CNVD API调用
	return c.mockCNVDData(), nil
}

// collectFromExploitDB 从ExploitDB收集
func (c *Collector) collectFromExploitDB(ctx context.Context, source SourceConfig) ([]VulnerabilityData, error) {
	log.Infof("从ExploitDB收集漏洞情报: %s", source.URL)
	// 实现ExploitDB API调用
	return c.mockExploitDBData(), nil
}

// collectFromSecurityFocus 从SecurityFocus收集
func (c *Collector) collectFromSecurityFocus(ctx context.Context, source SourceConfig) ([]VulnerabilityData, error) {
	log.Infof("从SecurityFocus收集漏洞情报: %s", source.URL)
	// 实现SecurityFocus API调用
	return c.mockSecurityFocusData(), nil
}

// collectFromCustom 从自定义源收集
func (c *Collector) collectFromCustom(ctx context.Context, source SourceConfig) ([]VulnerabilityData, error) {
	log.Infof("从自定义源收集漏洞情报: %s", source.URL)

	// 支持多种自定义源格式
	method := "GET"
	if m, ok := source.Config["method"].(string); ok && m != "" {
		method = m
	}

	req := c.client.R().SetContext(ctx)
	
	// 设置请求头
	if headers, ok := source.Config["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if s, ok := v.(string); ok {
				req.SetHeader(k, s)
			}
		}
	}

	// 设置请求体
	if body, ok := source.Config["body"].(map[string]interface{}); ok {
		req.SetBody(body)
	}

	// 执行请求
	var resp *resty.Response
	var err error

	switch strings.ToUpper(method) {
	case "GET":
		resp, err = req.Get(source.URL)
	case "POST":
		resp, err = req.Post(source.URL)
	case "PUT":
		resp, err = req.Put(source.URL)
	case "DELETE":
		resp, err = req.Delete(source.URL)
	default:
		return nil, fmt.Errorf("不支持的HTTP方法: %s", method)
	}

	if err != nil {
		return nil, fmt.Errorf("自定义源请求失败: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("自定义源返回错误状态码: %d", resp.StatusCode())
	}

	// 解析响应
	var data interface{}
	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		return nil, fmt.Errorf("解析自定义源响应失败: %v", err)
	}

	// 调用自定义解析器
	return c.parseCustomResponse(data, source.Config)
}

// parseNVDResponse 解析NVD API响应
func (c *Collector) parseNVDResponse(response map[string]interface{}) []VulnerabilityData {
	var vulnerabilities []VulnerabilityData

	// 解析NVD JSON结构
	if vulnerabilitiesData, ok := response["vulnerabilities"].([]interface{}); ok {
		for _, vulnItem := range vulnerabilitiesData {
			if vulnMap, ok := vulnItem.(map[string]interface{}); ok {
				if cveData, ok := vulnMap["cve"].(map[string]interface{}); ok {
					vuln := VulnerabilityData{
						Source:   "nvd",
						SourceID: getString(cveData, "id"),
						Title:    getString(cveData, "id"),
						RawData:  vulnMap,
					}

					// 解析描述
					if descriptions, ok := cveData["descriptions"].([]interface{}); ok && len(descriptions) > 0 {
						if desc, ok := descriptions[0].(map[string]interface{}); ok {
							vuln.Description = getString(desc, "value")
						}
					}

					// 解析严重程度
					if metrics, ok := cveData["metrics"].(map[string]interface{}); ok {
						if cvssV3, ok := metrics["cvssMetricV31"].([]interface{}); ok && len(cvssV3) > 0 {
							if metric, ok := cvssV3[0].(map[string]interface{}); ok {
								if cvssData, ok := metric["cvssData"].(map[string]interface{}); ok {
									vuln.CVSSScore = getFloat64(cvssData, "baseScore")
									vuln.CVSSVector = getString(cvssData, "vectorString")
									vuln.Severity = getSeverityFromCVSS(vuln.CVSSScore)
								}
							}
						} else if cvssV2, ok := metrics["cvssMetricV2"].([]interface{}); ok && len(cvssV2) > 0 {
							if metric, ok := cvssV2[0].(map[string]interface{}); ok {
								if cvssData, ok := metric["cvssData"].(map[string]interface{}); ok {
									vuln.CVSSScore = getFloat64(cvssData, "baseScore")
									vuln.Severity = getSeverityFromCVSS(vuln.CVSSScore)
								}
							}
						}
					}

					// 解析CVE ID
					vuln.CVEID = vuln.SourceID

					// 解析引用
					if references, ok := cveData["references"].([]interface{}); ok {
						for _, ref := range references {
							if refMap, ok := ref.(map[string]interface{}); ok {
								if url := getString(refMap, "url"); url != "" {
									vuln.References = append(vuln.References, url)
								}
							}
						}
					}

					// 解析发布时间
					if published, ok := cveData["published"].(string); ok {
						if t, err := time.Parse(time.RFC3339, published); err == nil {
							vuln.PublishedDate = &t
						}
					}

					// 解析最后修改时间
					if lastModified, ok := cveData["lastModified"].(string); ok {
						if t, err := time.Parse(time.RFC3339, lastModified); err == nil {
							vuln.LastModifiedDate = &t
						}
					}

					vulnerabilities = append(vulnerabilities, vuln)
				}
			}
		}
	}

	return vulnerabilities
}

// parseCustomResponse 解析自定义响应
func (c *Collector) parseCustomResponse(data interface{}, config map[string]interface{}) ([]VulnerabilityData, error) {
	// 这里可以根据配置中的解析规则来解析数据
	// 暂时返回空数组
	return []VulnerabilityData{}, nil
}

// SaveToDatabase 保存漏洞数据到数据库
func (c *Collector) SaveToDatabase(ctx context.Context, vulnerabilities []VulnerabilityData) error {
	if len(vulnerabilities) == 0 {
		return nil
	}

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	for _, vuln := range vulnerabilities {
		// 检查是否已存在
		var count int
		err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM vulnerability_intelligence 
			WHERE source = $1 AND source_id = $2
		`, vuln.Source, vuln.SourceID).Scan(&count)

		if err != nil {
			log.Errorf("检查漏洞是否存在失败: %v", err)
			continue
		}

		rawDataJSON, _ := json.Marshal(vuln.RawData)
		affectedProductsJSON, _ := json.Marshal(vuln.AffectedProducts)
		referencesJSON, _ := json.Marshal(vuln.References)

		if count > 0 {
			// 更新现有记录
			_, err = tx.ExecContext(ctx, `
				UPDATE vulnerability_intelligence SET
					title = $1,
					description = $2,
					severity = $3,
					cve_id = $4,
					cvss_score = $5,
					cvss_vector = $6,
					affected_products = $7,
					references = $8,
					published_date = $9,
					last_modified_date = $10,
					raw_data = $11,
					updated_at = CURRENT_TIMESTAMP
				WHERE source = $12 AND source_id = $13
			`,
				vuln.Title,
				vuln.Description,
				vuln.Severity,
				vuln.CVEID,
				vuln.CVSSScore,
				vuln.CVSSVector,
				string(affectedProductsJSON),
				string(referencesJSON),
				vuln.PublishedDate,
				vuln.LastModifiedDate,
				string(rawDataJSON),
				vuln.Source,
				vuln.SourceID,
			)
		} else {
			// 插入新记录
			_, err = tx.ExecContext(ctx, `
				INSERT INTO vulnerability_intelligence (
					source, source_id, title, description, severity, cve_id,
					cvss_score, cvss_vector, affected_products, references,
					published_date, last_modified_date, raw_data
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			`,
				vuln.Source,
				vuln.SourceID,
				vuln.Title,
				vuln.Description,
				vuln.Severity,
				vuln.CVEID,
				vuln.CVSSScore,
				vuln.CVSSVector,
				string(affectedProductsJSON),
				string(referencesJSON),
				vuln.PublishedDate,
				vuln.LastModifiedDate,
				string(rawDataJSON),
			)
		}

		if err != nil {
			log.Errorf("保存漏洞数据失败: %v", err)
			continue
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	log.Infof("成功保存 %d 条漏洞数据到数据库", len(vulnerabilities))
	return nil
}

// UpdateSourceSyncStatus 更新情报源同步状态
func (c *Collector) UpdateSourceSyncStatus(ctx context.Context, sourceID int, status string, message string) error {
	now := time.Now()
	_, err := c.db.ExecContext(ctx, `
		UPDATE intelligence_source SET
			last_sync = $1,
			last_sync_status = $2,
			last_sync_message = $3,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`, now, status, message, sourceID)
	return err
}

// 辅助函数
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}

func getSeverityFromCVSS(score float64) string {
	if score >= 9.0 {
		return "critical"
	} else if score >= 7.0 {
		return "high"
	} else if score >= 4.0 {
		return "medium"
	} else if score > 0 {
		return "low"
	}
	return "unknown"
}

// 模拟数据函数（用于测试）
func (c *Collector) mockAVDData() []VulnerabilityData {
	return []VulnerabilityData{
		{
			Source:           "avd",
			SourceID:         "AVD-2024-0001",
			Title:            "Apache Log4j2 远程代码执行漏洞",
			Description:      "Apache Log4j2 存在远程代码执行漏洞，攻击者可通过构造恶意请求触发漏洞",
			Severity:         "critical",
			CVEID:            "CVE-2021-44228",
			CVSSScore:        10.0,
			CVSSVector:       "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H",
			AffectedProducts: []string{"Apache Log4j2 2.0-beta9 to 2.14.1"},
			References:       []string{"https://avd.aliyun.com/detail?id=AVD-2024-0001"},
			PublishedDate:    &[]time.Time{time.Now().Add(-24 * time.Hour)}[0],
			RawData:          map[string]interface{}{"source": "avd"},
		},
	}
}

func (c *Collector) mockCNNVDData() []VulnerabilityData {
	return []VulnerabilityData{
		{
			Source:           "cnnvd",
			SourceID:         "CNNVD-202401-001",
			Title:            "某国产操作系统权限提升漏洞",
			Description:      "某国产操作系统存在权限提升漏洞，本地攻击者可利用该漏洞提升权限",
			Severity:         "high",
			CVEID:            "CVE-2024-12345",
			CVSSScore:        7.8,
			CVSSVector:       "CVSS:3.1/AV:L/AC:L/PR:L/UI:N/S:U/C:H/I:H/A:H",
			AffectedProducts: []string{"某国产操作系统 V1.0"},
			References:       []string{"https://www.cnnvd.org.cn/web/xxk/ldxqById.tag?CNNVD=CNNVD-202401-001"},
			PublishedDate:    &[]time.Time{time.Now().Add(-48 * time.Hour)}[0],
			RawData:          map[string]interface{}{"source": "cnnvd"},
		},
	}
}

func (c *Collector) mockCNVDData() []VulnerabilityData {
	return []VulnerabilityData{
		{
			Source:           "cnvd",
			SourceID:         "CNVD-2024-00001",
			Title:            "某Web应用SQL注入漏洞",
			Description:      "某Web应用存在SQL注入漏洞，攻击者可利用该漏洞获取数据库敏感信息",
			Severity:         "high",
			CVEID:            "CVE-2024-54321",
			CVSSScore:        8.5,
			CVSSVector:       "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N",
			AffectedProducts: []string{"某Web应用 V2.0"},
			References:       []string{"https://www.cnvd.org.cn/flaw/show/CNVD-2024-00001"},
			PublishedDate:    &[]time.Time{time.Now().Add(-72 * time.Hour)}[0],
			RawData:          map[string]interface{}{"source": "cnvd"},
		},
	}
}

func (c *Collector) mockExploitDBData() []VulnerabilityData {
	return []VulnerabilityData{
		{
			Source:           "exploitdb",
			SourceID:         "EDB-12345",
			Title:            "WordPress Plugin XSS Vulnerability",
			Description:      "WordPress某插件存在跨站脚本漏洞，攻击者可注入恶意脚本",
			Severity:         "medium",
			CVEID:            "CVE-2024-11111",
			CVSSScore:        6.1,
			CVSSVector:       "CVSS:3.1/AV:N/AC:L/PR:N/UI:R/S:C/C:L/I:L/A:N",
			AffectedProducts: []string{"WordPress Plugin < 1.2.3"},
			References:       []string{"https://www.exploit-db.com/exploits/12345"},
			PublishedDate:    &[]time.Time{time.Now().Add(-96 * time.Hour)}[0],
			RawData:          map[string]interface{}{"source": "exploitdb"},
		},
	}
}

func (c *Collector) mockSecurityFocusData() []VulnerabilityData {
	return []VulnerabilityData{
		{
			Source:           "securityfocus",
			SourceID:         "BID-123456",
			Title:            "Linux Kernel Privilege Escalation",
			Description:      "Linux内核存在权限提升漏洞，本地用户可利用该漏洞获取root权限",
			Severity:         "critical",
			CVEID:            "CVE-2024-22222",
			CVSSScore:        9.8,
			CVSSVector:       "CVSS:3.1/AV:L/AC:L/PR:L/UI:N/S:C/C:H/I:H/A:H",
			AffectedProducts: []string{"Linux Kernel 5.10-5.15"},
			References:       []string{"https://www.securityfocus.com/bid/123456"},
			PublishedDate:    &[]time.Time{time.Now().Add(-120 * time.Hour)}[0],
			RawData:          map[string]interface{}{"source": "securityfocus"},
		},
	}
}