package log

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLogLevelConstants(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{FATAL, "FATAL"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if name, ok := levelNames[tt.level]; !ok || name != tt.expected {
				t.Errorf("levelNames[%v] = %v, want %v", tt.level, name, tt.expected)
			}
		})
	}
}

func TestLogLevelOrder(t *testing.T) {
	// 测试日志级别顺序
	if DEBUG >= INFO {
		t.Errorf("DEBUG should be less than INFO")
	}
	if INFO >= WARN {
		t.Errorf("INFO should be less than WARN")
	}
	if WARN >= ERROR {
		t.Errorf("WARN should be less than ERROR")
	}
	if ERROR >= FATAL {
		t.Errorf("ERROR should be less than FATAL")
	}
}

func TestLoggerInitialization(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := ioutil.TempDir("", "logtest")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 测试创建Logger实例
	logger := &Logger{
		level:      INFO,
		logDir:     tempDir,
		maxSize:    1024, // 1KB for testing
		maxBackups: 2,
	}

	if logger.level != INFO {
		t.Errorf("logger.level = %v, want %v", logger.level, INFO)
	}
	if logger.logDir != tempDir {
		t.Errorf("logger.logDir = %v, want %v", logger.logDir, tempDir)
	}
	if logger.maxSize != 1024 {
		t.Errorf("logger.maxSize = %v, want %v", logger.maxSize, 1024)
	}
	if logger.maxBackups != 2 {
		t.Errorf("logger.maxBackups = %v, want %v", logger.maxBackups, 2)
	}
}

func TestLogLevelFiltering(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "logtest")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试logger
	logger := &Logger{
		level:  WARN, // 只记录WARN及以上级别的日志
		logDir: tempDir,
	}

	// 测试日志级别过滤
	testCases := []struct {
		level    LogLevel
		message  string
		shouldLog bool
	}{
		{DEBUG, "Debug message", false},
		{INFO, "Info message", false},
		{WARN, "Warning message", true},
		{ERROR, "Error message", true},
		{FATAL, "Fatal message", true},
	}

	for _, tc := range testCases {
		// 这里我们只测试级别比较逻辑
		// 实际日志记录需要文件操作，我们在其他测试中验证
		if (tc.level >= logger.level) != tc.shouldLog {
			t.Errorf("Level %v with threshold %v: expected shouldLog=%v", tc.level, logger.level, tc.shouldLog)
		}
	}
}

func TestLogFileCreation(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "logtest")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logger := &Logger{
		level:      INFO,
		logDir:     tempDir,
		maxSize:    1024 * 1024, // 1MB
		maxBackups: 3,
	}

	// 测试打开日志文件
	err = logger.openLogFile()
	if err != nil {
		t.Fatalf("打开日志文件失败: %v", err)
	}
	defer logger.file.Close()

	// 验证文件是否存在
	if logger.file == nil {
		t.Error("日志文件句柄为nil")
	}

	// 验证文件路径
	expectedFilename := fmt.Sprintf("goattack-%s.log", time.Now().Format("2006-01-02"))
	expectedPath := filepath.Join(tempDir, expectedFilename)
	
	// 获取实际文件路径
	actualPath := logger.file.Name()
	if !strings.Contains(actualPath, expectedFilename) {
		t.Errorf("日志文件路径不包含预期文件名: got %v, want contains %v", actualPath, expectedFilename)
	}
}

func TestLogRotation(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "logtest")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建小尺寸限制的logger用于测试轮转
	logger := &Logger{
		level:      INFO,
		logDir:     tempDir,
		maxSize:    100, // 非常小的限制，便于测试轮转
		maxBackups: 2,
	}

	// 初始化日志文件
	err = logger.openLogFile()
	if err != nil {
		t.Fatalf("打开日志文件失败: %v", err)
	}
	defer logger.file.Close()

	// 写入足够多的日志以触发轮转
	for i := 0; i < 10; i++ {
		logger.log(INFO, "Test log message %d for rotation testing", i)
	}

	// 检查日志文件数量
	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("读取日志目录失败: %v", err)
	}

	// 应该至少有1个日志文件
	if len(files) == 0 {
		t.Error("没有找到日志文件")
	}
}

func TestLogFormat(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "logtest")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logger := &Logger{
		level:  DEBUG, // 设置为DEBUG以记录所有级别
		logDir: tempDir,
	}

	// 初始化日志文件
	err = logger.openLogFile()
	if err != nil {
		t.Fatalf("打开日志文件失败: %v", err)
	}
	defer logger.file.Close()

	// 记录测试日志
	testMessage := "Test log message with number: %d"
	testArg := 42
	logger.log(INFO, testMessage, testArg)

	// 读取日志文件内容
	content, err := ioutil.ReadFile(logger.file.Name())
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	// 验证日志格式包含时间戳和级别
	logContent := string(content)
	if !strings.Contains(logContent, "INFO") {
		t.Error("日志内容不包含级别信息")
	}
	if !strings.Contains(logContent, fmt.Sprintf(testMessage, testArg)) {
		t.Error("日志内容不包含消息内容")
	}
}

func TestConcurrentLogging(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "logtest")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logger := &Logger{
		level:  INFO,
		logDir: tempDir,
	}

	// 初始化日志文件
	err = logger.openLogFile()
	if err != nil {
		t.Fatalf("打开日志文件失败: %v", err)
	}
	defer logger.file.Close()

	// 并发写入测试
	done := make(chan bool)
	concurrency := 10
	messagesPerGoroutine := 100

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			for j := 0; j < messagesPerGoroutine; j++ {
				logger.log(INFO, "Goroutine %d message %d", id, j)
			}
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < concurrency; i++ {
		<-done
	}

	// 验证日志文件不为空
	fileInfo, err := logger.file.Stat()
	if err != nil {
		t.Fatalf("获取文件信息失败: %v", err)
	}

	if fileInfo.Size() == 0 {
		t.Error("并发写入后日志文件为空")
	}
}

func TestLogFunctions(t *testing.T) {
	// 测试各个日志函数（不实际写入文件）
	// 这里主要测试函数调用不会panic

	// 注意：由于defaultLogger可能为nil，我们创建临时logger进行测试
	tempDir, err := ioutil.TempDir("", "logtest")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试logger
	testLogger := &Logger{
		level:  INFO,
		logDir: tempDir,
	}

	// 保存原始logger，测试后恢复
	originalLogger := defaultLogger
	defaultLogger = testLogger
	defer func() { defaultLogger = originalLogger }()

	// 初始化日志文件
	err = testLogger.openLogFile()
	if err != nil {
		t.Fatalf("打开日志文件失败: %v", err)
	}
	defer testLogger.file.Close()

	// 测试各个日志级别函数
	t.Run("Debug", func(t *testing.T) {
		Debug("Debug test message")
	})

	t.Run("Info", func(t *testing.T) {
		Info("Info test message")
	})

	t.Run("Warn", func(t *testing.T) {
		Warn("Warning test message")
	})

	t.Run("Error", func(t *testing.T) {
		Error("Error test message")
	})

	t.Run("Debugf", func(t *testing.T) {
		Debugf("Debug formatted: %s", "test")
	})

	t.Run("Infof", func(t *testing.T) {
		Infof("Info formatted: %d", 123)
	})

	t.Run("Warnf", func(t *testing.T) {
		Warnf("Warning formatted: %v", fmt.Errorf("test error"))
	})

	t.Run("Errorf", func(t *testing.T) {
		Errorf("Error formatted: %s", "test error")
	})
}

func TestCloseLogger(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "logtest")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logger := &Logger{
		level:  INFO,
		logDir: tempDir,
	}

	// 初始化日志文件
	err = logger.openLogFile()
	if err != nil {
		t.Fatalf("打开日志文件失败: %v", err)
	}

	// 写入一些日志
	logger.log(INFO, "Test message before close")

	// 关闭logger
	logger.mu.Lock()
	if logger.file != nil {
		logger.file.Close()
		logger.file = nil
	}
	logger.mu.Unlock()

	// 验证文件已关闭
	// 尝试写入应该不会panic（但可能不会实际写入）
	logger.log(INFO, "Test message after close")
}

func TestLogLevelString(t *testing.T) {
	// 测试levelNames映射
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{FATAL, "FATAL"},
		{LogLevel(999), ""}, // 不存在的级别
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := levelNames[tt.level]
			if got != tt.expected {
				t.Errorf("levelNames[%v] = %v, want %v", tt.level, got, tt.expected)
			}
		})
	}
}

func TestLogWithDifferentFormats(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "logtest")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logger := &Logger{
		level:  INFO,
		logDir: tempDir,
	}

	// 初始化日志文件
	err = logger.openLogFile()
	if err != nil {
		t.Fatalf("打开日志文件失败: %v", err)
	}
	defer logger.file.Close()

	// 测试不同的日志格式
	testCases := []struct {
		name    string
		level   LogLevel
		format  string
		args    []interface{}
	}{
		{"Simple string", INFO, "Simple message", nil},
		{"With integer", INFO, "Number: %d", []interface{}{42}},
		{"With string", INFO, "String: %s", []interface{}{"test"}},
		{"With multiple args", INFO, "Multiple: %d %s %v", []interface{}{1, "two", true}},
		{"With error", ERROR, "Error: %v", []interface{}{fmt.Errorf("test error")}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.args == nil {
				logger.log(tc.level, tc.format)
			} else {
				logger.log(tc.level, tc.format, tc.args...)
			}
		})
	}
}

func TestBufferLogging(t *testing.T) {
	// 测试缓冲区日志记录
	var buf bytes.Buffer
	
	// 创建一个临时logger，重定向输出到缓冲区
	tempDir, err := ioutil.TempDir("", "logtest")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logger := &Logger{
		level:  INFO,
		logDir: tempDir,
	}

	// 初始化日志文件
	err = logger.openLogFile()
	if err != nil {
		t.Fatalf("打开日志文件失败: %v", err)
	}
	defer logger.file.Close()

	// 记录测试消息
	testMessage := "Buffer test message"
	logger.log(INFO, testMessage)

	// 读取文件内容验证
	content, err := ioutil.ReadFile(logger.file.Name())
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	if !strings.Contains(string(content), testMessage) {
		t.Errorf("日志文件不包含测试消息: %s", testMessage)
	}
}