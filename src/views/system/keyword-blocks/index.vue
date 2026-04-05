<template>
  <el-card shadow="never">
    <template #header>
      <div class="header">
        <span>关键词屏蔽词</span>
        <span class="sub">匹配到屏蔽词的资源标题将不会出现在前台搜索结果中</span>
      </div>
    </template>

    <div class="add-row">
      <el-input
        v-model="newKeyword"
        placeholder="请输入要屏蔽的关键词（将做模糊子串匹配）"
        clearable
        style="max-width: 520px"
      />
      <el-switch v-model="newEnabled" />
      <span class="hint">{{ newEnabled ? '启用' : '停用' }}</span>
      <el-button type="primary" :loading="saving" @click="onAdd">添加</el-button>
    </div>

    <el-table :data="list" border style="width: 100%; margin-top: 16px" v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="keyword" label="屏蔽词" min-width="280" show-overflow-tooltip />
      <el-table-column label="启用" width="120">
        <template #default="{ row }">
          <el-switch
            v-model="row.enabled"
            :active-value="true"
            :inactive-value="false"
            @change="() => onToggle(row)"
          />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button size="small" text type="danger" @click="onDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus'
import { keywordBlockCreate, keywordBlockDelete, keywordBlockList, keywordBlockUpdate, type IKeywordBlock } from '@/api/keywordBlock'

defineOptions({ name: 'Keyword-blocksView' })

const list = ref<IKeywordBlock[]>([])
const loading = ref(false)
const saving = ref(false)

const newKeyword = ref('')
const newEnabled = ref(true)

const load = async () => {
  loading.value = true
  try {
    const { data: res } = await keywordBlockList()
    if (res.code !== 200) return
    list.value = res.data?.list || []
  } finally {
    loading.value = false
  }
}

const onAdd = async () => {
  const kw = newKeyword.value.trim()
  if (!kw) {
    ElMessage.warning('请输入关键词')
    return
  }
  saving.value = true
  try {
    const { data: res } = await keywordBlockCreate({ keyword: kw, enabled: newEnabled.value })
    if (res.code !== 200) return
    ElMessage.success('添加成功')
    newKeyword.value = ''
    newEnabled.value = true
    await load()
  } finally {
    saving.value = false
  }
}

const onToggle = async (row: IKeywordBlock) => {
  // 行内切换后立刻同步到后端
  const { data: res } = await keywordBlockUpdate(Number(row.id), { enabled: row.enabled })
  if (res.code !== 200) return
  ElMessage.success('已更新')
}

const onDelete = async (row: IKeywordBlock) => {
  await ElMessageBox.confirm(`确认删除屏蔽词「${row.keyword}」吗？`, '提示', { type: 'warning' })
  const { data: res } = await keywordBlockDelete(Number(row.id))
  if (res.code !== 200) return
  ElMessage.success('已删除')
  await load()
}

onMounted(load)
</script>

<style scoped lang="scss">
.header {
  font-weight: 600;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.sub {
  font-size: 12px;
  font-weight: 400;
  color: var(--el-text-color-secondary);
}
.add-row {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 8px;
}
.hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  width: 60px;
}
</style>

