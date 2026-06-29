import request from './index'
import type { Result } from './index'

export interface CaptchaGenerateResponse {
  token: string
  bg: string
  bgWidth: number
  bgHeight: number
  chars: string
}

export interface CaptchaPoint {
  x: number
  y: number
}

export interface CaptchaVerifyResponse {
  success: boolean
  token: string
  message: string
}

export function getCaptcha() {
  return request.get<any, Result<CaptchaGenerateResponse>>('/captcha/generate')
}

export function verifyCaptcha(data: { token: string; points: CaptchaPoint[] }) {
  return request.post<any, Result<CaptchaVerifyResponse>>('/captcha/verify', data)
}
