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
  show_site_title?: boolean
  /** 前台首页是否显示排行榜模块 */
  home_rank_board_enabled?: boolean
  douban_cover_proxy_url?: string
  /** TG 等资源外链封面返代，如 https://wsrv.nl/?url= */
  tg_image_proxy_url?: string
  /** TMDB v4 Read Access Token（用于前台搜索补充影视信息） */
  tmdb_bearer_token?: string
  /** TMDB 请求代理地址（可选），如 http://127.0.0.1:7890 */
  tmdb_proxy_url?: string
  /** 豆瓣聚合接口基地址（默认 https://api.iyuns.com） */
  iyuns_api_base_url?: string
  footer_quick_links?: IFriendLinkItem[]
  footer_hot_platforms?: string[]
  footer_social_links?: IFriendLinkItem[]
  footer_wechat?: string
  auto_delete_invalid_links?: boolean
  hide_invalid_links_in_search?: boolean

  /** Meilisearch：开启后搜索优先走 Meili（失败回退 MySQL） */
  meili_enabled?: boolean
  meili_url?: string
  meili_api_key?: string
  meili_index?: string
}

export const getSystemConfig = () => request.get<ICommonResponse<ISystemConfig>>('/system/config')

export const updateSystemConfig = (data: ISystemConfig) =>
  request.put<ICommonResponse<unknown>>('/system/config', data)

export const meiliTest = () => request.post<ICommonResponse<any>>('/system/config/meili/test')

export const meiliReindex = (batch_size = 500) =>
  request.post<ICommonResponse<any>>(`/system/config/meili/reindex?batch_size=${batch_size}`)
