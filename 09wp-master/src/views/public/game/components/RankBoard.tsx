import { Card, Typography } from '@douyinfe/semi-ui'
import type { RankItem } from '../mockData'
import styles from '../GameHome.module.scss'

const { Text } = Typography

type RankBoardProps = {
  list: RankItem[]
}

export default function RankBoard({ list }: RankBoardProps) {
  return (
    <Card className={styles.blockCard} bordered={false}>
      <div className={styles.blockTitle}>本周热游榜</div>
      <div className={styles.rankGrid}>
        {list.slice(0, 12).map((item) => (
          <div key={item.id} className={styles.rankRow}>
            <span className={item.rank <= 3 ? styles.rankNumHot : styles.rankNum}>{String(item.rank).padStart(2, '0')}</span>
            <img src={item.cover} alt={item.title} className={styles.rankCover} />
            <div className={styles.rankMeta}>
              <div className={styles.rankTitle}>{item.title}</div>
              <Text type="tertiary" size="small">
                ★ {item.score} · {item.size}
              </Text>
            </div>
          </div>
        ))}
      </div>
    </Card>
  )
}

