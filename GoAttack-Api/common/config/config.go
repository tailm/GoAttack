package config

import (
	"os"
	"strconv"
)

// PostgreSQL 配置
var PGUser = getEnv("PG_USER", "zwj")
var PGPassword = getEnv("PG_PASSWORD", "tailm123") // 请修改为你的密码
var PGHost = getEnv("PG_HOST", "127.0.0.1")
var PGPort = getEnvInt("PG_PORT", 5432)
var PGDBName = getEnv("PG_DB", "goattack")
var PGSSLMode = getEnv("PG_SSLMODE", "disable")

// Redis 配置
var RedisPassword = getEnv("REDIS_PASSWORD", "")
var RedisHost = getEnv("REDIS_HOST", "127.0.0.1")
var RedisPort = getEnvInt("REDIS_PORT", 6379)

// Redis Addr in format host:port for easier docker environment mapping
var RedisAddr = getEnv("REDIS_ADDR", "")

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}
