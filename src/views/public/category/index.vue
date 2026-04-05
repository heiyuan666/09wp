<template>
  <el-card class="panel" shadow="hover">
    <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
      <div>
        <div class="panel-title">分类资源：{{ slug }}</div>
        <div class="sub">共 {{ total }} 条资源</div>
      </div>
      <el-select v-model="platform" class="!w-[150px]">
        <el-option label="全部平台" value="" />
        <el-option label="百度网盘" value="baidu" />
        <el-option label="阿里云盘" value="aliyun" />
        <el-option label="夸克网盘" value="quark" />
        <el-option label="其他" value="other" />
      </el-select>
    </div>

    <div class="resource-grid">
      <div v-for="row in shownList" :key="row.id" class="item-card" @click="goDetail(row.id)">
        <img v-if="row.cover" class="cover" :src="row.cover" alt="cover" />
        <div v-else class="cover placeholder">网盘资源</div>
        <div class="item-title">{{ row.title }}</div>
        <div class="badges">
          <el-tag size="small">{{ platformText(row.link) }}</el-tag>
          <el-tag v-if="row.link_valid === false" type="danger" size="small">已失效</el-tag>
        </div>
        <div class="item-meta">
          <span>浏览 {{ row.view_count || 0 }}</span>
          <span v-if="row.extract_code">提取码 {{ row.extract_code }}</span>
        </div>
      </div>
    </div>
    <el-empty v-if="list.length === 0" description="该分类暂无资源" :image-size="80" />

    <div class="pager">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        :total="total"
        @change="load"
      />
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { siteCategories, siteResourcePage } from '@/api/netdisk'

const route = useRoute()
const router = useRouter()
const slug = computed(() => route.params.slug as string)

const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref<any[]>([])
const platform = ref('')
const categoryId = ref<string>('')

const goDetail = (id: any) => router.push(`/r/${id}`)
const platformText = (link: string) => {
  const u = String(link || '').toLowerCase()
  if (u.includes('pan.baidu.com')) return '百度网盘'
  if (u.includes('aliyundrive.com')) return '阿里云盘'
  if (u.includes('pan.quark.cn')) return '夸克网盘'
  return '其他'
}
const matchPlatform = (link: string) => {
  if (!platform.value) return true
  const p = platformText(link)
  if (platform.value === 'baidu') return p === '百度网盘'
  if (platform.value === 'aliyun') return p === '阿里云盘'
  if (platform.value === 'quark') return p === '夸克网盘'
  return p === '其他'
}
const shownList = computed(() => list.value.filter((x) => matchPlatform(x.link)))

const load = async () => {
  if (!categoryId.value) return
  const { data: res } = await siteResourcePage({
    category_id: categoryId.value,
    page: page.value,
    page_size: pageSize.value,
    sort: 'latest',
  })
  if (res.code !== 200) return
  list.value = res.data.list || []
  total.value = res.data.total || 0
}

onMounted(async () => {
  const { data: res } = await siteCategories()
  if (res.code !== 200) return
  const cat = (res.data || []).find((c: any) => c.slug === slug.value)
  categoryId.value = cat?.id ? String(cat.id) : ''
  await load()
})
</script>

<style scoped>
.panel {
  border-radius: 14px;
}
.panel-title {
  font-weight: 700;
}
.sub {
  margin-top: 4px;
  font-size: 12px;
  opacity: 0.8;
}
.resource-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 12px;
}
.item-card {
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s ease;
  background: #fff;
}
.item-card:hover {
  transform: translateY(-2px);
  border-color: rgba(80, 130, 255, 0.7);
}
.cover {
  width: 100%;
  height: 120px;
  object-fit: cover;
  display: block;
}
.placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f3f4f6;
  font-size: 12px;
  opacity: 0.8;
}
.item-title {
  padding: 8px 10px 0;
  font-weight: 700;
  line-height: 1.4;
  min-height: 44px;
}
.badges {
  padding: 6px 10px 0;
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}
.item-meta {
  padding: 6px 10px 10px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 12px;
  opacity: 0.85;
}
.pager {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}
</style>

