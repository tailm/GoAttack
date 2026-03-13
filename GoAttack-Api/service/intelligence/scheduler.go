package intelligence

import (
	"GoAttack/common/log"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Scheduler 漏洞情报收集调度器
type Scheduler struct {
	collector *Collector
	db        *sql.DB
	jobs      map[int]*Job
	mu        sync.RWMutex
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// Job 收集任务
type Job struct {
	SourceID   int
	SourceName string
	Interval   time.Duration
	StopChan   chan struct{}
	Running    bool
}

// NewScheduler 创建新的调度器
func NewScheduler(db *sql.DB) *Scheduler {
	collector := NewCollector(db)
	return &Scheduler{
		collector: collector,
		db:        db,
		jobs:      make(map[int]*Job),
		stopChan:  make(chan struct{}),
	}
}

// Start 启动调度器
func (s *Scheduler) Start(ctx context.Context) error {
	log.Info("启动漏洞情报收集调度器")

	// 加载所有启用的情报源
	sources, err := s.loadEnabledSources(ctx)
	if err != nil {
		return err
	}

	// 为每个源创建定时任务
	for _, source := range sources {
		s.addJob(source)
	}

	// 启动监控goroutine
	s.wg.Add(1)
	go s.monitorSources(ctx)

	log.Infof("调度器已启动，共 %d 个收集任务", len(s.jobs))
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	log.Info("停止漏洞情报收集调度器")

	// 停止所有任务
	s.mu.Lock()
	for _, job := range s.jobs {
		if job.Running {
			close(job.StopChan)
			job.Running = false
		}
	}
	s.mu.Unlock()

	// 停止监控goroutine
	close(s.stopChan)
	s.wg.Wait()

	log.Info("调度器已停止")
}

// addJob 添加收集任务
func (s *Scheduler) addJob(source SourceConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果任务已存在，先停止
	if job, exists := s.jobs[source.ID]; exists {
		if job.Running {
			close(job.StopChan)
		}
	}

	// 创建新任务
	job := &Job{
		SourceID:   source.ID,
		SourceName: source.Name,
		Interval:   time.Duration(source.SyncInterval) * time.Second,
		StopChan:   make(chan struct{}),
		Running:    true,
	}

	s.jobs[source.ID] = job

	// 启动任务goroutine
	s.wg.Add(1)
	go s.runJob(job, source)
}

// removeJob 移除收集任务
func (s *Scheduler) removeJob(sourceID int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job, exists := s.jobs[sourceID]; exists {
		if job.Running {
			close(job.StopChan)
			job.Running = false
		}
		delete(s.jobs, sourceID)
	}
}

// runJob 运行收集任务
func (s *Scheduler) runJob(job *Job, source SourceConfig) {
	defer s.wg.Done()

	log.Infof("启动收集任务: %s (ID: %d, 间隔: %v)", source.Name, source.ID, job.Interval)

	// 立即执行一次
	s.executeCollection(job, source)

	// 定时执行
	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.executeCollection(job, source)
		case <-job.StopChan:
			log.Infof("停止收集任务: %s (ID: %d)", source.Name, source.ID)
			return
		}
	}
}

// executeCollection 执行收集操作
func (s *Scheduler) executeCollection(job *Job, source SourceConfig) {
	log.Debugf("开始收集: %s", source.Name)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 更新状态为进行中
	err := s.collector.UpdateSourceSyncStatus(ctx, source.ID, "pending", "开始收集")
	if err != nil {
		log.Errorf("更新源 %s 状态失败: %v", source.Name, err)
		return
	}

	// 执行收集
	vulnerabilities, err := s.collector.CollectFromSource(ctx, source)
	if err != nil {
		log.Errorf("收集源 %s 失败: %v", source.Name, err)
		_ = s.collector.UpdateSourceSyncStatus(ctx, source.ID, "failed", err.Error())
		return
	}

	// 保存到数据库
	if len(vulnerabilities) > 0 {
		err = s.collector.SaveToDatabase(ctx, vulnerabilities)
		if err != nil {
			log.Errorf("保存漏洞数据失败: %v", err)
			_ = s.collector.UpdateSourceSyncStatus(ctx, source.ID, "failed", "保存数据失败: "+err.Error())
			return
		}
	}

	// 更新状态为成功
	message := fmt.Sprintf("成功收集 %d 条漏洞数据", len(vulnerabilities))
	err = s.collector.UpdateSourceSyncStatus(ctx, source.ID, "success", message)
	if err != nil {
		log.Errorf("更新源 %s 状态失败: %v", source.Name, err)
		return
	}

	log.Infof("完成收集: %s, 收集到 %d 条漏洞数据", source.Name, len(vulnerabilities))
}

// monitorSources 监控情报源变化
func (s *Scheduler) monitorSources(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkSources(ctx)
		case <-s.stopChan:
			return
		}
	}
}

// checkSources 检查情报源变化
func (s *Scheduler) checkSources(ctx context.Context) {
	// 加载所有启用的情报源
	sources, err := s.loadEnabledSources(ctx)
	if err != nil {
		log.Errorf("加载情报源失败: %v", err)
		return
	}

	// 检查需要添加或更新的任务
	sourceMap := make(map[int]SourceConfig)
	for _, source := range sources {
		sourceMap[source.ID] = source

		s.mu.RLock()
		job, exists := s.jobs[source.ID]
		s.mu.RUnlock()

		if !exists {
			// 新源，添加任务
			log.Infof("发现新情报源: %s (ID: %d)", source.Name, source.ID)
			s.addJob(source)
		} else if job.Interval != time.Duration(source.SyncInterval)*time.Second {
			// 间隔变化，更新任务
			log.Infof("情报源 %s 间隔变化: %v -> %v", source.Name, job.Interval, time.Duration(source.SyncInterval)*time.Second)
			s.removeJob(source.ID)
			s.addJob(source)
		}
	}

	// 检查需要移除的任务
	s.mu.RLock()
	for sourceID, job := range s.jobs {
		if _, exists := sourceMap[sourceID]; !exists {
			// 源已禁用或删除，移除任务
			log.Infof("移除已禁用情报源: %s (ID: %d)", job.SourceName, sourceID)
			s.mu.RUnlock()
			s.removeJob(sourceID)
			s.mu.RLock()
		}
	}
	s.mu.RUnlock()
}

// loadEnabledSources 加载所有启用的情报源
func (s *Scheduler) loadEnabledSources(ctx context.Context) ([]SourceConfig, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, url, type, enabled, sync_interval, config, last_sync, last_sync_status
		FROM intelligence_source
		WHERE enabled = true
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []SourceConfig
	for rows.Next() {
		var source SourceConfig
		var configJSON []byte
		var lastSync *time.Time
		var lastSyncStatus *string

		err := rows.Scan(
			&source.ID,
			&source.Name,
			&source.URL,
			&source.Type,
			&source.Enabled,
			&source.SyncInterval,
			&configJSON,
			&lastSync,
			&lastSyncStatus,
		)
		if err != nil {
			log.Errorf("扫描情报源行失败: %v", err)
			continue
		}

		// 解析配置
		if len(configJSON) > 0 {
			var config map[string]interface{}
			if err := json.Unmarshal(configJSON, &config); err != nil {
				log.Errorf("解析情报源配置失败: %v", err)
			} else {
				source.Config = config
			}
		}

		source.LastSync = lastSync
		if lastSyncStatus != nil {
			source.LastSyncStatus = *lastSyncStatus
		}

		sources = append(sources, source)
	}

	return sources, nil
}

// TriggerManualSync 手动触发同步
func (s *Scheduler) TriggerManualSync(ctx context.Context, sourceID int) error {
	s.mu.RLock()
	job, exists := s.jobs[sourceID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("未找到源ID: %d", sourceID)
	}

	// 加载源配置
	var source SourceConfig
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, url, type, enabled, sync_interval, config
		FROM intelligence_source
		WHERE id = $1
	`, sourceID).Scan(
		&source.ID,
		&source.Name,
		&source.URL,
		&source.Type,
		&source.Enabled,
		&source.SyncInterval,
		&source.Config,
	)
	if err != nil {
		return fmt.Errorf("加载源配置失败: %v", err)
	}

	// 执行一次收集
	go s.executeCollection(job, source)

	return nil
}

// GetJobStatus 获取任务状态
func (s *Scheduler) GetJobStatus() map[int]JobStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := make(map[int]JobStatus)
	for id, job := range s.jobs {
		status[id] = JobStatus{
			SourceID:   job.SourceID,
			SourceName: job.SourceName,
			Interval:   job.Interval,
			Running:    job.Running,
		}
	}

	return status
}

// JobStatus 任务状态
type JobStatus struct {
	SourceID   int           `json:"source_id"`
	SourceName string        `json:"source_name"`
	Interval   time.Duration `json:"interval"`
	Running    bool          `json:"running"`
}

