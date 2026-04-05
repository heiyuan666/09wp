<template>
  <div class="form-content-inner">
    <h2 class="title">找回密码</h2>
    <p class="subtitle">请输入您的邮箱地址来重置密码</p>

    <el-form v-if="step === 'request'" :model="forgotPasswordForm" label-position="top" class="forgot-password-form">
      <el-form-item>
        <el-input v-model="forgotPasswordForm.email" placeholder="请输入注册时的邮箱地址" />
      </el-form-item>
      <el-button type="primary" class="submit-btn" :loading="loading" @click="handleRequest">
        发送重置链接
      </el-button>

      <div class="hint">
        <span>演示环境：接口会直接返回重置令牌（生产应通过邮箱发送）。</span>
      </div>

      <div class="back-link">
        <el-link :underline="false" @click="emits('goToMode', 'login')">
          <el-icon><component :is="menuStore.iconComponents['Element:ArrowLeft']" /></el-icon>
          返回登录
        </el-link>
      </div>
    </el-form>

    <el-form v-else :model="resetForm" label-position="top" class="forgot-password-form">
      <div class="reset-token">
        <div class="reset-token-title">重置令牌</div>
        <div class="reset-token-value">{{ resetForm.token || '-' }}</div>
      </div>

      <el-form-item>
        <el-input v-model="resetForm.newPassword" type="password" show-password placeholder="请输入新密码（至少 6 位）" />
      </el-form-item>
      <el-form-item>
        <el-input v-model="resetForm.confirmPassword" type="password" show-password placeholder="请再次输入新密码" />
      </el-form-item>
      <el-button type="primary" class="submit-btn" :loading="loading" @click="handleReset">重置密码</el-button>

      <div class="back-link">
        <el-link :underline="false" @click="step = 'request'">
          返回上一步
        </el-link>
      </div>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import type { IEmits } from '@/types/login'
import { sitePasswordForgot, sitePasswordReset } from '@/api/netdisk'

const emits = defineEmits<IEmits>()
const menuStore = useMenuStore()
const route = useRoute()

const forgotPasswordForm = ref({
  email: '',
})

type Step = 'request' | 'reset'
const step = ref<Step>('request')
const loading = ref(false)

const resetForm = ref({
  token: '',
  newPassword: '',
  confirmPassword: '',
})

watch(
  () => route.query.token,
  (t) => {
    const token = String(t || '').trim()
    if (!token) return
    resetForm.value.token = token
    step.value = 'reset'
  },
  { immediate: true },
)

const handleRequest = async () => {
  const email = String(forgotPasswordForm.value.email || '').trim()
  if (!email) {
    ElMessage.warning('请输入邮箱')
    return
  }

  loading.value = true
  try {
    const { data: res } = await sitePasswordForgot({ email })
    if (res.code !== 200) return

    const token = String(res.data?.reset_token || '').trim()
    if (!token) {
      ElMessage.success('如果邮箱存在，将发送重置链接到你的邮箱')
      return
    }

    resetForm.value.token = token
    step.value = 'reset'
    ElMessage.success('已获取重置令牌，请设置新密码')
  } finally {
    loading.value = false
  }
}

const handleReset = async () => {
  const token = String(resetForm.value.token || '').trim()
  if (!token) {
    ElMessage.warning('重置令牌缺失，请返回上一步重新获取')
    return
  }
  if (!resetForm.value.newPassword) {
    ElMessage.warning('请输入新密码')
    return
  }
  if (resetForm.value.newPassword !== resetForm.value.confirmPassword) {
    ElMessage.warning('两次输入的密码不一致')
    return
  }

  loading.value = true
  try {
    const { data: res } = await sitePasswordReset({
      token,
      newPassword: resetForm.value.newPassword,
      confirmPassword: resetForm.value.confirmPassword,
    })
    if (res.code !== 200) return

    ElMessage.success('密码已重置，请重新登录')
    emits('goToMode', 'login')
  } finally {
    loading.value = false
  }
}
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

  .forgot-password-form {
    :deep(.el-input__wrapper),
    :deep(.el-select__wrapper) {
      padding: 0.5rem 1rem;
      border-radius: 0.5rem;
      box-shadow: 0 0 0 1px var(--el-border-color) inset;
      min-height: 2.75rem;

      &.is-focus {
        box-shadow: 0 0 0 1px var(--el-color-primary) inset;
      }
    }

    .submit-btn {
      width: 100%;
      height: 2.75rem;
      border-radius: 0.75rem;
      font-size: 1rem;
      font-weight: 600;
      margin-top: 0.9rem;
      margin-bottom: 1.5rem;
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

  .hint {
    margin-top: -8px;
    margin-bottom: 14px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .reset-token {
    margin-bottom: 14px;
    border: 1px solid var(--el-border-color-light);
    border-radius: 12px;
    padding: 12px;
    background: var(--el-fill-color-light);

    .reset-token-title {
      font-size: 12px;
      font-weight: 800;
      color: var(--el-text-color-secondary);
      margin-bottom: 8px;
    }

    .reset-token-value {
      word-break: break-all;
      font-size: 12px;
      color: var(--el-text-color-regular);
    }
  }
}
</style>
