import SubmitClient from "./submit-client"

export default async function SubmitPage({
  searchParams,
}: {
  searchParams?: Promise<Record<string, string | string[] | undefined>>
}) {
  const sp = (await searchParams) || {}
  const raw = Array.isArray(sp.game_id) ? sp.game_id[0] : sp.game_id
  const n = Number((raw || "").trim())
  const prefillGameId = Number.isFinite(n) && n > 0 ? n : 0
  return <SubmitClient prefillGameId={prefillGameId} />
}

