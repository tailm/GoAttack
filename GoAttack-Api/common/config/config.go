// Package config 提供 GoAttack 系统的配置管理功能。
//
// 该包负责从环境变量中读取系统配置，支持以下配置项：
//   - PostgreSQL 数据库连接配置
//   - Redis 缓存连接配置
//
// 配置优先级：
//   1. 环境变量（最高优先级）
//   2. 默认值（当环境变量未设置时使用）
//
// 支持的 PostgreSQL 环境变量：
//   - PG_USER: PostgreSQL 用户名（默认：zwj）
//   - PG_PASSWORD: PostgreSQL 密码（默认：tailm123）
//   - PG_HOST: PostgreSQL 主机地址（默认：127.0.0.1）
//   - PG_PORT: PostgreSQL 端口（默认：5432）
//   - PG_DB: PostgreSQL 数据库名（默认：goattack）
//   - PG_SSLMODE: PostgreSQL SSL 模式（默认：disable）
//
// 支持的 Redis 环境变量：
//   - REDIS_PASSWORD: Redis 密码（默认：空）
//   - REDIS_HOST: Redis 主机地址（默认：127.0.0.1）
//   - REDIS_PORT: Redis 端口（默认：6379）
//   - REDIS_ADDR: Redis 地址（格式：host:port，覆盖 HOST 和 PORT）
//
// 使用示例：
//   export PG_USER=admin
//   export PG_PASSWORD=secret
//   export REDIS_PASSWORD=redispass
//   go run main.go
package config

import (
	"os"
	"strconv"
)

// PostgreSQL 配置变量
//
// 这些变量存储 PostgreSQL 数据库的连接配置，可以通过环境变量覆盖默认值。
var (
	// PGUser PostgreSQL 用户名
	// 环境变量: PG_USER
	// 默认值: "zwj"
	PGUser = getEnv("PG_USER", "zwj")
	
	// PGPassword PostgreSQL 密码
	// 环境变量: PG_PASSWORD
	// 默认值: "tailm123"
	// 注意：在生产环境中请务必修改此默认值
	PGPassword = getEnv("PG_PASSWORD", "tailm123")
	
	// PGHost PostgreSQL 主机地址
	// 环境变量: PG_HOST
	// 默认值: "127.0.0.1"
	PGHost = getEnv("PG_HOST", "127.0.0.1")
	
	// PGPort PostgreSQL 端口
	// 环境变量: PG_PORT
	// 默认值: 5432
	PGPort = getEnvInt("PG_PORT", 5432)
	
	// PGDBName PostgreSQL 数据库名
	// 环境变量: PG_DB
	// 默认值: "goattack"
	PGDBName = getEnv("PG_DB", "goattack")
	
	// PGSSLMode PostgreSQL SSL 模式
	// 环境变量: PG_SSLMODE
	// 默认值: "disable"
	// 可选值: "disable", "require", "verify-ca", "verify-full"
	PGSSLMode = getEnv("PG_SSLMODE", "disable")
)

// Redis 配置变量
//
// 这些变量存储 Redis 缓存的连接配置，可以通过环境变量覆盖默认值。
var (
	// RedisPassword Redis 密码
	// 环境变量: REDIS_PASSWORD
	// 默认值: "" (空字符串，表示无密码)
	RedisPassword = getEnv("REDIS_PASSWORD", "")
	
	// RedisHost Redis 主机地址
	// 环境变量: REDIS_HOST
	// 默认值: "127.0.0.1"
	RedisHost = getEnv("REDIS_HOST", "127.0.0.1")
	
	// RedisPort Redis 端口
	// 环境变量: REDIS_PORT
	// 默认值: 6379
	RedisPort = getEnvInt("REDIS_PORT", 6379)
	
	// RedisAddr Redis 地址（格式：host:port）
	// 环境变量: REDIS_ADDR
	// 默认值: "" (空字符串，表示使用 RedisHost:RedisPort)
	// 注意：如果设置了此变量，将覆盖 RedisHost 和 RedisPort 的设置
	RedisAddr = getEnv("REDIS_ADDR", "")
)

// getEnv 从环境变量获取字符串值，如果不存在则返回默认值。
//
// 参数：
//   - key: 环境变量名
//   - fallback: 默认值，当环境变量不存在时返回
//
// 返回值：
//   string: 环境变量的值或默认值
//
// 示例：
//   dbHost := getEnv("DB_HOST", "localhost")
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvInt 从环境变量获取整数值，如果不存在或解析失败则返回默认值。
//
// 参数：
//   - key: 环境变量名
//   - fallback: 默认值，当环境变量不存在或解析失败时返回
//
// 返回值：
//   int: 环境变量的整数值或默认值
//
// 示例：
//   port := getEnvInt("PORT", 8080)
func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}
