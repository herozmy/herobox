<template>
  <div class="sing-box-proxy">
    <div class="page-header">
      <h2>代理/出站</h2>
    </div>

    <!-- 服务状态 -->
    <ServiceStatus service-name="sing-box" />

    <!-- 节点分组和列表 -->
    <el-card class="proxy-list-card">
      <template #header>
        <div class="card-header">
          <span>代理/出站管理</span>
          <div class="header-actions">
            <el-button 
              v-if="activeGroup === '代理'" 
              type="primary" 
              size="small" 
              @click="addProxy">
              <el-icon><Plus /></el-icon>
              添加节点
            </el-button>
            <el-button 
              v-if="activeGroup === '应用分流'" 
              type="primary" 
              size="small" 
              @click="addAppProxy">
              <el-icon><Plus /></el-icon>
              添加分流
            </el-button>
            <el-button 
              v-if="activeGroup === '节点过滤'" 
              type="primary" 
              size="small" 
              @click="addFilterProxy">
              <el-icon><Plus /></el-icon>
              添加过滤器
            </el-button>
            <el-button type="text" @click="testAllNodes">
              <el-icon><Connection /></el-icon>
              测试所有节点
            </el-button>
            <el-button type="text" @click="refreshNodes">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </div>
      </template>

      <!-- 协议分组选择 -->
      <div class="group-selector">
        <el-tabs v-model="activeGroup" @tab-change="handleGroupChange" type="card">
          <el-tab-pane 
            v-for="group in nodeGroups" 
            :key="group.name"
            :name="group.name">
            <template #label>
              <div class="tab-label">
                <span class="protocol-icon">{{ getProtocolIcon(group.name) }}</span>
                <span class="protocol-name">{{ group.name }}</span>
                <span class="node-count">({{ group.nodes.length }})</span>
              </div>
            </template>
          </el-tab-pane>
        </el-tabs>
      </div>

      <!-- 节点列表 -->
      <el-table :data="currentGroupNodes" stripe>
        <el-table-column label="状态" width="80">
          <template #default="scope">
            <el-tag type="success" size="small">已保存</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="tag" label="节点名称" width="200">
          <template #default="scope">
            <div class="node-name">
              {{ scope.row.tag }}
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="120">
          <template #default="scope">
            <el-tag :type="getProxyTypeTag(scope.row.type)">
              {{ scope.row.type.toUpperCase() }}
            </el-tag>
          </template>
        </el-table-column>
        <!-- 只在代理分组显示服务器、端口列 -->
        <el-table-column v-if="activeGroup === '代理'" prop="server" label="服务器" width="150" />
        <el-table-column v-if="activeGroup === '代理'" prop="server_port" label="端口" width="80" />
        <el-table-column label="操作" :width="activeGroup === '代理' ? 250 : 180">
          <template #default="scope">
            <!-- 只在代理分组显示延迟测试按钮 -->
            <el-button 
              v-if="activeGroup === '代理'"
              type="primary" 
              size="small" 
              @click="testNode(scope.row)"
              :loading="scope.row.testing">
              <el-icon><Connection /></el-icon>
              测试
            </el-button>
            <el-button 
              type="warning" 
              size="small" 
              @click="editProxy(scope.row)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button 
              type="danger" 
              size="small" 
              @click="deleteProxy(scope.row)">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="!currentGroupNodes.length" class="empty-state">
        <el-empty description="当前分组暂无节点">
          <el-button type="primary" @click="addProxy">添加第一个节点</el-button>
        </el-empty>
      </div>
    </el-card>

    <!-- 添加/编辑节点对话框 -->
    <el-dialog 
      v-model="dialogVisible" 
      :title="getDialogTitle()"
      width="600px">
      <el-form :model="proxyForm" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item label="节点名称" prop="name">
          <el-input v-model="proxyForm.name" placeholder="请输入节点名称" />
        </el-form-item>
        
        <!-- 应用分流节点的编辑界面 -->
        <template v-if="activeGroup === '应用分流'">
          <el-form-item label="节点类型" prop="type">
            <el-select v-model="proxyForm.type" placeholder="请选择节点类型" :disabled="!!editingProxy">
              <el-option label="Selector" value="selector" />
              <el-option label="URLTest" value="urltest" />
              <el-option label="LoadBalance" value="loadbalance" />
            </el-select>
          </el-form-item>
          
          <el-form-item label="出站节点" prop="outbounds">
            <div class="outbound-selection">
              <div class="selection-header">
                <span>选择出站节点：</span>
                <el-button size="small" @click="selectAllOutbounds">全选</el-button>
                <el-button size="small" @click="clearAllOutbounds">清空</el-button>
              </div>
              
              <!-- 当前已选择的出站节点预览 -->
              <div v-if="proxyForm.outbounds && proxyForm.outbounds.length > 0" class="current-selection">
                <div class="selection-title">
                  <span>已选择 ({{ proxyForm.outbounds.length }})：</span>
                </div>
                <div class="selected-tags">
                  <el-tag 
                    v-for="outbound in proxyForm.outbounds" 
                    :key="outbound"
                    size="small" 
                    closable
                    @close="removeOutbound(outbound)"
                    class="selected-tag">
                    {{ outbound }}
                  </el-tag>
                </div>
              </div>
              <div v-else-if="isEditingAppNode()" class="no-selection">
                <span class="text-muted">当前没有选择任何出站节点</span>
              </div>
              
              <div class="outbound-list">
                <el-checkbox-group v-model="proxyForm.outbounds">
                  <!-- 系统内置节点分组 -->
                  <div class="outbound-group">
                    <div class="group-title">系统节点 (2)</div>
                    <div class="group-nodes">
                      <el-checkbox 
                        label="direct"
                        class="outbound-checkbox">
                        <span class="node-info">
                          <span class="node-name">direct</span>
                          <el-tag size="small" type="success">
                            DIRECT
                          </el-tag>
                        </span>
                      </el-checkbox>
                      <el-checkbox 
                        label="block"
                        class="outbound-checkbox">
                        <span class="node-info">
                          <span class="node-name">block</span>
                          <el-tag size="small" type="danger">
                            BLOCK
                          </el-tag>
                        </span>
                      </el-checkbox>
                    </div>
                  </div>
                  
                  <!-- 其他节点分组 -->
                  <div v-if="availableOutbounds.length === 0" class="no-outbounds">
                    <el-empty description="暂无可用的代理节点" :image-size="60" />
                  </div>
                  <div v-else>
                    <div v-for="group in availableOutbounds" :key="group.name" class="outbound-group">
                      <div class="group-title">{{ group.name }} ({{ group.nodes?.length || 0 }})</div>
                      <div v-if="group.nodes && group.nodes.length > 0" class="group-nodes">
                        <el-checkbox 
                          v-for="node in group.nodes" 
                          :key="node.tag"
                          :label="node.tag"
                          class="outbound-checkbox">
                          <span class="node-info">
                            <span class="node-name">{{ node.tag }}</span>
                            <el-tag size="small" :type="getProxyTypeTag(node.type)">
                              {{ node.type?.toUpperCase() || 'UNKNOWN' }}
                            </el-tag>
                          </span>
                        </el-checkbox>
                      </div>
                      <div v-else class="no-nodes">
                        <span class="text-muted">该分组暂无节点</span>
                      </div>
                    </div>
                  </div>
                </el-checkbox-group>
              </div>
            </div>
          </el-form-item>
          
          <el-form-item label="默认出站" prop="default">
            <el-select v-model="proxyForm.default" placeholder="请选择默认出站节点">
              <el-option 
                v-for="outbound in getAllOutboundOptions()" 
                :key="outbound.tag" 
                :label="getOutboundDisplayName(outbound.tag)" 
                :value="outbound.tag" />
            </el-select>
          </el-form-item>
        </template>
        
        <!-- 节点过滤器的编辑界面 -->
        <template v-if="activeGroup === '节点过滤'">
          <el-form-item label="包含" prop="include">
            <el-input 
              v-model="proxyForm.include" 
              placeholder="例如: (?i)日本|东京|JP|Japan" 
              type="textarea" 
              :rows="2" />
            <div style="font-size: 12px; color: #999; margin-top: 4px;">
              正则表达式，匹配包含这些关键字的节点名称
            </div>
          </el-form-item>
          
          <el-form-item label="排除" prop="exclude">
            <el-input 
              v-model="proxyForm.exclude" 
              placeholder="例如: (?i)到期|过期|expire" 
              type="textarea" 
              :rows="2" />
            <div style="font-size: 12px; color: #999; margin-top: 4px;">
              正则表达式，排除包含这些关键字的节点名称
            </div>
          </el-form-item>
          
          <el-form-item label="使用所有订阅">
            <el-checkbox v-model="proxyForm.useAllProviders">使用所有订阅节点作为筛选范围</el-checkbox>
          </el-form-item>
          
          <el-form-item label="手动指定节点" prop="outbounds">
            <el-select v-model="proxyForm.outbounds" placeholder="可选：手动指定特定节点" multiple>
              <el-option-group label="系统节点">
                <el-option label="直连" value="direct" />
                <el-option label="拒绝" value="block" />
              </el-option-group>
              <el-option-group 
                v-for="group in availableOutbounds" 
                :key="group.name" 
                :label="group.name">
                <el-option 
                  v-for="node in group.nodes" 
                  :key="node.tag" 
                  :label="node.tag" 
                  :value="node.tag" />
              </el-option-group>
            </el-select>
            <div style="font-size: 12px; color: #999; margin-top: 4px;">
              可选项，不选择时将根据上述筛选条件自动匹配节点
            </div>
          </el-form-item>
          
          <el-form-item label="模式" prop="type">
            <el-select v-model="proxyForm.type" placeholder="请选择模式">
              <el-option label="selector" value="selector" />
              <el-option label="urltest" value="urltest" />
              <el-option label="loadbalance" value="loadbalance" />
            </el-select>
          </el-form-item>
        </template>
        
        <!-- 代理节点的编辑界面 -->
        <template v-else-if="activeGroup === '代理'">
          <el-form-item label="代理类型" prop="type">
            <el-select v-model="proxyForm.type" placeholder="请选择代理类型">
              <el-option label="Shadowsocks" value="shadowsocks" />
              <el-option label="VMess" value="vmess" />
              <el-option label="VLESS" value="vless" />
              <el-option label="Trojan" value="trojan" />
              <el-option label="Hysteria" value="hysteria" />
            </el-select>
          </el-form-item>
          <el-form-item label="服务器地址" prop="server">
            <el-input v-model="proxyForm.server" placeholder="请输入服务器地址" />
          </el-form-item>
          <el-form-item label="端口" prop="port">
            <el-input-number 
              v-model="proxyForm.port" 
              :min="1" 
              :max="65535" 
              placeholder="请输入端口" 
              style="width: 100%" />
          </el-form-item>
          <el-form-item label="加密方式" prop="method">
            <el-input v-model="proxyForm.method" placeholder="如: aes-256-gcm" />
          </el-form-item>
          <el-form-item label="密码" prop="password">
            <el-input v-model="proxyForm.password" type="password" placeholder="请输入密码" />
          </el-form-item>
          <el-form-item label="UUID" prop="uuid">
            <el-input v-model="proxyForm.uuid" placeholder="请输入UUID（VLESS/VMess）" />
          </el-form-item>
        </template>
        
        <!-- 启用状态 - 所有类型都显示 -->
        <el-form-item label="启用状态">
          <el-switch v-model="proxyForm.enabled" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="saveProxy" :loading="saving">
            保存
          </el-button>
        </span>
      </template>
    </el-dialog>


  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Plus, Connection, Refresh, Edit, Delete, CircleCheck, Document
} from '@element-plus/icons-vue'
import ServiceStatus from '../components/ServiceStatus.vue'
import { 
  apiGetSingBoxOutbounds, 
  apiCreateSingBoxOutbound, 
  apiUpdateSingBoxOutbound,
  apiDeleteSingBoxOutbound,
  apiRestartSingBoxService,
  apiValidateOutboundsChanges,
  apiBatchSaveOutbounds
} from '../utils/api'

// 响应式数据
const nodeGroups = ref([])
const activeGroup = ref('默认')
const dialogVisible = ref(false)
const saveConfirmDialogVisible = ref(false)
const editingProxy = ref(null)
const finalSaving = ref(false)
const validationMessage = ref('')
const saveOption = ref('save')
const pendingOperation = ref(null)
const saving = ref(false)
const savingAll = ref(false)
const pendingChanges = ref([])
const originalConfig = ref(null)
const formRef = ref()
const availableOutbounds = ref([])


// 表单数据
const proxyForm = reactive({
  name: '',
  type: '',
  server: '',
  port: null,
  method: '',
  password: '',
  uuid: '',
  default: 'direct',
  enabled: true,
  outbounds: [],
  include: '',
  exclude: '',
  useAllProviders: false
})

// 表单验证规则
const formRules = computed(() => {
  const baseRules = {
    name: [{ required: true, message: '请输入节点名称', trigger: 'blur' }],
    type: [{ required: true, message: '请选择类型', trigger: 'change' }]
  }
  
  // 只有代理节点需要验证服务器和端口
  if (activeGroup.value === '代理') {
    baseRules.server = [{ required: true, message: '请输入服务器地址', trigger: 'blur' }]
    baseRules.port = [{ required: true, message: '请输入端口', trigger: 'blur' }]
  }
  
  return baseRules
})

// 计算属性：是否有待保存的更改
const hasPendingChanges = computed(() => {
  return pendingChanges.value.length > 0
})

// 计算属性：待保存更改数量
const pendingChangesCount = computed(() => {
  return pendingChanges.value.length
})

// 计算属性

const currentGroupNodes = computed(() => {
  const group = nodeGroups.value.find(g => g.name === activeGroup.value)
  return group ? group.nodes : []
})

// 获取代理类型标签样式
const getProxyTypeTag = (type) => {
  const typeMap = {
    'shadowsocks': 'primary',
    'vmess': 'success',
    'vless': 'warning',
    'trojan': 'danger',
    'hysteria': 'info'
  }
  return typeMap[type] || 'default'
}

// 获取延迟样式类
const getLatencyClass = (latency) => {
  if (latency < 100) return 'latency-good'
  if (latency < 300) return 'latency-normal'
  return 'latency-poor'
}

// 获取协议图标
const getProtocolIcon = (protocolName) => {
  const iconMap = {
    '代理': '🚀',
    '应用分流': '📱',
    '节点过滤': '🌍',
    'DEFAULT': '📡'
  }
  return iconMap[protocolName] || iconMap['DEFAULT']
}

// 判断是否需要加载出站节点列表（应用分流和节点过滤都需要）
const isEditingAppNode = () => {
  return (
    activeGroup.value === '应用分流' || 
    activeGroup.value === '节点过滤'
  )
}

// 获取对话框标题
const getDialogTitle = () => {
  if (!editingProxy.value) {
    // 新增节点时根据当前分组确定标题
    if (activeGroup.value === '应用分流') {
      return '添加应用分流'
    } else if (activeGroup.value === '节点过滤') {
      return '添加节点过滤器'
    } else {
      return '添加代理节点'
    }
  }
  
  // 编辑现有节点时根据节点类型确定标题
  if (isEditingAppNode()) {
    return '编辑应用分流节点'
  }
  
  return '编辑代理节点'
}

// 加载可用的出站节点
const loadAvailableOutbounds = () => {
  const outboundGroups = []
  const allNodes = new Map() // 用于去重和快速查找
  
  // 收集所有节点到Map中
  nodeGroups.value.forEach(group => {
    if (group.nodes) {
      group.nodes.forEach(node => {
        allNodes.set(node.tag, node)
      })
    }
  })
  
  // 获取代理分组的节点
  const proxyGroup = nodeGroups.value.find(group => group.name === '代理')
  const proxyNodes = []
  if (proxyGroup && proxyGroup.nodes && proxyGroup.nodes.length > 0) {
    proxyNodes.push(...proxyGroup.nodes)
  }
  
  // 获取节点过滤分组的节点
  const filterGroup = nodeGroups.value.find(group => group.name === '节点过滤')
  const filterNodes = []
  if (filterGroup && filterGroup.nodes && filterGroup.nodes.length > 0) {
    filterNodes.push(...filterGroup.nodes)
  }
  
  // 检查当前已配置的出站节点，确保它们都在可选列表中
  const currentOutbounds = proxyForm.outbounds || []
  const configuredNodes = []
  
  currentOutbounds.forEach(outboundName => {
    const node = allNodes.get(outboundName)
    if (node) {
      // 检查这个节点是否已经在代理或过滤分组中
      const inProxy = proxyNodes.some(n => n.tag === node.tag)
      const inFilter = filterNodes.some(n => n.tag === node.tag)
      
      if (!inProxy && !inFilter) {
        configuredNodes.push(node)
      }
    } else {
      // 如果节点不存在于当前数据中，创建一个占位节点
      configuredNodes.push({
        tag: outboundName,
        type: 'unknown'
      })
    }
  })
  
  // 构建分组
  if (proxyNodes.length > 0) {
    outboundGroups.push({
      name: '代理',
      nodes: proxyNodes
    })
  }
  
  if (filterNodes.length > 0) {
    outboundGroups.push({
      name: '节点过滤',
      nodes: filterNodes
    })
  }
  
  // 不再创建"当前配置"分组，让已选择的节点在原始分组中显示为选中状态
  
  // 如果没有找到任何分组，尝试从所有分组中收集节点
  if (outboundGroups.length === 0) {
    const fallbackProxyNodes = []
    const fallbackFilterNodes = []
    
    nodeGroups.value.forEach(group => {
      if (group.nodes) {
        group.nodes.forEach(node => {
          // 真正的代理协议节点
          if (node.type && ['shadowsocks', 'vmess', 'vless', 'trojan', 'hysteria', 'hysteria2', 'tuic', 'wireguard', 'ssh', 'shadowtls', 'shadowsocksr'].includes(node.type.toLowerCase())) {
            fallbackProxyNodes.push(node)
          }
          // 逻辑节点（selector, urltest, loadbalance等）
          else if (node.type && ['selector', 'urltest', 'loadbalance'].includes(node.type.toLowerCase())) {
            fallbackFilterNodes.push(node)
          }
        })
      }
    })
    
    if (fallbackProxyNodes.length > 0) {
      outboundGroups.push({
        name: '代理节点',
        nodes: fallbackProxyNodes
      })
    }
    
    if (fallbackFilterNodes.length > 0) {
      outboundGroups.push({
        name: '节点过滤',
        nodes: fallbackFilterNodes
      })
    }
  }
  
  availableOutbounds.value = outboundGroups
  console.log('Available outbounds loaded:', availableOutbounds.value)
  console.log('Current form outbounds:', proxyForm.outbounds)
  console.log('Current configured outbounds:', currentOutbounds)
}

// 全选出站节点
const selectAllOutbounds = () => {
  const allOutbounds = []
  availableOutbounds.value.forEach(group => {
    group.nodes.forEach(node => {
      allOutbounds.push(node.name)
    })
  })
  proxyForm.outbounds = allOutbounds
}

// 清空出站节点选择
const clearAllOutbounds = () => {
  proxyForm.outbounds = []
}

// 移除单个出站节点
const removeOutbound = (outboundName) => {
  const index = proxyForm.outbounds.indexOf(outboundName)
  if (index > -1) {
    proxyForm.outbounds.splice(index, 1)
  }
}

// 获取已选择的出站节点选项（仅显示在出站节点中已勾选的节点）
const getAllOutboundOptions = () => {
  const selectedNodes = []
  
  // 只返回在proxyForm.outbounds中已选择的节点
  if (proxyForm.outbounds && proxyForm.outbounds.length > 0) {
    // 检查系统内置节点
    const systemNodes = ['direct', 'block']
    systemNodes.forEach(systemNode => {
      if (proxyForm.outbounds.includes(systemNode)) {
        selectedNodes.push({ tag: systemNode, type: systemNode })
      }
    })
    
    // 遍历所有分组找到对应的节点信息
    nodeGroups.value.forEach(group => {
      if (group.nodes) {
        group.nodes.forEach(node => {
          // 如果该节点在已选择列表中，则添加到选项中
          if (proxyForm.outbounds.includes(node.tag)) {
            selectedNodes.push({ tag: node.tag, type: node.type })
          }
        })
      }
    })
  }
  
  return selectedNodes
}

// 获取出站节点的显示名称
const getOutboundDisplayName = (tag) => {
  switch (tag) {
    case 'direct':
      return '直连'
    case 'block':
      return '拒绝'
    default:
      return tag
  }
}

// 确保必要的标签页始终存在并按正确顺序排列
const ensureRequiredGroups = (groups) => {
  const requiredGroups = ['代理', '应用分流', '节点过滤']
  const result = []
  
  // 按照预定义的顺序添加标签页
  requiredGroups.forEach(groupName => {
    const existingGroup = groups.find(group => group.name === groupName)
    if (existingGroup) {
      // 如果存在，使用现有的数据
      result.push(existingGroup)
    } else {
      // 如果不存在，创建空的分组
      result.push({
        name: groupName,
        nodes: []
      })
    }
  })
  
  // 添加其他不在预定义列表中的分组（如果有的话）
  groups.forEach(group => {
    if (!requiredGroups.includes(group.name)) {
      result.push(group)
    }
  })
  
  return result
}

// 刷新节点列表
const refreshNodes = async (keepCurrentGroup = true) => {
  try {
    const currentActiveGroup = activeGroup.value
    const response = await apiGetSingBoxOutbounds()
    if (response.code === 200) {
      // 确保必要的标签页始终存在
      nodeGroups.value = ensureRequiredGroups(response.data || [])
      
      // 如果需要保持当前分组且当前分组存在，则保持不变
      if (keepCurrentGroup && currentActiveGroup !== '默认') {
        const groupExists = nodeGroups.value.some(group => group.name === currentActiveGroup)
        if (groupExists) {
          activeGroup.value = currentActiveGroup
        } else if (nodeGroups.value.length > 0) {
          // 当前分组不存在了，切换到第一个分组
          activeGroup.value = nodeGroups.value[0].name
        }
      } else {
        // 设置默认激活的分组（首次加载或明确要求重置）
        if (nodeGroups.value.length > 0) {
          activeGroup.value = nodeGroups.value[0].name
        }
      }
    } else {
      throw new Error(response.message || '获取节点列表失败')
    }
  } catch (error) {
    console.error('获取节点列表失败:', error)
    ElMessage.error('获取节点列表失败: ' + (error.response?.data?.message || error.message))
    // 失败时也要确保必要的标签页存在
    nodeGroups.value = ensureRequiredGroups([{
      name: '默认',
      nodes: []
    }])
    activeGroup.value = nodeGroups.value[0].name
  }
}

// 分组切换
const handleGroupChange = (groupName) => {
  activeGroup.value = groupName
}

// 添加代理节点
const addProxy = () => {
  editingProxy.value = null
  Object.assign(proxyForm, {
    name: '',
    type: '',
    server: '',
    port: null,
    method: '',
    password: '',
    uuid: '',
    default: 'direct',
    enabled: true,
    outbounds: [],
    include: '',
    exclude: '',
    useAllProviders: false
  })
  
  dialogVisible.value = true
}

// 添加应用分流节点
const addAppProxy = () => {
  editingProxy.value = null
  Object.assign(proxyForm, {
    name: '',
    type: 'selector', // 应用分流通常使用selector类型
    server: '',
    port: null,
    method: '',
    password: '',
    uuid: '',
    default: 'direct',
    enabled: true,
    outbounds: [],
    include: '',
    exclude: '',
    useAllProviders: false
  })
  
  // 加载可用的出站节点
  loadAvailableOutbounds()
  
  dialogVisible.value = true
}

// 添加节点过滤器
const addFilterProxy = () => {
  editingProxy.value = null
  Object.assign(proxyForm, {
    name: '',
    type: 'selector', // 过滤器通常使用selector类型
    server: '',
    port: null,
    method: '',
    password: '',
    uuid: '',
    default: 'direct',
    enabled: true,
    outbounds: [],
    include: '',
    exclude: '',
    useAllProviders: false
  })
  
  // 加载可用的出站节点
  loadAvailableOutbounds()
  
  dialogVisible.value = true
}

// 编辑代理节点
const editProxy = (proxy) => {
  editingProxy.value = proxy
  
  // 直接使用配置文件中的字段，不做额外处理
  const formData = {
    name: proxy.tag || '',
    type: proxy.type || '',
    server: proxy.server || '',
    port: proxy.server_port || null,
    method: proxy.method || '',
    password: proxy.password || '',
    uuid: proxy.uuid || '',
    default: proxy.default || 'direct',
    enabled: true, // UI显示用
    outbounds: proxy.outbounds || [],
    include: proxy.include || '',
    exclude: proxy.exclude || '',
    useAllProviders: proxy.use_all_providers || false
  }
  
  Object.assign(proxyForm, formData)
  
  // 如果是应用分流或节点过滤节点，加载可用的出站节点
  if (isEditingAppNode()) {
    loadAvailableOutbounds()
  }
  
  dialogVisible.value = true
}

// 保存代理节点 (直接保存模式)
const saveProxy = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    saving.value = true
    
    // 准备节点数据
    const outboundData = {
      tag: proxyForm.name,
      type: proxyForm.type
    }
    
    // 根据节点类型添加相应的字段
    if (activeGroup.value === '节点过滤') {
      // 节点过滤器配置
      if (proxyForm.include) outboundData.include = proxyForm.include
      if (proxyForm.exclude) outboundData.exclude = proxyForm.exclude
      if (proxyForm.useAllProviders) outboundData.use_all_providers = proxyForm.useAllProviders
      if (proxyForm.outbounds && proxyForm.outbounds.length > 0) {
        outboundData.outbounds = proxyForm.outbounds
      }
    } else if (activeGroup.value === '应用分流') {
      // 应用分流配置
      if (proxyForm.outbounds && proxyForm.outbounds.length > 0) {
        outboundData.outbounds = proxyForm.outbounds
      }
      if (proxyForm.default) outboundData.default = proxyForm.default
    } else {
      // 代理节点配置
      if (proxyForm.server) outboundData.server = proxyForm.server
      if (proxyForm.port) outboundData.server_port = proxyForm.port
      if (proxyForm.method) outboundData.method = proxyForm.method
      if (proxyForm.password) outboundData.password = proxyForm.password
      if (proxyForm.uuid) outboundData.uuid = proxyForm.uuid
    }
    
    // 调用API直接保存节点
    let saveSuccess = false
    if (editingProxy.value) {
      const { apiUpdateSingBoxOutbound } = await import('../utils/api')
      await apiUpdateSingBoxOutbound(editingProxy.value.id, outboundData)
      ElMessage.success('节点更新成功')
      saveSuccess = true
    } else {
      const { apiCreateSingBoxOutbound } = await import('../utils/api')
      await apiCreateSingBoxOutbound(outboundData)
      ElMessage.success('节点添加成功')
      saveSuccess = true
    }
    
    // 保存成功后清理状态并关闭对话框
    editingProxy.value = null
    dialogVisible.value = false
    await refreshNodes() // 刷新节点列表
    
    // 只有在保存成功后才询问是否重启服务
    if (saveSuccess) {
      try {
        await ElMessageBox.confirm(
          '✅ 配置验证成功并已保存！\n\n是否重启 Sing-Box 服务以应用更改？',
          '重启服务',
          {
            confirmButtonText: '重启服务',
            cancelButtonText: '稍后手动重启',
            type: 'info',
            closeOnClickModal: false,
            closeOnPressEscape: false,
            showClose: true
          }
        )
        
        // 用户选择重启服务
        try {
          const { apiRestartSingBoxService } = await import('../utils/api')
          await apiRestartSingBoxService()
          ElMessage.success('Sing-Box 服务重启成功')
        } catch (restartError) {
          console.error('重启服务失败:', restartError)
          ElMessage.error('重启服务失败: ' + (restartError.response?.data?.message || restartError.message))
        }
      } catch (confirmError) {
        // 用户取消重启，不做任何操作
        console.log('用户选择稍后手动重启服务')
      }
    }
    
  } catch (error) {
    console.error('保存节点失败:', error)
    ElMessage.error('保存失败: ' + (error.response?.data?.message || error.message))
  } finally {
    saving.value = false
  }
}

// 批量保存所有更改 (先验证，确认后再保存)
const saveAllChanges = async () => {
  if (pendingChanges.value.length === 0) return
  
  try {
    savingAll.value = true
    ElMessage.info('正在验证配置...')
    
    // 只验证，不保存
    const response = await apiValidateOutboundsChanges(pendingChanges.value)
    
    // 显示验证成功，准备保存确认
    ElMessage.success('✅ 配置验证通过!')
    validationMessage.value = '所有配置更改已通过 Sing-Box 官方验证，可以选择只保存或保存并重启 Sing-Box 服务'
    saveOption.value = 'save'
    saveConfirmDialogVisible.value = true
    
    // 不在这里清空 pendingChanges，等用户确认后再清空
    
  } catch (error) {
    console.error('配置验证失败:', error)
    ElMessage.error('配置验证失败: ' + (error.response?.data?.error || error.message))
  } finally {
    savingAll.value = false
  }
}

// 刷新节点列表并应用待保存的更改
const refreshNodesWithPendingChanges = async () => {
  // 先刷新原始数据，保持当前分组
  await refreshNodes(true)
  
  // 然后应用待保存的更改到显示数据
  applyPendingChangesToDisplay()
}

// 将待保存的更改应用到显示数据
const applyPendingChangesToDisplay = () => {
  const processedChanges = new Set()
  
  pendingChanges.value.forEach(change => {
    const changeKey = `${change.type}-${change.data.tag}`
    if (processedChanges.has(changeKey)) return
    processedChanges.add(changeKey)
    
    if (change.type === 'create') {
      // 添加新节点到相应分组
      const group = findOrCreateGroup(change.data.type, change.data.tag, change.data)
      if (group) {
        const newNode = {
          ...change.data,
          id: change.id,
          isPending: true,
          pendingType: 'create'
        }
        group.nodes.push(newNode)
      }
    } else if (change.type === 'update') {
      // 更新现有节点
      nodeGroups.value.forEach(group => {
        const nodeIndex = group.nodes.findIndex(node => 
          node.tag === change.originalProxy.tag
        )
        if (nodeIndex !== -1) {
          group.nodes[nodeIndex] = {
            ...group.nodes[nodeIndex],
            ...change.data,
            isPending: true,
            pendingType: 'update'
          }
        }
      })
    }
  })
}

// 找到或创建节点分组
const findOrCreateGroup = (nodeType, nodeTag, nodeData = null) => {
  // 判断是否为带正则表达式的节点过滤器
  const hasRegexFilter = nodeData && (
    (nodeData.include && nodeData.include.trim() !== '') || 
    (nodeData.exclude && nodeData.exclude.trim() !== '')
  )
  
  let groupName
  if (hasRegexFilter) {
    // 只有带正则表达式的节点才归类为节点过滤
    groupName = '节点过滤'
  } else if (activeGroup.value === '应用分流' || activeGroup.value === '节点过滤') {
    // 在应用分流或节点过滤页面添加但没有正则表达式的节点都归类为应用分流
    groupName = '应用分流'
  } else {
    // 对于代理节点，使用原来的分类逻辑
    groupName = categorizeNode(nodeType, nodeTag)
  }
  
  let group = nodeGroups.value.find(g => g.name === groupName)
  if (!group) {
    group = { name: groupName, nodes: [] }
    nodeGroups.value.push(group)
  }
  return group
}

// 节点分类逻辑 (简化版)
const categorizeNode = (type, tag) => {
  const proxyProtocols = ['shadowsocks', 'vmess', 'vless', 'trojan', 'hysteria', 'hysteria2']
  if (proxyProtocols.includes(type?.toLowerCase())) {
    return '代理'
  }
  
  // 应用分流检测
  const appPattern = /(telegram|twitter|youtube|google|apple|netflix|spotify|github|discord|tiktok|instagram|facebook|whatsapp|微信|QQ|百度|淘宝|支付宝|抖音|微博|知乎|bilibili|steam|选择|规则|漏网之鱼)/i
  if (appPattern.test(tag)) {
    return '应用分流'
  }
  
  return '节点过滤'
}

// 获取待保存状态标签类型
const getPendingTagType = (pendingType) => {
  switch (pendingType) {
    case 'create': return 'primary'
    case 'update': return 'warning'
    case 'delete': return 'danger'
    default: return 'info'
  }
}

// 获取待保存状态标签文本
const getPendingTagText = (pendingType) => {
  switch (pendingType) {
    case 'create': return '新增'
    case 'update': return '修改'
    case 'delete': return '删除'
    default: return '待保存'
  }
}

// 确认保存操作
const confirmSave = async () => {
  try {
    finalSaving.value = true
    
    // 批量保存所有更改
    await apiBatchSaveOutbounds(pendingChanges.value)
    
    // 如果选择保存并重启，执行重启操作
    if (saveOption.value === 'save-restart') {
      await apiRestartSingBoxService()
      ElMessage.success('配置已保存并成功重启 Sing-Box 服务')
    } else {
      ElMessage.success('配置已保存')
    }
    
    // 只有在用户确认保存后才清空待保存更改
    pendingChanges.value = []
    
    saveConfirmDialogVisible.value = false
    pendingOperation.value = null
    
    // 刷新节点列表但保持在当前分组页面
    await refreshNodes(true)
  } catch (error) {
    console.error('操作失败:', error)
    ElMessage.error('操作失败: ' + (error.response?.data?.error || error.message))
  } finally {
    finalSaving.value = false
  }
}

// 处理保存确认对话框关闭
const handleSaveDialogClose = () => {
  // 当用户点击X关闭对话框时，不清空 pendingChanges
  // 保持保存按钮可见，让用户可以重新尝试保存
  saveConfirmDialogVisible.value = false
  pendingOperation.value = null
  
  // 显示提示，告知用户配置尚未最终保存
  ElMessage.warning('配置尚未最终保存，请重新点击"保存配置"按钮')
}

// 删除代理节点 (直接删除模式)
const deleteProxy = (proxy) => {
  ElMessageBox.confirm(
    `确定要删除节点 "${proxy.tag || proxy.name}" 吗？`,
    '确认删除',
    {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      const { apiDeleteSingBoxOutbound } = await import('../utils/api')
      await apiDeleteSingBoxOutbound(proxy.id)
      ElMessage.success('节点删除成功')
      await refreshNodes() // 刷新节点列表
      
      // 询问是否重启服务
      try {
        await ElMessageBox.confirm(
          '节点已删除。是否重启 Sing-Box 服务以应用更改？',
          '重启服务',
          {
            confirmButtonText: '重启服务',
            cancelButtonText: '稍后手动重启',
            type: 'info',
            closeOnClickModal: false,
            closeOnPressEscape: false,
            showClose: true
          }
        )
        
        // 用户选择重启服务
        try {
          const { apiRestartSingBoxService } = await import('../utils/api')
          await apiRestartSingBoxService()
          ElMessage.success('Sing-Box 服务重启成功')
        } catch (restartError) {
          console.error('重启服务失败:', restartError)
          ElMessage.error('重启服务失败: ' + (restartError.response?.data?.message || restartError.message))
        }
      } catch (confirmError) {
        // 用户取消重启，不做任何操作
        console.log('用户选择稍后手动重启服务')
      }
    } catch (error) {
      console.error('删除节点失败:', error)
      ElMessage.error('删除失败: ' + (error.response?.data?.message || error.message))
    }
  }).catch(() => {
    // 用户取消删除
  })
}

// 测试单个节点
const testNode = async (node) => {
  node.testing = true
  try {
    // TODO: 调用API测试节点
    await new Promise(resolve => setTimeout(resolve, 2000)) // 模拟测试
    node.latency = Math.floor(Math.random() * 300) + 20
    node.status = 'online'
    ElMessage.success(`节点 ${node.name} 测试完成`)
  } catch (error) {
    node.status = 'offline'
    ElMessage.error(`节点 ${node.name} 测试失败`)
  } finally {
    node.testing = false
  }
}

// 测试所有节点
const testAllNodes = async () => {
  ElMessage.info('开始测试所有节点...')
  for (const group of nodeGroups.value) {
    for (const node of group.nodes) {
      await testNode(node)
    }
  }
  ElMessage.success('所有节点测试完成')
}


// 组件挂载时获取数据
onMounted(() => {
  // 首次加载时不保持当前分组
  refreshNodes(false)
})
</script>

<style scoped>
.sing-box-proxy {
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
  font-size: 20px;
}

.header-controls {
  display: flex;
  gap: 10px;
}


.proxy-list-card {
  margin-bottom: 15px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.group-selector {
  margin-bottom: 20px;
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 6px;
}

.protocol-icon {
  font-size: 16px;
}

.protocol-name {
  font-weight: 500;
}

.node-count {
  font-size: 12px;
  color: #909399;
  font-weight: normal;
}

.node-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-online {
  background-color: #67c23a;
}

.status-offline {
  background-color: #f56c6c;
}

.latency-good {
  color: #67c23a;
  font-weight: bold;
}

.latency-normal {
  color: #e6a23c;
}

.latency-poor {
  color: #f56c6c;
}

/* 出站节点选择样式 */
.outbound-selection {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 12px;
  max-height: 400px;
  overflow-y: auto;
}

.selection-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #ebeef5;
}

.selection-header span {
  font-weight: 500;
  color: #606266;
}

.outbound-group {
  margin-bottom: 16px;
}

.outbound-group:last-child {
  margin-bottom: 0;
}

.group-title {
  font-weight: 500;
  color: #409eff;
  margin-bottom: 8px;
  font-size: 14px;
}

.group-nodes {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.outbound-checkbox {
  margin: 0;
  width: 100%;
}

.outbound-checkbox .el-checkbox__label {
  width: 100%;
}

.node-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.node-info:hover {
  background-color: #f5f7fa;
}

.node-name {
  flex: 1;
  font-size: 13px;
}

.no-outbounds {
  text-align: center;
  padding: 20px;
}

.no-nodes {
  padding: 8px 12px;
  text-align: center;
}

/* 当前选择预览样式 */
.current-selection {
  margin-bottom: 12px;
  padding: 12px;
  background-color: #f8f9fa;
  border-radius: 4px;
  border: 1px solid #e9ecef;
}

.selection-title {
  font-size: 13px;
  color: #606266;
  margin-bottom: 8px;
  font-weight: 500;
}

.selected-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.selected-tag {
  margin: 0;
}

.text-muted {
  color: #909399;
}

.empty-state {
  text-align: center;
  padding: 40px 0;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

/* 保存确认对话框样式 */
.save-confirm-content {
  padding: 10px 0;
}

.validation-status {
  text-align: center;
  padding: 20px 0;
}

.success-icon {
  color: #67c23a;
  margin-bottom: 12px;
}

.status-title {
  font-size: 18px;
  color: #67c23a;
  margin: 0 0 8px 0;
  font-weight: 600;
}

.status-description {
  font-size: 14px;
  color: #606266;
  margin: 0;
  line-height: 1.5;
}

.save-options {
  text-align: left;
}

.save-options h4 {
  font-size: 16px;
  color: #303133;
  margin: 0 0 16px 0;
  font-weight: 600;
}

.save-radio-group {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.save-radio-group .el-radio {
  margin-right: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  padding: 12px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  transition: all 0.2s;
}

.save-radio-group .el-radio:hover {
  border-color: #409eff;
  background-color: #f0f9ff;
}

.save-radio-group .el-radio.is-checked {
  border-color: #409eff;
  background-color: #f0f9ff;
}

.radio-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
  display: block;
}

.radio-desc {
  font-size: 12px;
  color: #909399;
  line-height: 1.4;
  margin-left: 20px;
}
</style>

