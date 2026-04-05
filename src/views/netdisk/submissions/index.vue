<template>
  <div class="panel">
    <div class="panel-title">用户投稿审核</div>

    <div class="filters">
      <el-input v-model="query.q" placeholder="按标题 / 链接 / 标签搜索" clearable style="width: 280px" />
      <el-select v-model="query.status" placeholder="审核状态" clearable style="width: 160px">
        <el-option label="待审核" value="pending" />
        <el-option label="已通过" value="approved" />
        <el-option label="已驳回" value="rejected" />
      </el-select>
      <el-input v-model="query.user_id" placeholder="用户 ID" clearable style="width: 140px" />
      <el-button type="warning" plain @click="onlyPending">只看待审核</el-button>
      <el-button type="primary" @click="load">查询</el-button>
      <el-button @click="reset">重置</el-button>
    </div>

    <el-table :data="list" size="small" style="width: 100%">
      <el-table-column prop="id" label="ID" width="90" />
      <el-table-column prop="user_id" label="用户ID" width="100" />
      <el-table-column prop="title" label="标题" min-width="220" show-overflow-tooltip />
      <el-table-column label="游戏归属" width="140">
        <template #default="{ row }">
          <el-tag v-if="row.game_id" type="success">游戏 #{{ row.game_id }}</el-tag>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column prop="link" label="链接" min-width="260" show-overflow-tooltip />
      <el-table-column prop="category_id" label="站点分类" width="160">
        <template #default="{ row }">
          {{ categoryName(row.category_id) }}
        </template>
      </el-table-column>
      <el-table-column prop="extract_code" label="提取码" width="100" />
      <el-table-column prop="status" label="状态" width="110">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="review_msg" label="驳回原因" min-width="200" show-overflow-tooltip />
      <el-table-column prop="created_at" label="提交时间" width="180" />
      <el-table-column label="操作" width="220">
        <template #default="{ row }">
          <el-button link type="success" :disabled="row.status !== 'pending'" @click="approve(row)">审核通过</el-button>
          <el-button link type="danger" :disabled="row.status !== 'pending'" @click="openReject(row)">驳回</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.page_size"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        :total="total"
        @change="load"
      />
    </div>

    <el-dialog v-model="rejectVisible" title="驳回原因" width="520px">
      <el-input v-model="rejectReason" type="textarea" :rows="4" placeholder="请输入驳回原因（可选）" />
      <template #footer>
        <el-button @click="rejectVisible = false">取消</el-button>
        <el-button type="danger" @click="doReject">确认驳回</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { adminCategoryList, adminSubmissionApprove, adminSubmissionList, adminSubmissionReject } from '@/api/netdisk'

defineOptions({ name: 'NetdiskSubmissionsView' })

const categories = ref<any[]>([])
const list = ref<any[]>([])
const total = ref(0)

const query = reactive({
  page: 1,
  page_size: 20,
  status: 'pending',
  q: '',
  user_id: '',
})

const rejectVisible = ref(false)
const rejectReason = ref('')
const rejectRowID = ref<number>(0)

const loadCategories = async () => {
  const { data: res } = await adminCategoryList()
  if (res.code !== 200) return
  categories.value = res.data.list || res.data || []
}

const load = async () => {
  const { data: res } = await adminSubmissionList(query)
  if (res.code !== 200) return
  list.value = res.data.list || []
  total.value = res.data.total || 0
}

const reset = () => {
  Object.assign(query, { page: 1, page_size: 20, status: 'pending', q: '', user_id: '' })
  load()
}

const onlyPending = () => {
  query.page = 1
  query.status = 'pending'
  load()
}

const categoryName = (cid: any) => {
  if (!cid) return '-'
  const c = categories.value.find((x) => String(x.id) === String(cid))
  return c?.name || '-'
}

const statusText = (status: string) => {
  if (status === 'pending') return '待审核'
  if (status === 'approved') return '已通过'
  if (status === 'rejected') return '已驳回'
  return status || '-'
}

const statusType = (status: string) => {
  if (status === 'pending') return 'warning'
  if (status === 'approved') return 'success'
  if (status === 'rejected') return 'danger'
  return 'info'
}

const approve = async (row: any) => {
  const { data: res } = await adminSubmissionApprove(row.id)
  if (res.code !== 200) return
  const resourceID = res.data?.resource_id ? `站点资源 #${res.data.resource_id}` : ''
  const gameResourceID = res.data?.game_resource_id ? `详情资源 #${res.data.game_resource_id}` : ''
  const message = [resourceID, gameResourceID].filter(Boolean).join(' / ')
  ElMessage.success(message ? `审核通过，已写入 ${message}` : '审核通过')
  await load()
}

const openReject = (row: any) => {
  rejectRowID.value = Number(row.id)
  rejectReason.value = ''
  rejectVisible.value = true
}

const doReject = async () => {
  if (!rejectRowID.value) return
  const { data: res } = await adminSubmissionReject(rejectRowID.value, { reason: rejectReason.value.trim() })
  if (res.code !== 200) return
  rejectVisible.value = false
  ElMessage.success('已驳回')
  await load()
}

onMounted(async () => {
  await loadCategories()
  await load()
})
</script>

<style scoped>
.panel {
  padding: 14px;
  border-radius: 14px;
  background: #fff;
}

.panel-title {
  margin-bottom: 10px;
  font-weight: 800;
}

.filters {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  align-items: center;
  margin-bottom: 10px;
}

.pager {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}
</style>
