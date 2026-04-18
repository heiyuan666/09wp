<script setup lang="ts">
import type { Software } from '../types/software'
import { statusLabels, platformIcons } from '../types/software'

const props = defineProps<{
  software: Software
  size?: 'sm' | 'md' | 'lg'
}>()

const cardSize = computed(() => props.size || 'sm')
</script>

<template>
  <NuxtLink
    :to="`/software/${software.id}`"
    class="group block"
  >
    <!-- Large size (Featured) -->
    <UCard
      v-if="cardSize === 'lg'"
      class="h-full transition-all group-hover:ring-2 group-hover:ring-primary group-hover:shadow-lg"
    >
      <template #header>
        <div class="flex items-center gap-4">
          <img
            :src="software.icon"
            :alt="software.name"
            class="size-14 rounded-lg object-cover bg-muted"
          />
          <div class="flex-1 min-w-0">
            <h3 class="font-semibold text-lg truncate">{{ software.name }}</h3>
            <div class="flex items-center gap-2 mt-1">
              <UBadge :color="statusLabels[software.status].color" variant="subtle" size="sm">
                {{ statusLabels[software.status].label }}
              </UBadge>
              <span class="text-muted text-sm">v{{ software.version }}</span>
            </div>
          </div>
        </div>
      </template>
      <p class="text-muted text-sm line-clamp-2">{{ software.description }}</p>
      <template #footer>
        <div class="flex items-center gap-2">
          <UIcon
            v-for="platform in software.platform"
            :key="platform"
            :name="platformIcons[platform] || 'i-lucide-box'"
            class="size-4 text-muted"
          />
        </div>
      </template>
    </UCard>

    <!-- Medium size (Category page) -->
    <UCard
      v-else-if="cardSize === 'md'"
      class="h-full transition-all group-hover:ring-2 group-hover:ring-primary group-hover:shadow-lg"
    >
      <div class="flex items-start gap-4">
        <img
          :src="software.icon"
          :alt="software.name"
          class="size-16 rounded-xl object-cover bg-muted shrink-0"
        />
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2">
            <h3 class="font-semibold text-lg truncate">{{ software.name }}</h3>
          </div>
          <div class="flex items-center gap-2 mt-1">
            <UBadge
              :color="statusLabels[software.status].color"
              variant="subtle"
              size="sm"
            >
              {{ statusLabels[software.status].label }}
            </UBadge>
            <span class="text-muted text-sm">v{{ software.version }}</span>
          </div>
        </div>
      </div>

      <p class="text-muted text-sm mt-4 line-clamp-2">{{ software.description }}</p>

      <div class="flex items-center justify-between mt-4 pt-4 border-t border-default">
        <div class="flex items-center gap-2">
          <UIcon
            v-for="platform in software.platform"
            :key="platform"
            :name="platformIcons[platform] || 'i-lucide-box'"
            class="size-4 text-muted"
          />
        </div>
        <span class="text-muted text-sm">{{ software.size }}</span>
      </div>
    </UCard>

    <!-- Small size (Default, list page) -->
    <UCard
      v-else
      class="h-full transition-all group-hover:ring-2 group-hover:ring-primary group-hover:shadow-lg"
    >
      <div class="flex items-start gap-3">
        <img
          :src="software.icon"
          :alt="software.name"
          class="size-12 rounded-lg object-cover bg-muted shrink-0"
        />
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2">
            <h3 class="font-semibold truncate">{{ software.name }}</h3>
            <UBadge
              :color="statusLabels[software.status].color"
              variant="subtle"
              size="xs"
            >
              {{ statusLabels[software.status].label }}
            </UBadge>
          </div>
          <p class="text-muted text-xs mt-1">{{ software.categoryLabel }}</p>
        </div>
      </div>
      <p class="text-muted text-sm mt-3 line-clamp-2">{{ software.description }}</p>
      <div class="flex items-center justify-between mt-4 pt-3 border-t border-default">
        <div class="flex items-center gap-1.5">
          <UIcon
            v-for="platform in software.platform.slice(0, 3)"
            :key="platform"
            :name="platformIcons[platform] || 'i-lucide-box'"
            class="size-4 text-muted"
          />
          <span v-if="software.platform.length > 3" class="text-muted text-xs">
            +{{ software.platform.length - 3 }}
          </span>
        </div>
        <span class="text-muted text-xs">{{ software.size }}</span>
      </div>
    </UCard>
  </NuxtLink>
</template>
