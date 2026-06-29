import request from './index'
import type { Result } from './index'

export interface DashboardStats {
  userCount: number
  roleCount: number
  menuCount: number
  deptCount: number
  postCount: number
  configCount: number
  logCount: number
}

export function getDashboardStats() {
  return request.get<any, Result<DashboardStats>>('/dashboard/stats')
}
