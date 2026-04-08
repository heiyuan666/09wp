import Link from "next/link"
import { notFound } from "next/navigation"
import { Header } from "@/components/game/header"
import { Footer } from "@/components/home/footer"
import { TrendingSection, type GameCardItem, type TrendingTabId } from "@/components/home/trending-section"
import { absolutizeGameMediaUrls, fetchGameCategoryList, fetchGameList, splitToList } from "@/lib/api/game"

export const revalidate = 300

function centsToYuanText(cents: number) {
  if (!Number.isFinite(cents) || cents <= 0) return "0"
  return String(Math.round(cents / 100))
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

export default async function CategoryDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const categoryId = (id || "").trim()
  if (!categoryId) return notFound()

  const [cats, listRes] = await Promise.all([fetchGameCategoryList(), fetchGameList({ page: 1, page_size: 60, category_id: Number(categoryId) })])
  const cat = (cats || []).find((c) => String(c.id) === categoryId)
  const list = (listRes.list || []).map(absolutizeGameMediaUrls)

  const trendingSorted = [...list].sort((a, b) => {
    const sa = (a.downloads || 0) * 1.2 + (a.likes || 0) * 0.7 + (a.rating || 0) * 100
    const sb = (b.downloads || 0) * 1.2 + (b.likes || 0) * 0.7 + (b.rating || 0) * 100
    return sb - sa
  })
  const newSorted = [...list].sort((a, b) => {
    const ta = Date.parse(a.created_at) || a.id
    const tb = Date.parse(b.created_at) || b.id
    return tb - ta
  })

  const gamesByTab: Record<TrendingTabId, GameCardItem[]> = {
    trending: buildGameCardItems(trendingSorted.slice(0, 24), { isTrending: true }),
    new: buildGameCardItems(newSorted.slice(0, 24), { isNew: true }),
    upcoming: buildGameCardItems(newSorted.slice(0, 24)),
    free: buildGameCardItems(list.filter((g) => g.price_final === 0 || g.price_text === "免费").slice(0, 24)),
  }

  return (
    <div className="min-h-screen bg-background">
      <Header />
      <main className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-10">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold">{cat?.name || `分类 ${categoryId}`}</h1>
            {cat?.description ? <p className="text-sm text-muted-foreground mt-1">{cat.description}</p> : null}
          </div>
          <Link className="text-sm text-primary hover:underline" href="/category">
            返回分类列表
          </Link>
        </div>

        <TrendingSection gamesByTab={gamesByTab} />
      </main>
      <Footer />
    </div>
  )
}

