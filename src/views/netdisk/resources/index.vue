<template>
  <div class="panel">
    <div class="panel-title">资源管理</div>

    <div class="filters">
      <el-input v-model="query.title" placeholder="按标题搜索" clearable style="width: 260px" />
      <el-select v-model="query.category_id" placeholder="分类" clearable style="width: 200px">
        <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="String(c.id)" />
      </el-select>
      <el-select v-model="query.status" placeholder="状态" clearable style="width: 160px">
        <el-option label="显示" value="1" />
        <el-option label="隐藏" value="0" />
      </el-select>
      <el-button type="primary" @click="load">查询</el-button>
      <el-button @click="reset">重置</el-button>
      <div class="spacer" />
      <el-button @click="syncFromTelegram">同步全部 TG 频道</el-button>
      <el-button type="primary" @click="openCreate">新增资源</el-button>
      <el-button type="danger" :disabled="selectedIds.length === 0" @click="batchDelete">批量删除</el-button>
      <el-button @click="checkAllLinks">全量检测</el-button>
      <el-button @click="checkAllLinksOneByOne">全量逐条检测</el-button>
      <el-button :disabled="selectedIds.length === 0" @click="checkLinks(true)">批量检测</el-button>
      <el-button :disabled="selectedIds.length === 0" @click="checkLinks(true, undefined, true)">
        批量逐条检测
      </el-button>
      <el-button :disabled="selectedIds.length === 0" @click="batchStatus(1)">批量显示</el-button>
      <el-button :disabled="selectedIds.length === 0" @click="batchStatus(0)">批量隐藏</el-button>
      <el-button
        type="success"
        :disabled="selectedRows.length === 0"
        :loading="batchTransferLoading"
        @click="batchTransfer"
      >
        批量转存
      </el-button>
      <el-button type="warning" :loading="importTransferLoading" @click="openImportTransferDialog">
        批量导入转存
      </el-button>
      <el-button type="primary" @click="openTableImportDialog">表格导入</el-button>
      <el-button type="success" :loading="tableExportLoading" @click="openTableExportDialog">表格导出</el-button>
    </div>

    <el-table :data="list" size="small" style="width: 100%" @selection-change="onSelect">
      <el-table-column type="selection" width="48" />
      <el-table-column prop="id" label="ID" width="90" />
      <el-table-column prop="title" label="标题" min-width="220" />
      <el-table-column label="网盘" width="120">
        <template #default="{ row }">
          <span v-if="String(row.link || '').trim()" class="drive-name">
            {{ platformText(row.link)
            }}<template v-if="(row.extra_links?.length || 0) > 0"> +{{ row.extra_links.length }}</template>
          </span>
          <span v-else class="drive-empty">—</span>
        </template>
      </el-table-column>
      <el-table-column prop="category_id" label="分类" width="160">
        <template #default="{ row }">
          {{ categoryName(row.category_id) }}
        </template>
      </el-table-column>
      <el-table-column prop="view_count" label="点击" width="90" />
      <el-table-column prop="status" label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">
            {{ row.status === 1 ? '显示' : '隐藏' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="链接检测" width="120">
        <template #default="{ row }">
          <el-tag :type="row.link_valid === false ? 'danger' : row.link_valid === true ? 'success' : 'info'">
            {{ row.link_valid === false ? '失效' : row.link_valid === true ? '有效' : '未检测' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="转存状态" width="180">
        <template #default="{ row }">
          <el-tag :type="row.transfer_status === 'success' ? 'success' : row.transfer_status === 'failed' ? 'danger' : 'info'">
            {{
              row.transfer_status === 'success'
                ? '成功'
                : row.transfer_status === 'failed'
                  ? '失败'
                  : row.transfer_status === 'pending'
                    ? '进行中'
                    : '未转存'
            }}
          </el-tag>
          <div v-if="row.transfer_msg" class="transfer-msg" :title="row.transfer_msg">{{ row.transfer_msg }}</div>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="340">
        <template #default="{ row }">
          <el-button link @click="checkLinks(false, row)">检测</el-button>
          <el-button link type="warning" @click="checkLinks(false, row, true)">逐条</el-button>
          <el-button link type="success" @click="retryTransfer(row)">重试转存</el-button>
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="info" @click="openTransferLogs(row)">日志</el-button>
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

    <el-dialog v-model="visible" :title="form.id ? '编辑资源' : '新增资源'" width="720px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="标题">
          <el-input v-model="form.title" />
        </el-form-item>
        <el-form-item label="主链接">
          <el-input v-model="form.link" placeholder="主要展示的网盘分享链接" />
        </el-form-item>
        <el-form-item label="附加链接">
          <div class="extra-links-editor">
            <div v-for="(_u, i) in form.extra_links" :key="i" class="extra-link-row">
              <el-input v-model="form.extra_links[i]" placeholder="其它网盘链接（夸克/百度等）" />
              <el-button type="danger" link @click="removeExtraLink(i)">删除</el-button>
            </div>
            <el-button size="small" @click="addExtraLink">添加一条链接</el-button>
          </div>
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="form.category_id" style="width: 100%">
            <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="Number(c.id)" />
          </el-select>
        </el-form-item>
        <el-form-item label="提取码">
          <el-input v-model="form.extract_code" />
        </el-form-item>
        <el-form-item label="封面">
          <el-input v-model="form.cover" />
        </el-form-item>
        <el-form-item label="标签">
          <div class="tag-editor">
            <div v-if="formTagList.length" class="tag-list">
              <el-tag
                v-for="tag in formTagList"
                :key="tag"
                closable
                effect="light"
                @close="removeFormTag(tag)"
              >
                {{ tag }}
              </el-tag>
            </div>
            <el-input
              v-model="tagInput"
              placeholder="输入一个完整关键词后回车添加，可保留逗号"
              @keydown.enter.prevent="appendFormTags()"
              @blur="appendFormTags()"
            />
          </div>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="4" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort_order" :min="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="logsVisible"
      :title="logsResourceId ? `转存日志 - ${logsResourceId}` : '转存日志'"
      width="980px"
    >
      <div v-if="logsLoading" class="logs-loading">加载中...</div>
      <el-table v-else :data="transferLogs" size="small" style="width: 100%">
        <el-table-column prop="attempt" label="尝试" width="90" />
        <el-table-column prop="platform" label="平台" width="120" />
        <el-table-column prop="status" label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : row.status === 'failed' ? 'danger' : 'info'">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="说明" min-width="180" show-overflow-tooltip />
        <el-table-column prop="old_link" label="原链接" min-width="240" show-overflow-tooltip />
        <el-table-column prop="new_link" label="新链接" min-width="240" show-overflow-tooltip />
        <el-table-column prop="created_at" label="时间" min-width="180" show-overflow-tooltip />
      </el-table>
    </el-dialog>

    <el-dialog v-model="batchTransferVisible" title="批量转存结果" width="980px">
      <el-table :data="batchTransferResults" size="small" style="width: 100%">
        <el-table-column prop="index" label="#" width="70" />
        <el-table-column prop="platform" label="平台" width="100" />
        <el-table-column prop="link" label="原链接" min-width="220" show-overflow-tooltip />
        <el-table-column label="结果" width="100">
          <template #default="{ row }">
            <el-tag :type="row.success ? 'success' : 'danger'">{{ row.success ? '成功' : '失败' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="说明" min-width="160" show-overflow-tooltip />
        <el-table-column prop="own_share_url" label="我的网盘链接" min-width="260" show-overflow-tooltip />
      </el-table>
    </el-dialog>

    <el-dialog v-model="importTransferVisible" title="批量导入转存" width="860px">
      <el-form label-width="110px">
        <el-form-item label="目标分类">
          <el-select v-model="importTransferCategoryID" placeholder="请选择分类" style="width: 100%">
            <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="Number(c.id)" />
          </el-select>
        </el-form-item>
        <el-form-item label="链接列表">
          <el-input
            v-model="importTransferText"
            type="textarea"
            :rows="10"
            placeholder="每行一个网盘分享链接，支持夸克、百度、迅雷等，可混合粘贴。"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="importTransferVisible = false">取消</el-button>
        <el-button type="primary" :loading="importTransferLoading" @click="submitImportTransfer">
          开始转存并入库
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="tableImportVisible" title="表格导入资源（固定字段）" width="720px">
      <el-form label-width="110px">
        <el-form-item label="选择文件">
          <input
            ref="tableImportFileInput"
            class="file-input"
            type="file"
            accept=".csv,.xlsx"
            @change="onTableImportFileChange"
          />
          <div v-if="tableImportFileName" class="file-name">已选择：{{ tableImportFileName }}</div>
        </el-form-item>
        <el-form-item label="表头要求">
          <div class="hint">
            请使用系统“表格导出”生成的 CSV/XLSX 文件导回导入，字段与导出保持一致，这样可以直接还原资源管理数据。
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="tableImportVisible = false">取消</el-button>
        <el-button type="primary" :loading="tableImportLoading" :disabled="!tableImportFile" @click="submitTableImport">
          开始导入并入库
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="tableExportVisible" title="表格导出资源（导出全部记录）" width="820px">
      <el-form label-width="120px">
        <el-form-item label="导出格式">
          <el-select v-model="tableExportFormat" style="width: 200px">
            <el-option label="XLSX" value="xlsx" />
            <el-option label="CSV" value="csv" />
          </el-select>
        </el-form-item>
        <el-form-item label="导出数量上限">
          <el-input-number v-model="tableExportLimit" :min="1" :max="50000" />
        </el-form-item>
        <el-form-item label="导出说明">
          <div class="hint">导出资源管理中的全部记录到 CSV/XLSX 文件，可直接再导回系统。</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="tableExportVisible = false">取消</el-button>
        <el-button type="primary" :loading="tableExportLoading" @click="submitTableExport">
          生成导出文件
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { onMounted, reactive, ref } from 'vue'
import {
  adminCategoryList,
  adminResourceBatchDelete,
  adminResourceBatchStatus,
  adminResourceCheckLinks,
  adminResourceCreate,
  adminResourceDelete,
  adminResourceExportTable,
  adminResourceImportTable,
  adminResourcePage,
  adminResourceRetryTransfer,
  adminResourceSyncTelegram,
  adminResourceTransferLogs,
  adminResourceUpdate,
  netdiskTransferBatchByLinks,
} from '@/api/netdisk'
import { platformText } from '@/views/public/search/searchHelpers'

const categories = ref<any[]>([])
const list = ref<any[]>([])
const total = ref(0)
const visible = ref(false)
const selectedIds = ref<number[]>([])
const selectedRows = ref<any[]>([])
const batchTransferLoading = ref(false)
const batchTransferVisible = ref(false)
const batchTransferResults = ref<any[]>([])
const importTransferVisible = ref(false)
const importTransferLoading = ref(false)
const importTransferText = ref('')
const importTransferCategoryID = ref<number | undefined>(undefined)

const tableImportVisible = ref(false)
const tableImportLoading = ref(false)
const tableImportFile = ref<File | null>(null)
const tableImportFileName = ref('')
const tableImportFileInput = ref<HTMLInputElement | null>(null)

const tableExportVisible = ref(false)
const tableExportLoading = ref(false)
const tableExportFormat = ref<'xlsx' | 'csv'>('xlsx')
const tableExportLimit = ref(50000)

const logsVisible = ref(false)
const logsResourceId = ref<number | null>(null)
const logsLoading = ref(false)
const transferLogs = ref<any[]>([])

const query = reactive<any>({
  page: 1,
  page_size: 20,
  title: '',
  category_id: '',
  status: '',
})

const form = reactive<any>({
  id: '',
  title: '',
  link: '',
  extra_links: [] as string[],
  category_id: undefined,
  description: '',
  extract_code: '',
  cover: '',
  tags: '',
  sort_order: 0,
  status: 1,
})

const tagInput = ref('')
const formTagList = ref<string[]>([])

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

const syncFormTags = (value: unknown) => {
  formTagList.value = parseTagList(value)
  form.tags = formTagList.value.length ? JSON.stringify(formTagList.value) : ''
}

const appendFormTags = () => {
  const value = String(tagInput.value || '').trim()
  if (!value) return
  if (!formTagList.value.includes(value)) {
    formTagList.value.push(value)
  }
  form.tags = JSON.stringify(formTagList.value)
  tagInput.value = ''
}

const removeFormTag = (tag: string) => {
  formTagList.value = formTagList.value.filter((item) => item !== tag)
  form.tags = formTagList.value.length ? JSON.stringify(formTagList.value) : ''
}

const loadCategories = async () => {
  const { data: res } = await adminCategoryList()
  if (res.code !== 200) return
  categories.value = res.data.list || res.data || []
}

const load = async () => {
  const { data: res } = await adminResourcePage(query)
  if (res.code !== 200) return
  list.value = res.data.list || []
  total.value = res.data.total || 0
}

const reset = () => {
  Object.assign(query, { page: 1, page_size: 20, title: '', category_id: '', status: '' })
  load()
}

const categoryName = (cid: any) => {
  const c = categories.value.find((x) => String(x.id) === String(cid))
  return c?.name || '-'
}

const onSelect = (rows: any[]) => {
  selectedRows.value = rows
  selectedIds.value = rows.map((r) => Number(r.id))
}

const addExtraLink = () => {
  if (!Array.isArray(form.extra_links)) form.extra_links = []
  form.extra_links.push('')
}

const removeExtraLink = (i: number) => {
  if (!Array.isArray(form.extra_links)) return
  form.extra_links.splice(i, 1)
}

const openCreate = () => {
  Object.assign(form, {
    id: '',
    title: '',
    link: '',
    extra_links: [],
    category_id: categories.value?.[0]?.id ? Number(categories.value[0].id) : undefined,
    description: '',
    extract_code: '',
    cover: '',
    tags: '',
    sort_order: 0,
    status: 1,
  })
  tagInput.value = ''
  syncFormTags('')
  visible.value = true
}

const openEdit = (row: any) => {
  const ex = Array.isArray(row.extra_links) ? row.extra_links.map((x: unknown) => String(x || '').trim()).filter(Boolean) : []
  Object.assign(form, row, { category_id: Number(row.category_id), extra_links: ex.length ? ex : [] })
  tagInput.value = ''
  syncFormTags(row.tags)
  visible.value = true
}

const save = async () => {
  const extraClean = (Array.isArray(form.extra_links) ? form.extra_links : [])
    .map((x: string) => String(x || '').trim())
    .filter(Boolean)
  const payload = {
    title: form.title,
    link: form.link,
    extra_links: extraClean,
    category_id: Number(form.category_id),
    description: form.description,
    extract_code: form.extract_code,
    cover: form.cover,
    tags: formTagList.value.length ? JSON.stringify(formTagList.value) : '',
    sort_order: form.sort_order,
    status: form.status,
  }
  if (form.id) await adminResourceUpdate(String(form.id), payload)
  else await adminResourceCreate(payload)
  visible.value = false
  await load()
}

const remove = async (row: any) => {
  await adminResourceDelete(String(row.id))
  await load()
}

const retryTransfer = async (row: any) => {
  const { data: res } = await adminResourceRetryTransfer(String(row.id))
  if (res.code !== 200) return
  ElMessage.success('重试转存已提交')
  await load()
}

const openTransferLogs = async (row: any) => {
  logsResourceId.value = Number(row.id)
  logsVisible.value = true
  logsLoading.value = true
  try {
    const { data: res } = await adminResourceTransferLogs(String(row.id), { page: 1, page_size: 50 })
    if (res.code !== 200) return
    transferLogs.value = res.data.list || []
  } finally {
    logsLoading.value = false
  }
}

const batchDelete = async () => {
  await adminResourceBatchDelete(selectedIds.value)
  selectedIds.value = []
  await load()
}

const batchStatus = async (status: number) => {
  await adminResourceBatchStatus(selectedIds.value, status)
  selectedIds.value = []
  await load()
}

const syncFromTelegram = async () => {
  const { data: res } = await adminResourceSyncTelegram()
  if (res.code !== 200) return
  ElMessage.success(
    `TG 同步完成：频道 ${res.data?.synced ?? 0}，新增 ${res.data?.added ?? 0}，跳过 ${res.data?.skipped ?? 0}`,
  )
  await load()
}

const checkLinks = async (batch = false, row?: any, oneByOne = false) => {
  const ids = batch ? selectedIds.value : row?.id ? [Number(row.id)] : []
  const { data: res } = await adminResourceCheckLinks(ids, [], oneByOne)
  if (res.code !== 200) return
  const details = (res.data?.details || []) as any[]
  const detailText = oneByOne
    ? `；明细：${details
        .map((d: any) => `${d.status === 'valid' ? '有效' : d.status === 'invalid' ? '失效' : '未知'}:${d.link}`)
        .slice(0, 3)
        .join(' | ')}${details.length > 3 ? ' ...' : ''}`
    : ''
  ElMessage.success(
    `检测完成：共 ${res.data?.checked ?? 0}，有效 ${res.data?.valid ?? 0}，失效 ${res.data?.invalid ?? 0}，未知 ${res.data?.unknown ?? 0}${detailText}`,
  )
  if (batch) selectedIds.value = []
  await load()
}

const checkAllLinks = async () => {
  const { data: res } = await adminResourceCheckLinks([], [], false)
  if (res.code !== 200) return
  ElMessage.success(
    `全量检测完成：共 ${res.data?.checked ?? 0}，有效 ${res.data?.valid ?? 0}，失效 ${res.data?.invalid ?? 0}，未知 ${res.data?.unknown ?? 0}`,
  )
  await load()
}

const checkAllLinksOneByOne = async () => {
  const { data: res } = await adminResourceCheckLinks([], [], true)
  if (res.code !== 200) return
  ElMessage.success(
    `全量逐条检测完成：共 ${res.data?.checked ?? 0}，有效 ${res.data?.valid ?? 0}，失效 ${res.data?.invalid ?? 0}，未知 ${res.data?.unknown ?? 0}`,
  )
  await load()
}

const batchTransfer = async () => {
  if (selectedRows.value.length === 0) return
  batchTransferLoading.value = true
  try {
    const items = selectedRows.value
      .map((r) => ({
        link: String(r.link || '').trim(),
        passcode: String(r.extract_code || '').trim() || undefined,
      }))
      .filter((x) => x.link)
    if (items.length === 0) {
      ElMessage.warning('所选资源没有可用链接')
      return
    }
    const { data: res } = await netdiskTransferBatchByLinks({ items })
    if (res.code !== 200) {
      ElMessage.error(res.message || '批量转存失败')
      return
    }
    const rawResults = (res.data?.results || []) as any[]
    batchTransferResults.value = rawResults.map((r) => {
      const data = r.data || {}
      const ownShareURL = String(data.own_share_url || '').trim()
      return {
        index: Number(r.index ?? 0) + 1,
        platform: r.platform || '-',
        link: r.link || '',
        success: !!r.success,
        message: r.success ? (data.message || '转存成功') : r.message || '转存失败',
        own_share_url: ownShareURL,
      }
    })
    batchTransferVisible.value = true
    ElMessage.success(`批量转存完成：成功 ${res.data?.success ?? 0}，失败 ${res.data?.failed ?? 0}`)
    selectedIds.value = []
    selectedRows.value = []
    await load()
  } finally {
    batchTransferLoading.value = false
  }
}

const openImportTransferDialog = () => {
  importTransferCategoryID.value = categories.value?.[0]?.id ? Number(categories.value[0].id) : undefined
  importTransferText.value = ''
  importTransferVisible.value = true
}

const openTableImportDialog = () => {
  tableImportVisible.value = true
  tableImportLoading.value = false
  tableImportFile.value = null
  tableImportFileName.value = ''
  if (tableImportFileInput.value) tableImportFileInput.value.value = ''
}

const onTableImportFileChange = (e: Event) => {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0] || null
  tableImportFile.value = file
  tableImportFileName.value = file?.name || ''
}

const submitTableImport = async () => {
  if (!tableImportFile.value) return
  tableImportLoading.value = true
  try {
    const { data: res } = await adminResourceImportTable(tableImportFile.value)
    if (res.code !== 200) return
    ElMessage.success(`导入完成：新增 ${res.data?.added ?? 0}，更新 ${res.data?.updated ?? 0}，跳过 ${res.data?.skipped ?? 0}`)
    tableImportVisible.value = false
    await load()
  } finally {
    tableImportLoading.value = false
  }
}

const openTableExportDialog = () => {
  tableExportVisible.value = true
  tableExportLoading.value = false
  tableExportFormat.value = 'xlsx'
  tableExportLimit.value = 50000
}

const resolveExportDownloadURL = (link: string) => {
  const value = String(link || '').trim()
  if (!value) return ''
  if (/^https?:\/\//i.test(value)) return value

  const apiBaseURL = String(import.meta.env.VITE_API_BASE_URL || '').trim()
  if (/^https?:\/\//i.test(apiBaseURL)) {
    try {
      const backendBase = apiBaseURL.replace(/\/api\/v\d+\/?$/i, '/')
      return new URL(value, backendBase).toString()
    } catch {}
  }

  return new URL(value, window.location.origin).toString()
}

const submitTableExport = async () => {
  tableExportLoading.value = true
  try {
    const params: any = {
      format: tableExportFormat.value,
      export_all: true,
      limit: tableExportLimit.value,
    }

    Object.keys(params).forEach((k) => {
      if (params[k] === '' || params[k] === null) delete params[k]
    })

    const { data: res } = await adminResourceExportTable(params)
    if (res.code !== 200) return

    const link = String(res.data?.link || '').trim()
    const filename = String(res.data?.filename || '').trim()
    const count = Number(res.data?.count ?? 0)
    ElMessage.success(`导出完成：${count} 条，文件：${filename || '-'}，已准备下载`)
    if (link) {
      const href = resolveExportDownloadURL(link)
      const a = document.createElement('a')
      a.href = href
      a.download = filename || ''
      a.click()
    }
    tableExportVisible.value = false
    await load()
  } finally {
    tableExportLoading.value = false
  }
}

const extractLinksFromText = (input: string) => {
  const urlReg = /https?:\/\/[^\s]+/g
  const found = input.match(urlReg) || []
  const uniq: string[] = []
  const seen = new Set<string>()
  for (const raw of found) {
    const link = raw.trim().replace(/[),.;]+$/, '')
    if (!link || seen.has(link)) continue
    seen.add(link)
    uniq.push(link)
  }
  return uniq
}

const pickTitle = (row: any) => {
  const data = row?.data || {}
  const title = String(data.title || '').trim()
  if (title) return title
  const link = String(row?.link || '').trim()
  if (!link) return '未命名资源'
  try {
    const u = new URL(link)
    return `${u.hostname} 资源`
  } catch {
    return '未命名资源'
  }
}

const submitImportTransfer = async () => {
  const categoryID = Number(importTransferCategoryID.value || 0)
  if (!categoryID) {
    ElMessage.warning('请先选择目标分类')
    return
  }
  const links = extractLinksFromText(importTransferText.value)
  if (links.length === 0) {
    ElMessage.warning('未识别到可用链接')
    return
  }

  importTransferLoading.value = true
  try {
    const items = links.map((link) => ({ link }))
    const { data: res } = await netdiskTransferBatchByLinks({ items })
    if (res.code !== 200) {
      ElMessage.error(res.message || '批量转存失败')
      return
    }
    const rawResults = (res.data?.results || []) as any[]
    let inserted = 0
    let skipped = 0
    for (const row of rawResults) {
      if (!row?.success) {
        skipped += 1
        continue
      }
      const ownShareURL = String(row?.data?.own_share_url || '').trim()
      if (!ownShareURL) {
        skipped += 1
        continue
      }
      const payload = {
        title: pickTitle(row),
        link: ownShareURL,
        category_id: categoryID,
        description: '',
        extract_code: '',
        cover: '',
        tags: '',
        sort_order: 0,
        status: 1,
      }
      const createRes = await adminResourceCreate(payload)
      if (createRes.data?.code === 200) inserted += 1
      else skipped += 1
    }

    batchTransferResults.value = rawResults.map((r, idx) => ({
      index: Number(r.index ?? idx) + 1,
      platform: r.platform || '-',
      link: r.link || '',
      success: !!r.success,
      message: r.success ? (r.data?.message || '转存成功') : r.message || '转存失败',
      own_share_url: String(r.data?.own_share_url || '').trim(),
    }))
    batchTransferVisible.value = true
    importTransferVisible.value = false
    ElMessage.success(`处理完成：入库 ${inserted} 条，跳过 ${skipped} 条`)
    await load()
  } finally {
    importTransferLoading.value = false
  }
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
  font-weight: 800;
  margin-bottom: 10px;
}

.filters {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 10px;
  align-items: center;
}

.spacer {
  flex: 1;
}

.pager {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}

.transfer-msg {
  margin-top: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  max-width: 160px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-input {
  width: 100%;
}

.file-name {
  margin-top: 6px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.hint {
  font-size: 12px;
  line-height: 1.8;
  color: var(--el-text-color-secondary);
}

.tag-editor {
  width: 100%;
}

.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 10px;
}

.drive-name {
  font-size: 13px;
  line-height: 1.4;
  color: var(--el-text-color-regular);
}

.drive-empty {
  color: var(--el-text-color-placeholder);
}

.extra-links-editor {
  width: 100%;
}
.extra-link-row {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}
.extra-link-row .el-input {
  flex: 1;
}
</style>
