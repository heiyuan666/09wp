<template>
  <el-card class="panel" shadow="hover">
    <div class="panel-title">用户登录</div>
    <el-form label-position="top" class="space-y-1">
      <el-form-item label="账号">
        <el-input v-model="form.username" placeholder="用户名或邮箱" />
      </el-form-item>
      <el-form-item label="密码">
        <el-input v-model="form.password" type="password" show-password placeholder="请输入密码" />
      </el-form-item>
    </el-form>
    <div class="actions">
      <el-button type="primary" @click="submit">登录</el-button>
      <el-button @click="goRegister">去注册</el-button>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { useRouter } from 'vue-router'
import { siteLogin } from '@/api/netdisk'

const router = useRouter()
const form = reactive({ username: '', password: '' })

const submit = async () => {
  const { data: res } = await siteLogin(form)
  if (res.code !== 200) return
  localStorage.setItem('user_token', res.data.token)
  router.push('/u/me')
}

const goRegister = () => router.push('/u/register')
</script>

<style scoped>
.panel {
  max-width: 520px;
  margin: 0 auto;
  border-radius: 14px;
}
.panel-title {
  font-weight: 800;
  margin-bottom: 10px;
}
.actions {
  display: flex;
  gap: 8px;
}
</style>

