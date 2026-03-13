import request from '@/utils/request'

// 漏洞情报模型
export interface VulnerabilityIntelligence {
  id: number
  cve_id: string
  cnvd_id?: string
  cnnvd_id?: string
  title: string
  description?: string
  severity: string
  cvss_score?: number
  cvss_vector?: string
  affected_products?: string[]
  affected_versions?: string[]
  vuln_references?: Array<{
    title?: string
    url: string
    source?: string
  }>
  exploit_available: boolean
  poc_available: boolean
  published_at?: string
  last_modified?: string
  source: string
  is_0day: boolean
  created_at: string
  updated_at: string
}

// 漏洞情报查询参数
export interface VulnerabilityIntelligenceQuery {
  page?: number
  pageSize?: number
  search?: string
  severity?: string
  source?: string
  cveId?: string
  startDate?: string
  endDate?: string
  is0Day?: boolean
  exploitAvailable?: boolean
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}

// 漏洞情报响应数据
export interface VulnerabilityIntelligenceResponse {
  code: number
  msg: string
  data?: {
    total: number
    page: number
    pageSize: number
    items: VulnerabilityIntelligence[]
  }
}

// 漏洞情报详情响应
export interface VulnerabilityIntelligenceDetailResponse {
  code: number
  msg: string
  data?: VulnerabilityIntelligence
}

// 漏洞情报统计响应
export interface VulnerabilityStatsResponse {
  code: number
  msg: string
  data?: {
    total: number
    bySeverity: Record<string, number>
    bySource: Record<string, number>
    byDay: Array<{
      date: string
      count: number
    }>
    zeroDayCount: number
    exploitCount: number
    pocCount: number
  }
}

// 漏洞情报同步响应
export interface SyncResponse {
  code: number
  msg: string
  data?: {
    syncId: string
    status: string
    message: string
  }
}

// 同步状态响应
export interface SyncStatusResponse {
  code: number
  msg: string
  data?: {
    syncId: string
    status: string
    progress: number
    total: number
    processed: number
    success: number
    failed: number
    startTime: string
    endTime?: string
    error?: string
  }
}

// 获取漏洞情报列表
export function getVulnerabilityIntelligenceList(params: VulnerabilityIntelligenceQuery) {
  return request.get<VulnerabilityIntelligenceResponse>('/api/vulnerability/intelligence', {
    params,
  })
}

// 获取漏洞情报详情
export function getVulnerabilityIntelligenceDetail(id: number) {
  return request.get<VulnerabilityIntelligenceDetailResponse>(
    `/api/vulnerability/intelligence/${id}`,
  )
}

// 同步漏洞情报
export function syncVulnerabilityIntelligence() {
  return request.post<SyncResponse>('/api/vulnerability/intelligence/sync')
}

// 获取同步状态
export function getSyncStatus(syncId: string) {
  return request.get<SyncStatusResponse>(`/api/vulnerability/intelligence/sync/${syncId}`)
}

// 获取漏洞统计
export function getVulnerabilityStats() {
  return request.get<VulnerabilityStatsResponse>('/api/vulnerability/intelligence/stats')
}

// 搜索漏洞情报
export function searchVulnerabilityIntelligence(keyword: string) {
  return request.get<VulnerabilityIntelligenceResponse>('/api/vulnerability/intelligence/search', {
    params: { keyword },
  })
}