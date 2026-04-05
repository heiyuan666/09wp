import { useState } from 'react'
import { Button, ConfigProvider, Empty, Pagination, Skeleton, Toast } from '@douyinfe/semi-ui'
import zh_CN from '@douyinfe/semi-ui/lib/es/locale/source/zh_CN'
import { runtimeConfig } from '@/config/runtimeConfig'
import SearchHeader from '../components/SearchHeader'
import FilterPanel from '../components/FilterPanel'
import ResultCard from '../components/ResultCard'
import styles from '../styles/SearchPage.module.scss'
import type { SearchBridge } from '../useSearchPage'
import { useSearchPage } from '../useSearchPage'

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
    list,
    categories,
    filters,
    setFilter,
    onSearch,
    onGoDetail,
  } = useSearchPage(bridge)

  const keyword = String(bridge.routeQueryQ || '').trim() || qInput.trim()
  const isDark = themeMode === 'dark'
  const siteTitle = runtimeConfig.siteTitle || '懒盘搜索'

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
              <span>条，耗时</span>
              <span className={styles.summaryCount}>{elapsedMs}ms</span>
            </div>

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
