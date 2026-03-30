import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import router from '@/router'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
  withCredentials: true, // 携带 HttpOnly Cookie (refresh_token)
})

// 请求拦截：自动注入 Access Token
api.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.accessToken) {
    config.headers.Authorization = `Bearer ${auth.accessToken}`
  }
  return config
})

// 响应拦截：自动刷新 Token（全局只刷新一次，其余排队等待）
let isRefreshing = false
let pendingQueue: Array<{ resolve: (token: string) => void; reject: (err: any) => void }> = []

function processPendingQueue(token: string | null, error?: any) {
  pendingQueue.forEach(({ resolve, reject }) => {
    token ? resolve(token) : reject(error)
  })
  pendingQueue = []
}

api.interceptors.response.use(
  (response) => response.data,
  async (error) => {
    const originalRequest = error.config
    const auth = useAuthStore()

    if (
      error.response?.status === 401 &&
      !originalRequest._retry &&
      !originalRequest.url?.includes('/refresh')
    ) {
      // 如果已经在刷新中，当前请求排队等待
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          pendingQueue.push({
            resolve: (newToken: string) => {
              originalRequest.headers.Authorization = `Bearer ${newToken}`
              resolve(api(originalRequest))
            },
            reject,
          })
        })
      }

      originalRequest._retry = true
      isRefreshing = true

      try {
        const res = await axios.post('/api/v1/user/refresh', null, {
          withCredentials: true,
        })
        if (res.data.code === 0) {
          const newToken = res.data.data.access_token
          auth.setAccessToken(newToken)
          originalRequest.headers.Authorization = `Bearer ${newToken}`
          // 通知所有排队的请求用新 token 重发
          processPendingQueue(newToken)
          return api(originalRequest)
        }
      } catch {
        processPendingQueue(null, error)
        auth.logout()
        router.push('/login')
        return Promise.reject(error)
      } finally {
        isRefreshing = false
      }
    }

    return Promise.reject(error.response?.data || error)
  }
)

export default api

// --- API 函数 ---

export const authApi = {
  register: (data: { phone: string; password: string; nickname?: string }) =>
    api.post('/user/register', data),
  login: (data: { phone: string; password: string }) =>
    api.post('/user/login', data),
  refresh: () => api.post('/user/refresh'),
  getUserInfo: () => api.get('/user/info'),
  updatePassword: (data: any) => api.post('/user/password', data),
}

export const emailApi = {
  bind: (data: any) => api.post('/email/bind', data),
  getAccounts: () => api.get('/email/accounts'),
  deleteAccount: (id: string) => api.delete(`/email/accounts/${id}`),
}

export const billApi = {
  getTrendStats: () => api.get('/bill/stats/trend'),
  getCategoryStats: () => api.get('/bill/stats/category'),
  getDashboardStats: () => api.get('/bill/dashboard'),
  getBillDetail: (id: string) => api.get(`/bill/${id}`),
  getBillList: (params: { page: number; size: number; keyword?: string; category?: string; date?: string }) => api.get('/bill/list', { params }),
  uploadImage: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/bill/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 60000, // AI分析需要更长时间
    })
  },
  updateBillRemark: (id: string, remark: string) => api.put(`/bill/${id}/remark`, { remark }),
  updateBill: (id: string, data: any) => api.put(`/bill/${id}`, data),
  deleteBill: (id: string) => api.delete(`/bill/${id}`),
  createBill: (data: { amount: number; merchant: string; category: string; remark: string; created_at: string }) => api.post('/bill/manual', data),
}
