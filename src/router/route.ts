/**
 * name: 路由名称, 也用于keepAlive缓存，需要与组件名称一致
 * meta.keepAlive: 是否需要缓存
 *
 */
export const staticRoutes = [
  // 前台公开站点（无需登录）
  {
    path: '/',
    name: 'PublicLayout',
    component: () => import('@/layouts/public.vue'),
    meta: { public: true, keepAlive: false },
    children: [
      {
        path: '',
        name: 'PublicHomeView',
        component: () => import('@/views/public/home/index.vue'),
        meta: { public: true, keepAlive: false },
      },
      {
        path: 'search',
        name: 'PublicSearchView',
        component: () => import('@/views/public/search/index.vue'),
        meta: { public: true, keepAlive: false },
      },
      {
        path: 'tag/:tag',
        name: 'PublicTagView',
        component: () => import('@/views/public/tag/index.vue'),
        meta: { public: true, keepAlive: false },
      },
      {
        path: 'games',
        name: 'PublicGameHomeView',
        component: () => import('@/views/public/game/index.vue'),
        meta: { public: true, keepAlive: false },
      },
      {
        path: 'games/:id',
        name: 'PublicGameDetailView',
        component: () => import('@/views/public/game/detail.react-host.vue'),
        meta: { public: true, keepAlive: false },
      },
      {
        path: 'docs/netdisk-api',
        name: 'PublicNetdiskApiDocView',
        component: () => import('@/views/public/docs/netdisk-api.vue'),
        meta: { public: true, keepAlive: false },
      },
      {
        path: 'c/:slug',
        name: 'PublicCategoryView',
        component: () => import('@/views/public/category/index.vue'),
        meta: { public: true, keepAlive: false },
      },
      {
        path: 'r/:id',
        name: 'PublicResourceDetailView',
        component: () => import('@/views/public/resource/index.vue'),
        meta: { public: true, keepAlive: false },
      },
      {
        path: 'haoka',
        name: 'PublicHaokaView',
        component: () => import('@/views/public/haoka/index.vue'),
        meta: { public: true, keepAlive: false },
      },
      {
        path: 'haoka/:id',
        name: 'PublicHaokaDetailView',
        component: () => import('@/views/public/haoka/detail.vue'),
        meta: { public: true, keepAlive: false },
      },
      {
        path: 'u/login/qr',
        name: 'PublicUserQrLoginView',
        component: () => import('@/views/public/auth/qrLogin.vue'),
        meta: { public: true, keepAlive: false, hidden: true },
      },
      {
        path: 'u/login',
        name: 'PublicUserLoginView',
        component: () => import('@/views/public/auth/login.vue'),
        meta: { public: true, keepAlive: false, hidden: true },
      },
      {
        path: 'u/register',
        name: 'PublicUserRegisterView',
        component: () => import('@/views/public/auth/register.vue'),
        meta: { public: true, keepAlive: false, hidden: true },
      },
      {
        path: 'u/me',
        name: 'PublicMeView',
        component: () => import('@/views/public/me/index.vue'),
        meta: { public: true, keepAlive: false },
      },
    ],
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/login/index.vue'),
    meta: { keepAlive: false },
  },
  // 重定向路由(暂时注释掉，因为redirect路由会导致加载缓慢)
  // {
  //   path: '/redirect/:path(.*)',
  //   name: 'redirect',
  //   component: () => import('@/views/redirect/index.vue'),
  //   meta: { hidden: true },
  // },
  {
    path: '/admin',
    name: 'layout',
    component: () => import('@/layouts/index.vue'),
    children: [
      {
        path: '',
        name: 'AdminIndexRedirect',
        redirect: '/admin/dashboard',
        meta: { title: '控制台', icon: 'HOutline:Squares2X2Icon', hidden: true },
      },
      {
        path: 'dashboard',
        name: 'DashboardView',
        component: () => import('@/views/dashboard/index.vue'),
        meta: { title: '控制台', icon: 'HOutline:Squares2X2Icon', keepAlive: true },
      },
      {
        path: '/netdisk/submissions',
        name: 'NetdiskSubmissionsView',
        component: () => import('@/views/netdisk/submissions/index.vue'),
        meta: { title: '用户提交审核', icon: 'HOutline:ClipboardDocumentCheckIcon', keepAlive: true },
      },
      {
        path: '/game/reviews',
        name: 'GameReviewsView',
        component: () => import('@/views/game/reviews/index.vue'),
        meta: { title: '游戏评论', icon: 'HOutline:ChatBubbleLeftRightIcon', keepAlive: false },
      },
      {
        path: '/game/feedbacks',
        name: 'GameFeedbacksView',
        component: () => import('@/views/game/feedbacks/index.vue'),
        meta: { title: '资源失效反馈', icon: 'HOutline:ExclamationTriangleIcon', keepAlive: false },
      },
      // 游戏管理 - 站点设置（复用系统配置页）
      {
        path: '/game/settings',
        name: 'GameSiteSettingsView',
        component: () => import('@/views/game/settings/index.vue'),
        meta: { title: '站点设置', icon: 'HOutline:AdjustmentsHorizontalIcon', keepAlive: false },
      },
      // 游戏管理 - 导航栏设置（复用导航菜单管理页）
      {
        path: '/game/nav-menu',
        name: 'GameNavMenuSettingsView',
        component: () => import('@/views/game/nav-menu/index.vue'),
        meta: { title: '导航栏设置', icon: 'HOutline:Bars3BottomLeftIcon', keepAlive: false },
      },
      {
        path: '/system/keyword-blocks',
        name: 'Keyword-blocksView',
        component: () => import('@/views/system/keyword-blocks/index.vue'),
        meta: { title: '关键词屏蔽', icon: 'HOutline:ShieldCheckIcon', keepAlive: true },
      },
      {
        path: '/system/cleanup-logs',
        name: 'SystemCleanupLogsView',
        component: () => import('@/views/system/cleanup-logs/index.vue'),
        meta: { title: '清理任务日志', icon: 'HOutline:ClipboardDocumentListIcon', keepAlive: false },
      },
      {
        path: '/system/global-search',
        name: 'SystemGlobalSearchView',
        component: () => import('@/views/system/global-search/index.vue'),
        meta: { title: '全网搜接口', icon: 'HOutline:GlobeAsiaAustraliaIcon', keepAlive: false },
      },
      {
        path: '/profile',
        name: 'ProfileView',
        component: () => import('@/views/profile/index.vue'),
        meta: { title: '个人中心', icon: 'HOutline:UserCircleIcon', keepAlive: true },
      },
      {
        path: '/exception/403',
        name: '403',
        component: () => import('@/views/exception/403/index.vue'),
        meta: { title: '403', icon: 'HOutline:NoSymbolIcon', keepAlive: true },
      },
      {
        path: '/:pathMatch(.*)*',
        name: '404',
        component: () => import('@/views/exception/404/index.vue'),
        meta: { title: '404', icon: 'HOutline:QuestionMarkCircleIcon', keepAlive: true },
      },
    ],
  },
]
