import request from '@/utils/request'
import type { ICommonResponse } from '@/types/common'

export interface INavMenu {
  id?: number
  title: string
  path: string
  position: 'top_nav' | 'home_promo'
  sort_order: number
  visible: boolean
}

export const navMenuPage = (params?: any) =>
  request.get<ICommonResponse<{ list: INavMenu[]; total: number }>>('/nav-menus', { params })

export const navMenuCreate = (data: INavMenu) =>
  request.post<ICommonResponse<unknown>>('/nav-menus', data)

export const navMenuUpdate = (id: number, data: INavMenu) =>
  request.put<ICommonResponse<unknown>>(`/nav-menus/${id}`, data)

export const navMenuDelete = (id: number) =>
  request.delete<ICommonResponse<unknown>>(`/nav-menus/${id}`)

