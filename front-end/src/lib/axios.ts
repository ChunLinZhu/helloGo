import axios, { AxiosError, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { useAppStore } from '../stores/app'
import { toast } from 'sonner'

export const api = axios.create()

// 请求拦截器：注入 baseURL、Authorization、CSRF Token
api.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const { apiUrl, token, csrfToken } = useAppStore.getState()
  config.baseURL = apiUrl
  if (token) {
    config.headers.set('Authorization', `Bearer ${token}`)
  }
  if (csrfToken) {
    config.headers.set('X-CSRF-Token', csrfToken)
  }
  console.log('→', config.method?.toUpperCase(), config.baseURL + config.url)
  return config
})

// 响应拦截器：日志 + 错误提示
api.interceptors.response.use(
  (resp: AxiosResponse) => {
    console.log('←', resp.status, resp.config.url, resp.data)
    return resp
  },
  (err: AxiosError) => {
    const msg = (err.response?.data as Record<string, unknown>)?.message || err.message
    toast.error(String(msg))
    console.error('✗', err.response?.status, err.config?.url, err.response?.data)
    return Promise.reject(err)
  }
)
