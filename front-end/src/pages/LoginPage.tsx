import { useState, useEffect } from 'react'
import { api } from '../lib/axios'
import { useAppStore } from '../stores/app'

export default function LoginPage() {
  const { setToken, setSessionId, setRefreshToken, setCsrfToken, csrfToken } = useAppStore()
  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('admin123')
  const [resp, setResp] = useState<unknown>(null)

  // 获取 CSRF token
  useEffect(() => {
    api.get('/api/csrf-token')
      .then((r) => {
        const token = r.data?.csrfToken || ''
        if (token) setCsrfToken(token)
      })
      .catch(() => {})
  }, [setCsrfToken])

  const handleLogin = async () => {
    try {
      const r = await api.post('/api/auth/login', { username, password })
      const data = r.data?.data || r.data
      setToken(data?.accessToken || '')
      setSessionId(data?.sessionId || '')
      setRefreshToken(data?.refreshToken || '')
      setResp(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown }; message?: string }
      setResp(err?.response?.data || { error: err?.message })
    }
  }

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">登录认证</h2>
      <p className="text-sm text-gray-500">
        POST /api/auth/login — 获取 JWT 双令牌（access + refresh）
      </p>

      <div className="p-4 bg-white border rounded space-y-3">
        <div className="text-xs text-gray-400">
          CSRF Token: {csrfToken ? csrfToken.slice(0, 20) + '...' : '未获取'}
        </div>
        <div className="flex gap-2">
          <input
            className="border px-2 py-1"
            placeholder="用户名"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
          />
          <input
            className="border px-2 py-1"
            placeholder="密码"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          <button className="border px-4 py-1 bg-blue-50 hover:bg-blue-100 rounded" onClick={handleLogin}>
            登录
          </button>
        </div>
      </div>

      <pre className="bg-gray-100 p-3 overflow-auto text-xs rounded">{JSON.stringify(resp, null, 2) || ''}</pre>
    </div>
  )
}
