<template>
  <div class="sing-box-inbound">
    <div class="page-header">
      <h2>入站设置</h2>
    </div>

    <!-- 服务状态 -->
    <ServiceStatus service-name="sing-box" />

    <!-- 入站规则列表 -->
    <el-card class="inbound-list-card">
      <template #header>
        <div class="card-header">
          <span>入站规则列表</span>
          <el-button type="text" @click="refreshInbounds">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>

      <el-table :data="inbounds" stripe>
        <el-table-column prop="tag" label="标签" width="150" />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="scope">
            <el-tag :type="getTypeTagType(scope.row.type)">
              {{ scope.row.type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="listen" label="监听地址" width="150" />
        <el-table-column prop="listen_port" label="端口" width="100" />
        <el-table-column prop="protocol" label="协议" width="120" />
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.enabled ? 'success' : 'danger'">
              {{ scope.row.enabled ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="!inbounds.length" class="empty-state">
        <el-empty description="暂无入站规则" />
      </div>
    </el-card>

  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { apiGetSingBoxInbounds } from '../utils/api'
import ServiceStatus from '../components/ServiceStatus.vue'

// 响应式数据
const inbounds = ref([])

// 获取类型标签样式
const getTypeTagType = (type) => {
  const typeMap = {
    'http': 'primary',
    'socks': 'success',
    'mixed': 'warning',
    'tun': 'danger',
    'shadowsocks': 'info'
  }
  return typeMap[type] || 'default'
}

// 刷新入站规则列表
const refreshInbounds = async () => {
  try {
    const response = await apiGetSingBoxInbounds()
    if (response.code === 200) {
      inbounds.value = response.data
    } else {
      throw new Error(response.message || '获取入站规则失败')
    }
  } catch (error) {
    console.error('获取入站规则失败:', error)
    ElMessage.error('获取入站规则失败: ' + (error.response?.data?.message || error.message))
    // 失败时使用空数组
    inbounds.value = []
  }
}

// 组件挂载时获取数据
onMounted(() => {
  refreshInbounds()
})
</script>

<style scoped>
.sing-box-inbound {
  padding: 15px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.page-header h2 {
  margin: 0;
  color: #303133;
}

.inbound-list-card {
  margin-bottom: 15px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.empty-state {
  text-align: center;
  padding: 40px 0;
}
</style>
