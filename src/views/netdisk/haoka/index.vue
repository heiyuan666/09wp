<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getSystemConfig, updateSystemConfig, type ISystemConfig } from '@/api/systemConfig'
import {
  haokaCategories,
  haokaListProducts,
  haokaQueryProducts,
  haokaSync,
  haokaUpsertFromExternal,
  haokaProductDetail,
  haokaSetProductFlag,
  haokaUpdateProduct,
  haokaCreateProduct,
  type IHaokaCategory,
  type IHaokaExternalProduct,
  type IHaokaProductDetail,
  type IHaokaSkuItem,
} from '@/api/haoka'

defineOptions({ name: 'HaokaView' })

const loading = ref(false)
const list = ref<any[]>([])
const categories = ref<IHaokaCategory[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const filters = reactive({
  category_id: '',
  operator: '',
  flag: '', // "true" | "false" | ''
})

const credential = reactive({
  user_id: '',
  secret: '',
})

const haokaSyncEnabled = ref(false)
const haokaSyncInterval = ref(3600)
const savingHaokaConfig = ref(false)

// 添加号卡（来自外部查询）
const addDialogVisible = ref(false)
const addDialogProductID = ref('')
const addDialogLoading = ref(false)
const addQueryResults = ref<IHaokaExternalProduct[]>([])
const selectedProduct = ref<IHaokaExternalProduct | null>(null)

// 编辑/新增号卡
const editDialogVisible = ref(false)
const editLoading = ref(false)
const editProductID = ref<number>(0)
const editForm = reactive({
  product_id: 0,
  category_id: 0,
  operator: '',
  product_name: '',
  main_pic: '',
  area: '',
  disable_area: '',
  little_picture: '',
  net_addr: '',
  flag: true,
  number_sel: 0,
  back_money_type: '',
  taocan: '',
  rule: '',
  age1: 18,
  age2: 60,
  price_time: '',
  skus: [] as IHaokaSkuItem[],
})

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
      page: page.value,
      page_size: pageSize.value,
    })
    if (res.code !== 200) {
      ElMessage.error(res.message || '列表查询失败')
      return
    }
    list.value = res.data?.list || []
    total.value = Number(res.data?.total || 0)
  } finally {
    loading.value = false
  }
}

watch(
  () => [filters.category_id, filters.operator, filters.flag],
  () => {
    page.value = 1
  },
)

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
  addDialogVisible.value = true
  addDialogProductID.value = ''
  addQueryResults.value = []
  selectedProduct.value = null
}

const queryAddDialogProducts = async () => {
  if (!credential.user_id || !credential.secret) {
    ElMessage.warning('请先填写 user_id 和 secret')
    return
  }
  addDialogLoading.value = true
  try {
    const { data: res } = await haokaQueryProducts({
      user_id: credential.user_id,
      secret: credential.secret,
      product_id: addDialogProductID.value || '',
    })
    if (res.code !== 200) {
      ElMessage.error(res.message || '查询失败')
      return
    }
    addQueryResults.value = res.data?.list || []
    selectedProduct.value = addQueryResults.value[0] || null
  } finally {
    addDialogLoading.value = false
  }
}

const saveSelectedAsLocal = async () => {
  if (!selectedProduct.value) {
    ElMessage.warning('请先选择一个号卡套餐')
    return
  }
  addDialogLoading.value = true
  try {
    const { data: res } = await haokaUpsertFromExternal(selectedProduct.value)
    if (res.code !== 200) {
      ElMessage.error(res.message || '保存失败')
      return
    }
    ElMessage.success('号卡已保存')
    addDialogVisible.value = false
    await loadList()
  } finally {
    addDialogLoading.value = false
  }
}

const setShelf = async (id: number, flag: boolean) => {
  try {
    const { data: res } = await haokaSetProductFlag(id, { flag })
    if (res.code !== 200) {
      ElMessage.error(res.message || '更新失败')
      return
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '更新失败')
  } finally {
    await loadList()
  }
}

const resetEditForm = () => {
  editProductID.value = 0
  editForm.product_id = 0
  editForm.category_id = 0
  editForm.operator = ''
  editForm.product_name = ''
  editForm.main_pic = ''
  editForm.area = ''
  editForm.disable_area = ''
  editForm.little_picture = ''
  editForm.net_addr = ''
  editForm.flag = true
  editForm.number_sel = 0
  editForm.back_money_type = ''
  editForm.taocan = ''
  editForm.rule = ''
  editForm.age1 = 18
  editForm.age2 = 60
  editForm.price_time = ''
  editForm.skus = []
}

const openEditDialog = async (row: any) => {
  resetEditForm()
  editProductID.value = Number(row.id || 0)
  editLoading.value = true
  try {
    const { data: res } = await haokaProductDetail(editProductID.value)
    if (res.code !== 200) {
      ElMessage.error(res.message || '获取详情失败')
      return
    }
    const detail = res.data as {
      product: IHaokaProductDetail
      skus: IHaokaSkuItem[]
    }
    editForm.product_id = detail.product.product_id
    editForm.category_id = detail.product.category_id
    editForm.operator = detail.product.operator || ''
    editForm.product_name = detail.product.product_name
    editForm.main_pic = detail.product.main_pic || ''
    editForm.area = detail.product.area || ''
    editForm.disable_area = detail.product.disable_area || ''
    editForm.little_picture = detail.product.little_picture || ''
    editForm.net_addr = detail.product.net_addr || ''
    editForm.flag = Boolean(detail.product.flag)
    editForm.number_sel = detail.product.number_sel || 0
    editForm.back_money_type = detail.product.back_money_type || ''
    editForm.taocan = detail.product.taocan || ''
    editForm.rule = detail.product.rule || ''
    editForm.age1 = detail.product.age1 || 0
    editForm.age2 = detail.product.age2 || 0
    editForm.price_time = detail.product.price_time || ''
    editForm.skus = detail.skus || []
    editDialogVisible.value = true
  } finally {
    editLoading.value = false
  }
}

const saveEdit = async () => {
  if (editProductID.value <= 0) {
    // 新建
    if (!editForm.product_id) {
      ElMessage.error('请填写 ProductID')
      return
    }
  }
  editLoading.value = true
  try {
    const payload = {
      product_id: editForm.product_id,
      category_id: editForm.category_id,
      operator: editForm.operator,
      product_name: editForm.product_name,
      main_pic: editForm.main_pic,
      area: editForm.area,
      disable_area: editForm.disable_area,
      little_picture: editForm.little_picture,
      net_addr: editForm.net_addr,
      flag: editForm.flag,
      number_sel: editForm.number_sel,
      back_money_type: editForm.back_money_type,
      taocan: editForm.taocan,
      rule: editForm.rule,
      age1: editForm.age1,
      age2: editForm.age2,
      price_time: editForm.price_time,
      skus: editForm.skus,
    }
    const { data: res } =
      editProductID.value > 0
        ? await haokaUpdateProduct(editProductID.value, payload)
        : await haokaCreateProduct(payload)
    if (res.code !== 200) {
      ElMessage.error(res.message || '保存失败')
      return
    }
    ElMessage.success('保存成功')
    editDialogVisible.value = false
    await loadList()
  } finally {
    editLoading.value = false
  }
}

onMounted(async () => {
  // 读取系统配置里的号卡 user_id / secret / 定时同步开关
  try {
    const { data: res } = await getSystemConfig()
    if (res.code === 200 && res.data) {
      const cfg = res.data as ISystemConfig
      credential.user_id = cfg.haoka_user_id || ''
      credential.secret = cfg.haoka_secret || ''
      haokaSyncEnabled.value = cfg.haoka_sync_enabled ?? false
      haokaSyncInterval.value = cfg.haoka_sync_interval ?? 3600
    }
  } catch {
    // ignore
  }
  await loadCategories()
  await loadList()
})

const saveHaokaConfig = async () => {
  if (!credential.user_id || !credential.secret) {
    ElMessage.warning('请先填写 user_id 和接口 secret')
    return
  }
  savingHaokaConfig.value = true
  try {
    const { data: res } = await getSystemConfig()
    if (res.code !== 200 || !res.data) return
    const current = res.data as ISystemConfig
    const payload: ISystemConfig = {
      ...current,
      haoka_user_id: credential.user_id,
      haoka_secret: credential.secret,
      haoka_sync_enabled: haokaSyncEnabled.value,
      haoka_sync_interval: haokaSyncInterval.value,
    }
    const { data: putRes } = await updateSystemConfig(payload)
    if (putRes.code !== 200) {
      ElMessage.error(putRes.message || '保存失败')
      return
    }
    ElMessage.success('号卡配置已保存')
  } finally {
    savingHaokaConfig.value = false
  }
}
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

        <el-form-item label="定时同步">
          <el-switch v-model="haokaSyncEnabled" />
        </el-form-item>
        <el-form-item label="同步间隔(秒)">
          <el-input-number v-model="haokaSyncInterval" :min="300" :max="86400" />
        </el-form-item>
      </el-form>

      <div class="bar">
        <el-button type="primary" :loading="loading" @click="syncAll">同步数据</el-button>
        <el-button @click="openAddDialog">添加号卡</el-button>
        <el-button type="success" :loading="savingHaokaConfig" @click="saveHaokaConfig">保存配置</el-button>
      </div>

      <el-divider />

      <div class="filters">
        <el-select v-model="filters.category_id" placeholder="分类" clearable style="width: 220px">
          <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="String(c.id)" />
        </el-select>

        <el-select v-model="filters.flag" placeholder="上架状态" clearable style="width: 160px">
          <el-option label="上架(true)" value="true" />
          <el-option label="已下架(false)" value="false" />
        </el-select>

        <el-input v-model="filters.operator" placeholder="运营商(电信/移动/联通)" clearable style="width: 220px" />

        <el-button
          type="primary"
          @click="
            () => {
              page = 1 as any
              loadList()
            }
          "
          v-if="!loading"
        >
          查询
        </el-button>
      </div>

      <el-table :data="list" border style="width: 100%; margin-top: 12px" v-loading="loading">
        <el-table-column prop="product_id" label="ProductID" width="120" />
        <el-table-column prop="product_name" label="套餐名称" min-width="220" />
        <el-table-column prop="category_name" label="分类" width="140" />
        <el-table-column prop="operator" label="运营商" width="140" />
        <el-table-column prop="main_pic" label="主图" width="120">
          <template #default="{ row }">
            <el-image v-if="row.main_pic" :src="row.main_pic" style="width: 60px; height: 40px" fit="cover" />
          </template>
        </el-table-column>

        <el-table-column label="上架" width="120">
          <template #default="{ row }">
            <el-switch
              :model-value="row.flag"
              inline-prompt
              active-text="上架"
              inactive-text="下架"
              @change="(v) => setShelf(row.id, Boolean(v))"
            />
          </template>
        </el-table-column>

        <el-table-column prop="area" label="归属地" width="140" />
        <el-table-column prop="disable_area" label="禁发区域" min-width="180" show-overflow-tooltip />
        <el-table-column prop="number_sel" label="选号" width="90" />
        <el-table-column prop="back_money_type" label="返佣类型" width="120" />
        <el-table-column prop="price_time" label="优惠时间" width="120" />
        <el-table-column prop="age1" label="年龄(最小)" width="120" />
        <el-table-column prop="age2" label="年龄(最大)" width="120" />
        <el-table-column prop="taocan" label="套餐说明" min-width="200" show-overflow-tooltip />

        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEditDialog(row)">编辑</el-button>
            <el-button link type="success" @click="() => { resetEditForm(); editDialogVisible = true; editForm.product_id = row.product_id }">
              复制新增
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pager">
        <el-pagination
          background
          layout="total, prev, pager, next, sizes"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          v-model:current-page="page"
          v-model:page-size="pageSize"
          @current-change="loadList"
          @size-change="
            () => {
              page = 1 as any
              loadList()
            }
          "
        />
      </div>
    </el-card>
  </div>

  <el-dialog v-model="addDialogVisible" title="添加号卡（来自外部上架查询接口）" width="980px" :close-on-click-modal="false">
    <el-form label-width="90px">
      <el-form-item label="ProductID(可空)" style="margin-bottom: 0">
        <el-input v-model="addDialogProductID" placeholder="不填返回所有上架商品" style="max-width: 420px" />
      </el-form-item>
    </el-form>

    <div class="dialog-bar">
      <el-button type="primary" :loading="addDialogLoading" @click="queryAddDialogProducts">查询上架商品</el-button>
      <el-button @click="saveSelectedAsLocal" :loading="addDialogLoading" type="success">保存为号卡</el-button>
    </div>

    <el-table :data="addQueryResults" border style="width: 100%; margin-top: 12px" v-loading="addDialogLoading" @row-click="selectedProduct = $event">
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

  <el-dialog v-model="editDialogVisible" :title="editProductID > 0 ? '编辑号卡' : '手动新增号卡'" width="1100px" :close-on-click-modal="false">
    <div class="edit-grid">
      <el-form label-width="120px" class="edit-form">
        <el-form-item label="ProductID">
          <el-input-number v-model="editForm.product_id" :min="1" :max="999999999" />
        </el-form-item>

        <el-form-item label="分类">
          <el-select v-model="editForm.category_id" placeholder="请选择分类" style="width: 260px">
            <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="运营商">
          <el-input v-model="editForm.operator" style="width: 260px" />
        </el-form-item>

        <el-form-item label="套餐名称">
          <el-input v-model="editForm.product_name" style="width: 420px" />
        </el-form-item>

        <el-form-item label="主图URL">
          <el-input v-model="editForm.main_pic" style="width: 420px" />
        </el-form-item>

        <el-form-item label="归属地">
          <el-input v-model="editForm.area" style="width: 260px" />
        </el-form-item>

        <el-form-item label="禁发区域">
          <el-input v-model="editForm.disable_area" style="width: 420px" />
        </el-form-item>

        <el-form-item label="详情小图URL">
          <el-input v-model="editForm.little_picture" style="width: 420px" />
        </el-form-item>

        <el-form-item label="资料介绍URL">
          <el-input v-model="editForm.net_addr" style="width: 420px" />
        </el-form-item>

        <el-form-item label="上架">
          <el-switch v-model="editForm.flag" active-text="上架" inactive-text="下架" />
        </el-form-item>

        <el-form-item label="选号">
          <el-input-number v-model="editForm.number_sel" :min="0" :max="2" />
        </el-form-item>

        <el-form-item label="返佣类型">
          <el-input v-model="editForm.back_money_type" style="width: 260px" />
        </el-form-item>

        <el-form-item label="套餐说明">
          <el-input v-model="editForm.taocan" type="textarea" :rows="3" style="width: 520px" />
        </el-form-item>

        <el-form-item label="结算规则">
          <el-input v-model="editForm.rule" type="textarea" :rows="3" style="width: 520px" />
        </el-form-item>

        <el-form-item label="年龄最小">
          <el-input-number v-model="editForm.age1" :min="0" :max="120" />
        </el-form-item>

        <el-form-item label="年龄最大">
          <el-input-number v-model="editForm.age2" :min="0" :max="120" />
        </el-form-item>

        <el-form-item label="优惠时间">
          <el-input v-model="editForm.price_time" style="width: 260px" />
        </el-form-item>
      </el-form>

      <div class="sku-area">
        <div class="sku-title">SKUs</div>
        <el-button type="primary" plain style="margin-bottom: 10px" @click="editForm.skus.push({ sku_id: 0, sku_name: '', desc: '' })">
          添加 SKU
        </el-button>
        <el-table :data="editForm.skus" border style="width: 100%">
          <el-table-column label="SkuID" width="120">
            <template #default="{ row }">
              <el-input-number v-model="row.sku_id" :min="0" />
            </template>
          </el-table-column>
          <el-table-column label="SkuName" min-width="200">
            <template #default="{ row }">
              <el-input v-model="row.sku_name" />
            </template>
          </el-table-column>
          <el-table-column label="Desc" min-width="280">
            <template #default="{ row }">
              <el-input v-model="row.desc" />
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <template #footer>
      <div class="footer-bar">
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editLoading" @click="saveEdit">保存</el-button>
      </div>
    </template>
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
.pager {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}
.dialog-bar {
  margin-top: 10px;
  display: flex;
  gap: 12px;
}
.edit-grid {
  display: flex;
  gap: 18px;
  align-items: flex-start;
}
.edit-form {
  flex: 1;
  min-width: 520px;
}
.sku-area {
  width: 420px;
}
.sku-title {
  font-weight: 700;
  margin-bottom: 8px;
}
.footer-bar {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>

