import { apiGet } from "@/lib/api/client"

export type PublicSystemConfig = {
  site_title: string
  logo_url: string
  favicon_url: string
  seo_keywords: string
  seo_description: string
}

export async function fetchPublicSystemConfig() {
  // 游戏站点：独立配置接口
  return apiGet<PublicSystemConfig>("/game/public/config", { cache: "force-cache" })
}

