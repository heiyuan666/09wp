<template>
  <div ref="mountEl" class="react-home-root"></div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { createElement } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import HomeReact from './HomeReact'

const mountEl = ref<HTMLElement | null>(null)
let root: Root | null = null

onMounted(() => {
  if (!mountEl.value) return
  root = createRoot(mountEl.value)
  root.render(createElement(HomeReact))
})

onBeforeUnmount(() => {
  root?.unmount()
  root = null
})
</script>

<style scoped>
.react-home-root {
  width: 100%;
}
</style>
