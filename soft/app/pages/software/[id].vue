<script setup lang="ts">
import { statusLabels, platformIcons } from '../../types/software'
import { netdiskLinkLabel } from '../../utils/netdiskLabel'

const { data: sitePublic } = useSoftwareSiteConfig()

const route = useRoute()
const id = computed(() => route.params.id as string)

const { fetchSoftwareDetailMapped } = useSoftwareApi()

const { data: software, error } = await useAsyncData(
  () => `software-detail-${id.value}`,
  () => fetchSoftwareDetailMapped(id.value),
  { watch: [id] },
)

const toast = useToast()

function copyPassword(pwd?: string) {
  const t = pwd || software.value?.cloudPassword
  if (t) {
    navigator.clipboard.writeText(t)
    toast.add({
      title: '复制成功',
      description: '提取码已复制到剪贴板',
      icon: 'i-lucide-check',
      color: 'success',
    })
  }
}

const screenshotModalOpen = ref(false)
const currentScreenshot = ref('')

function openScreenshot(url: string) {
  currentScreenshot.value = url
  screenshotModalOpen.value = true
}

const detailPageTitle = computed(() => {
  const brand = sitePublic.value?.site_title?.trim() || '软件库'
  return software.value ? `${software.value.name} - ${brand}` : `软件详情 - ${brand}`
})
const detailPageDesc = computed(
  () => software.value?.description?.trim() || sitePublic.value?.seo_description || '软件详情页',
)
const detailPageKeywords = computed(() => sitePublic.value?.seo_keywords || '')

useSeoMeta({
  title: detailPageTitle,
  description: detailPageDesc,
  keywords: detailPageKeywords,
})
</script>

<template>
  <UContainer class="py-8">
    <UEmpty
      v-if="error || !software"
      icon="i-lucide-file-question"
      title="软件未找到"
      description="您访问的软件不存在或已下架"
    >
      <template #actions>
        <UButton to="/" label="返回首页" icon="i-lucide-home" />
      </template>
    </UEmpty>

    <template v-else>
      <UBreadcrumb
        :items="[
          { label: '首页', to: '/', icon: 'i-lucide-home' },
          { label: software.categoryLabel, to: software.category ? `/category/${software.category}` : '/' },
          { label: software.name },
        ]"
        class="mb-6"
      />

      <div class="flex flex-col lg:flex-row gap-8 mb-8">
        <div class="flex items-start gap-6">
          <img
            :src="software.icon"
            :alt="software.name"
            class="size-24 lg:size-32 rounded-2xl object-cover bg-muted shadow-lg"
          />
          <div>
            <h1 class="text-2xl lg:text-3xl font-bold">{{ software.name }}</h1>
            <p class="text-muted mt-1">{{ software.categoryLabel }}</p>
            <div class="flex items-center gap-3 mt-3">
              <UBadge
                :color="statusLabels[software.status].color"
                variant="subtle"
              >
                {{ statusLabels[software.status].label }}
              </UBadge>
              <span class="text-muted">v{{ software.version }}</span>
            </div>
            <div class="flex items-center gap-2 mt-3">
              <UIcon
                v-for="platform in software.platform"
                :key="platform"
                :name="platformIcons[platform] || 'i-lucide-box'"
                class="size-5 text-muted"
              />
            </div>
          </div>
        </div>

        <div class="lg:ml-auto flex flex-col gap-3">
          <UButton
            v-if="software.directDownloads?.length"
            :to="software.directDownloads[0]"
            target="_blank"
            icon="i-lucide-download"
            label="直接下载"
            size="lg"
          />
          <UButton
            v-if="software.website"
            :to="software.website"
            target="_blank"
            icon="i-lucide-external-link"
            label="访问官网"
            color="neutral"
            variant="outline"
          />
        </div>
      </div>

      <div class="grid lg:grid-cols-3 gap-8">
        <div class="lg:col-span-2 space-y-8">
          <UCard>
            <template #header>
              <h2 class="font-semibold text-lg flex items-center gap-2">
                <UIcon name="i-lucide-info" class="size-5" />
                软件简介
              </h2>
            </template>
            <p class="text-muted leading-relaxed">{{ software.description }}</p>
          </UCard>

          <UCard v-if="software.cover">
            <template #header>
              <h2 class="font-semibold text-lg flex items-center gap-2">
                <UIcon name="i-lucide-image" class="size-5" />
                封面预览
              </h2>
            </template>
            <img
              :src="software.cover"
              :alt="`${software.name} 封面`"
              class="w-full rounded-lg object-cover max-h-96"
            />
          </UCard>

          <UCard v-if="software.screenshots?.length">
            <template #header>
              <h2 class="font-semibold text-lg flex items-center gap-2">
                <UIcon name="i-lucide-images" class="size-5" />
                软件截图
              </h2>
            </template>
            <div class="grid grid-cols-2 gap-4">
              <button
                v-for="(screenshot, index) in software.screenshots"
                :key="index"
                class="relative aspect-video rounded-lg overflow-hidden bg-muted group"
                @click="openScreenshot(screenshot)"
              >
                <img
                  :src="screenshot"
                  :alt="`截图 ${index + 1}`"
                  class="w-full h-full object-cover transition-transform group-hover:scale-105"
                />
                <div class="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-colors flex items-center justify-center">
                  <UIcon name="i-lucide-zoom-in" class="size-8 text-white opacity-0 group-hover:opacity-100 transition-opacity" />
                </div>
              </button>
            </div>
          </UCard>

          <UCard v-if="software.versionRows?.length">
            <template #header>
              <h2 class="font-semibold text-lg flex items-center gap-2">
                <UIcon name="i-lucide-layers" class="size-5" />
                历史版本
              </h2>
            </template>
            <div class="space-y-6">
              <div
                v-for="row in software.versionRows"
                :key="row.id"
                class="border border-default rounded-lg p-4 space-y-3"
              >
                <div class="flex flex-wrap items-center gap-2">
                  <span class="font-medium">v{{ row.version }}</span>
                  <UBadge v-if="row.publishedAt" color="neutral" variant="subtle" size="sm">
                    {{ row.publishedAt }}
                  </UBadge>
                </div>
                <p v-if="row.releaseNotes" class="text-sm text-muted whitespace-pre-wrap">{{ row.releaseNotes }}</p>
                <div v-if="row.directDownloads.length" class="space-y-2">
                  <span class="text-sm font-medium">直链</span>
                  <div class="flex flex-wrap gap-2">
                    <UButton
                      v-for="(link, i) in row.directDownloads"
                      :key="i"
                      :to="link"
                      target="_blank"
                      size="sm"
                      color="neutral"
                      variant="subtle"
                      icon="i-lucide-external-link"
                    >
                      直链 {{ i + 1 }}
                    </UButton>
                  </div>
                </div>
                <div v-if="row.cloudDownloads.length" class="space-y-2">
                  <span class="text-sm font-medium">网盘下载</span>
                  <div class="space-y-2">
                    <div
                      v-for="(link, i) in row.cloudDownloads"
                      :key="i"
                      class="flex flex-wrap items-center gap-2 rounded-lg border border-default p-2"
                    >
                      <UBadge color="primary" variant="subtle">{{ netdiskLinkLabel(link) }}</UBadge>
                      <UButton
                        :to="link"
                        target="_blank"
                        size="sm"
                        color="neutral"
                        variant="subtle"
                        icon="i-lucide-external-link"
                      >
                        打开分享
                      </UButton>
                    </div>
                  </div>
                  <div v-if="row.cloudPassword" class="flex items-center gap-2 p-2 bg-muted rounded-lg text-sm">
                    <span class="text-muted">提取码：</span>
                    <code class="font-mono">{{ row.cloudPassword }}</code>
                    <UButton
                      icon="i-lucide-copy"
                      size="xs"
                      color="neutral"
                      variant="ghost"
                      @click="copyPassword(row.cloudPassword)"
                    />
                  </div>
                </div>
              </div>
            </div>
          </UCard>

          <UCard>
            <template #header>
              <h2 class="font-semibold text-lg flex items-center gap-2">
                <UIcon name="i-lucide-download" class="size-5" />
                下载地址（当前版本）
              </h2>
            </template>

            <div class="space-y-6">
              <div v-if="software.directDownloads?.length">
                <h3 class="font-medium mb-3 flex items-center gap-2">
                  <UIcon name="i-lucide-link" class="size-4 text-primary" />
                  直链下载
                </h3>
                <div class="space-y-2">
                  <UButton
                    v-for="(link, index) in software.directDownloads"
                    :key="index"
                    :to="link"
                    target="_blank"
                    color="neutral"
                    variant="subtle"
                    class="w-full justify-start"
                    icon="i-lucide-external-link"
                  >
                    下载链接 {{ index + 1 }}
                  </UButton>
                </div>
              </div>

              <div v-if="software.cloudDownloads?.length">
                <h3 class="font-medium mb-3 flex items-center gap-2">
                  <UIcon name="i-lucide-cloud" class="size-4 text-primary" />
                  网盘下载
                </h3>
                <div class="space-y-2">
                  <div
                    v-for="(link, index) in software.cloudDownloads"
                    :key="index"
                    class="flex flex-wrap items-center gap-2 rounded-lg border border-default p-3"
                  >
                    <UBadge color="primary" variant="subtle">{{ netdiskLinkLabel(link) }}</UBadge>
                    <UButton
                      :to="link"
                      target="_blank"
                      color="neutral"
                      variant="subtle"
                      class="min-w-0 flex-1 justify-start"
                      icon="i-lucide-external-link"
                    >
                      打开分享
                    </UButton>
                  </div>
                </div>
                <div v-if="software.cloudPassword" class="mt-3 flex items-center gap-2 p-3 bg-muted rounded-lg">
                  <UIcon name="i-lucide-key" class="size-4 text-muted" />
                  <span class="text-sm text-muted">提取码：</span>
                  <code class="px-2 py-0.5 bg-elevated rounded text-sm font-mono">{{ software.cloudPassword }}</code>
                  <UButton
                    icon="i-lucide-copy"
                    size="xs"
                    color="neutral"
                    variant="ghost"
                    @click="copyPassword()"
                  />
                </div>
              </div>

              <UEmpty
                v-if="!software.directDownloads?.length && !software.cloudDownloads?.length"
                icon="i-lucide-download-off"
                title="暂无下载"
                description="该软件暂未提供下载链接"
              />
            </div>
          </UCard>
        </div>

        <div class="space-y-6">
          <UCard>
            <template #header>
              <h2 class="font-semibold text-lg flex items-center gap-2">
                <UIcon name="i-lucide-file-text" class="size-5" />
                软件信息
              </h2>
            </template>
            <dl class="space-y-4">
              <div class="flex justify-between">
                <dt class="text-muted">版本号</dt>
                <dd class="font-medium">{{ software.version }}</dd>
              </div>
              <USeparator />
              <div class="flex justify-between">
                <dt class="text-muted">文件大小</dt>
                <dd class="font-medium">{{ software.size }}</dd>
              </div>
              <USeparator />
              <div class="flex justify-between">
                <dt class="text-muted">状态</dt>
                <dd>
                  <UBadge
                    :color="statusLabels[software.status].color"
                    variant="subtle"
                    size="sm"
                  >
                    {{ statusLabels[software.status].label }}
                  </UBadge>
                </dd>
              </div>
              <USeparator />
              <div class="flex justify-between">
                <dt class="text-muted">分类</dt>
                <dd>
                  <NuxtLink
                    v-if="software.category"
                    :to="`/category/${software.category}`"
                    class="text-primary hover:underline"
                  >
                    {{ software.categoryLabel }}
                  </NuxtLink>
                  <span v-else>{{ software.categoryLabel }}</span>
                </dd>
              </div>
              <USeparator v-if="software.releaseDate" />
              <div v-if="software.releaseDate" class="flex justify-between">
                <dt class="text-muted">发布日期</dt>
                <dd class="font-medium">{{ software.releaseDate }}</dd>
              </div>
              <USeparator v-if="software.updateDate" />
              <div v-if="software.updateDate" class="flex justify-between">
                <dt class="text-muted">更新时间</dt>
                <dd class="font-medium">{{ software.updateDate }}</dd>
              </div>
            </dl>
          </UCard>

          <UCard>
            <template #header>
              <h2 class="font-semibold text-lg flex items-center gap-2">
                <UIcon name="i-lucide-monitor" class="size-5" />
                支持平台
              </h2>
            </template>
            <div class="flex flex-wrap gap-2">
              <UBadge
                v-for="platform in software.platform"
                :key="platform"
                color="neutral"
                variant="subtle"
                class="flex items-center gap-1.5"
              >
                <UIcon :name="platformIcons[platform] || 'i-lucide-box'" class="size-4" />
                {{ platform }}
              </UBadge>
              <span v-if="!software.platform.length" class="text-muted text-sm">未标注</span>
            </div>
          </UCard>

          <UCard v-if="software.website">
            <template #header>
              <h2 class="font-semibold text-lg flex items-center gap-2">
                <UIcon name="i-lucide-globe" class="size-5" />
                官方网站
              </h2>
            </template>
            <UButton
              :to="software.website"
              target="_blank"
              color="neutral"
              variant="subtle"
              class="w-full justify-start"
              icon="i-lucide-external-link"
            >
              {{ software.website }}
            </UButton>
          </UCard>
        </div>
      </div>
    </template>

    <UModal v-model:open="screenshotModalOpen" fullscreen>
      <template #body>
        <div class="flex items-center justify-center min-h-full p-4">
          <img
            :src="currentScreenshot"
            alt="Screenshot"
            class="max-w-full max-h-full rounded-lg"
          />
        </div>
      </template>
    </UModal>
  </UContainer>
</template>
