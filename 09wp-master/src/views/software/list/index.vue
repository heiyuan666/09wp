<template>
  <div class="panel">
    <div class="panel-title">软件管理</div>

    <div class="bar">
      <el-form inline>
        <el-form-item label="关键词">
          <el-input v-model="keyword" placeholder="软件名称" clearable />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="categoryId" placeholder="全部" clearable style="width: 180px">
            <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="版本">
          <el-input v-model="versionFilter" placeholder="版本号" clearable style="width: 160px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchList">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
          <el-button type="success" @click="openCreate">新增软件</el-button>
        </el-form-item>
      </el-form>
    </div>

    <el-table :data="list" size="small" style="width: 100%" v-loading="loading">
      <el-table-column prop="id" label="ID" width="90" />
      <el-table-column label="图标" width="72">
        <template #default="{ row }">
          <el-image
            v-if="row.icon_thumb || row.icon"
            :src="row.icon_thumb || row.icon"
            fit="cover"
            style="width: 40px; height: 40px; border-radius: 8px"
            :preview-src-list="row.icon ? [row.icon] : row.icon_thumb ? [row.icon_thumb] : []"
            preview-teleported
          />
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="封面" width="90">
        <template #default="{ row }">
          <el-image
            v-if="row.cover_thumb || row.cover"
            :src="row.cover_thumb || row.cover"
            fit="cover"
            style="width: 48px; height: 48px; border-radius: 6px"
            :preview-src-list="row.cover ? [row.cover] : []"
            preview-teleported
          />
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip />
      <el-table-column prop="version" label="版本" width="120" />
      <el-table-column prop="size" label="大小" width="110" />
      <el-table-column prop="platforms" label="平台" min-width="140" show-overflow-tooltip />
      <el-table-column prop="status" label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '上架' : '下架' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="updated_at" label="更新时间" width="180" />
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="primary" @click="goVersions(row)">版本</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        :total="total"
        @size-change="fetchList"
        @current-change="fetchList"
      />
    </div>

    <el-dialog v-model="visible" :title="form.id ? '编辑软件' : '新增软件'" width="860px">
      <el-form :model="form" label-width="110px">
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="软件名称">
              <el-input v-model="form.name" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="软件分类">
              <el-select v-model="form.category_id" placeholder="请选择分类" style="width: 100%">
                <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="软件简介">
          <el-input v-model="form.summary" type="textarea" :rows="3" />
        </el-form-item>

        <el-row :gutter="12">
          <el-col :span="8">
            <el-form-item label="版本号">
              <el-input v-model="form.version" placeholder="如 1.2.3" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="软件大小">
              <el-input v-model="form.size" placeholder="如 128MB" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="状态">
              <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="系统平台">
          <el-checkbox-group v-model="form.platforms_arr">
            <el-checkbox label="Windows" />
            <el-checkbox label="Mac" />
            <el-checkbox label="Linux" />
            <el-checkbox label="Android" />
            <el-checkbox label="iOS" />
          </el-checkbox-group>
        </el-form-item>

        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="发布日期">
              <el-date-picker
                v-model="form.published_at"
                type="date"
                value-format="YYYY-MM-DD"
                placeholder="可选"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="更新时间">
              <el-date-picker
                v-model="form.updated_at_override"
                type="date"
                value-format="YYYY-MM-DD"
                placeholder="可选"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="软件官网">
          <el-input v-model="form.website" placeholder="https://..." />
        </el-form-item>

        <el-form-item label="软件图标">
          <div class="media-upload-row">
            <el-upload
              :show-file-list="false"
              :auto-upload="true"
              :before-upload="beforeUpload"
              :http-request="uploadIcon"
            >
              <el-button type="primary">上传图标</el-button>
            </el-upload>
            <span class="media-hint">方形小图，用于列表；建议约 256×256，将生成缩略图</span>
            <div v-if="form.icon" class="thumb-with-remove">
              <el-image
                :src="form.icon_thumb || form.icon"
                fit="cover"
                style="width: 64px; height: 64px; border-radius: 8px"
                :preview-src-list="[form.icon]"
                preview-teleported
              />
              <el-button type="danger" circle size="small" class="remove-img-btn" @click="clearIcon">
                <el-icon><Close /></el-icon>
              </el-button>
            </div>
          </div>
        </el-form-item>

        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="封面上传">
              <div class="media-upload-row">
                <el-upload
                  :show-file-list="false"
                  :auto-upload="true"
                  :before-upload="beforeUpload"
                  :http-request="uploadCover"
                >
                  <el-button type="primary">上传封面</el-button>
                </el-upload>
                <div v-if="form.cover" class="thumb-with-remove">
                  <el-image
                    :src="form.cover_thumb || form.cover"
                    fit="cover"
                    style="width: 64px; height: 64px; border-radius: 8px"
                    :preview-src-list="[form.cover]"
                    preview-teleported
                  />
                  <el-button type="danger" circle size="small" class="remove-img-btn" @click="clearCover">
                    <el-icon><Close /></el-icon>
                  </el-button>
                </div>
              </div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="截图（可选）">
              <el-input v-model="form.screenshots_text" type="textarea" :rows="2" placeholder="每行一个图片 URL" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="直链下载">
              <el-input
                v-model="form.download_direct_text"
                type="textarea"
                :rows="3"
                placeholder="每行一个直链 URL"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="网盘下载">
              <el-input v-model="form.download_pan_text" type="textarea" :rows="3" placeholder="每行一个网盘 URL" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="网盘提取码">
          <el-input v-model="form.download_extract" placeholder="可选" style="max-width: 240px" />
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
import { ref, onMounted, reactive } from 'vue'
import { Close } from '@element-plus/icons-vue'
import router from '@/router'
import {
  softwareCategoryList,
  softwareCreate,
  softwareDelete,
  softwareList,
  softwareUpdate,
  softwareUploadCover,
  type ISoftwareCategory,
  type ISoftwareItem,
} from '@/api/software'

const loading = ref(false)
const keyword = ref('')
const categoryId = ref<number | undefined>(undefined)
const versionFilter = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref<ISoftwareItem[]>([])

const categories = ref<ISoftwareCategory[]>([])
const visible = ref(false)
const form = reactive<any>({
  id: 0,
  category_id: 0,
  name: '',
  summary: '',
  version: '',
  cover: '',
  cover_thumb: '',
  icon: '',
  icon_thumb: '',
  screenshots_text: '',
  size: '',
  platforms_arr: [] as string[],
  website: '',
  download_direct_text: '',
  download_pan_text: '',
  download_extract: '',
  published_at: '',
  updated_at_override: '',
  status: 1,
})

const parseLines = (raw: string) =>
  String(raw || '')
    .split('\n')
    .map((s) => s.trim())
    .filter(Boolean)

const splitPlatforms = (raw: string) =>
  String(raw || '')
    .split(/[\r\n\t,，;；/|]+/g)
    .map((s) => s.trim())
    .filter(Boolean)

const loadCategories = async () => {
  const { data: res } = await softwareCategoryList()
  if (res.code !== 200) return
  categories.value = Array.isArray(res.data?.list) ? res.data.list : []
}

const fetchList = async () => {
  loading.value = true
  try {
    const { data: res } = await softwareList({
      page: page.value,
      page_size: pageSize.value,
      keyword: keyword.value || undefined,
      category_id: categoryId.value || undefined,
      version: versionFilter.value || undefined,
    })
    if (res.code === 200 && res.data) {
      list.value = Array.isArray(res.data.list) ? res.data.list : []
      total.value = Number(res.data.total || 0)
    }
  } finally {
    loading.value = false
  }
}

const resetSearch = () => {
  keyword.value = ''
  categoryId.value = undefined
  versionFilter.value = ''
  page.value = 1
  fetchList()
}

const resetForm = () => {
  Object.assign(form, {
    id: 0,
    category_id: categories.value?.[0]?.id || 0,
    name: '',
    summary: '',
    version: '',
    cover: '',
    cover_thumb: '',
    icon: '',
    icon_thumb: '',
    screenshots_text: '',
    size: '',
    platforms_arr: [] as string[],
    website: '',
    download_direct_text: '',
    download_pan_text: '',
    download_extract: '',
    published_at: '',
    updated_at_override: '',
    status: 1,
  })
}

const openCreate = () => {
  resetForm()
  visible.value = true
}

const openEdit = (row: ISoftwareItem) => {
  resetForm()
  Object.assign(form, row)
  form.icon = row.icon || ''
  form.icon_thumb = row.icon_thumb || ''
  form.platforms_arr = splitPlatforms(row.platforms)
  form.screenshots_text = (Array.isArray(row.screenshots) ? row.screenshots : []).join('\n')
  form.download_direct_text = (Array.isArray(row.download_direct) ? row.download_direct : []).join('\n')
  form.download_pan_text = (Array.isArray(row.download_pan) ? row.download_pan : []).join('\n')
  form.published_at = row.published_at ? String(row.published_at).slice(0, 10) : ''
  form.updated_at_override = row.updated_at_override ? String(row.updated_at_override).slice(0, 10) : ''
  visible.value = true
}

const save = async () => {
  const payload: any = {
    name: form.name,
    summary: form.summary,
    category_id: form.category_id,
    version: form.version,
    cover: form.cover,
    cover_thumb: form.cover_thumb,
    icon: form.icon,
    icon_thumb: form.icon_thumb,
    screenshots: parseLines(form.screenshots_text),
    size: form.size,
    platforms: form.platforms_arr,
    website: form.website,
    download_direct: parseLines(form.download_direct_text),
    download_pan: parseLines(form.download_pan_text),
    download_extract: form.download_extract,
    published_at: form.published_at || '',
    updated_at_override: form.updated_at_override || '',
    status: form.status,
  }
  if (form.id) await softwareUpdate(String(form.id), payload)
  else await softwareCreate(payload)
  visible.value = false
  await fetchList()
}

const remove = async (row: ISoftwareItem) => {
  await softwareDelete(String(row.id))
  await fetchList()
}

const goVersions = (row: ISoftwareItem) => {
  router.push({
    path: '/admin/software/versions',
    query: { software_id: String(row.id) },
  })
}

const beforeUpload = (file: File) => {
  const ok = ['image/jpeg', 'image/png', 'image/webp'].includes(file.type)
  if (!ok) return false
  const maxMB = 5
  if (file.size > maxMB * 1024 * 1024) return false
  return true
}

const uploadCover = async (opt: any) => {
  const file = opt?.file as File
  const { data: res } = await softwareUploadCover(file, 'cover')
  if (res.code !== 200) return
  form.cover = res.data?.url || ''
  form.cover_thumb = res.data?.thumb_url || ''
}

const uploadIcon = async (opt: any) => {
  const file = opt?.file as File
  const { data: res } = await softwareUploadCover(file, 'icon')
  if (res.code !== 200) return
  form.icon = res.data?.url || ''
  form.icon_thumb = res.data?.thumb_url || ''
}

const clearCover = () => {
  form.cover = ''
  form.cover_thumb = ''
}

const clearIcon = () => {
  form.icon = ''
  form.icon_thumb = ''
}

onMounted(async () => {
  await loadCategories()
  await fetchList()
})
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
.bar {
  margin-bottom: 10px;
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 8px;
}
.pager {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}
.media-upload-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}
.media-hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  flex: 1 1 200px;
}
.thumb-with-remove {
  position: relative;
  display: inline-block;
  vertical-align: middle;
}
.remove-img-btn {
  position: absolute;
  top: -6px;
  right: -6px;
  padding: 4px !important;
}
</style>
