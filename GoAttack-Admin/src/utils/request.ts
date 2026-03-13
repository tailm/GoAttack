import axios from 'axios'
import { getToken } from './auth'

// 创建axios实例
const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:3000',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    // 从localStorage获取token
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    // 如果返回的数据有code字段，根据code判断请求状态
    if (response.data && typeof response.data === 'object' && 'code' in response.data) {
      const { code, msg } = response.data
      if (code === 20000) {
        return response.data
      } else {
        // 非20000状态码，抛出错误
        const error = new Error(msg || '请求失败')
        ;(error as any).code = code
        ;(error as any).response = response
        return Promise.reject(error)
      }
    }
    return response.data
  },
  (error) => {
    // 处理HTTP错误
    if (error.response) {
      switch (error.response.status) {
        case 401:
          // token过期，跳转到登录页
          window.location.href = '/login'
          break
        case 403:
          console.error('权限不足')
          break
        case 404:
          console.error('请求的资源不存在')
          break
        case 500:
          console.error('服务器内部错误')
          break
        default:
          console.error('请求错误', error.response.status)
      }
    } else if (error.request) {
      console.error('网络错误，请检查网络连接')
    } else {
      console.error('请求配置错误', error.message)
    }
    return Promise.reject(error)
  }
)

export default request