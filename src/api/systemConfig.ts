import request from '@/utils/request'
import type { ICommonResponse } from '@/types/common'

/** 友情链接（后台系统配置） */
export interface IFriendLinkItem {
  title: string
  url: string
}

export interface ISystemConfig {
  id?: number
  site_title: string
  admin_email: string
  support_email: string
  contact_phone: string
  contact_qq: string
  logo_url: string
  favicon_url: string
  seo_keywords: string
  seo_description: string
  icp_record: string
  footer_text: string
  clarity_project_id?: string
  clarity_enabled?: boolean
  friend_links?: IFriendLinkItem[]
  allow_register: boolean
  submission_need_review?: boolean
  submission_auto_transfer?: boolean
  resource_detail_auto_transfer?: boolean
  haoka_user_id?: string
  haoka_secret?: string
  haoka_sync_enabled?: boolean
  haoka_sync_interval?: number
  // 前台号卡详情页按钮跳转链接（可配置）
  haoka_order_url?: string
  haoka_agent_reg_url?: string
  smtp_host: string
  smtp_port: number
  smtp_user: string
  smtp_pass: string
  smtp_from: string
  tg_bot_token?: string
  tg_proxy_url?: string
  tg_api_id?: number
  tg_api_hash?: string
  tg_session?: string
  pancheck_base_url?: string
  link_check_enabled?: boolean
  link_check_interval?: number
  tg_channel_chat_id?: string
  tg_sync_enabled?: boolean
  tg_sync_interval?: number
  tg_default_cat_id?: number
  douban_hot_nav_enabled?: boolean
  hot_search_enabled?: boolean
  /** 前台首页是否显示排行榜模块 */
  home_rank_board_enabled?: boolean
  douban_cover_proxy_url?: string
  auto_delete_invalid_links?: boolean
  hide_invalid_links_in_search?: boolean
}

export const getSystemConfig = () => request.get<ICommonResponse<ISystemConfig>>('/system/config')

export const updateSystemConfig = (data: ISystemConfig) =>
  request.put<ICommonResponse<unknown>>('/system/config', data)
