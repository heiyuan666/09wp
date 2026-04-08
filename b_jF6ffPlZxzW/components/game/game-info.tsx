import { Star, Calendar, Users, Globe, Monitor } from "lucide-react"
import { Badge } from "@/components/ui/badge"

interface GameInfoProps {
  title: string
  developer: string
  publisher: string
  releaseDate: string
  rating: number
  reviewCount: number
  genres: string[]
  platforms: string[]
  languages: string[]
}

export function GameInfo({
  title,
  developer,
  publisher,
  releaseDate,
  rating,
  reviewCount,
  genres,
  platforms,
  languages,
}: GameInfoProps) {
  return (
    <div className="space-y-6">
      {/* Title */}
      <div>
        <h1 className="text-3xl font-bold text-foreground md:text-4xl lg:text-5xl text-balance">
          {title}
        </h1>
        <p className="mt-2 text-muted-foreground">
          由 <span className="text-foreground">{developer}</span> 开发 · 
          <span className="text-foreground"> {publisher}</span> 发行
        </p>
      </div>

      {/* Rating */}
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2">
          <div className="flex">
            {[1, 2, 3, 4, 5].map((star) => (
              <Star
                key={star}
                className={`h-5 w-5 ${
                  star <= rating
                    ? "fill-primary text-primary"
                    : "fill-muted text-muted"
                }`}
              />
            ))}
          </div>
          <span className="text-lg font-semibold text-foreground">{rating.toFixed(1)}</span>
        </div>
        <span className="text-muted-foreground">({reviewCount.toLocaleString()} 条评价)</span>
      </div>

      {/* Genres */}
      <div className="flex flex-wrap gap-2">
        {genres.map((genre) => (
          <Badge key={genre} variant="secondary" className="bg-secondary text-secondary-foreground">
            {genre}
          </Badge>
        ))}
      </div>

      {/* Details */}
      <div className="space-y-3 rounded-lg bg-card p-4">
        <div className="flex items-center gap-3 text-sm">
          <Calendar className="h-4 w-4 text-muted-foreground" />
          <span className="text-muted-foreground">发售日期:</span>
          <span className="text-foreground">{releaseDate}</span>
        </div>
        <div className="flex items-center gap-3 text-sm">
          <Monitor className="h-4 w-4 text-muted-foreground" />
          <span className="text-muted-foreground">平台:</span>
          <span className="text-foreground">{platforms.join(", ")}</span>
        </div>
        <div className="flex items-center gap-3 text-sm">
          <Globe className="h-4 w-4 text-muted-foreground" />
          <span className="text-muted-foreground">语言:</span>
          <span className="text-foreground">{languages.join(", ")}</span>
        </div>
        <div className="flex items-center gap-3 text-sm">
          <Users className="h-4 w-4 text-muted-foreground" />
          <span className="text-muted-foreground">游戏模式:</span>
          <span className="text-foreground">单人</span>
        </div>
      </div>
    </div>
  )
}
