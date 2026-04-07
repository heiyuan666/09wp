/**
 * 应用全局配置
 * 用户可以在这里自定义项目的各种配置项
 */

/** 未配置 VITE_API_BASE_URL 时使用 /api/v1，避免 fetch('/public/...') 与 Vite public 目录冲突 */
export const API_BASE_URL = String(import.meta.env.VITE_API_BASE_URL || '/api/v1').replace(/\/+$/, '')

export const APP_CONFIG = {
  // 是否启用 MSW
  enableMSW: false,
  // MSW 监听的请求路径
  listenMSWPath: '/DFAN-admin-api',

  // 项目名称
  name: '网盘资源',

  // Favicon src - 根据环境动态设置 base path
  faviconSrc: `${import.meta.env.VITE_STATIC_URL}favicon.ico`,

  // Logo src
  logoSrc: new URL('@/assets/logo.svg', import.meta.url).href,

  // 是否展示主题配置
  showThemeConfig: true,
}
