<template>
  <el-card shadow="never">
    <template #header>
      <div class="header">
        <span>软件库站点设置</span>
      </div>
    </template>

    <el-form :model="form" label-width="120px" class="config-form">
      <el-form-item label="网站标题">
        <el-input v-model="form.site_title" placeholder="例如：XX 软件库" />
      </el-form-item>
      <el-form-item label="Logo URL">
        <el-input v-model="form.logo_url" placeholder="https://... 或上传封面后填入 /public/covers/..." />
      </el-form-item>
      <el-form-item label="Favicon URL">
        <el-input v-model="form.favicon_url" placeholder="https://... 或 .ico 地址" />
      </el-form-item>
      <el-form-item label="SEO 关键词">
        <el-input v-model="form.seo_keywords" placeholder="例如：软件,下载,工具" />
      </el-form-item>
      <el-form-item label="SEO 描述">
        <el-input v-model="form.seo_description" type="textarea" :rows="3" />
      </el-form-item>

      <el-form-item>
        <el-button type="primary" :loading="saving" @click="save">保存配置</el-button>
        <el-button @click="load">重新加载</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import {
  getSoftwareSiteConfig,
  updateSoftwareSiteConfig,
  type ISoftwareSiteConfig,
} from '@/api/softwareSiteConfig'

defineOptions({ name: 'SoftwareSiteSettingsView' })

const saving = ref(false)
const form = reactive<ISoftwareSiteConfig>({
  site_title: '',
  logo_url: '',
  favicon_url: '',
  seo_keywords: '',
  seo_description: '',
})

const load = async () => {
  const { data: res } = await getSoftwareSiteConfig()
  if (res.code !== 200) return
  Object.assign(form, res.data || {})
}

const save = async () => {
  saving.value = true
  try {
    const { data: res } = await updateSoftwareSiteConfig(form)
    if (res.code !== 200) return
    ElMessage.success('保存成功')
    await load()
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped lang="scss">
.header {
  font-weight: 600;
}
</style>
