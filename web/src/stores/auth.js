import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  // 内网环境，无需实际认证，始终返回已认证状态
  const token = ref('no-auth-needed')
  const user = ref({
    username: 'admin',
    is_admin: true
  })

  const isAuthenticated = computed(() => true) // 内网环境始终认证通过

  const login = async (username, password) => {
    // 内网环境直接返回成功
    return true
  }

  const logout = () => {
    // 内网环境无需退出逻辑
  }

  const getUserInfo = async () => {
    // 内网环境返回默认用户信息
    return user.value
  }

  return {
    token,
    user,
    isAuthenticated,
    login,
    logout,
    getUserInfo
  }
})
