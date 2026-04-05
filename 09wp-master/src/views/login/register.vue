<template>
  <div class="form-content-inner">
    <h2 class="title">创建账号</h2>
    <p class="subtitle">加入 {{ runtimeConfig.siteTitle }}，开始您的管理之旅</p>

    <el-form :model="registerForm" label-position="top" class="register-form">
      <el-form-item>
        <el-input v-model="registerForm.username" placeholder="设置用户名" />
      </el-form-item>
      <el-form-item>
        <el-input v-model="registerForm.email" placeholder="输入电子邮箱" />
      </el-form-item>
      <el-form-item>
        <div class="captcha-row">
          <el-input v-model="registerForm.captchaCode" placeholder="输入图形验证码" />
          <button class="captcha-img" type="button" @click="refreshCaptcha" :disabled="captchaLoading">
            <span v-if="captchaLoading" class="captcha-loading">加载中...</span>
            <span v-else v-html="captchaSvg" />
          </button>
        </div>
      </el-form-item>
      <el-form-item>
        <div class="email-code-row">
          <el-input v-model="registerForm.emailCode" placeholder="输入邮箱验证码" />
          <el-button
            class="send-code-btn"
            :loading="sendingCode"
            :disabled="sendCooldown > 0"
            @click="handleSendEmailCode"
          >
            {{ sendCooldown > 0 ? `${sendCooldown}s 后重试` : '发送验证码' }}
          </el-button>
        </div>
      </el-form-item>
      <el-form-item>
        <el-input
          v-model="registerForm.password"
          type="password"
          show-password
          placeholder="设置登录密码"
        />
      </el-form-item>
      <el-form-item>
        <el-input
          v-model="registerForm.confirmPassword"
          type="password"
          show-password
          placeholder="确认您的密码"
        />
      </el-form-item>
      <el-button type="primary" class="submit-btn" :loading="loading" @click="handleRegister">
        立即注册
      </el-button>
      <div class="back-link">
        <span class="have-account">已有账号？</span>
        <el-link :underline="false" @click="emits('goToMode', 'login')">
          <el-icon><component :is="menuStore.iconComponents['Element:ArrowLeft']" /></el-icon>
          返回登录
        </el-link>
      </div>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { siteCaptcha, siteRegister, siteRegisterSendCode } from '@/api/netdisk'
import { runtimeConfig } from '@/config/runtimeConfig'
import { ElMessage } from 'element-plus'
import type { IEmits } from '@/types/login'

defineOptions({ name: 'RegisterComponent' })

const emits = defineEmits<IEmits>()
const menuStore = useMenuStore()

const registerForm = ref({
  username: '',
  email: '',
  captchaId: '',
  captchaCode: '',
  emailCode: '',
  password: '',
  confirmPassword: '',
})

const loading = ref(false)
const captchaLoading = ref(false)
const captchaSvg = ref('')
const sendingCode = ref(false)
const sendCooldown = ref(0)
let cooldownTimer: number | undefined

const refreshCaptcha = async () => {
  captchaLoading.value = true
  try {
    const { data: res } = await siteCaptcha()
    if (res.code !== 200) {
      ElMessage.error(res.message || '获取验证码失败')
      return
    }
    captchaSvg.value = res.data.svg
    registerForm.value.captchaId = res.data.captcha_id
    registerForm.value.captchaCode = ''
  } finally {
    captchaLoading.value = false
  }
}

onMounted(() => {
  refreshCaptcha()
})

onBeforeUnmount(() => {
  if (cooldownTimer) window.clearInterval(cooldownTimer)
})

const startCooldown = (seconds: number) => {
  if (cooldownTimer) window.clearInterval(cooldownTimer)
  sendCooldown.value = seconds
  cooldownTimer = window.setInterval(() => {
    sendCooldown.value -= 1
    if (sendCooldown.value <= 0) {
      sendCooldown.value = 0
      if (cooldownTimer) window.clearInterval(cooldownTimer)
      cooldownTimer = undefined
    }
  }, 1000)
}

const handleSendEmailCode = async () => {
  if (!runtimeConfig.allowRegister) {
    ElMessage.warning('当前系统已关闭注册')
    return
  }
  const email = registerForm.value.email.trim()
  if (!email) {
    ElMessage.warning('请输入邮箱')
    return
  }
  const emailReg = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailReg.test(email)) {
    ElMessage.warning('邮箱格式不正确')
    return
  }
  if (!registerForm.value.captchaId || !registerForm.value.captchaCode.trim()) {
    ElMessage.warning('请先输入图形验证码')
    return
  }

  sendingCode.value = true
  try {
    const { data: res } = await siteRegisterSendCode({
      email,
      captcha_id: registerForm.value.captchaId,
      captcha_code: registerForm.value.captchaCode.trim(),
    })
    if (res.code !== 200) {
      ElMessage.error(res.message || '发送失败')
      await refreshCaptcha()
      return
    }
    ElMessage.success('验证码已发送，请查收邮箱')
    startCooldown(60)
  } finally {
    sendingCode.value = false
  }
}

const handleRegister = async () => {
  if (!runtimeConfig.allowRegister) {
    ElMessage.warning('当前系统已关闭注册')
    return
  }

  const { username, email, password, confirmPassword, emailCode, captchaId, captchaCode } = registerForm.value

  if (!username.trim()) {
    ElMessage.warning('请输入用户名')
    return
  }
  if (!email.trim()) {
    ElMessage.warning('请输入邮箱')
    return
  }
  const emailReg = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailReg.test(email)) {
    ElMessage.warning('邮箱格式不正确')
    return
  }
  if (!password) {
    ElMessage.warning('请输入密码')
    return
  }
  if (password.length < 6) {
    ElMessage.warning('密码至少 6 位')
    return
  }
  if (password !== confirmPassword) {
    ElMessage.warning('两次密码输入不一致')
    return
  }
  if (!captchaId || !captchaCode.trim()) {
    ElMessage.warning('请输入图形验证码')
    return
  }
  if (!emailCode.trim()) {
    ElMessage.warning('请输入邮箱验证码')
    return
  }

  loading.value = true
  try {
    const { data: res } = await siteRegister({
      username: username.trim(),
      email: email.trim(),
      password,
      email_code: emailCode.trim(),
      captcha_id: captchaId,
      captcha_code: captchaCode.trim(),
    })
    if (res.code !== 200) {
      ElMessage.error(res.message || '注册失败')
      await refreshCaptcha()
      return
    }
    ElMessage.success('注册成功，请登录')
    registerForm.value = {
      username: '',
      email: '',
      captchaId: '',
      captchaCode: '',
      emailCode: '',
      password: '',
      confirmPassword: '',
    }
    await refreshCaptcha()
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

  .register-form {
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

    .captcha-row,
    .email-code-row {
      width: 100%;
      display: flex;
      align-items: stretch;
      gap: 0.75rem;
    }

    .captcha-img {
      width: 128px;
      min-width: 128px;
      border: 1px solid var(--el-border-color);
      border-radius: 0.5rem;
      background: var(--el-bg-color);
      padding: 0;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      cursor: pointer;
      overflow: hidden;
      user-select: none;

      &:disabled {
        cursor: not-allowed;
        opacity: 0.8;
      }

      :deep(svg) {
        display: block;
      }
    }

    .captcha-loading {
      font-size: 12px;
      color: var(--el-text-color-secondary);
      padding: 0 10px;
      white-space: nowrap;
    }

    .send-code-btn {
      width: 128px;
      min-width: 128px;
      border-radius: 0.5rem;
      font-weight: 600;
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
      gap: 0.5rem;

      .have-account {
        font-size: 0.875rem;
        color: var(--el-text-color-secondary);
      }

      .el-link {
        font-size: 0.9rem;
        font-weight: 600;
        transition: all 0.3s;
        color: var(--el-text-color-secondary);

        &:hover {
          color: var(--el-color-primary);
          transform: translateX(-4px);
        }
      }
    }
  }
}
</style>
