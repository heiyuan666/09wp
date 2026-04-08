import { apiGet, apiPost } from "@/lib/api/client"

export type UserSubmissionItem = {
  id: number
  user_id: number
  game_id?: number | null
  title: string
  link: string
  category_id: number
  description?: string
  extract_code?: string
  tags?: string
  status: "pending" | "approved" | "rejected" | string
  review_msg?: string
  created_at?: string
  updated_at?: string
}

export async function createUserSubmission(input: {
  token: string
  title: string
  link: string
  game_id?: number
  category_id?: number
  description?: string
  extract_code?: string
  tags?: string
}) {
  return apiPost<UserSubmissionItem>(
    "/user/submissions",
    {
      title: input.title,
      link: input.link,
      game_id: input.game_id || 0,
      category_id: input.category_id || 0,
      description: input.description || "",
      extract_code: input.extract_code || "",
      tags: input.tags || "",
    },
    { headers: { Authorization: `Bearer ${input.token}` } },
  )
}

export async function fetchMySubmissions(token: string) {
  return apiGet<UserSubmissionItem[]>("/user/submissions", {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  })
}

