<template>
  <div class="panel">
    <div class="panel-title">TG频道管理</div>
    <div class="toolbar">
      <el-button type="primary" @click="openCreate">新增频道</el-button>
      <el-button @click="syncAll">同步全部启用频道</el-button>
    </div>

    <el-card shadow="never" class="session-card">
      <template #header>MTProto 登录会话（api_id/api_hash）</template>
      <div class="session-row">
        <el-tag :type="sessionStatus.has_session ? 'success' : 'info'">
          {{ sessionStatus.has_session ? '已登录' : '未登录' }}
        </el-tag>
        <span class="session-meta">手机号：{{ sessionStatus.phone || '-' }}</span>
        <span class="session-meta">需2FA：{{ sessionStatus.need_password ? '是' : '否' }}</span>
        <el-button size="small" @click="loadSessionStatus">刷新状态</el-button>
      </div>
      <div class="session-row">
        <el-input v-model="sessionForm.phone" placeholder="+8613xxxxxxxxx" style="width: 220px" />
        <el-button @click="sendCode">发送验证码</el-button>
        <el-input v-model="sessionForm.code" placeholder="验证码" style="width: 160px" />
        <el-button @click="signInByCode">验证码登录</el-button>
        <el-input v-model="sessionForm.password" placeholder="2FA密码(如需要)" show-password style="width: 180px" />
        <el-button @click="checkPassword">提交2FA</el-button>
      </div>
    </el-card>

    <el-table :data="list" size="small">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="频道名称" min-width="160" />
      <el-table-column prop="channel_chat_id" label="Chat ID" min-width="180" />
      <el-table-column prop="proxy_url" label="代理" min-width="220">
        <template #default="{ row }">
          <span>{{ row.proxy_url || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="default_cat_id" label="默认分类ID" width="110" />
      <el-table-column prop="sync_interval" label="间隔(秒)" width="100" />
      <el-table-column prop="enabled" label="启用" width="90">
        <template #default="{ row }">
          <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '是' : '否' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="last_sync_status" label="最近状态" width="110" />
      <el-table-column label="操作" width="320">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link @click="testOne(row)">测试</el-button>
          <el-button link @click="syncOne(row)">同步</el-button>
          <el-button link type="success" @click="backfillOne(row)">回溯</el-button>
          <el-button link type="warning" @click="showError(row)">错误详情</el-button>
          <el-button link type="danger" @click="removeOne(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="visible" :title="form.id ? '编辑频道' : '新增频道'" width="680px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="频道名称">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="Bot Token">
          <el-input v-model="form.bot_token" placeholder="留空则使用系统配置的全局 Bot Token" />
        </el-form-item>
        <el-form-item label="频道 Chat ID">
          <el-input v-model="form.channel_chat_id" placeholder="-100xxxxxxxxxx" />
        </el-form-item>
        <el-form-item label="代理地址">
          <el-input v-model="form.proxy_url" placeholder="留空则使用系统配置全局代理；也可填 http:// 或 socks5://" />
        </el-form-item>
        <el-form-item label="默认分类ID">
          <el-input-number v-model="form.default_cat_id" :min="0" />
        </el-form-item>
        <el-form-item label="同步间隔(秒)">
          <el-input-number v-model="form.sync_interval" :min="30" :max="86400" />
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
        <el-descriptions-item label="频道名称">{{ errorRow.name || '-' }}</el-descriptions-item>
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
  telegramChannelBackfill,
  telegramChannelCreate,
  telegramChannelDelete,
  telegramChannelList,
  telegramSessionCheckPassword,
  telegramSessionSendCode,
  telegramSessionSignIn,
  telegramSessionStatus,
  telegramChannelSync,
  telegramChannelSyncAll,
  telegramChannelTest,
  telegramChannelUpdate,
} from '@/api/netdisk'
import { ElMessageBox } from 'element-plus'

defineOptions({ name: 'TgChannelsView' })

const list = ref<any[]>([])
const visible = ref(false)
const sessionStatus = reactive<any>({ has_api: false, has_session: false, need_password: false, phone: '' })
const sessionForm = reactive<any>({ phone: '', code: '', password: '' })
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
  bot_token: '',
  channel_chat_id: '',
  proxy_url: '',
  default_cat_id: 0,
  enabled: true,
  sync_interval: 300,
})

const load = async () => {
  const { data: res } = await telegramChannelList()
  if (res.code !== 200) return
  list.value = res.data?.list || []
}

const loadSessionStatus = async () => {
  const { data: res } = await telegramSessionStatus()
  if (res.code !== 200) return
  Object.assign(sessionStatus, res.data || {})
}

const openCreate = () => {
  Object.assign(form, {
    id: 0,
    name: '',
    bot_token: '',
    channel_chat_id: '',
    proxy_url: '',
    default_cat_id: 0,
    enabled: true,
    sync_interval: 300,
  })
  visible.value = true
}

const openEdit = (row: any) => {
  Object.assign(form, row)
  visible.value = true
}

const save = async () => {
  const payload = {
    name: form.name,
    bot_token: form.bot_token,
    channel_chat_id: form.channel_chat_id,
    proxy_url: form.proxy_url || '',
    default_cat_id: Number(form.default_cat_id || 0),
    enabled: !!form.enabled,
    sync_interval: Number(form.sync_interval || 300),
  }
  if (form.id) await telegramChannelUpdate(form.id, payload)
  else await telegramChannelCreate(payload)
  visible.value = false
  await load()
}

const removeOne = async (row: any) => {
  await telegramChannelDelete(row.id)
  await load()
}

const syncOne = async (row: any) => {
  const { data: res } = await telegramChannelSync(row.id)
  if (res.code !== 200) return
  ElMessage.success(`同步完成：新增 ${res.data?.added ?? 0}，跳过 ${res.data?.skipped ?? 0}`)
  await load()
}

const backfillOne = async (row: any) => {
  const promptResult = await ElMessageBox.prompt('请输入回溯条数（默认 2000，最大 5000）', '回溯同步', {
    inputValue: '2000',
    inputPattern: /^(?:[1-9]\d{0,3}|5000)$/,
    inputErrorMessage: '请输入 1-5000 的整数',
    confirmButtonText: '开始回溯',
    cancelButtonText: '取消',
  }).catch(() => null)
  if (!promptResult?.value) return
  const limit = Math.min(5000, Math.max(1, Number(promptResult.value) || 2000))
  const { data: res } = await telegramChannelBackfill(row.id, { limit })
  if (res.code !== 200) return
  ElMessage.success(
    `回溯完成：扫描 ${res.data?.scanned ?? 0}，新增 ${res.data?.added ?? 0}，跳过 ${res.data?.skipped ?? 0}`,
  )
  await load()
}

const testCurrent = async () => {
  const { data: res } = await telegramChannelTest({
    bot_token: String(form.bot_token || ''),
    channel_chat_id: String(form.channel_chat_id || ''),
    proxy_url: String(form.proxy_url || ''),
  })
  if (res.code !== 200) return
  ElMessage.success(res.data?.message || '测试成功')
}

const testOne = async (row: any) => {
  const { data: res } = await telegramChannelTest({
    bot_token: String(row.bot_token || ''),
    channel_chat_id: String(row.channel_chat_id || ''),
    proxy_url: String(row.proxy_url || ''),
  })
  if (res.code !== 200) return
  ElMessage.success(res.data?.message || '测试成功')
}

const syncAll = async () => {
  const { data: res } = await telegramChannelSyncAll()
  if (res.code !== 200) return
  ElMessage.success(
    `全部同步完成：频道 ${res.data?.synced ?? 0}，新增 ${res.data?.added ?? 0}，跳过 ${res.data?.skipped ?? 0}`,
  )
  await load()
}

const sendCode = async () => {
  const { data: res } = await telegramSessionSendCode({ phone: String(sessionForm.phone || '') })
  if (res.code !== 200) return
  ElMessage.success(res.data?.message || '验证码已发送')
  await loadSessionStatus()
}

const signInByCode = async () => {
  const { data: res } = await telegramSessionSignIn({ code: String(sessionForm.code || '') })
  if (res.code !== 200) return
  ElMessage.success(res.data?.message || '登录成功')
  await loadSessionStatus()
}

const checkPassword = async () => {
  const { data: res } = await telegramSessionCheckPassword({ password: String(sessionForm.password || '') })
  if (res.code !== 200) return
  ElMessage.success(res.data?.message || '2FA登录成功')
  await loadSessionStatus()
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
  await load()
  await loadSessionStatus()
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
.session-card {
  margin-bottom: 12px;
}
.session-row {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}
.session-meta {
  color: #666;
  font-size: 12px;
}
.error-text {
  white-space: pre-wrap;
  word-break: break-all;
}
</style>

