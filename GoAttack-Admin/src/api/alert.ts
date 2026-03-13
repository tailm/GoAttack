import request from '@/utils/request'

// 预警模型
export interface Alert {
  id: number
  title: string
  description: string
  severity: 'critical' | 'high' | 'medium' | 'low'
  type: string
  source?: string
  status: 'unread' | 'read' | 'resolved'
  data?: Record<string, any>
  created_at: string
  read_at?: string
  resolved_at?: string
}

// 预警查询参数
export interface AlertQuery {
  page?: number
  pageSize?: number
  search?: string
  severity?: string
  status?: string
  type?: string
  startDate?: string
  endDate?: string
}

// 预警响应数据
export interface AlertResponse {
  code: number
  msg: string
  data?: {
    total: number
    page: number
    pageSize: number
    items: Alert[]
  }
}

// 预警详情响应
export interface AlertDetailResponse {
  code: number
  msg: string
  data?: Alert
}

// 预警创建请求
export interface AlertCreateRequest {
  title: string
  description: string
  severity: 'critical' | 'high' | 'medium' | 'low'
  type: string
  source?: string
  data?: Record<string, any>
}

// 预警状态更新请求
export interface AlertStatusRequest {
  status: 'unread' | 'read' | 'resolved'
}

// 预警统计响应
export interface AlertStatsResponse {
  code: number
  msg: string
  data?: {
    total: number
    unread: number
    critical: number
    high: number
    medium: number
    low: number
    byType: Record<string, number>
    byDay: Array<{
      date: string
      count: number
    }>
  }
}

// 预警订阅请求
export interface AlertSubscribeRequest {
  email?: boolean
  webhook?: boolean
  dingtalk?: boolean
  sms?: boolean
  severity_filter?: string[]
  type_filter?: string[]
  frequency?: 'realtime' | 'hourly' | 'daily' | 'weekly'
}

// 预警订阅响应
export interface AlertSubscribeResponse {
  code: number
  msg: string
  data?: {
    subscribed: boolean
    channels: string[]
    filters: Record<string, any>
  }
}

// 获取预警列表
export function getAlertList(params: AlertQuery) {
  return request.get<AlertResponse>('/api/alerts', {
    params,
  })
}

// 创建预警
export function createAlert(data: AlertCreateRequest) {
  return request.post<AlertDetailResponse>('/api/alerts', data)
}

// 更新预警状态
export function updateAlertStatus(id: number, status: string) {
  return request.put<AlertDetailResponse>(`/api/alerts/${id}/status`, { status })
}

// 删除预警
export function deleteAlert(id: number) {
  return request.delete<{ code: number; msg: string }>(`/api/alerts/${id}`)
}

// 获取预警统计
export function getAlertStats() {
  return request.get<AlertStatsResponse>('/api/alerts/stats')
}

// 订阅预警
export function subscribeAlert(data: AlertSubscribeRequest) {
  return request.post<AlertSubscribeResponse>('/api/alerts/subscribe', data)
}