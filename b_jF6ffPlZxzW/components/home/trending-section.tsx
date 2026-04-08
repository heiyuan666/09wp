"use client"

import { useState } from "react"
import { ArrowRight } from "lucide-react"
import { Button } from "@/components/ui/button"
import { GameCard } from "./game-card"
import { cn } from "@/lib/utils"

const tabs = [
  { id: "trending", label: "热门游戏" },
  { id: "new", label: "新品上架" },
  { id: "upcoming", label: "即将推出" },
  { id: "free", label: "免费游戏" },
]

export type TrendingTabId = (typeof tabs)[number]["id"]

export type GameCardItem = {
  id: number
  title: string
  image: string
  price: string
  originalPrice?: string
  discount?: number
  rating?: number
  releaseDate?: string
  tags?: string[]
  isNew?: boolean
  isTrending?: boolean
}

export function TrendingSection({ gamesByTab }: { gamesByTab: Record<TrendingTabId, GameCardItem[]> }) {
  const [activeTab, setActiveTab] = useState("trending")

  const currentGames = gamesByTab[activeTab as TrendingTabId] || []

  return (
    <section>
      {/* Header with Tabs */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
        <div className="flex items-center gap-2 overflow-x-auto pb-2 sm:pb-0">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                "whitespace-nowrap rounded-full px-4 py-2 text-sm font-medium transition-all duration-300",
                activeTab === tab.id
                  ? "bg-primary text-primary-foreground"
                  : "bg-secondary text-muted-foreground hover:bg-secondary/80 hover:text-foreground"
              )}
            >
              {tab.label}
            </button>
          ))}
        </div>
        <Button variant="ghost" className="gap-2 text-primary self-end">
          查看全部
          <ArrowRight className="h-4 w-4" />
        </Button>
      </div>

      {/* Games Grid */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-4">
        {currentGames.map((game) => (
          <GameCard key={game.id} {...game} />
        ))}
      </div>
    </section>
  )
}
