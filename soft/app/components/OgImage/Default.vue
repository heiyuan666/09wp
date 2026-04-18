<script setup lang="ts">
const { title = '软件库', description = '' } = defineProps<{
  title?: string
  description?: string
}>()
</script>

<template>
  <div
    :style="{
      display: 'flex',
      flexDirection: 'column',
      justifyContent: 'center',
      alignItems: 'center',
      width: '100%',
      height: '100%',
      padding: '60px',
      backgroundColor: '#0c0a09',
      color: '#fafaf9',
    }"
  >
    <h1
      :style="{
        fontSize: '64px',
        fontWeight: 'bold',
        margin: 0,
        lineHeight: 1.1,
        textAlign: 'center',
      }"
    >
      {{ title }}
    </h1>
    <p
      v-if="description"
      :style="{
        fontSize: '28px',
        opacity: 0.75,
        marginTop: '20px',
        textAlign: 'center',
        maxWidth: '900px',
      }"
    >
      {{ description }}
    </p>
  </div>
</template>
