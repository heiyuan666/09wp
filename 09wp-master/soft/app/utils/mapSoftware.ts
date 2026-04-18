import type {
  ApiSoftware,
  ApiSoftwareCategory,
  ApiSoftwareVersion,
  Software,
  SoftwareVersionRow,
} from '../types/software'
import { resolveAssetUrl } from './assetUrl'
import { sortNetdiskLinks } from './netdiskLabel'

function parsePlatforms(s: string): string[] {
  if (!s?.trim()) return []
  return s
    .split(/[,，;；]/)
    .map(x => x.trim())
    .filter(Boolean)
}

function mapStatus(st: number): Software['status'] {
  if (st === 1) return 'active'
  return 'deprecated'
}

function fmtDate(iso?: string | null): string | undefined {
  if (!iso) return undefined
  try {
    return new Date(iso).toLocaleDateString('zh-CN')
  } catch {
    return iso
  }
}

export function mapApiSoftware(
  raw: ApiSoftware,
  apiBase: string,
  category?: ApiSoftwareCategory | null,
): Software {
  const icon = resolveAssetUrl(raw.icon_thumb || raw.icon, apiBase)
  const coverPath = raw.cover_thumb || raw.cover
  const cover = coverPath ? resolveAssetUrl(coverPath, apiBase) : undefined
  const screenshots = (raw.screenshots || []).map(s => resolveAssetUrl(s, apiBase))

  return {
    id: String(raw.id),
    name: raw.name,
    category: category?.slug || '',
    categoryLabel: category?.name || '未分类',
    description: raw.summary || '',
    version: raw.version || '',
    size: raw.size || '-',
    status: mapStatus(raw.status),
    platform: parsePlatforms(raw.platforms),
    releaseDate: fmtDate(raw.published_at),
    updateDate: fmtDate(raw.updated_at_override) || fmtDate(raw.updated_at),
    website: raw.website || undefined,
    icon,
    cover,
    screenshots: screenshots.length ? screenshots : undefined,
    directDownloads: raw.download_direct?.length ? [...raw.download_direct] : undefined,
    cloudDownloads: raw.download_pan?.length ? sortNetdiskLinks(raw.download_pan) : undefined,
    cloudPassword: raw.download_extract || undefined,
  }
}

export function mapApiVersion(v: ApiSoftwareVersion): SoftwareVersionRow {
  return {
    id: String(v.id),
    version: v.version,
    releaseNotes: v.release_notes || undefined,
    publishedAt: fmtDate(v.published_at),
    directDownloads: v.download_direct?.length ? [...v.download_direct] : [],
    cloudDownloads: sortNetdiskLinks(v.download_pan),
    cloudPassword: v.download_extract || undefined,
  }
}

export function buildCategoryMap(list: ApiSoftwareCategory[]): Map<number, ApiSoftwareCategory> {
  const m = new Map<number, ApiSoftwareCategory>()
  for (const c of list) m.set(c.id, c)
  return m
}

export function mapApiSoftwareList(
  rawList: ApiSoftware[],
  apiBase: string,
  categories: ApiSoftwareCategory[],
): Software[] {
  const catMap = buildCategoryMap(categories)
  return rawList.map(s => mapApiSoftware(s, apiBase, catMap.get(s.category_id)))
}
