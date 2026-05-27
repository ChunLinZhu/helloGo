import { useState } from 'react'
import { api } from '../lib/axios'

export default function SystemPage() {
  const [menus, setMenus] = useState<unknown>(null)
  const [depts, setDepts] = useState<unknown>(null)
  const [dicts, setDicts] = useState<unknown>(null)

  const fetchMenus = async () => {
    try {
      const r = await api.get('/api/menus/tree')
      setMenus(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setMenus(err?.response?.data)
    }
  }

  const fetchDepts = async () => {
    try {
      const r = await api.get('/api/departments/tree')
      setDepts(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setDepts(err?.response?.data)
    }
  }

  const fetchDicts = async () => {
    try {
      const r = await api.get('/api/dicts?page=1&limit=50')
      setDicts(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setDicts(err?.response?.data)
    }
  }

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">系统数据</h2>
      <p className="text-sm text-gray-500">菜单树、部门树、字典管理（需要 JWT）</p>

      <div className="p-4 bg-white border rounded space-y-3">
        <div className="flex items-center gap-3">
          <h3 className="font-medium">菜单树 GET /api/menus/tree</h3>
          <button className="border px-3 py-1 bg-blue-50 hover:bg-blue-100 rounded text-sm" onClick={fetchMenus}>
            获取
          </button>
        </div>
        {!!menus && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded max-h-60">{JSON.stringify(menus, null, 2) || ''}</pre>}
      </div>

      <div className="p-4 bg-white border rounded space-y-3">
        <div className="flex items-center gap-3">
          <h3 className="font-medium">部门树 GET /api/departments/tree</h3>
          <button className="border px-3 py-1 bg-blue-50 hover:bg-blue-100 rounded text-sm" onClick={fetchDepts}>
            获取
          </button>
        </div>
        {!!depts && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded max-h-60">{JSON.stringify(depts, null, 2) || ''}</pre>}
      </div>

      <div className="p-4 bg-white border rounded space-y-3">
        <div className="flex items-center gap-3">
          <h3 className="font-medium">字典列表 GET /api/dicts</h3>
          <button className="border px-3 py-1 bg-blue-50 hover:bg-blue-100 rounded text-sm" onClick={fetchDicts}>
            获取
          </button>
        </div>
        {!!dicts && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded max-h-60">{JSON.stringify(dicts, null, 2) || ''}</pre>}
      </div>
    </div>
  )
}
