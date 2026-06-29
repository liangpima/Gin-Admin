import request from './index'

export interface AgreementItem {
  id: number
  title: string
  content: string
  type: string
  sort: number
  status: number
  createdAt: string
}

export interface AgreementQuery {
  name?: string
  type?: string
  status?: number
  page: number
  pageSize: number
}

export function getAgreementList(params: AgreementQuery) {
  return request.get('/system/agreement/list', { params })
}

export function createAgreement(data: Partial<AgreementItem>) {
  return request.post('/system/agreement', data)
}

export function updateAgreement(data: Partial<AgreementItem>) {
  return request.put('/system/agreement', data)
}

export function deleteAgreement(id: number) {
  return request.delete(`/system/agreement/${id}`)
}

export function getAgreementByType(type: string) {
  return request.get(`/system/agreement/type/${type}`)
}
