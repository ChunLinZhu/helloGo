import { useState } from 'react'
import { api } from '../lib/axios'

export default function UserPage() {
  // 分页查询
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(10)
  const [listResp, setListResp] = useState<unknown>(null)

  // 创建用户
  const [newUser, setNewUser] = useState({ username: '', password: '', email: '' })
  const [createResp, setCreateResp] = useState<unknown>(null)

  // 查询/更新/删除
  const [userId, setUserId] = useState('')
  const [detailResp, setDetailResp] = useState<unknown>(null)
  const [updateEmail, setUpdateEmail] = useState('')
  const [updateResp, setUpdateResp] = useState<unknown>(null)
  const [deleteResp, setDeleteResp] = useState<unknown>(null)

  const fetchList = async () => {
    try {
      const r = await api.get(`/api/users?page=${page}&limit=${limit}`)
      setListResp(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setListResp(err?.response?.data)
    }
  }

  const createUser = async () => {
    try {
      const r = await api.post('/api/users', newUser)
      setCreateResp(r.data)
      fetchList()
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setCreateResp(err?.response?.data)
    }
  }

  const getUser = async () => {
    try {
      const r = await api.get(`/api/users/${userId}`)
      setDetailResp(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setDetailResp(err?.response?.data)
    }
  }

  const updateUser = async () => {
    try {
      const r = await api.patch(`/api/users/${userId}`, { email: updateEmail })
      setUpdateResp(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setUpdateResp(err?.response?.data)
    }
  }

  const deleteUser = async () => {
    try {
      const r = await api.delete(`/api/users/${userId}`)
      setDeleteResp(r.data)
      fetchList()
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setDeleteResp(err?.response?.data)
    }
  }

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">用户管理</h2>
      <p className="text-sm text-gray-500">需要 admin 角色 token，先登录后操作</p>

      {/* 创建用户 */}
      <div className="p-4 bg-white border rounded space-y-3">
        <h3 className="font-medium">创建用户 POST /api/users</h3>
        <div className="flex gap-2">
          <input className="border px-2 py-1" placeholder="用户名" value={newUser.username} onChange={(e) => setNewUser({ ...newUser, username: e.target.value })} />
          <input className="border px-2 py-1" placeholder="密码" value={newUser.password} onChange={(e) => setNewUser({ ...newUser, password: e.target.value })} />
          <input className="border px-2 py-1" placeholder="邮箱（可选）" value={newUser.email} onChange={(e) => setNewUser({ ...newUser, email: e.target.value })} />
          <button className="border px-3 py-1 bg-green-50 hover:bg-green-100 rounded" onClick={createUser}>创建</button>
        </div>
        {!!createResp && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded">{JSON.stringify(createResp, null, 2) || ''}</pre>}
      </div>

      {/* 分页查询 */}
      <div className="p-4 bg-white border rounded space-y-3">
        <h3 className="font-medium">分页查询 GET /api/users</h3>
        <div className="flex gap-2 items-center">
          <label className="text-sm">page</label>
          <input className="border px-2 py-1 w-20" type="number" value={page} onChange={(e) => setPage(Number(e.target.value) || 1)} />
          <label className="text-sm">limit</label>
          <input className="border px-2 py-1 w-20" type="number" value={limit} onChange={(e) => setLimit(Number(e.target.value) || 10)} />
          <button className="border px-3 py-1 bg-blue-50 hover:bg-blue-100 rounded" onClick={fetchList}>查询</button>
        </div>
        {!!listResp && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded max-h-60">{JSON.stringify(listResp, null, 2) || ''}</pre>}
      </div>

      {/* 详情/更新/删除 */}
      <div className="p-4 bg-white border rounded space-y-3">
        <h3 className="font-medium">按 ID 操作</h3>
        <div className="flex gap-2 items-center flex-wrap">
          <input className="border px-2 py-1" placeholder="User ID" value={userId} onChange={(e) => setUserId(e.target.value)} />
          <button className="border px-3 py-1 bg-blue-50 hover:bg-blue-100 rounded" onClick={getUser}>查询</button>
          <input className="border px-2 py-1" placeholder="新邮箱" value={updateEmail} onChange={(e) => setUpdateEmail(e.target.value)} />
          <button className="border px-3 py-1 bg-yellow-50 hover:bg-yellow-100 rounded" onClick={updateUser}>更新</button>
          <button className="border px-3 py-1 bg-red-50 hover:bg-red-100 rounded" onClick={deleteUser}>删除</button>
        </div>
        {!!detailResp && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded">{JSON.stringify(detailResp, null, 2) || ''}</pre>}
        {!!updateResp && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded">{JSON.stringify(updateResp, null, 2) || ''}</pre>}
        {!!deleteResp && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded">{JSON.stringify(deleteResp, null, 2) || ''}</pre>}
      </div>
    </div>
  )
}
