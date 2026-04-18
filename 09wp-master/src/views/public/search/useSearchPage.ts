import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  siteCategories,
  siteDoubanInfoSearch,
  siteGlobalGetLink,
  siteGlobalSearch,
  sitePublicConfig,
  siteResourcePage,
  siteSearch,
  siteTMDBSearch,
  type IDoubanInfoItem,
  type ITMDBItem,
} from '@/api/netdisk'
import { type ICategory, type ISearchResource, type SearchFiltersState } from './searchHelpers'

export type SearchBridge = {
  routeQueryQ: string
  onReplaceSearch: (q: string) => void
  onGoDetail: (id: string | number) => void
}

export function useSearchPage({ routeQueryQ, onReplaceSearch, onGoDetail }: SearchBridge) {
  const [themeMode, setThemeMode] = useState<'light' | 'dark'>(() => {
    const cached = localStorage.getItem('themeMode')
    if (cached === 'light' || cached === 'dark') return cached
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  })
  const [qInput, setQInput] = useState(routeQueryQ)
  const qInputRef = useRef(qInput)
  qInputRef.current = qInput

  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [elapsedMs, setElapsedMs] = useState(0)
  const [list, setList] = useState<ISearchResource[]>([])
  const [categories, setCategories] = useState<ICategory[]>([])
  const [tmdbEnabled, setTmdbEnabled] = useState(false)
  const [tmdbItem, setTmdbItem] = useState<ITMDBItem | null>(null)
  const [doubanEnabled, setDoubanEnabled] = useState(false)
  const [doubanItem, setDoubanItem] = useState<IDoubanInfoItem | null>(null)
  const [globalLoading, setGlobalLoading] = useState(false)
  /** 最近一次全网搜接口耗时（ms），与本地库搜索 elapsedMs 分开统计 */
  const [globalSearchElapsedMs, setGlobalSearchElapsedMs] = useState(0)
  const [globalCloudType, setGlobalCloudType] = useState('')
  /** 后台「全网搜默认网盘类型」，用于前台未选手动筛选时与接口默认一致 */
  const [globalSearchCloudTypesFromServer, setGlobalSearchCloudTypesFromServer] = useState('')
  const [thunderDownloadEnabled, setThunderDownloadEnabled] = useState(false)
  const [globalList, setGlobalList] = useState<
    Array<{
      url: string
      password?: string
      note?: string
      datetime?: string
      source?: string
      cloud_type?: string
      link_status?: 'valid' | 'invalid' | 'pending' | 'unknown'
      images?: string[]
    }>
  >([])

  const [filters, setFilters] = useState<SearchFiltersState>({
    sort: 'relevance',
    categoryId: '',
    platform: '',
    shareTime: '',
    shareYear: '',
    fileType: '',
    exactMode: false,
    dedupMode: false,
  })

  const filtersKey = useMemo(
    () =>
      [
        filters.sort,
        filters.categoryId,
        filters.platform,
        filters.shareTime,
        filters.shareYear,
        filters.fileType,
        String(filters.exactMode),
        String(filters.dedupMode),
      ].join('|'),
    [filters],
  )

  const globalSearchCloudTypesForApi = useMemo(() => {
    const manual = globalCloudType.trim()
    const server = globalSearchCloudTypesFromServer.trim()
    const merged = manual || server
    return merged || undefined
  }, [globalCloudType, globalSearchCloudTypesFromServer])

  /** 下拉框展示值：手动优先；否则用后台默认的首个类型（多类型时仅展示第一项） */
  const globalCloudTypeForSelect = useMemo(() => {
    const manual = globalCloudType.trim()
    if (manual) return manual
    const server = globalSearchCloudTypesFromServer.trim()
    if (!server) return ''
    return server.split(',')[0]?.trim() || ''
  }, [globalCloudType, globalSearchCloudTypesFromServer])

  useEffect(() => {
    setQInput(routeQueryQ)
    setPage(1)
  }, [routeQueryQ])

  useEffect(() => {
    void (async () => {
      const { data: res } = await siteCategories()
      if (res.code === 200 && Array.isArray(res.data)) {
        setCategories(res.data)
      }
    })()
  }, [])

  useEffect(() => {
    localStorage.setItem('themeMode', themeMode)
    document.body.setAttribute('theme-mode', themeMode)
    document.documentElement.setAttribute('theme-mode', themeMode)
  }, [themeMode])

  const fetchList = useCallback(async () => {
    const params: Record<string, string | number> = {
      page,
      page_size: pageSize,
      sort: filters.sort,
    }
    if (filters.categoryId) params.category_id = filters.categoryId
    if (filters.platform) params.platform = filters.platform

    const routeKeyword = String(routeQueryQ || '').trim()
    const keyword = routeKeyword || qInputRef.current.trim()

    setLoading(true)
    const startedAt = performance.now()
    try {
      const { data: res } = keyword ? await siteSearch({ ...params, q: keyword }) : await siteResourcePage(params)
      if (res.code !== 200) {
        setList([])
        setTotal(0)
        return
      }
      setList(res.data?.list || [])
      setTotal(res.data?.total || 0)
      setElapsedMs(Math.max(1, Math.round(performance.now() - startedAt)))
    } finally {
      setLoading(false)
    }
  }, [routeQueryQ, page, pageSize, filters.sort, filters.categoryId, filters.platform])

  useEffect(() => {
    void fetchList()
  }, [routeQueryQ, page, pageSize, filtersKey, fetchList])

  useEffect(() => {
    let cancelled = false
    void (async () => {
      try {
        const { data: res } = await sitePublicConfig()
        if (cancelled || res.code !== 200 || !res.data) return
        setThunderDownloadEnabled(Boolean(res.data.thunder_download_enabled))
        const gs = String((res.data as Record<string, unknown>).global_search_cloud_types || '')
          .split(',')
          .map((x) => x.trim().toLowerCase())
          .filter(Boolean)
          .join(',')
        if (!cancelled) setGlobalSearchCloudTypesFromServer(gs)
      } catch {
        if (!cancelled) setThunderDownloadEnabled(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    const keyword = String(routeQueryQ || '').trim() || qInputRef.current.trim()
    if (!keyword) {
      setGlobalList([])
      setGlobalSearchElapsedMs(0)
      setGlobalLoading(false)
      return
    }
    let cancelled = false
    setGlobalLoading(true)
    setGlobalSearchElapsedMs(0)
    const startedAt = performance.now()
    void (async () => {
      try {
        const { data: res } = await siteGlobalSearch({
          q: keyword,
          ...(globalSearchCloudTypesForApi ? { cloud_types: globalSearchCloudTypesForApi } : {}),
        })
        if (cancelled || res.code !== 200) return
        setGlobalList(Array.isArray(res.data?.list) ? res.data.list.slice(0, 12) : [])
      } catch {
        if (!cancelled) setGlobalList([])
      } finally {
        if (cancelled) return
        setGlobalSearchElapsedMs(Math.max(1, Math.round(performance.now() - startedAt)))
        setGlobalLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [routeQueryQ, globalSearchCloudTypesForApi])

  // 实时效果：有“检测中”项时，短轮询刷新全网搜状态（后端会过滤 invalid）。
  useEffect(() => {
    const keyword = String(routeQueryQ || '').trim() || qInputRef.current.trim()
    if (!keyword || globalLoading || globalList.length === 0) return
    const hasPending = globalList.some((it) => (it.link_status || 'pending') === 'pending')
    if (!hasPending) return
    let cancelled = false
    const timer = window.setInterval(() => {
      void (async () => {
        try {
          const { data: res } = await siteGlobalSearch({
            q: keyword,
            ...(globalSearchCloudTypesForApi ? { cloud_types: globalSearchCloudTypesForApi } : {}),
          })
          if (cancelled || res.code !== 200) return
          setGlobalList(Array.isArray(res.data?.list) ? res.data.list.slice(0, 12) : [])
        } catch {
          // 静默失败，避免打断当前体验
        }
      })()
    }, 3500)
    return () => {
      cancelled = true
      window.clearInterval(timer)
    }
  }, [routeQueryQ, globalSearchCloudTypesForApi, globalLoading, globalList])

  useEffect(() => {
    const keyword = String(routeQueryQ || '').trim() || qInputRef.current.trim()
    if (!keyword) {
      setTmdbItem(null)
      setDoubanItem(null)
      return
    }
    let cancelled = false
    void (async () => {
      try {
        const { data: res } = await siteTMDBSearch({ q: keyword })
        if (cancelled || res.code !== 200) return
        setTmdbEnabled(Boolean(res.data?.enabled))
        setTmdbItem(res.data?.item || null)
      } catch {
        if (!cancelled) {
          setTmdbItem(null)
        }
      }
    })()
    void (async () => {
      try {
        const { data: res } = await siteDoubanInfoSearch({ q: keyword })
        if (cancelled || res.code !== 200) return
        setDoubanEnabled(Boolean(res.data?.enabled))
        setDoubanItem(res.data?.item || null)
      } catch {
        if (!cancelled) {
          setDoubanItem(null)
        }
      }
    })()
    return () => {
      cancelled = true
    }
  }, [routeQueryQ])

  const onSearch = useCallback(async () => {
    setPage(1)
    const keyword = qInputRef.current.trim()
    const routeKeyword = String(routeQueryQ || '').trim()
    if (routeKeyword === keyword) {
      await fetchList()
      return
    }
    onReplaceSearch(keyword)
  }, [routeQueryQ, fetchList, onReplaceSearch])

  const resetFilters = useCallback(() => {
    setFilters({
      sort: 'relevance',
      categoryId: '',
      platform: '',
      shareTime: '',
      shareYear: '',
      fileType: '',
      exactMode: false,
      dedupMode: false,
    })
    setPage(1)
  }, [])

  const setFilter = useCallback(<K extends keyof SearchFiltersState>(key: K, value: SearchFiltersState[K]) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
    setPage(1)
  }, [])

  const toggleTheme = useCallback(() => {
    setThemeMode((prev) => (prev === 'light' ? 'dark' : 'light'))
  }, [])

  const claimGlobalLink = useCallback(
    async (item: {
      url: string
      password?: string
      note?: string
      source?: string
      cloud_type?: string
      images?: string[]
    }) => {
      const { data: res } = await siteGlobalGetLink({
        url: item.url,
        password: item.password || '',
      })
      if (res.code !== 200) {
        const errMsg = String(res.message || '').trim() || '获取链接失败'
        throw new Error(errMsg)
      }
      if (!res.data?.link) return null
      return {
        link: String(res.data.link),
        platform: String(res.data.platform || ''),
        message: String(res.data.message || ''),
        status: String(res.data.status || ''),
        ownShareSource: String(res.data.own_share_source || ''),
        fallbackReason: String(res.data.fallback_reason || ''),
      }
    },
    [],
  )

  return {
    themeMode,
    toggleTheme,
    qInput,
    setQInput,
    page,
    setPage,
    pageSize,
    setPageSize,
    total,
    loading,
    elapsedMs,
    tmdbEnabled,
    tmdbItem,
    doubanEnabled,
    doubanItem,
    list,
    globalLoading,
    globalSearchElapsedMs,
    globalCloudType,
    globalCloudTypeForSelect,
    setGlobalCloudType,
    globalList,
    thunderDownloadEnabled,
    categories,
    filters,
    setFilter,
    onSearch,
    resetFilters,
    onGoDetail,
    claimGlobalLink,
  }
}
