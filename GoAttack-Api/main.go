// Package main 是 GoAttack 漏洞检测与情报系统的入口点。
//
// GoAttack 是一个集成了漏洞情报收集、漏洞检测、预警通知和任务调度的安全平台。
// 该系统提供以下核心功能：
//   - 多源漏洞情报自动收集（NVD、CNNVD、CNVD、ExploitDB等）
//   - 基于规则的漏洞检测引擎
//   - 实时预警通知系统
//   - 任务调度和管理
//   - RESTful API 接口
//
// 主要组件：
//   - 漏洞情报收集器：从多个公开漏洞库自动同步漏洞信息
//   - 检测引擎：基于规则匹配进行漏洞检测
//   - 预警系统：通过多种渠道发送安全预警
//   - API服务器：提供管理界面和数据接口
//
// 启动流程：
//   1. 初始化日志系统
//   2. 连接 PostgreSQL 数据库
//   3. 连接 Redis 缓存
//   4. 启动定时任务调度器
//   5. 初始化各服务模块
//   6. 启动 HTTP API 服务器
package main

import (
	"GoAttack/api"
	"GoAttack/common/config"
	"GoAttack/common/log"
	"GoAttack/common/postgres"
	"GoAttack/common/redis"
	"GoAttack/service"
	"GoAttack/service/alert"
	"GoAttack/service/detection"
	"GoAttack/service/intelligence"
	"fmt"
)

// main 是 GoAttack 应用程序的入口函数。
//
// 该函数负责：
//   - 初始化所有系统组件
//   - 建立数据库和缓存连接
//   - 启动后台服务
//   - 运行 HTTP API 服务器
//
// 默认监听端口：3000
//
// 环境变量配置：
//   - POSTGRES_HOST: PostgreSQL 主机地址（默认：localhost）
//   - POSTGRES_PORT: PostgreSQL 端口（默认：5432）
//   - POSTGRES_USER: PostgreSQL 用户名（默认：postgres）
//   - POSTGRES_PASSWORD: PostgreSQL 密码（默认：goattack）
//   - POSTGRES_DB: PostgreSQL 数据库名（默认：goattack）
//   - REDIS_HOST: Redis 主机地址（默认：localhost）
//   - REDIS_PORT: Redis 端口（默认：6379）
//   - REDIS_PASSWORD: Redis 密码（可选）
//
// 示例：
//   POSTGRES_HOST=localhost POSTGRES_PASSWORD=secret go run main.go
func main() {
	// 初始化日志系统
	log.InitLogger()
	defer log.Close()

	log.Info("==================== GoAttack 启动 ====================")

	// 初始化 PostgreSQL 数据库
	err := postgres.InitDB()
	if err != nil {
		log.Fatal("PostgreSQL 初始化失败: %v", err)
	}
	defer postgres.Close()
	log.Info("PostgreSQL 数据库连接成功")

	// 初始化Redis连接
	err = redis.InitRedis(redis.RedisConfig{
		Host:     config.RedisHost,
		Port:     config.RedisPort,
		Password: config.RedisPassword,
		DB:       0,
	})
	if err != nil {
		log.Warn("Redis 连接失败: %v，将在没有 Redis 的情况下继续运行", err)
	} else {
		log.Info("Redis 连接成功")
	}
	defer redis.Close()

	fmt.Println("Starting GoAttack API Server...")
	log.Info("启动 GoAttack API 服务器，监听端口: 3000")

	// 启动定时任务调度器
	service.StartTaskScheduler()

	// 初始化漏洞情报服务
	if err := intelligence.Init(postgres.DB); err != nil {
		log.Warn("漏洞情报服务初始化失败: %v，将在没有漏洞情报功能的情况下继续运行", err)
	} else {
		defer intelligence.Stop()
		log.Info("漏洞情报服务初始化成功")
	}

	// 初始化检测引擎服务
	if err := detection.Init(postgres.DB); err != nil {
		log.Warn("检测引擎服务初始化失败: %v，将在没有检测功能的情况下继续运行", err)
	} else {
		log.Info("检测引擎服务初始化成功")
	}

	// 初始化预警通知服务
	if err := alert.Init(postgres.DB); err != nil {
		log.Warn("预警通知服务初始化失败: %v，将在没有预警功能的情况下继续运行", err)
	} else {
		log.Info("预警通知服务初始化成功")
	}

	r := api.SetupRouter()
	if err := r.Run(":3000"); err != nil {
		log.Fatal("启动服务器失败: %v", err)
	}
}
