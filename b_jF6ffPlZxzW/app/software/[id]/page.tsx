import { Header } from "@/components/game/header"
import { Footer } from "@/components/home/footer"
import { fetchSoftwareDetail } from "@/lib/api/software"
import { netdiskLinkLabel, sortNetdiskLinks } from "@/lib/netdiskLabel"

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
              src={
                software.cover ||
                software.cover_thumb ||
                software.icon_thumb ||
                software.icon ||
                "/placeholder.svg"
              }
              alt={software.name}
              className="h-36 w-36 rounded-lg object-cover bg-muted"
            />
            <div className="flex-1">
              <div className="flex flex-wrap items-center gap-3">
                {(software.icon_thumb || software.icon) && (software.cover || software.cover_thumb) ? (
                  <img
                    src={software.icon_thumb || software.icon || "/placeholder.svg"}
                    alt=""
                    className="h-10 w-10 shrink-0 rounded-md object-cover bg-muted border border-border"
                  />
                ) : null}
                <h1 className="text-2xl font-bold text-foreground">{software.name}</h1>
              </div>
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

        {((software.download_direct?.length ?? 0) > 0 || (software.download_pan?.length ?? 0) > 0 || software.download_extract) ? (
          <section className="mt-6 rounded-xl border border-border bg-card p-6">
            <h2 className="text-lg font-semibold text-foreground">下载地址</h2>
            <div className="mt-4 space-y-3 text-sm">
              {software.download_direct?.map((u, i) => (
                <a key={`sd-${i}`} href={u} target="_blank" rel="noreferrer" className="block text-primary hover:underline">
                  直链下载 {i + 1}
                </a>
              ))}
              {sortNetdiskLinks(software.download_pan).map((u, i) => (
                <div key={`sp-${i}`} className="flex flex-wrap items-center gap-2 rounded-lg border border-border p-2">
                  <span className="rounded-md bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">{netdiskLinkLabel(u)}</span>
                  <a href={u} target="_blank" rel="noreferrer" className="text-primary hover:underline">
                    打开分享
                  </a>
                </div>
              ))}
              {software.download_extract ? <div className="text-muted-foreground">提取码：{software.download_extract}</div> : null}
            </div>
          </section>
        ) : null}

        <section className="mt-6 rounded-xl border border-border bg-card p-6">
          <h2 className="text-lg font-semibold text-foreground">版本列表</h2>
          {versions.length === 0 ? (
            <p className="mt-3 text-sm text-muted-foreground">暂无版本信息</p>
          ) : (
            <div className="mt-4 space-y-4">
              {versions.map((v) => {
                const panLinks = sortNetdiskLinks(v.download_pan)
                return (
                <div key={v.id} className="rounded-lg border border-border p-4">
                  <div className="font-medium text-foreground">{v.version}</div>
                  <div className="mt-1 text-xs text-muted-foreground">{v.published_at ? String(v.published_at).slice(0, 10) : "未填写发布时间"}</div>
                  {v.release_notes ? <p className="mt-2 text-sm text-muted-foreground whitespace-pre-wrap">{v.release_notes}</p> : null}
                  <div className="mt-3 space-y-2 text-sm">
                    {v.download_direct?.map((u, i) => (
                      <a key={`d-${i}`} href={u} target="_blank" rel="noreferrer" className="block text-primary hover:underline">
                        直链下载 {i + 1}
                      </a>
                    ))}
                    {panLinks.length > 0 ? (
                      <div className="text-xs font-medium text-muted-foreground">网盘下载</div>
                    ) : null}
                    {panLinks.map((u, i) => (
                      <div key={`p-${i}`} className="flex flex-wrap items-center gap-2 rounded-lg border border-border p-2">
                        <span className="rounded-md bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">{netdiskLinkLabel(u)}</span>
                        <a href={u} target="_blank" rel="noreferrer" className="text-primary hover:underline">
                          打开分享
                        </a>
                      </div>
                    ))}
                    {v.download_extract ? <div className="text-muted-foreground">提取码：{v.download_extract}</div> : null}
                  </div>
                </div>
                )
              })}
            </div>
          )}
        </section>
      </main>
      <Footer />
    </div>
  )
}
