import axios from 'axios'
import type { ApiResponse } from '@/types/egg'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器：自动添加 Token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器
api.interceptors.response.use(
  (response) => response.data,
  (error) => {
    const message = error.response?.data?.message || '网络异常'
    console.error('[API Error]', message)
    return Promise.reject(error)
  }
)

/** GET 请求 */
export function get<T>(url: string, params?: Record<string, unknown>): Promise<ApiResponse<T>> {
  return api.get(url, { params })
}

/** POST 请求 */
export function post<T>(url: string, data?: Record<string, unknown>): Promise<ApiResponse<T>> {
  return api.post(url, data)
}

export default api
