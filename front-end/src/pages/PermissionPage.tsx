import { useState } from 'react'
import { api } from '../lib/axios'

export default function PermissionPage() {
  const [listResp, setListResp] = useState<unknown>(null)
  const [newPerm, setNewPerm] = useState({ code: '', name: '', description: '' })
  const [createResp, setCreateResp] = useState<unknown>(null)

  const [permId, setPermId] = useState('')
  const [detailResp, setDetailResp] = useState<unknown>(null)
  const [deleteResp, setDeleteResp] = useState<unknown>(null)

  const fetchList = async () => {
    try {
      const r = await api.get('/api/permissions?page=1&limit=50')
      setListResp(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setListResp(err?.response?.data)
    }
  }

  const createPerm = async () => {
    try {
      const r = await api.post('/api/permissions', newPerm)
      setCreateResp(r.data)
      fetchList()
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setCreateResp(err?.response?.data)
    }
  }

  const getPerm = async () => {
    try {
      const r = await api.get(`/api/permissions/${permId}`)
      setDetailResp(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setDetailResp(err?.response?.data)
    }
  }

  const deletePerm = async () => {
    try {
      const r = await api.delete(`/api/permissions/${permId}`)
      setDeleteResp(r.data)
      fetchList()
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setDeleteResp(err?.response?.data)
    }
  }

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">权限管理</h2>

      {/* 列表 */}
      <div className="p-4 bg-white border rounded space-y-3">
        <h3 className="font-medium">权限列表 GET /api/permissions</h3>
        <button className="border px-3 py-1 bg-blue-50 hover:bg-blue-100 rounded" onClick={fetchList}>刷新列表</button>
        {!!listResp && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded max-h-60">{JSON.stringify(listResp, null, 2) || ''}</pre>}
      </div>

      {/* 创建 */}
      <div className="p-4 bg-white border rounded space-y-3">
        <h3 className="font-medium">创建权限 POST /api/permissions</h3>
        <div className="flex gap-2">
          <input className="border px-2 py-1" placeholder="Code (如 user:view)" value={newPerm.code} onChange={(e) => setNewPerm({ ...newPerm, code: e.target.value })} />
          <input className="border px-2 py-1" placeholder="Name" value={newPerm.name} onChange={(e) => setNewPerm({ ...newPerm, name: e.target.value })} />
          <input className="border px-2 py-1" placeholder="描述（可选）" value={newPerm.description} onChange={(e) => setNewPerm({ ...newPerm, description: e.target.value })} />
          <button className="border px-3 py-1 bg-green-50 hover:bg-green-100 rounded" onClick={createPerm}>创建</button>
        </div>
        {!!createResp && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded">{JSON.stringify(createResp, null, 2) || ''}</pre>}
      </div>

      {/* 详情/删除 */}
      <div className="p-4 bg-white border rounded space-y-3">
        <h3 className="font-medium">按 ID 操作</h3>
        <div className="flex gap-2 items-center">
          <input className="border px-2 py-1" placeholder="Permission ID" value={permId} onChange={(e) => setPermId(e.target.value)} />
          <button className="border px-3 py-1 bg-blue-50 hover:bg-blue-100 rounded" onClick={getPerm}>查询</button>
          <button className="border px-3 py-1 bg-red-50 hover:bg-red-100 rounded" onClick={deletePerm}>删除</button>
        </div>
        {!!detailResp && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded">{JSON.stringify(detailResp, null, 2) || ''}</pre>}
        {!!deleteResp && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded">{JSON.stringify(deleteResp, null, 2) || ''}</pre>}
      </div>
    </div>
  )
}
