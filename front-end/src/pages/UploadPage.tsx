import { useState } from 'react'
import { api } from '../lib/axios'

export default function UploadPage() {
  const [file, setFile] = useState<File | null>(null)
  const [uploadResp, setUploadResp] = useState<unknown>(null)

  // 分片上传
  const [chunkSize, setChunkSize] = useState(512 * 1024) // 512KB
  const [chunkResp, setChunkResp] = useState<unknown>(null)

  // 文件列表
  const [listResp, setListResp] = useState<unknown>(null)

  // 删除
  const [fileId, setFileId] = useState('')
  const [deleteResp, setDeleteResp] = useState<unknown>(null)

  const handleUpload = async () => {
    if (!file) return
    try {
      const fd = new FormData()
      fd.append('file', file)
      const r = await api.post('/api/uploads', fd, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      setUploadResp(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setUploadResp(err?.response?.data)
    }
  }

  const handleChunkUpload = async () => {
    if (!file) return
    try {
      const id = `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
      const total = Math.ceil(file.size / chunkSize)

      for (let i = 0; i < total; i++) {
        const start = i * chunkSize
        const end = Math.min(start + chunkSize, file.size)
        const blob = file.slice(start, end)
        const fd = new FormData()
        fd.append('chunk', new File([blob], `chunk-${i}.bin`, { type: file.type }))
        fd.append('fileId', id)
        fd.append('index', String(i))
        fd.append('total', String(total))
        fd.append('filename', file.name)
        fd.append('mimetype', file.type)
        await api.post('/api/uploads/chunk', fd, {
          headers: { 'Content-Type': 'multipart/form-data' },
        })
      }

      const r = await api.post('/api/uploads/merge', {
        fileId: id,
        total,
        filename: file.name,
        mimetype: file.type,
      })
      setChunkResp(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown }; message?: string }
      setChunkResp(err?.response?.data || { error: err?.message })
    }
  }

  const fetchList = async () => {
    try {
      const r = await api.get('/api/uploads?page=1&limit=20')
      setListResp(r.data)
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setListResp(err?.response?.data)
    }
  }

  const deleteFile = async () => {
    try {
      const r = await api.delete(`/api/uploads/${fileId}`)
      setDeleteResp(r.data)
      fetchList()
    } catch (e: unknown) {
      const err = e as { response?: { data?: unknown } }
      setDeleteResp(err?.response?.data)
    }
  }

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">文件上传</h2>
      <p className="text-sm text-gray-500">需要先登录获取 token</p>

      <div className="p-4 bg-white border rounded space-y-3">
        <h3 className="font-medium">选择文件</h3>
        <input type="file" onChange={(e) => setFile(e.target.files?.[0] || null)} />
        {file && <p className="text-xs text-gray-500">{file.name} ({(file.size / 1024).toFixed(1)} KB)</p>}
      </div>

      {/* 普通上传 */}
      <div className="p-4 bg-white border rounded space-y-3">
        <h3 className="font-medium">普通上传 POST /api/uploads</h3>
        <button className="border px-3 py-1 bg-green-50 hover:bg-green-100 rounded" onClick={handleUpload}>
          上传文件
        </button>
        {!!uploadResp && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded">{JSON.stringify(uploadResp, null, 2) || ''}</pre>}
      </div>

      {/* 分片上传 */}
      <div className="p-4 bg-white border rounded space-y-3">
        <h3 className="font-medium">分片上传 POST /api/uploads/chunk + /api/uploads/merge</h3>
        <div className="flex gap-2 items-center">
          <label className="text-sm">分片大小(字节)</label>
          <input className="border px-2 py-1 w-28" type="number" value={chunkSize} onChange={(e) => setChunkSize(Number(e.target.value) || 524288)} />
          <button className="border px-3 py-1 bg-yellow-50 hover:bg-yellow-100 rounded" onClick={handleChunkUpload}>
            分片上传并合并
          </button>
        </div>
        {!!chunkResp && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded">{JSON.stringify(chunkResp, null, 2) || ''}</pre>}
      </div>

      {/* 文件列表 */}
      <div className="p-4 bg-white border rounded space-y-3">
        <h3 className="font-medium">文件列表 GET /api/uploads</h3>
        <button className="border px-3 py-1 bg-blue-50 hover:bg-blue-100 rounded" onClick={fetchList}>查询</button>
        {!!listResp && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded max-h-60">{JSON.stringify(listResp, null, 2) || ''}</pre>}
      </div>

      {/* 删除 */}
      <div className="p-4 bg-white border rounded space-y-3">
        <h3 className="font-medium">删除文件 DELETE /api/uploads/:id</h3>
        <div className="flex gap-2">
          <input className="border px-2 py-1" placeholder="File ID" value={fileId} onChange={(e) => setFileId(e.target.value)} />
          <button className="border px-3 py-1 bg-red-50 hover:bg-red-100 rounded" onClick={deleteFile}>删除</button>
        </div>
        {!!deleteResp && <pre className="bg-gray-100 p-2 text-xs overflow-auto rounded">{JSON.stringify(deleteResp, null, 2) || ''}</pre>}
      </div>
    </div>
  )
}
