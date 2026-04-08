import request from '@/utils/request'
import type { ICommonResponse } from '@/types/common'

export interface IGameSiteConfig {
  id?: number
  site_title: string
  logo_url: string
  favicon_url: string
  seo_keywords: string
  seo_description: string
}

export const getGameSiteConfig = () => request.get<ICommonResponse<IGameSiteConfig>>('/game/config')

export const updateGameSiteConfig = (data: IGameSiteConfig) =>
  request.put<ICommonResponse<unknown>>('/game/config', data)

