/**
 * 迅雷下载 JS-SDK：https://open.thunderurl.com/thunder-link.js
 * 用于 magnet / ed2k 等协议唤起迅雷客户端。
 */

declare global {
  interface Window {
    thunderLink?: {
      newTask: (opts: ThunderNewTaskOptions) => void
    }
  }
}

export type ThunderTaskItem = {
  url: string
  name?: string
  size?: number
  dir?: string
}

export type ThunderNewTaskOptions = {
  downloadDir?: string
  taskGroupName?: string
  tasks: ThunderTaskItem[]
  excludePath?: string
  installFile?: string
  threadCount?: string
  referer?: string
  userAgent?: string
}

const SCRIPT_SRC = '//open.thunderurl.com/thunder-link.js'
const SCRIPT_ATTR = 'data-thunder-link-sdk'

export function isMagnetUrl(raw: string): boolean {
  return /^magnet:\?/i.test(String(raw || '').trim())
}

export function isEd2kUrl(raw: string): boolean {
  return /^ed2k:\/\//i.test(String(raw || '').trim())
}

export function isThunderFriendlyUrl(raw: string): boolean {
  const u = String(raw || '').trim()
  if (!u) return false
  return isMagnetUrl(u) || isEd2kUrl(u) || /^thunder:\/\//i.test(u)
}

/** 从 magnet 的 dn 参数尽量解析展示文件名 */
export function guessNameFromMagnet(magnet: string): string | undefined {
  const s = String(magnet || '').trim()
  if (!isMagnetUrl(s)) return undefined
  try {
    const idx = s.indexOf('?')
    const qs = idx >= 0 ? s.slice(idx + 1) : ''
    const params = new URLSearchParams(qs)
    const dn = params.get('dn')
    if (dn) {
      return decodeURIComponent(dn.replace(/\+/g, ' ')).slice(0, 200)
    }
  } catch {
    /* ignore */
  }
  return undefined
}

function loadScriptOnce(): Promise<void> {
  if (typeof window === 'undefined') return Promise.resolve()
  if (window.thunderLink && typeof window.thunderLink.newTask === 'function') {
    return Promise.resolve()
  }
  const existing = document.querySelector<HTMLScriptElement>(`script[${SCRIPT_ATTR}]`)
  if (existing) {
    return new Promise((resolve, reject) => {
      const deadline = Date.now() + 15000
      const tick = () => {
        if (window.thunderLink && typeof window.thunderLink.newTask === 'function') {
          resolve()
          return
        }
        if (Date.now() > deadline) {
          reject(new Error('迅雷 JS-SDK 加载超时'))
          return
        }
        window.setTimeout(tick, 50)
      }
      tick()
    })
  }
  return new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.src = SCRIPT_SRC
    script.async = true
    script.setAttribute(SCRIPT_ATTR, '1')
    script.onload = () => {
      const deadline = Date.now() + 10000
      const tick = () => {
        if (window.thunderLink && typeof window.thunderLink.newTask === 'function') {
          resolve()
          return
        }
        if (Date.now() > deadline) {
          reject(new Error('迅雷 JS-SDK 未就绪'))
          return
        }
        window.setTimeout(tick, 30)
      }
      tick()
    }
    script.onerror = () => reject(new Error('迅雷 JS-SDK 脚本加载失败'))
    document.head.appendChild(script)
  })
}

/**
 * 使用迅雷客户端创建下载任务（单条）。
 */
export async function thunderDownloadSingle(
  url: string,
  options?: { name?: string; size?: number; downloadDir?: string },
): Promise<void> {
  const u = String(url || '').trim()
  if (!u) throw new Error('下载地址为空')
  await loadScriptOnce()
  const tl = window.thunderLink
  if (!tl || typeof tl.newTask !== 'function') {
    throw new Error('迅雷 JS-SDK 不可用')
  }
  const name = options?.name?.trim() || guessNameFromMagnet(u)
  tl.newTask({
    downloadDir: options?.downloadDir || '全网搜',
    tasks: [{ url: u, name: name || undefined, size: options?.size }],
  })
}
