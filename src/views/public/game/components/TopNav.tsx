import { IconBell, IconSearch, IconUser } from '@douyinfe/semi-icons'
import { Button, Input, Space } from '@douyinfe/semi-ui'
import styles from '../GameHome.module.scss'

export type NavItem = {
  title: string
  path?: string
}

type TopNavProps = {
  navItems: NavItem[]
  keyword: string
  onKeywordChange: (value: string) => void
  onSearch: () => void
}

export default function TopNav({ navItems, keyword, onKeywordChange, onSearch }: TopNavProps) {
  return (
    <header className={styles.topHeader}>
      <div className={styles.brandArea}>
        <img src="https://picsum.photos/seed/logo66/34/34" alt="logo" className={styles.logo} />
        <span className={styles.siteTitle}>游戏资源站</span>
      </div>

      <nav className={styles.navLinks}>
        {navItems.map((item, idx) => (
          <a
            key={`${item.title}-${idx}`}
            className={idx === 0 ? styles.navActive : styles.navLink}
            href={item.path || '#'}
          >
            {item.title}
          </a>
        ))}
      </nav>

      <div className={styles.searchWrap}>
        <Input
          prefix={<IconSearch />}
          value={keyword}
          onChange={onKeywordChange}
          onEnterPress={onSearch}
          placeholder="搜索游戏 / 标签 / 开发商"
          size="large"
          showClear
          className={styles.searchInput}
        />
        <Button type="primary" theme="solid" icon={<IconSearch />} className={styles.searchButton} onClick={onSearch}>
          搜索
        </Button>
      </div>

      <Space spacing={8}>
        <Button icon={<IconBell />} theme="borderless" />
        <Button icon={<IconUser />} theme="borderless" />
      </Space>
    </header>
  )
}
