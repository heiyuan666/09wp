"use client"

import { useRouter } from "next/navigation"
import { useEffect, useState } from "react"
import { Header } from "@/components/game/header"
import { Footer } from "@/components/home/footer"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import Link from "next/link"
import { fetchUserProfile, changeUserPassword, type UserProfile } from "@/lib/api/user"
import { fetchMySubmissions, type UserSubmissionItem } from "@/lib/api/submission"

export default function MePage() {
  const router = useRouter()
  const [token, setToken] = useState("")
  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState("")

  const [oldPassword, setOldPassword] = useState("")
  const [newPassword, setNewPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [savingPwd, setSavingPwd] = useState(false)
  const [pwdMsg, setPwdMsg] = useState("")

  const [subs, setSubs] = useState<UserSubmissionItem[]>([])
  const [subsLoading, setSubsLoading] = useState(false)

  useEffect(() => {
    try {
      setToken(localStorage.getItem("token") || "")
    } catch {
      setToken("")
    }
  }, [])

  useEffect(() => {
    if (!token) {
      setProfile(null)
      setSubs([])
      return
    }
    setLoading(true)
    setError("")
    fetchUserProfile(token)
      .then((p) => setProfile(p))
      .catch((e) => setError(e instanceof Error ? e.message : "加载失败"))
      .finally(() => setLoading(false))
  }, [token])

  useEffect(() => {
    if (!token) return
    setSubsLoading(true)
    fetchMySubmissions(token)
      .then((items) => setSubs(items || []))
      .catch(() => {
        /* ignore */
      })
      .finally(() => setSubsLoading(false))
  }, [token])

  function logout() {
    try {
      localStorage.removeItem("token")
    } catch {
      /* ignore */
    }
    router.push("/")
    router.refresh()
  }

  async function onChangePassword() {
    setPwdMsg("")
    setError("")
    if (!token) {
      setError("请先登录")
      return
    }
    if (!oldPassword || !newPassword || !confirmPassword) {
      setPwdMsg("请填写完整密码信息")
      return
    }
    if (newPassword.length < 6) {
      setPwdMsg("新密码至少 6 位")
      return
    }
    if (newPassword !== confirmPassword) {
      setPwdMsg("两次输入的新密码不一致")
      return
    }
    setSavingPwd(true)
    try {
      await changeUserPassword({ token, oldPassword, newPassword, confirmPassword })
      setPwdMsg("密码已更新")
      setOldPassword("")
      setNewPassword("")
      setConfirmPassword("")
    } catch (e) {
      setPwdMsg(e instanceof Error ? e.message : "更新失败")
    } finally {
      setSavingPwd(false)
    }
  }

  return (
    <div className="min-h-screen bg-background">
      <Header />
      <main className="mx-auto max-w-3xl px-4 py-14 sm:px-6">
        <div className="rounded-xl border border-border bg-card p-6">
          <h1 className="text-2xl font-semibold text-foreground">我的账号</h1>
          {token ? (
            <>
              <p className="mt-2 text-sm text-muted-foreground">已登录</p>

              {loading ? <div className="mt-4 text-sm text-muted-foreground">加载中...</div> : null}
              {error ? <div className="mt-4 text-sm text-destructive">{error}</div> : null}

              {profile ? (
                <div className="mt-6 grid gap-6 md:grid-cols-2">
                  <div className="rounded-lg border border-border p-4">
                    <div className="text-sm font-medium text-foreground">个人信息</div>
                    <div className="mt-3 space-y-3 text-sm">
                      <div className="flex items-center justify-between gap-3">
                        <span className="text-muted-foreground">用户名</span>
                        <span className="text-foreground">{profile.username}</span>
                      </div>
                      <div className="flex items-center justify-between gap-3">
                        <span className="text-muted-foreground">昵称</span>
                        <span className="text-foreground">{profile.name || "-"}</span>
                      </div>
                      <div className="flex items-center justify-between gap-3">
                        <span className="text-muted-foreground">邮箱</span>
                        <span className="text-foreground">{profile.email || "-"}</span>
                      </div>
                      <div className="flex items-center justify-between gap-3">
                        <span className="text-muted-foreground">状态</span>
                        <span className="text-foreground">{profile.status || "-"}</span>
                      </div>
                      <div className="pt-2">
                        <div className="text-muted-foreground">简介</div>
                        <Textarea value={profile.bio || ""} readOnly className="mt-2" />
                      </div>
                      <div>
                        <div className="text-muted-foreground">标签</div>
                        <div className="mt-2 text-foreground">{profile.tags || "-"}</div>
                      </div>
                    </div>
                  </div>

                  <div className="rounded-lg border border-border p-4">
                    <div className="text-sm font-medium text-foreground">修改密码</div>
                    <div className="mt-3 space-y-3">
                      <Input
                        type="password"
                        placeholder="原密码"
                        value={oldPassword}
                        onChange={(e) => setOldPassword(e.target.value)}
                      />
                      <Input
                        type="password"
                        placeholder="新密码（至少 6 位）"
                        value={newPassword}
                        onChange={(e) => setNewPassword(e.target.value)}
                      />
                      <Input
                        type="password"
                        placeholder="确认新密码"
                        value={confirmPassword}
                        onChange={(e) => setConfirmPassword(e.target.value)}
                      />
                      {pwdMsg ? (
                        <div className={pwdMsg.includes("已") ? "text-sm text-primary" : "text-sm text-destructive"}>
                          {pwdMsg}
                        </div>
                      ) : null}
                      <div className="flex items-center justify-between gap-3">
                        <Button variant="secondary" onClick={logout}>
                          退出登录
                        </Button>
                        <Button onClick={onChangePassword} disabled={savingPwd}>
                          {savingPwd ? "保存中..." : "保存密码"}
                        </Button>
                      </div>
                    </div>
                  </div>
                </div>
              ) : null}

              <div className="mt-6 rounded-lg border border-border p-4">
                <div className="flex items-center justify-between gap-3">
                  <div className="text-sm font-medium text-foreground">我的投稿</div>
                  <Button asChild>
                    <Link href="/submit">我要投稿</Link>
                  </Button>
                </div>
                {subsLoading ? (
                  <div className="mt-3 text-sm text-muted-foreground">加载中...</div>
                ) : subs.length ? (
                  <div className="mt-3 space-y-2">
                    {subs.slice(0, 20).map((s) => (
                      <div key={s.id} className="rounded-md border border-border bg-secondary/20 p-3 text-sm">
                        <div className="flex items-center justify-between gap-3">
                          <div className="min-w-0">
                            <div className="truncate font-medium text-foreground">{s.title}</div>
                            <div className="mt-1 truncate text-xs text-muted-foreground">{s.link}</div>
                          </div>
                          <div className="flex flex-col items-end gap-1">
                            <span
                              className={
                                s.status === "approved"
                                  ? "text-xs text-primary"
                                  : s.status === "rejected"
                                    ? "text-xs text-destructive"
                                    : "text-xs text-yellow-500"
                              }
                            >
                              {s.status === "approved" ? "已通过" : s.status === "rejected" ? "已驳回" : "待审核"}
                            </span>
                            {s.game_id ? (
                              <Link className="text-xs text-muted-foreground underline" href={`/game/${s.game_id}`}>
                                去游戏
                              </Link>
                            ) : null}
                          </div>
                        </div>
                        {s.review_msg && s.status === "rejected" ? (
                          <div className="mt-2 text-xs text-destructive">驳回原因：{s.review_msg}</div>
                        ) : null}
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="mt-3 text-sm text-muted-foreground">暂无投稿记录</div>
                )}
              </div>
            </>
          ) : (
            <>
              <p className="mt-2 text-sm text-muted-foreground">当前未登录</p>
              <div className="mt-6">
                <Button onClick={() => router.push("/login")}>去登录</Button>
              </div>
            </>
          )}
        </div>
      </main>
      <Footer />
    </div>
  )
}

