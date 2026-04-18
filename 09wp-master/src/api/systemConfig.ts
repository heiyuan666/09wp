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
  /** 详情页每次点击重新转存并返回新的本人分享链接，不覆盖库内展示链接 */
  resource_detail_each_click_fresh_share?: boolean
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
  /** 豆瓣信息卡搜索缓存 TTL（秒），0 使用默认 SearchTTL */
  douban_search_cache_ttl?: number
  /** 前台搜索页是否启用豆瓣信息卡 */
  douban_search_enabled?: boolean
  /** TMDB v4 Read Access Token（用于前台搜索补充影视信息） */
  tmdb_bearer_token?: string
  /** 前台搜索页是否启用 TMDB 信息卡 */
  tmdb_enabled?: boolean
  /** TMDB 信息卡搜索缓存 TTL（秒），0 使用默认 SearchTTL */
  tmdb_search_cache_ttl?: number
  /** TMDB 请求代理地址（可选），如 http://127.0.0.1:7890 */
  tmdb_proxy_url?: string
  /** 豆瓣聚合接口基地址（默认 https://api.iyuns.com） */
  iyuns_api_base_url?: string
  global_search_enabled?: boolean
  global_search_link_check_enabled?: boolean
  global_search_api_url?: string
  global_search_cloud_types?: string
  global_search_default_category_id?: number
  global_search_auto_transfer?: boolean
  global_search_cleanup_enabled?: boolean
  global_search_cleanup_days?: number
  global_search_cleanup_minutes?: number
  global_search_cleanup_delete_netdisk_files?: boolean
  footer_quick_links?: IFriendLinkItem[]
  footer_hot_platforms?: string[]
  footer_social_links?: IFriendLinkItem[]
  footer_wechat?: string
  auto_delete_invalid_links?: boolean
  hide_invalid_links_in_search?: boolean
  /** 前台搜索全网搜磁力链是否显示「迅雷下载」 */
  thunder_download_enabled?: boolean

  /** 定时清理夸克指定目录内创建时间过久的文件/文件夹（仅当前目录一层） */
  quark_cleanup_enabled?: boolean
  /** 目录 fid，空则尝试用网盘凭证中的夸克转存目录（根目录 0 不会清理） */
  quark_cleanup_folder_id?: string
  /** 早于「现在减该分钟数」的条目将被删除，默认 60 */
  quark_cleanup_older_than_minutes?: number
  /** 任务间隔（分钟），默认 5 */
  quark_cleanup_interval_minutes?: number

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
export const globalSearchTest = (q = '测试', cloudTypes = '') =>
  request.post<ICommonResponse<any>>(
    `/system/config/global-search/test?q=${encodeURIComponent(q)}&cloud_types=${encodeURIComponent(cloudTypes)}`,
  )

export const meiliReindex = (batch_size = 500, target: 'resources' | 'games' = 'resources') =>
  request.post<ICommonResponse<any>>(
    `/system/config/meili/reindex?batch_size=${batch_size}&target=${encodeURIComponent(target)}`,
  )
