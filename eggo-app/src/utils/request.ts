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

/** DELETE 请求 */
export function del<T>(url: string): Promise<ApiResponse<T>> {
  return api.delete(url)
}

// ─── 客户端缓存（localStorage，TTL 5 分钟）────────────────────────────────
const CACHE_TTL = 5 * 60 * 1000 // 5 分钟

interface CacheEntry<T> {
  data: T
  ts: number
}

function cacheKey(url: string, params?: Record<string, unknown>): string {
  return 'mkt_cache:' + url + (params ? JSON.stringify(params) : '')
}

/** 读取缓存，过期返回 null */
export function getCache<T>(url: string, params?: Record<string, unknown>): T | null {
  try {
    const raw = localStorage.getItem(cacheKey(url, params))
    if (!raw) return null
    const entry: CacheEntry<T> = JSON.parse(raw)
    if (Date.now() - entry.ts > CACHE_TTL) return null
    return entry.data
  } catch {
    return null
  }
}

/** 写入缓存 */
export function setCache<T>(url: string, data: T, params?: Record<string, unknown>): void {
  try {
    const entry: CacheEntry<T> = { data, ts: Date.now() }
    localStorage.setItem(cacheKey(url, params), JSON.stringify(entry))
  } catch {
    // localStorage 满了就跳过
  }
}

export default api
