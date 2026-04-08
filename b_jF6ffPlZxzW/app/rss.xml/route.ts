import { fetchGameList } from "@/lib/api/game"
import { fetchPublicSystemConfig } from "@/lib/api/system"
import { getSiteOrigin } from "@/lib/seo/site"

export const revalidate = 300

function xmlEscape(s: string) {
  return (s ?? "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&apos;")
}

export async function GET() {
  const origin = await getSiteOrigin()
  const cfg = await fetchPublicSystemConfig().catch(() => ({
    site_title: "游戏资源站",
    seo_description: "",
    logo_url: "",
    favicon_url: "",
    seo_keywords: "",
  }))

  const siteTitle = (cfg.site_title || "游戏资源站").trim() || "游戏资源站"
  const siteDesc = (cfg.seo_description || "").trim()

  const res = await fetchGameList({ page: 1, page_size: 30 })
  const items = (res.list || []).slice(0, 30)

  const now = new Date().toUTCString()
  const channel =
    `<channel>` +
    `<title>${xmlEscape(siteTitle)}</title>` +
    `<link>${xmlEscape(origin + "/")}</link>` +
    `<description>${xmlEscape(siteDesc || siteTitle)}</description>` +
    `<lastBuildDate>${xmlEscape(now)}</lastBuildDate>` +
    items
      .map((g) => {
        const link = `${origin}/${g.id}`
        const pubDate = g.updated_at ? new Date(g.updated_at).toUTCString() : now
        const desc = (g.short_description || g.description || "").toString()
        return (
          `<item>` +
          `<title>${xmlEscape(g.title)}</title>` +
          `<link>${xmlEscape(link)}</link>` +
          `<guid isPermaLink="true">${xmlEscape(link)}</guid>` +
          `<pubDate>${xmlEscape(pubDate)}</pubDate>` +
          `<description>${xmlEscape(desc)}</description>` +
          `</item>`
        )
      })
      .join("") +
    `</channel>`

  const xml =
    `<?xml version="1.0" encoding="UTF-8"?>` +
    `<rss version="2.0">` +
    channel +
    `</rss>`

  return new Response(xml, {
    headers: {
      "Content-Type": "application/rss+xml; charset=utf-8",
      "Cache-Control": "public, max-age=300",
    },
  })
}

