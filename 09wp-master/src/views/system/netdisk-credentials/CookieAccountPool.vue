<template>
  <div class="cookie-pool">
    <p class="hint">{{ hint }}</p>
    <div v-for="(row, i) in accounts" :key="i" class="row">
      <div class="row-line">
        <el-input v-model="row.name" placeholder="备注名，如 账号一" class="name-inp" />
        <el-input
          v-if="showFolderId"
          v-model="row.target_folder_id"
          :placeholder="folderIdPlaceholder"
          class="folder-inp"
          clearable
        />
        <div class="acts">
          <el-checkbox v-model="row.disabled">暂停</el-checkbox>
          <el-button type="danger" link @click="remove(i)">删除</el-button>
        </div>
      </div>
      <el-input
        v-model="row.cookie"
        type="textarea"
        :rows="2"
        :placeholder="cookiePlaceholder"
        class="cookie-inp"
      />
      <el-input
        v-if="showTargetPath"
        v-model="row.target_path"
        :placeholder="pathPlaceholder"
        class="path-inp"
        clearable
      />
    </div>
    <el-button size="small" @click="add">添加账号</el-button>
  </div>
</template>

<script setup lang="ts">
import type { INetdiskCookieAccount } from '@/api/netdiskCredential'

withDefaults(
  defineProps<{
    hint?: string
    /** 是否显示「转存目录 ID」列（夸克/UC/115/天翼/123/阿里/迅雷等） */
    showFolderId?: boolean
    /** 是否显示百度「转存路径」 */
    showTargetPath?: boolean
    folderIdPlaceholder?: string
    pathPlaceholder?: string
    cookiePlaceholder?: string
  }>(),
  {
    hint: '多个账号时，每次转存请求按顺序轮流使用，可减轻单账号频率限制；此处全空则仅使用上方主 Cookie。',
    showFolderId: true,
    showTargetPath: false,
    folderIdPlaceholder: '转存目录 ID（留空用上方全局）',
    pathPlaceholder: '转存路径，如 /我的资源（留空用上方全局）',
    cookiePlaceholder: 'Cookie / Token',
  },
)

const accounts = defineModel<INetdiskCookieAccount[]>({ default: () => [] })

function add() {
  accounts.value.push({
    name: '',
    cookie: '',
    disabled: false,
    target_folder_id: '',
    target_path: '',
  })
}

function remove(i: number) {
  accounts.value.splice(i, 1)
}
</script>

<style scoped lang="scss">
.cookie-pool {
  width: 100%;
}
.hint {
  margin: 0 0 10px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.5;
}
.row {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 10px;
  padding: 10px;
  background: var(--el-fill-color-light);
  border-radius: 6px;
}
.row-line {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}
.name-inp {
  width: 140px;
  flex-shrink: 0;
}
.folder-inp {
  width: 220px;
  flex: 1;
  min-width: 160px;
}
.cookie-inp {
  width: 100%;
}
.path-inp {
  width: 100%;
}
.acts {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  margin-left: auto;
}
</style>
