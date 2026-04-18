export default defineNuxtConfig({
  compatibilityDate: '2026-04-11',
  devtools: { enabled: true },
  modules: [
    '@nuxt/ui',
    // SEO：见 https://nuxt.com/modules?category=SEO
    '@nuxtjs/robots',
    '@nuxtjs/sitemap',
    'nuxt-schema-org',
    'nuxt-og-image',
    'nuxt-module-feed',
  ],
  css: ['~/assets/css/main.css'],
  runtimeConfig: {
    public: {
      /** 网盘后端 API 根（/api/v1 之前的主机部分） */
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://127.0.0.1:8080',
      /** 本站完整 URL（sitemap / OG / RSS 绝对链接，生产环境务必配置） */
      siteUrl: process.env.NUXT_PUBLIC_SITE_URL || 'http://localhost:3000',
    },
  },
  /** nuxt-site-config（@nuxtjs/sitemap / schema-org / og-image 依赖） */
  site: {
    url: process.env.NUXT_PUBLIC_SITE_URL || 'http://localhost:3000',
    name: '软件库',
    description: '发现、下载优质软件，涵盖多平台多分类的软件资源库',
    defaultLocale: 'zh',
  },
  robots: {
    allow: ['/'],
    sitemap: ['/sitemap.xml'],
  },
  sitemap: {
    /** 自动收录 Nuxt 路由；动态软件页可在后续用 server 源扩展 */
    autoLastmod: true,
  },
  schemaOrg: {
    identity: {
      type: 'Organization',
      name: '软件库',
      description: '发现、下载优质软件，涵盖多平台多分类的软件资源库',
    },
  },
  ogImage: {
    defaults: {
      width: 1200,
      height: 630,
    },
  },
  feed: {
    sources: [
      { path: '/rss.xml', type: 'rss2', cacheTime: 60 * 15 },
      { path: '/atom.xml', type: 'atom1', cacheTime: 60 * 15 },
    ],
  },
  // @nuxt/ui 会启用 @nuxt/fonts；默认会拉取 Google Fonts / Material Symbols，国内网络易超时。
  fonts: {
    providers: {
      google: false,
      googleicons: false,
    },
    priority: ['bunny', 'fontsource', 'local', 'npm'],
  },
  app: {
    head: {
      title: '软件库',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: '发现、下载优质软件，涵盖多平台多分类的软件资源库' },
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/icon.svg' },
        { rel: 'alternate', type: 'application/rss+xml', title: 'RSS', href: '/rss.xml' },
      ],
    },
  },
})
