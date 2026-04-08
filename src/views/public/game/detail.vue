<template>
  <div ref="mountEl" class="react-game-detail-root"></div>
</template>

<script setup lang="ts">
import { createElement } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import GameDetailReact from './GameDetailReact'

defineOptions({
  name: 'PublicGameDetailView',
})

const route = useRoute()
const mountEl = ref<HTMLElement | null>(null)
let root: Root | null = null

const renderApp = () => {
  if (!mountEl.value) return
  if (!root) root = createRoot(mountEl.value)
  root.render(createElement(GameDetailReact, { gameId: String(route.params.id || '') }))
}

onMounted(() => {
  renderApp()
})

watch(
  () => route.params.id,
  () => {
    renderApp()
  },
)

onBeforeUnmount(() => {
  root?.unmount()
  root = null
})
</script>

<style scoped>
.react-game-detail-root {
  width: 100%;
}
</style>
