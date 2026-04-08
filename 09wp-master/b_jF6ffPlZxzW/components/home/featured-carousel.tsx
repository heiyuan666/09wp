"use client"

import { useState, useEffect, useCallback } from "react"
import Image from "next/image"
import Link from "next/link"
import { ChevronLeft, ChevronRight, Play } from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

export interface FeaturedGame {
  id: number
  title: string
  subtitle: string
  description: string
  image: string
  price: string
  discount?: number
  tags: string[]
}

export function FeaturedCarousel({ games }: { games: FeaturedGame[] }) {
  const [currentIndex, setCurrentIndex] = useState(0)
  const [isAutoPlaying, setIsAutoPlaying] = useState(true)

  const nextSlide = useCallback(() => {
    if (!games?.length) return
    setCurrentIndex((prev) => (prev + 1) % games.length)
  }, [games])

  const prevSlide = () => {
    if (!games?.length) return
    setCurrentIndex((prev) => (prev - 1 + games.length) % games.length)
  }

  useEffect(() => {
    if (!isAutoPlaying) return
    const interval = setInterval(nextSlide, 5000)
    return () => clearInterval(interval)
  }, [isAutoPlaying, nextSlide])

  const currentGame = games?.[currentIndex]
  const tags = currentGame?.tags ?? []

  if (!currentGame) {
    return null
  }

  return (
    <section 
      className="relative"
      onMouseEnter={() => setIsAutoPlaying(false)}
      onMouseLeave={() => setIsAutoPlaying(true)}
    >
      {/* Main Hero */}
      <div className="relative h-[500px] md:h-[600px] lg:h-[700px] overflow-hidden rounded-2xl">
        {games.map((game, index) => (
          <div
            key={game.id}
            className={cn(
              "absolute inset-0 transition-all duration-700 ease-in-out",
              index === currentIndex ? "opacity-100 scale-100" : "opacity-0 scale-105"
            )}
          >
            <Image
              src={game.image}
              alt={game.title}
              fill
              className="object-cover"
              priority={index === 0}
            />
            {/* Gradient Overlay */}
            <div className="absolute inset-0 bg-gradient-to-r from-background via-background/60 to-transparent" />
            <div className="absolute inset-0 bg-gradient-to-t from-background via-transparent to-transparent" />
          </div>
        ))}

        {/* Content */}
        <div className="absolute inset-0 flex items-center">
          <div className="mx-auto max-w-7xl w-full px-4 sm:px-6 lg:px-8">
            <div className="max-w-xl">
              {/* Tags */}
              <div className="flex flex-wrap gap-2 mb-4">
                {tags.map((tag) => (
                  <span
                    key={tag}
                    className="rounded-full bg-primary/20 px-3 py-1 text-xs font-medium text-primary"
                  >
                    {tag}
                  </span>
                ))}
              </div>

              {/* Title */}
              <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold text-foreground mb-2 text-balance">
                {currentGame.title}
              </h1>
              <p className="text-xl md:text-2xl text-primary font-medium mb-4">
                {currentGame.subtitle}
              </p>

              {/* Description */}
              <p className="text-muted-foreground text-base md:text-lg mb-6 text-pretty">
                {currentGame.description}
              </p>

              {/* Price & CTA */}
              <div className="flex flex-wrap items-center gap-4">
                <div className="flex items-center gap-3">
                  {currentGame.discount && (
                    <span className="rounded-md bg-primary px-2 py-1 text-sm font-bold text-primary-foreground">
                      -{currentGame.discount}%
                    </span>
                  )}
                  <div className="flex items-baseline gap-2">
                    {currentGame.discount && (
                      <span className="text-muted-foreground line-through text-sm">
                        ¥{currentGame.price}
                      </span>
                    )}
                    <span className="text-2xl font-bold text-foreground">
                      ¥{currentGame.discount 
                        ? Math.round(Number(currentGame.price) * (1 - currentGame.discount / 100)) 
                        : currentGame.price}
                    </span>
                  </div>
                </div>
                <Button size="lg" className="gap-2" asChild>
                  <Link href={`/${currentGame.id}#download`}>
                    <Play className="h-4 w-4" />
                    立即下载
                  </Link>
                </Button>
                <Button size="lg" variant="secondary">
                  了解更多
                </Button>
              </div>
            </div>
          </div>
        </div>

        {/* Navigation Arrows */}
        <div className="absolute inset-y-0 left-4 flex items-center">
          <Button
            variant="secondary"
            size="icon"
            className="h-12 w-12 rounded-full bg-background/50 backdrop-blur-sm hover:bg-background/80"
            onClick={prevSlide}
          >
            <ChevronLeft className="h-6 w-6" />
          </Button>
        </div>
        <div className="absolute inset-y-0 right-4 flex items-center">
          <Button
            variant="secondary"
            size="icon"
            className="h-12 w-12 rounded-full bg-background/50 backdrop-blur-sm hover:bg-background/80"
            onClick={nextSlide}
          >
            <ChevronRight className="h-6 w-6" />
          </Button>
        </div>

        {/* Dots Indicator */}
        <div className="absolute bottom-6 left-1/2 -translate-x-1/2 flex gap-2">
          {games.map((_, index) => (
            <button
              key={index}
              onClick={() => setCurrentIndex(index)}
              className={cn(
                "h-2 rounded-full transition-all duration-300",
                index === currentIndex 
                  ? "w-8 bg-primary" 
                  : "w-2 bg-foreground/30 hover:bg-foreground/50"
              )}
            />
          ))}
        </div>
      </div>

      {/* Thumbnail Strip */}
      <div className="mt-4 flex gap-4 overflow-x-auto pb-2 scrollbar-hide">
        {games.map((game, index) => (
          <button
            key={game.id}
            onClick={() => setCurrentIndex(index)}
            className={cn(
              "relative flex-shrink-0 overflow-hidden rounded-lg transition-all duration-300",
              index === currentIndex 
                ? "ring-2 ring-primary ring-offset-2 ring-offset-background" 
                : "opacity-60 hover:opacity-100"
            )}
          >
            <div className="relative h-20 w-36">
              <Image
                src={game.image}
                alt={game.title}
                fill
                className="object-cover"
              />
            </div>
            <div className="absolute inset-0 bg-gradient-to-t from-background/80 to-transparent" />
            <span className="absolute bottom-2 left-2 text-xs font-medium text-foreground">
              {game.title}
            </span>
          </button>
        ))}
      </div>
    </section>
  )
}
