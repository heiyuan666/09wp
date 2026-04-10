import { apiGet } from "@/lib/api/client"

export type SoftwareItem = {
  id: number
  category_id: number
  name: string
  summary: string
  version: string
  cover: string
  cover_thumb: string
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

export type SoftwareVersionItem = {
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

export type SoftwareListResult = {
  list: SoftwareItem[]
  total: number
}

export async function fetchSoftwareList(params?: {
  page?: number
  page_size?: number
  keyword?: string
  category_id?: number
  version?: string
}) {
  const qs = new URLSearchParams()
  if (params?.page) qs.set("page", String(params.page))
  if (params?.page_size) qs.set("page_size", String(params.page_size))
  if (params?.keyword) qs.set("keyword", params.keyword)
  if (typeof params?.category_id === "number") qs.set("category_id", String(params.category_id))
  if (params?.version) qs.set("version", params.version)
  const query = qs.toString()
  return await apiGet<SoftwareListResult>(`/software/list${query ? `?${query}` : ""}`, { cache: "force-cache" })
}

export async function fetchSoftwareDetail(id: string | number) {
  return await apiGet<{ software: SoftwareItem; versions: SoftwareVersionItem[] }>(`/software/detail/${id}`, { cache: "force-cache" })
}
