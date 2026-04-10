<template>
  <div class="panel">
    <div class="panel-title">软件分类管理</div>

    <div class="bar">
      <el-button type="primary" @click="openCreate">新增分类</el-button>
    </div>

    <el-table :data="list" size="small" style="width: 100%" v-loading="loading">
      <el-table-column prop="id" label="ID" width="90" />
      <el-table-column prop="name" label="名称" min-width="160" />
      <el-table-column prop="slug" label="Slug" min-width="180" />
      <el-table-column label="排序" width="140">
        <template #default="{ row }">
          <el-input-number
            v-model="row.sort_order"
            :min="0"
            size="small"
            @change="(val: number | undefined) => changeSort(row, val)"
          />
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180" />
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="visible" :title="form.id ? '编辑分类' : '新增分类'" width="520px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="名称">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="Slug">
          <el-input v-model="form.slug" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort_order" :min="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import {
  softwareCategoryCreate,
  softwareCategoryDelete,
  softwareCategoryList,
  softwareCategorySort,
  softwareCategoryUpdate,
  type ISoftwareCategory,
} from '@/api/software'

const loading = ref(false)
const list = ref<ISoftwareCategory[]>([])

const visible = ref(false)
const form = reactive<any>({
  id: '',
  name: '',
  slug: '',
  sort_order: 0,
  status: 1,
})

const fetchList = async () => {
  loading.value = true
  try {
    const { data: res } = await softwareCategoryList()
    if (res.code === 200 && res.data) {
      list.value = Array.isArray(res.data.list) ? res.data.list : []
    }
  } finally {
    loading.value = false
  }
}

const openCreate = () => {
  Object.assign(form, { id: '', name: '', slug: '', sort_order: 0, status: 1 })
  visible.value = true
}

const openEdit = (row: ISoftwareCategory) => {
  Object.assign(form, row)
  visible.value = true
}

const save = async () => {
  const payload = {
    name: form.name,
    slug: form.slug,
    sort_order: form.sort_order,
    status: form.status,
  }
  if (form.id) await softwareCategoryUpdate(String(form.id), payload)
  else await softwareCategoryCreate(payload)
  visible.value = false
  await fetchList()
}

const remove = async (row: ISoftwareCategory) => {
  await softwareCategoryDelete(String(row.id))
  await fetchList()
}

const changeSort = async (row: ISoftwareCategory, val: number | undefined) => {
  await softwareCategorySort(String(row.id), Number(val || 0))
}

onMounted(fetchList)
</script>

<style scoped>
.panel {
  padding: 14px;
  border-radius: 14px;
  background: #fff;
}
.panel-title {
  font-weight: 800;
  margin-bottom: 10px;
}
.bar {
  margin-bottom: 10px;
  display: flex;
  justify-content: flex-end;
}
</style>
