/** 后端 Software JSON（与 Go model 对齐） */
export interface ApiSoftware {
  id: number
  category_id: number
  name: string
  summary: string
  version: string
  cover: string
  cover_thumb: string
  icon: string
  icon_thumb: string
  screenshots: string[]
  size: string
  platforms: string
  website: string
  download_direct: string[]
  download_pan: string[]
  download_extract: string
  published_at?: string
  updated_at_override?: string
  status: number
  created_at: string
  updated_at: string
}

export interface ApiSoftwareVersion {
  id: number
  software_id: number
  version: string
  release_notes: string
  published_at?: string
  download_direct: string[]
  download_pan: string[]
  download_extract: string
  created_at: string
  updated_at: string
}

export interface ApiSoftwareCategory {
  id: number
  name: string
  slug: string
  sort_order: number
  status: number
  created_at: string
  updated_at: string
}

/** 前端展示用 */
export interface SoftwareVersionRow {
  id: string
  version: string
  releaseNotes?: string
  publishedAt?: string
  directDownloads: string[]
  cloudDownloads: string[]
  cloudPassword?: string
}

export interface Software {
  id: string
  name: string
  /** 分类 slug，用于链接 /category/:slug */
  category: string
  categoryLabel: string
  description: string
  version: string
  size: string
  status: 'active' | 'deprecated' | 'beta'
  platform: string[]
  releaseDate?: string
  updateDate?: string
  website?: string
  icon: string
  cover?: string
  screenshots?: string[]
  directDownloads?: string[]
  cloudDownloads?: string[]
  cloudPassword?: string
  /** 详情页：各版本下载 */
  versionRows?: SoftwareVersionRow[]
}

export const statusLabels: Record<string, { label: string; color: 'success' | 'warning' | 'error' }> = {
  active: { label: '正常', color: 'success' },
  deprecated: { label: '已停用', color: 'error' },
  beta: { label: '测试版', color: 'warning' },
}

export const platformIcons: Record<string, string> = {
  Windows: 'i-simple-icons-windows',
  macOS: 'i-simple-icons-apple',
  Linux: 'i-simple-icons-linux',
  Android: 'i-simple-icons-android',
  iOS: 'i-simple-icons-ios',
  Web: 'i-lucide-globe',
}

/** 无后端 slug 映射时的默认图标 */
const SLUG_ICONS: Record<string, string> = {
  productivity: 'i-lucide-zap',
  development: 'i-lucide-code-2',
  design: 'i-lucide-palette',
  system: 'i-lucide-settings',
  media: 'i-lucide-play-circle',
  network: 'i-lucide-wifi',
  dev: 'i-lucide-code-2',
  tool: 'i-lucide-wrench',
}

export function iconForCategorySlug(slug: string): string {
  return SLUG_ICONS[slug] || 'i-lucide-folder'
}
