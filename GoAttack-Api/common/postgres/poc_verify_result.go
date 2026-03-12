package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"GoAttack/common/log"
)

// PocVerifyResult POC验证结果
type PocVerifyResult struct {
	ID            int64                  `json:"id"`
	Target        string                 `json:"target"`
	PocID         int64                  `json:"poc_id"`
	TemplateID    string                 `json:"template_id"`
	TemplateName  string                 `json:"template_name"`
	Matched       bool                   `json:"matched"`
	Severity      string                 `json:"severity"`
	Description   string                 `json:"description"`
	Request       string                 `json:"request"`
	Response      string                 `json:"response"`
	MatchedAt     string                 `json:"matched_at"`
	ExtractedData map[string]interface{} `json:"extracted_data,omitempty"`
	Error         string                 `json:"error,omitempty"`
	VerifiedBy    string                 `json:"verified_by"`
	VerifiedAt    time.Time              `json:"verified_at"`
}

// SavePocVerifyResult 保存POC验证结果
func SavePocVerifyResult(result *PocVerifyResult) error {
	// 序列化 ExtractedData
	var extractedDataJSON []byte
	var err error
	if result.ExtractedData != nil {
		// 清理ExtractedData，确保所有值都可以序列化为JSON
		cleanedData := cleanJSONData(result.ExtractedData)
		if len(cleanedData) == 0 {
			// 如果清理后为空，设置为nil
			result.ExtractedData = nil
		} else {
			extractedDataJSON, err = json.Marshal(cleanedData)
			if err != nil {
				// 如果序列化失败，记录错误但继续执行，将ExtractedData设置为空
				log.Warn("序列化extracted_data失败: %v，将跳过此字段", err)
				result.ExtractedData = nil
			}
		}
	}

	query := `
		INSERT INTO poc_verify_result (
			target, poc_id, template_id, template_name, matched, severity, description,
			request, response, matched_at, extracted_data, error, verified_by, verified_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING id
	`

	err = DB.QueryRow(query,
		result.Target, result.PocID, result.TemplateID, result.TemplateName,
		result.Matched, result.Severity, result.Description,
		result.Request, result.Response, result.MatchedAt,
		extractedDataJSON, result.Error, result.VerifiedBy, result.VerifiedAt,
	).Scan(&result.ID)

	if err != nil {
		return fmt.Errorf("保存验证结果失败: %v", err)
	}

	return nil
}

// GetPocVerifyResultByID 根据ID获取验证结果
func GetPocVerifyResultByID(id int64) (*PocVerifyResult, error) {
	query := `
		SELECT 
			id, target, poc_id, template_id, template_name, matched, severity, description,
			request, response, matched_at, extracted_data, error, verified_by, verified_at
		FROM poc_verify_result
		WHERE id = $1
	`

	result := &PocVerifyResult{}
	var extractedDataJSON []byte

	err := DB.QueryRow(query, id).Scan(
		&result.ID, &result.Target, &result.PocID, &result.TemplateID, &result.TemplateName,
		&result.Matched, &result.Severity, &result.Description,
		&result.Request, &result.Response, &result.MatchedAt,
		&extractedDataJSON, &result.Error, &result.VerifiedBy, &result.VerifiedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询验证结果失败: %v", err)
	}

	// 解析 ExtractedData
	if len(extractedDataJSON) > 0 {
		if err := json.Unmarshal(extractedDataJSON, &result.ExtractedData); err != nil {
			return nil, fmt.Errorf("解析extracted_data失败: %v", err)
		}
	}

	return result, nil
}

// ListPocVerifyResults 获取验证结果列表（分页）
func ListPocVerifyResults(page, pageSize int, filters map[string]interface{}) ([]*PocVerifyResult, int, error) {
	// 构建查询条件
	where := "WHERE 1=1"
	args := make([]interface{}, 0)
	argIndex := 1

	if target, ok := filters["target"].(string); ok && target != "" {
		where += fmt.Sprintf(" AND target LIKE $%d", argIndex)
		args = append(args, "%"+target+"%")
		argIndex++
	}

	if pocID, ok := filters["poc_id"].(int64); ok && pocID > 0 {
		where += fmt.Sprintf(" AND poc_id = $%d", argIndex)
		args = append(args, pocID)
		argIndex++
	}

	if templateID, ok := filters["template_id"].(string); ok && templateID != "" {
		where += fmt.Sprintf(" AND template_id = $%d", argIndex)
		args = append(args, templateID)
		argIndex++
	}

	if matched, ok := filters["matched"].(bool); ok {
		where += fmt.Sprintf(" AND matched = $%d", argIndex)
		args = append(args, matched)
		argIndex++
	}

	if severity, ok := filters["severity"].(string); ok && severity != "" {
		where += fmt.Sprintf(" AND severity = $%d", argIndex)
		args = append(args, severity)
		argIndex++
	}

	// 查询总数
	var total int
	countQuery := "SELECT COUNT(*) FROM poc_verify_result " + where
	err := DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %v", err)
	}

	// 查询列表
	offset := (page - 1) * pageSize
	query := `
		SELECT 
			id, target, poc_id, template_id, template_name, matched, severity, description,
			request, response, matched_at, extracted_data, error, verified_by, verified_at
		FROM poc_verify_result
	` + where + fmt.Sprintf(`
		ORDER BY verified_at DESC
		LIMIT $%d OFFSET $%d
	`, argIndex, argIndex+1)

	args = append(args, pageSize, offset)
	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询验证结果列表失败: %v", err)
	}
	defer rows.Close()

	results := make([]*PocVerifyResult, 0)
	for rows.Next() {
		result := &PocVerifyResult{}
		var extractedDataJSON []byte

		err := rows.Scan(
			&result.ID, &result.Target, &result.PocID, &result.TemplateID, &result.TemplateName,
			&result.Matched, &result.Severity, &result.Description,
			&result.Request, &result.Response, &result.MatchedAt,
			&extractedDataJSON, &result.Error, &result.VerifiedBy, &result.VerifiedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描结果失败: %v", err)
		}

		// 解析 ExtractedData
		if len(extractedDataJSON) > 0 {
			if err := json.Unmarshal(extractedDataJSON, &result.ExtractedData); err != nil {
				return nil, 0, fmt.Errorf("解析extracted_data失败: %v", err)
			}
		}

		results = append(results, result)
	}

	return results, total, nil
}

// DeletePocVerifyResult 删除验证结果
func DeletePocVerifyResult(id int64) error {
	query := "DELETE FROM poc_verify_result WHERE id = $1"
	_, err := DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除验证结果失败: %v", err)
	}
	return nil
}

// BatchDeletePocVerifyResults 批量删除验证结果
func BatchDeletePocVerifyResults(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := "DELETE FROM poc_verify_result WHERE id IN (" + strings.Join(placeholders, ",") + ")"

	_, err := DB.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("批量删除验证结果失败: %v", err)
	}
	return nil
}

// UpdatePocVerifyResult 更新POC验证结果
func UpdatePocVerifyResult(result *PocVerifyResult) error {
	// 序列化 ExtractedData
	var extractedDataJSON []byte
	var err error
	if result.ExtractedData != nil {
		// 清理ExtractedData，确保所有值都可以序列化为JSON
		cleanedData := cleanJSONData(result.ExtractedData)
		if len(cleanedData) == 0 {
			// 如果清理后为空，设置为nil
			result.ExtractedData = nil
		} else {
			extractedDataJSON, err = json.Marshal(cleanedData)
			if err != nil {
				// 如果序列化失败，记录错误但继续执行，将ExtractedData设置为空
				log.Warn("序列化extracted_data失败: %v，将跳过此字段", err)
				result.ExtractedData = nil
			}
		}
	}

	query := `
		UPDATE poc_verify_result SET
			target = $1, poc_id = $2, template_id = $3, template_name = $4, matched = $5, severity = $6, description = $7,
			request = $8, response = $9, matched_at = $10, extracted_data = $11, error = $12, verified_by = $13, verified_at = $14
		WHERE id = $15
	`

	_, err = DB.Exec(query,
		result.Target, result.PocID, result.TemplateID, result.TemplateName,
		result.Matched, result.Severity, result.Description,
		result.Request, result.Response, result.MatchedAt,
		extractedDataJSON, result.Error, result.VerifiedBy, result.VerifiedAt,
		result.ID,
	)

	if err != nil {
		return fmt.Errorf("更新验证结果失败: %v", err)
	}

	return nil
}

// cleanJSONData 清理数据，确保所有值都可以序列化为JSON
func cleanJSONData(data map[string]interface{}) map[string]interface{} {
	cleaned := make(map[string]interface{})
	
	for key, value := range data {
		cleanedValue := cleanJSONValue(value)
		if cleanedValue != nil {
			cleaned[key] = cleanedValue
		}
	}
	
	return cleaned
}

// cleanJSONValue 递归清理JSON值
func cleanJSONValue(value interface{}) interface{} {
	switch v := value.(type) {
	case nil:
		return nil
	case string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		// 基本类型可以直接返回
		return v
	case []byte:
		// 字节数组转换为字符串
		return string(v)
	case []interface{}:
		// 递归清理数组
		cleanedArray := make([]interface{}, 0, len(v))
		for _, item := range v {
			cleanedItem := cleanJSONValue(item)
			if cleanedItem != nil {
				cleanedArray = append(cleanedArray, cleanedItem)
			}
		}
		if len(cleanedArray) == 0 {
			return nil
		}
		return cleanedArray
	case map[string]interface{}:
		// 递归清理map
		cleanedMap := make(map[string]interface{})
		for k, item := range v {
			cleanedItem := cleanJSONValue(item)
			if cleanedItem != nil {
				cleanedMap[k] = cleanedItem
			}
		}
		if len(cleanedMap) == 0 {
			return nil
		}
		return cleanedMap
	case map[interface{}]interface{}:
		// 处理interface{}键的map，转换为string键
		cleanedMap := make(map[string]interface{})
		for k, item := range v {
			keyStr := fmt.Sprintf("%v", k)
			cleanedItem := cleanJSONValue(item)
			if cleanedItem != nil {
				cleanedMap[keyStr] = cleanedItem
			}
		}
		if len(cleanedMap) == 0 {
			return nil
		}
		return cleanedMap
	default:
		// 其他类型转换为字符串
		return fmt.Sprintf("%v", v)
	}
}
