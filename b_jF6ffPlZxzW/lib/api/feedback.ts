import { apiPost } from "@/lib/api/client"

export async function createGameResourceFeedback(input: {
  game_id: number
  game_resource_id: number
  download_url?: string
  extract_code?: string
  type: "link_invalid" | "other"
  content?: string
  contact?: string
}) {
  return apiPost<unknown>("/game/feedbacks", {
    game_id: input.game_id,
    game_resource_id: input.game_resource_id,
    download_url: input.download_url || "",
    extract_code: input.extract_code || "",
    type: input.type,
    content: input.content || "",
    contact: input.contact || "",
  })
}

