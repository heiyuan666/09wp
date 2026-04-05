<template>
  <div class="doc-page">
    <el-card shadow="hover" class="block">
      <template #header>
        <span class="head-title">网盘开放 API 说明</span>
      </template>
      <p class="lead">
        以下为前台可匿名调用的只读接口，用于第三方站点同步资源元数据。实际请求前缀以部署为准，开发环境一般为
        <code>{{ defaultBase }}</code>。
      </p>
    </el-card>

    <el-card shadow="hover" class="block">
      <template #header>资源列表</template>
      <p class="mono method">GET {{ baseExample }}/open/netdisk/resources</p>
      <p class="hint">查询参数（均为可选，与站内资源筛选一致）：</p>
      <el-table :data="listParams" border size="small" class="param-table">
        <el-table-column prop="name" label="参数" width="140" />
        <el-table-column prop="desc" label="说明" />
      </el-table>
      <p class="hint mt">成功时 <code>data</code> 形状：</p>
      <pre class="code">{{ listResponse }}</pre>
    </el-card>

    <el-card shadow="hover" class="block">
      <template #header>资源详情</template>
      <p class="mono method">GET {{ baseExample }}/open/netdisk/resources/:id</p>
      <p class="hint"><code>:id</code> 为资源数字 ID。未上架或不存在时返回业务错误码。</p>
      <p class="hint">成功时 <code>data</code> 在列表项字段基础上，可能包含 <code>latest_transfer</code>（最近一次转存记录摘要）。</p>
    </el-card>

    <el-card shadow="hover" class="block">
      <template #header>通用响应</template>
      <pre class="code">{{ commonEnvelope }}</pre>
      <p class="hint">列表接口的 <code>data.list</code> 单项主要字段：<code>id</code>、<code>title</code>、
        <code>category_id</code>、<code>category_name</code>、<code>platform</code>、<code>link</code>、
        <code>extract_code</code>、<code>description</code>、<code>cover</code>、<code>tags</code>（数组）、
        <code>link_valid</code>、<code>view_count</code>、<code>created_at</code> 等。</p>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

defineOptions({
  name: 'PublicNetdiskApiDocView',
})

const defaultBase = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'

const baseExample = computed(() =>
  String(import.meta.env.VITE_API_BASE_URL || 'http://你的域名/api/v1').replace(/\/$/, ''),
)

const listParams = [
  { name: 'page', desc: '页码，默认 1' },
  { name: 'page_size', desc: '每页条数，默认 20，最大 100' },
  { name: 'sort', desc: '排序：latest（默认）或 hot（按浏览量）' },
  { name: 'category_id', desc: '分类 ID' },
  { name: 'platform', desc: 'baidu / aliyun / quark / xunlei / uc / tianyi / yidong / pan115 / pan123 / other' },
  { name: 'link_valid', desc: '1 或 true 仅有效；0 或 false 仅失效' },
  { name: 'q', desc: '关键词，匹配标题、描述、标签' },
]

const listResponse = `{
  "code": 200,
  "message": "获取成功",
  "data": {
    "list": [ /* 资源对象数组 */ ],
    "total": 0,
    "page": 1,
    "page_size": 20
  }
}`

const commonEnvelope = `{
  "code": 200,
  "message": "获取成功",
  "data": { }
}`
</script>

<style scoped>
.doc-page {
  max-width: 900px;
  margin: 0 auto;
  padding: 16px 16px 40px;
}
.block {
  border-radius: 12px;
  margin-bottom: 16px;
}
.head-title {
  font-weight: 700;
}
.lead {
  margin: 0;
  line-height: 1.7;
  color: #334155;
  font-size: 14px;
}
.lead code {
  font-size: 13px;
  padding: 2px 6px;
  background: #f1f5f9;
  border-radius: 4px;
}
.method {
  margin: 0 0 12px;
  font-size: 13px;
  word-break: break-all;
  color: #0f172a;
}
.hint {
  margin: 0 0 10px;
  font-size: 13px;
  color: #64748b;
  line-height: 1.6;
}
.hint code {
  font-size: 12px;
}
.param-table {
  margin-bottom: 12px;
}
.mt {
  margin-top: 12px;
}
.code {
  margin: 0;
  padding: 12px 14px;
  background: #0f172a;
  color: #e2e8f0;
  border-radius: 8px;
  font-size: 12px;
  line-height: 1.5;
  overflow-x: auto;
}
</style>
