import { Card, Tag, Typography } from '@douyinfe/semi-ui'
import type { UpdateItem } from '../mockData'
import styles from '../GameHome.module.scss'

const { Text } = Typography

type UpdateStripProps = {
  list: UpdateItem[]
}

export default function UpdateStrip({ list }: UpdateStripProps) {
  return (
    <Card className={styles.blockCard} bordered={false}>
      <div className={styles.blockTitle}>今日更新</div>
      <div className={styles.updateGrid}>
        {list.map((item) => (
          <article key={item.id} className={styles.updateCard}>
            <img src={item.cover} alt={item.title} className={styles.updateCover} />
            <div className={styles.updateBody}>
              <div className={styles.updateTitle}>{item.title}</div>
              <Text type="tertiary" size="small">
                发行日期：{item.date}
              </Text>
            </div>
            <Tag color="blue" size="small">
              {item.tag}
            </Tag>
          </article>
        ))}
      </div>
    </Card>
  )
}

