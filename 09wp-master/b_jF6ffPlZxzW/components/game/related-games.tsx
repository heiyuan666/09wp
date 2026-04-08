import Image from "next/image"
import { Star } from "lucide-react"

interface RelatedGame {
  id: string
  title: string
  coverUrl: string
  price: number
  rating: number
}

interface RelatedGamesProps {
  games: RelatedGame[]
}

export function RelatedGames({ games }: RelatedGamesProps) {
  return (
    <div className="space-y-4">
      <h3 className="text-xl font-semibold text-foreground">相似游戏</h3>
      
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {games.map((game) => (
          <a
            key={game.id}
            href="#"
            className="group rounded-lg bg-card overflow-hidden transition-transform hover:scale-[1.02]"
          >
            <div className="relative aspect-[3/4] w-full overflow-hidden">
              <Image
                src={game.coverUrl}
                alt={game.title}
                fill
                className="object-cover transition-transform duration-500 group-hover:scale-110"
              />
              <div className="absolute inset-0 bg-gradient-to-t from-background/80 via-transparent to-transparent" />
              <div className="absolute bottom-3 left-3 right-3">
                <h4 className="font-semibold text-foreground text-sm line-clamp-2">{game.title}</h4>
                <div className="mt-1 flex items-center justify-between">
                  <div className="flex items-center gap-1">
                    <Star className="h-3 w-3 fill-primary text-primary" />
                    <span className="text-xs text-foreground">{game.rating.toFixed(1)}</span>
                  </div>
                  <span className="text-sm font-bold text-primary">¥{game.price}</span>
                </div>
              </div>
            </div>
          </a>
        ))}
      </div>
    </div>
  )
}
