import { fetchGameCategoryList, fetchGameList } from "@/lib/api/game"
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

  const urls: Array<{ loc: string; lastmod?: string }> = []

  // static
  urls.push({ loc: `${origin}/` })
  urls.push({ loc: `${origin}/category` })

  // categories
  try {
    const cats = await fetchGameCategoryList()
    for (const c of cats) {
      urls.push({ loc: `${origin}/category/${c.id}`, lastmod: c.updated_at })
    }
  } catch {
    // ignore
  }

  // games (paged)
  const pageSize = 200
  let page = 1
  let seen = 0
  const hardCap = 5000

  while (seen < hardCap) {
    const res = await fetchGameList({ page, page_size: pageSize })
    for (const g of res.list || []) {
      urls.push({ loc: `${origin}/${g.id}`, lastmod: g.updated_at })
      seen++
      if (seen >= hardCap) break
    }
    if (!res.list || res.list.length < pageSize) break
    page++
  }

  const body =
    `<?xml version="1.0" encoding="UTF-8"?>` +
    `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` +
    urls
      .map((u) => {
        const lastmod = u.lastmod ? `<lastmod>${xmlEscape(new Date(u.lastmod).toISOString())}</lastmod>` : ""
        return `<url><loc>${xmlEscape(u.loc)}</loc>${lastmod}</url>`
      })
      .join("") +
    `</urlset>`

  return new Response(body, {
    headers: {
      "Content-Type": "application/xml; charset=utf-8",
      "Cache-Control": "public, max-age=300",
    },
  })
}

