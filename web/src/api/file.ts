import request from './index'
import type { Result, PageResult } from './index'

export interface FileItem {
  id: number
  name: string
  storageName: string
  path: string
  url: string
  size: number
  mimeType: string
  storageType: number
  createBy: number
  createdAt: string
}

export function uploadFile(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return request.post<any, Result<{ id: number; name: string; url: string; size: number }>>(
    '/system/file/upload',
    formData,
    { headers: { 'Content-Type': 'multipart/form-data' } },
  )
}

export function getFileList(params: { name?: string; mimeType?: string; sortOrder?: string; page: number; pageSize: number }) {
  return request.get<any, PageResult<FileItem>>('/system/file/list', { params })
}

export function deleteFile(id: number) {
  return request.delete<any, Result>(`/system/file/${id}`)
}
