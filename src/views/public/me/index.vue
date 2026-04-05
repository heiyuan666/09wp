<template>
  <el-card class="panel" shadow="hover">
    <div class="panel-title">用户中心</div>

    <div v-if="!token" class="empty">
      <el-empty description="你还没登录" :image-size="80" />
      <el-button type="primary" @click="goLogin">去登录</el-button>
    </div>

    <template v-else>
      <div class="profile-hero">
        <div>
          <div class="profile-name">{{ me?.username || '用户' }}</div>
          <div class="profile-email">{{ me?.email || '-' }}</div>
        </div>
        <el-button @click="logout">退出登录</el-button>
      </div>

      <div class="section-title">资源投稿中心</div>
      <el-tabs v-model="activeTab" class="submit-tabs">
        <el-tab-pane label="网盘资源" name="netdisk">
          <div class="tab-intro">适合文档、软件、影视、教程等普通网盘资源投稿。</div>
          <el-form :model="netdiskForm" label-position="top" class="submit-form">
            <el-form-item label="标题">
              <el-input v-model="netdiskForm.title" placeholder="例如：大学英语四级资料合集" />
            </el-form-item>
            <el-form-item label="链接">
              <el-input v-model="netdiskForm.link" placeholder="粘贴分享链接（支持百度 / 夸克 / 阿里 / 天翼等）" />
            </el-form-item>
            <div class="form-grid">
              <el-form-item label="分类">
                <el-select v-model="netdiskForm.category_id" placeholder="请选择分类" style="width: 100%">
                  <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="Number(c.id)" />
                </el-select>
              </el-form-item>
              <el-form-item label="提取码">
                <el-input v-model="netdiskForm.extract_code" placeholder="如有提取码请填写" />
              </el-form-item>
            </div>
            <el-form-item label="标签">
              <el-input v-model="netdiskForm.tags" placeholder="用英文逗号分隔，如：英语,四级,真题" />
            </el-form-item>
            <el-form-item label="资源说明">
              <el-input
                v-model="netdiskForm.description"
                type="textarea"
                :rows="4"
                placeholder="补充资源内容、适用对象、使用说明等"
              />
            </el-form-item>
            <div class="submit-actions">
              <el-button type="primary" :loading="submittingNetdisk" @click="submitNetdiskResource">提交审核</el-button>
              <el-button @click="resetNetdiskForm">清空</el-button>
            </div>
            <div class="submit-hint">提交后会进入审核队列，通过后会在资源站展示。</div>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="游戏资源" name="game">
          <div class="tab-intro">适合为具体游戏投稿本体、补丁、Mod、修改器或整合资源。</div>
          <el-form :model="gameForm" label-position="top" class="submit-form">
            <div class="form-grid">
              <el-form-item label="选择游戏">
                <el-select
                  v-model="gameForm.game_id"
                  filterable
                  clearable
                  placeholder="请选择游戏"
                  style="width: 100%"
                >
                  <el-option
                    v-for="g in gameOptions"
                    :key="g.id"
                    :label="`${g.title} (#${g.id})`"
                    :value="Number(g.id)"
                  />
                </el-select>
              </el-form-item>
              <el-form-item label="投稿标题">
                <el-input v-model="gameForm.title" placeholder="例如：Dota 2 夸克网盘整合包" />
              </el-form-item>
            </div>
            <el-form-item label="下载链接">
              <el-input v-model="gameForm.link" placeholder="粘贴游戏资源分享链接" />
            </el-form-item>
            <div class="form-grid">
              <el-form-item label="提取码">
                <el-input v-model="gameForm.extract_code" placeholder="没有可不填" />
              </el-form-item>
              <el-form-item label="标签">
                <el-input v-model="gameForm.tags" placeholder="如：补丁,MOD,整合包" />
              </el-form-item>
            </div>
            <el-form-item label="资源说明">
              <el-input
                v-model="gameForm.description"
                type="textarea"
                :rows="4"
                placeholder="可以写版本、适用说明、安装方法、资源类型等"
              />
            </el-form-item>
            <div class="submit-actions">
              <el-button type="primary" :loading="submittingGame" @click="submitGameResource">提交审核</el-button>
              <el-button @click="resetGameForm">清空</el-button>
            </div>
            <div class="submit-hint">审核通过后，这些资源会出现在对应游戏详情页的下载资源区。</div>
          </el-form>
        </el-tab-pane>
      </el-tabs>

      <div class="section-title">我的投稿记录</div>
      <el-tabs v-model="recordTab" class="record-tabs">
        <el-tab-pane :label="`全部 (${mySubmissions.length})`" name="all" />
        <el-tab-pane :label="`网盘资源 (${netdiskSubmissions.length})`" name="netdisk" />
        <el-tab-pane :label="`游戏资源 (${gameSubmissions.length})`" name="game" />
      </el-tabs>

      <el-table v-if="filteredSubmissions.length > 0" :data="filteredSubmissions" size="small" style="width: 100%">
        <el-table-column prop="title" label="标题" min-width="220" />
        <el-table-column label="类型" width="120">
          <template #default="{ row }">
            <el-tag :type="row.game_id ? 'success' : 'info'" size="small">
              {{ row.game_id ? '游戏资源' : '网盘资源' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag v-if="row.status === 'pending'" type="warning">待审核</el-tag>
            <el-tag v-else-if="row.status === 'approved'" type="success">已通过</el-tag>
            <el-tag v-else type="danger">已驳回</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="关联内容" min-width="180">
          <template #default="{ row }">
            <span v-if="row.game_id">
              游戏 #{{ row.game_id }}
              <el-link type="primary" @click="goGameDetail(row.game_id)">查看</el-link>
            </span>
            <span v-else>普通网盘资源</span>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="review_msg" label="备注" min-width="200" />
      </el-table>
      <el-empty v-else description="暂无提交记录" :image-size="80" />

      <div class="section-title">我的收藏</div>
      <el-table v-if="favorites.length > 0" :data="favorites" size="small" style="width: 100%">
        <el-table-column prop="title" label="标题" min-width="240">
          <template #default="{ row }">
            <el-link type="primary" @click="goDetail(row.id)">{{ row.title }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="view_count" label="点击" width="90" />
      </el-table>
      <el-empty v-else description="暂无收藏" :image-size="80" />
    </template>
  </el-card>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { siteCategories, siteFavorites, siteMe, siteMySubmissions, siteSubmissionCreate } from '@/api/netdisk'
import { publicGameList } from '@/views/public/game/api'

defineOptions({
  name: 'PublicMeView',
})

const router = useRouter()
const token = computed(() => localStorage.getItem('user_token') || '')

const activeTab = ref('netdisk')
const recordTab = ref('all')

const me = ref<any>(null)
const favorites = ref<any[]>([])
const categories = ref<Array<{ id: number | string; name: string }>>([])
const mySubmissions = ref<any[]>([])
const gameOptions = ref<Array<{ id: number; title: string }>>([])

const submittingNetdisk = ref(false)
const submittingGame = ref(false)

const netdiskForm = ref({
  title: '',
  link: '',
  category_id: undefined as number | undefined,
  extract_code: '',
  tags: '',
  description: '',
})

const gameForm = ref({
  game_id: undefined as number | undefined,
  title: '',
  link: '',
  extract_code: '',
  tags: '',
  description: '',
})

const netdiskSubmissions = computed(() => mySubmissions.value.filter((item) => !Number(item.game_id || 0)))
const gameSubmissions = computed(() => mySubmissions.value.filter((item) => Number(item.game_id || 0) > 0))
const filteredSubmissions = computed(() => {
  if (recordTab.value === 'netdisk') return netdiskSubmissions.value
  if (recordTab.value === 'game') return gameSubmissions.value
  return mySubmissions.value
})

const goLogin = () => router.push('/login')
const goDetail = (id: any) => router.push(`/r/${id}`)
const goGameDetail = (id: any) => router.push(`/games/${id}`)

const formatDateTime = (value?: string) => {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return String(value)
  const yyyy = date.getFullYear()
  const mm = String(date.getMonth() + 1).padStart(2, '0')
  const dd = String(date.getDate()).padStart(2, '0')
  const hh = String(date.getHours()).padStart(2, '0')
  const mi = String(date.getMinutes()).padStart(2, '0')
  return `${yyyy}-${mm}-${dd} ${hh}:${mi}`
}

const resetNetdiskForm = () => {
  netdiskForm.value = {
    title: '',
    link: '',
    category_id: undefined,
    extract_code: '',
    tags: '',
    description: '',
  }
}

const resetGameForm = () => {
  gameForm.value = {
    game_id: undefined,
    title: '',
    link: '',
    extract_code: '',
    tags: '',
    description: '',
  }
}

const load = async () => {
  if (!token.value) return
  const [{ data: meRes }, { data: favRes }, { data: catRes }, { data: subRes }, gameRes] = await Promise.all([
    siteMe(),
    siteFavorites(),
    siteCategories(),
    siteMySubmissions(),
    publicGameList({ page: 1, page_size: 100 }),
  ])

  if (meRes.code === 200) me.value = meRes.data
  if (favRes.code === 200) favorites.value = favRes.data || []
  if (catRes.code === 200 && Array.isArray(catRes.data)) categories.value = catRes.data
  if (subRes.code === 200 && Array.isArray(subRes.data)) mySubmissions.value = subRes.data

  if (gameRes.data.code === 200 && Array.isArray(gameRes.data.data?.list)) {
    gameOptions.value = gameRes.data.data.list.map((item: any) => ({
      id: Number(item.id),
      title: String(item.title || '未命名游戏'),
    }))
  }
}

const submitNetdiskResource = async () => {
  const title = String(netdiskForm.value.title || '').trim()
  const link = String(netdiskForm.value.link || '').trim()
  const categoryId = Number(netdiskForm.value.category_id || 0)
  if (!title) return ElMessage.warning('请填写标题')
  if (!link) return ElMessage.warning('请填写链接')
  if (!categoryId) return ElMessage.warning('请选择分类')

  submittingNetdisk.value = true
  try {
    const { data: res } = await siteSubmissionCreate({
      title,
      link,
      category_id: categoryId,
      extract_code: String(netdiskForm.value.extract_code || '').trim(),
      tags: String(netdiskForm.value.tags || '').trim(),
      description: String(netdiskForm.value.description || '').trim(),
    })
    if (res.code !== 200) return
    ElMessage.success('网盘资源提交成功，等待审核')
    resetNetdiskForm()
    await load()
    recordTab.value = 'netdisk'
  } finally {
    submittingNetdisk.value = false
  }
}

const submitGameResource = async () => {
  const title = String(gameForm.value.title || '').trim()
  const link = String(gameForm.value.link || '').trim()
  const gameId = Number(gameForm.value.game_id || 0)
  if (!gameId) return ElMessage.warning('请选择游戏')
  if (!title) return ElMessage.warning('请填写投稿标题')
  if (!link) return ElMessage.warning('请填写下载链接')

  submittingGame.value = true
  try {
    const { data: res } = await siteSubmissionCreate({
      title,
      link,
      game_id: gameId,
      extract_code: String(gameForm.value.extract_code || '').trim(),
      tags: String(gameForm.value.tags || '').trim(),
      description: String(gameForm.value.description || '').trim(),
    })
    if (res.code !== 200) return
    ElMessage.success('游戏资源提交成功，等待审核')
    resetGameForm()
    await load()
    recordTab.value = 'game'
  } finally {
    submittingGame.value = false
  }
}

const logout = () => {
  localStorage.removeItem('user_token')
  router.push('/')
}

onMounted(load)
</script>

<style scoped>
.panel {
  border-radius: 18px;
}

.panel-title {
  font-weight: 800;
  font-size: 20px;
  margin-bottom: 14px;
}

.profile-hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 18px;
  border-radius: 16px;
  background: linear-gradient(135deg, #f0fdf4, #ecfeff);
  border: 1px solid #d9f3e5;
  margin-bottom: 18px;
}

.profile-name {
  font-size: 22px;
  font-weight: 800;
  color: #0f172a;
}

.profile-email {
  margin-top: 6px;
  font-size: 14px;
  color: #64748b;
}

.section-title {
  margin-top: 18px;
  margin-bottom: 12px;
  font-size: 17px;
  font-weight: 800;
  color: #0f172a;
}

.submit-tabs,
.record-tabs {
  margin-bottom: 10px;
}

.tab-intro {
  margin-bottom: 12px;
  color: #64748b;
  font-size: 13px;
}

.submit-form {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 14px;
  padding: 14px;
  background: var(--el-bg-color);
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.submit-actions {
  display: flex;
  gap: 8px;
}

.submit-hint {
  margin-top: 8px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.empty {
  display: grid;
  place-items: center;
  gap: 10px;
}

@media (max-width: 720px) {
  .profile-hero {
    align-items: flex-start;
    flex-direction: column;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
