<template>
  <el-card shadow="never">
    <template #header>
      <div class="header">
        <span>反馈管理</span>
      </div>
    </template>

    <el-form :inline="true" label-width="80px" class="filter-form">
      <el-form-item label="状态">
        <el-select v-model="filters.status" placeholder="全部" style="width: 160px">
          <el-option label="全部" value="" />
          <el-option label="待处理" value="pending" />
          <el-option label="已处理" value="processed" />
        </el-select>
      </el-form-item>

      <el-form-item label="类型">
        <el-select v-model="filters.type" placeholder="全部类型" style="width: 220px">
          <el-option label="全部类型" value="" />
          <el-option v-for="option in typeOptions" :key="option.value" :label="option.label" :value="option.value" />
        </el-select>
      </el-form-item>

      <el-button type="primary" :loading="loading" @click="onSearch">查询</el-button>
      <el-button @click="reset">重置</el-button>
    </el-form>

    <div class="quick-filters">
      <span class="quick-title">快捷筛选</span>
      <el-button
        v-for="option in typeOptions"
        :key="option.value"
        size="small"
        :type="filters.type === option.value ? 'primary' : 'default'"
        @click="setTypeQuick(option.value)"
      >
        {{ option.label }}
      </el-button>
    </div>

    <el-table :data="list" border style="width: 100%; margin-top: 12px" v-loading="loading">
      <el-table-column prop="id" label="ID" width="90" />
      <el-table-column prop="resource_id" label="资源ID" width="120" />
      <el-table-column label="类型" width="160">
        <template #default="{ row }">
          <el-tag size="small">{{ feedbackTypeLabel(row.type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="content" label="内容" min-width="320" show-overflow-tooltip />
      <el-table-column prop="contact" label="联系方式" min-width="180" show-overflow-tooltip />
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="row.status === 'processed' ? 'success' : 'warning'" size="small">
            {{ row.status === 'processed' ? '已处理' : '待处理' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="提交时间" min-width="170" />
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <div class="action-cell">
            <el-button link type="primary" @click="openResource(row.resource_id)">查看资源</el-button>
            <el-button
              v-if="row.status !== 'processed'"
              link
              type="success"
              :loading="updatingId === row.id"
              @click="updateStatus(row.id, 'processed')"
            >
              标记已处理
            </el-button>
            <el-button
              v-else
              link
              type="warning"
              :loading="updatingId === row.id"
              @click="updateStatus(row.id, 'pending')"
            >
              标记待处理
            </el-button>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination
        background
        layout="total, prev, pager, next"
        :total="total"
        :page-size="pageSize"
        v-model:current-page="page"
        @current-change="load"
      />
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { feedbackAdminList, feedbackAdminUpdateStatus, type IResourceFeedback } from '@/api/feedback'

defineOptions({ name: 'FeedbacksView' })

const list = ref<IResourceFeedback[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const updatingId = ref<number | null>(null)

const filters = reactive({
  status: '',
  type: '',
})

const typeOptions = [
  { label: '失效举报', value: 'link_invalid' },
  { label: '密码错误', value: 'password_error' },
  { label: '内容异常', value: 'content_error' },
  { label: '举报反馈', value: 'report_feedback' },
  { label: '其他', value: 'other' },
]

const typeLabelMap: Record<string, string> = {
  link_invalid: '失效举报',
  password_error: '密码错误',
  content_error: '内容异常',
  report_feedback: '举报反馈',
  other: '其他',
}

const feedbackTypeLabel = (type: string) => typeLabelMap[type] || type || '-'

const load = async () => {
  loading.value = true
  try {
    const { data: res } = await feedbackAdminList({
      page: page.value,
      page_size: pageSize.value,
      status: filters.status,
      type: filters.type,
    })
    if (res.code !== 200) return
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (error: any) {
    ElMessage.error(error?.message || '查询失败')
  } finally {
    loading.value = false
  }
}

const onSearch = async () => {
  page.value = 1
  await load()
}

const reset = async () => {
  filters.status = ''
  filters.type = ''
  page.value = 1
  await load()
}

const setTypeQuick = async (type: string) => {
  filters.type = type
  page.value = 1
  await load()
}

const updateStatus = async (id: number, status: 'pending' | 'processed') => {
  updatingId.value = id
  try {
    const { data: res } = await feedbackAdminUpdateStatus(id, status)
    if (res.code !== 200) return
    ElMessage.success(status === 'processed' ? '已标记为已处理' : '已标记为待处理')
    await load()
  } catch (error: any) {
    ElMessage.error(error?.message || '更新状态失败')
  } finally {
    updatingId.value = null
  }
}

const openResource = (resourceId: number) => {
  if (!resourceId) return
  window.open(`/r/${resourceId}`, '_blank')
}

onMounted(load)
</script>

<style scoped lang="scss">
.header {
  font-weight: 600;
}

.filter-form {
  margin-top: 10px;
  margin-bottom: 10px;
  gap: 10px;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
}

.quick-filters {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin: 4px 0 10px;
}

.quick-title {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.action-cell {
  display: flex;
  gap: 8px;
  align-items: center;
}

.pager {
  margin-top: 14px;
  display: flex;
  justify-content: flex-end;
}
</style>
