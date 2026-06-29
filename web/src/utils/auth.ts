import Cookies from 'js-cookie'

const TokenKey = 'go_admin_token'
const RefreshTokenKey = 'go_admin_refresh_token'

export function getToken(): string | undefined {
  return Cookies.get(TokenKey)
}

export function setToken(token: string, expires: number = 24): void {
  Cookies.set(TokenKey, token, {
    expires,
    sameSite: 'Lax',
    secure: window.location.protocol === 'https:',
  })
}

export function removeToken(): void {
  Cookies.remove(TokenKey)
  Cookies.remove(RefreshTokenKey)
}

export function getRefreshToken(): string | undefined {
  return Cookies.get(RefreshTokenKey)
}

export function setRefreshToken(token: string, expires: number = 7): void {
  Cookies.set(RefreshTokenKey, token, {
    expires,
    sameSite: 'Lax',
    secure: window.location.protocol === 'https:',
  })
}
