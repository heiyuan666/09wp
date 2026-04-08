export type SharePayload = {
  title: string
  text?: string
  url: string
}

export async function tryNativeShare(payload: SharePayload): Promise<boolean> {
  const url = (payload.url || "").trim()
  if (!url) return false
  try {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const navAny: any = typeof navigator !== "undefined" ? navigator : null
    if (!navAny?.share) return false
    await navAny.share({
      title: payload.title || (typeof document !== "undefined" ? document.title : "分享"),
      text: payload.text || "",
      url,
    })
    return true
  } catch {
    return false
  }
}

export async function copyTextToClipboard(text: string): Promise<boolean> {
  const t = (text || "").trim()
  if (!t) return false
  try {
    await navigator.clipboard.writeText(t)
    return true
  } catch {
    return false
  }
}

