import { useState } from 'react'
import { api } from '../lib/axios'

export default function AccountPage() {
  const [username, setUsername] = useState('admin')
  const [newPassword, setNewPassword] = useState('newpass123')
  const [resetToken, setResetToken] = useState('')
  const [resp, setResp] = useState<unknown>(null)

  const requestReset = async () => {
    try {
      const r = await api.post('/api/auth/password/request-reset', { username })
      const data = r.data?.data || r.data
      setResetToken(data?.token || data?.resetToken || '')
      setResp(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown }; message?: string }
      setResp(err?.response?.data || { error: err?.message })
    }
  }

  const executeReset = async () => {
    try {
      const r = await api.post('/api/auth/password/reset', {
        username,
        newPassword,
        token: resetToken,
      })
      setResp(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown }; message?: string }
      setResp(err?.response?.data || { error: err?.message })
    }
  }

  const wrongLogin = async () => {
    try {
      const r = await api.post('/api/auth/login', { username, password: 'wrong-password-123' })
      setResp(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown }; message?: string }
      setResp(err?.response?.data || { error: err?.message })
    }
  }

  const unlock = async () => {
    try {
      const r = await api.post('/api/auth/unlock', { username })
      setResp(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown }; message?: string }
      setResp(err?.response?.data || { error: err?.message })
    }
  }

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">账户安全</h2>
      <p className="text-sm text-gray-500">密码重置、登录失败锁定与解锁测试</p>

      <div className="p-4 bg-white border rounded space-y-3">
        <div className="flex gap-2 items-center">
          <label className="text-sm">用户名</label>
          <input className="border px-2 py-1" value={username} onChange={(e) => setUsername(e.target.value)} />
          <label className="text-sm">新密码</label>
          <input className="border px-2 py-1" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} />
        </div>
      </div>

      <div className="flex gap-2 flex-wrap">
        <button className="border px-3 py-1 bg-blue-50 hover:bg-blue-100 rounded" onClick={requestReset}>
          请求重置密码
        </button>
        <button className="border px-3 py-1 bg-green-50 hover:bg-green-100 rounded" onClick={executeReset}>
          执行重置
        </button>
        <button className="border px-3 py-1 bg-yellow-50 hover:bg-yellow-100 rounded" onClick={wrongLogin}>
          模拟错误登录
        </button>
        <button className="border px-3 py-1 bg-red-50 hover:bg-red-100 rounded" onClick={unlock}>
          解锁账户
        </button>
      </div>

      <div className="text-xs text-gray-500">
        重置 Token: {resetToken ? resetToken.slice(0, 20) + '...' : '无'}
      </div>

      <pre className="bg-gray-100 p-3 overflow-auto text-xs rounded">{JSON.stringify(resp, null, 2) || ''}</pre>

      <p className="text-xs text-gray-500">
        操作流程：1) 请求重置获取 token → 2) 执行重置 → 3) 多次错误登录触发锁定 → 4) 解锁账户
      </p>
    </div>
  )
}
