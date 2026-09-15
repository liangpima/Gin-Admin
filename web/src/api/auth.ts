import request from './index'
import type { Result } from './index'

export interface LoginParams {
  username: string
  password: string
  /** 验证码校验通过后返回的 token，后端据此确认本次登录已过人机校验 */
  captchaToken: string
}

export interface LoginResult {
  accessToken: string
  refreshToken: string
  expiresIn: number
  tokenType: string
}

export interface UserInfoResult {
  id: number
  username: string
  nickname: string
  avatar: string
  email: string
  phone: string
  roles: { id: number; name: string; code: string }[]
  buttons: string[]
  menus: any[]
}

export function login(data: LoginParams) {
  return request.post<any, Result<LoginResult>>('/auth/login', data)
}

export function refreshToken(data: { refreshToken: string }) {
  return request.post<any, Result<LoginResult>>('/auth/refresh', data)
}

export function logout(refreshToken?: string) {
  return request.post<any, Result>('/auth/logout', { refreshToken })
}

export function getUserInfo() {
  return request.get<any, Result<UserInfoResult>>('/auth/userInfo')
}
