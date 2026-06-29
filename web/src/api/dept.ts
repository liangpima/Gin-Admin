import request from './index'
import type { Result } from './index'

export interface DeptItem {
  id: number
  parentId: number
  name: string
  sort: number
  leader: string
  phone: string
  email: string
  status: number
  children?: DeptItem[]
}

export function getDeptTree() {
  return request.get<any, Result<DeptItem[]>>('/system/dept/tree')
}

export function createDept(data: any) {
  return request.post<any, Result>('/system/dept', data)
}

export function updateDept(data: any) {
  return request.put<any, Result>('/system/dept', data)
}

export function deleteDept(id: number) {
  return request.delete<any, Result>(`/system/dept/${id}`)
}
