import request from './index'
import type { Result, PageResult } from './index'

export interface RoleItem {
  id: number
  name: string
  code: string
  sort: number
  status: number
  dataScope: number
  createdAt: string
}

export function getRoleList(params: any) {
  return request.get<any, Result<PageResult<RoleItem>>>('/system/role/list', { params })
}

export function getAllRoles() {
  return request.get<any, Result<RoleItem[]>>('/system/role/all')
}

export function getRoleById(id: number) {
  return request.get<any, Result<RoleItem>>(`/system/role/${id}`)
}

export function createRole(data: any) {
  return request.post<any, Result>('/system/role', data)
}

export function updateRole(data: any) {
  return request.put<any, Result>('/system/role', data)
}

export function deleteRole(id: number) {
  return request.delete<any, Result>(`/system/role/${id}`)
}
