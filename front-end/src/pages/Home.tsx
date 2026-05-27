import { Link } from 'react-router-dom'
import { useAppStore } from '../stores/app'

export default function Home() {
  const { apiUrl, setApiUrl, token } = useAppStore()

  const pages = [
    { to: '/login', label: '登录认证', desc: 'JWT 双令牌登录' },
    { to: '/users', label: '用户管理', desc: 'CRUD + 分页查询' },
    { to: '/roles', label: '角色管理', desc: '角色 CRUD + 分配权限' },
    { to: '/permissions', label: '权限管理', desc: '权限 CRUD' },
    { to: '/upload', label: '文件上传', desc: '普通上传 + 分片上传' },
    { to: '/account', label: '账户安全', desc: '密码重置、账户解锁' },
    { to: '/system', label: '系统数据', desc: '菜单树、部门树、字典' },
    { to: '/logs', label: '审计日志', desc: '操作日志查询' },
    { to: '/metrics', label: 'Prometheus', desc: '指标文本输出' },
    { to: '/health', label: '健康检查', desc: '存活 + 就绪检查' },
    { to: '/csrf', label: 'CSRF', desc: 'CSRF Token 获取与验证' },
  ]

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-bold">helloGo 前端测试面板</h1>
      <p className="text-sm text-gray-500">Go + Fiber v2 管理后台 API 手动测试工具</p>

      <div className="flex gap-3 items-center p-3 bg-white border rounded">
        <label className="text-sm font-medium">API 地址</label>
        <input
          className="border px-2 py-1 flex-1 text-sm"
          value={apiUrl}
          onChange={(e) => setApiUrl(e.target.value)}
        />
        <span className={`text-xs px-2 py-0.5 rounded ${token ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}`}>
          {token ? '已登录' : '未登录'}
        </span>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
        {pages.map((p) => (
          <Link
            key={p.to}
            to={p.to}
            className="border p-3 bg-white hover:bg-gray-50 rounded transition"
          >
            <div className="font-medium">{p.label}</div>
            <div className="text-xs text-gray-500 mt-1">{p.desc}</div>
          </Link>
        ))}
      </div>
    </div>
  )
}
