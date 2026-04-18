<script setup lang="ts">
import { mapApiSoftwareList } from '../../utils/mapSoftware'
import { iconForCategorySlug } from '../../types/software'
import type { ApiSoftwareCategory } from '../../types/software'

const route = useRoute()
const slug = computed(() => route.params.slug as string)

const { fetchCategories, fetchSoftwarePage, apiBase } = useSoftwareApi()
const { data: sitePublic } = useSoftwareSiteConfig()

const page = ref(1)
const pageSize = 12
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

const { data: categoryRow, error: catError } = await useAsyncData(
  () => `sw-cat-${slug.value}`,
  async () => {
    const list = categories.value?.list
    if (!list?.length) {
      throw createError({ statusCode: 500, statusMessage: '分类数据未加载' })
    }
    const c = list.find((x: ApiSoftwareCategory) => x.slug === slug.value)
    if (!c) {
      throw createError({ statusCode: 404, statusMessage: '分类不存在' })
    }
    return c
  },
  { watch: [slug] },
)

const categoryId = computed(() => categoryRow.value?.id)

const { data: listPage, pending, error: listError } = await useAsyncData(
  () => `sw-cat-list-${slug.value}`,
  () => {
    if (categoryId.value == null) {
      return Promise.resolve({ list: [], total: 0 })
    }
    return fetchSoftwarePage({
      page: page.value,
      page_size: pageSize,
      keyword: debouncedSearch.value || undefined,
      category_id: categoryId.value,
    })
  },
  { watch: [page, debouncedSearch, categoryId, slug] },
)

const softwareList = computed(() =>
  mapApiSoftwareList(listPage.value?.list ?? [], apiBase(), categoryList.value),
)

const totalItems = computed(() => Number(listPage.value?.total ?? 0))

const otherCategories = computed(() =>
  categoryList.value.filter((c: ApiSoftwareCategory) => c.slug !== slug.value),
)

const catPageTitle = computed(() => {
  const brand = sitePublic.value?.site_title?.trim() || '软件库'
  return categoryRow.value ? `${categoryRow.value.name} - ${brand}` : `分类 - ${brand}`
})
const catPageDesc = computed(() =>
  categoryRow.value
    ? `浏览 ${categoryRow.value.name} 分类下的所有软件。${sitePublic.value?.seo_description || ''}`.trim()
    : sitePublic.value?.seo_description || '浏览软件分类',
)
const catPageKeywords = computed(() => sitePublic.value?.seo_keywords || '')

useSeoMeta({
  title: catPageTitle,
  description: catPageDesc,
  keywords: catPageKeywords,
})
</script>

<template>
  <UContainer>
    <UAlert
      v-if="catError || listError"
      color="error"
      variant="subtle"
      title="加载失败"
      :description="(catError || listError)?.message || '请检查接口配置'"
      class="my-4"
    />

    <UPageHeader
      v-if="categoryRow"
      :title="categoryRow.name"
      :description="`浏览 ${categoryRow.name} 下的所有软件`"
      class="py-8"
    >
      <template #headline>
        <UBreadcrumb
          :items="[
            { label: '首页', to: '/', icon: 'i-lucide-home' },
            { label: '分类' },
            { label: categoryRow.name },
          ]"
        />
      </template>
    </UPageHeader>

    <div v-if="categoryRow" class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-6">
      <UInput
        v-model="searchInput"
        icon="i-lucide-search"
        :placeholder="`在 ${categoryRow.name} 中搜索...`"
        size="lg"
        class="w-full sm:max-w-md"
      />
      <UBadge color="neutral" variant="subtle">
        {{ totalItems }} 款软件
      </UBadge>
    </div>

    <div v-if="pending && !softwareList.length" class="text-muted py-12 text-center">
      加载中…
    </div>

    <UEmpty
      v-else-if="categoryRow && !softwareList.length"
      icon="i-lucide-inbox"
      title="暂无软件"
      description="该分类下暂时没有软件"
      class="py-12"
    />

    <template v-else-if="categoryRow && softwareList.length">
      <div class="grid sm:grid-cols-2 lg:grid-cols-3 gap-6 pb-8">
        <SoftwareCard
          v-for="item in softwareList"
          :key="item.id"
          :software="item"
          size="md"
        />
      </div>

      <div v-if="totalItems > pageSize" class="flex justify-center pb-12">
        <UPagination
          v-model="page"
          :total="totalItems"
          :items-per-page="pageSize"
          show-edges
        />
      </div>
    </template>

    <UPageSection
      headline="其他分类"
      title="浏览更多"
      class="border-t border-default"
    >
      <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4">
        <NuxtLink
          v-for="cat in otherCategories"
          :key="cat.id"
          :to="`/category/${cat.slug}`"
        >
          <UCard class="hover:ring-2 hover:ring-primary transition-all">
            <div class="flex flex-col items-center gap-3 py-2">
              <UIcon :name="iconForCategorySlug(cat.slug)" class="size-8 text-primary" />
              <span class="font-medium text-sm">{{ cat.name }}</span>
            </div>
          </UCard>
        </NuxtLink>
      </div>
    </UPageSection>
  </UContainer>
</template>
