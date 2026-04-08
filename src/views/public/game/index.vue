<template>
  <div ref="mountEl" class="react-game-root"></div>
</template>

<script setup lang="ts">
import { createElement } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import { onBeforeUnmount, onMounted, ref } from 'vue'
import GameHomeStoreReact from './GameHomeStoreReact'

defineOptions({
  name: 'PublicGameHomeView',
})

const mountEl = ref<HTMLElement | null>(null)
let root: Root | null = null

onMounted(() => {
  if (!mountEl.value) return
  root = createRoot(mountEl.value)
  root.render(createElement(GameHomeStoreReact))
})

onBeforeUnmount(() => {
  root?.unmount()
  root = null
})
</script>

<style scoped>
.react-game-root {
  width: 100%;
}
</style>

