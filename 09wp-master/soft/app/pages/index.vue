<script setup lang="ts">
import { mapApiSoftwareList } from '../utils/mapSoftware'
import { iconForCategorySlug } from '../types/software'
import type { ApiSoftwareCategory } from '../types/software'

const { fetchCategories, fetchSoftwarePage, apiBase } = useSoftwareApi()
const { data: sitePublic } = useSoftwareSiteConfig()

/** 与后台「站点标题」一致，不再拼接固定后缀 */
const pageBrand = computed(() => sitePublic.value?.site_title?.trim() || '软件库')
const pageTitle = computed(() => pageBrand.value)
const pageDesc = computed(
  () =>
    sitePublic.value?.seo_description?.trim() ||
    '涵盖多平台、多分类的软件资源库，为您精选各类实用工具、媒体处理、办公效率等软件',
)
const pageOgDesc = computed(
  () => sitePublic.value?.seo_description?.trim() || '涵盖多平台、多分类的软件资源库',
)
const pageKeywords = computed(() => sitePublic.value?.seo_keywords || '')

useSeoMeta({
  title: pageTitle,
  description: pageDesc,
  ogTitle: pageTitle,
  ogDescription: pageOgDesc,
  keywords: pageKeywords,
})

const page = ref(1)
const pageSize = 12
const selectedCategoryId = ref<number | null>(null)
const searchInput = ref('')
const debouncedSearch = ref('')

let searchTimer: ReturnType<typeof setTimeout> | null = null
watch(searchInput, v => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    debouncedSearch.value = v
    page.value = 1
  }, 400)
})

const { data: categories } = await useAsyncData('sw-categories', () => fetchCategories(), {
  default: () => ({ list: [] as ApiSoftwareCategory[] }),
})

const categoryList = computed(() => categories.value?.list ?? [])

const categoryCards = computed(() =>
  categoryList.value.map((c: ApiSoftwareCategory) => ({
    id: c.id,
    label: c.name,
    slug: c.slug,
    icon: iconForCategorySlug(c.slug),
  })),
)

const { data: listPage, pending, error: listError } = await useAsyncData(
  'sw-list',
  () =>
    fetchSoftwarePage({
      page: page.value,
      page_size: pageSize,
      keyword: debouncedSearch.value || undefined,
      category_id: selectedCategoryId.value ?? undefined,
    }),
  { watch: [page, selectedCategoryId, debouncedSearch] },
)

const { data: featuredPage } = await useAsyncData('sw-featured', () =>
  fetchSoftwarePage({ page: 1, page_size: 3 }),
)

const softwareList = computed(() =>
  mapApiSoftwareList(listPage.value?.list ?? [], apiBase(), categoryList.value),
)

const totalItems = computed(() => Number(listPage.value?.total ?? 0))

const featuredSoftware = computed(() =>
  mapApiSoftwareList(featuredPage.value?.list ?? [], apiBase(), categoryList.value),
)

const selectedLabel = computed(() => {
  if (selectedCategoryId.value == null) return '全部软件'
  const c = categoryList.value.find((x: ApiSoftwareCategory) => x.id === selectedCategoryId.value)
  return c?.name || '全部软件'
})

function selectCategory(id: number | null) {
  selectedCategoryId.value = selectedCategoryId.value === id ? null : id
  page.value = 1
}

watch(selectedCategoryId, () => {
  page.value = 1
})
</script>

<template>
  <div>
    <UPageHero
      title="发现优质软件"
      description="涵盖多平台、多分类的软件资源库，为您精选各类实用工具"
      :links="[
        {
          label: '浏览全部软件',
          icon: 'i-lucide-arrow-down',
          to: '#software-list',
        },
      ]"
      class="bg-gradient-to-b from-primary/5 to-transparent"
    />

    <UPageSection
      headline="软件分类"
      title="按类别浏览"
      description="选择您需要的软件类型"
    >
      <div v-if="!categoryCards.length && pending" class="text-muted text-sm py-6">
        加载分类中…
      </div>
      <div v-else class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4">
        <UCard
          v-for="cat in categoryCards"
          :key="cat.id"
          class="cursor-pointer transition-all hover:ring-2 hover:ring-primary"
          :class="{ 'ring-2 ring-primary bg-primary/5': selectedCategoryId === cat.id }"
          @click="selectCategory(cat.id)"
        >
          <div class="flex flex-col items-center gap-3 py-2">
            <UIcon :name="cat.icon" class="size-8 text-primary" />
            <span class="font-medium text-sm">{{ cat.label }}</span>
          </div>
        </UCard>
      </div>
    </UPageSection>

    <UPageSection
      v-if="selectedCategoryId == null && !debouncedSearch"
      headline="精选推荐"
      title="热门软件"
      class="bg-muted/30"
    >
      <div v-if="!featuredSoftware.length && pending" class="text-muted text-sm py-6">
        加载中…
      </div>
      <div v-else class="grid md:grid-cols-3 gap-6">
        <SoftwareCard
          v-for="software in featuredSoftware"
          :key="software.id"
          :software="software"
          size="lg"
        />
      </div>
    </UPageSection>

    <UPageSection
      id="software-list"
      :headline="selectedLabel"
      :title="selectedCategoryId ? `${selectedLabel}列表` : '软件列表'"
    >
      <template #headline>
        <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between w-full mb-4 gap-4">
          <div class="flex items-center gap-3">
            <span class="text-sm text-primary font-medium uppercase tracking-wider">
              {{ selectedLabel }}
            </span>
            <UBadge color="neutral" variant="subtle" size="sm">
              {{ totalItems }} 款软件
            </UBadge>
          </div>
          <UInput
            v-model="searchInput"
            icon="i-lucide-search"
            placeholder="搜索软件..."
            class="w-full sm:w-64"
          />
        </div>
      </template>

      <UAlert
        v-if="listError"
        color="error"
        variant="subtle"
        title="加载失败"
        :description="listError.message || '请检查 NUXT_PUBLIC_API_BASE 与后端服务是否可用'"
        class="mb-4"
      />

      <div v-if="pending && !softwareList.length" class="text-muted py-12 text-center">
        加载中…
      </div>

      <UEmpty
        v-else-if="!pending && !softwareList.length"
        icon="i-lucide-inbox"
        title="暂无软件"
        description="没有找到符合条件的软件"
      />

      <template v-else>
        <div class="grid sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          <SoftwareCard
            v-for="software in softwareList"
            :key="software.id"
            :software="software"
            size="sm"
          />
        </div>

        <div v-if="totalItems > pageSize" class="flex justify-center mt-8">
          <UPagination
            v-model="page"
            :total="totalItems"
            :items-per-page="pageSize"
            show-edges
          />
        </div>
      </template>
    </UPageSection>
  </div>
</template>
