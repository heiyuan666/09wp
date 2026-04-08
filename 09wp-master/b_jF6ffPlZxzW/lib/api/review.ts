import { apiGet, apiPost } from "@/lib/api/client"

export type GameReviewUser = {
  id: number
  username: string
  avatar?: string | null
}

export type GameReviewItem = {
  id: number
  game_id: number
  rating: number
  content: string
  helpful: number
  unhelpful: number
  created_at: string
  user: GameReviewUser
}

export type GameReviewDistributionItem = {
  stars: number
  count: number
  percentage: number
}

export type GameReviewListResp = {
  list: GameReviewItem[]
  total: number
  average: number
  distribution: GameReviewDistributionItem[]
}

export async function fetchGameReviews(params: { game_id: number; page?: number; page_size?: number; sort?: "recent" | "helpful" }) {
  const qs = new URLSearchParams()
  qs.set("game_id", String(params.game_id))
  if (params.page) qs.set("page", String(params.page))
  if (params.page_size) qs.set("page_size", String(params.page_size))
  if (params.sort) qs.set("sort", params.sort)
  return apiGet<GameReviewListResp>(`/game/reviews?${qs.toString()}`)
}

export async function createGameReview(input: { token: string; game_id: number; rating: number; content: string }) {
  return apiPost<{ id: number }>(
    "/game/reviews",
    { game_id: input.game_id, rating: input.rating, content: input.content },
    { headers: { Authorization: `Bearer ${input.token}` } },
  )
}

export async function voteGameReview(input: { token: string; review_id: number; vote: -1 | 0 | 1 }) {
  return apiPost<{ ok: boolean }>(
    `/game/reviews/${input.review_id}/vote`,
    { vote: input.vote },
    { headers: { Authorization: `Bearer ${input.token}` } },
  )
}

