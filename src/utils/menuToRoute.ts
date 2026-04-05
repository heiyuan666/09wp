import { type RouteComponent, type RouteRecordRaw } from 'vue-router'
import type { IMenuItem } from '@/types/system/menu'

const modules = import.meta.glob('@/views/**/*.vue')

const componentName = (path: string) => {
  const parts = path.split('/').filter(Boolean)
  const base = parts
    .map((part) => part.charAt(0).toUpperCase() + part.substring(1))
    .join('')
  return `${base || 'Page'}View`
}

const normalizeMenuPath = (rawPath: string) => {
  return rawPath.replace(/^\/+/, '').replace(/^admin\/+/, '')
}

export const menuToRoute = (menuList: IMenuItem[]) => {
  const dynamicRoute: RouteRecordRaw[] = []

  menuList.forEach((menu) => {
    if (menu.type === 'menu') {
      const path = normalizeMenuPath(menu.path || '')
      if (!path) return

      const component = modules[`/src/views/${path}/index.vue`] as RouteComponent | undefined
      if (!component) return

      dynamicRoute.push({
        path,
        name: componentName(path),
        component,
        meta: {
          icon: menu.icon,
          title: menu.title,
          id: menu.id,
          parentId: menu.parentId,
          keepAlive: true,
        },
      })
    }

    if (menu.type === 'directory' && menu.children?.length) {
      dynamicRoute.push(...menuToRoute(menu.children))
    }
  })

  return dynamicRoute
}
