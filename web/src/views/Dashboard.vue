<template>
  <div class="dashboard">
    <div class="page-header">
      <h2>仪表板</h2>
      <el-button @click="refreshData" :loading="loading">
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
    </div>

    <!-- 三栏布局：系统信息、Sing-Box管理、MosDNS管理 -->
    <el-row :gutter="20">
      <!-- 系统信息 -->
      <el-col :span="8">
        <el-card class="system-card">
          <template #header>
            <div class="card-header">
              <span>系统信息</span>
              <el-icon><Monitor /></el-icon>
            </div>
          </template>
          
          <div v-if="systemInfo" class="system-info">
            <div class="info-item">
              <span class="label">操作系统:</span>
              <span class="value">{{ systemInfo.os_info }}</span>
            </div>
            <div class="info-item">
              <span class="label">运行时间:</span>
              <span class="value">{{ systemInfo.uptime }}</span>
            </div>
          </div>
          
          <el-skeleton v-else :rows="2" animated />
        </el-card>
      </el-col>

      <!-- Sing-Box 管理 -->
      <el-col :span="8">
        <el-card class="service-management-card">
          <template #header>
            <div class="card-header">
              <span>Sing-Box 管理</span>
              <el-icon><Setting /></el-icon>
            </div>
          </template>
          
          <div v-if="services && services['sing-box']" class="service-management">
            <!-- 服务状态 -->
            <div class="status-section">
              <div class="service-status-row">
                <span class="service-label">状态:</span>
                <el-tag :type="getStatusType(services['sing-box'].status)">
                  <span :class="'status-dot status-' + services['sing-box'].status"></span>
                  {{ getStatusText(services['sing-box'].status) }}
                </el-tag>
              </div>
              
              <div class="service-details">
                <div v-if="services['sing-box'].status === 'running'" class="detail-content">
                  <div class="detail-item">
                    <span class="label">PID:</span>
                    <span class="value">{{ services['sing-box'].pid || 'N/A' }}</span>
                  </div>
                  <div class="detail-item">
                    <span class="label">运行时间:</span>
                    <span class="value">{{ services['sing-box'].uptime || 'N/A' }}</span>
                  </div>
                </div>
                <div v-else class="service-stopped-info">
                  <div v-if="services['sing-box'].status === 'not_installed'" class="not-installed-content">
                    <div class="status-message">{{ getServiceStoppedText(services['sing-box'].status) }}</div>
                    <div class="install-guide">
                      <div class="guide-title">📖 安装指南</div>
                      <div class="guide-buttons">
                        <el-button 
                          type="primary" 
                          size="small" 
                          plain
                          @click="openLink('https://sing-box.sagernet.org/installation/')"
                          class="guide-button">
                          <el-icon><Document /></el-icon>
                          官方文档
                        </el-button>
                        <el-button 
                          type="info" 
                          size="small" 
                          plain
                          @click="openLink('https://github.com/SagerNet/sing-box')"
                          class="guide-button">
                          <el-icon><Link /></el-icon>
                          GitHub 仓库
                        </el-button>
                      </div>
                    </div>
                  </div>
                  <div v-else class="simple-status">
                    {{ getServiceStoppedText(services['sing-box'].status) }}
                  </div>
                </div>
              </div>
            </div>

            <!-- 操作按钮 -->
            <div v-if="services['sing-box'].status !== 'not_installed'" class="action-buttons">
              <el-button 
                type="success" 
                size="small"
                :disabled="services['sing-box'].status === 'running'"
                :loading="actionLoading === 'sing-box-start'"
                @click="controlService('sing-box', 'start')"
              >
                启动
              </el-button>
              <el-button 
                type="danger" 
                size="small"
                :disabled="services['sing-box'].status !== 'running'"
                :loading="actionLoading === 'sing-box-stop'"
                @click="controlService('sing-box', 'stop')"
              >
                停止
              </el-button>
              <el-button 
                type="warning" 
                size="small"
                :disabled="services['sing-box'].status !== 'running'"
                :loading="actionLoading === 'sing-box-restart'"
                @click="controlService('sing-box', 'restart')"
              >
                重启
              </el-button>
            </div>
            <div v-else class="no-service-message">
              <span>请先安装服务后再进行操作</span>
            </div>
          </div>
          
          <el-skeleton v-else :rows="4" animated />
        </el-card>
      </el-col>

      <!-- MosDNS 管理 -->
      <el-col :span="8">
        <el-card class="service-management-card">
          <template #header>
            <div class="card-header">
              <span>MosDNS 管理</span>
              <el-icon><Setting /></el-icon>
            </div>
          </template>
          
          <div v-if="services && services['mosdns']" class="service-management">
            <!-- 服务状态 -->
            <div class="status-section">
              <div class="service-status-row">
                <span class="service-label">状态:</span>
                <el-tag :type="getStatusType(services['mosdns'].status)">
                  <span :class="'status-dot status-' + services['mosdns'].status"></span>
                  {{ getStatusText(services['mosdns'].status) }}
                </el-tag>
              </div>
              
              <div class="service-details">
                <div v-if="services['mosdns'].status === 'running'" class="detail-content">
                  <div class="detail-item">
                    <span class="label">PID:</span>
                    <span class="value">{{ services['mosdns'].pid || 'N/A' }}</span>
                  </div>
                  <div class="detail-item">
                    <span class="label">运行时间:</span>
                    <span class="value">{{ services['mosdns'].uptime || 'N/A' }}</span>
                  </div>
                </div>
                <div v-else class="service-stopped-info">
                  <div v-if="services['mosdns'].status === 'not_installed'" class="not-installed-content">
                    <div class="status-message">{{ getServiceStoppedText(services['mosdns'].status) }}</div>
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
                  <div v-else class="simple-status">
                    {{ getServiceStoppedText(services['mosdns'].status) }}
                  </div>
                </div>
              </div>
            </div>

            <!-- 操作按钮 -->
            <div v-if="services['mosdns'].status !== 'not_installed'" class="action-buttons">
              <el-button 
                type="success" 
                size="small"
                :disabled="services['mosdns'].status === 'running'"
                :loading="actionLoading === 'mosdns-start'"
                @click="controlService('mosdns', 'start')"
              >
                启动
              </el-button>
              <el-button 
                type="danger" 
                size="small"
                :disabled="services['mosdns'].status !== 'running'"
                :loading="actionLoading === 'mosdns-stop'"
                @click="controlService('mosdns', 'stop')"
              >
                停止
              </el-button>
              <el-button 
                type="warning" 
                size="small"
                :disabled="services['mosdns'].status !== 'running'"
                :loading="actionLoading === 'mosdns-restart'"
                @click="controlService('mosdns', 'restart')"
              >
                重启
              </el-button>
            </div>
            <div v-else class="no-service-message">
              <span>请先安装服务后再进行操作</span>
            </div>
          </div>
          
          <el-skeleton v-else :rows="4" animated />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { apiGetDashboard, apiControlService } from '../utils/api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Monitor, Setting, Link, Refresh, Document } from '@element-plus/icons-vue'

const loading = ref(false)
const actionLoading = ref('')
const systemInfo = ref(null)
const services = ref(null)
const recentLogs = ref(null)

const getStatusType = (status) => {
  const types = {
    running: 'success',
    stopped: 'warning',
    failed: 'danger',
    not_installed: 'info',
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

const openLink = (url) => {
  window.open(url, '_blank')
}

const controlService = async (serviceName, action) => {
  const actionNames = {
    start: '启动',
    stop: '停止',
    restart: '重启',
    reload: '重载'
  }

  const serviceNames = {
    'mosdns': 'MosDNS',
    'sing-box': 'Sing-Box'
  }

  try {
    await ElMessageBox.confirm(
      `确定要${actionNames[action]} ${serviceNames[serviceName]} 服务吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    const loadingKey = `${serviceName}-${action}`
    actionLoading.value = loadingKey

    const response = await apiControlService(serviceName, action)
    if (response.code === 200) {
      ElMessage.success(`${serviceNames[serviceName]} 服务${actionNames[action]}成功`)
      // 更新服务信息
      if (response.data.service_info) {
        services.value[serviceName] = response.data.service_info
      }
      // 延迟刷新数据
      setTimeout(() => {
        loadDashboardData()
      }, 1000)
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(`服务操作失败: ${error.message || error}`)
      console.error(error)
    }
  } finally {
    actionLoading.value = ''
  }
}

const loadDashboardData = async () => {
  loading.value = true
  try {
    const response = await apiGetDashboard()
    if (response.code === 200) {
      const data = response.data
      systemInfo.value = data.system_info
      services.value = data.services
      recentLogs.value = data.recent_logs
    }
  } catch (error) {
    ElMessage.error('加载仪表板数据失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const refreshData = () => {
  loadDashboardData()
}

onMounted(() => {
  loadDashboardData()
})
</script>

<style scoped>
.dashboard {
  height: 100%;
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

.system-info,
.services-list {
  padding: 10px 0;
}

.info-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;
}

.info-item:last-child {
  border-bottom: none;
}

.label {
  font-weight: 500;
  color: #606266;
}

.value {
  color: #303133;
}

.service-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding: 10px;
  background-color: #fafafa;
  border-radius: 6px;
}

.service-name {
  display: flex;
  align-items: center;
  font-weight: 500;
}

.service-info {
  color: #909399;
  font-size: 12px;
}

/* 新增的服务管理样式 */
.service-management-card {
  height: 280px;
}

.service-management {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.status-section {
  flex: 1;
  margin-bottom: 15px;
}

.service-status-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.service-label {
  font-weight: 500;
  color: #606266;
}

.service-details {
  margin-top: 10px;
  min-height: 80px;
  display: flex;
  align-items: center;
}

.detail-content {
  width: 100%;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  padding: 5px 0;
  font-size: 14px;
}

.service-stopped-info {
  width: 100%;
  color: #909399;
  font-size: 14px;
  padding: 15px;
  background-color: #fafafa;
  border-radius: 6px;
}

.simple-status {
  text-align: center;
}

.not-installed-content {
  text-align: left;
  width: 100%;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.status-message {
  text-align: center;
  margin-bottom: 12px;
  font-weight: 500;
  color: #909399;
}

.install-guide {
  border-top: 1px solid #e4e7ed;
  padding-top: 10px;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.guide-title {
  font-size: 13px;
  font-weight: 500;
  color: #606266;
  margin-bottom: 8px;
  text-align: center;
}

.guide-buttons {
  display: flex;
  flex-direction: row;
  gap: 8px;
  width: 100%;
}

.guide-button {
  flex: 1 !important;
  height: 32px !important;
  display: flex !important;
  justify-content: center !important;
  align-items: center !important;
  border-radius: 6px !important;
  min-width: 0 !important;
}

.guide-button .el-icon {
  margin-right: 6px !important;
  font-size: 14px !important;
}

.guide-button span {
  font-size: 12px !important;
  white-space: nowrap !important;
  overflow: hidden !important;
  text-overflow: ellipsis !important;
}

.detail-item .label {
  color: #606266;
}

.no-service-message {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 40px;
  background-color: #f5f7fa;
  border: 1px dashed #dcdfe6;
  border-radius: 6px;
  color: #909399;
  font-size: 13px;
}

.detail-item .value {
  color: #303133;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.action-buttons .el-button {
  flex: 1;
  min-width: 60px;
}

.log-content {
  background-color: #1e1e1e;
  color: #d4d4d4;
  padding: 15px;
  border-radius: 6px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.5;
  max-height: 300px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header span {
  font-weight: 500;
}
</style>

