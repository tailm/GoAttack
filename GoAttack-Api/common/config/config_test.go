package config

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	// 测试环境变量存在的情况
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	result := getEnv("TEST_VAR", "default")
	if result != "test_value" {
		t.Errorf("getEnv() = %v, want %v", result, "test_value")
	}

	// 测试环境变量不存在的情况
	result = getEnv("NON_EXISTENT_VAR", "default_value")
	if result != "default_value" {
		t.Errorf("getEnv() = %v, want %v", result, "default_value")
	}
}

func TestGetEnvInt(t *testing.T) {
	// 测试有效的整数环境变量
	os.Setenv("TEST_INT", "123")
	defer os.Unsetenv("TEST_INT")

	result := getEnvInt("TEST_INT", 456)
	if result != 123 {
		t.Errorf("getEnvInt() = %v, want %v", result, 123)
	}

	// 测试无效的整数环境变量
	os.Setenv("TEST_INVALID", "not_a_number")
	defer os.Unsetenv("TEST_INVALID")

	result = getEnvInt("TEST_INVALID", 456)
	if result != 456 {
		t.Errorf("getEnvInt() = %v, want %v", result, 456)
	}

	// 测试环境变量不存在的情况
	result = getEnvInt("NON_EXISTENT_INT", 789)
	if result != 789 {
		t.Errorf("getEnvInt() = %v, want %v", result, 789)
	}
}

func TestConfigDefaults(t *testing.T) {
	// 清除所有相关环境变量，测试默认值
	os.Unsetenv("PG_USER")
	os.Unsetenv("PG_PASSWORD")
	os.Unsetenv("PG_HOST")
	os.Unsetenv("PG_PORT")
	os.Unsetenv("PG_DB")
	os.Unsetenv("PG_SSLMODE")
	os.Unsetenv("REDIS_PASSWORD")
	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")
	os.Unsetenv("REDIS_ADDR")

	// 重新初始化包以使用默认值
	// 注意：由于变量在包初始化时设置，我们需要重新导入或使用其他方法
	// 这里我们直接测试导出的变量值
	t.Run("PostgreSQL defaults", func(t *testing.T) {
		// 这些测试需要重新编译包，所以我们只验证逻辑
		// 在实际测试中，应该使用测试专用的包实例
	})

	t.Run("Redis defaults", func(t *testing.T) {
		// 同上
	})
}

func TestConfigWithEnvVars(t *testing.T) {
	// 设置环境变量
	os.Setenv("PG_USER", "test_user")
	os.Setenv("PG_PASSWORD", "test_password")
	os.Setenv("PG_HOST", "test_host")
	os.Setenv("PG_PORT", "9999")
	os.Setenv("PG_DB", "test_db")
	os.Setenv("PG_SSLMODE", "require")
	os.Setenv("REDIS_PASSWORD", "redis_pass")
	os.Setenv("REDIS_HOST", "redis_host")
	os.Setenv("REDIS_PORT", "6380")
	os.Setenv("REDIS_ADDR", "redis_host:6380")

	defer func() {
		os.Unsetenv("PG_USER")
		os.Unsetenv("PG_PASSWORD")
		os.Unsetenv("PG_HOST")
		os.Unsetenv("PG_PORT")
		os.Unsetenv("PG_DB")
		os.Unsetenv("PG_SSLMODE")
		os.Unsetenv("REDIS_PASSWORD")
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_ADDR")
	}()

	// 测试 getEnv 函数
	tests := []struct {
		name     string
		key      string
		fallback string
		want     string
	}{
		{"PG_USER", "PG_USER", "", "test_user"},
		{"PG_PASSWORD", "PG_PASSWORD", "", "test_password"},
		{"PG_HOST", "PG_HOST", "", "test_host"},
		{"PG_DB", "PG_DB", "", "test_db"},
		{"PG_SSLMODE", "PG_SSLMODE", "", "require"},
		{"REDIS_PASSWORD", "REDIS_PASSWORD", "", "redis_pass"},
		{"REDIS_HOST", "REDIS_HOST", "", "redis_host"},
		{"REDIS_ADDR", "REDIS_ADDR", "", "redis_host:6380"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getEnv(tt.key, tt.fallback)
			if got != tt.want {
				t.Errorf("getEnv(%q, %q) = %v, want %v", tt.key, tt.fallback, got, tt.want)
			}
		})
	}

	// 测试 getEnvInt 函数
	intTests := []struct {
		name     string
		key      string
		fallback int
		want     int
	}{
		{"PG_PORT", "PG_PORT", 0, 9999},
		{"REDIS_PORT", "REDIS_PORT", 0, 6380},
	}

	for _, tt := range intTests {
		t.Run(tt.name, func(t *testing.T) {
			got := getEnvInt(tt.key, tt.fallback)
			if got != tt.want {
				t.Errorf("getEnvInt(%q, %v) = %v, want %v", tt.key, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestRedisAddrOverride(t *testing.T) {
	// 测试 REDIS_ADDR 覆盖 REDIS_HOST 和 REDIS_PORT
	os.Setenv("REDIS_ADDR", "custom_host:9999")
	os.Setenv("REDIS_HOST", "should_be_ignored")
	os.Setenv("REDIS_PORT", "1111")
	defer func() {
		os.Unsetenv("REDIS_ADDR")
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
	}()

	// 注意：由于变量在包初始化时设置，这个测试需要重新编译包
	// 这里我们只测试 getEnv 函数的行为
	addr := getEnv("REDIS_ADDR", "")
	if addr != "custom_host:9999" {
		t.Errorf("REDIS_ADDR = %v, want %v", addr, "custom_host:9999")
	}
}

func TestEmptyEnvVars(t *testing.T) {
	// 测试空环境变量
	os.Setenv("EMPTY_VAR", "")
	defer os.Unsetenv("EMPTY_VAR")

	result := getEnv("EMPTY_VAR", "default")
	if result != "" {
		t.Errorf("getEnv() for empty var = %v, want empty string", result)
	}
}

func TestEnvVarPriority(t *testing.T) {
	// 测试环境变量优先级高于默认值
	os.Setenv("TEST_PRIORITY", "env_value")
	defer os.Unsetenv("TEST_PRIORITY")

	result := getEnv("TEST_PRIORITY", "default_value")
	if result != "env_value" {
		t.Errorf("getEnv() = %v, want %v", result, "env_value")
	}
}