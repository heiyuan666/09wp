<template>
  <div class="profile-page">
    <section class="hero-card">
      <div class="hero-left">
        <div class="hero-copy">
          <span class="eyebrow">ACCOUNT CENTER</span>
          <h1>{{ displayName }}</h1>
          <p>维护管理员账号的登录名与联系邮箱，修改后会立即作用于后台登录。</p>
          <div class="meta-row">
            <span class="meta-pill">ID {{ userStore.userInfo?.id || '-' }}</span>
            <span class="meta-pill">账号 {{ profileForm.username || '-' }}</span>
            <span class="meta-pill">邮箱 {{ profileForm.email || '未填写' }}</span>
          </div>
        </div>
      </div>
      <div class="hero-side">
        <div class="hero-stat">
          <span>当前状态</span>
          <strong>{{ userStore.userInfo?.status === 'active' ? '启用' : '停用' }}</strong>
        </div>
        <div class="hero-stat">
          <span>最近更新</span>
          <strong>{{ userStore.userInfo?.updateTime || '-' }}</strong>
        </div>
      </div>
    </section>

    <div class="content-grid">
      <section class="panel-card">
        <div class="panel-head">
          <div>
            <h2>基本资料</h2>
            <p>这里仅维护管理员账号当前支持保存的资料字段。</p>
          </div>
          <el-button type="primary" :loading="savingProfile" @click="saveProfile">保存资料</el-button>
        </div>

        <el-form ref="profileFormRef" :model="profileForm" :rules="profileRules" label-position="top">
          <div class="form-grid two-cols">
            <el-form-item label="用户名" prop="username">
              <el-input v-model.trim="profileForm.username" placeholder="请输入用户名" maxlength="50" />
            </el-form-item>
            <el-form-item label="邮箱" prop="email">
              <el-input v-model.trim="profileForm.email" placeholder="请输入邮箱" maxlength="100" />
            </el-form-item>
          </div>
        </el-form>
      </section>

      <aside class="side-stack">
        <section class="panel-card compact-card">
          <div class="panel-head compact-head">
            <div>
              <h2>修改密码</h2>
              <p>定期更新登录密码，保护你的账号安全。</p>
            </div>
          </div>

          <el-form ref="passwordFormRef" :model="passwordForm" :rules="passwordRules" label-position="top">
            <el-form-item label="当前密码" prop="oldPassword">
              <el-input v-model.trim="passwordForm.oldPassword" type="password" show-password placeholder="请输入当前密码" />
            </el-form-item>
            <el-form-item label="新密码" prop="newPassword">
              <el-input v-model.trim="passwordForm.newPassword" type="password" show-password placeholder="至少 6 位" />
            </el-form-item>
            <el-form-item label="确认密码" prop="confirmPassword">
              <el-input
                v-model.trim="passwordForm.confirmPassword"
                type="password"
                show-password
                placeholder="请再次输入新密码"
              />
            </el-form-item>
            <el-button class="full-btn" type="danger" :loading="savingPassword" @click="savePassword">更新密码</el-button>
          </el-form>
        </section>

        <section class="panel-card compact-card info-card">
          <div class="panel-head compact-head">
            <div>
              <h2>账号信息</h2>
              <p>这里展示你的账号基础信息和最近状态。</p>
            </div>
          </div>
          <div class="info-list">
            <div class="info-item"><span>用户 ID</span><strong>{{ userStore.userInfo?.id || '-' }}</strong></div>
            <div class="info-item"><span>角色</span><strong>{{ userStore.userRoleName }}</strong></div>
            <div class="info-item"><span>创建时间</span><strong>{{ userStore.userInfo?.createTime || '-' }}</strong></div>
            <div class="info-item"><span>更新时间</span><strong>{{ userStore.userInfo?.updateTime || '-' }}</strong></div>
          </div>
        </section>
      </aside>
    </div>

  </div>
</template>

<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'

const userStore = useUserStore()
const profileFormRef = useTemplateRef<FormInstance>('profileFormRef')
const passwordFormRef = useTemplateRef<FormInstance>('passwordFormRef')

const savingProfile = ref(false)
const savingPassword = ref(false)

const profileForm = reactive({
  username: '',
  email: '',
})

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

const syncProfileForm = () => {
  const info = userStore.userInfo
  profileForm.username = String(info?.username || '')
  profileForm.email = String(info?.email || '')
}

const displayName = computed(() => profileForm.username || '未命名用户')

const validateUsername = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
  const v = String(value || '').trim()
  if (!v) return callback(new Error('请输入用户名'))
  if (v.length < 3) return callback(new Error('用户名至少 3 位'))
  callback()
}

const validateEmail = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
  const v = String(value || '').trim()
  if (!v) return callback()
  const ok = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v)
  if (!ok) return callback(new Error('请输入正确的邮箱地址'))
  callback()
}

const validateNewPassword = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
  const v = String(value || '').trim()
  if (!v) return callback(new Error('请输入新密码'))
  if (v.length < 6) return callback(new Error('密码至少 6 位'))
  callback()
}

const validateConfirmPassword = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
  if (!String(value || '').trim()) return callback(new Error('请再次输入密码'))
  if (value !== passwordForm.newPassword) return callback(new Error('两次输入的密码不一致'))
  callback()
}

const profileRules: FormRules = {
  username: [{ required: true, validator: validateUsername, trigger: 'blur' }],
  email: [{ validator: validateEmail, trigger: 'blur' }],
}

const passwordRules: FormRules = {
  oldPassword: [{ required: true, message: '请输入当前密码', trigger: 'blur' }],
  newPassword: [{ required: true, validator: validateNewPassword, trigger: 'blur' }],
  confirmPassword: [{ required: true, validator: validateConfirmPassword, trigger: 'blur' }],
}

const saveProfile = async () => {
  await profileFormRef.value?.validate()
  savingProfile.value = true
  try {
    await userStore.updateUserProfile({ ...profileForm })
    await userStore.getUserInfo()
    syncProfileForm()
  } finally {
    savingProfile.value = false
  }
}

const savePassword = async () => {
  await passwordFormRef.value?.validate()
  savingPassword.value = true
  try {
    await userStore.updatePassword({ ...passwordForm })
  } finally {
    savingPassword.value = false
  }
}

watch(
  () => userStore.userInfo,
  () => syncProfileForm(),
  { immediate: true },
)

onMounted(async () => {
  if (!userStore.userInfo) {
    await userStore.getUserInfo()
    syncProfileForm()
  }
})
</script>

<style scoped lang="scss">
.profile-page {
  display: grid;
  gap: 22px;
  padding: 24px;
}

.hero-card,
.panel-card {
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 24px;
  background:
    radial-gradient(circle at top right, rgba(14, 165, 233, 0.12), transparent 30%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(247, 249, 252, 0.98));
  box-shadow: 0 18px 48px rgba(15, 23, 42, 0.07);
}

.hero-card {
  display: flex;
  align-items: stretch;
  justify-content: space-between;
  gap: 24px;
  padding: 28px;
}

.hero-left {
  display: flex;
  gap: 20px;
  flex: 1;
}

.avatar-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.avatar-btn {
  width: 100%;
}

.hero-copy {
  flex: 1;
}

.eyebrow {
  display: inline-flex;
  align-items: center;
  padding: 6px 12px;
  border-radius: 999px;
  background: rgba(14, 165, 233, 0.1);
  color: #0284c7;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.hero-copy h1 {
  margin: 12px 0 10px;
  font-size: 32px;
  line-height: 1.15;
  color: #0f172a;
}

.hero-copy p,
.panel-head p {
  margin: 0;
  color: #64748b;
  line-height: 1.75;
}

.meta-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 18px;
}

.meta-pill {
  padding: 8px 14px;
  border-radius: 999px;
  background: #eef6ff;
  color: #1e3a8a;
  font-size: 13px;
  font-weight: 600;
}

.hero-side {
  width: 240px;
  display: grid;
  gap: 12px;
}

.hero-stat {
  display: grid;
  gap: 8px;
  padding: 18px;
  border-radius: 18px;
  background: #0f172a;
  color: #e2e8f0;
}

.hero-stat span {
  color: rgba(226, 232, 240, 0.7);
  font-size: 12px;
}

.hero-stat strong {
  font-size: 18px;
  word-break: break-word;
}

.content-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.65fr) minmax(320px, 0.9fr);
  gap: 22px;
}

.side-stack {
  display: grid;
  gap: 22px;
  align-content: start;
}

.panel-card {
  padding: 24px;
}

.panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}

.compact-head {
  margin-bottom: 14px;
}

.panel-head h2 {
  margin: 0 0 6px;
  font-size: 22px;
  color: #0f172a;
}

.form-grid {
  display: grid;
  gap: 16px;
}

.form-grid.two-cols {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.full-btn {
  width: 100%;
  margin-top: 6px;
}

.info-list {
  display: grid;
  gap: 12px;
}

.info-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 16px;
  border-radius: 16px;
  background: #f8fafc;
  border: 1px solid rgba(148, 163, 184, 0.16);
}

.info-item span {
  color: #64748b;
}

.info-item strong {
  color: #0f172a;
  text-align: right;
}

@media (max-width: 1024px) {
  .content-grid {
    grid-template-columns: 1fr;
  }

  .hero-card {
    flex-direction: column;
  }

  .hero-side {
    width: 100%;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 720px) {
  .profile-page {
    padding: 16px;
  }

  .hero-left,
  .form-grid.two-cols,
  .hero-side {
    grid-template-columns: 1fr;
    flex-direction: column;
  }

  .panel-head,
  .meta-row {
    align-items: stretch;
  }

  .panel-head {
    flex-direction: column;
  }
}
</style>
