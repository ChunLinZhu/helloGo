import { useState } from 'react'
import { api } from '../lib/axios'

export default function MetricsPage() {
  const [text, setText] = useState('')

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">Prometheus 指标</h2>
      <p className="text-sm text-gray-500">GET /api/metrics — Prometheus 文本格式输出</p>

      <button
        className="border px-3 py-1 bg-blue-50 hover:bg-blue-100 rounded"
        onClick={async () => {
          const r = await api.get('/api/metrics', { responseType: 'text' })
          setText(r.data as string)
        }}
      >
        拉取指标
      </button>

      <pre className="bg-gray-100 p-3 overflow-auto text-xs whitespace-pre-wrap rounded max-h-[600px]">
        {text || '点击"拉取指标"查看 Prometheus 文本输出'}
      </pre>
    </div>
  )
}
