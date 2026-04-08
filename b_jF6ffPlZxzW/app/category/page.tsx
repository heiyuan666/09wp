import Link from "next/link"
import { fetchGameCategoryList } from "@/lib/api/game"

export default async function CategoryIndexPage() {
  const cats = await fetchGameCategoryList()
  return (
    <main className="mx-auto max-w-4xl px-4 py-10">
      <h1 className="text-2xl font-bold mb-6">游戏分类</h1>
      <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
        {(cats || []).map((c) => (
          <Link
            key={c.id}
            href={`/category/${c.id}`}
            className="rounded-lg border border-border bg-card px-4 py-3 hover:border-primary/50 transition-colors"
          >
            <div className="font-medium">{c.name}</div>
            {c.description ? <div className="text-xs text-muted-foreground mt-1 line-clamp-2">{c.description}</div> : null}
          </Link>
        ))}
      </div>
    </main>
  )
}

