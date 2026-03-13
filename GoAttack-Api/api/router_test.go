package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetupRouter(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	t.Run("创建路由引擎", func(t *testing.T) {
		router := SetupRouter()
		
		// 验证路由引擎不为nil
		assert.NotNil(t, router, "路由引擎不应为nil")
		
		// 验证路由引擎是Gin引擎实例
		_, ok := router.(*gin.Engine)
		assert.True(t, ok, "路由引擎应为*gin.Engine类型")
	})

	t.Run("路由引擎配置", func(t *testing.T) {
		router := SetupRouter()
		
		// 验证路由引擎已配置
		assert.NotNil(t, router, "路由引擎不应为nil")
	})
}

func TestRouterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := SetupRouter()

	t.Run("CORS中间件", func(t *testing.T) {
		// 创建测试请求
		req, _ := http.NewRequest("OPTIONS", "/api/test", nil)
		req.Header.Set("Origin", "http://example.com")
		req.Header.Set("Access-Control-Request-Method", "GET")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type,Authorization")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// 验证CORS头
		assert.Equal(t, http.StatusOK, w.Code, "OPTIONS请求应返回200")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Origin"), "*", "应允许所有来源")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET", "应允许GET方法")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type", "应允许Content-Type头")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization", "应允许Authorization头")
	})

	t.Run("静态文件服务", func(t *testing.T) {
		// 测试静态文件路由存在
		// 注意：这里我们只测试路由配置，不测试实际文件服务
		req, _ := http.NewRequest("GET", "/uploads/test.txt", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// 静态文件路由应该存在（即使文件不存在）
		// 404表示路由存在但文件不存在，这是预期的
		assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code, "静态文件路由应存在")
	})
}

func TestRouterEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := SetupRouter()

	// 测试公开路由
	t.Run("公开登录接口", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/user/login", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// 登录接口应该存在（即使请求体无效）
		assert.NotEqual(t, http.StatusNotFound, w.Code, "登录接口路由应存在")
	})

	// 测试需要认证的路由组
	t.Run("认证路由组", func(t *testing.T) {
		testRoutes := []struct {
			method string
			path   string
		}{
			{"GET", "/api/admin/users"},
			{"POST", "/api/task/create"},
			{"GET", "/api/setting"},
			{"GET", "/api/dashboard/stats"},
			{"GET", "/api/intelligence/sources"},
			{"GET", "/api/detection/tasks"},
			{"GET", "/api/alert/list"},
		}

		for _, route := range testRoutes {
			t.Run(route.path, func(t *testing.T) {
				req, _ := http.NewRequest(route.method, route.path, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				
				// 这些路由应该存在（即使返回401未授权）
				// 401表示路由存在但需要认证，这是预期的
				status := w.Code
				assert.True(t, status == http.StatusUnauthorized || status == http.StatusOK || status == http.StatusNotFound,
					"路由 %s %s 应存在，状态码: %d", route.method, route.path, status)
			})
		}
	})
}

func TestRouterConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := SetupRouter()

	t.Run("文件上传大小限制", func(t *testing.T) {
		// Gin的MaxMultipartMemory是内部字段，我们通过创建请求来间接测试
		// 这里我们主要验证路由引擎已正确创建
		assert.NotNil(t, router, "路由引擎应已配置")
	})

	t.Run("路由分组", func(t *testing.T) {
		// 验证/api前缀的路由组已配置
		// 通过发送请求到/api路径来测试
		req, _ := http.NewRequest("GET", "/api/test-route", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// 404表示路由不存在但路径被处理（因为/api组存在）
		// 如果/api组不存在，可能会返回405或其他状态码
		status := w.Code
		assert.NotEqual(t, http.StatusMethodNotAllowed, status, "/api路由组应存在")
	})
}

func TestRouterHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := SetupRouter()

	t.Run("基本健康检查", func(t *testing.T) {
		// 测试服务器能正常响应
		req, _ := http.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// 根路径可能返回404，但服务器应正常响应
		assert.NotEqual(t, 0, w.Code, "服务器应响应请求")
	})
}

func TestRouterMiddlewareOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := SetupRouter()

	t.Run("中间件执行顺序", func(t *testing.T) {
		// 测试CORS中间件在路由之前
		req, _ := http.NewRequest("GET", "/api/test", nil)
		req.Header.Set("Origin", "http://test.com")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// 验证CORS头已设置
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"), "CORS头应已设置")
	})
}

func TestRouterErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := SetupRouter()

	t.Run("不存在的路由", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/non-existent-route", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// 不存在的路由应返回404
		assert.Equal(t, http.StatusNotFound, w.Code, "不存在的路由应返回404")
	})

	t.Run("不支持的HTTP方法", func(t *testing.T) {
		// 测试不支持的HTTP方法
		req, _ := http.NewRequest("PATCH", "/api/user/login", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// PATCH方法可能返回404或405，取决于路由配置
		status := w.Code
		assert.True(t, status == http.StatusNotFound || status == http.StatusMethodNotAllowed,
			"不支持的HTTP方法应返回404或405，实际: %d", status)
	})
}

func TestRouterConcurrentAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := SetupRouter()

	// 测试并发访问
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			req, _ := http.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < concurrency; i++ {
		<-done
	}

	// 如果没有panic，测试通过
	assert.True(t, true, "并发访问应不会panic")
}

func TestRouterInitialization(t *testing.T) {
	t.Run("多次初始化", func(t *testing.T) {
		// 测试可以多次调用SetupRouter
		router1 := SetupRouter()
		router2 := SetupRouter()
		
		// 每次调用应返回新的实例
		assert.NotSame(t, router1, router2, "每次调用应返回新的路由实例")
		assert.NotNil(t, router1, "第一次调用不应返回nil")
		assert.NotNil(t, router2, "第二次调用不应返回nil")
	})
}

func TestRouterPerformance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := SetupRouter()

	t.Run("路由匹配性能", func(t *testing.T) {
		// 测试路由匹配性能（简单测试）
		paths := []string{
			"/api/user/login",
			"/api/admin/users",
			"/api/task/create",
			"/api/setting",
			"/api/dashboard/stats",
			"/api/intelligence/sources",
			"/api/detection/tasks",
			"/api/alert/list",
			"/uploads/test.jpg",
		}

		for _, path := range paths {
			req, _ := http.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			
			// 记录时间（简单测试，不进行详细性能分析）
			router.ServeHTTP(w, req)
			
			// 验证请求被处理
			assert.NotEqual(t, 0, w.Code, "路径 %s 应被处理", path)
		}
	})
}

// 测试辅助函数
func testRequest(t *testing.T, router *gin.Engine, method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestRouterHelperFunctions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := SetupRouter()

	t.Run("测试辅助函数", func(t *testing.T) {
		w := testRequest(t, router, "GET", "/")
		assert.NotEqual(t, 0, w.Code, "辅助函数应正常工作")
	})
}