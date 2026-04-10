import { Header } from "@/components/game/header"
import { Footer } from "@/components/home/footer"
import { fetchSoftwareDetail } from "@/lib/api/software"

export const revalidate = 120

export default async function SoftwareDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const data = await fetchSoftwareDetail(id)
  const software = data.software
  const versions = Array.isArray(data.versions) ? data.versions : []

  return (
    <div className="min-h-screen bg-background">
      <Header />
      <main className="mx-auto max-w-5xl px-4 sm:px-6 lg:px-8 py-8">
        <div className="rounded-xl border border-border bg-card p-6">
          <div className="flex flex-col md:flex-row gap-6">
            <img
              src={software.cover || software.cover_thumb || "/placeholder.svg"}
              alt={software.name}
              className="h-36 w-36 rounded-lg object-cover bg-muted"
            />
            <div className="flex-1">
              <h1 className="text-2xl font-bold text-foreground">{software.name}</h1>
              <p className="mt-2 text-sm text-muted-foreground">{software.summary || "暂无简介"}</p>
              <div className="mt-3 text-sm text-muted-foreground">
                版本：{software.version || "-"} · 平台：{software.platforms || "-"} · 大小：{software.size || "-"}
              </div>
              {software.website ? (
                <a className="mt-3 inline-block text-sm text-primary hover:underline" href={software.website} target="_blank" rel="noreferrer">
                  官方网站
                </a>
              ) : null}
            </div>
          </div>
        </div>

        <section className="mt-6 rounded-xl border border-border bg-card p-6">
          <h2 className="text-lg font-semibold text-foreground">版本列表</h2>
          {versions.length === 0 ? (
            <p className="mt-3 text-sm text-muted-foreground">暂无版本信息</p>
          ) : (
            <div className="mt-4 space-y-4">
              {versions.map((v) => (
                <div key={v.id} className="rounded-lg border border-border p-4">
                  <div className="font-medium text-foreground">{v.version}</div>
                  <div className="mt-1 text-xs text-muted-foreground">{v.published_at ? String(v.published_at).slice(0, 10) : "未填写发布时间"}</div>
                  {v.release_notes ? <p className="mt-2 text-sm text-muted-foreground whitespace-pre-wrap">{v.release_notes}</p> : null}
                  <div className="mt-3 space-y-1 text-sm">
                    {v.download_direct?.map((u, i) => (
                      <a key={`d-${i}`} href={u} target="_blank" rel="noreferrer" className="block text-primary hover:underline">
                        直链下载 {i + 1}
                      </a>
                    ))}
                    {v.download_pan?.map((u, i) => (
                      <a key={`p-${i}`} href={u} target="_blank" rel="noreferrer" className="block text-primary hover:underline">
                        网盘下载 {i + 1}
                      </a>
                    ))}
                    {v.download_extract ? <div className="text-muted-foreground">提取码：{v.download_extract}</div> : null}
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      </main>
      <Footer />
    </div>
  )
}
