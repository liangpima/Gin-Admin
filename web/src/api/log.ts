import request from './index'

export interface OperationLogItem {
  id: number
  title: string
  operatorName: string
  requestMethod: string
  requestUrl: string
  status: number
  ip: string
  costTime: number
  createdAt: string
}

export interface LoginLogItem {
  id: number
  username: string
  ip: string
  browser: string
  os: string
  status: number
  msg: string
  loginTime: string
}

export interface LogQuery {
  page: number
  pageSize: number
  title?: string
  username?: string
}

export function getOperationLogs(params: LogQuery) {
  return request.get('/system/log/operation', { params })
}

export function getLoginLogs(params: LogQuery) {
  return request.get('/system/log/login', { params })
}
