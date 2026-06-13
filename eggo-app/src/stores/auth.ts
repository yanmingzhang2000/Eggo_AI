import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { post } from '@/utils/request'

interface UserInfo {
  id: string
  username: string
  email: string
}

interface AuthResponse {
  token: string
  user: UserInfo
  isGuest: boolean
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const user = ref<UserInfo | null>(null)
  const isGuest = ref(false)

  const isLoggedIn = computed(() => !!token.value)
  const username = computed(() => user.value?.username || '游客')

  // 初始化时从 localStorage 恢复
  function init() {
    const savedToken = localStorage.getItem('token')
    const savedUser = localStorage.getItem('user')
    const savedIsGuest = localStorage.getItem('isGuest')

    if (savedToken) {
      token.value = savedToken
      isGuest.value = savedIsGuest === 'true'
      if (savedUser) {
        try {
          user.value = JSON.parse(savedUser)
        } catch (e) {
          user.value = null
        }
      }
    }
  }

  // 游客登录
  async function guestLogin() {
    try {
      const res = await post<AuthResponse>('/auth/guest')
      if (res.code === 0 && res.data) {
        setAuth(res.data.token, res.data.user, true)
        return true
      }
      return false
    } catch (err) {
      console.error('Guest login failed:', err)
      return false
    }
  }

  // 用户登录
  async function login(username: string, password: string) {
    try {
      const res = await post<AuthResponse>('/auth/login', { username, password })
      if (res.code === 0 && res.data) {
        setAuth(res.data.token, res.data.user, false)
        return { success: true }
      }
      return { success: false, message: res.message }
    } catch (err: any) {
      return { success: false, message: err.response?.data?.message || '登录失败' }
    }
  }

  // 用户注册
  async function register(username: string, email: string, password: string) {
    try {
      const res = await post<AuthResponse>('/auth/register', { username, email, password })
      if (res.code === 0 && res.data) {
        setAuth(res.data.token, res.data.user, false)
        return { success: true }
      }
      return { success: false, message: res.message }
    } catch (err: any) {
      return { success: false, message: err.response?.data?.message || '注册失败' }
    }
  }

  // 退出登录
  function logout() {
    token.value = ''
    user.value = null
    isGuest.value = false
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    localStorage.removeItem('isGuest')
  }

  // 设置认证信息
  function setAuth(newToken: string, newUser: UserInfo, guest: boolean) {
    token.value = newToken
    user.value = newUser
    isGuest.value = guest
    localStorage.setItem('token', newToken)
    localStorage.setItem('user', JSON.stringify(newUser))
    localStorage.setItem('isGuest', String(guest))
  }

  // 初始化
  init()

  return {
    token,
    user,
    isGuest,
    isLoggedIn,
    username,
    guestLogin,
    login,
    register,
    logout,
  }
})
