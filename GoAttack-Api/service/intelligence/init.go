package intelligence

import (
	"GoAttack/common/log"
	"context"
	"database/sql"
	"sync"
)

var (
	serviceInstance *Service
	serviceOnce     sync.Once
)

// Init 初始化漏洞情报服务
func Init(db *sql.DB) error {
	var initErr error
	serviceOnce.Do(func() {
		serviceInstance = NewService(db)
		
		// 启动调度器
		ctx := context.Background()
		if err := serviceInstance.StartScheduler(ctx); err != nil {
			log.Errorf("启动漏洞情报收集调度器失败: %v", err)
			initErr = err
			return
		}
		
		log.Info("漏洞情报服务初始化完成")
	})
	return initErr
}

// GetService 获取服务实例
func GetService() *Service {
	return serviceInstance
}

// Stop 停止服务
func Stop() {
	if serviceInstance != nil {
		serviceInstance.StopScheduler()
		log.Info("漏洞情报服务已停止")
	}
}