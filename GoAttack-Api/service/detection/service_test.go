package detection

import (
	"context"
	"database/sql"
	"encoding/json"
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
		assert.NotNil(t, service.engine, "Engine不应为nil")
	})

	t.Run("空数据库连接", func(t *testing.T) {
		// 测试空数据库连接
		service := NewService(nil)
		
		// Service应该仍然被创建
		assert.NotNil(t, service, "Service实例不应为nil，即使数据库为nil")
		assert.Nil(t, service.db, "数据库连接应为nil")
		assert.NotNil(t, service.engine, "Engine不应为nil")
	})
}

func TestCreateTask(t *testing.T) {
	// 创建模拟数据库
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "创建模拟数据库失败")
	defer db.Close()

	// 创建Service实例
	service := NewService(db)

	t.Run("创建有效任务", func(t *testing.T) {
		ctx := context.Background()
		
		// 创建测试任务请求
		taskRequest := CreateTaskRequest{
			Name:        "测试扫描任务",
			Description: "这是一个测试任务",
			Targets:     []string{"192.168.1.1", "example.com"},
			Rules:       []string{"CVE-2021-1234", "CVE-2021-5678"},
			Config: map[string]interface{}{
				"timeout": 30,
				"threads": 10,
			},
			CreatedBy: "testuser",
		}

		// 序列化JSON字段
		targetsJSON, _ := json.Marshal(taskRequest.Targets)
		rulesJSON, _ := json.Marshal(taskRequest.Rules)
		configJSON, _ := json.Marshal(taskRequest.Config)

		// 设置模拟查询
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO detection_task`).
			WithArgs(
				taskRequest.Name,
				taskRequest.Description,
				sqlmock.AnyArg(), // status
				sqlmock.AnyArg(), // progress
				targetsJSON,
				rulesJSON,
				configJSON,
				taskRequest.CreatedBy,
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		// 调用CreateTask
		taskID, err := service.CreateTask(ctx, taskRequest)
		
		// 验证结果
		assert.NoError(t, err, "CreateTask不应返回错误")
		assert.Equal(t, 1, taskID, "任务ID应为1")
		
		// 验证所有期望的查询都执行了
		assert.NoError(t, mock.ExpectationsWereMet(), "所有期望的查询都应执行")
	})

	t.Run("创建任务验证失败", func(t *testing.T) {
		ctx := context.Background()
		
		// 创建无效任务请求（空名称）
		taskRequest := CreateTaskRequest{
			Name:        "", // 空名称
			Description: "无效任务",
			Targets:     []string{},
			Rules:       []string{},
			Config:      map[string]interface{}{},
			CreatedBy:   "testuser",
		}

		// 调用CreateTask
		taskID, err := service.CreateTask(ctx, taskRequest)
		
		// 验证结果
		assert.Error(t, err, "CreateTask应返回验证错误")
		assert.Equal(t, 0, taskID, "任务ID应为0")
		assert.Contains(t, err.Error(), "任务名称", "错误消息应包含任务名称")
	})

	t.Run("JSON序列化失败", func(t *testing.T) {
		ctx := context.Background()
		
		// 创建包含无效JSON值的任务请求
		taskRequest := CreateTaskRequest{
			Name:        "测试任务",
			Description: "测试",
			Targets:     []string{"valid"},
			Rules:       []string{"valid"},
			Config: map[string]interface{}{
				"invalid": make(chan int), // 通道无法序列化为JSON
			},
			CreatedBy: "testuser",
		}

		// 调用CreateTask
		taskID, err := service.CreateTask(ctx, taskRequest)
		
		// 验证结果
		assert.Error(t, err, "CreateTask应返回JSON序列化错误")
		assert.Equal(t, 0, taskID, "任务ID应为0")
		assert.Contains(t, err.Error(), "序列化", "错误消息应包含序列化相关")
	})

	t.Run("数据库错误", func(t *testing.T) {
		ctx := context.Background()
		
		// 创建有效任务请求
		taskRequest := CreateTaskRequest{
			Name:        "测试任务",
			Description: "测试",
			Targets:     []string{"192.168.1.1"},
			Rules:       []string{"CVE-2021-1234"},
			Config:      map[string]interface{}{},
			CreatedBy:   "testuser",
		}

		// 序列化JSON字段
		targetsJSON, _ := json.Marshal(taskRequest.Targets)
		rulesJSON, _ := json.Marshal(taskRequest.Rules)
		configJSON, _ := json.Marshal(taskRequest.Config)

		// 设置模拟查询返回错误
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO detection_task`).
			WithArgs(
				taskRequest.Name,
				taskRequest.Description,
				sqlmock.AnyArg(), // status
				sqlmock.AnyArg(), // progress
				targetsJSON,
				rulesJSON,
				configJSON,
				taskRequest.CreatedBy,
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
			).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		// 调用CreateTask
		taskID, err := service.CreateTask(ctx, taskRequest)
		
		// 验证结果
		assert.Error(t, err, "CreateTask应返回数据库错误")
		assert.Equal(t, 0, taskID, "任务ID应为0")
		assert.Contains(t, err.Error(), "connection is done", "错误消息应包含连接错误")
		
		// 验证所有期望的查询都执行了
		assert.NoError(t, mock.ExpectationsWereMet(), "所有期望的查询都应执行")
	})
}

func TestGetTask(t *testing.T) {
	// 创建模拟数据库
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "创建模拟数据库失败")
	defer db.Close()

	// 创建Service实例
	service := NewService(db)

	t.Run("获取存在的任务", func(t *testing.T) {
		ctx := context.Background()
		taskID := 1
		
		// 设置模拟查询
		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "status", "progress",
			"targets", "rules", "config", "created_by",
			"started_at", "finished_at", "created_at", "updated_at",
		}).AddRow(
			taskID,
			"测试任务",
			"任务描述",
			"pending",
			0.0,
			`["192.168.1.1"]`,
			`["CVE-2021-1234"]`,
			`{"timeout":30}`,
			"testuser",
			nil, // started_at
			nil, // finished_at
			time.Now(),
			time.Now(),
		)

		mock.ExpectQuery(`SELECT id, name, description, status, progress, targets, rules, config, created_by, started_at, finished_at, created_at, updated_at FROM detection_task WHERE id = \$1`).
			WithArgs(taskID).
			WillReturnRows(rows)

		// 调用GetTask
		task, err := service.GetTask(ctx, taskID)
		
		// 验证结果
		assert.NoError(t, err, "GetTask不应返回错误")
		assert.NotNil(t, task, "任务不应为nil")
		assert.Equal(t, taskID, task.ID, "任务ID应匹配")
		assert.Equal(t, "测试任务", task.Name, "任务名称应匹配")
		assert.Equal(t, "pending", task.Status, "任务状态应匹配")
		
		// 验证所有期望的查询都执行了
		assert.NoError(t, mock.ExpectationsWereMet(), "所有期望的查询都应执行")
	})

	t.Run("获取不存在的任务", func(t *testing.T) {
		ctx := context.Background()
		taskID := 999
		
		// 设置模拟查询返回空结果
		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "status", "progress",
			"targets", "rules", "config", "created_by",
			"started_at", "finished_at", "created_at", "updated_at",
		})

		mock.ExpectQuery(`SELECT id, name, description, status, progress, targets, rules, config, created_by, started_at, finished_at, created_at, updated_at FROM detection_task WHERE id = \$1`).
			WithArgs(taskID).
			WillReturnRows(rows)

		// 调用GetTask
		task, err := service.GetTask(ctx, taskID)
		
		// 验证结果
		assert.Error(t, err, "GetTask应返回错误")
		assert.Nil(t, task, "任务应为nil")
		assert.Equal(t, sql.ErrNoRows, err, "错误应为sql.ErrNoRows")
		
		// 验证所有期望的查询都执行了
		assert.NoError(t, mock.ExpectationsWereMet(), "所有期望的查询都应执行")
	})

	t.Run("数据库错误", func(t *testing.T) {
		ctx := context.Background()
		taskID := 1
		
		// 设置模拟查询返回错误
		mock.ExpectQuery(`SELECT id, name, description, status, progress, targets, rules, config, created_by, started_at, finished_at, created_at, updated_at FROM detection_task WHERE id = \$1`).
			WithArgs(taskID).
			WillReturnError(sql.ErrConnDone)

		// 调用GetTask
		task, err := service.GetTask(ctx, taskID)
		
		// 验证结果
		assert.Error(t, err, "GetTask应返回错误")
		assert.Nil(t, task, "任务应为nil")
		assert.Contains(t, err.Error(), "connection is done", "错误消息应包含连接错误")
		
		// 验证所有期望的查询都执行了
		assert.NoError(t, mock.ExpectationsWereMet(), "所有期望的查询都应执行")
	})

	t.Run("JSON解析错误", func(t *testing.T) {
		ctx := context.Background()
		taskID := 1
		
		// 设置模拟查询返回无效的JSON
		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "status", "progress",
			"targets", "rules", "config", "created_by",
			"started_at", "finished_at", "created_at", "updated_at",
		}).AddRow(
			taskID,
			"测试任务",
			"任务描述",
			"pending",
			0.0,
			`{invalid json}`, // 无效的JSON
			`["CVE-2021-1234"]`,
			`{"timeout":30}`,
			"testuser",
			nil, // started_at
			nil, // finished_at
			time.Now(),
			time.Now(),
		)

		mock.ExpectQuery(`SELECT id, name, description, status, progress, targets, rules, config, created_by, started_at, finished_at, created_at, updated_at FROM detection_task WHERE id = \$1`).
			WithArgs(taskID).
			WillReturnRows(rows)

		// 调用GetTask
		task, err := service.GetTask(ctx, taskID)
		
		// 验证结果
		assert.Error(t, err, "GetTask应返回JSON解析错误")
		assert.Nil(t, task, "任务应为nil")
		assert.Contains(t, err.Error(), "JSON", "错误消息应包含JSON相关")
		
		// 验证所有期望的查询都执行了
		assert.NoError(t, mock.ExpectationsWereMet(), "所有期望的查询都应执行")
	})
}

func TestListTasks(t *testing.T) {
	// 创建模拟数据库
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "创建模拟数据库失败")
	defer db.Close()

	// 创建Service实例
	service := NewService(db)

	t.Run("列出任务", func(t *testing.T) {
		ctx := context.Background()
		
		// 设置模拟查询
		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "status", "progress",
			"targets", "rules", "config", "created_by",
			"started_at", "finished_at", "created_at", "updated_at",
		}).
			AddRow(
				1, "任务1", "描述1", "pending", 0.0,
				`["target1"]`, `["rule1"]`, `{}`, "user1",
				nil, nil, time.Now(), time.Now(),
			).
			AddRow(
				2, "任务2", "描述2", "running", 50.0,
				`["target2"]`, `["rule2"]`, `{"timeout":30}`, "user2",
				time.Now(), nil, time.Now(), time.Now(),
			)

		mock.ExpectQuery(`SELECT id, name, description, status, progress, targets, rules, config, created_by, started_at, finished_at, created_at, updated_at FROM detection_task`).
			WillReturnRows(rows)
		
		// 设置计数查询
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM detection_task`).
			WillReturnRows(countRows)

		// 调用ListTasks
		tasks, total, err := service.ListTasks(ctx, 1, 10, "", "")
		
		// 验证结果
		assert.NoError(t, err, "ListTasks不应返回错误")
		assert.Equal(t, 2, total, "总数应为2")
		assert.Len(t, tasks, 2, "应返回2个任务")
		
		if len(tasks) >= 2 {
			assert.Equal(t, 1, tasks[0].ID, "第一个任务ID应为1")
			assert.Equal(t, "任务1", tasks[0].Name, "第一个任务名称应匹配")
			assert.Equal(t, "pending", tasks[0].Status, "第一个任务状态应匹配")
			
			assert.Equal(t, 2, tasks[1].ID, "第二个任务ID应为2")
			assert.Equal(t, "任务2", tasks[1].Name, "第二个任务名称应匹配")
			assert.Equal(t, "running", tasks[1].Status, "第二个任务状态应匹配")
			assert.Equal(t, 50.0, tasks[1].Progress, "第二个任务进度应匹配")
		}
		
		// 验证所有期望的查询都执行了
		assert.NoError(t, mock.ExpectationsWereMet(), "所有期望的查询都应执行")
	})

	t.Run("带搜索条件列出任务", func(t *testing.T) {
		ctx := context.Background()
		searchName := "测试"
		searchStatus := "running"
		
		// 设置模拟查询
		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "status", "progress",
			"targets", "rules", "config", "created_by",
			"started_at", "finished_at", "created_at", "updated_at",
		}).
			AddRow(
				1, "测试任务", "描述", "running", 75.0,
				`["target"]`, `["rule"]`, `{}`, "user",
				time.Now(), nil, time.Now(), time.Now(),
			)

		mock.ExpectQuery(`SELECT id, name, description, status, progress, targets, rules, config, created_by, started_at, finished_at, created_at, updated_at FROM detection_task WHERE 1=1 AND name LIKE '%' \|\| \$1 \|\| '%' AND status = \$2`).
			WithArgs(searchName, searchStatus).
			WillReturnRows(rows)
		
		// 设置计数查询
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM detection_task WHERE 1=1 AND name LIKE '%' \|\| \$1 \|\| '%' AND status = \$2`).
			WithArgs(searchName, searchStatus).
			WillReturnRows(countRows)

		// 调用ListTasks
		tasks, total, err := service.ListTasks(ctx, 1, 10, searchName, searchStatus)
		
		// 验证结果
		assert.NoError(t, err, "ListTasks不应返回错误")
		assert.Equal(t, 1, total, "总数应为1")
		assert.Len(t, tasks, 1, "应返回1个任务")
		
		// 验证所有期望的查询都执行了
		assert.NoError(t, mock.ExpectationsWereMet(), "所有期望的查询都应执行")
	})

	t.Run("空结果", func(t *testing.T) {
		ctx := context.Background()
		
		// 设置模拟查询返回空结果
		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "status", "progress",
			"targets", "rules", "config", "created_by",
			"started_at", "finished_at", "created_at", "updated_at",
		})

		mock.ExpectQuery(`SELECT id, name, description, status, progress, targets, rules, config, created_by, started_at, finished_at, created_at, updated_at FROM detection_task`).
			WillReturnRows(rows)
		
		// 设置计数查询
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM detection_task`).
			WillReturnRows(countRows)

		// 调用ListTasks
		tasks, total, err := service.ListTasks(ctx, 1, 10, "", "")
		
		// 验证结果
		assert.NoError(t, err, "ListTasks不应返回错误")
		assert.Equal(t, 0, total, "总数应为0")
		assert.Empty(t, tasks, "应返回空切片")
		
		// 验证所有期望的查询都执行了
		assert.NoError(t, mock.ExpectationsWereMet(), "所有期望的查询都应执行")
	})
}

func TestUpdateTaskStatus(t *testing.T) {
	// 创建模拟数据库
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "创建模拟数据库失败")
	defer db.Close()

	// 创建Service实例
	service := NewService(db)

	t.Run("更新任务状态", func(t *testing.T) {
		ctx := context.Background()
		taskID := 1
		status := "running"
		progress := 25.5
		
		// 设置模拟查询
		mock.ExpectExec(`UPDATE detection_task SET status = \$1, progress = \$2, updated_at = NOW\(\) WHERE id = \$3`).
			WithArgs(status, progress, taskID).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 影响1行

		// 调用UpdateTaskStatus
		err := service.UpdateTaskStatus(ctx, taskID, status, progress)
		
		// 验证结果
		assert.NoError(t, err, "UpdateTaskStatus不应返回错误")
		
		// 验证所有期望的查询都执行了
		assert.NoError(t, mock.ExpectationsWereMet(), "所有期望的查询都应执行")
	})

	t.Run("更新不存在的任务", func(t *testing.T) {
		ctx := context.Background()
		taskID := 999
		status := "running"
		progress := 25.5
		
		// 设置模拟查询返回0行受影响
		mock.ExpectExec(`UPDATE detection_task SET status = \$1, progress = \$2, updated_at = NOW\(\) WHERE id = \$3`).
			WithArgs(status, progress, taskID).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 影响0行

		// 调用UpdateTaskStatus
		err := service.UpdateTaskStatus(ctx, taskID, status, progress)
		
		// 验证结果
		assert.NoError(t, err, "UpdateTaskStatus不应返回错误（任务不存在）")
		
		// 验证所有期望的查询都执行了
		assert.NoError(t, mock.ExpectationsWereMet(), "所有期望的查询都应执行")
	})

	t.Run("数据库错误", func(t *testing.T) {
		ctx := context.Background()
		taskID := 1
		status := "running"
		progress := 25.5
		
		// 设置模拟查询返回错误
		mock.ExpectExec(`UPDATE detection_task SET status = \$1, progress = \$2, updated_at = NOW\(\) WHERE id = \$3`).
			WithArgs(status, progress, taskID).
			WillReturnError(sql.ErrConnDone)

		// 调用UpdateTaskStatus
		err := service.UpdateTaskStatus(ctx, taskID, status, progress)
		
		// 验证结果
		assert.Error(t, err, "UpdateTaskStatus应返回错误")
		assert.Contains(t, err.Error(), "connection is done", "错误消息应包含连接错误")
		
		// 验证所有期望的查询都执行了
		assert.NoError(t, mock.ExpectationsWereMet(), "所有期望的查询都应执行")
	})
}

func TestServiceConcurrentAccess(t *testing.T) {
	// 创建模拟数据库
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "创建模拟数据库失败")
	defer db.Close()

	// 创建Service实例
	service := NewService(db)

	t.Run("并发创建任务", func(t *testing.T) {
		concurrency := 3
		done := make(chan bool, concurrency)

		// 设置多个模拟查询
		for i := 0; i < concurrency; i++ {
			taskID := i + 1
			
			mock.ExpectBegin()
			mock.ExpectQuery(`INSERT INTO detection_task`).
				WithArgs(
					sqlmock.AnyArg(), // name
					sqlmock.AnyArg(), // description
					sqlmock.AnyArg(), // status
					sqlmock.AnyArg(), // progress
					sqlmock.AnyArg(), // targets
					sqlmock.AnyArg(), // rules
					sqlmock.AnyArg(), // config
					sqlmock.AnyArg(), // created_by
					sqlmock.AnyArg(), // created_at
					sqlmock.AnyArg(), // updated_at
				).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(taskID))
			mock.ExpectCommit()
		}

		// 并发调用CreateTask
		for i := 0; i < concurrency; i++ {
			go func(id int) {
				ctx := context.Background()
				taskRequest := CreateTaskRequest{
					Name:        "并发任务",
					Description: "并发测试",
					Targets:     []string{"192.168.1.1"},
					Rules:       []string{"CVE-2021-1234"},
					Config:      map[string]interface{}{},
					CreatedBy:   "testuser",
				}
				
				_, err := service.CreateTask(ctx, taskRequest)
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

func TestServiceErrorScenarios(t *testing.T) {
	// 创建模拟数据库
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "创建模拟数据库失败")
	defer db.Close()

	// 创建Service实例
	service := NewService(db)

	t.Run("无效的任务状态", func(t *testing.T) {
		ctx := context.Background()
		taskID := 1
		
		testCases := []struct {
			name     string
			status   string
			progress float64
			shouldErr bool
		}{
			{"有效状态", "pending", 0.0, false},
			{"有效状态", "running", 50.0, false},
			{"有效状态", "completed", 100.0, false},
			{"有效状态", "failed", 0.0, false},
			{"无效状态", "invalid", 0.0, true},
			{"无效进度", "running", -10.0, true},
			{"无效进度", "running", 150.0, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				if !tc.shouldErr {
					// 设置模拟查询
					mock.ExpectExec(`UPDATE detection_task SET status = \$1, progress = \$2, updated_at = NOW\(\) WHERE id = \$3`).
						WithArgs(tc.status, tc.progress, taskID).
						WillReturnResult(sqlmock.NewResult(0, 1))
				}

				err := service.UpdateTaskStatus(ctx, taskID, tc.status, tc.progress)
				
				if tc.shouldErr {
					assert.Error(t, err, "UpdateTaskStatus应返回错误")
				} else {
					assert.NoError(t, err, "UpdateTaskStatus不应返回错误")
				}
			})
		}
		
		// 验证所有期望的查询都执行了
		assert.NoError(t, mock.ExpectationsWereMet(), "所有期望的查询都应执行")
	})
}