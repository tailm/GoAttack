package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskStruct(t *testing.T) {
	t.Run("Task结构体字段", func(t *testing.T) {
		now := time.Now()
		task := Task{
			ID:          1,
			Name:        "测试任务",
			Target:      "192.168.1.0/24",
			Type:        "port",
			Status:      "pending",
			Progress:    0,
			Creator:     "admin",
			Description: "端口扫描测试",
			Options:     `{"threads": 10, "timeout": 30}`,
			CreatedAt:   now,
			UpdatedAt:   now,
			StartedAt:   nil,
			CompletedAt: nil,
		}

		// 验证字段值
		assert.Equal(t, 1, task.ID, "ID应匹配")
		assert.Equal(t, "测试任务", task.Name, "名称应匹配")
		assert.Equal(t, "192.168.1.0/24", task.Target, "目标应匹配")
		assert.Equal(t, "port", task.Type, "类型应匹配")
		assert.Equal(t, "pending", task.Status, "状态应匹配")
		assert.Equal(t, 0, task.Progress, "进度应匹配")
		assert.Equal(t, "admin", task.Creator, "创建者应匹配")
		assert.Equal(t, "端口扫描测试", task.Description, "描述应匹配")
		assert.Equal(t, `{"threads": 10, "timeout": 30}`, task.Options, "选项应匹配")
		assert.Equal(t, now, task.CreatedAt, "创建时间应匹配")
		assert.Equal(t, now, task.UpdatedAt, "更新时间应匹配")
		assert.Nil(t, task.StartedAt, "开始时间应为nil")
		assert.Nil(t, task.CompletedAt, "完成时间应为nil")
	})

	t.Run("Task JSON标签", func(t *testing.T) {
		now := time.Now()
		task := Task{
			ID:          1,
			Name:        "测试任务",
			Target:      "192.168.1.1",
			Type:        "web",
			Status:      "running",
			Progress:    50,
			Creator:     "user",
			Description: "Web扫描测试",
			Options:     "{}",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		// 测试JSON序列化
		data, err := json.Marshal(task)
		require.NoError(t, err, "JSON序列化不应失败")

		// 验证JSON包含正确的字段
		var decoded map[string]interface{}
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err, "JSON反序列化不应失败")

		assert.Equal(t, float64(1), decoded["id"], "JSON id字段应匹配")
		assert.Equal(t, "测试任务", decoded["name"], "JSON name字段应匹配")
		assert.Equal(t, "192.168.1.1", decoded["target"], "JSON target字段应匹配")
		assert.Equal(t, "web", decoded["type"], "JSON type字段应匹配")
		assert.Equal(t, "running", decoded["status"], "JSON status字段应匹配")
		assert.Equal(t, float64(50), decoded["progress"], "JSON progress字段应匹配")
		assert.Equal(t, "user", decoded["creator"], "JSON creator字段应匹配")
		assert.Equal(t, "Web扫描测试", decoded["description"], "JSON description字段应匹配")
		assert.Equal(t, "{}", decoded["options"], "JSON options字段应匹配")
	})

	t.Run("Task指针字段", func(t *testing.T) {
		now := time.Now()
		startedAt := now.Add(-1 * time.Hour)
		completedAt := now

		task := Task{
			ID:          2,
			Name:        "已完成任务",
			Target:      "example.com",
			Type:        "vuln",
			Status:      "completed",
			Progress:    100,
			Creator:     "scanner",
			Description: "漏洞扫描",
			Options:     `{"enable_weak_password": true}`,
			CreatedAt:   now.Add(-2 * time.Hour),
			UpdatedAt:   now,
			StartedAt:   &startedAt,
			CompletedAt: &completedAt,
		}

		// 验证指针字段
		assert.NotNil(t, task.StartedAt, "开始时间不应为nil")
		assert.NotNil(t, task.CompletedAt, "完成时间不应为nil")
		assert.Equal(t, startedAt, *task.StartedAt, "开始时间值应匹配")
		assert.Equal(t, completedAt, *task.CompletedAt, "完成时间值应匹配")
	})
}

func TestScanOptionsStruct(t *testing.T) {
	t.Run("ScanOptions默认值", func(t *testing.T) {
		options := ScanOptions{}

		// 验证默认值
		assert.False(t, options.EnableHostDiscovery, "EnableHostDiscovery默认应为false")
		assert.Equal(t, "", options.Ports, "Ports默认应为空")
		assert.False(t, options.EnableServiceDet, "EnableServiceDet默认应为false")
		assert.False(t, options.EnableWeakPassword, "EnableWeakPassword默认应为false")
		assert.False(t, options.EnableSubdomainEnum, "EnableSubdomainEnum默认应为false")
		assert.False(t, options.EnableDirScan, "EnableDirScan默认应为false")
		assert.False(t, options.EnableReverse, "EnableReverse默认应为false")
		assert.Equal(t, 0, options.Threads, "Threads默认应为0")
		assert.Equal(t, 0, options.Timeout, "Timeout默认应为0")
		assert.Equal(t, "", options.Advanced, "Advanced默认应为空")
		assert.Equal(t, "", options.ScheduledTime, "ScheduledTime默认应为空")
		assert.Equal(t, "", options.BlacklistPorts, "BlacklistPorts默认应为空")
		assert.Equal(t, "", options.BlacklistHosts, "BlacklistHosts默认应为空")
	})

	t.Run("ScanOptions完整配置", func(t *testing.T) {
		options := ScanOptions{
			EnableHostDiscovery: true,
			Ports:               "1-1000,8080,8443",
			EnableServiceDet:    true,
			EnableWeakPassword:  true,
			EnableSubdomainEnum: true,
			EnableDirScan:       true,
			EnableReverse:       true,
			Threads:            50,
			Timeout:            300,
			Advanced:           `{"aggressive": true, "rate_limit": 100}`,
			ScheduledTime:      "2024-01-01 10:00:00",
			BlacklistPorts:     "22,3389",
			BlacklistHosts:     "192.168.1.100,example.com",
		}

		// 验证字段值
		assert.True(t, options.EnableHostDiscovery, "EnableHostDiscovery应为true")
		assert.Equal(t, "1-1000,8080,8443", options.Ports, "Ports应匹配")
		assert.True(t, options.EnableServiceDet, "EnableServiceDet应为true")
		assert.True(t, options.EnableWeakPassword, "EnableWeakPassword应为true")
		assert.True(t, options.EnableSubdomainEnum, "EnableSubdomainEnum应为true")
		assert.True(t, options.EnableDirScan, "EnableDirScan应为true")
		assert.True(t, options.EnableReverse, "EnableReverse应为true")
		assert.Equal(t, 50, options.Threads, "Threads应匹配")
		assert.Equal(t, 300, options.Timeout, "Timeout应匹配")
		assert.Equal(t, `{"aggressive": true, "rate_limit": 100}`, options.Advanced, "Advanced应匹配")
		assert.Equal(t, "2024-01-01 10:00:00", options.ScheduledTime, "ScheduledTime应匹配")
		assert.Equal(t, "22,3389", options.BlacklistPorts, "BlacklistPorts应匹配")
		assert.Equal(t, "192.168.1.100,example.com", options.BlacklistHosts, "BlacklistHosts应匹配")
	})

	t.Run("ScanOptions JSON序列化", func(t *testing.T) {
		options := ScanOptions{
			EnableHostDiscovery: true,
			Ports:               "80,443",
			EnableServiceDet:    true,
			Threads:            10,
			Timeout:            60,
		}

		// 测试JSON序列化
		data, err := json.Marshal(options)
		require.NoError(t, err, "JSON序列化不应失败")

		// 验证JSON包含正确的字段
		var decoded map[string]interface{}
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err, "JSON反序列化不应失败")

		assert.Equal(t, true, decoded["enable_host_discovery"], "JSON enable_host_discovery字段应匹配")
		assert.Equal(t, "80,443", decoded["ports"], "JSON ports字段应匹配")
		assert.Equal(t, true, decoded["enable_service_det"], "JSON enable_service_det字段应匹配")
		assert.Equal(t, false, decoded["enable_weak_password"], "JSON enable_weak_password字段应匹配")
		assert.Equal(t, false, decoded["enable_subdomain_enum"], "JSON enable_subdomain_enum字段应匹配")
		assert.Equal(t, false, decoded["enable_dir_scan"], "JSON enable_dir_scan字段应匹配")
		assert.Equal(t, false, decoded["enable_reverse"], "JSON enable_reverse字段应匹配")
		assert.Equal(t, float64(10), decoded["threads"], "JSON threads字段应匹配")
		assert.Equal(t, float64(60), decoded["timeout"], "JSON timeout字段应匹配")
		assert.Equal(t, "", decoded["advanced"], "JSON advanced字段应匹配")
		assert.Equal(t, "", decoded["scheduled_time"], "JSON scheduled_time字段应匹配")
		assert.Equal(t, "", decoded["blacklist_ports"], "JSON blacklist_ports字段应匹配")
		assert.Equal(t, "", decoded["blacklist_hosts"], "JSON blacklist_hosts字段应匹配")
	})

	t.Run("ScanOptions JSON反序列化", func(t *testing.T) {
		jsonStr := `{
			"enable_host_discovery": true,
			"ports": "1-1024",
			"enable_service_det": false,
			"enable_weak_password": true,
			"enable_subdomain_enum": false,
			"enable_dir_scan": true,
			"enable_reverse": false,
			"threads": 20,
			"timeout": 120,
			"advanced": "{\"custom\": true}",
			"scheduled_time": "2024-12-31 23:59:59",
			"blacklist_ports": "22,23",
			"blacklist_hosts": "10.0.0.1"
		}`

		var options ScanOptions
		err := json.Unmarshal([]byte(jsonStr), &options)
		require.NoError(t, err, "JSON反序列化不应失败")

		// 验证字段值
		assert.True(t, options.EnableHostDiscovery, "EnableHostDiscovery应为true")
		assert.Equal(t, "1-1024", options.Ports, "Ports应匹配")
		assert.False(t, options.EnableServiceDet, "EnableServiceDet应为false")
		assert.True(t, options.EnableWeakPassword, "EnableWeakPassword应为true")
		assert.False(t, options.EnableSubdomainEnum, "EnableSubdomainEnum应为false")
		assert.True(t, options.EnableDirScan, "EnableDirScan应为true")
		assert.False(t, options.EnableReverse, "EnableReverse应为false")
		assert.Equal(t, 20, options.Threads, "Threads应匹配")
		assert.Equal(t, 120, options.Timeout, "Timeout应匹配")
		assert.Equal(t, `{"custom": true}`, options.Advanced, "Advanced应匹配")
		assert.Equal(t, "2024-12-31 23:59:59", options.ScheduledTime, "ScheduledTime应匹配")
		assert.Equal(t, "22,23", options.BlacklistPorts, "BlacklistPorts应匹配")
		assert.Equal(t, "10.0.0.1", options.BlacklistHosts, "BlacklistHosts应匹配")
	})
}

func TestTaskStatusValidation(t *testing.T) {
	t.Run("有效状态值", func(t *testing.T) {
		validStatuses := []string{"pending", "running", "completed", "failed", "cancelled"}
		
		for _, status := range validStatuses {
			t.Run(status, func(t *testing.T) {
				task := Task{Status: status}
				// 状态值应该被接受（这里只是测试结构体可以设置这些值）
				assert.Equal(t, status, task.Status, "状态值应被接受")
			})
		}
	})

	t.Run("状态转换", func(t *testing.T) {
		// 测试状态字段可以更新
		task := Task{Status: "pending"}
		assert.Equal(t, "pending", task.Status, "初始状态应为pending")
		
		task.Status = "running"
		assert.Equal(t, "running", task.Status, "状态应可更新为running")
		
		task.Status = "completed"
		assert.Equal(t, "completed", task.Status, "状态应可更新为completed")
		
		task.Status = "failed"
		assert.Equal(t, "failed", task.Status, "状态应可更新为failed")
	})
}

func TestTaskProgressValidation(t *testing.T) {
	t.Run("进度范围", func(t *testing.T) {
		testCases := []struct {
			name     string
			progress int
			valid    bool
		}{
			{"负进度", -10, false},
			{"零进度", 0, true},
			{"正常进度", 50, true},
			{"满进度", 100, true},
			{"超进度", 150, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				task := Task{Progress: tc.progress}
				// 进度值应该被接受（这里只是测试结构体可以设置这些值）
				// 实际验证应该在业务逻辑中
				assert.Equal(t, tc.progress, task.Progress, "进度值应被设置")
			})
		}
	})
}

func TestTaskTimestamps(t *testing.T) {
	t.Run("时间戳操作", func(t *testing.T) {
		now := time.Now()
		oneHourAgo := now.Add(-1 * time.Hour)
		twoHoursAgo := now.Add(-2 * time.Hour)

		task := Task{
			CreatedAt: twoHoursAgo,
			UpdatedAt: oneHourAgo,
		}

		// 验证时间戳
		assert.True(t, task.CreatedAt.Before(now), "创建时间应在当前时间之前")
		assert.True(t, task.UpdatedAt.Before(now), "更新时间应在当前时间之前")
		assert.True(t, task.CreatedAt.Before(task.UpdatedAt), "创建时间应在更新时间之前")

		// 更新更新时间
		task.UpdatedAt = now
		assert.True(t, task.UpdatedAt.After(oneHourAgo), "更新时间应已更新")
	})

	t.Run("指针时间戳", func(t *testing.T) {
		now := time.Now()
		startedAt := now.Add(-1 * time.Hour)
		completedAt := now

		task := Task{
			StartedAt:   &startedAt,
			CompletedAt: &completedAt,
		}

		// 验证指针时间戳
		require.NotNil(t, task.StartedAt, "StartedAt不应为nil")
		require.NotNil(t, task.CompletedAt, "CompletedAt不应为nil")
		assert.True(t, task.StartedAt.Before(*task.CompletedAt), "开始时间应在完成时间之前")
		assert.True(t, (*task.CompletedAt).Sub(*task.StartedAt) > 0, "时间差应为正")
	})
}

func TestScanOptionsValidation(t *testing.T) {
	t.Run("端口格式", func(t *testing.T) {
		testCases := []struct {
			name     string
			ports    string
			expected string
		}{
			{"单个端口", "80", "80"},
			{"多个端口", "80,443,8080", "80,443,8080"},
			{"端口范围", "1-1000", "1-1000"},
			{"混合格式", "80,443,8080,1-1000", "80,443,8080,1-1000"},
			{"空格处理", "80, 443, 8080", "80, 443, 8080"}, // 注意：空格会被保留
			{"空字符串", "", ""},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				options := ScanOptions{Ports: tc.ports}
				assert.Equal(t, tc.expected, options.Ports, "端口字符串应匹配")
			})
		}
	})

	t.Run("线程数验证", func(t *testing.T) {
		testCases := []struct {
			name     string
			threads  int
			expected int
		}{
			{"零线程", 0, 0},
			{"正数线程", 10, 10},
			{"大线程数", 1000, 1000},
			{"负数线程", -5, -5}, // 注意：负数应该被业务逻辑拒绝
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				options := ScanOptions{Threads: tc.threads}
				assert.Equal(t, tc.expected, options.Threads, "线程数应匹配")
			})
		}
	})

	t.Run("超时时间验证", func(t *testing.T) {
		testCases := []struct {
			name     string
			timeout  int
			expected int
		}{
			{"零超时", 0, 0},
			{"正数超时", 30, 30},
			{"长超时", 3600, 3600},
			{"负数超时", -10, -10}, // 注意：负数应该被业务逻辑拒绝
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				options := ScanOptions{Timeout: tc.timeout}
				assert.Equal(t, tc.expected, options.Timeout, "超时时间应匹配")
			})
		}
	})
}

func TestModelJSONCompatibility(t *testing.T) {
	t.Run("Task JSON往返", func(t *testing.T) {
		now := time.Now()
		startedAt := now.Add(-1 * time.Hour)
		completedAt := now

		original := Task{
			ID:          123,
			Name:        "JSON测试任务",
			Target:      "test.example.com",
			Type:        "web",
			Status:      "completed",
			Progress:    100,
			Creator:     "jsonuser",
			Description: "JSON序列化测试",
			Options:     `{"key": "value"}`,
			CreatedAt:   now,
			UpdatedAt:   now,
			StartedAt:   &startedAt,
			CompletedAt: &completedAt,
		}

		// 序列化为JSON
		data, err := json.Marshal(original)
		require.NoError(t, err, "序列化不应失败")

		// 反序列化
		var decoded Task
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err, "反序列化不应失败")

		// 验证字段
		assert.Equal(t, original.ID, decoded.ID, "ID应匹配")
		assert.Equal(t, original.Name, decoded.Name, "名称应匹配")
		assert.Equal(t, original.Target, decoded.Target, "目标应匹配")
		assert.Equal(t, original.Type, decoded.Type, "类型应匹配")
		assert.Equal(t, original.Status, decoded.Status, "状态应匹配")
		assert.Equal(t, original.Progress, decoded.Progress, "进度应匹配")
		assert.Equal(t, original.Creator, decoded.Creator, "创建者应匹配")
		assert.Equal(t, original.Description, decoded.Description, "描述应匹配")
		assert.Equal(t, original.Options, decoded.Options, "选项应匹配")
		// 注意：时间字段可能因时区而略有不同，所以我们只检查它们是否被设置
		assert.False(t, decoded.CreatedAt.IsZero(), "创建时间应被设置")
		assert.False(t, decoded.UpdatedAt.IsZero(), "更新时间应被设置")
		assert.NotNil(t, decoded.StartedAt, "开始时间不应为nil")
		assert.NotNil(t, decoded.CompletedAt, "完成时间不应为nil")
	})

	t.Run("ScanOptions JSON往返", func(t *testing.T) {
		original := ScanOptions{
			EnableHostDiscovery: true,
			Ports:               "1-65535",
			EnableServiceDet:    true,
			EnableWeakPassword:  false,
			EnableSubdomainEnum: true,
			EnableDirScan:       false,
			EnableReverse:       true,
			Threads:            100,
			Timeout:            600,
			Advanced:           `{"custom": true, "rate": 10}`,
			ScheduledTime:      "2024-01-01 00:00:00",
			BlacklistPorts:     "22,3389,5900",
			BlacklistHosts:     "192.168.1.1,10.0.0.1",
		}

		// 序列化为JSON
		data, err := json.Marshal(original)
		require.NoError(t, err, "序列化不应失败")

		// 反序列化
		var decoded ScanOptions
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err, "反序列化不应失败")

		// 验证所有字段
		assert.Equal(t, original.EnableHostDiscovery, decoded.EnableHostDiscovery, "EnableHostDiscovery应匹配")
		assert.Equal(t, original.Ports, decoded.Ports, "Ports应匹配")
		assert.Equal(t, original.EnableServiceDet, decoded.EnableServiceDet, "EnableServiceDet应匹配")
		assert.Equal(t, original.EnableWeakPassword, decoded.EnableWeakPassword, "EnableWeakPassword应匹配")
		assert.Equal(t, original.EnableSubdomainEnum, decoded.EnableSubdomainEnum, "EnableSubdomainEnum应匹配")
		assert.Equal(t, original.EnableDirScan, decoded.EnableDirScan, "EnableDirScan应匹配")
		assert.Equal(t, original.EnableReverse, decoded.EnableReverse, "EnableReverse应匹配")
		assert.Equal(t, original.Threads, decoded.Threads, "Threads应匹配")
		assert.Equal(t, original.Timeout, decoded.Timeout, "Timeout应匹配")
		assert.Equal(t, original.Advanced, decoded.Advanced, "Advanced应匹配")
		assert.Equal(t, original.ScheduledTime, decoded.ScheduledTime, "ScheduledTime应匹配")
		assert.Equal(t, original.BlacklistPorts, decoded.BlacklistPorts, "BlacklistPorts应匹配")
		assert.Equal(t, original.BlacklistHosts, decoded.BlacklistHosts, "BlacklistHosts应匹配")
	})
}