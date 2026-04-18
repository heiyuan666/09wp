<template>
  <div class="page-wrap">
    <el-card shadow="never" class="card-block">
      <template #header>
        <div class="header"><span>全网搜站点设置</span></div>
      </template>
      <p class="intro">
        总开关、兜底接口与清理策略在此配置；<strong>多线路聚合</strong>在下方表格中按条维护（每条可单独地址与默认网盘类型）。当存在<strong>至少一条已启用</strong>的线路时，只走表格，不再使用此处「兜底聚合地址」。
      </p>
      <el-form :model="settings" label-width="160px" class="settings-form">
        <el-form-item label="启用全网搜">
          <el-switch v-model="settings.global_search_enabled" />
          <span class="item-desc">关闭时前台全网搜接口返回「全网搜未开启」。</span>
        </el-form-item>
        <el-form-item label="全网搜链接检测">
          <el-switch v-model="settings.global_search_link_check_enabled" />
          <span class="item-desc">开启后结果会异步做失效检测并显示状态标签。</span>
        </el-form-item>
        <el-form-item label="链接清洗正则">
          <el-input
            v-model="settings.global_search_url_sanitize_regex"
            type="textarea"
            :rows="2"
            placeholder="可选。若上游把说明文字和链接混在一起，可写正则提取真实 URL；留空则使用内置规则（提取首段 http(s)/magnet）。"
          />
          <span class="item-desc">应用于搜索结果与「获取链接」转存前，减少链接字段异常导致的转存失败。</span>
        </el-form-item>
        <el-divider content-position="left">兜底（无已启用线路时）</el-divider>
        <el-form-item label="聚合接口地址">
          <el-input
            v-model="settings.global_search_api_url"
            placeholder="如 https://api.iyuns.com/api/wpysso；空则按 IYUNS 基址拼接"
          />
        </el-form-item>
        <el-form-item label="默认网盘类型">
          <el-input v-model="settings.global_search_cloud_types" placeholder="可选，逗号分隔，如 baidu,quark" />
        </el-form-item>
        <el-divider content-position="left">一键入库</el-divider>
        <el-form-item label="默认分类 ID">
          <el-input-number v-model="settings.global_search_default_category_id" :min="0" :step="1" />
          <span class="item-desc">0 表示由后端自动选首个可用分类。</span>
        </el-form-item>
        <el-form-item label="获取链接后自动转存">
          <el-switch v-model="settings.global_search_auto_transfer" />
        </el-form-item>
        <el-divider content-position="left">定时清理（全网搜来源资源）</el-divider>
        <el-form-item label="启用定时清理">
          <el-switch v-model="settings.global_search_cleanup_enabled" />
        </el-form-item>
        <el-form-item label="清理间隔(分钟)">
          <el-input-number v-model="settings.global_search_cleanup_minutes" :min="0" :step="30" />
          <span class="item-desc">大于 0 时优先生效；0 则按「保留天数」。</span>
        </el-form-item>
        <el-form-item label="保留天数">
          <el-input-number v-model="settings.global_search_cleanup_days" :min="1" :step="1" />
        </el-form-item>
        <el-form-item label="清理时删网盘文件">
          <el-switch v-model="settings.global_search_cleanup_delete_netdisk_files" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="settingsSaving" @click="saveSettings">保存站点设置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="card-block">
      <template #header>
        <div class="header"><span>接口线路（多线路）</span></div>
      </template>
      <div class="toolbar">
        <el-button type="primary" @click="openDialog()">新增线路</el-button>
      </div>

      <el-table :data="list" border style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="名称" width="180" />
        <el-table-column prop="api_url" label="接口地址" min-width="320" show-overflow-tooltip />
        <el-table-column prop="cloud_types" label="默认网盘类型" min-width="220" show-overflow-tooltip />
        <el-table-column prop="sort_order" label="排序" width="90" />
        <el-table-column label="启用" width="100">
          <template #default="{ row }">
            <el-switch
              :model-value="row.enabled"
              @change="(v: string | number | boolean) => onToggleLine(row, Boolean(v))"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" text type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button size="small" text type="danger" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>

  <el-dialog v-model="dialogVisible" :title="dialogTitle" width="560px">
    <el-form :model="form" label-width="120px">
      <el-form-item label="线路名称"><el-input v-model="form.name" placeholder="例如：IYUNS 主接口" /></el-form-item>
      <el-form-item label="接口地址"><el-input v-model="form.api_url" placeholder="https://api.iyuns.com/api/wpysso" /></el-form-item>
      <el-form-item label="默认网盘类型"><el-input v-model="form.cloud_types" placeholder="baidu,quark（可空）" /></el-form-item>
      <el-form-item label="排序"><el-input-number v-model="form.sort_order" :min="0" /></el-form-item>
      <el-form-item label="启用"><el-switch v-model="form.enabled" /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="onSave">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  globalSearchAPIList,
  globalSearchAPICreate,
  globalSearchAPIUpdate,
  globalSearchAPIDelete,
  globalSearchSettingsGet,
  globalSearchSettingsPut,
  type IGlobalSearchAPIItem,
  type IGlobalSearchSettings,
} from '@/api/globalSearchApi'

defineOptions({ name: 'SystemGlobalSearchView' })

const defaultSettings = (): IGlobalSearchSettings => ({
  global_search_enabled: false,
  global_search_link_check_enabled: false,
  global_search_api_url: '',
  global_search_cloud_types: '',
  global_search_default_category_id: 0,
  global_search_auto_transfer: true,
  global_search_cleanup_enabled: false,
  global_search_cleanup_days: 7,
  global_search_cleanup_minutes: 0,
  global_search_cleanup_delete_netdisk_files: false,
  global_search_url_sanitize_regex: '',
})

const settings = reactive<IGlobalSearchSettings>(defaultSettings())
const settingsSaving = ref(false)

const list = ref<IGlobalSearchAPIItem[]>([])
const dialogVisible = ref(false)
const dialogTitle = ref('新增线路')
const saving = ref(false)

const form = reactive<IGlobalSearchAPIItem>({
  name: '',
  api_url: '',
  cloud_types: '',
  enabled: true,
  sort_order: 0,
})

const resetForm = () => {
  form.id = undefined
  form.name = ''
  form.api_url = ''
  form.cloud_types = ''
  form.enabled = true
  form.sort_order = 0
}

const loadSettings = async () => {
  const { data: res } = await globalSearchSettingsGet()
  if (res.code !== 200 || !res.data) return
  Object.assign(settings, defaultSettings(), res.data)
}

const saveSettings = async () => {
  settingsSaving.value = true
  try {
    const { data: res } = await globalSearchSettingsPut({ ...settings })
    if (res.code !== 200) return
    ElMessage.success('站点设置已保存')
  } finally {
    settingsSaving.value = false
  }
}

const load = async () => {
  const { data: res } = await globalSearchAPIList()
  if (res.code !== 200) return
  list.value = res.data?.list || []
}

const openDialog = (row?: IGlobalSearchAPIItem) => {
  if (row) {
    dialogTitle.value = '编辑线路'
    Object.assign(form, row)
  } else {
    dialogTitle.value = '新增线路'
    resetForm()
  }
  dialogVisible.value = true
}

const onSave = async () => {
  if (!String(form.name || '').trim() || !String(form.api_url || '').trim()) {
    ElMessage.warning('名称和接口地址不能为空')
    return
  }
  saving.value = true
  try {
    if (form.id) {
      const { data: res } = await globalSearchAPIUpdate(form.id, form)
      if (res.code !== 200) return
    } else {
      const { data: res } = await globalSearchAPICreate(form)
      if (res.code !== 200) return
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    await load()
  } finally {
    saving.value = false
  }
}

const onToggleLine = async (row: IGlobalSearchAPIItem, enabled: boolean) => {
  if (!row.id) return
  const { data: res } = await globalSearchAPIUpdate(row.id, { ...row, enabled })
  if (res.code !== 200) return
  ElMessage.success(enabled ? '已启用' : '已禁用')
  await load()
}

const onDelete = async (row: IGlobalSearchAPIItem) => {
  await ElMessageBox.confirm(`确认删除「${row.name}」吗？`, '提示', { type: 'warning' })
  const { data: res } = await globalSearchAPIDelete(row.id!)
  if (res.code !== 200) return
  ElMessage.success('已删除')
  await load()
}

onMounted(async () => {
  await loadSettings()
  await load()
})
</script>

<style scoped lang="scss">
.page-wrap {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.card-block {
  .header {
    font-weight: 600;
  }
}
.intro {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  line-height: 1.6;
  margin: 0 0 16px;
}
.settings-form {
  max-width: 720px;
}
.item-desc {
  margin-left: 10px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.toolbar {
  margin-bottom: 10px;
  display: flex;
  justify-content: flex-end;
}
</style>
