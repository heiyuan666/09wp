<template>
  <el-card shadow="never">
    <template #header>
      <div class="header">
        <span>系统配置</span>
      </div>
    </template>

    <el-form :model="form" label-width="120px" class="config-form">
      <el-divider content-position="left">基础信息</el-divider>
      <el-form-item label="网站标题">
        <el-input v-model="form.site_title" placeholder="例如：09 管理后台" />
      </el-form-item>
      <el-form-item label="管理员邮箱">
        <el-input v-model="form.admin_email" placeholder="admin@example.com" />
      </el-form-item>
      <el-form-item label="支持邮箱">
        <el-input v-model="form.support_email" placeholder="support@example.com" />
      </el-form-item>
      <el-form-item label="联系电话">
        <el-input v-model="form.contact_phone" placeholder="请输入联系电话" />
      </el-form-item>
      <el-form-item label="QQ">
        <el-input v-model="form.contact_qq" placeholder="请输入 QQ" />
      </el-form-item>
      <el-form-item label="Logo URL">
        <el-input v-model="form.logo_url" placeholder="https://..." />
      </el-form-item>
      <el-form-item label="Favicon URL">
        <el-input v-model="form.favicon_url" placeholder="https://..." />
      </el-form-item>
      <el-form-item label="允许注册">
        <el-switch v-model="form.allow_register" />
      </el-form-item>
      <el-form-item label="用户提交需要审核">
        <el-switch v-model="form.submission_need_review" />
        <span class="item-desc">开启后，用户提交资源需要管理员审核，审核通过后才会对外展示。</span>
      </el-form-item>
      <el-form-item label="用户提交自动转存">
        <el-switch v-model="form.submission_auto_transfer" />
        <span class="item-desc">开启后，用户提交资源时将自动触发转存，无需手动操作。</span>
      </el-form-item>
      <el-form-item label="详情页点击转存">
        <el-switch v-model="form.resource_detail_auto_transfer" />
        <span class="item-desc">开启后，用户在资源详情页点击“查看资源”时，后台才会触发自动转存。</span>
      </el-form-item>

      <el-divider content-position="left">号卡配置</el-divider>
      <el-form-item label="代理 user_id">
        <el-input v-model="form.haoka_user_id" placeholder="172号卡登录账号" />
      </el-form-item>
      <el-form-item label="接口 secret">
        <el-input v-model="form.haoka_secret" type="password" show-password placeholder="接口秘钥" />
      </el-form-item>
      <el-form-item label="定时同步">
        <el-switch v-model="form.haoka_sync_enabled" />
        <span class="item-desc">开启后，将定时同步号卡产品到本系统。</span>
      </el-form-item>
      <el-form-item label="同步间隔(秒)">
        <el-input-number v-model="form.haoka_sync_interval" :min="300" :max="86400" />
      </el-form-item>

      <el-divider content-position="left">前台号卡链接</el-divider>
      <el-form-item label="前台详情页跳转链接">
        <el-input v-model="form.haoka_order_url" placeholder="https://..." />
        <span class="item-desc">用于前台号卡详情页按钮跳转的链接（可配置）。</span>
      </el-form-item>
      <el-form-item label="号卡代理注册链接">
        <el-input v-model="form.haoka_agent_reg_url" placeholder="https://..." />
        <span class="item-desc">用于号卡代理注册的回调/接口地址（可配置）。</span>
      </el-form-item>

      <el-divider content-position="left">SEO 配置</el-divider>
      <el-form-item label="SEO 关键词">
        <el-input v-model="form.seo_keywords" placeholder="例如：资源,网盘,转存" />
      </el-form-item>
      <el-form-item label="SEO 描述">
        <el-input v-model="form.seo_description" type="textarea" :rows="3" />
      </el-form-item>
      <el-form-item label="ICP备案号">
        <el-input v-model="form.icp_record" />
      </el-form-item>
      <el-form-item label="页脚文字">
        <el-input v-model="form.footer_text" />
      </el-form-item>
      <el-form-item label="Clarity 项目 ID">
        <el-input v-model="form.clarity_project_id" placeholder="例如：abc123de45" />
        <span class="item-desc">启用后需要填写 Microsoft Clarity 的 Project ID，否则无法采集数据。</span>
      </el-form-item>
      <el-form-item label="启用 Clarity">
        <el-switch v-model="form.clarity_enabled" />
        <span class="item-desc">开启后会加载 Clarity；请确保已正确填写 Project ID。</span>
      </el-form-item>

      <el-divider content-position="left">前台首页 / 热榜</el-divider>
      <el-form-item label="显示豆瓣热榜导航">
        <el-switch v-model="form.douban_hot_nav_enabled" />
      </el-form-item>
      <el-form-item label="首页显示热门搜索">
        <el-switch v-model="form.hot_search_enabled" />
        <span class="item-desc">关闭后首页不再展示热搜词标签区。</span>
      </el-form-item>
      <el-form-item label="显示站点名称">
        <el-switch v-model="form.show_site_title" />
        <span class="item-desc">控制前台左上角品牌区是否显示网站名称（仅保留 Logo 时更简洁）。</span>
      </el-form-item>
      <el-form-item label="首页显示排行榜">
        <el-switch v-model="form.home_rank_board_enabled" />
        <span class="item-desc">含热门资源、最新资源、豆瓣热门；关闭后可减少首页请求。</span>
      </el-form-item>

      <el-divider content-position="left">豆瓣封面代理</el-divider>
      <el-form-item label="封面代理地址">
        <el-input
          v-model="form.douban_cover_proxy_url"
          type="textarea"
          :rows="2"
          placeholder="例如： https://image.baidu.com/search/down?url={url}"
        />
      </el-form-item>

      <el-divider content-position="left">TG 资源图片返代</el-divider>
      <el-form-item label="图片代理地址">
        <el-input
          v-model="form.tg_image_proxy_url"
          type="textarea"
          :rows="2"
          placeholder="例如：https://wsrv.nl/?url= 或 https://wsrv.nl/?url={url}"
        />
        <span class="item-desc">
          对来源为 Telegram 同步、且封面为 http(s) 外链的图片生效（本地已落盘到 /public/covers 的不走代理）。规则与上方豆瓣封面代理相同，支持
          <code>{url}</code> 或以 <code>url=</code> 结尾。
        </span>
      </el-form-item>

      <el-divider content-position="left">TMDB API</el-divider>
      <el-form-item label="TMDB Token">
        <el-input
          v-model="form.tmdb_bearer_token"
          type="password"
          show-password
          placeholder="填写 TMDB v4 Read Access Token"
        />
        <el-button
          class="ml8"
          type="primary"
          link
          @click="openTMDBApply"
        >
          申请 API
        </el-button>
        <span class="item-desc">用于前台搜索页展示 TMDB 影视信息卡，不填则不展示。</span>
      </el-form-item>
      <el-form-item label="TMDB 代理地址">
        <el-input
          v-model="form.tmdb_proxy_url"
          placeholder="可选，如 http://127.0.0.1:7890"
        />
        <span class="item-desc">当服务器无法直连 TMDB 时可填写；仅影响 TMDB 请求。</span>
      </el-form-item>
      <el-form-item label="IYUNS 接口地址">
        <el-input
          v-model="form.iyuns_api_base_url"
          placeholder="默认：https://api.iyuns.com"
        />
        <span class="item-desc">用于豆瓣检索（wpysso）和详情（dbys）接口基地址。</span>
      </el-form-item>

      <el-divider content-position="left">链接有效性检查</el-divider>
      <el-form-item label="自动删除无效链接">
        <el-switch v-model="form.auto_delete_invalid_links" />
        <span class="item-desc">开启后将自动删除 resources 中的无效链接，减少坏数据。</span>
      </el-form-item>
      <el-form-item label="搜索中隐藏无效链接">
        <el-switch v-model="form.hide_invalid_links_in_search" />
      </el-form-item>

      <el-divider content-position="left">Meilisearch 搜索</el-divider>
      <el-form-item label="启用 Meili 搜索">
        <el-switch v-model="form.meili_enabled" />
        <span class="item-desc">开启后：搜索优先走 Meilisearch；关闭则使用 MySQL。</span>
      </el-form-item>
      <el-form-item label="Meili URL">
        <el-input v-model="form.meili_url" placeholder="http://127.0.0.1:7700" />
      </el-form-item>
      <el-form-item label="Meili API Key">
        <el-input v-model="form.meili_api_key" type="password" show-password placeholder="可选：master key / API key" />
      </el-form-item>
      <el-form-item label="Meili Index">
        <el-input v-model="form.meili_index" placeholder="resources" />
        <el-button class="ml8" type="primary" plain :loading="meiliTesting" @click="testMeili">
          测试连接
        </el-button>
        <el-button class="ml8" type="warning" plain :loading="meiliReindexing" @click="reindexMeili">
          重建索引
        </el-button>
      </el-form-item>

      <el-divider content-position="left">友情链接</el-divider>
      <el-form-item label="友情链接列表">
        <div class="friend-links-editor">
          <div v-for="(row, idx) in form.friend_links" :key="idx" class="friend-row">
            <el-input v-model="row.title" placeholder="标题" class="friend-title" />
            <el-input v-model="row.url" placeholder="https://..." class="friend-url" />
            <el-button type="danger" text @click="removeFriend(idx)">删除</el-button>
          </div>
          <el-button type="primary" plain @click="addFriend">添加友情链接</el-button>
        </div>
      </el-form-item>
      <el-divider content-position="left">页脚快捷链接</el-divider>
      <el-form-item label="快捷链接列表">
        <div class="friend-links-editor">
          <div v-for="(row, idx) in form.footer_quick_links" :key="`fql-${idx}`" class="friend-row">
            <el-input v-model="row.title" placeholder="标题" class="friend-title" />
            <el-input v-model="row.url" placeholder="https://... 或 /search 或 #contact" class="friend-url" />
            <el-button type="danger" text @click="removeFooterQuickLink(idx)">删除</el-button>
          </div>
          <el-button type="primary" plain @click="addFooterQuickLink">添加快捷链接</el-button>
        </div>
      </el-form-item>

      <el-divider content-position="left">页脚热门网盘</el-divider>
      <el-form-item label="热门网盘列表">
        <div class="friend-links-editor">
          <div v-for="(name, idx) in form.footer_hot_platforms" :key="`fhp-${idx}`" class="friend-row">
            <el-input v-model="form.footer_hot_platforms![idx]" placeholder="例如：夸克网盘" class="friend-url" />
            <el-button type="danger" text @click="removeFooterHotPlatform(idx)">删除</el-button>
          </div>
          <el-button type="primary" plain @click="addFooterHotPlatform">添加热门网盘</el-button>
        </div>
      </el-form-item>

      <el-divider content-position="left">页脚社交媒体</el-divider>
      <el-form-item label="社交媒体列表">
        <div class="friend-links-editor">
          <div v-for="(row, idx) in form.footer_social_links" :key="`fsl-${idx}`" class="friend-row">
            <el-input v-model="row.title" placeholder="名称，如 Twitter" class="friend-title" />
            <el-input v-model="row.url" placeholder="https://...（可留空仅显示名称）" class="friend-url" />
            <el-button type="danger" text @click="removeFooterSocialLink(idx)">删除</el-button>
          </div>
          <el-button type="primary" plain @click="addFooterSocialLink">添加社交媒体</el-button>
        </div>
      </el-form-item>

      <el-form-item label="微信公众号">
        <el-input v-model="form.footer_wechat" placeholder="例如：网盘资源导航" />
      </el-form-item>

      <el-divider content-position="left">SMTP 邮件</el-divider>
      <el-form-item label="SMTP Host">
        <el-input v-model="form.smtp_host" placeholder="smtp.example.com" />
      </el-form-item>
      <el-form-item label="SMTP Port">
        <el-input-number v-model="form.smtp_port" :min="1" :max="65535" />
      </el-form-item>
      <el-form-item label="SMTP 用户名">
        <el-input v-model="form.smtp_user" />
      </el-form-item>
      <el-form-item label="SMTP 密码">
        <el-input v-model="form.smtp_pass" type="password" show-password />
      </el-form-item>
      <el-form-item label="发件人邮箱">
        <el-input v-model="form.smtp_from" />
      </el-form-item>

      <el-divider content-position="left">Telegram 机器人</el-divider>
      <el-form-item label="全局 Bot Token">
        <el-input v-model="form.tg_bot_token" placeholder="例如：123456:ABCDEF..." />
      </el-form-item>
      <el-form-item label="TG 代理">
        <el-input v-model="form.tg_proxy_url" placeholder="http://127.0.0.1:7890 或 socks5://127.0.0.1:1080" />
      </el-form-item>
      <el-form-item label="TG API ID">
        <el-input-number v-model="form.tg_api_id" :min="0" placeholder="请输入 TG API ID" />
      </el-form-item>
      <el-form-item label="TG API HASH">
        <el-input v-model="form.tg_api_hash" placeholder="例如：0123456789abcdef0123456789abcdef" />
      </el-form-item>
      <el-form-item label="TG Session">
        <el-input v-model="form.tg_session" type="textarea" :rows="3" placeholder="可选：MTProto 登录会话字符串" />
      </el-form-item>
      <el-form-item label="盘查服务地址">
        <el-input v-model="form.pancheck_base_url" placeholder="https://pancheck.116818.xyz" />
      </el-form-item>
      <el-form-item label="链接校验">
        <el-switch v-model="form.link_check_enabled" />
      </el-form-item>
      <el-form-item label="链接校验间隔(秒)">
        <el-input-number v-model="form.link_check_interval" :min="60" :max="86400" />
      </el-form-item>
      <el-alert
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 12px"
        title="提示：保存 Cookie / Token 后将用于请求外部服务，请确认填写正确。"
      />

      <el-form-item>
        <el-button type="primary" :loading="saving" @click="save" v-permission="['config:update']">
          保存配置
        </el-button>
        <el-button @click="load">加载配置</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<script setup lang="ts">
import {
  getSystemConfig,
  type IFriendLinkItem,
  type ISystemConfig,
  meiliReindex,
  meiliTest,
  updateSystemConfig,
} from '@/api/systemConfig'
import { loadRuntimeConfig, runtimeConfig } from '@/config/runtimeConfig'

defineOptions({ name: 'ConfigView' })

const saving = ref(false)
const meiliTesting = ref(false)
const meiliReindexing = ref(false)
const form = reactive<ISystemConfig>({
  site_title: '',
  admin_email: '',
  support_email: '',
  contact_phone: '',
  contact_qq: '',
  logo_url: '',
  favicon_url: '',
  seo_keywords: '',
  seo_description: '',
  icp_record: '',
  footer_text: '',
  clarity_project_id: '',
  clarity_enabled: false,
  friend_links: [] as IFriendLinkItem[],
  footer_quick_links: [] as IFriendLinkItem[],
  footer_hot_platforms: [] as string[],
  footer_social_links: [] as IFriendLinkItem[],
  footer_wechat: '',
  allow_register: true,
  submission_need_review: true,
  submission_auto_transfer: false,
  resource_detail_auto_transfer: false,
  haoka_user_id: '',
  haoka_secret: '',
  haoka_sync_enabled: false,
  haoka_sync_interval: 3600,
  haoka_order_url: '',
  haoka_agent_reg_url: '',
  smtp_host: '',
  smtp_port: 25,
  smtp_user: '',
  smtp_pass: '',
  smtp_from: '',
  tg_bot_token: '',
  tg_proxy_url: '',
  tg_api_id: 0,
  tg_api_hash: '',
  tg_session: '',
  pancheck_base_url: '',
  link_check_enabled: false,
  link_check_interval: 3600,
  douban_hot_nav_enabled: false,
  hot_search_enabled: true,
  show_site_title: true,
  home_rank_board_enabled: true,
  douban_cover_proxy_url: '',
  tg_image_proxy_url: '',
  tmdb_bearer_token: '',
  tmdb_proxy_url: '',
  iyuns_api_base_url: 'https://api.iyuns.com',
  auto_delete_invalid_links: false,
  hide_invalid_links_in_search: false,
  meili_enabled: false,
  meili_url: 'http://127.0.0.1:7700',
  meili_api_key: '',
  meili_index: 'resources',
})

const openTMDBApply = () => {
  window.open('https://www.themoviedb.org/settings/api', '_blank', 'noopener,noreferrer')
}

const load = async () => {
  const { data: res } = await getSystemConfig()
  if (res.code !== 200) return
  Object.assign(form, res.data)
  if (!Array.isArray(form.friend_links)) {
    form.friend_links = []
  }
  if (!Array.isArray(form.footer_quick_links)) {
    form.footer_quick_links = []
  }
  if (!Array.isArray(form.footer_hot_platforms)) {
    form.footer_hot_platforms = []
  }
  if (!Array.isArray(form.footer_social_links)) {
    form.footer_social_links = []
  }
}

const testMeili = async () => {
  meiliTesting.value = true
  try {
    const { data: res } = await meiliTest()
    if (res.code !== 200) return
    if (res.data?.ok) ElMessage.success(res.data?.message || '连接正常')
    else ElMessage.warning(res.data?.message || '连接失败')
  } finally {
    meiliTesting.value = false
  }
}

const reindexMeili = async () => {
  meiliReindexing.value = true
  try {
    const { data: res } = await meiliReindex(500)
    if (res.code !== 200) return
    ElMessage.success(`已提交重建：indexed=${res.data?.indexed ?? 0} / total=${res.data?.total ?? 0}`)
  } finally {
    meiliReindexing.value = false
  }
}

const addFriend = () => {
  form.friend_links!.push({ title: '', url: '' })
}

const removeFriend = (idx: number) => {
  form.friend_links!.splice(idx, 1)
}

const addFooterQuickLink = () => {
  form.footer_quick_links!.push({ title: '', url: '' })
}

const removeFooterQuickLink = (idx: number) => {
  form.footer_quick_links!.splice(idx, 1)
}

const addFooterHotPlatform = () => {
  form.footer_hot_platforms!.push('')
}

const removeFooterHotPlatform = (idx: number) => {
  form.footer_hot_platforms!.splice(idx, 1)
}

const addFooterSocialLink = () => {
  form.footer_social_links!.push({ title: '', url: '' })
}

const removeFooterSocialLink = (idx: number) => {
  form.footer_social_links!.splice(idx, 1)
}

const save = async () => {
  saving.value = true
  try {
    const payload: ISystemConfig = {
      ...form,
      friend_links: [...(form.friend_links || [])],
      footer_quick_links: [...(form.footer_quick_links || [])],
      footer_hot_platforms: [...(form.footer_hot_platforms || [])],
      footer_social_links: [...(form.footer_social_links || [])],
      tg_api_id: Number(form.tg_api_id) || 0,
    }
    const { data: res } = await updateSystemConfig(payload)
    if (res.code !== 200) return
    ElMessage.success('保存成功')
    Object.assign(runtimeConfig, {
      siteTitle: form.site_title || runtimeConfig.siteTitle,
      logoUrl: form.logo_url || runtimeConfig.logoUrl,
      faviconUrl: form.favicon_url || runtimeConfig.faviconUrl,
      footerText: form.footer_text || runtimeConfig.footerText,
      clarityProjectId: form.clarity_project_id || '',
      clarityEnabled: form.clarity_enabled ?? false,
      seoKeywords: form.seo_keywords || '',
      seoDescription: form.seo_description || '',
      icpRecord: form.icp_record || '',
      allowRegister: form.allow_register,
      supportEmail: form.support_email || '',
      contactPhone: form.contact_phone || '',
      friendLinks: [...(form.friend_links || [])],
      footerQuickLinks: [...(form.footer_quick_links || [])],
      footerHotPlatforms: [...(form.footer_hot_platforms || [])],
      footerSocialLinks: [...(form.footer_social_links || [])],
      footerWechat: form.footer_wechat || '',
      doubanHotNavEnabled: form.douban_hot_nav_enabled ?? false,
      hotSearchEnabled: form.hot_search_enabled ?? true,
      showSiteTitle: form.show_site_title ?? true,
      homeRankBoardEnabled: form.home_rank_board_enabled ?? true,
      doubanCoverProxyUrl: form.douban_cover_proxy_url || '',
      tgImageProxyUrl: form.tg_image_proxy_url || '',
      haokaOrderUrl: form.haoka_order_url || runtimeConfig.haokaOrderUrl,
      haokaAgentRegUrl: form.haoka_agent_reg_url || runtimeConfig.haokaAgentRegUrl,
    })
    await loadRuntimeConfig()
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped lang="scss">
.header {
  font-weight: 600;
}

.config-form {
  max-width: 900px;
}

.friend-links-editor {
  width: 100%;
}
.friend-row {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}
.friend-title {
  flex: 0 0 160px;
}
.friend-url {
  flex: 1;
  min-width: 0;
}
.ml8 {
  margin-left: 8px;
}

.item-desc {
  margin-left: 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
