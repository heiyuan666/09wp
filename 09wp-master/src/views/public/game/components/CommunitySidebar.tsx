import { Card } from '@douyinfe/semi-ui'
import styles from '../GameHome.module.scss'

export type PromoItem = {
  id: number
  title: string
  path?: string
}

type CommunitySidebarProps = {
  promos: PromoItem[]
}

export default function CommunitySidebar({ promos }: CommunitySidebarProps) {
  return (
    <aside className={styles.rightAside}>
      {promos.map((item) => (
        <Card key={item.id} bordered={false} className={styles.adCard}>
          <a href={item.path || '#'} className={styles.promoLink}>
            <div className={styles.promoBody}>
              <div className={styles.blockTitle}>推荐位</div>
              <div className={styles.promoTitle}>{item.title}</div>
            </div>
          </a>
        </Card>
      ))}
    </aside>
  )
}
