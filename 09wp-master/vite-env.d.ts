/** 构建时由 vite.config 从 package.json 注入 */
declare const __APP_VERSION__: string

interface ImportMetaEnv {
  // 接口基础URL
  readonly VITE_API_BASE_URL: string
  // 静态资源URL
  readonly VITE_STATIC_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
