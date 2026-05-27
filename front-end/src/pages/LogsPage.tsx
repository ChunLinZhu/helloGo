import { useState } from 'react'
import { api } from '../lib/axios'

export default function LogsPage() {
  const [resp, setResp] = useState<unknown>(null)
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(20)

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">审计日志</h2>
      <p className="text-sm text-gray-500">GET /api/logs — 操作审计日志查询（需要 JWT）</p>

      <div className="p-4 bg-white border rounded space-y-3">
        <div className="flex gap-2 items-center">
          <label className="text-sm">page</label>
          <input className="border px-2 py-1 w-20" type="number" value={page} onChange={(e) => setPage(Number(e.target.value) || 1)} />
          <label className="text-sm">limit</label>
          <input className="border px-2 py-1 w-20" type="number" value={limit} onChange={(e) => setLimit(Number(e.target.value) || 20)} />
          <button
            className="border px-3 py-1 bg-blue-50 hover:bg-blue-100 rounded"
            onClick={async () => {
              try {
                const r = await api.get(`/api/logs?page=${page}&limit=${limit}`)
                setResp(r.data)
              } catch (e: unknown) {
                const err = e as { response?: { data?: unknown } }
                setResp(err?.response?.data || { error: '请求失败' })
              }
            }}
          >
            查询日志
          </button>
        </div>
      </div>

      <pre className="bg-gray-100 p-3 overflow-auto text-xs rounded max-h-[500px]">{JSON.stringify(resp, null, 2) || ''}</pre>
    </div>
  )
}
