<template>
  <div ref="mountEl" class="search-react-root"></div>
</template>

<script setup lang="ts">
import { createElement } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import SearchReact from '@/views/public/search/SearchReact'

defineOptions({
  name: 'PublicTagPage',
})

const mountEl = ref<HTMLElement | null>(null)
const route = useRoute()
const router = useRouter()
let root: Root | null = null

function routeTag(): string {
  const tag = route.params.tag
  if (Array.isArray(tag)) return String(tag[0] || '').trim()
  return String(tag || '').trim()
}

function renderSearch() {
  if (!mountEl.value) return
  if (!root) root = createRoot(mountEl.value)
  root.render(
    createElement(SearchReact, {
      routeQueryQ: routeTag(),
      onReplaceSearch: (q: string) => {
        const keyword = String(q || '').trim()
        if (!keyword) {
          router.replace('/search')
          return
        }
        router.replace(`/tag/${encodeURIComponent(keyword)}`)
      },
      onGoDetail: (id: string | number) => router.push(`/r/${id}`),
    }),
  )
}

let stopWatch: (() => void) | undefined

onMounted(() => {
  renderSearch()
  stopWatch = watch(() => [route.params.tag, route.path], renderSearch)
})

onBeforeUnmount(() => {
  stopWatch?.()
  root?.unmount()
  root = null
})
</script>

<style scoped>
.search-react-root {
  width: 100%;
}
</style>
