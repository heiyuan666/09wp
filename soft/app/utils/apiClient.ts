/** 统一请求 /api/v1，解析 { code, data, message }（业务错误时 HTTP 仍可能为 200） */
export async function apiFetch<T>(path: string, opts?: { query?: Record<string, string | number | undefined> }) {
  const config = useRuntimeConfig()
  const base = String(config.public.apiBase || '').replace(/\/$/, '')
  const p = path.startsWith('/') ? path : `/${path}`
  const url = `${base}/api/v1${p}`
  const q = opts?.query
  const res = await $fetch<{ code: number; message?: string; data: T }>(url, {
    query: q as Record<string, string | number | boolean | undefined>,
  })
  if (res.code !== 200) {
    const code = res.code >= 400 && res.code < 600 ? res.code : 500
    throw createError({
      statusCode: code,
      statusMessage: res.message || '请求失败',
    })
  }
  return res.data
}
