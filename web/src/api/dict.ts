import request from './index'

export interface DictTypeItem {
  id: number
  name: string
  type: string
  status: number
}

export interface DictDataItem {
  id: number
  label: string
  value: string
  sort: number
  status: number
  dictType: string
}

export interface DictQuery {
  page: number
  pageSize: number
}

export function getDictTypeList(params: DictQuery) {
  return request.get('/system/dict/type/list', { params })
}

export function createDictType(data: Partial<DictTypeItem>) {
  return request.post('/system/dict/type', data)
}

export function deleteDictType(id: number) {
  return request.delete(`/system/dict/type/${id}`)
}

export function getDictDataList(params: DictQuery & { dictType: string }) {
  return request.get('/system/dict/data/list', { params })
}

export function createDictData(data: Partial<DictDataItem>) {
  return request.post('/system/dict/data', data)
}

export function deleteDictData(id: number) {
  return request.delete(`/system/dict/data/${id}`)
}
