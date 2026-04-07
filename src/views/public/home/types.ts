export type PublicConfig = {
  site_title: string
  support_email: string
  contact_phone: string
  contact_qq: string
  logo_url: string
  favicon_url: string
  seo_keywords: string
  seo_description: string
  icp_record: string
  footer_text: string
  allow_register: boolean
  hot_search_enabled?: boolean
  show_site_title?: boolean
  /** 首页是否展示排行榜（热门资源 / 最新资源 / 豆瓣热门） */
  home_rank_board_enabled?: boolean
  friend_links?: Array<{ title: string; url: string }>
}

export type NavMenuItem = {
  id: number
  title: string
  path: string
}

export type HotSearchItem = { keyword: string; search_count: number }

export type HomeResourceItem = {
  id: number
  title: string
  cover?: string
  view_count?: number
}

export type DoubanHotItem = {
  title: string
  cover?: string
  url?: string
}

export const siteNameFallback = '懒盘搜索'

export const defaultTopNav: NavMenuItem[] = [
  { id: 1, title: '聚合搜索', path: '#' },
  { id: 2, title: '提交资源', path: '#' },
  { id: 3, title: '侵权屏蔽', path: '#' },
  { id: 4, title: '本站搭建', path: '#' },
]

export const defaultPromos: NavMenuItem[] = [
  { id: 1, title: '本站服务器', path: '#' },
  { id: 2, title: '淘宝隐藏优惠券', path: '#' },
  { id: 3, title: '(0905)本站 APP 下载', path: '#' },
  { id: 4, title: '临时域名(可保存书签，比较快)', path: '#' },
]

export const defaultConfig: PublicConfig = {
  site_title: '网盘资源导航系统',
  support_email: 'support@example.com',
  contact_phone: '',
  contact_qq: '',
  logo_url: '',
  favicon_url: '',
  seo_keywords: '网盘,资源,导航',
  seo_description: '网盘资源导航管理系统',
  icp_record: '',
  footer_text: '©09cdn www.09cdn.com',
  allow_register: true,
  hot_search_enabled: true,
  show_site_title: true,
  home_rank_board_enabled: true,
}
