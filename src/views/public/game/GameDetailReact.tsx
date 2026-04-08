import '@semi-ui-styles/semi.min.css'

import {
  IconClock,
  IconLikeHeart,
  IconPlayCircle,
  IconSearch,
  IconThumbUpStroked,
} from '@douyinfe/semi-icons'
import { Button, Card, ConfigProvider, Empty, Input, Modal, Skeleton, Tag, Toast, Typography } from '@douyinfe/semi-ui'
import zh_CN from '@douyinfe/semi-ui/lib/es/locale/source/zh_CN'
import dayjs from 'dayjs'
import { useEffect, useMemo, useState } from 'react'
import { siteSubmissionCreate } from '@/api/netdisk'
import {
  publicGameDetail,
  publicGameList,
  publicGameResourceList,
  type PublicGameItem,
  type PublicGameResource,
} from './api'
import styles from './GameDetail.module.scss'

const { Title, Paragraph, Text } = Typography

type GameDetailReactProps = {
  gameId: string
}

type HeroMedia = {
  type: 'video' | 'image'
  src: string
  poster: string
  title: string
}

type ResourceSort = 'latest' | 'hot'
type RankTab = 'week' | 'month' | 'history'
type ResourceCategoryKey = 'game' | 'mod' | 'trainer' | 'submission'
type PublicGameDetailLike = PublicGameItem & { video_url?: string }

const fallbackGameDetail: PublicGameDetailLike = {
  id: 9527,
  title: '红色沙漠',
  cover: 'https://images.unsplash.com/photo-1542751110-97427bbecf20?auto=format&fit=crop&w=1200&q=80',
  banner: 'https://images.unsplash.com/photo-1511512578047-dfb367046420?auto=format&fit=crop&w=1600&q=80',
  header_image: 'https://images.unsplash.com/photo-1545239351-1141bd82e8a6?auto=format&fit=crop&w=1600&q=80',
  short_description: '以幻想大陆为背景的开放世界动作冒险游戏。',
  description: '在广阔荒原上自由探索，参与骑战与近身战斗，以沉浸式叙事体验一段史诗冒险。',
  website: 'https://store.steampowered.com',
  publishers: 'Pearl Abyss',
  genres: '动作,冒险',
  tags: '开放世界,剧情,单机,高画质',
  release_date: '2026-03-20',
  size: '150GB',
  developer: 'Pearl Abyss',
  type: '动作 / 冒险',
  rating: 4.6,
  steam_score: 46,
  downloads: 17000,
  likes: 34,
  dislikes: 16,
  gallery: [
    'https://images.unsplash.com/photo-1511882150382-421056c89033?auto=format&fit=crop&w=1600&q=80',
    'https://images.unsplash.com/photo-1511919884226-fd3cad34687c?auto=format&fit=crop&w=1600&q=80',
    'https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?auto=format&fit=crop&w=1600&q=80',
    'https://images.unsplash.com/photo-1511512578047-dfb367046420?auto=format&fit=crop&w=1600&q=80',
  ],
  created_at: '2026-03-30T08:00:00Z',
  updated_at: '2026-04-01T12:00:00Z',
  video_url: 'https://www.w3schools.com/html/mov_bbb.mp4',
}

const fallbackResources: PublicGameResource[] = [
  {
    id: 1,
    game_id: 9527,
    title: '【亲测可用】解压即玩 红色沙漠官方中文版',
    resource_type: 'game',
    version: 'V1.01.01 / Build.22560074',
    size: '150GB',
    download_type: '解压即玩',
    pan_type: '迅雷/百度',
    download_url: 'https://pan.xunlei.com/s/mock-red-desert-1\nhttps://pan.baidu.com/s/mock-red-desert-1',
    download_urls: ['https://pan.xunlei.com/s/mock-red-desert-1', 'https://pan.baidu.com/s/mock-red-desert-1'],
    tested: true,
    author: 'BigWolf_老狼',
    publish_date: '2026-04-01 20:28',
  },
  {
    id: 2,
    game_id: 9527,
    title: '【亲测可玩】解压即玩 V1.01.01 数字豪华版 + DLC',
    resource_type: 'game',
    version: 'V1.01.01',
    size: '150GB',
    download_type: '虚拟机',
    pan_type: '夸克',
    download_url: 'https://pan.quark.cn/s/mock-red-desert-2',
    download_urls: ['https://pan.quark.cn/s/mock-red-desert-2'],
    tested: true,
    author: 'BigWolf_老狼',
    publish_date: '2026-04-01 15:16',
  },
  {
    id: 3,
    game_id: 9527,
    title: '【MOD】高清材质包与性能优化补丁',
    resource_type: 'mod',
    version: 'v2.3',
    size: '8.6GB',
    download_type: 'MOD',
    pan_type: '阿里',
    download_url: 'https://www.alipan.com/s/mock-red-desert-mod',
    download_urls: ['https://www.alipan.com/s/mock-red-desert-mod'],
    tested: true,
    author: 'DFAN_ModTeam',
    publish_date: '2026-04-01 09:32',
  },
  {
    id: 4,
    game_id: 9527,
    title: '【修改器】二十项属性修改工具',
    resource_type: 'trainer',
    version: 'Build 20260401',
    size: '120MB',
    download_type: '修改器',
    pan_type: '夸克',
    download_url: 'https://pan.quark.cn/s/mock-red-desert-trainer',
    download_urls: ['https://pan.quark.cn/s/mock-red-desert-trainer'],
    tested: true,
    author: '风灵月影',
    publish_date: '2026-03-31 23:15',
  },
]

const fallbackRankList: PublicGameItem[] = [
  { ...fallbackGameDetail, id: 101, title: '红色沙漠', size: '150GB', rating: 4.6, cover: 'https://picsum.photos/seed/rank-1/120/120' },
  { ...fallbackGameDetail, id: 102, title: 'Slay the Spire 2', size: '4GB', rating: 9.4, cover: 'https://picsum.photos/seed/rank-2/120/120' },
  { ...fallbackGameDetail, id: 103, title: '死亡搁浅 2', size: '150GB', rating: 9.6, cover: 'https://picsum.photos/seed/rank-3/120/120' },
  { ...fallbackGameDetail, id: 104, title: '剑星', size: '75GB', rating: 9.3, cover: 'https://picsum.photos/seed/rank-4/120/120' },
  { ...fallbackGameDetail, id: 105, title: '赛博朋克 2077', size: '70GB', rating: 8.5, cover: 'https://picsum.photos/seed/rank-5/120/120' },
]

const panIconMap: Record<string, string> = {
  '百度': '/baidu.png',
  '夸克': '/quark.png',
  '迅雷': '/xunlei.png',
  '阿里': '/al.png',
  '天翼': '/tainyi.png',
  UC: '/uc.png',
  '移动': '/yidong.png',
}

const buttonToneClassMap: Record<string, string> = {
  '百度': styles.linkButtonBaidu,
  '夸克': styles.linkButtonQuark,
  '迅雷': styles.linkButtonXunlei,
  '阿里': styles.linkButtonAliyun,
  UC: styles.linkButtonUc,
  '天翼': styles.linkButtonTianyi,
  '移动': styles.linkButtonMobile,
  '115': styles.linkButtonCloud115,
  '123': styles.linkButtonPan123,
  '下载': styles.linkButtonDirect,
}

const toNumber = (value?: number, fallback = 0) => {
  const result = Number(value)
  return Number.isFinite(result) ? result : fallback
}

const ensureAbsolute = (value?: string) => {
  const raw = String(value || '').trim()
  if (!raw) return ''
  if (/^https?:\/\//i.test(raw)) return raw
  if (raw.startsWith('//')) return 'https:' + raw
  return raw
}

const cleanList = (value?: string[]) => (Array.isArray(value) ? value.filter((item) => String(item || '').trim()) : [])

const splitByCommonSeparators = (value?: string) =>
  String(value || '')
    .split(/[\n,，/|、\s]+/)
    .map((item) => item.trim())
    .filter(Boolean)

const formatDate = (value?: string, format = 'M月D日 HH:mm') => {
  if (!value) return '-'
  const date = dayjs(value)
  return date.isValid() ? date.format(format) : value
}

const formatSimpleDate = (value?: string) => {
  if (!value) return '-'
  const date = dayjs(value)
  return date.isValid() ? date.format('YYYY/M/D') : value
}

const formatDownloadCount = (value?: number) => {
  const count = toNumber(value)
  return count >= 10000 ? (count / 10000).toFixed(1) + 'W' : String(count)
}

const normalizeResourceLinks = (item: PublicGameResource) => {
  const merged = [
    ...(Array.isArray(item.download_urls) ? item.download_urls : []),
    ...String(item.download_url || '')
      .split(/[\n,，]+/)
      .map((url) => url.trim())
      .filter(Boolean),
  ]
  return Array.from(new Set(merged))
}

const normalizePanLabel = (value?: string) => {
  const raw = String(value || '').trim()
  if (!raw) return ''
  if (raw.includes('百度')) return '百度'
  if (raw.includes('夸克')) return '夸克'
  if (raw.includes('迅雷')) return '迅雷'
  if (raw.includes('阿里')) return '阿里'
  if (raw.includes('UC')) return 'UC'
  if (raw.includes('天翼')) return '天翼'
  if (raw.includes('移动')) return '移动'
  if (raw.includes('115')) return '115'
  if (raw.includes('123')) return '123'
  return raw
}

const getLinkMeta = (link: string) => {
  const lower = link.toLowerCase()
  if (lower.includes('pan.baidu.com')) return { label: '百度', icon: panIconMap['百度'] }
  if (lower.includes('pan.quark.cn')) return { label: '夸克', icon: panIconMap['夸克'] }
  if (lower.includes('pan.xunlei.com')) return { label: '迅雷', icon: panIconMap['迅雷'] }
  if (lower.includes('alipan.com') || lower.includes('aliyundrive.com')) return { label: '阿里', icon: panIconMap['阿里'] }
  if (lower.includes('drive.uc.cn')) return { label: 'UC', icon: panIconMap.UC }
  if (lower.includes('cloud.189.cn')) return { label: '天翼', icon: panIconMap['天翼'] }
  if (lower.includes('caiyun.139.com')) return { label: '移动', icon: panIconMap['移动'] }
  return { label: '下载', icon: '' }
}

const getButtonMeta = (resource: PublicGameResource, link: string, index: number) => {
  const configuredPanTypes = splitByCommonSeparators(resource.pan_type).map((item) => normalizePanLabel(item))
  const configuredLabel = configuredPanTypes[index] || configuredPanTypes[0] || ''
  const detected = getLinkMeta(link)
  const label = configuredLabel || detected.label
  return {
    label,
    icon: panIconMap[label] || detected.icon || '',
    isNetdisk: Boolean(panIconMap[label] || detected.icon),
    toneClass: buttonToneClassMap[label] || styles.linkButtonDirect,
  }
}

const toHeroMedia = (detail: PublicGameDetailLike): HeroMedia[] => {
  const imagePool = [ensureAbsolute(detail.banner), ensureAbsolute(detail.header_image), ensureAbsolute(detail.cover), ...cleanList(detail.gallery).map(ensureAbsolute)].filter(Boolean)
  const uniqueImages = Array.from(new Set(imagePool))
  const items: HeroMedia[] = uniqueImages.map((src, index) => ({
    type: 'image',
    src,
    poster: src,
    title: detail.title + ' 画面 ' + (index + 1),
  }))
  const video = ensureAbsolute(detail.video_url)
  if (video) {
    items.unshift({
      type: 'video',
      src: video,
      poster: uniqueImages[0] || ensureAbsolute(detail.cover) || '',
      title: detail.title + ' 宣传视频',
    })
  }
  return items.length > 0 ? items : [{ type: 'image', src: fallbackGameDetail.cover || '', poster: fallbackGameDetail.cover || '', title: detail.title + ' 默认图' }]
}

const resourceCategoryOptions: Array<{ key: ResourceCategoryKey; label: string; badge: string; emptyTitle: string; emptyDescription: string }> = [
  { key: 'game', label: '本体', badge: 'GAME', emptyTitle: '暂无本体资源', emptyDescription: '当前游戏还没有配置本体下载内容。' },
  { key: 'mod', label: 'MOD', badge: 'MOD', emptyTitle: '暂无 MOD 资源', emptyDescription: '当前游戏还没有可用的 MOD 资源。' },
  { key: 'trainer', label: '修改器', badge: 'TR', emptyTitle: '暂无修改器资源', emptyDescription: '当前游戏还没有可用的修改器资源。' },
  { key: 'submission', label: '用户投稿', badge: 'UP', emptyTitle: '暂无用户投稿', emptyDescription: '这个分类下还没有投稿内容，可以抢先提交。' },
]

const getResourceCategory = (resource: PublicGameResource): ResourceCategoryKey => {
  if (resource.resource_type === 'game' || resource.resource_type === 'mod' || resource.resource_type === 'trainer' || resource.resource_type === 'submission') {
    return resource.resource_type
  }
  const title = String(resource.title || '').toLowerCase()
  const downloadType = String(resource.download_type || '').toLowerCase()
  const author = String(resource.author || '').toLowerCase()
  const combined = [title, downloadType, author].join(' ')

  if (
    combined.includes('投稿') ||
    combined.includes('用户') ||
    combined.includes('玩家') ||
    combined.includes('网友') ||
    combined.includes('submission')
  ) {
    return 'submission'
  }

  if (
    combined.includes('修改器') ||
    combined.includes('trainer') ||
    combined.includes('wemod') ||
    combined.includes('风灵月影') ||
    combined.includes('ce table')
  ) {
    return 'trainer'
  }

  if (
    combined.includes('mod') ||
    combined.includes('补丁') ||
    combined.includes('材质') ||
    combined.includes('地图') ||
    combined.includes('皮肤') ||
    combined.includes('整合包') ||
    combined.includes('存档')
  ) {
    return 'mod'
  }

  return 'game'
}

const GameDetailReact = ({ gameId }: GameDetailReactProps) => {
  const [loading, setLoading] = useState(true)
  const [detail, setDetail] = useState<PublicGameDetailLike | null>(null)
  const [resources, setResources] = useState<PublicGameResource[]>([])
  const [rankGames, setRankGames] = useState<PublicGameItem[]>([])
  const [activeHero, setActiveHero] = useState(0)
  const [videoVisible, setVideoVisible] = useState(false)
  const [currentVideo, setCurrentVideo] = useState('')
  const [resourceSort, setResourceSort] = useState<ResourceSort>('latest')
  const [resourceCategory, setResourceCategory] = useState<ResourceCategoryKey>('game')
  const [rankTab, setRankTab] = useState<RankTab>('week')
  const [introExpanded, setIntroExpanded] = useState(false)
  const [submissionVisible, setSubmissionVisible] = useState(false)
  const [submissionLoading, setSubmissionLoading] = useState(false)
  const [submissionForm, setSubmissionForm] = useState({
    title: '',
    link: '',
    description: '',
    extract_code: '',
    tags: '',
  })

  useEffect(() => {
    let disposed = false

    const load = async () => {
      setLoading(true)
      try {
        const detailRes = await publicGameDetail(gameId)
        const detailData = detailRes.data.code === 200 && detailRes.data.data
          ? (detailRes.data.data as PublicGameDetailLike)
          : { ...fallbackGameDetail, id: Number(gameId || 0) || fallbackGameDetail.id }

        const [resourceRes, rankRes] = await Promise.all([
          publicGameResourceList(gameId).catch(() => null),
          publicGameList({ page: 1, page_size: 10, category_id: detailData.category_id }).catch(() => null),
        ])

        if (disposed) return

        const detailResources = Array.isArray(detailData.resources) ? detailData.resources : []
        const fetchedResources = resourceRes?.data?.code === 200 && Array.isArray(resourceRes.data.data) ? resourceRes.data.data : []
        const nextRankGames = rankRes?.data?.code === 200 && Array.isArray(rankRes.data.data?.list) && rankRes.data.data.list.length > 0
          ? rankRes.data.data.list.filter((item) => String(item.id) !== String(gameId)).slice(0, 10)
          : fallbackRankList

        setDetail(detailData)
        setResources(detailResources.length > 0 ? detailResources : fetchedResources)
        setRankGames(nextRankGames)
      } catch {
        if (disposed) return
        setDetail({ ...fallbackGameDetail, id: Number(gameId || 0) || fallbackGameDetail.id })
        setResources(fallbackResources)
        setRankGames(fallbackRankList)
      } finally {
        if (!disposed) setLoading(false)
      }
    }

    load()
    return () => {
      disposed = true
    }
  }, [gameId])

  const safeDetail = detail || fallbackGameDetail
  const gameIdNumber = Number(safeDetail.id || gameId || 0) || undefined
  const heroMedia = useMemo(() => toHeroMedia(safeDetail), [safeDetail])
  const currentMedia = heroMedia[activeHero] || heroMedia[0]
  const genreTags = splitByCommonSeparators(safeDetail.genres)
  const gameTags = splitByCommonSeparators(safeDetail.tags)

  const sortedResources = useMemo(() => [...resources].sort((a, b) => {
    const aTime = dayjs(a.publish_date || a.created_at).valueOf() || 0
    const bTime = dayjs(b.publish_date || b.created_at).valueOf() || 0
    if (resourceSort === 'hot') {
      return Number(Boolean(b.tested)) - Number(Boolean(a.tested)) || bTime - aTime || b.id - a.id
    }
    return bTime - aTime || b.id - a.id
  }), [resourceSort, resources])

  const categorizedResources = useMemo(() => {
    const result: Record<ResourceCategoryKey, PublicGameResource[]> = {
      game: [],
      mod: [],
      trainer: [],
      submission: [],
    }

    sortedResources.forEach((resource) => {
      result[getResourceCategory(resource)].push(resource)
    })

    return result
  }, [sortedResources])

  const currentResources = categorizedResources[resourceCategory]
  const currentCategory = resourceCategoryOptions.find((item) => item.key === resourceCategory) || resourceCategoryOptions[0]

  const rankedGames = useMemo(() => {
    if (rankTab === 'history') return rankGames
    if (rankTab === 'month') return [...rankGames].reverse()
    return rankGames
  }, [rankGames, rankTab])

  useEffect(() => {
    if (activeHero >= heroMedia.length) setActiveHero(0)
  }, [activeHero, heroMedia.length])

  useEffect(() => {
    if (heroMedia.length <= 1) return undefined
    const timer = window.setInterval(() => {
      setActiveHero((value) => (value + 1) % heroMedia.length)
    }, 5000)
    return () => window.clearInterval(timer)
  }, [heroMedia.length])

  const openHeroVideo = () => {
    const videoItem = heroMedia.find((item) => item.type === 'video')
    if (!videoItem) {
      Toast.info('暂无视频资源')
      return
    }
    setCurrentVideo(videoItem.src)
    setVideoVisible(true)
  }

  const handleSubmissionChange = (field: keyof typeof submissionForm, value: string) => {
    setSubmissionForm((prev) => ({ ...prev, [field]: value }))
  }

  const handleSubmitSubmission = async () => {
    if (!submissionForm.title.trim()) {
      Toast.error('请填写投稿标题')
      return
    }
    if (!submissionForm.link.trim()) {
      Toast.error('请填写下载链接')
      return
    }

    setSubmissionLoading(true)
    try {
      await siteSubmissionCreate({
        title: submissionForm.title.trim(),
        link: submissionForm.link.trim(),
        game_id: gameIdNumber,
        category_id: Number(safeDetail.category_id || 0) || undefined,
        description: [safeDetail.title, submissionForm.description.trim()].filter(Boolean).join('\n'),
        extract_code: submissionForm.extract_code.trim() || undefined,
        tags: [safeDetail.title, '用户投稿', submissionForm.tags.trim()].filter(Boolean).join(','),
      })
      Toast.success('投稿已提交，等待审核')
      setSubmissionVisible(false)
      setSubmissionForm({
        title: '',
        link: '',
        description: '',
        extract_code: '',
        tags: '',
      })
    } catch {
      Toast.error('投稿提交失败，请确认登录状态后重试')
    } finally {
      setSubmissionLoading(false)
    }
  }

  return (
    <ConfigProvider locale={zh_CN}>
      <div className={styles.page}>
        <div className={styles.shell}>
          <div className={styles.breadcrumb}>
            <button type="button" onClick={() => (window.location.href = '/')}>
              首页
            </button>
            <span className={styles.crumbSep}>{'>'}</span>
            <strong>{safeDetail.title}</strong>
          </div>

          {loading ? (
            <div className={styles.loadingState}>
              <Skeleton placeholder={<Skeleton.Image style={{ width: '100%', height: 320 }} />} loading />
              <Skeleton placeholder={<Skeleton.Paragraph rows={10} />} loading />
            </div>
          ) : (
            <div className={styles.mainLayout}>
              <div className={styles.leftColumn}>
                <Card bordered={false} className={styles.summaryCard}>
                  <div className={styles.summaryGrid}>
                    <div className={styles.galleryPanel}>
                      <div className={styles.previewFrame}>
                        <img src={currentMedia?.poster || currentMedia?.src || fallbackGameDetail.cover} alt={currentMedia?.title || safeDetail.title} className={styles.previewImage} />
                        {currentMedia?.type === 'video' && (
                          <button type="button" className={styles.previewPlay} onClick={openHeroVideo}>
                            <IconPlayCircle size="large" />
                            <span>播放视频</span>
                          </button>
                        )}
                        <button type="button" className={styles.navArrow + ' ' + styles.navPrev} onClick={() => setActiveHero((heroMedia.length + activeHero - 1) % heroMedia.length)}>
                          {'‹'}
                        </button>
                        <button type="button" className={styles.navArrow + ' ' + styles.navNext} onClick={() => setActiveHero((activeHero + 1) % heroMedia.length)}>
                          {'›'}
                        </button>
                      </div>

                      <div className={styles.thumbRow}>
                        {heroMedia.slice(0, 5).map((item, index) => (
                          <button key={item.src + '-' + index} type="button" className={styles.thumbButton + ' ' + (activeHero === index ? styles.thumbButtonActive : '')} onClick={() => setActiveHero(index)}>
                            <img src={item.poster || item.src} alt={item.title} className={styles.thumbImage} />
                            {item.type === 'video' && <span className={styles.thumbVideo}>视频</span>}
                          </button>
                        ))}
                      </div>
                    </div>

                    <div className={styles.metaPanel}>
                      <div className={styles.titleRow}>
                        <Title heading={3} className={styles.gameTitle}>{safeDetail.title}</Title>
                        <div className={styles.downloadCount}>{formatDownloadCount(safeDetail.downloads)} 下载</div>
                      </div>

                      <Paragraph className={styles.summaryText} ellipsis={{ rows: 4 }}>
                        {safeDetail.short_description || safeDetail.description || '暂无游戏简介'}
                      </Paragraph>

                      <div className={styles.infoList}>
                        <div className={styles.infoItem}>
                          <span>发布时间</span>
                          <strong>{formatSimpleDate(safeDetail.release_date || safeDetail.created_at)}</strong>
                        </div>
                        <div className={styles.infoItem}>
                          <span>游戏类型</span>
                          <div className={styles.infoBadges}>
                            {genreTags.length > 0 ? genreTags.slice(0, 3).map((tag) => <Tag key={tag}>{tag}</Tag>) : <Tag>{safeDetail.type || '动作'}</Tag>}
                          </div>
                        </div>
                        <div className={styles.infoItem}>
                          <span>开发厂商</span>
                          <strong>{safeDetail.developer || safeDetail.publishers || '-'}</strong>
                        </div>
                        <div className={styles.infoItem}>
                          <span>Steam好评率</span>
                          <div className={styles.scoreBadge}>Steam 推荐度 {toNumber(safeDetail.steam_score, 46)}%</div>
                        </div>
                      </div>

                      <div className={styles.voteRow}>
                        <div className={styles.voteCard}>
                          <span className={styles.voteEmoji}>YES</span>
                          <strong>{toNumber(safeDetail.likes, 34)}</strong>
                        </div>
                        <div className={styles.voteCard}>
                          <span className={styles.voteEmoji}>NO</span>
                          <strong>{toNumber(safeDetail.dislikes, 16)}</strong>
                        </div>
                      </div>
                    </div>
                  </div>
                </Card>

                <Card bordered={false} className={styles.sectionCard}>
                  <div className={styles.sectionHeader}>
                    <div className={styles.sectionTitleWrap}>
                      <div className={styles.sectionHeading}>
                        <span className={styles.sectionBadge + ' ' + styles[`sectionBadge${resourceCategory.charAt(0).toUpperCase()}${resourceCategory.slice(1)}`]}>{
                          currentCategory.badge
                        }</span>
                        <h3 className={styles.sectionTitle}>下载分类</h3>
                      </div>
                      <span className={styles.sectionLine} />
                    </div>
                    <div className={styles.sectionActions}>
                      <button type="button" className={styles.sortButton + ' ' + (resourceSort === 'latest' ? styles.sortButtonActive : '')} onClick={() => setResourceSort('latest')}>
                        最近发布
                      </button>
                      <button type="button" className={styles.sortButton + ' ' + (resourceSort === 'hot' ? styles.sortButtonActive : '')} onClick={() => setResourceSort('hot')}>
                        最热
                      </button>
                      <Button theme="solid" type="primary" onClick={() => setSubmissionVisible(true)}>用户投稿</Button>
                    </div>
                  </div>

                  <div className={styles.categoryTabs}>
                    {resourceCategoryOptions.map((item) => {
                      const count = categorizedResources[item.key].length
                      return (
                        <button
                          key={item.key}
                          type="button"
                          className={
                            styles.categoryTab + ' ' +
                            styles[`categoryTab${item.key.charAt(0).toUpperCase()}${item.key.slice(1)}`] + ' ' +
                            (resourceCategory === item.key ? styles.categoryTabActive : '')
                          }
                          onClick={() => setResourceCategory(item.key)}
                        >
                              <span>{item.label}</span>
                              <em>{count}</em>
                        </button>
                      )
                    })}
                  </div>

                  <div className={styles.resourceList}>
                    {currentResources.length > 0 ? (
                      currentResources.map((resource) => {
                        const links = normalizeResourceLinks(resource)
                        return (
                          <article key={resource.id} className={styles.resourceItem}>
                            <div className={styles.resourceMain}>
                              <span className={styles.resourceMark + ' ' + styles[`resourceMark${resourceCategory.charAt(0).toUpperCase()}${resourceCategory.slice(1)}`]}>{currentCategory.badge}</span>
                              <div className={styles.resourceBody}>
                                <h4 className={styles.resourceTitle}>{resource.title || safeDetail.title}</h4>
                                <div className={styles.resourceMeta}>
                                  {resource.tested && <span className={styles.metaBadgeSuccess}>已测试</span>}
                                  {resource.download_type && <span className={styles.metaBadgeWarn}>{resource.download_type}</span>}
                                  {resource.version && <span className={styles.metaText}>版本: {resource.version}</span>}
                                  {resource.size && <span className={styles.metaText}>文件大小: {resource.size}</span>}
                                </div>
                                <div className={styles.authorRow}>
                                  <span className={styles.authorDot} />
                                  <span>{resource.author || '匿名用户'}</span>
                                </div>
                              </div>
                            </div>

                            <div className={styles.resourceAside}>
                              <div className={styles.resourceLinks}>
                                {links.map((link, index) => {
                                  const meta = getButtonMeta(resource, link, index)
                                  return (
                                    <button key={resource.id + '-' + index} type="button" className={styles.linkButton + ' ' + (meta.isNetdisk ? styles.linkButtonNetdisk : styles.linkButtonDirect) + ' ' + meta.toneClass} onClick={() => window.open(link, '_blank', 'noopener,noreferrer')}>
                                      {meta.icon ? <img src={meta.icon} alt={meta.label} className={styles.linkButtonIcon} /> : null}
                                      <span>{meta.label}</span>
                                    </button>
                                  )
                                })}
                              </div>
                              <Text className={styles.resourceTime}>
                                <IconClock /> {formatDate(resource.publish_date || resource.created_at)}
                              </Text>
                            </div>
                          </article>
                        )
                      })
                    ) : (
                      <Empty title={currentCategory.emptyTitle} description={currentCategory.emptyDescription} />
                    )}
                  </div>
                </Card>

                <Card bordered={false} className={styles.sectionCard}>
                  <div className={styles.sectionHeader}>
                    <div className={styles.sectionTitleWrap}>
                      <h3 className={styles.sectionTitle}>游戏介绍</h3>
                    </div>
                    <button type="button" className={styles.expandButton} onClick={() => setIntroExpanded((value) => !value)}>
                      {introExpanded ? '收起' : '展开'}
                    </button>
                  </div>
                  <div className={styles.introContent + ' ' + (introExpanded ? styles.introExpanded : '')}>
                    <Paragraph className={styles.introText}>{safeDetail.description || '暂无游戏介绍'}</Paragraph>
                    {gameTags.length > 0 && (
                      <div className={styles.tagRow}>
                        {gameTags.map((tag) => (
                          <Tag key={tag} color="cyan">{tag}</Tag>
                        ))}
                      </div>
                    )}
                  </div>
                </Card>
              </div>

              <aside className={styles.rightColumn}>
                <Card bordered={false} className={styles.adCard}>
                  <div className={styles.adHeader}>
                    <span className={styles.adStripe} />
                    <span className={styles.adText}>推荐专区</span>
                    <span className={styles.adStripe} />
                  </div>
                  <img src={ensureAbsolute(safeDetail.banner || safeDetail.cover) || fallbackGameDetail.cover} alt={safeDetail.title} className={styles.adImage} />
                </Card>

                <Card bordered={false} className={styles.rankCard}>
                  <div className={styles.rankHeader}>
                    <button type="button" className={styles.rankTab + ' ' + (rankTab === 'week' ? styles.rankTabActive : '')} onClick={() => setRankTab('week')}>本周热门</button>
                    <button type="button" className={styles.rankTab + ' ' + (rankTab === 'month' ? styles.rankTabActive : '')} onClick={() => setRankTab('month')}>本月热门</button>
                    <button type="button" className={styles.rankTab + ' ' + (rankTab === 'history' ? styles.rankTabActive : '')} onClick={() => setRankTab('history')}>历史热门</button>
                  </div>
                  <div className={styles.rankList}>
                    {rankedGames.map((game, index) => (
                      <button key={game.id} type="button" className={styles.rankItem} onClick={() => (window.location.href = '/games/' + game.id)}>
                        <span className={styles.rankNo + ' ' + (index < 3 ? styles.rankNoHot : '')}>{String(index + 1).padStart(2, '0')}</span>
                        <img src={ensureAbsolute(game.cover || game.banner || game.header_image) || fallbackGameDetail.cover} alt={game.title} className={styles.rankThumb} />
                        <div className={styles.rankBody}>
                          <div className={styles.rankTitle}>{game.title}</div>
                          <div className={styles.rankMeta}>
                            <span className={styles.rankScore}><IconLikeHeart /> {toNumber(game.rating, 8.6).toFixed(1)}</span>
                            <span>{game.size || '150GB'}</span>
                            <span>{splitByCommonSeparators(game.type)[0] || '动作'}</span>
                          </div>
                        </div>
                        {index < 3 && <span className={styles.hotBadge}>HOT</span>}
                      </button>
                    ))}
                  </div>
                </Card>

                <Card bordered={false} className={styles.searchCard}>
                  <Input suffix={<IconSearch />} placeholder="搜索游戏 / MOD / 资源" showClear />
                </Card>

                <Card bordered={false} className={styles.feedCard}>
                  <div className={styles.feedBanner}>最新发布</div>
                  <div className={styles.feedList}>
                    {currentResources.slice(0, 4).map((resource, index) => (
                      <div key={resource.id} className={styles.feedItem}>
                        <div className={styles.feedAvatar}>{index + 1}</div>
                        <div className={styles.feedBody}>
                          <div className={styles.feedName}>{resource.author || '匿名用户'}</div>
                          <div className={styles.feedText}>{resource.title}</div>
                        </div>
                        <div className={styles.feedLike}>
                          <IconThumbUpStroked />
                          <span>{index}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                </Card>
              </aside>
            </div>
          )}
        </div>

        <Modal title="视频预览" visible={videoVisible} onCancel={() => setVideoVisible(false)} footer={null} width={960}>
          {currentVideo ? <video className={styles.videoPlayer} src={currentVideo} controls autoPlay playsInline /> : <Empty title="暂无视频" />}
        </Modal>
        <Modal
          title="用户投稿"
          visible={submissionVisible}
          onCancel={() => setSubmissionVisible(false)}
          onOk={handleSubmitSubmission}
          okText="提交投稿"
          confirmLoading={submissionLoading}
          className={styles.submitModal}
        >
          <div className={styles.submitForm}>
            <div className={styles.submitTarget}>
              <strong>{safeDetail.title}</strong>
              <span>审核通过后归档到 用户投稿</span>
            </div>
            <div className={styles.submitField}>
              <span>投稿标题</span>
              <Input value={submissionForm.title} onChange={(value) => handleSubmissionChange('title', value)} placeholder="例如：红色沙漠 高清材质 MOD" />
            </div>
            <div className={styles.submitField}>
              <span>下载链接</span>
              <Input value={submissionForm.link} onChange={(value) => handleSubmissionChange('link', value)} placeholder="填写网盘或直链地址" />
            </div>
            <div className={styles.submitFieldGrid}>
              <div className={styles.submitField}>
                <span>提取码</span>
                <Input value={submissionForm.extract_code} onChange={(value) => handleSubmissionChange('extract_code', value)} placeholder="有的话再填" />
              </div>
              <div className={styles.submitField}>
                <span>标签</span>
                <Input value={submissionForm.tags} onChange={(value) => handleSubmissionChange('tags', value)} placeholder="MOD, 修改器, 优化补丁" />
              </div>
            </div>
            <div className={styles.submitField}>
              <span>补充说明</span>
              <Input
                value={submissionForm.description}
                onChange={(value) => handleSubmissionChange('description', value)}
                placeholder="补充版本、适配说明、安装方法"
                maxLength={300}
                showClear
              />
            </div>
            <p className={styles.submitHint}>投稿会自动绑定当前游戏，审核通过后会进入该详情页的“用户投稿”分类。</p>
          </div>
        </Modal>
      </div>
    </ConfigProvider>
  )
}

export default GameDetailReact
