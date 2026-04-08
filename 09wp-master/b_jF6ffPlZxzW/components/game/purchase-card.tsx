"use client"

import { useState } from "react"
import { ShoppingCart, Heart, Share2, Gift, Check } from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

interface PurchaseCardProps {
  price: number
  originalPrice?: number
  discount?: number
}

export function PurchaseCard({ price, originalPrice, discount }: PurchaseCardProps) {
  const [isWishlisted, setIsWishlisted] = useState(false)
  const [isAddedToCart, setIsAddedToCart] = useState(false)

  return (
    <div className="rounded-lg bg-card p-6 space-y-4">
      {/* Price */}
      <div className="space-y-1">
        {discount && originalPrice && (
          <div className="flex items-center gap-2">
            <span className="rounded bg-primary px-2 py-1 text-sm font-bold text-primary-foreground">
              -{discount}%
            </span>
            <span className="text-muted-foreground line-through">
              ¥{originalPrice.toFixed(2)}
            </span>
          </div>
        )}
        <div className="text-3xl font-bold text-foreground">
          ¥{price.toFixed(2)}
        </div>
      </div>

      {/* Buttons */}
      <div className="space-y-2">
        <Button 
          className="w-full bg-primary text-primary-foreground hover:bg-primary/90"
          size="lg"
          onClick={() => setIsAddedToCart(true)}
        >
          {isAddedToCart ? (
            <>
              <Check className="mr-2 h-5 w-5" />
              已加入购物车
            </>
          ) : (
            <>
              <ShoppingCart className="mr-2 h-5 w-5" />
              加入购物车
            </>
          )}
        </Button>
        
        <div className="grid grid-cols-2 gap-2">
          <Button 
            variant="secondary"
            className={cn(
              "bg-secondary text-secondary-foreground hover:bg-secondary/80",
              isWishlisted && "bg-primary/20 text-primary"
            )}
            onClick={() => setIsWishlisted(!isWishlisted)}
          >
            <Heart className={cn("mr-2 h-4 w-4", isWishlisted && "fill-current")} />
            {isWishlisted ? "已收藏" : "收藏"}
          </Button>
          <Button variant="secondary" className="bg-secondary text-secondary-foreground hover:bg-secondary/80">
            <Gift className="mr-2 h-4 w-4" />
            赠送
          </Button>
        </div>
        
        <Button variant="ghost" className="w-full text-muted-foreground hover:text-foreground">
          <Share2 className="mr-2 h-4 w-4" />
          分享
        </Button>
      </div>

      {/* Features */}
      <div className="border-t border-border pt-4 space-y-2">
        <p className="text-sm text-muted-foreground">
          <Check className="mr-2 inline h-4 w-4 text-primary" />
          支持云存档
        </p>
        <p className="text-sm text-muted-foreground">
          <Check className="mr-2 inline h-4 w-4 text-primary" />
          手柄支持
        </p>
        <p className="text-sm text-muted-foreground">
          <Check className="mr-2 inline h-4 w-4 text-primary" />
          成就系统
        </p>
      </div>
    </div>
  )
}
