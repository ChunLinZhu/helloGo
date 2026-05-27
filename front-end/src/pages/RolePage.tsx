import { useState } from 'react'
import { api } from '../lib/axios'

export default function RolePage() {
  const [listResp, setListResp] = useState<unknown>(null)
  const [newRole, setNewRole] = useState({ code: '', name: '' })
  const [createResp, setCreateResp] = useState<unknown>(null)

  // 分配权限
  const [roleId, setRoleId] = useState('')
  const [permCodes, setPermCodes] = useState('')
  const [assignResp, setAssignResp] = useState<unknown>(null)

  // 详情/删除
  const [detailId, setDetailId] = useState('')
  const [detailResp, setDetailResp] = useState<unknown>(null)
  const [deleteResp, setDeleteResp] = useState<unknown>(null)

  const fetchList = async () => {
    try {
      const r = await api.get('/api/roles?page=1&limit=50')
      setListResp(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setListResp(err?.response?.data)
    }
  }

  const createRole = async () => {
    try {
      const r = await api.post('/api/roles', newRole)
      setCreateResp(r.data)
      fetchList()
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setCreateResp(err?.response?.data)
    }
  }

  const assignPermissions = async () => {
    try {
      const perms = permCodes.split(',').map((s) => s.trim()).filter(Boolean)
      const r = await api.post(`/api/roles/${roleId}/permissions`, { permissionCodes: perms })
      setAssignResp(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setAssignResp(err?.response?.data)
    }
  }

  const getRole = async () => {
    try {
      const r = await api.get(`/api/roles/${detailId}`)
      setDetailResp(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setDetailResp(err?.response?.data)
    }
  }

  const deleteRole = async () => {
    try {
      const r = await api.delete(`/api/roles/${detailId}`)
      setDeleteResp(r.data)
      fetchList()
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setDeleteResp(err?.response?.data)
    }
  }

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">角色管理</h2>

      {/* 角色列表 */}
      <div className="p-4 bg-white border rounded space-y-3">
        <h3 className="font-medium">角色列表 GET /api/roles</h3>
        <button className="border px-3 py-1 bg-blue-50 hover:bg-blue-100 rounded" onClick={fetchList}>刷新列表</button>
        {!!listResp && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded max-h-60">{JSON.stringify(listResp, null, 2) || ''}</pre>}
      </div>

      {/* 创建角色 */}
      <div className="p-4 bg-white border rounded space-y-3">
        <h3 className="font-medium">创建角色 POST /api/roles</h3>
        <div className="flex gap-2">
          <input className="border px-2 py-1" placeholder="Code (如 editor)" value={newRole.code} onChange={(e) => setNewRole({ ...newRole, code: e.target.value })} />
          <input className="border px-2 py-1" placeholder="Name (如 编辑)" value={newRole.name} onChange={(e) => setNewRole({ ...newRole, name: e.target.value })} />
          <button className="border px-3 py-1 bg-green-50 hover:bg-green-100 rounded" onClick={createRole}>创建</button>
        </div>
        {!!createResp && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded">{JSON.stringify(createResp, null, 2) || ''}</pre>}
      </div>

      {/* 分配权限 */}
      <div className="p-4 bg-white border rounded space-y-3">
        <h3 className="font-medium">分配权限 POST /api/roles/:id/permissions</h3>
        <div className="flex gap-2">
          <input className="border px-2 py-1" placeholder="Role ID" value={roleId} onChange={(e) => setRoleId(e.target.value)} />
          <input className="border px-2 py-1 flex-1" placeholder="权限 Code（逗号分隔）" value={permCodes} onChange={(e) => setPermCodes(e.target.value)} />
          <button className="border px-3 py-1 bg-yellow-50 hover:bg-yellow-100 rounded" onClick={assignPermissions}>分配</button>
        </div>
        {!!assignResp && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded">{JSON.stringify(assignResp, null, 2) || ''}</pre>}
      </div>

      {/* 详情/删除 */}
      <div className="p-4 bg-white border rounded space-y-3">
        <h3 className="font-medium">按 ID 操作</h3>
        <div className="flex gap-2 items-center">
          <input className="border px-2 py-1" placeholder="Role ID" value={detailId} onChange={(e) => setDetailId(e.target.value)} />
          <button className="border px-3 py-1 bg-blue-50 hover:bg-blue-100 rounded" onClick={getRole}>查询</button>
          <button className="border px-3 py-1 bg-red-50 hover:bg-red-100 rounded" onClick={deleteRole}>删除</button>
        </div>
        {!!detailResp && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded">{JSON.stringify(detailResp, null, 2) || ''}</pre>}
        {!!deleteResp && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded">{JSON.stringify(deleteResp, null, 2) || ''}</pre>}
      </div>
    </div>
  )
}
