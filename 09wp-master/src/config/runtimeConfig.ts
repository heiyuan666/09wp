import { APP_CONFIG } from '@/config/app.config'
import { reactive } from 'vue'
import type { IFriendLinkItem } from '@/api/systemConfig'

interface IRuntimeConfig {
  siteTitle: string
  logoUrl: string
  faviconUrl: string
  footerText: string
  seoKeywords: string
  seoDescription: string
  icpRecord: string
  allowRegister: boolean
  clarityProjectId: string
  clarityEnabled: boolean
  supportEmail: string
  contactPhone: string
  friendLinks: IFriendLinkItem[]
  doubanHotNavEnabled: boolean
  hotSearchEnabled: boolean
  homeRankBoardEnabled: boolean
  doubanCoverProxyUrl: string
  /** TG 外链封面返代模板（与后台系统配置一致） */
  tgImageProxyUrl: string
  // 号卡前台按钮（可配置跳转链接）
  haokaOrderUrl?: string
  haokaAgentRegUrl?: string
}

const CACHE_KEY = 'runtime_system_config'

const defaultConfig: IRuntimeConfig = {
  siteTitle: APP_CONFIG.name,
  logoUrl: APP_CONFIG.logoSrc,
  faviconUrl: APP_CONFIG.faviconSrc,
  footerText: '©️零九cdn www.09cdn.com',
  seoKeywords: '',
  seoDescription: '',
  icpRecord: '',
  allowRegister: true,
  clarityProjectId: '',
  clarityEnabled: false,
  supportEmail: '',
  contactPhone: '',
  friendLinks: [],
  doubanHotNavEnabled: false,
  hotSearchEnabled: true,
  homeRankBoardEnabled: true,
  doubanCoverProxyUrl: '',
  tgImageProxyUrl: '',
  haokaOrderUrl: '',
  haokaAgentRegUrl: '',
}

export const runtimeConfig = reactive<IRuntimeConfig>({ ...defaultConfig })

const applyDocumentConfig = () => {
  document.title = runtimeConfig.siteTitle || APP_CONFIG.name

  let faviconLink = document.querySelector("link[rel~='icon']") as HTMLLinkElement
  if (!faviconLink) {
    faviconLink = document.createElement('link')
    faviconLink.rel = 'icon'
    document.head.appendChild(faviconLink)
  }
  faviconLink.href = runtimeConfig.faviconUrl || APP_CONFIG.faviconSrc

  const setMeta = (name: string, content: string) => {
    let meta = document.querySelector(`meta[name='${name}']`) as HTMLMetaElement
    if (!meta) {
      meta = document.createElement('meta')
      meta.name = name
      document.head.appendChild(meta)
    }
    meta.content = content
  }
  if (runtimeConfig.seoKeywords) setMeta('keywords', runtimeConfig.seoKeywords)
  if (runtimeConfig.seoDescription) setMeta('description', runtimeConfig.seoDescription)
}

const applyClarityConfig = () => {
  const scriptId = 'microsoft-clarity-script'
  const existing = document.getElementById(scriptId) as HTMLScriptElement | null
  const projectId = String(runtimeConfig.clarityProjectId || '').trim()

  if (!runtimeConfig.clarityEnabled || !projectId) {
    existing?.remove()
    return
  }

  if (existing?.dataset.projectId === projectId) return
  existing?.remove()

  const win = window as Window & {
    clarity?: (...args: any[]) => void
    [key: string]: any
  }

  win.clarity =
    win.clarity ||
    function (...args: any[]) {
      ;(win.clarity as any).q = (win.clarity as any).q || []
      ;(win.clarity as any).q.push(args)
    }

  const script = document.createElement('script')
  script.id = scriptId
  script.async = true
  script.dataset.projectId = projectId
  script.src = `https://www.clarity.ms/tag/${encodeURIComponent(projectId)}`
  document.head.appendChild(script)
}

export const loadRuntimeConfig = async () => {
  // 先读本地缓存，避免首屏闪烁
  const cache = localStorage.getItem(CACHE_KEY)
  if (cache) {
    try {
      Object.assign(runtimeConfig, { ...defaultConfig, ...JSON.parse(cache) })
      applyDocumentConfig()
      applyClarityConfig()
    } catch {
      // ignore invalid cache
    }
  } else {
    applyDocumentConfig()
    applyClarityConfig()
  }

  // 再拉取后端最新配置
  try {
    const baseURL = (import.meta.env.VITE_API_BASE_URL || '').replace(/\/+$/, '')
    const url = `${baseURL}/public/config`
    const resp = await fetch(url)
    const res = await resp.json()
    if (res?.code === 200 && res?.data) {
      const data = res.data
      Object.assign(runtimeConfig, {
        siteTitle: data.site_title || APP_CONFIG.name,
        logoUrl: data.logo_url || APP_CONFIG.logoSrc,
        faviconUrl: data.favicon_url || APP_CONFIG.faviconSrc,
        footerText: data.footer_text || defaultConfig.footerText,
        seoKeywords: data.seo_keywords || '',
        seoDescription: data.seo_description || '',
        icpRecord: data.icp_record || '',
        allowRegister: data.allow_register ?? true,
        clarityProjectId: data.clarity_project_id || '',
        clarityEnabled: data.clarity_enabled ?? false,
        supportEmail: data.support_email || '',
        contactPhone: data.contact_phone || '',
        friendLinks: Array.isArray(data.friend_links) ? data.friend_links : [],
        doubanHotNavEnabled: data.douban_hot_nav_enabled ?? false,
        hotSearchEnabled: data.hot_search_enabled ?? true,
        homeRankBoardEnabled: data.home_rank_board_enabled ?? true,
        doubanCoverProxyUrl: data.douban_cover_proxy_url || '',
        tgImageProxyUrl: data.tg_image_proxy_url || '',
        haokaOrderUrl: data.haoka_order_url || '',
        haokaAgentRegUrl: data.haoka_agent_reg_url || '',
      })
      localStorage.setItem(CACHE_KEY, JSON.stringify(runtimeConfig))
      applyDocumentConfig()
      applyClarityConfig()
    }
  } catch {
    // 网络失败使用本地缓存/默认配置
  }
}
