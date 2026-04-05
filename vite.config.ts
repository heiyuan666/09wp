import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath, URL } from 'node:url'
import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import react from '@vitejs/plugin-react'
import vueDevTools from 'vite-plugin-vue-devtools'
import tailwindcss from '@tailwindcss/vite'
import UnoCSS from 'unocss/vite'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import AppLoading from 'vite-plugin-app-loading'

const pkg = JSON.parse(readFileSync(fileURLToPath(new URL('./package.json', import.meta.url)), 'utf-8')) as {
  version: string
}

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const apiBaseURL = env.VITE_API_BASE_URL || ''
  // 从 VITE_API_BASE_URL 推导后端根地址：例如 http://localhost:8080/api/v1 -> http://localhost:8080
  const backendRootURL = apiBaseURL.startsWith('http')
    ? apiBaseURL.replace(/\/api\/v1\/?$/, '')
    : 'http://localhost:8080'
  return {
    base: env.VITE_STATIC_URL || '/',
    define: {
      __APP_VERSION__: JSON.stringify(pkg.version),
    },
    plugins: [
      vue(),
      react(),
      UnoCSS(),
      tailwindcss(),
      vueDevTools(),
      AppLoading(),
      AutoImport({
        imports: ['vue', 'vue-router', 'pinia'],
        dirs: ['src/stores'],
        resolvers: [ElementPlusResolver()],
      }),
      Components({
        resolvers: [ElementPlusResolver()],
        dirs: ['src/components'], // 指定组件目录,注册为全局组件
      }),
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
        // @douyinfe/semi-ui 的 package exports 未包含 dist/css，首页按需引入完整样式
        '@semi-ui-styles/semi.min.css': path.resolve(
          fileURLToPath(new URL('.', import.meta.url)),
          'node_modules/@douyinfe/semi-ui/dist/css/semi.min.css',
        ),
      },
    },
    server: {
      host: '0.0.0.0',
      port: 3007,
      open: true,
      proxy: {
        // 让 crawlers/浏览器直接访问站点根目录的 sitemap/rss
        '/sitemap.xml': {
          target: backendRootURL,
          changeOrigin: true,
        },
        '/rss.xml': {
          target: backendRootURL,
          changeOrigin: true,
        },
      },
    },
  }
})
