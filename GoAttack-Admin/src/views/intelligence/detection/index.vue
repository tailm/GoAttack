<template>
  <div class="detection-tasks-page">
    <a-card :bordered="false">
      <template #title>
        <a-space>
          <icon-play-circle />
          <span>{{ $t('menu.intelligence.detectionTasks') }}</span>
        </a-space>
      </template>

      <!-- 操作按钮区域 -->
      <div class="action-area">
        <a-space>
          <a-button type="primary" @click="handleCreateTask">
            <template #icon>
              <icon-plus />
            </template>
            新建检测任务
          </a-button>
          <a-button @click="handleRefresh" :loading="loading">
            <template #icon>
              <icon-refresh />
            </template>
            刷新
          </a-button>
        </a-space>
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
      >
        <template #columns>
          <a-table-column title="任务名称" data-index="name" :width="200">
            <template #cell="{ record }">
              <a-link @click="handleViewDetail(record)">{{ record.name }}</a-link>
            </template>
          </a-table-column>
          <a-table-column title="检测类型" data-index="detection_type" :width="120">
            <template #cell="{ record }">
              <a-tag :color="getTypeColor(record.detection_type)">
                {{ getTypeText(record.detection_type) }}
              </a-tag>
            </template>
          </a-table-column>
          <a-table-column title="状态" data-index="status" :width="100">
            <template #cell="{ record }">
              <a-tag :color="getStatusColor(record.status)">
                {{ getStatusText(record.status) }}
              </a-tag>
            </template>
          </a-table-column>
          <a-table-column title="目标数量" data-index="target_count" :width="100">
            <template #cell="{ record }">
              {{ record.target_count || 0 }}
            </template>
          </a-table-column>
          <a-table-column title="漏洞数量" data-index="vulnerability_count" :width="100">
            <template #cell="{ record }">
              <a-tag v-if="record.vulnerability_count > 0" color="red">
                {{ record.vulnerability_count }}
              </a-tag>
              <span v-else>0</span>
            </template>
          </a-table-column>
          <a-table-column title="风险等级" data-index="risk_level" :width="100">
            <template #cell="{ record }">
              <a-tag v-if="record.risk_level" :color="getRiskColor(record.risk_level)">
                {{ getRiskText(record.risk_level) }}
              </a-tag>
              <span v-else>-</span>
            </template>
          </a-table-column>
          <a-table-column title="创建时间" data-index="created_at" :width="160">
            <template #cell="{ record }">
              {{ formatDate(record.created_at) }}
            </template>
          </a-table-column>
          <a-table-column title="最后运行" data-index="last_run" :width="160">
            <template #cell="{ record }">
              {{ record.last_run ? formatDate(record.last_run) : '-' }}
            </template>
          </a-table-column>
          <a-table-column title="操作" :width="180" fixed="right">
            <template #cell="{ record }">
              <a-space>
                <a-button type="text" size="small" @click="handleViewDetail(record)">
                  <template #icon>
                    <icon-eye />
                  </template>
                </a-button>
                <a-button type="text" size="small" @click="handleStartTask(record)" :disabled="record.status === 'running'">
                  <template #icon>
                    <icon-play-circle />
                  </template>
                </a-button>
                <a-button type="text" size="small" @click="handleStopTask(record)" :disabled="record.status !== 'running'">
                  <template #icon>
                    <icon-pause />
                  </template>
                </a-button>
                <a-button type="text" size="small" status="danger" @click="handleDeleteTask(record)">
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

    <!-- 创建任务对话框 -->
    <a-modal v-model:visible="createVisible" title="新建检测任务" width="600px" @ok="handleCreateConfirm" @cancel="handleCreateCancel">
      <a-form :model="createForm" :label-col-props="{ span: 6 }" :wrapper-col-props="{ span: 18 }">
        <a-form-item field="name" label="任务名称" required>
          <a-input v-model="createForm.name" placeholder="请输入任务名称" />
        </a-form-item>
        <a-form-item field="description" label="任务描述">
          <a-textarea v-model="createForm.description" placeholder="请输入任务描述" :rows="3" />
        </a-form-item>
        <a-form-item field="detectionType" label="检测类型" required>
          <a-select v-model="createForm.detectionType" placeholder="请选择检测类型">
            <a-option value="vulnerability">漏洞检测</a-option>
            <a-option value="compliance">合规检测</a-option>
            <a-option value="baseline">基线检测</a-option>
            <a-option value="custom">自定义检测</a-option>
          </a-select>
        </a-form-item>
        <a-form-item field="targets" label="检测目标" required>
          <a-textarea v-model="createForm.targets" placeholder="请输入检测目标，每行一个（支持IP、域名、CIDR）" :rows="4" />
        </a-form-item>
        <a-form-item field="config" label="检测配置">
          <a-textarea v-model="createForm.config" placeholder="请输入JSON格式的检测配置" :rows="4" />
        </a-form-item>
        <a-form-item field="schedule" label="调度计划">
          <a-input v-model="createForm.schedule" placeholder="Cron表达式，如：0 0 * * *（每天0点）" />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 任务详情对话框 -->
    <a-modal v-model:visible="detailVisible" :title="detailTitle" width="800px" :footer="false" @cancel="handleCloseDetail">
      <div v-if="currentRecord">
        <a-descriptions :column="2" bordered>
          <a-descriptions-item label="任务ID">{{ currentRecord.id }}</a-descriptions-item>
          <a-descriptions-item label="任务名称">{{ currentRecord.name }}</a-descriptions-item>
          <a-descriptions-item label="检测类型">
            <a-tag :color="getTypeColor(currentRecord.detection_type)">
              {{ getTypeText(currentRecord.detection_type) }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="getStatusColor(currentRecord.status)">
              {{ getStatusText(currentRecord.status) }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="风险等级">
            <a-tag v-if="currentRecord.risk_level" :color="getRiskColor(currentRecord.risk_level)">
              {{ getRiskText(currentRecord.risk_level) }}
            </a-tag>
            <span v-else>-</span>
          </a-descriptions-item>
          <a-descriptions-item label="目标数量">{{ currentRecord.target_count || 0 }}</a-descriptions-item>
          <a-descriptions-item label="漏洞数量">
            <a-tag v-if="currentRecord.vulnerability_count > 0" color="red">
              {{ currentRecord.vulnerability_count }}
            </a-tag>
            <span v-else>0</span>
          </a-descriptions-item>
          <a-descriptions-item label="创建时间">{{ formatDate(currentRecord.created_at) }}</a-descriptions-item>
          <a-descriptions-item label="最后运行">
            {{ currentRecord.last_run ? formatDate(currentRecord.last_run) : '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="下次运行">
            {{ currentRecord.next_run ? formatDate(currentRecord.next_run) : '-' }}
          </a-descriptions-item>
        </a-descriptions>

        <a-divider />

        <a-typography-title :heading="6">任务描述</a-typography-title>
        <a-typography-paragraph>
          {{ currentRecord.description || '暂无描述' }}
        </a-typography-paragraph>

        <a-divider />

        <a-typography-title :heading="6">检测配置</a-typography-title>
        <pre v-if="currentRecord.config" style="background: #f5f5f5; padding: 12px; border-radius: 4px; overflow: auto"
          >{{ JSON.stringify(currentRecord.config, null, 2) }}
        </pre>
        <a-typography-paragraph v-else>暂无配置信息</a-typography-paragraph>

        <a-divider />

        <a-typography-title :heading="6">调度计划</a-typography-title>
        <pre v-if="currentRecord.schedule" style="background: #f5f5f5; padding: 12px; border-radius: 4px; overflow: auto"
          >{{ JSON.stringify(currentRecord.schedule, null, 2) }}
        </pre>
        <a-typography-paragraph v-else>暂无调度计划</a-typography-paragraph>
      </div>
    </a-modal>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import {
  getDetectionTaskList,
  createDetectionTask,
  updateDetectionTaskStatus,
  deleteDetectionTask,
  type DetectionTask,
  type DetectionTaskQuery,
} from '@/api/detection'

// 表格数据
const tableData = ref<DetectionTask[]>([])
const loading = ref(false)
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showTotal: true,
  showPageSize: true,
  pageSizeOptions: [10, 20, 50, 100],
})

// 创建任务表单
const createVisible = ref(false)
const createForm = reactive({
  name: '',
  description: '',
  detectionType: 'vulnerability',
  targets: '',
  config: '{}',
  schedule: '',
})

// 详情对话框
const detailVisible = ref(false)
const currentRecord = ref<DetectionTask | null>(null)
const detailTitle = ref('任务详情')

// 获取数据
const fetchData = async () => {
  loading.value = true
  try {
    const query: DetectionTaskQuery = {
      page: pagination.current,
      pageSize: pagination.pageSize,
    }

    const response = await getDetectionTaskList(query)
    if (response.code === 20000) {
      tableData.value = response.data?.items || []
      pagination.total = response.data?.total || 0
    } else {
      Message.error(response.msg || '获取数据失败')
    }
  } catch (error) {
    Message.error('获取数据失败')
    // 获取检测任务失败
  } finally {
    loading.value = false
  }
}

// 刷新数据
const handleRefresh = () => {
  fetchData()
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

// 创建任务
const resetCreateForm = () => {
  createForm.name = ''
  createForm.description = ''
  createForm.detectionType = 'vulnerability'
  createForm.targets = ''
  createForm.config = '{}'
  createForm.schedule = ''
}

const handleCreateTask = () => {
  createVisible.value = true
}

const handleCreateConfirm = async () => {
  if (!createForm.name.trim()) {
    Message.error('请输入任务名称')
    return
  }

  if (!createForm.targets.trim()) {
    Message.error('请输入检测目标')
    return
  }

  try {
    const targets = createForm.targets.split('\n').filter((t) => t.trim())
    const config = createForm.config ? JSON.parse(createForm.config) : {}
    const schedule = createForm.schedule ? { cron: createForm.schedule } : undefined

    const response = await createDetectionTask({
      name: createForm.name,
      description: createForm.description,
      detection_type: createForm.detectionType,
      targets,
      config,
      schedule,
    })

    if (response.code === 20000) {
      Message.success('任务创建成功')
      createVisible.value = false
      resetCreateForm()
      fetchData()
    } else {
      Message.error(response.msg || '任务创建失败')
    }
  } catch (error) {
    Message.error('任务创建失败')
    // 创建检测任务失败
  }
}

const handleCreateCancel = () => {
  createVisible.value = false
  resetCreateForm()
}

// 查看详情
const handleViewDetail = (record: DetectionTask) => {
  currentRecord.value = record
  detailTitle.value = `任务详情 - ${record.name}`
  detailVisible.value = true
}

// 关闭详情
const handleCloseDetail = () => {
  detailVisible.value = false
  currentRecord.value = null
}

// 启动任务
const handleStartTask = async (record: DetectionTask) => {
  try {
    const response = await updateDetectionTaskStatus(record.id, 'running')
    if (response.code === 20000) {
      Message.success('任务已启动')
      fetchData()
    } else {
      Message.error(response.msg || '启动任务失败')
    }
  } catch (error) {
    Message.error('启动任务失败')
    // 启动检测任务失败
  }
}

// 停止任务
const handleStopTask = async (record: DetectionTask) => {
  try {
    const response = await updateDetectionTaskStatus(record.id, 'stopped')
    if (response.code === 20000) {
      Message.success('任务已停止')
      fetchData()
    } else {
      Message.error(response.msg || '停止任务失败')
    }
  } catch (error) {
    Message.error('停止任务失败')
    // 停止检测任务失败
  }
}

// 删除任务
const handleDeleteTask = (record: DetectionTask) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除任务 "${record.name}" 吗？此操作不可恢复。`,
    okText: '删除',
    cancelText: '取消',
    okButtonProps: {
      status: 'danger',
    },
    onOk: async () => {
      try {
        const response = await deleteDetectionTask(record.id)
        if (response.code === 20000) {
          Message.success('任务删除成功')
          fetchData()
        } else {
          Message.error(response.msg || '删除任务失败')
        }
      } catch (error) {
        Message.error('删除任务失败')
        // 删除检测任务失败
      }
    },
  })
}

// 工具函数
const getTypeColor = (type: string) => {
  switch (type) {
    case 'vulnerability':
      return 'red'
    case 'compliance':
      return 'blue'
    case 'baseline':
      return 'green'
    case 'custom':
      return 'purple'
    default:
      return 'gray'
  }
}

const getTypeText = (type: string) => {
  switch (type) {
    case 'vulnerability':
      return '漏洞检测'
    case 'compliance':
      return '合规检测'
    case 'baseline':
      return '基线检测'
    case 'custom':
      return '自定义检测'
    default:
      return type
  }
}

const getStatusColor = (status: string) => {
  switch (status) {
    case 'pending':
      return 'orange'
    case 'running':
      return 'green'
    case 'completed':
      return 'blue'
    case 'failed':
      return 'red'
    case 'stopped':
      return 'gray'
    default:
      return 'gray'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'pending':
      return '等待中'
    case 'running':
      return '运行中'
    case 'completed':
      return '已完成'
    case 'failed':
      return '失败'
    case 'stopped':
      return '已停止'
    default:
      return status
  }
}

const getRiskColor = (risk: string) => {
  switch (risk?.toLowerCase()) {
    case 'critical':
      return 'red'
    case 'high':
      return 'orange'
    case 'medium':
      return 'yellow'
    case 'low':
      return 'green'
    default:
      return 'gray'
  }
}

const getRiskText = (risk: string) => {
  switch (risk?.toLowerCase()) {
    case 'critical':
      return '严重'
    case 'high':
      return '高危'
    case 'medium':
      return '中危'
    case 'low':
      return '低危'
    default:
      return risk
  }
}

const formatDate = (dateString?: string) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN')
}

// 初始化
onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.detection-tasks-page {
  padding: 20px;
}

.action-area {
  margin-bottom: 20px;
}

:deep(.arco-table-cell) {
  padding: 12px 8px;
}

:deep(.arco-descriptions-item-label) {
  font-weight: 600;
  width: 120px;
}
</style>
