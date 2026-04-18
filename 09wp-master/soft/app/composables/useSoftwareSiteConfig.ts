import { apiFetch } from '../utils/apiClient'
import { resolveAssetUrl } from '../utils/assetUrl'

export interface SoftwarePublicSiteConfig {
  site_title: string
  logo_url: string
  favicon_url: string
  seo_keywords: string
  seo_description: string
}

/** 全站共享：软件库前台站点配置（与 app.vue、各页 SEO 共用缓存） */
export function useSoftwareSiteConfig() {
  const config = useRuntimeConfig()
  const apiBase = () => String(config.public.apiBase || '').replace(/\/$/, '')

  const asyncData = useAsyncData('software-public-config', async () => {
    try {
      return await apiFetch<SoftwarePublicSiteConfig>('/software/public/config')
    } catch {
      return null
    }
  })

  const resolvedLogo = computed(() => {
    const u = asyncData.data.value?.logo_url
    return u ? resolveAssetUrl(u, apiBase()) : ''
  })

  const resolvedFavicon = computed(() => {
    const u = asyncData.data.value?.favicon_url
    return u ? resolveAssetUrl(u, apiBase()) : ''
  })

  const siteTitle = computed(() => asyncData.data.value?.site_title?.trim() || '软件库')

  // 勿展开 ...asyncData，避免与 Nuxt 内部字段冲突导致 resolvedLogo 等丢失
  return {
    data: asyncData.data,
    pending: asyncData.pending,
    error: asyncData.error,
    refresh: asyncData.refresh,
    status: asyncData.status,
    execute: asyncData.execute,
    clear: asyncData.clear,
    apiBase,
    resolvedLogo,
    resolvedFavicon,
    siteTitle,
  }
}
