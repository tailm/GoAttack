import request from '@/utils/request'

// 检测任务模型
export interface DetectionTask {
  id: number
  name: string
  description?: string
  detection_type: string
  targets: string[]
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
  config?: Record<string, any>
  schedule?: {
    cron?: string
    interval?: number
    enabled?: boolean
  }
  created_at: string
  updated_at: string
  started_at?: string
  completed_at?: string
  created_by: string
}

// 检测任务查询参数
export interface DetectionTaskQuery {
  page?: number
  pageSize?: number
  status?: string
  detectionType?: string
  search?: string
}

// 检测任务响应数据
export interface DetectionTaskResponse {
  code: number
  msg: string
  data?: {
    total: number
    page: number
    pageSize: number
    items: DetectionTask[]
  }
}

// 检测任务详情响应
export interface DetectionTaskDetailResponse {
  code: number
  msg: string
  data?: DetectionTask
}

// 检测任务创建请求
export interface DetectionTaskCreateRequest {
  name: string
  description?: string
  detection_type: string
  targets: string[]
  config?: Record<string, any>
  schedule?: {
    cron?: string
    interval?: number
    enabled?: boolean
  }
}

// 检测任务更新请求
export interface DetectionTaskUpdateRequest {
  name?: string
  description?: string
  detection_type?: string
  targets?: string[]
  config?: Record<string, any>
  schedule?: {
    cron?: string
    interval?: number
    enabled?: boolean
  }
  enabled?: boolean
}

// 检测任务状态更新请求
export interface DetectionTaskStatusRequest {
  status: 'pending' | 'running' | 'completed' | 'failed' | 'stopped'
}

// 检测结果响应
export interface DetectionResultResponse {
  code: number
  msg: string
  data?: {
    total: number
    page: number
    pageSize: number
    items: DetectionResult[]
  }
}

// 检测统计响应
export interface DetectionStatsResponse {
  code: number
  msg: string
  data?: {
    totalTasks: number
    runningTasks: number
    completedTasks: number
    failedTasks: number
    totalVulnerabilities: number
    bySeverity: Record<string, number>
    byType: Record<string, number>
    recentTasks: DetectionTask[]
  }
}

// 检测任务模型
export interface DetectionTask {
  id: number
  task_id: number
  name: string
  description?: string
  detection_type: string
  config?: Record<string, any>
  schedule?: {
    cron?: string
    interval?: number
    enabled?: boolean
  }
  status: 'pending' | 'running' | 'completed' | 'failed' | 'stopped'
  target_count?: number
  vulnerability_count?: number
  risk_level?: string
  last_run?: string
  next_run?: string
  enabled: boolean
  created_at: string
  updated_at: string
}

// 检测结果模型
export interface DetectionResult {
  id: number
  task_id: number
  asset_id?: number
  vulnerability_id?: number
  status: 'pending' | 'scanning' | 'completed' | 'failed'
  risk_level: 'critical' | 'high' | 'medium' | 'low' | 'info'
  details?: Record<string, any>
  evidence?: string
  verified: boolean
  verified_by?: string
  verified_at?: string
  created_at: string
  updated_at: string
}

// 获取检测任务列表
export function getDetectionTaskList(params: DetectionTaskQuery) {
  return request.get<DetectionTaskResponse>('/api/detection/tasks', {
    params
  })
}

// 获取检测任务详情
export function getDetectionTaskDetail(id: number) {
  return request.get<DetectionTaskDetailResponse>(`/api/detection/tasks/${id}`)
}

// 创建检测任务
export function createDetectionTask(data: DetectionTaskCreateRequest) {
  return request.post<DetectionTaskDetailResponse>('/api/detection/tasks', data)
}

// 更新检测任务状态
export function updateDetectionTaskStatus(id: number, status: string) {
  return request.put<DetectionTaskDetailResponse>(`/api/detection/tasks/${id}/status`, { status })
}

// 删除检测任务
export function deleteDetectionTask(id: number) {
  return request.delete<{ code: number; msg: string }>(`/api/detection/tasks/${id}`)
}

// 获取检测结果
export function getDetectionResults(taskId: number, params?: { page?: number; pageSize?: number }) {
  return request.get<DetectionResultResponse>(`/api/detection/tasks/${taskId}/results`, {
    params
  })
}

// 获取检测统计
export function getDetectionStats() {
  return request.get<DetectionStatsResponse>('/api/detection/stats')
}