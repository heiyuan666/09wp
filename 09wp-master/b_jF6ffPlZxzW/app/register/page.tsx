"use client"

import { useEffect, useMemo, useState } from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { Header } from "@/components/game/header"
import { Footer } from "@/components/home/footer"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { fetchCaptcha, registerUser, sendRegisterEmailCode } from "@/lib/api/auth"

export default function RegisterPage() {
  const router = useRouter()
  const [username, setUsername] = useState("")
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [captchaId, setCaptchaId] = useState("")
  const [captchaSvg, setCaptchaSvg] = useState("")
  const [captchaCode, setCaptchaCode] = useState("")
  const [emailCode, setEmailCode] = useState("")

  const [loading, setLoading] = useState(false)
  const [sending, setSending] = useState(false)
  const [error, setError] = useState("")
  const [hint, setHint] = useState("")

  const captchaSvgHtml = useMemo(() => ({ __html: captchaSvg }), [captchaSvg])

  async function reloadCaptcha() {
    setError("")
    try {
      const cap = await fetchCaptcha()
      setCaptchaId(cap.captcha_id)
      setCaptchaSvg(cap.svg)
      setCaptchaCode("")
    } catch (err) {
      setError(err instanceof Error ? err.message : "获取验证码失败")
    }
  }

  async function onSendCode() {
    setError("")
    setHint("")
    setSending(true)
    try {
      if (!captchaId || !captchaSvg) await reloadCaptcha()
      await sendRegisterEmailCode({
        email: email.trim(),
        captcha_id: captchaId,
        captcha_code: captchaCode.trim(),
      })
      setHint("验证码已发送，请查收邮箱（10 分钟内有效）")
    } catch (err) {
      setError(err instanceof Error ? err.message : "发送失败")
    } finally {
      setSending(false)
    }
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError("")
    setHint("")
    setLoading(true)
    try {
      if (!captchaId || !captchaSvg) await reloadCaptcha()
      await registerUser({
        username: username.trim(),
        email: email.trim(),
        password,
        email_code: emailCode.trim(),
        captcha_id: captchaId,
        captcha_code: captchaCode.trim(),
      })
      router.push("/login")
    } catch (err) {
      setError(err instanceof Error ? err.message : "注册失败")
    } finally {
      setLoading(false)
    }
  }

  // load captcha once
  useEffect(() => {
    reloadCaptcha()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div className="min-h-screen bg-background">
      <Header />
      <main className="mx-auto max-w-md px-4 py-14 sm:px-6">
        <div className="rounded-xl border border-border bg-card p-6">
          <h1 className="text-2xl font-semibold text-foreground">注册</h1>
          <p className="mt-2 text-sm text-muted-foreground">注册需要图形验证码 + 邮箱验证码</p>

          <form className="mt-6 space-y-4" onSubmit={onSubmit}>
            <div className="space-y-2">
              <div className="text-sm font-medium text-foreground">用户名</div>
              <Input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="请输入用户名" />
            </div>
            <div className="space-y-2">
              <div className="text-sm font-medium text-foreground">邮箱</div>
              <Input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="请输入邮箱" />
            </div>
            <div className="space-y-2">
              <div className="text-sm font-medium text-foreground">密码</div>
              <Input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="至少 6 位" />
            </div>

            <div className="space-y-2">
              <div className="text-sm font-medium text-foreground">图形验证码</div>
              <div className="flex items-center gap-3">
                <div
                  className="h-10 w-[120px] rounded-md border border-border bg-background cursor-pointer"
                  title="点击刷新"
                  onClick={reloadCaptcha}
                  dangerouslySetInnerHTML={captchaSvgHtml}
                />
                <Input
                  value={captchaCode}
                  onChange={(e) => setCaptchaCode(e.target.value)}
                  placeholder="输入图形验证码"
                />
              </div>
              <div className="text-xs text-muted-foreground">看不清？点击图片刷新</div>
            </div>

            <div className="space-y-2">
              <div className="text-sm font-medium text-foreground">邮箱验证码</div>
              <div className="flex items-center gap-3">
                <Input value={emailCode} onChange={(e) => setEmailCode(e.target.value)} placeholder="输入邮箱验证码" />
                <Button type="button" variant="secondary" onClick={onSendCode} disabled={sending}>
                  {sending ? "发送中..." : "发送验证码"}
                </Button>
              </div>
            </div>

            {hint ? <div className="text-sm text-primary">{hint}</div> : null}
            {error ? <div className="text-sm text-destructive">{error}</div> : null}

            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? "提交中..." : "注册"}
            </Button>
          </form>

          <div className="mt-6 text-sm text-muted-foreground">
            已有账号？{" "}
            <Link href="/login" className="text-primary hover:underline">
              去登录
            </Link>
          </div>
        </div>
      </main>
      <Footer />
    </div>
  )
}

