package main

import (
	"GoAttack/api"
	"GoAttack/common/config"
	"GoAttack/common/log"
	"GoAttack/common/postgres"
	"GoAttack/common/redis"
	"GoAttack/service"
	"GoAttack/service/intelligence"
	"fmt"
)

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

	r := api.SetupRouter()
	if err := r.Run(":3000"); err != nil {
		log.Fatal("启动服务器失败: %v", err)
	}
}
