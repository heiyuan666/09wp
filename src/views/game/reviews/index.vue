<template>
  <el-card shadow="never">
    <template #header>
      <div class="header">
        <span>游戏评论</span>
      </div>
    </template>

    <div class="toolbar">
      <el-input v-model="q.keyword" placeholder="搜索评论内容" style="width: 260px" clearable />
      <el-input v-model="q.game_id" placeholder="game_id（可选）" style="width: 160px" clearable />
      <el-select v-model="q.status" placeholder="状态" style="width: 140px" clearable>
        <el-option label="展示" value="1" />
        <el-option label="隐藏" value="0" />
      </el-select>
      <el-button type="primary" @click="load">查询</el-button>
    </div>

    <el-table :data="list" border style="width: 100%">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="game_id" label="游戏ID" width="100" />
      <el-table-column prop="user_id" label="用户ID" width="100" />
      <el-table-column prop="rating" label="评分" width="80" />
      <el-table-column prop="helpful_count" label="有帮助" width="90" />
      <el-table-column prop="unhelpful_count" label="无帮助" width="90" />
      <el-table-column prop="status" label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '展示' : '隐藏' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="content" label="内容" min-width="280" show-overflow-tooltip />
      <el-table-column prop="created_at" label="创建时间" width="180" />
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button size="small" text type="warning" @click="toggleStatus(row)">
            {{ row.status === 1 ? '隐藏' : '展示' }}
          </el-button>
          <el-button size="small" text type="danger" @click="onDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination
        background
        layout="total, prev, pager, next, sizes"
        :total="total"
        :page-size="q.page_size"
        :current-page="q.page"
        @current-change="(p:number)=>{q.page=p;load()}"
        @size-change="(s:number)=>{q.page_size=s;q.page=1;load()}"
      />
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminGameReviewDelete, adminGameReviewPage, adminGameReviewSetStatus, type IGameReview } from '@/api/gameReview'

defineOptions({ name: 'GameReviewsView' })

const q = reactive({
  keyword: '',
  game_id: '',
  status: '',
  page: 1,
  page_size: 20,
})

const list = ref<IGameReview[]>([])
const total = ref(0)

const load = async () => {
  const params: any = {
    page: q.page,
    page_size: q.page_size,
  }
  if (q.keyword) params.keyword = q.keyword
  if (q.game_id) params.game_id = q.game_id
  if (q.status !== '') params.status = q.status

  const { data: res } = await adminGameReviewPage(params)
  if (res.code !== 200) return
  list.value = res.data?.list || []
  total.value = res.data?.total || 0
}

const toggleStatus = async (row: IGameReview) => {
  const next = row.status === 1 ? 0 : 1
  const { data: res } = await adminGameReviewSetStatus(row.id, next as 0 | 1)
  if (res.code !== 200) return
  ElMessage.success('已更新')
  await load()
}

const onDelete = async (row: IGameReview) => {
  await ElMessageBox.confirm(`确认删除评论 #${row.id} 吗？`, '提示', { type: 'warning' })
  const { data: res } = await adminGameReviewDelete(row.id)
  if (res.code !== 200) return
  ElMessage.success('已删除')
  await load()
}

onMounted(load)
</script>

<style scoped lang="scss">
.header {
  font-weight: 600;
}
.toolbar {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  align-items: center;
  margin-bottom: 12px;
}
.pager {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}
</style>

