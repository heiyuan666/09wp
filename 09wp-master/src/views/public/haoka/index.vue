<template>
  <div class="haoka-page">
    <section class="hero">
      <div class="hero-copy">
        <p class="eyebrow">HOT CARD COLLECTION</p>
        <h1 class="hero-title">号卡专区</h1>
        <p class="hero-desc">一站式筛选运营商、优惠周期与归属规则，快速找到适合你的套餐。</p>
      </div>

      <div class="hero-metrics">
        <article class="metric">
          <span class="metric-label">在售套餐</span>
          <strong class="metric-value">{{ total }}</strong>
        </article>
        <article class="metric">
          <span class="metric-label">分类数量</span>
          <strong class="metric-value">{{ categories.length }}</strong>
        </article>
        <article class="metric">
          <span class="metric-label">已启用筛选</span>
          <strong class="metric-value">{{ activeFilterCount }}</strong>
        </article>
      </div>
    </section>

    <section class="filter-panel">
      <div class="operator-tabs" role="tablist" aria-label="运营商筛选">
        <button class="op-chip" :class="{ active: operator === '' }" @click="setOperator('')">全部</button>
        <button v-for="item in operatorOptions" :key="item" class="op-chip" :class="{ active: operator === item }" @click="setOperator(item)">
          {{ item }}
        </button>
      </div>

      <div class="filter-grid">
        <label class="field-wrap">
          <span class="field-label">分类</span>
          <select v-model="categoryId" class="field">
            <option value="">全部分类</option>
            <option v-for="c in categories" :key="c.id" :value="String(c.id)">
              {{ c.name }}
            </option>
          </select>
        </label>

        <label class="field-wrap field-search">
          <span class="field-label">搜索</span>
          <input v-model="q" class="field" type="text" placeholder="输入套餐关键字" @keyup.enter="triggerSearch" />
        </label>

        <div class="actions">
          <button class="btn btn-primary" @click="triggerSearch">查询</button>
          <button class="btn btn-secondary" @click="resetFilters">重置</button>
        </div>
      </div>
    </section>

    <section v-if="list.length > 0" class="cards">
      <article v-for="row in list" :key="row.id" class="card" @click="goDetail(row.id)">
        <div class="card-cover">
          <img v-if="row.main_pic" class="card-img" :src="row.main_pic" :alt="row.product_name" />
          <div v-else class="card-img card-placeholder">暂无主图</div>
          <span class="card-tag">在售</span>
        </div>

        <div class="card-body">
          <h3 class="card-title" :title="row.product_name">{{ row.product_name }}</h3>
          <div class="pill-row">
            <span class="pill">{{ row.operator || '-' }}</span>
            <span class="pill pill-muted">{{ row.area || '全国可办' }}</span>
            <span v-if="row.price_time" class="pill pill-emphasis">{{ row.price_time }}</span>
          </div>

          <div class="meta-grid">
            <p class="meta-item">
              <span class="meta-key">返佣类型</span>
              <span class="meta-value">{{ row.back_money_type || '-' }}</span>
            </p>
            <p class="meta-item">
              <span class="meta-key">选号规则</span>
              <span class="meta-value">{{ numberSelText(row.number_sel) }}</span>
            </p>
          </div>

          <button class="detail-btn" @click.stop="goDetail(row.id)">查看详情</button>
        </div>
      </article>
    </section>

    <el-empty v-else description="暂无号卡套餐" :image-size="72" />

    <div class="pager" v-if="total > 0">
      <el-pagination
        background
        layout="total, sizes, prev, pager, next"
        :total="total"
        :page-sizes="[10, 20, 50]"
        v-model:current-page="page"
        v-model:page-size="pageSize"
        @change="load"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { haokaPublicCategories, haokaPublicProducts, type IHaokaCategory, type IHaokaProductItem } from '@/api/haoka_public'
import { runtimeConfig } from '@/config/runtimeConfig'

defineOptions({ name: 'PublicHaokaView' })

const router = useRouter()

const categories = ref<IHaokaCategory[]>([])
const list = ref<IHaokaProductItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const categoryId = ref<string>('')
const operator = ref<string>('')
const q = ref<string>('')

const operatorOptions = ['电信', '移动', '联通']

const activeFilterCount = computed(() => [categoryId.value, operator.value, q.value.trim()].filter(Boolean).length)

const goDetail = (id: number) => {
  router.push(`/haoka/${id}`)
}

const setOperator = (value: string) => {
  operator.value = value
}

const numberSelText = (n?: number) => {
  const v = Number(n ?? 0)
  if (v === 0) return '不支持选号'
  if (v === 1) return '收货地非归属地'
  if (v === 2) return '收货地为归属地'
  return '-'
}

const load = async () => {
  const { data: res } = await haokaPublicProducts({
    category_id: categoryId.value ? Number(categoryId.value) : undefined,
    operator: operator.value || undefined,
    q: q.value || undefined,
    page: page.value,
    page_size: pageSize.value,
    flag: 'true',
  })
  if (res.code !== 200) return
  list.value = res.data?.list || []
  total.value = Number(res.data?.total || 0)
}

const triggerSearch = () => {
  page.value = 1
  load()
}

const resetFilters = () => {
  categoryId.value = ''
  operator.value = ''
  q.value = ''
  triggerSearch()
}

onMounted(async () => {
  // 列表页基础 SEO（可被搜索引擎抓取）
  const siteTitle = runtimeConfig.siteTitle || '网盘资源导航系统'
  document.title = `号卡套餐 - ${siteTitle}`

  const desc = runtimeConfig.seoDescription || '号卡专区：聚合电信/移动/联通上架套餐，查看返佣与选号规则，快速找到适合你的号卡。'
  const setMeta = (name: string, content: string) => {
    if (!content) return
    let meta = document.querySelector(`meta[name='${name}']`) as HTMLMetaElement
    if (!meta) {
      meta = document.createElement('meta')
      meta.name = name
      document.head.appendChild(meta)
    }
    meta.content = content
  }
  const setOpenGraphMeta = (property: string, content: string) => {
    if (!content) return
    let meta = document.querySelector(`meta[property='${property}']`) as HTMLMetaElement
    if (!meta) {
      meta = document.createElement('meta')
      meta.setAttribute('property', property)
      document.head.appendChild(meta)
    }
    meta.content = content
  }

  setMeta('description', desc)
  const keywords = [runtimeConfig.seoKeywords, '号卡', 'SIM卡', '电信', '移动', '联通', '套餐'].filter(Boolean).join(',')
  setMeta('keywords', keywords.slice(0, 120))
  setOpenGraphMeta('og:title', document.title)
  setOpenGraphMeta('og:description', desc)
  setOpenGraphMeta('og:type', 'website')

  const { data: res } = await haokaPublicCategories()
  if (res.code === 200) categories.value = res.data || []
  await load()
})
</script>

<style scoped>
.haoka-page {
  --ink: #0f1723;
  --muted: rgba(15, 23, 35, 0.64);
  --line: rgba(15, 23, 35, 0.12);
  --brand: #1f7a5e;
  --brand-soft: rgba(31, 122, 94, 0.14);
  position: relative;
  padding: 22px 14px 20px;
  color: var(--ink);
  background:
    radial-gradient(circle at 92% -10%, rgba(31, 122, 94, 0.26), transparent 34%),
    radial-gradient(circle at -8% 12%, rgba(26, 76, 146, 0.18), transparent 32%),
    linear-gradient(140deg, #faf9f5 0%, #f2f2ee 100%);
}

.hero {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(0, 1fr);
  gap: 18px;
  margin-bottom: 16px;
}

.hero-copy {
  background: rgba(255, 255, 255, 0.58);
  border: 1px solid var(--line);
  border-radius: 22px;
  padding: 24px 22px 22px;
  backdrop-filter: blur(8px);
}

.eyebrow {
  margin: 0;
  font-size: 11px;
  letter-spacing: 0.18em;
  font-weight: 800;
  color: var(--muted);
}

.hero-title {
  margin: 8px 0 0;
  font-size: clamp(28px, 6vw, 46px);
  line-height: 1.05;
  font-weight: 900;
}

.hero-desc {
  margin: 14px 0 0;
  color: var(--muted);
  line-height: 1.75;
  font-size: 14px;
  max-width: 560px;
}

.hero-metrics {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
}

.metric {
  min-height: 104px;
  border-radius: 18px;
  border: 1px solid var(--line);
  background: rgba(255, 255, 255, 0.62);
  padding: 18px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  backdrop-filter: blur(8px);
}

.metric-label {
  font-size: 12px;
  color: var(--muted);
  font-weight: 700;
}

.metric-value {
  margin-top: 8px;
  line-height: 1;
  font-size: clamp(28px, 3vw, 38px);
  font-weight: 900;
}

.filter-panel {
  border-radius: 20px;
  border: 1px solid var(--line);
  background: rgba(255, 255, 255, 0.62);
  backdrop-filter: blur(10px);
  padding: 14px;
  margin-bottom: 16px;
}

.operator-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}

.op-chip {
  height: 44px;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.7);
  color: var(--ink);
  padding: 0 16px;
  font-weight: 800;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.op-chip:hover {
  transform: translateY(-1px);
}

.op-chip.active {
  border-color: rgba(31, 122, 94, 0.3);
  background: var(--brand-soft);
  color: #0e4e3c;
}

.filter-grid {
  display: grid;
  grid-template-columns: 180px minmax(0, 1fr) auto;
  gap: 10px;
  align-items: end;
}

.field-wrap {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-label {
  font-size: 12px;
  color: var(--muted);
  font-weight: 700;
}

.field {
  height: 44px;
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 0 12px;
  background: rgba(255, 255, 255, 0.84);
  outline: none;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.field:focus {
  border-color: rgba(31, 122, 94, 0.4);
  box-shadow: 0 0 0 3px rgba(31, 122, 94, 0.12);
}

.actions {
  display: flex;
  gap: 8px;
}

.btn {
  min-width: 88px;
  height: 44px;
  border-radius: 12px;
  font-weight: 800;
  border: 1px solid transparent;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn:hover {
  transform: translateY(-1px);
}

.btn-primary {
  background: linear-gradient(145deg, #1f7a5e, #16644d);
  color: #fff;
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.76);
  border-color: var(--line);
  color: var(--ink);
}

.cards {
  display: grid;
  gap: 14px;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
}

.card {
  border: 1px solid var(--line);
  border-radius: 18px;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.72);
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.card:hover {
  transform: translateY(-3px);
  box-shadow: 0 16px 34px rgba(15, 23, 35, 0.08);
}

.card-cover {
  position: relative;
  height: 168px;
  background: rgba(15, 23, 35, 0.04);
}

.card-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.card-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(15, 23, 35, 0.45);
  font-weight: 700;
}

.card-tag {
  position: absolute;
  top: 10px;
  right: 10px;
  padding: 5px 10px;
  border-radius: 999px;
  font-size: 11px;
  color: white;
  background: rgba(11, 18, 30, 0.62);
  border: 1px solid rgba(255, 255, 255, 0.32);
}

.card-body {
  padding: 14px;
}

.card-title {
  margin: 0;
  font-size: 18px;
  line-height: 1.3;
  min-height: 46px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.pill-row {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
}

.pill {
  border-radius: 999px;
  padding: 5px 10px;
  font-size: 12px;
  font-weight: 800;
  border: 1px solid rgba(31, 122, 94, 0.24);
  background: var(--brand-soft);
}

.pill-muted {
  border-color: var(--line);
  background: rgba(15, 23, 35, 0.04);
  color: var(--muted);
}

.pill-emphasis {
  border-color: rgba(26, 76, 146, 0.3);
  background: rgba(26, 76, 146, 0.12);
}

.meta-grid {
  margin-top: 12px;
  display: grid;
  grid-template-columns: 1fr;
  gap: 7px;
}

.meta-item {
  margin: 0;
  display: flex;
  justify-content: space-between;
  gap: 10px;
  font-size: 12px;
}

.meta-key {
  color: var(--muted);
}

.meta-value {
  color: var(--ink);
  font-weight: 700;
  text-align: right;
}

.detail-btn {
  margin-top: 14px;
  width: 100%;
  height: 44px;
  border: 1px solid var(--line);
  border-radius: 11px;
  background: rgba(255, 255, 255, 0.85);
  color: var(--ink);
  font-weight: 800;
  cursor: pointer;
}

.detail-btn:hover {
  border-color: rgba(31, 122, 94, 0.3);
  background: rgba(31, 122, 94, 0.07);
}

.pager {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 1024px) {
  .hero {
    grid-template-columns: 1fr;
  }

  .hero-metrics {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .filter-grid {
    grid-template-columns: 1fr 1fr;
  }

  .field-search,
  .actions {
    grid-column: 1 / -1;
  }
}

@media (max-width: 768px) {
  .haoka-page {
    padding: 16px 10px 18px;
  }

  .hero-copy {
    padding: 18px 16px;
  }

  .hero-desc {
    font-size: 13px;
  }

  .hero-metrics {
    grid-template-columns: 1fr;
  }

  .filter-grid {
    grid-template-columns: 1fr;
  }

  .actions {
    width: 100%;
  }

  .btn {
    flex: 1;
  }

  .cards {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .card-cover {
    height: 180px;
  }
}
</style>
