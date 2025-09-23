import axios from 'axios'

// 创建axios实例
const api = axios.create({
  baseURL: '/api',
  timeout: 10000
})

// 响应拦截器 - 简化版，无需认证
api.interceptors.response.use(
  response => {
    // 安全地返回响应数据
    if (response && response.data !== undefined) {
      return response.data
    }
    console.warn('API响应格式异常:', response)
    return null
  },
  error => {
    console.error('API请求错误:', error)
    return Promise.reject(error)
  }
)

// 认证相关API - 保留接口以兼容现有代码，但实际不使用
export const apiLogin = (username, password) => {
  // 内网环境直接返回成功，无需实际认证
  return Promise.resolve({
    code: 200,
    message: 'success',
    data: {
      access_token: 'no-auth-needed',
      token_type: 'Bearer',
      expires_in: 86400
    }
  })
}

export const apiGetUser = () => {
  // 返回默认用户信息
  return Promise.resolve({
    code: 200,
    message: 'success',
    data: {
      username: 'admin',
      is_admin: true
    }
  })
}

// 仪表板API
export const apiGetDashboard = () => {
  return api.get('/dashboard')
}

// 服务管理API
export const apiGetServices = () => {
  return api.get('/services')
}

export const apiGetService = (name) => {
  return api.get(`/services/${name}`)
}

export const apiGetServiceInfo = (name) => {
  return api.get(`/services/${name}`)
}

// Sing-Box 配置相关 API
export const apiGetSingBoxConfig = () => {
  return api.get('/singbox/config')
}

export const apiUpdateSingBoxConfig = (config, options = {}) => {
  const { backup = true, autoRestart = false, enableRollback = false } = options
  return api.put('/singbox/config', { 
    config, 
    backup, 
    auto_restart: autoRestart,
    enable_rollback: enableRollback
  })
}

export const apiValidateSingBoxConfig = (config) => {
  return api.post('/singbox/config/validate', { config })
}

export const apiValidateCurrentSingBoxConfig = () => {
  return api.post('/singbox/config/validate-current')
}

// 规则集管理API
export const apiCreateRuleSet = (ruleSetData) => {
  return api.post('/singbox/rulesets', ruleSetData)
}

export const apiUpdateRuleSet = (id, ruleSetData) => {
  return api.put(`/singbox/rulesets/${id}`, ruleSetData)
}

export const apiDeleteRuleSet = (id) => {
  return api.delete(`/singbox/rulesets/${id}`)
}

// 内核更新相关 API
export const apiDetectSingBoxPath = () => {
  return api.get('/singbox/kernel/detect-path')
}

export const apiCheckSingBoxUpdate = () => {
  return api.get('/singbox/kernel/check-update')
}

export const apiUpdateSingBoxKernel = (data) => {
  return api.post('/singbox/kernel/update', data)
}


export const apiControlService = (name, action) => {
  return api.post(`/services/${name}/action`, { action })
}

// 配置管理API
export const apiGetConfig = (service) => {
  return api.get(`/config/${service}`)
}

export const apiUpdateConfig = (service, content, backup = true) => {
  return api.put(`/config/${service}`, { content, backup })
}

// 日志API
export const apiGetLogs = (service, lines = 100, filter = '') => {
  return api.get(`/logs/${service}`, {
    params: { lines, filter_keyword: filter }
  })
}

// Sing-Box模块配置API
export const apiGetSingBoxInbounds = () => api.get('/singbox/inbounds')
export const apiGetSingBoxOutbounds = () => api.get('/singbox/outbounds')  
export const apiGetSingBoxRules = () => api.get('/singbox/rules')

// Sing-Box节点管理API
export const apiCreateSingBoxOutbound = (outbound) => api.post('/singbox/outbounds', outbound)
export const apiUpdateSingBoxOutbound = (id, outbound) => api.put(`/singbox/outbounds/${id}`, outbound)
export const apiDeleteSingBoxOutbound = (id) => api.delete(`/singbox/outbounds/${id}`)

// 重启 Sing-Box 服务
export const apiRestartSingBoxService = () => api.post('/singbox/restart')

// 验证出站节点更改（只验证，不保存）
export const apiValidateOutboundsChanges = (changes) => api.post('/singbox/outbounds/validate', { changes })

// 批量保存出站节点更改
export const apiBatchSaveOutbounds = (changes) => api.post('/singbox/outbounds/batch-save', { changes })

// 路由规则管理
export const apiCreateRouteRule = (rule) => api.post('/singbox/rules/route', rule)
export const apiUpdateRouteRule = (id, rule) => api.put(`/singbox/rules/route/${id}`, rule)
export const apiDeleteRouteRule = (id) => api.delete(`/singbox/rules/route/${id}`)

// 路由规则排序
export const apiMoveRouteRuleUp = (id) => api.post(`/singbox/rules/route/${id}/move-up`)
export const apiMoveRouteRuleDown = (id) => api.post(`/singbox/rules/route/${id}/move-down`)
export const apiReorderRouteRules = (ruleIds) => api.post('/singbox/rules/route/reorder', { rule_ids: ruleIds })

export default api
