<template>
  <div class="game-detail">
    <div class="toolbar">
      <el-button text type="primary" @click="goList">← 返回游戏列表</el-button>
    </div>

    <el-skeleton v-if="loading" :rows="12" animated />

    <template v-else-if="game">
      <div
        class="hero"
        :style="heroBg ? { backgroundImage: `linear-gradient(180deg, rgba(0,0,0,.55), rgba(15,23,42,.92)), url(${heroBg})` } : undefined"
      >
        <div class="hero-inner">
          <img v-if="coverUrl" :src="coverUrl" class="cover" :alt="game.title" />
          <div class="title-block">
            <h1 class="title">{{ game.title }}</h1>
            <div class="chips">
              <el-tag v-if="game.developer" size="small">{{ game.developer }}</el-tag>
              <el-tag v-if="game.type" size="small" type="info">{{ game.type }}</el-tag>
              <el-tag v-if="game.release_date" size="small" type="success">发行 {{ game.release_date }}</el-tag>
              <span v-if="game.price_text" class="price">{{ game.price_text }}</span>
            </div>
            <p v-if="game.short_description" class="short">{{ game.short_description }}</p>
            <div class="meta-grid">
              <div v-if="game.publishers" class="meta-item">
                <span class="meta-label">发行商</span>
                <span class="meta-value">{{ game.publishers }}</span>
              </div>
              <div v-if="game.genres" class="meta-item">
                <span class="meta-label">类型</span>
                <span class="meta-value">{{ game.genres }}</span>
              </div>
              <div v-if="game.tags" class="meta-item">
                <span class="meta-label">标签</span>
                <span class="meta-value">{{ game.tags }}</span>
              </div>
              <div v-if="game.website" class="meta-item">
                <span class="meta-label">官网</span>
                <a class="meta-link" :href="game.website" target="_blank" rel="noopener noreferrer">{{ game.website }}</a>
              </div>
            </div>
          </div>
        </div>
      </div>

      <el-card v-if="videoUrl" class="section" shadow="hover">
        <template #header>视频播放</template>
        <div class="video-wrap">
          <button v-if="!videoStarted" type="button" class="video-poster" @click="videoStarted = true">
            <img v-if="videoPoster" :src="videoPoster" :alt="game.title" class="video-poster-image" />
            <div v-else class="video-poster-fallback"></div>
            <div class="video-poster-overlay">
              <span class="video-play-button">▶</span>
              <span class="video-play-text">点击播放视频</span>
            </div>
          </button>
          <div v-else class="video-player">
            <ReactPlayerWrapper :url="videoUrl" :playing="true" :controls="true" />
          </div>
        </div>
      </el-card>

      <el-card v-if="galleryUrls.length" class="section" shadow="hover">
        <template #header>截图 / 画廊</template>
        <div class="gallery">
          <el-carousel height="220px" :interval="0" arrow="always" indicator-position="outside">
            <el-carousel-item v-for="(u, i) in galleryUrls" :key="`${i}-${u}`">
              <div class="gallery-slide" @click="openGalleryViewer(u)">
                <img :src="u" class="g-img" alt="" loading="lazy" />
              </div>
            </el-carousel-item>
          </el-carousel>
        </div>
      </el-card>

      <el-dialog
        v-model="galleryViewerVisible"
        title="查看截图"
        width="860px"
        :show-close="true"
        destroy-on-close
      >
        <div class="gallery-viewer">
          <img v-if="galleryViewerUrl" :src="galleryViewerUrl" class="viewer-img" alt="" />
        </div>
      </el-dialog>

      <el-card v-if="descriptionHtml" class="section" shadow="hover">
        <template #header>游戏介绍</template>
        <div class="desc" v-html="descriptionHtml"></div>
      </el-card>

      <el-card v-if="resourceGroups.length" class="section" shadow="hover">
        <template #header>下载资源</template>
        <div class="resource-groups">
          <section v-for="group in resourceGroups" :key="group.key" class="resource-group">
            <div class="resource-group-head">
              <h3>{{ group.label }}</h3>
              <span>{{ group.items.length }} 个资源</span>
            </div>

            <div class="resource-list">
              <article v-for="item in group.items" :key="item.id" class="resource-card">
                <div class="resource-main">
                  <div class="resource-title-row">
                    <h4 class="resource-title">{{ item.title }}</h4>
                    <el-tag v-if="item.tested" type="success" size="small">已测试</el-tag>
                  </div>
                  <div class="resource-meta">
                    <span v-if="item.version" class="meta-pill">
                      <span class="meta-pill-icon">V</span>
                      <span>版本 {{ item.version }}</span>
                    </span>
                    <span v-if="item.size" class="meta-pill">
                      <span class="meta-pill-icon">S</span>
                      <span>{{ item.size }}</span>
                    </span>
                    <span v-if="item.pan_type" class="meta-pill">
                      <span class="meta-pill-icon">P</span>
                      <span>{{ item.pan_type }}</span>
                    </span>
                    <span v-if="item.download_type" class="meta-pill">
                      <span class="meta-pill-icon">D</span>
                      <span>{{ item.download_type }}</span>
                    </span>
                    <span v-if="item.author" class="meta-pill">
                      <span class="meta-pill-icon">A</span>
                      <span>作者 {{ item.author }}</span>
                    </span>
                    <span v-if="item.publish_date" class="meta-pill">
                      <span class="meta-pill-icon">T</span>
                      <span>{{ formatDate(item.publish_date) }}</span>
                    </span>
                  </div>
                </div>

                <div class="resource-links">
                  <div v-for="(u, i) in item.link_list" :key="`${item.id}-${i}`" class="download-action">
                    <el-link
                      :href="u"
                      target="_blank"
                      rel="noopener noreferrer"
                      type="primary"
                      class="download-link"
                    >
                      {{ getDownloadLabel(u, i, item.link_list.length) }}
                    </el-link>
                    <el-button size="small" text type="primary" @click="copyLink(u)">复制链接</el-button>
                  </div>
                </div>
              </article>
            </div>
          </section>
        </div>
      </el-card>

      <el-card class="section" shadow="hover">
        <template #header>用户提交资源</template>
        <template v-if="userToken">
          <el-form :model="submissionForm" label-position="top" class="submission-form">
            <el-form-item label="资源标题">
              <el-input v-model="submissionForm.title" placeholder="例如：Dota 2 夸克网盘整合包" />
            </el-form-item>
            <el-form-item label="下载链接">
              <el-input v-model="submissionForm.link" placeholder="粘贴网盘分享链接，支持夸克 / 百度 / 阿里 / 迅雷等" />
            </el-form-item>
            <div class="submission-grid">
              <el-form-item label="提取码">
                <el-input v-model="submissionForm.extract_code" placeholder="没有可不填" />
              </el-form-item>
              <el-form-item label="标签">
                <el-input v-model="submissionForm.tags" placeholder="如：补丁,整合包,教程" />
              </el-form-item>
            </div>
            <el-form-item label="资源说明">
              <el-input
                v-model="submissionForm.description"
                type="textarea"
                :rows="4"
                placeholder="可以补充版本、适用说明、安装方式等信息"
              />
            </el-form-item>
            <div class="submission-actions">
              <el-button type="primary" :loading="submittingSubmission" @click="submitGameResource">提交资源</el-button>
              <el-button @click="resetSubmissionForm">清空</el-button>
            </div>
            <div class="submission-hint">提交后会进入审核；审核通过后，会自动出现在当前游戏的资源列表里。</div>
          </el-form>

          <div class="submission-history">
            <div class="submission-history-title">我的投稿记录</div>
            <el-table v-if="myGameSubmissions.length > 0" :data="myGameSubmissions" size="small" style="width: 100%">
              <el-table-column prop="title" label="标题" min-width="220" />
              <el-table-column prop="status" label="状态" width="110">
                <template #default="{ row }">
                  <el-tag v-if="row.status === 'pending'" type="warning">待审核</el-tag>
                  <el-tag v-else-if="row.status === 'approved'" type="success">已通过</el-tag>
                  <el-tag v-else type="danger">已驳回</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="review_msg" label="备注" min-width="200" />
            </el-table>
            <el-empty v-else description="你还没有为这个游戏提交过资源" :image-size="72" />
          </div>
        </template>
        <div v-else class="submission-login">
          <el-empty description="登录后可为这个游戏提交资源" :image-size="72" />
          <el-button type="primary" @click="router.push('/login')">去登录</el-button>
        </div>
      </el-card>
    </template>

    <el-empty v-else description="未找到该游戏或已下架" />
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { siteMySubmissions, siteSubmissionCreate } from '@/api/netdisk'
import ReactPlayerWrapper from '@/components/ReactPlayerWrapper.vue'
import {
  publicGameDetail,
  publicGameResourceList,
  type PublicGameItem,
  type PublicGameResource,
} from './api'

defineOptions({
  name: 'PublicGameDetailView',
})

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const game = ref<PublicGameItem | null>(null)
const resources = ref<PublicGameResource[]>([])
const videoStarted = ref(false)
const submittingSubmission = ref(false)
const mySubmissions = ref<any[]>([])

const submissionForm = ref({
  title: '',
  link: '',
  extract_code: '',
  tags: '',
  description: '',
})

const gameId = computed(() => {
  const raw = route.params.id as string
  const n = Number(raw)
  return Number.isFinite(n) && n > 0 ? n : 0
})

const userToken = computed(() => {
  if (typeof window === 'undefined') return ''
  return localStorage.getItem('user_token') || ''
})

const coverUrl = computed(() => {
  const g = game.value
  if (!g) return ''
  return String(g.cover || g.header_image || '').trim()
})

const heroBg = computed(() => {
  const g = game.value
  if (!g) return ''
  return String(g.banner || g.cover || g.header_image || '').trim()
})

const videoPoster = computed(() => {
  const g = game.value
  if (!g) return ''
  return String(g.banner || g.header_image || g.cover || '').trim()
})

const videoUrl = computed(() => {
  const g = game.value
  if (!g) return ''
  return String(g.video_url || '').trim()
})

const galleryUrls = computed(() => {
  const g = game.value
  const raw = g?.gallery
  if (!Array.isArray(raw)) return []
  return raw.map((x) => String(x).trim()).filter(Boolean)
})

const galleryViewerVisible = ref(false)
const galleryViewerUrl = ref('')
const openGalleryViewer = (u: string) => {
  galleryViewerUrl.value = u
  galleryViewerVisible.value = true
}

const descriptionHtml = computed(() => {
  const d = game.value?.description
  return d ? String(d) : ''
})

const mergedResources = computed(() => {
  const fromDetail = game.value?.resources
  if (Array.isArray(fromDetail) && fromDetail.length) return fromDetail
  return resources.value
})

const normalizeLinks = (item: PublicGameResource) => {
  const fromArray = Array.isArray(item.download_urls) ? item.download_urls.map((x) => String(x).trim()).filter(Boolean) : []
  if (fromArray.length > 0) return fromArray
  const raw = String(item.download_url || '').trim()
  if (!raw) return []
  return raw
    .split(/[\n\r\t ,;，；]+/)
    .map((x) => x.trim())
    .filter((x) => /^https?:\/\//i.test(x))
}

const resourceLabelMap: Record<string, string> = {
  game: '本体下载',
  mod: 'Mod 下载',
  trainer: '修改器下载',
  submission: '玩家投稿',
}

const resourceGroups = computed(() => {
  const grouped = new Map<string, Array<PublicGameResource & { link_list: string[] }>>()
  for (const item of mergedResources.value as PublicGameResource[]) {
    const key = String(item.resource_type || 'game').trim() || 'game'
    const linkList = normalizeLinks(item)
    if (linkList.length === 0) continue
    if (!grouped.has(key)) grouped.set(key, [])
    grouped.get(key)?.push({ ...item, link_list: linkList })
  }

  return Array.from(grouped.entries()).map(([key, items]) => ({
    key,
    label: resourceLabelMap[key] || key,
    items,
  }))
})

const myGameSubmissions = computed(() =>
  mySubmissions.value.filter((item) => Number(item.game_id || 0) === Number(gameId.value)),
)

const formatDate = (value?: string) => {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return String(value)
  return date.toISOString().slice(0, 10)
}

const detectPanLabel = (link: string) => {
  const url = String(link).toLowerCase()
  if (url.includes('pan.quark.cn')) return '夸克网盘'
  if (url.includes('pan.baidu.com')) return '百度网盘'
  if (url.includes('pan.xunlei.com')) return '迅雷网盘'
  if (url.includes('aliyundrive.com') || url.includes('alipan.com')) return '阿里云盘'
  if (url.includes('cloud.189.cn')) return '天翼云盘'
  if (url.includes('drive.uc.cn') || url.includes('drive-h.uc.cn')) return 'UC 网盘'
  if (url.includes('115.com')) return '115 网盘'
  if (url.includes('123pan') || url.includes('123684.com') || url.includes('123685.com')) return '123 云盘'
  return '下载链接'
}

const getDownloadLabel = (link: string, index: number, total: number) => {
  const label = detectPanLabel(link)
  return total > 1 ? `${label} ${index + 1}` : label
}

const goList = () => router.push('/games')

const resetSubmissionForm = () => {
  submissionForm.value = {
    title: '',
    link: '',
    extract_code: '',
    tags: '',
    description: '',
  }
}

const copyLink = async (link: string) => {
  try {
    await navigator.clipboard.writeText(link)
    ElMessage.success('链接已复制')
  } catch {
    ElMessage.error('复制失败，请手动复制')
  }
}

const loadMySubmissions = async () => {
  if (!userToken.value) {
    mySubmissions.value = []
    return
  }
  const { data: res } = await siteMySubmissions()
  if (res.code === 200 && Array.isArray(res.data)) {
    mySubmissions.value = res.data
  }
}

const submitGameResource = async () => {
  const title = String(submissionForm.value.title || '').trim()
  const link = String(submissionForm.value.link || '').trim()
  if (!title) return ElMessage.warning('请填写资源标题')
  if (!link) return ElMessage.warning('请填写下载链接')
  if (!gameId.value) return ElMessage.warning('当前游戏 ID 无效')

  submittingSubmission.value = true
  try {
    const { data: res } = await siteSubmissionCreate({
      title,
      link,
      game_id: gameId.value,
      extract_code: String(submissionForm.value.extract_code || '').trim(),
      tags: String(submissionForm.value.tags || '').trim(),
      description: String(submissionForm.value.description || '').trim(),
    })
    if (res.code !== 200) return
    ElMessage.success('提交成功，等待审核')
    resetSubmissionForm()
    await loadMySubmissions()
  } finally {
    submittingSubmission.value = false
  }
}

const load = async () => {
  if (!gameId.value) {
    loading.value = false
    game.value = null
    resources.value = []
    videoStarted.value = false
    return
  }
  loading.value = true
  try {
    const [dRes, rRes] = await Promise.all([
      publicGameDetail(gameId.value),
      publicGameResourceList(gameId.value),
    ])
    if (dRes.data.code !== 200 || !dRes.data.data) {
      game.value = null
      resources.value = []
      return
    }
    game.value = dRes.data.data
    resources.value = rRes.data.code === 200 && Array.isArray(rRes.data.data) ? rRes.data.data : []
    videoStarted.value = false
    await loadMySubmissions()
  } finally {
    loading.value = false
  }
}

watch(
  () => gameId.value,
  () => {
    void load()
  },
  { immediate: true },
)
</script>

<style scoped>
.game-detail {
  max-width: 1100px;
  margin: 0 auto;
  padding: 12px 16px 32px;
}
.toolbar {
  margin-bottom: 8px;
}
.hero {
  border-radius: 14px;
  overflow: hidden;
  background: linear-gradient(135deg, #1e293b, #0f172a);
  background-size: cover;
  background-position: center;
  margin-bottom: 16px;
}
.hero-inner {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  padding: 24px;
  align-items: flex-start;
}
.cover {
  width: 200px;
  max-width: 40vw;
  border-radius: 10px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.35);
  object-fit: cover;
  aspect-ratio: 2/3;
  background: #334155;
}
.title-block {
  flex: 1;
  min-width: 200px;
  color: #f8fafc;
}
.title {
  margin: 0 0 12px;
  font-size: 1.75rem;
  line-height: 1.25;
}
.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  margin-bottom: 10px;
}
.price {
  font-size: 14px;
  opacity: 0.9;
}
.short {
  margin: 0;
  font-size: 14px;
  line-height: 1.6;
  opacity: 0.88;
  max-width: 720px;
}
.meta-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 10px;
  margin-top: 16px;
}
.meta-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px 12px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.08);
}
.meta-label {
  font-size: 12px;
  opacity: 0.75;
}
.meta-value,
.meta-link {
  color: #f8fafc;
  font-size: 13px;
  word-break: break-word;
}
.video-wrap {
  max-width: 820px;
  border-radius: 12px;
  overflow: hidden;
  background: #0f172a;
}
.video-poster {
  position: relative;
  width: 100%;
  padding: 0;
  border: 0;
  display: block;
  cursor: pointer;
  background: #020617;
}
.video-poster-image,
.video-poster-fallback {
  width: 100%;
  aspect-ratio: 16 / 9;
  display: block;
  object-fit: cover;
}
.video-poster-fallback {
  background: linear-gradient(135deg, #0f172a, #1e293b);
}
.video-poster-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #fff;
  background: linear-gradient(180deg, rgba(2, 6, 23, 0.16), rgba(2, 6, 23, 0.6));
}
.video-play-button {
  width: 64px;
  height: 64px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  background: rgba(15, 23, 42, 0.78);
  box-shadow: 0 12px 28px rgba(0, 0, 0, 0.28);
}
.video-play-text {
  font-size: 14px;
  font-weight: 600;
}
.video-player {
  width: 100%;
  aspect-ratio: 16 / 9;
  display: block;
  background: #000;
  position: relative;
  overflow: hidden;
}

.video-player :deep(.rp-wrap) {
  position: absolute;
  inset: 0;
}
.section {
  border-radius: 14px;
  margin-bottom: 16px;
}
.gallery {
  width: 100%;
}

.gallery-slide {
  width: 100%;
  height: 100%;
  cursor: pointer;
  border-radius: 8px;
  overflow: hidden;
  background: #f1f5f9;
}

.g-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.gallery-viewer {
  width: 100%;
  padding: 10px 0;
  display: grid;
  place-items: center;
}

.viewer-img {
  max-width: 100%;
  max-height: 70vh;
  object-fit: contain;
  border-radius: 10px;
}
.desc {
  font-size: 14px;
  line-height: 1.7;
  color: #334155;
  word-break: break-word;
}
.desc :deep(img) {
  max-width: 100%;
  height: auto;
}
.links {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.resource-groups {
  display: flex;
  flex-direction: column;
  gap: 18px;
}
.resource-group {
  padding-top: 4px;
}
.resource-group-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}
.resource-group-head h3 {
  margin: 0;
  font-size: 16px;
  color: #0f172a;
}
.resource-group-head span {
  color: #64748b;
  font-size: 12px;
}
.resource-list {
  display: grid;
  gap: 12px;
}
.resource-card {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  justify-content: space-between;
  padding: 14px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #f8fafc;
}
.resource-main {
  flex: 1;
  min-width: 240px;
}
.resource-title-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 8px;
}
.resource-title {
  margin: 0;
  font-size: 15px;
  color: #0f172a;
}
.resource-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  color: #475569;
  font-size: 13px;
}
.meta-pill {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-radius: 999px;
  background: #eef6ff;
  border: 1px solid #d7e8fb;
  color: #35506b;
}
.meta-pill-icon {
  width: 20px;
  height: 20px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #dcecff;
  color: #255ea8;
  font-size: 11px;
  font-weight: 800;
}
.resource-links {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}
.download-action {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.download-link {
  padding: 8px 14px;
  border-radius: 999px;
  background: linear-gradient(135deg, #ecfeff, #dcfce7);
  font-weight: 700;
}
.muted {
  color: #94a3b8;
  font-size: 13px;
}
.submission-form {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 14px;
  background: #fafcff;
}
.submission-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}
.submission-actions {
  display: flex;
  gap: 8px;
}
.submission-hint {
  margin-top: 8px;
  font-size: 12px;
  color: #64748b;
}
.submission-history {
  margin-top: 18px;
}
.submission-history-title {
  margin-bottom: 10px;
  font-size: 15px;
  font-weight: 700;
  color: #0f172a;
}
.submission-login {
  display: grid;
  place-items: center;
  gap: 10px;
  padding: 10px 0 4px;
}
@media (max-width: 720px) {
  .hero-inner {
    padding: 18px;
  }
  .cover {
    width: 150px;
  }
  .resource-card {
    padding: 12px;
  }
  .submission-grid {
    grid-template-columns: 1fr;
  }
}
</style>
