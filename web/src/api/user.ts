import request from './index'
import type { Result, PageResult } from './index'

export interface UserItem {
  id: number
  username: string
  nickname: string
  email: string
  phone: string
  avatar: string
  status: number
  deptId: number
  roles: { id: number; name: string; code: string }[]
  createdAt: string
}

export interface UserListParams {
  username?: string
  phone?: string
  status?: number
  deptId?: number
  page: number
  pageSize: number
}

export function getUserList(params: UserListParams) {
  return request.get<any, Result<PageResult<UserItem>>>('/system/user/list', { params })
}

export function getUserById(id: number) {
  return request.get<any, Result<UserItem>>(`/system/user/${id}`)
}

export function createUser(data: any) {
  return request.post<any, Result>('/system/user', data)
}

export function updateUser(data: any) {
  return request.put<any, Result>('/system/user', data)
}

export function deleteUser(id: number) {
  return request.delete<any, Result>(`/system/user/${id}`)
}

export function updateUserStatus(data: { id: number; status: number }) {
  return request.put<any, Result>('/system/user/status', data)
}

export function updateUserRoles(data: { id: number; roleIds: number[] }) {
  return request.put<any, Result>('/system/user/roles', data)
}

export function updateUserDept(data: { id: number; deptId: number }) {
  return request.put<any, Result>('/system/user/dept', data)
}

export function resetPassword(data: { id: number; password: string }) {
  return request.put<any, Result>('/system/user/resetPwd', data)
}

export function changePassword(data: { oldPassword: string; newPassword: string }) {
  return request.put<any, Result>('/system/user/changePwd', data)
}
