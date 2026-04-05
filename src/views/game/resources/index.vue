<template>
  <div class="panel">
    <div class="panel-title">游戏下载资源</div>
    <div class="filters">
      <el-select
        v-model="query.game_id"
        placeholder="选择已发布游戏"
        style="width: 340px"
        filterable
        @change="onGameChange"
      >
        <el-option v-for="g in games" :key="g.id" :label="`${g.title} (#${g.id})`" :value="g.id" />
      </el-select>
      <el-select v-model="query.resource_type" placeholder="资源分类" clearable style="width: 160px" @change="load">
        <el-option v-for="item in resourceTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
      </el-select>
      <div class="spacer" />
      <el-button type="primary" :disabled="!query.game_id" @click="openCreate">新增资源</el-button>
    </div>

    <el-table :data="list" style="width: 100%">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="title" label="资源标题" min-width="220" />
      <el-table-column label="资源分类" width="120">
        <template #default="{ row }">
          <el-tag :type="resourceTypeTagType(row.resource_type)">{{ resourceTypeLabel(row.resource_type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="version" label="版本" width="120" />
      <el-table-column prop="size" label="大小" width="100" />
      <el-table-column prop="download_type" label="下载方式" min-width="120" />
      <el-table-column prop="pan_type" label="网盘类型" min-width="140" />
      <el-table-column label="下载地址" min-width="220">
        <template #default="{ row }">
          <div class="link-cell">
            <span>{{ linkCount(row.download_url) }} 条</span>
            <el-link
              v-if="firstLink(row.download_url)"
              :href="firstLink(row.download_url)"
              target="_blank"
              type="primary"
              :underline="false"
            >
              查看首条
            </el-link>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="tested" label="是否测试" width="100">
        <template #default="{ row }">
          <el-tag :type="row.tested ? 'success' : 'info'">{{ row.tested ? '已测试' : '未测试' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="260">
        <template #default="{ row }">
          <el-button link type="success" @click="transferAndReplaceRow(row)">转存替换</el-button>
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="visible" :title="form.id ? '编辑资源' : '新增资源'" width="760px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="资源标题">
          <el-input v-model="form.title" />
        </el-form-item>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="资源分类">
              <el-select v-model="form.resource_type" style="width: 100%">
                <el-option v-for="item in resourceTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="版本号">
              <el-input v-model="form.version" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="资源大小">
              <el-input v-model="form.size" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="下载方式">
              <el-input v-model="form.download_type" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="网盘类型">
              <el-select v-model="form.pan_type" style="width: 100%">
                <el-option label="自动识别" value="" />
                <el-option label="迅雷" value="迅雷" />
                <el-option label="百度" value="百度" />
                <el-option label="夸克" value="夸克" />
                <el-option label="阿里" value="阿里" />
                <el-option label="115" value="115" />
                <el-option label="天翼" value="天翼" />
                <el-option label="123" value="123" />
                <el-option label="UC" value="UC" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="是否测试">
              <el-switch v-model="form.tested" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="下载地址">
          <el-input
            v-model="form.download_url"
            type="textarea"
            :autosize="{ minRows: 4, maxRows: 10 }"
            placeholder="支持多个网盘地址，每行一个链接。保存时自动去重并识别网盘类型。"
          />
        </el-form-item>
        <el-form-item label="转存选项">
          <el-switch
            v-model="form.transfer_before_save"
            active-text="保存前转存并替换为本人链接"
            inactive-text="直接保存原始链接"
          />
          <el-button style="margin-left: 12px" type="success" plain :loading="transferLoading" @click="previewTransfer">
            立即转存并预填
          </el-button>
        </el-form-item>
        <el-form-item label="识别结果">
          <div class="recognized">
            <el-tag v-for="tag in detectedPanTypes" :key="tag" type="success" effect="light">{{ tag }}</el-tag>
            <span v-if="detectedPanTypes.length === 0" class="hint">暂未识别到有效网盘链接</span>
          </div>
        </el-form-item>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="发布者">
              <el-input v-model="form.author" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="发布日期">
              <el-date-picker v-model="form.publish_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { gameList, gameResourceCreate, gameResourceDelete, gameResourceList, gameResourceUpdate } from '@/api/game'
import { netdiskTransferBatchByLinks } from '@/api/netdisk'

const splitDownloadSep = /[\r\n\t ,;，；]+/
const resourceTypeOptions = [
  { label: '本体', value: 'game' },
  { label: 'MOD', value: 'mod' },
  { label: '修改器', value: 'trainer' },
  { label: '用户投稿', value: 'submission' },
]

const normalizeLinks = (raw: string) => {
  const items = String(raw || '').split(splitDownloadSep)
  const set = new Set<string>()
  const links: string[] = []
  for (const part of items) {
    const v = part.trim().replace(/[),.;，；]+$/, '')
    if (!v) continue
    if (!/^https?:\/\//i.test(v)) continue
    if (set.has(v)) continue
    set.add(v)
    links.push(v)
  }
  return links
}

const detectPanTypeByLink = (link: string) => {
  const u = String(link || '').toLowerCase()
  if (u.includes('pan.baidu.com')) return '百度'
  if (u.includes('pan.quark.cn')) return '夸克'
  if (u.includes('pan.xunlei.com')) return '迅雷'
  if (u.includes('aliyundrive.com') || u.includes('alipan.com')) return '阿里'
  if (u.includes('cloud.189.cn')) return '天翼'
  if (u.includes('115.com')) return '115'
  if (u.includes('123684.com') || u.includes('123pan.com')) return '123'
  if (u.includes('drive.uc.cn')) return 'UC'
  return '其他'
}

const resourceTypeLabel = (value?: string) => {
  return resourceTypeOptions.find((item) => item.value === value)?.label || '本体'
}

const resourceTypeTagType = (value?: string) => {
  if (value === 'mod') return 'success'
  if (value === 'trainer') return 'warning'
  if (value === 'submission') return 'danger'
  return 'primary'
}

const games = ref<any[]>([])
const list = ref<any[]>([])
const visible = ref(false)
const route = useRoute()
const router = useRouter()

const query = reactive<any>({
  game_id: undefined,
  resource_type: '',
})

const form = reactive<any>({
  id: 0,
  title: '',
  resource_type: 'game',
  version: '',
  size: '',
  download_type: '',
  pan_type: '',
  download_url: '',
  tested: false,
  author: '',
  publish_date: '',
  transfer_before_save: false,
})
const transferLoading = ref(false)

const detectedPanTypes = computed(() => {
  const links = normalizeLinks(form.download_url)
  const set = new Set<string>()
  for (const link of links) set.add(detectPanTypeByLink(link))
  return Array.from(set)
})

const currentGame = computed(() => games.value.find((x: any) => Number(x.id) === Number(query.game_id)))
const linkCount = (downloadURL: string) => normalizeLinks(downloadURL).length
const firstLink = (downloadURL: string) => normalizeLinks(downloadURL)[0] || ''

const parseRouteGameID = () => {
  const raw = route.query.game_id
  const id = Number(Array.isArray(raw) ? raw[0] : raw)
  return Number.isFinite(id) && id > 0 ? id : undefined
}

const syncRouteGameID = async (id?: number) => {
  const nextQuery = { ...route.query } as Record<string, any>
  if (id) nextQuery.game_id = String(id)
  else delete nextQuery.game_id
  await router.replace({ query: nextQuery })
}

const loadGames = async () => {
  const { data: res } = await gameList({ page: 1, page_size: 200 })
  if (res.code !== 200) return
  games.value = res.data?.list || []

  const routeGameID = parseRouteGameID()
  if (routeGameID) {
    const exists = games.value.some((x: any) => Number(x.id) === routeGameID)
    if (exists) {
      query.game_id = routeGameID
      return
    }
  }

  if (!query.game_id && games.value.length > 0) query.game_id = games.value[0].id
}

const load = async () => {
  if (!query.game_id) {
    list.value = []
    return
  }
  const { data: res } = await gameResourceList(query.game_id)
  if (res.code !== 200) return
  const rows = Array.isArray(res.data) ? res.data : []
  list.value = query.resource_type ? rows.filter((item: any) => item.resource_type === query.resource_type) : rows
}

const onGameChange = async (val: number) => {
  query.game_id = val
  await syncRouteGameID(Number(val))
  await load()
}

const openCreate = () => {
  Object.assign(form, {
    id: 0,
    title: currentGame.value?.title || '',
    resource_type: 'game',
    version: '',
    size: '',
    download_type: '',
    pan_type: '',
    download_url: '',
    tested: false,
    author: '',
    publish_date: '',
    transfer_before_save: false,
  })
  visible.value = true
}

const openEdit = (row: any) => {
  Object.assign(form, {
    ...row,
    resource_type: row.resource_type || 'game',
    download_url: String(row.download_url || ''),
    transfer_before_save: false,
  })
  visible.value = true
}

const transferLinksToOwn = async (links: string[]) => {
  if (links.length === 0) return { ownLinks: [] as string[], success: 0, failed: 0 }
  const { data: res } = await netdiskTransferBatchByLinks({ items: links.map((link) => ({ link })) })
  if (res.code !== 200) {
    throw new Error(res.message || '转存失败')
  }
  const ownLinks: string[] = []
  const seen = new Set<string>()
  const results = Array.isArray(res.data?.results) ? res.data.results : []
  for (const row of results) {
    const ownShareURL = String(row?.data?.own_share_url || '').trim()
    if (!ownShareURL || seen.has(ownShareURL)) continue
    seen.add(ownShareURL)
    ownLinks.push(ownShareURL)
  }
  return {
    ownLinks,
    success: Number(res.data?.success || 0),
    failed: Number(res.data?.failed || 0),
  }
}

const previewTransfer = async () => {
  const links = normalizeLinks(form.download_url)
  if (links.length === 0) {
    ElMessage.warning('请先填写有效下载地址')
    return
  }
  transferLoading.value = true
  try {
    const result = await transferLinksToOwn(links)
    if (result.ownLinks.length === 0) {
      ElMessage.warning('转存完成，但未拿到本人分享链接，请检查网盘凭证配置')
      return
    }
    form.download_url = result.ownLinks.join('\n')
    form.pan_type = ''
    ElMessage.success(`转存完成：成功 ${result.success}，失败 ${result.failed}，已回填 ${result.ownLinks.length} 条本人链接`)
  } catch (err: any) {
    ElMessage.error(err?.message || '转存失败')
  } finally {
    transferLoading.value = false
  }
}

const save = async () => {
  let links = normalizeLinks(form.download_url)
  if (links.length === 0) {
    ElMessage.warning('请至少填写一条有效下载地址')
    return
  }

  if (form.transfer_before_save) {
    transferLoading.value = true
    try {
      const result = await transferLinksToOwn(links)
      if (result.ownLinks.length === 0) {
        ElMessage.warning('未获取到本人分享链接，已取消保存，请检查网盘凭证或链接有效性')
        return
      }
      links = result.ownLinks
      form.download_url = links.join('\n')
      form.pan_type = ''
      ElMessage.success(`转存完成：成功 ${result.success}，失败 ${result.failed}，将保存本人链接`)
    } catch (err: any) {
      ElMessage.error(err?.message || '转存失败')
      return
    } finally {
      transferLoading.value = false
    }
  }

  const payload = {
    game_id: Number(query.game_id),
    title: form.title,
    resource_type: form.resource_type,
    version: form.version,
    size: form.size,
    download_type: form.download_type,
    pan_type: form.pan_type,
    download_url: links.join('\n'),
    download_urls: links,
    tested: !!form.tested,
    author: form.author,
    publish_date: form.publish_date,
  }
  if (form.id) await gameResourceUpdate(form.id, payload)
  else await gameResourceCreate(payload)
  visible.value = false
  await load()
}

const remove = async (row: any) => {
  await gameResourceDelete(row.id)
  await load()
}

const transferAndReplaceRow = async (row: any) => {
  const links = normalizeLinks(String(row.download_url || ''))
  if (links.length === 0) {
    ElMessage.warning('该资源没有可转存的下载地址')
    return
  }
  transferLoading.value = true
  try {
    const result = await transferLinksToOwn(links)
    if (result.ownLinks.length === 0) {
      ElMessage.warning('转存完成，但未拿到本人分享链接')
      return
    }
    await gameResourceUpdate(row.id, {
      download_url: result.ownLinks.join('\n'),
      download_urls: result.ownLinks,
      pan_type: '',
    })
    ElMessage.success(`已替换为本人链接，共 ${result.ownLinks.length} 条`)
    await load()
  } catch (err: any) {
    ElMessage.error(err?.message || '转存替换失败')
  } finally {
    transferLoading.value = false
  }
}

onMounted(async () => {
  await loadGames()
  await syncRouteGameID(query.game_id ? Number(query.game_id) : undefined)
  await load()
})

watch(
  () => route.query.game_id,
  async () => {
    const routeGameID = parseRouteGameID()
    if (!routeGameID || Number(routeGameID) === Number(query.game_id)) return
    const exists = games.value.some((x: any) => Number(x.id) === routeGameID)
    if (!exists) return
    query.game_id = routeGameID
    await load()
  },
)
</script>

<style scoped>
.panel {
  padding: 14px;
  border-radius: 14px;
  background: #fff;
}

.panel-title {
  font-weight: 800;
  margin-bottom: 10px;
}

.filters {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 10px;
}

.spacer {
  flex: 1;
}

.recognized {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.hint {
  font-size: 12px;
  color: #909399;
}

.link-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
