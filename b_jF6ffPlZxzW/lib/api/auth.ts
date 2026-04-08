import { apiGet, apiPost } from "@/lib/api/client"

export type CaptchaDTO = {
  captcha_id: string
  svg: string
  expires_at: string
}

export type LoginResponse = {
  token: string
  user: { id: number | string; username: string; email?: string }
}

export async function fetchCaptcha() {
  return apiGet<CaptchaDTO>("/auth/captcha")
}

export async function sendRegisterEmailCode(input: { email: string; captcha_id: string; captcha_code: string }) {
  return apiPost<{ expires_at: string }>("/auth/register/send-code", input)
}

export async function registerUser(input: {
  username: string
  email: string
  password: string
  email_code: string
  captcha_id: string
  captcha_code: string
}) {
  return apiPost<{ id: number; username: string; email: string }>("/auth/register", input)
}

export async function loginUser(input: { username: string; password: string }) {
  return apiPost<LoginResponse>("/auth/login", input)
}

