<template>
  <div ref="mountEl" class="react-game-detail-root" />
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { createElement } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import { useRoute } from 'vue-router'
import GameDetailStoreReact from './GameDetailStoreReact'

defineOptions({
  name: 'PublicGameDetailReactHost',
})

const mountEl = ref<HTMLElement | null>(null)
let root: Root | null = null

const route = useRoute()

const gameId = computed(() => {
  const raw = route.params.id as string | undefined
  const n = Number(raw)
  return Number.isFinite(n) && n > 0 ? n : 0
})

const render = () => {
  if (!mountEl.value) return
  if (!root) root = createRoot(mountEl.value)
  root.render(createElement(GameDetailStoreReact, { gameId: gameId.value }))
}

onMounted(() => {
  render()
})

watch(
  () => gameId.value,
  () => {
    render()
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

