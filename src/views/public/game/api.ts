import type { ICommonResponse } from '@/types/common'
import publicRequest from '@/utils/publicRequest'

export type PublicGameCategory = {
  id: number
  name: string
  slug: string
  description: string
}

export type PublicGameItem = {
  id: number
  category_id?: number
  steam_appid?: number
  title: string
  cover?: string
  banner?: string
  header_image?: string
  video_url?: string
  short_description?: string
  description?: string
  website?: string
  publishers?: string
  genres?: string
  tags?: string
  price_text?: string
  release_date?: string
  size?: string
  developer?: string
  type?: string
  rating?: number
  steam_score?: number
  downloads?: number
  likes?: number
  dislikes?: number
  gallery?: string[]
  resources?: PublicGameResource[]
  created_at?: string
  updated_at?: string
}

export type PublicGameResource = {
  id: number
  game_id: number
  title: string
  resource_type?: 'game' | 'mod' | 'trainer' | 'submission'
  version?: string
  size?: string
  download_type?: string
  pan_type?: string
  download_url: string
  download_urls?: string[]
  tested?: boolean
  author?: string
  publish_date?: string
  created_at?: string
  updated_at?: string
}

export const publicGameCategoryList = () => publicRequest.get<ICommonResponse<PublicGameCategory[]>>('/game/category/list')

export const publicGameList = (params?: Record<string, unknown>) =>
  publicRequest.get<ICommonResponse<{ list: PublicGameItem[]; total: number }>>('/game/list', { params })

export const publicGameDetail = (id: number | string) =>
  publicRequest.get<ICommonResponse<PublicGameItem>>(`/game/detail/${id}`)

export const publicGameResourceList = (gameId: number | string) =>
  publicRequest.get<ICommonResponse<PublicGameResource[]>>('/game/resource/list', {
    params: { game_id: gameId },
  })

export type PublicNavMenu = {
  id: number
  title: string
  path?: string
  position: string
  sort_order?: number
  visible?: boolean
}

export const publicNavMenuList = (position: 'top_nav' | 'home_promo') =>
  publicRequest.get<ICommonResponse<{ list: PublicNavMenu[] }>>('/public/nav-menus', {
    params: { position },
  })

export type PublicResourceItem = {
  id: number
  title: string
  cover?: string
  created_at?: string
}

export const publicResourceList = (params?: Record<string, unknown>) =>
  publicRequest.get<ICommonResponse<{ list: PublicResourceItem[]; total: number }>>('/resources', { params })
