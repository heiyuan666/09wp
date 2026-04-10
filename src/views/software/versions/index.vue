<template>
  <div class="panel">
    <div class="panel-title">版本管理</div>

    <div class="bar">
      <el-form inline>
        <el-form-item label="软件">
          <el-select
            v-model="softwareId"
            filterable
            clearable
            placeholder="请选择软件"
            style="width: 360px"
            @change="loadVersions"
          >
            <el-option v-for="s in softwares" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :disabled="!softwareId" @click="openCreate">新增版本</el-button>
        </el-form-item>
      </el-form>
    </div>

    <el-table :data="versions" size="small" style="width: 100%" v-loading="loading">
      <el-table-column prop="id" label="ID" width="90" />
      <el-table-column prop="version" label="版本号" width="160" />
      <el-table-column prop="published_at" label="发布时间" width="140">
        <template #default="{ row }">
          {{ row.published_at ? String(row.published_at).slice(0, 10) : '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="release_notes" label="更新说明" min-width="260" show-overflow-tooltip />
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="visible" :title="form.id ? '编辑版本' : '新增版本'" width="720px">
      <el-form :model="form" label-width="110px">
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="版本号">
              <el-input v-model="form.version" placeholder="如 1.2.3" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="发布时间">
              <el-date-picker v-model="form.published_at" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="更新说明">
          <el-input v-model="form.release_notes" type="textarea" :rows="4" placeholder="可选" />
        </el-form-item>

        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="直链下载">
              <el-input v-model="form.download_direct_text" type="textarea" :rows="3" placeholder="每行一个直链 URL" />
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
import { onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import {
  softwareList,
  softwareVersionCreate,
  softwareVersionDelete,
  softwareVersionList,
  softwareVersionUpdate,
  type ISoftwareItem,
  type ISoftwareVersionItem,
} from '@/api/software'

const route = useRoute()

const loading = ref(false)
const softwares = ref<ISoftwareItem[]>([])
const softwareId = ref<number | undefined>(undefined)
const versions = ref<ISoftwareVersionItem[]>([])

const visible = ref(false)
const form = reactive<any>({
  id: 0,
  version: '',
  release_notes: '',
  published_at: '',
  download_direct_text: '',
  download_pan_text: '',
  download_extract: '',
})

const parseLines = (raw: string) =>
  String(raw || '')
    .split('\n')
    .map((s) => s.trim())
    .filter(Boolean)

const loadSoftwares = async () => {
  const { data: res } = await softwareList({ page: 1, page_size: 200 })
  if (res.code !== 200) return
  softwares.value = Array.isArray(res.data?.list) ? res.data.list : []
}

const loadVersions = async () => {
  if (!softwareId.value) {
    versions.value = []
    return
  }
  loading.value = true
  try {
    const { data: res } = await softwareVersionList(String(softwareId.value))
    if (res.code !== 200) return
    versions.value = Array.isArray(res.data?.list) ? res.data.list : []
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  Object.assign(form, {
    id: 0,
    version: '',
    release_notes: '',
    published_at: '',
    download_direct_text: '',
    download_pan_text: '',
    download_extract: '',
  })
}

const openCreate = () => {
  resetForm()
  visible.value = true
}

const openEdit = (row: ISoftwareVersionItem) => {
  resetForm()
  Object.assign(form, row)
  form.download_direct_text = (Array.isArray(row.download_direct) ? row.download_direct : []).join('\n')
  form.download_pan_text = (Array.isArray(row.download_pan) ? row.download_pan : []).join('\n')
  form.published_at = row.published_at ? String(row.published_at).slice(0, 10) : ''
  visible.value = true
}

const save = async () => {
  if (!softwareId.value) return
  const payload: any = {
    version: form.version,
    release_notes: form.release_notes,
    published_at: form.published_at || '',
    download_direct: parseLines(form.download_direct_text),
    download_pan: parseLines(form.download_pan_text),
    download_extract: form.download_extract,
  }
  if (form.id) await softwareVersionUpdate(String(form.id), payload)
  else await softwareVersionCreate(String(softwareId.value), payload)
  visible.value = false
  await loadVersions()
}

const remove = async (row: ISoftwareVersionItem) => {
  await softwareVersionDelete(String(row.id))
  await loadVersions()
}

onMounted(async () => {
  await loadSoftwares()
  const fromQuery = String(route.query.software_id || '').trim()
  if (fromQuery) {
    const idNum = Number(fromQuery)
    if (Number.isFinite(idNum) && idNum > 0) softwareId.value = idNum
  }
  await loadVersions()
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
</style>

