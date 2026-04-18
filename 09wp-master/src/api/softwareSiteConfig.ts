import request from '@/utils/request'
import type { ICommonResponse } from '@/types/common'

export interface ISoftwareSiteConfig {
  id?: number
  site_title: string
  logo_url: string
  favicon_url: string
  seo_keywords: string
  seo_description: string
}

/** 管理端：/api/v1/game/software/site-config */
export const getSoftwareSiteConfig = () =>
  request.get<ICommonResponse<ISoftwareSiteConfig>>('/game/software/site-config')

export const updateSoftwareSiteConfig = (data: Partial<ISoftwareSiteConfig>) =>
  request.put<ICommonResponse<unknown>>('/game/software/site-config', data)
