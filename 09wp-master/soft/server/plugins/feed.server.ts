import type { Feed } from 'feed'

/**
 * nuxt-module-feed：填充 RSS / Atom 条目（可按需接后端软件列表）
 */
export default defineNitroPlugin((nitroApp) => {
  nitroApp.hooks.hook('feed:generate', async ({ feed }: { feed: Feed }) => {
    const siteUrl = (process.env.NUXT_PUBLIC_SITE_URL || 'http://localhost:3000').replace(/\/$/, '')
    const title = '软件库'
    const description = '发现、下载优质软件，涵盖多平台多分类的软件资源库'

    feed.options = {
      id: siteUrl,
      title,
      description,
      link: siteUrl,
      language: 'zh-CN',
      favicon: `${siteUrl}/icon.svg`,
      copyright: `© ${new Date().getFullYear()} ${title}`,
    }

    feed.addItem({
      title: '首页',
      id: `${siteUrl}/`,
      link: `${siteUrl}/`,
      description,
      date: new Date(),
    })
  })
})
