<template>
  <div class="form-content-inner">
    <h2 class="title">欢迎回来</h2>
    <p class="subtitle">请输入您的账号信息登录系统</p>

    <!-- 登录表单 -->
    <el-form
      ref="loginFormRef"
      :model="loginForm"
      :rules="loginRules"
      label-position="top"
      class="login-form"
      @keyup.enter="handleLogin"
    >
      <el-form-item prop="username">
        <el-input v-model="loginForm.username" placeholder="请输入用户名/邮箱" />
      </el-form-item>

      <el-form-item prop="password">
        <el-input
          v-model="loginForm.password"
          type="password"
          show-password
          placeholder="请输入密码"
        />
      </el-form-item>

      <div class="form-options">
        <el-link type="primary" :underline="false" @click="emits('goToMode', 'forgot')"
          >忘记密码？</el-link
        >
      </div>

      <el-button type="primary" class="submit-btn" :loading="loading" @click="handleLogin">
        登录
      </el-button>
    </el-form>

    <!-- 其他登录方式 -->
    <div class="divider">
      <el-divider>
        <span class="divider-text">或使用以下方式登录</span>
      </el-divider>
    </div>

    <div class="social-login">
      <el-button class="social-btn" @click="emits('goToMode', 'qr')">
        <template #icon>
          <el-icon>
            <component :is="menuStore.iconComponents['Element:FullScreen']" />
          </el-icon>
        </template>
        扫码登录
      </el-button>
    </div>

    <p class="register-link">
      <span>还没有账号？</span>
      <el-link type="primary" :underline="false" @click="emits('goToMode', 'register')"
        >立即注册</el-link
      >
    </p>
  </div>
</template>

<script setup lang="ts">
import { login } from '@/api/login'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import type { ILoginMode } from '@/types/login'

interface IEmits {
  (e: 'goToMode', mode: ILoginMode): void
}

const emits = defineEmits<IEmits>()

const router = useRouter()
const menuStore = useMenuStore()
const loginFormRef = useTemplateRef<FormInstance>('loginFormRef')
const loading = ref(false)

const loginForm = ref({
  username: '',
  password: '',
})

// 登录
const handleLogin = async () => {
  await loginFormRef.value?.validate()
  loading.value = true
  try {
    const { data: res } = await login({
      username: loginForm.value.username,
      password: loginForm.value.password,
    } as any)
    if ((res as any)?.code !== 200) return
    // 后台管理员 token
    localStorage.setItem('token', (res as any).data.token)
    // 避免混用
    localStorage.removeItem('user_token')
    ElMessage.success('登录成功')
    router.push('/profile')
  } finally {
    loading.value = false
  }
}

const loginRules = reactive<FormRules>({
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
})

// 公网用户登录：不再使用管理员角色预设与登录日志
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
    margin-bottom: 1.7rem;
  }

  .login-form {
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
    .form-options {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 1.5rem;
    }

    .submit-btn {
      width: 100%;
      height: 2.75rem;
      border-radius: 0.75rem;
      font-size: 1rem;
      font-weight: 600;
      margin-bottom: 1rem;
      letter-spacing: 0.5rem;
    }
  }

  .divider {
    margin-bottom: 2rem;
    .divider-text {
      font-size: 0.75rem;
      color: var(--el-text-color-placeholder);
    }
  }

  .social-login {
    display: flex;
    justify-content: center;
    margin-bottom: 1.5rem;
    .social-btn {
      flex: 1;
      height: 2.75rem;
      border-radius: 8px;

      .social-icon {
        width: 18px;
        height: 18px;
      }
    }
  }

  .register-link {
    display: flex;
    justify-content: center;
    align-items: center;
    font-size: 0.875rem;
    color: var(--el-text-color-secondary);
    .el-link {
      margin-left: 0.5rem;
      font-weight: 600;
    }
  }
}
</style>
