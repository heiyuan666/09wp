import request from '@/utils/request'
import type { ICommonResponse } from '@/types/common'

export interface IGameCategory {
  id: number
  name: string
  slug: string
  description: string
  created_at: string
  updated_at: string
}

export interface IGameItem {
  id: number
  category_id?: number
  steam_appid?: number
  required_age?: number
  is_free?: boolean
  title: string
  cover: string
  banner: string
  video_url: string
  short_description?: string
  supported_languages?: string
  reviews?: string
  pc_requirements?: string
  mac_requirements?: string
  linux_requirements?: string
  header_image?: string
  website?: string
  developers?: string
  publishers?: string
  platforms?: string
  genres?: string
  tags?: string
  price_text?: string
  price_currency?: string
  price_initial?: number
  price_final?: number
  price_discount?: number
  metacritic_score?: number
  description: string
  release_date?: string
  size: string
  type: string
  developer: string
  rating: number
  steam_score: number
  recommendations_total?: number
  downloads: number
  likes: number
  dislikes: number
  gallery: string[]
  created_at: string
  updated_at: string
}

export interface IGameResource {
  id: number
  game_id: number
  title: string
  resource_type: 'game' | 'mod' | 'trainer' | 'submission'
  version: string
  size: string
  download_type: string
  pan_type: string
  download_url: string
  download_urls?: string[]
  tested: boolean
  author: string
  publish_date?: string
  created_at: string
  updated_at: string
}

export interface ISteamAppDetail {
  appid: number
  required_age?: number
  name: string
  type: string
  is_free: boolean
  short_description: string
  detailed_description: string
  about_the_game: string
  supported_languages?: string
  reviews?: string
  pc_requirements?: string
  mac_requirements?: string
  linux_requirements?: string
  header_image: string
  capsule_image: string
  background: string
  background_raw: string
  website: string
  developers: string[]
  publishers: string[]
  categories: string[]
  genres: string[]
  tags: string[]
  platforms?: { windows?: boolean; mac?: boolean; linux?: boolean }
  screenshots: string[]
  video_url: string
  price_text: string
  price_currency: string
  price_initial: number
  price_final: number
  price_discount: number
  metacritic_score: number
  metacritic_url: string
  recommendations_total?: number
  release_date: string
  coming_soon: boolean
  cc: string
  l: string
}

export const gameCategoryList = () => request.get<ICommonResponse<IGameCategory[]>>('/game/category/list')
export const gameCategoryCreate = (data: Partial<IGameCategory>) =>
  request.post<ICommonResponse<unknown>>('/game/category/create', data)
export const gameCategoryUpdate = (id: number | string, data: Partial<IGameCategory>) =>
  request.put<ICommonResponse<unknown>>(`/game/category/${id}`, data)
export const gameCategoryDelete = (id: number | string) =>
  request.delete<ICommonResponse<unknown>>(`/game/category/${id}`)

export const gameList = (params?: Record<string, unknown>) =>
  request.get<ICommonResponse<{ list: IGameItem[]; total: number }>>('/game/list', { params })
export const gameDetail = (id: number | string) => request.get<ICommonResponse<IGameItem>>(`/game/detail/${id}`)
export const gameCreate = (data: Partial<IGameItem>) => request.post<ICommonResponse<unknown>>('/game/create', data)
export const gameUpdate = (id: number | string, data: Partial<IGameItem>) =>
  request.put<ICommonResponse<unknown>>(`/game/${id}`, data)
export const gameDelete = (id: number | string) => request.delete<ICommonResponse<unknown>>(`/game/${id}`)
export const gameSteamAppDetail = (appid: number | string, params?: { cc?: string; l?: string }) =>
  request.get<ICommonResponse<ISteamAppDetail>>(`/game/steam/app/${appid}`, { params })
export const gameSteamSearch = (name: string, params?: { cc?: string; l?: string }) =>
  request.get<
    ICommonResponse<{
      search_term: string
      count: number
      hint?: string
      data: Array<{ appid: number; name: string; icon: string; match_score: number }>
    }>
  >(
    '/game/steam/search',
    { params: { name, ...(params || {}) } },
  )

export const gameResourceList = (gameId: number | string) =>
  request.get<ICommonResponse<IGameResource[]>>('/game/resource/list', { params: { game_id: gameId } })
export const gameResourceCreate = (data: Partial<IGameResource>) =>
  request.post<ICommonResponse<unknown>>('/game/resource/create', data)
export const gameResourceUpdate = (id: number | string, data: Partial<IGameResource>) =>
  request.put<ICommonResponse<unknown>>(`/game/resource/${id}`, data)
export const gameResourceDelete = (id: number | string) =>
  request.delete<ICommonResponse<unknown>>(`/game/resource/${id}`)

export const gameUpload = (file: File, dir = 'game-gallery') => {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('dir', dir)
  return request.post<ICommonResponse<{ url: string }>>('/game/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}
