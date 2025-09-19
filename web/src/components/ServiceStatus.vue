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
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { 
  Refresh, 
  VideoPlay, 
  VideoPause, 
  RefreshRight, 
  Loading,
  Edit
} from '@element-plus/icons-vue'
import { apiGetServiceInfo, apiControlService } from '../utils/api'

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
const router = useRouter()
const editConfig = () => {
  // 根据服务名跳转到对应的配置页面
  if (props.serviceName === 'sing-box') {
    router.push('/singbox-manage')
  } else if (props.serviceName === 'mosdns') {
    router.push('/mosdns-manage')
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
</style>
