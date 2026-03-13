// Package main 包含测试辅助函数和工具
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestDB 创建测试数据库连接
func TestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建模拟数据库失败: %v", err)
	}

	cleanup := func() {
		db.Close()
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("未满足的模拟期望: %v", err)
		}
	}

	return db, mock, cleanup
}

// TestContext 创建测试上下文
func TestContext(t *testing.T) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// TestTempDir 创建临时目录
func TestTempDir(t *testing.T, prefix string) (string, func()) {
	tempDir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}

	cleanup := func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("清理临时目录失败: %v", err)
		}
	}

	return tempDir, cleanup
}

// TestFile 创建测试文件
func TestFile(t *testing.T, dir, filename, content string) string {
	filepath := filepath.Join(dir, filename)
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}
	return filepath
}

// AssertNoError 断言没有错误
func AssertNoError(t *testing.T, err error, message string, args ...interface{}) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", fmt.Sprintf(message, args...), err)
	}
}

// AssertError 断言有错误
func AssertError(t *testing.T, err error, message string, args ...interface{}) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: 期望错误但得到nil", fmt.Sprintf(message, args...))
	}
}

// AssertEqual 断言相等
func AssertEqual[T comparable](t *testing.T, got, want T, message string, args ...interface{}) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: 得到 %v, 期望 %v", fmt.Sprintf(message, args...), got, want)
	}
}

// AssertNotEqual 断言不相等
func AssertNotEqual[T comparable](t *testing.T, got, want T, message string, args ...interface{}) {
	t.Helper()
	if got == want {
		t.Fatalf("%s: 不期望 %v 等于 %v", fmt.Sprintf(message, args...), got, want)
	}
}

// AssertNil 断言为nil
func AssertNil(t *testing.T, value interface{}, message string, args ...interface{}) {
	t.Helper()
	if value != nil {
		t.Fatalf("%s: 期望nil但得到 %v", fmt.Sprintf(message, args...), value)
	}
}

// AssertNotNil 断言不为nil
func AssertNotNil(t *testing.T, value interface{}, message string, args ...interface{}) {
	t.Helper()
	if value == nil {
		t.Fatalf("%s: 期望非nil但得到nil", fmt.Sprintf(message, args...))
	}
}

// AssertTrue 断言为true
func AssertTrue(t *testing.T, condition bool, message string, args ...interface{}) {
	t.Helper()
	if !condition {
		t.Fatalf("%s: 条件为假", fmt.Sprintf(message, args...))
	}
}

// AssertFalse 断言为false
func AssertFalse(t *testing.T, condition bool, message string, args ...interface{}) {
	t.Helper()
	if condition {
		t.Fatalf("%s: 条件为真", fmt.Sprintf(message, args...))
	}
}

// AssertContains 断言包含
func AssertContains(t *testing.T, s, substr string, message string, args ...interface{}) {
	t.Helper()
	if !contains(s, substr) {
		t.Fatalf("%s: 字符串 %q 不包含 %q", fmt.Sprintf(message, args...), s, substr)
	}
}

// AssertNotContains 断言不包含
func AssertNotContains(t *testing.T, s, substr string, message string, args ...interface{}) {
	t.Helper()
	if contains(s, substr) {
		t.Fatalf("%s: 字符串 %q 包含 %q", fmt.Sprintf(message, args...), s, substr)
	}
}

// AssertLen 断言长度
func AssertLen[T any](t *testing.T, slice []T, length int, message string, args ...interface{}) {
	t.Helper()
	if len(slice) != length {
		t.Fatalf("%s: 切片长度 %d, 期望 %d", fmt.Sprintf(message, args...), len(slice), length)
	}
}

// AssertEmpty 断言为空
func AssertEmpty[T any](t *testing.T, slice []T, message string, args ...interface{}) {
	t.Helper()
	if len(slice) != 0 {
		t.Fatalf("%s: 切片非空, 长度 %d", fmt.Sprintf(message, args...), len(slice))
	}
}

// AssertNotEmpty 断言非空
func AssertNotEmpty[T any](t *testing.T, slice []T, message string, args ...interface{}) {
	t.Helper()
	if len(slice) == 0 {
		t.Fatalf("%s: 切片为空", fmt.Sprintf(message, args...))
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestLogger 创建测试日志器
type TestLogger struct {
	t *testing.T
}

// NewTestLogger 创建新的测试日志器
func NewTestLogger(t *testing.T) *TestLogger {
	return &TestLogger{t: t}
}

// Log 记录日志
func (l *TestLogger) Log(format string, args ...interface{}) {
	l.t.Helper()
	l.t.Logf(format, args...)
}

// Error 记录错误
func (l *TestLogger) Error(format string, args ...interface{}) {
	l.t.Helper()
	l.t.Errorf(format, args...)
}

// Fatal 记录致命错误
func (l *TestLogger) Fatal(format string, args ...interface{}) {
	l.t.Helper()
	l.t.Fatalf(format, args...)
}

// TestClock 测试时钟
type TestClock struct {
	now time.Time
}

// NewTestClock 创建新的测试时钟
func NewTestClock(now time.Time) *TestClock {
	return &TestClock{now: now}
}

// Now 返回当前时间
func (c *TestClock) Now() time.Time {
	return c.now
}

// Advance 前进时间
func (c *TestClock) Advance(d time.Duration) {
	c.now = c.now.Add(d)
}

// Set 设置时间
func (c *TestClock) Set(t time.Time) {
	c.now = t
}

// TestConfig 测试配置
type TestConfig struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	RedisHost  string
	RedisPort  int
}

// DefaultTestConfig 默认测试配置
func DefaultTestConfig() TestConfig {
	return TestConfig{
		DBHost:     "localhost",
		DBPort:     5432,
		DBUser:     "test",
		DBPassword: "test",
		DBName:     "testdb",
		RedisHost:  "localhost",
		RedisPort:  6379,
	}
}

// SetupTestEnv 设置测试环境变量
func SetupTestEnv(t *testing.T, config TestConfig) func() {
	t.Helper()
	
	// 保存原始环境变量
	originalEnv := make(map[string]string)
	envVars := map[string]string{
		"PG_USER":     config.DBUser,
		"PG_PASSWORD": config.DBPassword,
		"PG_HOST":     config.DBHost,
		"PG_PORT":     fmt.Sprintf("%d", config.DBPort),
		"PG_DB":       config.DBName,
		"REDIS_HOST":  config.RedisHost,
		"REDIS_PORT":  fmt.Sprintf("%d", config.RedisPort),
	}
	
	// 设置环境变量
	for key, value := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Setenv(key, value)
	}
	
	// 返回清理函数
	return func() {
		for key, originalValue := range originalEnv {
			if originalValue == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, originalValue)
			}
		}
	}
}

// Retry 重试函数
func Retry(t *testing.T, maxAttempts int, delay time.Duration, fn func() error) error {
	t.Helper()
	
	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			if i < maxAttempts-1 {
				time.Sleep(delay)
			}
		}
	}
	return lastErr
}

// WaitFor 等待条件成立
func WaitFor(t *testing.T, timeout time.Duration, condition func() bool) bool {
	t.Helper()
	
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// TestCleanup 测试清理注册
type TestCleanup struct {
	cleanupFuncs []func()
}

// NewTestCleanup 创建新的测试清理器
func NewTestCleanup() *TestCleanup {
	return &TestCleanup{}
}

// Add 添加清理函数
func (c *TestCleanup) Add(fn func()) {
	c.cleanupFuncs = append(c.cleanupFuncs, fn)
}

// Cleanup 执行所有清理函数
func (c *TestCleanup) Cleanup() {
	for i := len(c.cleanupFuncs) - 1; i >= 0; i-- {
		c.cleanupFuncs[i]()
	}
}

// Defer 延迟执行清理函数
func (c *TestCleanup) Defer(t *testing.T) {
	t.Helper()
	t.Cleanup(c.Cleanup)
}