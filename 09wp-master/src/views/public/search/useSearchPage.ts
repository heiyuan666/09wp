import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { siteCategories, siteResourcePage, siteSearch, siteTMDBSearch, type ITMDBItem } from '@/api/netdisk'
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
    const keyword = String(routeQueryQ || '').trim() || qInputRef.current.trim()
    if (!keyword) {
      setTmdbItem(null)
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
    list,
    categories,
    filters,
    setFilter,
    onSearch,
    resetFilters,
    onGoDetail,
  }
}
