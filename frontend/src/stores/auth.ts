import { defineStore } from 'pinia'
import { ref } from 'vue'
import { authApi } from '@/api'
import router from '@/router'

interface UserInfo {
  id: string
  uid: string
  phone: string
  nickname: string
  avatar: string
  email: string
  last_login: number
  role: string
}

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref(localStorage.getItem('access_token') || '')
  const user = ref<UserInfo | null>(null)
  const isLoggedIn = ref(!!accessToken.value)

  function setAccessToken(token: string) {
    accessToken.value = token
    localStorage.setItem('access_token', token)
    isLoggedIn.value = true
  }

  function logout() {
    accessToken.value = ''
    user.value = null
    isLoggedIn.value = false
    localStorage.removeItem('access_token')
  }

  async function fetchUserInfo() {
    try {
      const res: any = await authApi.getUserInfo()
      if (res.code === 0) {
        user.value = res.data
      } else {
        throw new Error(res.message || 'Failed to fetch user info')
      }
    } catch {
      logout()
      router.push('/login')
    }
  }

  return {
    accessToken,
    user,
    isLoggedIn,
    setAccessToken,
    logout,
    fetchUserInfo,
  }
})
