<template>
  <div class="panel">
    <div class="panel-title">游戏分类管理</div>
    <div class="bar">
      <el-button type="primary" @click="openCreate">新增分类</el-button>
    </div>

    <el-table :data="list" style="width: 100%">
      <el-table-column prop="id" label="ID" width="90" />
      <el-table-column prop="name" label="分类名称" min-width="160" />
      <el-table-column prop="slug" label="分类标识" min-width="180" />
      <el-table-column prop="description" label="分类描述" min-width="240" show-overflow-tooltip />
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="visible" :title="form.id ? '编辑分类' : '新增分类'" width="520px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="分类名称">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="分类标识">
          <el-input v-model="form.slug" />
        </el-form-item>
        <el-form-item label="分类描述">
          <el-input v-model="form.description" type="textarea" :rows="3" />
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
import { gameCategoryCreate, gameCategoryDelete, gameCategoryList, gameCategoryUpdate } from '@/api/game'

const list = ref<any[]>([])
const visible = ref(false)
const form = reactive<any>({
  id: 0,
  name: '',
  slug: '',
  description: '',
})

const load = async () => {
  const { data: res } = await gameCategoryList()
  if (res.code !== 200) return
  list.value = res.data || []
}

const openCreate = () => {
  Object.assign(form, { id: 0, name: '', slug: '', description: '' })
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
    description: form.description,
  }
  if (form.id) await gameCategoryUpdate(form.id, payload)
  else await gameCategoryCreate(payload)
  visible.value = false
  await load()
}

const remove = async (row: any) => {
  await gameCategoryDelete(row.id)
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

