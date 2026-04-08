import { apiGet, getApiBaseUrl, toAbsoluteUrl } from "@/lib/api/client"

export type GameResourceDTO = {
  id: number
  game_id: number
  title: string
  resource_type: string
  version: string
  size: string
  download_type: string
  pan_type: string
  download_url: string
  download_urls: string[]
  extract_code?: string
  tested: boolean
  author: string
  publish_date?: string
  created_at: string
  updated_at: string
}

export type GameDTO = {
  id: number
  category_id?: number
  steam_appid: number
  required_age?: number
  is_free?: boolean
  title: string
  cover: string
  banner: string
  video_url: string
  short_description: string
  supported_languages?: string
  reviews?: string
  pc_requirements?: string
  mac_requirements?: string
  linux_requirements?: string
  header_image: string
  website: string
  developers?: string
  publishers: string
  platforms?: string
  genres: string
  tags: string
  price_text: string
  price_currency: string
  price_initial: number
  price_final: number
  price_discount: number
  metacritic_score: number
  description: string
  release_date?: string
  size: string
  type: string
  developer: string
  rating: number
  steam_score: number
  recommendations_total?: number
  downloads: number
  likes: number
  dislikes: number
  gallery: string[]
  created_at: string
  updated_at: string
  resources?: GameResourceDTO[]
}

export async function fetchGameDetail(id: string | number) {
  return await apiGet<GameDTO>(`/game/detail/${id}`, { cache: "force-cache" })
}

export type GameListResult = {
  list: GameDTO[]
  total: number
}

export type GameCategoryDTO = {
  id: number
  name: string
  slug: string
  description: string
  created_at: string
  updated_at: string
}

export async function fetchGameList(params: {
  page?: number
  page_size?: number
  keyword?: string
  category_id?: number
  type?: string
}) {
  const qs = new URLSearchParams()
  if (params.page) qs.set("page", String(params.page))
  if (params.page_size) qs.set("page_size", String(params.page_size))
  if (params.keyword) qs.set("keyword", params.keyword)
  if (typeof params.category_id === "number") qs.set("category_id", String(params.category_id))
  if (params.type) qs.set("type", params.type)
  const query = qs.toString()
  return await apiGet<GameListResult>(`/game/list${query ? `?${query}` : ""}`, { cache: "force-cache" })
}

export async function fetchGameCategoryList() {
  return await apiGet<GameCategoryDTO[]>(`/game/category/list`, { cache: "force-cache" })
}

export function getBackendOriginFromApiBaseUrl() {
  const apiBase = getApiBaseUrl()
  try {
    const u = new URL(apiBase)
    return `${u.protocol}//${u.host}`
  } catch {
    return "http://localhost:8080"
  }
}

export function absolutizeGameMediaUrls(game: GameDTO) {
  const origin = getBackendOriginFromApiBaseUrl()
  return {
    ...game,
    cover: toAbsoluteUrl(game.cover, origin),
    banner: toAbsoluteUrl(game.banner, origin),
    header_image: toAbsoluteUrl(game.header_image, origin),
    video_url: toAbsoluteUrl(game.video_url, origin),
    gallery: Array.isArray(game.gallery) ? game.gallery.map((x) => toAbsoluteUrl(x, origin)) : [],
  }
}

export function splitToList(input: string) {
  const raw = (input ?? "").trim()
  if (!raw) return []
  const parts = raw
    .split(/[\r\n\t,，;；/|]+/g)
    .map((s) => s.trim())
    .filter(Boolean)
  return Array.from(new Set(parts))
}

export type DownloadCardPanType =
  | "quark"
  | "aliyun"
  | "baidu"
  | "lanzou"
  | "123pan"
  | "tianyi"
  | "mega"
  | "onedrive"

export function detectPanTypeByUrl(url: string): DownloadCardPanType | null {
  const u = (url ?? "").toLowerCase()
  if (!u) return null
  if (u.includes("pan.quark.cn")) return "quark"
  if (u.includes("aliyundrive.com") || u.includes("alipan.com")) return "aliyun"
  if (u.includes("pan.baidu.com")) return "baidu"
  if (u.includes("cloud.189.cn")) return "tianyi"
  if (u.includes("123pan.com") || u.includes("123684.com")) return "123pan"
  if (u.includes("mega.nz") || u.includes("mega.io")) return "mega"
  if (u.includes("1drv.ms") || u.includes("onedrive.live.com")) return "onedrive"
  // 后端支持的“蓝奏/UC/115/迅雷”等，这里先降级：走链接探测不到就返回 null
  return null
}

