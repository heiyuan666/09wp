<template>
  <div class="haoka-detail">
    <el-skeleton :loading="loading" animated :rows="8">
      <div v-if="product" class="inner">
        <div class="bg" />

        <div class="topbar">
          <button class="back" @click="goBack">返回号卡</button>
          <div class="crumb">
            <span class="crumb-pill">{{ categoryName || '-' }}</span>
            <span class="crumb-sep">/</span>
            <span class="crumb-muted">套餐详情</span>
          </div>
        </div>

        <div class="hero">
          <div class="cover">
            <img v-if="product.main_pic" :src="product.main_pic" :alt="product.product_name" class="cover-img" />
            <div v-else class="cover-img placeholder">无主图</div>

            <div class="cover-badges">
              <span class="badge badge-dark">上架</span>
              <span class="badge badge-soft">返佣：{{ product.back_money_type || '-' }}</span>
            </div>
          </div>

          <div class="main">
            <h1 class="title">{{ product.product_name }}</h1>

            <div class="pill-row">
              <span class="pill">{{ product.operator || '-' }}</span>
              <span class="pill pill-soft">{{ product.area || '-' }}</span>
              <span v-if="product.price_time" class="pill pill-primary">优惠 {{ product.price_time }}</span>
            </div>

            <div class="grid2">
              <div class="info">
                <div class="info-k">选号规则</div>
                <div class="info-v">{{ numberSelText(product.number_sel) }}</div>
              </div>
              <div class="info">
                <div class="info-k">年龄范围</div>
                <div class="info-v">{{ product.age1 || 0 }} - {{ product.age2 || 0 }}</div>
              </div>
              <div class="info" v-if="product.disable_area">
                <div class="info-k">禁发区域</div>
                <div class="info-v">{{ product.disable_area }}</div>
              </div>
            </div>

            <div class="actions">
              <button class="btn btn-primary" :disabled="!product.net_addr" @click="copyText(product.net_addr || '')">
                复制资料地址
              </button>
              <button class="btn btn-ghost" :disabled="!product.net_addr" @click="openNetAddr">打开资料</button>
              <button
                class="btn btn-primary"
                :disabled="!runtimeConfig.haokaOrderUrl"
                @click="openOrder"
              >
                店铺下单
              </button>
              <button
                class="btn btn-ghost"
                :disabled="!runtimeConfig.haokaAgentRegUrl"
                @click="openAgentReg"
              >
                代理注册
              </button>
            </div>
          </div>
        </div>

        <div class="section">
          <div class="section-title">套餐说明</div>
          <div v-if="product.taocan" class="rich">{{ fullText(product.taocan) }}</div>
          <div v-else class="empty">暂无套餐说明</div>
        </div>

        <div class="section">
          <div class="section-title">结算规则</div>
          <div v-if="product.rule" class="rich">{{ fullText(product.rule) }}</div>
          <div v-else class="empty">暂无结算规则</div>
        </div>

        <div class="section">
          <div class="section-title">可办理区域（SKUs）</div>

          <div v-if="skus && skus.length > 0" class="sku-grid">
            <div v-for="s in skus" :key="s.sku_id" class="sku-card">
              <div class="sku-head">
                <div class="sku-name">{{ s.sku_name }}</div>
                <button v-if="s.desc" class="link" @click="toggleSku(s.sku_id)"> {{ expandedSkus[s.sku_id] ? '收起' : '展开' }} </button>
              </div>

              <div class="sku-desc" :class="{ expanded: expandedSkus[s.sku_id] }" v-if="s.desc">
                {{ fullText(s.desc) }}
              </div>
              <div v-else class="sku-desc empty">暂无描述</div>
            </div>
          </div>

          <el-empty v-else description="暂无SKU" :image-size="64" />
        </div>
      </div>
    </el-skeleton>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { haokaPublicProductDetail, type IHaokaProductDetail, type IHaokaSkuItem } from '@/api/haoka_public'
import { useClipboard } from '@vueuse/core'
import { runtimeConfig } from '@/config/runtimeConfig'

defineOptions({ name: 'PublicHaokaDetailView' })

const route = useRoute()
const router = useRouter()
const clipboard = useClipboard()

const loading = ref(true)
const product = ref<IHaokaProductDetail | null>(null)
const categoryName = ref<string>('')
const skus = ref<IHaokaSkuItem[]>([])
const expandedSkus = reactive<Record<number, boolean>>({})

const fullText = (s?: string) => String(s || '').trim()

const numberSelText = (n?: number) => {
  const v = Number(n ?? 0)
  if (v === 0) return '不支持'
  if (v === 1) return '收货地非归属地'
  if (v === 2) return '收货地为归属地'
  return '-'
}

const setMeta = (name: string, content: string) => {
  const c = String(content || '').trim()
  if (!c) return
  let meta = document.querySelector(`meta[name='${name}']`) as HTMLMetaElement
  if (!meta) {
    meta = document.createElement('meta')
    meta.name = name
    document.head.appendChild(meta)
  }
  meta.content = c
}

const setOpenGraphMeta = (property: string, content: string) => {
  const c = String(content || '').trim()
  if (!c) return
  let meta = document.querySelector(`meta[property='${property}']`) as HTMLMetaElement
  if (!meta) {
    meta = document.createElement('meta')
    meta.setAttribute('property', property)
    document.head.appendChild(meta)
  }
  meta.content = c
}

const applyDetailSeo = () => {
  const d = product.value
  if (!d) return

  const siteTitle = runtimeConfig.siteTitle || '网盘资源导航系统'
  const title = String(d.product_name || '').trim()
  document.title = title ? `${title} - ${siteTitle}` : `号卡详情 - ${siteTitle}`

  const rawDesc = String(d.taocan || '').trim() || String(d.rule || '').trim()
  const desc = rawDesc ? rawDesc.slice(0, 180) : runtimeConfig.seoDescription || ''

  setMeta('description', desc)

  const keywords = [title, d.operator, d.area, numberSelText(d.number_sel)]
    .filter(Boolean)
    .slice(0, 12)
    .join(',')
  setMeta('keywords', keywords)

  // OpenGraph：便于社交平台展示
  setOpenGraphMeta('og:title', document.title)
  setOpenGraphMeta('og:description', desc)
  setOpenGraphMeta('og:type', 'website')

  const coverRaw = String(d.main_pic || '').trim()
  if (coverRaw) {
    const coverAbs =
      coverRaw.startsWith('http://') || coverRaw.startsWith('https://')
        ? coverRaw
        : `${window.location.origin}${coverRaw.startsWith('/') ? '' : '/'}${coverRaw}`
    setOpenGraphMeta('og:image', coverAbs)
    setOpenGraphMeta('twitter:image', coverAbs)
  }
}

const load = async () => {
  loading.value = true
  try {
    const id = Number(route.params.id)
    if (!id) return
    const { data: res } = await haokaPublicProductDetail(id)
    if (res.code !== 200) return
    product.value = res.data?.product || null
    categoryName.value = res.data?.category_name || ''
    skus.value = res.data?.skus || []
    applyDetailSeo()
  } finally {
    loading.value = false
  }
}

const go = (url: string) => window.open(url, '_blank', 'noopener,noreferrer')

const goBack = () => {
  router.push('/haoka')
}

const openNetAddr = () => {
  if (!product.value?.net_addr) return
  go(product.value.net_addr)
}

const openOrder = () => {
  const url = String(runtimeConfig.haokaOrderUrl || '').trim()
  if (!url) return
  go(url)
}

const openAgentReg = () => {
  const url = String(runtimeConfig.haokaAgentRegUrl || '').trim()
  if (!url) return
  go(url)
}

const copyText = async (t: string) => {
  const text = String(t || '').trim()
  if (!text) return
  try {
    await clipboard.copy(text)
    ElMessage.success('已复制')
  } catch {
    ElMessage.error('复制失败')
  }
}

const toggleSku = (skuId: number) => {
  expandedSkus[skuId] = !expandedSkus[skuId]
}

onMounted(load)
</script>

<style scoped>
.haoka-detail {
  padding: 14px 12px 18px;
  position: relative;
}

.inner {
  position: relative;
}

.bg {
  position: absolute;
  inset: -80px -60px auto -60px;
  height: 280px;
  background: radial-gradient(540px 220px at 25% 35%, rgba(64, 158, 255, 0.22), transparent 62%),
    radial-gradient(480px 240px at 80% 20%, rgba(255, 170, 74, 0.16), transparent 65%),
    linear-gradient(180deg, rgba(0, 0, 0, 0.02), transparent);
  pointer-events: none;
  z-index: 0;
}

.topbar {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.back {
  height: 38px;
  padding: 0 14px;
  border-radius: 12px;
  border: 1px solid var(--el-border-color-lighter);
  background: rgba(255, 255, 255, 0.78);
  cursor: pointer;
  font-weight: 900;
  transition: transform 0.15s ease;
}
.back:hover {
  transform: translateY(-1px);
}

.crumb {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  opacity: 0.78;
}

.crumb-pill {
  font-weight: 900;
  padding: 6px 10px;
  border-radius: 999px;
  border: 1px solid rgba(64, 158, 255, 0.22);
  background: rgba(64, 158, 255, 0.08);
}
.crumb-sep {
  opacity: 0.4;
}
.crumb-muted {
  color: rgba(22, 22, 24, 0.72);
}

.hero {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 16px;
}

.cover {
  border-radius: 16px;
  overflow: hidden;
  border: 1px solid var(--el-border-color-lighter);
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(10px);
  position: relative;
}

.cover-img {
  width: 100%;
  height: 240px;
  object-fit: cover;
  display: block;
}
.cover-img.placeholder {
  height: 240px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 900;
  color: var(--el-text-color-secondary);
}

.cover-badges {
  position: absolute;
  inset: 12px auto auto 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  max-width: calc(100% - 24px);
}

.badge {
  font-size: 12px;
  font-weight: 950;
  padding: 7px 10px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.18);
}
.badge-dark {
  background: rgba(0, 0, 0, 0.52);
  color: white;
}
.badge-soft {
  background: rgba(255, 255, 255, 0.72);
  color: rgba(22, 22, 24, 0.88);
}

.main {
  border-radius: 16px;
  border: 1px solid var(--el-border-color-lighter);
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(10px);
  padding: 14px 14px 12px;
}

.title {
  font-size: 28px;
  line-height: 1.2;
  font-weight: 950;
  margin: 0;
}

.pill-row {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.pill {
  font-size: 12px;
  font-weight: 950;
  padding: 7px 10px;
  border-radius: 999px;
  border: 1px solid rgba(64, 158, 255, 0.18);
  background: rgba(64, 158, 255, 0.08);
}
.pill-soft {
  border-color: rgba(0, 0, 0, 0.08);
  background: rgba(0, 0, 0, 0.03);
  color: rgba(22, 22, 24, 0.74);
}
.pill-primary {
  border-color: rgba(64, 158, 255, 0.35);
  background: rgba(64, 158, 255, 0.16);
}

.grid2 {
  margin-top: 14px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.info {
  border-radius: 14px;
  border: 1px solid rgba(0, 0, 0, 0.06);
  background: rgba(255, 255, 255, 0.65);
  padding: 10px 12px;
}

.info-k {
  font-size: 12px;
  font-weight: 900;
  opacity: 0.7;
}
.info-v {
  margin-top: 6px;
  font-weight: 950;
  line-height: 1.4;
}

.actions {
  margin-top: 14px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.btn {
  height: 40px;
  padding: 0 14px;
  border-radius: 12px;
  border: 1px solid var(--el-border-color-lighter);
  cursor: pointer;
  font-weight: 950;
}

.btn-primary {
  border-color: rgba(64, 158, 255, 0.4);
  background: linear-gradient(180deg, rgba(64, 158, 255, 0.95), rgba(64, 158, 255, 0.7));
  color: white;
}
.btn-ghost {
  background: rgba(255, 255, 255, 0.76);
  color: rgba(22, 22, 24, 0.88);
}

.section {
  position: relative;
  z-index: 1;
  margin-top: 16px;
  border-radius: 16px;
  border: 1px solid var(--el-border-color-lighter);
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(10px);
  padding: 14px;
}

.section-title {
  font-weight: 950;
  font-size: 16px;
  margin-bottom: 10px;
}

.rich {
  font-size: 14px;
  line-height: 1.8;
  white-space: pre-wrap;
  word-break: break-word;
}

.empty {
  opacity: 0.65;
  font-size: 14px;
}

.sku-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 12px;
}

.sku-card {
  border-radius: 14px;
  border: 1px solid rgba(0, 0, 0, 0.06);
  background: rgba(255, 255, 255, 0.65);
  padding: 12px;
}

.sku-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.sku-name {
  font-weight: 950;
}

.link {
  height: 30px;
  border-radius: 10px;
  padding: 0 10px;
  border: 1px solid rgba(64, 158, 255, 0.22);
  background: rgba(64, 158, 255, 0.08);
  cursor: pointer;
  font-weight: 900;
  color: rgba(22, 22, 24, 0.9);
}

.sku-desc {
  margin-top: 10px;
  font-size: 13px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
  opacity: 0.92;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.sku-desc.expanded {
  -webkit-line-clamp: unset;
  display: block;
}

.sku-desc.empty {
  opacity: 0.55;
}

@media (max-width: 992px) {
  .hero {
    grid-template-columns: 1fr;
  }
  .grid2 {
    grid-template-columns: 1fr;
  }
  .cover-img,
  .cover-img.placeholder {
    height: 220px;
  }
}
</style>

