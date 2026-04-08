import { useEffect, useMemo, useState } from 'react'
import { Play, Search, ShoppingCart, User, Menu as MenuIcon, X, Gamepad2, ChevronLeft, ChevronRight, ArrowRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { publicGameCategoryList, publicGameList, publicNavMenuList, type PublicGameItem } from './api'

import './GameHomeStoreReact.scss'

type CategoryOption = { id: number; name: string }

type CardGame = {
  id: number
  title: string
  cover: string
  category: string
  subType: string
  releaseDateText: string
  updateText: string
  downloads: number
  description: string
  developer: string
  publisher: string
  score: number
  tags: string[]
}

const PAGE_SIZE = 10
const RECOMMEND_PAGE_SIZE = 50
const HERO_SLIDE_COUNT = 5
const HERO_AUTOPLAY_MS = 5000

const fallbackCover = (id: number) => `https://picsum.photos/seed/fallback${id}/480/260`

const toDateText = (value?: string) => {
  if (!value) return '-'
  const t = new Date(value)
  if (Number.isNaN(t.getTime())) return String(value)
  return t.toISOString().slice(0, 10)
}

const toCardItem = (item: PublicGameItem, categoryName: string): CardGame => {
  const cover =
    String(item.cover || item.banner || item.header_image || '').trim() ||
    fallbackCover(Number(item.id || 0))

  const ratingRaw = Number(item.rating || 0)
  const steamScoreRaw = Number(item.steam_score || 0)
  const score = ratingRaw > 0 ? ratingRaw : Number((steamScoreRaw / 10).toFixed(1))

  const createdAt = String(item.created_at || item.updated_at || '')

  const tags =
    typeof item.tags === 'string'
      ? String(item.tags)
          .split(/[，,]/)
          .map((x) => x.trim())
          .filter(Boolean)
          .slice(0, 6)
      : []

  return {
    id: Number(item.id || 0),
    title: String(item.title || '未命名游戏'),
    cover,
    category: categoryName || String(item.type || '未分类'),
    subType: String(item.type || '单机'),
    releaseDateText: toDateText(String(item.release_date || createdAt)),
    updateText: createdAt ? `${toDateText(createdAt)} 更新` : '最近更新',
    downloads: Number(item.downloads || 0),
    description: String(item.short_description || item.description || '').trim(),
    developer: String(item.developer || '').trim(),
    publisher: String(item.publishers || '').trim(),
    score: score > 0 ? score : 7.5,
    tags,
  }
}

const readKeywordFromURL = () => {
  if (typeof window === 'undefined') return ''
  return new URLSearchParams(window.location.search).get('keyword')?.trim() || ''
}

const syncKeywordToURL = (keyword: string) => {
  if (typeof window === 'undefined') return
  const url = new URL(window.location.href)
  const kw = keyword.trim()
  if (kw) url.searchParams.set('keyword', kw)
  else url.searchParams.delete('keyword')
  window.history.replaceState({}, '', url.toString())
}

const GameStoreHeader = ({
  keyword,
  onKeywordChange,
  onSearch,
}: {
  keyword: string
  onKeywordChange: (v: string) => void
  onSearch: () => void
}) => {
  const [isMenuOpen, setIsMenuOpen] = useState(false)

  const navItems = ['商店', '社区', '关于', '支持']

  return (
    <header className="gamestore-header">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="flex h-16 items-center justify-between">
          <div className="flex items-center gap-2">
            <Gamepad2 className="h-8 w-8 text-primary" />
            <span className="text-xl font-bold text-foreground">GameStore</span>
          </div>

          <nav className="hidden md:flex items-center gap-8">
            {navItems.map((item) => (
              <a key={item} href="#" className="gamestore-nav-link">
                {item}
              </a>
            ))}
          </nav>

          <div className="flex items-center gap-2">
            <div className="hidden sm:flex items-center gap-2 mr-2">
              <div className="gamestore-search-shell">
                <Search className="h-4 w-4 text-muted-foreground" />
                <input
                  className="gamestore-search-input"
                  value={keyword}
                  onChange={(e) => onKeywordChange(e.target.value)}
                  placeholder="搜索游戏 / 标签 / 开发商"
                />
              </div>
              <Button size="icon" variant="ghost" onClick={onSearch} aria-label="搜索">
                <Search className="h-4 w-4" />
              </Button>
            </div>

            <Button variant="ghost" size="icon" className="text-muted-foreground hover:text-foreground relative">
              <ShoppingCart className="h-5 w-5" />
              <span className="gamestore-badge-count">2</span>
            </Button>
            <Button variant="ghost" size="icon" className="text-muted-foreground hover:text-foreground">
              <User className="h-5 w-5" />
            </Button>

            <Button
              variant="ghost"
              size="icon"
              className="md:hidden text-muted-foreground hover:text-foreground"
              onClick={() => setIsMenuOpen(!isMenuOpen)}
              aria-label="菜单"
            >
              {isMenuOpen ? <X className="h-5 w-5" /> : <MenuIcon className="h-5 w-5" />}
            </Button>
          </div>
        </div>

        <div
          className={cn(
            'md:hidden overflow-hidden transition-all duration-300',
            isMenuOpen ? 'max-h-48 pb-4' : 'max-h-0',
          )}
        >
          <div className="flex flex-col gap-3">
            <div className="gamestore-search-shell w-full">
              <Search className="h-4 w-4 text-muted-foreground" />
              <input
                className="gamestore-search-input"
                value={keyword}
                onChange={(e) => onKeywordChange(e.target.value)}
                placeholder="搜索游戏 / 标签 / 开发商"
              />
              <button type="button" className="gamestore-search-btn" onClick={onSearch}>
                搜索
              </button>
            </div>
            <nav className="flex flex-col gap-2">
              {navItems.map((item) => (
                <a key={item} href="#" className="gamestore-nav-mobile-link">
                  {item}
                </a>
              ))}
            </nav>
          </div>
        </div>
      </div>
    </header>
  )
}

const FeaturedCarousel = ({
  games,
  onGoDetail,
}: {
  games: CardGame[]
  onGoDetail: (id: number) => void
}) => {
  const [currentIndex, setCurrentIndex] = useState(0)
  const [isAutoPlaying, setIsAutoPlaying] = useState(true)

  useEffect(() => {
    if (!isAutoPlaying) return
    if (games.length <= 1) return
    const t = window.setInterval(() => {
      setCurrentIndex((prev) => (prev + 1) % games.length)
    }, HERO_AUTOPLAY_MS)
    return () => window.clearInterval(t)
  }, [isAutoPlaying, games.length])

  useEffect(() => {
    setCurrentIndex(0)
  }, [games.map((g) => g.id).join('|')])

  const prevSlide = () => {
    if (games.length <= 1) return
    setCurrentIndex((prev) => (prev - 1 + games.length) % games.length)
  }
  const nextSlide = () => {
    if (games.length <= 1) return
    setCurrentIndex((prev) => (prev + 1) % games.length)
  }

  const currentGame = games[currentIndex]

  return (
    <section
      className="relative"
      onMouseEnter={() => setIsAutoPlaying(false)}
      onMouseLeave={() => setIsAutoPlaying(true)}
    >
      <div
        className="relative h-[500px] md:h-[600px] lg:h-[700px] overflow-hidden rounded-2xl"
      >
        {games.map((game, index) => (
          <div
            key={game.id}
            className={cn(
              'absolute inset-0 transition-all duration-700 ease-in-out',
              index === currentIndex ? 'opacity-100 scale-100' : 'opacity-0 scale-105',
            )}
          >
            <img src={game.cover} alt={game.title} className="object-cover h-full w-full" />
            <div className="absolute inset-0 bg-gradient-to-r from-background via-background/60 to-transparent" />
            <div className="absolute inset-0 bg-gradient-to-t from-background via-transparent to-transparent" />
          </div>
        ))}

        <div className="absolute inset-0 flex items-center">
          <div className="mx-auto max-w-7xl w-full px-4 sm:px-6 lg:px-8">
            <div className="max-w-xl">
              <div className="flex flex-wrap gap-2 mb-4">
                {(currentGame?.tags || []).slice(0, 6).map((tag) => (
                  <span key={tag} className="rounded-full bg-primary/20 px-3 py-1 text-xs font-medium text-primary">
                    {tag}
                  </span>
                ))}
              </div>

              <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold text-foreground mb-2 text-balance">
                {currentGame?.title || '—'}
              </h1>
              <p className="text-xl md:text-2xl text-primary font-medium mb-4">
                {currentGame?.subType || currentGame?.category || ''}
              </p>

              <p className="text-muted-foreground text-base md:text-lg mb-6 text-pretty">
                {currentGame?.description || '真实接口返回的当前精选游戏。'}
              </p>

              <div className="flex flex-wrap items-center gap-4">
                <div className="flex items-center gap-3">
                  {/* 当前项目后端可能不包含 price/discount，这里用下载热度保持布局占位 */}
                  <div className="flex items-baseline gap-2">
                    <span className="text-2xl font-bold text-foreground">
                      {currentGame?.downloads ? `¥${currentGame.downloads}` : '¥--'}
                    </span>
                  </div>
                </div>

                <Button size="lg" className="gap-2" onClick={() => currentGame && onGoDetail(currentGame.id)}>
                    <Play className="h-4 w-4" />
                    立即购买
                </Button>
                <Button size="lg" variant="secondary" onClick={() => currentGame && onGoDetail(currentGame.id)}>
                    了解更多
                  </Button>
                </div>
              </div>
            </div>
          </div>

        {games.length > 1 ? (
          <>
            <div className="absolute inset-y-0 left-4 flex items-center">
              <Button
                variant="secondary"
                size="icon"
                className="h-12 w-12 rounded-full bg-background/50 backdrop-blur-sm hover:bg-background/80"
                onClick={prevSlide}
                aria-label="上一张"
              >
                <ChevronLeft className="h-6 w-6" />
              </Button>
            </div>
            <div className="absolute inset-y-0 right-4 flex items-center">
              <Button
                variant="secondary"
                size="icon"
                className="h-12 w-12 rounded-full bg-background/50 backdrop-blur-sm hover:bg-background/80"
                onClick={nextSlide}
                aria-label="下一张"
              >
                <ChevronRight className="h-6 w-6" />
              </Button>
            </div>
            <div className="absolute bottom-6 left-1/2 -translate-x-1/2 flex gap-2">
              {games.map((g, i) => (
                <button
                  key={g.id}
                  type="button"
                  onClick={() => setCurrentIndex(i)}
                  className={cn(
                    'h-2 rounded-full transition-all duration-300',
                    i === currentIndex ? 'w-8 bg-primary' : 'w-2 bg-foreground/30 hover:bg-foreground/50',
                  )}
                />
              ))}
            </div>
          </>
        ) : null}
      </div>

      {games.length > 1 ? (
        <div className="mt-4 flex gap-4 overflow-x-auto pb-2 scrollbar-hide gamestore-featured-thumbs">
          {games.map((g, i) => (
            <button
              key={g.id}
              type="button"
              onClick={() => setCurrentIndex(i)}
              className={cn(
                'relative flex-shrink-0 overflow-hidden rounded-lg transition-all duration-300',
                i === currentIndex
                  ? 'ring-2 ring-primary ring-offset-2 ring-offset-background'
                  : 'opacity-60 hover:opacity-100',
              )}
            >
              <div className="relative h-20 w-36">
                <img src={g.cover} alt={g.title} className="object-cover h-full w-full" />
              </div>
              <div className="absolute inset-0 bg-gradient-to-t from-background/70 via-background/30 to-transparent" />
            </button>
          ))}
        </div>
      ) : null}
    </section>
  )
}

export default function GameHomeStoreReact() {
  const [loading, setLoading] = useState(false)
  const initialKeyword = useMemo(() => readKeywordFromURL(), [])
  const [keywordInput, setKeywordInput] = useState(initialKeyword)
  const [searchKeyword, setSearchKeyword] = useState(initialKeyword)
  const [activeCategory, setActiveCategory] = useState('全部')
  const [currentPage, setCurrentPage] = useState(1)

  const [categories, setCategories] = useState<CategoryOption[]>([])
  const [games, setGames] = useState<CardGame[]>([])
  const [featuredGames, setFeaturedGames] = useState<CardGame[]>([])
  const [total, setTotal] = useState(0)

  const [navItems, setNavItems] = useState<string[]>([])

  const categoryNameToId = useMemo(() => new Map(categories.map((c) => [c.name, c.id])), [categories])
  const hasSearchKeyword = searchKeyword.trim().length > 0

  const gotoGameDetail = (id: number) => {
    window.location.href = `/games/${id}`
  }

  const triggerSearch = () => {
    const nextKeyword = keywordInput.trim()
    syncKeywordToURL(nextKeyword)
    setSearchKeyword(nextKeyword)
    setCurrentPage(1)
  }

  useEffect(() => {
    let disposed = false
    const run = async () => {
      try {
        const { data: navRes } = await publicNavMenuList('home_promo')
        if (disposed) return
        if (navRes.code === 200 && Array.isArray(navRes.data?.list)) {
          const mapped = navRes.data.list
            .filter((v: any) => String(v.title || '').trim())
            .map((v: any) => String(v.title))
          if (mapped.length) setNavItems(mapped.slice(0, 6))
        }
      } catch {}
    }
    void run()
    return () => {
      disposed = true
    }
  }, [])

  useEffect(() => {
    let disposed = false
    const run = async () => {
      try {
        const { data: res } = await publicGameCategoryList()
        if (disposed) return
        if (res.code === 200 && Array.isArray(res.data)) {
          setCategories(
            res.data
              .map((c: any) => ({ id: Number(c.id), name: String(c.name || '') }))
              .filter((c: CategoryOption) => c.name),
          )
        }
      } catch {}
    }
    void run()
    return () => {
      disposed = true
    }
  }, [])

  useEffect(() => {
    let disposed = false
    const run = async () => {
      setLoading(true)
      try {
        const categoryID = activeCategory !== '全部' ? categoryNameToId.get(activeCategory) : undefined

        const listParams: Record<string, unknown> = { page: currentPage, page_size: PAGE_SIZE }
        const recommendParams: Record<string, unknown> = { page: 1, page_size: RECOMMEND_PAGE_SIZE }

        if (searchKeyword.trim()) {
          listParams.keyword = searchKeyword.trim()
          recommendParams.keyword = searchKeyword.trim()
        }
        if (categoryID) {
          listParams.category_id = categoryID
          recommendParams.category_id = categoryID
        }

        const [{ data: listRes }, { data: recommendRes }] = await Promise.all([
          publicGameList(listParams),
          publicGameList(recommendParams),
        ])

        if (disposed) return

        if (listRes.code !== 200) {
          setGames([])
          setTotal(0)
          return
        }

        const list = Array.isArray(listRes.data?.list) ? (listRes.data.list as PublicGameItem[]) : []
        setGames(
          list.map((item) => {
            const cat = categories.find((c) => c.id === Number(item.category_id))
            return toCardItem(item, cat?.name || activeCategory)
          }),
        )
        setTotal(Number(listRes.data?.total || list.length))

        const recommendList = Array.isArray(recommendRes.data?.list) ? (recommendRes.data.list as PublicGameItem[]) : []
        const ranked = [...recommendList]
          .sort((a, b) => Number(b.downloads || 0) - Number(a.downloads || 0))
          .slice(0, HERO_SLIDE_COUNT)
          .map((item) => {
            const cat = categories.find((c) => c.id === Number(item.category_id))
            return toCardItem(item, cat?.name || activeCategory)
          })
        setFeaturedGames(ranked)
      } finally {
        if (!disposed) setLoading(false)
      }
    }
    void run()
    return () => {
      disposed = true
    }
  }, [activeCategory, categories, currentPage, searchKeyword, categoryNameToId])

  const categoryCards = useMemo(() => {
    return [{ id: 0, name: '全部' }, ...categories]
  }, [categories])

  // 如果推荐/热度接口为空，使用当前列表做兜底，保证轮播结构不会“消失”
  const featuredSlides = useMemo(() => {
    if (featuredGames.length) return featuredGames.slice(0, HERO_SLIDE_COUNT)
    return games.slice(0, HERO_SLIDE_COUNT)
  }, [featuredGames, games])

  return (
    <div className="gamestore-scope bg-background text-foreground min-h-screen">
      <GameStoreHeader keyword={keywordInput} onKeywordChange={setKeywordInput} onSearch={triggerSearch} />

      <main className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-8">
        {featuredSlides.length ? (
          <section className="mb-12">
            <FeaturedCarousel games={featuredSlides} onGoDetail={gotoGameDetail} />
          </section>
        ) : null}

        <section className="mb-12">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-2xl font-bold text-foreground">游戏分类</h2>
            <button type="button" className="text-sm font-medium text-primary hover:underline">
              查看全部
            </button>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-8 gap-4">
            {categoryCards.map((c, idx) => {
              const isActive = c.name === activeCategory
              const bg = idx % 2 === 0 ? 'bg-primary/10 text-primary hover:bg-primary/20' : 'bg-accent/10 text-accent hover:bg-accent/20'
              return (
                <button
                  key={c.name}
                  type="button"
                  onClick={() => {
                    setActiveCategory(c.name)
                    setCurrentPage(1)
                  }}
                  className={cn(
                    'flex flex-col items-center gap-3 rounded-xl p-4 transition-all duration-300 border border-border/60',
                    bg,
                    isActive ? 'ring-2 ring-primary' : 'hover:border-primary/50',
                  )}
                >
                  <div className="text-center">
                    <p className="font-medium text-sm">{c.name}</p>
                    <p className="text-xs opacity-70">{c.name === '全部' ? '精选游戏' : '分类游戏'}</p>
                  </div>
                </button>
              )
            })}
          </div>
        </section>

        <section className="mb-12">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-4">
              <h2 className="text-2xl font-bold text-foreground">限时热度</h2>
              <span className="rounded-full bg-primary/20 px-3 py-1 text-xs font-medium text-primary">下载热度</span>
            </div>
            <Button variant="ghost" className="gap-2 text-primary" onClick={() => setActiveCategory('全部')}>
              查看全部
              <ArrowRight className="h-4 w-4" />
            </Button>
          </div>

          <div className="grid gap-4 md:grid-cols-3">
            {featuredSlides.slice(0, 3).map((g) => (
              <div
                key={g.id}
                className="group relative overflow-hidden rounded-xl bg-card border border-border transition-all duration-300 hover:border-primary/50 p-4"
              >
                <div className="flex gap-4">
                  <div className="relative h-24 w-24 flex-shrink-0 overflow-hidden rounded-lg bg-background">
                    <img src={g.cover} alt={g.title} className="object-cover h-full w-full" />
                  </div>
                  <div className="flex flex-col justify-between flex-1 min-w-0">
                    <div>
                      <h3 className="font-semibold text-foreground truncate group-hover:text-primary transition-colors">{g.title}</h3>
                      <div className="mt-2 text-muted-foreground text-sm">热度：{g.downloads}</div>
                    </div>
                    <div className="flex items-center justify-between mt-2">
                      <span className="rounded bg-primary px-2 py-0.5 text-xs font-bold text-primary-foreground">
                        {g.score.toFixed(1)} 分
                      </span>
                      <Button size="sm" onClick={() => gotoGameDetail(g.id)}>
                        查看
                      </Button>
                    </div>
                  </div>
                </div>
              </div>
            ))}
            {featuredSlides.length === 0 ? <div className="text-muted-foreground">暂无热度数据</div> : null}
          </div>
        </section>

        <section className="mb-12">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
            <div className="flex items-center gap-2 overflow-x-auto pb-2 sm:pb-0">
              <div className="whitespace-nowrap rounded-full px-4 py-2 text-sm font-medium transition-all duration-300 bg-primary/20 text-primary">
                {hasSearchKeyword ? `搜索：${searchKeyword.trim()}` : activeCategory === '全部' ? '发现游戏' : `${activeCategory} 游戏`}
              </div>
              <div className="whitespace-nowrap rounded-full px-4 py-2 text-sm font-medium transition-all duration-300 bg-secondary text-muted-foreground hover:bg-secondary/80 hover:text-foreground">
                共 {total} 款
              </div>
            </div>

            <Button variant="ghost" className="gap-2 text-primary self-end" onClick={() => setCurrentPage(1)}>
              刷新
            </Button>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-4">
            {games.map((g) => (
              <div
                key={g.id}
                className="group rounded-lg bg-card overflow-hidden border border-border transition-transform hover:scale-[1.02] cursor-pointer"
                onClick={() => gotoGameDetail(g.id)}
                role="button"
                tabIndex={0}
              >
                <div className="relative aspect-[3/4] overflow-hidden">
                  <img src={g.cover} alt={g.title} className="object-cover transition-transform duration-500 group-hover:scale-110 h-full w-full" />
                  <div className="absolute inset-0 bg-gradient-to-t from-background/80 via-transparent to-transparent" />
                </div>
                <div className="p-4">
                  <h4 className="font-semibold text-foreground text-sm line-clamp-2">{g.title}</h4>
                  <div className="mt-2 flex items-center justify-between">
                    <span className="text-xs text-muted-foreground">{g.category}</span>
                    <span className="text-sm font-bold text-primary">{g.downloads}</span>
                  </div>
                  {g.tags.length ? (
                    <div className="mt-3 flex flex-wrap gap-1">
                      {g.tags.slice(0, 2).map((t) => (
                        <span key={t} className="rounded-full bg-secondary px-2 py-0.5 text-[10px] font-medium text-muted-foreground">
                          {t}
                        </span>
                      ))}
                    </div>
                  ) : null}
                </div>
              </div>
            ))}
          </div>

          {loading ? <div className="mt-6 text-muted-foreground">加载中...</div> : null}
          {!loading && games.length === 0 ? <div className="mt-6 text-muted-foreground">暂无游戏数据</div> : null}

          <div className="mt-6 flex items-center justify-end gap-3">
            <Button variant="secondary" disabled={currentPage <= 1} onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}>
              上一页
            </Button>
            <div className="text-sm text-muted-foreground">
              第 {currentPage} 页 / 共 {Math.max(1, Math.ceil(total / PAGE_SIZE))} 页
            </div>
            <Button
              variant="secondary"
              disabled={currentPage >= Math.ceil(total / PAGE_SIZE)}
              onClick={() => setCurrentPage((p) => p + 1)}
            >
              下一页
            </Button>
          </div>
        </section>

        <section className="mb-12">
          <div className="relative overflow-hidden rounded-2xl bg-gradient-to-r from-primary/20 via-card to-primary/10 border border-border p-8 md:p-12">
            <div className="relative z-10 max-w-2xl">
              <h2 className="text-2xl md:text-3xl font-bold text-foreground mb-4 text-balance">订阅我们的新闻通讯</h2>
              <p className="text-muted-foreground mb-6 text-pretty">第一时间获取最新游戏资讯、独家优惠和限时折扣活动。</p>
              <div className="flex flex-col sm:flex-row gap-3">
                <input
                  type="email"
                  placeholder="输入您的邮箱地址"
                  className="flex-1 rounded-lg bg-background border border-border px-4 py-3 text-foreground placeholder:text-muted-foreground outline-none"
                />
                <button type="button" className="rounded-lg bg-primary px-6 py-3 font-medium text-primary-foreground hover:bg-primary/90 transition-colors">
                  立即订阅
                </button>
              </div>
            </div>
            <div className="absolute top-0 right-0 w-64 h-64 bg-primary/10 rounded-full blur-3xl" />
            <div className="absolute bottom-0 right-1/4 w-48 h-48 bg-primary/5 rounded-full blur-2xl" />
          </div>
        </section>
      </main>

      <footer className="gamestore-footer">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-12">
          <div className="grid grid-cols-2 md:grid-cols-5 gap-8">
            <div className="col-span-2 md:col-span-1">
              <div className="flex items-center gap-2 mb-4">
                <Gamepad2 className="h-8 w-8 text-primary" />
                <span className="text-xl font-bold text-foreground">GameStore</span>
              </div>
              <p className="text-sm text-muted-foreground">探索无限游戏世界，发现你的下一个最爱。</p>
            </div>
            <div>
              <h3 className="font-semibold text-foreground mb-4">产品</h3>
              <ul className="space-y-2">
                {['游戏商店', '社区中心', '创意工坊', '直播'].map((t) => (
                  <li key={t}>
                    <a href="#" className="gamestore-footer-link">
                      {t}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
            <div>
              <h3 className="font-semibold text-foreground mb-4">支持</h3>
              <ul className="space-y-2">
                {['帮助中心', '联系我们', '退款政策', '账户安全'].map((t) => (
                  <li key={t}>
                    <a href="#" className="gamestore-footer-link">
                      {t}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
            <div>
              <h3 className="font-semibold text-foreground mb-4">公司</h3>
              <ul className="space-y-2">
                {['关于我们', '招贤纳士', '新闻动态', '合作伙伴'].map((t) => (
                  <li key={t}>
                    <a href="#" className="gamestore-footer-link">
                      {t}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
            <div>
              <h3 className="font-semibold text-foreground mb-4">法律</h3>
              <ul className="space-y-2">
                {['服务条款', '隐私政策', 'Cookie 政策', '版权声明'].map((t) => (
                  <li key={t}>
                    <a href="#" className="gamestore-footer-link">
                      {t}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          </div>

          <div className="mt-12 pt-8 border-t border-border flex flex-col sm:flex-row justify-between items-center gap-4">
            <p className="text-sm text-muted-foreground">© 2026 GameStore. 保留所有权利。</p>
            <div className="flex items-center gap-6">
              <a href="#" className="text-muted-foreground hover:text-foreground transition-colors" aria-label="Twitter">
                <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M24 4.557c-.883.392-1.832.656-2.828.775 1.017-.609 1.798-1.574 2.165-2.724-.951.564-2.005.974-3.127 1.195-.897-.957-2.178-1.555-3.594-1.555-3.179 0-5.515 2.966-4.797 6.045-4.091-.205-7.719-2.165-10.148-5.144-1.29 2.213-.669 5.108 1.523 6.574-.806-.026-1.566-.247-2.229-.616-.054 2.281 1.581 4.415 3.949 4.89-.693.188-1.452.232-2.224.084.626 1.956 2.444 3.379 4.6 3.419-2.07 1.623-4.678 2.348-7.29 2.04 2.179 1.397 4.768 2.212 7.548 2.212 9.142 0 14.307-7.721 13.995-14.646.962-.695 1.797-1.562 2.457-2.549z" />
                </svg>
              </a>
              <a href="#" className="text-muted-foreground hover:text-foreground transition-colors" aria-label="Facebook">
                <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23 1.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
                </svg>
              </a>
              <a href="#" className="text-muted-foreground hover:text-foreground transition-colors" aria-label="Instagram">
                <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028 14.09 14.09 0 0 0 1.226-1.994.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 1 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.956-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.946 2.418-2.157 2.418z" />
                </svg>
              </a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}

