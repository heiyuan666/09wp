import { apiGet } from "@/lib/api/client"

export type PublicNavMenuItem = {
  id: number
  title: string
  path: string
  position: string
  sort_order: number
  visible: boolean
}

export async function fetchPublicNavMenus(position: "top_nav" | "home_promo" = "top_nav") {
  // 游戏站点：独立导航菜单接口
  const data = await apiGet<{ list: PublicNavMenuItem[] }>(
    `/game/public/nav-menus?position=${encodeURIComponent(position)}`,
    { cache: "force-cache" },
  )
  return Array.isArray(data.list) ? data.list : []
}

