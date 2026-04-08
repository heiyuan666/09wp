"use client"

import { useEffect, useMemo, useState } from "react"
import Image from "next/image"
import Link from "next/link"
import { 
  Star, 
  ChevronDown,
  ThumbsUp,
  ThumbsDown,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { Textarea } from "@/components/ui/textarea"
import { createGameReview, fetchGameReviews, voteGameReview, type GameReviewDistributionItem, type GameReviewItem } from "@/lib/api/review"

function formatDateZh(isoLike?: string) {
  if (!isoLike) return ""
  const d = new Date(isoLike)
  if (Number.isNaN(d.getTime())) return isoLike
  return d.toLocaleDateString("zh-CN", { year: "numeric", month: "long", day: "numeric" })
}

function fallbackDist(): GameReviewDistributionItem[] {
  return [
    { stars: 5, count: 0, percentage: 0 },
    { stars: 4, count: 0, percentage: 0 },
    { stars: 3, count: 0, percentage: 0 },
    { stars: 2, count: 0, percentage: 0 },
    { stars: 1, count: 0, percentage: 0 },
  ]
}

export function UserReviews({ gameId }: { gameId: number }) {
  const [reviews, setReviews] = useState<GameReviewItem[]>([])
  const [total, setTotal] = useState(0)
  const [average, setAverage] = useState(0)
  const [distribution, setDistribution] = useState<GameReviewDistributionItem[]>(fallbackDist())

  const [loading, setLoading] = useState(false)
  const [posting, setPosting] = useState(false)
  const [error, setError] = useState("")

  const [isWriting, setIsWriting] = useState(false)
  const [rating, setRating] = useState(5)
  const [content, setContent] = useState("")

  const [sortBy, setSortBy] = useState<"helpful" | "recent">("recent")
  const [showAll, setShowAll] = useState(false)

  const displayedReviews = showAll ? reviews : reviews.slice(0, 5)

  const hasToken = useMemo(() => {
    try {
      return Boolean(localStorage.getItem("token"))
    } catch {
      return false
    }
  }, [])

  async function load() {
    setLoading(true)
    setError("")
    try {
      const res = await fetchGameReviews({ game_id: gameId, page: 1, page_size: 20, sort: sortBy })
      setReviews(res.list || [])
      setTotal(res.total || 0)
      setAverage(res.average || 0)
      setDistribution((res.distribution && res.distribution.length ? res.distribution : fallbackDist()) as GameReviewDistributionItem[])
    } catch (e) {
      setError(e instanceof Error ? e.message : "加载失败")
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [gameId, sortBy])

  async function onSubmit() {
    setError("")
    if (content.trim().length < 3) {
      setError("评论内容至少 3 个字")
      return
    }
    let token = ""
    try {
      token = localStorage.getItem("token") || ""
    } catch {
      token = ""
    }
    if (!token) {
      setError("请先登录再评论")
      return
    }
    setPosting(true)
    try {
      await createGameReview({ token, game_id: gameId, rating, content: content.trim() })
      setContent("")
      setIsWriting(false)
      await load()
    } catch (e) {
      setError(e instanceof Error ? e.message : "发布失败")
    } finally {
      setPosting(false)
    }
  }

  async function onVote(reviewId: number, vote: -1 | 0 | 1) {
    setError("")
    let token = ""
    try {
      token = localStorage.getItem("token") || ""
    } catch {
      token = ""
    }
    if (!token) {
      setError("请先登录再投票")
      return
    }
    try {
      await voteGameReview({ token, review_id: reviewId, vote })
      await load()
    } catch (e) {
      setError(e instanceof Error ? e.message : "投票失败")
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-foreground">用户评价</h2>
        <Button variant="outline" size="sm" onClick={() => setIsWriting((v) => !v)}>
          {isWriting ? "取消" : "撰写评价"}
        </Button>
      </div>

      {isWriting ? (
        <div className="rounded-xl bg-card border border-border p-5 space-y-4">
          <div className="flex items-center justify-between">
            <div className="text-sm font-medium text-foreground">你的评分</div>
            <div className="flex items-center gap-1">
              {[1, 2, 3, 4, 5].map((s) => (
                <button
                  key={s}
                  type="button"
                  className="p-1"
                  onClick={() => setRating(s)}
                  aria-label={`rate-${s}`}
                >
                  <Star className={cn("h-5 w-5", s <= rating ? "fill-primary text-primary" : "text-muted-foreground")} />
                </button>
              ))}
            </div>
          </div>
          <Textarea value={content} onChange={(e) => setContent(e.target.value)} placeholder="写下你的评价..." />
          {!hasToken ? (
            <div className="flex items-center justify-between gap-3 text-xs text-muted-foreground">
              <span>提示：需要登录后才能发布评论</span>
              <Button variant="secondary" size="sm" asChild>
                <Link href="/login">去登录</Link>
              </Button>
            </div>
          ) : null}
          {error ? <div className="text-sm text-destructive">{error}</div> : null}
          <div className="flex justify-end">
            <Button onClick={onSubmit} disabled={posting}>
              {posting ? "发布中..." : "发布评价"}
            </Button>
          </div>
        </div>
      ) : null}

      {/* Rating Overview */}
      <div className="grid gap-6 md:grid-cols-2 rounded-xl bg-card border border-border p-6">
        {/* Average Rating */}
        <div className="flex items-center gap-6">
          <div className="text-center">
            <div className="text-5xl font-bold text-foreground">{average ? average.toFixed(1) : "0.0"}</div>
            <div className="flex items-center justify-center gap-1 my-2">
              {[1, 2, 3, 4, 5].map((star) => (
                <Star
                  key={star}
                  className={cn(
                    "h-5 w-5",
                    star <= Math.round(average)
                      ? "fill-primary text-primary"
                      : "text-muted-foreground"
                  )}
                />
              ))}
            </div>
            <div className="text-sm text-muted-foreground">
              {total.toLocaleString()} 条评价
            </div>
          </div>
        </div>

        {/* Rating Distribution */}
        <div className="space-y-2">
          {distribution.map((item) => (
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
          variant={sortBy === "recent" ? "secondary" : "ghost"}
          size="sm"
          onClick={() => setSortBy("recent")}
        >
          最新评价
        </Button>
        <Button
          variant={sortBy === "helpful" ? "secondary" : "ghost"}
          size="sm"
          onClick={() => setSortBy("helpful")}
        >
          最有帮助
        </Button>
      </div>

      {/* Reviews List */}
      <div className="space-y-4">
        {loading ? <div className="text-sm text-muted-foreground">加载中...</div> : null}
        {!loading && error && !isWriting ? <div className="text-sm text-destructive">{error}</div> : null}
        {!loading && !error && reviews.length === 0 ? (
          <div className="text-sm text-muted-foreground">暂无评论，快来抢沙发。</div>
        ) : null}
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
                    src={review.user.avatar || "/images/game-cover.jpg"}
                    alt={review.user.username || "用户"}
                    fill
                    className="object-cover"
                  />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-foreground">{review.user.username || "匿名用户"}</span>
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {formatDateZh(review.created_at)}
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
                        star <= (review.rating || 0)
                          ? "fill-primary text-primary"
                          : "text-muted-foreground"
                      )}
                    />
                  ))}
                </div>
              </div>
            </div>

            {/* Content */}
            <p className="text-foreground leading-relaxed mb-4">{review.content}</p>

            <div className="flex items-center justify-between pt-4 border-t border-border">
              <div className="text-xs text-muted-foreground">评分与评论为用户发布内容，仅供参考。</div>
              <div className="flex items-center gap-2">
                <Button variant="ghost" size="sm" className="h-8 gap-1.5" onClick={() => onVote(review.id, 1)}>
                  <ThumbsUp className="h-3.5 w-3.5" />
                  <span>{review.helpful || 0}</span>
                </Button>
                <Button variant="ghost" size="sm" className="h-8 gap-1.5" onClick={() => onVote(review.id, -1)}>
                  <ThumbsDown className="h-3.5 w-3.5" />
                  <span>{review.unhelpful || 0}</span>
                </Button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Show More */}
      {reviews.length > 5 && (
        <Button
          variant="outline"
          className="w-full"
          onClick={() => setShowAll(!showAll)}
        >
          <ChevronDown className={cn(
            "mr-2 h-4 w-4 transition-transform",
            showAll && "rotate-180"
          )} />
          {showAll ? "收起评价" : `查看更多评价 (${reviews.length - 5})`}
        </Button>
      )}
    </div>
  )
}
