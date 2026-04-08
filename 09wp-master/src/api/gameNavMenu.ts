import request from '@/utils/request'
import type { ICommonResponse } from '@/types/common'

export interface IGameNavMenu {
  id?: number
  title: string
  path: string
  position: 'top_nav' | 'home_promo'
  sort_order: number
  visible: boolean
}

export const gameNavMenuPage = (params?: any) =>
  request.get<ICommonResponse<{ list: IGameNavMenu[]; total: number }>>('/game/nav-menus', { params })

export const gameNavMenuCreate = (data: IGameNavMenu) =>
  request.post<ICommonResponse<unknown>>('/game/nav-menus', data)

export const gameNavMenuUpdate = (id: number, data: IGameNavMenu) =>
  request.put<ICommonResponse<unknown>>(`/game/nav-menus/${id}`, data)

export const gameNavMenuDelete = (id: number) =>
  request.delete<ICommonResponse<unknown>>(`/game/nav-menus/${id}`)

