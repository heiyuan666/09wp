<template>
  <div class="panel">
    <div class="panel-title">RSS订阅抓取</div>
    <div class="toolbar">
      <el-button type="primary" @click="openCreate">新增订阅</el-button>
      <el-button @click="syncAll">同步全部启用订阅</el-button>
    </div>

    <el-table :data="list" size="small">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="订阅名称" min-width="160" />
      <el-table-column prop="feed_url" label="RSS地址" min-width="360" show-overflow-tooltip />
      <el-table-column prop="default_cat_id" label="默认分类ID" width="110" />
      <el-table-column prop="max_items" label="单次抓取" width="110" />
      <el-table-column prop="sync_interval" label="间隔(秒)" width="110" />
      <el-table-column prop="enabled" label="启用" width="90">
        <template #default="{ row }">
          <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '是' : '否' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="last_sync_status" label="最近状态" width="120" />
      <el-table-column prop="last_sync_at" label="最近同步" min-width="170" />
      <el-table-column label="操作" width="260">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link @click="testOne(row)">测试</el-button>
          <el-button link @click="syncOne(row)">同步</el-button>
          <el-button link type="warning" @click="showError(row)">错误详情</el-button>
          <el-button link type="danger" @click="removeOne(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="visible" :title="form.id ? '编辑RSS订阅' : '新增RSS订阅'" width="720px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="订阅名称">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="RSS地址">
          <el-input v-model="form.feed_url" placeholder="https://rsshub.rssforever.com/telegram/channel/Quark_Movies" />
        </el-form-item>
        <el-form-item label="默认分类">
          <el-select v-model="form.default_cat_id" style="width: 100%" clearable>
            <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="Number(c.id)" />
          </el-select>
        </el-form-item>
        <el-form-item label="抓取间隔(秒)">
          <el-input-number v-model="form.sync_interval" :min="60" :max="86400" />
        </el-form-item>
        <el-form-item label="单次抓取条数">
          <el-input-number v-model="form.max_items" :min="1" :max="200" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="visible = false">取消</el-button>
        <el-button @click="testCurrent">测试连接</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="errorVisible" title="错误详情" width="640px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="订阅名称">{{ errorRow.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="最近状态">{{ errorRow.last_sync_status || '-' }}</el-descriptions-item>
        <el-descriptions-item label="最近同步时间">{{ errorRow.last_sync_at || '-' }}</el-descriptions-item>
        <el-descriptions-item label="错误信息">
          <div class="error-text">{{ errorRow.last_sync_msg || '暂无错误信息' }}</div>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button type="primary" @click="errorVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import {
  adminCategoryList,
  rssSubscriptionCreate,
  rssSubscriptionDelete,
  rssSubscriptionList,
  rssSubscriptionSync,
  rssSubscriptionSyncAll,
  rssSubscriptionTest,
  rssSubscriptionUpdate,
} from '@/api/netdisk'

defineOptions({ name: 'RssSubscriptionsView' })

const categories = ref<any[]>([])
const list = ref<any[]>([])
const visible = ref(false)
const errorVisible = ref(false)
const errorRow = reactive<any>({
  name: '',
  last_sync_status: '',
  last_sync_at: '',
  last_sync_msg: '',
})

const form = reactive<any>({
  id: 0,
  name: '',
  feed_url: '',
  default_cat_id: undefined,
  enabled: true,
  sync_interval: 1800,
  max_items: 50,
})

const loadCategories = async () => {
  const { data: res } = await adminCategoryList()
  if (res.code !== 200) return
  categories.value = res.data.list || res.data || []
}

const load = async () => {
  const { data: res } = await rssSubscriptionList()
  if (res.code !== 200) return
  list.value = res.data?.list || []
}

const openCreate = () => {
  Object.assign(form, {
    id: 0,
    name: '',
    feed_url: '',
    default_cat_id: categories.value?.[0]?.id ? Number(categories.value[0].id) : undefined,
    enabled: true,
    sync_interval: 1800,
    max_items: 50,
  })
  visible.value = true
}

const openEdit = (row: any) => {
  Object.assign(form, {
    ...row,
    default_cat_id: row.default_cat_id ? Number(row.default_cat_id) : undefined,
  })
  visible.value = true
}

const save = async () => {
  const payload = {
    name: String(form.name || '').trim(),
    feed_url: String(form.feed_url || '').trim(),
    default_cat_id: Number(form.default_cat_id || 0),
    enabled: !!form.enabled,
    sync_interval: Number(form.sync_interval || 1800),
    max_items: Number(form.max_items || 50),
  }
  if (form.id) await rssSubscriptionUpdate(form.id, payload)
  else await rssSubscriptionCreate(payload)
  visible.value = false
  await load()
}

const removeOne = async (row: any) => {
  await rssSubscriptionDelete(row.id)
  await load()
}

const testCurrent = async () => {
  const { data: res } = await rssSubscriptionTest({ feed_url: String(form.feed_url || '').trim() })
  if (res.code !== 200) return
  ElMessage.success(res.data?.message || 'RSS测试成功')
}

const testOne = async (row: any) => {
  const { data: res } = await rssSubscriptionTest({ feed_url: String(row.feed_url || '').trim() })
  if (res.code !== 200) return
  ElMessage.success(res.data?.message || 'RSS测试成功')
}

const syncOne = async (row: any) => {
  const { data: res } = await rssSubscriptionSync(row.id)
  if (res.code !== 200) return
  ElMessage.success(`同步完成：新增 ${res.data?.added ?? 0}，跳过 ${res.data?.skipped ?? 0}`)
  await load()
}

const syncAll = async () => {
  const { data: res } = await rssSubscriptionSyncAll()
  if (res.code !== 200) return
  ElMessage.success(
    `全部同步完成：订阅 ${res.data?.synced ?? 0}，新增 ${res.data?.added ?? 0}，跳过 ${res.data?.skipped ?? 0}`,
  )
  await load()
}

const showError = (row: any) => {
  Object.assign(errorRow, {
    name: row.name || '',
    last_sync_status: row.last_sync_status || '',
    last_sync_at: row.last_sync_at || '',
    last_sync_msg: row.last_sync_msg || '',
  })
  errorVisible.value = true
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
.toolbar {
  display: flex;
  gap: 10px;
  margin-bottom: 10px;
}
.error-text {
  white-space: pre-wrap;
  word-break: break-all;
}
</style>

