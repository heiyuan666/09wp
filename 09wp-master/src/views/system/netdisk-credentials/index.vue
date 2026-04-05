<template>
  <el-card shadow="never">
    <template #header>
      <div class="header">
        <span>网盘凭证</span>
        <span class="sub">用于一键转存：请填写各网盘登录后的 Cookie / Token，勿泄露给他人</span>
      </div>
    </template>

    <el-form :model="form" label-width="140px" class="cred-form">
      <el-alert
        type="info"
        show-icon
        :closable="false"
        class="mb16"
        title="转存后替换为本人分享链接"
        description="开启后：自动或手动转存成功时，会尽量把资源管理里的链接改为您自己网盘新生成的分享地址。已对接：阿里云盘、百度、夸克、UC、115、123 云盘、天翼云盘（各盘以接口实际返回为准，若失败会保留原链接并附说明）。"
      />
      <el-form-item label="替换为本人链接">
        <el-switch v-model="form.replace_link_after_transfer" />
        <span class="hint">需同时开启对应网盘的「自动转存」或手动点转存</span>
      </el-form-item>

      <el-divider content-position="left">百度网盘</el-divider>
      <el-form-item label="百度 Cookie">
        <el-input
          v-model="form.baidu_cookie"
          type="textarea"
          :rows="4"
          placeholder="浏览器登录 pan.baidu.com 后复制整段 Cookie（需含 BDUSS、STOKEN 等）"
        />
        <CookieAccountPool
          v-model="form.baidu_cookie_accounts"
          class="pool-mt"
          :show-folder-id="false"
          :show-target-path="true"
          hint="多账号轮询：每行可单独填转存路径；路径留空则用上方「转存目录路径」。Cookie 填完整百度 Cookie。"
        />
      </el-form-item>
      <el-form-item label="转存目录路径">
        <el-input v-model="form.baidu_target_path" placeholder="默认 /（网盘根目录），如 /我的资源" />
      </el-form-item>
      <el-form-item label="自动转存">
        <el-switch v-model="form.baidu_auto_save" />
      </el-form-item>

      <el-divider content-position="left">迅雷网盘</el-divider>
      <el-form-item label="迅雷 refresh_token">
        <el-input
          v-model="form.xunlei_cookie"
          type="textarea"
          :rows="3"
          placeholder="登录 pan.xunlei.com 后抓取 /v1/auth/token 使用的 refresh_token"
        />
        <CookieAccountPool
          v-model="form.xunlei_cookie_accounts"
          class="pool-mt"
          cookie-placeholder="refresh_token"
          folder-id-placeholder="转存目录 ID（留空用上方全局，默认 0）"
          hint="多个迅雷账号时按顺序轮流使用 refresh_token；Token 填在下方输入框。目录 ID 可每账号单独设置，留空则用上方全局。"
        />
      </el-form-item>
      <el-form-item label="转存目录 ID">
        <el-input v-model="form.xunlei_target_folder_id" placeholder="默认 0（根目录）" />
      </el-form-item>
      <el-form-item label="自动转存">
        <el-switch v-model="form.xunlei_auto_save" />
      </el-form-item>

      <el-divider content-position="left">夸克网盘</el-divider>
      <el-form-item label="夸克 Cookie">
        <el-input v-model="form.quark_cookie" type="textarea" :rows="3" placeholder="用于自动转存，建议填写完整 cookie" />
        <CookieAccountPool
          v-model="form.quark_cookie_accounts"
          class="pool-mt"
          folder-id-placeholder="转存目录 fid（留空用上方全局，默认 0）"
        />
      </el-form-item>
      <el-form-item label="转存目录ID">
        <el-input v-model="form.quark_target_folder_id" placeholder="默认 0（根目录）" />
      </el-form-item>
      <el-form-item label="自动转存">
        <el-switch v-model="form.quark_auto_save" />
      </el-form-item>
      <el-form-item label="广告过滤">
        <el-switch v-model="form.quark_ad_filter_enabled" />
        <span class="hint">开启后：转存完成会递归扫描并删除命中广告词的文件/文件夹</span>
      </el-form-item>
      <el-form-item label="广告关键词">
        <el-input
          v-model="form.quark_banned_keywords"
          type="textarea"
          :rows="3"
          placeholder="逗号分隔，例如：广告,福利,公众号"
        />
      </el-form-item>

      <el-divider content-position="left">UC 网盘</el-divider>
      <el-form-item label="UC Cookie">
        <el-input
          v-model="form.uc_cookie"
          type="textarea"
          :rows="3"
          placeholder="浏览器登录 drive.uc.cn 后复制整段 Cookie（转存接口与夸克同源）"
        />
        <CookieAccountPool
          v-model="form.uc_cookie_accounts"
          class="pool-mt"
          folder-id-placeholder="转存目录 ID（留空用上方全局，默认 0）"
        />
      </el-form-item>
      <el-form-item label="转存目录 ID">
        <el-input v-model="form.uc_target_folder_id" placeholder="默认 0（根目录）" />
      </el-form-item>
      <el-form-item label="自动转存">
        <el-switch v-model="form.uc_auto_save" />
      </el-form-item>

      <el-divider content-position="left">115 网盘</el-divider>
      <el-form-item label="115 Cookie">
        <el-input v-model="form.pan115_cookie" type="textarea" :rows="3" placeholder="登录 115 后复制 webapi 所需 Cookie" />
        <CookieAccountPool
          v-model="form.pan115_cookie_accounts"
          class="pool-mt"
          folder-id-placeholder="转存目录 cid（留空用上方全局）"
        />
      </el-form-item>
      <el-form-item label="转存目录 cid">
        <el-input v-model="form.pan115_target_folder_id" placeholder="留空为根目录" />
      </el-form-item>
      <el-form-item label="自动转存">
        <el-switch v-model="form.pan115_auto_save" />
      </el-form-item>

      <el-divider content-position="left">天翼云盘</el-divider>
      <el-form-item label="天翼 Cookie">
        <el-input v-model="form.tianyi_cookie" type="textarea" :rows="3" placeholder="浏览器登录 cloud.189.cn 后复制整段 Cookie" />
        <CookieAccountPool
          v-model="form.tianyi_cookie_accounts"
          class="pool-mt"
          folder-id-placeholder="转存目录 ID（留空用上方全局，默认 -11）"
        />
      </el-form-item>
      <el-form-item label="转存目录 ID">
        <el-input v-model="form.tianyi_target_folder_id" placeholder="默认 -11（个人云根目录）" />
      </el-form-item>
      <el-form-item label="自动转存">
        <el-switch v-model="form.tianyi_auto_save" />
      </el-form-item>

      <el-divider content-position="left">123 云盘</el-divider>
      <el-form-item label="123 Token">
        <el-input
          v-model="form.pan123_cookie"
          type="textarea"
          :rows="3"
          placeholder="登录 www.123pan.com 后从请求头复制 authorization: Bearer ..."
        />
        <CookieAccountPool
          v-model="form.pan123_cookie_accounts"
          class="pool-mt"
          folder-id-placeholder="转存目录 fileId（留空用上方全局，默认 0）"
        />
      </el-form-item>
      <el-form-item label="转存目录 fileId">
        <el-input v-model="form.pan123_target_folder_id" placeholder="默认 0（根目录）" />
      </el-form-item>
      <el-form-item label="自动转存">
        <el-switch v-model="form.pan123_auto_save" />
      </el-form-item>

      <el-divider content-position="left">阿里云盘</el-divider>
      <el-form-item label="refresh_token">
        <el-input
          v-model="form.aliyun_refresh_token"
          type="textarea"
          :rows="3"
          placeholder="阿里云盘开放平台 OAuth refresh_token（与 AList 等工具所用一致；勿泄露）"
        />
        <CookieAccountPool
          v-model="form.aliyun_refresh_token_accounts"
          class="pool-mt"
          cookie-placeholder="refresh_token"
          folder-id-placeholder="转存到目录 file_id（留空用上方全局，默认 root）"
          hint="多个阿里云账号轮流使用 refresh_token；每行可单独指定转存目录 file_id，留空则用上方全局。"
        />
      </el-form-item>
      <el-form-item label="转存到目录 file_id">
        <el-input v-model="form.aliyun_target_parent_file_id" placeholder="默认 root（根目录）；子目录填对应 folder 的 file_id" />
      </el-form-item>
      <el-form-item label="自动转存">
        <el-switch v-model="form.aliyun_auto_save" />
      </el-form-item>

      <el-form-item>
        <el-button type="primary" :loading="saving" @click="save" v-permission="['config:update']">保存</el-button>
        <el-button @click="load">重新加载</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<script setup lang="ts">
import {
  getNetdiskCredentials,
  type INetdiskCredential,
  updateNetdiskCredentials,
} from '@/api/netdiskCredential'
import CookieAccountPool from './CookieAccountPool.vue'

defineOptions({ name: 'Netdisk-credentialsView' })

const saving = ref(false)
const form = reactive<INetdiskCredential>({
  quark_cookie: '',
  quark_target_folder_id: '0',
  quark_auto_save: false,
  quark_ad_filter_enabled: false,
  quark_banned_keywords: '',
  pan115_cookie: '',
  pan115_target_folder_id: '',
  pan115_auto_save: false,
  tianyi_cookie: '',
  tianyi_target_folder_id: '-11',
  tianyi_auto_save: false,
  pan123_cookie: '',
  pan123_target_folder_id: '0',
  pan123_auto_save: false,
  baidu_cookie: '',
  baidu_target_path: '/',
  baidu_auto_save: false,
  xunlei_cookie: '',
  xunlei_target_folder_id: '0',
  xunlei_auto_save: false,
  uc_cookie: '',
  uc_target_folder_id: '0',
  uc_auto_save: false,
  aliyun_refresh_token: '',
  aliyun_target_parent_file_id: 'root',
  aliyun_auto_save: false,
  replace_link_after_transfer: false,
  quark_cookie_accounts: [],
  uc_cookie_accounts: [],
  pan115_cookie_accounts: [],
  tianyi_cookie_accounts: [],
  pan123_cookie_accounts: [],
  baidu_cookie_accounts: [],
  aliyun_refresh_token_accounts: [],
  xunlei_cookie_accounts: [],
})

function normalizeCookieAccounts(data: INetdiskCredential) {
  const keys = [
    'quark_cookie_accounts',
    'uc_cookie_accounts',
    'pan115_cookie_accounts',
    'tianyi_cookie_accounts',
    'pan123_cookie_accounts',
    'baidu_cookie_accounts',
    'aliyun_refresh_token_accounts',
    'xunlei_cookie_accounts',
  ] as const
  for (const k of keys) {
    if (!Array.isArray(data[k])) {
      ;(data as Record<string, unknown>)[k] = []
    }
  }
}

const load = async () => {
  const { data: res } = await getNetdiskCredentials()
  if (res.code !== 200) return
  const data = { ...res.data }
  normalizeCookieAccounts(data)
  Object.assign(form, data)
}

const save = async () => {
  saving.value = true
  try {
    const payload: INetdiskCredential = { ...form }
    const { data: res } = await updateNetdiskCredentials(payload)
    if (res.code !== 200) return
    ElMessage.success('保存成功')
    await load()
  } finally {
    saving.value = false
  }
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
.cred-form {
  max-width: 900px;
}
.mb16 {
  margin-bottom: 16px;
}
.hint {
  margin-left: 8px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.pool-mt {
  margin-top: 12px;
}
</style>
