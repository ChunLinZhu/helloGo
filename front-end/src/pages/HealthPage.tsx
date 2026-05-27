import { useState } from 'react'
import { api } from '../lib/axios'

export default function HealthPage() {
  const [health, setHealth] = useState<unknown>(null)
  const [ready, setReady] = useState<unknown>(null)

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">健康检查</h2>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="p-4 bg-white border rounded space-y-3">
          <h3 className="font-medium">存活检查 /api/health</h3>
          <p className="text-xs text-gray-500">不依赖外部服务，仅检查进程是否存活</p>
          <button
            className="border px-3 py-1 bg-blue-50 hover:bg-blue-100 rounded"
            onClick={async () => {
              const r = await api.get('/api/health')
              setHealth(r.data)
            }}
          >
            检查
          </button>
          <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded">{JSON.stringify(health, null, 2) || ''}</pre>
        </div>

        <div className="p-4 bg-white border rounded space-y-3">
          <h3 className="font-medium">就绪检查 /api/health/ready</h3>
          <p className="text-xs text-gray-500">验证 DB + Redis 连接状态</p>
          <button
            className="border px-3 py-1 bg-blue-50 hover:bg-blue-100 rounded"
            onClick={async () => {
              try {
                const r = await api.get('/api/health/ready')
                setReady(r.data)
              } catch (e: unknown) {
                const err = e as { response?: { data?: unknown } }
                setReady(err?.response?.data || { error: '请求失败' })
              }
            }}
          >
            检查
          </button>
          <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded">{JSON.stringify(ready, null, 2) || ''}</pre>
        </div>
      </div>
    </div>
  )
}
