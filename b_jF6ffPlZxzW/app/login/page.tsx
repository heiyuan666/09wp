"use client"

import { useRouter } from "next/navigation"
import Link from "next/link"
import { useState } from "react"
import { Header } from "@/components/game/header"
import { Footer } from "@/components/home/footer"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { loginUser } from "@/lib/api/auth"

export default function LoginPage() {
  const router = useRouter()
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState("")

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError("")
    setLoading(true)
    try {
      const res = await loginUser({ username: username.trim(), password })
      localStorage.setItem("token", res.token)
      router.push("/")
      router.refresh()
    } catch (err) {
      setError(err instanceof Error ? err.message : "登录失败")
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-background">
      <Header />
      <main className="mx-auto max-w-md px-4 py-14 sm:px-6">
        <div className="rounded-xl border border-border bg-card p-6">
          <h1 className="text-2xl font-semibold text-foreground">登录</h1>
          <p className="mt-2 text-sm text-muted-foreground">使用用户名或邮箱登录</p>

          <form className="mt-6 space-y-4" onSubmit={onSubmit}>
            <div className="space-y-2">
              <div className="text-sm font-medium text-foreground">账号</div>
              <Input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="用户名 / 邮箱" />
            </div>
            <div className="space-y-2">
              <div className="text-sm font-medium text-foreground">密码</div>
              <Input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="请输入密码" />
            </div>
            {error ? <div className="text-sm text-destructive">{error}</div> : null}
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? "登录中..." : "登录"}
            </Button>
          </form>

          <div className="mt-6 text-sm text-muted-foreground">
            还没有账号？{" "}
            <Link href="/register" className="text-primary hover:underline">
              去注册
            </Link>
          </div>
        </div>
      </main>
      <Footer />
    </div>
  )
}

