"use client"

import { useState } from "react"
import { X, Megaphone, ArrowRight } from "lucide-react"
import Link from "next/link"
import { cn } from "@/lib/utils"

interface AnnouncementBannerProps {
  message: string
  link?: string
  linkText?: string
}

export function AnnouncementBanner({ 
  message = "本站所有资源仅供学习交流使用，请于下载后24小时内删除，支持正版游戏！",
  link,
  linkText = "了解更多"
}: AnnouncementBannerProps) {
  const [isVisible, setIsVisible] = useState(true)

  if (!isVisible) return null

  return (
    <div className="relative bg-primary/10 border-b border-primary/20">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-center gap-3 py-2.5 text-sm">
          <Megaphone className="h-4 w-4 text-primary flex-shrink-0" />
          <p className="text-foreground/80">
            {message}
          </p>
          {link && (
            <Link 
              href={link}
              className="inline-flex items-center gap-1 font-medium text-primary hover:text-primary/80 transition-colors"
            >
              {linkText}
              <ArrowRight className="h-3 w-3" />
            </Link>
          )}
        </div>
      </div>
      <button
        onClick={() => setIsVisible(false)}
        className={cn(
          "absolute right-2 top-1/2 -translate-y-1/2 p-1.5 rounded-full",
          "text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors"
        )}
      >
        <X className="h-4 w-4" />
      </button>
    </div>
  )
}
