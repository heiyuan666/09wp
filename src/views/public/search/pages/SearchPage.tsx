import { useEffect, useRef, useState } from 'react'
import {
  Banner,
  Button,
  Card,
  ConfigProvider,
  Empty,
  Input,
  Modal,
  Pagination,
  Select,
  Skeleton,
  Spin,
  Tag,
  Toast,
} from '@douyinfe/semi-ui'
import zh_CN from '@douyinfe/semi-ui/lib/es/locale/source/zh_CN'
import { runtimeConfig } from '@/config/runtimeConfig'
import { buildProxiedImageSrc } from '@/utils/coverProxy'
import { guessNameFromMagnet, isMagnetUrl, thunderDownloadSingle } from '@/utils/thunderLinkSdk'
import SearchHeader from '../components/SearchHeader'
import FilterPanel from '../components/FilterPanel'
import ResultCard from '../components/ResultCard'
import styles from '../styles/SearchPage.module.scss'
import type { SearchBridge } from '../useSearchPage'
import { useSearchPage } from '../useSearchPage'

/** 全网搜「获取链接」里不展示的套话；网盘返回的具体错误（如分享已取消）仍展示 */
function isBoilerplateGetLinkMessage(msg: string): boolean {
  const m = String(msg || '').trim()
  if (!m) return true
  if (m.includes('暂不支持自动转存')) return true
  if (m.includes('已返回原始链接')) return true
  if (m.includes('暂未生成本人分享链接')) return true
  if (m.includes('未生成可用链接')) return true
  return false
}

export default function SearchPage(bridge: SearchBridge) {
  const [mobileFilterOpen, setMobileFilterOpen] = useState(false)
  const {
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
    onGoDetail,
    claimGlobalLink,
  } = useSearchPage(bridge)
  const [claimingMap, setClaimingMap] = useState<Record<string, boolean>>({})
  const [linkModalVisible, setLinkModalVisible] = useState(false)
  const [linkModalValue, setLinkModalValue] = useState('')
  const [linkModalMessage, setLinkModalMessage] = useState('')
  const [statusNotice, setStatusNotice] = useState('')
  const [statusNoticeVisible, setStatusNoticeVisible] = useState(false)
  const lastPendingCountRef = useRef(0)

  useEffect(() => {
    const pendingCount = globalList.filter((it) => (it.link_status || 'pending') === 'pending').length
    const prevPending = lastPendingCountRef.current
    if (prevPending > 0 && pendingCount === 0) {
      setStatusNotice('链接检测已完成，结果已自动刷新。')
      setStatusNoticeVisible(true)
      window.setTimeout(() => setStatusNoticeVisible(false), 2200)
    }
    lastPendingCountRef.current = pendingCount
  }, [globalList])

  const keyword = String(bridge.routeQueryQ || '').trim() || qInput.trim()
  const totalSearchMs =
    keyword.trim().length > 0 ? Math.max(0, elapsedMs) + Math.max(0, globalSearchElapsedMs) : elapsedMs
  const isDark = themeMode === 'dark'
  const siteTitle = runtimeConfig.siteTitle || '09网盘搜索'
  const tmdbPoster = buildProxiedImageSrc(tmdbItem?.poster, String(runtimeConfig.tgImageProxyUrl || '').trim())
  const doubanPoster = buildProxiedImageSrc(doubanItem?.poster, String(runtimeConfig.doubanCoverProxyUrl || '').trim())

  return (
    <ConfigProvider locale={zh_CN}>
      <div className={`${styles.page} ${isDark ? styles.pageDark : ''}`}>
        <SearchHeader
          siteTitle={siteTitle}
          keyword={qInput}
          loading={loading}
          onKeywordChange={setQInput}
          onSearch={onSearch}
          onToggleTheme={toggleTheme}
        />

        <main className={styles.main}>
          <aside className={`${styles.aside} ${mobileFilterOpen ? '' : styles.asideCollapsedMobile}`}>
            <FilterPanel
              filters={filters}
              categories={categories}
              setFilter={setFilter}
              onTipsClick={() => Toast.info('支持关键词、分类、网盘类型组合筛选')}
            />
          </aside>

          <section className={styles.content}>
            <Button
              className={styles.mobileFilterToggle}
              theme="light"
              onClick={() => setMobileFilterOpen((prev) => !prev)}
            >
              {mobileFilterOpen ? '收起筛选' : '显示筛选'}
            </Button>

            <div className={styles.summary}>
              <span>{siteTitle}为您提供</span>
              <span className={styles.summaryKeyword}>{keyword || '全部'}</span>
              <span>搜索结果</span>
              <span className={styles.summaryCount}>{total}</span>
              <span>条</span>
              {keyword.trim().length > 0 ? (
                <>
                  <span>，</span>
                  {loading ? (
                    <span style={{ opacity: 0.85 }}>本地搜索中…</span>
                  ) : (
                    <>
                      <span>本地耗时</span>
                      <span className={styles.summaryCount}>{elapsedMs}ms</span>
                    </>
                  )}
                  {globalLoading ? (
                    <span style={{ marginLeft: 6, color: 'var(--semi-color-primary)' }}>· 全网搜正在搜索中…</span>
                  ) : (
                    <>
                      <span style={{ marginLeft: 6 }}>· 全网搜</span>
                      <span className={styles.summaryCount}>{globalSearchElapsedMs}ms</span>
                    </>
                  )}
                  {!loading && !globalLoading ? (
                    <>
                      <span>（合计</span>
                      <span className={styles.summaryCount}>{totalSearchMs}ms</span>
                      <span>）</span>
                    </>
                  ) : null}
                </>
              ) : (
                <>
                  <span>，耗时</span>
                  <span className={styles.summaryCount}>{elapsedMs}ms</span>
                </>
              )}
            </div>
            {statusNotice ? (
              <div
                style={{
                  marginBottom: 10,
                  padding: '8px 10px',
                  borderRadius: 8,
                  fontSize: 12,
                  color: 'var(--semi-color-primary)',
                  background: 'var(--semi-color-primary-light-default)',
                  border: '1px solid var(--semi-color-primary-light-active)',
                  opacity: statusNoticeVisible ? 1 : 0,
                  transform: statusNoticeVisible ? 'translateY(0)' : 'translateY(-6px)',
                  transition: 'all .28s ease',
                  pointerEvents: 'none',
                }}
              >
                {statusNotice}
              </div>
            ) : null}

            {tmdbEnabled && tmdbItem ? (
              <a className={styles.tmdbCard} href={tmdbItem.url || '#'} target="_blank" rel="noreferrer">
                {tmdbPoster ? <img className={styles.tmdbPoster} src={tmdbPoster} alt={tmdbItem.title} /> : null}
                <div className={styles.tmdbBody}>
                  <div className={styles.tmdbTitle}>
                    TMDB：{tmdbItem.title}
                    {tmdbItem.release_date ? ` (${tmdbItem.release_date.slice(0, 4)})` : ''}
                  </div>
                  <div className={styles.tmdbMeta}>
                    {tmdbItem.media_type === 'tv' ? '剧集' : '电影'}
                    {typeof tmdbItem.rating === 'number' ? ` · 评分 ${tmdbItem.rating.toFixed(1)}` : ''}
                  </div>
                  <div className={styles.tmdbOverview}>{tmdbItem.overview || '暂无简介'}</div>
                </div>
              </a>
            ) : null}
            {doubanEnabled && doubanItem ? (
              <a className={styles.tmdbCard} href={doubanItem.url || '#'} target="_blank" rel="noreferrer">
                {doubanPoster ? <img className={styles.tmdbPoster} src={doubanPoster} alt={doubanItem.title} /> : null}
                <div className={styles.tmdbBody}>
                  <div className={styles.tmdbTitle}>
                    豆瓣：{doubanItem.title}
                    {doubanItem.year ? ` (${String(doubanItem.year).trim()})` : ''}
                  </div>
                  <div className={styles.tmdbMeta}>
                    {doubanItem.rating ? `评分 ${doubanItem.rating}` : '暂无评分'}
                  </div>
                  <div className={styles.tmdbOverview}>{doubanItem.overview || '暂无简介'}</div>
                </div>
              </a>
            ) : null}

            {keyword ? (
              <Card
                shadows="hover"
                style={{ marginBottom: 12 }}
                title={`全网搜结果（${globalList.length}）`}
                headerExtraContent={
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <Select
                      value={globalCloudTypeForSelect}
                      onChange={(v) => setGlobalCloudType(String(v || ''))}
                      style={{ width: 180 }}
                      optionList={[
                        { label: '全部网盘', value: '' },
                        { label: '磁力', value: 'magnet' },
                        { label: '夸克', value: 'quark' },
                        { label: '百度', value: 'baidu' },
                        { label: 'UC', value: 'uc' },
                        { label: '迅雷', value: 'xunlei' },
                        { label: '阿里云盘', value: 'aliyun' },
                        { label: '天翼云盘', value: 'tianyi' },
                        { label: '115', value: '115' },
                        { label: '123', value: '123' },
                      ]}
                    />
                    <span style={{ fontSize: 12, opacity: 0.75 }}>数据源：wpysso</span>
                  </div>
                }
              >
                <Spin
                  spinning={globalLoading}
                  tip="正在全网聚合搜索，请稍候…"
                  size="large"
                  style={{ width: '100%', display: 'block' }}
                >
                  <div style={{ minHeight: globalLoading ? 140 : undefined, paddingTop: globalLoading ? 4 : 0 }}>
                    {globalLoading ? (
                      <div aria-busy="true" aria-live="polite">
                        <Skeleton.Paragraph rows={4} />
                        <div
                          style={{
                            marginTop: 12,
                            fontSize: 12,
                            color: 'var(--semi-color-text-2)',
                            textAlign: 'center',
                            opacity: 0.85,
                          }}
                        >
                          正在匹配网盘与磁力线路…
                        </div>
                      </div>
                    ) : globalList.length ? (
                      <>
                        <Banner
                          type="warning"
                          fullMode={false}
                          bordered
                          title="资源约 10 分钟有效"
                          description={
                            <span>
                              全网搜返回的链接会较快失效，请尽快点击「获取链接」并转存或下载保存。
                              <span style={{ marginLeft: 6, opacity: 0.9 }}>（机智提醒：别等过期了才想起来～）</span>
                            </span>
                          }
                          style={{ marginBottom: 12 }}
                        />
                        <div style={{ display: 'grid', gap: 10 }}>
                    {globalList.map((it) => {
                      const key = `${it.url}|${it.password || ''}`
                      const magnet = isMagnetUrl(it.url)
                      return (
                        <div
                          key={key}
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'space-between',
                            gap: 10,
                            border: '1px solid var(--semi-color-border)',
                            borderRadius: 8,
                            padding: '10px 12px',
                          }}
                        >
                          <div style={{ minWidth: 0 }}>
                            <div style={{ fontWeight: 600, marginBottom: 4 }}>{it.note || '全网搜资源'}</div>
                            <div style={{ marginTop: 6, display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                              {it.cloud_type ? <Tag color="blue">网盘：{it.cloud_type}</Tag> : <Tag color="blue">网盘：未知</Tag>}
                              {magnet ? <Tag color="purple">磁力</Tag> : null}
                              {it.password ? <Tag color="green">提取码 {it.password}</Tag> : null}
                              {it.link_status === 'valid' ? <Tag color="light-green">状态：有效</Tag> : null}
                              {it.link_status === 'pending' ? <Tag color="orange">状态：检测中</Tag> : null}
                              {it.link_status === 'unknown' ? <Tag color="grey">状态：未知</Tag> : null}
                            </div>
                          </div>
                          <div style={{ display: 'flex', flexShrink: 0, gap: 8, flexWrap: 'wrap', justifyContent: 'flex-end' }}>
                            {thunderDownloadEnabled && magnet ? (
                              <Button
                                theme="solid"
                                type="tertiary"
                                onClick={async () => {
                                  try {
                                    await thunderDownloadSingle(it.url, {
                                      name: guessNameFromMagnet(it.url) || it.note || undefined,
                                      downloadDir: '全网搜',
                                    })
                                    Toast.success('已发送至迅雷下载')
                                  } catch (error: any) {
                                    Toast.error(error?.message || '迅雷下载唤起失败，请检查是否已安装迅雷')
                                  }
                                }}
                              >
                                迅雷下载
                              </Button>
                            ) : null}
                            <Button
                              loading={Boolean(claimingMap[key])}
                              onClick={async () => {
                                setClaimingMap((prev) => ({ ...prev, [key]: true }))
                                try {
                                  const out = await claimGlobalLink(it)
                                  if (!out) {
                                    Toast.error('获取链接失败')
                                    return
                                  }
                                  setLinkModalValue(out.link)
                                  const tip = String(out.message || '').trim()
                                  setLinkModalMessage(isBoilerplateGetLinkMessage(tip) ? '' : tip)
                                  setLinkModalVisible(true)
                                  const showTip = tip && !isBoilerplateGetLinkMessage(tip)
                                  if (out.status === 'success') {
                                    if (showTip) Toast.success(tip)
                                    else Toast.success('已获取链接')
                                  } else if (out.status === 'pending') {
                                    if (showTip) Toast.warning(tip)
                                    else Toast.warning('转存处理中')
                                  } else {
                                    if (showTip) Toast.info(tip)
                                  }
                                } catch (error: any) {
                                  Toast.error(error?.message || '获取链接失败')
                                } finally {
                                  setClaimingMap((prev) => ({ ...prev, [key]: false }))
                                }
                              }}
                            >
                              获取链接
                            </Button>
                          </div>
                        </div>
                      )
                    })}
                        </div>
                      </>
                    ) : (
                      <Empty description="全网搜暂无结果" imageStyle={{ height: 60 }} />
                    )}
                  </div>
                </Spin>
              </Card>
            ) : null}
            <Modal
              title="小主人，资源拿到啦～快点击查看吧"
              visible={linkModalVisible}
              onCancel={() => setLinkModalVisible(false)}
              footer={
                <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8, flexWrap: 'wrap' }}>
                  <Button onClick={() => setLinkModalVisible(false)}>关闭</Button>
                  {thunderDownloadEnabled && isMagnetUrl(linkModalValue) ? (
                    <Button
                      theme="solid"
                      type="tertiary"
                      onClick={async () => {
                        try {
                          await thunderDownloadSingle(linkModalValue, {
                            name: guessNameFromMagnet(linkModalValue) || undefined,
                            downloadDir: '全网搜',
                          })
                          Toast.success('已发送至迅雷下载')
                        } catch (error: any) {
                          Toast.error(error?.message || '迅雷下载唤起失败，请检查是否已安装迅雷')
                        }
                      }}
                    >
                      迅雷下载
                    </Button>
                  ) : null}
                  <Button
                    theme="light"
                    onClick={() => {
                      if (!linkModalValue) return
                      window.open(linkModalValue, '_blank', 'noopener,noreferrer')
                    }}
                  >
                    打开页面
                  </Button>
                  <Button
                    theme="solid"
                    type="primary"
                    onClick={async () => {
                      try {
                        await navigator.clipboard.writeText(linkModalValue)
                        Toast.success('链接已复制')
                      } catch {
                        Toast.error('复制失败，请手动复制')
                      }
                    }}
                  >
                    复制链接
                  </Button>
                </div>
              }
            >
              {linkModalMessage ? (
                <div style={{ marginBottom: 8, fontSize: 12, color: 'var(--semi-color-text-2)' }}>{linkModalMessage}</div>
              ) : null}
              <Input value={linkModalValue} readonly />
            </Modal>

            <Skeleton
              loading={loading}
              active
              placeholder={
                <div style={{ display: 'grid', gap: 16 }}>
                  <Skeleton.Paragraph rows={6} />
                  <Skeleton.Paragraph rows={6} />
                </div>
              }
            >
              <div className={styles.resultList}>
                {list.map((item) => (
                  <ResultCard
                    key={String(item.id)}
                    item={item}
                    keyword={keyword}
                    categories={categories}
                    onGoDetail={onGoDetail}
                  />
                ))}
              </div>
            </Skeleton>

            {!loading && list.length === 0 ? (
              <Empty description="暂无匹配结果，换个关键词试试" imageStyle={{ height: 80 }} />
            ) : null}

            {total > 0 ? (
              <div className={styles.pager}>
                <Pagination
                  total={total}
                  currentPage={page}
                  pageSize={pageSize}
                  pageSizeOpts={[10, 20, 50]}
                  showSizeChanger
                  showTotal
                  onChange={(current, size) => {
                    setPage(current)
                    if (size !== pageSize) setPageSize(size)
                  }}
                />
              </div>
            ) : null}
          </section>
        </main>
      </div>
    </ConfigProvider>
  )
}
