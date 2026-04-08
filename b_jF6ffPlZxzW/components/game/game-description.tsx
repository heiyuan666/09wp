"use client"

import { useState } from "react"
import { ChevronDown, ChevronUp } from "lucide-react"
import { cn } from "@/lib/utils"

interface GameDescriptionProps {
  shortDescription: string
  fullDescription: string
  features: string[]
}

export function GameDescription({ shortDescription, fullDescription, features }: GameDescriptionProps) {
  const [isExpanded, setIsExpanded] = useState(false)

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-xl font-semibold text-foreground mb-3">游戏介绍</h3>
        <p className="text-muted-foreground leading-relaxed">{shortDescription}</p>
        
        <div className={cn(
          "overflow-hidden transition-all duration-500",
          isExpanded ? "max-h-[1000px] mt-4" : "max-h-0"
        )}>
          <p className="text-muted-foreground leading-relaxed whitespace-pre-line">
            {fullDescription}
          </p>
        </div>
        
        <button
          onClick={() => setIsExpanded(!isExpanded)}
          className="mt-3 flex items-center gap-1 text-sm font-medium text-primary hover:text-primary/80 transition-colors"
        >
          {isExpanded ? (
            <>
              收起 <ChevronUp className="h-4 w-4" />
            </>
          ) : (
            <>
              展开更多 <ChevronDown className="h-4 w-4" />
            </>
          )}
        </button>
      </div>

      <div>
        <h4 className="text-lg font-semibold text-foreground mb-3">游戏特色</h4>
        <ul className="grid gap-2 sm:grid-cols-2">
          {features.map((feature, index) => (
            <li key={index} className="flex items-start gap-2">
              <span className="mt-1.5 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-primary" />
              <span className="text-sm text-muted-foreground">{feature}</span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}
