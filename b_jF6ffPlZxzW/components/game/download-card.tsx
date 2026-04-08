"use client"

import { useState } from "react"
import { 
  Download, 
  Heart, 
  Share2, 
  Copy, 
  Check, 
  ExternalLink,
  HardDrive,
  Cloud,
  FileDown,
  ChevronDown,
  Info,
  AlertCircle
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

type PanType = "quark" | "aliyun" | "baidu" | "lanzou" | "123pan" | "tianyi" | "mega" | "onedrive"

interface DownloadLink {
  id: string
  name: string
  type: PanType
  url: string
  password?: string
  size: string
  speed: "fast" | "medium" | "slow"
  isRecommended?: boolean
}

interface DownloadCardProps {
  fileSize: string
  updateDate: string
  version: string
  downloads?: DownloadLink[]
}

// 网盘颜色和标签配置
const panConfig: Record<PanType, { color: string; label: string; bgColor: string }> = {
  quark: { color: "#6366F1", label: "夸", bgColor: "bg-indigo-500" },
  aliyun: { color: "#FF6A00", label: "阿", bgColor: "bg-orange-500" },
  baidu: { color: "#2196F3", label: "百", bgColor: "bg-blue-500" },
  lanzou: { color: "#10B981", label: "蓝", bgColor: "bg-emerald-500" },
  "123pan": { color: "#F59E0B", label: "123", bgColor: "bg-amber-500" },
  tianyi: { color: "#3B82F6", label: "天", bgColor: "bg-blue-600" },
  mega: { color: "#D93025", label: "M", bgColor: "bg-red-600" },
  onedrive: { color: "#0078D4", label: "O", bgColor: "bg-sky-600" },
}

// 网盘图标组件
function PanIcon({ type, className }: { type: PanType; className?: string }) {
  const config = panConfig[type]
  return (
    <div className={cn(
      "flex items-center justify-center rounded-lg text-white font-bold",
      config.bgColor,
      className
    )}>
      <span className={type === "123pan" ? "text-[10px]" : "text-sm"}>{config.label}</span>
    </div>
  )
}

const defaultDownloads: DownloadLink[] = [
  {
    id: "quark",
    name: "夸克网盘",
    type: "quark",
    url: "https://pan.quark.cn/s/xxxxx",
    password: "abcd",
    size: "65.2 GB",
    speed: "fast",
    isRecommended: true,
  },
  {
    id: "aliyun",
    name: "阿里云盘",
    type: "aliyun",
    url: "https://www.aliyundrive.com/s/xxxxx",
    size: "65.2 GB",
    speed: "fast",
  },
  {
    id: "baidu",
    name: "百度网盘",
    type: "baidu",
    url: "https://pan.baidu.com/s/xxxxx",
    password: "1234",
    size: "65.2 GB",
    speed: "slow",
  },
  {
    id: "lanzou",
    name: "蓝奏云",
    type: "lanzou",
    url: "https://lanzou.com/xxxxx",
    password: "game",
    size: "65.2 GB",
    speed: "medium",
  },
  {
    id: "123pan",
    name: "123云盘",
    type: "123pan",
    url: "https://www.123pan.com/s/xxxxx",
    size: "65.2 GB",
    speed: "medium",
  },
  {
    id: "tianyi",
    name: "天翼云盘",
    type: "tianyi",
    url: "https://cloud.189.cn/t/xxxxx",
    password: "ty88",
    size: "65.2 GB",
    speed: "medium",
  },
]

export function DownloadCard({ 
  fileSize, 
  updateDate, 
  version,
  downloads = defaultDownloads 
}: DownloadCardProps) {
  const [isWishlisted, setIsWishlisted] = useState(false)
  const [copiedId, setCopiedId] = useState<string | null>(null)
  const [showAllLinks, setShowAllLinks] = useState(false)

  const copyPassword = (id: string, password: string) => {
    navigator.clipboard.writeText(password)
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  const copyLink = (url: string, password?: string) => {
    const text = password ? `${url}\n提取码: ${password}` : url
    navigator.clipboard.writeText(text)
  }

  const speedColors = {
    fast: "text-primary bg-primary/10",
    medium: "text-yellow-500 bg-yellow-500/10",
    slow: "text-red-500 bg-red-500/10",
  }

  const speedLabels = {
    fast: "高速",
    medium: "中速",
    slow: "限速",
  }

  const displayedDownloads = showAllLinks ? downloads : downloads.slice(0, 4)

  return (
    <div className="rounded-xl bg-card border border-border overflow-hidden shadow-lg">
      {/* Header */}
      <div className="bg-gradient-to-r from-primary/20 to-primary/5 border-b border-border p-5">
        <div className="flex items-center gap-3 mb-3">
          <div className="p-2 rounded-lg bg-primary/20">
            <FileDown className="h-5 w-5 text-primary" />
          </div>
          <div>
            <h3 className="font-bold text-lg text-foreground">资源下载</h3>
            <p className="text-xs text-muted-foreground">多网盘高速下载</p>
          </div>
        </div>
        <div className="grid grid-cols-3 gap-3 text-sm">
          <div className="flex items-center gap-2 p-2 rounded-lg bg-background/50">
            <HardDrive className="h-4 w-4 text-primary" />
            <div>
              <p className="text-[10px] text-muted-foreground">文件大小</p>
              <p className="font-medium text-foreground">{fileSize}</p>
            </div>
          </div>
          <div className="flex items-center gap-2 p-2 rounded-lg bg-background/50">
            <Cloud className="h-4 w-4 text-primary" />
            <div>
              <p className="text-[10px] text-muted-foreground">版本</p>
              <p className="font-medium text-foreground">{version}</p>
            </div>
          </div>
          <div className="flex items-center gap-2 p-2 rounded-lg bg-background/50">
            <Info className="h-4 w-4 text-primary" />
            <div>
              <p className="text-[10px] text-muted-foreground">更新</p>
              <p className="font-medium text-foreground">{updateDate}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Download Links */}
      <div className="p-4 space-y-2">
        {displayedDownloads.map((link) => (
          <div
            key={link.id}
            className={cn(
              "group rounded-lg border p-3 transition-all duration-200 hover:shadow-md",
              link.isRecommended 
                ? "border-primary/50 bg-primary/5 hover:bg-primary/10" 
                : "border-border bg-secondary/30 hover:bg-secondary/50"
            )}
          >
            <div className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-3 min-w-0">
                <PanIcon type={link.type} className="h-10 w-10 flex-shrink-0" />
                <div className="min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="font-semibold text-foreground">{link.name}</span>
                    {link.isRecommended && (
                      <span className="rounded-full bg-primary px-2 py-0.5 text-[10px] font-bold text-primary-foreground animate-pulse">
                        推荐
                      </span>
                    )}
                    <span className={cn(
                      "rounded-full px-2 py-0.5 text-[10px] font-medium",
                      speedColors[link.speed]
                    )}>
                      {speedLabels[link.speed]}
                    </span>
                  </div>
                  <div className="flex items-center gap-2 mt-0.5">
                    <span className="text-xs text-muted-foreground">{link.size}</span>
                    {link.password && (
                      <span className="text-xs text-muted-foreground">
                        提取码: <span className="font-mono text-foreground">{link.password}</span>
                      </span>
                    )}
                  </div>
                </div>
              </div>
              
              <div className="flex items-center gap-2 flex-shrink-0">
                {link.password && (
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 px-2 text-xs hidden sm:flex"
                    onClick={() => copyPassword(link.id, link.password!)}
                  >
                    {copiedId === link.id ? (
                      <>
                        <Check className="mr-1 h-3 w-3 text-primary" />
                        已复制
                      </>
                    ) : (
                      <>
                        <Copy className="mr-1 h-3 w-3" />
                        复制码
                      </>
                    )}
                  </Button>
                )}
                <Button
                  size="sm"
                  className={cn(
                    "h-9 px-4 font-medium",
                    link.isRecommended 
                      ? "bg-primary text-primary-foreground hover:bg-primary/90 shadow-md shadow-primary/25" 
                      : "bg-secondary text-foreground hover:bg-secondary/80"
                  )}
                  onClick={() => copyLink(link.url, link.password)}
                  asChild
                >
                  <a href={link.url} target="_blank" rel="noopener noreferrer">
                    <Download className="mr-1.5 h-4 w-4" />
                    下载
                    <ExternalLink className="ml-1 h-3 w-3 opacity-50" />
                  </a>
                </Button>
              </div>
            </div>
          </div>
        ))}

        {/* Show More Button */}
        {downloads.length > 4 && (
          <Button
            variant="ghost"
            className="w-full text-muted-foreground hover:text-foreground mt-2"
            onClick={() => setShowAllLinks(!showAllLinks)}
          >
            <ChevronDown className={cn(
              "mr-2 h-4 w-4 transition-transform duration-200",
              showAllLinks && "rotate-180"
            )} />
            {showAllLinks ? "收起更多" : `展开更多网盘 (${downloads.length - 4})`}
          </Button>
        )}
      </div>

      {/* Actions */}
      <div className="border-t border-border p-4 bg-secondary/20">
        <div className="grid grid-cols-2 gap-3">
          <Button 
            variant="outline"
            className={cn(
              "h-11 border-border",
              isWishlisted && "bg-primary/10 border-primary/50 text-primary"
            )}
            onClick={() => setIsWishlisted(!isWishlisted)}
          >
            <Heart className={cn("mr-2 h-4 w-4", isWishlisted && "fill-current")} />
            {isWishlisted ? "已收藏" : "收藏游戏"}
          </Button>
          <Button variant="outline" className="h-11 border-border">
            <Share2 className="mr-2 h-4 w-4" />
            分享资源
          </Button>
        </div>
      </div>

      {/* Tips */}
      <div className="border-t border-border bg-amber-500/5 px-4 py-3">
        <div className="flex gap-2">
          <AlertCircle className="h-4 w-4 text-amber-500 flex-shrink-0 mt-0.5" />
          <p className="text-xs text-muted-foreground leading-relaxed">
            推荐使用<span className="text-primary font-medium">夸克网盘</span>或<span className="text-primary font-medium">阿里云盘</span>下载，速度更快且无需会员。如遇链接失效请在评论区反馈，我们会及时补链。
          </p>
        </div>
      </div>
    </div>
  )
}
