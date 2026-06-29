import request from './index'
import type { Result } from './index'

export interface MenuItem {
  id: number
  parentId: number
  name: string
  path: string
  component: string
  redirect: string
  icon: string
  title: string
  type: number
  permission: string
  sort: number
  visible: number
  status: number
  isExternal: number
  isCache: number
  children?: MenuItem[]
}

export function getMenuTree() {
  return request.get<any, Result<MenuItem[]>>('/system/menu/tree')
}

export function getAllMenus() {
  return request.get<any, Result<MenuItem[]>>('/system/menu/all')
}

export function createMenu(data: any) {
  return request.post<any, Result>('/system/menu', data)
}

export function updateMenu(data: any) {
  return request.put<any, Result>('/system/menu', data)
}

export function deleteMenu(id: number) {
  return request.delete<any, Result>(`/system/menu/${id}`)
}
