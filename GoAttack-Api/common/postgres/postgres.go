package postgres

import (
	"GoAttack/common/config"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitDB 初始化数据库连接
func InitDB() error {
	// 1. 先连接到 postgres 默认数据库，用于创建目标数据库
	dsnWithoutDB := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		config.PGHost,
		config.PGPort,
		config.PGUser,
		config.PGPassword,
		config.PGSSLMode,
	)

	tempDB, err := sql.Open("postgres", dsnWithoutDB)
	if err != nil {
		return fmt.Errorf("open postgres (without db) failed: %v", err)
	}

	if err = tempDB.Ping(); err != nil {
		tempDB.Close()
		return fmt.Errorf("ping postgres (without db) failed: %v", err)
	}

	// 2. 创建数据库（如果不存在）
	var exists bool
	err = tempDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", config.PGDBName).Scan(&exists)
	if err != nil {
		tempDB.Close()
		return fmt.Errorf("check database existence failed: %v", err)
	}

	if !exists {
		createDBQuery := fmt.Sprintf("CREATE DATABASE %s ENCODING 'UTF8'", config.PGDBName)
		if _, err := tempDB.Exec(createDBQuery); err != nil {
			tempDB.Close()
			return fmt.Errorf("create database failed: %v", err)
		}
	}
	tempDB.Close()

	// 3. 连接到指定的数据库
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.PGHost,
		config.PGPort,
		config.PGUser,
		config.PGPassword,
		config.PGDBName,
		config.PGSSLMode,
	)

	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open postgres failed: %v", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("ping postgres failed: %v", err)
	}

	// 4. 检查是否需要初始化表结构
	if err := checkAndInitTables(); err != nil {
		return fmt.Errorf("init tables failed: %v", err)
	}

	// 自动创建新增的表（幂等，不影响已有数据）
	autoMigrate()

	log.Println("PostgreSQL connected successfully")
	return nil
}

// checkAndInitTables 检查表是否存在，不存在则执行 init.sql
func checkAndInitTables() error {
	var tableName string
	// 检查 user 表是否存在
	err := DB.QueryRow("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'user' LIMIT 1").Scan(&tableName)
	if err == sql.ErrNoRows {
		// user 表不存在，需要执行 init.sql
		log.Println("Database tables not found, initializing from common/sql/init.sql...")
		sqlBytes, err := os.ReadFile("common/sql/init.sql")
		if err != nil {
			return fmt.Errorf("read init.sql failed: %v", err)
		}

		// 执行整个 sql 脚本
		if _, err := DB.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("execute init.sql failed: %v", err)
		}
		log.Println("Database tables initialized successfully.")
		return nil
	} else if err != nil {
		return fmt.Errorf("check table existence failed: %v", err)
	}

	return nil
}

// autoMigrate 自动创建新增表（使用 CREATE TABLE IF NOT EXISTS，不影响已有数据）
func autoMigrate() {
	sqls := []string{
		`CREATE TABLE IF NOT EXISTS notification_read_time (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) NOT NULL UNIQUE,
			last_read_at TIMESTAMP DEFAULT '2000-01-01 00:00:00',
			last_cleared_at TIMESTAMP DEFAULT '2000-01-01 00:00:00'
		)`,
		`CREATE INDEX IF NOT EXISTS idx_notification_read_time_username ON notification_read_time(username)`,
	}
	for _, s := range sqls {
		if _, err := DB.Exec(s); err != nil {
			log.Printf("[AutoMigrate] 执行失败: %v", err)
		}
	}
}

// GetDB 获取数据库连接实例
func GetDB() *sql.DB {
	return DB
}

// Close 关闭数据库连接
func Close() {
	if DB != nil {
		DB.Close()
	}
}
