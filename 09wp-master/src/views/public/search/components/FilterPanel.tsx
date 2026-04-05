import { Button, Card, Checkbox, Select } from '@douyinfe/semi-ui'
import { type ICategory, platformOptions, type SearchFiltersState } from '../searchHelpers'
import styles from '../styles/SearchPage.module.scss'

const sortOptions = [
  { label: '相关度优先', value: 'relevance' },
  { label: '最新发布', value: 'latest' },
  { label: '热度优先', value: 'hot' },
]

const shareTimeOptions = [
  { label: '不限时间', value: '' },
  { label: '今天', value: 'today' },
  { label: '近一周', value: 'week' },
  { label: '近一月', value: 'month' },
  { label: '近一年', value: 'year' },
]

const yearOptions = [
  { label: '不限年份', value: '' },
  { label: '2026', value: '2026' },
  { label: '2025', value: '2025' },
  { label: '2024', value: '2024' },
]

const fileTypeOptions = [
  { label: '文件类型', value: '' },
  { label: 'PDF', value: 'pdf' },
  { label: 'Word', value: 'doc' },
  { label: 'Excel', value: 'xls' },
  { label: 'PPT', value: 'ppt' },
  { label: '图片', value: 'jpg' },
  { label: '视频', value: 'mp4' },
  { label: '压缩包', value: 'zip' },
]

type FilterPanelProps = {
  filters: SearchFiltersState
  categories: ICategory[]
  setFilter: <K extends keyof SearchFiltersState>(key: K, value: SearchFiltersState[K]) => void
  onTipsClick: () => void
}

export default function FilterPanel({ filters, categories, setFilter, onTipsClick }: FilterPanelProps) {
  return (
    <Card className={styles.filterCard} bodyStyle={{ padding: 16 }}>
      <div className={styles.filterItem}>
        <div className={styles.filterLabel}>网盘类型</div>
        <Select
          value={filters.platform}
          onChange={(value) => setFilter('platform', value as SearchFiltersState['platform'])}
          optionList={platformOptions}
          style={{ width: '100%' }}
        />
      </div>

      <div className={styles.filterItem}>
        <div className={styles.filterLabel}>排序方式</div>
        <Select
          value={filters.sort}
          onChange={(value) => setFilter('sort', value as SearchFiltersState['sort'])}
          optionList={sortOptions}
          style={{ width: '100%' }}
        />
      </div>

      <div className={styles.filterItem}>
        <div className={styles.filterLabel}>分享时间</div>
        <Select
          value={filters.shareTime}
          onChange={(value) => setFilter('shareTime', String(value))}
          optionList={shareTimeOptions}
          style={{ width: '100%' }}
        />
      </div>

      <div className={styles.filterItem}>
        <div className={styles.filterLabel}>分享年份</div>
        <Select
          value={filters.shareYear}
          onChange={(value) => setFilter('shareYear', String(value))}
          optionList={yearOptions}
          style={{ width: '100%' }}
        />
      </div>

      <div className={styles.filterItem}>
        <div className={styles.filterLabel}>文件类型</div>
        <Select
          value={filters.fileType}
          onChange={(value) => setFilter('fileType', String(value))}
          optionList={fileTypeOptions}
          style={{ width: '100%' }}
        />
      </div>

      <div className={styles.filterItem}>
        <div className={styles.filterLabel}>资源分类</div>
        <Select
          value={filters.categoryId}
          onChange={(value) => setFilter('categoryId', String(value))}
          optionList={[
            { label: '所有分类', value: '' },
            ...categories.map((item) => ({ label: item.name, value: String(item.id) })),
          ]}
          style={{ width: '100%' }}
        />
      </div>

      <div className={styles.checkboxArea}>
        <Checkbox
          checked={filters.exactMode}
          onChange={(e) => setFilter('exactMode', Boolean(e.target.checked))}
        >
          精确模式
        </Checkbox>
        <Checkbox
          checked={filters.dedupMode}
          onChange={(e) => setFilter('dedupMode', Boolean(e.target.checked))}
        >
          去重模式
        </Checkbox>
      </div>

      <Button block theme="light" onClick={onTipsClick}>
        搜索技巧
      </Button>
    </Card>
  )
}
