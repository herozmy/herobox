<template>
  <div class="logs">
    <div class="page-header">
      <h2>日志查看</h2>
      <div class="header-actions">
        <el-select v-model="currentService" @change="loadLogs" placeholder="选择服务">
          <el-option label="MosDNS" value="mosdns" />
          <el-option label="Sing-Box" value="sing-box" />
        </el-select>
        <el-input
          v-model="filterKeyword"
          placeholder="过滤关键词"
          style="width: 200px;"
          @change="loadLogs"
          clearable
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-select v-model="logLines" @change="loadLogs" style="width: 120px;">
          <el-option label="50行" :value="50" />
          <el-option label="100行" :value="100" />
          <el-option label="200行" :value="200" />
          <el-option label="500行" :value="500" />
          <el-option label="1000行" :value="1000" />
        </el-select>
        <el-button @click="loadLogs" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button @click="clearLogs">
          <el-icon><Delete /></el-icon>
          清空显示
        </el-button>
        <el-switch
          v-model="autoRefresh"
          active-text="自动刷新"
          inactive-text="手动刷新"
          @change="toggleAutoRefresh"
        />
      </div>
    </div>

    <el-card v-if="currentService" class="logs-card">
      <template #header>
        <div class="card-header">
          <span>{{ currentService === 'mosdns' ? 'MosDNS' : 'Sing-Box' }} 日志</span>
          <div class="log-info">
            <el-tag size="small" type="info">
              显示 {{ logData?.filtered_lines || 0 }} / {{ logData?.total_lines || 0 }} 行
            </el-tag>
            <el-tag v-if="autoRefresh" size="small" type="success">
              <el-icon><Clock /></el-icon>
              自动刷新中
            </el-tag>
          </div>
        </div>
      </template>

      <div class="logs-content">
        <div v-if="logData && logData.content" class="log-display">
          <pre class="log-text">{{ logData.content }}</pre>
        </div>
        
        <div v-else-if="loading" class="log-loading">
          <el-skeleton :rows="15" animated />
        </div>
        
        <div v-else class="log-empty">
          <el-empty description="暂无日志数据" />
        </div>
      </div>

      <!-- 日志控制工具栏 -->
      <div class="logs-toolbar">
        <div class="toolbar-left">
          <el-button size="small" @click="scrollToTop">
            <el-icon><Top /></el-icon>
            回到顶部
          </el-button>
          <el-button size="small" @click="scrollToBottom">
            <el-icon><Bottom /></el-icon>
            跳到底部
          </el-button>
        </div>
        
        <div class="toolbar-right">
          <el-button size="small" @click="copyLogs" :disabled="!logData?.content">
            <el-icon><DocumentCopy /></el-icon>
            复制日志
          </el-button>
          <el-button size="small" @click="downloadLogs" :disabled="!logData?.content">
            <el-icon><Download /></el-icon>
            下载日志
          </el-button>
        </div>
      </div>
    </el-card>

    <el-empty v-else description="请选择要查看的服务日志" />
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { apiGetLogs } from '../utils/api'
import { ElMessage } from 'element-plus'

const loading = ref(false)
const currentService = ref('mosdns')
const filterKeyword = ref('')
const logLines = ref(100)
const autoRefresh = ref(false)
const logData = ref(null)
const refreshTimer = ref(null)

const loadLogs = async () => {
  if (!currentService.value) return

  loading.value = true
  try {
    const response = await apiGetLogs(currentService.value, logLines.value, filterKeyword.value)
    if (response.code === 200) {
      logData.value = response.data
      
      // 如果开启自动刷新，滚动到底部
      if (autoRefresh.value) {
        nextTick(() => {
          scrollToBottom()
        })
      }
    }
  } catch (error) {
    ElMessage.error('加载日志失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const clearLogs = () => {
  logData.value = null
}

const toggleAutoRefresh = (enabled) => {
  if (enabled) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
}

const startAutoRefresh = () => {
  if (refreshTimer.value) return
  
  refreshTimer.value = setInterval(() => {
    loadLogs()
  }, 5000) // 每5秒刷新一次
}

const stopAutoRefresh = () => {
  if (refreshTimer.value) {
    clearInterval(refreshTimer.value)
    refreshTimer.value = null
  }
}

const scrollToTop = () => {
  const logDisplay = document.querySelector('.log-display')
  if (logDisplay) {
    logDisplay.scrollTop = 0
  }
}

const scrollToBottom = () => {
  const logDisplay = document.querySelector('.log-display')
  if (logDisplay) {
    logDisplay.scrollTop = logDisplay.scrollHeight
  }
}

const copyLogs = async () => {
  if (!logData.value?.content) return
  
  try {
    await navigator.clipboard.writeText(logData.value.content)
    ElMessage.success('日志已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

const downloadLogs = () => {
  if (!logData.value?.content) return
  
  const content = logData.value.content
  const blob = new Blob([content], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  
  const a = document.createElement('a')
  a.href = url
  a.download = `${currentService.value}-logs-${new Date().toISOString().slice(0, 19)}.txt`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  
  ElMessage.success('日志下载成功')
}

onMounted(() => {
  loadLogs()
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped>
.logs {
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

.header-actions {
  display: flex;
  gap: 10px;
  align-items: center;
}

.logs-card {
  height: calc(100vh - 180px);
  display: flex;
  flex-direction: column;
}

.logs-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.log-display {
  flex: 1;
  background-color: #1e1e1e;
  border-radius: 6px;
  overflow: auto;
  padding: 15px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.4;
}

.log-text {
  color: #d4d4d4;
  margin: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.log-loading {
  padding: 20px;
}

.log-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
}

.logs-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #e4e7ed;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  gap: 10px;
}

.log-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header span {
  font-weight: 500;
}

/* 滚动条样式 */
.log-display::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.log-display::-webkit-scrollbar-track {
  background: #2d2d2d;
  border-radius: 4px;
}

.log-display::-webkit-scrollbar-thumb {
  background: #555;
  border-radius: 4px;
}

.log-display::-webkit-scrollbar-thumb:hover {
  background: #777;
}
</style>
