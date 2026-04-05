<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  haokaCategories,
  haokaListProducts,
  haokaQueryProducts,
  haokaSync,
  haokaUpsertFromExternal,
  type IHaokaCategory,
  type IHaokaExternalProduct,
} from '@/api/haoka'

defineOptions({ name: 'HaokaView' })

const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)

const categories = ref<IHaokaCategory[]>([])

const filters = reactive({
  category_id: '',
  operator: '',
  flag: '', // "true" | "false" | ''
})

const credential = reactive({
  user_id: '',
  secret: '',
})

const dialogVisible = ref(false)
const dialogProductID = ref('')
const dialogLoading = ref(false)
const queryResults = ref<IHaokaExternalProduct[]>([])
const selectedProduct = ref<IHaokaExternalProduct | null>(null)

const loadCategories = async () => {
  const { data: res } = await haokaCategories()
  if (res.code !== 200) {
    ElMessage.error(res.message || '分类查询失败')
    return
  }
  categories.value = res.data || []
}

const loadList = async () => {
  loading.value = true
  try {
    const { data: res } = await haokaListProducts({
      category_id: filters.category_id || undefined,
      operator: filters.operator || undefined,
      flag: filters.flag || undefined,
    })
    if (res.code !== 200) {
      ElMessage.error(res.message || '列表查询失败')
      return
    }
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } finally {
    loading.value = false
  }
}

const syncAll = async () => {
  if (!credential.user_id || !credential.secret) {
    ElMessage.warning('请先填写 user_id 和 secret')
    return
  }
  loading.value = true
  try {
    const { data: res } = await haokaSync({
      user_id: credential.user_id,
      secret: credential.secret,
      product_id: '',
    })
    if (res.code !== 200) {
      ElMessage.error(res.message || '同步失败')
      return
    }
    ElMessage.success(`同步完成：total=${res.data?.total ?? 0}`)
    await loadList()
  } finally {
    loading.value = false
  }
}

const openAddDialog = () => {
  dialogVisible.value = true
  dialogProductID.value = ''
  queryResults.value = []
  selectedProduct.value = null
}

const queryDialogProducts = async () => {
  if (!credential.user_id || !credential.secret) {
    ElMessage.warning('请先填写 user_id 和 secret')
    return
  }
  dialogLoading.value = true
  try {
    const { data: res } = await haokaQueryProducts({
      user_id: credential.user_id,
      secret: credential.secret,
      product_id: dialogProductID.value || '',
    })
    if (res.code !== 200) {
      ElMessage.error(res.message || '查询失败')
      return
    }
    queryResults.value = res.data?.list || []
    selectedProduct.value = queryResults.value[0] || null
  } finally {
    dialogLoading.value = false
  }
}

const saveSelected = async () => {
  if (!selectedProduct.value) {
    ElMessage.warning('请先在列表中选择一个号卡套餐')
    return
  }
  dialogLoading.value = true
  try {
    const { data: res } = await haokaUpsertFromExternal(selectedProduct.value)
    if (res.code !== 200) {
      ElMessage.error(res.message || '保存失败')
      return
    }
    ElMessage.success('号卡已保存')
    dialogVisible.value = false
    await loadList()
  } finally {
    dialogLoading.value = false
  }
}

const onDialogRowClick = (row: IHaokaExternalProduct) => {
  selectedProduct.value = row
}

onMounted(async () => {
  await loadCategories()
  await loadList()
})
</script>

<template>
  <div class="panel">
    <el-card shadow="never">
      <template #header>
        <div class="header">
          <span>号卡管理</span>
        </div>
      </template>

      <el-form label-width="90px" class="cred-form">
        <el-form-item label="代理 user_id">
          <el-input v-model="credential.user_id" placeholder="172号卡登录账号" style="max-width: 420px" />
        </el-form-item>
        <el-form-item label="接口 secret">
          <el-input v-model="credential.secret" placeholder="接口秘钥" type="password" show-password style="max-width: 420px" />
        </el-form-item>
      </el-form>

      <div class="bar">
        <el-button type="primary" :loading="loading" @click="syncAll">同步数据</el-button>
        <el-button @click="openAddDialog">添加号卡</el-button>
      </div>

      <el-divider />

      <div class="filters">
        <el-select v-model="filters.category_id" placeholder="分类（电信/移动/联通）" clearable style="width: 220px">
          <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="String(c.id)" />
        </el-select>

        <el-select v-model="filters.flag" placeholder="上架状态" clearable style="width: 160px" class="ml">
          <el-option label="上架(true)" value="true" />
          <el-option label="已下架(false)" value="false" />
        </el-select>

        <el-button type="primary" class="ml" @click="loadList">查询</el-button>
        <el-button @click="() => Object.assign(filters, { category_id: '', operator: '', flag: '' })">重置</el-button>
      </div>

      <el-table :data="list" border style="width: 100%; margin-top: 12px" v-loading="loading">
        <el-table-column prop="product_id" label="ProductID" width="120" />
        <el-table-column prop="product_name" label="套餐名称" min-width="220" />
        <el-table-column prop="category_name" label="分类" width="140" />
        <el-table-column prop="operator" label="运营商" width="140" />
        <el-table-column prop="flag" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="row.flag ? 'success' : 'info'">{{ row.flag ? '上架' : '已下架' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="主图" width="120">
          <template #default="{ row }">
            <el-image v-if="row.main_pic" :src="row.main_pic" :preview-src-list="[row.main_pic]" style="width: 60px; height: 40px" fit="cover" />
          </template>
        </el-table-column>
        <el-table-column prop="area" label="归属地" width="140" />
        <el-table-column prop="price_time" label="优惠时间" width="160" />
        <el-table-column prop="age1" label="年龄(最小)" width="120" />
        <el-table-column prop="age2" label="年龄(最大)" width="120" />
      </el-table>
    </el-card>
  </div>

  <el-dialog v-model="dialogVisible" title="添加号卡（来自外部上架查询接口）" width="980px" :close-on-click-modal="false">
    <el-form label-width="90px">
      <el-form-item label="ProductID(可空)" style="margin-bottom: 0">
        <el-input v-model="dialogProductID" placeholder="不填返回所有上架商品" style="max-width: 420px" />
      </el-form-item>
    </el-form>

    <div class="dialog-bar">
      <el-button type="primary" :loading="dialogLoading" @click="queryDialogProducts">查询上架商品</el-button>
      <el-button @click="saveSelected" :loading="dialogLoading" type="success">保存为号卡</el-button>
    </div>

    <el-table
      :data="queryResults"
      border
      style="width: 100%; margin-top: 12px"
      v-loading="dialogLoading"
      @row-click="onDialogRowClick"
    >
      <el-table-column prop="productID" label="ProductID" width="120" />
      <el-table-column prop="productName" label="套餐名称" min-width="220" />
      <el-table-column prop="operator" label="运营商" width="140" />
      <el-table-column prop="flag" label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="row.flag ? 'success' : 'info'">{{ row.flag ? '上架' : '已下架' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="主图" width="120">
        <template #default="{ row }">
          <el-image v-if="row.mainPic" :src="row.mainPic" style="width: 60px; height: 40px" fit="cover" />
        </template>
      </el-table-column>
      <el-table-column prop="area" label="归属地" width="140" />
    </el-table>
  </el-dialog>
</template>

<style scoped lang="scss">
.panel {
  padding: 0;
}
.header {
  font-weight: 700;
}
.cred-form {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 24px;
}
.bar {
  margin-top: 10px;
  display: flex;
  gap: 12px;
}
.filters {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.ml {
  margin-left: 8px;
}
.dialog-bar {
  margin-top: 10px;
  display: flex;
  gap: 12px;
}
</style>

