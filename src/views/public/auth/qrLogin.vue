<template>
  <el-card class="panel" shadow="hover">
    <div class="panel-title">扫码登录（用户）</div>
    <p class="hint">
      使用 App 扫描下方二维码，在手机上输入<strong>用户账号与密码</strong>确认后，本页将自动登录。
    </p>

    <div class="qr-wrap">
      <div v-if="loading && !qrDataUrl" class="center muted">生成二维码中…</div>
      <img v-else-if="qrDataUrl" :src="qrDataUrl" alt="扫码登录" class="qr-img" />
      <div v-else class="center">
        <p class="err">{{ loadError || '无法生成二维码' }}</p>
        <el-button type="primary" link @click="bootstrap">重试</el-button>
      </div>
      <div v-if="expired" class="mask">
        <p>二维码已失效</p>
        <el-button type="primary" link @click="refreshQr">刷新</el-button>
      </div>
    </div>

    <div class="actions">
      <el-button @click="router.push('/u/login')">返回账号登录</el-button>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import QRCode from 'qrcode'
import { siteQrLoginCreate, siteQrLoginStatus } from '@/api/netdisk'

const router = useRouter()
const loading = ref(true)
const loadError = ref('')
const qrDataUrl = ref('')
const currentSid = ref('')
const expired = ref(false)
let pollTimer: ReturnType<typeof setInterval> | null = null

const stopPoll = () => {
  if (pollTimer != null) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

const pollOnce = async () => {
  if (!currentSid.value || expired.value) return
  try {
    const { data: res } = await siteQrLoginStatus(currentSid.value)
    if (res.code !== 200 || !res.data) return
    const st = res.data.status
    if (st === 'expired') {
      stopPoll()
      expired.value = true
      return
    }
    if (st === 'confirmed' && res.data.token) {
      stopPoll()
      localStorage.setItem('user_token', res.data.token)
      localStorage.removeItem('token')
      ElMessage.success('登录成功')
      router.push('/u/me')
    }
  } catch {
    /* publicRequest 已提示 */
  }
}

const bootstrap = async () => {
  loading.value = true
  loadError.value = ''
  qrDataUrl.value = ''
  expired.value = false
  stopPoll()
  try {
    const { data: res } = await siteQrLoginCreate()
    if (res.code !== 200 || !res.data?.qr_payload) {
      loadError.value = (res as { message?: string }).message || '创建会话失败'
      return
    }
    currentSid.value = res.data.sid
    qrDataUrl.value = await QRCode.toDataURL(res.data.qr_payload, {
      width: 220,
      margin: 2,
      errorCorrectionLevel: 'M',
    })
    pollTimer = setInterval(pollOnce, 2000)
    void pollOnce()
  } catch (e) {
    loadError.value = '网络异常，请确认后端已启动且跨域正常'
    console.error(e)
  } finally {
    loading.value = false
  }
}

const refreshQr = () => {
  expired.value = false
  void bootstrap()
}

onMounted(() => {
  void bootstrap()
})

onBeforeUnmount(() => {
  stopPoll()
})
</script>

<style scoped>
.panel {
  max-width: 520px;
  margin: 0 auto;
  border-radius: 14px;
}
.panel-title {
  font-weight: 800;
  margin-bottom: 8px;
}
.hint {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-bottom: 16px;
  line-height: 1.5;
}
.qr-wrap {
  position: relative;
  width: 240px;
  height: 240px;
  margin: 0 auto 16px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--el-fill-color-blank);
}
.qr-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.center {
  text-align: center;
  padding: 12px;
}
.muted {
  color: var(--el-text-color-secondary);
  font-size: 14px;
}
.err {
  color: var(--el-color-danger);
  font-size: 13px;
}
.mask {
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.92);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  font-size: 14px;
}
.actions {
  display: flex;
  justify-content: center;
}
</style>
