import request from './index'
import type { Result, PageResult } from './index'

export interface PayOrder {
  id: number
  orderNo: string
  tradeNo: string
  subject: string
  body: string
  amount: number
  currency: string
  channel: string
  status: number
  paidAt: string
  refundAt: string
  refundAmt: number
  openId: string
  createdAt: string
}

export interface CreateOrderParams {
  subject: string
  body?: string
  amount: number
  channel: 'wechat' | 'alipay'
  openId?: string
  extra?: string
}

export function createPayOrder(data: CreateOrderParams) {
  return request.post<any, Result<{
    orderNo: string
    amount: number
    formUrl?: string
    codeUrl?: string
    paySign?: string
    appId?: string
    timeStamp?: string
    nonceStr?: string
    package?: string
    orderString?: string
  }>>('/system/pay/order', data)
}

export function getPayOrder(orderNo: string) {
  return request.get<any, Result<PayOrder>>('/system/pay/order', { params: { orderNo } })
}

export function closePayOrder(orderNo: string) {
  return request.post<any, Result>('/system/pay/order/close', { orderNo })
}

export function getPayOrderList(params: { subject?: string; status?: string; channel?: string; page: number; pageSize: number }) {
  return request.get<any, PageResult<PayOrder>>('/system/pay/order/list', { params })
}

export function queryPayOrder(orderNo: string) {
  return request.get<any, Result<{ orderNo: string; status: number; paidAt: string }>>('/system/pay/order/query', { params: { orderNo } })
}
