import { headers } from "next/headers"

function trimTrailingSlash(s: string) {
  return s.replace(/\/+$/, "")
}

export async function getSiteOrigin() {
  const fromEnv = process.env.NEXT_PUBLIC_SITE_ORIGIN || process.env.SITE_ORIGIN
  if (fromEnv && fromEnv.trim()) return trimTrailingSlash(fromEnv.trim())

  const h = await headers()
  const host = h.get("x-forwarded-host") || h.get("host") || "localhost:3000"
  const proto = h.get("x-forwarded-proto") || "http"
  return `${proto}://${host}`
}

