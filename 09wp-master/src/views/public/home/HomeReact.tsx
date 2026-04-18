import '@semi-ui-styles/semi.min.css'

import { IconContrast, IconSearch } from '@douyinfe/semi-icons'
import zh_CN from '@douyinfe/semi-ui/lib/es/locale/source/zh_CN'
import { Button, ConfigProvider, Input, Layout, Space, TabPane, Tabs, Tag, Typography } from '@douyinfe/semi-ui'
import { CloudIcon } from './HomeIcons'
import styles from './HomePage.module.scss'
import { defaultConfig } from './types'
import { usePublicHomeData } from './usePublicHomeData'

const { Header, Content, Footer } = Layout
const { Text, Paragraph } = Typography

export default function HomeReact() {
  const {
    keyword,
    setKeyword,
    config,
    topNav,
    homePromos,
    hotSearches,
    hotResources,
    hotByCategory,
    latestResources,
    doubanHot,
    friendLinks,
    year,
    siteTitle,
    siteDescription,
    search,
    openPromo,
    openHotKeyword,
    toggleTheme,
  } = usePublicHomeData()

  const logoNode = config.logo_url ? (
    <img src={config.logo_url} alt="logo" className={styles.logoImg} />
  ) : (
    <CloudIcon size={22} style={{ color: 'var(--semi-color-text-2)' }} />
  )

  return (
    <ConfigProvider locale={zh_CN}>
      <Layout className={styles.page}>
        <Header className={styles.headerBar}>
          <div className={styles.headerInner}>
            <a className={styles.brand} href="/">
              {logoNode}
              {config.show_site_title !== false ? <span className={styles.brandTitle}>{siteTitle}</span> : null}
            </a>

            <nav className={styles.nav} aria-label="顶部导航">
              {topNav.length > 0
                ? topNav.map((item) => (
                    <a
                      key={item.id}
                      className={styles.navLink}
                      href={item.path || '#'}
                      target={item.path?.startsWith('http') ? '_blank' : '_self'}
                      rel={item.path?.startsWith('http') ? 'noopener noreferrer' : undefined}
                    >
                      {item.title}
                    </a>
                  ))
                : null}
            </nav>

            <div className={styles.headerActions}>
              <Button theme="borderless" type="tertiary" icon={<IconContrast />} onClick={toggleTheme} />
              <a className={styles.loginLink} href="/login">
                登录
              </a>
            </div>
          </div>
        </Header>

        <Content className={styles.main}>
          <div className={styles.heroTitle}>
            <CloudIcon size={42} style={{ color: 'var(--semi-color-text-2)' }} />
            <h1>{siteTitle}</h1>
          </div>

          <Paragraph className={styles.subtitle}>{siteDescription}</Paragraph>

          <div className={styles.searchShell}>
            <Input
              className={styles.searchInput}
              value={keyword}
              onChange={setKeyword}
              onEnterPress={search}
              placeholder="输入关键词进行搜索"
              size="large"
              borderless
            />
            <Button theme="borderless" type="tertiary" icon={<IconSearch />} onClick={search} />
          </div>

          <div className={styles.promoGrid}>
            {homePromos.length > 0
              ? homePromos.slice(0, 4).map((item) => (
                  <Button key={item.id} className={styles.promoBtn} theme="light" size="large" onClick={() => openPromo(item)}>
                    {item.title}
                  </Button>
                ))
              : null}
          </div>

          {config.hot_search_enabled !== false ? (
            <section className={styles.hotPanel}>
              <div className={styles.hotHead}>
                <span className={styles.hotTitle}>热门搜索</span>
              </div>
              <Space className={styles.hotWrap} spacing={10} wrap>
                {hotSearches.length > 0 ? (
                  hotSearches.slice(0, 12).map((item, index) => (
                    <Tag
                      key={item.keyword}
                      className={styles.hotTag}
                      color="white"
                      onClick={() => openHotKeyword(item.keyword)}
                      style={{ animationDelay: `${Math.min(index, 8) * 60}ms` }}
                    >
                      <span className={styles.hotIndex}>{index + 1}</span>
                      <span className={styles.hotKeyword}>{item.keyword}</span>
                    </Tag>
                  ))
                ) : (
                  <Text type="tertiary" size="small">
                    暂无热搜词
                  </Text>
                )}
              </Space>
            </section>
          ) : null}

          {config.home_rank_board_enabled !== false ? (
          <section className={styles.rankPanel} aria-label="排行榜">
            <div className={styles.rankHead}>
              <span className={styles.rankTitle}>排行榜</span>
            </div>
            <Tabs type="line" className={styles.rankTabs}>
              <TabPane tab="热门资源" itemKey="hot">
                <div className={styles.rankList}>
                  {hotResources.length > 0 ? (
                    hotResources.slice(0, 10).map((it, idx) => (
                      <a key={it.id} className={styles.rankRow} href={`/r/${it.id}`}>
                        <span className={styles.rankNo}>{idx + 1}</span>
                        <span className={styles.rankText}>{it.title}</span>
                      </a>
                    ))
                  ) : (
                    <Text type="tertiary" size="small">
                      暂无数据
                    </Text>
                  )}
                </div>
              </TabPane>
              <TabPane tab="分类热门" itemKey="hot_by_category">
                <div className={styles.rankList}>
                  {hotByCategory.length > 0 ? (
                    hotByCategory.slice(0, 8).map((group) => (
                      <div key={group.category_id} style={{ marginBottom: 14 }}>
                        <Text strong>{group.category_name}</Text>
                        {group.resources.slice(0, 5).map((it, idx) => (
                          <a key={it.id} className={styles.rankRow} href={`/r/${it.id}`}>
                            <span className={styles.rankNo}>{idx + 1}</span>
                            <span className={styles.rankText}>{it.title}</span>
                          </a>
                        ))}
                      </div>
                    ))
                  ) : (
                    <Text type="tertiary" size="small">
                      暂无数据
                    </Text>
                  )}
                </div>
              </TabPane>
              <TabPane tab="最新资源" itemKey="latest">
                <div className={styles.rankList}>
                  {latestResources.length > 0 ? (
                    latestResources.slice(0, 10).map((it, idx) => (
                      <a key={it.id} className={styles.rankRow} href={`/r/${it.id}`}>
                        <span className={styles.rankNo}>{idx + 1}</span>
                        <span className={styles.rankText}>{it.title}</span>
                        <span className={styles.rankMeta}>最新</span>
                      </a>
                    ))
                  ) : (
                    <Text type="tertiary" size="small">
                      暂无数据
                    </Text>
                  )}
                </div>
              </TabPane>
              <TabPane tab="豆瓣热门" itemKey="douban">
                <div className={styles.rankList}>
                  {doubanHot.length > 0 ? (
                    doubanHot.slice(0, 10).map((it, idx) => (
                      <button
                        key={`${it.title}-${idx}`}
                        className={styles.rankRowBtn}
                        onClick={() => openHotKeyword(it.title)}
                      >
                        <span className={styles.rankNo}>{idx + 1}</span>
                        <span className={styles.rankText}>{it.title}</span>
                        <span className={styles.rankMeta}>搜索</span>
                      </button>
                    ))
                  ) : (
                    <Text type="tertiary" size="small">
                      暂无数据
                    </Text>
                  )}
                </div>
              </TabPane>
            </Tabs>
          </section>
          ) : null}
        </Content>

        <Footer className={styles.footerBar}>
          <div className={styles.footerInner}>
            <Paragraph className={styles.disclaimer} size="small" type="tertiary">
              敬告与声明：本站不产生、存储任何数据，也从未参与录制、上传，所有资源均来自网络及网友提交；无意冒犯任何公司、用户的权益、版权。
            </Paragraph>

            {friendLinks.length > 0 ? (
              <div className={styles.friendLinks}>
                <span className={styles.friendTitle}>友情链接：</span>
                {friendLinks.map((fl) => (
                  <a key={`${fl.title}-${fl.url}`} href={fl.url || '#'} target="_blank" rel="noopener noreferrer">
                    {fl.title}
                  </a>
                ))}
              </div>
            ) : null}

            <Text type="tertiary" size="small">
              {config.footer_text || defaultConfig.footer_text}
              {config.icp_record ? ` | ${config.icp_record}` : ''}
              {` | © ${year} ${siteTitle}`}
            </Text>
          </div>
        </Footer>
      </Layout>
    </ConfigProvider>
  )
}
