import { useState } from 'react'
import { api } from '../lib/axios'
import { useAppStore } from '../stores/app'

export default function CsrfPage() {
  const { setCsrfToken, csrfToken } = useAppStore()
  const [resp, setResp] = useState<unknown>(null)

  const fetchToken = async () => {
    try {
      const r = await api.get('/api/csrf-token')
      const token = r.data?.csrfToken || ''
      if (token) setCsrfToken(token)
      setResp(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown }; message?: string }
      setResp(err?.response?.data || { error: err?.message })
    }
  }

  const testLogin = async () => {
    try {
      const r = await api.post('/api/auth/login', { username: 'admin', password: 'admin123' })
      setResp(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown }; message?: string }
      setResp(err?.response?.data || { error: err?.message })
    }
  }

  const testWithoutCsrf = async () => {
    // 临时清除 CSRF token 来测试不带 token 的请求
    const saved = csrfToken
    setCsrfToken('')
    try {
      const r = await api.post('/api/auth/login', { username: 'admin', password: 'admin123' })
      setResp(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown }; message?: string }
      setResp(err?.response?.data || { error: err?.message })
    } finally {
      setCsrfToken(saved)
    }
  }

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">CSRF 防护测试</h2>
      <p className="text-sm text-gray-500">
        验证 CSRF header 模式：POST/PUT/DELETE 请求必须携带有效的 X-CSRF-Token 头
      </p>

      <div className="p-4 bg-white border rounded space-y-3">
        <h3 className="font-medium">1. 获取 CSRF Token</h3>
        <p className="text-xs text-gray-500">GET /api/csrf-token — 返回 JWT 签名的 token（30 分钟有效）</p>
        <button className="border px-3 py-1 bg-blue-50 hover:bg-blue-100 rounded" onClick={fetchToken}>
          获取 Token
        </button>
        <div className="text-xs text-gray-500 mt-2">
          当前 Token: {csrfToken ? csrfToken.slice(0, 30) + '...' : '无'}
        </div>
      </div>

      <div className="p-4 bg-white border rounded space-y-3">
        <h3 className="font-medium">2. 携带 Token 登录</h3>
        <p className="text-xs text-gray-500">POST /api/auth/login — axios 拦截器会自动注入 X-CSRF-Token 头</p>
        <button className="border px-3 py-1 bg-green-50 hover:bg-green-100 rounded" onClick={testLogin}>
          携带 Token 登录
        </button>
      </div>

      <div className="p-4 bg-white border rounded space-y-3">
        <h3 className="font-medium">3. 不带 Token 登录（应被拒绝）</h3>
        <p className="text-xs text-gray-500">不带 X-CSRF-Token 头发起 POST 请求，期望返回 403</p>
        <button className="border px-3 py-1 bg-red-50 hover:bg-red-100 rounded" onClick={testWithoutCsrf}>
          不带 Token 登录
        </button>
      </div>

      <pre className="bg-gray-100 p-3 overflow-auto text-xs rounded">{JSON.stringify(resp, null, 2) || ''}</pre>

      <p className="text-xs text-gray-500">
        操作流程：1) 获取 Token → 2) 携带 Token 登录（应成功）→ 3) 不带 Token 登录（应返回 403）
      </p>
    </div>
  )
}
