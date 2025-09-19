<template>
  <div class="mosdns-manage">
    <div class="page-header">
      <h2>MosDNS 管理</h2>
      <el-button @click="refreshData" :loading="loading">
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
    </div>

    <!-- 服务状态和DNS配置管理 -->
    <el-row :gutter="20" class="service-section">
      <el-col :span="12">
        <el-card class="service-status-card">
          <template #header>
            <div class="card-header">
              <span>服务状态</span>
              <div v-if="serviceInfo && serviceInfo.status !== 'not_installed'" class="header-controls">
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
              </div>
              <el-icon v-else><Monitor /></el-icon>
            </div>
          </template>
          
          <div v-if="serviceInfo" class="service-content">
            <div class="status-row">
              <span class="label">状态:</span>
              <el-tag :type="getStatusType(serviceInfo.status)">
                <span :class="'status-dot status-' + serviceInfo.status"></span>
                {{ getStatusText(serviceInfo.status) }}
              </el-tag>
            </div>
            
            <div v-if="serviceInfo.status === 'running'" class="service-details">
              <div class="detail-item">
                <span class="label">进程ID:</span>
                <span class="value">{{ serviceInfo.pid || 'N/A' }}</span>
              </div>
              <div class="detail-item">
                <span class="label">运行时间:</span>
                <span class="value">{{ serviceInfo.uptime || 'N/A' }}</span>
              </div>
            </div>
            
            <div v-else-if="serviceInfo.status === 'not_installed'" class="not-installed-content">
              <div class="status-message">服务未安装</div>
              <div class="install-guide">
                <div class="guide-title">📖 安装指南</div>
                <div class="guide-buttons">
                  <el-button 
                    type="info" 
                    size="small" 
                    plain
                    @click="openLink('https://github.com/IrineSistiana/mosdns')"
                    class="guide-button">
                    <el-icon><Link /></el-icon>
                    GitHub 仓库
                  </el-button>
                  <el-button 
                    type="warning" 
                    size="small" 
                    plain
                    @click="openLink('https://github.com/IrineSistiana/mosdns/wiki')"
                    class="guide-button">
                    <el-icon><Document /></el-icon>
                    安装Wiki
                  </el-button>
                </div>
              </div>
            </div>
            
            <div v-else class="service-stopped">
              {{ getServiceStoppedText(serviceInfo.status) }}
            </div>
          </div>
          
          <el-skeleton v-else :rows="3" animated />
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card class="config-card" v-if="serviceInfo && serviceInfo.status !== 'not_installed'">
          <template #header>
            <div class="card-header">
              <span>DNS配置管理</span>
              <el-icon><Document /></el-icon>
            </div>
          </template>
          
          <div class="config-content">
            <div class="config-item">
              <h4>配置文件</h4>
              <p>管理 MosDNS 配置文件</p>
              <el-button type="primary" @click="openConfigEditor" class="config-btn">
                <el-icon><Edit /></el-icon>
                编辑配置
              </el-button>
            </div>
            
            <div class="config-item">
              <h4>规则管理</h4>
              <p>管理DNS解析规则</p>
              <el-button type="info" @click="manageRules" class="config-btn">
                <el-icon><List /></el-icon>
                管理规则
              </el-button>
            </div>
            
            <div class="config-item">
              <h4>配置验证</h4>
              <p>验证配置文件格式</p>
              <el-button type="success" @click="validateConfig" class="config-btn">
                <el-icon><CircleCheck /></el-icon>
                验证配置
              </el-button>
            </div>
          </div>
        </el-card>
        
        <el-card class="install-help-card" v-else-if="serviceInfo && serviceInfo.status === 'not_installed'">
          <template #header>
            <div class="card-header">
              <span>安装帮助</span>
              <el-icon><QuestionFilled /></el-icon>
            </div>
          </template>
          
          <div class="install-help-content">
            <el-empty 
              description="服务未安装，请先安装 MosDNS 服务"
              :image-size="100">
              <el-button type="primary" @click="openLink('https://github.com/IrineSistiana/mosdns/wiki')">
                查看安装指南
              </el-button>
            </el-empty>
          </div>
        </el-card>
        
        <el-skeleton v-else :rows="4" animated />
      </el-col>
    </el-row>

    <!-- DNS统计信息 -->
    <el-row :gutter="20" v-if="serviceInfo && serviceInfo.status === 'running'">
      <el-col :span="12">
        <el-card class="stats-card">
          <template #header>
            <div class="card-header">
              <span>DNS统计</span>
              <el-icon><DataAnalysis /></el-icon>
            </div>
          </template>
          
          <div class="stats-content">
            <div class="stat-item">
              <span class="stat-label">查询总数:</span>
              <span class="stat-value">{{ dnsStats.totalQueries || '0' }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">缓存命中:</span>
              <span class="stat-value">{{ dnsStats.cacheHits || '0' }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">被阻止:</span>
              <span class="stat-value">{{ dnsStats.blockedQueries || '0' }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">平均延迟:</span>
              <span class="stat-value">{{ dnsStats.avgLatency || '0ms' }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card class="tools-card">
          <template #header>
            <div class="card-header">
              <span>DNS工具</span>
              <el-icon><Tools /></el-icon>
            </div>
          </template>
          
          <div class="tools-content">
            <div class="tool-item">
              <el-button type="primary" @click="flushCache">
                <el-icon><Delete /></el-icon>
                清空DNS缓存
              </el-button>
            </div>
            <div class="tool-item">
              <el-button type="info" @click="testDNS">
                <el-icon><Search /></el-icon>
                DNS解析测试
              </el-button>
            </div>
            <div class="tool-item">
              <el-button type="warning" @click="reloadRules">
                <el-icon><RefreshRight /></el-icon>
                重载规则
              </el-button>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 日志查看 -->
    <el-card class="logs-section" v-if="serviceInfo && serviceInfo.status !== 'not_installed'">
      <template #header>
        <div class="card-header">
          <span>服务日志</span>
          <el-icon><DocumentCopy /></el-icon>
        </div>
      </template>
      
      <div class="logs-content">
        <div class="logs-toolbar">
          <el-button size="small" @click="refreshLogs" :loading="logsLoading">
            <el-icon><Refresh /></el-icon>
            刷新日志
          </el-button>
          <el-button size="small" @click="clearLogs">
            <el-icon><Delete /></el-icon>
            清空显示
          </el-button>
          <el-select v-model="logLevel" size="small" style="width: 120px; margin-left: 10px;">
            <el-option label="全部" value="all" />
            <el-option label="错误" value="error" />
            <el-option label="警告" value="warn" />
            <el-option label="信息" value="info" />
            <el-option label="调试" value="debug" />
          </el-select>
        </div>
        
        <div class="logs-container">
          <pre v-if="logs.length > 0" class="logs-text">{{ logs.join('\n') }}</pre>
          <el-empty v-else description="暂无日志数据" :image-size="80" />
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Refresh, Monitor, Setting, Document, Link, VideoPlay, VideoPause, 
  RefreshRight, Edit, List, CircleCheck, DataAnalysis, Tools, Delete, 
  Search, DocumentCopy, QuestionFilled 
} from '@element-plus/icons-vue'
import { apiGetServiceInfo, apiControlService, apiGetLogs } from '../utils/api'

const loading = ref(false)
const actionLoading = ref('')
const logsLoading = ref(false)
const serviceInfo = ref(null)
const logs = ref([])
const logLevel = ref('all')
const dnsStats = ref({})

const getStatusType = (status) => {
  const types = {
    running: 'success',
    stopped: 'info',
    failed: 'danger',
    not_installed: 'warning',
    unknown: 'info'
  }
  return types[status] || 'info'
}

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

const getServiceStoppedText = (status) => {
  const texts = {
    stopped: '服务未运行',
    failed: '服务启动失败',
    not_installed: '服务未安装',
    unknown: '服务状态未知'
  }
  return texts[status] || '服务当前未运行'
}

const getControlDescription = (status) => {
  const descriptions = {
    running: '服务正在运行中，可以执行停止或重启操作',
    stopped: '服务已停止，可以执行启动操作',
    failed: '服务启动失败，建议检查配置后重新启动',
    unknown: '服务状态未知，请检查服务状态'
  }
  return descriptions[status] || '请选择合适的操作'
}

const getAlertType = (status) => {
  const types = {
    running: 'success',
    stopped: 'warning',
    failed: 'error',
    unknown: 'info'
  }
  return types[status] || 'info'
}

const openLink = (url) => {
  window.open(url, '_blank')
}

const refreshData = async () => {
  loading.value = true
  try {
    const response = await apiGetServiceInfo('mosdns')
    serviceInfo.value = response.data
    
    // 如果服务运行中，获取DNS统计
    if (serviceInfo.value.status === 'running') {
      // 这里可以调用获取DNS统计的API
      dnsStats.value = {
        totalQueries: '1,234',
        cacheHits: '987',
        blockedQueries: '45',
        avgLatency: '12ms'
      }
    }
  } catch (error) {
    console.error('获取服务信息失败:', error)
    ElMessage.error('获取服务信息失败')
  } finally {
    loading.value = false
  }
}

const controlService = async (action) => {
  const actionNames = {
    start: '启动',
    stop: '停止',
    restart: '重启'
  }
  
  try {
    await ElMessageBox.confirm(
      `确定要${actionNames[action]} MosDNS 服务吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    actionLoading.value = action
    const response = await apiControlService('mosdns', action)
    
    if (response.data.success) {
      ElMessage.success(`${actionNames[action]}操作执行成功`)
      // 延迟刷新状态
      setTimeout(refreshData, 1000)
    } else {
      ElMessage.error(response.data.message || `${actionNames[action]}操作失败`)
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('控制服务失败:', error)
      ElMessage.error(`${actionNames[action]}操作失败`)
    }
  } finally {
    actionLoading.value = ''
  }
}

const openConfigEditor = () => {
  ElMessage.info('配置编辑功能开发中...')
}

const manageRules = () => {
  ElMessage.info('规则管理功能开发中...')
}

const validateConfig = () => {
  ElMessage.info('配置验证功能开发中...')
}

const flushCache = () => {
  ElMessage.info('清空DNS缓存功能开发中...')
}

const testDNS = () => {
  ElMessage.info('DNS解析测试功能开发中...')
}

const reloadRules = () => {
  ElMessage.info('重载规则功能开发中...')
}

const refreshLogs = async () => {
  logsLoading.value = true
  try {
    const response = await apiGetLogs('mosdns')
    logs.value = response.data.logs || []
  } catch (error) {
    console.error('获取日志失败:', error)
    ElMessage.error('获取日志失败')
  } finally {
    logsLoading.value = false
  }
}

const clearLogs = () => {
  logs.value = []
}

onMounted(() => {
  refreshData()
  refreshLogs()
})
</script>

<style scoped>
.mosdns-manage {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  color: #303133;
}

.service-section {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-controls {
  display: flex;
  gap: 8px;
  align-items: center;
}

.header-controls .el-button {
  margin: 0;
}

.service-status-card,
.config-card,
.install-help-card {
  height: 100%;
}

.service-content {
  min-height: 200px;
}

.status-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 15px;
  background-color: #f8f9fa;
  border-radius: 8px;
}

.label {
  font-weight: 500;
  color: #606266;
}

.value {
  color: #303133;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}

.service-details {
  padding: 15px;
  background-color: #f0f9ff;
  border-radius: 8px;
  border: 1px solid #e1f5fe;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10px;
  padding: 5px 0;
}

.detail-item:last-child {
  margin-bottom: 0;
}

.not-installed-content {
  text-align: center;
  padding: 20px;
}

.status-message {
  margin-bottom: 15px;
  font-weight: 500;
  color: #909399;
}

.install-guide {
  border-top: 1px solid #e4e7ed;
  padding-top: 15px;
}

.guide-title {
  font-size: 14px;
  font-weight: 500;
  color: #606266;
  margin-bottom: 10px;
}

.guide-buttons {
  display: flex;
  justify-content: center;
  gap: 10px;
}

.guide-button {
  min-width: 120px;
}

.service-stopped {
  text-align: center;
  padding: 30px;
  color: #909399;
  background-color: #fafafa;
  border-radius: 8px;
}

.control-content {
  min-height: 200px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.control-buttons {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.control-btn {
  width: 100%;
  height: 50px;
  font-size: 16px;
}

.no-service-content {
  min-height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.config-section,
.logs-section {
  margin-bottom: 20px;
}

.config-content {
  padding: 10px 0;
}

.config-item {
  text-align: center;
  padding: 20px 15px;
  margin-bottom: 15px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background-color: #fafafa;
}

.config-item:last-child {
  margin-bottom: 0;
}

.config-btn {
  width: 100%;
}

.install-help-content {
  padding: 20px;
}

.config-item h4 {
  margin: 0 0 10px 0;
  color: #303133;
}

.config-item p {
  margin: 0 0 15px 0;
  color: #606266;
  font-size: 14px;
}

.stats-card,
.tools-card {
  margin-bottom: 20px;
}

.stats-content {
  padding: 20px 0;
}

.stat-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding: 10px 15px;
  background-color: #f8f9fa;
  border-radius: 6px;
}

.stat-item:last-child {
  margin-bottom: 0;
}

.stat-label {
  font-weight: 500;
  color: #606266;
}

.stat-value {
  font-weight: 600;
  color: #303133;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}

.tools-content {
  padding: 20px 0;
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.tool-item {
  width: 100%;
}

.tool-item .el-button {
  width: 100%;
}

.logs-content {
  padding: 20px 0;
}

.logs-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 15px;
}

.logs-container {
  height: 300px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  overflow: auto;
  background-color: #f8f9fa;
}

.logs-text {
  padding: 15px;
  margin: 0;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.5;
  color: #303133;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 6px;
}

.status-running {
  background-color: #67c23a;
}

.status-stopped {
  background-color: #909399;
}

.status-failed {
  background-color: #f56c6c;
}

.status-not_installed {
  background-color: #e6a23c;
}

.status-unknown {
  background-color: #909399;
}
</style>
