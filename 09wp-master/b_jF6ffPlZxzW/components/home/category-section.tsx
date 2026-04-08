"use client"

import Link from "next/link"
import { 
  Sword, 
  Car, 
  Ghost, 
  Users, 
  Puzzle, 
  Trophy, 
  Wand2, 
  Rocket 
} from "lucide-react"
import { cn } from "@/lib/utils"

export interface CategoryItem {
  id: string
  name: string
  icon: "action" | "racing" | "horror" | "multiplayer" | "puzzle" | "sports" | "rpg" | "scifi"
  count?: number
  color: string
  queryValue?: string
}

const fallbackCategories: CategoryItem[] = [
  { id: "action", name: "动作", icon: "action", count: 0, color: "bg-red-500/10 text-red-500 hover:bg-red-500/20" },
  { id: "racing", name: "竞速", icon: "racing", count: 0, color: "bg-blue-500/10 text-blue-500 hover:bg-blue-500/20" },
  { id: "horror", name: "恐怖", icon: "horror", count: 0, color: "bg-purple-500/10 text-purple-500 hover:bg-purple-500/20" },
  { id: "multiplayer", name: "多人", icon: "multiplayer", count: 0, color: "bg-green-500/10 text-green-500 hover:bg-green-500/20" },
  { id: "puzzle", name: "解谜", icon: "puzzle", count: 0, color: "bg-yellow-500/10 text-yellow-500 hover:bg-yellow-500/20" },
  { id: "sports", name: "体育", icon: "sports", count: 0, color: "bg-orange-500/10 text-orange-500 hover:bg-orange-500/20" },
  { id: "rpg", name: "角色扮演", icon: "rpg", count: 0, color: "bg-pink-500/10 text-pink-500 hover:bg-pink-500/20" },
  { id: "scifi", name: "科幻", icon: "scifi", count: 0, color: "bg-cyan-500/10 text-cyan-500 hover:bg-cyan-500/20" },
]

export function CategorySection({
  categories,
  selectedCategory,
}: {
  categories: CategoryItem[]
  selectedCategory?: string
}) {
  const list = categories?.length ? categories : fallbackCategories
  const iconMap = {
    action: Sword,
    racing: Car,
    horror: Ghost,
    multiplayer: Users,
    puzzle: Puzzle,
    sports: Trophy,
    rpg: Wand2,
    scifi: Rocket,
  } as const
  return (
    <section>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-foreground">游戏分类</h2>
        <button className="text-sm font-medium text-primary hover:underline">
          查看全部
        </button>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-8 gap-4">
        {list.map((category) => {
          const Icon = iconMap[category.icon]
          const active = !!selectedCategory && selectedCategory === (category.queryValue || category.id)
          return (
            <Link
              key={category.id}
              href={`/?category=${encodeURIComponent(category.queryValue || category.id)}`}
              className={cn(
                "flex flex-col items-center gap-3 rounded-xl p-4 transition-all duration-300",
                category.color,
                active && "ring-2 ring-primary ring-offset-2 ring-offset-background"
              )}
            >
              <Icon className="h-8 w-8" />
              <div className="text-center">
                <p className="font-medium text-sm">{category.name}</p>
                {typeof category.count === "number" && (
                  <p className="text-xs opacity-70">{category.count} 款游戏</p>
                )}
              </div>
            </Link>
          )
        })}
      </div>
    </section>
  )
}
