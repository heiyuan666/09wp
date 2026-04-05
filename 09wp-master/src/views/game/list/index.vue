<template>
  <div class="panel">
    <div class="panel-title">游戏列表</div>
    <div class="filters">
      <el-input v-model="query.keyword" placeholder="按游戏名称搜索" clearable style="width: 260px" />
      <el-select v-model="query.category_id" clearable placeholder="分类" style="width: 180px">
        <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
      </el-select>
      <el-button type="primary" @click="load">查询</el-button>
      <el-button @click="reset">重置</el-button>
      <div class="spacer" />
      <el-button type="primary" @click="openCreate">新增游戏</el-button>
    </div>

    <el-table :data="list" style="width: 100%">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="title" label="游戏名称" min-width="200" />
      <el-table-column prop="steam_appid" label="Steam AppID" width="120" />
      <el-table-column prop="price_text" label="价格" width="120" />
      <el-table-column prop="type" label="类型" min-width="120" show-overflow-tooltip />
      <el-table-column prop="developer" label="开发商" min-width="140" show-overflow-tooltip />
      <el-table-column prop="release_date" label="发行日期" width="120" />
      <el-table-column prop="tags" label="标签" min-width="200" show-overflow-tooltip />
      <el-table-column label="操作" width="320">
        <template #default="{ row }">
          <el-button link type="success" @click="goResources(row)">资源管理</el-button>
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.page_size"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        :total="total"
        @change="load"
      />
    </div>

    <el-dialog v-model="visible" :title="form.id ? '编辑游戏' : '新增游戏'" width="980px" destroy-on-close>
      <el-form :model="form" label-width="110px">
        <el-row :gutter="12">
          <el-col :span="8">
            <el-form-item label="Steam AppID">
              <el-input v-model="form.steam_appid" placeholder="例如 570" />
            </el-form-item>
          </el-col>
          <el-col :span="5">
            <el-form-item label="地区">
              <el-input v-model="form.steam_cc" placeholder="cn" />
            </el-form-item>
          </el-col>
          <el-col :span="5">
            <el-form-item label="语言">
              <el-input v-model="form.steam_l" placeholder="schinese" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label=" ">
              <el-button type="primary" plain :loading="steamLoading" @click="fetchFromSteam">
                Steam 一键填充
              </el-button>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="游戏名称">
              <el-input v-model="form.title" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="游戏分类">
              <el-select v-model="form.category_id" clearable style="width: 100%">
                <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="封面">
              <el-input v-model="form.cover" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="顶部大图">
              <el-input v-model="form.banner" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="Header 图">
              <el-input v-model="form.header_image" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="视频">
              <el-input v-model="form.video_url" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="官网">
              <el-input v-model="form.website" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="发行日期">
              <el-date-picker v-model="form.release_date" type="date" style="width: 100%" value-format="YYYY-MM-DD" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="12">
          <el-col :span="8">
            <el-form-item label="价格文案">
              <el-input v-model="form.price_text" placeholder="如：¥ 198 / 免费" />
            </el-form-item>
          </el-col>
          <el-col :span="5">
            <el-form-item label="币种">
              <el-input v-model="form.price_currency" placeholder="CNY" />
            </el-form-item>
          </el-col>
          <el-col :span="5">
            <el-form-item label="折扣%">
              <el-input-number v-model="form.price_discount" :min="0" :max="100" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="现价(分)">
              <el-input-number v-model="form.price_final" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="12">
          <el-col :span="8">
            <el-form-item label="原价(分)">
              <el-input-number v-model="form.price_initial" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="类型">
              <el-input v-model="form.type" placeholder="动作/RPG/射击等" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="开发商">
              <el-input v-model="form.developer" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="12">
          <el-col :span="8">
            <el-form-item label="大小">
              <el-input v-model="form.size" placeholder="例如 70GB" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="发行商">
              <el-input v-model="form.publishers" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="标签">
              <el-input v-model="form.tags" placeholder="例如：多人/竞技/FPS" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="游戏类型标签">
              <el-input v-model="form.genres" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="简短介绍">
              <el-input v-model="form.short_description" type="textarea" :rows="2" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="12">
          <el-col :span="8">
            <el-form-item label="Meta 评分">
              <el-input-number v-model="form.metacritic_score" :min="0" :max="100" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="综合评分">
              <el-input-number v-model="form.rating" :min="0" :max="10" :step="0.1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="Steam 好评率">
              <el-input-number v-model="form.steam_score" :min="0" :max="100" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="12">
          <el-col :span="8">
            <el-form-item label="下载次数">
              <el-input-number v-model="form.downloads" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="点赞数">
              <el-input-number v-model="form.likes" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="点踩数">
              <el-input-number v-model="form.dislikes" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="截图(每行1个URL)">
          <el-input v-model="form.gallery_text" type="textarea" :rows="4" />
        </el-form-item>
        <el-form-item label="游戏介绍">
          <el-input v-model="form.description" type="textarea" :rows="4" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { gameCategoryList, gameCreate, gameDelete, gameList, gameSteamAppDetail, gameUpdate } from '@/api/game'

const router = useRouter()
const categories = ref<any[]>([])
const list = ref<any[]>([])
const total = ref(0)
const visible = ref(false)
const steamLoading = ref(false)

const query = reactive({
  page: 1,
  page_size: 20,
  keyword: '',
  category_id: undefined as number | undefined,
})

const form = reactive({
  id: 0,
  category_id: undefined as number | undefined,
  steam_appid: '',
  steam_cc: 'cn',
  steam_l: 'schinese',
  title: '',
  cover: '',
  banner: '',
  video_url: '',
  short_description: '',
  header_image: '',
  website: '',
  publishers: '',
  genres: '',
  tags: '',
  price_text: '',
  price_currency: '',
  price_initial: 0,
  price_final: 0,
  price_discount: 0,
  metacritic_score: 0,
  description: '',
  release_date: '',
  size: '',
  type: '',
  developer: '',
  rating: 0,
  steam_score: 0,
  downloads: 0,
  likes: 0,
  dislikes: 0,
  gallery_text: '',
})

const loadCategories = async () => {
  const { data: res } = await gameCategoryList()
  if (res.code !== 200) return
  categories.value = res.data || []
}

const load = async () => {
  const { data: res } = await gameList(query as Record<string, unknown>)
  if (res.code !== 200) return
  list.value = res.data?.list || []
  total.value = res.data?.total || 0
}

const reset = async () => {
  Object.assign(query, { page: 1, page_size: 20, keyword: '', category_id: undefined })
  await load()
}

const resetForm = () => {
  Object.assign(form, {
    id: 0,
    category_id: undefined,
    steam_appid: '',
    steam_cc: 'cn',
    steam_l: 'schinese',
    title: '',
    cover: '',
    banner: '',
    video_url: '',
    short_description: '',
    header_image: '',
    website: '',
    publishers: '',
    genres: '',
    tags: '',
    price_text: '',
    price_currency: '',
    price_initial: 0,
    price_final: 0,
    price_discount: 0,
    metacritic_score: 0,
    description: '',
    release_date: '',
    size: '',
    type: '',
    developer: '',
    rating: 0,
    steam_score: 0,
    downloads: 0,
    likes: 0,
    dislikes: 0,
    gallery_text: '',
  })
}

const openCreate = () => {
  resetForm()
  visible.value = true
}

const openEdit = (row: any) => {
  resetForm()
  Object.assign(form, {
    ...row,
    steam_appid: row.steam_appid ? String(row.steam_appid) : '',
    steam_cc: 'cn',
    steam_l: 'schinese',
    gallery_text: (row.gallery || []).join('\n'),
  })
  visible.value = true
}

const goResources = (row: any) => {
  router.push({
    path: '/admin/game/resources',
    query: { game_id: String(row.id) },
  })
}

const parseLines = (raw: string) =>
  String(raw || '')
    .split('\n')
    .map((s) => s.trim())
    .filter(Boolean)

const normalizeSteamReleaseDate = (raw: string) => {
  const s = String(raw || '').trim()
  if (!s) return ''

  const iso = s.match(/^(\d{4})-(\d{1,2})-(\d{1,2})/)
  if (iso?.[1] && iso[2] && iso[3]) {
    return `${iso[1]}-${iso[2].padStart(2, '0')}-${iso[3].padStart(2, '0')}`
  }

  const cn = s.match(/^(\d{4})\s*年\s*(\d{1,2})\s*月\s*(\d{1,2})\s*日$/)
  if (cn?.[1] && cn[2] && cn[3]) {
    return `${cn[1]}-${cn[2].padStart(2, '0')}-${cn[3].padStart(2, '0')}`
  }

  const parsed = new Date(s)
  if (!Number.isNaN(parsed.getTime())) {
    const y = parsed.getFullYear()
    const m = `${parsed.getMonth() + 1}`.padStart(2, '0')
    const d = `${parsed.getDate()}`.padStart(2, '0')
    return `${y}-${m}-${d}`
  }
  return ''
}

const fetchFromSteam = async () => {
  const appid = String(form.steam_appid || '').trim()
  if (!appid) {
    ElMessage.warning('请先输入 Steam AppID')
    return
  }

  steamLoading.value = true
  try {
    const { data: res } = await gameSteamAppDetail(appid, {
      cc: String(form.steam_cc || 'cn').trim() || 'cn',
      l: String(form.steam_l || 'schinese').trim() || 'schinese',
    })
    if (res.code !== 200 || !res.data) {
      ElMessage.error(res.message || '获取 Steam 信息失败')
      return
    }

    const d = res.data
    const screenshots = Array.isArray(d.screenshots) ? d.screenshots : []
    const genres = Array.isArray(d.genres) ? d.genres : []
    const tags = Array.isArray(d.tags) ? d.tags : []
    const categoriesFromSteam = Array.isArray(d.categories) ? d.categories : []
    const publishers = Array.isArray(d.publishers) ? d.publishers : []
    const developers = Array.isArray(d.developers) ? d.developers : []
    const releaseDate = normalizeSteamReleaseDate(String(d.release_date || ''))

    form.title = d.name || form.title
    form.cover = d.header_image || form.cover
    form.header_image = d.header_image || form.header_image
    form.banner = d.background_raw || d.capsule_image || form.banner
    form.video_url = d.video_url || form.video_url
    form.short_description = d.short_description || form.short_description
    form.description = d.detailed_description || d.about_the_game || form.description
    form.website = d.website || form.website
    form.publishers = publishers.join('/') || form.publishers
    form.genres = genres.join('/') || form.genres
    form.tags = tags.join('/') || categoriesFromSteam.join('/') || form.tags
    form.type = genres.join('/') || categoriesFromSteam.join('/') || form.type
    form.developer = developers.join('/') || form.developer
    form.metacritic_score = Number(d.metacritic_score || 0)
    form.steam_score = Number(d.metacritic_score || form.steam_score || 0)
    form.price_text = String(d.price_text || '')
    form.price_currency = String(d.price_currency || '')
    form.price_initial = Number(d.price_initial || 0)
    form.price_final = Number(d.price_final || 0)
    form.price_discount = Number(d.price_discount || 0)
    if (releaseDate) form.release_date = releaseDate
    if (screenshots.length > 0) form.gallery_text = screenshots.slice(0, 20).join('\n')

    ElMessage.success('Steam 信息已填充，请检查后保存')
  } finally {
    steamLoading.value = false
  }
}

const save = async () => {
  const payload = {
    category_id: form.category_id || undefined,
    steam_appid: Number(form.steam_appid || 0),
    title: form.title,
    cover: form.cover,
    banner: form.banner,
    video_url: form.video_url,
    short_description: form.short_description,
    header_image: form.header_image,
    website: form.website,
    publishers: form.publishers,
    genres: form.genres,
    tags: form.tags,
    price_text: form.price_text,
    price_currency: form.price_currency,
    price_initial: Number(form.price_initial || 0),
    price_final: Number(form.price_final || 0),
    price_discount: Number(form.price_discount || 0),
    metacritic_score: Number(form.metacritic_score || 0),
    description: form.description,
    release_date: form.release_date,
    size: form.size,
    type: form.type,
    developer: form.developer,
    rating: Number(form.rating || 0),
    steam_score: Number(form.steam_score || 0),
    downloads: Number(form.downloads || 0),
    likes: Number(form.likes || 0),
    dislikes: Number(form.dislikes || 0),
    gallery: parseLines(form.gallery_text),
  }

  if (!payload.title?.trim()) {
    ElMessage.warning('请填写游戏名称')
    return
  }

  if (form.id) {
    await gameUpdate(form.id, payload)
  } else {
    await gameCreate(payload)
  }
  visible.value = false
  await load()
}

const remove = async (row: any) => {
  await gameDelete(row.id)
  await load()
}

onMounted(async () => {
  await loadCategories()
  await load()
})
</script>

<style scoped>
.panel {
  padding: 14px;
  border-radius: 14px;
  background: #fff;
}

.panel-title {
  font-weight: 700;
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

.pager {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}
</style>
