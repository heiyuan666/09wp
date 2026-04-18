/** 将后端返回的 /public/... 路径补全为可访问绝对地址 */
export function resolveAssetUrl(path: string | undefined, apiBase: string): string {
  const p = String(path || '').trim()
  if (!p) return '/icon.svg'
  if (/^https?:\/\//i.test(p)) return p
  const origin = apiBase.replace(/\/$/, '')
  return p.startsWith('/') ? `${origin}${p}` : `${origin}/${p}`
}
