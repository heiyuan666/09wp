<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'
import { iconForCategorySlug } from './types/software'
import type { ApiSoftwareCategory } from './types/software'

const { fetchCategories } = useSoftwareApi()
const { resolvedLogo, resolvedFavicon, siteTitle } = await useSoftwareSiteConfig()

const { data: categories } = await useAsyncData('sw-categories', () => fetchCategories(), {
  default: () => ({ list: [] as ApiSoftwareCategory[] }),
})

const categoryNav = computed(() =>
  (categories.value?.list ?? []).map((c: ApiSoftwareCategory) => ({
    label: c.name,
    to: `/category/${c.slug}`,
    icon: iconForCategorySlug(c.slug),
  })),
)

const navItems = computed<NavigationMenuItem[]>(() => [
  {
    label: '首页',
    to: '/',
    icon: 'i-lucide-home',
  },
  {
    label: '分类',
    icon: 'i-lucide-grid-3x3',
    children:
      categoryNav.value.length > 0
        ? categoryNav.value.map(c => ({ label: c.label, to: c.to }))
        : [
            { label: '效率工具', to: '/category/productivity' },
            { label: '开发工具', to: '/category/development' },
            { label: '设计软件', to: '/category/design' },
            { label: '系统工具', to: '/category/system' },
            { label: '媒体播放', to: '/category/media' },
            { label: '网络工具', to: '/category/network' },
          ],
  },
])

/** 使用 useHead 的 getter（非 computed 包一层），与 @unhead/vue 2 的解析兼容 */
useHead(() => {
  const fav = resolvedFavicon?.value ?? ''
  const name = siteTitle?.value ?? '软件库'
  return {
    ...(fav ? { link: [{ rel: 'icon', href: fav }] } : {}),
    meta: [
      { property: 'og:site_name', content: name },
      { name: 'twitter:card', content: 'summary_large_image' },
    ],
  }
})
</script>

<template>
  <UApp>
    <UHeader>
      <template #title>
        <NuxtLink to="/" class="flex items-center gap-2 min-w-0">
          <img
            v-if="resolvedLogo"
            :src="resolvedLogo"
            :alt="siteTitle"
            class="h-8 w-auto max-w-[160px] object-contain object-left"
          />
          <template v-else>
            <UIcon name="i-lucide-box" class="size-7 text-primary shrink-0" />
            <span class="font-bold text-lg truncate">{{ siteTitle }}</span>
          </template>
        </NuxtLink>
      </template>

      <UNavigationMenu :items="navItems" />

      <template #right>
        <UColorModeButton />
      </template>

      <template #body>
        <UNavigationMenu
          :items="navItems"
          orientation="vertical"
          class="-mx-2.5"
        />
      </template>
    </UHeader>

    <UMain>
      <NuxtPage />
    </UMain>

    <UFooter>
      <template #left>
        <p class="text-muted text-sm">
          © {{ new Date().getFullYear() }} {{ siteTitle }} · 保留所有权利
        </p>
      </template>
      <template #right>
        <UButton
          icon="i-simple-icons-github"
          color="neutral"
          variant="ghost"
          to="https://github.com"
          target="_blank"
        />
      </template>
    </UFooter>
  </UApp>
</template>
