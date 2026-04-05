<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { adminStats, type IAdminStats } from '@/api/adminStats'
import { appVersion } from '@/config/appVersion'
import { runtimeConfig } from '@/config/runtimeConfig'

defineOptions({ name: 'DashboardView' })

const loading = ref(false)
const router = useRouter()
const stats = ref<IAdminStats>({
  resources_total: 0,
  resources_online: 0,
  users_total: 0,
  categories_total: 0,
  submissions_pending: 0,
  feedbacks_total: 0,
  searches_total: 0,
})

const go = (path: string) => {
  router.push(path)
}

const openClarity = () => {
  window.open('https://clarity.microsoft.com/', '_blank', 'noopener,noreferrer')
}

const goConfig = () => {
  router.push('/admin/system/config')
}

const backendVersion = ref('')

const loadBackendVersion = async () => {
  try {
    const base = (import.meta.env.VITE_API_BASE_URL || '').replace(/\/+$/, '')
    if (!base) return
    const r = await fetch(`${base}/public/version`)
    const j = await r.json()
    if (j?.code === 200 && j?.data?.version) {
      backendVersion.value = String(j.data.version)
    }
  } catch {
    /* ignore */
  }
}

const load = async () => {
  loading.value = true
  try {
    const { data: res } = await adminStats()
    if (res?.code === 200 && res.data) stats.value = res.data
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void load()
  void loadBackendVersion()
})
</script>

<template>
  <div class="das" v-loading="loading">
    <div class="das-head">
      <div>
        <div class="title">控制台</div>
        <div class="ver-line">
          前端 v{{ appVersion }}<template v-if="backendVersion"> · 后端 v{{ backendVersion }}</template>
        </div>
      </div>
      <div class="head-actions">
        <el-button
          v-if="runtimeConfig.clarityEnabled && runtimeConfig.clarityProjectId"
          type="success"
          plain
          @click="openClarity"
        >
          查看 Clarity 统计
        </el-button>
        <el-button type="primary" plain @click="load">刷新</el-button>
      </div>
    </div>

    <el-row :gutter="12">
      <el-col :xs="12" :sm="8" :md="6">
        <div class="stat-card">
          <div class="k">资源总数</div>
          <div class="v clickable" @click="go('/admin/netdisk/resources')">
            {{ stats.resources_total }}
          </div>
          <div class="s">
            在线：            <span class="clickable" @click="go('/admin/netdisk/resources')">{{ stats.resources_online }}</span>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <div class="stat-card">
          <div class="k">用户总数</div>
          <div class="v clickable" @click="go('/admin/system/user')">{{ stats.users_total }}</div>
          <div class="s">后台/前台用户合计</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <div class="stat-card">
          <div class="k">分类数量</div>
          <div class="v clickable" @click="go('/admin/netdisk/categories')">{{ stats.categories_total }}</div>
          <div class="s">用于前台导航与筛选</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <div class="stat-card warn">
          <div class="k">待审核提交</div>
          <div class="v clickable" @click="go('/admin/netdisk/submissions')">{{ stats.submissions_pending }}</div>
          <div class="s">需要管理员处理</div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="12" class="mt">
      <el-col :xs="12" :sm="8" :md="6">
        <div class="stat-card">
          <div class="k">反馈总数</div>
          <div class="v clickable" @click="go('/admin/system/feedbacks')">{{ stats.feedbacks_total }}</div>
          <div class="s">资源反馈累计</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <div class="stat-card">
          <div class="k">搜索总次数</div>
          <div class="v">{{ stats.searches_total }}</div>
          <div class="s">来自前台搜索关键词统计</div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="16" :md="12">
        <div class="stat-card clarity-card">
          <div class="clarity-head">
            <div class="k">Microsoft Clarity</div>
            <el-tag :type="runtimeConfig.clarityEnabled && runtimeConfig.clarityProjectId ? 'success' : 'info'">
              {{ runtimeConfig.clarityEnabled && runtimeConfig.clarityProjectId ? '已启用' : '未启用' }}
            </el-tag>
          </div>
          <div class="clarity-meta">Project ID: {{ runtimeConfig.clarityProjectId || '-' }}</div>
          <div class="clarity-actions">
            <el-button type="success" plain :disabled="!(runtimeConfig.clarityEnabled && runtimeConfig.clarityProjectId)" @click="openClarity">
              打开统计后台
            </el-button>
            <el-button plain @click="goConfig">前往系统配置</el-button>
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.das {
  padding: 12px;
}
.das-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}
.head-actions {
  display: flex;
  gap: 8px;
}
.title {
  font-size: 18px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}
.ver-line {
  margin-top: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.stat-card {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  padding: 12px;
  background: var(--el-bg-color);
  min-height: 92px;
}
.stat-card.warn {
  border-color: rgba(245, 108, 108, 0.35);
  background: linear-gradient(180deg, rgba(245, 108, 108, 0.08), rgba(245, 108, 108, 0.02));
}
.k {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.v {
  font-size: 26px;
  font-weight: 800;
  line-height: 1.2;
  margin-top: 6px;
  color: var(--el-text-color-primary);
}
.clickable {
  cursor: pointer;
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.clickable:hover {
  opacity: 0.85;
  transform: translateY(-1px);
}
.s {
  margin-top: 6px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.mt {
  margin-top: 12px;
}
.clarity-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.clarity-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.clarity-meta {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  word-break: break-all;
}
.clarity-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
</style>




