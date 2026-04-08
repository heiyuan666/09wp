export type ApiEnvelope<T> = {
  code: number
  message: string
  data: T
}

export class ApiError extends Error {
  public readonly code: number
  constructor(message: string, code: number) {
    super(message)
    this.code = code
  }
}

function trimTrailingSlash(s: string) {
  return s.replace(/\/+$/, "")
}

function trimLeadingSlash(s: string) {
  return s.replace(/^\/+/, "")
}

export function getApiBaseUrl() {
  const fromEnv = process.env.NEXT_PUBLIC_API_BASE_URL
  if (fromEnv && fromEnv.trim()) return trimTrailingSlash(fromEnv.trim())
  return "http://localhost:8080/api/v1"
}

export function toAbsoluteUrl(pathOrUrl: string, origin: string) {
  const raw = (pathOrUrl ?? "").trim()
  if (!raw) return ""
  if (raw.startsWith("http://") || raw.startsWith("https://")) return raw
  return `${trimTrailingSlash(origin)}/${trimLeadingSlash(raw)}`
}

export async function apiGet<T>(path: string, init?: RequestInit): Promise<T> {
  const baseUrl = getApiBaseUrl()
  const url = `${baseUrl}${path.startsWith("/") ? "" : "/"}${path}`

  const res = await fetch(url, {
    ...init,
    method: "GET",
    headers: {
      Accept: "application/json",
      ...(init?.headers ?? {}),
    },
    cache: "no-store",
  })

  if (!res.ok) {
    throw new ApiError(`HTTP ${res.status}`, res.status)
  }

  const json = (await res.json()) as ApiEnvelope<T>
  if (!json || typeof json.code !== "number") {
    throw new ApiError("后端返回格式不正确", 500)
  }
  if (json.code !== 200) {
    throw new ApiError(json.message || "请求失败", json.code)
  }
  return json.data
}

