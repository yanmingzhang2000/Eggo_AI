import type { ApiResponse } from '@/types/egg'

const BASE_URL = 'http://localhost:8080/api/v1'

interface RequestOptions {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: Record<string, unknown>
  header?: Record<string, string>
}

/** 封装 uni.request 为 Promise */
function request<T>(options: RequestOptions): Promise<ApiResponse<T>> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}${options.url}`,
      method: options.method || 'GET',
      data: options.data,
      header: {
        'Content-Type': 'application/json',
        ...options.header,
      },
      success: (res) => {
        const statusCode = res.statusCode
        const data = res.data as ApiResponse<T>

        if (statusCode >= 200 && statusCode < 300) {
          resolve(data)
        } else {
          uni.showToast({ title: data.message || '请求失败', icon: 'none' })
          reject(data)
        }
      },
      fail: (err) => {
        uni.showToast({ title: '网络异常', icon: 'none' })
        reject(err)
      },
    })
  })
}

/** GET 请求 */
export function get<T>(url: string, data?: Record<string, unknown>) {
  return request<T>({ url, method: 'GET', data })
}

/** POST 请求 */
export function post<T>(url: string, data?: Record<string, unknown>) {
  return request<T>({ url, method: 'POST', data })
}
