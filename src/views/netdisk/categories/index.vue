<template>
  <div class="panel">
    <div class="panel-title">分类管理</div>

    <div class="bar">
      <el-button type="primary" @click="openCreate">新增分类</el-button>
    </div>

    <el-table :data="list" size="small" style="width: 100%">
      <el-table-column prop="id" label="ID" width="90" />
      <el-table-column prop="name" label="名称" min-width="160" />
      <el-table-column prop="slug" label="Slug" min-width="160" />
      <el-table-column prop="sort_order" label="排序" width="90" />
      <el-table-column prop="status" label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">
            {{ row.status === 1 ? '显示' : '隐藏' }}
          </el-tag>
        </template>
      </el-table-column>
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
import { onMounted, reactive, ref } from 'vue'
import {
  adminCategoryCreate,
  adminCategoryDelete,
  adminCategoryList,
  adminCategoryUpdate,
} from '@/api/netdisk'

const list = ref<any[]>([])
const visible = ref(false)
const form = reactive<any>({
  id: '',
  name: '',
  slug: '',
  sort_order: 0,
  status: 1,
})

const load = async () => {
  const { data: res } = await adminCategoryList()
  if (res.code !== 200) return
  list.value = res.data.list || res.data || []
}

const openCreate = () => {
  Object.assign(form, { id: '', name: '', slug: '', sort_order: 0, status: 1 })
  visible.value = true
}
const openEdit = (row: any) => {
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
  if (form.id) await adminCategoryUpdate(String(form.id), payload)
  else await adminCategoryCreate(payload)
  visible.value = false
  await load()
}

const remove = async (row: any) => {
  await adminCategoryDelete(String(row.id))
  await load()
}

onMounted(load)
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

