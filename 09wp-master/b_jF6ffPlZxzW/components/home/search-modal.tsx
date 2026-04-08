"use client"

import { useState, useEffect, useRef } from "react"
import Image from "next/image"
import Link from "next/link"
import { Search, X, Clock, TrendingUp, Star } from "lucide-react"
import { cn } from "@/lib/utils"
import { fetchGameList, absolutizeGameMediaUrls, splitToList } from "@/lib/api/game"

interface SearchResult {
  id: number
  title: string
  image: string
  price: string
  rating: number
  genre: string
}

const recentSearches = ["赛博朋克", "开放世界", "角色扮演", "科幻游戏"]
const trendingSearches = ["艾尔登法环", "博德之门3", "原神", "黑神话悟空", "星空"]

interface SearchModalProps {
  isOpen: boolean
  onClose: () => void
}

export function SearchModal({ isOpen, onClose }: SearchModalProps) {
  const [query, setQuery] = useState("")
  const [results, setResults] = useState<SearchResult[]>([])
  const [loading, setLoading] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (isOpen && inputRef.current) {
      inputRef.current.focus()
    }
  }, [isOpen])

  useEffect(() => {
    const q = query.trim()
    if (!q) {
      setResults([])
      return
    }

    const handle = setTimeout(async () => {
      setLoading(true)
      try {
        const res = await fetchGameList({ page: 1, page_size: 10, keyword: q })
        const list = (res.list || []).map(absolutizeGameMediaUrls)
        const mapped: SearchResult[] = list.map((g) => {
          const genres = splitToList(g.genres)
          const tags = splitToList(g.tags)
          const genre = (genres[0] || tags[0] || "").trim()
          const priceText = g.price_final === 0 || g.price_text === "免费" ? "免费" : String(Math.round((g.price_final || 0) / 100))
          return {
            id: g.id,
            title: g.title,
            image: g.cover || g.header_image || "/images/game-cover.jpg",
            price: priceText,
            rating: Number(g.rating || 0),
            genre: genre || "游戏",
          }
        })
        setResults(mapped)
      } finally {
        setLoading(false)
      }
    }, 250)

    return () => clearTimeout(handle)
  }, [query])

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        onClose()
      }
    }
    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [onClose])

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50">
      {/* Backdrop */}
      <div 
        className="absolute inset-0 bg-background/80 backdrop-blur-sm"
        onClick={onClose}
      />
      
      {/* Modal */}
      <div className="relative mx-auto max-w-2xl pt-[15vh] px-4">
        <div className="overflow-hidden rounded-2xl border border-border bg-card shadow-2xl">
          {/* Search Input */}
          <div className="flex items-center gap-3 border-b border-border px-4 py-4">
            <Search className="h-5 w-5 text-muted-foreground" />
            <input
              ref={inputRef}
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="搜索游戏、类型、发行商..."
              className="flex-1 bg-transparent text-foreground placeholder:text-muted-foreground focus:outline-none"
            />
            {query && (
              <button onClick={() => setQuery("")} className="text-muted-foreground hover:text-foreground">
                <X className="h-4 w-4" />
              </button>
            )}
            <kbd className="hidden sm:inline-flex items-center gap-1 rounded border border-border bg-secondary px-2 py-1 text-xs text-muted-foreground">
              ESC
            </kbd>
          </div>

          {/* Content */}
          <div className="max-h-[60vh] overflow-y-auto p-4">
            {query.length === 0 ? (
              <div className="space-y-6">
                {/* Recent Searches */}
                <div>
                  <div className="flex items-center gap-2 mb-3">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm font-medium text-muted-foreground">最近搜索</span>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    {recentSearches.map((search) => (
                      <button
                        key={search}
                        onClick={() => setQuery(search)}
                        className="rounded-full bg-secondary px-3 py-1.5 text-sm text-foreground hover:bg-secondary/80 transition-colors"
                      >
                        {search}
                      </button>
                    ))}
                  </div>
                </div>

                {/* Trending Searches */}
                <div>
                  <div className="flex items-center gap-2 mb-3">
                    <TrendingUp className="h-4 w-4 text-primary" />
                    <span className="text-sm font-medium text-muted-foreground">热门搜索</span>
                  </div>
                  <div className="space-y-2">
                    {trendingSearches.map((search, index) => (
                      <button
                        key={search}
                        onClick={() => setQuery(search)}
                        className="flex items-center gap-3 w-full rounded-lg px-3 py-2 text-left hover:bg-secondary transition-colors"
                      >
                        <span className={cn(
                          "flex h-6 w-6 items-center justify-center rounded-full text-xs font-bold",
                          index < 3 ? "bg-primary text-primary-foreground" : "bg-secondary text-muted-foreground"
                        )}>
                          {index + 1}
                        </span>
                        <span className="text-foreground">{search}</span>
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            ) : results.length > 0 ? (
              <div className="space-y-2">
                {results.map((result) => (
                  <Link
                    key={result.id}
                    href={`/game/${result.id}`}
                    onClick={onClose}
                    className="flex items-center gap-4 rounded-lg p-3 hover:bg-secondary transition-colors"
                  >
                    <div className="relative h-16 w-16 flex-shrink-0 overflow-hidden rounded-lg">
                      <Image
                        src={result.image}
                        alt={result.title}
                        fill
                        className="object-cover"
                      />
                    </div>
                    <div className="flex-1 min-w-0">
                      <h3 className="font-semibold text-foreground truncate">{result.title}</h3>
                      <div className="flex items-center gap-3 text-sm text-muted-foreground">
                        <span>{result.genre}</span>
                        {!!result.rating && (
                          <div className="flex items-center gap-1">
                            <Star className="h-3 w-3 fill-primary text-primary" />
                            <span>{result.rating}</span>
                          </div>
                        )}
                      </div>
                    </div>
                    <span className="text-lg font-bold text-foreground">¥{result.price}</span>
                  </Link>
                ))}
              </div>
            ) : (
              <div className="py-12 text-center">
                {loading ? (
                  <p className="text-muted-foreground">搜索中...</p>
                ) : (
                  <>
                    <p className="text-muted-foreground">未找到相关游戏</p>
                    <p className="text-sm text-muted-foreground mt-1">请尝试其他关键词</p>
                  </>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
