import type { Metadata } from 'next'
import { Geist, Geist_Mono } from 'next/font/google'
import { Analytics } from '@vercel/analytics/next'
import './globals.css'
import { fetchPublicSystemConfig } from '@/lib/api/system'
import { getSiteOrigin } from '@/lib/seo/site'
import { JsonLd } from '@/components/seo/jsonld'

export const revalidate = 300

const _geist = Geist({ subsets: ["latin"] });
const _geistMono = Geist_Mono({ subsets: ["latin"] });

export async function generateMetadata(): Promise<Metadata> {
  try {
    const cfg = await fetchPublicSystemConfig()
    const title = (cfg.site_title || 'GameStore').trim() || 'GameStore'
    const description = (cfg.seo_description || '').trim() || '发现并下载您喜爱的游戏资源'
    const favicon = (cfg.favicon_url || '').trim()
    const origin = await getSiteOrigin()
    return {
      title,
      description,
      keywords: (cfg.seo_keywords || '').trim() || undefined,
      icons: favicon ? { icon: [{ url: favicon }] } : undefined,
      alternates: { canonical: origin },
      openGraph: {
        title,
        description,
        url: origin,
        siteName: title,
        type: "website",
      },
    }
  } catch {
    return {
      title: 'GameStore',
      description: '发现并下载您喜爱的游戏资源',
    }
  }
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="zh-CN">
      <body className="font-sans antialiased">
        {/* Google / SEO: JSON-LD (WebSite) */}
        {/* JSON-LD 不依赖 Nuxt SEO 模块，Next.js 也可直接输出 */}
        <JsonLd
          data={{
            "@context": "https://schema.org",
            "@type": "WebSite",
            name: "GameSite",
            url: process.env.NEXT_PUBLIC_SITE_ORIGIN || undefined,
            potentialAction: {
              "@type": "SearchAction",
              target: `${process.env.NEXT_PUBLIC_SITE_ORIGIN || ""}/?q={search_term_string}`,
              "query-input": "required name=search_term_string",
            },
          }}
        />
        {children}
        {process.env.NODE_ENV === 'production' && <Analytics />}
      </body>
    </html>
  )
}
