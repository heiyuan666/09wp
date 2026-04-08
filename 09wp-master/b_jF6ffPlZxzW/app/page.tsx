import { Header } from "@/components/game/header"
import { FeaturedCarousel, type FeaturedGame } from "@/components/home/featured-carousel"
import { CategorySection, type CategoryItem } from "@/components/home/category-section"
import { DealsSection, type Deal } from "@/components/home/deals-section"
import { TrendingSection, type GameCardItem, type TrendingTabId } from "@/components/home/trending-section"
import { Footer } from "@/components/home/footer"
import { absolutizeGameMediaUrls, fetchGameCategoryList, fetchGameList, splitToList } from "@/lib/api/game"
import { redirect } from "next/navigation"

export const revalidate = 300

function centsToYuanText(cents: number) {
  if (!Number.isFinite(cents) || cents <= 0) return "0"
  return String(Math.round(cents / 100))
}

function buildFeaturedGames(list: ReturnType<typeof absolutizeGameMediaUrls>[]) {
  const out: FeaturedGame[] = []
  for (const g of list.slice(0, 5)) {
    const tags = splitToList(g.tags).slice(0, 5)
    out.push({
      id: g.id,
      title: g.title,
      subtitle: g.short_description ? g.short_description.slice(0, 14) : "精选推荐",
      description: g.short_description || "",
      image: g.banner || g.header_image || g.cover || "/images/featured-1.jpg",
      price: centsToYuanText(g.price_final),
      discount: g.price_discount || undefined,
      tags,
    })
  }
  return out
}

function buildDealGames(list: ReturnType<typeof absolutizeGameMediaUrls>[]) {
  const now = Date.now()
  const deals: Deal[] = []
  for (const g of list) {
    const discount = g.price_discount || 0
    if (discount <= 0) continue
    const original = g.price_initial || g.price_final
    deals.push({
      id: g.id,
      title: g.title,
      image: g.cover || g.header_image || "/images/game-cover.jpg",
      originalPrice: centsToYuanText(original),
      salePrice: centsToYuanText(g.price_final),
      discount,
      endTime: new Date(now + 2 * 24 * 60 * 60 * 1000).toISOString(),
    })
  }
  deals.sort((a, b) => b.discount - a.discount)
  return deals.slice(0, 3)
}

function buildGameCardItems(list: ReturnType<typeof absolutizeGameMediaUrls>[], opts?: { isNew?: boolean; isTrending?: boolean }) {
  const out: GameCardItem[] = []
  for (const g of list) {
    const priceText = g.price_final === 0 || g.price_text === "免费" ? "免费" : centsToYuanText(g.price_final)
    const original =
      g.price_discount && g.price_initial && g.price_initial > g.price_final ? centsToYuanText(g.price_initial) : undefined
    out.push({
      id: g.id,
      title: g.title,
      image: g.cover || g.header_image || "/images/game-cover.jpg",
      price: priceText,
      originalPrice: original,
      discount: g.price_discount || undefined,
      rating: g.rating || undefined,
      releaseDate: g.release_date ? String(g.release_date).slice(0, 10) : undefined,
      tags: splitToList(g.genres).slice(0, 2),
      isNew: opts?.isNew,
      isTrending: opts?.isTrending,
    })
  }
  return out
}

export default async function HomePage({
  searchParams,
}: {
  searchParams?: Promise<{ category?: string }>
}) {
  const resolvedSearch = (await searchParams) || {}
  const selectedCategory = (resolvedSearch.category || "").trim()
  if (selectedCategory) {
    redirect(`/category/${encodeURIComponent(selectedCategory)}`)
  }

  const [catList, listRes] = await Promise.all([
    fetchGameCategoryList(),
    fetchGameList({ page: 1, page_size: 60 }),
  ])

  const allGames = (listRes.list || []).map(absolutizeGameMediaUrls)
  const scopedGames = allGames

  // 热门：按 downloads / likes / rating 做一个简单混合排序
  const trendingSorted = [...scopedGames].sort((a, b) => {
    const sa = (a.downloads || 0) * 1.2 + (a.likes || 0) * 0.7 + (a.rating || 0) * 100
    const sb = (b.downloads || 0) * 1.2 + (b.likes || 0) * 0.7 + (b.rating || 0) * 100
    return sb - sa
  })

  // 新品：按 created_at / id 倒序
  const newSorted = [...scopedGames].sort((a, b) => {
    const ta = Date.parse(a.created_at) || a.id
    const tb = Date.parse(b.created_at) || b.id
    return tb - ta
  })

  // 免费
  const freeSorted = scopedGames.filter((g) => g.price_final === 0 || g.price_text === "免费")

  // 折扣
  const dealsSorted = scopedGames
    .filter((g) => (g.price_discount || 0) > 0)
    .sort((a, b) => (b.price_discount || 0) - (a.price_discount || 0))

  const featured = buildFeaturedGames(trendingSorted)
  const deals = buildDealGames(dealsSorted)

  const gamesByTab: Record<TrendingTabId, GameCardItem[]> = {
    trending: buildGameCardItems(trendingSorted.slice(0, 12), { isTrending: true }),
    new: buildGameCardItems(newSorted.slice(0, 12), { isNew: true }),
    upcoming: buildGameCardItems(newSorted.slice(0, 12)),
    free: buildGameCardItems(freeSorted.slice(0, 12)),
  }

  const palette = [
    "bg-red-500/10 text-red-500 hover:bg-red-500/20",
    "bg-blue-500/10 text-blue-500 hover:bg-blue-500/20",
    "bg-purple-500/10 text-purple-500 hover:bg-purple-500/20",
    "bg-green-500/10 text-green-500 hover:bg-green-500/20",
    "bg-yellow-500/10 text-yellow-500 hover:bg-yellow-500/20",
    "bg-orange-500/10 text-orange-500 hover:bg-orange-500/20",
    "bg-pink-500/10 text-pink-500 hover:bg-pink-500/20",
    "bg-cyan-500/10 text-cyan-500 hover:bg-cyan-500/20",
  ]

  const iconKeys: CategoryItem["icon"][] = ["action", "racing", "horror", "multiplayer", "puzzle", "sports", "rpg", "scifi"]
  const orderBySlug = new Map(iconKeys.map((k, i) => [k, i]))
  const sortedCats = [...(catList || [])]
    .filter((c) => orderBySlug.has((c.slug || "") as CategoryItem["icon"]))
    .sort((a, b) => (orderBySlug.get(a.slug as CategoryItem["icon"]) || 0) - (orderBySlug.get(b.slug as CategoryItem["icon"]) || 0))

  const categoryItems: CategoryItem[] = sortedCats.slice(0, 8).map((c, idx) => ({
    id: c.slug || String(c.id),
    name: c.name,
    icon: (c.slug as CategoryItem["icon"]) || iconKeys[idx % iconKeys.length],
    count: undefined,
    color: palette[idx % palette.length],
    queryValue: String(c.id),
  }))

  return (
    <div className="min-h-screen bg-background">
      <Header />
      
      <main className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-8">
        {/* Featured Carousel */}
        <section className="mb-12">
          <FeaturedCarousel games={featured.length ? featured : []} />
        </section>

        {/* Category Section */}
        <section className="mb-12">
          <CategorySection categories={categoryItems} selectedCategory={selectedCategory} />
        </section>

        {/* Deals Section */}
        <section className="mb-12">
          <DealsSection deals={deals} />
        </section>

        {/* Trending Games */}
        <section className="mb-12">
          <TrendingSection gamesByTab={gamesByTab} />
        </section>

        {/* Newsletter Banner */}
        <section className="mb-12">
          <div className="relative overflow-hidden rounded-2xl bg-gradient-to-r from-primary/20 via-card to-primary/10 border border-border p-8 md:p-12">
            <div className="relative z-10 max-w-2xl">
              <h2 className="text-2xl md:text-3xl font-bold text-foreground mb-4 text-balance">
                订阅我们的新闻通讯
              </h2>
              <p className="text-muted-foreground mb-6 text-pretty">
                第一时间获取最新游戏资讯、独家优惠和限时折扣活动。
              </p>
              <div className="flex flex-col sm:flex-row gap-3">
                <input
                  type="email"
                  placeholder="输入您的邮箱地址"
                  className="flex-1 rounded-lg bg-background border border-border px-4 py-3 text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                />
                <button className="rounded-lg bg-primary px-6 py-3 font-medium text-primary-foreground hover:bg-primary/90 transition-colors">
                  立即订阅
                </button>
              </div>
            </div>
            {/* Decorative elements */}
            <div className="absolute top-0 right-0 w-64 h-64 bg-primary/10 rounded-full blur-3xl" />
            <div className="absolute bottom-0 right-1/4 w-48 h-48 bg-primary/5 rounded-full blur-2xl" />
          </div>
        </section>
      </main>

      <Footer />
    </div>
  )
}
