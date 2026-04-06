/** 资源是否应按「TG 图片返代」处理（与后台 source / external_id / 封面域一致） */
export function shouldApplyTgCoverProxy(
  resource: Record<string, unknown> | null | undefined,
  coverRaw: string,
): boolean {
  if (!resource) return false
  const src = String(resource.source ?? '').trim().toLowerCase()
  if (src === 'telegram') return true
  const ext = String(resource.external_id ?? '').trim().toLowerCase()
  if (ext.startsWith('tg:')) return true
  const u = coverRaw.trim().toLowerCase()
  if (!u) return false
  // 常见 TG 网页/图片 CDN（历史数据可能未写 source）
  if (u.includes('telesco.pe')) return true
  if (u.includes('telegram-cdn.org')) return true
  if (u.includes('cdn.telegram.org')) return true
  if (u.includes('telegram.org/file')) return true
  return false
}

/**
 * 按系统配置的返代模板拼接图片地址（豆瓣封面、TG 外链图等）。
 * 支持 `{url}`、以 `url=` 结尾、或自动追加 `?url=` / `&url=`。
 */
export function buildProxiedImageSrc(cover: string | undefined | null, template: string): string {
  const c = String(cover || '').trim()
  if (!c) return ''

  const tmpl = String(template || '').trim()
  if (!tmpl) return c

  const enc = encodeURIComponent(c)

  if (tmpl.includes('{url}')) return tmpl.replace(/\{url\}/g, enc)

  if (tmpl.endsWith('url=')) return tmpl + enc

  const idx = tmpl.indexOf('url=')
  if (idx >= 0) {
    const prefix = tmpl.slice(0, idx + 'url='.length)
    return prefix + enc
  }

  if (tmpl.includes('?')) {
    if (tmpl.endsWith('?') || tmpl.endsWith('&')) return tmpl + 'url=' + enc
    return tmpl + '&url=' + enc
  }
  return tmpl.replace(/\/?$/, '') + '?url=' + enc
}
