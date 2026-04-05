<template>
  <div ref="mountEl" class="search-react-root"></div>
</template>

<script setup lang="ts">
import { createElement } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import SearchReact from './SearchReact'

defineOptions({
  name: 'PublicSearchPage',
})

const mountEl = ref<HTMLElement | null>(null)
const route = useRoute()
const router = useRouter()
let root: Root | null = null

function queryQ(): string {
  const q = route.query.q
  if (Array.isArray(q)) return String(q[0] || '')
  return String(q || '')
}

function renderSearch() {
  if (!mountEl.value) return
  if (!root) root = createRoot(mountEl.value)
  root.render(
    createElement(SearchReact, {
      routeQueryQ: queryQ(),
      onReplaceSearch: (q: string) => router.replace({ path: '/search', query: q ? { q } : {} }),
      onGoDetail: (id: string | number) => router.push(`/r/${id}`),
    }),
  )
}

let stopWatch: (() => void) | undefined

onMounted(() => {
  renderSearch()
  stopWatch = watch(() => [route.query.q, route.path], renderSearch)
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
