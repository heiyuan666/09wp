import { IconSun } from '@douyinfe/semi-icons'
import { Button, Typography } from '@douyinfe/semi-ui'
import { CloudIcon } from '../../home/HomeIcons'
import styles from '../styles/SearchPage.module.scss'
import SearchBar from './SearchBar'

const { Text } = Typography

type SearchHeaderProps = {
  siteTitle: string
  keyword: string
  loading: boolean
  onKeywordChange: (value: string) => void
  onSearch: () => void | Promise<void>
  onToggleTheme: () => void
}

export default function SearchHeader({
  siteTitle,
  keyword,
  loading,
  onKeywordChange,
  onSearch,
  onToggleTheme,
}: SearchHeaderProps) {
  return (
    <header className={styles.header}>
      <div className={styles.brand}>
        <CloudIcon size={22} style={{ color: 'rgb(var(--semi-blue-5))' }} />
        <Text strong className={styles.brandName}>
          {siteTitle}
        </Text>
      </div>

      <SearchBar value={keyword} loading={loading} onChange={onKeywordChange} onSearch={onSearch} />

      <div className={styles.headerActions}>
        <Button theme="borderless" type="tertiary" icon={<IconSun />} aria-label="切换主题" onClick={onToggleTheme} />
        <Button theme="solid" type="primary" onClick={() => (window.location.href = '/login')}>
          登录
        </Button>
      </div>
    </header>
  )
}
