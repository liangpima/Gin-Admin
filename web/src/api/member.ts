import request from './index'
import type { Result, PageResult } from './index'

export interface MemberItem {
  id: number
  memberNo: string
  username: string
  nickname: string
  avatar: string
  phone: string
  gender: number
  birthday: string
  levelId: number
  status: number
  points: number
  wechatOpenid: string
  registerTime: string
  lastVisitTime: string
  tags: { id: number; name: string; color: string }[]
  createdAt: string
}

export interface MemberLevelItem {
  id: number
  name: string
  minPoints: number
  discount: number
  icon: string
  sort: number
  status: number
}

export interface MemberTagItem {
  id: number
  name: string
  color: string
  sort: number
  status: number
}

export interface PointsLogItem {
  id: number
  memberId: number
  change: number
  type: number
  source: string
  orderNo: string
  createdAt: string
}

export function getMemberList(params: { phone?: string; nickname?: string; levelId?: number; status?: number; page: number; pageSize: number }) {
  return request.get<any, PageResult<MemberItem>>('/member/list', { params })
}

export function createMember(data: any) {
  return request.post<any, Result>('/member', data)
}

export function updateMember(data: any) {
  return request.put<any, Result>('/member', data)
}

export function deleteMember(id: number) {
  return request.delete<any, Result>(`/member/${id}`)
}

export function updateMemberStatus(data: { id: number; status: number }) {
  return request.put<any, Result>('/member/status', data)
}

export function updateMemberTags(data: { id: number; tagIds: number[] }) {
  return request.put<any, Result>('/member/tags', data)
}

export function getMemberLevelList(params: { name?: string; page: number; pageSize: number }) {
  return request.get<any, PageResult<MemberLevelItem>>('/member/level/list', { params })
}

export function getAllMemberLevels() {
  return request.get<any, Result<MemberLevelItem[]>>('/member/level/all')
}

export function createMemberLevel(data: any) {
  return request.post<any, Result>('/member/level', data)
}

export function updateMemberLevel(data: any) {
  return request.put<any, Result>('/member/level', data)
}

export function deleteMemberLevel(id: number) {
  return request.delete<any, Result>(`/member/level/${id}`)
}

export function getMemberTagList(params: { name?: string; page: number; pageSize: number }) {
  return request.get<any, PageResult<MemberTagItem>>('/member/tag/list', { params })
}

export function getAllMemberTags() {
  return request.get<any, Result<MemberTagItem[]>>('/member/tag/all')
}

export function createMemberTag(data: any) {
  return request.post<any, Result>('/member/tag', data)
}

export function updateMemberTag(data: any) {
  return request.put<any, Result>('/member/tag', data)
}

export function deleteMemberTag(id: number) {
  return request.delete<any, Result>(`/member/tag/${id}`)
}

export function getPointsLogList(params: { memberId?: number; type?: number; page: number; pageSize: number }) {
  return request.get<any, PageResult<PointsLogItem>>('/member/points/list', { params })
}
