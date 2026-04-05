import dayjs from 'dayjs'

export { CloudIcon } from '../home/HomeIcons'

export interface ICategory {
  id: number | string
  name: string
}

export interface IFileItem {
  name: string
  size?: string
  type?: string
}

export interface ISearchResource {
  id: number | string
  title: string
  description?: string
  category_id: number | string
  link: string
  extract_code?: string
  link_valid?: boolean
  tags?: string
  view_count?: number
  created_at?: string
  files?: IFileItem[]
}

export type PlatformValue =
  | ''
  | 'baidu'
  | 'aliyun'
  | 'quark'
  | 'xunlei'
  | 'uc'
  | 'tianyi'
  | 'yidong'
  | 'pan115'
  | 'pan123'
  | 'other'

export type SearchFiltersState = {
  sort: 'relevance' | 'latest' | 'hot'
  categoryId: string
  platform: PlatformValue
  shareTime: string
  shareYear: string
  fileType: string
  exactMode: boolean
  dedupMode: boolean
}

export const platformOptions: Array<{ label: string; value: PlatformValue }> = [
  { label: '所有网盘', value: '' },
  { label: '百度网盘', value: 'baidu' },
  { label: '阿里云盘', value: 'aliyun' },
  { label: '夸克网盘', value: 'quark' },
  { label: '迅雷网盘', value: 'xunlei' },
  { label: 'UC 网盘', value: 'uc' },
  { label: '天翼云盘', value: 'tianyi' },
  { label: '移动云盘', value: 'yidong' },
  { label: '115 网盘', value: 'pan115' },
  { label: '123 网盘', value: 'pan123' },
  { label: '其他网盘', value: 'other' },
]

export function driveIconSrc(link: string) {
  const text = String(link || '').toLowerCase()
  if (text.includes('pan.baidu.com')) return '/baidu.png'
  if (text.includes('aliyundrive.com') || text.includes('alipan.com')) return '/al.png'
  if (text.includes('pan.quark.cn')) return '/quark.png'
  if (text.includes('pan.xunlei.com')) return '/xunlei.png'
  if (
    text.includes('drive-h.uc.cn') ||
    text.includes('drive.uc.cn') ||
    text.includes('yun.uc.cn') ||
    text.includes('uc.cn')
  ) {
    return '/uc.png'
  }
  if (text.includes('cloud.189.cn') || text.includes('caiyun.189') || text.includes('tianyi')) return '/tainyi.png'
  if (text.includes('yun.139.com') || text.includes('caiyun.139.com')) return '/yidong.png'
  if (text.includes('115.com') || text.includes('115cdn.com')) return '/kuaitu.png'
  if (
    text.includes('123pan') ||
    text.includes('123684') ||
    text.includes('123685') ||
    text.includes('123912') ||
    text.includes('123592') ||
    text.includes('123865') ||
    text.includes('123.net')
  ) {
    return '/lanzou.png'
  }
  return '/uc.png'
}

export function categoryText(categories: ICategory[], id: number | string) {
  return categories.find((item) => String(item.id) === String(id))?.name || '未分类'
}

/** 用于匹配「用户点的网盘」与转存后刷新得到的链接（按平台而非完整 URL）。 */
export type NetdiskPlatformKey =
  | 'quark'
  | 'baidu'
  | 'aliyun'
  | 'xunlei'
  | 'uc'
  | 'tianyi'
  | 'yidong'
  | 'pan115'
  | 'pan123'
  | 'other'

export function netdiskPlatformKey(link: string): NetdiskPlatformKey {
  const t = String(link || '').toLowerCase()
  if (t.includes('pan.quark.cn')) return 'quark'
  if (t.includes('pan.baidu.com')) return 'baidu'
  if (t.includes('aliyundrive.com') || t.includes('alipan.com')) return 'aliyun'
  if (t.includes('pan.xunlei.com')) return 'xunlei'
  if (t.includes('drive-h.uc.cn') || t.includes('drive.uc.cn') || t.includes('yun.uc.cn')) return 'uc'
  if (t.includes('cloud.189.cn') || t.includes('caiyun.189') || t.includes('tianyi')) return 'tianyi'
  if (t.includes('yun.139.com') || t.includes('caiyun.139.com')) return 'yidong'
  if (t.includes('115.com') || t.includes('115cdn.com')) return 'pan115'
  if (
    t.includes('123pan') ||
    t.includes('123684') ||
    t.includes('123685') ||
    t.includes('123912') ||
    t.includes('123592') ||
    t.includes('123865') ||
    t.includes('123.net')
  ) {
    return 'pan123'
  }
  return 'other'
}

/** 详情页多链接展示顺序：数值越小越靠前（夸克优先于百度，与常见资源站习惯一致）。 */
export function netdiskLinkDisplayRank(link: string): number {
  const t = String(link || '').toLowerCase()
  if (t.includes('pan.quark.cn')) return 1
  if (t.includes('pan.baidu.com')) return 2
  if (t.includes('aliyundrive.com') || t.includes('alipan.com')) return 3
  if (t.includes('pan.xunlei.com')) return 4
  if (t.includes('drive-h.uc.cn') || t.includes('drive.uc.cn') || t.includes('yun.uc.cn')) return 5
  if (t.includes('cloud.189.cn') || t.includes('caiyun.189') || t.includes('tianyi')) return 6
  if (t.includes('yun.139.com') || t.includes('caiyun.139.com')) return 7
  if (t.includes('115.com') || t.includes('115cdn.com')) return 8
  if (
    t.includes('123pan') ||
    t.includes('123684') ||
    t.includes('123685') ||
    t.includes('123912') ||
    t.includes('123592') ||
    t.includes('123865') ||
    t.includes('123.net')
  ) {
    return 9
  }
  return 100
}

export function platformText(link: string) {
  const text = String(link || '').toLowerCase()
  if (text.includes('pan.baidu.com')) return '百度网盘'
  if (text.includes('aliyundrive.com') || text.includes('alipan.com')) return '阿里云盘'
  if (text.includes('pan.quark.cn')) return '夸克网盘'
  if (text.includes('pan.xunlei.com')) return '迅雷网盘'
  if (text.includes('drive-h.uc.cn') || text.includes('drive.uc.cn') || text.includes('yun.uc.cn')) return 'UC 网盘'
  if (text.includes('cloud.189.cn') || text.includes('caiyun.189') || text.includes('tianyi')) return '天翼云盘'
  if (text.includes('yun.139.com') || text.includes('caiyun.139.com')) return '移动云盘'
  if (text.includes('115.com') || text.includes('115cdn.com')) return '115 网盘'
  if (
    text.includes('123pan') ||
    text.includes('123684') ||
    text.includes('123685') ||
    text.includes('123912') ||
    text.includes('123592') ||
    text.includes('123865') ||
    text.includes('123.net')
  ) {
    return '123 网盘'
  }
  return '其他网盘'
}

export function tagsOf(raw?: string) {
  return String(raw || '')
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
}

export function extractFiles(item: ISearchResource): IFileItem[] {
  if (Array.isArray(item.files) && item.files.length > 0) {
    return item.files.slice(0, 6)
  }

  const lines = String(item.description || '')
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)

  const files = lines
    .map((line) => line.replace(/^file:\s*/i, '').trim())
    .filter((line) => /\.[a-z0-9]{2,6}$/i.test(line))
    .slice(0, 6)

  return files.map((name) => ({ name }))
}

export function formatDate(val?: string) {
  return val ? dayjs(val).format('YYYY-MM-DD') : '-'
}
