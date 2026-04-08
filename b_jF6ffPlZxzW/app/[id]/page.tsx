import { Header } from "@/components/game/header"
import { VideoPlayer } from "@/components/game/video-player"
import { ScreenshotGallery } from "@/components/game/screenshot-gallery"
import { GameInfo } from "@/components/game/game-info"
import { DownloadCard } from "@/components/game/download-card"
import { SystemRequirements } from "@/components/game/system-requirements"
import { GameDescription } from "@/components/game/game-description"
import { RelatedGames } from "@/components/game/related-games"
import { UserReviews } from "@/components/game/user-reviews"
import { Footer } from "@/components/home/footer"
import {
  absolutizeGameMediaUrls,
  detectPanTypeByUrl,
  fetchGameDetail,
  fetchGameList,
  splitToList,
  type GameDTO,
  type GameResourceDTO,
} from "@/lib/api/game"

type SystemSpec = {
  os: string
  processor: string
  memory: string
  graphics: string
  storage: string
  directX?: string
}

function formatDateZh(isoLike?: string) {
  if (!isoLike) return ""
  const d = new Date(isoLike)
  if (Number.isNaN(d.getTime())) return isoLike
  return d.toLocaleDateString("zh-CN", { year: "numeric", month: "long", day: "numeric" })
}

function stripHtml(input?: string) {
  if (!input) return ""
  return input
    .replace(/<br\s*\/?>/gi, "\n")
    .replace(/<\/li>/gi, "\n")
    .replace(/<[^>]+>/g, "")
    .replace(/&nbsp;/g, " ")
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/\r/g, "")
    .replace(/\n{3,}/g, "\n\n")
    .trim()
}

function extractReqValue(text: string, labels: string[]) {
  for (const label of labels) {
    const reg = new RegExp(`${label}\\s*[:：]\\s*([^\\n]+)`, "i")
    const m = text.match(reg)
    if (m?.[1]) return m[1].trim()
  }
  return ""
}

function parseRequirement(raw?: string): SystemSpec {
  const txt = stripHtml(raw)
  return {
    os: extractReqValue(txt, ["操作系统", "OS"]) || "未提供",
    processor: extractReqValue(txt, ["处理器", "Processor"]) || "未提供",
    memory: extractReqValue(txt, ["内存", "Memory"]) || "未提供",
    graphics: extractReqValue(txt, ["显卡", "Graphics", "Video"]) || "未提供",
    storage: extractReqValue(txt, ["存储空间", "Storage"]) || "未提供",
    directX: extractReqValue(txt, ["DirectX 版本", "DirectX"]) || undefined,
  }
}

function parseLanguages(raw?: string) {
  const txt = stripHtml(raw)
  return txt
    .replace(/\*具有完全音频支持的语言/gi, "")
    .split(/[,，]+/g)
    .map((s) => s.replace(/\*/g, "").trim())
    .filter(Boolean)
}

function buildDownloadCardDownloads(resources: GameResourceDTO[]) {
  const nameByType = {
    quark: "夸克网盘",
    aliyun: "阿里云盘",
    baidu: "百度网盘",
    lanzou: "蓝奏云",
    "123pan": "123云盘",
    tianyi: "天翼云盘",
    mega: "MEGA",
    onedrive: "OneDrive",
  } as const

  const items: Array<{
    id: string
    name: string
    type: keyof typeof nameByType
    url: string
    size: string
    speed: "fast" | "medium" | "slow"
    isRecommended?: boolean
  }> = []

  for (const res of resources) {
    const size = res.size || ""
    const urls = Array.isArray(res.download_urls) ? res.download_urls : []
    for (const url of urls) {
      const t = detectPanTypeByUrl(url)
      if (!t) continue
      items.push({
        id: `${res.id}-${t}`,
        name: nameByType[t],
        type: t,
        url,
        size,
        speed: "medium",
      })
    }
  }

  if (items.length === 0) {
    for (const res of resources) {
      if (res.download_url?.trim()) {
        items.push({
          id: String(res.id),
          name: res.title || "下载链接",
          type: "quark",
          url: res.download_url,
          size: res.size || "",
          speed: "medium",
        })
      }
    }
  }

  const uniqueByUrl = new Map<string, (typeof items)[number]>()
  for (const it of items) {
    if (!uniqueByUrl.has(it.url)) uniqueByUrl.set(it.url, it)
  }
  const result = Array.from(uniqueByUrl.values())
  if (result.length > 0) result[0].isRecommended = true
  return result
}

async function fetchRelatedGames(current: GameDTO) {
  const cid = typeof current.category_id === "number" ? current.category_id : undefined
  const listRes = await fetchGameList({ page: 1, page_size: 8, category_id: cid })
  return listRes.list
    .filter((g) => g.id !== current.id)
    .slice(0, 4)
    .map(absolutizeGameMediaUrls)
}

export default async function GameDetailPage({
  params,
}: {
  params: { id?: string | string[] } | Promise<{ id?: string | string[] }>
}) {
  const resolvedParams = await Promise.resolve(params)
  const rawId = Array.isArray(resolvedParams?.id) ? resolvedParams.id[0] : resolvedParams?.id
  const id = String(rawId || "").trim()

  if (!/^\d+$/.test(id)) {
    return (
      <div className="min-h-screen bg-background">
        <Header />
        <main className="mx-auto max-w-5xl px-4 py-16 sm:px-6 lg:px-8">
          <div className="rounded-xl border border-border bg-card p-8">
            <h1 className="text-2xl font-semibold text-foreground">游戏详情加载失败</h1>
            <p className="mt-3 text-sm text-muted-foreground">当前路由参数无效，无法解析游戏 ID。</p>
          </div>
        </main>
        <Footer />
      </div>
    )
  }

  let game: GameDTO
  let related: GameDTO[] = []
  try {
    const raw = await fetchGameDetail(id)
    game = absolutizeGameMediaUrls(raw)
    related = await fetchRelatedGames(game)
  } catch (error) {
    return (
      <div className="min-h-screen bg-background">
        <Header />
        <main className="mx-auto max-w-5xl px-4 py-16 sm:px-6 lg:px-8">
          <div className="rounded-xl border border-border bg-card p-8">
            <h1 className="text-2xl font-semibold text-foreground">游戏详情加载失败</h1>
            <p className="mt-3 text-sm text-muted-foreground">
              请确认 Go 后端已启动，且 `NEXT_PUBLIC_API_BASE_URL` 指向正确接口地址。
            </p>
            <p className="mt-2 text-xs text-muted-foreground">
              错误信息：{error instanceof Error ? error.message : "unknown error"}
            </p>
          </div>
        </main>
        <Footer />
      </div>
    )
  }

  const screenshots = (game.gallery?.length ? game.gallery : [game.header_image, game.cover]).filter(Boolean)
  const genres = splitToList(game.genres)
  const platforms = splitToList(game.platforms || game.tags).filter((x) => /pc|windows|mac|linux|ps|xbox|switch/i.test(x))
  const languages = parseLanguages(game.supported_languages)
  const systemRequirements = {
    windows: parseRequirement(game.pc_requirements),
    mac: parseRequirement(game.mac_requirements || game.pc_requirements),
    linux: parseRequirement(game.linux_requirements || game.pc_requirements),
  }

  const resources = Array.isArray(game.resources) ? game.resources : []
  const bestRes = resources[0]
  const downloads = buildDownloadCardDownloads(resources)

  const releaseDate = game.release_date ? formatDateZh(game.release_date) : ""
  const updateDate = bestRes?.updated_at ? formatDateZh(bestRes.updated_at) : formatDateZh(game.updated_at)

  return (
    <div className="min-h-screen bg-background">
      <Header />

      <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="grid gap-8 lg:grid-cols-3">
          <div className="lg:col-span-2 space-y-6">
            {game.video_url ? (
              <VideoPlayer
                videoUrl={game.video_url}
                posterUrl={game.header_image || game.cover || "/images/game-cover.jpg"}
                title="官方预告片"
              />
            ) : (
              <div className="relative aspect-video w-full overflow-hidden rounded-lg bg-secondary">
                <img
                  src={game.header_image || game.cover || "/images/game-cover.jpg"}
                  alt={game.title}
                  className="h-full w-full object-cover"
                />
                <div className="absolute inset-0 bg-gradient-to-t from-background/70 via-transparent to-transparent" />
              </div>
            )}

            {screenshots.length > 0 && <ScreenshotGallery screenshots={screenshots} />}
          </div>

          <div className="space-y-6">
            <GameInfo
              title={game.title}
              developer={game.developers || game.developer || "未知开发商"}
              publisher={game.publishers || "未知发行商"}
              releaseDate={releaseDate || "未知"}
              rating={Number.isFinite(game.rating) ? game.rating : 0}
              reviewCount={game.recommendations_total || 0}
              genres={genres.length ? genres : ["未分类"]}
              platforms={platforms.length ? platforms : ["PC"]}
              languages={languages.length ? languages : ["简体中文"]}
            />

            <DownloadCard
              fileSize={bestRes?.size || game.size || "未知"}
              updateDate={updateDate || "未知"}
              version={bestRes?.version || "最新"}
              downloads={downloads}
            />
          </div>
        </div>

        <section className="mt-12 border-t border-border pt-8">
          <GameDescription
            shortDescription={game.short_description || "暂无简介"}
            fullDescription={stripHtml(game.description || game.reviews || "")}
            features={splitToList(game.tags).slice(0, 10)}
          />
        </section>

        <section className="mt-12 border-t border-border pt-8">
          <SystemRequirements
            windows={systemRequirements.windows}
            mac={systemRequirements.mac}
            linux={systemRequirements.linux}
          />
        </section>

        <section className="mt-12 border-t border-border pt-8">
          <UserReviews />
        </section>

        <section className="mt-12 border-t border-border pt-8">
          <RelatedGames
            games={related.map((g) => ({
              id: String(g.id),
              title: g.title,
              coverUrl: g.cover || g.header_image,
              price: g.price_final ? Math.round(g.price_final / 100) : 0,
              rating: g.rating || 0,
            }))}
          />
        </section>
      </main>

      <Footer />
    </div>
  )
}

