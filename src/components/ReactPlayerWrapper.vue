<template>
  <div ref="el" class="rp-wrap" />
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import React from 'react'
import ReactDOM from 'react-dom/client'
import ReactPlayer from 'react-player'

type Props = {
  url: string
  playing?: boolean
  controls?: boolean
  className?: string
}

const props = withDefaults(defineProps<Props>(), {
  playing: false,
  controls: true,
})

const el = ref<HTMLDivElement | null>(null)
let root: ReturnType<typeof ReactDOM.createRoot> | null = null

const render = () => {
  if (!el.value) return
  if (!props.url) return
  if (!root) root = ReactDOM.createRoot(el.value)
  root.render(
    React.createElement(ReactPlayer, {
      src: props.url,
      playing: !!props.playing,
      controls: props.controls,
      width: '100%',
      height: '100%',
      className: props.className || undefined,
      // 某些移动端需要显式 playsinline 才不会强制全屏
      playsInline: true,
    }),
  )
}

watch(
  () => [props.url, props.playing, props.controls, props.className],
  () => render(),
  { immediate: true },
)

onMounted(() => render())
onBeforeUnmount(() => {
  try {
    root?.unmount()
  } catch {
    // ignore
  }
})
</script>

<style scoped>
.rp-wrap {
  width: 100%;
  height: 100%;
}
</style>

