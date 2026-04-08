<template>
  <el-card shadow="never">
    <template #header>
      <div class="header">
        <span>游戏导航菜单管理</span>
      </div>
    </template>

    <el-tabs v-model="activePosition" @tab-change="load">
      <el-tab-pane label="顶部导航" name="top_nav" />
      <el-tab-pane label="首页按钮" name="home_promo" />
    </el-tabs>

    <div class="toolbar">
      <el-button type="primary" @click="openDialog()">新增</el-button>
    </div>

    <el-table :data="list" border style="width: 100%">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="title" label="标题" min-width="160" />
      <el-table-column prop="path" label="链接" min-width="220" show-overflow-tooltip />
      <el-table-column prop="position" label="位置" width="100">
        <template #default="{ row }">
          <el-tag type="info">
            {{ row.position === 'home_promo' ? '首页按钮' : '顶部导航' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="sort_order" label="排序" width="80" />
      <el-table-column prop="visible" label="显示" width="90">
        <template #default="{ row }">
          <el-switch v-model="row.visible" disabled />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button size="small" text type="primary" @click="openDialog(row)">编辑</el-button>
          <el-button size="small" text type="danger" @click="onDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>

  <el-dialog v-model="dialogVisible" :title="dialogTitle" width="480px">
    <el-form :model="form" label-width="100px">
      <el-form-item label="标题">
        <el-input v-model="form.title" placeholder="导航名称" />
      </el-form-item>
      <el-form-item label="链接地址">
        <el-input v-model="form.path" placeholder="/ 或 /category 或 https://..." />
      </el-form-item>
      <el-form-item label="位置">
        <el-select v-model="form.position" style="width: 220px">
          <el-option label="顶部导航" value="top_nav" />
          <el-option label="首页按钮" value="home_promo" />
        </el-select>
      </el-form-item>
      <el-form-item label="排序">
        <el-input-number v-model="form.sort_order" :min="0" />
      </el-form-item>
      <el-form-item label="是否显示">
        <el-switch v-model="form.visible" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="onSave">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  gameNavMenuPage,
  gameNavMenuCreate,
  gameNavMenuUpdate,
  gameNavMenuDelete,
  type IGameNavMenu,
} from '@/api/gameNavMenu'

defineOptions({ name: 'GameNavMenuSettingsView' })

const activePosition = ref<'top_nav' | 'home_promo'>('top_nav')
const list = ref<IGameNavMenu[]>([])

const dialogVisible = ref(false)
const dialogTitle = ref('新增菜单')
const saving = ref(false)

const form = reactive<IGameNavMenu>({
  title: '',
  path: '',
  position: 'top_nav',
  sort_order: 0,
  visible: true,
})

const resetForm = () => {
  form.id = undefined
  form.title = ''
  form.path = ''
  form.position = activePosition.value
  form.sort_order = 0
  form.visible = true
}

const load = async () => {
  const { data: res } = await gameNavMenuPage({ position: activePosition.value })
  if (res.code !== 200) return
  list.value = res.data?.list || []
}

const openDialog = (row?: IGameNavMenu) => {
  if (row) {
    dialogTitle.value = '编辑菜单'
    Object.assign(form, row)
  } else {
    dialogTitle.value = '新增菜单'
    resetForm()
  }
  dialogVisible.value = true
}

const onSave = async () => {
  if (!form.title || !form.title.trim()) {
    ElMessage.warning('标题不能为空')
    return
  }
  saving.value = true
  try {
    if (form.id) {
      const { data: res } = await gameNavMenuUpdate(form.id, form)
      if (res.code !== 200) return
    } else {
      const { data: res } = await gameNavMenuCreate(form)
      if (res.code !== 200) return
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    await load()
  } finally {
    saving.value = false
  }
}

const onDelete = async (row: IGameNavMenu) => {
  await ElMessageBox.confirm(`确认删除「${row.title}」吗？`, '提示', { type: 'warning' })
  const { data: res } = await gameNavMenuDelete(row.id!)
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
  margin: 10px 0;
  display: flex;
  justify-content: flex-end;
}
</style>

