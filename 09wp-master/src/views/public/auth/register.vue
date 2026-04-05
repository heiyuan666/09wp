<template>
  <el-card class="panel" shadow="hover">
    <div class="panel-title">用户注册</div>
    <el-form label-position="top" class="space-y-1">
      <el-form-item label="用户名">
        <el-input v-model="form.username" />
      </el-form-item>
      <el-form-item label="邮箱">
        <el-input v-model="form.email" />
      </el-form-item>
      <el-form-item label="密码">
        <el-input v-model="form.password" type="password" show-password />
      </el-form-item>
    </el-form>
    <div class="actions">
      <el-button type="primary" @click="submit">注册</el-button>
      <el-button @click="goLogin">去登录</el-button>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { useRouter } from 'vue-router'
import { siteRegister } from '@/api/netdisk'

const router = useRouter()
const form = reactive({ username: '', email: '', password: '' })

const submit = async () => {
  const { data: res } = await siteRegister(form)
  if (res.code !== 200) return
  router.push('/u/login')
}

const goLogin = () => router.push('/u/login')
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

