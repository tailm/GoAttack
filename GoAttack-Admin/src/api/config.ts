import request from '@/utils/request'

// 情报源模型
export interface IntelligenceSource {
  id: number
  name: string
  url: string
  type: 'avd' | 'nvd' | 'cnnvd' | 'cnvd' | 'exploitdb' | 'securityfocus' | 'custom'
  enabled: boolean
  sync_interval?: number
  last_sync?: string
  last_sync_status?: 'success' | 'failed' | 'pending'
  config?: Record<string, any>
  created_at: string
  updated_at: string
}

// 检测规则模型
export interface DetectionRule {
  id: number
  name: string
  description?: string
  type: 'port' | 'service' | 'version' | 'cve' | 'custom'
  condition: string
  action: 'alert' | 'block' | 'log' | 'report'
  severity: 'critical' | 'high' | 'medium' | 'low'
  enabled: boolean
  priority?: number
  tags?: string[]
  created_at: string
  updated_at: string
}

// 预警配置模型
export interface AlertConfig {
  id: number
  name: string
  description?: string
  severity_filter?: string[]
  source_filter?: string[]
  notification_channels: {
    email?: boolean
    webhook?: boolean
    dingtalk?: boolean
    sms?: boolean
  }
  enabled: boolean
  created_at: string
  updated_at: string
}

// 情报源响应
export interface IntelligenceSourcesResponse {
  code: number
  msg: string
  data?: IntelligenceSource[]
}

// 检测规则响应
export interface DetectionRulesResponse {
  code: number
  msg: string
  data?: DetectionRule[]
}

// 预警配置响应
export interface AlertConfigsResponse {
  code: number
  msg: string
  data?: AlertConfig[]
}

// 情报源更新请求
export interface IntelligenceSourceUpdateRequest {
  name?: string
  url?: string
  type?: string
  enabled?: boolean
  sync_interval?: number
  config?: Record<string, any>
}

// 检测规则创建/更新请求
export interface DetectionRuleRequest {
  name: string
  description?: string
  type: string
  condition: string
  action: string
  severity: string
  enabled: boolean
  priority?: number
  tags?: string[]
}

// 同步情报源响应
export interface SyncSourceResponse {
  code: number
  msg: string
  data?: {
    job_id: string
    status: string
    message: string
  }
}

// 获取情报源列表
export function getIntelligenceSources() {
  return request.get<IntelligenceSourcesResponse>('/api/config/intelligence-sources')
}

// 更新情报源
export function updateIntelligenceSource(id: number, data: IntelligenceSourceUpdateRequest) {
  return request.put<{ code: number; msg: string }>(`/api/config/intelligence-sources/${id}`, data)
}

// 同步情报源
export function syncIntelligenceSource(id: number) {
  return request.post<SyncSourceResponse>(`/api/config/intelligence-sources/${id}/sync`)
}

// 获取检测规则列表
export function getDetectionRules() {
  return request.get<DetectionRulesResponse>('/api/config/detection-rules')
}

// 创建检测规则
export function createDetectionRule(data: DetectionRuleRequest) {
  return request.post<{ code: number; msg: string; data?: DetectionRule }>('/api/config/detection-rules', data)
}

// 更新检测规则
export function updateDetectionRule(id: number, data: DetectionRuleRequest) {
  return request.put<{ code: number; msg: string; data?: DetectionRule }>(`/api/config/detection-rules/${id}`, data)
}

// 删除检测规则
export function deleteDetectionRule(id: number) {
  return request.delete<{ code: number; msg: string }>(`/api/config/detection-rules/${id}`)
}

// 获取预警配置列表
export function getAlertConfigs() {
  return request.get<AlertConfigsResponse>('/api/config/alert-configs')
}