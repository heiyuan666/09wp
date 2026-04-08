"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import Link from "next/link"
import { Header } from "@/components/game/header"
import { Footer } from "@/components/home/footer"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { createUserSubmission } from "@/lib/api/submission"

export default function SubmitClient({ prefillGameId }: { prefillGameId: number }) {
  const router = useRouter()

  const [token, setToken] = useState("")
  const [title, setTitle] = useState("")
  const [link, setLink] = useState("")
  const [gameId, setGameId] = useState<number>(0)
  const [extractCode, setExtractCode] = useState("")
  const [tags, setTags] = useState("")
  const [description, setDescription] = useState("")

  const [submitting, setSubmitting] = useState(false)
  const [msg, setMsg] = useState("")

  useEffect(() => {
    try {
      setToken(localStorage.getItem("token") || "")
    } catch {
      setToken("")
    }
  }, [])

  useEffect(() => {
    if (prefillGameId > 0) setGameId(prefillGameId)
  }, [prefillGameId])

  async function onSubmit() {
    setMsg("")
    if (!token) {
      setMsg("请先登录后再投稿")
      return
    }
    const t = title.trim()
    const l = link.trim()
    if (!t || !l) {
      setMsg("请填写标题和网盘链接")
      return
    }
    if (!gameId || gameId <= 0) {
      setMsg("请填写游戏 ID（投稿到某个游戏详情页）")
      return
    }

    setSubmitting(true)
    try {
      await createUserSubmission({
        token,
        title: t,
        link: l,
        game_id: gameId,
        extract_code: extractCode.trim(),
        tags: tags.trim(),
        description: description.trim(),
      })
      setMsg("提交成功：等待审核")
      setTitle("")
      setLink("")
      setExtractCode("")
      setTags("")
      setDescription("")
      router.push("/me")
      router.refresh()
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "提交失败")
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="min-h-screen bg-background">
      <Header />
      <main className="mx-auto max-w-3xl px-4 py-14 sm:px-6">
        <div className="rounded-xl border border-border bg-card p-6">
          <h1 className="text-2xl font-semibold text-foreground">提交网盘资源</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            提交后会进入审核流程，审核通过后会显示在对应游戏的详情页下载区。
          </p>

          {!token ? (
            <div className="mt-6 rounded-lg border border-border bg-secondary/30 p-4">
              <div className="text-sm text-foreground">当前未登录</div>
              <div className="mt-3 flex gap-3">
                <Button asChild>
                  <Link href="/login">去登录</Link>
                </Button>
                <Button variant="secondary" asChild>
                  <Link href="/register">去注册</Link>
                </Button>
              </div>
            </div>
          ) : null}

          <div className="mt-6 space-y-4">
            <div className="grid gap-3 md:grid-cols-2">
              <Input placeholder="资源标题（必填）" value={title} onChange={(e) => setTitle(e.target.value)} />
              <Input
                placeholder="游戏 ID（必填，可从详情页 URL 获取）"
                value={gameId ? String(gameId) : ""}
                onChange={(e) => setGameId(Number(e.target.value || 0))}
                inputMode="numeric"
              />
            </div>

            <Input placeholder="网盘链接（必填）" value={link} onChange={(e) => setLink(e.target.value)} />

            <div className="grid gap-3 md:grid-cols-2">
              <Input
                placeholder="提取码（可选）"
                value={extractCode}
                onChange={(e) => setExtractCode(e.target.value)}
              />
              <Input
                placeholder="标签（可选，逗号分隔）"
                value={tags}
                onChange={(e) => setTags(e.target.value)}
              />
            </div>

            <Textarea
              placeholder="补充说明（可选）"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="min-h-28"
            />

            {msg ? (
              <div className={msg.includes("成功") ? "text-sm text-primary" : "text-sm text-destructive"}>{msg}</div>
            ) : null}

            <div className="flex items-center justify-between gap-3">
              <Button variant="secondary" onClick={() => router.push("/me")}>
                返回个人中心
              </Button>
              <Button onClick={onSubmit} disabled={submitting}>
                {submitting ? "提交中..." : "提交审核"}
              </Button>
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </div>
  )
}

