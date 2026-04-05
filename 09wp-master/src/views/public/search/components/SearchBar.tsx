import { IconSearch } from '@douyinfe/semi-icons'
import { Button, Input } from '@douyinfe/semi-ui'
import styles from '../styles/SearchPage.module.scss'

type SearchBarProps = {
  value: string
  loading: boolean
  onChange: (value: string) => void
  onSearch: () => void | Promise<void>
}

export default function SearchBar({ value, loading, onChange, onSearch }: SearchBarProps) {
  return (
    <div className={styles.searchBar}>
      <Input
        size="large"
        value={value}
        onChange={onChange}
        onEnterPress={() => void onSearch()}
        showClear
        prefix={<IconSearch />}
        placeholder="输入关键词搜索网盘资源"
      />
      <Button type="primary" size="large" loading={loading} onClick={() => void onSearch()}>
        搜索
      </Button>
    </div>
  )
}
