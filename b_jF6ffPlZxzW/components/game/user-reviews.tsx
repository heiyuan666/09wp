"use client"

import { useState } from "react"
import Image from "next/image"
import { 
  Star, 
  ThumbsUp, 
  ThumbsDown, 
  MessageSquare, 
  ChevronDown,
  Flag,
  MoreHorizontal
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

interface Review {
  id: string
  user: {
    name: string
    avatar: string
    level: number
    gamesOwned: number
  }
  rating: number
  content: string
  pros?: string[]
  cons?: string[]
  date: string
  playtime: string
  helpful: number
  unhelpful: number
  replies: number
}

const mockReviews: Review[] = [
  {
    id: "1",
    user: {
      name: "夜之城漫游者",
      avatar: "/images/screenshot-1.jpg",
      level: 28,
      gamesOwned: 156,
    },
    rating: 5,
    content: "绝对是近年来最好的开放世界游戏之一！画面效果惊艳，剧情深度令人叹服。虽然刚发售时有些问题，但经过多次更新后，现在的体验已经非常完美了。强烈推荐给所有喜欢角色扮演游戏的玩家。",
    pros: ["画面精美", "剧情深度", "自由度高", "音乐出色"],
    cons: ["配置要求高", "部分任务重复"],
    date: "2024年3月15日",
    playtime: "128.5 小时",
    helpful: 256,
    unhelpful: 12,
    replies: 23,
  },
  {
    id: "2",
    user: {
      name: "硬核玩家小明",
      avatar: "/images/screenshot-2.jpg",
      level: 42,
      gamesOwned: 312,
    },
    rating: 4,
    content: "游戏整体质量很高，尤其是夜之城的氛围营造得非常到位。不过个人觉得战斗系统可以更丰富一些，有时候会感觉有点单调。但瑕不掩瑜，依然是值得一玩的佳作。",
    pros: ["世界观完整", "支线任务丰富"],
    cons: ["战斗系统一般", "AI有时不够智能"],
    date: "2024年3月10日",
    playtime: "86.2 小时",
    helpful: 128,
    unhelpful: 8,
    replies: 15,
  },
  {
    id: "3",
    user: {
      name: "游戏收藏家",
      avatar: "/images/screenshot-3.jpg",
      level: 15,
      gamesOwned: 89,
    },
    rating: 5,
    content: "入手之后根本停不下来！每一个角落都有故事，每一个NPC都有自己的生活。这种沉浸感是其他游戏很难给予的。DLC更是锦上添花，强烈推荐购买完整版！",
    date: "2024年3月8日",
    playtime: "200+ 小时",
    helpful: 89,
    unhelpful: 3,
    replies: 8,
  },
]

const ratingStats = {
  average: 4.5,
  total: 125847,
  distribution: [
    { stars: 5, percentage: 68 },
    { stars: 4, percentage: 22 },
    { stars: 3, percentage: 6 },
    { stars: 2, percentage: 2 },
    { stars: 1, percentage: 2 },
  ],
}

export function UserReviews() {
  const [reviews] = useState<Review[]>(mockReviews)
  const [sortBy, setSortBy] = useState<"helpful" | "recent">("helpful")
  const [showAll, setShowAll] = useState(false)

  const displayedReviews = showAll ? reviews : reviews.slice(0, 2)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-foreground">用户评价</h2>
        <Button variant="outline" size="sm">
          撰写评价
        </Button>
      </div>

      {/* Rating Overview */}
      <div className="grid gap-6 md:grid-cols-2 rounded-xl bg-card border border-border p-6">
        {/* Average Rating */}
        <div className="flex items-center gap-6">
          <div className="text-center">
            <div className="text-5xl font-bold text-foreground">{ratingStats.average}</div>
            <div className="flex items-center justify-center gap-1 my-2">
              {[1, 2, 3, 4, 5].map((star) => (
                <Star
                  key={star}
                  className={cn(
                    "h-5 w-5",
                    star <= Math.round(ratingStats.average)
                      ? "fill-primary text-primary"
                      : "text-muted-foreground"
                  )}
                />
              ))}
            </div>
            <div className="text-sm text-muted-foreground">
              {ratingStats.total.toLocaleString()} 条评价
            </div>
          </div>
        </div>

        {/* Rating Distribution */}
        <div className="space-y-2">
          {ratingStats.distribution.map((item) => (
            <div key={item.stars} className="flex items-center gap-3">
              <div className="flex items-center gap-1 w-12">
                <span className="text-sm text-muted-foreground">{item.stars}</span>
                <Star className="h-3 w-3 fill-primary text-primary" />
              </div>
              <div className="flex-1 h-2 rounded-full bg-secondary overflow-hidden">
                <div
                  className="h-full bg-primary rounded-full transition-all duration-500"
                  style={{ width: `${item.percentage}%` }}
                />
              </div>
              <span className="text-sm text-muted-foreground w-12 text-right">
                {item.percentage}%
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* Sort Options */}
      <div className="flex items-center gap-2">
        <span className="text-sm text-muted-foreground">排序：</span>
        <Button
          variant={sortBy === "helpful" ? "secondary" : "ghost"}
          size="sm"
          onClick={() => setSortBy("helpful")}
        >
          最有帮助
        </Button>
        <Button
          variant={sortBy === "recent" ? "secondary" : "ghost"}
          size="sm"
          onClick={() => setSortBy("recent")}
        >
          最新评价
        </Button>
      </div>

      {/* Reviews List */}
      <div className="space-y-4">
        {displayedReviews.map((review) => (
          <div
            key={review.id}
            className="rounded-xl bg-card border border-border p-5 transition-colors hover:border-border/80"
          >
            {/* User Info */}
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center gap-3">
                <div className="relative h-10 w-10 overflow-hidden rounded-full">
                  <Image
                    src={review.user.avatar}
                    alt={review.user.name}
                    fill
                    className="object-cover"
                  />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-foreground">{review.user.name}</span>
                    <span className="rounded bg-secondary px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
                      Lv.{review.user.level}
                    </span>
                  </div>
                  <div className="text-xs text-muted-foreground">
                    已拥有 {review.user.gamesOwned} 款游戏
                  </div>
                </div>
              </div>
              
              <div className="flex items-center gap-2">
                <div className="flex items-center gap-1">
                  {[1, 2, 3, 4, 5].map((star) => (
                    <Star
                      key={star}
                      className={cn(
                        "h-4 w-4",
                        star <= review.rating
                          ? "fill-primary text-primary"
                          : "text-muted-foreground"
                      )}
                    />
                  ))}
                </div>
                <Button variant="ghost" size="icon" className="h-8 w-8">
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </div>
            </div>

            {/* Content */}
            <p className="text-foreground leading-relaxed mb-4">{review.content}</p>

            {/* Pros & Cons */}
            {(review.pros || review.cons) && (
              <div className="flex flex-wrap gap-4 mb-4">
                {review.pros && (
                  <div className="flex flex-wrap gap-1.5">
                    {review.pros.map((pro) => (
                      <span
                        key={pro}
                        className="rounded-full bg-primary/10 px-2.5 py-1 text-xs font-medium text-primary"
                      >
                        + {pro}
                      </span>
                    ))}
                  </div>
                )}
                {review.cons && (
                  <div className="flex flex-wrap gap-1.5">
                    {review.cons.map((con) => (
                      <span
                        key={con}
                        className="rounded-full bg-destructive/10 px-2.5 py-1 text-xs font-medium text-destructive"
                      >
                        - {con}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            )}

            {/* Meta & Actions */}
            <div className="flex items-center justify-between pt-4 border-t border-border">
              <div className="flex items-center gap-4 text-xs text-muted-foreground">
                <span>{review.date}</span>
                <span>游玩时长: {review.playtime}</span>
              </div>
              <div className="flex items-center gap-2">
                <Button variant="ghost" size="sm" className="h-8 gap-1.5">
                  <ThumbsUp className="h-3.5 w-3.5" />
                  <span>{review.helpful}</span>
                </Button>
                <Button variant="ghost" size="sm" className="h-8 gap-1.5">
                  <ThumbsDown className="h-3.5 w-3.5" />
                  <span>{review.unhelpful}</span>
                </Button>
                <Button variant="ghost" size="sm" className="h-8 gap-1.5">
                  <MessageSquare className="h-3.5 w-3.5" />
                  <span>{review.replies}</span>
                </Button>
                <Button variant="ghost" size="sm" className="h-8">
                  <Flag className="h-3.5 w-3.5" />
                </Button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Show More */}
      {reviews.length > 2 && (
        <Button
          variant="outline"
          className="w-full"
          onClick={() => setShowAll(!showAll)}
        >
          <ChevronDown className={cn(
            "mr-2 h-4 w-4 transition-transform",
            showAll && "rotate-180"
          )} />
          {showAll ? "收起评价" : `查看更多评价 (${reviews.length - 2})`}
        </Button>
      )}
    </div>
  )
}
