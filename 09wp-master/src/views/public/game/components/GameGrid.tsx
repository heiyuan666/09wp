import type { GameItem } from '../mockData'
import styles from '../GameHome.module.scss'

type GameGridProps = {
  games: GameItem[]
}

export default function GameGrid({ games }: GameGridProps) {
  return (
    <section className={styles.gameFeed}>
      {games.map((game) => (
        <article
          key={game.id}
          className={styles.gameFeedItem}
          role="button"
          tabIndex={0}
          onClick={() => {
            window.location.href = `/games/${game.id}`
          }}
          onKeyDown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') {
              e.preventDefault()
              window.location.href = `/games/${game.id}`
            }
          }}
        >
          <img src={game.cover} alt={game.title} className={styles.gameFeedThumb} />
          <div className={styles.gameFeedBody}>
            <div className={styles.gameFeedTop}>
              <div className={styles.gameFeedHeading}>
                <h4 className={styles.gameFeedTitle}>{game.title}</h4>
                <div className={styles.gameFeedTags}>
                  <span className={styles.gameFeedTag}>{game.category}</span>
                  <span className={styles.gameFeedTag}>{game.subType}</span>
                  {game.size && game.size !== '-' ? <span className={styles.gameFeedTag}>{game.size}</span> : null}
                </div>
              </div>
              <div className={styles.score}>★ {game.score}</div>
            </div>

            {game.description ? <p className={styles.gameFeedDesc}>{game.description}</p> : null}

            <div className={styles.gameFeedMeta}>
              <span>发行日期：{game.releaseDate}</span>
              <span>{game.downloads} 下载</span>
              <span>{game.updateText}</span>
            </div>

            {game.publisher || game.developer ? (
              <div className={styles.gameFeedMeta}>
                {game.developer ? <span>开发：{game.developer}</span> : null}
                {game.publisher ? <span>发行：{game.publisher}</span> : null}
              </div>
            ) : null}
          </div>
        </article>
      ))}
    </section>
  )
}
