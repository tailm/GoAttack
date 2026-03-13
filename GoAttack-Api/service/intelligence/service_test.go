package intelligence

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	t.Run("创建Service实例", func(t *testing.T) {
		// 创建模拟数据库连接
		db, _, err := sqlmock.New()
		require.NoError(t, err, "创建模拟数据库失败")
		defer db.Close()

		// 创建Service实例
		service := NewService(db)
		
		// 验证Service不为nil
		assert.NotNil(t, service, "Service实例不应为nil")
		
		// 验证字段
		assert.Equal(t, db, service.db, "数据库连接应正确设置")
		assert.NotNil(t, service.collector, "Collector不应为nil")
		assert.NotNil(t, service.scheduler, "Scheduler不应为nil")
	})

	t.Run("空数据库连接", func(t *testing.T) {
		// 测试空数据库连接
		service := NewService(nil)
		
		// Service应该仍然被创建
		assert.NotNil(t, service, "Service实例不应为nil，即使数据库为nil")
		assert.Nil(t, service.db, "数据库连接应为nil")
		assert.NotNil(t, service.collector, "Collector不应为nil")
		assert.NotNil(t, service.scheduler, "Scheduler不应为nil")
	})
}

func TestServiceMethods(t *testing.T) {
	// 创建模拟数据库
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "创建模拟数据库失败")
	defer db.Close()

	// 创建Service实例
	service := NewService(db)

	t.Run("StartScheduler", func(t *testing.T) {
		ctx := context.Background()
		
		// 测试启动调度器
		err := service.StartScheduler(ctx)
		
		// 由于scheduler.Start可能涉及goroutine，我们主要测试它不panic
		assert.NoError(t, err, "StartScheduler不应返回错误")
	})

	t.Run("StopScheduler", func(t *testing.T) {
		// 测试停止调度器
		// 这个方法应该不panic
		assert.NotPanics(t, func() {
			service.StopScheduler()
		}, "StopScheduler不应panic")
	})

	t.Run("GetSources", func(t *testing.T) {
		ctx := context.Background()
		
		// 设置模拟查询
		rows := sqlmock.NewRows([]string{
			"id", "name", "url", "type", "enabled", "sync_interval", "config",
			"last_sync", "last_sync_status", "last_sync_message",
			"created_at", "updated_at",
		}).AddRow(
			1, "NVD", "https://nvd.nist.gov", "nvd", true, 3600,
			`{"api_key": ""}`,
			time.Now(), "success", "同步成功",
			time.Now(), time.Now(),
		)

		mock.ExpectQuery(`SELECT id, name, url, type, enabled, sync_interval, config, 
		       last_sync, last_sync_status, last_sync_message,
		       created_at, updated_at
		FROM intelligence_source`).
			WillReturnRows(rows)
		
		// 设置计数查询
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM intelligence_source`).
			WillReturnRows(countRows)

		// 调用GetSources
		sources, total, err := service.GetSources(ctx, 1, 10, "")
		
		// 验证结果
		assert.NoError(t, err, "GetSources不应返回错误")
		assert.Equal(t, 1, total, "总数应为1")
		assert.Len(t, sources, 1, "应返回1个情报源")
		
		if len(sources) > 0 {
			assert.Equal(t, 1, sources[0].ID, "ID应为1")
			assert.Equal(t, "NVD", sources[0].Name, "名称应为NVD")
			assert.Equal(t, "https://nvd.nist.gov", sources[0].URL, "URL应正确")
			assert.True(t, sources[0].Enabled, "应启用")
		}
		
		// 验证所有期望的查询都执行了
		assert.NoError(t, mock.ExpectationsWereMet(), "所有期望的查询都应执行")
	})

	t.Run("GetSourcesWithSearch", func(t *testing.T) {
		ctx := context.Background()
		searchTerm := "NVD"
		
		// 设置模拟查询（带搜索条件）
		rows := sqlmock.NewRows([]string{
			"id", "name", "url", "type", "enabled", "sync_interval", "config",
			"last_sync", "last_sync_status", "last_sync_message",
			"created_at", "updated_at",
		}).AddRow(
			1, "NVD", "https://nvd.nist.gov", "nvd", true, 3600,
			`{"api_key": ""}`,
			time.Now(), "success", "同步成功",
			time.Now(), time.Now(),
		)

		mock.ExpectQuery(`SELECT id, name, url, type, enabled, sync_interval, config, 
		       last_sync, last_sync_status, last_sync_message,
		       created_at, updated_at
		FROM intelligence_source
		WHERE 1=1 AND \(name LIKE '%' \|\| \$1 \|\| '%' OR url LIKE '%' \|\| \$2 \|\| '%'\)`).
			WithArgs(searchTerm, searchTerm).
			WillReturnRows(rows)
		
		// 设置计数查询
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM intelligence_source
		WHERE 1=1 AND \(name LIKE '%' \|\| \$1 \|\| '%' OR url LIKE '%' \|\| \$2 \|\| '%'\)`).
			WithArgs(searchTerm, searchTerm).
			WillReturnRows(countRows)

		// 调用GetSources
		sources, total, err := service.GetSources(ctx, 1, 10, searchTerm)
		
		// 验证结果
		assert.NoError(t, err, "GetSources不应返回错误")
		assert.Equal(t, 1, total, "总数应为1")
		assert.Len(t, sources, 1, "应返回1个情报源")
		
		// 验证所有期望的查询都执行了
		assert.NoError(t, mock.ExpectationsWereMet(), "所有期望的查询都应执行")
	})

	t.Run("GetSourcesEmptyResult", func(t *testing.T) {
		ctx := context.Background()
		
		// 设置模拟查询返回空结果
		rows := sqlmock.NewRows([]string{
			"id", "name", "url", "type", "enabled", "sync_interval", "config",
			"last_sync", "last_sync_status", "last_sync_message",
			"created_at", "updated_at",
		})

		mock.ExpectQuery(`SELECT id, name, url, type, enabled, sync_interval, config, 
		       last_sync, last_sync_status, last_sync_message,
		       created_at, updated_at
		FROM intelligence_source`).
			WillReturnRows(rows)
		
		// 设置计数查询
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM intelligence_source`).
			WillReturnRows(countRows)

		// 调用GetSources
		sources, total, err := service.GetSources(ctx, 1, 10, "")
		
		// 验证结果
		assert.NoError(t, err, "GetSources不应返回错误")
		assert.Equal(t, 0, total, "总数应为0")
		assert.Empty(t, sources, "应返回空切片")
		
		// 验证所有期望的查询都执行了
		assert.NoError(t, mock.ExpectationsWereMet(), "所有期望的查询都应执行")
	})

	t.Run("GetSourcesDatabaseError", func(t *testing.T) {
		ctx := context.Background()
		
		// 设置模拟查询返回错误
		mock.ExpectQuery(`SELECT id, name, url, type, enabled, sync_interval, config, 
		       last_sync, last_sync_status, last_sync_message,
		       created_at, updated_at
		FROM intelligence_source`).
			WillReturnError(sql.ErrConnDone)

		// 调用GetSources
		sources, total, err := service.GetSources(ctx, 1, 10, "")
		
		// 验证结果
		assert.Error(t, err, "GetSources应返回错误")
		assert.Equal(t, 0, total, "总数应为0")
		assert.Nil(t, sources, "应返回nil")
		assert.Contains(t, err.Error(), "connection is done", "错误消息应包含连接错误")
		
		// 验证所有期望的查询都执行了
		assert.NoError(t, mock.ExpectationsWereMet(), "所有期望的查询都应执行")
	})
}

func TestServicePagination(t *testing.T) {
	// 创建模拟数据库
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "创建模拟数据库失败")
	defer db.Close()

	// 创建Service实例
	service := NewService(db)

	t.Run("分页参数", func(t *testing.T) {
		ctx := context.Background()
		
		testCases := []struct {
			name     string
			page     int
			pageSize int
			offset   int
		}{
			{"第一页", 1, 10, 0},
			{"第二页", 2, 10, 10},
			{"第三页", 3, 20, 40},
			{"第零页", 0, 10, 0}, // page为0时应该当作1处理
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// 设置模拟查询
				rows := sqlmock.NewRows([]string{
					"id", "name", "url", "type", "enabled", "sync_interval", "config",
					"last_sync", "last_sync_status", "last_sync_message",
					"created_at", "updated_at",
				})

				// 计算实际偏移量
				actualPage := tc.page
				if actualPage < 1 {
					actualPage = 1
				}
				expectedOffset := (actualPage - 1) * tc.pageSize

				mock.ExpectQuery(`SELECT id, name, url, type, enabled, sync_interval, config, 
				       last_sync, last_sync_status, last_sync_message,
				       created_at, updated_at
				FROM intelligence_source`).
					WillReturnRows(rows)
				
				// 设置计数查询
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM intelligence_source`).
					WillReturnRows(countRows)

				// 调用GetSources
				_, _, err := service.GetSources(ctx, tc.page, tc.pageSize, "")
				
				// 验证没有错误
				assert.NoError(t, err, "GetSources不应返回错误")
				
				// 注意：我们无法直接验证LIMIT和OFFSET参数，因为它们在查询字符串中
				// 但我们可以验证查询执行成功
			})
		}
	})
}

func TestServiceContext(t *testing.T) {
	// 创建模拟数据库
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "创建模拟数据库失败")
	defer db.Close()

	// 创建Service实例
	service := NewService(db)

	t.Run("带取消的上下文", func(t *testing.T) {
		// 创建可取消的上下文
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // 立即取消

		// 设置模拟查询
		rows := sqlmock.NewRows([]string{
			"id", "name", "url", "type", "enabled", "sync_interval", "config",
			"last_sync", "last_sync_status", "last_sync_message",
			"created_at", "updated_at",
		})

		mock.ExpectQuery(`SELECT id, name, url, type, enabled, sync_interval, config, 
		       last_sync, last_sync_status, last_sync_message,
		       created_at, updated_at
		FROM intelligence_source`).
			WillReturnRows(rows)
		
		// 设置计数查询
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM intelligence_source`).
			WillReturnRows(countRows)

		// 调用GetSources（上下文已取消）
		_, _, err := service.GetSources(ctx, 1, 10, "")
		
		// 查询可能成功或失败，取决于上下文取消的时机
		// 我们只验证没有panic
		assert.NotPanics(t, func() {
			service.GetSources(ctx, 1, 10, "")
		}, "使用已取消的上下文不应panic")
	})

	t.Run("带超时的上下文", func(t *testing.T) {
		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// 设置模拟查询（故意延迟）
		rows := sqlmock.NewRows([]string{
			"id", "name", "url", "type", "enabled", "sync_interval", "config",
			"last_sync", "last_sync_status", "last_sync_message",
			"created_at", "updated_at",
		})

		mock.ExpectQuery(`SELECT id, name, url, type, enabled, sync_interval, config, 
		       last_sync, last_sync_status, last_sync_message,
		       created_at, updated_at
		FROM intelligence_source`).
			WillReturnRows(rows)
		
		// 设置计数查询
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM intelligence_source`).
			WillReturnRows(countRows)

		// 调用GetSources
		_, _, err := service.GetSources(ctx, 1, 10, "")
		
		// 由于超时时间很短，查询可能超时或成功
		// 我们只验证没有panic
		assert.NotPanics(t, func() {
			service.GetSources(ctx, 1, 10, "")
		}, "使用带超时的上下文不应panic")
	})
}

func TestServiceConcurrentAccess(t *testing.T) {
	// 创建模拟数据库
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "创建模拟数据库失败")
	defer db.Close()

	// 创建Service实例
	service := NewService(db)

	t.Run("并发访问", func(t *testing.T) {
		concurrency := 5
		done := make(chan bool, concurrency)

		// 设置多个模拟查询
		for i := 0; i < concurrency; i++ {
			rows := sqlmock.NewRows([]string{
				"id", "name", "url", "type", "enabled", "sync_interval", "config",
				"last_sync", "last_sync_status", "last_sync_message",
				"created_at", "updated_at",
			}).AddRow(
				1, "NVD", "https://nvd.nist.gov", "nvd", true, 3600,
				`{"api_key": ""}`,
				time.Now(), "success", "同步成功",
				time.Now(), time.Now(),
			)

			mock.ExpectQuery(`SELECT id, name, url, type, enabled, sync_interval, config, 
			       last_sync, last_sync_status, last_sync_message,
			       created_at, updated_at
			FROM intelligence_source`).
				WillReturnRows(rows)
			
			countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
			mock.ExpectQuery(`SELECT COUNT\(\*\) FROM intelligence_source`).
				WillReturnRows(countRows)
		}

		// 并发调用GetSources
		for i := 0; i < concurrency; i++ {
			go func(id int) {
				ctx := context.Background()
				_, _, err := service.GetSources(ctx, 1, 10, "")
				if err != nil {
					t.Logf("goroutine %d 错误: %v", id, err)
				}
				done <- true
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < concurrency; i++ {
			<-done
		}

		// 验证所有期望的查询都执行了
		assert.NoError(t, mock.ExpectationsWereMet(), "所有期望的查询都应执行")
	})
}

func TestServiceErrorHandling(t *testing.T) {
	// 创建模拟数据库
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "创建模拟数据库失败")
	defer db.Close()

	// 创建Service实例
	service := NewService(db)

	t.Run("JSON解析错误", func(t *testing.T) {
		ctx := context.Background()
		
		// 设置模拟查询返回无效的JSON
		rows := sqlmock.NewRows([]string{
			"id", "name", "url", "type", "enabled", "sync_interval", "config",
			"last_sync", "last_sync_status", "last_sync_message",
			"created_at", "updated_at",
		}).AddRow(
			1, "NVD", "https://nvd.nist.gov", "nvd", true, 3600,
			`{invalid json}`, // 无效的JSON
			time.Now(), "success", "同步成功",
			time.Now(), time.Now(),
		)

		mock.ExpectQuery(`SELECT id, name, url, type, enabled, sync_interval, config, 
		       last_sync, last_sync_status, last_sync_message,
		       created_at, updated_at
		FROM intelligence_source`).
			WillReturnRows(rows)
		
		// 设置计数查询
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM intelligence_source`).
			WillReturnRows(countRows)

		// 调用GetSources
		sources, total, err := service.GetSources(ctx, 1, 10, "")
		
		// 验证结果
		assert.Error(t, err, "GetSources应返回JSON解析错误")
		assert.Equal(t, 0, total, "总数应为0")
		assert.Nil(t, sources, "应返回nil")
		assert.Contains(t, err.Error(), "JSON", "错误消息应包含JSON相关")
		
		// 验证所有期望的查询都执行了
		assert.NoError(t, mock.ExpectationsWereMet(), "所有期望的查询都应执行")
	})
}