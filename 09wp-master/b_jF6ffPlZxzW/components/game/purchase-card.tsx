"use client"

import { useState } from "react"
import { Download, Heart, Share2, Gift, Check } from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { copyTextToClipboard, tryNativeShare } from "@/lib/share"

interface PurchaseCardProps {
  price: number
  originalPrice?: number
  discount?: number
  share?: { title: string; text?: string; url: string }
}

export function PurchaseCard({ price, originalPrice, discount, share }: PurchaseCardProps) {
  const [isWishlisted, setIsWishlisted] = useState(false)
  const [clicked, setClicked] = useState(false)
  const [shareMsg, setShareMsg] = useState("")

  const shareCurrentPage = async () => {
    setShareMsg("")
    try {
      const url = share?.url?.trim() || (typeof window !== "undefined" ? window.location.href : "")
      const title = share?.title?.trim() || (typeof document !== "undefined" ? document.title : "分享")
      const text = (share?.text || "").trim()

      if (await tryNativeShare({ title, text, url })) {
        setShareMsg("已打开系统分享")
        return
      }
      const ok = await copyTextToClipboard(url)
      if (ok) {
        setShareMsg("已复制页面链接")
        return
      }
      setShareMsg("无法获取当前页面链接")
    } catch {
      setShareMsg("分享失败")
    }
  }

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
          onClick={() => setClicked(true)}
        >
          {clicked ? (
            <>
              <Check className="mr-2 h-5 w-5" />
              已准备下载
            </>
          ) : (
            <>
              <Download className="mr-2 h-5 w-5" />
              立即下载
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
        
        <Button variant="ghost" className="w-full text-muted-foreground hover:text-foreground" onClick={shareCurrentPage}>
          <Share2 className="mr-2 h-4 w-4" />
          分享
        </Button>
        {shareMsg ? <div className="text-xs text-muted-foreground">{shareMsg}</div> : null}
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
