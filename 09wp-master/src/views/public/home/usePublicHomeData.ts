import { useEffect, useMemo, useState } from 'react'
import { API_BASE_URL } from '@/config/app.config'
import {
  defaultConfig,
  siteNameFallback,
  type DoubanHotItem,
  type HomeHotCategoryItem,
  type HomeResourceItem,
  type HotSearchItem,
  type NavMenuItem,
  type PublicConfig,
} from './types'

export function usePublicHomeData() {
  const [keyword, setKeyword] = useState('')
  const [config, setConfig] = useState<PublicConfig>(defaultConfig)
  const [topNav, setTopNav] = useState<NavMenuItem[]>([])
  const [homePromos, setHomePromos] = useState<NavMenuItem[]>([])
  const [hotSearches, setHotSearches] = useState<HotSearchItem[]>([])
  const [hotResources, setHotResources] = useState<HomeResourceItem[]>([])
  const [hotByCategory, setHotByCategory] = useState<HomeHotCategoryItem[]>([])
  const [latestResources, setLatestResources] = useState<HomeResourceItem[]>([])
  const [doubanHot, setDoubanHot] = useState<DoubanHotItem[]>([])
  const year = useMemo(() => new Date().getFullYear(), [])
  const friendLinks = Array.isArray(config.friend_links) ? config.friend_links : []
  const siteTitle = (config.site_title || '').trim() || siteNameFallback
  const siteDescription = (config.seo_description || '').trim() || siteTitle || defaultConfig.seo_description
  const [themeMode, setThemeMode] = useState<'light' | 'dark'>(() => {
    const cached = localStorage.getItem('themeMode')
    if (cached === 'light' || cached === 'dark') return cached
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  })

  useEffect(() => {
    localStorage.setItem('themeMode', themeMode)
    document.body.setAttribute('theme-mode', themeMode)
    document.documentElement.setAttribute('theme-mode', themeMode)
  }, [themeMode])

  useEffect(() => {
    const controller = new AbortController()
    const base = API_BASE_URL

    fetch(`${base}/public/config`, { signal: controller.signal })
      .then(async (res) => res.json())
      .then((res) => {
        if (res?.code !== 200 || !res?.data) return
        setConfig((prev) => ({ ...prev, ...res.data }))
        const hotSearchOn = res.data.hot_search_enabled !== false
        const rankOn = res.data.home_rank_board_enabled !== false

        if (hotSearchOn || rankOn) {
          fetch(`${base}/home`, { signal: controller.signal })
            .then(async (r) => r.json())
            .then((homeRes) => {
              if (homeRes?.code !== 200 || !homeRes?.data) return
              const d = homeRes.data
              if (hotSearchOn && Array.isArray(d.hot_searches)) {
                setHotSearches(d.hot_searches)
              }
              if (rankOn) {
                if (Array.isArray(d.hot)) setHotResources(d.hot)
                if (Array.isArray(d.hot_by_category)) setHotByCategory(d.hot_by_category)
                if (Array.isArray(d.latest)) setLatestResources(d.latest)
              }
            })
            .catch(() => {})
        }

        if (rankOn) {
          fetch(`${base}/public/douban-hot?limit=12`, { signal: controller.signal })
            .then((r) => r.json())
            .then((dRes) => {
              const list = dRes?.data?.list
              if (dRes?.code === 200 && Array.isArray(list)) setDoubanHot(list)
            })
            .catch(() => {})
        }
      })
      .catch(() => {})

    fetch(`${base}/public/nav-menus?position=top_nav`, { signal: controller.signal })
      .then((res) => res.json())
      .then((res) => {
        const list = res?.data?.list
        if (res?.code === 200 && Array.isArray(list)) {
          setTopNav(
            list.map((x: { id: number; title: string; path?: string }) => ({
              id: x.id,
              title: x.title,
              path: x.path || '#',
            })),
          )
        }
      })
      .catch(() => {})

    fetch(`${base}/public/nav-menus?position=home_promo`, { signal: controller.signal })
      .then((res) => res.json())
      .then((res) => {
        const list = res?.data?.list
        if (res?.code === 200 && Array.isArray(list)) {
          setHomePromos(
            list.map((x: { id: number; title: string; path?: string }) => ({
              id: x.id,
              title: x.title,
              path: x.path || '#',
            })),
          )
        }
      })
      .catch(() => {})

    return () => controller.abort()
  }, [])

  useEffect(() => {
    document.title = siteTitle
  }, [siteTitle])

  const search = () => {
    const q = keyword.trim()
    if (!q) return
    window.location.href = `/search?q=${encodeURIComponent(q)}`
  }

  const openPromo = (item: NavMenuItem) => {
    if (!item.path) return
    if (item.path.startsWith('http')) {
      window.open(item.path, '_blank', 'noopener')
      return
    }
    window.location.href = item.path
  }

  const openHotKeyword = (kw: string) => {
    window.location.href = `/search?q=${encodeURIComponent(kw)}`
  }

  const toggleTheme = () => {
    setThemeMode((prev) => (prev === 'light' ? 'dark' : 'light'))
  }

  return {
    keyword,
    setKeyword,
    config,
    topNav,
    homePromos,
    hotSearches,
    hotResources,
    hotByCategory,
    latestResources,
    doubanHot,
    friendLinks,
    year,
    siteTitle,
    siteDescription,
    search,
    openPromo,
    openHotKeyword,
    toggleTheme,
  }
}
