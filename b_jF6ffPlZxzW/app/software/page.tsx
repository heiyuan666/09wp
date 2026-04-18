import Link from "next/link"
import { Header } from "@/components/game/header"
import { Footer } from "@/components/home/footer"
import { fetchSoftwareList } from "@/lib/api/software"

export const revalidate = 120

export default async function SoftwarePage() {
  const data = await fetchSoftwareList({ page: 1, page_size: 60 })
  const list = Array.isArray(data.list) ? data.list : []

  return (
    <div className="min-h-screen bg-background">
      <Header />
      <main className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-6">
          <h1 className="text-2xl md:text-3xl font-bold text-foreground">软件库</h1>
          <p className="mt-2 text-sm text-muted-foreground">精选常用软件，支持多版本与多下载地址</p>
        </div>

        {list.length === 0 ? (
          <div className="rounded-lg border border-border bg-card p-8 text-sm text-muted-foreground">暂无软件数据</div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {list.map((item) => (
              <Link
                key={item.id}
                href={`/software/${item.id}`}
                className="rounded-lg border border-border bg-card p-4 hover:border-primary/40 transition-colors"
              >
                <div className="flex gap-3">
                  <img
                    src={
                      item.icon_thumb ||
                      item.icon ||
                      item.cover_thumb ||
                      item.cover ||
                      "/placeholder.svg"
                    }
                    alt={item.name}
                    className="h-16 w-16 rounded object-cover bg-muted"
                  />
                  <div className="min-w-0 flex-1">
                    <h2 className="truncate font-semibold text-foreground">{item.name}</h2>
                    <p className="mt-1 text-xs text-muted-foreground line-clamp-2">{item.summary || "暂无简介"}</p>
                    <div className="mt-2 text-xs text-muted-foreground">
                      版本：{item.version || "-"} · 大小：{item.size || "-"}
                    </div>
                  </div>
                </div>
              </Link>
            ))}
          </div>
        )}
      </main>
      <Footer />
    </div>
  )
}
