import React from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route, Link } from 'react-router-dom'
import { Toaster } from 'sonner'
import './styles.css'
import Home from './pages/Home'
import LoginPage from './pages/LoginPage'
import UserPage from './pages/UserPage'
import RolePage from './pages/RolePage'
import PermissionPage from './pages/PermissionPage'
import UploadPage from './pages/UploadPage'
import AccountPage from './pages/AccountPage'
import SystemPage from './pages/SystemPage'
import LogsPage from './pages/LogsPage'
import MetricsPage from './pages/MetricsPage'
import HealthPage from './pages/HealthPage'
import CsrfPage from './pages/CsrfPage'
import { useAppStore } from './stores/app'
import { api } from './lib/axios'

function LogoutButton() {
  const { sessionId, logout } = useAppStore()
  if (!sessionId) return null
  return (
    <button
      className="ml-auto text-sm text-red-600 hover:underline"
      onClick={async () => {
        try {
          await api.post('/api/auth/logout', { sessionId })
        } catch {
          // 即使请求失败也要登出
        } finally {
          logout()
          window.location.href = '/login'
        }
      }}
    >
      退出登录
    </button>
  )
}

function App() {
  return (
    <BrowserRouter>
      <div className="min-h-screen bg-gray-50 text-gray-900">
        <nav className="flex gap-3 p-3 bg-white shadow sticky top-0 z-10 items-center text-sm">
          <Link to="/" className="font-semibold hover:underline">首页</Link>
          <Link to="/login" className="hover:underline">登录</Link>
          <Link to="/users" className="hover:underline">用户</Link>
          <Link to="/roles" className="hover:underline">角色</Link>
          <Link to="/permissions" className="hover:underline">权限</Link>
          <Link to="/upload" className="hover:underline">上传</Link>
          <Link to="/account" className="hover:underline">账户安全</Link>
          <Link to="/system" className="hover:underline">系统</Link>
          <Link to="/logs" className="hover:underline">日志</Link>
          <Link to="/metrics" className="hover:underline">指标</Link>
          <Link to="/health" className="hover:underline">健康</Link>
          <Link to="/csrf" className="hover:underline">CSRF</Link>
          <LogoutButton />
        </nav>
        <main className="p-4 max-w-6xl mx-auto">
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/login" element={<LoginPage />} />
            <Route path="/users" element={<UserPage />} />
            <Route path="/roles" element={<RolePage />} />
            <Route path="/permissions" element={<PermissionPage />} />
            <Route path="/upload" element={<UploadPage />} />
            <Route path="/account" element={<AccountPage />} />
            <Route path="/system" element={<SystemPage />} />
            <Route path="/logs" element={<LogsPage />} />
            <Route path="/metrics" element={<MetricsPage />} />
            <Route path="/health" element={<HealthPage />} />
            <Route path="/csrf" element={<CsrfPage />} />
          </Routes>
        </main>
      </div>
      <Toaster position="top-right" richColors />
    </BrowserRouter>
  )
}

createRoot(document.getElementById('root')!).render(<App />)
