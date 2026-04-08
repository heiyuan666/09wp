import { useEffect, useMemo, useState } from 'react'
import { Modal, Toast } from '@douyinfe/semi-ui'
import ReactPlayer from 'react-player'
import type { PublicGameItem, PublicGameResource } from './api'
import { publicGameDetail, publicGameResourceList } from './api'
import { siteMySubmissions, siteSubmissionCreate } from '@/api/netdisk'
import './GameDetailReact.scss'

type Props = {
  gameId: number
}

export default function GameDetailReact({ gameId }: Props) {
  const [loading, setLoading] = useState(true)
  const [game, setGame] = useState<PublicGameItem | null>(null)
  const [resources, setResources] = useState<PublicGameResource[]>([])

  const [videoStarted, setVideoStarted] = useState(false)

  const [galleryViewerVisible, setGalleryViewerVisible] = useState(false)
  const [galleryViewerUrl, setGalleryViewerUrl] = useState('')

  const [galleryIndex, setGalleryIndex] = useState(0)

  const [submissionForm, setSubmissionForm] = useState({
    title: '',
    link: '',
    extract_code: '',
    tags: '',
    description: '',
  })
  const [submittingSubmission, setSubmittingSubmission] = useState(false)
  const [mySubmissions, setMySubmissions] = useState<any[]>([])

  const coverUrl = useMemo(() => {
    const g = game
    if (!g) return ''
    return String(g.cover || g.header_image || '').trim()
  }, [game])

  const heroBg = useMemo(() => {
    const g = game
    if (!g) return ''
    return String(g.banner || g.cover || g.header_image || '').trim()
  }, [game])

  const videoPoster = useMemo(() => {
    const g = game
    if (!g) return ''
    return String(g.banner || g.header_image || g.cover || '').trim()
  }, [game])

  const videoUrl = useMemo(() => {
    const g = game
    if (!g) return ''
    return String(g.video_url || '').trim()
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

  const userToken = useMemo(() => {
    if (typeof window === 'undefined') return ''
    return localStorage.getItem('user_token') || ''
  }, [])

  const mergedResources = useMemo(() => {
    const fromDetail = game?.resources
    if (Array.isArray(fromDetail) && fromDetail.length) return fromDetail
    return resources
  }, [game, resources])

  const normalizeLinks = (item: PublicGameResource) => {
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

  const resourceLabelMap: Record<string, string> = {
    game: '本体下载',
    mod: 'Mod 下载',
    trainer: '修改器下载',
    submission: '玩家投稿',
  }

  const resourceGroups = useMemo(() => {
    const grouped = new Map<string, Array<PublicGameResource & { link_list: string[] }>>()
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

  const myGameSubmissions = useMemo(() => {
    const gid = Number(gameId || 0)
    return mySubmissions.filter((item) => Number(item.game_id || 0) === gid)
  }, [mySubmissions, gameId])

  const formatDate = (value?: string) => {
    if (!value) return ''
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return String(value)
    return date.toISOString().slice(0, 10)
  }

  const detectPanLabel = (link: string) => {
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

  const getDownloadLabel = (link: string, index: number, total: number) => {
    const label = detectPanLabel(link)
    return total > 1 ? `${label} ${index + 1}` : label
  }

  const openGalleryViewer = (u: string) => {
    setGalleryViewerUrl(u)
    setGalleryViewerVisible(true)
  }

  const goList = () => {
    window.location.href = '/games'
  }

  const goLogin = () => {
    window.location.href = '/login'
  }

  const resetSubmissionForm = () => {
    setSubmissionForm({
      title: '',
      link: '',
      extract_code: '',
      tags: '',
      description: '',
    })
  }

  const copyLink = async (link: string) => {
    try {
      await navigator.clipboard.writeText(link)
      Toast.success('链接已复制')
    } catch {
      Toast.error('复制失败，请手动复制')
    }
  }

  const loadMySubmissions = async () => {
    if (typeof window === 'undefined') {
      setMySubmissions([])
      return
    }
    const token = localStorage.getItem('user_token') || ''
    if (!token) {
      setMySubmissions([])
      return
    }

    const { data: res } = await siteMySubmissions()
    if (res.code === 200 && Array.isArray(res.data)) {
      setMySubmissions(res.data)
      return
    }
    setMySubmissions([])
  }

  const submitGameResource = async () => {
    const title = String(submissionForm.title || '').trim()
    const link = String(submissionForm.link || '').trim()
    if (!title) return Toast.warning('请填写资源标题')
    if (!link) return Toast.warning('请填写下载链接')
    if (!gameId) return Toast.warning('当前游戏 ID 无效')

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
      Toast.success('提交成功，等待审核')
      resetSubmissionForm()
      await loadMySubmissions()
    } finally {
      setSubmittingSubmission(false)
    }
  }

  useEffect(() => {
    let disposed = false

    const load = async () => {
      if (!gameId || !Number.isFinite(gameId) || gameId <= 0) {
        setLoading(false)
        setGame(null)
        setResources([])
        setVideoStarted(false)
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
        setVideoStarted(false)
        await loadMySubmissions()
      } finally {
        if (!disposed) setLoading(false)
      }
    }

    load()
    return () => {
      disposed = true
    }
  }, [gameId])

  useEffect(() => {
    setGalleryIndex(0)
  }, [galleryUrls.length, gameId])

  const canShowEmpty = !loading && (!game || !game.title)

  const currentGalleryUrl = galleryUrls[galleryIndex] || ''
  const hasGallery = galleryUrls.length > 0

  const gotoPrevGallery = () => {
    if (!hasGallery) return
    setGalleryIndex((prev) => (prev - 1 + galleryUrls.length) % galleryUrls.length)
  }

  const gotoNextGallery = () => {
    if (!hasGallery) return
    setGalleryIndex((prev) => (prev + 1) % galleryUrls.length)
  }

  return (
    <div className="game-detail">
      <div className="toolbar">
        <button type="button" className="back-btn" onClick={goList}>
          ← 返回游戏列表
        </button>
      </div>

      {loading ? (
        <div className="loading-wrap">加载中...</div>
      ) : game ? (
        <>
          <div
            className="hero"
            style={
              heroBg
                ? {
                    backgroundImage: `linear-gradient(180deg, rgba(0,0,0,.55), rgba(15,23,42,.92)), url(${heroBg})`,
                  }
                : undefined
            }
          >
            <div className="hero-inner">
              {coverUrl ? <img src={coverUrl} className="cover" alt={game.title} /> : null}
              <div className="title-block">
                <h1 className="title">{game.title}</h1>
                <div className="chips">
                  {game.developer ? <span className="chip chip-default">{game.developer}</span> : null}
                  {game.type ? <span className="chip chip-info">{game.type}</span> : null}
                  {game.release_date ? <span className="chip chip-success">发行 {game.release_date}</span> : null}
                  {game.price_text ? <span className="price">{game.price_text}</span> : null}
                </div>
                {game.short_description ? <p className="short">{game.short_description}</p> : null}

                <div className="meta-grid">
                  {game.publishers ? (
                    <div className="meta-item">
                      <span className="meta-label">发行商</span>
                      <span className="meta-value">{game.publishers}</span>
                    </div>
                  ) : null}
                  {game.genres ? (
                    <div className="meta-item">
                      <span className="meta-label">类型</span>
                      <span className="meta-value">{game.genres}</span>
                    </div>
                  ) : null}
                  {game.tags ? (
                    <div className="meta-item">
                      <span className="meta-label">标签</span>
                      <span className="meta-value">{game.tags}</span>
                    </div>
                  ) : null}
                  {game.website ? (
                    <div className="meta-item">
                      <span className="meta-label">官网</span>
                      <a className="meta-link" href={game.website} target="_blank" rel="noreferrer">
                        {game.website}
                      </a>
                    </div>
                  ) : null}
                </div>
              </div>
            </div>
          </div>

          {videoUrl ? (
            <div className="section section-card">
              <div className="section-header">视频播放</div>
              <div className="video-wrap">
                {!videoStarted ? (
                  <button type="button" className="video-poster" onClick={() => setVideoStarted(true)}>
                    {videoPoster ? (
                      <img src={videoPoster} alt={game.title} className="video-poster-image" />
                    ) : (
                      <div className="video-poster-fallback" />
                    )}
                    <div className="video-poster-overlay">
                      <span className="video-play-button">▶</span>
                      <span className="video-play-text">点击播放视频</span>
                    </div>
                  </button>
                ) : (
                  <div className="video-player">
                    <div className="rp-wrap">
                      <ReactPlayer
                        src={videoUrl}
                        playing
                        controls
                        width="100%"
                        height="100%"
                        playsInline
                      />
                    </div>
                  </div>
                )}
              </div>
            </div>
          ) : null}

          {hasGallery ? (
            <div className="section section-card">
              <div className="section-header">截图 / 画廊</div>

              <div className="gallery-carousel" aria-label="截图轮播">
                <button type="button" className="gallery-arrow gallery-arrow-left" onClick={gotoPrevGallery} aria-label="上一张">
                  ‹
                </button>

                <button type="button" className="gallery-arrow gallery-arrow-right" onClick={gotoNextGallery} aria-label="下一张">
                  ›
                </button>

                <div
                  className="gallery-slide"
                  role="button"
                  tabIndex={0}
                  onClick={() => openGalleryViewer(currentGalleryUrl)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' || e.key === ' ') openGalleryViewer(currentGalleryUrl)
                  }}
                >
                  <img src={currentGalleryUrl} className="g-img" alt="" loading="lazy" />
                </div>

                <div className="gallery-dots">
                  {galleryUrls.map((u, i) => (
                    <button
                      key={`${u}-${i}`}
                      type="button"
                      className={i === galleryIndex ? 'gallery-dot gallery-dot-active' : 'gallery-dot'}
                      onClick={() => setGalleryIndex(i)}
                      aria-label={`切换到第 ${i + 1} 张`}
                    />
                  ))}
                </div>
              </div>

              <Modal
                visible={galleryViewerVisible}
                title="查看截图"
                onCancel={() => setGalleryViewerVisible(false)}
                footer={null}
                width={860}
                destroyOnClose
              >
                <div className="gallery-viewer">
                  {galleryViewerUrl ? <img src={galleryViewerUrl} className="viewer-img" alt="" /> : null}
                </div>
              </Modal>
            </div>
          ) : null}

          {descriptionHtml ? (
            <div className="section section-card">
              <div className="section-header">游戏介绍</div>
              <div className="desc" dangerouslySetInnerHTML={{ __html: descriptionHtml }} />
            </div>
          ) : null}

          {resourceGroups.length > 0 ? (
            <div className="section section-card">
              <div className="section-header">下载资源</div>
              <div className="resource-groups">
                {resourceGroups.map((group) => (
                  <section key={group.key} className="resource-group">
                    <div className="resource-group-head">
                      <h3>{group.label}</h3>
                      <span>{group.items.length} 个资源</span>
                    </div>

                    <div className="resource-list">
                      {group.items.map((item) => (
                        <article key={item.id} className="resource-card">
                          <div className="resource-main">
                            <div className="resource-title-row">
                              <h4 className="resource-title">{item.title}</h4>
                              {item.tested ? <span className="resource-tested">已测试</span> : null}
                            </div>

                            <div className="resource-meta">
                              {item.version ? (
                                <span className="meta-pill">
                                  <span className="meta-pill-icon">V</span>
                                  <span>版本 {item.version}</span>
                                </span>
                              ) : null}
                              {item.size ? (
                                <span className="meta-pill">
                                  <span className="meta-pill-icon">S</span>
                                  <span>{item.size}</span>
                                </span>
                              ) : null}
                              {item.pan_type ? (
                                <span className="meta-pill">
                                  <span className="meta-pill-icon">P</span>
                                  <span>{item.pan_type}</span>
                                </span>
                              ) : null}
                              {item.download_type ? (
                                <span className="meta-pill">
                                  <span className="meta-pill-icon">D</span>
                                  <span>{item.download_type}</span>
                                </span>
                              ) : null}
                              {item.author ? (
                                <span className="meta-pill">
                                  <span className="meta-pill-icon">A</span>
                                  <span>作者 {item.author}</span>
                                </span>
                              ) : null}
                              {item.publish_date ? (
                                <span className="meta-pill">
                                  <span className="meta-pill-icon">T</span>
                                  <span>{formatDate(item.publish_date)}</span>
                                </span>
                              ) : null}
                            </div>
                          </div>

                          <div className="resource-links">
                            {item.link_list.map((u, idx) => (
                              <div key={`${item.id}-${idx}`} className="download-action">
                                <a
                                  href={u}
                                  target="_blank"
                                  rel="noreferrer"
                                  className="download-link"
                                >
                                  {getDownloadLabel(u, idx, item.link_list.length)}
                                </a>
                                <button type="button" className="copy-link-btn" onClick={() => copyLink(u)}>
                                  复制链接
                                </button>
                              </div>
                            ))}
                          </div>
                        </article>
                      ))}
                    </div>
                  </section>
                ))}
              </div>
            </div>
          ) : null}

          <div className="section section-card">
            <div className="section-header">用户提交资源</div>
            {userToken ? (
              <>
                <div className="submission-form">
                  <div className="submission-field">
                    <label className="submission-label">资源标题</label>
                    <input
                      className="form-input"
                      value={submissionForm.title}
                      onChange={(e) => setSubmissionForm((p) => ({ ...p, title: e.target.value }))}
                      placeholder="例如：Dota 2 夸克网盘整合包"
                    />
                  </div>

                  <div className="submission-field">
                    <label className="submission-label">下载链接</label>
                    <input
                      className="form-input"
                      value={submissionForm.link}
                      onChange={(e) => setSubmissionForm((p) => ({ ...p, link: e.target.value }))}
                      placeholder="粘贴网盘分享链接，支持夸克 / 百度 / 阿里 / 迅雷等"
                    />
                  </div>

                  <div className="submission-grid">
                    <div className="submission-field">
                      <label className="submission-label">提取码</label>
                      <input
                        className="form-input"
                        value={submissionForm.extract_code}
                        onChange={(e) => setSubmissionForm((p) => ({ ...p, extract_code: e.target.value }))}
                        placeholder="没有可不填"
                      />
                    </div>
                    <div className="submission-field">
                      <label className="submission-label">标签</label>
                      <input
                        className="form-input"
                        value={submissionForm.tags}
                        onChange={(e) => setSubmissionForm((p) => ({ ...p, tags: e.target.value }))}
                        placeholder="如：补丁,整合包,教程"
                      />
                    </div>
                  </div>

                  <div className="submission-field">
                    <label className="submission-label">资源说明</label>
                    <textarea
                      className="form-textarea"
                      rows={4}
                      value={submissionForm.description}
                      onChange={(e) => setSubmissionForm((p) => ({ ...p, description: e.target.value }))}
                      placeholder="可以补充版本、适用说明、安装方式等信息"
                    />
                  </div>

                  <div className="submission-actions">
                    <button
                      type="button"
                      className="btn-primary"
                      onClick={() => void submitGameResource()}
                      disabled={submittingSubmission}
                    >
                      {submittingSubmission ? '提交中...' : '提交资源'}
                    </button>
                    <button type="button" className="btn-secondary" onClick={resetSubmissionForm} disabled={submittingSubmission}>
                      清空
                    </button>
                  </div>

                  <div className="submission-hint">
                    提交后会进入审核；审核通过后，会自动出现在当前游戏的资源列表里。
                  </div>
                </div>

                <div className="submission-history">
                  <div className="submission-history-title">我的投稿记录</div>
                  {myGameSubmissions.length > 0 ? (
                    <table className="submissions-table">
                      <thead>
                        <tr>
                          <th>标题</th>
                          <th style={{ width: 110 }}>状态</th>
                          <th>备注</th>
                        </tr>
                      </thead>
                      <tbody>
                        {myGameSubmissions.map((row) => (
                          <tr key={row.id}>
                            <td>{row.title}</td>
                            <td>
                              {row.status === 'pending' ? (
                                <span className="badge badge-warning">待审核</span>
                              ) : row.status === 'approved' ? (
                                <span className="badge badge-success">已通过</span>
                              ) : (
                                <span className="badge badge-danger">已驳回</span>
                              )}
                            </td>
                            <td>{row.review_msg}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  ) : (
                    <div className="empty-state">你还没有为这个游戏提交过资源</div>
                  )}
                </div>
              </>
            ) : (
              <div className="submission-login">
                <div className="empty-state">登录后可为这个游戏提交资源</div>
                <button type="button" className="btn-primary" onClick={goLogin}>
                  去登录
                </button>
              </div>
            )}
          </div>
        </>
      ) : (
        canShowEmpty && <div className="empty-state-page">未找到该游戏或已下架</div>
      )}
    </div>
  )
}

