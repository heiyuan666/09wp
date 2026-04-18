<template>
  <el-card shadow="never">
    <template #header>
      <div class="header">
        <span>清理任务日志</span>
      </div>
    </template>

    <el-form :inline="true" label-width="80px" class="filter-form">
      <el-form-item label="状态">
        <el-select v-model="filters.status" placeholder="全部" style="width: 160px">
          <el-option label="全部" value="" />
          <el-option label="成功" value="success" />
          <el-option label="失败" value="failed" />
          <el-option label="跳过" value="skipped" />
        </el-select>
      </el-form-item>
      <el-form-item label="平台">
        <el-select v-model="filters.platform" placeholder="全部平台" style="width: 160px">
          <el-option label="全部平台" value="" />
          <el-option label="夸克" value="quark" />
          <el-option label="百度" value="baidu" />
          <el-option label="UC" value="uc" />
          <el-option label="迅雷" value="xunlei" />
        </el-select>
      </el-form-item>
      <el-form-item label="资源ID">
        <el-input v-model="filters.resource_id" placeholder="可选" clearable style="width: 160px" />
      </el-form-item>
      <el-button type="primary" :loading="loading" @click="onSearch">查询</el-button>
      <el-button @click="onReset">重置</el-button>
    </el-form>

    <el-table :data="list" border style="width: 100%; margin-top: 12px" v-loading="loading">
      <el-table-column prop="id" label="ID" width="90" />
      <el-table-column prop="resource_id" label="资源ID" width="110" />
      <el-table-column prop="platform" label="平台" width="100" />
      <el-table-column prop="action" label="动作" width="180" />
      <el-table-column label="状态" width="110">
        <template #default="{ row }">
          <el-tag :type="row.status === 'success' ? 'success' : row.status === 'failed' ? 'danger' : 'info'">
            {{ row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="message" label="消息" min-width="320" show-overflow-tooltip />
      <el-table-column prop="created_at" label="时间" min-width="180" />
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
import { adminCleanupLogList, type ICleanupTaskLogItem } from '@/api/cleanupLog'

defineOptions({ name: 'SystemCleanupLogsView' })

const list = ref<ICleanupTaskLogItem[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const filters = reactive({
  status: '',
  platform: '',
  resource_id: '',
})

const load = async () => {
  loading.value = true
  try {
    const { data: res } = await adminCleanupLogList({
      page: page.value,
      page_size: pageSize.value,
      task: 'global_search_cleanup',
      status: filters.status,
      platform: filters.platform,
      resource_id: filters.resource_id,
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

const onReset = async () => {
  filters.status = ''
  filters.platform = ''
  filters.resource_id = ''
  page.value = 1
  await load()
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
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
}
.pager {
  margin-top: 14px;
  display: flex;
  justify-content: flex-end;
}
</style>

