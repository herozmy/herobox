<template>
  <div class="service-status-compact">
    <div class="status-bar">
      <div class="status-info">
        <span class="service-title">服务状态</span>
        <div v-if="!serviceInfo" class="loading-status">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>获取中...</span>
        </div>
        <div v-else-if="serviceInfo.status === 'not_installed'" class="status-content">
          <el-tag type="warning" size="small">
            <span class="status-dot status-warning"></span>
            未安装
          </el-tag>
        </div>
        <div v-else class="status-content">
          <el-tag :type="getStatusType(serviceInfo.status)" size="small">
            <span :class="'status-dot status-' + serviceInfo.status"></span>
            {{ getStatusText(serviceInfo.status) }}
          </el-tag>
          <span v-if="serviceInfo.status === 'running' && serviceInfo.uptime" class="uptime-text">
            {{ serviceInfo.uptime }}
          </span>
        </div>
      </div>
      
      <div class="control-buttons">
        <el-button 
          size="small"
          @click="refreshServiceStatus"
          :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <template v-if="serviceInfo && serviceInfo.status !== 'not_installed'">
          <el-button 
            type="success" 
            size="small"
            :disabled="serviceInfo.status === 'running'"
            :loading="actionLoading === 'start'"
            @click="controlService('start')">
            <el-icon><VideoPlay /></el-icon>
            启动
          </el-button>
          <el-button 
            type="danger" 
            size="small"
            :disabled="serviceInfo.status !== 'running'"
            :loading="actionLoading === 'stop'"
            @click="controlService('stop')">
            <el-icon><VideoPause /></el-icon>
            停止
          </el-button>
          <el-button 
            type="warning" 
            size="small"
            :disabled="serviceInfo.status !== 'running'"
            :loading="actionLoading === 'restart'"
            @click="controlService('restart')">
            <el-icon><RefreshRight /></el-icon>
            重启
          </el-button>
          <el-button 
            type="primary" 
            size="small"
            @click="editConfig">
            <el-icon><Edit /></el-icon>
            编辑配置
          </el-button>
        </template>
      </div>
    </div>

    <!-- 配置编辑对话框 -->
    <el-dialog 
      v-model="configDialogVisible" 
      :title="`${serviceName === 'sing-box' ? 'Sing-Box' : 'MosDNS'} 配置编辑器`"
      width="80%"
      :before-close="closeConfigDialog"
      class="config-dialog">
      
      <div v-loading="configLoading" class="config-editor-container">
        <div class="editor-toolbar">
          <div class="toolbar-left">
            <el-button 
              type="primary" 
              size="small"
              @click="validateConfig"
              :loading="validating">
              <el-icon><CircleCheck /></el-icon>
              验证配置
            </el-button>
            <el-button 
              size="small"
              @click="formatConfig">
              <el-icon><Document /></el-icon>
              格式化
            </el-button>
          </div>
          <div class="toolbar-right">
            <el-button 
              type="success" 
              size="small"
              @click="saveConfig"
              :loading="saving">
              <el-icon><Select /></el-icon>
              保存配置
            </el-button>
          </div>
        </div>
        
        <div class="editor-content">
          <textarea 
            ref="configEditor"
            v-model="configContent"
            class="config-textarea"
            placeholder="配置内容..."
            spellcheck="false">
          </textarea>
        </div>
        
        <!-- 验证结果显示 -->
        <div v-if="validationResult" class="validation-result">
          <el-alert
            :title="validationResult.valid ? '✅ 配置验证通过' : '❌ 配置验证失败'"
            :type="validationResult.valid ? 'success' : 'error'"
            :closable="false"
            show-icon>
            <div v-if="validationResult.message">{{ validationResult.message }}</div>
            <div v-if="validationResult.validation_method" class="validation-method">
              <small>验证方式: {{ validationResult.validation_method }}</small>
            </div>
            <div v-if="!validationResult.valid && validationResult.error" class="error-details">
              <pre>{{ validationResult.error }}</pre>
            </div>
          </el-alert>
        </div>
      </div>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="closeConfigDialog">取消</el-button>
          <el-button 
            type="primary" 
            @click="saveConfig"
            :loading="saving">
            保存
          </el-button>
        </span>
      </template>
    </el-dialog>

  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  Refresh, 
  VideoPlay, 
  VideoPause, 
  RefreshRight, 
  Loading,
  Edit,
  CircleCheck,
  Document,
  Select
} from '@element-plus/icons-vue'
import { 
  apiGetServiceInfo, 
  apiControlService, 
  apiGetSingBoxConfig, 
  apiUpdateSingBoxConfig, 
  apiValidateSingBoxConfig,
  apiGetConfig,
  apiUpdateConfig
} from '../utils/api'

// Props
const props = defineProps({
  serviceName: {
    type: String,
    default: 'sing-box'
  },
  autoRefresh: {
    type: Boolean,
    default: true
  },
  refreshInterval: {
    type: Number,
    default: 10000 // 10秒
  }
})

// 响应式数据
const serviceInfo = ref(null)
const loading = ref(false)
const actionLoading = ref('')

// 配置编辑相关
const configDialogVisible = ref(false)
const configLoading = ref(false)
const saving = ref(false)
const validating = ref(false)
const configContent = ref('')
const validationResult = ref(null)
const configEditor = ref()

// 自动刷新定时器
let refreshTimer = null

// 获取状态类型
const getStatusType = (status) => {
  const types = {
    running: 'success',
    stopped: 'info',
    failed: 'danger',
    not_installed: 'warning',
    unknown: 'warning'
  }
  return types[status] || 'info'
}

// 获取状态文本
const getStatusText = (status) => {
  const texts = {
    running: '运行中',
    stopped: '未运行',
    failed: '失败',
    not_installed: '未安装',
    unknown: '未知'
  }
  return texts[status] || '未知'
}

// 获取状态描述
const getStatusDescription = (status) => {
  const descriptions = {
    running: '服务正在正常运行',
    stopped: '服务已停止，可以执行启动操作',
    failed: '服务启动失败，建议检查配置后重新启动',
    unknown: '服务状态未知，请检查服务状态'
  }
  return descriptions[status] || '请选择合适的操作'
}

// 获取警告类型
const getAlertType = (status) => {
  const types = {
    running: 'success',
    stopped: 'info',
    failed: 'error',
    unknown: 'warning'
  }
  return types[status] || 'info'
}

// 获取服务状态
const getServiceStatus = async () => {
  if (loading.value) return
  
  loading.value = true
  try {
    const response = await apiGetServiceInfo(props.serviceName)
    serviceInfo.value = response.data
  } catch (error) {
    console.error('获取服务状态失败:', error)
    ElMessage.error('获取服务状态失败: ' + (error.response?.data?.message || error.message))
  } finally {
    loading.value = false
  }
}

// 刷新服务状态
const refreshServiceStatus = async () => {
  await getServiceStatus()
}

// 控制服务
const controlService = async (action) => {
  if (actionLoading.value) return
  
  actionLoading.value = action
  try {
    const response = await apiControlService(props.serviceName, action)
    ElMessage.success(response.data.message || `服务${action}操作成功`)
    
    // 延迟刷新状态
    setTimeout(() => {
      getServiceStatus()
    }, 2000)
  } catch (error) {
    console.error(`服务${action}操作失败:`, error)
    ElMessage.error(`服务${action}操作失败: ` + (error.response?.data?.message || error.message))
  } finally {
    actionLoading.value = ''
  }
}

// 开始自动刷新
const startAutoRefresh = () => {
  if (props.autoRefresh && props.refreshInterval > 0) {
    refreshTimer = setInterval(() => {
      getServiceStatus()
    }, props.refreshInterval)
  }
}

// 停止自动刷新
const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

// 编辑配置
const editConfig = async () => {
  try {
    configLoading.value = true
    configDialogVisible.value = true
    validationResult.value = null
    
    // 根据服务类型获取配置
    let response
    if (props.serviceName === 'sing-box') {
      response = await apiGetSingBoxConfig()
    } else if (props.serviceName === 'mosdns') {
      response = await apiGetConfig('mosdns')
    } else {
      throw new Error('不支持的服务类型')
    }
    
    configContent.value = typeof response.data === 'string' 
      ? response.data 
      : JSON.stringify(response.data, null, 2)
      
  } catch (error) {
    console.error('获取配置失败:', error)
    ElMessage.error('获取配置失败: ' + (error.response?.data?.message || error.message))
    configDialogVisible.value = false
  } finally {
    configLoading.value = false
  }
}

// 关闭配置编辑对话框
const closeConfigDialog = () => {
  configDialogVisible.value = false
  configContent.value = ''
  validationResult.value = null
}

// 验证配置
const validateConfig = async () => {
  if (!configContent.value.trim()) {
    ElMessage.warning('配置内容不能为空')
    return
  }
  
  try {
    validating.value = true
    validationResult.value = null
    
    let config
    try {
      config = JSON.parse(configContent.value)
    } catch (parseError) {
      validationResult.value = {
        valid: false,
        message: 'JSON 格式错误',
        error: parseError.message
      }
      return
    }
    
    // 根据服务类型调用验证API
    let response
    if (props.serviceName === 'sing-box') {
      response = await apiValidateSingBoxConfig(config)
    } else {
      // mosdns暂时只做JSON格式验证
      validationResult.value = {
        valid: true,
        message: 'JSON 格式验证通过'
      }
      return
    }
    
    const data = response.data
    validationResult.value = {
      valid: data.valid,
      message: data.message,
      error: data.errors && data.errors.length > 0 ? data.errors.join('\n') : '',
      validation_method: data.validation_method
    }
    
  } catch (error) {
    console.error('验证配置失败:', error)
    validationResult.value = {
      valid: false,
      message: '验证失败',
      error: error.response?.data?.message || error.message
    }
  } finally {
    validating.value = false
  }
}

// 格式化配置
const formatConfig = () => {
  try {
    const config = JSON.parse(configContent.value)
    configContent.value = JSON.stringify(config, null, 2)
    ElMessage.success('配置格式化成功')
  } catch (error) {
    ElMessage.error('JSON 格式错误，无法格式化')
  }
}

// 保存配置
const saveConfig = async () => {
  if (!configContent.value.trim()) {
    ElMessage.warning('配置内容不能为空')
    return
  }
  
  try {
    saving.value = true
    
    // 先验证配置格式
    let config
    try {
      config = JSON.parse(configContent.value)
    } catch (parseError) {
      ElMessage.error('JSON 格式错误: ' + parseError.message)
      return
    }
    
    // 根据服务类型保存配置
    if (props.serviceName === 'sing-box') {
      await apiUpdateSingBoxConfig(config)
    } else if (props.serviceName === 'mosdns') {
      await apiUpdateConfig('mosdns', configContent.value)
    }
    
    ElMessage.success('配置保存成功')
    closeConfigDialog()
    
    // 刷新服务状态
    setTimeout(() => {
      getServiceStatus()
    }, 1000)
    
  } catch (error) {
    console.error('保存配置失败:', error)
    ElMessage.error('保存配置失败: ' + (error.response?.data?.message || error.message))
  } finally {
    saving.value = false
  }
}

// 生命周期
onMounted(() => {
  getServiceStatus()
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
})

// 暴露方法给父组件
defineExpose({
  refreshServiceStatus,
  getServiceStatus
})
</script>

<style scoped>
.service-status-compact {
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 6px;
  padding: 12px 16px;
  margin-bottom: 16px;
}

.status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  min-height: 32px;
}

.status-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.service-title {
  font-weight: 600;
  font-size: 14px;
  color: #303133;
}

.status-content {
  display: flex;
  align-items: center;
  gap: 8px;
}

.loading-status {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #909399;
  font-size: 12px;
}

.uptime-text {
  color: #909399;
  font-size: 12px;
}

.control-buttons {
  display: flex;
  align-items: center;
  gap: 6px;
}

.status-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  margin-right: 4px;
}

.status-running {
  background-color: #67c23a;
  animation: pulse 2s infinite;
}

.status-stopped {
  background-color: #909399;
}

.status-failed {
  background-color: #f56c6c;
}

.status-unknown,
.status-warning {
  background-color: #e6a23c;
}

@keyframes pulse {
  0% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
  100% {
    opacity: 1;
  }
}

/* 响应式设计 */
@media (max-width: 768px) {
  .status-bar {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
    min-height: auto;
  }
  
  .control-buttons {
    width: 100%;
    justify-content: flex-end;
    flex-wrap: wrap;
  }
  
  .service-status-compact {
    padding: 10px 12px;
  }
}

/* 配置编辑器对话框样式 */
.config-dialog {
  .el-dialog__body {
    padding: 20px;
  }
}

.config-editor-container {
  display: flex;
  flex-direction: column;
  height: 60vh;
  min-height: 400px;
}

.editor-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding: 10px 0;
  border-bottom: 1px solid #e4e7ed;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  gap: 10px;
}

.editor-content {
  flex: 1;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
}

.config-textarea {
  width: 100%;
  height: 100%;
  border: none;
  outline: none;
  padding: 15px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 14px;
  line-height: 1.5;
  background-color: #fafafa;
  resize: none;
  color: #2c3e50;
}

.config-textarea:focus {
  background-color: #fff;
}

.validation-result {
  margin-top: 15px;
}

.validation-result .validation-method {
  margin-top: 8px;
  color: #909399;
  font-style: italic;
}

.validation-result .error-details {
  margin-top: 10px;
}

.validation-result pre {
  background-color: #f8f8f8;
  padding: 10px;
  border-radius: 4px;
  margin: 0;
  font-size: 12px;
  line-height: 1.4;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>
