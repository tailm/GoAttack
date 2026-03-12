<template>
  <div class="alerts-page">
    <a-card :bordered="false">
      <template #title>
        <a-space>
          <icon-notification />
          <span>{{ $t('menu.intelligence.alerts') }}</span>
        </a-space>
      </template>
      
      <!-- 操作按钮区域 -->
      <div class="action-area">
        <a-space>
          <a-button type="primary" @click="handleCreateAlert">
            <template #icon>
              <icon-plus />
            </template>
            新建预警
          </a-button>
          <a-button @click="handleMarkAllRead" :disabled="unreadCount === 0">
            <template #icon>
              <icon-check-circle />
            </template>
            全部标记已读
          </a-button>
          <a-button @click="handleRefresh" :loading="loading">
            <template #icon>
              <icon-refresh />
            </template>
            刷新
          </a-button>
          <a-badge :count="unreadCount" :offset="[10, -5]">
            <a-button @click="handleFilterUnread">
              <template #icon>
                <icon-eye-invisible />
              </template>
              未读预警
            </a-button>
          </a-badge>
        </a-space>
      </div>

      <!-- 搜索和过滤区域 -->
      <div class="search-area">
        <a-form :model="searchForm" layout="inline">
          <a-form-item field="search" :label="$t('alert.search.label')">
            <a-input
              v-model="searchForm.search"
              :placeholder="$t('alert.search.placeholder')"
              allow-clear
              @change="handleSearch"
            />
          </a-form-item>
          <a-form-item field="severity" :label="$t('alert.severity.label')">
            <a-select
              v-model="searchForm.severity"
              :placeholder="$t('alert.severity.placeholder')"
              allow-clear
              @change="handleSearch"
            >
              <a-option value="critical">{{ $t('alert.severity.critical') }}</a-option>
              <a-option value="high">{{ $t('alert.severity.high') }}</a-option>
              <a-option value="medium">{{ $t('alert.severity.medium') }}</a-option>
              <a-option value="low">{{ $t('alert.severity.low') }}</a-option>
            </a-select>
          </a-form-item>
          <a-form-item field="status" :label="$t('alert.status.label')">
            <a-select
              v-model="searchForm.status"
              :placeholder="$t('alert.status.placeholder')"
              allow-clear
              @change="handleSearch"
            >
              <a-option value="unread">{{ $t('alert.status.unread') }}</a-option>
              <a-option value="read">{{ $t('alert.status.read') }}</a-option>
              <a-option value="resolved">{{ $t('alert.status.resolved') }}</a-option>
            </a-select>
          </a-form-item>
          <a-form-item field="type" :label="$t('alert.type.label')">
            <a-select
              v-model="searchForm.type"
              :placeholder="$t('alert.type.placeholder')"
              allow-clear
              @change="handleSearch"
            >
              <a-option value="vulnerability">{{ $t('alert.type.vulnerability') }}</a-option>
              <a-option value="compliance">{{ $t('alert.type.compliance') }}</a-option>
              <a-option value="baseline">{{ $t('alert.type.baseline') }}</a-option>
              <a-option value="system">{{ $t('alert.type.system') }}</a-option>
            </a-select>
          </a-form-item>
          <a-form-item>
            <a-button type="primary" @click="handleSearch">
              <template #icon>
                <icon-search />
              </template>
              {{ $t('alert.search.button') }}
            </a-button>
            <a-button @click="handleReset">
              <template #icon>
                <icon-refresh />
              </template>
              {{ $t('alert.reset.button') }}
            </a-button>
          </a-form-item>
        </a-form>
      </div>

      <!-- 数据表格 -->
      <a-table
        :data="tableData"
        :loading="loading"
        :pagination="pagination"
        @page-change="handlePageChange"
        @page-size-change="handlePageSizeChange"
        row-key="id"
        :bordered="false"
        :row-class="getRowClass"
      >
        <template #columns>
          <a-table-column title="状态" data-index="status" :width="80">
            <template #cell="{ record }">
              <a-badge v-if="record.status === 'unread'" status="processing" dot />
              <a-badge v-else-if="record.status === 'read'" status="success" />
              <a-badge v-else status="default" />
            </template>
          </a-table-column>
          <a-table-column title="标题" data-index="title" :ellipsis="true" :tooltip="true">
            <template #cell="{ record }">
              <a-space>
                <span :class="{ 'unread-title': record.status === 'unread' }">{{ record.title }}</span>
                <a-tag v-if="record.status === 'unread'" size="small" color="red">新</a-tag>
              </a-space>
            </template>
          </a-table-column>
          <a-table-column title="严重程度" data-index="severity" :width="100">
            <template #cell="{ record }">
              <a-tag :color="getSeverityColor(record.severity)">
                {{ getSeverityText(record.severity) }}
              </a-tag>
            </template>
          </a-table-column>
          <a-table-column title="类型" data-index="type" :width="100">
            <template #cell="{ record }">
              <a-tag>{{ getTypeText(record.type) }}</a-tag>
            </template>
          </a-table-column>
          <a-table-column title="来源" data-index="source" :width="120">
            <template #cell="{ record }">
              <a-tag>{{ record.source || '-' }}</a-tag>
            </template>
          </a-table-column>
          <a-table-column title="创建时间" data-index="created_at" :width="160">
            <template #cell="{ record }">
              {{ formatDate(record.created_at) }}
            </template>
          </a-table-column>
          <a-table-column title="操作" :width="150" fixed="right">
            <template #cell="{ record }">
              <a-space>
                <a-button 
                  type="text" 
                  size="small" 
                  @click="handleViewDetail(record)"
                >
                  <template #icon>
                    <icon-eye />
                  </template>
                </a-button>
                <a-button 
                  v-if="record.status === 'unread'"
                  type="text" 
                  size="small" 
                  @click="handleMarkAsRead(record)"
                >
                  <template #icon>
                    <icon-check />
                  </template>
                </a-button>
                <a-button 
                  v-if="record.status !== 'resolved'"
                  type="text" 
                  size="small" 
                  @click="handleResolve(record)"
                >
                  <template #icon>
                    <icon-check-circle />
                  </template>
                </a-button>
                <a-button 
                  type="text" 
                  size="small" 
                  status="danger"
                  @click="handleDelete(record)"
                >
                  <template #icon>
                    <icon-delete />
                  </template>
                </a-button>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </a-card>

    <!-- 预警详情对话框 -->
    <a-modal
      v-model:visible="detailVisible"
      :title="detailTitle"
      width="800px"
      :footer="false"
      @cancel="handleCloseDetail"
    >
      <div v-if="currentRecord">
        <a-descriptions :column="2" bordered>
          <a-descriptions-item label="预警ID">{{ currentRecord.id }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="getStatusColor(currentRecord.status)">
              {{ getStatusText(currentRecord.status) }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="标题">{{ currentRecord.title }}</a-descriptions-item>
          <a-descriptions-item label="严重程度">
            <a-tag :color="getSeverityColor(currentRecord.severity)">
              {{ getSeverityText(currentRecord.severity) }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="类型">{{ getTypeText(currentRecord.type) }}</a-descriptions-item>
          <a-descriptions-item label="来源">{{ currentRecord.source || '-' }}</a-descriptions-item>
          <a-descriptions-item label="创建时间">{{ formatDate(currentRecord.created_at) }}</a-descriptions-item>
          <a-descriptions-item label="阅读时间">
            {{ currentRecord.read_at ? formatDate(currentRecord.read_at) : '未读' }}
          </a-descriptions-item>
          <a-descriptions-item label="解决时间">
            {{ currentRecord.resolved_at ? formatDate(currentRecord.resolved_at) : '未解决' }}
          </a-descriptions-item>
        </a-descriptions>
        
        <a-divider />
        
        <a-typography-title :heading="6">预警描述</a-typography-title>
        <a-typography-paragraph>
          {{ currentRecord.description || '暂无描述' }}
        </a-typography-paragraph>
        
        <a-divider />
        
        <a-typography-title :heading="6">预警数据</a-typography-title>
        <pre v-if="currentRecord.data" style="background: #f5f5f5; padding: 12px; border-radius: 4px; overflow: auto;">
{{ JSON.stringify(currentRecord.data, null, 2) }}
        </pre>
        <a-typography-paragraph v-else>暂无数据</a-typography-paragraph>
        
        <a-divider />
        
        <div class="action-buttons" v-if="currentRecord.status !== 'resolved'">
          <a-space>
            <a-button 
              v-if="currentRecord.status === 'unread'"
              type="primary" 
              @click="handleMarkCurrentAsRead"
            >
              <template #icon>
                <icon-check />
              </template>
              标记为已读
            </a-button>
            <a-button 
              type="primary" 
              status="success"
              @click="handleResolveCurrent"
            >
              <template #icon>
                <icon-check-circle />
              </template>
              标记为已解决
            </a-button>
          </a-space>
        </div>
      </div>
    </a-modal>

    <!-- 创建预警对话框 -->
    <a-modal
      v-model:visible="createVisible"
      title="新建预警"
      width="600px"
      @ok="handleCreateConfirm"
      @cancel="handleCreateCancel"
    >
      <a-form :model="createForm" :label-col-props="{ span: 6 }" :wrapper-col-props="{ span: 18 }">
        <a-form-item field="title" label="预警标题" required>
          <a-input v-model="createForm.title" placeholder="请输入预警标题" />
        </a-form-item>
        <a-form-item field="description" label="预警描述" required>
          <a-textarea v-model="createForm.description" placeholder="请输入预警描述" :rows="3" />
        </a-form-item>
        <a-form-item field="severity" label="严重程度" required>
          <a-select v-model="createForm.severity" placeholder="请选择严重程度">
            <a-option value="critical">严重</a-option>
            <a-option value="high">高危</a-option>
            <a-option value="medium">中危</a-option>
            <a-option value="low">低危</a-option>
          </a-select>
        </a-form-item>
        <a-form-item field="type" label="预警类型" required>
          <a-select v-model="createForm.type" placeholder="请选择预警类型">
            <a-option value="vulnerability">漏洞预警</a-option>
            <a-option value="compliance">合规预警</a-option>
            <a-option value="baseline">基线预警</a-option>
            <a-option value="system">系统预警</a-option>
          </a-select>
        </a-form-item>
        <a-form-item field="source" label="预警来源">
          <a-input v-model="createForm.source" placeholder="请输入预警来源" />
        </a-form-item>
        <a-form-item field="data" label="预警数据">
          <a-textarea 
            v-model="createForm.data" 
            placeholder="请输入JSON格式的预警数据" 
            :rows="4" 
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import { 
  getAlertList,
  createAlert,
  updateAlertStatus,
  deleteAlert,
  getAlertStats,
  type Alert,
  type AlertQuery
} from '@/api/alert'

// 搜索表单
const searchForm = reactive({
  search: '',
  severity: '',
  status: '',
  type: '',
  startDate: '',
  endDate: ''
})

// 表格数据
const tableData = ref<Alert[]>([])
const loading = ref(false)
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showTotal: true,
  showPageSize: true,
  pageSizeOptions: [10, 20, 50, 100]
})

// 统计信息
const stats = ref({
  total: 0,
  unread: 0,
  critical: 0,
  high: 0,
  medium: 0,
  low: 0
})

// 创建预警表单
const createVisible = ref(false)
const createForm = reactive({
  title: '',
  description: '',
  severity: 'medium',
  type: 'vulnerability',
  source: '',
  data: '{}'
})

// 详情对话框
const detailVisible = ref(false)
const currentRecord = ref<Alert | null>(null)
const detailTitle = ref('预警详情')

// 计算未读数量
const unreadCount = computed(() => {
  return tableData.value.filter(item => item.status === 'unread').length
})

// 获取数据
const fetchData = async () => {
  loading.value = true
  try {
    const query: AlertQuery = {
      page: pagination.current,
      pageSize: pagination.pageSize,
      search: searchForm.search || undefined,
      severity: searchForm.severity || undefined,
      status: searchForm.status || undefined,
      type: searchForm.type || undefined,
      startDate: searchForm.startDate || undefined,
      endDate: searchForm.endDate || undefined
    }
    
    const response = await getAlertList(query)
    if (response.code === 20000) {
      tableData.value = response.data?.items || []
      pagination.total = response.data?.total || 0
    } else {
      Message.error(response.msg || '获取数据失败')
    }
  } catch (error) {
    Message.error('获取数据失败')
    console.error('获取预警列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 获取统计信息
const fetchStats = async () => {
  try {
    const response = await getAlertStats()
    if (response.code === 20000 && response.data) {
      stats.value = response.data
    }
  } catch (error) {
    console.error('获取预警统计失败:', error)
  }
}

// 搜索处理
const handleSearch = () => {
  pagination.current = 1
  fetchData()
}

// 重置搜索
const handleReset = () => {
  searchForm.search = ''
  searchForm.severity = ''
  searchForm.status = ''
  searchForm.type = ''
  searchForm.startDate = ''
  searchForm.endDate = ''
  pagination.current = 1
  fetchData()
}

// 刷新数据
const handleRefresh = () => {
  fetchData()
  fetchStats()
}

// 过滤未读预警
const handleFilterUnread = () => {
  searchForm.status = 'unread'
  handleSearch()
}

// 分页处理
const handlePageChange = (page: number) => {
  pagination.current = page
  fetchData()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.pageSize = pageSize
  pagination.current = 1
  fetchData()
}

// 创建预警
const handleCreateAlert = () => {
  createVisible.value = true
}

const handleCreateConfirm = async () => {
  if (!createForm.title.trim()) {
    Message.error('请输入预警标题')
    return
  }
  
  if (!createForm.description.trim()) {
    Message.error('请输入预警描述')
    return
  }
  
  try {
    const data = createForm.data ? JSON.parse(createForm.data) : {}
    
    const response = await createAlert({
      title: createForm.title,
      description: createForm.description,
      severity: createForm.severity,
      type: createForm.type,
      source: createForm.source || undefined,
      data
    })
    
    if (response.code === 20000) {
      Message.success('预警创建成功')
      createVisible.value = false
      resetCreateForm()
      fetchData()
      fetchStats()
    } else {
      Message.error(response.msg || '预警创建失败')
    }
  } catch (error) {
    Message.error('预警创建失败')
    console.error('创建预警失败:', error)
  }
}

const handleCreateCancel = () => {
  createVisible.value = false
  resetCreateForm()
}

const resetCreateForm = () => {
  createForm.title = ''
  createForm.description = ''
  createForm.severity = 'medium'
  createForm.type = 'vulnerability'
  createForm.source = ''
  createForm.data = '{}'
}

// 查看详情
const handleViewDetail = (record: Alert) => {
  currentRecord.value = record
  detailTitle.value = `预警详情 - ${record.title}`
  detailVisible.value = true
  
  // 如果是未读状态，自动标记为已读
  if (record.status === 'unread') {
    handleMarkAsRead(record)
  }
}

// 关闭详情
const handleCloseDetail = () => {
  detailVisible.value = false
  currentRecord.value = null
}

// 标记为已读
const handleMarkAsRead = async (record: Alert) => {
  try {
    const response = await updateAlertStatus(record.id, 'read')
    if (response.code === 20000) {
      record.status = 'read'
      record.read_at = new Date().toISOString()
      Message.success('已标记为已读')
      fetchStats()
    } else {
      Message.error(response.msg || '标记失败')
    }
  } catch (error) {
    Message.error('标记失败')
    console.error('标记预警为已读失败:', error)
  }
}

// 标记当前为已读
const handleMarkCurrentAsRead = async () => {
  if (currentRecord.value) {
    await handleMarkAsRead(currentRecord.value)
  }
}

// 标记为已解决
const handleResolve = async (record: Alert) => {
  try {
    const response = await updateAlertStatus(record.id, 'resolved')
    if (response.code === 20000) {
      record.status = 'resolved'
      record.resolved_at = new Date().toISOString()
      Message.success('已标记为已解决')
      fetchStats()
    } else {
      Message.error(response.msg || '标记失败')
    }
  } catch (error) {
    Message.error('标记失败')
    console.error('标记预警为已解决失败:', error)
  }
}

// 标记当前为已解决
const handleResolveCurrent = async () => {
  if (currentRecord.value) {
    await handleResolve(currentRecord.value)
  }
}

// 全部标记为已读
const handleMarkAllRead = async () => {
  Modal.confirm({
    title: '确认操作',
    content: `确定要将所有未读预警标记为已读吗？`,
    okText: '确认',
    cancelText: '取消',
    onOk: async () => {
      try {
        // 这里需要调用批量标记已读的API
        // 暂时先标记当前页的所有未读预警
        const unreadAlerts = tableData.value.filter(item => item.status === 'unread')
        for (const alert of unreadAlerts) {
          await updateAlertStatus(alert.id, 'read')
          alert.status = 'read'
          alert.read_at = new Date().toISOString()
        }
        Message.success(`已标记 ${unreadAlerts.length} 条预警为已读`)
        fetchStats()
      } catch (error) {
        Message.error('标记失败')
        console.error('批量标记预警为已读失败:', error)
      }
    }
  })
}

// 删除预警
const handleDelete = (record: Alert) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除预警 "${record.title}" 吗？此操作不可恢复。`,
    okText: '删除',
    cancelText: '取消',
    okButtonProps: {
      status: 'danger'
    },
    onOk: async () => {
      try {
        const response = await deleteAlert(record.id)
        if (response.code === 20000) {
          Message.success('预警删除成功')
          fetchData()
          fetchStats()
        } else {
          Message.error(response.msg || '删除失败')
        }
      } catch (error) {
        Message.error('删除失败')
        console.error('删除预警失败:', error)
      }
    }
  })
}

// 工具函数
const getSeverityColor = (severity: string) => {
  switch (severity?.toLowerCase()) {
    case 'critical': return 'red'
    case 'high': return 'orange'
    case 'medium': return 'yellow'
    case 'low': return 'green'
    default: return 'gray'
  }
}

const getSeverityText = (severity: string) => {
  switch (severity?.toLowerCase()) {
    case 'critical': return '严重'
    case 'high': return '高危'
    case 'medium': return '中危'
    case 'low': return '低危'
    default: return '未知'
  }
}

const getTypeText = (type: string) => {
  switch (type) {
    case 'vulnerability': return '漏洞预警'
    case 'compliance': return '合规预警'
    case 'baseline': return '基线预警'
    case 'system': return '系统预警'
    default: return type
  }
}

const getStatusColor = (status: string) => {
  switch (status) {
    case 'unread': return 'red'
    case 'read': return 'green'
    case 'resolved': return 'blue'
    default: return 'gray'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'unread': return '未读'
    case 'read': return '已读'
    case 'resolved': return '已解决'
    default: return status
  }
}

const getRowClass = (record: Alert) => {
  return record.status === 'unread' ? 'unread-row' : ''
}

const formatDate = (dateString?: string) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN')
}

// 初始化
onMounted(() => {
  fetchData()
  fetchStats()
})
</script>

<style scoped>
.alerts-page {
  padding: 20px;
}

.action-area {
  margin-bottom: 20px;
}

.search-area {
  margin-bottom: 20px;
}

:deep(.arco-table-cell) {
  padding: 12px 8px;
}

:deep(.arco-descriptions-item-label) {
  font-weight: 600;
  width: 120px;
}

:deep(.unread-row) {
  background-color: #fff2f0;
}

.unread-title {
  font-weight: 600;
}

.action-buttons {
  margin-top: 20px;
  text-align: right;
}
</style>