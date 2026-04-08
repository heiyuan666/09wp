<template>
  <div class="api-doc-page">
    <section class="hero-card">
      <div class="hero-copy">
        <span class="eyebrow">ADMIN API</span>
        <h1>网盘资源管理 API 文档</h1>
        <p>
          这页对应的是后台“网盘资源管理”接口，不是游戏详情接口。核心围绕
          <code>/api/v1/admin/resources</code>
          这一组接口，适合后台系统对接、脚本批量管理、运营工具调用。
        </p>
        <div class="hero-actions">
          <el-button type="primary" size="large" @click="scrollToTester">在线调试</el-button>
          <el-button size="large" @click="copyText(resourceListUrl)">复制列表接口</el-button>
        </div>
      </div>
      <div class="hero-panel">
        <div class="hero-panel-label">Authorization</div>
        <div class="hero-panel-url">Bearer &lt;admin_token&gt;</div>
        <div class="hero-panel-tip">除登录接口外，资源管理接口均需管理员 token</div>
      </div>
    </section>

    <section class="doc-grid two-up">
      <article class="doc-card">
        <div class="section-head">
          <div>
            <h2>认证流程</h2>
            <p>先登录管理员接口，拿到 token 后再访问资源管理接口。</p>
          </div>
        </div>
        <div class="endpoint-list">
          <div class="endpoint-item">
            <span class="method post">POST</span>
            <div class="endpoint-main">
              <div class="endpoint-path">/api/v1/login</div>
              <div class="endpoint-desc">管理员登录，返回后台 JWT token</div>
            </div>
            <el-button text @click="copyText(loginUrl)">复制</el-button>
          </div>
        </div>
        <pre class="code-block"><code>{{ loginExample }}</code></pre>
      </article>

      <article class="doc-card">
        <div class="section-head">
          <div>
            <h2>返回格式</h2>
            <p>项目当前统一返回结构如下。</p>
          </div>
          <el-button text @click="copyText(responseExample)">复制示例</el-button>
        </div>
        <pre class="code-block"><code>{{ responseExample }}</code></pre>
      </article>
    </section>

    <section class="doc-card">
      <div class="section-head">
        <div>
          <h2>资源管理接口总览</h2>
          <p>以下接口都属于后台“网盘资源管理”。</p>
        </div>
      </div>
      <div class="endpoint-list">
        <div v-for="item in endpointList" :key="item.path + item.method" class="endpoint-item">
          <span :class="['method', item.methodClass]">{{ item.method }}</span>
          <div class="endpoint-main">
            <div class="endpoint-path">{{ item.path }}</div>
            <div class="endpoint-desc">{{ item.desc }}</div>
          </div>
          <el-button text @click="copyText(item.fullUrl)">复制</el-button>
        </div>
      </div>
    </section>

    <section class="doc-grid">
      <article class="doc-card">
        <div class="section-head">
          <div>
            <h2>列表查询参数</h2>
            <p>用于 `GET /api/v1/admin/resources`。</p>
          </div>
        </div>
        <el-table :data="listParams" stripe class="param-table">
          <el-table-column prop="name" label="参数" min-width="120" />
          <el-table-column prop="type" label="类型" width="120" />
          <el-table-column prop="required" label="必填" width="100" />
          <el-table-column prop="desc" label="说明" min-width="260" />
        </el-table>
      </article>

      <article class="doc-card">
        <div class="section-head">
          <div>
            <h2>新增/编辑字段</h2>
            <p>用于 `POST /api/v1/admin/resources` 和 `PUT /api/v1/admin/resources/:id`。</p>
          </div>
        </div>
        <el-table :data="payloadFields" stripe class="param-table">
          <el-table-column prop="name" label="字段" min-width="120" />
          <el-table-column prop="type" label="类型" width="120" />
          <el-table-column prop="required" label="必填" width="100" />
          <el-table-column prop="desc" label="说明" min-width="260" />
        </el-table>
      </article>
    </section>

    <section class="doc-grid">
      <article class="doc-card">
        <div class="section-head">
          <div>
            <h2>示例请求</h2>
            <p>下面这些例子都基于后台网盘资源管理接口。</p>
          </div>
        </div>
        <div class="snippet-group">
          <div class="snippet-head">
            <span>查询资源列表</span>
            <el-button text @click="copyText(curlListExample)">复制</el-button>
          </div>
          <pre class="code-block"><code>{{ curlListExample }}</code></pre>
        </div>
        <div class="snippet-group">
          <div class="snippet-head">
            <span>新增资源</span>
            <el-button text @click="copyText(curlCreateExample)">复制</el-button>
          </div>
          <pre class="code-block"><code>{{ curlCreateExample }}</code></pre>
        </div>
        <div class="snippet-group">
          <div class="snippet-head">
            <span>更新资源</span>
            <el-button text @click="copyText(curlUpdateExample)">复制</el-button>
          </div>
          <pre class="code-block"><code>{{ curlUpdateExample }}</code></pre>
        </div>
        <div class="snippet-group">
          <div class="snippet-head">
            <span>批量删除</span>
            <el-button text @click="copyText(curlBatchDeleteExample)">复制</el-button>
          </div>
          <pre class="code-block"><code>{{ curlBatchDeleteExample }}</code></pre>
        </div>
      </article>

      <article ref="testerRef" class="doc-card">
        <div class="section-head">
          <div>
            <h2>在线调试</h2>
            <p>这里调的是后台资源列表接口，需要管理员 token。</p>
          </div>
        </div>

        <el-form label-position="top" class="tester-form">
          <el-form-item label="管理员 Token">
            <el-input v-model="tester.token" type="textarea" :rows="3" placeholder="粘贴 Bearer token，不需要写 Bearer 前缀" />
          </el-form-item>
          <div class="tester-row">
            <el-form-item label="title">
              <el-input v-model="tester.title" placeholder="按标题搜索" clearable />
            </el-form-item>
            <el-form-item label="category_id">
              <el-input v-model="tester.category_id" placeholder="分类 ID" clearable />
            </el-form-item>
            <el-form-item label="status">
              <el-select v-model="tester.status" clearable placeholder="全部状态">
                <el-option label="1" value="1" />
                <el-option label="0" value="0" />
              </el-select>
            </el-form-item>
          </div>
          <div class="tester-row compact">
            <el-form-item label="page">
              <el-input-number v-model="tester.page" :min="1" :step="1" />
            </el-form-item>
            <el-form-item label="page_size">
              <el-input-number v-model="tester.page_size" :min="1" :max="100" :step="1" />
            </el-form-item>
          </div>
          <div class="tester-actions">
            <el-button type="primary" :loading="loading" @click="runTester">请求列表</el-button>
            <el-button @click="copyText(testerUrl)">复制请求地址</el-button>
          </div>
        </el-form>

        <div class="request-preview">
          <div class="request-preview-label">当前请求</div>
          <code>{{ testerUrl }}</code>
        </div>

        <pre class="code-block result-block"><code>{{ testerResult }}</code></pre>
      </article>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'

const rawBase = (import.meta.env.VITE_API_BASE_URL || '/api/v1').replace(/\/+$/, '')
const origin = typeof window !== 'undefined' ? window.location.origin : ''
const apiBaseUrl = rawBase.startsWith('http') ? rawBase : `${origin}${rawBase}`

const loginUrl = `${apiBaseUrl}/login`
const resourceListUrl = `${apiBaseUrl}/admin/resources`
const resourceCreateUrl = `${apiBaseUrl}/admin/resources`
const resourceUpdateUrl = `${apiBaseUrl}/admin/resources/:id`
const resourceDeleteUrl = `${apiBaseUrl}/admin/resources/:id`
const resourceRetryUrl = `${apiBaseUrl}/admin/resources/:id/retry-transfer`
const resourceLogsUrl = `${apiBaseUrl}/admin/resources/:id/transfer-logs`
const resourceBatchDeleteUrl = `${apiBaseUrl}/admin/resources/batch-delete`
const resourceBatchStatusUrl = `${apiBaseUrl}/admin/resources/batch-status`
const resourceCheckUrl = `${apiBaseUrl}/admin/resources/check-links`
const resourceSyncTelegramUrl = `${apiBaseUrl}/admin/resources/sync-telegram`

const endpointList = [
  { method: 'GET', methodClass: 'get', path: '/api/v1/admin/resources', desc: '资源列表分页查询', fullUrl: resourceListUrl },
  { method: 'POST', methodClass: 'post', path: '/api/v1/admin/resources', desc: '新增资源', fullUrl: resourceCreateUrl },
  { method: 'PUT', methodClass: 'put', path: '/api/v1/admin/resources/:id', desc: '更新资源', fullUrl: resourceUpdateUrl },
  { method: 'DELETE', methodClass: 'delete', path: '/api/v1/admin/resources/:id', desc: '删除单条资源', fullUrl: resourceDeleteUrl },
  { method: 'POST', methodClass: 'post', path: '/api/v1/admin/resources/batch-delete', desc: '批量删除资源', fullUrl: resourceBatchDeleteUrl },
  { method: 'POST', methodClass: 'post', path: '/api/v1/admin/resources/batch-status', desc: '批量修改资源状态', fullUrl: resourceBatchStatusUrl },
  { method: 'POST', methodClass: 'post', path: '/api/v1/admin/resources/:id/retry-transfer', desc: '手动重试转存', fullUrl: resourceRetryUrl },
  { method: 'GET', methodClass: 'get', path: '/api/v1/admin/resources/:id/transfer-logs', desc: '查看某条资源的转存日志', fullUrl: resourceLogsUrl },
  { method: 'POST', methodClass: 'post', path: '/api/v1/admin/resources/check-links', desc: '批量检测链接有效性', fullUrl: resourceCheckUrl },
  { method: 'POST', methodClass: 'post', path: '/api/v1/admin/resources/sync-telegram', desc: '同步 Telegram 资源', fullUrl: resourceSyncTelegramUrl },
]

const listParams = [
  { name: 'page', type: 'number', required: '否', desc: '页码，默认 1。' },
  { name: 'page_size', type: 'number', required: '否', desc: '每页条数，默认 20，最大 100。' },
  { name: 'title', type: 'string', required: '否', desc: '标题模糊搜索。' },
  { name: 'category_id', type: 'number', required: '否', desc: '分类 ID 过滤。' },
  { name: 'status', type: 'number', required: '否', desc: '状态过滤，通常 1 为显示，0 为隐藏。' },
]

const payloadFields = [
  { name: 'title', type: 'string', required: '是', desc: '资源标题。' },
  { name: 'link', type: 'string', required: '是', desc: '网盘分享链接。' },
  { name: 'category_id', type: 'number', required: '是', desc: '所属分类 ID。' },
  { name: 'description', type: 'string', required: '否', desc: '资源描述。' },
  { name: 'extract_code', type: 'string', required: '否', desc: '提取码。' },
  { name: 'cover', type: 'string', required: '否', desc: '封面图链接。' },
  { name: 'tags', type: 'string', required: '否', desc: '标签字符串，逗号分隔。' },
  { name: 'sort_order', type: 'number', required: '否', desc: '排序值。' },
  { name: 'status', type: 'number', required: '否', desc: '状态值。' },
]

const loginExample = `curl -X POST "${loginUrl}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "username": "admin",
    "password": "123456"
  }'`

const responseExample = `{
  "code": 200,
  "message": "获取成功",
  "data": {
    "list": [],
    "total": 0
  }
}`

const curlListExample = `curl "${resourceListUrl}?page=1&page_size=20&title=黑神话&status=1" \\
  -H "Authorization: Bearer <admin_token>"`

const curlCreateExample = `curl -X POST "${resourceCreateUrl}" \\
  -H "Authorization: Bearer <admin_token>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "title": "黑神话：悟空 夸克版",
    "link": "https://pan.quark.cn/s/xxxx",
    "category_id": 3,
    "description": "含本体与更新补丁",
    "extract_code": "8a2d",
    "cover": "https://example.com/cover.jpg",
    "tags": "动作,单机,夸克",
    "sort_order": 100,
    "status": 1
  }'`

const curlUpdateExample = `curl -X PUT "${apiBaseUrl}/admin/resources/12" \\
  -H "Authorization: Bearer <admin_token>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "title": "黑神话：悟空 夸克版",
    "link": "https://pan.quark.cn/s/xxxx",
    "category_id": 3,
    "description": "更新后的描述",
    "extract_code": "8a2d",
    "cover": "https://example.com/cover.jpg",
    "tags": "动作,单机,夸克",
    "sort_order": 120,
    "status": 1
  }'`

const curlBatchDeleteExample = `curl -X POST "${resourceBatchDeleteUrl}" \\
  -H "Authorization: Bearer <admin_token>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "ids": [12, 18, 21]
  }'`

const testerRef = ref<HTMLElement>()
const loading = ref(false)
const testerResult = ref(`{
  "tip": "这里会显示后台资源列表接口的返回结果"
}`)

const tester = ref({
  token: '',
  title: '',
  category_id: '',
  status: '',
  page: 1,
  page_size: 10,
})

const testerUrl = computed(() => {
  const params = new URLSearchParams()
  if (tester.value.title) params.set('title', tester.value.title)
  if (tester.value.category_id) params.set('category_id', tester.value.category_id)
  if (tester.value.status) params.set('status', tester.value.status)
  if (tester.value.page) params.set('page', String(tester.value.page))
  if (tester.value.page_size) params.set('page_size', String(tester.value.page_size))
  const query = params.toString()
  return query ? `${resourceListUrl}?${query}` : resourceListUrl
})

const copyText = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制')
  } catch {
    ElMessage.error('复制失败')
  }
}

const scrollToTester = () => {
  testerRef.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

const runTester = async () => {
  if (!tester.value.token.trim()) {
    ElMessage.warning('请先填写管理员 token')
    return
  }

  loading.value = true
  try {
    const resp = await fetch(testerUrl.value, {
      headers: {
        Authorization: `Bearer ${tester.value.token.trim()}`,
      },
    })
    const json = await resp.json()
    testerResult.value = JSON.stringify(json, null, 2)
  } catch (error: any) {
    testerResult.value = JSON.stringify(
      {
        error: true,
        message: error?.message || 'request failed',
      },
      null,
      2,
    )
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.api-doc-page {
  width: min(1180px, calc(100vw - 32px));
  margin: 28px auto 56px;
  display: grid;
  gap: 20px;
}

.hero-card,
.doc-card {
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 24px;
  background:
    radial-gradient(circle at top right, rgba(14, 165, 233, 0.14), transparent 28%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(247, 249, 252, 0.98));
  box-shadow: 0 18px 46px rgba(15, 23, 42, 0.08);
}

.hero-card {
  display: grid;
  grid-template-columns: 1.55fr 0.95fr;
  gap: 20px;
  padding: 34px;
}

.hero-copy h1 {
  margin: 10px 0 14px;
  font-size: 34px;
  line-height: 1.15;
  color: #111827;
}

.hero-copy p,
.section-head p,
.endpoint-desc {
  color: #5b6475;
  line-height: 1.7;
}

.hero-copy code {
  padding: 2px 8px;
  border-radius: 999px;
  background: rgba(15, 23, 42, 0.06);
  color: #0f172a;
}

.eyebrow {
  display: inline-flex;
  padding: 6px 12px;
  border-radius: 999px;
  background: rgba(14, 165, 233, 0.1);
  color: #0284c7;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.hero-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
}

.hero-panel {
  padding: 22px;
  border-radius: 20px;
  background: linear-gradient(180deg, #0f172a, #172033);
  color: #e5eefc;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.hero-panel-label {
  font-size: 12px;
  letter-spacing: 0.08em;
  opacity: 0.72;
  text-transform: uppercase;
}

.hero-panel-url {
  margin-top: 12px;
  font-size: 22px;
  font-weight: 700;
  word-break: break-all;
}

.hero-panel-tip {
  margin-top: 10px;
  color: rgba(229, 238, 252, 0.72);
  line-height: 1.7;
}

.doc-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20px;
}

.doc-grid.two-up {
  align-items: start;
}

.doc-card {
  padding: 26px;
}

.section-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}

.section-head h2 {
  margin: 0 0 6px;
  font-size: 22px;
  color: #111827;
}

.endpoint-list {
  display: grid;
  gap: 14px;
}

.endpoint-item {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px 18px;
  border-radius: 18px;
  background: #f8fafc;
  border: 1px solid rgba(148, 163, 184, 0.24);
}

.method {
  min-width: 64px;
  padding: 8px 0;
  border-radius: 999px;
  text-align: center;
  font-size: 12px;
  font-weight: 700;
}

.method.get {
  color: #065f46;
  background: #d1fae5;
}

.method.post {
  color: #9a3412;
  background: #ffedd5;
}

.method.put {
  color: #1d4ed8;
  background: #dbeafe;
}

.method.delete {
  color: #b91c1c;
  background: #fee2e2;
}

.endpoint-main {
  flex: 1;
  min-width: 0;
}

.endpoint-path {
  font-family: Consolas, Monaco, monospace;
  font-size: 14px;
  color: #0f172a;
  word-break: break-all;
}

.code-block {
  margin: 0;
  padding: 18px;
  border-radius: 18px;
  background: #0f172a;
  color: #dbeafe;
  overflow: auto;
  font-size: 13px;
  line-height: 1.7;
}

.param-table :deep(.el-table__cell) {
  vertical-align: top;
}

.snippet-group + .snippet-group {
  margin-top: 16px;
}

.snippet-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
  color: #111827;
  font-weight: 600;
}

.tester-form {
  margin-top: 10px;
}

.tester-row {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.tester-row.compact {
  grid-template-columns: repeat(2, minmax(0, 220px));
}

.tester-actions {
  display: flex;
  gap: 12px;
  margin-top: 6px;
}

.request-preview {
  margin: 18px 0 14px;
  padding: 14px 16px;
  border-radius: 16px;
  background: #f8fafc;
  border: 1px solid rgba(148, 163, 184, 0.22);
}

.request-preview-label {
  margin-bottom: 6px;
  color: #64748b;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.request-preview code {
  word-break: break-all;
  color: #0f172a;
}

.result-block {
  min-height: 320px;
}

@media (max-width: 960px) {
  .hero-card,
  .doc-grid,
  .tester-row,
  .tester-row.compact {
    grid-template-columns: 1fr;
  }

  .hero-card,
  .doc-card {
    padding: 20px;
  }

  .hero-copy h1 {
    font-size: 28px;
  }

  .hero-actions,
  .tester-actions {
    flex-direction: column;
  }
}
</style>
