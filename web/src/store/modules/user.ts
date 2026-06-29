import { defineStore } from 'pinia'
import { login as loginApi, getUserInfo, logout as logoutApi } from '@/api/auth'
import { setToken, getToken, removeToken, setRefreshToken, getRefreshToken } from '@/utils/auth'
import type { UserInfoResult } from '@/api/auth'
import { usePermissionStore } from './permission'
import router from '@/router'

interface UserState {
  token: string
  refreshToken: string
  userInfo: UserInfoResult | null
  roles: string[]
  buttons: string[]
}

export const useUserStore = defineStore('user', {
  state: (): UserState => ({
    token: getToken() || '',
    refreshToken: getRefreshToken() || '',
    userInfo: null,
    roles: [],
    buttons: [],
  }),

  actions: {
    async login(username: string, password: string) {
      const res = await loginApi({ username, password })
      this.token = res.data.accessToken
      this.refreshToken = res.data.refreshToken
      setToken(res.data.accessToken)
      setRefreshToken(res.data.refreshToken)
    },

    async getInfo() {
      const res = await getUserInfo()
      this.userInfo = res.data
      this.roles = res.data.roles.map((r) => r.code)
      this.buttons = res.data.buttons
      return res.data
    },

    async logout() {
      try {
        await logoutApi()
      } catch {}
      this.token = ''
      this.refreshToken = ''
      this.userInfo = null
      this.roles = []
      this.buttons = []
      removeToken()

      const permissionStore = usePermissionStore()
      permissionStore.routes = []
      permissionStore.addRoutes = []
      router.getRoutes().forEach((route) => {
        if (route.name && !['Login', 'Dashboard', 'NotFound'].includes(route.name as string)) {
          router.removeRoute(route.name)
        }
      })
    },

    hasButton(code: string): boolean {
      return this.buttons.includes(code)
    },

    hasRole(code: string): boolean {
      return this.roles.includes(code)
    },
  },
})
