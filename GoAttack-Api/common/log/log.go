// Package log 提供 GoAttack 系统的日志记录功能。
//
// 该包实现了一个线程安全的日志系统，支持以下特性：
//   - 多级别日志记录（DEBUG、INFO、WARN、ERROR、FATAL）
//   - 日志文件轮转（按大小和备份数量）
//   - 控制台和文件双重输出
//   - 格式化日志输出
//   - 线程安全操作
//
// 日志级别：
//   - DEBUG: 调试信息，用于开发阶段
//   - INFO: 普通信息，记录系统运行状态
//   - WARN: 警告信息，表示潜在问题
//   - ERROR: 错误信息，表示操作失败但系统可继续运行
//   - FATAL: 致命错误，表示系统无法继续运行
//
// 日志文件：
//   - 默认日志目录: ./logs/
//   - 日志文件格式: goattack-YYYY-MM-DD.log
//   - 最大文件大小: 10MB
//   - 最大备份数量: 7个
//
// 使用示例：
//   log.InitLogger()
//   defer log.Close()
//   log.Info("系统启动成功")
//   log.Errorf("连接数据库失败: %v", err)
package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel 表示日志级别类型。
//
// 日志级别从低到高依次为：DEBUG < INFO < WARN < ERROR < FATAL。
// 只有大于等于当前设置级别的日志才会被记录。
type LogLevel int

const (
	// DEBUG 调试级别，记录最详细的日志信息。
	DEBUG LogLevel = iota
	
	// INFO 信息级别，记录系统运行状态信息。
	INFO
	
	// WARN 警告级别，记录潜在问题。
	WARN
	
	// ERROR 错误级别，记录操作失败信息。
	ERROR
	
	// FATAL 致命级别，记录系统无法继续运行的错误。
	FATAL
)

var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// Logger 日志器结构体，提供线程安全的日志记录功能。
//
// Logger 管理日志文件的创建、写入和轮转，支持多级别日志记录。
// 所有日志操作都是线程安全的，可以在并发环境中安全使用。
//
// 字段：
//   - mu: 互斥锁，保证线程安全
//   - level: 当前日志级别，低于此级别的日志将被忽略
//   - file: 当前日志文件句柄
//   - logDir: 日志文件存储目录
//   - maxSize: 单个日志文件最大大小（字节），默认10MB
//   - maxBackups: 保留的旧日志文件数量，默认7个
//   - currentSize: 当前日志文件大小（字节）
type Logger struct {
	mu          sync.Mutex
	level       LogLevel
	file        *os.File
	logDir      string
	maxSize     int64 // 单个日志文件最大大小（字节）
	maxBackups  int   // 保留的旧日志文件数量
	currentSize int64
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// InitLogger 初始化全局日志系统。
//
// 该函数使用单例模式确保只初始化一次日志系统，主要功能包括：
//   - 创建日志目录（默认：./logs/）
//   - 初始化默认日志器实例
//   - 设置日志级别为 INFO
//   - 配置日志文件轮转参数（最大10MB，保留10个备份）
//   - 打开当前日志文件
//
// 注意：
//   - 该函数是线程安全的，多次调用只会执行一次初始化
//   - 如果初始化失败，程序会退出并返回错误码1
//   - 建议在程序启动时尽早调用此函数
//
// 使用示例：
//   func main() {
//       log.InitLogger()
//       defer log.Close()
//       // ... 其他初始化代码
//   }
func InitLogger() {
	once.Do(func() {
		logDir := "./logs"
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "创建日志目录失败: %v\n", err)
			os.Exit(1)
		}

		defaultLogger = &Logger{
			level:      INFO,
			logDir:     logDir,
			maxSize:    10 * 1024 * 1024, // 10MB
			maxBackups: 10,
		}

		if err := defaultLogger.openLogFile(); err != nil {
			fmt.Fprintf(os.Stderr, "初始化日志文件失败: %v\n", err)
			os.Exit(1)
		}

		Info("[日志系统] 已初始化，日志目录: %s", logDir)
	})
}

// openLogFile 打开日志文件
func (l *Logger) openLogFile() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 关闭旧文件
	if l.file != nil {
		l.file.Close()
	}

	// 创建新日志文件
	filename := filepath.Join(l.logDir, fmt.Sprintf("goattack_%s.log", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}

	// 获取文件大小
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	l.file = file
	l.currentSize = stat.Size()

	return nil
}

// rotateLog 日志轮转
func (l *Logger) rotateLog() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.file.Close()
	}

	// 生成备份文件名
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	oldName := filepath.Join(l.logDir, fmt.Sprintf("goattack_%s.log", time.Now().Format("2006-01-02")))
	newName := filepath.Join(l.logDir, fmt.Sprintf("goattack_%s.log", timestamp))

	// 重命名当前文件
	if err := os.Rename(oldName, newName); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("重命名日志文件失败: %v", err)
	}

	// 清理旧日志文件
	l.cleanOldLogs()

	// 创建新文件
	return l.openLogFile()
}

// cleanOldLogs 清理旧日志文件
func (l *Logger) cleanOldLogs() {
	files, err := filepath.Glob(filepath.Join(l.logDir, "goattack_*.log"))
	if err != nil {
		return
	}

	if len(files) <= l.maxBackups {
		return
	}

	// 删除最旧的文件
	for i := 0; i < len(files)-l.maxBackups; i++ {
		os.Remove(files[i])
	}
}

// checkRotate 检查是否需要轮转
func (l *Logger) checkRotate() {
	if l.currentSize >= l.maxSize {
		if err := l.rotateLog(); err != nil {
			fmt.Fprintf(os.Stderr, "日志轮转失败: %v\n", err)
		}
	}
}

// log 记录日志
func (l *Logger) log(level LogLevel, format string, v ...interface{}) {
	if level < l.level {
		return
	}

	// 格式化日志消息
	message := fmt.Sprintf(format, v...)
	logLine := fmt.Sprintf("%s [%s] %s\n",
		time.Now().Format("2006/01/02 15:04:05"),
		levelNames[level],
		message)

	l.mu.Lock()
	defer l.mu.Unlock()

	// 同时输出到控制台和文件
	fmt.Print(logLine)
	if l.file != nil {
		l.file.WriteString(logLine)
	}

	// 更新当前文件大小
	l.currentSize += int64(len(logLine))

	// 检查是否需要轮转
	go l.checkRotate()
}

// SetLevel 设置日志级别
func SetLevel(level LogLevel) {
	if defaultLogger != nil {
		defaultLogger.level = level
	}
}

// Debug 调试日志
func Debug(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(DEBUG, format, v...)
	}
}

// Info 记录 INFO 级别的日志信息。
//
// 参数：
//   - format: 格式化字符串，支持 fmt.Printf 风格的格式化
//   - v: 可变参数，用于格式化字符串
//
// 示例：
//   log.Info("系统启动成功")
//   log.Info("用户 %s 登录成功", username)
func Info(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(INFO, format, v...)
	}
}

// Warn 记录 WARN 级别的日志信息。
//
// 参数：
//   - format: 格式化字符串，支持 fmt.Printf 风格的格式化
//   - v: 可变参数，用于格式化字符串
//
// 示例：
//   log.Warn("磁盘空间不足")
//   log.Warn("连接 %s 超时", host)
func Warn(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(WARN, format, v...)
	}
}

// Error 记录 ERROR 级别的日志信息。
//
// 参数：
//   - format: 格式化字符串，支持 fmt.Printf 风格的格式化
//   - v: 可变参数，用于格式化字符串
//
// 示例：
//   log.Error("数据库连接失败")
//   log.Error("文件 %s 读取失败: %v", filename, err)
func Error(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(ERROR, format, v...)
	}
}

// Fatal 记录 FATAL 级别的日志信息并退出程序。
//
// 该函数会：
//   1. 记录 FATAL 级别的日志
//   2. 调用 os.Exit(1) 退出程序
//
// 参数：
//   - format: 格式化字符串，支持 fmt.Printf 风格的格式化
//   - v: 可变参数，用于格式化字符串
//
// 示例：
//   log.Fatal("配置文件不存在")
//   log.Fatal("初始化数据库失败: %v", err)
func Fatal(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(FATAL, format, v...)
		os.Exit(1)
	}
}

// Close 关闭日志系统并释放资源。
//
// 该函数会：
//   1. 获取日志文件的互斥锁
//   2. 关闭当前日志文件
//   3. 释放文件句柄
//
// 注意：
//   - 建议在程序退出前调用此函数
//   - 通常与 defer 一起使用
//
// 示例：
//   func main() {
//       log.InitLogger()
//       defer log.Close()
//       // ... 程序逻辑
//   }
func Close() {
	if defaultLogger != nil && defaultLogger.file != nil {
		defaultLogger.mu.Lock()
		defer defaultLogger.mu.Unlock()
		defaultLogger.file.Close()
	}
}

// Debugf 记录 DEBUG 级别的日志信息（兼容标准库命名）。
//
// 这是 Debug 函数的别名，提供与标准库 log.Printf 类似的命名。
//
// 参数：
//   - format: 格式化字符串，支持 fmt.Printf 风格的格式化
//   - v: 可变参数，用于格式化字符串
func Debugf(format string, v ...interface{}) {
	Debug(format, v...)
}

// Infof 记录 INFO 级别的日志信息（兼容标准库命名）。
//
// 这是 Info 函数的别名，提供与标准库 log.Printf 类似的命名。
//
// 参数：
//   - format: 格式化字符串，支持 fmt.Printf 风格的格式化
//   - v: 可变参数，用于格式化字符串
func Infof(format string, v ...interface{}) {
	Info(format, v...)
}

// Warnf 记录 WARN 级别的日志信息（兼容标准库命名）。
//
// 这是 Warn 函数的别名，提供与标准库 log.Printf 类似的命名。
//
// 参数：
//   - format: 格式化字符串，支持 fmt.Printf 风格的格式化
//   - v: 可变参数，用于格式化字符串
func Warnf(format string, v ...interface{}) {
	Warn(format, v...)
}

// Errorf 记录 ERROR 级别的日志信息（兼容标准库命名）。
//
// 这是 Error 函数的别名，提供与标准库 log.Printf 类似的命名。
//
// 参数：
//   - format: 格式化字符串，支持 fmt.Printf 风格的格式化
//   - v: 可变参数，用于格式化字符串
func Errorf(format string, v ...interface{}) {
	Error(format, v...)
}

// Fatalf 记录 FATAL 级别的日志信息并退出程序（兼容标准库命名）。
//
// 这是 Fatal 函数的别名，提供与标准库 log.Fatalf 类似的命名。
//
// 参数：
//   - format: 格式化字符串，支持 fmt.Printf 风格的格式化
//   - v: 可变参数，用于格式化字符串
func Fatalf(format string, v ...interface{}) {
	Fatal(format, v...)
}

// Printf 普通日志（默认INFO级别）
func Printf(format string, v ...interface{}) {
	Info(format, v...)
}

// Print 普通日志（兼容标准库，默认INFO级别）
func Print(v ...interface{}) {
	Info(fmt.Sprint(v...))
}

// Println 普通日志（兼容标准库，默认INFO级别）
func Println(v ...interface{}) {
	Info(fmt.Sprint(v...))
}

// Panicf 触发panic的日志
func Panicf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	Error(msg)
	panic(msg)
}

// Panic 触发panic的日志（兼容标准库）
func Panic(v ...interface{}) {
	msg := fmt.Sprint(v...)
	Error(msg)
	panic(msg)
}

// Panicln 触发panic的日志（兼容标准库）
func Panicln(v ...interface{}) {
	msg := fmt.Sprint(v...)
	Error(msg)
	panic(msg)
}
