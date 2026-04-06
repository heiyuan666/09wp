<template>
  <div class="form-content-inner">
    <h2 class="title">扫码登录</h2>
    <p class="subtitle">请使用 {{ runtimeConfig.siteTitle }} App 扫描下方二维码</p>

    <div class="qr-container">
      <div class="qr-placeholder">
        <div v-if="loading && !qrDataUrl" class="qr-loading">
          <el-icon class="is-loading" :size="40"><Loading /></el-icon>
          <p>正在生成二维码…</p>
        </div>
        <img v-else-if="qrDataUrl" :src="qrDataUrl" alt="扫码登录" class="qr-image" />
        <div v-else class="qr-empty">
          <el-icon :size="48" color="var(--el-text-color-placeholder)"><WarningFilled /></el-icon>
          <p>{{ loadError || '无法展示二维码' }}</p>
          <el-button type="primary" link @click="bootstrap">重试</el-button>
        </div>

        <div class="qr-mask" v-if="qrExpired">
          <p>二维码已失效</p>
          <el-button type="primary" link @click="refreshQr">点击刷新</el-button>
        </div>
      </div>

      <el-alert
        type="warning"
        :closable="false"
        show-icon
        class="role-alert"
        title="本页为管理后台登录：二维码仅支持「管理员账号」。普通用户请使用下方链接前往用户扫码登录。"
      />
      <div class="user-qr-link">
        <span class="user-qr-label">普通用户（前台账号）扫码登录：</span>
        <el-link type="primary" :underline="false" @click="goUserQrLogin">
          前往用户扫码登录页
        </el-link>
      </div>
      <p class="qr-tip">
        打开 App → 登录页 → 扫码登录，使用<strong>管理员用户名与密码</strong>（非前台用户）确认后，本页将自动登录
      </p>
      <p v-if="sidShort" class="sid-hint">会话 ID：{{ sidShort }}</p>

      <div class="back-link">
        <el-link :underline="false" @click="goBack">
          <el-icon><component :is="menuStore.iconComponents['Element:ArrowLeft']" /></el-icon>
          返回账号登录
        </el-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import QRCode from 'qrcode'
import { Loading, WarningFilled } from '@element-plus/icons-vue'
import { adminQrLoginCreate, adminQrLoginStatus } from '@/api/login'
import { runtimeConfig } from '@/config/runtimeConfig'
import type { IEmits } from '@/types/login'

const emits = defineEmits<IEmits>()
const menuStore = useMenuStore()
const router = useRouter()

const loading = ref(true)
const loadError = ref('')
const qrDataUrl = ref('')
const currentSid = ref('')
const qrExpired = ref(false)
let pollTimer: ReturnType<typeof setInterval> | null = null

const sidShort = computed(() => {
  const s = currentSid.value
  if (!s || s.length < 12) return ''
  return `${s.slice(0, 8)}…${s.slice(-4)}`
})

const stopPoll = () => {
  if (pollTimer != null) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

const pollOnce = async () => {
  if (!currentSid.value || qrExpired.value) return
  try {
    const { data: res } = await adminQrLoginStatus(currentSid.value)
    if (res.code !== 200 || !res.data) return
    const st = res.data.status
    if (st === 'expired') {
      stopPoll()
      qrExpired.value = true
      return
    }
    if (st === 'confirmed' && res.data.token) {
      stopPoll()
      localStorage.setItem('token', res.data.token)
      localStorage.removeItem('user_token')
      ElMessage.success('登录成功')
      router.push('/profile')
    }
  } catch {
    /* axios 已提示 */
  }
}

const bootstrap = async () => {
  loading.value = true
  loadError.value = ''
  qrDataUrl.value = ''
  qrExpired.value = false
  stopPoll()
  try {
    const { data: res } = await adminQrLoginCreate()
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
    loadError.value = '网络异常，请检查 API 地址与后端服务'
    console.error(e)
  } finally {
    loading.value = false
  }
}

const refreshQr = () => {
  qrExpired.value = false
  void bootstrap()
}

const goBack = () => {
  stopPoll()
  emits('goToMode', 'login')
}

/** 跳转到前台用户扫码登录（/u/login/qr），与管理端扫码会话无关 */
const goUserQrLogin = () => {
  stopPoll()
  router.push('/u/login/qr')
}

onMounted(() => {
  void bootstrap()
})

onBeforeUnmount(() => {
  stopPoll()
})
</script>

<style scoped lang="scss">
.form-content-inner {
  .title {
    font-size: 1.75rem;
    font-weight: 700;
    color: var(--el-text-color-primary);
    margin-bottom: 0.5rem;
  }

  .subtitle {
    font-size: 0.95rem;
    color: var(--el-text-color-secondary);
    margin-bottom: 2rem;
  }

  .qr-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 1.25rem 0;

    .qr-placeholder {
      position: relative;
      width: 12.5rem;
      height: 12.5rem;
      border: 1px solid var(--el-border-color-lighter);
      border-radius: 1rem;
      display: flex;
      align-items: center;
      justify-content: center;
      margin-bottom: 20px;
      background: var(--el-fill-color-blank);
      overflow: hidden;

      .qr-image {
        width: 100%;
        height: 100%;
        object-fit: contain;
      }

      .qr-loading,
      .qr-empty {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 0.5rem;
        padding: 1rem;
        text-align: center;
        font-size: 0.85rem;
        color: var(--el-text-color-secondary);
      }

      .qr-mask {
        position: absolute;
        inset: 0;
        background: rgba(255, 255, 255, 0.92);
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        border-radius: 1rem;
        p {
          margin-bottom: 0.5rem;
          font-size: 0.9rem;
          color: var(--el-text-color-primary);
        }
      }
    }

    .role-alert {
      max-width: 22rem;
      margin-bottom: 1rem;
      text-align: left;
    }

    .user-qr-link {
      display: flex;
      flex-wrap: wrap;
      align-items: center;
      justify-content: center;
      gap: 6px;
      max-width: 22rem;
      margin-bottom: 1rem;
      font-size: 0.9rem;
      text-align: center;
      line-height: 1.5;

      .user-qr-label {
        color: var(--el-text-color-secondary);
      }
    }

    .qr-tip {
      font-size: 0.9rem;
      color: var(--el-text-color-secondary);
      margin-bottom: 0.75rem;
      text-align: center;
      max-width: 22rem;
      line-height: 1.5;
    }

    .sid-hint {
      font-size: 0.75rem;
      color: var(--el-text-color-placeholder);
      margin-bottom: 1rem;
      font-family: ui-monospace, monospace;
    }

    .back-link {
      display: flex;
      justify-content: center;
      align-items: center;

      .el-link {
        font-size: 0.875rem;
        color: var(--el-text-color-secondary);
        font-weight: 500;
        transition: all 0.3s;

        &:hover {
          color: var(--el-color-primary);
          transform: translateX(-4px);
        }
      }
    }
  }
}
</style>
