<template>
  <div class="resource-detail-page">
    <el-skeleton :loading="loading" animated :rows="7">
      <template v-if="data">
        <div class="top-grid">
          <el-card class="main-card" shadow="never">
            <div class="detail-layout">
              <div class="poster-col">
                <div class="cover-box poster-box">
                  <img
                    v-if="showResourceCover"
                    :src="coverDisplayUrl"
                    :alt="String(data.title || '资源封面')"
                    class="cover-image poster-image"
                    loading="lazy"
                    decoding="async"
                    referrerpolicy="no-referrer"
                    @error="onCoverImageError"
                  />
                  <div v-else class="poster-placeholder">暂无封面</div>
                </div>
                <div class="qr-box poster-qr">
                  <img v-if="qrDataUrl" :src="qrDataUrl" alt="二维码" />
                  <div v-else class="qr-placeholder">生成中…</div>
                  <span>扫码保存链接</span>
                </div>
              </div>

              <div class="info-col">
                <div class="label-row">
                  <div class="seq-badge" title="资源编号">
                    <span class="seq-icon" aria-hidden="true">
                      <svg
                        viewBox="0 0 24 24"
                        width="14"
                        height="14"
                        fill="none"
                        xmlns="http://www.w3.org/2000/svg"
                      >
                        <path
                          d="M10.5 13.5L13.5 10.5"
                          stroke="currentColor"
                          stroke-width="2"
                          stroke-linecap="round"
                          stroke-linejoin="round"
                        />
                        <path
                          d="M9 6.5H7.8C5.253 6.5 4 7.753 4 10.3V11.5"
                          stroke="currentColor"
                          stroke-width="2"
                          stroke-linecap="round"
                        />
                        <path
                          d="M15 17.5H16.2C18.747 17.5 20 16.247 20 13.7V12.5"
                          stroke="currentColor"
                          stroke-width="2"
                          stroke-linecap="round"
                        />
                        <path
                          d="M6.5 9V7.8C6.5 5.253 7.753 4 10.3 4H11.5"
                          stroke="currentColor"
                          stroke-width="2"
                          stroke-linecap="round"
                        />
                        <path
                          d="M17.5 15V16.2C17.5 18.747 16.247 20 13.7 20H12.5"
                          stroke="currentColor"
                          stroke-width="2"
                          stroke-linecap="round"
                        />
                      </svg>
                    </span>
                    <span class="seq-text">No. {{ data.id || id }}</span>
                  </div>
                  <el-tag type="primary" effect="light" round>资源详情</el-tag>
                </div>

                <h1 class="title" :title="String(data.title || '未命名资源')">
                  {{ data.title || '未命名资源' }}
                </h1>

                <div class="desc-box">{{ fullText }}</div>

                <div class="meta-row">
                  <span class="dot" />分享时间
                  <span class="value">{{ formatDate(data.share_at || data.created_at) }}</span>
                </div>
                <div class="meta-row">
                  <span class="dot" />收录时间
                  <span class="value">{{ formatDate(data.created_at) }}</span>
                </div>
                <div class="meta-row">
                  <span class="dot" />网盘平台
                  <span class="value">{{ platformSummary }}</span>
                </div>
                <div class="meta-row tags-row">
                  <span class="dot" />关键词标签
                  <div class="tags">
                    <el-tag
                      v-for="tag in tags"
                      :key="tag"
                      effect="light"
                      size="small"
                      class="clickable-tag"
                      @click="goTagSearch(tag)"
                    >
                      {{ tag }}
                    </el-tag>
                    <span v-if="tags.length === 0" class="empty">暂无关键词</span>
                  </div>
                </div>

                <div class="bottom-actions">
                  <div ref="actionButtonsMount" class="bottom-actions-mount" />
                </div>

                <div class="action-area">
                  <div class="tip-box">
                    <strong>温馨提示</strong>
                    如遇网盘资源失效或其他问题，请点击上方「查看资源 / 反馈问题」按钮向我们反馈。
                  </div>
                </div>
              </div>
            </div>
          </el-card>

          <el-card class="side-card" shadow="never">
            <template #header>
              <div class="side-title">最新资源</div>
            </template>
            <div class="side-list">
              <button
                v-for="(item, idx) in latestList"
                :key="item.id"
                class="side-item"
                @click="goDetail(item.id)"
              >
                <span class="item-idx">{{ idx + 1 }}</span>
                <span class="item-title">{{ item.title }}</span>
              </button>
              <el-empty v-if="latestList.length === 0" description="暂无资源" :image-size="64" />
            </div>
          </el-card>
        </div>

        <el-card v-if="transferLog?.exists" class="filterlog-card" shadow="never">
          <template #header>
            <div class="block-title">文件内容</div>
          </template>
          <div class="filterlog-meta">
            <div class="meta-row">
              <span class="dot" />时间
              <span class="value">{{ formatDate(String(transferLog.created_at || '')) }}</span>
            </div>
          </div>

          <div v-if="transferLog.filter_log" class="filterlog-body">
            <div class="filterlog-stats">
              <el-tag effect="light" type="primary">文件 {{ transferLog.filter_log.total_files ?? 0 }}</el-tag>
              <el-tag effect="light" type="primary">文件夹 {{ transferLog.filter_log.total_folders ?? 0 }}</el-tag>
              <el-button
                v-if="(transferLog.filter_log.ad_files || []).length > 0"
                size="small"
                type="primary"
                plain
                @click="adFilesDialogVisible = true"
              >
                查看广告文件
              </el-button>
            </div>
            <div class="filterlog-tree">
              <div
                v-for="(line, idx) in transferLog.filter_log.structure || []"
                :key="idx"
                class="tree-line"
              >
                {{ line }}
              </div>
              <el-empty
                v-if="(transferLog.filter_log.structure || []).length === 0"
                description="暂无目录结构"
                :image-size="64"
              />
            </div>
          </div>
        </el-card>

        <el-dialog v-model="adFilesDialogVisible" title="广告文件清理明细" width="980px">
          <el-table :data="adFiles" size="small" style="width: 100%">
            <el-table-column prop="name" label="名称" min-width="180" show-overflow-tooltip />
            <el-table-column prop="path" label="路径" min-width="280" show-overflow-tooltip />
            <el-table-column prop="keyword" label="命中词" width="140" show-overflow-tooltip />
            <el-table-column prop="fid" label="FID" width="160" show-overflow-tooltip />
          </el-table>
          <template #footer>
            <el-button @click="adFilesDialogVisible = false">关闭</el-button>
          </template>
        </el-dialog>

        <el-alert class="safe-alert" type="info" show-icon :closable="false">
          资源链接由网络收集整理，请先自行甄别安全性与有效性，如失效或内容有误可提交反馈。
        </el-alert>

        <div class="bottom-grid">
          <el-card class="list-card" shadow="never">
            <template #header>
              <div class="block-title">相关推荐</div>
            </template>
            <div class="list-inner">
              <button
                v-for="(item, idx) in recommendList"
                :key="item.id"
                class="line-item"
                @click="goDetail(item.id)"
              >
                <span class="item-idx small">{{ idx + 1 }}</span>
                <span class="item-title line">{{ item.title }}</span>
              </button>
              <el-empty v-if="recommendList.length === 0" description="暂无推荐" :image-size="64" />
            </div>
          </el-card>

          <el-card class="list-card" shadow="never">
            <template #header>
              <div class="block-title">同类新资源</div>
            </template>
            <div class="list-inner">
              <button
                v-for="(item, idx) in userOtherList"
                :key="item.id"
                class="line-item"
                @click="goDetail(item.id)"
              >
                <span class="item-idx small">{{ idx + 1 }}</span>
                <span class="item-title line">{{ item.title }}</span>
              </button>
              <el-empty v-if="userOtherList.length === 0" description="暂无内容" :image-size="64" />
            </div>
          </el-card>
        </div>
      </template>

      <el-empty v-else description="资源不存在或已下线" :image-size="96" />
    </el-skeleton>

    <el-dialog
      v-model="resolvingDialogVisible"
      title="正在获取"
      width="420px"
      :show-close="false"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :destroy-on-close="false"
      append-to-body
    >
      <div class="resolve-loading">
        <span class="resolve-spinner" aria-hidden="true" />
        <div class="resolve-copy">
          <div class="resolve-title">正在为你获取可用链接</div>
          <p>{{ resolvingText }}</p>
        </div>
      </div>
    </el-dialog>

    <el-dialog
      v-model="feedbackDialogVisible"
      title="问题反馈"
      width="520px"
      :destroy-on-close="true"
      @closed="feedbackForm.content = ''"
    >
      <el-form :model="feedbackForm" label-width="90px">
        <el-form-item label="问题类型">
          <el-select v-model="feedbackForm.type" style="width: 100%">
            <el-option label="链接失效" value="link_invalid" />
            <el-option label="提取码错误/缺失" value="password_error" />
            <el-option label="资源内容不符/缺失" value="content_error" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="反馈内容">
          <el-input
            v-model="feedbackForm.content"
            type="textarea"
            :rows="5"
            placeholder="请描述你遇到的问题，方便我们尽快处理"
          />
        </el-form-item>
        <el-form-item label="联系方式">
          <el-input v-model="feedbackForm.contact" placeholder="邮箱 / QQ / 微信 / 手机号（选填）" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="feedbackSubmitting" @click="feedbackSubmit">提交</el-button>
          <el-button @click="feedbackDialogVisible = false">取消</el-button>
        </el-form-item>
      </el-form>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'
import { createElement } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import QRCode from 'qrcode'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { feedbackCreate } from '@/api/feedback'
import ActionButtonsReact from './ActionButtonsReact'
import {
  siteResourceAccessLink,
  siteResourceDetail,
  siteResourceLatestTransferLog,
  siteResourcePage,
} from '@/api/netdisk'
import { runtimeConfig } from '@/config/runtimeConfig'
import { buildProxiedImageSrc, shouldApplyTgCoverProxy } from '@/utils/coverProxy'
import {
  netdiskLinkDisplayRank,
  netdiskPlatformKey,
  platformText,
} from '@/views/public/search/searchHelpers'

const route = useRoute()
const router = useRouter()

const id = computed(() => String(route.params.id || ''))
const loading = ref(false)
const data = ref<any | null>(null)
const latestList = ref<any[]>([])
const recommendList = ref<any[]>([])
const userOtherList = ref<any[]>([])
const transferLog = ref<any | null>(null)
const adFilesDialogVisible = ref(false)
const resolvingLink = ref(false)
const resolvingDialogVisible = ref(false)
const resolvingText = ref('正在检查资源并生成可用链接...')
const actionButtonsMount = ref<HTMLElement | null>(null)
let actionButtonsRoot: Root | null = null

const adFiles = computed(() => {
  const list = (transferLog.value?.filter_log?.ad_files || []) as any[]
  return list.map((x) => ({
    name: String(x?.name || '').trim(),
    path: String(x?.path || '').trim(),
    keyword: String(x?.keyword || '').trim(),
    fid: String(x?.fid || '').trim(),
  }))
})

const feedbackDialogVisible = ref(false)
const feedbackSubmitting = ref(false)
const feedbackForm = ref({
  type: 'link_invalid' as 'link_invalid' | 'password_error' | 'content_error' | 'other',
  content: '',
  contact: '',
})

const feedbackOpen = () => {
  feedbackForm.value.type = 'link_invalid'
  feedbackForm.value.content = ''
  feedbackForm.value.contact = ''
  feedbackDialogVisible.value = true
}

const feedbackSubmit = async () => {
  const rid = Number(data.value?.id || 0)
  if (!rid) {
    ElMessage.warning('资源不存在')
    return
  }
  const content = String(feedbackForm.value.content || '').trim()
  if (!content) {
    ElMessage.warning('请填写反馈内容')
    return
  }

  feedbackSubmitting.value = true
  try {
    const payload = {
      resource_id: rid,
      type: feedbackForm.value.type,
      content,
      contact: String(feedbackForm.value.contact || '').trim(),
    }
    const { data: res } = await feedbackCreate(payload)
    if (res.code !== 200) return
    ElMessage.success('反馈提交成功')
    feedbackDialogVisible.value = false
  } finally {
    feedbackSubmitting.value = false
  }
}

const setMeta = (name: string, content: string) => {
  if (!content) return
  let meta = document.querySelector(`meta[name='${name}']`) as HTMLMetaElement | null
  if (!meta) {
    meta = document.createElement('meta')
    meta.name = name
    document.head.appendChild(meta)
  }
  meta.content = content
}

const setOpenGraphMeta = (property: string, content: string) => {
  if (!content) return
  let meta = document.querySelector(`meta[property='${property}']`) as HTMLMetaElement | null
  if (!meta) {
    meta = document.createElement('meta')
    meta.setAttribute('property', property)
    document.head.appendChild(meta)
  }
  meta.content = content
}

const setJsonLd = (payload: Record<string, any>) => {
  const nodeId = 'resource-jsonld'
  let script = document.getElementById(nodeId) as HTMLScriptElement | null
  if (!script) {
    script = document.createElement('script')
    script.id = nodeId
    script.type = 'application/ld+json'
    document.head.appendChild(script)
  }
  script.textContent = JSON.stringify(payload)
}

const parseTagList = (value: unknown) => {
  if (Array.isArray(value)) {
    return value.map((item) => String(item || '').trim()).filter(Boolean)
  }
  const text = String(value || '').trim()
  if (!text) return []

  if (text.startsWith('[') && text.endsWith(']')) {
    try {
      const parsed = JSON.parse(text)
      if (Array.isArray(parsed)) {
        return parsed.map((item) => String(item || '').trim()).filter(Boolean)
      }
    } catch {}
  }

  return text
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
}

const applyDetailSeo = () => {
  const d = data.value
  if (!d) return

  const siteTitle = runtimeConfig.siteTitle || '资源库'
  const title = String(d.title || '').trim()
  document.title = title ? `${title} - ${siteTitle}` : `资源详情 - ${siteTitle}`

  const desc = String(d.description || '').trim() || runtimeConfig.seoDescription || ''
  setMeta('description', desc)

  const keywords = parseTagList(d.tags)
  setMeta('keywords', [title, ...keywords].filter(Boolean).slice(0, 12).join(','))

  setOpenGraphMeta('og:title', document.title)
  setOpenGraphMeta('og:description', desc)
  setOpenGraphMeta('og:type', 'website')

  const coverAbs = coverDisplayUrl.value

  if (coverAbs) {
    setOpenGraphMeta('og:image', coverAbs)
    setOpenGraphMeta('twitter:image', coverAbs)
  }

  setJsonLd({
    '@context': 'https://schema.org',
    '@type': 'CreativeWork',
    name: title || '资源详情',
    description: desc,
    keywords,
    url: window.location.href,
    image: coverAbs || undefined,
    datePublished: d.created_at || undefined,
    dateModified: d.updated_at || d.created_at || undefined,
    identifier: String(d.id || ''),
  })
}

const tags = computed(() => parseTagList(data.value?.tags))

const allShareLinks = computed(() => {
  const main = String(data.value?.link || '').trim()
  const raw = data.value?.extra_links
  const extras = Array.isArray(raw)
    ? raw.map((x: unknown) => String(x || '').trim()).filter(Boolean)
    : []
  const out: string[] = []
  const seen = new Set<string>()
  for (const u of [main, ...extras]) {
    if (!u || seen.has(u)) continue
    seen.add(u)
    out.push(u)
  }
  out.sort((a, b) => {
    const ra = netdiskLinkDisplayRank(a)
    const rb = netdiskLinkDisplayRank(b)
    if (ra !== rb) return ra - rb
    return a.localeCompare(b)
  })
  return out
})

const platformSummary = computed(() => {
  const links = allShareLinks.value
  if (links.length === 0) return '—'
  const names = links.map((l) => platformText(l))
  return [...new Set(names)].join(' · ')
})

const coverImageError = ref(false)

const coverDisplayUrl = computed(() => {
  const coverRaw = String(data.value?.cover || '').trim()
  if (!coverRaw) return ''
  let resolved: string
  if (coverRaw.startsWith('//')) {
    try {
      resolved = `${window.location.protocol}${coverRaw}`
    } catch {
      return ''
    }
  } else if (coverRaw.startsWith('http://') || coverRaw.startsWith('https://')) {
    resolved = coverRaw
  } else {
    resolved = `${window.location.origin}${coverRaw.startsWith('/') ? '' : '/'}${coverRaw}`
  }

  const isRemoteCover = /^https?:\/\//i.test(coverRaw) || coverRaw.startsWith('//')
  const tmpl = String(runtimeConfig.tgImageProxyUrl || '').trim()
  if (tmpl && isRemoteCover && shouldApplyTgCoverProxy(data.value, coverRaw)) {
    return buildProxiedImageSrc(resolved, tmpl)
  }
  return resolved
})

const showResourceCover = computed(() => Boolean(coverDisplayUrl.value) && !coverImageError.value)

const onCoverImageError = () => {
  coverImageError.value = true
}

watch(
  () => `${id.value}|${String(data.value?.cover || '').trim()}`,
  () => {
    coverImageError.value = false
  },
)

const fullText = computed(() => {
  const description = String(data.value?.description || '').trim()
  if (!description) return '暂无资源简介'
  return description
})

const currentLink = computed(() => {
  const link = allShareLinks.value[0] || String(data.value?.link || '').trim()
  if (link) return link
  return window.location.href
})

const qrDataUrl = ref('')

const updateLocalQr = async () => {
  const text = String(currentLink.value || '').trim()
  if (!text) {
    qrDataUrl.value = ''
    return
  }
  try {
    qrDataUrl.value = await QRCode.toDataURL(text, {
      width: 100,
      margin: 1,
      errorCorrectionLevel: 'M',
      color: { dark: '#000000ff', light: '#ffffffff' },
    })
  } catch {
    qrDataUrl.value = ''
  }
}

watch(currentLink, () => {
  void updateLocalQr()
}, { immediate: true })

const formatDate = (value?: string) => (value ? dayjs(value).format('YYYY-MM-DD HH:mm:ss') : '-')

const loadAsideLists = async () => {
  const [latestRes, hotRes] = await Promise.all([
    siteResourcePage({ page: 1, page_size: 10, sort: 'latest' }),
    siteResourcePage({ page: 1, page_size: 10, sort: 'hot' }),
  ])
  latestList.value = (latestRes.data.code === 200 ? latestRes.data.data?.list : []) || []
  recommendList.value = (hotRes.data.code === 200 ? hotRes.data.data?.list : []) || []
  userOtherList.value = latestList.value.filter((x) => String(x.id) !== id.value).slice(0, 10)
}

const load = async () => {
  if (!id.value) {
    data.value = null
    return
  }

  loading.value = true
  try {
    const [{ data: detailRes }, { data: tlogRes }] = await Promise.all([
      siteResourceDetail(id.value),
      siteResourceLatestTransferLog(id.value),
      loadAsideLists(),
    ])
    data.value = detailRes.code === 200 ? detailRes.data : null
    transferLog.value = tlogRes.code === 200 ? tlogRes.data : null
    applyDetailSeo()

    if (data.value?.id) {
      userOtherList.value = userOtherList.value.filter((x) => String(x.id) !== String(data.value.id))
      latestList.value = latestList.value.filter((x) => String(x.id) !== String(data.value.id))
      recommendList.value = recommendList.value.filter((x) => String(x.id) !== String(data.value.id))
    }
  } finally {
    loading.value = false
  }
}

const goDetail = (targetId: string | number) => {
  if (String(targetId) === id.value) return
  router.push(`/r/${targetId}`)
}

const goTagSearch = (tag: string) => {
  const keyword = String(tag || '').trim()
  if (!keyword) return
  router.push(`/tag/${encodeURIComponent(keyword)}`)
}

const sleep = (ms: number) => new Promise((resolve) => window.setTimeout(resolve, ms))

/** 将 access-link 返回的 links / extra_links 写回详情，便于未整页刷新时更新多链接列表 */
const mergeLinksFromAccessPayload = (payload: Record<string, unknown>) => {
  if (!data.value) return
  const rawLinks = payload.links
  const arr = Array.isArray(rawLinks)
    ? rawLinks.map((x) => String(x || '').trim()).filter(Boolean)
    : []
  if (arr.length > 0) {
    data.value = {
      ...data.value,
      link: arr[0],
      extra_links: arr.slice(1),
    }
    void updateLocalQr()
    return
  }
  const main = String(payload.link || data.value.link || '').trim()
  const ex = payload.extra_links
  const extras = Array.isArray(ex)
    ? ex.map((x) => String(x || '').trim()).filter(Boolean)
    : []
  if (main || extras.length > 0) {
    data.value = {
      ...data.value,
      link: main || data.value.link,
      extra_links: extras,
    }
    void updateLocalQr()
  }
}

/** 走 access-link 流程后，打开用户所选网盘（转存后 URL 可能变化，按平台匹配）。 */
const openAccessLinkForRow = async (clickedUrl: string) => {
  if (!id.value || resolvingLink.value) return
  const prefKey = netdiskPlatformKey(clickedUrl)

  resolvingLink.value = true
  resolvingDialogVisible.value = true
  resolvingText.value = '正在检查资源并生成可用链接...'

  try {
    for (let attempt = 0; attempt < 20; attempt += 1) {
      const { data: res } = await siteResourceAccessLink(id.value)
      if (res.code !== 200) {
        throw new Error(res.message || '获取链接失败')
      }

      const payload = (res.data || {}) as Record<string, unknown>
      const status = String(payload.status || '')
      const link = String(payload.link || '').trim()
      const freshShare = Boolean(payload.fresh_share)
      const message = String(payload.message || '').trim() || '正在检查资源并生成可用链接...'
      resolvingText.value = message

      if ((status === 'success' || status === 'direct') && link) {
        // 每次点击重新分享：仅合并本次返回的链接打开网盘，不整页刷新（避免把库内展示链覆盖回来）
        if (status === 'success' && !freshShare) {
          await load()
        } else {
          mergeLinksFromAccessPayload(payload)
        }
        const links = allShareLinks.value
        const openUrl =
          links.find((u) => u === clickedUrl) ||
          links.find((u) => netdiskPlatformKey(u) === prefKey) ||
          links[0] ||
          link
        window.open(openUrl, '_blank')
        return
      }

      if (status === 'failed') {
        throw new Error(message || '获取链接失败')
      }

      await sleep(1500)
    }

    throw new Error('获取超时，请稍后重试')
  } catch (error: any) {
    ElMessage.error(String(error?.message || '获取链接失败'))
  } finally {
    resolvingLink.value = false
    resolvingDialogVisible.value = false
  }
}

const copyText = async (text: string, okText: string) => {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success(okText)
  } catch {
    ElMessage.error('复制失败，请手动复制')
  }
}

const copyLink = async () => {
  const lines = allShareLinks.value
  const text =
    lines.length > 1 ? lines.join('\n') : lines[0] || String(data.value?.link || '').trim() || window.location.href
  await copyText(text, lines.length > 1 ? '全部链接已复制' : '链接已复制')
}

const shareCurrent = async () => {
  const page = window.location.href
  const lines = allShareLinks.value
  if (lines.length === 0) {
    await copyText(page, '页面地址已复制')
    return
  }
  const text = `${page}\n\n${lines.join('\n')}`
  await copyText(text, '页面地址与资源链接已复制')
}

const shareActionRows = computed(() =>
  allShareLinks.value.map((url) => ({
    label: platformText(url),
    url,
  })),
)

const renderActionButtons = () => {
  const mountEl = actionButtonsMount.value
  if (!mountEl) {
    actionButtonsRoot?.unmount()
    actionButtonsRoot = null
    return
  }
  if (!actionButtonsRoot) actionButtonsRoot = createRoot(mountEl)
  actionButtonsRoot.render(
    createElement(ActionButtonsReact, {
      shareRows: shareActionRows.value,
      opening: resolvingLink.value,
      onCopyLink: copyLink,
      onOpenRow: openAccessLinkForRow,
      onCopyPage: shareCurrent,
      onFeedback: feedbackOpen,
    }),
  )
}

onMounted(() => {
  renderActionButtons()
})

onBeforeUnmount(() => {
  actionButtonsRoot?.unmount()
  actionButtonsRoot = null
})

watch(
  () => [
    actionButtonsMount.value,
    data.value?.id,
    String(data.value?.link || ''),
    JSON.stringify(data.value?.extra_links || []),
    JSON.stringify(shareActionRows.value),
    resolvingLink.value,
  ],
  () => {
    renderActionButtons()
  },
  { immediate: true },
)

watch(
  () => id.value,
  async () => {
    await load()
  },
  { immediate: true },
)

watch(
  () => data.value?.id,
  () => {
    if (data.value?.id) applyDetailSeo()
  },
)
</script>

<style scoped>
.resource-detail-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 16px 8px;
}

.top-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  gap: 14px;
  align-items: start;
}

.main-card {
  border-radius: 12px;
  border-top: 3px solid var(--el-color-primary);
}

.detail-layout {
  display: grid;
  grid-template-columns: 248px minmax(0, 1fr);
  gap: 20px;
  align-items: start;
}

.poster-col {
  display: grid;
  gap: 12px;
  align-self: start;
}

.poster-box {
  margin-bottom: 0;
}

.poster-image {
  width: 100%;
  aspect-ratio: 2 / 3;
  max-height: unset;
  object-fit: cover;
}

.poster-placeholder {
  width: 100%;
  aspect-ratio: 2 / 3;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: var(--el-fill-color-light);
  color: var(--el-text-color-secondary);
  font-size: 13px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.poster-qr {
  margin-top: 4px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  padding: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.poster-qr img {
  width: 110px;
  height: 110px;
  object-fit: cover;
}

.info-col {
  min-width: 0;
}

.label-row {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 14px;
}

.seq-badge {
  display: flex;
  gap: 8px;
  align-items: center;
  padding: 4px 10px;
  border-radius: 999px;
  border: 1px solid var(--el-border-color-lighter);
  background: color-mix(in srgb, var(--el-color-primary) 8%, transparent);
  color: var(--el-text-color-primary);
}

.seq-icon {
  width: 22px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: color-mix(in srgb, var(--el-color-primary) 18%, transparent);
  color: var(--el-color-primary);
}

.seq-text {
  font-size: 12px;
  font-weight: 700;
}

.item-idx {
  width: 26px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: color-mix(in srgb, var(--el-color-primary) 12%, transparent);
  color: var(--el-color-primary);
  font-size: 12px;
  flex-shrink: 0;
}

.item-idx.small {
  width: 24px;
  height: 20px;
  font-size: 11px;
}

.item-title {
  flex: 1;
  min-width: 0;
  color: inherit;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-word;
}

.item-title.line {
  -webkit-line-clamp: 2;
}

.title {
  margin: 0 0 14px;
  font-size: 34px;
  line-height: 1.25;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-word;
}

.desc-box {
  max-height: 340px;
  overflow: auto;
  white-space: pre-wrap;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  padding: 12px;
  line-height: 1.8;
  margin-bottom: 14px;
  color: var(--el-text-color-regular);
}

.cover-box {
  margin-bottom: 14px;
}

.cover-image {
  display: block;
  width: 100%;
  max-height: 360px;
  object-fit: contain;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: #f6f7f9;
}

.meta-row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 34px;
  color: var(--el-text-color-regular);
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.value {
  margin-left: 14px;
  color: var(--el-text-color-primary);
}

.dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--el-color-primary);
}

.tags-row {
  align-items: flex-start;
  padding: 8px 0;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-left: 14px;
}

.clickable-tag {
  cursor: pointer;
  transition: transform 0.15s ease;
}

.clickable-tag:hover {
  transform: translateY(-1px);
}

.empty {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.action-area {
  margin-top: 14px;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 140px;
  gap: 12px;
  align-items: center;
}

.tip-box {
  min-height: 82px;
  border: 1px solid color-mix(in srgb, var(--el-color-warning) 30%, var(--el-border-color-lighter));
  border-radius: 8px;
  background: color-mix(in srgb, var(--el-color-warning-light-9) 72%, #fff);
  color: var(--el-text-color-secondary);
  padding: 10px 12px;
  line-height: 1.8;
}

.tip-box strong {
  color: var(--el-color-warning-dark-2);
  margin-right: 8px;
}

.qr-box {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  padding: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.qr-box img {
  width: 100px;
  height: 100px;
  object-fit: cover;
}

.qr-placeholder {
  width: 100px;
  height: 100px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-light);
  border-radius: 4px;
}

.qr-box span {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.bottom-actions {
  margin-top: 12px;
}

.bottom-actions-mount {
  width: 100%;
}

.side-card,
.list-card {
  border-radius: 12px;
}

.side-title,
.block-title {
  font-weight: 700;
  border-left: 3px solid var(--el-color-primary);
  padding-left: 8px;
}

.side-list,
.list-inner {
  display: grid;
  gap: 8px;
}

.side-item,
.line-item {
  text-align: left;
  border: 0;
  background: transparent;
  padding: 6px 0;
  color: var(--el-text-color-regular);
  line-height: 1.6;
  cursor: pointer;
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.side-item:hover,
.line-item:hover {
  color: var(--el-color-primary);
}

.safe-alert {
  margin-top: 14px;
  margin-bottom: 14px;
}

.resolve-loading {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 8px 4px 12px;
}

.resolve-spinner {
  width: 42px;
  height: 42px;
  border-radius: 999px;
  border: 3px solid color-mix(in srgb, var(--el-color-primary) 18%, white);
  border-top-color: var(--el-color-primary);
  animation: resolve-spin 0.9s linear infinite;
  flex: 0 0 auto;
}

.resolve-copy {
  min-width: 0;
}

.resolve-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--el-text-color-primary);
  margin-bottom: 6px;
}

.resolve-copy p {
  margin: 0;
  color: var(--el-text-color-regular);
  line-height: 1.7;
}

@keyframes resolve-spin {
  to {
    transform: rotate(360deg);
  }
}

.filterlog-card {
  border-radius: 12px;
  margin-top: 14px;
}

.filterlog-stats {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 10px;
}

.filterlog-tree {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  padding: 10px 12px;
  max-height: 320px;
  overflow: auto;
  background: var(--el-fill-color-lighter);
}

.tree-line {
  font-family:
    ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New',
    monospace;
  font-size: 12px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}

.bottom-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

@media (max-width: 960px) {
  .top-grid {
    grid-template-columns: 1fr;
  }

  .detail-layout {
    grid-template-columns: 1fr;
  }

  .poster-col {
    grid-template-columns: minmax(0, 240px) 1fr;
    align-items: start;
  }

  .bottom-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .poster-col {
    grid-template-columns: 1fr;
  }

  .title {
    font-size: 24px;
    -webkit-line-clamp: 3;
  }

  .action-area {
    grid-template-columns: 1fr;
  }

  .bottom-actions {
    flex-wrap: wrap;
    gap: 8px;
  }
}

</style>
