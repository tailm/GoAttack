package config

import (
	"GoAttack/common/postgres"
	"database/sql"
)

var db *sql.DB

// Init 初始化数据库连接
func Init() {
	if postgres.DB != nil {
		db = postgres.DB
	}
}