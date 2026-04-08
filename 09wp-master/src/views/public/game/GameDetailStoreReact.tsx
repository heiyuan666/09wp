import { useEffect, useMemo, useRef, useState, type MouseEvent } from 'react'
import {
  Play,
  Pause,
  Volume2,
  VolumeX,
  Maximize,
  SkipBack,
  SkipForward,
  X,
  ChevronLeft,
  ChevronRight,
  Gamepad2,
} from 'lucide-react'

import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

import type { PublicGameItem, PublicGameResource } from './api'
import { publicGameDetail, publicGameResourceList } from './api'
import { siteMySubmissions, siteSubmissionCreate } from '@/api/netdisk'

type ToastType = 'success' | 'warning' | 'error'
type ToastItem = { id: string; message: string; type: ToastType }

type ToastContext = {
  push: (message: string, type?: ToastType) => void
}

type Props = {
  gameId: number
}

function formatDate(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return String(value)
  return date.toISOString().slice(0, 10)
}

function detectPanLabel(link: string) {
  const url = String(link).toLowerCase()
  if (url.includes('pan.quark.cn')) return '夸克网盘'
  if (url.includes('pan.baidu.com')) return '百度网盘'
  if (url.includes('pan.xunlei.com')) return '迅雷网盘'
  if (url.includes('aliyundrive.com') || url.includes('alipan.com')) return '阿里云盘'
  if (url.includes('cloud.189.cn')) return '天翼云盘'
  if (url.includes('drive.uc.cn') || url.includes('drive-h.uc.cn')) return 'UC 网盘'
  if (url.includes('115.com')) return '115 网盘'
  if (url.includes('123pan') || url.includes('123684.com') || url.includes('123685.com')) return '123 云盘'
  return '下载链接'
}

function getDownloadLabel(link: string, index: number, total: number) {
  const label = detectPanLabel(link)
  return total > 1 ? `${label} ${index + 1}` : label
}

function normalizeLinks(item: PublicGameResource) {
  const fromArray =
    Array.isArray(item.download_urls) ? item.download_urls.map((x) => String(x).trim()).filter(Boolean) : []
  if (fromArray.length > 0) return fromArray

  const raw = String(item.download_url || '').trim()
  if (!raw) return []

  return raw
    .split(/[\n\r\t ,;，；]+/)
    .map((x) => x.trim())
    .filter((x) => /^https?:\/\//i.test(x))
}

function VideoPlayer({ videoUrl, posterUrl, title, onToast }: { videoUrl: string; posterUrl: string; title?: string; onToast: ToastContext }) {
  const [isPlaying, setIsPlaying] = useState(false)
  const [isMuted, setIsMuted] = useState(false)
  const [progress, setProgress] = useState(0)
  const [showControls, setShowControls] = useState(true)
  const videoRef = useRef<HTMLVideoElement | null>(null)

  const togglePlay = () => {
    const v = videoRef.current
    if (!v) return
    if (isPlaying) {
      v.pause()
    } else {
      void v.play().catch(() => onToast.push('视频播放失败，请稍后重试', 'error'))
    }
    setIsPlaying(!isPlaying)
  }

  const toggleMute = () => {
    const v = videoRef.current
    if (!v) return
    v.muted = !isMuted
    setIsMuted(!isMuted)
  }

  const handleTimeUpdate = () => {
    const v = videoRef.current
    if (!v) return
    if (!v.duration || Number.isNaN(v.duration)) return
    const p = (v.currentTime / v.duration) * 100
    setProgress(p)
  }

  const handleProgressClick = (e: MouseEvent<HTMLDivElement>) => {
    const v = videoRef.current
    if (!v) return
    if (!v.duration || Number.isNaN(v.duration)) return
    const rect = e.currentTarget.getBoundingClientRect()
    const pos = (e.clientX - rect.left) / rect.width
    v.currentTime = pos * v.duration
  }

  const handleFullscreen = () => {
    const v = videoRef.current
    if (!v) return
    if (document.fullscreenElement) document.exitFullscreen()
    else void v.requestFullscreen().catch(() => onToast.push('全屏失败', 'error'))
  }

  const skip = (seconds: number) => {
    const v = videoRef.current
    if (!v) return
    v.currentTime += seconds
  }

  return (
    <div
      className="relative aspect-video w-full overflow-hidden rounded-lg bg-secondary group"
      onMouseEnter={() => setShowControls(true)}
      onMouseLeave={() => setShowControls(isPlaying ? false : true)}
    >
      <video
        ref={videoRef}
        src={videoUrl}
        poster={posterUrl}
        className="h-full w-full object-cover"
        onTimeUpdate={handleTimeUpdate}
        onClick={togglePlay}
        onEnded={() => setIsPlaying(false)}
        playsInline
      />

      {!isPlaying ? (
        <div
          className="absolute inset-0 flex items-center justify-center bg-background/40 cursor-pointer"
          onClick={togglePlay}
          role="button"
          tabIndex={0}
        >
          <div className="flex h-20 w-20 items-center justify-center rounded-full bg-primary/90 transition-transform hover:scale-110">
            <Play className="h-10 w-10 text-primary-foreground" fill="currentColor" />
          </div>
        </div>
      ) : null}

      <div
        className={cn(
          'absolute bottom-0 left-0 right-0 bg-gradient-to-t from-background/90 to-transparent p-4 transition-opacity duration-300',
          showControls || !isPlaying ? 'opacity-100' : 'opacity-0',
        )}
      >
        {title ? <p className="mb-2 text-sm font-medium text-foreground">{title}</p> : null}

        <div className="mb-3 h-1 w-full cursor-pointer rounded-full bg-muted" onClick={handleProgressClick}>
          <div className="h-full rounded-full bg-primary transition-all" style={{ width: `${progress}%` }} />
        </div>

        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <button type="button" onClick={() => skip(-10)} className="text-foreground/70 hover:text-foreground transition-colors">
              <SkipBack className="h-5 w-5" />
            </button>
            <button
              type="button"
              onClick={togglePlay}
              className="flex h-10 w-10 items-center justify-center rounded-full bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
            >
              {isPlaying ? <Pause className="h-5 w-5" /> : <Play className="h-5 w-5 ml-0.5" />}
            </button>
            <button type="button" onClick={() => skip(10)} className="text-foreground/70 hover:text-foreground transition-colors">
              <SkipForward className="h-5 w-5" />
            </button>
          </div>

          <div className="flex items-center gap-3">
            <button type="button" onClick={toggleMute} className="text-foreground/70 hover:text-foreground transition-colors">
              {isMuted ? <VolumeX className="h-5 w-5" /> : <Volume2 className="h-5 w-5" />}
            </button>
            <button type="button" onClick={handleFullscreen} className="text-foreground/70 hover:text-foreground transition-colors">
              <Maximize className="h-5 w-5" />
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

function ScreenshotGallery({ screenshots, onToast }: { screenshots: string[]; onToast: ToastContext }) {
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [isLightboxOpen, setIsLightboxOpen] = useState(false)

  useEffect(() => {
    setSelectedIndex(0)
  }, [screenshots.map((s) => s).join('|')])

  const goToNext = () => setSelectedIndex((prev) => (prev + 1) % screenshots.length)
  const goToPrev = () => setSelectedIndex((prev) => (prev - 1 + screenshots.length) % screenshots.length)

  if (!screenshots.length) return null

  return (
    <>
      <div className="space-y-4">
        <div
          className="relative aspect-video w-full overflow-hidden rounded-lg cursor-pointer group"
          onClick={() => setIsLightboxOpen(true)}
          role="button"
          tabIndex={0}
          onKeyDown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') setIsLightboxOpen(true)
          }}
        >
          <img src={screenshots[selectedIndex]} alt={`游戏截图 ${selectedIndex + 1}`} className="object-cover w-full h-full" />
          <div className="absolute inset-0 bg-background/20 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
            <span className="text-foreground text-sm font-medium">点击放大</span>
          </div>
        </div>

        <div className="flex gap-2 overflow-x-auto pb-2 scrollbar-hide">
          {screenshots.map((screenshot, index) => (
            <button
              key={index}
              type="button"
              onClick={() => setSelectedIndex(index)}
              className={cn(
                'relative h-16 w-28 flex-shrink-0 overflow-hidden rounded-md transition-all',
                selectedIndex === index ? 'ring-2 ring-primary' : 'opacity-60 hover:opacity-100',
              )}
            >
              <img src={screenshot} alt={`缩略图 ${index + 1}`} className="object-cover w-full h-full" />
            </button>
          ))}
        </div>
      </div>

      {isLightboxOpen ? (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-background/95"
          onClick={() => setIsLightboxOpen(false)}
          role="dialog"
          aria-modal="true"
        >
          <button
            type="button"
            className="absolute right-4 top-4 text-foreground/70 hover:text-foreground transition-colors"
            onClick={(e) => {
              e.stopPropagation()
              setIsLightboxOpen(false)
            }}
            aria-label="关闭"
          >
            <X className="h-8 w-8" />
          </button>

          <button
            type="button"
            className="absolute left-4 top-1/2 -translate-y-1/2 flex h-12 w-12 items-center justify-center rounded-full bg-secondary text-foreground hover:bg-secondary/80 transition-colors"
            onClick={(e) => {
              e.stopPropagation()
              goToPrev()
            }}
            aria-label="上一张"
          >
            <ChevronLeft className="h-6 w-6" />
          </button>

          <div
            className="relative h-[80vh] w-[90vw] max-w-6xl"
            onClick={(e) => e.stopPropagation()}
          >
            <img src={screenshots[selectedIndex]} alt={`游戏截图 ${selectedIndex + 1}`} className="object-contain w-full h-full" />
          </div>

          <button
            type="button"
            className="absolute right-4 top-1/2 -translate-y-1/2 flex h-12 w-12 items-center justify-center rounded-full bg-secondary text-foreground hover:bg-secondary/80 transition-colors"
            onClick={(e) => {
              e.stopPropagation()
              goToNext()
            }}
            aria-label="下一张"
          >
            <ChevronRight className="h-6 w-6" />
          </button>

          <div className="absolute bottom-4 left-1/2 -translate-x-1/2 text-foreground/70">
            {selectedIndex + 1} / {screenshots.length}
          </div>
        </div>
      ) : null}
    </>
  )
}

export default function GameDetailStoreReact({ gameId }: Props) {
  const [loading, setLoading] = useState(true)
  const [game, setGame] = useState<PublicGameItem | null>(null)
  const [resources, setResources] = useState<PublicGameResource[]>([])
  const [mySubmissions, setMySubmissions] = useState<any[]>([])

  const [toasts, setToasts] = useState<ToastItem[]>([])
  const toastPush: ToastContext['push'] = (message, type = 'success') => {
    const id = `${Date.now()}_${Math.random().toString(16).slice(2)}`
    setToasts((prev) => [{ id, message, type }, ...prev].slice(0, 3))
    window.setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id))
    }, 2500)
  }

  const canShowNotFound = !loading && (!game || !game.title)

  const coverUrl = useMemo(() => {
    if (!game) return ''
    return String(game.cover || game.header_image || '').trim()
  }, [game])

  const heroBg = useMemo(() => {
    if (!game) return ''
    return String(game.banner || game.cover || game.header_image || '').trim()
  }, [game])

  const videoPoster = useMemo(() => {
    if (!game) return ''
    return String(game.banner || game.header_image || game.cover || '').trim()
  }, [game])

  const videoUrl = useMemo(() => {
    if (!game) return ''
    return String(game.video_url || '').trim()
  }, [game])

  const galleryUrls = useMemo(() => {
    const raw = game?.gallery
    if (!Array.isArray(raw)) return []
    return raw.map((x) => String(x).trim()).filter(Boolean)
  }, [game])

  const descriptionHtml = useMemo(() => {
    const d = game?.description
    return d ? String(d) : ''
  }, [game])

  const mergedResources = useMemo(() => {
    const fromDetail = game?.resources
    if (Array.isArray(fromDetail) && fromDetail.length) return fromDetail
    return resources
  }, [game, resources])

  const resourceGroups = useMemo(() => {
    const grouped = new Map<string, Array<PublicGameResource & { link_list: string[] }>>()
    const resourceLabelMap: Record<string, string> = {
      game: '本体下载',
      mod: 'Mod 下载',
      trainer: '修改器下载',
      submission: '玩家投稿',
    }

    for (const item of mergedResources as PublicGameResource[]) {
      const key = String(item.resource_type || 'game').trim() || 'game'
      const linkList = normalizeLinks(item)
      if (linkList.length === 0) continue
      if (!grouped.has(key)) grouped.set(key, [])
      grouped.get(key)?.push({ ...item, link_list: linkList })
    }

    return Array.from(grouped.entries()).map(([key, items]) => ({
      key,
      label: resourceLabelMap[key] || key,
      items,
    }))
  }, [mergedResources])

  const [submissionForm, setSubmissionForm] = useState({
    title: '',
    link: '',
    extract_code: '',
    tags: '',
    description: '',
  })
  const [submittingSubmission, setSubmittingSubmission] = useState(false)

  const userToken = useMemo(() => {
    if (typeof window === 'undefined') return ''
    return localStorage.getItem('user_token') || ''
  }, [])

  const myGameSubmissions = useMemo(() => {
    const gid = Number(gameId || 0)
    return mySubmissions.filter((item) => Number(item.game_id || 0) === gid)
  }, [mySubmissions, gameId])

  const goList = () => {
    window.location.href = '/games'
  }
  const goLogin = () => {
    window.location.href = '/login'
  }

  const loadMySubmissions = async () => {
    if (typeof window === 'undefined') {
      setMySubmissions([])
      return
    }
    if (!userToken) {
      setMySubmissions([])
      return
    }
    try {
      const { data: res } = await siteMySubmissions()
      if (res.code === 200 && Array.isArray(res.data)) setMySubmissions(res.data)
      else setMySubmissions([])
    } catch {
      setMySubmissions([])
    }
  }

  const submitGameResource = async () => {
    const title = String(submissionForm.title || '').trim()
    const link = String(submissionForm.link || '').trim()
    if (!title) return toastPush('请填写资源标题', 'warning')
    if (!link) return toastPush('请填写下载链接', 'warning')
    if (!gameId || !Number.isFinite(gameId)) return toastPush('当前游戏 ID 无效', 'warning')

    setSubmittingSubmission(true)
    try {
      const { data: res } = await siteSubmissionCreate({
        title,
        link,
        game_id: gameId,
        extract_code: String(submissionForm.extract_code || '').trim(),
        tags: String(submissionForm.tags || '').trim(),
        description: String(submissionForm.description || '').trim(),
      })

      if (res.code !== 200) return
      toastPush('提交成功，等待审核', 'success')
      setSubmissionForm({ title: '', link: '', extract_code: '', tags: '', description: '' })
      await loadMySubmissions()
    } catch {
      toastPush('提交失败，请稍后重试', 'error')
    } finally {
      setSubmittingSubmission(false)
    }
  }

  useEffect(() => {
    let disposed = false

    const run = async () => {
      if (!gameId || !Number.isFinite(gameId) || gameId <= 0) {
        setGame(null)
        setResources([])
        setLoading(false)
        return
      }

      setLoading(true)
      try {
        const [dRes, rRes] = await Promise.all([publicGameDetail(gameId), publicGameResourceList(gameId)])
        if (disposed) return

        if (dRes.data.code !== 200 || !dRes.data.data) {
          setGame(null)
          setResources([])
          return
        }

        setGame(dRes.data.data)
        setResources(Array.isArray(rRes.data.data) && rRes.data.code === 200 ? rRes.data.data : [])
        await loadMySubmissions()
      } finally {
        if (!disposed) setLoading(false)
      }
    }

    void run()
    return () => {
      disposed = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [gameId])

  useEffect(() => {
    if (!loading) void loadMySubmissions()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loading])

  const toastByType = (t: ToastType) => {
    if (t === 'success') return 'bg-green-500/15 text-green-700 border-green-500/25'
    if (t === 'warning') return 'bg-amber-500/15 text-amber-700 border-amber-500/25'
    return 'bg-red-500/15 text-red-700 border-red-500/25'
  }

  const copyLink = async (link: string) => {
    try {
      await navigator.clipboard.writeText(link)
      toastPush('链接已复制', 'success')
    } catch {
      toastPush('复制失败，请手动复制', 'error')
    }
  }

  return (
    <div className="gamestore-scope min-h-screen bg-background text-foreground">
      {/* 顶部栏：简化为返回 + 品牌 */}
      <header className="sticky top-0 z-40 bg-background/80 backdrop-blur-lg border-b border-border">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-3">
              <button type="button" className="text-muted-foreground hover:text-foreground font-semibold" onClick={goList}>
                ← 返回
              </button>
              <div className="flex items-center gap-2">
                <Gamepad2 className="h-6 w-6 text-primary" />
                <span className="font-bold text-foreground">GameStore</span>
              </div>
            </div>
            <div className="text-muted-foreground text-sm">{formatDate(game?.release_date)}</div>
          </div>
        </div>
      </header>

      {loading ? (
        <div className="max-w-7xl mx-auto px-4 py-10 text-muted-foreground">加载中...</div>
      ) : canShowNotFound ? (
        <div className="max-w-7xl mx-auto px-4 py-10 text-muted-foreground">未找到该游戏或已下架</div>
      ) : (
        <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
          <div
            className="rounded-2xl border border-border bg-card p-6 mb-8"
            style={
              heroBg
                ? {
                    backgroundImage: `linear-gradient(180deg, rgba(0,0,0,.55), rgba(15,23,42,.92)), url(${heroBg})`,
                  }
                : undefined
            }
          >
            <div className="grid gap-8 lg:grid-cols-3">
              <div className="lg:col-span-2">
                <div className="flex items-center gap-6 mb-6">
                  {coverUrl ? <img src={coverUrl} alt={game?.title || ''} className="w-36 max-w-full rounded-lg shadow" /> : null}
                  <div className="min-w-0">
                    <h1 className="text-3xl md:text-4xl font-bold text-foreground line-clamp-2">{game?.title}</h1>
                    <p className="mt-2 text-muted-foreground">
                      由 <span className="text-foreground">{game?.developer || '—'}</span> 开发 ·{' '}
                      <span className="text-foreground">{game?.publishers || '—'}</span> 发行
                    </p>
                    <div className="mt-3 flex flex-wrap gap-2">
                      {game?.type ? (
                        <span className="rounded-full bg-primary/20 px-3 py-1 text-xs font-medium text-primary">{game.type}</span>
                      ) : null}
                      {game?.release_date ? (
                        <span className="rounded-full bg-secondary px-3 py-1 text-xs font-medium text-muted-foreground">
                          {formatDate(game.release_date)}
                        </span>
                      ) : null}
                      {game?.price_text ? (
                        <span className="rounded-full bg-primary px-3 py-1 text-xs font-medium text-primary-foreground">
                          {game.price_text}
                        </span>
                      ) : null}
                    </div>
                    {game?.short_description ? <p className="mt-4 text-muted-foreground">{game.short_description}</p> : null}
                  </div>
                </div>

                <div className="space-y-6">
                  {videoUrl ? (
                    <VideoPlayer
                      videoUrl={videoUrl}
                      posterUrl={videoPoster}
                      title="官方预告片"
                      onToast={{ push: toastPush }}
                    />
                  ) : null}

                  <ScreenshotGallery screenshots={galleryUrls} onToast={{ push: toastPush }} />
                </div>
              </div>

              <aside className="space-y-6">
                <div className="rounded-xl bg-card p-5 border border-border">
                  <h3 className="text-xl font-semibold mb-4">游戏信息</h3>
                  <div className="space-y-3 text-sm text-muted-foreground">
                    <div className="flex items-center justify-between">
                      <span>类型</span>
                      <span className="text-foreground font-medium">{game?.type || '—'}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span>发行</span>
                      <span className="text-foreground font-medium">{formatDate(game?.release_date)}</span>
                    </div>
                    {game?.genres ? (
                      <div className="flex items-center justify-between">
                        <span>类型</span>
                        <span className="text-foreground font-medium line-clamp-1">{game.genres}</span>
                      </div>
                    ) : null}
                    {game?.tags ? (
                      <div className="flex items-center justify-between">
                        <span>标签</span>
                        <span className="text-foreground font-medium line-clamp-1">{game.tags}</span>
                      </div>
                    ) : null}
                    <div className="flex items-center justify-between">
                      <span>下载热度</span>
                      <span className="text-foreground font-bold">{game?.downloads || 0}</span>
                    </div>
                  </div>
                </div>

                <div className="rounded-xl bg-card p-5 border border-border">
                  <h3 className="text-xl font-semibold mb-4">价格/购买</h3>
                  <div className="text-foreground font-bold text-2xl mb-3">{game?.price_text || '—'}</div>
                  <Button
                    className="w-full"
                    onClick={() => toastPush('此页面为资源聚合展示，暂不提供真实购买操作', 'warning')}
                  >
                    立即购买（展示）
                  </Button>
                  {game?.website ? (
                    <a href={game.website} target="_blank" rel="noreferrer" className="block mt-3 text-sm text-primary hover:underline">
                      打开官网
                    </a>
                  ) : null}
                </div>
              </aside>
            </div>
          </div>

          {/* 介绍 */}
          {descriptionHtml ? (
            <section className="mt-8 border-t border-border pt-8">
              <div className="rounded-xl bg-card border border-border p-6">
                <h3 className="text-xl font-semibold mb-4">游戏介绍</h3>
                <div className="text-sm text-muted-foreground leading-relaxed" dangerouslySetInnerHTML={{ __html: descriptionHtml }} />
              </div>
            </section>
          ) : null}

          {/* 下载资源 */}
          {resourceGroups.length > 0 ? (
            <section className="mt-10">
              <div className="space-y-6">
                <div className="rounded-xl bg-card border border-border p-6">
                  <h3 className="text-xl font-semibold mb-4">下载资源</h3>
                  <div className="space-y-6">
                    {resourceGroups.map((group) => (
                      <div key={group.key}>
                        <div className="flex items-center justify-between mb-4">
                          <h4 className="font-semibold text-foreground">{group.label}</h4>
                          <span className="text-sm text-muted-foreground">{group.items.length} 个资源</span>
                        </div>
                        <div className="space-y-4">
                          {group.items.map((item) => (
                            <div key={item.id} className="rounded-lg border border-border bg-background p-4">
                              <div className="flex items-start justify-between gap-4 flex-wrap">
                                <div className="min-w-0">
                                  <div className="font-semibold text-foreground">{item.title}</div>
                                  {item.tested ? <div className="mt-2 text-xs text-green-600 font-bold">已测试</div> : null}
                                </div>

                                <div className="flex flex-wrap gap-2 justify-end">
                                  {item.version ? (
                                    <span className="rounded-full bg-secondary px-3 py-1 text-xs text-muted-foreground">版本 {item.version}</span>
                                  ) : null}
                                  {item.size ? (
                                    <span className="rounded-full bg-secondary px-3 py-1 text-xs text-muted-foreground">大小 {item.size}</span>
                                  ) : null}
                                  {item.pan_type ? (
                                    <span className="rounded-full bg-secondary px-3 py-1 text-xs text-muted-foreground">平台 {item.pan_type}</span>
                                  ) : null}
                                  {item.download_type ? (
                                    <span className="rounded-full bg-secondary px-3 py-1 text-xs text-muted-foreground">类型 {item.download_type}</span>
                                  ) : null}
                                  {item.author ? (
                                    <span className="rounded-full bg-secondary px-3 py-1 text-xs text-muted-foreground">作者 {item.author}</span>
                                  ) : null}
                                  {item.publish_date ? (
                                    <span className="rounded-full bg-secondary px-3 py-1 text-xs text-muted-foreground">时间 {formatDate(item.publish_date)}</span>
                                  ) : null}
                                </div>
                              </div>

                              <div className="mt-4 flex flex-wrap gap-3 items-center">
                                {item.link_list.map((u, idx) => (
                                  <div key={`${item.id}-${idx}`} className="flex items-center gap-2 flex-wrap">
                                    <a
                                      href={u}
                                      target="_blank"
                                      rel="noreferrer"
                                      className="rounded-full bg-primary px-4 py-2 text-xs font-bold text-primary-foreground hover:bg-primary/90 transition-colors"
                                    >
                                      {getDownloadLabel(u, idx, item.link_list.length)}
                                    </a>
                                    <Button size="sm" variant="secondary" onClick={() => copyLink(u)}>
                                      复制链接
                                    </Button>
                                  </div>
                                ))}
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </section>
          ) : null}

          {/* 玩家投稿 */}
          <section className="mt-10 mb-10">
            <div className="rounded-xl bg-card border border-border p-6">
              <h3 className="text-xl font-semibold mb-4">用户提交资源</h3>

              {userToken ? (
                <div className="space-y-8">
                  <div className="rounded-lg border border-border bg-background p-5">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <label className="text-sm font-semibold text-muted-foreground">资源标题</label>
                        <input
                          className="w-full rounded-lg border border-border bg-background px-4 py-2 text-foreground outline-none"
                          value={submissionForm.title}
                          onChange={(e) => setSubmissionForm((p) => ({ ...p, title: e.target.value }))}
                          placeholder="例如：Dota 2 夸克网盘整合包"
                        />
                      </div>

                      <div className="space-y-2">
                        <label className="text-sm font-semibold text-muted-foreground">下载链接</label>
                        <input
                          className="w-full rounded-lg border border-border bg-background px-4 py-2 text-foreground outline-none"
                          value={submissionForm.link}
                          onChange={(e) => setSubmissionForm((p) => ({ ...p, link: e.target.value }))}
                          placeholder="粘贴网盘分享链接，支持夸克 / 百度 / 阿里 / 迅雷等"
                        />
                      </div>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
                      <div className="space-y-2">
                        <label className="text-sm font-semibold text-muted-foreground">提取码</label>
                        <input
                          className="w-full rounded-lg border border-border bg-background px-4 py-2 text-foreground outline-none"
                          value={submissionForm.extract_code}
                          onChange={(e) => setSubmissionForm((p) => ({ ...p, extract_code: e.target.value }))}
                          placeholder="没有可不填"
                        />
                      </div>
                      <div className="space-y-2">
                        <label className="text-sm font-semibold text-muted-foreground">标签</label>
                        <input
                          className="w-full rounded-lg border border-border bg-background px-4 py-2 text-foreground outline-none"
                          value={submissionForm.tags}
                          onChange={(e) => setSubmissionForm((p) => ({ ...p, tags: e.target.value }))}
                          placeholder="如：补丁,整合包,教程"
                        />
                      </div>
                    </div>

                    <div className="space-y-2 mt-4">
                      <label className="text-sm font-semibold text-muted-foreground">资源说明</label>
                      <textarea
                        className="w-full rounded-lg border border-border bg-background px-4 py-2 text-foreground outline-none"
                        rows={4}
                        value={submissionForm.description}
                        onChange={(e) => setSubmissionForm((p) => ({ ...p, description: e.target.value }))}
                        placeholder="可以补充版本、适用说明、安装方式等信息"
                      />
                    </div>

                    <div className="flex gap-3 mt-5 flex-wrap">
                      <Button onClick={() => void submitGameResource()} disabled={submittingSubmission}>
                        {submittingSubmission ? '提交中...' : '提交资源'}
                      </Button>
                      <Button
                        variant="secondary"
                        disabled={submittingSubmission}
                        onClick={() =>
                          setSubmissionForm({
                            title: '',
                            link: '',
                            extract_code: '',
                            tags: '',
                            description: '',
                          })
                        }
                      >
                        清空
                      </Button>
                    </div>

                    <div className="mt-3 text-sm text-muted-foreground">
                      提交后会进入审核；审核通过后，会自动出现在当前游戏的资源列表里。
                    </div>
                  </div>

                  <div className="rounded-lg border border-border bg-background p-5">
                    <div className="text-lg font-semibold mb-4">我的投稿记录</div>
                    {myGameSubmissions.length > 0 ? (
                      <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                          <thead>
                            <tr className="text-left text-muted-foreground">
                              <th className="py-2 px-3">标题</th>
                              <th className="py-2 px-3" style={{ width: 120 }}>
                                状态
                              </th>
                              <th className="py-2 px-3">备注</th>
                            </tr>
                          </thead>
                          <tbody>
                            {myGameSubmissions.map((row) => (
                              <tr key={row.id} className="border-t border-border">
                                <td className="py-3 px-3">{row.title}</td>
                                <td className="py-3 px-3">
                                  {row.status === 'pending' ? (
                                    <span className="rounded-full border border-amber-500/25 bg-amber-500/10 px-3 py-1 text-amber-700 font-bold">
                                      待审核
                                    </span>
                                  ) : row.status === 'approved' ? (
                                    <span className="rounded-full border border-green-500/25 bg-green-500/10 px-3 py-1 text-green-700 font-bold">
                                      已通过
                                    </span>
                                  ) : (
                                    <span className="rounded-full border border-red-500/25 bg-red-500/10 px-3 py-1 text-red-700 font-bold">
                                      已驳回
                                    </span>
                                  )}
                                </td>
                                <td className="py-3 px-3 text-muted-foreground">{row.review_msg}</td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    ) : (
                      <div className="text-muted-foreground">你还没有为这个游戏提交过资源</div>
                    )}
                  </div>
                </div>
              ) : (
                <div className="text-center py-8">
                  <div className="text-muted-foreground mb-4">登录后可为这个游戏提交资源</div>
                  <Button onClick={goLogin}>去登录</Button>
                </div>
              )}
            </div>
          </section>
        </main>
      )}

      {/* Toasts */}
      <div className="fixed right-4 bottom-4 z-[60] space-y-2">
        {toasts.map((t) => (
          <div key={t.id} className={cn('rounded-lg border px-4 py-3 text-sm shadow-sm', toastByType(t.type))}>
            {t.message}
          </div>
        ))}
      </div>
    </div>
  )
}

