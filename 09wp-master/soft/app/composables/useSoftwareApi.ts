import type { ApiSoftware, ApiSoftwareCategory, ApiSoftwareVersion } from '../types/software'
import { apiFetch } from '../utils/apiClient'
import { buildCategoryMap, mapApiSoftware, mapApiVersion } from '../utils/mapSoftware'
import type { Software } from '../types/software'

export function useSoftwareApi() {
  const config = useRuntimeConfig()
  const apiBase = () => String(config.public.apiBase || '').replace(/\/$/, '')

  async function fetchCategories() {
    return await apiFetch<{ list: ApiSoftwareCategory[] }>('/software/categories')
  }

  async function fetchSoftwarePage(params: {
    page?: number
    page_size?: number
    keyword?: string
    category_id?: number
  }) {
    const q: Record<string, string | number | undefined> = {
      page: params.page ?? 1,
      page_size: params.page_size ?? 12,
    }
    if (params.keyword?.trim()) q.keyword = params.keyword.trim()
    if (params.category_id != null) q.category_id = params.category_id
    return await apiFetch<{ list: ApiSoftware[]; total: number }>('/software/list', { query: q })
  }

  async function fetchSoftwareDetailMapped(id: string): Promise<Software> {
    const base = apiBase()
    const [cats, detail] = await Promise.all([
      apiFetch<{ list: ApiSoftwareCategory[] }>('/software/categories'),
      apiFetch<{ software: ApiSoftware; versions: ApiSoftwareVersion[] }>(`/software/detail/${id}`),
    ])
    const catMap = buildCategoryMap(cats.list)
    const cat = catMap.get(detail.software.category_id)
    const software = mapApiSoftware(detail.software, base, cat)
    const versionRows = (detail.versions || []).map(mapApiVersion)
    if (versionRows.length) software.versionRows = versionRows
    return software
  }

  return {
    apiBase,
    fetchCategories,
    fetchSoftwarePage,
    fetchSoftwareDetailMapped,
  }
}
