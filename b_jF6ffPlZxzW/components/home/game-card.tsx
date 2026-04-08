"use client"

import { useState } from "react"
import Image from "next/image"
import Link from "next/link"
import { Heart, ShoppingCart, Star, Clock } from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

interface GameCardProps {
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

export function GameCard({
  id,
  title,
  image,
  price,
  originalPrice,
  discount,
  rating,
  releaseDate,
  tags = [],
  isNew,
  isTrending,
}: GameCardProps) {
  const [isWishlisted, setIsWishlisted] = useState(false)
  const [isHovered, setIsHovered] = useState(false)

  return (
    <div
      className="group relative flex flex-col overflow-hidden rounded-xl bg-card border border-border transition-all duration-300 hover:border-primary/50 hover:shadow-lg hover:shadow-primary/5"
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      {/* Image Container */}
      <Link href={`/${id}`} className="relative aspect-[3/4] overflow-hidden block">
        <Image
          src={image}
          alt={title}
          fill
          className={cn("object-cover transition-transform duration-500", isHovered && "scale-110")}
        />
        
        {/* Badges */}
        <div className="absolute top-3 left-3 flex flex-col gap-2">
          {discount && (
            <span className="rounded-md bg-primary px-2 py-1 text-xs font-bold text-primary-foreground">
              -{discount}%
            </span>
          )}
          {isNew && (
            <span className="rounded-md bg-blue-500 px-2 py-1 text-xs font-bold text-foreground">
              新品
            </span>
          )}
          {isTrending && (
            <span className="rounded-md bg-orange-500 px-2 py-1 text-xs font-bold text-foreground">
              热门
            </span>
          )}
        </div>

        {/* Wishlist Button */}
        <button
          onClick={() => setIsWishlisted(!isWishlisted)}
          className={cn(
            "absolute top-3 right-3 flex h-8 w-8 items-center justify-center rounded-full transition-all duration-300",
            isWishlisted 
              ? "bg-primary text-primary-foreground" 
              : "bg-background/50 text-foreground backdrop-blur-sm hover:bg-background/80"
          )}
        >
          <Heart className={cn("h-4 w-4", isWishlisted && "fill-current")} />
        </button>

        {/* Hover Overlay */}
        <div className={cn(
          "absolute inset-0 flex items-end bg-gradient-to-t from-background via-background/50 to-transparent p-4 transition-opacity duration-300",
          isHovered ? "opacity-100" : "opacity-0"
        )}>
          <Button className="w-full gap-2" size="sm">
            <ShoppingCart className="h-4 w-4" />
            加入购物车
          </Button>
        </div>
      </Link>

      {/* Content */}
      <div className="flex flex-1 flex-col p-4">
        {/* Tags */}
        {tags.length > 0 && (
          <div className="flex flex-wrap gap-1 mb-2">
            {tags.slice(0, 2).map((tag) => (
              <span
                key={tag}
                className="rounded-full bg-secondary px-2 py-0.5 text-[10px] font-medium text-muted-foreground"
              >
                {tag}
              </span>
            ))}
          </div>
        )}

        {/* Title */}
        <Link href={`/${id}`} className="font-semibold text-foreground line-clamp-2 mb-2 group-hover:text-primary transition-colors">
          {title}
        </Link>

        {/* Meta Info */}
        <div className="mt-auto flex items-center gap-3 text-xs text-muted-foreground">
          {rating && (
            <div className="flex items-center gap-1">
              <Star className="h-3 w-3 fill-primary text-primary" />
              <span>{rating.toFixed(1)}</span>
            </div>
          )}
          {releaseDate && (
            <div className="flex items-center gap-1">
              <Clock className="h-3 w-3" />
              <span>{releaseDate}</span>
            </div>
          )}
        </div>

        {/* Price */}
        <div className="mt-3 flex items-center justify-between border-t border-border pt-3">
          <div className="flex items-baseline gap-2">
            {originalPrice && (
              <span className="text-xs text-muted-foreground line-through">
                ¥{originalPrice}
              </span>
            )}
            <span className="text-lg font-bold text-foreground">
              ¥{price}
            </span>
          </div>
        </div>
      </div>
    </div>
  )
}
