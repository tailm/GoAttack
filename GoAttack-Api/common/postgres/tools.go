package postgres

import (
	"database/sql"
	"time"
)

// ToolConfig 空间测绘工具配置
type ToolConfig struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	APIKey    string    `json:"api_key"`
	APIEmail  string    `json:"api_email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetToolConfig 获取指定工具配置
func GetToolConfig(name string) (*ToolConfig, error) {
	cfg := &ToolConfig{}
	query := `
	SELECT id, name, api_key, api_email, created_at, updated_at
	FROM tools WHERE name = $1 LIMIT 1`
	err := DB.QueryRow(query, name).Scan(
		&cfg.ID, &cfg.Name, &cfg.APIKey, &cfg.APIEmail, &cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return &ToolConfig{Name: name}, nil
	}
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// GetAllToolConfigs 获取所有工具配置
func GetAllToolConfigs() ([]ToolConfig, error) {
	query := `
	SELECT id, name, api_key, api_email, created_at, updated_at
	FROM tools`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	configs := make([]ToolConfig, 0)
	for rows.Next() {
		var cfg ToolConfig
		if err := rows.Scan(&cfg.ID, &cfg.Name, &cfg.APIKey, &cfg.APIEmail, &cfg.CreatedAt, &cfg.UpdatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}
	return configs, rows.Err()
}

// UpsertToolConfig 插入或更新工具配置
func UpsertToolConfig(name, apiKey, apiEmail string) error {
	query := `
	INSERT INTO tools (name, api_key, api_email)
	VALUES ($1, $2, $3)
	ON CONFLICT (name) DO UPDATE SET
		api_key = EXCLUDED.api_key,
		api_email = EXCLUDED.api_email,
		updated_at = CURRENT_TIMESTAMP`
	_, err := DB.Exec(query, name, apiKey, apiEmail)
	return err
}
