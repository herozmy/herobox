<template>
  <div class="sing-box-rules">
    <div class="page-header">
      <h2>规则设置</h2>
    </div>

    <!-- 保存/撤销操作栏 -->
    <div v-if="hasUnsavedChanges" class="save-bar">
      <div class="save-bar-content">
        <div class="save-bar-info">
          <el-icon class="warning-icon"><InfoFilled /></el-icon>
          <span>您有未保存的规则排序更改</span>
        </div>
        <div class="save-bar-actions">
          <el-button type="default" @click="cancelChanges" :loading="isReverting">
            撤销更改
          </el-button>
          <el-button type="primary" @click="saveChanges" :loading="isSaving">
            保存更改
          </el-button>
        </div>
      </div>
    </div>

    <!-- 重启服务提醒栏 -->
    <div v-if="needsRestart" class="restart-bar">
      <div class="restart-bar-content">
        <div class="restart-bar-info">
          <el-icon class="restart-icon"><RefreshRight /></el-icon>
          <span>配置已保存，需要重启 Sing-Box 服务以应用更改</span>
        </div>
        <div class="restart-bar-actions">
          <el-button type="success" @click="restartService" :loading="isRestarting">
            重启服务
          </el-button>
          <el-button type="default" @click="dismissRestart">
            稍后重启
          </el-button>
        </div>
      </div>
    </div>

    <!-- 服务状态 -->
    <ServiceStatus service-name="sing-box" />

    <!-- 规则管理 - 左右布局 -->
    <div class="rules-content-flex">
      <!-- 左侧：路由规则表格 -->
      <div class="rules-left">
        <el-card class="rules-list-card">
          <template #header>
            <div class="card-header">
              <span>路由规则 (route.rules)</span>
              <el-button type="primary" size="small" @click="addRouteRule">
                <el-icon><Plus /></el-icon>
                添加规则
              </el-button>
            </div>
          </template>

          <el-table 
            :data="routeRules" 
            stripe 
            table-layout="auto" 
            size="small"
            ref="routeRulesTable"
            :row-key="(row) => row.id"
            @row-click="handleRowClick">
            <el-table-column type="index" label="序号" width="50" />
            <el-table-column label="匹配条件" min-width="200" show-overflow-tooltip>
              <template #default="scope">
                <div class="rule-conditions">
                  <div v-for="condition in scope.row.conditions" :key="condition.type" class="condition-item">
                    <el-tag size="small" :type="getConditionTypeTag(condition.type)">
                      {{ getConditionTypeLabel(condition.type) }}
                    </el-tag>
                    <code class="condition-content" :title="condition.content">
                      {{ condition.content.length > 40 ? condition.content.substring(0, 40) + '...' : condition.content }}
                    </code>
                  </div>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="其他条件" width="120" show-overflow-tooltip>
              <template #default="scope">
                <div class="other-conditions">
                  <el-tag v-if="scope.row.invert" size="small" type="warning">反向</el-tag>
                  <el-tag v-if="scope.row.ip_version" size="small" type="info">IPv{{ scope.row.ip_version }}</el-tag>
                  <el-tag v-if="scope.row.ip_is_private !== undefined" size="small" type="warning">
                    {{ scope.row.ip_is_private ? '私有IP' : '公有IP' }}
                  </el-tag>
                  <el-tag v-if="scope.row.network" size="small" type="default">{{ scope.row.network }}</el-tag>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="出站/动作" width="100">
              <template #default="scope">
                <el-tag :type="getOutboundOrActionTag(scope.row)" size="small">
                  {{ getOutboundOrActionDisplay(scope.row) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="220">
              <template #default="scope">
                <div class="table-actions">
                  <el-button 
                    size="small" 
                    :icon="ArrowUp"
                    @click="moveRouteRuleUp(scope.row, scope.$index)"
                    :disabled="scope.$index === 0 || hasUnsavedChanges"
                    title="上移"
                  />
                  <el-button 
                    size="small" 
                    :icon="ArrowDown"
                    @click="moveRouteRuleDown(scope.row, scope.$index)"
                    :disabled="scope.$index === routeRules.length - 1 || hasUnsavedChanges"
                    title="下移"
                  />
                  <el-button 
                    type="primary" 
                    size="small" 
                    @click="editRouteRule(scope.row)"
                    :icon="Edit">
                    编辑
                  </el-button>
                  <el-button 
                    type="danger" 
                    size="small" 
                    @click="deleteRouteRule(scope.row, scope.$index)"
                    :icon="Delete">
                    删除
                  </el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>

          <div v-if="!routeRules.length" class="empty-state">
            <el-empty description="暂无路由规则配置">
              <el-button type="primary" @click="addRouteRule">添加第一个路由规则</el-button>
            </el-empty>
          </div>
        </el-card>
      </div>

      <!-- 右侧：规则集 -->
      <div class="rules-right">
        <el-card class="rules-list-card">
          <template #header>
            <div class="card-header">
              <span>规则集 (route.rule_set)</span>
              <el-button type="primary" size="small" @click="addRuleSet">
                <el-icon><Plus /></el-icon>
                添加规则集
              </el-button>
            </div>
          </template>

          <div class="rule-sets-list-compact">
            <div 
              v-for="ruleSet in ruleSets" 
              :key="ruleSet.id"
              class="rule-set-card-compact">
              <div class="rule-set-header-compact">
                <div class="rule-set-info-compact">
                  <h4 class="rule-set-name-compact">{{ ruleSet.tag }}</h4>
                  <div class="rule-set-properties-compact">
                    <el-tag v-if="ruleSet.type" :type="getRuleSetTypeTag(ruleSet.type)" size="small">
                      {{ ruleSet.type }}
                    </el-tag>
                    <el-tag v-if="ruleSet.format" size="small">{{ ruleSet.format }}</el-tag>
                  </div>
                </div>
                <div class="rule-set-actions-compact">
                  <el-button type="primary" size="small" @click="editRuleSet(ruleSet)">
                    编辑
                  </el-button>
                  <el-button type="danger" size="small" @click="deleteRuleSet(ruleSet)">
                    删除
                  </el-button>
                </div>
              </div>
              <div v-if="ruleSet.url || ruleSet.path" class="rule-set-source-compact">
                <code class="source-url-compact">{{ (ruleSet.url || ruleSet.path).length > 40 ? (ruleSet.url || ruleSet.path).substring(0, 40) + '...' : (ruleSet.url || ruleSet.path) }}</code>
              </div>
            </div>
          </div>

          <div v-if="!ruleSets.length" class="empty-state">
            <el-empty description="暂无规则集配置">
              <el-button type="primary" @click="addRuleSet">添加第一个规则集</el-button>
            </el-empty>
          </div>
        </el-card>
      </div>
    </div>

    <!-- 路由规则编辑对话框 -->
    <el-dialog 
      v-model="ruleDialogVisible" 
      :title="editingRule ? '编辑路由规则' : '添加路由规则'"
      width="600px"
      :close-on-click-modal="false"
      @closed="onRuleDialogClosed">
      <el-form :model="ruleForm" :rules="ruleFormRules" ref="ruleFormRef" label-width="80px">
        <!-- 出站选择 -->
        <el-form-item label="出站" prop="outbound">
          <el-select v-model="ruleForm.outbound" placeholder="请选择出站" style="width: 100%">
            <el-option 
              v-for="outbound in availableOutbounds" 
              :key="outbound" 
              :label="outbound" 
              :value="outbound" />
          </el-select>
        </el-form-item>

        <!-- 规则集选择 -->
        <el-form-item label="规则集">
          <el-select
            v-model="selectedRuleSets"
            multiple
            placeholder="请选择规则集"
            style="width: 100%"
            collapse-tags
            collapse-tags-tooltip
            :max-collapse-tags="3"
            clearable
            filterable>
            <el-option
              v-for="ruleSet in availableRuleSets"
              :key="ruleSet"
              :label="ruleSet"
              :value="ruleSet">
              <span style="float: left">{{ ruleSet }}</span>
              <span style="float: right; color: #8492a6; font-size: 12px">规则集</span>
            </el-option>
          </el-select>
        </el-form-item>

        <!-- 传统规则 (可选) -->
        <el-collapse v-model="activeCollapse" class="rule-collapse">
          <el-collapse-item title="▶ ▶ 传统规则 (可选)" name="traditional">
            <div class="traditional-rules">
              <!-- 域名规则 -->
              <el-form-item label="域名">
                <el-input 
                  v-model="ruleForm.domain" 
                  type="textarea" 
                  :rows="2"
                  placeholder="完整域名匹配，每行一个&#10;example.com" />
              </el-form-item>
              
              <el-form-item label="域名后缀">
                <el-input 
                  v-model="ruleForm.domain_suffix" 
                  type="textarea" 
                  :rows="2"
                  placeholder="域名后缀匹配，每行一个&#10;.google.com" />
              </el-form-item>
              
              <el-form-item label="域名关键词">
                <el-input 
                  v-model="ruleForm.domain_keyword" 
                  type="textarea" 
                  :rows="2"
                  placeholder="域名关键词匹配，每行一个&#10;google" />
              </el-form-item>

              <el-form-item label="GeoSite">
                <el-input 
                  v-model="ruleForm.geosite" 
                  placeholder="GeoSite规则，如: google,youtube" />
              </el-form-item>

              <!-- IP规则 -->
              <el-form-item label="IP地址段">
                <el-input 
                  v-model="ruleForm.ip_cidr" 
                  type="textarea" 
                  :rows="2"
                  placeholder="IP CIDR，每行一个&#10;192.168.1.0/24" />
              </el-form-item>

              <el-form-item label="GeoIP">
                <el-input 
                  v-model="ruleForm.geoip" 
                  placeholder="GeoIP规则，如: cn,us" />
              </el-form-item>

              <!-- 规则集 -->
              <el-form-item label="规则集">
                <el-input 
                  v-model="ruleForm.rule_set" 
                  placeholder="规则集，如: geosite-google,geosite-youtube" />
              </el-form-item>

              <!-- 其他规则 -->
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="端口">
                    <el-input 
                      v-model="ruleForm.port" 
                      placeholder="80,443,8080-8090" />
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="协议">
                    <el-input 
                      v-model="ruleForm.protocol" 
                      placeholder="http,tls" />
                  </el-form-item>
                </el-col>
              </el-row>

              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="入站">
                    <el-input 
                      v-model="ruleForm.inbound" 
                      placeholder="入站标签" />
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="进程名">
                    <el-input 
                      v-model="ruleForm.process_name" 
                      placeholder="chrome.exe" />
                  </el-form-item>
                </el-col>
              </el-row>

              <!-- 高级选项 -->
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="IP版本">
                    <el-select v-model="ruleForm.ip_version" placeholder="不限制" clearable style="width: 100%">
                      <el-option label="IPv4" :value="4" />
                      <el-option label="IPv6" :value="6" />
                    </el-select>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="私有IP">
                    <el-select v-model="ruleForm.ip_is_private" placeholder="不限制" clearable style="width: 100%">
                      <el-option label="匹配私有IP" :value="true" />
                      <el-option label="匹配公有IP" :value="false" />
                    </el-select>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="反向匹配">
                    <el-switch v-model="ruleForm.invert" />
                  </el-form-item>
                </el-col>
              </el-row>
            </div>
          </el-collapse-item>
        </el-collapse>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="cancelRuleEdit">取消</el-button>
          <el-button type="primary" @click="saveRouteRule" :loading="saving">
            保存
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 规则集编辑对话框 -->
    <el-dialog
      v-model="showRuleSetDialog"
      :title="editingRuleSet ? '编辑规则集' : '添加规则集'"
      width="500px"
      :close-on-click-modal="false">
      
      <el-form
        ref="ruleSetFormRef"
        :model="ruleSetForm"
        label-width="120px"
        :rules="ruleSetFormRules">
        
        <el-form-item label="标签 (Tag)" prop="tag">
          <el-input
            v-model="ruleSetForm.tag"
            placeholder="请输入规则集标签，如: geosite-google"
            clearable />
        </el-form-item>

        <el-form-item label="类型" prop="type">
          <el-select
            v-model="ruleSetForm.type"
            placeholder="请选择规则集类型"
            style="width: 100%">
            <el-option label="远程" value="remote" />
            <el-option label="本地" value="local" />
          </el-select>
        </el-form-item>

        <el-form-item label="格式" prop="format">
          <el-select
            v-model="ruleSetForm.format"
            placeholder="请选择规则集格式"
            style="width: 100%">
            <el-option label="Binary" value="binary" />
            <el-option label="Source" value="source" />
          </el-select>
        </el-form-item>

        <el-form-item
          v-if="ruleSetForm.type === 'remote'"
          label="URL"
          prop="url">
          <el-input
            v-model="ruleSetForm.url"
            placeholder="请输入规则集URL"
            clearable />
        </el-form-item>

        <el-form-item
          v-if="ruleSetForm.type === 'local'"
          label="路径"
          prop="path">
          <el-input
            v-model="ruleSetForm.path"
            placeholder="请输入本地文件路径"
            clearable />
        </el-form-item>

        <el-form-item
          v-if="ruleSetForm.type === 'remote'"
          label="下载代理">
          <el-select
            v-model="ruleSetForm.download_detour"
            placeholder="请选择下载代理（可选）"
            clearable
            style="width: 100%">
            <el-option
              v-for="outbound in availableOutbounds"
              :key="outbound"
              :label="outbound"
              :value="outbound" />
          </el-select>
        </el-form-item>

        <el-form-item
          v-if="ruleSetForm.type === 'remote'"
          label="更新间隔">
          <el-input
            v-model="ruleSetForm.update_interval"
            placeholder="如: 1d, 12h, 30m (可选)"
            clearable />
        </el-form-item>
      </el-form>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="handleRuleSetCancel">取消</el-button>
          <el-button type="primary" @click="handleRuleSetSave" :loading="ruleSetLoading">
            {{ editingRuleSet ? '更新' : '添加' }}
          </el-button>
        </span>
      </template>
    </el-dialog>

  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus, Edit, Delete, Document, ArrowUp, ArrowDown, InfoFilled, RefreshRight } from '@element-plus/icons-vue'
import ServiceStatus from '../components/ServiceStatus.vue'
import { apiGetSingBoxRules, apiGetSingBoxOutbounds, apiCreateRouteRule, apiUpdateRouteRule, apiDeleteRouteRule, apiMoveRouteRuleUp, apiMoveRouteRuleDown, apiReorderRouteRules, apiValidateCurrentSingBoxConfig } from '../utils/api'
import Sortable from 'sortablejs'

// 响应式数据
const activeTab = ref('route-rules')
const routeRules = ref([])
const ruleSets = ref([])
const availableOutbounds = ref([])
const availableRuleSets = ref([])

// 规则集管理相关
const showRuleSetDialog = ref(false)
const editingRuleSet = ref(null)
const ruleSetFormRef = ref()
const ruleSetLoading = ref(false)
const ruleSetForm = reactive({
  tag: '',
  type: 'remote',
  format: 'binary',
  url: '',
  path: '',
  download_detour: '',
  update_interval: ''
})

// 规则集表单验证规则
const ruleSetFormRules = reactive({
  tag: [
    { required: true, message: '请输入规则集标签', trigger: 'blur' },
    { min: 1, max: 50, message: '标签长度应为 1-50 个字符', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择规则集类型', trigger: 'change' }
  ],
  url: [
    { 
      validator: (rule, value, callback) => {
        if (ruleSetForm.type === 'remote' && !value) {
          callback(new Error('远程规则集必须提供URL'))
        } else {
          callback()
        }
      }, 
      trigger: 'blur' 
    }
  ],
  path: [
    { 
      validator: (rule, value, callback) => {
        if (ruleSetForm.type === 'local' && !value) {
          callback(new Error('本地规则集必须提供文件路径'))
        } else {
          callback()
        }
      }, 
      trigger: 'blur' 
    }
  ]
})

// 拖拽相关
const routeRulesTable = ref()
let sortableInstance = null
const isDragging = ref(false)

// 保存状态管理
const hasUnsavedChanges = ref(false)
const isSaving = ref(false)
const isReverting = ref(false)
const originalRulesOrder = ref([]) // 保存原始规则顺序
const currentRulesOrder = ref([])   // 当前显示的规则顺序

// 重启服务状态管理
const needsRestart = ref(false)
const isRestarting = ref(false)

// 路由规则编辑相关
const ruleDialogVisible = ref(false)
const editingRule = ref(null)
const editingRuleIndex = ref(-1)
const originalRuleBackup = ref(null) // 用于保存编辑前的原始状态
const saving = ref(false)
const ruleFormRef = ref()
const activeCollapse = ref([])
const selectedRuleSets = ref([])
const movingRuleIds = ref(new Set())

// 路由规则表单数据
const ruleForm = reactive({
  outbound: '',
  ip_version: null,
  invert: false,
  domain: '',
  domain_suffix: '',
  domain_keyword: '',
  domain_regex: '',
  geosite: '',
  ip_cidr: '',
  source_ip_cidr: '',
  geoip: '',
  source_geoip: '',
  network: '',
  protocol: '',
  inbound: '',
  port: '',
  port_range: '',
  source_port: '',
  source_port_range: '',
  process_name: '',
  process_path: '',
  auth_user: '',
  rule_set: '',
  ip_is_private: null
})

// 表单验证规则
const ruleFormRules = {
  outbound: [{ required: true, message: '请选择出站', trigger: 'change' }]
}

// 工具函数

// 获取匹配条件类型标签样式
const getConditionTypeTag = (type) => {
  const typeMap = {
    'domain': 'primary',
    'domain_suffix': 'success',
    'domain_keyword': 'warning',
    'domain_regex': 'primary',
    'ip_cidr': 'danger',
    'geoip': 'info',
    'geosite': 'primary',
    'source_geoip': 'info',
    'inbound': 'default',
    'protocol': 'default',
    'network': 'default',
    'auth_user': 'default',
    'port': 'warning',
    'port_range': 'warning',
    'source_port': 'warning',
    'source_port_range': 'warning',
    'source_ip_cidr': 'danger',
    'process_name': 'info',
    'process_path': 'info',
    'package_name': 'info',
    'user': 'default',
    'user_id': 'default',
    'clash_mode': 'default',
    'rule_set': 'primary',
    'ip_is_private': 'warning'
  }
  return typeMap[type] || 'default'
}

// 获取匹配条件类型标签
const getConditionTypeLabel = (type) => {
  const typeMap = {
    'domain': '域名',
    'domain_suffix': '域名后缀',
    'domain_keyword': '域名关键词',
    'domain_regex': '域名正则',
    'ip_cidr': 'IP地址段',
    'geoip': 'GeoIP',
    'geosite': 'GeoSite',
    'source_geoip': '源GeoIP',
    'inbound': '入站',
    'protocol': '协议',
    'network': '网络',
    'auth_user': '认证用户',
    'port': '端口',
    'port_range': '端口范围',
    'source_port': '源端口',
    'source_port_range': '源端口范围',
    'source_ip_cidr': '源IP',
    'process_name': '进程名',
    'process_path': '进程路径',
    'package_name': '包名',
    'user': '用户',
    'user_id': '用户ID',
    'clash_mode': 'Clash模式',
    'rule_set': '规则集',
    'ip_is_private': '私有IP'
  }
  return typeMap[type] || type
}

// 获取规则集类型标签样式
const getRuleSetTypeTag = (type) => {
  const typeMap = {
    'remote': 'primary',
    'local': 'success'
  }
  return typeMap[type] || 'default'
}

// 获取出站标签样式
const getOutboundTag = (outbound) => {
  const outboundMap = {
    'direct': 'success',
    'proxy': 'primary',
    'block': 'danger',
    'reject': 'warning'
  }
  return outboundMap[outbound] || 'default'
}

// 获取出站或动作的显示文本
const getOutboundOrActionDisplay = (rule) => {
  // 优先显示动作
  if (rule.action) {
    const actionMap = {
      'sniff': '嗅探'
    }
    return actionMap[rule.action] || rule.action
  }
  
  // 如果没有动作，显示出站
  return rule.outbound || '未设置'
}

// 获取出站或动作的标签样式
const getOutboundOrActionTag = (rule) => {
  // 动作的样式
  if (rule.action) {
    const actionTagMap = {
      'sniff': 'info'
    }
    return actionTagMap[rule.action] || 'default'
  }
  
  // 出站的样式
  return getOutboundTag(rule.outbound)
}

// 获取可用出站列表
const loadAvailableOutbounds = async () => {
  try {
    const response = await apiGetSingBoxOutbounds()
    if (response.code === 200) {
      const outbounds = ['direct', 'block', 'dns-out'] // 默认出站
      response.data.forEach(group => {
        group.nodes.forEach(node => {
          if (!outbounds.includes(node.tag)) {
            outbounds.push(node.tag)
          }
        })
      })
      availableOutbounds.value = outbounds
    }
  } catch (error) {
    console.error('获取出站列表失败:', error)
    availableOutbounds.value = ['direct', 'block', 'dns-out']
  }
}

// 刷新规则列表
const refreshRules = async () => {
  try {
    const response = await apiGetSingBoxRules()
    if (response.code === 200) {
      // 处理路由规则数据
      routeRules.value = response.data.routeRules || []
      // 保存原始规则顺序（用于撤销）
      originalRulesOrder.value = [...routeRules.value]
      currentRulesOrder.value = routeRules.value.map(rule => rule.id)
      // 重置未保存状态
      hasUnsavedChanges.value = false
      // 数据刷新后可以清除重启提醒（说明配置已同步）
      // needsRestart.value = false
      // 处理规则集数据
      ruleSets.value = response.data.ruleSets || []
      // 更新可用规则集列表
      availableRuleSets.value = ruleSets.value.map(rs => rs.tag).filter(tag => tag)
    } else {
      throw new Error(response.message || '获取规则配置失败')
    }
  } catch (error) {
    console.error('获取规则配置失败:', error)
    ElMessage.error('获取规则配置失败: ' + (error.response?.data?.message || error.message))
    // 失败时使用空数组
    routeRules.value = []
    ruleSets.value = []
    availableRuleSets.value = []
  } finally {
    // 数据更新后重新初始化拖拽
    setTimeout(initSortable, 100)
  }
}

// 重置路由规则表单
const resetRuleForm = () => {
  Object.assign(ruleForm, {
    outbound: '',
    ip_version: null,
    invert: false,
    domain: '',
    domain_suffix: '',
    domain_keyword: '',
    domain_regex: '',
    geosite: '',
    ip_cidr: '',
    source_ip_cidr: '',
    geoip: '',
    source_geoip: '',
    network: '',
    protocol: '',
    inbound: '',
    port: '',
    port_range: '',
    source_port: '',
    source_port_range: '',
    process_name: '',
    process_path: '',
    auth_user: '',
    rule_set: '',
    ip_is_private: null
  })
  selectedRuleSets.value = []
  activeCollapse.value = []
}

// 规则集操作已简化为下拉选择，无需额外函数

// 添加路由规则
const addRouteRule = () => {
  editingRule.value = null
  editingRuleIndex.value = -1
  originalRuleBackup.value = null // 新增规则没有原始状态
  resetRuleForm()
  ruleDialogVisible.value = true
}

// 编辑路由规则
const editRouteRule = (rule) => {
  editingRuleIndex.value = routeRules.value.findIndex(r => r.id === rule.id)
  
  // 备份原始状态，用于取消时恢复
  originalRuleBackup.value = JSON.parse(JSON.stringify(rule))
  
  // 创建编辑副本，避免直接修改原始数据
  editingRule.value = JSON.parse(JSON.stringify(rule))
  
  // 填充表单数据
  resetRuleForm()
  ruleForm.outbound = rule.outbound || ''
  ruleForm.ip_version = rule.ip_version || null
  ruleForm.invert = rule.invert || false
  ruleForm.network = rule.network || ''
  ruleForm.ip_is_private = rule.ip_is_private !== undefined ? rule.ip_is_private : null
  
  // 从conditions中提取各种条件
  if (rule.conditions) {
    rule.conditions.forEach(condition => {
      switch (condition.type) {
        case 'domain':
          ruleForm.domain = condition.content
          break
        case 'domain_suffix':
          ruleForm.domain_suffix = condition.content
          break
        case 'domain_keyword':
          ruleForm.domain_keyword = condition.content
          break
        case 'domain_regex':
          ruleForm.domain_regex = condition.content
          break
        case 'geosite':
          ruleForm.geosite = condition.content
          break
        case 'ip_cidr':
          ruleForm.ip_cidr = condition.content
          break
        case 'source_ip_cidr':
          ruleForm.source_ip_cidr = condition.content
          break
        case 'geoip':
          ruleForm.geoip = condition.content
          break
        case 'source_geoip':
          ruleForm.source_geoip = condition.content
          break
        case 'protocol':
          ruleForm.protocol = condition.content
          break
        case 'inbound':
          ruleForm.inbound = condition.content
          break
        case 'port':
          ruleForm.port = condition.content
          break
        case 'source_port':
          ruleForm.source_port = condition.content
          break
        case 'process_name':
          ruleForm.process_name = condition.content
          break
        case 'process_path':
          ruleForm.process_path = condition.content
          break
        case 'rule_set':
          // 将rule_set内容同时设置到选择器和表单字段
          const ruleSets = condition.content.split(/[,\n]/).map(v => v.trim()).filter(v => v)
          selectedRuleSets.value = [...ruleSets]
          ruleForm.rule_set = condition.content
          break
        case 'ip_is_private':
          ruleForm.ip_is_private = condition.content === 'true'
          break
      }
    })
  }
  
  ruleDialogVisible.value = true
}

// 保存路由规则
const saveRouteRule = async () => {
  if (!ruleFormRef.value) return
  
  try {
    await ruleFormRef.value.validate()
    saving.value = true
    
    // 构建规则对象
    const ruleData = {
      outbound: ruleForm.outbound,
    }
    
    // 添加可选字段
    if (ruleForm.ip_version) ruleData.ip_version = ruleForm.ip_version
    if (ruleForm.invert) ruleData.invert = ruleForm.invert
    if (ruleForm.network) ruleData.network = ruleForm.network
    if (ruleForm.ip_is_private !== null) ruleData.ip_is_private = ruleForm.ip_is_private
    
    // 处理规则集选择器的值
    if (selectedRuleSets.value.length > 0) {
      const combinedRuleSets = [...selectedRuleSets.value]
      // 如果传统规则集字段也有值，合并它们
      if (ruleForm.rule_set && ruleForm.rule_set.trim()) {
        const traditionalRuleSets = ruleForm.rule_set.split(',').map(v => v.trim()).filter(v => v)
        combinedRuleSets.push(...traditionalRuleSets)
      }
      
      if (combinedRuleSets.length === 1) {
        ruleData.rule_set = combinedRuleSets[0]
      } else if (combinedRuleSets.length > 1) {
        ruleData.rule_set = combinedRuleSets
      }
    } else if (ruleForm.rule_set && ruleForm.rule_set.trim()) {
      // 只有传统规则集字段有值
      const values = ruleForm.rule_set.split(',').map(v => v.trim()).filter(v => v)
      if (values.length === 1) {
        ruleData.rule_set = values[0]
      } else if (values.length > 1) {
        ruleData.rule_set = values
      }
    }
    
    // 添加各种规则条件 (排除rule_set，已在上面单独处理)
    const fields = [
      'domain', 'domain_suffix', 'domain_keyword', 'domain_regex',
      'ip_cidr', 'geoip', 'geosite', 'source_geoip', 'source_ip_cidr',
      'inbound', 'protocol', 'auth_user', 'port', 'port_range',
      'source_port', 'source_port_range', 'process_name', 'process_path',
      'package_name', 'user', 'user_id', 'clash_mode'
    ]
    
    fields.forEach(field => {
      const value = ruleForm[field]
      if (value !== undefined && value !== null && value !== '') {
        // 对于字符串字段，检查是否为空字符串
        if (typeof value === 'string' && !value.trim()) {
          return
        }
        
        // 处理数组类型字段
        if (['domain', 'domain_suffix', 'domain_keyword', 'domain_regex', 
             'ip_cidr', 'geoip', 'geosite', 'source_geoip', 'source_ip_cidr',
             'inbound', 'protocol', 'auth_user', 'port', 'port_range',
             'source_port', 'source_port_range', 'process_name', 'process_path',
             'package_name', 'user', 'user_id', 'clash_mode'].includes(field)) {
          // 将逗号分隔的字符串转换为数组
          const values = value.split(',').map(v => v.trim()).filter(v => v)
          if (values.length === 1) {
            ruleData[field] = values[0]
          } else if (values.length > 1) {
            ruleData[field] = values
          }
        } else {
          ruleData[field] = value
        }
      }
    })
    
    // 调用API保存规则
    let saveSuccess = false
    if (editingRule.value) {
      await apiUpdateRouteRule(editingRule.value.id, ruleData)
      ElMessage.success('路由规则更新成功')
      saveSuccess = true
    } else {
      await apiCreateRouteRule(ruleData)
      ElMessage.success('路由规则添加成功')
      saveSuccess = true
    }
    
    // 保存成功后清理状态并关闭对话框
    editingRule.value = null
    editingRuleIndex.value = -1
    originalRuleBackup.value = null // 清除备份，避免恢复
    ruleDialogVisible.value = false
    await refreshRules() // 确保刷新完成
    
    // 标记需要重启服务
    if (saveSuccess) {
      needsRestart.value = true
    }
  } catch (error) {
    console.error('保存失败:', error)
    if (error !== 'validation failed') {
      const errorMsg = error.response?.data?.error || error.message || '保存失败'
      ElMessage.error(errorMsg)
    }
  } finally {
    saving.value = false
  }
}

// 取消编辑路由规则
const cancelRuleEdit = () => {
  // 直接关闭对话框，不保存任何更改
  // onRuleDialogClosed 会处理状态清理
  ruleDialogVisible.value = false
}

// 对话框关闭事件处理
const onRuleDialogClosed = () => {
  // 清理编辑状态
  editingRule.value = null
  editingRuleIndex.value = -1
  originalRuleBackup.value = null
  resetRuleForm()
}

// 删除路由规则
const deleteRouteRule = (rule, index) => {
  ElMessageBox.confirm(
    `确定要删除第 ${index + 1} 条路由规则吗？`,
    '确认删除',
    {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      await apiDeleteRouteRule(rule.id)
      ElMessage.success('路由规则删除成功')
      await refreshRules() // 确保刷新完成
      
      // 标记需要重启服务
      needsRestart.value = true
    } catch (error) {
      const errorMsg = error.response?.data?.error || error.message || '删除失败'
      ElMessage.error(errorMsg)
    }
  }).catch(() => {
    // 用户取消删除
  })
}

// 规则集管理方法
const addRuleSet = () => {
  showRuleSetDialog.value = true
  editingRuleSet.value = null
  resetRuleSetForm()
}

const editRuleSet = (ruleSet) => {
  showRuleSetDialog.value = true
  editingRuleSet.value = ruleSet
  // 填充表单数据
  ruleSetForm.tag = ruleSet.tag || ''
  ruleSetForm.type = ruleSet.type || 'remote'
  ruleSetForm.format = ruleSet.format || 'binary'
  ruleSetForm.url = ruleSet.url || ''
  ruleSetForm.path = ruleSet.path || ''
  ruleSetForm.download_detour = ruleSet.download_detour || ''
  ruleSetForm.update_interval = ruleSet.update_interval || ''
}

const deleteRuleSet = (ruleSet) => {
  ElMessageBox.confirm(
    `确定要删除规则集 "${ruleSet.tag}" 吗？`,
    '确认删除',
    {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      ruleSetLoading.value = true
      const { apiDeleteRuleSet } = await import('../utils/api')
      await apiDeleteRuleSet(ruleSet.id)
      ElMessage.success('规则集删除成功')
      needsRestart.value = true
      // 重新加载数据
      await refreshRules()
    } catch (error) {
      ElMessage.error('删除失败: ' + (error.response?.data?.error || error.message))
    } finally {
      ruleSetLoading.value = false
    }
  }).catch(() => {
    // 用户取消删除
  })
}

// 规则集表单管理
const resetRuleSetForm = () => {
  Object.assign(ruleSetForm, {
    tag: '',
    type: 'remote',
    format: 'binary',
    url: '',
    path: '',
    download_detour: '',
    update_interval: ''
  })
  // 清除表单验证状态
  if (ruleSetFormRef.value) {
    ruleSetFormRef.value.clearValidate()
  }
}

const handleRuleSetSave = async () => {
  // 表单验证
  if (!ruleSetFormRef.value) return
  
  const valid = await ruleSetFormRef.value.validate().catch(() => false)
  if (!valid) return

  try {
    ruleSetLoading.value = true
    
    if (editingRuleSet.value) {
      // 编辑模式
      const { apiUpdateRuleSet } = await import('../utils/api')
      await apiUpdateRuleSet(editingRuleSet.value.id, { ...ruleSetForm })
      ElMessage.success('规则集更新成功')
    } else {
      // 添加模式
      const { apiCreateRuleSet } = await import('../utils/api')
      await apiCreateRuleSet({ ...ruleSetForm })
      ElMessage.success('规则集创建成功')
    }
    
    needsRestart.value = true
    showRuleSetDialog.value = false
    await refreshRules()
  } catch (error) {
    ElMessage.error(`操作失败: ${error.response?.data?.error || error.message}`)
  } finally {
    ruleSetLoading.value = false
  }
}

const handleRuleSetCancel = () => {
  showRuleSetDialog.value = false
  resetRuleSetForm()
}

const moveRuleSetUp = (ruleSet) => {
  // TODO: 实现规则集上移功能
  ElMessage.info('规则集排序功能开发中...')
}

const moveRuleSetDown = (ruleSet) => {
  // TODO: 实现规则集下移功能
  ElMessage.info('规则集排序功能开发中...')
}

// 路由规则排序
const isRuleMoving = (rule) => movingRuleIds.value.has(rule.id)

const moveRouteRuleUp = (rule, index) => {
  if (index === 0 || hasUnsavedChanges.value) return
  
  // 本地移动（不立即保存）
  const newOrder = [...routeRules.value]
  const temp = newOrder[index]
  newOrder[index] = newOrder[index - 1]
  newOrder[index - 1] = temp
  
  // 更新显示
  routeRules.value = newOrder
  currentRulesOrder.value = newOrder.map(rule => rule.id)
  hasUnsavedChanges.value = true
  
  ElMessage.info('规则顺序已更改，请点击"保存更改"按钮确认')
}

const moveRouteRuleDown = (rule, index) => {
  if (index === routeRules.value.length - 1 || hasUnsavedChanges.value) return
  
  // 本地移动（不立即保存）
  const newOrder = [...routeRules.value]
  const temp = newOrder[index]
  newOrder[index] = newOrder[index + 1]
  newOrder[index + 1] = temp
  
  // 更新显示
  routeRules.value = newOrder
  currentRulesOrder.value = newOrder.map(rule => rule.id)
  hasUnsavedChanges.value = true
  
  ElMessage.info('规则顺序已更改，请点击"保存更改"按钮确认')
}

// 初始化拖拽排序
const initSortable = () => {
  if (!routeRulesTable.value || routeRules.value.length === 0) return
  
  // 销毁之前的实例
  if (sortableInstance) {
    sortableInstance.destroy()
  }
  
  // 获取表格的tbody元素
  const tbody = routeRulesTable.value.$el.querySelector('.el-table__body-wrapper tbody')
  if (!tbody) return

  sortableInstance = Sortable.create(tbody, {
    animation: 150,
    ghostClass: 'sortable-ghost',
    chosenClass: 'sortable-chosen',
    dragClass: 'sortable-drag',
    onStart: () => {
      isDragging.value = true
    },
    onEnd: (evt) => {
      isDragging.value = false
      
      // 如果位置没有变化，直接返回
      if (evt.oldIndex === evt.newIndex) return
      
      // 更新本地显示顺序（不立即保存）
      const newOrder = [...routeRules.value]
      const movedItem = newOrder.splice(evt.oldIndex, 1)[0]
      newOrder.splice(evt.newIndex, 0, movedItem)
      
      // 更新当前显示的规则顺序
      routeRules.value = newOrder
      currentRulesOrder.value = newOrder.map(rule => rule.id)
      
      // 标记有未保存的更改
      hasUnsavedChanges.value = true
      
      ElMessage.info('规则顺序已更改，请点击"保存更改"按钮确认')
    }
  })
}

// 处理行点击事件（阻止拖拽时的编辑）
const handleRowClick = () => {
  // 拖拽时不响应行点击
  if (isDragging.value) return
}

// 保存更改
const saveChanges = async () => {
  if (!hasUnsavedChanges.value) return
  
  try {
    isSaving.value = true
    
    // 调用批量重排序接口
    await apiReorderRouteRules(currentRulesOrder.value)
    
    ElMessage.success('规则排序保存成功')
    
    // 重新加载数据以确保同步
    await refreshRules()
    
    // 标记需要重启服务
    needsRestart.value = true
    
  } catch (error) {
    console.error('保存失败:', error)
    ElMessage.error(error.response?.data?.error || error.message || '保存失败')
  } finally {
    isSaving.value = false
  }
}

// 撤销更改
const cancelChanges = () => {
  if (!hasUnsavedChanges.value) return
  
  ElMessageBox.confirm(
    '确定要撤销当前的排序更改吗？所有未保存的排序调整将被丢弃。',
    '确认撤销',
    {
      confirmButtonText: '确定撤销',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    // 恢复原始顺序
    routeRules.value = [...originalRulesOrder.value]
    currentRulesOrder.value = originalRulesOrder.value.map(rule => rule.id)
    hasUnsavedChanges.value = false
    
    ElMessage.success('已撤销排序更改')
    
    // 重新初始化拖拽以更新DOM
    setTimeout(initSortable, 100)
  }).catch(() => {
    // 用户取消撤销
  })
}

// 重启服务
const restartService = async () => {
  try {
    isRestarting.value = true
    
    // 第一步：验证当前配置
    ElMessage.info('正在验证配置文件...')
    const validationResult = await apiValidateCurrentSingBoxConfig()
    
    if (!validationResult.data.valid) {
      // 配置验证失败
      ElMessageBox.alert(
        `配置文件验证失败，无法重启服务：\n\n${validationResult.data.error || '未知错误'}`,
        '配置验证失败',
        {
          confirmButtonText: '确定',
          type: 'error',
          dangerouslyUseHTMLString: false
        }
      )
      return
    }
    
    // 第二步：配置验证通过，确认重启
    await ElMessageBox.confirm(
      '✅ 配置文件验证通过！\n\n确定要重启 Sing-Box 服务吗？',
      '确认重启服务',
      {
        confirmButtonText: '重启服务',
        cancelButtonText: '取消',
        type: 'info'
      }
    )
    
    // 第三步：执行重启
    ElMessage.info('正在重启服务...')
    const { apiRestartSingBoxService } = await import('../utils/api')
    await apiRestartSingBoxService()
    
    ElMessage.success('Sing-Box 服务重启成功')
    needsRestart.value = false
    
  } catch (error) {
    // 区分不同类型的错误
    if (error === 'cancel') {
      ElMessage.info('已取消重启服务')
    } else if (error.response?.data?.error) {
      ElMessage.error('重启服务失败: ' + error.response.data.error)
    } else if (error.message) {
      ElMessage.error('重启服务失败: ' + error.message)
    } else {
      ElMessage.error('重启服务失败')
    }
  } finally {
    isRestarting.value = false
  }
}

// 稍后重启（隐藏提醒栏）
const dismissRestart = () => {
  needsRestart.value = false
  ElMessage.info('已隐藏重启提醒，您可以稍后手动重启服务')
}

// 组件挂载时获取数据
onMounted(async () => {
  await loadAvailableOutbounds()
  await refreshRules()
  // 数据加载完成后初始化拖拽
  setTimeout(initSortable, 100)
})
</script>

<style scoped>
.sing-box-rules {
  padding: 15px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  color: #303133;
  font-size: 20px;
}

/* 保存/撤销操作栏样式 */
.save-bar {
  background-color: #fff7e6;
  border: 1px solid #ffd591;
  border-radius: 6px;
  margin-bottom: 20px;
  padding: 0;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
}

.save-bar-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
}

.save-bar-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #d46b08;
  font-size: 14px;
  font-weight: 500;
}

.warning-icon {
  color: #fa8c16;
  font-size: 16px;
}

.save-bar-actions {
  display: flex;
  gap: 8px;
}

/* 重启服务提醒栏样式 */
.restart-bar {
  background-color: #f0f9ff;
  border: 1px solid #91d5ff;
  border-radius: 6px;
  margin-bottom: 20px;
  padding: 0;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
}

.restart-bar-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
}

.restart-bar-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1677ff;
  font-size: 14px;
  font-weight: 500;
}

.restart-icon {
  color: #52c41a;
  font-size: 16px;
}

.restart-bar-actions {
  display: flex;
  gap: 8px;
}

/* 移除不再使用的标签页样式 */

.rules-list-card {
  margin-bottom: 15px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.rule-conditions {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.condition-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.condition-content {
  background-color: #f1f1f1;
  padding: 2px 8px;
  border-radius: 3px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  color: #606266;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.url-content {
  background-color: #f1f1f1;
  padding: 2px 8px;
  border-radius: 3px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  word-break: break-all;
}

.other-conditions {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

/* 规则集选择已改为下拉式，移除旧样式 */

/* 折叠面板样式 */
.rule-collapse {
  margin-top: 16px;
}

.rule-collapse .el-collapse-item__header {
  background-color: #f5f7fa;
  border: 1px solid #e4e7ed;
  padding-left: 12px;
  font-size: 14px;
  color: #606266;
}

.traditional-rules {
  padding: 16px 0;
}

/* 规则集卡片样式 */
.rule-sets-container {
  padding: 0;
}

.rule-sets-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 0 4px;
}

.rule-sets-header .title {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.rule-sets-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.rule-set-card {
  background: white;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 16px;
  transition: all 0.3s ease;
}

.rule-set-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border-color: #c0c4cc;
}

.rule-set-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 12px;
}

.rule-set-icon {
  width: 40px;
  height: 40px;
  background: #f5f7fa;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
  font-size: 18px;
  flex-shrink: 0;
}

.rule-set-info {
  flex: 1;
  min-width: 0;
}

.rule-set-name {
  margin: 0 0 4px 0;
  font-size: 16px;
  font-weight: 500;
  color: #303133;
  line-height: 1.2;
}

.rule-set-details {
  color: #606266;
  font-size: 13px;
  line-height: 1.4;
}

.detail-item {
  display: inline-block;
  margin-right: 16px;
}

.rule-set-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.rule-set-content {
  margin-left: 52px;
}

.rule-set-properties {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 8px;
}

.property {
  display: inline-block;
}

.rule-set-source {
  margin-top: 8px;
}

.source-url {
  display: block;
  background-color: #f5f7fa;
  padding: 8px 12px;
  border-radius: 4px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  color: #606266;
  word-break: break-all;
  line-height: 1.4;
}

.empty-state {
  text-align: center;
  padding: 40px 0;
}

/* 左右布局样式 */
.rules-content-flex {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  width: 100%;
}

.rules-left {
  flex: 0 0 calc(65% - 8px);
  min-width: 600px;
}

.rules-right {
  flex: 0 0 calc(35% - 8px);
  min-width: 300px;
}

/* 紧凑的规则集样式 */
.rule-sets-list-compact {
  display: flex;
  flex-direction: column;
  gap: 8px;
  height: fit-content;
}

.rule-set-card-compact {
  background: #ffffff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 0;
  transition: all 0.3s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.rule-set-card-compact:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  border-color: #409eff;
  transform: translateY(-1px);
}

.rule-set-header-compact {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 10px;
}

.rule-set-info-compact {
  flex: 1;
  min-width: 0;
  margin-right: 10px;
}

.rule-set-name-compact {
  margin: 0 0 8px 0;
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  line-height: 1.3;
  word-break: break-all;
}

.rule-set-properties-compact {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 4px;
}

.rule-set-properties-compact .el-tag {
  font-size: 11px;
  padding: 2px 6px;
  height: 20px;
  line-height: 16px;
}

.rule-set-actions-compact {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex-shrink: 0;
}

.rule-set-actions-compact .el-button {
  padding: 4px 10px;
  font-size: 11px;
  height: 26px;
  min-width: 50px;
}

.rule-set-source-compact {
  margin-top: 10px;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
}

.source-url-compact {
  display: block;
  background-color: #f8f9fa;
  padding: 8px 10px;
  border-radius: 4px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 11px;
  color: #606266;
  word-break: break-all;
  line-height: 1.4;
  border-left: 3px solid #409eff;
}

/* 表格操作按钮对齐 */
.table-actions {
  display: flex;
  gap: 4px;
  align-items: center;
  justify-content: flex-start;
}

.table-actions .el-button {
  padding: 5px 10px;
  font-size: 12px;
  height: 28px;
  min-height: 28px;
}

/* 表格样式优化 */
.rules-list-card .el-table {
  font-size: 13px;
}

.rules-list-card .el-table .el-table__cell {
  padding: 6px 8px;
}

.rules-list-card .el-table .el-table__row {
  height: auto;
  min-height: 36px;
}

.rules-list-card .el-table .el-table__header th {
  padding: 8px;
  font-size: 13px;
  font-weight: 600;
}

/* 条件标签优化 */
.rule-conditions {
  line-height: 1.3;
}

.rule-conditions .condition-item {
  margin-bottom: 2px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.rule-conditions .condition-item:last-child {
  margin-bottom: 0;
}


.condition-content {
  font-size: 11px !important;
  padding: 1px 4px !important;
}

/* 其他条件列优化 */
.other-conditions {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  line-height: 1.3;
}

.other-conditions .el-tag {
  margin: 0;
  font-size: 11px;
  padding: 2px 6px;
  height: 20px;
  line-height: 16px;
}

/* 紧凑的其他条件列 */
.other-conditions-compact {
  display: flex;
  flex-direction: column;
  gap: 2px;
  line-height: 1.1;
}

.other-conditions-compact .el-tag {
  margin: 0;
  font-size: 10px;
  padding: 0 4px;
  height: 18px;
  line-height: 18px;
}

/* 拖拽排序样式 */
.el-table tbody.sortable-ghost {
  opacity: 0.4;
}

.el-table tbody tr.sortable-chosen {
  background-color: #f5f7fa;
}

.el-table tbody tr.sortable-drag {
  background-color: #409eff;
  color: white;
}

.el-table tbody tr {
  cursor: move;
}

.el-table tbody tr:hover {
  background-color: #f5f7fa;
}

</style>
