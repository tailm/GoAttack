// Package api 提供 GoAttack 系统的 RESTful API 接口。
//
// 该包包含所有 HTTP 路由的配置和中间件设置，主要功能包括：
//   - 路由配置和分组
//   - 跨域资源共享 (CORS) 设置
//   - JWT 认证中间件
//   - 静态文件服务
//   - API 路由注册
//
// 路由结构：
//   - /api/user/login (公开): 用户登录接口
//   - /api/* (需要认证): 所有需要认证的 API 接口
//   - /uploads (公开): 静态文件服务
//
// 认证机制：
//   使用 JWT (JSON Web Token) 进行用户认证，所有 /api/* 路径的请求都需要有效的 JWT Token。
//
// 支持的 API 模块：
//   - admin: 用户管理和认证
//   - task: 扫描任务管理
//   - setting: 系统设置
//   - poc: POC 漏洞验证管理
//   - tools: 工具管理
//   - dashboard: 仪表盘数据
//   - plugin: 插件管理
//   - dict: 字典管理
//   - notification: 通知功能
//   - intelligence: 漏洞情报管理
//   - detection: 检测任务管理
//   - alert: 预警通知管理
//   - config: 系统配置管理
package api

import (
	"GoAttack/api/admin"
	"GoAttack/api/alert"
	"GoAttack/api/config"
	"GoAttack/api/dashboard"
	"GoAttack/api/detection"
	"GoAttack/api/dict"
	"GoAttack/api/intelligence"
	"GoAttack/api/notification"
	"GoAttack/api/plugin"
	"GoAttack/api/poc"
	"GoAttack/api/setting"
	"GoAttack/api/task"
	"GoAttack/api/tools"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter 创建并配置 Gin 路由引擎。
//
// 该函数负责：
//   - 创建默认的 Gin 引擎实例
//   - 配置 CORS 中间件以支持跨域请求
//   - 设置文件上传大小限制（1GB）
//   - 配置静态文件服务
//   - 设置公开的登录接口
//   - 注册所有需要认证的 API 路由
//
// 返回值：
//   *gin.Engine: 配置好的 Gin 路由引擎实例
//
// 示例：
//   r := SetupRouter()
//   r.Run(":3000")
//
// 路由分组：
//   - 公开路由:
//     - POST /api/user/login: 用户登录
//     - GET /uploads/*: 静态文件访问
//   - 需要认证的路由 (所有 /api/* 路径):
//     - 用户管理、任务管理、系统设置等所有业务接口
//
// 中间件：
//   - CORS: 允许所有来源的跨域请求
//   - AuthMiddleware: JWT 认证中间件，保护 /api/* 路径
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 设置上传文件时的最大内存限制，默认是32MB，对于巨量的POC文件批量导入可能不够，扩大到 1GB
	r.MaxMultipartMemory = 1024 << 20
	// CORS 设置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 静态文件服务（用于访问上传的头像等文件）
	r.Static("/uploads", "./uploads")

	// 登录接口（公开，不需要认证）
	r.POST("/api/user/login", admin.Login)

	// 所有 /api/* 路径都需要认证
	api := r.Group("/api")
	api.Use(admin.AuthMiddleware())
	{
		// 用户管理
		admin.RegisterRoutes(api)

		// 任务管理
		task.RegisterRoutes(api)

		// 系统设置
		setting.RegisterRoutes(api)

		// POC 管理
		poc.RegisterRoutes(api)
		tools.RegisterRoutes(api)

		// 仪表盘
		dashboard.RegisterRoutes(api)

		// 插件管理
		plugin.RegisterRoutes(api)

		// 字典管理
		dict.RegisterRoutes(api)

		// 通知功能
		notification.RegisterRoutes(api)

		// 漏洞情报管理
		intelligence.RegisterRoutes(api)

		// 检测任务管理
		detection.RegisterRoutes(api)

		// 预警通知管理
		alert.RegisterRoutes(api)

		// 系统配置管理
		config.RegisterRoutes(api)
	}

	return r
}
