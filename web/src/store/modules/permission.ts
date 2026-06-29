import { defineStore } from 'pinia'
import type { RouteRecordRaw } from 'vue-router'
import { constantRoutes } from '@/router/routes/static'
import Layout from '@/layout/index.vue'

const viewModules = import.meta.glob('@/views/**/*.vue')

interface PermissionState {
  routes: RouteRecordRaw[]
  addRoutes: RouteRecordRaw[]
}

export const usePermissionStore = defineStore('permission', {
  state: (): PermissionState => ({
    routes: [],
    addRoutes: [],
  }),

  actions: {
    setRoutes(routes: RouteRecordRaw[]) {
      this.addRoutes = routes
      this.routes = constantRoutes.concat(routes)
    },

    generateRoutes(menus: any[]) {
      const accessedRoutes = filterAsyncRoutes(menus)
      this.setRoutes(accessedRoutes)
      return accessedRoutes
    },
  },
})

function filterAsyncRoutes(menus: any[]): RouteRecordRaw[] {
  const res: RouteRecordRaw[] = []
  menus.forEach((menu) => {
    if (menu.type === 2) return

    const hasChildren = menu.children && menu.children.filter((c: any) => c.type !== 2).length > 0
    const componentPath = menu.component
      ? viewModules[`/src/views/${menu.component}.vue`]
      : (hasChildren ? Layout : undefined)

    if (!componentPath) return

    const route: any = {
      path: menu.path || '',
      name: menu.name,
      component: componentPath,
      meta: {
        title: menu.title || menu.name,
        icon: menu.icon,
        hidden: menu.visible === 0,
        noCache: menu.isCache === 0,
      },
    }

    if (hasChildren) {
      route.children = filterAsyncRoutes(menu.children)
    }

    res.push(route)
  })
  return res
}
