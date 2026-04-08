<template>
  <div :class="['public-shell', { 'search-shell': isSearchRoute }]">
    <header v-if="!isHomeRoute && !isSearchRoute && !isGameRoute" class="public-header">
      <div class="inner">
        <div class="brand animate__animated animate__fadeInLeft animate__faster" @click="goHome">
          <img v-if="runtimeConfig.logoUrl" :src="runtimeConfig.logoUrl" alt="logo" class="brand-logo" />
          <span v-if="runtimeConfig.showSiteTitle !== false" class="brand-title">{{ runtimeConfig.siteTitle || '盘小子' }}</span>
        </div>
        <nav class="nav animate__animated animate__fadeInRight animate__faster">
          <el-button text class="nav-link" @click="goHome">首页</el-button>
          <el-button text class="nav-link" @click="goResource">资源</el-button>
  <el-popover v-if="runtimeConfig.doubanHotNavEnabled" placement="bottom-start" trigger="click" width="420">
            <template #reference>
              <el-button text class="nav-link">豆瓣热门</el-button>
            </template>
            <div v-if="doubanHot.length" class="douban-list">
              <div v-for="it in doubanHot" :key="it.title" class="douban-row" @click="applyDouban(it.title)">
                <img v-if="it.cover" :src="buildDoubanCoverSrc(it.cover)" class="douban-cover" alt="cover" />
                <div class="douban-title">{{ it.title }}</div>
              </div>
            </div>
            <div v-else class="douban-empty">暂无</div>
          </el-popover>
          <el-button text class="nav-link" @click="goContact">联系我们</el-button>
        </nav>
      </div>
    </header>

    <main :class="['public-main', { 'is-home': isHomeRoute, 'is-search': isSearchRoute }]">
      <router-view />
    </main>
    <footer v-if="!isHomeRoute && !isGameRoute" class="public-footer" id="contact">
      <div class="inner footer-grid">
        <div class="about">
          <div class="footer-brand clickable" @click="goHome">
            <img v-if="runtimeConfig.logoUrl" :src="runtimeConfig.logoUrl" alt="logo" class="brand-logo" />
            <span v-if="runtimeConfig.showSiteTitle !== false" class="brand-title">{{ runtimeConfig.siteTitle || '盘小子' }}</span>
          </div>
          <p>{{ runtimeConfig.seoDescription || `${runtimeConfig.siteTitle || '盘小子'}是一个聚合网盘资源的搜索平台。` }}</p>
        </div>
        <div>
          <h4>快捷链接</h4>
          <a v-for="(lnk, i) in footerQuickLinks" :key="`q-${i}`" :href="lnk.url || '#'">
            {{ lnk.title || lnk.url }}
          </a>
          <template v-if="runtimeConfig.friendLinks?.length">
            <a
              v-for="(fl, i) in runtimeConfig.friendLinks"
              :key="i"
              :href="fl.url || '#'"
              target="_blank"
              rel="noopener noreferrer"
            >
              {{ fl.title || fl.url }}
            </a>
          </template>
        </div>
        <div>
          <h4>热门网盘</h4>
          <span v-for="(it, i) in footerHotPlatforms" :key="`hp-${i}`">{{ it }}</span>
        </div>
        <div>
          <h4>联系我们</h4>
          <span>{{ runtimeConfig.footerWechat || '微信公众号' }}</span>
          <span>{{ runtimeConfig.supportEmail || 'support@example.com' }}</span>
          <h4 class="social-title">社交媒体</h4>
          <template v-for="(it, i) in footerSocialLinks" :key="`soc-${i}`">
            <a v-if="it.url" :href="it.url" target="_blank" rel="noopener noreferrer">{{ it.title || it.url }}</a>
            <span v-else>{{ it.title }}</span>
          </template>
        </div>
      </div>
      <div class="inner copyright">© 2026 {{ runtimeConfig.siteTitle || '盘小子' }}. All rights reserved.</div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { runtimeConfig } from '@/config/runtimeConfig'
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { siteDoubanHot, type IDoubanHotItem } from '@/api/netdisk'
import { buildProxiedImageSrc } from '@/utils/coverProxy'

const router = useRouter()
const route = useRoute()
const isHomeRoute = computed(() => route.path === '/')
const isSearchRoute = computed(() => route.path.startsWith('/search') || route.path.startsWith('/tag/'))
const isGameRoute = computed(() => route.path.startsWith('/games'))

const goHome = () => router.push('/')
const goResource = () => router.push('/search')
const goContact = () => document.getElementById('contact')?.scrollIntoView({ behavior: 'smooth' })

const doubanHot = ref<IDoubanHotItem[]>([])

const loadDoubanHot = async () => {
  if (!runtimeConfig.doubanHotNavEnabled) return
  try {
    const { data: res } = await siteDoubanHot({ limit: 8 })
    if (res.code === 200 && Array.isArray(res.data?.list)) {
      doubanHot.value = res.data.list
    } else {
      doubanHot.value = []
    }
  } catch {
    doubanHot.value = []
  }
}

const applyDouban = async (kw: string) => {
  doubanHot.value = doubanHot.value // keep
  await router.push({ path: '/search', query: kw ? { q: kw } : {} })
}

const buildDoubanCoverSrc = (cover?: string) =>
  buildProxiedImageSrc(cover, String(runtimeConfig.doubanCoverProxyUrl || '').trim())

const footerQuickLinks = computed(() => {
  if (Array.isArray(runtimeConfig.footerQuickLinks) && runtimeConfig.footerQuickLinks.length > 0) {
    return runtimeConfig.footerQuickLinks
  }
  return [
    { title: '首页', url: '/' },
    { title: '资源列表', url: '/search' },
    { title: '联系我们', url: '#contact' },
  ]
})

const footerHotPlatforms = computed(() => {
  if (Array.isArray(runtimeConfig.footerHotPlatforms) && runtimeConfig.footerHotPlatforms.length > 0) {
    return runtimeConfig.footerHotPlatforms
  }
  return ['夸克网盘', '阿里云盘', '百度网盘', '迅雷云盘']
})

const footerSocialLinks = computed(() => {
  if (Array.isArray(runtimeConfig.footerSocialLinks) && runtimeConfig.footerSocialLinks.length > 0) {
    return runtimeConfig.footerSocialLinks
  }
  return [{ title: 'Twitter', url: '' }]
})

watch(
  () => runtimeConfig.doubanHotNavEnabled,
  async () => {
    await loadDoubanHot()
  },
  { immediate: true },
)

const setCanonical = () => {
  let link = document.querySelector("link[rel='canonical']") as HTMLLinkElement
  if (!link) {
    link = document.createElement('link')
    link.rel = 'canonical'
    document.head.appendChild(link)
  }
  link.href = `${window.location.origin}${route.fullPath}`
}

const setMetaTag = (name: string, content: string) => {
  let meta = document.querySelector(`meta[name='${name}']`) as HTMLMetaElement
  if (!meta) {
    meta = document.createElement('meta')
    meta.name = name
    document.head.appendChild(meta)
  }
  meta.content = content
}

const setOgTag = (property: string, content: string) => {
  let meta = document.querySelector(`meta[property='${property}']`) as HTMLMetaElement
  if (!meta) {
    meta = document.createElement('meta')
    meta.setAttribute('property', property)
    document.head.appendChild(meta)
  }
  meta.content = content
}

watch(
  () => route.fullPath,
  () => {
    const suffix = runtimeConfig.siteTitle || '网盘资源站'
    if (route.path.startsWith('/search')) document.title = `搜索资源 - ${suffix}`
    else if (route.path.startsWith('/tag/')) {
      const tag = Array.isArray(route.params.tag)
        ? String(route.params.tag[0] || '').trim()
        : String(route.params.tag || '').trim()
      document.title = tag ? `${tag} 标签资源 - ${suffix}` : `标签资源 - ${suffix}`
    }
    else if (route.path.startsWith('/c/')) document.title = `分类资源 - ${suffix}`
    else if (route.path.startsWith('/r/')) document.title = `资源详情 - ${suffix}`
    else document.title = suffix
    setCanonical()
  },
  { immediate: true },
)

watch(
  () => [route.fullPath, runtimeConfig.siteTitle, runtimeConfig.seoKeywords, runtimeConfig.seoDescription],
  () => {
    if (!route.path.startsWith('/search') && !route.path.startsWith('/tag/')) return
    const suffix = runtimeConfig.siteTitle || '网盘资源站'
    const q = route.path.startsWith('/tag/')
      ? (Array.isArray(route.params.tag)
          ? String(route.params.tag[0] || '').trim()
          : String(route.params.tag || '').trim())
      : (Array.isArray(route.query.q) ? String(route.query.q[0] || '').trim() : String(route.query.q || '').trim())
    const title = q ? `${q} 搜索结果 - ${suffix}` : `网盘搜索 - ${suffix}`
    const keywords = [runtimeConfig.seoKeywords, q, '网盘搜索', '资源搜索'].filter(Boolean).join(',')
    const description = q
      ? `搜索“${q}”的网盘资源结果，来自${suffix}。`
      : runtimeConfig.seoDescription || `${suffix}网盘资源搜索聚合平台。`

    document.title = title
    setMetaTag('keywords', keywords)
    setMetaTag('description', description)
    setOgTag('og:title', title)
    setOgTag('og:description', description)
    setOgTag('og:type', 'website')
    setCanonical()
  },
  { immediate: true },
)
</script>

<style scoped>
.public-shell {
  min-height: 100vh;
  background: var(--el-bg-color-page);
  color: var(--el-text-color-primary);
}
.public-shell.search-shell {
  background: #ffffff;
}
.clickable {
  cursor: pointer;
}
.public-header {
  position: sticky;
  top: 0;
  z-index: 20;
  background: color-mix(in srgb, var(--el-bg-color) 96%, transparent);
  border-bottom: 1px solid var(--el-border-color-light);
  backdrop-filter: blur(8px);
}
.inner {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 16px;
  width: 100%;
}
.public-header .inner {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.brand {
  cursor: pointer;
  user-select: none;
  display: flex;
  align-items: center;
  gap: 8px;
}
.brand-logo {
  width: 26px;
  height: 26px;
  object-fit: contain;
  border-radius: 6px;
}
.brand-title {
  font-weight: 600;
  font-size: 16px;
  color: var(--el-text-color-regular);
  letter-spacing: 0.2px;
}
.nav {
  display: flex;
  gap: 6px;
  align-items: center;
}
.nav-link {
  height: 34px;
  color: var(--el-text-color-regular);
  padding: 0 10px;
  border-radius: 8px;
  transition: all 0.2s ease;
}
.nav-link:hover {
  background: color-mix(in srgb, var(--el-color-primary) 12%, var(--el-bg-color));
  color: var(--el-color-primary);
}

.douban-list {
  max-height: 360px;
  overflow: auto;
  padding: 4px 0;
}
.douban-row {
  display: flex;
  gap: 10px;
  align-items: center;
  padding: 8px 12px;
  cursor: pointer;
  border-radius: 10px;
}
.douban-row:hover {
  background: color-mix(in srgb, var(--el-color-primary) 10%, transparent);
}
.douban-cover {
  width: 34px;
  height: 48px;
  object-fit: cover;
  border-radius: 6px;
}
.douban-title {
  font-size: 13px;
  line-height: 1.3;
  color: var(--el-text-color-primary);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.douban-empty {
  padding: 10px 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.public-main {
  padding: 24px 0 0;
}
.public-main.is-home,
.public-main.is-search {
  padding-top: 0;
}
.public-footer {
  margin-top: 36px;
  border-top: 1px solid var(--el-border-color-light);
  background: var(--el-bg-color);
  padding: 26px 0 18px;
  color: var(--el-text-color-secondary);
}
.footer-grid {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr;
  gap: 20px;
}
.footer-brand {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}
.about p {
  line-height: 1.8;
  font-size: 13px;
}
.public-footer h4 {
  margin: 0 0 8px;
  font-size: 14px;
  color: var(--el-text-color-primary);
}
.public-footer a,
.public-footer span {
  display: block;
  color: var(--el-text-color-secondary);
  text-decoration: none;
  margin-bottom: 6px;
  font-size: 13px;
}
.social-title {
  margin-top: 14px !important;
}
.copyright {
  margin-top: 14px;
  border-top: 1px solid var(--el-border-color-lighter);
  padding-top: 12px;
  font-size: 12px;
}
@media (max-width: 860px) {
  .footer-grid {
    grid-template-columns: 1fr 1fr;
  }
}
@media (max-width: 640px) {
  .footer-grid {
    grid-template-columns: 1fr;
  }
}
</style>
