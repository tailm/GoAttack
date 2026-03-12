<template>
  <div class="intelligence-config-page">
    <a-tabs default-active-key="sources">
      <!-- 漏洞情报源配置 -->
      <a-tab-pane key="sources" title="情报源配置">
        <a-card :bordered="false">
          <template #title>
            <a-space>
              <icon-link />
              <span>漏洞情报源配置</span>
            </a-space>
          </template>
          
          <div class="action-area">
            <a-button type="primary" @click="handleAddSource">
              <template #icon>
                <icon-plus />
              </template>
              添加情报源
            </a-button>
            <a-button @click="handleSyncAll" :loading="syncingAll">
              <template #icon>
                <icon-sync />
              </template>
              同步所有源
            </a-button>
          </div>

          <a-table
            :data="sourcesData"
            :loading="sourcesLoading"
            row-key="id"
            :bordered="false"
          >
            <template #columns>
              <a-table-column title="名称" data-index="name" :width="150">
                <template #cell="{ record }">
                  <a-space>
                    <icon-link />
                    <span>{{ record.name }}</span>
                  </a-space>
                </template>
              </a-table-column>
              <a-table-column title="类型" data-index="type" :width="100">
                <template #cell="{ record }">
                  <a-tag :color="getSourceTypeColor(record.type)">
                    {{ getSourceTypeText(record.type) }}
                  </a-tag>
                </template>
              </a-table-column>
              <a-table-column title="URL" data-index="url" :ellipsis="true" :tooltip="true">
                <template #cell="{ record }">
                  <a-link :href="record.url" target="_blank">{{ record.url }}</a-link>
                </template>
              </a-table-column>
              <a-table-column title="状态" data-index="enabled" :width="80">
                <template #cell="{ record }">
                  <a-switch 
                    v-model="record.enabled" 
                    :checked-value="true"
                    :unchecked-value="false"
                    @change="handleToggleSource(record)"
                  />
                </template>
              </a-table-column>
              <a-table-column title="同步间隔" data-index="sync_interval" :width="120">
                <template #cell="{ record }">
                  {{ formatInterval(record.sync_interval) }}
                </template>
              </a-table-column>
              <a-table-column title="最后同步" data-index="last_sync" :width="160">
                <template #cell="{ record }">
                  <div v-if="record.last_sync">
                    {{ formatDate(record.last_sync) }}
                    <a-tag v-if="record.last_sync_status === 'success'" size="small" color="green">成功</a-tag>
                    <a-tag v-else-if="record.last_sync_status === 'failed'" size="small" color="red">失败</a-tag>
                    <a-tag v-else size="small" color="gray">未知</a-tag>
                  </div>
                  <span v-else>-</span>
                </template>
              </a-table-column>
              <a-table-column title="操作" :width="150" fixed="right">
                <template #cell="{ record }">
                  <a-space>
                    <a-button type="text" size="small" @click="handleEditSource(record)">
                      <template #icon>
                        <icon-edit />
                      </template>
                    </a-button>
                    <a-button 
                      type="text" 
                      size="small" 
                      @click="handleSyncSource(record)"
                      :loading="record.id === syncingSourceId"
                    >
                      <template #icon>
                        <icon-sync />
                      </template>
                    </a-button>
                    <a-button 
                      type="text" 
                      size="small" 
                      status="danger"
                      @click="handleDeleteSource(record)"
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
      </a-tab-pane>

      <!-- 检测规则配置 -->
      <a-tab-pane key="rules" title="检测规则">
        <a-card :bordered="false">
          <template #title>
            <a-space>
              <icon-settings />
              <span>检测规则配置</span>
            </a-space>
          </template>
          
          <div class="action-area">
            <a-button type="primary" @click="handleAddRule">
              <template #icon>
                <icon-plus />
              </template>
              添加检测规则
            </a-button>
          </div>

          <a-table
            :data="rulesData"
            :loading="rulesLoading"
            row-key="id"
            :bordered="false"
          >
            <template #columns>
              <a-table-column title="规则名称" data-index="name" :width="200">
                <template #cell="{ record }">
                  <a-space>
                    <icon-settings />
                    <span>{{ record.name }}</span>
                  </a-space>
                </template>
              </a-table-column>
              <a-table-column title="描述" data-index="description" :ellipsis="true" :tooltip="true" />
              <a-table-column title="类型" data-index="type" :width="100">
                <template #cell="{ record }">
                  <a-tag>{{ getRuleTypeText(record.type) }}</a-tag>
                </template>
              </a-table-column>
              <a-table-column title="严重程度" data-index="severity" :width="100">
                <template #cell="{ record }">
                  <a-tag :color="getSeverityColor(record.severity)">
                    {{ getSeverityText(record.severity) }}
                  </a-tag>
                </template>
              </a-table-column>
              <a-table-column title="动作" data-index="action" :width="100">
                <template #cell="{ record }">
                  <a-tag>{{ getActionText(record.action) }}</a-tag>
                </template>
              </a-table-column>
              <a-table-column title="状态" data-index="enabled" :width="80">
                <template #cell="{ record }">
                  <a-switch 
                    v-model="record.enabled" 
                    :checked-value="true"
                    :unchecked-value="false"
                    @change="handleToggleRule(record)"
                  />
                </template>
              </a-table-column>
              <a-table-column title="优先级" data-index="priority" :width="80">
                <template #cell="{ record }">
                  <a-tag>{{ record.priority }}</a-tag>
                </template>
              </a-table-column>
              <a-table-column title="操作" :width="120" fixed="right">
                <template #cell="{ record }">
                  <a-space>
                    <a-button type="text" size="small" @click="handleEditRule(record)">
                      <template #icon>
                        <icon-edit />
                      </template>
                    </a-button>
                    <a-button 
                      type="text" 
                      size="small" 
                      status="danger"
                      @click="handleDeleteRule(record)"
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
      </a-tab-pane>

      <!-- 预警配置 -->
      <a-tab-pane key="alerts" title="预警配置">
        <a-card :bordered="false">
          <template #title>
            <a-space>
              <icon-notification />
              <span>预警通知配置</span>
            </a-space>
          </template>
          
          <div class="action-area">
            <a-button type="primary" @click="handleAddAlertConfig">
              <template #icon>
                <icon-plus />
              </template>
              添加预警配置
            </a-button>
          </div>

          <a-table
            :data="alertConfigsData"
            :loading="alertConfigsLoading"
            row-key="id"
            :bordered="false"
          >
            <template #columns>
              <a-table-column title="配置名称" data-index="name" :width="150">
                <template #cell="{ record }">
                  <a-space>
                    <icon-notification />
                    <span>{{ record.name }}</span>
                  </a-space>
                </template>
              </a-table-column>
              <a-table-column title="描述" data-index="description" :ellipsis="true" :tooltip="true" />
              <a-table-column title="严重程度过滤" data-index="severity_filter" :width="150">
                <template #cell="{ record }">
                  <a-space v-if="record.severity_filter && record.severity_filter.length > 0">
                    <a-tag v-for="severity in record.severity_filter" :key="severity" size="small" :color="getSeverityColor(severity)">
                      {{ getSeverityText(severity) }}
                    </a-tag>
                  </a-space>
                  <span v-else>全部</span>
                </template>
              </a-table-column>
              <a-table-column title="来源过滤" data-index="source_filter" :width="150">
                <template #cell="{ record }">
                  <a-space v-if="record.source_filter && record.source_filter.length > 0">
                    <a-tag v-for="source in record.source_filter" :key="source" size="small">
                      {{ source }}
                    </a-tag>
                  </a-space>
                  <span v-else>全部</span>
                </template>
              </a-table-column>
              <a-table-column title="通知渠道" data-index="notification_channels" :width="200">
                <template #cell="{ record }">
                  <a-space v-if="record.notification_channels">
                    <a-tag v-if="record.notification_channels.email" size="small" color="blue">邮件</a-tag>
                    <a-tag v-if="record.notification_channels.webhook" size="small" color="green">Webhook</a-tag>
                    <a-tag v-if="record.notification_channels.dingtalk" size="small" color="orange">钉钉</a-tag>
                    <a-tag v-if="record.notification_channels.sms" size="small" color="purple">短信</a-tag>
                  </a-space>
                  <span v-else>-</span>
                </template>
              </a-table-column>
              <a-table-column title="状态" data-index="enabled" :width="80">
                <template #cell="{ record }">
                  <a-switch 
                    v-model="record.enabled" 
                    :checked-value="true"
                    :unchecked-value="false"
                    @change="handleToggleAlertConfig(record)"
                  />
                </template>
              </a-table-column>
              <a-table-column title="操作" :width="120" fixed="right">
                <template #cell="{ record }">
                  <a-space>
                    <a-button type="text" size="small" @click="handleEditAlertConfig(record)">
                      <template #icon>
                        <icon-edit />
                      </template>
                    </a-button>
                    <a-button 
                      type="text" 
                      size="small" 
                      status="danger"
                      @click="handleDeleteAlertConfig(record)"
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
      </a-tab-pane>
    </a-tabs>

    <!-- 情报源编辑对话框 -->
    <a-modal
      v-model:visible="sourceDialogVisible"
      :title="sourceDialogTitle"
      width="600px"
      @ok="handleSaveSource"
      @cancel="handleCancelSource"
    >
      <a-form :model="sourceForm" :label-col-props="{ span: 6 }" :wrapper-col-props="{ span: 18 }">
        <a-form-item field="name" label="名称" required>
          <a-input v-model="sourceForm.name" placeholder="请输入情报源名称" />
        </a-form-item>
        <a-form-item field="url" label="URL" required>
          <a-input v-model="sourceForm.url" placeholder="请输入API地址" />
        </a-form-item>
        <a-form-item field="type" label="类型" required>
          <a-select v-model="sourceForm.type" placeholder="请选择情报源类型">
            <a-option value="avd">阿里云漏洞库 (AVD)</a-option>
            <a-option value="nvd">国家漏洞数据库 (NVD)</a-option>
            <a-option value="cnnvd">国家信息安全漏洞库 (CNNVD)</a-option>
            <a-option value="cnvd">国家信息安全漏洞共享平台 (CNVD)</a-option>
            <a-option value="exploitdb">Exploit Database</a-option>
            <a-option value="securityfocus">SecurityFocus</a-option>
            <a-option value="custom">自定义</a-option>
          </a-select>
        </a-form-item>
        <a-form-item field="enabled" label="启用状态">
          <a-switch v-model="sourceForm.enabled" />
        </a-form-item>
        <a-form-item field="sync_interval" label="同步间隔(秒)">
          <a-input-number v-model="sourceForm.sync_interval" :min="300" :step="300" placeholder="请输入同步间隔" />
          <template #extra>最小值为300秒（5分钟）</template>
        </a-form-item>
        <a-form-item field="config" label="配置参数">
          <a-textarea 
            v-model="sourceForm.config" 
            placeholder="请输入JSON格式的配置参数" 
            :rows="4" 
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 检测规则编辑对话框 -->
    <a-modal
      v-model:visible="ruleDialogVisible"
      :title="ruleDialogTitle"
      width="600px"
      @ok="handleSaveRule"
      @cancel="handleCancelRule"
    >
      <a-form :model="ruleForm" :label-col-props="{ span: 6 }" :wrapper-col-props="{ span: 18 }">
        <a-form-item field="name" label="规则名称" required>
          <a-input v-model="ruleForm.name" placeholder="请输入规则名称" />
        </a-form-item>
        <a-form-item field="description" label="规则描述">
          <a-textarea v-model="ruleForm.description" placeholder="请输入规则描述" :rows="3" />
        </a-form-item>
        <a-form-item field="type" label="规则类型" required>
          <a-select v-model="ruleForm.type" placeholder="请选择规则类型">
            <a-option value="port">端口检测</a-option>
            <a-option value="service">服务检测</a-option>
            <a-option value="version">版本检测</a-option>
            <a-option value="cve">CVE漏洞检测</a-option>
            <a-option value="custom">自定义检测</a-option>
          </a-select>
        </a-form-item>
        <a-form-item field="condition" label="检测条件" required>
          <a-textarea v-model="ruleForm.condition" placeholder="请输入检测条件表达式" :rows="3" />
        </a-form-item>
        <a-form-item field="action" label="动作" required>
          <a-select v-model="ruleForm.action" placeholder="请选择动作">
            <a-option value="alert">预警</a-option>
            <a-option value="block">阻断</a-option>
            <a-option value="log">记录</a-option>
            <a-option value="report">报告</a-option>
          </a-select>
        </a-form-item>
        <a-form-item field="severity" label="严重程度" required>
          <a-select v-model="ruleForm.severity" placeholder="请选择严重程度">
            <a-option value="critical">严重</a-option>
            <a-option value="high">高危</a-option>
            <a-option value="medium">中危</a-option>
            <a-option value="low">低危</a-option>
          </a-select>
        </a-form-item>
        <a-form-item field="enabled" label="启用状态">
          <a-switch v-model="ruleForm.enabled" />
        </a-form-item>
        <a-form-item field="priority" label="优先级">
          <a-input-number v-model="ruleForm.priority" :min="0" :max="100" placeholder="请输入优先级" />
          <template #extra>数值越小优先级越高（0-100）</template>
        </a-form-item>
        <a-form-item field="tags" label="标签">
          <a-input-tag v-model="ruleForm.tags" placeholder="请输入标签，按回车添加" allow-clear />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import { 
  getIntelligenceSources,
  updateIntelligenceSource,
  getDetectionRules,
  createDetectionRule,
  updateDetectionRule,
  deleteDetectionRule,
  getAlertConfigs,
  syncIntelligenceSource,
  type IntelligenceSource,
  type DetectionRule,
  type AlertConfig
} from '@/api/config'

// 情报源数据
const sourcesData = ref<IntelligenceSource[]>([])
const sourcesLoading = ref(false)

// 检测规则数据
const rulesData = ref<DetectionRule[]>([])
const rulesLoading = ref(false)

// 预警配置数据
const alertConfigsData = ref<AlertConfig[]>([])
const alertConfigsLoading = ref(false)

// 同步状态
const syncingAll = ref(false)
const syncingSourceId = ref<number | null>(null)

// 情报源对话框
const sourceDialogVisible = ref(false)
const sourceDialogTitle = ref('添加情报源')
const sourceForm = reactive({
  id: 0,
  name: '',
  url: '',
  type: 'avd',
  enabled: true,
  sync_interval: 3600,
  config: '{}'
})

// 检测规则对话框
const ruleDialogVisible = ref(false)
const ruleDialogTitle = ref('添加检测规则')
const ruleForm = reactive({
  id: 0,
  name: '',
  description: '',
  type: 'cve',
  condition: '',
  action: 'alert',
  severity: 'medium',
  enabled: true,
  priority: 0,
  tags: [] as string[]
})

// 获取情报源数据
const fetchSources = async () => {
  sourcesLoading.value = true
  try {
    const response = await getIntelligenceSources()
    if (response.code === 20000) {
      sourcesData.value = response.data || []
    } else {
      Message.error(response.msg || '获取情报源失败')
    }
  } catch (error) {
    Message.error('获取情报源失败')
    console.error('获取情报源失败:', error)
  } finally {
    sourcesLoading.value = false
  }
}

// 获取检测规则数据
const fetchRules = async () => {
  rulesLoading.value = true
  try {
    const response = await getDetectionRules()
    if (response.code === 20000) {
      rulesData.value = response.data || []
    } else {
      Message.error(response.msg || '获取检测规则失败')
    }
  } catch (error) {
    Message.error('获取检测规则失败')
    console.error('获取检测规则失败:', error)
  } finally {
    rulesLoading.value = false
  }
}

// 获取预警配置数据
const fetchAlertConfigs = async () => {
  alertConfigsLoading.value = true
  try {
    const response = await getAlertConfigs()
    if (response.code === 20000) {
      alertConfigsData.value = response.data || []
    } else {
      Message.error(response.msg || '获取预警配置失败')
    }
  } catch (error) {
    Message.error('获取预警配置失败')
    console.error('获取预警配置失败:', error)
  } finally {
    alertConfigsLoading.value = false
  }
}

// 添加情报源
const handleAddSource = () => {
  sourceDialogTitle.value = '添加情报源'
  sourceForm.id = 0
  sourceForm.name = ''
  sourceForm.url = ''
  sourceForm.type = 'avd'
  sourceForm.enabled = true
  sourceForm.sync_interval = 3600
  sourceForm.config = '{}'
  sourceDialogVisible.value = true
}

// 编辑情报源
const handleEditSource = (record: IntelligenceSource) => {
  sourceDialogTitle.value = '编辑情报源'
  sourceForm.id = record.id
  sourceForm.name = record.name
  sourceForm.url = record.url
  sourceForm.type = record.type
  sourceForm.enabled = record.enabled
  sourceForm.sync_interval = record.sync_interval || 3600
  sourceForm.config = JSON.stringify(record.config || {}, null, 2)
  sourceDialogVisible.value = true
}

// 保存情报源
const handleSaveSource = async () => {
  if (!sourceForm.name.trim()) {
    Message.error('请输入情报源名称')
    return
  }
  
  if (!sourceForm.url.trim()) {
    Message.error('请输入URL')
    return
  }
  
  try {
    const config = sourceForm.config ? JSON.parse(sourceForm.config) : {}
    
    const data = {
      name: sourceForm.name,
      url: sourceForm.url,
      type: sourceForm.type,
      enabled: sourceForm.enabled,
      sync_interval: sourceForm.sync_interval,
      config
    }
    
    if (sourceForm.id > 0) {
      // 更新
      const response = await updateIntelligenceSource(sourceForm.id, data)
      if (response.code === 20000) {
        Message.success('情报源更新成功')
        sourceDialogVisible.value = false
        fetchSources()
      } else {
        Message.error(response.msg || '更新失败')
      }
    } else {
      // 创建（TODO: 需要实现创建接口）
      Message.info('创建功能开发中...')
    }
  } catch (error) {
    Message.error('保存失败')
    console.error('保存情报源失败:', error)
  }
}

// 取消情报源编辑
const handleCancelSource = () => {
  sourceDialogVisible.value = false
}

// 切换情报源状态
const handleToggleSource = async (record: IntelligenceSource) => {
  try {
    const data = {
      enabled: record.enabled
    }
    
    const response = await updateIntelligenceSource(record.id, data)
    if (response.code === 20000) {
      Message.success(`情报源已${record.enabled ? '启用' : '禁用'}`)
    } else {
      Message.error(response.msg || '操作失败')
      // 回滚状态
      record.enabled = !record.enabled
    }
  } catch (error) {
    Message.error('操作失败')
    console.error('切换情报源状态失败:', error)
    // 回滚状态
    record.enabled = !record.enabled
  }
}

// 同步情报源
const handleSyncSource = async (record: IntelligenceSource) => {
  syncingSourceId.value = record.id
  try {
    const response = await syncIntelligenceSource(record.id)
    if (response.code === 20000) {
      Message.success('同步任务已启动')
      // 可以在这里添加轮询检查同步状态的功能
    } else {
      Message.error(response.msg || '同步失败')
    }
  } catch (error) {
    Message.error('同步失败')
    console.error('同步情报源失败:', error)
  } finally {
    syncingSourceId.value = null
  }
}

// 同步所有情报源
const handleSyncAll = async () => {
  syncingAll.value = true
  try {
    // TODO: 实现批量同步接口
    Message.info('批量同步功能开发中...')
  } catch (error) {
    Message.error('同步失败')
    console.error('同步所有情报源失败:', error)
  } finally {
    syncingAll.value = false
  }
}

// 删除情报源
const handleDeleteSource = (record: IntelligenceSource) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除情报源 "${record.name}" 吗？此操作不可恢复。`,
    okText: '删除',
    cancelText: '取消',
    okButtonProps: {
      status: 'danger'
    },
    onOk: async () => {
      try {
        // TODO: 实现删除接口
        Message.info('删除功能开发中...')
      } catch (error) {
        Message.error('删除失败')
        console.error('删除情报源失败:', error)
      }
    }
  })
}

// 添加检测规则
const handleAddRule = () => {
  ruleDialogTitle.value = '添加检测规则'
  ruleForm.id = 0
  ruleForm.name = ''
  ruleForm.description = ''
  ruleForm.type = 'cve'
  ruleForm.condition = ''
  ruleForm.action = 'alert'
  ruleForm.severity = 'medium'
  ruleForm.enabled = true
  ruleForm.priority = 0
  ruleForm.tags = []
  ruleDialogVisible.value = true
}

// 编辑检测规则
const handleEditRule = (record: DetectionRule) => {
  ruleDialogTitle.value = '编辑检测规则'
  ruleForm.id = record.id
  ruleForm.name = record.name
  ruleForm.description = record.description || ''
  ruleForm.type = record.type
  ruleForm.condition = record.condition
  ruleForm.action = record.action
  ruleForm.severity = record.severity
  ruleForm.enabled = record.enabled
  ruleForm.priority = record.priority || 0
  ruleForm.tags = record.tags || []
  ruleDialogVisible.value = true
}

// 保存检测规则
const handleSaveRule = async () => {
  if (!ruleForm.name.trim()) {
    Message.error('请输入规则名称')
    return
  }
  
  if (!ruleForm.condition.trim()) {
    Message.error('请输入检测条件')
    return
  }
  
  try {
    const data = {
      name: ruleForm.name,
      description: ruleForm.description,
      type: ruleForm.type,
      condition: ruleForm.condition,
      action: ruleForm.action,
      severity: ruleForm.severity,
      enabled: ruleForm.enabled,
      priority: ruleForm.priority,
      tags: ruleForm.tags
    }
    
    if (ruleForm.id > 0) {
      // 更新
      const response = await updateDetectionRule(ruleForm.id, data)
      if (response.code === 20000) {
        Message.success('检测规则更新成功')
        ruleDialogVisible.value = false
        fetchRules()
      } else {
        Message.error(response.msg || '更新失败')
      }
    } else {
      // 创建
      const response = await createDetectionRule(data)
      if (response.code === 20000) {
        Message.success('检测规则创建成功')
        ruleDialogVisible.value = false
        fetchRules()
      } else {
        Message.error(response.msg || '创建失败')
      }
    }
  } catch (error) {
    Message.error('保存失败')
    console.error('保存检测规则失败:', error)
  }
}

// 取消检测规则编辑
const handleCancelRule = () => {
  ruleDialogVisible.value = false
}

// 切换检测规则状态
const handleToggleRule = async (record: DetectionRule) => {
  try {
    const data = {
      enabled: record.enabled
    }
    
    const response = await updateDetectionRule(record.id, data)
    if (response.code === 20000) {
      Message.success(`检测规则已${record.enabled ? '启用' : '禁用'}`)
    } else {
      Message.error(response.msg || '操作失败')
      // 回滚状态
      record.enabled = !record.enabled
    }
  } catch (error) {
    Message.error('操作失败')
    console.error('切换检测规则状态失败:', error)
    // 回滚状态
    record.enabled = !record.enabled
  }
}

// 删除检测规则
const handleDeleteRule = (record: DetectionRule) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除检测规则 "${record.name}" 吗？此操作不可恢复。`,
    okText: '删除',
    cancelText: '取消',
    okButtonProps: {
      status: 'danger'
    },
    onOk: async () => {
      try {
        const response = await deleteDetectionRule(record.id)
        if (response.code === 20000) {
          Message.success('检测规则删除成功')
          fetchRules()
        } else {
          Message.error(response.msg || '删除失败')
        }
      } catch (error) {
        Message.error('删除失败')
        console.error('删除检测规则失败:', error)
      }
    }
  })
}

// 添加预警配置
const handleAddAlertConfig = () => {
  Message.info('添加预警配置功能开发中...')
}

// 编辑预警配置
const handleEditAlertConfig = (record: AlertConfig) => {
  Message.info('编辑预警配置功能开发中...')
}

// 切换预警配置状态
const handleToggleAlertConfig = async (record: AlertConfig) => {
  Message.info('切换预警配置状态功能开发中...')
  // 回滚状态
  record.enabled = !record.enabled
}

// 删除预警配置
const handleDeleteAlertConfig = (record: AlertConfig) => {
  Message.info('删除预警配置功能开发中...')
}

// 工具函数
const getSourceTypeColor = (type: string) => {
  switch (type) {
    case 'avd': return 'red'
    case 'nvd': return 'blue'
    case 'cnnvd': return 'green'
    case 'cnvd': return 'orange'
    case 'exploitdb': return 'purple'
    case 'securityfocus': return 'cyan'
    default: return 'gray'
  }
}

const getSourceTypeText = (type: string) => {
  switch (type) {
    case 'avd': return 'AVD'
    case 'nvd': return 'NVD'
    case 'cnnvd': return 'CNNVD'
    case 'cnvd': return 'CNVD'
    case 'exploitdb': return 'ExploitDB'
    case 'securityfocus': return 'SecurityFocus'
    default: return type
  }
}

const getRuleTypeText = (type: string) => {
  switch (type) {
    case 'port': return '端口检测'
    case 'service': return '服务检测'
    case 'version': return '版本检测'
    case 'cve': return 'CVE检测'
    case 'custom': return '自定义'
    default: return type
  }
}

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

const getActionText = (action: string) => {
  switch (action) {
    case 'alert': return '预警'
    case 'block': return '阻断'
    case 'log': return '记录'
    case 'report': return '报告'
    default: return action
  }
}

const formatInterval = (seconds?: number) => {
  if (!seconds) return '-'
  if (seconds < 60) return `${seconds}秒`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}分钟`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}小时`
  return `${Math.floor(seconds / 86400)}天`
}

const formatDate = (dateString?: string) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN')
}

// 初始化
onMounted(() => {
  fetchSources()
  fetchRules()
  fetchAlertConfigs()
})
</script>

<style scoped>
.intelligence-config-page {
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