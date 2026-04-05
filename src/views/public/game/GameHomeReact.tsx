import '@semi-ui-styles/semi.min.css'

import { IconChevronLeft, IconChevronRight } from '@douyinfe/semi-icons'
import { Banner, Empty, Pagination, Spin, Tag, Toast } from '@douyinfe/semi-ui'
import { startTransition, useDeferredValue, useEffect, useMemo, useState } from 'react'
import {
  publicGameCategoryList,
  publicGameList,
  publicNavMenuList,
  type PublicGameItem,
} from './api'
import CategoryNav from './components/CategoryNav'
import GameGrid from './components/GameGrid'
import TopNav, { type NavItem } from './components/TopNav'
import styles from './GameHome.module.scss'
import { type GameItem } from './mockData'

const PAGE_SIZE = 10
const RECOMMEND_PAGE_SIZE = 8
const HERO_SLIDE_COUNT = 5
const HERO_AUTOPLAY_MS = 5000

const fallbackNav: NavItem[] = [
  { title: '首页', path: '/' },
  { title: '游戏', path: '/games' },
]

type CategoryOption = {
  id: number
  name: string
}

const readKeywordFromURL = () => {
  if (typeof window === 'undefined') return ''
  return new URLSearchParams(window.location.search).get('keyword')?.trim() || ''
}

const syncKeywordToURL = (keyword: string) => {
  if (typeof window === 'undefined') return
  const url = new URL(window.location.href)
  if (keyword.trim()) url.searchParams.set('keyword', keyword.trim())
  else url.searchParams.delete('keyword')
  window.history.replaceState({}, '', url.toString())
}

const toDateText = (value?: string) => {
  if (!value) return '-'
  const t = new Date(value)
  if (Number.isNaN(t.getTime())) return value
  return t.toISOString().slice(0, 10)
}

const toCardItem = (item: PublicGameItem, categoryName: string): GameItem => {
  const cover =
    String(item.cover || item.banner || item.header_image || '').trim() ||
    `https://picsum.photos/seed/fallback${item.id}/480/260`

  const scoreRaw = Number(item.rating || 0)
  const steamScoreRaw = Number(item.steam_score || 0)
  const score = scoreRaw > 0 ? scoreRaw : Number((steamScoreRaw / 10).toFixed(1))
  const createdAt = String(item.created_at || item.updated_at || '')

  return {
    id: Number(item.id || 0),
    title: String(item.title || '未命名游戏'),
    cover,
    category: categoryName || String(item.type || '未分类'),
    subType: String(item.type || '单机'),
    size: String(item.size || '-'),
    score: score > 0 ? score : 7.5,
    releaseDate: toDateText(String(item.release_date || createdAt)),
    updateText: createdAt ? `${toDateText(createdAt)} 更新` : '最近更新',
    downloads: String(item.downloads || 0),
    description: String(item.short_description || item.description || '').trim(),
    developer: String(item.developer || '').trim(),
    publisher: String(item.publishers || '').trim(),
  }
}

export default function GameHomeReact() {
  const initialKeyword = useMemo(() => readKeywordFromURL(), [])
  const [keywordInput, setKeywordInput] = useState(initialKeyword)
  const [searchKeyword, setSearchKeyword] = useState(initialKeyword)
  const [activeCategory, setActiveCategory] = useState('全部')
  const [currentPage, setCurrentPage] = useState(1)

  const [categories, setCategories] = useState<CategoryOption[]>([])
  const [loading, setLoading] = useState(false)
  const [games, setGames] = useState<GameItem[]>([])
  const [recommendGames, setRecommendGames] = useState<GameItem[]>([])
  const [total, setTotal] = useState(0)
  const [navItems, setNavItems] = useState<NavItem[]>(fallbackNav)
  const [activeHeroIndex, setActiveHeroIndex] = useState(0)

  const deferredKeyword = useDeferredValue(searchKeyword)
  const categoryNameToId = useMemo(() => new Map(categories.map((c) => [c.name, c.id])), [categories])
  const categoryTabs = useMemo(() => ['全部', ...categories.map((c) => c.name)], [categories])
  const hasSearchKeyword = deferredKeyword.trim().length > 0
  const heroSlides = useMemo(() => games.slice(0, HERO_SLIDE_COUNT), [games])
  const featuredGame = heroSlides[activeHeroIndex] || games[0]
  const gameList = useMemo(() => {
    if (!featuredGame) return games
    return games.filter((game) => game.id !== featuredGame.id)
  }, [featuredGame, games])

  useEffect(() => {
    let disposed = false

    const run = async () => {
      try {
        const { data: navRes } = await publicNavMenuList('top_nav')
        if (disposed) return

        if (navRes.code === 200 && Array.isArray(navRes.data?.list) && navRes.data.list.length > 0) {
          const mapped = navRes.data.list
            .filter((v) => String(v.title || '').trim())
            .map((v) => ({ title: String(v.title), path: String(v.path || '#') }))
          if (mapped.length > 0) setNavItems(mapped)
        }
      } catch {
        // keep fallback
      }
    }

    run()
    return () => {
      disposed = true
    }
  }, [])

  useEffect(() => {
    setActiveHeroIndex(0)
  }, [activeCategory, deferredKeyword, games.length])

  useEffect(() => {
    if (heroSlides.length <= 1) return
    const timer = window.setInterval(() => {
      setActiveHeroIndex((prev) => (prev + 1) % heroSlides.length)
    }, HERO_AUTOPLAY_MS)
    return () => {
      window.clearInterval(timer)
    }
  }, [heroSlides.length])

  useEffect(() => {
    let disposed = false

    const run = async () => {
      try {
        const { data: res } = await publicGameCategoryList()
        if (disposed) return

        if (res.code === 200 && Array.isArray(res.data)) {
          setCategories(
            res.data
              .map((c) => ({ id: Number(c.id), name: String(c.name || '') }))
              .filter((c) => c.name),
          )
        }
      } catch {
        // keep fallback
      }
    }

    run()
    return () => {
      disposed = true
    }
  }, [])

  useEffect(() => {
    let disposed = false

    const run = async () => {
      setLoading(true)
      try {
        const kw = deferredKeyword.trim()
        const categoryID = activeCategory !== '全部' ? categoryNameToId.get(activeCategory) : undefined
        const listParams: Record<string, unknown> = {
          page: currentPage,
          page_size: PAGE_SIZE,
        }
        const recommendParams: Record<string, unknown> = {
          page: 1,
          page_size: 50,
        }
        if (kw) {
          listParams.keyword = kw
          recommendParams.keyword = kw
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
          Toast.error('游戏列表接口返回异常')
          setGames([])
          setRecommendGames([])
          setTotal(0)
          return
        }

        const list = Array.isArray(listRes.data?.list) ? listRes.data.list : []
        setGames(
          list.map((item) => {
            const cat = categories.find((c) => c.id === Number(item.category_id))
            return toCardItem(item, cat?.name || activeCategory)
          }),
        )
        setTotal(Number(listRes.data?.total || list.length))

        const recommendList = Array.isArray(recommendRes.data?.list) ? recommendRes.data.list : []
        const ranked = [...recommendList]
          .sort((a, b) => Number(b.downloads || 0) - Number(a.downloads || 0))
          .slice(0, RECOMMEND_PAGE_SIZE)
          .map((item) => {
            const cat = categories.find((c) => c.id === Number(item.category_id))
            return toCardItem(item, cat?.name || activeCategory)
          })
        setRecommendGames(ranked)
      } catch {
        if (!disposed) {
          Toast.error('游戏列表加载失败，请检查后端接口')
          setGames([])
          setRecommendGames([])
          setTotal(0)
        }
      } finally {
        if (!disposed) setLoading(false)
      }
    }

    run()
    return () => {
      disposed = true
    }
  }, [activeCategory, categories, currentPage, deferredKeyword, categoryNameToId])

  const triggerSearch = () => {
    const nextKeyword = keywordInput.trim()
    syncKeywordToURL(nextKeyword)
    startTransition(() => {
      setSearchKeyword(nextKeyword)
      setCurrentPage(1)
    })
  }

  const clearSearch = () => {
    syncKeywordToURL('')
    startTransition(() => {
      setKeywordInput('')
      setSearchKeyword('')
      setCurrentPage(1)
    })
  }

  const changeHero = (direction: 'prev' | 'next') => {
    if (heroSlides.length <= 1) return
    setActiveHeroIndex((prev) => {
      if (direction === 'prev') return prev === 0 ? heroSlides.length - 1 : prev - 1
      return (prev + 1) % heroSlides.length
    })
  }

  return (
    <div className={styles.page}>
      <TopNav navItems={navItems} keyword={keywordInput} onKeywordChange={setKeywordInput} onSearch={triggerSearch} />

      <div className={`${styles.mainWrap} ${styles.mainWrapFull}`}>
        <div className={styles.homeShell}>
          <CategoryNav
            categories={categoryTabs}
            activeCategory={activeCategory}
            onCategoryChange={(value) => {
              startTransition(() => {
                setActiveCategory(value)
                setCurrentPage(1)
              })
            }}
          />

          <main className={styles.contentMain}>
            <div className={styles.listWrap}>
              {hasSearchKeyword ? (
                <Banner
                  type="info"
                  bordered
                  className={styles.searchBanner}
                  title={
                    <div className={styles.searchBannerTitle}>
                      <span>搜索结果</span>
                      <Tag color="blue" shape="circle">
                        {deferredKeyword}
                      </Tag>
                      <span className={styles.searchBannerMeta}>共找到 {total} 款相关游戏</span>
                      <button type="button" className={styles.searchClear} onClick={clearSearch}>
                        清空搜索
                      </button>
                    </div>
                  }
                />
              ) : null}

              {featuredGame ? (
                <section className={styles.heroCarousel}>
                  <div
                    className={styles.heroPanel}
                    role="button"
                    tabIndex={0}
                    onClick={() => {
                      window.location.href = `/games/${featuredGame.id}`
                    }}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault()
                        window.location.href = `/games/${featuredGame.id}`
                      }
                    }}
                  >
                    <img
                      key={featuredGame.id}
                      src={featuredGame.cover}
                      alt={featuredGame.title}
                      className={styles.heroCover}
                    />
                    <div className={styles.heroOverlay} />

                    {heroSlides.length > 1 ? (
                      <>
                        <button
                          type="button"
                          aria-label="上一张"
                          className={`${styles.heroArrow} ${styles.heroArrowLeft}`}
                          onClick={(e) => {
                            e.stopPropagation()
                            changeHero('prev')
                          }}
                        >
                          <IconChevronLeft />
                        </button>
                        <button
                          type="button"
                          aria-label="下一张"
                          className={`${styles.heroArrow} ${styles.heroArrowRight}`}
                          onClick={(e) => {
                            e.stopPropagation()
                            changeHero('next')
                          }}
                        >
                          <IconChevronRight />
                        </button>
                      </>
                    ) : null}

                    <div className={styles.heroContent}>
                      <div className={styles.heroBadge}>{activeCategory === '全部' ? 'TapTap 风格精选' : activeCategory}</div>
                      <h2 className={styles.heroTitle}>{featuredGame.title}</h2>
                      <p className={styles.heroDesc}>{featuredGame.description || '真实后端返回的当前分类精选游戏。'}</p>
                      <div className={styles.heroMeta}>
                        <span>{featuredGame.category}</span>
                        <span>{featuredGame.subType}</span>
                        <span>{featuredGame.releaseDate}</span>
                        <span>{featuredGame.downloads} 下载</span>
                      </div>
                    </div>
                  </div>

                  {heroSlides.length > 1 ? (
                    <div className={styles.heroRail}>
                      <div className={styles.heroDots}>
                        {heroSlides.map((game, index) => (
                          <button
                            key={game.id}
                            type="button"
                            aria-label={`切换到 ${game.title}`}
                            className={index === activeHeroIndex ? styles.heroDotActive : styles.heroDot}
                            onClick={() => setActiveHeroIndex(index)}
                          />
                        ))}
                      </div>

                      <div className={styles.heroThumbList}>
                        {heroSlides.map((game, index) => (
                          <button
                            key={game.id}
                            type="button"
                            className={index === activeHeroIndex ? styles.heroThumbActive : styles.heroThumb}
                            onClick={() => setActiveHeroIndex(index)}
                          >
                            <img src={game.cover} alt={game.title} className={styles.heroThumbCover} />
                            <div className={styles.heroThumbBody}>
                              <div className={styles.heroThumbTitle}>{game.title}</div>
                              <div className={styles.heroThumbMeta}>
                                <span>{game.category}</span>
                                <span>{game.downloads} 下载</span>
                              </div>
                            </div>
                          </button>
                        ))}
                      </div>
                    </div>
                  ) : null}
                </section>
              ) : null}

              <div className={styles.centerSectionHead}>
                <div>
                  <h3 className={styles.centerSectionTitle}>
                    {hasSearchKeyword ? '相关游戏' : activeCategory === '全部' ? '发现游戏' : `${activeCategory} 游戏`}
                  </h3>
                  <p className={styles.centerSectionSub}>左侧切分类，中间大图内容区，右侧展示游戏推荐</p>
                </div>
                <div className={styles.centerSectionCount}>{total} 款</div>
              </div>

              {loading ? (
                <div className={styles.loadingWrap}>
                  <Spin />
                </div>
              ) : games.length > 0 ? (
                <>
                  <GameGrid games={gameList} />
                  {total > PAGE_SIZE ? (
                    <div className={styles.pagerWrap}>
                      <Pagination
                        total={total}
                        pageSize={PAGE_SIZE}
                        currentPage={currentPage}
                        onPageChange={(page) => {
                          startTransition(() => setCurrentPage(page))
                        }}
                      />
                    </div>
                  ) : null}
                </>
              ) : (
                <Empty
                  title={hasSearchKeyword ? '没有找到相关游戏' : '暂无游戏数据'}
                  description={hasSearchKeyword ? '试试更换关键词或切换分类后再搜索。' : '当前分类没有返回游戏数据'}
                />
              )}
            </div>
          </main>

          <aside className={styles.recommendAside}>
            <div className={styles.recommendCard}>
              <div className={styles.recommendHeader}>
                <h3 className={styles.recommendTitle}>游戏推荐</h3>
                <span className={styles.recommendSub}>下载热度</span>
              </div>

              {recommendGames.length > 0 ? (
                <div className={styles.recommendList}>
                  {recommendGames.map((game, index) => (
                    <article
                      key={game.id}
                      className={styles.recommendItem}
                      role="button"
                      tabIndex={0}
                      onClick={() => {
                        window.location.href = `/games/${game.id}`
                      }}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                          e.preventDefault()
                          window.location.href = `/games/${game.id}`
                        }
                      }}
                    >
                      <span className={styles.recommendRank}>{String(index + 1).padStart(2, '0')}</span>
                      <img src={game.cover} alt={game.title} className={styles.recommendCover} />
                      <div className={styles.recommendBody}>
                        <div className={styles.recommendName}>{game.title}</div>
                        <div className={styles.recommendMeta}>
                          <span>{game.category}</span>
                          <span>{game.downloads} 下载</span>
                        </div>
                      </div>
                    </article>
                  ))}
                </div>
              ) : (
                <Empty title="暂无推荐" description="真实接口暂未返回推荐游戏" image={null} />
              )}
            </div>
          </aside>
        </div>
      </div>
    </div>
  )
}
