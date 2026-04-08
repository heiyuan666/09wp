<template>
  <el-card shadow="never">
    <template #header>
      <div class="header">
        <span>游戏站点设置</span>
      </div>
    </template>

    <el-form :model="form" label-width="120px" class="config-form">
      <el-form-item label="网站标题">
        <el-input v-model="form.site_title" placeholder="例如：XX 游戏资源站" />
      </el-form-item>
      <el-form-item label="Logo URL">
        <el-input v-model="form.logo_url" placeholder="https://..." />
      </el-form-item>
      <el-form-item label="Favicon URL">
        <el-input v-model="form.favicon_url" placeholder="https://..." />
      </el-form-item>
      <el-form-item label="SEO 关键词">
        <el-input v-model="form.seo_keywords" placeholder="例如：游戏,下载,资源" />
      </el-form-item>
      <el-form-item label="SEO 描述">
        <el-input v-model="form.seo_description" type="textarea" :rows="3" />
      </el-form-item>

      <el-form-item>
        <el-button type="primary" :loading="saving" @click="save">保存配置</el-button>
        <el-button @click="load">加载配置</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { getGameSiteConfig, updateGameSiteConfig, type IGameSiteConfig } from '@/api/gameSiteConfig'

defineOptions({ name: 'GameSiteSettingsView' })

const saving = ref(false)
const form = reactive<IGameSiteConfig>({
  site_title: '',
  logo_url: '',
  favicon_url: '',
  seo_keywords: '',
  seo_description: '',
})

const load = async () => {
  const { data: res } = await getGameSiteConfig()
  if (res.code !== 200) return
  Object.assign(form, res.data || {})
}

const save = async () => {
  saving.value = true
  try {
    const { data: res } = await updateGameSiteConfig(form)
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

