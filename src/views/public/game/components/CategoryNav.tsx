import styles from '../GameHome.module.scss'

type CategoryNavProps = {
  categories: string[]
  activeCategory: string
  onCategoryChange: (v: string) => void
}

export default function CategoryNav({
  categories,
  activeCategory,
  onCategoryChange,
}: CategoryNavProps) {
  return (
    <aside className={styles.categoryRail}>
      <div className={styles.categoryRailTitle}>游戏分类</div>
      <div className={styles.categoryRailList}>
        {categories.map((c) => (
          <button
            key={c}
            type="button"
            className={c === activeCategory ? styles.categoryRailItemActive : styles.categoryRailItem}
            onClick={() => onCategoryChange(c)}
          >
            <span>{c}</span>
          </button>
        ))}
      </div>
    </aside>
  )
}
