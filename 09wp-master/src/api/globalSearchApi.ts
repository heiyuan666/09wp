import request from '@/utils/request'
import type { ICommonResponse } from '@/types/common'

export interface IGlobalSearchAPIItem {
  id?: number
  name: string
  api_url: string
  cloud_types?: string
  enabled?: boolean
  sort_order?: number
}

/** 全网搜站点级配置（在「全网搜接口」页维护） */
export interface IGlobalSearchSettings {
  global_search_enabled: boolean
  global_search_link_check_enabled: boolean
  global_search_api_url: string
  global_search_cloud_types: string
  global_search_default_category_id: number
  global_search_auto_transfer: boolean
  global_search_cleanup_enabled: boolean
  global_search_cleanup_days: number
  global_search_cleanup_minutes: number
  global_search_cleanup_delete_netdisk_files: boolean
  /** 可选：从 url 字段中提取真实链接的正则（空则内置规则） */
  global_search_url_sanitize_regex?: string
}

export const globalSearchSettingsGet = () =>
  request.get<ICommonResponse<IGlobalSearchSettings>>('/system/global-search/settings')

export const globalSearchSettingsPut = (data: IGlobalSearchSettings) =>
  request.put<ICommonResponse<unknown>>('/system/global-search/settings', data)

export const globalSearchAPIList = () =>
  request.get<ICommonResponse<{ list: IGlobalSearchAPIItem[]; total: number }>>('/system/global-search/apis')

export const globalSearchAPICreate = (data: IGlobalSearchAPIItem) =>
  request.post<ICommonResponse<unknown>>('/system/global-search/apis', data)

export const globalSearchAPIUpdate = (id: number, data: IGlobalSearchAPIItem) =>
  request.put<ICommonResponse<unknown>>(`/system/global-search/apis/${id}`, data)

export const globalSearchAPIDelete = (id: number) =>
  request.delete<ICommonResponse<unknown>>(`/system/global-search/apis/${id}`)

