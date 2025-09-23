<template>
  <div class="sing-box-manage">
    <div class="page-header">
      <h2>Sing-Box 管理</h2>
      <el-button @click="refreshData" :loading="loading">
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
    </div>

    <!-- 服务状态和配置管理 -->
    <el-row :gutter="20" class="service-section">
      <el-col :span="12">
        <el-card class="service-status-card">
          <template #header>
            <div class="card-header">
              <div class="header-left">
                <span class="service-title">服务状态</span>
                <div v-if="serviceInfo && serviceInfo.status !== 'not_installed'" class="inline-status">
                  <el-tag :type="getStatusType(serviceInfo.status)" size="small">
                    <span :class="'status-dot status-' + serviceInfo.status"></span>
                    {{ getStatusText(serviceInfo.status) }}
                  </el-tag>
                  <span v-if="serviceInfo.status === 'running' && serviceInfo.uptime" class="uptime-text">
                    {{ serviceInfo.uptime }}
                  </span>
                </div>
              </div>
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
                <el-button 
                  type="primary" 
                  size="small"
                  @click="openConfigEditor">
                  <el-icon><Edit /></el-icon>
                  编辑配置
                </el-button>
              </div>
              <el-icon v-else><Monitor /></el-icon>
            </div>
          </template>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card class="install-help-card" v-if="serviceInfo && serviceInfo.status === 'not_installed'">
          <template #header>
            <div class="card-header">
              <span>安装帮助</span>
              <el-icon><QuestionFilled /></el-icon>
            </div>
          </template>
          
          <div class="install-help-content">
            <el-empty 
              description="服务未安装，请先安装 Sing-Box 服务"
              :image-size="100">
              <el-button type="primary" @click="openLink('https://sing-box.sagernet.org/installation/')">
                查看安装指南
              </el-button>
            </el-empty>
          </div>
        </el-card>
        
        <el-card class="help-card" v-else-if="loading && !serviceInfo">
          <el-skeleton :rows="4" animated />
        </el-card>
        
        <el-card class="kernel-info-card" v-else>
          <template #header>
            <div class="card-header">
              <div class="header-left">
                <span class="service-title">内核信息</span>
                <div v-if="kernelInfo" class="inline-status">
                  <el-tag type="primary" size="small" class="kernel-tag">
                    <el-icon class="kernel-icon"><Monitor /></el-icon>
                    {{ kernelInfo.name }}
                  </el-tag>
                  <el-tooltip v-if="kernelInfo.version && kernelInfo.version !== 'Unknown'" 
                              :content="'完整版本: ' + kernelInfo.version" 
                              placement="top">
                    <el-tag type="success" size="small" class="version-tag">
                      {{ kernelInfo.version }}
                    </el-tag>
                  </el-tooltip>
                  <el-tag v-if="kernelInfo.platform && kernelInfo.platform !== 'Unknown'" 
                          type="info" size="small" class="platform-tag">
                    {{ kernelInfo.platform }}
                  </el-tag>
                  <el-tag v-if="kernelInfo.branch && kernelInfo.branch !== ''" 
                          type="warning" size="small" class="branch-tag">
                    {{ kernelInfo.branch }}
                  </el-tag>
                </div>
                <div v-else class="inline-status">
                  <el-icon class="is-loading"><Loading /></el-icon>
                  <span>获取内核信息中...</span>
                </div>
              </div>
              <div class="header-controls">
                <el-button 
                  type="primary" 
                  size="small"
                  @click="showUpdateKernelDialog"
                  :loading="updateLoading">
                  <el-icon><Download /></el-icon>
                  更新内核
                </el-button>
              </div>
            </div>
          </template>
        </el-card>
      </el-col>
    </el-row>


    <!-- 配置编辑器对话框 -->
    <el-dialog 
      v-model="configDialogVisible" 
      title="Sing-Box 配置编辑器"
      width="80%"
      :before-close="closeConfigDialog"
      class="config-dialog">
      
      <div v-loading="configLoading" class="config-editor-container">
        <div class="editor-toolbar">
          <div class="toolbar-left">
            <el-button 
              type="primary" 
              @click="validateConfig" 
              :disabled="configLoading || configSaving">
              <el-icon><CircleCheck /></el-icon>
              验证配置
            </el-button>
            <el-button 
              @click="resetConfig" 
              :disabled="configLoading || configSaving">
              <el-icon><RefreshRight /></el-icon>
              重置
            </el-button>
          </div>
          <div class="toolbar-right">
            <div class="config-status">
              <el-tag v-if="configValidationStatus === 'valid'" type="success" size="small">
                <el-icon><CircleCheck /></el-icon>
                验证通过
              </el-tag>
              <el-tag v-else-if="configValidationStatus === 'invalid'" type="danger" size="small">
                <el-icon><CircleClose /></el-icon>
                验证失败
              </el-tag>
              <el-tag v-else-if="configValidationStatus === 'warning'" type="warning" size="small">
                <el-icon><Warning /></el-icon>
                有警告
              </el-tag>
            </div>
            <span class="config-path">/etc/sing-box/config.json</span>
          </div>
        </div>
        
        <div class="editor-content">
          <el-input
            v-model="configContent"
            type="textarea"
            :rows="20"
            placeholder="配置内容..."
            :disabled="configLoading || configSaving"
            class="config-textarea" />
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeConfigDialog" :disabled="configSaving">
            取消
          </el-button>
          <el-button 
            type="primary" 
            @click="saveConfig" 
            :loading="configSaving"
            :disabled="configLoading">
            保存配置
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 内核更新对话框 -->
    <el-dialog
      v-model="updateKernelDialogVisible"
      title="更新 Sing-Box 内核"
      width="600px"
      :close-on-click-modal="false">
      
      <div class="kernel-update-content">
        <div class="current-info">
          <h4>当前内核信息</h4>
          <div class="info-grid">
            <div class="info-item">
              <span class="label">版本:</span>
              <span class="value">{{ kernelInfo?.version || 'Unknown' }}</span>
            </div>
            <div class="info-item">
              <span class="label">平台:</span>
              <span class="value">{{ kernelInfo?.platform || 'Unknown' }}</span>
            </div>
            <div class="info-item">
              <span class="label">分支:</span>
              <span class="value">{{ kernelInfo?.branch || 'Unknown' }}</span>
            </div>
            <div class="info-item path-info">
              <span class="label">安装路径:</span>
              <pre class="value path-value">{{ updateInfo.currentPath || '检测中...' }}</pre>
            </div>
          </div>
        </div>

        <div v-if="updateInfo.latestVersion" class="latest-info">
          <h4>最新版本信息</h4>
          <div class="info-grid">
            <div class="info-item">
              <span class="label">版本:</span>
              <span class="value">{{ updateInfo.latestVersion }}</span>
            </div>
            <div class="info-item">
              <span class="label">发布时间:</span>
              <span class="value">{{ updateInfo.publishTime }}</span>
            </div>
            <div class="info-item">
              <span class="label">下载URL:</span>
              <span class="value download-url">{{ updateInfo.downloadUrl }}</span>
            </div>
          </div>
        </div>

        <div v-if="updateProgress.show" class="update-progress">
          <h4>更新进度</h4>
          <el-progress 
            :percentage="updateProgress.percentage" 
            :status="updateProgress.status"
            :stroke-width="8">
            <span class="progress-text">{{ updateProgress.text }}</span>
          </el-progress>
          <div v-if="updateProgress.logs.length > 0" class="progress-logs">
            <div v-for="(log, index) in updateProgress.logs" :key="index" class="log-item">
              <span class="log-time">{{ log.time }}</span>
              <span class="log-message">{{ log.message }}</span>
            </div>
          </div>
        </div>

        <div v-if="updateInfo.error" class="update-error">
          <el-alert
            :title="updateInfo.error"
            type="error"
            :closable="false">
          </el-alert>
        </div>
      </div>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="closeUpdateDialog" :disabled="updateProgress.show && updateProgress.percentage < 100">
            取消
          </el-button>
          <el-button 
            v-if="!updateProgress.show"
            type="primary" 
            @click="checkLatestVersion"
            :loading="updateLoading">
            检查更新
          </el-button>
          <el-button 
            v-if="updateInfo.latestVersion && !updateProgress.show"
            type="success" 
            @click="startKernelUpdate"
            :loading="updateLoading">
            开始更新
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, h, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Refresh, Monitor, Setting, Document, Link, VideoPlay, VideoPause, 
  RefreshRight, Edit, CircleCheck, CircleClose, Warning, Delete, QuestionFilled, Loading, Download
} from '@element-plus/icons-vue'
import { 
  apiGetServiceInfo, 
  apiControlService, 
  apiGetSingBoxConfig, 
  apiUpdateSingBoxConfig, 
  apiValidateSingBoxConfig,
  apiDetectSingBoxPath,
  apiCheckSingBoxUpdate,
  apiUpdateSingBoxKernel
} from '../utils/api'

const loading = ref(true)
const actionLoading = ref('')
const serviceInfo = ref(null)
const kernelInfo = ref(null)

// 内核更新相关
const updateKernelDialogVisible = ref(false)
const updateLoading = ref(false)
const updateInfo = reactive({
  currentPath: '',
  latestVersion: '',
  publishTime: '',
  downloadUrl: '',
  error: ''
})
const updateProgress = reactive({
  show: false,
  percentage: 0,
  status: '',
  text: '',
  logs: []
})

// EventSource 引用，用于清理
let currentEventSource = null

// 配置编辑器相关状态
const configDialogVisible = ref(false)
const configContent = ref('')
const configLoading = ref(false)
const configSaving = ref(false)
const originalConfig = ref('')
const configValidationStatus = ref('') // 'valid', 'invalid', 'warning', ''

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
    const response = await apiGetServiceInfo('sing-box')
    serviceInfo.value = response.data
    
    // 设置内核信息
    if (response.data && response.data.version) {
      kernelInfo.value = {
        name: 'Sing-Box',
        version: response.data.version,
        platform: response.data.platform || 'Unknown',
        branch: response.data.branch || ''
      }
    } else {
      // 如果服务信息中没有版本信息，设置默认值
      kernelInfo.value = {
        name: 'Sing-Box',
        version: 'Unknown',
        platform: 'Unknown',
        branch: ''
      }
    }
  } catch (error) {
    console.error('获取服务信息失败:', error)
    ElMessage.error('获取服务信息失败')
    // 错误时也设置默认的内核信息
    kernelInfo.value = {
      name: 'Sing-Box',
      version: 'Error',
      platform: 'Unknown',
      branch: ''
    }
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
      `确定要${actionNames[action]} Sing-Box 服务吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    actionLoading.value = action
    const response = await apiControlService('sing-box', action)
    
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

const openConfigEditor = async () => {
  try {
    configLoading.value = true
    configDialogVisible.value = true
    
    const response = await apiGetSingBoxConfig()
    const config = response.data
    
    // 将配置对象格式化为JSON字符串
    configContent.value = JSON.stringify(config, null, 2)
    originalConfig.value = configContent.value
    
  } catch (error) {
    console.error('获取配置失败:', error)
    ElMessage.error('获取配置失败: ' + (error.response?.data?.message || error.message))
    configDialogVisible.value = false
  } finally {
    configLoading.value = false
  }
}

const validateConfig = async () => {
  try {
    // 重置验证状态
    configValidationStatus.value = ''
    
    // 首先进行基本的JSON格式验证
    const configObject = JSON.parse(configContent.value)
    
    // 调用后端API进行详细验证
    const response = await apiValidateSingBoxConfig(configObject)
    const result = response.data
    
    if (result.valid) {
      if (result.warnings && result.warnings.length > 0) {
        configValidationStatus.value = 'warning'
        let message = '配置验证通过，但有警告：\n' + result.warnings.join('\n')
        ElMessage({
          message,
          type: 'warning',
          duration: 5000,
          showClose: true
        })
      } else {
        configValidationStatus.value = 'valid'
        ElMessage.success('✅ Sing-Box 配置验证通过')
      }
      return true
    } else {
      configValidationStatus.value = 'invalid'
      let errorMessage = '配置验证失败'
      if (result.errors && result.errors.length > 0) {
        errorMessage += '\n\n错误：\n' + result.errors.join('\n')
      }
      if (result.warnings && result.warnings.length > 0) {
        errorMessage += '\n\n警告：\n' + result.warnings.join('\n')
      }
      
      ElMessage({
        message: errorMessage,
        type: 'error',
        duration: 8000,
        showClose: true
      })
      return false
    }
  } catch (error) {
    configValidationStatus.value = 'invalid'
    let errorMessage = '配置验证失败'
    if (error.name === 'SyntaxError') {
      errorMessage = 'JSON格式错误: ' + error.message
    } else if (error.response?.data?.message) {
      errorMessage = error.response.data.message
    } else {
      errorMessage = '验证过程中发生错误: ' + error.message
    }
    
    ElMessage.error(errorMessage)
    return false
  }
}

const saveConfig = async () => {
  // 检查是否已经验证通过
  if (configValidationStatus.value !== 'valid' && configValidationStatus.value !== 'warning') {
    ElMessage({
      message: '请先验证配置后再保存',
      type: 'warning'
    })
    return
  }
  
  try {
    // 询问用户保存选项
    const action = await ElMessageBox({
      title: '配置保存选项',
      message: h('div', [
        h('p', '请选择保存后的操作：'),
        h('div', { style: 'margin: 15px 0;' }, [
          h('label', { style: 'display: block; margin: 8px 0;' }, [
            h('input', { 
              type: 'radio', 
              name: 'saveOption', 
              value: 'save_only',
              style: 'margin-right: 8px;',
              checked: true
            }),
            '仅保存配置（需要手动重启服务）'
          ]),
          h('label', { style: 'display: block; margin: 8px 0;' }, [
            h('input', { 
              type: 'radio', 
              name: 'saveOption', 
              value: 'save_and_restart',
              style: 'margin-right: 8px;'
            }),
            '保存并自动重启服务'
          ]),
          h('label', { style: 'display: block; margin: 8px 0;' }, [
            h('input', { 
              type: 'radio', 
              name: 'saveOption', 
              value: 'save_restart_rollback',
              style: 'margin-right: 8px;'
            }),
            '保存、重启并启用自动回滚（推荐）'
          ])
        ]),
        h('p', { style: 'color: #909399; font-size: 12px; margin-top: 10px;' }, 
          '自动回滚：如果重启后服务无法正常启动，将自动恢复之前的配置')
      ]),
      showCancelButton: true,
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      beforeClose: (action, instance, done) => {
        if (action === 'confirm') {
          const selected = document.querySelector('input[name="saveOption"]:checked')?.value || 'save_only'
          done(selected)
        } else {
          done()
        }
      }
    })
    
    if (!action || action === 'cancel') {
      return
    }
    
    configSaving.value = true
    
    // 解析配置内容
    const configObject = JSON.parse(configContent.value)
    
    // 根据用户选择设置保存选项
    const options = {
      backup: true,
      autoRestart: action === 'save_and_restart' || action === 'save_restart_rollback',
      enableRollback: action === 'save_restart_rollback'
    }
    
    const response = await apiUpdateSingBoxConfig(configObject, options)
    
    if (response.data.success) {
      originalConfig.value = configContent.value
      configDialogVisible.value = false
      
      if (response.data.restarting) {
        ElMessage({
          message: '配置保存成功，正在重启服务，请稍候...',
          type: 'success',
          duration: 3000
        })
        
        // 延迟刷新服务状态
        setTimeout(() => {
          refreshData()
        }, 8000)
      } else if (response.data.needs_restart) {
        ElMessageBox.alert(
          '✅ Sing-Box 配置验证成功并已保存！\n\n⚠️ 重要提醒：配置更改需要重启 Sing-Box 服务才能生效。\n\n您可以在服务状态区域手动重启服务。',
          '配置验证成功',
          {
            confirmButtonText: '我知道了',
            type: 'success',
            dangerouslyUseHTMLString: false
          }
        )
        ElMessage.success('✅ 配置验证成功并已保存，请重启服务使配置生效')
      } else {
        ElMessage.success('✅ 配置验证成功并已保存')
      }
    } else {
      ElMessage.error(response.data.message || '配置保存失败')
    }
    
  } catch (error) {
    if (error !== 'cancel') {
      console.error('保存配置失败:', error)
      ElMessage.error('保存配置失败: ' + (error.response?.data?.message || error.message))
    }
  } finally {
    configSaving.value = false
  }
}

const resetConfig = () => {
  configContent.value = originalConfig.value
  ElMessage.info('已重置为原始配置')
}

const closeConfigDialog = () => {
  if (configContent.value !== originalConfig.value) {
    ElMessageBox.confirm(
      '配置已修改但未保存，确定要关闭吗？',
      '确认关闭',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    ).then(() => {
      configDialogVisible.value = false
      configContent.value = ''
      originalConfig.value = ''
    }).catch(() => {
      // 用户取消关闭
    })
  } else {
    configDialogVisible.value = false
    configContent.value = ''
    originalConfig.value = ''
  }
}


// 监听配置内容变化，重置验证状态
watch(configContent, () => {
  if (configValidationStatus.value) {
    configValidationStatus.value = ''
  }
})

// 内核更新相关函数
const showUpdateKernelDialog = () => {
  updateKernelDialogVisible.value = true
  resetUpdateInfo()
  detectCurrentPath()
}


const resetUpdateInfo = () => {
  updateInfo.currentPath = ''
  updateInfo.latestVersion = ''
  updateInfo.publishTime = ''
  updateInfo.downloadUrl = ''
  updateInfo.error = ''
  updateProgress.show = false
  updateProgress.percentage = 0
  updateProgress.status = ''
  updateProgress.text = ''
  updateProgress.logs = []
}

const detectCurrentPath = async () => {
  console.log('开始检测当前路径...')
  updateInfo.currentPath = '检测中...'
  
  try {
    console.log('调用 apiDetectSingBoxPath...')
    const response = await apiDetectSingBoxPath()
    console.log('API 响应:', response)
    
    // 由于axios拦截器已经返回了response.data，所以response就是实际的数据
    console.log('响应结构分析:', {
      hasResponse: !!response,
      hasPath: !!(response && response.path),
      fullResponse: response
    })
    
    if (response && response.path) {
      updateInfo.currentPath = response.path
      console.log('成功检测到路径:', response.path)
    } else {
      updateInfo.currentPath = '未检测到路径'
      console.log('响应中没有路径信息:', response)
    }
  } catch (error) {
    console.error('路径检测详细错误:', error)
    
    // 安全地处理错误响应
    // 由于axios拦截器，错误响应的数据直接在error.response中
    if (error.response) {
      const errorData = error.response
      let debugText = '检测失败\n'
      
      // 显示主要错误信息
      if (errorData.error) {
        debugText += '错误: ' + errorData.error + '\n\n'
      }
      
      // 显示调试信息
      if (errorData.debug_info && Array.isArray(errorData.debug_info)) {
        debugText += '调试信息:\n'
        errorData.debug_info.forEach(info => {
          debugText += '- ' + info + '\n'
        })
      }
      
      // 显示检测方法
      if (errorData.detection_methods && Array.isArray(errorData.detection_methods)) {
        debugText += '\n检测方法:\n'
        errorData.detection_methods.forEach(method => {
          debugText += '- ' + method + '\n'
        })
      }
      
      // 显示检查过的路径
      if (errorData.checked_paths && Array.isArray(errorData.checked_paths)) {
        debugText += '\n检查过的路径:\n'
        errorData.checked_paths.forEach(path => {
          debugText += '- ' + path + '\n'
        })
      }
      
      updateInfo.currentPath = debugText
    } else if (error.message) {
      updateInfo.currentPath = '检测失败: ' + error.message
    } else {
      updateInfo.currentPath = '检测失败: 未知错误'
    }
  }
}

const checkLatestVersion = async () => {
  try {
    updateLoading.value = true
    updateInfo.error = ''
    
    const response = await apiCheckSingBoxUpdate()
    
    if (response.hasUpdate) {
      updateInfo.latestVersion = response.version
      updateInfo.publishTime = response.publishTime
      updateInfo.downloadUrl = response.downloadUrl
      ElMessage.success('发现新版本: ' + response.version)
    } else {
      ElMessage.info('当前已是最新版本')
    }
  } catch (error) {
    updateInfo.error = '检查更新失败: ' + (error.response?.error || error.message)
  } finally {
    updateLoading.value = false
  }
}

const startKernelUpdate = async () => {
  try {
    // 关闭之前的EventSource（如果存在）
    if (currentEventSource) {
      currentEventSource.close()
      currentEventSource = null
    }
    
    updateLoading.value = true
    updateProgress.show = true
    updateProgress.percentage = 0
    updateProgress.status = ''
    updateProgress.text = '准备更新...'
    updateProgress.logs = []
    
    addProgressLog('开始更新内核...')
    
    // 创建 SSE 连接来实时接收更新进度
    currentEventSource = new EventSource('/api/singbox/kernel/update-stream')
    
    currentEventSource.onmessage = (event) => {
      const data = JSON.parse(event.data)
      updateProgress.percentage = data.percentage || 0
      updateProgress.text = data.message || ''
      
      if (data.log) {
        addProgressLog(data.log)
      }
      
      if (data.status) {
        updateProgress.status = data.status
      }
      
      if (data.finished) {
        currentEventSource.close()
        currentEventSource = null
        updateLoading.value = false
        if (data.success) {
          ElMessage.success('内核更新成功！')
          setTimeout(() => {
            updateKernelDialogVisible.value = false
            refreshData() // 刷新服务信息
          }, 2000)
        } else {
          updateInfo.error = data.error || '更新失败'
        }
      }
    }
    
    currentEventSource.onerror = () => {
      if (currentEventSource) {
        currentEventSource.close()
        currentEventSource = null
      }
      updateLoading.value = false
      updateInfo.error = '更新连接中断'
    }
    
    // 启动更新
    await apiUpdateSingBoxKernel({
      downloadUrl: updateInfo.downloadUrl,
      targetPath: updateInfo.currentPath
    })
    
  } catch (error) {
    // 确保在错误情况下也关闭EventSource
    if (currentEventSource) {
      currentEventSource.close()
      currentEventSource = null
    }
    updateLoading.value = false
    updateProgress.show = false
    updateInfo.error = '更新失败: ' + (error.response?.error || error.message)
  }
}

const addProgressLog = (message) => {
  updateProgress.logs.push({
    time: new Date().toLocaleTimeString(),
    message: message
  })
}

const closeUpdateDialog = () => {
  if (updateProgress.show && updateProgress.percentage < 100) {
    ElMessageBox.confirm('更新正在进行中，确定要关闭吗？', '确认关闭', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(() => {
      // 关闭EventSource连接
      if (currentEventSource) {
        currentEventSource.close()
        currentEventSource = null
      }
      updateKernelDialogVisible.value = false
      resetUpdateInfo()
    })
  } else {
    // 关闭EventSource连接
    if (currentEventSource) {
      currentEventSource.close()
      currentEventSource = null
    }
    updateKernelDialogVisible.value = false
    resetUpdateInfo()
  }
}

onMounted(() => {
  refreshData()
})

onUnmounted(() => {
  // 清理EventSource连接
  if (currentEventSource) {
    currentEventSource.close()
    currentEventSource = null
  }
})
</script>

<style scoped>
.sing-box-manage {
  padding: 15px;
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
  margin-bottom: 15px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.service-title {
  font-weight: 500;
  color: #303133;
}

.inline-status {
  display: flex;
  align-items: center;
  gap: 8px;
}

.uptime-text {
  font-size: 14px;
  color: #606266;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}

.header-controls {
  display: flex;
  gap: 6px;
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
  min-height: 20px;
}

.status-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding: 12px;
  background-color: #f8f9fa;
  border-radius: 6px;
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
  padding: 12px;
  background-color: #f0f9ff;
  border-radius: 6px;
  border: 1px solid #e1f5fe;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  padding: 4px 0;
}

.detail-item:last-child {
  margin-bottom: 0;
}

.not-installed-content {
  text-align: center;
  padding: 15px;
}

.status-message {
  margin-bottom: 15px;
  font-weight: 500;
  color: #909399;
}

.install-guide {
  border-top: 1px solid #e4e7ed;
  padding-top: 12px;
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
  padding: 20px;
  color: #909399;
  background-color: #fafafa;
  border-radius: 6px;
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

.config-section {
  margin-bottom: 15px;
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

.kernel-tag {
  display: flex;
  align-items: center;
  gap: 3px;
  font-weight: 500;
}

.kernel-icon {
  font-size: 12px;
}

.version-tag {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-weight: 500;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: help;
}

.version-tag:hover {
  max-width: none;
  white-space: normal;
  word-break: break-all;
}

.platform-tag {
  font-size: 11px;
}

.branch-tag {
  font-size: 11px;
  font-weight: 500;
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

/* 配置编辑器对话框样式 */
.config-dialog {
  .el-dialog__body {
    padding: 20px;
  }
}

.config-editor-container {
  height: 600px;
  display: flex;
  flex-direction: column;
}

.editor-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding: 10px;
  background-color: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
}

.toolbar-left {
  display: flex;
  gap: 10px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: #606266;
}

.config-status {
  display: flex;
  align-items: center;
}

.config-path {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  background-color: #e4e7ed;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.editor-content {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.config-textarea {
  flex: 1;
}

.config-textarea .el-textarea__inner {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
  resize: none;
  height: 100% !important;
  min-height: 500px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

/* 内核更新对话框样式 */
.kernel-update-content {
  padding: 10px 0;
}

.current-info, .latest-info, .update-progress {
  margin-bottom: 20px;
}

.current-info h4, .latest-info h4, .update-progress h4 {
  margin: 0 0 10px 0;
  color: #409eff;
  font-size: 14px;
  font-weight: 600;
}

.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.info-item {
  display: flex;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 4px;
  border-left: 3px solid #409eff;
}

.info-item .label {
  font-weight: 600;
  color: #606266;
  min-width: 60px;
  margin-right: 8px;
}

.info-item .value {
  color: #303133;
  word-break: break-all;
}

.download-url {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
}

.progress-logs {
  margin-top: 15px;
  max-height: 200px;
  overflow-y: auto;
  background: #f5f7fa;
  border-radius: 4px;
  padding: 10px;
}

.log-item {
  display: flex;
  margin-bottom: 5px;
  font-size: 13px;
}

.log-time {
  color: #909399;
  margin-right: 10px;
  min-width: 80px;
}

.log-message {
  color: #303133;
}

.progress-text {
  font-size: 12px;
  color: #606266;
}

.update-error {
  margin-top: 15px;
}

.header-controls {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 路径信息特殊样式 */
.path-info {
  grid-column: 1 / -1; /* 占据整个宽度 */
  flex-direction: column !important;
  align-items: flex-start !important;
}

.path-value {
  width: 100%;
  margin: 5px 0 0 0;
  padding: 8px;
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  line-height: 1.4;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 150px;
  overflow-y: auto;
}
</style>
