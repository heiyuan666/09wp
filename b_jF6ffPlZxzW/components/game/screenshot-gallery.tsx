"use client"

import { useState } from "react"
import Image from "next/image"
import { ChevronLeft, ChevronRight, X } from "lucide-react"
import { cn } from "@/lib/utils"

interface ScreenshotGalleryProps {
  screenshots: string[]
}

export function ScreenshotGallery({ screenshots }: ScreenshotGalleryProps) {
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [isLightboxOpen, setIsLightboxOpen] = useState(false)

  const goToNext = () => {
    setSelectedIndex((prev) => (prev + 1) % screenshots.length)
  }

  const goToPrev = () => {
    setSelectedIndex((prev) => (prev - 1 + screenshots.length) % screenshots.length)
  }

  return (
    <>
      <div className="space-y-4">
        {/* Main preview */}
        <div 
          className="relative aspect-video w-full overflow-hidden rounded-lg cursor-pointer group"
          onClick={() => setIsLightboxOpen(true)}
        >
          <Image
            src={screenshots[selectedIndex]}
            alt={`游戏截图 ${selectedIndex + 1}`}
            fill
            className="object-cover transition-transform duration-500 group-hover:scale-105"
          />
          <div className="absolute inset-0 bg-background/20 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
            <span className="text-foreground text-sm font-medium">点击放大</span>
          </div>
        </div>

        {/* Thumbnail strip */}
        <div className="flex gap-2 overflow-x-auto pb-2 scrollbar-hide">
          {screenshots.map((screenshot, index) => (
            <button
              key={index}
              onClick={() => setSelectedIndex(index)}
              className={cn(
                "relative h-16 w-28 flex-shrink-0 overflow-hidden rounded-md transition-all",
                selectedIndex === index 
                  ? "ring-2 ring-primary" 
                  : "opacity-60 hover:opacity-100"
              )}
            >
              <Image
                src={screenshot}
                alt={`缩略图 ${index + 1}`}
                fill
                className="object-cover"
              />
            </button>
          ))}
        </div>
      </div>

      {/* Lightbox */}
      {isLightboxOpen && (
        <div 
          className="fixed inset-0 z-50 flex items-center justify-center bg-background/95"
          onClick={() => setIsLightboxOpen(false)}
        >
          <button 
            className="absolute right-4 top-4 text-foreground/70 hover:text-foreground transition-colors"
            onClick={() => setIsLightboxOpen(false)}
          >
            <X className="h-8 w-8" />
          </button>
          
          <button 
            className="absolute left-4 top-1/2 -translate-y-1/2 flex h-12 w-12 items-center justify-center rounded-full bg-secondary text-foreground hover:bg-secondary/80 transition-colors"
            onClick={(e) => { e.stopPropagation(); goToPrev(); }}
          >
            <ChevronLeft className="h-6 w-6" />
          </button>
          
          <div 
            className="relative h-[80vh] w-[90vw] max-w-6xl"
            onClick={(e) => e.stopPropagation()}
          >
            <Image
              src={screenshots[selectedIndex]}
              alt={`游戏截图 ${selectedIndex + 1}`}
              fill
              className="object-contain"
            />
          </div>
          
          <button 
            className="absolute right-4 top-1/2 -translate-y-1/2 flex h-12 w-12 items-center justify-center rounded-full bg-secondary text-foreground hover:bg-secondary/80 transition-colors"
            onClick={(e) => { e.stopPropagation(); goToNext(); }}
          >
            <ChevronRight className="h-6 w-6" />
          </button>

          <div className="absolute bottom-4 left-1/2 -translate-x-1/2 text-foreground/70">
            {selectedIndex + 1} / {screenshots.length}
          </div>
        </div>
      )}
    </>
  )
}
