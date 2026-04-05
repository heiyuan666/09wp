import '@/styles/index.css'
import 'virtual:uno.css'
import 'animate.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import 'element-plus/dist/index.css'
import 'nprogress/nprogress.css'
import { APP_CONFIG } from '@/config/app.config'
import { loadRuntimeConfig } from '@/config/runtimeConfig'
import { loadingFadeOut } from 'virtual:app-loading'
import { worker } from '@/mocks/browser'
import { initData } from '@/mocks/db/initData'
import { permissionDirective } from '@/directives/permission'
import { MotionPlugin } from '@vueuse/motion'
import VXETablePlugin from '@/plugins/vxeTable'
import '@/plugins/echarts'
import { createApp, nextTick } from 'vue'
import { createPinia } from 'pinia'
import App from '@/App.vue'
import router from '@/router/index'

const startMocksIfEnabled = async () => {
  if (!APP_CONFIG.enableMSW) return

  await worker.start({
    serviceWorker: {
      url: `${import.meta.env.VITE_STATIC_URL}mockServiceWorker.js`,
    },
    onUnhandledRequest(req, print) {
      if (req.url.includes(APP_CONFIG.listenMSWPath)) {
        print.warning()
      }
    },
  })

  await initData()
}

const startApp = async () => {
  await startMocksIfEnabled()
  await loadRuntimeConfig()

  const app = createApp(App)
  VXETablePlugin(app)

  app.use(createPinia())
  app.use(router)
  app.use(MotionPlugin)
  app.directive('permission', permissionDirective)

  app.mount('#app')
  await router.isReady()
  await nextTick()
  loadingFadeOut()
}

startApp().catch((error) => {
  console.error('Failed to start app:', error)
})
